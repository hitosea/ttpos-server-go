package printer

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/service"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 业务数据打印
 */
func (p *PrinterRepoImpl) PrintingBusinessData(
	businessData *template.PrintingBusinessData,
	startTime int64,
	endTime int64,
	deviceSnId ...string,
) (*resp.PrinterData, error) {
	var deviceSn string
	// 设备sn
	if len(deviceSnId) > 0 {
		deviceSn = deviceSnId[0]
	} else {
		deviceSn = p.ctx.GetDeviceSn()
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

	// 打印方式
	printMethod := p.SetPrinterMethod(settingPrinterInfo.PrintMethod)

	// 设置打印机宽度
	p.SetPrinterWidth(settingPrinterInfo.PrinterWidth)

	// 打印日志服务
	printerLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 获取打印内容
	printContent := p.getPrintingBusinessDataContent(settingPrinterInfo.PrinterType, businessData, startTime, endTime)
	if printContent == "" {
		return nil, errors.New("获取打印内容失败")
	}

	// 添加打印日志，依赖打印日志服务
	printerLogData, err := printerLogSrv.AddLog(p.ctx, resp.PrinterInfo{
		PrinterType: settingPrinterInfo.PrinterType,
	}, model.PrinterLog{
		PrintMethod:     printMethod,
		RelatedType:     0,
		RelatedUuid:     0,
		PrinterUuid:     settingPrinterInfo.PrinterUuid,
		CashierDeviceId: settingPrinterInfo.PrinterCashierDeviceSn,
		DataType:        constant.PrinterTemplateBusiness,
		Data:            printContent,
		Type:            1,
		FirstExecution:  1,
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
func (p *PrinterRepoImpl) getPrintingBusinessDataContent(
	printerType string, // 打印机类型
	businessData *template.PrintingBusinessData,
	startTime int64,
	endTime int64,
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
		constant.PrinterTypeSunmiLan,
		constant.PrinterTypeSunmiCloud,
		constant.PrinterTypeCashierSunmi,
	}, printerType) {
		base.IsSunMi = true
	}

	// 图片打印
	if p.IsImagePrinterMethod() {
		if !p.Is58mmPrinter() {
			return template.NewBusinessDataImgTemplate(base).GetPrintContent(businessData, startTime, endTime)
		} else {
			// 调用58mm模板
			return template.NewBusinessDataImgTemplate58mm(base).GetPrintContent58mm(businessData, startTime, endTime)
		}
	}

	/* *
	* 商米打印机
	 */
	if base.IsSunMi {
		return template.NewBusinessDataSunmiTemplate(base).GetPrintContent(
			printerType,
			businessData,
			startTime,
			endTime,
		)
	}

	/* *
	 * 芯烨打印机
	 */
	if slices.Contains([]string{
		constant.PrinterTypeXPrinterLan,
		constant.PrinterTypeXPrinterWifi,
		constant.PrinterTypeCodesoftLan,
		constant.PrinterTypeCodesoftWifi,
		constant.PrinterTypeGpCloud,
		constant.BrandA11510P,
	}, printerType) {
		return template.NewBusinessDataXprinterTemplate(base).GetPrintContent(
			printerType,
			businessData,
			startTime,
			endTime,
		)
	}

	return ""
}
