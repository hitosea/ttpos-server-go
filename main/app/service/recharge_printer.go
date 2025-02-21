package service

import (
	"errors"
	"github.com/spf13/viper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IRechargePrinterSrv 定义打印充值票据服务接口
type IRechargePrinterSrv interface {
	PrintTicket(ctx context.Context, param PrinterTicketParam) (model.PrinterLog, error) // 打印票据
}

// rechargePrinterSrv  打印充值票据服务结构体
type rechargePrinterSrv struct {
	dbm           *database.DBManager // 数据库管理器
	settingSrv    setting.ISrv
	printerLogSrv IPrinterLogSrv
}

func NewRechargePrinterSrv(dbm *database.DBManager, settingSrv setting.ISrv, printerLogSrv IPrinterLogSrv) IRechargePrinterSrv {
	return NewRechargePrinterSrvImpl(dbm, settingSrv, printerLogSrv)
}

func NewRechargePrinterSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv, printerLogSrv IPrinterLogSrv) IRechargePrinterSrv {
	return &rechargePrinterSrv{
		dbm:           dbm,
		settingSrv:    settingSrv,
		printerLogSrv: printerLogSrv,
	}
}

type PrinterTicketParam struct {
	RechargeOrder model.MemberRechargeOrder
	IsQueue       bool
	DeviceId      string
	PrintLang     string
}

// PrintTicket 打印票据
func (s *rechargePrinterSrv) PrintTicket(ctx context.Context, param PrinterTicketParam) (model.PrinterLog, error) {
	var printerLog model.PrinterLog
	// 获取打印设置
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		return printerLog, err
	}

	printerInfo := s.settingSrv.GetPrinterInfo(ctx, printerSetting, param.DeviceId)
	cashierBindKey := printerInfo.CashierBindKey
	printerType := printerInfo.PrinterBrand
	printer := printerInfo.Printer

	if !param.IsQueue && printerInfo.IsCashierOpen {
		return printerLog, errors.New("未开启打印, 请联系管理员")
	}
	if !param.IsQueue && printerInfo.Printer.Uuid == 0 {
		return printerLog, errors.New("未配置打印机, 请联系管理员")
	}
	if param.DeviceId != "" && printerInfo.IsCashierPrinter && viper.GetBool("IS_CLOUD_DEPLOY") {
		cashierBindKey = param.DeviceId
	}
	if printerInfo.IsCashierOpen {
		var content string
		if printerInfo.Printer.Uuid > 0 {
			printerType = printerInfo.Printer.PrinterType.Key
			content = s.getPrintContent(param.RechargeOrder, printerType)
		}
		var printerLogType int
		if printerInfo.IsCashierPrinter {
			printerLogType = constant.PrinterLogTypeCloud
		}
		var firstExecution int
		if !param.IsQueue {
			firstExecution = 1
		}
		// 添加打印日志，依赖打印日志服务
		printerLog, err = s.printerLogSrv.AddLog(ctx, printer, model.PrinterLog{
			PrinterUuid:     printer.Uuid,
			CashierDeviceId: cashierBindKey,
			RelatedType:     constant.PrinterLogRelatedRechargeOrder, // 关联
			RelatedUuid:     param.RechargeOrder.Uuid,
			DataType:        constant.PrinterLogDataTypeRecharge,
			Data:            content,
			Type:            printerLogType,
			FirstExecution:  firstExecution,
		}, param.DeviceId)

		if err != nil {
			return printerLog, errors.New("打印失败，未连接打印机")
		}
		return printerLog, nil
	}
	return printerLog, nil
}

// getPrintContent 获取打印内容
func (s *rechargePrinterSrv) getPrintContent(rechargeOrder model.MemberRechargeOrder, printerType string) string {
	return ""
}
