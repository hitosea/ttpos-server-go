// Package template 提供打印模板相关功能
package template

import (
	"fmt"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/pkg/utils"
)

// dishesXprinterTemplate xprinter菜品打印模板
type dishesXprinterTemplate struct {
	base *printerTemplate
}

// NewDishesXprinterTemplate 创建新的xprinter菜品打印模板
func NewDishesXprinterTemplate(
	base *printerTemplate,
) *dishesXprinterTemplate {
	return &dishesXprinterTemplate{
		base: base,
	}
}

// CompleteOrder 整单模版
func (t *dishesXprinterTemplate) CompleteOrder(
	tmpInfo model.PrinterTemplate,
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {
	// 检查是否有打印内容
	if len(products) == 0 {
		return ""
	}

	temp := tmpInfo.Template
	isShowSku := tmpInfo.IsShowSku

	// 人的翻译
	name := t.base.Translate("人")
	// 自助餐标记开关
	buffetSignOpen := t.base.PrinterSetting.BuffetSignOpen
	// 格式化时间
	updateTime := t.base.FormatUnixTimeDefault(order.UpdateTime)
	// 就餐人数
	mealNumStr := ""
	if order.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", order.MealNum, name)
	}
	// 打印机类型
	printerType := PrinterTypeXPrinterLan
	if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
		printerType = printerItem.Printer.PrinterType.Key
	}

	// 是否是商米打印机
	isSunmi := false
	if printerType == PrinterTypeSunmiLan || printerType == PrinterTypeSunmiCloud || printerType == PrinterTypeCashierSunmi || printerType == PrinterTypeCodesoftWifi {
		isSunmi = true
	}

	// 创建打印机实例
	printer := pkg.NewPrinter(567, "", "")
	if printerType != PrinterTypeXPrinterLan && printerType != PrinterTypeXPrinterWifi {
		printer.LineFeed(1)
	}

	/**
	 * 模版一
	 */
	if temp == 1 {
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetCharacterSize(2, 2, true)
		printer.SetPrintModes(true, true, false)

		// 桌号
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 60
			if t.base.IsMyText(order.SerialNo) {
				spacing = 80
			}
			if printerType == PrinterTypeXPrinterWifi {
				spacing = 120
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
			printer.RestoreDefaultLineSpacing()
			printer.LineFeed()
		} else if order.IsTakeout() {
			printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo + "\n")
		} else {
			printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
		}

		printer.LineFeed()
		printer.SetLineSpacing(50)

		printer.RestoreDefaultLineSpacing()
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{260, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)

		// 订单号
		printer.PrintInColumns(t.base.Translate("订单号"), order.OrderNo)
		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.RestoreDefaultLineSpacing()
		printer.PrintInColumns(t.base.Translate("时间"), updateTime)
		printer.LineFeed()

		// 设置行间距
		if printerType != PrinterTypeXPrinterWifi {
			printer.SetLineSpacing(40)
		}

		// 根据语言显示不同的文本
		if t.base.Lang == "my" {
			printer.AppendText("ကုန်စည်                                  ပမာဏ")
		} else {
			printer.AppendText(t.base.PrintText(t.base.Translate("商品"), "", t.base.Translate("数量"), 47))
		}

		// 添加分隔线
		printer.AppendText("\n------------------------------------------------\n")

		// 设置列
		printer.SetupColumns(
			[]int{500, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)

		// 处理产品
		var processProducts func(products printer_model.Products, isSubProduct bool)
		processProducts = func(products printer_model.Products, isSubProduct bool) {
			for _, product := range products {
				// 设置行间距
				printer.SetLineSpacing(45)
				// 处理自助餐文本
				buffetText := ""
				if buffetSignOpen == "1" {
					if product.IsBuffet {
						buffetText = t.base.Translate("自助餐") + "-"
					}
				}
				// 打包商品
				wrapText := ""
				if product.IsWrap && !isSubProduct {
					wrapText = "(" + t.base.Translate("打包") + ") "
				}
				// 产品名称
				productName := utils.IfString(isSubProduct, "-", "") + wrapText + buffetText + product.ProductName.GetLocale(t.base.Lang)
				// 打印产品名称和数量
				productNum := "x" + t.base.FloatToString(product.TotalNum)
				if len(productNum) >= 3 {
					printer.AppendText(t.base.PrintText(productName, "", productNum, 47))
					printer.LineFeed()
				} else {
					printer.PrintInColumns(productName, productNum)
				}

				// 设置字符大小和行间距
				printer.SetCharacterSize(1, 1)
				// 打印行间距和换行
				if isSubProduct {
					printer.LineFeed()
				}
				if printerType == PrinterTypeXPrinterLan {
					printer.SetLineSpacing(90)
				} else if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(10)
					printer.LineFeed()
					printer.SetLineSpacing(90)
				} else {
					printer.SetLineSpacing(45)
				}
				// 分割处理属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					if attr.GetLocale(t.base.Lang) == "" {
						continue
					}
					printer.AppendText(utils.IfString(isSubProduct, " ", "") + attr.GetLocale(t.base.Lang))
					printer.LineFeed()
				}

				// 处理备注
				if product.Remark != "" {
					// 设置行间距
					if t.base.IsMyText(product.Remark) {
						printer.SetLineSpacing(85)
					} else {
						printer.SetLineSpacing(55)
					}
					printer.SetCharacterSize(2, 2)
					printer.PrintInColumns(product.Remark)
					printer.SetCharacterSize(1, 1)
					printer.SetLineSpacing(20)
					printer.LineFeed()
				}

				// 打印行间距和换行
				printer.SetLineSpacing(12)
				printer.LineFeed()

				// 打印套餐子商品
				if len(product.SubProducts) > 0 {
					printer.LineFeed()
					printer.LineFeed()
					processProducts(product.SubProducts, true)
				}
			}
		}
		processProducts(products, false)

	} else {
		/**
		 * 模版二
		 */
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetCharacterSize(2, 2)
		printer.SetPrintModes(true, true, false)

		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 60
			if t.base.IsMyText(order.SerialNo) {
				spacing = 80
			}
			if printerType == PrinterTypeXPrinterWifi {
				spacing = 120
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
			printer.RestoreDefaultLineSpacing()
			printer.LineFeed()
		} else if order.IsTakeout() {
			printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo + "\n")
		} else if order.SerialNo != "" {
			printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
		}

		printer.LineFeed()
		printer.SetLineSpacing(50)

		printer.RestoreDefaultLineSpacing()
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{260, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)

		// 订单号
		printer.PrintInColumns(t.base.Translate("订单号"), order.OrderNo)

		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.RestoreDefaultLineSpacing()
		printer.PrintInColumns(t.base.Translate("时间"), updateTime)
		printer.LineFeed()

		// 设置行间距
		if printerType != PrinterTypeXPrinterWifi {
			printer.SetLineSpacing(40)
		}

		// 根据语言显示不同的文本
		if t.base.Lang == "my" {
			printer.AppendText("ကုန်စည်                                  ပမာဏ")
		} else {
			printer.AppendText(t.base.PrintText(t.base.Translate("商品"), "", t.base.Translate("数量"), 47))
		}

		// 添加分隔线
		printer.AppendText("\n------------------------------------------------\n")

		// 处理产品
		var processProducts func(products printer_model.Products, isSubProduct bool)
		processProducts = func(products printer_model.Products, isSubProduct bool) {
			for _, product := range products {
				// 处理自助餐文本
				buffetText := ""
				if buffetSignOpen == "1" {
					if product.IsBuffet {
						buffetText = t.base.Translate("自助餐") + "-"
					}
				}
				// 打包商品
				wrapText := ""
				if product.IsWrap && !isSubProduct {
					wrapText = "(" + t.base.Translate("打包") + ") "
				}
				// 产品名称
				productName := utils.IfString(isSubProduct, "-", "") + wrapText + buffetText + product.ProductName.GetLocale(t.base.Lang)

				// 设置列
				if t.base.IsThText(productName) {
					printer.SetupColumns(
						[]int{450, pkg.AlignLeft, 0},
						[]int{0, pkg.AlignRight, 0},
					)
				} else {
					printer.SetupColumns(
						[]int{480, pkg.AlignLeft, 0},
						[]int{0, pkg.AlignRight, 0},
					)
				}

				// 设置行间距
				if printerType == PrinterTypeXPrinterLan {
					printer.SetLineSpacing(20)
				} else if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(125)
				} else if t.base.IsMyText(productName) {
					printer.SetLineSpacing(80)
				} else {
					printer.SetLineSpacing(68)
				}
				// 设置字符大小
				printer.SetCharacterSize(2, 2)
				// 打印产品名称和数量
				productNum := "x" + t.base.FloatToString(product.TotalNum)
				if len(productNum) >= 3 {
					if isSunmi {
						printer.SetupColumns(
							[]int{490 - (len(productNum)-3)*30, pkg.AlignLeft, 0},
							[]int{0, pkg.AlignRight, 0},
						)
						printer.PrintInColumns(productName, productNum)
					} else {
						w := 20 - (len(productNum) - 3)
						printer.AppendText(t.base.PrintText(
							productName, "", productNum,
							w, w, 0, 0, 2,
						))
						printer.LineFeed()
					}
				} else {
					printer.PrintInColumns(productName, productNum)
				}
				// 设置字符大小和行间距
				printer.SetCharacterSize(1, 1)
				if printerType == PrinterTypeXPrinterLan {
					printer.SetLineSpacing(90)
				} else if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(20)
					printer.LineFeed()
					printer.SetLineSpacing(90)
				} else {
					printer.SetLineSpacing(50)
				}
				// 分割处理属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					if attr.GetLocale(t.base.Lang) == "" {
						continue
					}
					printer.AppendText(utils.IfString(isSubProduct, "  ", "") + attr.GetLocale(t.base.Lang))
					printer.SetLineSpacing(45)
					printer.LineFeed()
					if isSubProduct {
						printer.LineFeed()
					}
				}
				// 处理备注
				if product.Remark != "" {
					// 设置行间距
					if t.base.IsMyText(product.Remark) {
						printer.SetLineSpacing(85)
					} else {
						printer.SetLineSpacing(55)
					}
					printer.SetCharacterSize(2, 2)
					printer.PrintInColumns(product.Remark)
					printer.SetCharacterSize(1, 1)
					printer.SetLineSpacing(20)
					printer.LineFeed()
				}

				// 打印行间距和换行
				printer.SetLineSpacing(12)
				printer.LineFeed()

				// 打印套餐子商品
				if len(product.SubProducts) > 0 {
					processProducts(product.SubProducts, true)
				}
			}
		}
		processProducts(products, false)
	}

	// 恢复默认行间距
	printer.RestoreDefaultLineSpacing()

	// 打印额外行
	printer.LineFeed()
	printer.LineFeed()

	// 打印并退出页面模式
	printer.PrintAndExitPageMode()
	printer.LineFeed(4)
	printer.CutPaper(true)

	// 返回打印数据
	return printer.GetOrderData()
}

