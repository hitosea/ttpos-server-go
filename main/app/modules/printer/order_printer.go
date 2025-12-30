package printer

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/service"
	"ttpos-server-go/app/modules/printer/template"
	printer_request "ttpos-server-go/app/modules/printer/tyeps/request"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 结账单打印
 */
func (p *PrinterRepoImpl) PrintingStatementOrder(req *printer_request.PrintingStatementOrderReq) (*resp.PrinterData, error) {

	// 设备sn
	deviceSn := p.ctx.GetDeviceSn()
	// 如果不是收银端和助手端，从订单获取 获取设备品牌
	if p.ctx.GetSource() != constant.SourceCashier && p.ctx.GetSource() != constant.SourceAssistant {
		deviceRepo := repository.NewDeviceRepo(p.ctx.GetDB())
		deviceSn = deviceRepo.GetDeviceSn(deviceRepo.WhereUuid(req.SaleBill.DeviceUuid))
		if deviceSn == "" {
			return nil, errors.New("获取设备失败")
		}
	}

	// 获取打印设置
	settingPrinterInfo, err := p.setting.GetPrinterInfo(p.ctx, p.printerSetting, deviceSn)
	if err != nil {
		return nil, errors.WithMessage(err, "获取打印设置失败")
	}

	// 应用打印联数优先级规则：收银打印设置-结账单打印联数 > 打印设置-打印机打印联数
	// 仅对结账单（printType == 2）应用此规则
	if req.PrintType == 2 && p.printerSetting.EnableCustomCopies == "1" {
		copies := *p.printerSetting.CheckoutSlipCopies
		if copies > 0 {
			settingPrinterInfo.Copies = uint(copies)
		} else {
			// 如果设置为0，表示不打印
			settingPrinterInfo.Copies = 0
		}
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
	saleOrder := req.SaleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 获取打印模板
	tmp, tmpUuid, tmpData := p.GetPrinterTemplate(uint64(req.PrintType))

	// 打印方式
	printMethod := p.SetPrinterMethod(settingPrinterInfo.PrintMethod)
	if tmpUuid > 0 {
		printMethod = p.SetPrinterMethod(2)
	}

	// 设置打印机宽度
	p.SetPrinterWidth(settingPrinterInfo.PrinterWidth)

	// 打印日志服务
	printerLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 获取打印内容
	printContent := p.getPrintingStatementOrderContent(
		settingPrinterInfo,
		req.PrintType,
		req.SaleBill,
		saleOrder,
		req.PayMethodUuid,
		tmp,
		tmpUuid,
		tmpData,
		req.PayQrcode,
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
		RelatedUuid:     req.SaleOrderUuid,
		PrinterUuid:     settingPrinterInfo.PrinterUuid,
		CashierDeviceId: settingPrinterInfo.PrinterCashierDeviceSn,
		DataType:        req.PrintType,
		Data:            printContent,
		Type:            1,
		FirstExecution:  req.FirstExecution,
		Copies:          settingPrinterInfo.Copies,
		PrintSpeed:      settingPrinterInfo.PrintSpeed,
	}, "")
	if err != nil {
		logger.Logger.Error("添加打印日志失败", zap.Error(err))
		return nil, errors.WithMessage(err, "添加打印日志失败")
	}

	// 如果打印类型为结账单，且启用了自定义打印联数，且打印份数为0，则不打印
	if req.PrintType == 2 && p.printerSetting.EnableCustomCopies == "1" && printerLogData.Copies == 0 {
		return &resp.PrinterData{}, nil
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
func (p *PrinterRepoImpl) getPrintingStatementOrderContent(
	settingPrinterInfo settingResp.PrinterInfo, // 打印机设置
	printType int, // 打印类型 1-预结账单 2-结账单
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
	payMethodUuid uint64,
	tmp int,
	tmpUuid uint64,
	tmpData string,
	payQrcode string,
) string {

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
				payQrcode,
			)
		} else if !p.Is58mmPrinter() {
			return template.NewStatementOrderImgTemplate(base).GetPrintContent(
				settingPrinterInfo,
				printType,
				tmp,
				saleBill,
				saleOrder,
				payMethodUuid,
				payQrcode,
			)
		} else {
			return template.NewStatementOrderImg58mmTemplate(base).GetPrintContent58mm(
				settingPrinterInfo,
				printType,
				tmp,
				saleBill,
				saleOrder,
				payMethodUuid,
				payQrcode,
			)
		}
	}

	/* *
	* Compax 收银打印机 80mm 自带
	 */
	if settingPrinterInfo.PrinterType == printerConst.PrinterTypeCashierCompax {
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
	if slices.Contains([]string{
		printerConst.PrinterTypeXPrinterLan,
		printerConst.PrinterTypeXPrinterWifi,
		printerConst.PrinterTypeCashierImmin,
	}, settingPrinterInfo.PrinterType) {
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
	if slices.Contains([]string{printerConst.PrinterTypeCodesoftLan, printerConst.PrinterTypeCodesoftWifi, printerConst.PrinterTypeGpCloud}, settingPrinterInfo.PrinterType) {
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
