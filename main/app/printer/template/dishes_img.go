// Package template 提供打印模板相关功能
package template

import (
	"fmt"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/utils"
)

// dishesImgTemplate 图片菜品打印模板
type dishesImgTemplate struct {
	base *printerTemplate
}

// NewDishesImgTemplate 创建新的图片菜品打印模板
func NewDishesImgTemplate(
	base *printerTemplate,
) *dishesImgTemplate {
	return &dishesImgTemplate{
		base: base,
	}
}

// CompleteOrder 整单模版
func (t *dishesImgTemplate) CompleteOrder(
	tmpInfo model.PrinterTemplate,
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {
	// 返回打印结果
	if len(products) == 0 {
		return ""
	}
	tmp := tmpInfo.Template
	// 人的翻译
	name := t.base.Translate("人")
	// 格式化时间
	updateTime := t.base.FormatUnixTimeDefault(order.UpdateTime)
	// 就餐人数
	mealNumStr := ""
	if order.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", order.MealNum, name)
	}

	// 订单来源为外卖的文本
	orderSourceTakeoutText := t.base.GetOrderSourceTakeoutText(order.GetOrderSourceTakeoutText())

	// 创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetAlignment(pkg.AlignCenter)
	img.SetImagePadding(0) // 确保没有填充
	img.SetFontWeight(5)
	img.SetFontSize(30)

	// 桌号
	serialNoText := order.SerialNo
	if order.DeskUuid > 0 {
		// 判断文字是否包含缅甸语
		spacing := 50
		if t.base.IsMyText(serialNoText) {
			spacing = 68
		}
		img.SetTextLineHeight(spacing)
		img.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("桌号"), serialNoText, mealNumStr))
		img.SetTextLineHeight(45)
		img.LineFeed(1, 20)
	} else if order.IsTakeoutBill() {
		img.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText + "\n")
	} else {
		img.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText + "\n")
	}

	// 整单备注
	if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
		img.LineFeed(1, 20)
		img.AppendSplitLine()
		img.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
		img.LineFeed(1)
		img.AppendSplitLine()
	}

	// 订单号和时间
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.SetTextLineHeight(36)
	img.LineFeed(2)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("订单号"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: order.OrderNo, Width: 0, Align: pkg.AlignRight},
	)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("时间"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: updateTime, Width: 0, Align: pkg.AlignRight},
	)
	img.LineFeed(1)

	// 分批类型
	batchTagText := ""
	for _, product := range products {
		if product.BatchTagUuid > 0 {
			batchTag, err := repository.NewBatchTagRepo(t.base.Ctx.GetDB()).GetBatchTagInfo(product.BatchTagUuid)
			if err == nil {
				batchTagText = batchTag.MultiLanguageName.GetNameByLang(t.base.Lang)
				break
			}
		}
	}
	if batchTagText != "" && !order.IsTakeoutBill() {
		img.SetFontSize(28)
		img.SetFontWeight(2)
		img.SetAlignment(pkg.AlignCenter)
		img.AppendText(batchTagText)
		img.LineFeed(1, 60)
		img.SetFontSize(20)
		img.SetFontWeight(1)
	}

	// 商品和数量
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("商品"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.Translate("数量"), Width: 0, Align: pkg.AlignRight},
	)
	img.AppendSplitLine()
	img.LineFeed(1)
	img.SetTextLineHeight(utils.IfInt(tmp == 2 || tmp == 3, 50, 40))
	//
	t.base.PrintCompleteOrderImgProducts(img, tmpInfo, products)
	//
	img.LineFeed(3, 110)
	//
	return img.SetSegmentationHeight(200).Save("", !t.base.IsSunMi && printerItem.Printer.IsEnableSound(), 0)
}

