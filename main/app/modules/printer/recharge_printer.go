package printer

import (
	"slices"
	"ttpos-server-go/app/dto/resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/service"
	"ttpos-server-go/app/modules/printer/template"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 会员充值订单打印
 */
func (p *PrinterRepoImpl) PrintingRechargeOrder(
	order model.MemberRechargeOrder,
	firstExecution int,
) (*resp.PrinterData, error) {
	deviceSn := p.ctx.GetDeviceSn()

	// 获取打印设置
	settingPrinterInfo, err := p.setting.GetPrinterInfo(p.ctx, p.printerSetting, deviceSn)
	if err != nil {
		return nil, errors.WithMessage(err, "获取打印设置失败")
	}

	// 未开启打印
	if !settingPrinterInfo.IsCashierOpen {
		return nil, errors.New("未开启打印, 请联系管理员")
	}

	// 未配置打印机
	if settingPrinterInfo.PrinterType == "" {
		return nil, errors.New("未配置打印机, 请联系管理员")
	}

	// 打印方式
	printMethod := p.SetPrinterMethod(settingPrinterInfo.PrintMethod)

	// 打印日志服务
	printerLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 获取打印内容
	printContent := p.getPrintingRechargeOrderContent(settingPrinterInfo, order)
	if printContent == "" {
		return nil, errors.New("获取打印内容失败")
	}

	// 添加打印日志，依赖打印日志服务
	printerLogData, err := printerLogSrv.AddLog(p.ctx, resp.PrinterInfo{
		PrinterType: settingPrinterInfo.PrinterType,
	}, model.PrinterLog{
		PrintMethod:     printMethod,
		RelatedType:     2,
		RelatedUuid:     order.Uuid,
		PrinterUuid:     settingPrinterInfo.PrinterUuid,
		CashierDeviceId: settingPrinterInfo.PrinterCashierDeviceSn,
		DataType:        printerConst.PrinterTemplateRecharge,
		Data:            printContent,
		Type:            1,
		FirstExecution:  firstExecution,
		Copies:          settingPrinterInfo.Copies,
		PrintSpeed:      settingPrinterInfo.PrintSpeed,
	}, "")
	if err != nil {
		logger.Logger.Error("添加打印日志失败", zap.Error(err))
		return nil, errors.WithMessage(err, "添加打印日志失败")
	}

	// 代表由服务器进行发送打印
	if printerLogData.Type == 0 {
		return &resp.PrinterData{}, nil
	}

	// 打印
	return &resp.PrinterData{
		Data:              printerLogData.Data,
		PrintMethod:       printMethod,
		Uuid:              printerLogData.Uuid,
		Copies:            settingPrinterInfo.Copies,
		PrinterType:       settingPrinterInfo.PrinterType,
		PrinterConfig:     settingPrinterInfo.PrinterConfig,
		IsCashierPrinter:  settingPrinterInfo.IsCashierPrinter,
		IsUsbPrinter:      settingPrinterInfo.IsUsbPrinter,
		PrintingTime:      printerLogData.PrintingTime,
		EnableStatusCheck: settingPrinterInfo.EnableStatusCheck,
		TradeNo:           printerLogData.GetTradeNo(p.ctx.GetCompanyUuid()),
		PrintChunkSize:    printerLogData.GetPrintChunkSize(),
	}, nil
}

// 构建订单打印的内容
func (p *PrinterRepoImpl) getPrintingRechargeOrderContent(
	settingPrinterInfo settingResp.PrinterInfo,
	order model.MemberRechargeOrder,
) string {
	printerType := settingPrinterInfo.PrinterType

	// 创建打印机实例
	base := template.NewPrinterTemplate(
		p.ctx,
		p.setting,
		&p.storeSetting,
		&p.printerSetting,
		&p.currencySetting,
		false,
		p.Lang,
	)

	// 商米打印机
	if slices.Contains([]string{
		printerConst.PrinterTypeSunmiLan,
		printerConst.PrinterTypeSunmiCloud,
		printerConst.PrinterTypeCashierSunmi,
	}, printerType) {
		base.IsSunMi = true
	}

	// 图片打印
	if p.IsImagePrinterMethod() {
		return template.NewRechargeImgTemplate(base).GetPrintContent(settingPrinterInfo, order)
	}

	/* *
	* Compax 收银打印机 80mm 自带
	 */
	if printerType == printerConst.PrinterTypeCashierCompax {
		return template.NewRechargeCompaxTemplate(base).GetPrintContent(settingPrinterInfo, order)
	}

	/* *
	 * 芯烨打印机
	 */
	if slices.Contains([]string{
		printerConst.PrinterTypeXPrinterLan,
		printerConst.PrinterTypeXPrinterWifi,
		printerConst.PrinterTypeCashierImmin,
	}, printerType) {
		return template.NewRechargeXPrinterTemplate(base).GetPrintContent(settingPrinterInfo, order)
	}

	/* *
	* 商米打印机
	 */
	if base.IsSunMi {
		return template.NewRechargeSunmiTemplate(base).GetPrintContent(
			settingPrinterInfo,
			order,
		)
	}

	/* *
	* CODESOFT 打印机
	 */
	if slices.Contains([]string{printerConst.PrinterTypeCodesoftLan, printerConst.PrinterTypeCodesoftWifi, printerConst.PrinterTypeGpCloud}, printerType) {
		return template.NewRechargeCodesoftTemplate(base).GetPrintContent(settingPrinterInfo, order)
	}

	return ""
}
