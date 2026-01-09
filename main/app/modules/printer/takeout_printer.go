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
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 打印外卖平台订单小票（Grab、LINE MAN 等平台订单）
 * 与 PrintingTakeoutOrder 不同，这个方法处理的是第三方平台的外卖订单
 */
func (p *PrinterRepoImpl) PrintingPlatformTakeoutReceipt(
	order *takeoutModel.TakeoutOrder,
	receiptType string, // printerConst.TakeoutReceiptTypeMerchant、TakeoutReceiptTypeCustomer 或 TakeoutReceiptTypeRefund
	firstExecution int,
) (*resp.PrinterData, error) {
	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())

	// 设备sn
	deviceSn := p.ctx.GetDeviceSn()

	// 如果不是收银端，从主设备获取
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

	// 设置打印机宽度
	p.SetPrinterWidth(settingPrinterInfo.PrinterWidth)

	// 确定模板类型（商家联、顾客联或退单联）
	templateType := printerConst.PrinterTemplateTakeoutMerchant
	isMerchantReceipt := true // 默认是商家联
	if receiptType == printerConst.TakeoutReceiptTypeCustomer {
		templateType = printerConst.PrinterTemplateTakeoutCustomer
		isMerchantReceipt = false
	} else if receiptType == printerConst.TakeoutReceiptTypeRefund {
		templateType = printerConst.PrinterTemplateTakeoutRefund
		isMerchantReceipt = true // 退单联与商家联类似，显示完整信息
	}

	// 获取打印模板
	tmpId := uint64(templateType)
	_, tmpUuid, tmpDataStr := p.GetPrinterTemplate(tmpId)
	if tmpUuid == 0 || tmpDataStr == "" {
		printerSrv := service.NewPrinterSrv(p.dbm, p.cache)
		defaultCustomize, err := printerSrv.GetOrCreateDefaultCustomize(p.ctx, tmpId)
		if err != nil {
			return nil, errors.WithMessage(err, "获取或创建默认定制数据失败")
		}
		tmpData, err := printerSrv.UsePrinterCustomize(p.ctx, defaultCustomize.Uuid)
		if err != nil {
			return nil, errors.WithMessage(err, "使用打印机定制失败")
		}
		tmpDataStr = tmpData
	}

	// 获取打印内容
	printContent := p.getPlatformTakeoutPrintContent(
		settingPrinterInfo,
		tmpDataStr,
		order,
		isMerchantReceipt,
	)
	if printContent == "" {
		return nil, errors.New("获取打印内容失败")
	}

	// 打印方式
	printMethod := 2

	// 打印日志服务
	printerLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 添加打印日志
	printerLogData, err := printerLogSrv.AddLog(p.ctx, resp.PrinterInfo{
		PrinterType: settingPrinterInfo.PrinterType,
	}, model.PrinterLog{
		PrintMethod:     printMethod,
		RelatedType:     4, // 4: 外卖平台订单
		RelatedUuid:     order.Uuid,
		PrinterUuid:     settingPrinterInfo.PrinterUuid,
		CashierDeviceId: settingPrinterInfo.PrinterCashierDeviceSn,
		DataType:        templateType,
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

	// 返回打印数据
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

// getPlatformTakeoutPrintContent 获取外卖平台订单打印内容
func (p *PrinterRepoImpl) getPlatformTakeoutPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	tmpData string,
	order *takeoutModel.TakeoutOrder,
	isMerchantReceipt bool,
) string {
	// 创建打印机基础实例
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
	return template.NewPlatformTakeoutImgTemplate(base).GetPrintContent(
		settingPrinterInfo,
		tmpData,
		order,
		p.Is58mmPrinter(),
		isMerchantReceipt,
	)
}
