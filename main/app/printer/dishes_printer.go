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
	"ttpos-server-go/pkg/utils"

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

	// 获取订单信息
	db := p.dbm.GetDB(p.ctx.GetCompanyUuid())
	billInfo, err := repository.NewOrderRepo(db).GetSaleBillInfo(saleBillUuid, constant.No)
	if err != nil {
		return false
	}

	// 设备
	deviceRepo := repository.NewDeviceRepo(db)

	// 获取主设备
	mainDeviceSn := deviceRepo.GetDeviceSn(deviceRepo.WhereMain())

	// 打印日志服务
	pinterLogSrv := service.NewPrinterLogSrv(p.dbm, setting.NewSrv(p.dbm, p.cache))

	// 出菜单打印
	if printType == constant.PrinterProductTypeOutMenu {
		// 获取设备
		device, errDevice := deviceRepo.GetDevice(deviceRepo.WhereSource(p.ctx.GetSource()), deviceRepo.WhereSn(p.ctx.GetDeviceSn()))
		if errDevice != nil || device.RelatedPrinterUuid == 0 {
			return false
		}
		// 获取关联打印机
		printerRepo := repository.NewPrinterRepo(db)
		printer, err := printerRepo.GetPrinter(printerRepo.WhereUuid(device.RelatedPrinterUuid), printerRepo.WithPrinterType())
		if err != nil {
			return false
		}
		// 打印机类型
		printerType := constant.PrinterTypeXPrinterLan
		if printer.PrinterType != nil {
			printerType = printer.PrinterType.Key
		}
		// 打印方式
		printMethod := p.SetPrinterMethod(printer.PrintMethod, true)
		data := p.getPrintProductOutMenuContent(printer, billInfo, products)
		if data != "" {
			_, err = pinterLogSrv.AddLog(p.ctx, resp.PrinterInfo{
				PrinterType:   printerType,
				PrinterConfig: printer.ConfigJson,
			}, model.PrinterLog{
				PrintMethod: printMethod,
				RelatedType: 0,
				RelatedUuid: saleBillUuid,
				PrinterUuid: printer.Uuid,
				CashierDeviceId: func() string {
					cashierDeviceId := mainDeviceSn
					if printer.SourceDeviceSn != "" {
						cashierDeviceId = deviceRepo.GetDeviceSn(
							deviceRepo.WhereSource(constant.SourceCashier),
							deviceRepo.WhereSn(printer.SourceDeviceSn),
						)
						if cashierDeviceId == "" {
							cashierDeviceId = mainDeviceSn
						}
					}
					return cashierDeviceId
				}(),
				DataType:           constant.PrinterTemplateOutMenu,
				Data:               data,
				Type:               1,
				FirstExecution:     0,
				ProductPrinterUuid: 0,
				Copies:             printer.Copies,
			}, "")
			if err != nil {
				logger.Logger.Error("添加打印日志失败", zap.Error(err))
			}
		}

	} else {
		// 获取商品打印机列表
		productPrinters, err := p.getProductPrinterList(printType)
		if err != nil {
			logger.Logger.Error("获取商品打印机列表失败", zap.Error(err))
			return false
		}
		if len(productPrinters) == 0 {
			return false
		}

		// 当前订单对应的区域id
		regionUuid := uint64(0)
		if billInfo.Desk != nil {
			regionUuid = billInfo.Desk.RegionUuid
		}

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
				// 套餐商品列表 - 使用通用过滤方法
				subProducts := utils.FilterByContains(product.SubProducts, productIds, func(p printer_model.OrderProduct) uint64 {
					return p.ProductId
				})
				if !slices.Contains(productIds, product.ProductId) && len(subProducts) == 0 {
					continue
				}
				// 套餐商品
				product.SubProducts = subProducts
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

				// 获取打印机设备sn
				cashierDeviceId := mainDeviceSn
				if printerItem.Printer != nil && printerItem.Printer.SourceDeviceSn != "" {
					cashierDeviceId = deviceRepo.GetDeviceSn(
						deviceRepo.WhereSource(constant.SourceCashier),
						deviceRepo.WhereSn(printerItem.Printer.SourceDeviceSn),
					)
					if cashierDeviceId == "" {
						cashierDeviceId = mainDeviceSn
					}
				}

				// 退菜单打印
				if printType == constant.PrinterProductTypeBackFood {
					data := p.getPrintReturnProductContent(printerItem, billInfo, newProducts)
					if data != "" {
						_, err = pinterLogSrv.AddLog(p.ctx, resp.PrinterInfo{
							PrinterType:   printerType,
							PrinterConfig: printerItem.Printer.ConfigJson,
						}, model.PrinterLog{
							PrintMethod:        printMethod,
							RelatedType:        0,
							RelatedUuid:        saleBillUuid,
							PrinterUuid:        printerItem.PrinterUuid,
							CashierDeviceId:    cashierDeviceId,
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
						// 定义产品导出函数
						exportation := func(product printer_model.OrderProduct) {
							if data := p.getPrintProductOneContent(productPrinter, printerItem, billInfo, product); data != "" {
								_, err = pinterLogSrv.AddLog(p.ctx, resp.PrinterInfo{
									PrinterType:   printerType,
									PrinterConfig: printerItem.Printer.ConfigJson,
								}, model.PrinterLog{
									PrintMethod:        printMethod,
									RelatedType:        0,
									RelatedUuid:        saleBillUuid,
									PrinterUuid:        printerItem.PrinterUuid,
									CashierDeviceId:    cashierDeviceId,
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
						// 套餐商品列表
						subProducts := utils.FilterByContains(product.SubProducts, productIds, func(p printer_model.OrderProduct) uint64 {
							return p.ProductId
						})
						if len(subProducts) > 0 {
							for _, subProduct := range subProducts {
								exportation(subProduct)
							}
						} else {
							exportation(product)
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
						PrintMethod:        printMethod,
						RelatedType:        0,
						RelatedUuid:        saleBillUuid,
						PrinterUuid:        printerItem.PrinterUuid,
						CashierDeviceId:    cashierDeviceId,
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
	tmpInfo := p.GetPrinterTemplateInfo(constant.PrinterTemplateEntireOrder)

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
		return t.CompleteOrder(tmpInfo, printerItem, saleBill, products)
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
		return t.CompleteOrder(tmpInfo, printerItem, saleBill, products)
	}

	// 商米和芯烨打印机
	if printerItem.Printer != nil {
		t := template.NewDishesXprinterTemplate(base)
		return t.CompleteOrder(tmpInfo, printerItem, saleBill, products)
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
	tmpInfo := p.GetPrinterTemplateInfo(constant.PrinterTemplateOneDishOneMenu)

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
		return t.OneDishOneOrder(tmpInfo, productPrinter, printerItem, saleBill, []printer_model.OrderProduct{product})
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
		constant.PrinterTypeGpCloud,
	}, printerType) {
		t := template.NewDishesCodesoftTemplate(base)
		return t.OneDishOneOrder(tmpInfo, productPrinter, printerItem, saleBill, []printer_model.OrderProduct{product})
	}

	// 商米和芯烨打印机
	if printerItem.Printer != nil {
		t := template.NewDishesXprinterTemplate(base)
		return t.OneDishOneOrder(tmpInfo, productPrinter, printerItem, saleBill, []printer_model.OrderProduct{product})
	}

	return ""
}

// 构建退菜单打印的内容
func (p *PrinterRepoImpl) getPrintReturnProductContent(
	printerItem *model.ProductPrinterItem,
	saleBill model.SaleBill,
	products printer_model.Products,
) string {
	tmp, _, _ := p.GetPrinterTemplate(constant.PrinterTemplateReturnDish)

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
	printer model.Printer,
	saleBill model.SaleBill,
	products printer_model.Products,
) string {
	tmp, _, _ := p.GetPrinterTemplate(constant.PrinterTemplateOutMenu)

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
		return template.NewDishesImgTemplate(base).OutMenuTemplate(tmp, saleBill, products, p.GetFinishedTime())
	}

	// 商米和芯烨打印机
	return template.NewDishesXprinterTemplate(base).OutMenuTemplate(tmp, printer, saleBill, products, p.GetFinishedTime())
}