// OneDishOneOrder 一菜一单
func (t *dishesXprinterTemplate) OneDishOneOrder(
	tmpInfo model.PrinterTemplate,
	productPrinter model.ProductPrinter,
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {
	tmp := tmpInfo.Template
	isShowSku := tmpInfo.IsShowSku

	// 人的翻译
	name := t.base.Translate("人")
	// 自助餐标记开关
	buffetSignOpen := t.base.PrinterSetting.BuffetSignOpen
	// 格式化时间
	updateTime := t.base.FormatUnixTimeDefault(order.UpdateTime)
	// 就餐人数
	mealNumStr := ""
	if order.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", order.MealNum, name)
	}
	// 打印机类型
	printerType := PrinterTypeXPrinterLan
	if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
		printerType = printerItem.Printer.PrinterType.Key
	}

	// 是否是商米打印机
	isSunmi := false
	if printerType == PrinterTypeSunmiLan || printerType == PrinterTypeSunmiCloud || printerType == PrinterTypeCashierSunmi || printerType == PrinterTypeCodesoftWifi {
		isSunmi = true
	}

	// 是否有打印内容
	isPrinter := false

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.RestoreDefaultLineSpacing()

	/**
	 * 模版二
	 */
	if tmp == 2 {
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, true, false)
		printer.SetCharacterSize(2, 2)

		// 桌号
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 60
			if t.base.IsMyText(order.SerialNo) {
				spacing = 80
			}
			if printerType == PrinterTypeXPrinterWifi {
				spacing = 120
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
			printer.RestoreDefaultLineSpacing()
		} else if order.IsTakeout() {
			printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo)
		} else {
			printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + mealNumStr)
		}
		printer.LineFeed(3)

		// 设置对齐方式和列
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{490, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetPrintModes(false, false, false)

		// 遍历订单中的产品
		for _, product := range products {

			// 处理自助餐标记
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}
			// 打包商品
			wrapText := ""
			if product.IsWrap {
				wrapText = "(" + t.base.Translate("打包") + ") "
			}
			// 产品名称
			productName := wrapText + buffetText + product.ProductName.GetLocale(t.base.Lang)

			// 定义产品导出函数
			exportation := func(num float64) {
				// 设置行间距
				if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(120)
				} else {
					printer.SetLineSpacing(60)
				}
				printer.SetCharacterSize(2, 2)
				//
				productNum := "x" + t.base.FloatToString(num)
				if len(productNum) >= 3 {
					if isSunmi {
						printer.SetupColumns(
							[]int{490 - (len(productNum)-3)*30, pkg.AlignLeft, 0},
							[]int{0, pkg.AlignRight, 0},
						)
						printer.PrintInColumns(productName, productNum)
					} else {
						w := 20 - (len(productNum) - 3)
						printer.AppendText(t.base.PrintText(
							productName, "", productNum,
							w, w, 0, 0, 2,
						))
						printer.LineFeed()
					}
				} else {
					printer.PrintInColumns(productName, productNum)
				}
				//
				printer.SetCharacterSize(1, 2)
				printer.RestoreDefaultLineSpacing()

				// 行间距
				if printerType != PrinterTypeXPrinterLan {
					printer.LineFeed()
				}

				// 处理产品属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					printer.AppendText(attr.GetLocale(t.base.Lang))
					if printerType == PrinterTypeXPrinterWifi {
						printer.SetLineSpacing(40)
					} else {
						printer.SetLineSpacing(22)
					}
					printer.LineFeed(2)
					printer.RestoreDefaultLineSpacing()
				}

				// 处理备注
				if product.Remark != "" {
					// 设置行间距
					if printerType == PrinterTypeXPrinterWifi {
						printer.SetLineSpacing(120)
					} else {
						printer.SetLineSpacing(60)
					}
					printer.SetCharacterSize(2, 2)
					printer.AppendText(product.Remark)
					printer.SetCharacterSize(1, 1)
					if printerType == PrinterTypeXPrinterWifi {
						printer.SetLineSpacing(40)
					} else {
						printer.SetLineSpacing(22)
					}
					printer.LineFeed(2)
					printer.RestoreDefaultLineSpacing()
				}
				printer.LineFeed()
			}

			// 根据打印选择执行打印
			if productPrinter.PrintModeScene == 1 && product.NumType == 0 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1.0)
				}
			} else {
				exportation(product.TotalNum)
			}

			// 标记有内容被打印
			isPrinter = true
		}
		// 设置字符大小和打印模式
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.LineFeed()
		printer.AppendText("------------------------------------------------\n")
		printer.SetAlignment(pkg.AlignCenter)
		printer.AppendText(updateTime)

	} else {

		// 模版一
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetPrintModes(true, true, false)
		printer.SetCharacterSize(2, 2)

		// 处理桌号或取单号
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 60
			if t.base.IsMyText(order.SerialNo) {
				spacing = 80
			}
			if printerType == PrinterTypeXPrinterWifi {
				spacing = 120
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
			printer.RestoreDefaultLineSpacing()
		} else if order.IsTakeout() {
			printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo)
		} else {
			printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + mealNumStr)
		}

		// 行间距
		if printerType != PrinterTypeXPrinterLan {
			printer.LineFeed(1)
		}
		printer.LineFeed(1)

		// 设置字符大小和打印模式
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.AppendText(updateTime)
		printer.LineFeed(3)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{490, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)

		isPrinter = false

		// 遍历订单中的产品
		for _, product := range products {

			// 处理自助餐文本
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}
			// 打包商品
			wrapText := ""
			if product.IsWrap {
				wrapText = "(" + t.base.Translate("打包") + ") "
			}
			// 产品名称
			productName := wrapText + buffetText + product.ProductName.GetLocale(t.base.Lang)

			// 定义产品导出函数
			exportation := func(num float64) {
				// 设置行间距
				if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(120)
				} else {
					printer.SetLineSpacing(60)
				}
				printer.SetCharacterSize(2, 2)
				//
				productNum := "x" + t.base.FloatToString(num)
				if len(productNum) >= 3 {
					if isSunmi {
						printer.SetupColumns(
							[]int{490 - (len(productNum)-3)*30, pkg.AlignLeft, 0},
							[]int{0, pkg.AlignRight, 0},
						)
						printer.PrintInColumns(productName, productNum)
					} else {
						w := 20 - (len(productNum) - 3)
						printer.AppendText(t.base.PrintText(
							productName, "", productNum,
							w, w, 0, 0, 2,
						))
						printer.LineFeed()
					}
				} else {
					printer.PrintInColumns(productName, productNum)
				}
				//
				printer.SetCharacterSize(1, 2)
				printer.RestoreDefaultLineSpacing()

				// 行间距
				if printerType != PrinterTypeXPrinterLan {
					printer.LineFeed()
				}

				// 处理产品属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					printer.AppendText(attr.GetLocale(t.base.Lang))
					if printerType == PrinterTypeXPrinterWifi {
						printer.SetLineSpacing(40)
					} else {
						printer.SetLineSpacing(22)
					}
					printer.LineFeed(2)
					printer.RestoreDefaultLineSpacing()
				}

				// 处理备注
				if product.Remark != "" {
					printer.SetCharacterSize(2, 2)
					printer.AppendText(product.Remark)
					printer.SetCharacterSize(1, 1)
					if printerType == PrinterTypeXPrinterWifi {
						printer.SetLineSpacing(40)
					} else {
						printer.SetLineSpacing(22)
					}
					printer.LineFeed(2)
					printer.RestoreDefaultLineSpacing()
				}
				printer.LineFeed()
			}

			// 根据打印选择执行打印
			if productPrinter.PrintModeScene == 1 && product.NumType == 0 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1.0)
				}
			} else {
				exportation(product.TotalNum)
			}

			// 标记有内容被打印
			isPrinter = true
		}
	}
	// 如果没有内容被打印则返回空字符串
	if !isPrinter {
		return ""
	}

	printer.LineFeed()
	// 打印并退出页面模式
	printer.PrintAndExitPageMode()
	printer.LineFeed(6)
	printer.CutPaper(false)

	// 返回打印结果
	return printer.GetOrderData()
}

