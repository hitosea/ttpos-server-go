package tasks

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"
)

// 秒级任务
type usbPrintTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewUsbPrintTask 创建新的秒级任务
func NewUsbPrintTask(dbm *database.DBManager, cache cache.Cache) *usbPrintTask {
	return &usbPrintTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行任务
func (t *usbPrintTask) Execute() {
	// 根据 APP 表实例化数据库连接，并左关联 CompanySetting
	prefix := config.Database.TablePrefix
	var companies []model.Company
	if err := t.dbm.GetDB(0).
		Joins(fmt.Sprintf("LEFT JOIN %scompany_setting ON %scompany.uuid = %scompany_setting.company_uuid", prefix, prefix, prefix)).
		Where(fmt.Sprintf("%scompany.delete_time = 0", prefix)).
		Select(fmt.Sprintf("%scompany.uuid, %scompany.name", prefix, prefix)). // 选择需要的字段
		Find(&companies).Error; err != nil {
		log.Fatalf("Error querying companies with settings: %s", err)
		return
	}
	//
	for _, company := range companies {
		// 获取打印机列表
		usbPrinterLists, err := base.NewPrinterRepo(t.dbm.GetDB(company.Uuid)).GetUsbPrinterList()
		if err != nil {
			logger.Logger.Error("获取打印机列表失败", zap.Error(err))
			fmt.Println("获取打印机列表失败", zap.Error(err))
			return
		}
		if len(usbPrinterLists) > 0 {
			printerSetting, err := setting.NewSrv(t.dbm, t.cache).GetPrinterSetting(context.NewContext(
				context.WithCompanyUuid(company.Uuid),
				context.WithLogger(logger.Logger),
			), []dto.LanguageItem{})
			if err != nil {
				logger.Logger.Error("获取打印设置失败", zap.Error(err))
				fmt.Println("获取打印设置失败", zap.Error(err))
				return
			}
			//
			isUpdate := false
			for _, printer := range usbPrinterLists {
				// 打印机列表
				for i := range printerSetting.CashierPrinter {
					if fmt.Sprintf("%d", printer.Uuid) == printerSetting.CashierPrinter[i].PrinterUsbId {
						printerSetting.CashierPrinter[i].PrinterUsbId = "0"
						isUpdate = true
						// 推送更新打印机列表
						fmt.Println("打印机离线五分钟", printer.SourceDeviceSn)
						uuid, err := strconv.ParseUint(utils.Uint64OrStringToString(printerSetting.CashierPrinter[i].PrinterId), 10, 64)
						if err != nil {
							logger.Logger.Error("打印机ID格式错误", zap.Error(err))
							fmt.Println("打印机ID格式错误", zap.Error(err))
						}
						newPrinter, err := base.NewPrinterRepo(t.dbm.GetDB(company.Uuid)).GetPrinter(uint(uuid))
						if err != nil {
							logger.Logger.Error("获取打印机列表失败", zap.Error(err))
							fmt.Println("获取打印机列表失败", zap.Error(err))
						}
						// 推送更新打印机列表
						utils.Go(func() {
							websocket.PushClient(company.Uuid, websocket.SourceCashier, printer.SourceDeviceSn, websocket.UPDATE_SELECTED_PRINTER, map[string]interface{}{
								"type":             "update",
								"update_time":      time.Now().Unix(),
								"old_printer_name": printer.Name,
								"old_printer_sn":   printer.GetConfigJson().SN,
								"new_printer_name": utils.IfString(err == nil && newPrinter.Name != "", newPrinter.Name, "Built-in Printer"),
								"new_printer_sn":   "",
							})
						})
					}
				}
			}
			// 更新打印机列表
			if isUpdate {
				repository.NewSettingRepo(t.dbm.GetDB(company.Uuid)).Updates(constant.SettingPrinter, utils.ToJson(printerSetting))
				t.cache.Del(fmt.Sprintf("setting:company_id:%d", company.Uuid))
			}
		}
	}
}
