package printer

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/printer/service"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

/**
 * 菜品打印
 * @param int printType 打印类型 -2-为出菜单打印 -1-为退菜打印 0-付款打印 1-送厨打印
 */
func (p *PrinterRepoImpl) PrintingDishes(
	printType int,
	saleBillUuid uint64,
	products printer_model.Products,
) bool {
	// 设置语言
	p.Lang = p.printerSetting.KitchenLanguage

	// 获取商品打印机列表
	productPrinters, err := p.getProductPrinterList(printType)
	if err != nil {
		logger.Logger.Error("获取商品打印机列表失败", zap.Error(err))
		return false
	}
	if len(productPrinters) == 0 {
		return false
	}

	// 获取订单信息
	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())
	billInfo, err := repository.NewOrderRepo(db).GetSaleBillInfo(saleBillUuid, constant.No)
	if err != nil {
		return false
	}

	// 当前订单对应的区域id
	regionUuid := uint64(0)
	if billInfo.Desk != nil {
		regionUuid = billInfo.Desk.RegionUuid
	}

	// 打印日志服务
	pinterLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 循环商品打印机
	for _, productPrinter := range productPrinters {
		// 区域对的上才走
		regionUuids := productPrinter.GetPrinterRegionUuids()
		if len(regionUuids) != 0 && !slices.Contains(regionUuids, regionUuid) {
			continue
		}

		// 选中的商品才往下走
		productIds := productPrinter.GetPrinterProductIds()
		newProducts := make(printer_model.Products, 0)
		for _, product := range products {
			if !slices.Contains(productIds, product.ProductId) {
				continue
			}
			newProducts = append(newProducts, product)
		}

		// 循环下拉选中的打印机一个个打印
		for _, printerItem := range productPrinter.ProductPrinterItems {
			// 判断是否删除
			if printerItem.IsDelete() {
				continue
			}
			// 获取打印机类型
			var printerType string
			if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
				printerType = printerItem.Printer.PrinterType.Key
			}

			// 打印方式
			var printMethod int
			if printerItem.Printer != nil {
				printMethod = p.SetPrinterMethod(printerItem.Printer.PrintMethod, true)
			} else {
				printMethod = p.GetPrinterMethod(true)
			}

			// 出菜单打印
			if printType == constant.PrinterProductTypeOutMenu {
				data := p.getPrintProductOutMenuContent(productPrinter, printerItem, billInfo, newProducts)
				if data != "" {
					_, err = pinterLogSrv.AddLog(p.ctx, resp.PrinterInfo{
						PrinterType:   printerType,
						PrinterConfig: printerItem.Printer.ConfigJson,
					}, model.PrinterLog{
						PrintMethod: printMethod,
						RelatedType: 0,
						RelatedUuid: saleBillUuid,
						PrinterUuid: printerItem.PrinterUuid,
						CashierDeviceId: func() string {
							if printerItem.Printer != nil && printerItem.Printer.IsUsbPrinter() {
								return printerItem.Printer.SourceDeviceSn
							}
							return ""
						}(),
						DataType:           constant.PrinterTemplateOutMenu,
						Data:               data,
						Type:               1,
						FirstExecution:     0,
						ProductPrinterUuid: productPrinter.Uuid,
						Copies:             productPrinter.Copies,
					}, "")
					if err != nil {
						logger.Logger.Error("添加打印日志失败", zap.Error(err))
					}
				}
				continue
			}

			// 退菜单打印
			if printType == constant.PrinterProductTypeBackFood {
				data := p.getPrintReturnProductContent(printerItem, billInfo, newProducts)
				if data != "" {
					_, err = pinterLogSrv.AddLog(p.ctx, resp.PrinterInfo{
						PrinterType:   printerType,
						PrinterConfig: printerItem.Printer.ConfigJson,
					}, model.PrinterLog{
						PrintMethod: printMethod,
						RelatedType: 0,
						RelatedUuid: saleBillUuid,
						PrinterUuid: printerItem.PrinterUuid,
						CashierDeviceId: func() string {
							if printerItem.Printer != nil && printerItem.Printer.IsUsbPrinter() {
								return printerItem.Printer.SourceDeviceSn
							}
							return ""
						}(),
						DataType:           constant.PrinterTemplateReturnDish,
						Data:               data,
						Type:               1,
						FirstExecution:     0,
						ProductPrinterUuid: productPrinter.Uuid,
						Copies:             productPrinter.Copies,
					}, "")
					if err != nil {
						logger.Logger.Error("添加打印日志失败", zap.Error(err))
					}
				}
				continue
			}

			// 一菜一单打印
			if productPrinter.PrintMethod == constant.Yes || productPrinter.PrintMethod == constant.All {
				for _, product := range newProducts {
					if data := p.getPrintProductOneContent(productPrinter, printerItem, billInfo, product); data != "" {
						_, err = pinterLogSrv.AddLog(p.ctx, resp.PrinterInfo{
							PrinterType:   printerType,
							PrinterConfig: printerItem.Printer.ConfigJson,
						}, model.PrinterLog{
							PrintMethod: printMethod,
							RelatedType: 0,
							RelatedUuid: saleBillUuid,
							PrinterUuid: printerItem.PrinterUuid,
							CashierDeviceId: func() string {
								if printerItem.Printer != nil && printerItem.Printer.IsUsbPrinter() {
									return printerItem.Printer.SourceDeviceSn
								}
								return ""
							}(),
							DataType:           constant.PrinterTemplateOneDishOneMenu,
							Data:               data,
							Type:               1,
							FirstExecution:     0,
							ProductPrinterUuid: productPrinter.Uuid,
							Copies:             productPrinter.Copies,
						}, "")
						if err != nil {
							logger.Logger.Error("添加打印日志失败", zap.Error(err))
						}
					}
				}
				if productPrinter.PrintMethod != constant.All {
					continue
				}
			}

			// 整单打印
			if data := p.getPrintProductContent(productPrinter, printerItem, billInfo, newProducts); data != "" {
				// 添加打印日志，依赖打印日志服务
				_, err = pinterLogSrv.AddLog(p.ctx, resp.PrinterInfo{
					PrinterType: printerType,
				}, model.PrinterLog{
					PrintMethod: printMethod,
					RelatedType: 0,
					RelatedUuid: saleBillUuid,
					PrinterUuid: printerItem.PrinterUuid,
					CashierDeviceId: func() string {
						if printerItem.Printer != nil && printerItem.Printer.IsUsbPrinter() {
							return printerItem.Printer.SourceDeviceSn
						}
						return ""
					}(),
					DataType:           constant.PrinterTemplateEntireOrder,
					Data:               data,
					Type:               1,
					FirstExecution:     0,
					ProductPrinterUuid: productPrinter.Uuid,
					Copies:             productPrinter.Copies,
				}, "")
				if err != nil {
					logger.Logger.Error("添加打印日志失败", zap.Error(err))
				}
			}
		}
	}
	// 打印
	return true
}