// ReturnMenuTemplate 退菜单模版
func (t *dishesXprinterTemplate) ReturnMenuTemplate(
	tmp int,
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {

	if len(products) == 0 {
		return ""
	}

	// 人的翻译
	name := t.base.Translate("人")
	// 自助餐标记开关
	buffetSignOpen := t.base.PrinterSetting.BuffetSignOpen
	// 格式化时间
	updateTime := t.base.FormatUnixTimeDefault(order.UpdateTime)
	// 就餐人数
	mealNumStr := ""
	if order.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", order.MealNum, name)
	}
	// 打印机类型
	printerType := PrinterTypeXPrinterLan
	if printerItem.Printer != nil && printerItem.Printer.PrinterType != nil {
		printerType = printerItem.Printer.PrinterType.Key
	}

	// 是否是商米打印机
	isSunmi := false
	if printerType == PrinterTypeSunmiLan || printerType == PrinterTypeSunmiCloud || printerType == PrinterTypeCashierSunmi || printerType == PrinterTypeCodesoftWifi {
		isSunmi = true
	}

	// 创建打印机实例
	printer := pkg.NewPrinter(567, "", "")
	if printerType != PrinterTypeXPrinterLan && printerType != PrinterTypeXPrinterWifi {
		printer.LineFeed(1)
	}
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetLineSpacing(30)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 2)
	if tmp == 2 {
		printer.AppendText("************************")
		printer.LineFeed(1)
	}
	printer.AppendText(t.base.Translate("退菜单"))
	if tmp == 2 {
		printer.LineFeed(1)
		printer.AppendText("************************")
	}
	printer.LineFeed(2)
	if printerType == PrinterTypeXPrinterWifi {
		printer.LineFeed()
	}
	// 桌号
	if order.DeskUuid > 0 {
		// 判断文字是否包含缅甸语
		spacing := 60
		if t.base.IsMyText(order.SerialNo) {
			spacing = 80
		}
		if printerType == PrinterTypeXPrinterWifi {
			spacing = 120
		}
		printer.SetLineSpacing(spacing)
		printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
		printer.SetLineSpacing(30)
		printer.LineFeed()
	} else if order.IsTakeout() {
		printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo)
	} else {
		printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo)
	}

	printer.LineFeed()
	printer.SetLineSpacing(50)
	printer.LineFeed()
	printer.RestoreDefaultLineSpacing()
	printer.SetCharacterSize(1, 1)
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetupColumns(
		[]int{260, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)

	// 订单号
	printer.PrintInColumns(t.base.Translate("订单号"), order.OrderNo)
	printer.SetLineSpacing(20)
	printer.LineFeed()
	printer.RestoreDefaultLineSpacing()
	printer.PrintInColumns(t.base.Translate("时间"), updateTime)
	printer.LineFeed()

	// 根据语言显示不同的文本
	if t.base.Lang == "my" {
		printer.AppendText("ကုန်စည်                                  ပမာဏ")
	} else {
		printer.AppendText(t.base.PrintText(t.base.Translate("商品"), "", t.base.Translate("数量"), 47))
	}

	// 添加分隔线
	printer.AppendText("\n------------------------------------------------\n")

	// 设置列
	printer.SetupColumns(
		[]int{500, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)

	// 处理产品
	refundText := utils.IfString(t.base.Lang == "tr", "iptal", t.base.Translate("退"))
	var processProducts func(products printer_model.Products, isSubProduct bool)
	processProducts = func(products printer_model.Products, isSubProduct bool) {
		for _, product := range products {
			// 设置行间距
			printer.SetLineSpacing(45)
			// 处理自助餐文本
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}
			// 产品名称
			productName := utils.IfString(tmp == 2, "!!!", "") + "(" + refundText + ") " + buffetText + product.ProductName.GetLocale(t.base.Lang)
			if isSubProduct {
				productName = "-" + buffetText + product.ProductName.GetLocale(t.base.Lang)
			}

			// 设置字符大小和行间距
			if printerType == PrinterTypeXPrinterWifi {
				printer.SetLineSpacing(120)
			} else {
				printer.SetLineSpacing(60)
			}

			// 打印产品名称和数量
			printer.SetCharacterSize(2, 2)
			productNum := utils.IfString(tmp == 2, "-", "x") + t.base.FloatToString(product.TotalNum)
			if len(productNum) >= 3 {
				if isSunmi {
					printer.SetupColumns(
						[]int{490 - (len(productNum)-3)*30, pkg.AlignLeft, 0},
						[]int{0, pkg.AlignRight, 0},
					)
					printer.PrintInColumns(productName, productNum)
				} else {
					w := 20 - (len(productNum) - 3)
					printer.AppendText(t.base.PrintText(
						productName, "", productNum,
						w, w, 0, 0, 2,
					))
					printer.LineFeed()
				}
			} else {
				printer.PrintInColumns(productName, productNum)
			}
			printer.SetCharacterSize(1, 1)

			// 恢复默认行间距
			printer.RestoreDefaultLineSpacing()

			// 换行
			if printerType != PrinterTypeXPrinterLan {
				printer.LineFeed()
			}

			// 分割处理属性
			for _, attr := range product.ProductAttrList {
				if attr.GetLocale(t.base.Lang) == "" {
					continue
				}
				printer.AppendText(utils.IfString(isSubProduct, "  ", "") + attr.GetLocale(t.base.Lang))
				if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(40)
				} else {
					printer.SetLineSpacing(22)
				}
				printer.LineFeed(2)
				// 恢复默认行间距
				printer.RestoreDefaultLineSpacing()
			}

			// 处理备注
			if product.Remark != "" {
				printer.AppendText(product.Remark)
				printer.SetCharacterSize(1, 1)
				if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(40)
				} else {
					printer.SetLineSpacing(22)
				}
				printer.LineFeed(2)
				// 恢复默认行间距
				printer.RestoreDefaultLineSpacing()
			}

			// 打印行间距和换行
			printer.LineFeed()

			// 打印套餐子商品
			if len(product.SubProducts) > 0 {
				processProducts(product.SubProducts, true)
			}

			// 退菜原因
			if !isSubProduct {
				printer.AppendText("------------------------------------------------")
				printer.LineFeed(1, 34)
				// 获取退菜原因文本
				reasonText := product.Reason.GetLocale(t.base.Lang)
				// 如果有自定义原因，则添加
				if product.CustomReason != "" {
					if reasonText != "" {
						reasonText += "、"
					}
					reasonText += product.CustomReason
				}
				printer.SetAlignment(pkg.AlignLeft)
				printer.AppendText(fmt.Sprintf("%s： %s", t.base.Translate("退菜原因"), reasonText))
			}
		}
	}
	processProducts(products, false)

	// 换行
	if tmp == 2 {
		printer.SetLineSpacing(30)
		printer.SetPrintModes(true, true, false)
		printer.SetCharacterSize(2, 2)
		printer.LineFeed(3)
		printer.AppendText("************************")
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignCenter)
		printer.AppendText(t.base.Translate("请停止制作以上菜品！"))
		printer.LineFeed(1)
		printer.AppendText("************************")
		printer.RestoreDefaultLineSpacing()
	}

	// 打印额外行
	printer.LineFeed(2)
	if printerType == PrinterTypeXPrinterLan || printerType == PrinterTypeXPrinterWifi {
		printer.LineFeed()
	}

	// 打印并退出页面模式
	printer.PrintAndExitPageMode()
	printer.LineFeed(6)
	printer.CutPaper(true)

	// 返回打印数据
	return printer.GetOrderData()
}