// OneDishOneOrder 一菜一单模版
func (t *dishesImgTemplate) OneDishOneOrder(
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
	// 是否打印
	isPrinter := false
	// 创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)

	// 分批类型
	batchTagText := ""
	for _, product := range products {
		if product.BatchTagUuid > 0 {
			batchTag, err := repository.NewBatchTagRepo(t.base.Ctx.GetDB()).GetBatchTagInfo(product.BatchTagUuid)
			if err == nil {
				batchTagText = batchTag.MultiLanguageName.GetNameByLang(t.base.Lang)
				break
			}
		}
	}

	// 订单来源为外卖的文本
	orderSourceTakeoutText := t.base.GetOrderSourceTakeoutText(order.GetOrderSourceTakeoutText())

	if tmp == 2 {
		img.SetAlignment(pkg.AlignCenter)
		img.SetImagePadding(0) // 确保没有填充
		img.SetFontWeight(5)
		img.SetFontSize(30)

		// 桌号
		serialNoText := order.SerialNo
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 50
			if t.base.IsMyText(serialNoText) {
				spacing = 68
			}
			img.SetTextLineHeight(spacing)
			img.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("桌号"), serialNoText, mealNumStr))
			img.SetTextLineHeight(45)
			img.LineFeed(1, 20)
		} else if order.IsTakeoutBill() {
			img.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText)
		} else {
			img.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText)
		}

		// 整单备注
		if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
			img.LineFeed(1, 60)
			img.AppendSplitLine()
			img.SetAlignment(pkg.AlignLeft)
			img.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
			img.SetAlignment(pkg.AlignCenter)
			img.LineFeed(1)
			img.AppendSplitLine()
		}

		// 设置边距
		img.SetFontSize(20)
		img.SetFontWeight(1)
		img.SetTextLineHeight(36)
		img.LineFeed(1, 40)
		img.LineFeed(2)
		img.SetTextLineHeight(50)

		// 分批类型
		if batchTagText != "" && !order.IsTakeoutBill() {
			img.SetFontSize(28)
			img.SetFontWeight(2)
			img.AppendText(batchTagText)
			img.LineFeed(1, 60)
			img.SetFontSize(20)
			img.SetFontWeight(1)
		}

		// 商品和数量
		for _, product := range products {

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
			// 处理自助餐文本
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}

			// 产品名称
			productName := wrapText + packageText + buffetText + product.ProductName.GetLocale(t.base.Lang)

			// 定义产品导出函数
			exportation := func(num float64) {
				if t.base.Lang == "my" {
					img.SetTextLineHeight(90)
				} else {
					img.SetTextLineHeight(64)
				}
				img.LineFeed(1, 12)
				totalNum := "x" + t.base.FloatToString(num)
				productNameWidth := utils.IfInt(len(totalNum) >= 3, 470-(len(totalNum)*utils.IfInt(len(totalNum) > 5, 10, 7)), 480)
				img.PrintInColumns(
					pkg.ColumnConfig{Text: productName, Width: productNameWidth, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
					pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 50},
				)
				if t.base.Lang == "my" {
					img.LineFeed(1, 12)
				}

				// 分割处理属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					img.SetFontSize(24)
					if t.base.Lang == "my" {
						img.SetTextLineHeight(50)
					} else {
						img.SetTextLineHeight(40)
					}
					img.AppendText(attr.GetLocale(t.base.Lang))
					img.SetFontSize(20)
					img.LineFeed(1, 50)
				}

				// 处理备注（包含预设备注和自定义备注）
				var remarkText string
				if !product.RemarkLocale.IsNull() {
					// 使用预构建的多语言备注
					remarkText = product.RemarkLocale.GetLocale(t.base.Lang)
				} else {
					// 向后兼容：如果没有预构建的备注，使用原有备注
					remarkText = product.Remark
				}

				if remarkText != "" {
					if t.base.IsMyText(remarkText) {
						img.SetTextLineHeight(68)
					} else {
						img.SetTextLineHeight(50)
					}
					img.LineFeed(1, 12)
					img.SetFontSize(28)
					img.AppendText(remarkText)
					img.LineFeed(1, 50)
					img.SetFontSize(20)
				}

				img.LineFeed(1, 12)
				img.SetTextLineHeight(50)
			}

			// 根据打印选择执行打印
			if productPrinter.PrintModeScene == 1 && product.NumType == 0 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1.0)
				}
			} else {
				exportation(product.TotalNum)
			}
			// 标记已打印
			isPrinter = true
		}

		//
		img.LineFeed(1, 40)
		img.SetTextLineHeight(27)
		img.AppendSplitLine(pkg.WithLineFeed(true))
		img.SetFontSize(20)
		img.SetAlignment(pkg.AlignCenter)
		img.AppendText(updateTime)
		img.LineFeed(1, 70)
		//
	} else {
		img.SetAlignment(pkg.AlignLeft)
		img.SetImagePadding(0) // 确保没有填充
		img.SetFontWeight(5)
		img.SetFontSize(32)

		// 桌号
		serialNoText := order.SerialNo
		if order.DeskUuid > 0 {
			spacing := 50
			if t.base.IsMyText(serialNoText) {
				spacing = 68
			}
			img.SetTextLineHeight(spacing)
			img.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("桌号"), serialNoText, mealNumStr))
			img.SetTextLineHeight(45)
		} else if order.IsTakeoutBill() {
			img.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText)
		} else {
			img.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText)
		}

		// 整单备注
		if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
			img.LineFeed(1, 60)
			img.AppendSplitLine()
			img.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
			img.LineFeed(1)
			img.AppendSplitLine()
		}

		img.LineFeed(1, 50)
		img.SetFontSize(20)
		img.AppendText(updateTime)
		img.LineFeed(1)
		img.LineFeed(1, 24)

		// 分批类型
		if batchTagText != "" && !order.IsTakeoutBill() {
			img.SetFontSize(28)
			img.SetFontWeight(2)
			img.SetAlignment(pkg.AlignCenter)
			img.AppendText(batchTagText)
			img.LineFeed(1, 65)
			img.SetFontSize(20)
			img.SetFontWeight(1)
			img.SetAlignment(pkg.AlignLeft)
		}

		// 商品和数量
		for _, product := range products {
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
			// 处理自助餐文本
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}

			// 产品名称
			productName := wrapText + packageText + buffetText + product.ProductName.GetLocale(t.base.Lang)

			// 定义产品导出函数
			exportation := func(num float64) {
				if t.base.Lang == "my" {
					img.SetTextLineHeight(90)
				} else {
					img.SetTextLineHeight(64)
				}
				img.LineFeed(1, 12)
				totalNum := "x" + t.base.FloatToString(num)
				productNameWidth := utils.IfInt(len(totalNum) >= 3, 470-(len(totalNum)*utils.IfInt(len(totalNum) > 5, 10, 7)), 480)
				img.PrintInColumns(
					pkg.ColumnConfig{Text: productName, Width: productNameWidth, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
					pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 50},
				)
				if t.base.Lang == "my" {
					img.LineFeed(1, 12)
				}

				// 分割处理属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					img.SetFontSize(24)
					if t.base.Lang == "my" {
						img.SetTextLineHeight(50)
					} else {
						img.SetTextLineHeight(40)
					}
					img.AppendText(attr.GetLocale(t.base.Lang))
					img.SetFontSize(20)
					img.LineFeed(1, 50)
				}

				// 处理备注（包含预设备注和自定义备注）
				var remarkText string
				if !product.RemarkLocale.IsNull() {
					// 使用预构建的多语言备注
					remarkText = product.RemarkLocale.GetLocale(t.base.Lang)
				} else {
					// 向后兼容：如果没有预构建的备注，使用原有备注
					remarkText = product.Remark
				}

				if remarkText != "" {
					if t.base.IsMyText(remarkText) {
						img.SetTextLineHeight(68)
					} else {
						img.SetTextLineHeight(50)
					}
					img.LineFeed(1, 12)
					img.SetFontSize(28)
					img.AppendText(remarkText)
					img.LineFeed(1, 50)
					img.SetFontSize(20)
				}

				img.LineFeed(1, 12)
				img.SetTextLineHeight(50)
			}

			// 根据打印选择执行打印
			if productPrinter.PrintModeScene == 1 && product.NumType == 0 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1.0)
				}
			} else {
				exportation(product.TotalNum)
			}
			// 标记已打印
			isPrinter = true
		}
		img.LineFeed(2)
	}
	if !isPrinter {
		return ""
	}
	// Print and exit page mode
	img.SetTextLineHeight(30)
	img.LineFeed(4)
	//
	return img.SetSegmentationHeight(200).Save("", !t.base.IsSunMi && printerItem.Printer.IsEnableSound(), 0)
}

