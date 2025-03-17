package service

import (
	"encoding/json"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IPrinterLogSrv 定义打印日志服务接口
type IPrinterLogSrv interface {
	AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) // 添加打印日志
	GetPrinterList(ctx context.Context) ([]resp.PrinterLogData, error)                                                                       // 获取打印列表
	GetPrinterData(ctx context.Context) (*resp.PrinterDataList, error)                                                                       // 获取打印数据
}

type printerLogSrv struct {
	dbm        *database.DBManager
	settingSrv setting.ISrv
}

func NewPrinterLogSrv(dbm *database.DBManager, settingSrv setting.ISrv) IPrinterLogSrv {
	return NewPrinterLogSrvImpl(dbm, settingSrv)
}

func NewPrinterLogSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IPrinterLogSrv {
	return &printerLogSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

// GetPrinterList 获取打印列表
func (s *printerLogSrv) GetPrinterList(ctx context.Context) ([]resp.PrinterLogData, error) {
	// // 获取打印日志
	// printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	// printerLogList, _, err := printerLogRepo.PaginateGet(
	// 	1,
	// 	5,
	// 	printerLogRepo.WithPrinter(),
	// 	printerLogRepo.WithPrinterPrinterType(),
	// 	printerLogRepo.WithSaleBill(),
	// 	printerLogRepo.WithSaleBillDesk(),
	// 	printerLogRepo.WhereStatus(1),
	// 	printerLogRepo.WherePrinterTime(),
	// 	func(db *gorm.DB) *gorm.DB {
	// 		db.Where("print_method = ?", 2)
	// 		db.Where("first_execution = ?", 0)
	// 		db.Where("type = ?", constant.Yes)
	// 		return db
	// 	},
	// )
	// if err != nil {
	// 	return nil, errors.WithMessage(err)
	// }

	// // 转换为响应数据
	// results := make([]resp.PrinterLogData, 0, len(printerLogList))
	// for _, log := range printerLogList {
	// 	results = append(results, resp.PrinterLogData{
	// 		Id:          log.Id,
	// 		Data:        log.Data,
	// 		Reason:      log.Reason,
	// 		PrinterId:   log.PrinterUuid,
	// 		OrderId:     log.RelatedUuid,
	// 		CreateTime:  log.CreateTime,
	// 		PrintMethod: log.PrintMethod,
	// 	})
	// }

	// return results, nil

	return nil, nil
}

// GetPrinterData 获取打印数据
func (s *printerLogSrv) GetPrinterData(ctx context.Context) (*resp.PrinterDataList, error) {
	companyUuid := ctx.GetCompanyUuid()

	//禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(companyUuid)
		defer lock.NewSystemLock().UnlockUuid(companyUuid)
		ctx.AddLock()
	}

	// 获取打印日志
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(companyUuid))
	printerLogList, _, err := printerLogRepo.PaginateGet(
		1,
		5,
		printerLogRepo.WithPrinter(),
		printerLogRepo.WithPrinterPrinterType(),
		printerLogRepo.WhereType(1),
		printerLogRepo.WhereStatus(0),
		printerLogRepo.WherePrintMethod(2),
		printerLogRepo.WhereFirstExecution(0),
		func(db *gorm.DB) *gorm.DB {
			// 相同设备的
			db.Where("(cashier_device_id = ? OR cashier_device_id = '')", ctx.GetDeviceSn())
			// 0次或1次
			db.Where("(num in (0, 1))")
			// 1天内
			db.Where("(create_time > UNIX_TIMESTAMP() - 86400)")
			//
			return db
		},
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 转换为响应数据
	printerDataList := make([]resp.PrinterData, 0, len(printerLogList))
	for _, log := range printerLogList {
		printerDataList = append(printerDataList, resp.PrinterData{
			Uuid:        log.Uuid,
			Data:        log.Data,
			PrintMethod: log.PrintMethod,
			Copies:      log.Printer.Copies,
			PrinterType: log.Printer.PrinterType.Key,
			PrinterConfig: func() string {
				configJson, err := json.Marshal(log.Printer.GetConfigJson())
				if err != nil {
					return ""
				}
				return string(configJson)
			}(),
		})
	}

	return &resp.PrinterDataList{List: printerDataList}, nil
}

// AddLog 添加打印日志
func (s *printerLogSrv) AddLog(ctx context.Context, printer resp.PrinterInfo, printerLogData model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) {
	// 获取打印日志仓库
	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	// 标记进行中
	printerLogData.Status = constant.PrinterLogStatusInProgress
	// 获取商家设置，判断是否开启本地打印
	companySetting := ctx.GetCompanySetting()
	// 如果是商米云打印 - 就都队列打印
	if printer.PrinterType == constant.PrinterTypeSunmiCloud && companySetting.IsOpenLocalPrint == 0 {
		printerLogData.Type = constant.PrinterLogTypeDefault
		printerLogData.FirstExecution = 0
	}
	// 打印机不存在
	if printer.PrinterType == "" {
		printerLogData.Status = constant.PrinterLogStatusEnd
		printerLogData.Reason = "打印机不存在"
	}
	// 保存数据
	printerLogData.PrinterTime = time.Now().Unix()
	var printerLog model.PrinterLog
	copier.Copy(&printerLog, printerLogData)
	printerLog.Data = printerLog.CompressData()
	printerLog, err := printerLogRepo.Create(printerLog)
	if err != nil {
		logger.Logger.Error("保存数据打印日志失败", zap.Error(err))
		return model.PrinterLog{}, errors.WithMessage(err)
	}

	// 只保留7天的数据
	go func() {
		err = printerLogRepo.UpdateByWhere(map[string]any{"delete_time": time.Now().Unix()}, printerLogRepo.WhereCreatedBefore(7))
		if err != nil {
			logger.Logger.Error("删除n天前的打印日志失败", zap.Error(err))
		}
	}()

	return printerLog, nil
}
