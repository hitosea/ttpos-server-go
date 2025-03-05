package service

import (
	"github.com/spf13/viper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IRechargePrintSrv 定义打印充值票据服务接口
type IRechargePrintSrv interface {
	PrintTicket(ctx context.Context, printReq PrinterTicketReq) (resp.PrinterLogData, error) // 打印票据
}

// rechargePrintSrv  打印充值票据服务结构体
type rechargePrintSrv struct {
	dbm           *database.DBManager // 数据库管理器
	settingSrv    setting.ISrv
	printerLogSrv IPrinterLogSrv
}

func NewRechargePrinterSrv(dbm *database.DBManager, settingSrv setting.ISrv, printerLogSrv IPrinterLogSrv) IRechargePrintSrv {
	return NewRechargePrinterSrvImpl(dbm, settingSrv, printerLogSrv)
}

func NewRechargePrinterSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv, printerLogSrv IPrinterLogSrv) IRechargePrintSrv {
	return &rechargePrintSrv{
		dbm:           dbm,
		settingSrv:    settingSrv,
		printerLogSrv: printerLogSrv,
	}
}

type PrinterTicketReq struct {
	RechargeOrder model.MemberRechargeOrder
	IsQueue       bool
	DeviceId      string
	PrintLang     string
}

// PrintTicket 打印票据
func (s *rechargePrintSrv) PrintTicket(ctx context.Context, printReq PrinterTicketReq) (resp.PrinterLogData, error) {
	var printerLogData resp.PrinterLogData
	// 获取打印设置
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		return printerLogData, errors.WithMessage(err)
	}
	settingPrinterInfo, err := s.settingSrv.GetPrinterInfo(ctx, printerSetting, printReq.DeviceId)
	if err != nil {
		return printerLogData, errors.New("未开启打印, 请联系管理员")
	}
	cashierBindKey := settingPrinterInfo.CashierBindKey

	// 主动点击打印-但未开启打印
	if !printReq.IsQueue && settingPrinterInfo.IsCashierOpen {
		return printerLogData, errors.New("未开启打印, 请联系管理员")
	}
	// 主动点击打印-但未配置打印机
	if !printReq.IsQueue && settingPrinterInfo.PrinterType == "" {
		return printerLogData, errors.New("未配置打印机, 请联系管理员")
	}
	// 如果是云打印, cashierBindKey 等于自己
	if printReq.DeviceId != "" && settingPrinterInfo.IsCashierPrinter && viper.GetBool("IS_CLOUD_DEPLOY") {
		cashierBindKey = printReq.DeviceId
	}
	if settingPrinterInfo.IsCashierOpen {
		var printerType, content string
		// ToDo 根据品牌获取打印内容
		content = s.getPrintContent(printReq.RechargeOrder, printerType)

		printerLogType := constant.PrinterLogTypeDefault
		if settingPrinterInfo.IsCashierPrinter {
			printerLogType = constant.PrinterLogTypeCloud
		}
		firstExecution := 0
		if !printReq.IsQueue {
			firstExecution = 1
		}
		// 添加打印日志，依赖打印日志服务
		printerLogData, err = s.printerLogSrv.AddLog(ctx, resp.PrinterInfo{
			PrinterType:   settingPrinterInfo.PrinterType,
			PrinterConfig: settingPrinterInfo.PrinterConfig,
			PrintCopies:   settingPrinterInfo.Copies,
		}, resp.PrinterLogData{
			PrinterUuid:     settingPrinterInfo.PrinterUuid,
			CashierDeviceId: cashierBindKey,
			DataType:        constant.PrinterLogDataTypeRecharge,
			Data:            content,
			Type:            printerLogType,
			FirstExecution:  firstExecution,
		}, printReq.DeviceId)
		if err != nil {
			return printerLogData, errors.New("打印失败，未连接打印机")
		}
		return printerLogData, nil
	}
	return printerLogData, nil
}

// getPrintContent 获取打印内容
func (s *rechargePrintSrv) getPrintContent(rechargeOrder model.MemberRechargeOrder, printerType string) string {

	// ToDo 获取不同类型打印机的打印模板

	return ""
}