// ReturnMenuTemplate 退菜单模版
func (t *dishesImgTemplate) ReturnMenuTemplate(
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

	// 订单来源为外卖的文本
	orderSourceTakeoutText := t.base.GetOrderSourceTakeoutText(order.GetOrderSourceTakeoutText())

	// 创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetAlignment(pkg.AlignCenter)
	img.SetImagePadding(0) // 确保没有填充
	img.SetFontWeight(5)
	img.SetFontSize(30)
	if tmp == 2 {
		img.AppendText("***************************")
		img.LineFeed(1, 40)
	}
	menuTitle := t.base.Translate("退菜单")
	img.AppendText(menuTitle)
	if tmp == 2 {
		img.LineFeed(1, 48)
		img.AppendText("***************************")
	}
	img.LineFeed(1, 68)

	// 桌号
	serialNoText := order.SerialNo
	if order.DeskUuid > 0 {
		// 判断文字是否包含缅甸语
		spacing := 50
		if t.base.IsMyText(serialNoText) {
			spacing = 68
		}
		img.SetTextLineHeight(spacing)
		img.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("桌号"), serialNoText, mealNumStr))
		img.SetTextLineHeight(45)
		img.LineFeed(1)
	} else if order.IsTakeoutBill() {
		img.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText + "\n")
	} else {
		img.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText + "\n")
	}

	// 整单备注
	if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
		img.LineFeed(1, 20)
		img.AppendSplitLine()
		img.SetAlignment(pkg.AlignLeft)
		img.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
		img.SetAlignment(pkg.AlignCenter)
		img.LineFeed(1)
		img.AppendSplitLine()
		img.LineFeed(1, 20)
	}

	// 设置字体
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.SetTextLineHeight(36)
	img.LineFeed(2, 32)

	// 订单号和时间
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("订单号"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: order.OrderNo, Width: 0, Align: pkg.AlignRight},
	)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("时间"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: updateTime, Width: 0, Align: pkg.AlignRight},
	)
	img.LineFeed(1)

	// 商品和数量 - 标题
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("商品"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.Translate("数量"), Width: 0, Align: pkg.AlignRight},
	)
	img.AppendSplitLine()
	img.LineFeed(1)
	img.SetTextLineHeight(50)

	// 商品和数量
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

			// 产品名称
			productName := utils.IfString(tmp == 2, "!!!", "") + "(" + t.base.Translate("退") + ") " + buffetText + product.ProductName.GetLocale(t.base.Lang)
			if isSubProduct {
				productName = "--" + buffetText + product.ProductName.GetLocale(t.base.Lang)
			}

			// 设置行间距
			img.SetTextLineHeight(utils.IfInt(t.base.Lang == "my", 90, 64))
			img.LineFeed(1, 12)

			// 打印产品名称和数量
			totalNum := utils.IfString(tmp == 2, "-", "x") + t.base.FloatToString(product.TotalNum)
			productNameWidth := utils.IfInt(len(totalNum) >= 3, 470-(len(totalNum)*utils.IfInt(len(totalNum) > 5, 10, 7)), 480)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: productName, Width: productNameWidth, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
				pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 50},
			)
			if t.base.Lang == "my" {
				img.LineFeed(1, 12)
			}

			// 分割处理属性
			for _, attr := range product.ProductAttrList {
				if attr.GetLocale(t.base.Lang) == "" {
					continue
				}
				img.SetTextLineHeight(utils.IfInt(t.base.Lang == "my", 50, 40))
				img.AppendText(utils.IfString(isSubProduct, "  ", "") + attr.GetLocale(t.base.Lang))
				img.LineFeed(1, 50)
			}

			// 处理备注（包含预设备注和自定义备注）
			var remarkText string
			if !product.RemarkLocale.IsNull() {
				// 使用预构建的多语言备注
				remarkText = product.RemarkLocale.GetLocale(t.base.Lang)
			} else {
				// 向后兼容：如果没有预构建的备注，使用原有备注
				remarkText = product.Remark
			}

			if remarkText != "" {
				img.SetTextLineHeight(utils.IfInt(t.base.IsMyText(remarkText), 50, 40))
				img.AppendText(remarkText)
				img.LineFeed(1, 50)
			}

			img.LineFeed(1, 12)
			img.SetTextLineHeight(50)

			// 打印套餐子商品
			if len(product.SubProducts) > 0 {
				processProducts(product.SubProducts, true)
			}

			// 退菜原因
			if !isSubProduct {
				img.AppendSplitLine()
				img.LineFeed(1, 34)
				// 获取退菜原因文本
				reasonText := product.Reason.GetLocale(t.base.Lang)
				// 如果有自定义原因，则添加
				if product.CustomReason != "" {
					if reasonText != "" {
						reasonText += "、"
					}
					reasonText += product.CustomReason
				}
				img.AppendText(fmt.Sprintf("%s： %s", t.base.Translate("退菜原因"), reasonText))
			}

		}
	}
	processProducts(products, false)

	// 换行
	if tmp == 2 {
		img.LineFeed(1, 80)
		img.SetFontWeight(5)
		img.SetFontSize(30)
		img.SetAlignment(pkg.AlignCenter)
		img.AppendText("***************************")
		img.LineFeed(1, 40)
		img.AppendText(t.base.Translate("请停止制作以上菜品！"))
		img.LineFeed(1, 48)
		img.AppendText("***************************")
	} else {
		img.LineFeed(1, 50)
	}

	// 设置行间距
	img.SetTextLineHeight(50)
	// 换行
	img.LineFeed(3)
	//
	return img.SetSegmentationHeight(200).Save("", !t.base.IsSunMi && printerItem.Printer.IsEnableSound(), 0)
}

