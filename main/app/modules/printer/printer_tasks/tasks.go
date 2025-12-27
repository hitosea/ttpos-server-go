package printer_tasks

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// 秒级任务
type printerTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// 创建新的秒级任务
func NewPrinterTask(dbm *database.DBManager, cache cache.Cache) *printerTask {
	return &printerTask{
		dbm:   dbm,
		cache: cache,
	}
}

// 执行任务
func (t *printerTask) Execute() {
	// 根据 APP 表实例化数据库连接，并左关联 CompanySetting
	prefix := config.Database.TablePrefix
	var companies []model.Company
	if err := t.dbm.GetDB(0).
		Joins(fmt.Sprintf("LEFT JOIN %scompany_setting ON %scompany.uuid = %scompany_setting.company_uuid", prefix, prefix, prefix)).
		Where(fmt.Sprintf("%scompany.delete_time = 0", prefix)).
		Where(fmt.Sprintf("%scompany_setting.is_open_local_print = 0", prefix)).
		Select(fmt.Sprintf("%scompany.uuid, %scompany.name, %scompany_setting.is_open_local_print", prefix, prefix, prefix)). // 选择需要的字段
		Find(&companies).Error; err != nil {
		log.Fatalf("Error querying companies with settings: %s", err)
		return
	}

	// 每5个公司为一组进行处理
	for i := 0; i < len(companies); i += 5 {
		// 计算当前批次的结束索引
		end := i + 5
		if end > len(companies) {
			end = len(companies)
		}
		// 处理当前批次的公司
		utils.Go(func() {
			batch := companies[i:end]
			for _, company := range batch {
				// 使用闭包捕获变量
				t.sendPrinter(company.Uuid)
			}
		})
	}
}

// 执行任务
func (t *printerTask) sendPrinter(companyUuid uint64) {
	// 获取打印日志
	printerLogRepo := repository.NewPrinterLogRepo(t.dbm.GetDB(companyUuid))
	printerLogList, err := printerLogRepo.GetPrinterLogList(
		printerLogRepo.WithPrinter(),
		printerLogRepo.WhereStatus(1),
		printerLogRepo.WherePrinterTime(),
		printerLogRepo.WhereLimit(5),
		func(db *gorm.DB) *gorm.DB {
			if viper.GetString("CHECK_PRINT") != "false" {
				db.Where("print_method = ?", 2)
				db.Where("first_execution = ?", 0)
				db.Where("type = ?", constant.No)
			}
			return db
		},
	)
	if err != nil {
		log.Fatalf("执行打印任务失败 : %s", err)
		return
	}
	if len(printerLogList) == 0 {
		return
	}
	// 执行打印
	for _, printerLog := range printerLogList {
		t.ExecutePrinter(companyUuid, printerLog)
	}
}

// 执行任务
func (t *printerTask) ExecutePrinter(companyUuid uint64, printerLog model.PrinterLog) {
	printerLog.Num = printerLog.Num + 1
	if printerLog.Printer != nil {
		copies := printerLog.Copies
		if copies == 0 {
			copies = 1
		}
		content := func() string {
			if printerLog.PrinterLogData != nil {
				return printerLog.PrinterLogData.DecompressData()
			}
			if printerLog.Data != "-" && printerLog.Data != "" {
				return printerLog.DecompressData()
			}
			return ""
		}()
		configJson := printerLog.Printer.GetConfigJson()
		for i := uint(0); i < copies; i++ {
			// 执行打印
			var err error
			if slices.Contains([]string{
				constant.PrinterTypeSunmiLan,
				constant.PrinterTypeSunmiCloud,
				constant.PrinterTypeCashierSunmi,
			}, printerLog.PrinterType) {
				err = pkg.PrintSunmiTicket(configJson, content, printerLog.GetTradeNo(companyUuid))
			} else {
				err = pkg.PrintTicket(configJson.IP, configJson.PORT, content, printerLog.PrintMethod)
			}
			if err == nil {
				// 打印成功
				printerLog.Reason = "打印成功"
				printerLog.Status = 2
			} else {
				// 打印失败
				printerLog.Reason = err.Error()
				if printerLog.Num >= 3 {
					printerLog.Status = 0
				} else {
					printerLog.Status = 1
				}
			}
		}
	} else {
		// 打印机不存在
		printerLog.Reason = "The printer does not exist and has been deleted"
		printerLog.Status = 0
	}

	// 更新打印日志
	if err := repository.NewPrinterLogRepo(t.dbm.GetDB(companyUuid)).Update(printerLog.Uuid, map[string]any{
		"reason":       printerLog.Reason,
		"status":       printerLog.Status,
		"num":          printerLog.Num,
		"printer_time": time.Now().Unix(),
	}); err != nil {
		log.Fatalf("更新打印日志失败 : %s", err)
		return
	}

	// 递归执行
	if printerLog.Status == 1 && printerLog.Num < 3 {
		utils.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("打印任务panic: %v", r)
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			select {
			case <-time.After(3 * time.Second):
				t.ExecutePrinter(companyUuid, printerLog)
			case <-ctx.Done():
				log.Printf("打印任务超时，取消重试")
				return
			}
		})
	}
}
