package printer

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/service"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 打印发票
 */
func (p *PrinterRepoImpl) PrintingInvoice(
	saleBill *model.SaleBill,
	saleOrderUuid uint64,
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

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 打印方式
	printMethod := p.SetPrinterMethod(settingPrinterInfo.PrintMethod)

	// 设置打印机宽度
	p.SetPrinterWidth(settingPrinterInfo.PrinterWidth)

	// 打印日志服务
	printerLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 获取打印内容
	printContent := p.getPrintingInvoiceContent(
		settingPrinterInfo,
		settingPrinterInfo.PrinterType,
		saleBill,
		saleOrder,
		settingPrinterInfo.IsCashierPrinter,
	)
	if printContent == "" {
		return nil, errors.New("获取打印内容失败")
	}

	// 添加打印日志，依赖打印日志服务
	printerLogData, err := printerLogSrv.AddLog(p.ctx, resp.PrinterInfo{
		PrinterType: settingPrinterInfo.PrinterType,
	}, model.PrinterLog{
		PrintMethod:     printMethod,
		RelatedType:     1,
		RelatedUuid:     saleOrderUuid,
		PrinterUuid:     settingPrinterInfo.PrinterUuid,
		CashierDeviceId: settingPrinterInfo.PrinterCashierDeviceSn,
		DataType:        constant.PrinterTemplateInvoice,
		Data:            printContent,
		Type:            1,
		FirstExecution:  firstExecution,
		Copies:          settingPrinterInfo.Copies,
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
		Data:             printerLogData.Data,
		PrintMethod:      printMethod,
		Uuid:             printerLogData.Uuid,
		Copies:           settingPrinterInfo.Copies,
		PrinterType:      settingPrinterInfo.PrinterType,
		PrinterConfig:    settingPrinterInfo.PrinterConfig,
		IsCashierPrinter: settingPrinterInfo.IsCashierPrinter,
		IsUsbPrinter:     settingPrinterInfo.IsUsbPrinter,
		PrintingTime:     printerLogData.PrintingTime,
	}, nil
}

// 构建发票打印的内容
func (p *PrinterRepoImpl) getPrintingInvoiceContent(
	settingPrinterInfo respSetting.PrinterInfo,
	printerType string,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
	isCashierPrinter bool,
) string {
	// 获取打印模板
	tmpInfo := p.GetPrinterTemplateInfo(uint64(constant.PrinterTemplateInvoice))

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
		constant.PrinterTypeSunmiLan,
		constant.PrinterTypeSunmiCloud,
		constant.PrinterTypeCashierSunmi,
	}, printerType) {
		base.IsSunMi = true
	}

	// 图片打印
	if p.IsImagePrinterMethod() {
		if !p.Is58mmPrinter() {
			return template.NewInvoiceImgTemplate(base).GetPrintContent(
				settingPrinterInfo,
				tmpInfo,
				saleBill,
				saleOrder,
			)
		} else {
			return template.NewInvoiceImg58mmTemplate(base).GetPrintContent58mm(
				settingPrinterInfo,
				tmpInfo,
				saleBill,
				saleOrder,
			)
		}

	}

	/* *
	* Compax 收银打印机 80mm 自带
	 */
	if printerType == constant.PrinterTypeCashierCompax {
		return template.NewInvoiceCompaxTemplate(base).GetPrintContent(
			settingPrinterInfo,
			tmpInfo,
			saleBill,
			saleOrder,
			isCashierPrinter,
		)
	}

	/* *
	 * 芯烨打印机
	 */
	if slices.Contains([]string{constant.PrinterTypeXPrinterLan, constant.PrinterTypeXPrinterWifi}, printerType) {
		return template.NewInvoiceXprinterTemplate(base).GetPrintContent(
			settingPrinterInfo,
			printerType,
			tmpInfo,
			saleBill,
			saleOrder,
			isCashierPrinter,
		)
	}

	/* *
	* 商米打印机
	 */
	if base.IsSunMi {
		return template.NewInvoiceSunmiTemplate(base).GetPrintContent(
			settingPrinterInfo,
			printerType,
			tmpInfo,
			saleBill,
			saleOrder,
			isCashierPrinter,
		)
	}

	/* *
	* CODESOFT 打印机
	 */
	if slices.Contains([]string{constant.PrinterTypeCodesoftLan, constant.PrinterTypeCodesoftWifi}, printerType) {
		return template.NewInvoiceCodesoftTemplate(base).GetPrintContent(
			settingPrinterInfo,
			printerType,
			tmpInfo,
			saleBill,
			saleOrder,
			isCashierPrinter,
		)
	}

	return ""
}