// 构建订单菜品打印的内容
func (p *PrinterRepoImpl) getPrintProductContent(
	productPrinter model.ProductPrinter,
	printerItem *model.ProductPrinterItem,
	saleBill model.SaleBill,
	products printer_model.Products,
) string {
	tmp := p.GetPrinterTemplate(constant.PrinterTemplateEntireOrder)

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

	// 图片打印
	if p.IsImagePrinterMethod(true) {
		t := template.NewDishesImgTemplate(base)
		return t.CompleteOrder(tmp, printerItem, saleBill, products)
	}

	// 获取打印机类型
	var printerType string
	if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
		printerType = printerItem.Printer.PrinterType.Key
	}

	// CODESOFT 打印机
	if printerItem.Printer != nil && slices.Contains([]string{
		constant.PrinterTypeCodesoftLan,
		constant.PrinterTypeCodesoftWifi,
	}, printerType) {
		t := template.NewDishesCodesoftTemplate(base)
		return t.CompleteOrder(tmp, printerItem, saleBill, products)
	}

	// 商米和芯烨打印机
	if printerItem.Printer != nil {
		t := template.NewDishesXprinterTemplate(base)
		return t.CompleteOrder(tmp, printerItem, saleBill, products)
	}

	return ""
}

// 构建订单菜品（一菜一单）打印的内容
func (p *PrinterRepoImpl) getPrintProductOneContent(
	productPrinter model.ProductPrinter,
	printerItem *model.ProductPrinterItem,
	saleBill model.SaleBill,
	product printer_model.OrderProduct,
) string {
	tmp := p.GetPrinterTemplate(constant.PrinterTemplateOneDishOneMenu)

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

	// 图片打印
	if p.IsImagePrinterMethod(true) {
		t := template.NewDishesImgTemplate(base)
		return t.OneDishOneOrder(tmp, productPrinter, printerItem, saleBill, []printer_model.OrderProduct{product})
	}

	// 获取打印机类型
	var printerType string
	if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
		printerType = printerItem.Printer.PrinterType.Key
	}

	// CODESOFT 打印机
	if printerItem.Printer != nil && slices.Contains([]string{
		constant.PrinterTypeCodesoftLan,
		constant.PrinterTypeCodesoftWifi,
	}, printerType) {
		t := template.NewDishesCodesoftTemplate(base)
		return t.OneDishOneOrder(tmp, productPrinter, printerItem, saleBill, []printer_model.OrderProduct{product})
	}

	// 商米和芯烨打印机
	if printerItem.Printer != nil {
		t := template.NewDishesXprinterTemplate(base)
		return t.OneDishOneOrder(tmp, productPrinter, printerItem, saleBill, []printer_model.OrderProduct{product})
	}

	return ""
}

