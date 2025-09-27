package printer

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/service"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 结账单打印
 */
func (p *PrinterRepoImpl) PrintingStatementOrder(
	printType int, // 打印类型 1-预结账单 2-结账单 11-外送单
	saleBill *model.SaleBill,
	saleOrderUuid uint64,
	firstExecution int,
	payMethodUuid uint64,
) (*resp.PrinterData, error) {

	// 设备sn
	deviceSn := p.ctx.GetDeviceSn()
	// 如果不是收银端和助手端，从订单获取 获取设备品牌
	if p.ctx.GetSource() != constant.SourceCashier && p.ctx.GetSource() != constant.SourceAssistant {
		deviceRepo := repository.NewDeviceRepo(p.ctx.GetDB())
		deviceSn = deviceRepo.GetDeviceSn(deviceRepo.WhereUuid(saleBill.DeviceUuid))
		if deviceSn == "" {
			return nil, errors.New("获取设备失败")
		}
	}

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
	printContent := p.getPrintingStatementOrderContent(settingPrinterInfo, printType, saleBill, saleOrder, payMethodUuid)
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
		DataType:        printType,
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
	}, nil
}

// 构建订单打印的内容
func (p *PrinterRepoImpl) getPrintingStatementOrderContent(
	settingPrinterInfo settingResp.PrinterInfo, // 打印机设置
	printType int, // 打印类型 1-预结账单 2-结账单
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
	payMethodUuid uint64,
) string {
	// 获取打印模板
	tmp, tmpUuid, tmpData := p.GetPrinterTemplate(uint64(printType))

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
	}, settingPrinterInfo.PrinterType) {
		base.IsSunMi = true
	}

	// 图片打印
	if p.IsImagePrinterMethod() {
		if tmpUuid > 0 {
			return template.NewStatementOrderImgTemplateCustom(base).GetPrintContent(
				settingPrinterInfo,
				tmpData,
				saleBill,
				saleOrder,
				payMethodUuid,
				p.Is58mmPrinter(),
			)
		} else if !p.Is58mmPrinter() {
			return template.NewStatementOrderImgTemplate(base).GetPrintContent(
				settingPrinterInfo,
				printType,
				tmp,
				saleBill,
				saleOrder,
				payMethodUuid,
			)
		} else {
			return template.NewStatementOrderImg58mmTemplate(base).GetPrintContent58mm(
				settingPrinterInfo,
				printType,
				tmp,
				saleBill,
				saleOrder,
				payMethodUuid,
			)
		}
	}

	/* *
	* Compax 收银打印机 80mm 自带
	 */
	if settingPrinterInfo.PrinterType == constant.PrinterTypeCashierCompax {
		return template.NewStatementOrderCompaxTemplate(base).GetPrintContent(
			settingPrinterInfo,
			printType,
			tmp,
			saleBill,
			saleOrder,
		)
	}

	/* *
	 * 芯烨打印机
	 */
	if slices.Contains([]string{constant.PrinterTypeXPrinterLan, constant.PrinterTypeXPrinterWifi}, settingPrinterInfo.PrinterType) {
		return template.NewStatementOrderXprinterTemplate(base).GetPrintContent(
			settingPrinterInfo,
			printType,
			tmp,
			saleBill,
			saleOrder,
		)
	}

	/* *
	* 商米打印机
	 */
	if base.IsSunMi {
		return template.NewStatementOrderSunmiTemplate(base).GetPrintContent(
			settingPrinterInfo,
			printType,
			tmp,
			saleBill,
			saleOrder,
		)
	}

	/* *
	* CODESOFT 打印机
	 */
	if slices.Contains([]string{constant.PrinterTypeCodesoftLan, constant.PrinterTypeCodesoftWifi, constant.PrinterTypeGpCloud}, settingPrinterInfo.PrinterType) {
		return template.NewStatementOrderCodesoftTemplate(base).GetPrintContent(
			settingPrinterInfo,
			printType,
			tmp,
			saleBill,
			saleOrder,
		)
	}

	return ""
}
