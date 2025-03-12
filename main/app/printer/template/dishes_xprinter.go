// Package template 提供打印模板相关功能
package template

import (
	"fmt"

	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
)

// dishesXprinterTemplate xprinter菜品打印模板
type dishesXprinterTemplate struct {
	base *printerTemplate
}

// NewdishesXprinterTemplate 创建新的xprinter菜品打印模板
func NewDishesXprinterTemplate(
	ctx context.Context,
	setting *setting.Srv,
	storeSetting *respSetting.Store,
	printerSetting *respSetting.Printer,
	currencySetting *respSetting.Currency,
) *dishesXprinterTemplate {
	return &dishesXprinterTemplate{
		base: NewPrinterTemplate(ctx, setting, storeSetting, printerSetting, currencySetting, false),
	}
}

// CompleteOrder 整单模版
func (t *dishesXprinterTemplate) CompleteOrder(
	temp int,
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {
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
	// 是否有打印内容
	isPrinter := false

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
			productName := buffetText + product.ProductName.GetLocale(t.base.Lang)
			// 打印产品名称和数量
			printer.PrintInColumns(productName, "x"+fmt.Sprintf("%d", product.TotalNum))

			// 设置字符大小和行间距
			printer.SetCharacterSize(1, 1)
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
			for _, attr := range product.ProductAttrList {
				printer.AppendText(attr.GetLocale(t.base.Lang))
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

			// 标记已有打印内容
			isPrinter = true
		}
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
		for _, product := range products {
			// 处理自助餐文本
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}
			// 产品名称
			productName := buffetText + product.ProductName.GetLocale(t.base.Lang)

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
			printer.PrintInColumns(productName, "x"+fmt.Sprintf("%d", product.TotalNum))
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
			for _, attr := range product.ProductAttrList {
				printer.AppendText(attr.GetLocale(t.base.Lang))
				printer.SetLineSpacing(45)
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

			// 标记已有打印内容
			isPrinter = true
		}
	}

	// 恢复默认行间距
	printer.RestoreDefaultLineSpacing()

	// 检查是否有打印内容
	if !isPrinter {
		return ""
	}

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
	templateID int,
	productPrinter model.ProductPrinter,
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {
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
	// 是否有打印内容
	isPrinter := false

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.LineFeed()
	printer.RestoreDefaultLineSpacing()

	/**
	 * 模版二
	 */
	if templateID == 2 {
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

		// 遍历订单中的产品
		for _, product := range products {

			// 处理自助餐标记
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}
			// 产品名称
			productName := buffetText + product.ProductName.GetLocale(t.base.Lang)

			// 定义产品导出函数
			exportation := func(num uint) {
				// 设置行间距
				if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(120)
				} else {
					printer.SetLineSpacing(60)
				}
				printer.SetCharacterSize(2, 2)
				printer.PrintInColumns(productName, "x"+fmt.Sprintf("%d", num))
				printer.SetCharacterSize(1, 2)
				printer.RestoreDefaultLineSpacing()

				// 行间距
				if printerType != PrinterTypeXPrinterLan {
					printer.LineFeed()
				}

				// 处理产品属性
				for _, attr := range product.ProductAttrList {
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
			if productPrinter.PrintModeScene == 1 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1)
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

			// 产品名称
			productName := buffetText + product.ProductName.GetLocale(t.base.Lang)

			// 定义产品导出函数
			exportation := func(num int) {
				// 设置行间距
				if printerType == PrinterTypeXPrinterWifi {
					printer.SetLineSpacing(120)
				} else {
					printer.SetLineSpacing(60)
				}
				printer.SetCharacterSize(2, 2)
				printer.PrintInColumns(productName, "x"+fmt.Sprintf("%d", num))
				printer.SetCharacterSize(1, 2)
				printer.RestoreDefaultLineSpacing()

				// 行间距
				if printerType != PrinterTypeXPrinterLan {
					printer.LineFeed()
				}

				// 处理产品属性
				for _, attr := range product.ProductAttrList {
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
			if productPrinter.PrintModeScene == 1 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1)
				}
			} else {
				exportation(int(product.TotalNum))
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
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {
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
	// 是否有打印内容
	isPrinter := false

	// 创建打印机实例
	printer := pkg.NewPrinter(567, "", "")
	if printerType != PrinterTypeXPrinterLan && printerType != PrinterTypeXPrinterWifi {
		printer.LineFeed(1)
	}
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetLineSpacing(30)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.Translate("退菜单"))
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
		// 产品名称
		productName := "(" + t.base.Translate("退") + ") " + buffetText + product.ProductName.GetLocale(t.base.Lang)

		// 设置字符大小和行间距
		if printerType == PrinterTypeXPrinterWifi {
			printer.SetLineSpacing(120)
		} else {
			printer.SetLineSpacing(60)
		}

		// 打印产品名称和数量
		printer.SetCharacterSize(2, 2)
		printer.PrintInColumns(productName, "x"+fmt.Sprintf("%d", product.TotalNum))
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

		// 打印行间距和换行
		printer.LineFeed()

		// 退菜原因
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(1, 34)
		printer.AppendText(t.base.Translate("退菜原因") + "： " + product.ReturnReason)

		// 标记已有打印内容
		isPrinter = true
	}

	// 检查是否有打印内容
	if !isPrinter {
		return ""
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