// 构建退菜单打印的内容
func (p *PrinterRepoImpl) getPrintReturnProductContent(
	printerItem *model.ProductPrinterItem,
	saleBill model.SaleBill,
	products printer_model.Products,
) string {
	tmp := p.GetPrinterTemplate(constant.PrinterTemplateReturnDish)

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

	// 图片打印
	if p.IsImagePrinterMethod(true) {
		t := template.NewDishesImgTemplate(base)
		return t.ReturnMenuTemplate(tmp, printerItem, saleBill, products)
	}

	// 商米和芯烨打印机
	if printerItem.Printer != nil {
		t := template.NewDishesXprinterTemplate(base)
		return t.ReturnMenuTemplate(tmp, printerItem, saleBill, products)
	}
	return ""
}

// 构建订单菜品（出菜单）打印的内容
func (p *PrinterRepoImpl) getPrintProductOutMenuContent(
	productPrinter model.ProductPrinter,
	printerItem *model.ProductPrinterItem,
	saleBill model.SaleBill,
	products printer_model.Products,
) string {
	tmp := p.GetPrinterTemplate(constant.PrinterTemplateEntireOrder)

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

	// 图片打印
	if p.IsImagePrinterMethod(true) {
		t := template.NewDishesImgTemplate(base)
		return t.OutMenuTemplate(tmp, printerItem, saleBill, products)
	}

	// 获取打印机类型
	var printerType string
	if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
		printerType = printerItem.Printer.PrinterType.Key
	}

	// CODESOFT 打印机
	if printerItem.Printer != nil && slices.Contains([]string{
		constant.PrinterTypeCodesoftLan,
		constant.PrinterTypeCodesoftWifi,
	}, printerType) {
		t := template.NewDishesCodesoftTemplate(base)
		return t.CompleteOrder(tmp, printerItem, saleBill, products)
	}

	// 商米和芯烨打印机
	if printerItem.Printer != nil {
		t := template.NewDishesXprinterTemplate(base)
		return t.CompleteOrder(tmp, printerItem, saleBill, products)
	}

	return ""
}
