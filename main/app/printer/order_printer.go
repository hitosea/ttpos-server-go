package printer

import (
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
 * 结账单打印
 */
func (p *PrinterRepoImpl) PrintingStatementOrder(
	printType int,
	saleBill *model.SaleBill,
	saleOrderUuid uint64,
	FirstExecution int,
) (*resp.PrinterData, error) {
	// todo 设备id没对

	// 获取打印设置
	settingPrinterInfo, err := p.setting.GetPrinterInfo(p.ctx, p.printerSetting, p.ctx.GetDeviceSn())
	if err != nil {
		return nil, errors.WithMessage(err, "获取打印设置失败")
	}

	// 未开启打印
	if !settingPrinterInfo.IsCashierOpen {
		return nil, errors.WithMessage(err, "未开启打印, 请联系管理员")
	}

	// 未配置打印机
	if settingPrinterInfo.PrinterType == "" {
		return nil, errors.WithMessage(err, "未配置打印机, 请联系管理员")
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(err, "销售订单不存在")
	}

	// 打印方式
	printMethod := constant.PrinterLogPrintMethodText
	if p.printerSetting.PrintMethod == "2" {
		printMethod = constant.PrinterLogPrintMethodImage
	}

	// 打印日志服务
	printerLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 获取打印内容
	printContent := p.getPrintingStatementOrderContent(printType, saleBill, saleOrder)

	// 添加打印日志，依赖打印日志服务
	printerLogData, err := printerLogSrv.AddLog(p.ctx, resp.PrinterInfo{
		PrinterType: settingPrinterInfo.PrinterType,
	}, model.PrinterLog{
		PrintMethod:     printMethod,
		RelatedType:     1,
		RelatedUuid:     saleBill.Uuid,
		PrinterUuid:     settingPrinterInfo.PrinterUuid,
		CashierDeviceId: p.ctx.GetDeviceSn(),
		DataType:        constant.PrinterLogDataTypeReturnDish,
		Data:            printContent,
		Type:            1,
		FirstExecution:  FirstExecution,
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
		Data:          printerLogData.Data,
		PrintMethod:   printMethod,
		Uuid:          printerLogData.Uuid,
		Copies:        settingPrinterInfo.Copies,
		PrinterType:   settingPrinterInfo.PrinterType,
		PrinterConfig: settingPrinterInfo.PrinterConfig,
	}, nil
}

// 构建订单打印的内容
func (p *PrinterRepoImpl) getPrintingStatementOrderContent(
	printType int,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	// 获取打印模板
	tmp := p.GetPrinterTemplate(uint64(printType))

	// 图片打印
	if p.printerSetting.PrintMethod == "2" {
		return template.NewStatementOrderImgTemplate(
			p.ctx,
			p.setting,
			&p.storeSetting,
			&p.printerSetting,
			&p.currencySetting,
		).ImgPrint(printType, tmp, saleBill, saleOrder)
	}

	// 获取打印机类型
	// var printerType string
	// if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
	// 	printerType = printerItem.Printer.PrinterType.Key
	// }

	// // CODESOFT 打印机
	// if printerItem.Printer != nil && slices.Contains([]string{
	// 	constant.PrinterTypeCodesoftLan,
	// 	constant.PrinterTypeCodesoftWifi,
	// }, printerType) {
	// 	t := template.NewDishesCodesoftTemplate(p.ctx, p.setting, &p.storeSetting, &p.printerSetting, &p.currencySetting)
	// 	return t.CompleteOrder(tmp, printerItem, saleBill, products)
	// }

	// // 商米和芯烨打印机
	// if printerItem.Printer != nil {
	// 	t := template.NewDishesXprinterTemplate(p.ctx, p.setting, &p.storeSetting, &p.printerSetting, &p.currencySetting)
	// 	return t.CompleteOrder(tmp, printerItem, saleBill, products)
	// }

	return ""
}
