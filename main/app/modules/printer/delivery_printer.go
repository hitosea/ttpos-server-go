package printer

import (
	"slices"
	"ttpos-server-go/app/constant"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/dto/resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer/service"
	"ttpos-server-go/app/modules/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 外送单打印
 */
func (p *PrinterRepoImpl) PrintingTakeoutOrder(
	memberSaleOrder *model.MemberSaleOrder,
	saleBill *model.SaleBill,
	saleOrderUuid uint64,
) (*resp.PrinterData, error) {

	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())

	// 设备sn
	deviceSn := p.ctx.GetDeviceSn()

	// 如果不是收银端端，从主设备获取
	if deviceSn == "" || p.ctx.GetSource() != constant.SourceCashier {
		deviceRepo := repository.NewDeviceRepo(db)
		deviceSn = deviceRepo.GetDeviceSn(deviceRepo.WhereMain())
		if deviceSn == "" {
			return nil, errors.New("找不到收银机设备")
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

	// 打印日志服务
	printerLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 获取打印内容
	printContent := p.getPrintingContent(
		settingPrinterInfo,
		printerConst.PrinterTemplateTakeoutOrder,
		memberSaleOrder,
		saleBill,
		saleOrder,
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
		DataType:        printerConst.PrinterTemplateTakeoutOrder,
		Data:            printContent,
		Type:            1,
		FirstExecution:  0,
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
func (p *PrinterRepoImpl) getPrintingContent(
	settingPrinterInfo settingResp.PrinterInfo, // 打印机设置
	printType int, // 打印类型 11-外送单
	memberSaleOrder *model.MemberSaleOrder,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	// 获取打印模板
	tmp, _, _ := p.GetPrinterTemplate(uint64(printType))

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
		return template.NewTakeoutOrderImgTemplate(base).GetPrintContent(
			settingPrinterInfo,
			tmp,
			memberSaleOrder,
			saleBill,
			saleOrder,
		)
	}

	/* *
	* Compax 收银打印机 80mm 自带
	 */
	if settingPrinterInfo.PrinterType == printerConst.PrinterTypeCashierCompax {
		return template.NewTakeoutOrderCompaxTemplate(base).GetPrintContent(
			settingPrinterInfo,
			tmp,
			memberSaleOrder,
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
		return template.NewTakeoutOrderXprinterTemplate(base).GetPrintContent(
			settingPrinterInfo,
			tmp,
			memberSaleOrder,
			saleBill,
			saleOrder,
		)
	}

	/* *
	* 商米打印机
	 */
	if base.IsSunMi {
		return template.NewTakeoutOrderSunmiTemplate(base).GetPrintContent(
			settingPrinterInfo,
			tmp,
			memberSaleOrder,
			saleBill,
			saleOrder,
		)
	}

	/* *
	* CODESOFT 打印机
	 */
	if slices.Contains([]string{printerConst.PrinterTypeCodesoftLan, printerConst.PrinterTypeCodesoftWifi, printerConst.PrinterTypeGpCloud}, settingPrinterInfo.PrinterType) {
		return template.NewTakeoutOrderCodesoftTemplate(base).GetPrintContent(
			settingPrinterInfo,
			tmp,
			memberSaleOrder,
			saleBill,
			saleOrder,
		)
	}

	return ""
}
