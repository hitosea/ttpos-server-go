package service

import (
	"errors"
	"github.com/spf13/viper"
	"slices"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IPrinterLogSrv 定义打印日志服务接口
type IPrinterLogSrv interface {
	AddLog(ctx context.Context, printer model.Printer, printerLog model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) // 添加打印日志
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

// AddLog 添加打印日志
func (s *printerLogSrv) AddLog(ctx context.Context, printer model.Printer, printerLog model.PrinterLog, controlDeviceId string) (model.PrinterLog, error) {

	var respPrinterLog model.PrinterLog
	// 是否本机
	isLocal := printerLog.CashierDeviceId == controlDeviceId

	// 如果是点餐助手操作的，就永远不是本机
	if ctx.GetSource() == constant.SourceAssistant {
		isLocal = false
	}

	// 是否队列服务
	var isQueueService bool

	// 如果是局域网部署 - 就都下放打印
	if viper.GetBool("IS_CLOUD_DEPLOY") {
		printerLog.Type = constant.PrinterLogTypeCloud
	} else if !isLocal && printerLog.PrinterUuid > 0 {
		isQueueService = true
	}

	// 获取商家设置，判断是否开启本地打印
	companySetting := ctx.GetCompanySetting()
	if printer.PrinterType != nil && printer.PrinterType.Key == constant.PrinterTypeSunmiCloud && companySetting.IsOpenLocalPrint == 1 {
		isQueueService = true
		printerLog.Type = constant.PrinterLogTypeCloud
		printerLog.FirstExecution = 0
	}

	if printer.Uuid == 0 {
		printerLog.Status = constant.PrinterLogStatusEnd
		printerLog.Reason = "打印机不存在"
	} else {
		printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
		if err != nil {
			return respPrinterLog, err
		}
		// 默认收银打印
		printerLog.PrintMethod = constant.PrinterLogPintMethodText
		printMethodStr := printerSetting.PrintMethod
		// 这几种是送厨打印
		if slices.Contains([]int{constant.PrinterLogDataTypeOneDishOneMenu, constant.PrinterLogDataTypeEntireOrder, constant.PrinterLogDataTypeReturnDish}, printerLog.DataType) {
			printMethodStr = printerSetting.KitchenPrintMethod
		}
		printMethod, _ := strconv.Atoi(printMethodStr)
		if printMethod > 0 && slices.Contains([]int{constant.PrinterLogPintMethodText, constant.PrinterLogPintMethodImage}, printMethod) {
			printerLog.PrintMethod = printMethod
		}

		if !isQueueService { // 直接打印
			if printerLog.PrinterUuid > 0 && printerLog.Type == constant.PrinterLogTypeDefault && printerLog.FirstExecution == 1 {
				// todo 打印驱动，打印小票，返回错误
				if err := errors.New("123123"); err != nil {
					printerLog.Status = constant.PrinterLogStatusEnd
					printerLog.Reason = err.Error()
				} else {
					printerLog.Status = constant.PrinterLogStatusSuccess
					printerLog.Reason = "打印成功"
				}
				printerLog.Num = 1
			} else if printerLog.Type == constant.PrinterLogTypeCloud && printerLog.FirstExecution == 1 && isLocal {
				printerLog.Status = constant.PrinterLogStatusSuccess
			}
		}
	}

	printerLog.PrinterTime = time.Now().Unix()

	printerLogRepo := repository.NewPrinterLogRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	printerLog, err := printerLogRepo.Create(printerLog)
	if err != nil {
		return model.PrinterLog{}, err
	}

	// ToDo 添加打印日志
	return model.PrinterLog{}, nil

}