// OutMenuTemplate 出菜单模版
func (t *dishesImgTemplate) OutMenuTemplate(
	tmp int,
	printer model.Printer,
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
	// 订单来源为外卖的文本
	orderSourceTakeoutText := t.base.GetOrderSourceTakeoutText(order.GetOrderSourceTakeoutText())

	// 创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetAlignment(pkg.AlignCenter)
	img.SetImagePadding(0) // 确保没有填充
	img.SetFontWeight(5)
	img.SetFontSize(30)
	img.AppendText(t.base.Translate("出菜单"))
	img.LineFeed(1, 68)

	// 桌号
	if order.IsTakeoutBill() {
		img.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + order.SerialNo + "\n")
	} else if order.DeskUuid > 0 {
		// 判断文字是否包含缅甸语
		spacing := 50
		if t.base.IsMyText(order.SerialNo) {
			spacing = 68
		}
		img.SetTextLineHeight(spacing)
		img.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("桌号"), order.SerialNo, mealNumStr))
		img.SetTextLineHeight(45)
		img.LineFeed(1)
	} else {
		img.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
	}

	// 整单备注
	if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
		img.LineFeed(1, 20)
		img.AppendSplitLine()
		img.SetAlignment(pkg.AlignLeft)
		img.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
		img.SetAlignment(pkg.AlignCenter)
		img.LineFeed(1)
		img.AppendSplitLine()
		img.LineFeed(1, 20)
	}

	// 设置字体
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.SetTextLineHeight(36)
	img.LineFeed(2, 32)

	// 订单号和时间
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("订单号"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: order.OrderNo, Width: 0, Align: pkg.AlignRight},
	)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("时间"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: updateTime, Width: 0, Align: pkg.AlignRight},
	)
	img.LineFeed(1)

	// 商品和数量 - 标题
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("商品"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.Translate("数量"), Width: 0, Align: pkg.AlignRight},
	)
	img.AppendSplitLine()
	img.LineFeed(1)
	img.SetTextLineHeight(50)

	isMultiProduct := len(products) > 0
	var processProducts func(products printer_model.Products, isSubProduct bool)
	processProducts = func(products printer_model.Products, isSubProduct bool) {
		// 商品和数量
		for _, product := range products {

			// 打包商品
			wrapText := ""
			if product.IsWrap {
				wrapText = "(" + t.base.Translate("打包") + ") "
			}
			// 套餐
			packageText := ""
			if product.ProductType > constant.ProductTypeProduct && !isMultiProduct {
				packageText = t.base.Translate("套餐") + "-"
			}
			// 处理自助餐文本
			buffetText := ""
			if buffetSignOpen == "1" {
				if product.IsBuffet {
					buffetText = t.base.Translate("自助餐") + "-"
				}
			}
			// 产品名称
			productName := wrapText + packageText + buffetText + product.ProductName.GetLocale(t.base.Lang)
			// 是否是套餐子商品
			if isSubProduct {
				productName = "--" + buffetText + product.ProductName.GetLocale(t.base.Lang)
			}

			// 设置行间距
			img.SetTextLineHeight(utils.IfInt(t.base.Lang == "my", 90, 64))
			img.LineFeed(1, 12)

			// 打印产品名称和数量
			totalNum := "X" + t.base.FloatToString(product.TotalNum)
			productNameWidth := utils.IfInt(len(totalNum) >= 3, 470-(len(totalNum)*utils.IfInt(len(totalNum) > 5, 10, 7)), 480)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: productName, Width: productNameWidth, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
				pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 50},
			)
			if t.base.Lang == "my" {
				img.LineFeed(1, 12)
			}

			// 分割处理属性
			for _, attr := range product.ProductAttrList {
				if attr.GetLocale(t.base.Lang) != "" {
					img.SetTextLineHeight(utils.IfInt(t.base.Lang == "my", 50, 40))
					img.AppendText(attr.GetLocale(t.base.Lang))
					img.LineFeed(1, 50)
				}
			}

			// 处理备注（包含预设备注和自定义备注）
			var remarkText string
			if !product.RemarkLocale.IsNull() {
				// 使用预构建的多语言备注
				remarkText = product.RemarkLocale.GetLocale(t.base.Lang)
			} else {
				// 向后兼容：如果没有预构建的备注，使用原有备注
				remarkText = product.Remark
			}

			if remarkText != "" {
				img.SetTextLineHeight(utils.IfInt(t.base.IsMyText(remarkText), 50, 40))
				img.AppendText(remarkText)
				img.LineFeed(1, 50)
			}

			// 打印套餐子商品
			if len(product.SubProducts) > 0 {
				processProducts(product.SubProducts, true)
			}

			img.LineFeed(1, 12)
			img.SetTextLineHeight(50)
		}
	}
	processProducts(products, false)

	// 设置行间距
	img.SetTextLineHeight(50)
	// 换行
	img.LineFeed(3)
	//
	return img.SetSegmentationHeight(200).Save("", !t.base.IsSunMi && printer.IsEnableSound(), 0)
}