// OutMenuTemplate 出菜单模版
func (t *dishesXprinterTemplate) OutMenuTemplate(
	tmp int,
	mdPrinter model.Printer,
	order model.SaleBill,
	products printer_model.Products,
	finishedTime int64,
) string {

	if len(products) == 0 {
		return ""
	}

	// 人的翻译
	name := t.base.Translate("人")
	// 自助餐标记开关
	buffetSignOpen := t.base.PrinterSetting.BuffetSignOpen
	// 格式化时间
	updateTime := t.base.FormatUnixTimeDefault(finishedTime)
	// 就餐人数
	mealNumStr := ""
	if order.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", order.MealNum, name)
	}
	// 打印机类型
	printerType := PrinterTypeXPrinterLan
	if mdPrinter.PrinterType != nil {
		printerType = mdPrinter.PrinterType.Key
	}
	// 是否是商米打印机
	isSunmi := false
	if printerType == PrinterTypeSunmiLan || printerType == PrinterTypeSunmiCloud || printerType == PrinterTypeCashierSunmi || printerType == PrinterTypeCodesoftWifi {
		isSunmi = true
	}

	// 创建打印机实例
	printer := pkg.NewPrinter(567, "", "")
	if printerType != PrinterTypeXPrinterLan && printerType != PrinterTypeXPrinterWifi {
		printer.LineFeed(1)
	}
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetLineSpacing(30)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.Translate("出菜单"))
	printer.LineFeed(2)
	if printerType == PrinterTypeXPrinterWifi {
		printer.LineFeed()
	}
	// 桌号
	if order.IsTakeoutBill() {
		printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo)
	} else if order.DeskUuid > 0 {
		// 判断文字是否包含缅甸语
		spacing := 60
		if t.base.IsMyText(order.SerialNo) {
			spacing = 80
		}
		if printerType == PrinterTypeXPrinterWifi {
			spacing = 120
		}
		printer.SetLineSpacing(spacing)
		printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
		printer.SetLineSpacing(30)
		printer.LineFeed()
	} else if order.IsTakeout() {
		printer.AppendText(t.base.Translate("外送") + ": " + order.SerialNo)
	} else {
		printer.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo)
	}

	printer.LineFeed()
	printer.SetLineSpacing(50)
	printer.LineFeed()
	printer.RestoreDefaultLineSpacing()
	printer.SetCharacterSize(1, 1)
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetupColumns(
		[]int{260, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)

	// 订单号
	printer.PrintInColumns(t.base.Translate("订单号"), order.OrderNo)
	printer.SetLineSpacing(20)
	printer.LineFeed()
	printer.RestoreDefaultLineSpacing()
	printer.PrintInColumns(t.base.Translate("时间"), updateTime)
	printer.LineFeed()

	// 根据语言显示不同的文本
	if t.base.Lang == "my" {
		printer.AppendText("ကုန်စည်                                  ပမာဏ")
	} else {
		printer.AppendText(t.base.PrintText(t.base.Translate("商品"), "", t.base.Translate("数量"), 47))
	}

	// 添加分隔线
	printer.AppendText("\n------------------------------------------------\n")

	// 设置列
	printer.SetupColumns(
		[]int{500, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)

	// 处理产品
	for _, product := range products {
		// 设置行间距
		printer.SetLineSpacing(45)
		// 处理自助餐文本
		buffetText := ""
		if buffetSignOpen == "1" {
			if product.IsBuffet {
				buffetText = t.base.Translate("自助餐") + "-"
			}
		}
		// 打包商品
		wrapText := ""
		if product.IsWrap {
			wrapText = "(" + t.base.Translate("打包") + ") "
		}
		// 套餐
		packageText := ""
		if product.ProductType > constant.ProductTypeProduct {
			packageText = t.base.Translate("套餐") + "-"
		}
		// 产品名称
		productName := packageText + wrapText + buffetText + product.ProductName.GetLocale(t.base.Lang)

		// 设置字符大小和行间距
		if printerType == PrinterTypeXPrinterWifi {
			printer.SetLineSpacing(120)
		} else {
			printer.SetLineSpacing(60)
		}

		// 打印产品名称和数量
		printer.SetCharacterSize(2, 2)
		productNum := "x" + t.base.FloatToString(product.TotalNum)
		if len(productNum) >= 3 {
			if isSunmi {
				printer.SetupColumns(
					[]int{490 - (len(productNum)-3)*30, pkg.AlignLeft, 0},
					[]int{0, pkg.AlignRight, 0},
				)
				printer.PrintInColumns(productName, productNum)
			} else {
				w := 20 - (len(productNum) - 3)
				printer.AppendText(t.base.PrintText(
					productName, "", productNum,
					w, w, 0, 0, 2,
				))
				printer.LineFeed()
			}
		} else {
			printer.PrintInColumns(productName, productNum)
		}
		printer.SetCharacterSize(1, 1)

		// 恢复默认行间距
		printer.RestoreDefaultLineSpacing()

		// 换行
		if printerType != PrinterTypeXPrinterLan {
			printer.LineFeed()
		}

		// 分割处理属性
		for _, attr := range product.ProductAttrList {
			printer.AppendText(attr.GetLocale(t.base.Lang))
			if printerType == PrinterTypeXPrinterWifi {
				printer.SetLineSpacing(40)
			} else {
				printer.SetLineSpacing(22)
			}
			printer.LineFeed(2)
			// 恢复默认行间距
			printer.RestoreDefaultLineSpacing()
		}

		// 处理备注
		if product.Remark != "" {
			printer.AppendText(product.Remark)
			printer.SetCharacterSize(1, 1)
			if printerType == PrinterTypeXPrinterWifi {
				printer.SetLineSpacing(40)
			} else {
				printer.SetLineSpacing(22)
			}
			printer.LineFeed(2)
			// 恢复默认行间距
			printer.RestoreDefaultLineSpacing()
		}
	}

	// 打印额外行
	printer.LineFeed(2)
	if printerType == PrinterTypeXPrinterLan || printerType == PrinterTypeXPrinterWifi {
		printer.LineFeed()
	}

	// 打印并退出页面模式
	printer.PrintAndExitPageMode()
	printer.LineFeed(6)
	printer.CutPaper(true)

	// 返回打印数据
	return printer.GetOrderData()
}
