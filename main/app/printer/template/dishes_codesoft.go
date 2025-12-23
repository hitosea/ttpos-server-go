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

// printerTemplate Codesoft菜品打印模板
type dishesCodesoftTemplate struct {
	base *printerTemplate
}

// NewDishesCodesoftTemplate 创建新的Codesoft菜品打印模板
func NewDishesCodesoftTemplate(
	base *printerTemplate,
) *dishesCodesoftTemplate {
	return &dishesCodesoftTemplate{
		base: base,
	}
}

// CompleteOrder 整单模版
func (t *dishesCodesoftTemplate) CompleteOrder(
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

	// 分批类型
	batchTagText := ""
	for _, product := range products {
		if product.BatchTagUuid > 0 {
			batchTag, err := repository.NewBatchTagRepo(t.base.Ctx.GetDB()).GetBatchTagInfo(product.BatchTagUuid)
			if err == nil {
				batchTagText = batchTag.MultiLanguageName.GetNameByLang(t.base.Lang)
				if product.ShowDelayTag {
					postText := model.MultiLanguageName{
						EnName:   "Post Cooking",
						ZhName:   "稍后制作",
						ZhTwName: "稍後製作",
						ThName:   "ทำหลัง",
						MyName:   "အပြီးမှာဖွင့်ပါ",
						JaName:   "後で作る",
						KoName:   "나중에 만들어",
						TrName:   "Sonra yap",
						SvName:   "Gör efter",
					}
					batchTagText = batchTagText + " (" + postText.GetNameByLang(t.base.Lang) + ")"
				}
				break
			}
		}
	}

	// 订单来源为外卖的文本
	orderSourceTakeoutText := t.base.GetOrderSourceTakeoutText(order.GetOrderSourceTakeoutText())

	// 创建打印机实例
	printer := pkg.NewPrinter(567, "", "")
	printer.LineFeed(1)

	/**
	 * 模版一
	 */
	if temp == 1 {
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetCharacterSize(2, 2, true)
		printer.SetPrintModes(true, true, false)

		// 桌号
		serialNoText := order.SerialNo
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 60
			if t.base.IsMyText(serialNoText) {
				spacing = 80
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
			printer.RestoreDefaultLineSpacing()
			printer.LineFeed()
		} else if order.IsTakeoutBill() {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText + "\n")
		} else {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText + "\n")
		}

		// 整单备注
		if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
			printer.AppendText("\n------------------------------------------------\n")
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
			printer.AppendText("\n------------------------------------------------\n")
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

		// 分批类型
		if batchTagText != "" && !order.IsTakeoutBill() {
			printer.SetCharacterSize(2, 2)
			printer.SetPrintModes(true, true, false)
			printer.SetAlignment(pkg.AlignCenter)
			printer.AppendText(batchTagText)
			printer.LineFeed(2)
			printer.SetCharacterSize(1, 1)
			printer.SetPrintModes(false, false, false)
		}

		// 设置行间距
		printer.SetLineSpacing(40)

		// 根据语言显示不同的文本
		if t.base.Lang == "my" {
			printer.AppendText("ကုန်စည်                                  ပမာဏ")
		} else {
			printer.AppendText(t.base.PrintText(t.base.Translate("商品"), "", t.base.Translate("数量"), 47))
		}

		// 添加分隔线
		printer.AppendText("\n------------------------------------------------\n")

		printer.SetAlignment(pkg.AlignLeft)

		// 设置列
		printer.SetupColumns(
			[]int{490, pkg.AlignLeft, 0},
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
					w := 20 - (len(productNum) - 3)
					printer.AppendText(t.base.PrintText(productName, "", productNum, w, w, 0, 0, 2))
					printer.LineFeed()
				} else {
					printer.PrintInColumns(productName, productNum)
				}

				// 设置字符大小和行间距
				printer.SetCharacterSize(1, 1)
				// 打印行间距和换行
				printer.LineFeed()
				printer.SetLineSpacing(45)
				// 分割处理属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					if attr.GetLocale(t.base.Lang) == "" {
						continue
					}
					printer.AppendText(utils.IfString(isSubProduct, " ", "") + attr.GetLocale(t.base.Lang))
					printer.LineFeed()
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
					// 设置行间距
					if t.base.IsMyText(remarkText) {
						printer.SetLineSpacing(85)
					} else {
						printer.SetLineSpacing(55)
					}

					printer.SetCharacterSize(2, 2)
					printer.PrintInColumns(remarkText)
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
					processProducts(product.SubProducts, false)
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

		serialNoText := order.SerialNo
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 60
			if t.base.IsMyText(serialNoText) {
				spacing = 80
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
			printer.RestoreDefaultLineSpacing()
			printer.LineFeed()
		} else if order.IsTakeoutBill() {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText + "\n")
		} else {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText + "\n")
		}

		// 整单备注
		if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText("\n------------------------------------------------\n")
			printer.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
			printer.AppendText("\n------------------------------------------------\n")
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
		printer.PrintInColumns(t.base.Translate("时间"), updateTime) // 这里应该用i18n包来处理"时间"的翻译
		printer.LineFeed()

		// 分批类型
		if batchTagText != "" && !order.IsTakeoutBill() {
			printer.SetCharacterSize(2, 2)
			printer.SetPrintModes(true, true, false)
			printer.SetAlignment(pkg.AlignCenter)
			printer.AppendText(batchTagText)
			printer.LineFeed(2)
			printer.SetAlignment(pkg.AlignLeft)
			printer.SetCharacterSize(1, 1)
			printer.SetPrintModes(false, false, false)
		}

		// 设置行间距
		printer.SetLineSpacing(40)

		// 根据语言显示不同的文本
		if t.base.Lang == "my" {
			// 缅甸语文本
			printer.AppendText("ကုန်စည်                                  ပမာဏ")
		} else {
			printer.AppendText(t.base.PrintText(t.base.Translate("商品"), "", t.base.Translate("数量"), 47))
		}

		// 添加分隔线
		printer.AppendText("\n------------------------------------------------\n")

		// 设置列
		printer.SetupColumns(
			[]int{490, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)

		// 处理产品
		// 处理产品
		var processProducts func(products printer_model.Products, isSubProduct bool)
		processProducts = func(products printer_model.Products, isSubProduct bool) {
			for _, product := range products {
				// 设置字符大小
				printer.SetCharacterSize(2, 2)
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
				// 设置行间距
				if t.base.IsMyText(productName) {
					printer.SetLineSpacing(80)
				} else {
					printer.SetLineSpacing(68)
				}
				// 打印产品名称和数量
				if len(productNum) >= 3 {
					w := 20 - (len(productNum) - 3)
					printer.AppendText(t.base.PrintText(
						productName, "", productNum,
						w, w, 0, 0, 2,
					))
					printer.LineFeed()
				} else {
					printer.PrintInColumns(productName, productNum)
				}
				// 设置字符大小和行间距
				printer.SetCharacterSize(1, 1)
				printer.SetLineSpacing(50)
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
					// 设置行间距
					if t.base.IsMyText(remarkText) {
						printer.SetLineSpacing(85)
					} else {
						printer.SetLineSpacing(55)
					}
					printer.SetCharacterSize(2, 2)
					printer.PrintInColumns(remarkText)
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
					processProducts(product.SubProducts, false)
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
	printer.CutPaper(printerItem.Printer.IsEnableSound())

	// 返回打印数据
	return printer.GetOrderData()
}

// OneDishOneOrder 一菜一单
func (t *dishesCodesoftTemplate) OneDishOneOrder(
	tmpInfo model.PrinterTemplate,
	productPrinter model.ProductPrinter,
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {

	templateID := tmpInfo.Template
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
	// 是否有打印内容
	isPrinter := false

	// 分批类型
	batchTagText := ""
	for _, product := range products {
		if product.BatchTagUuid > 0 {
			batchTag, err := repository.NewBatchTagRepo(t.base.Ctx.GetDB()).GetBatchTagInfo(product.BatchTagUuid)
			if err == nil {
				batchTagText = batchTag.MultiLanguageName.GetNameByLang(t.base.Lang)
				if product.ShowDelayTag {
					postText := model.MultiLanguageName{
						EnName:   "Post Cooking",
						ZhName:   "稍后制作",
						ZhTwName: "稍後製作",
						ThName:   "ทำหลัง",
						MyName:   "အပြီးမှာဖွင့်ပါ",
						JaName:   "後で作る",
						KoName:   "나중에 만들어",
						TrName:   "Sonra yap",
						SvName:   "Gör efter",
					}
					batchTagText = batchTagText + " (" + postText.GetNameByLang(t.base.Lang) + ")"
				}
				break
			}
		}
	}

	// 订单来源为外卖的文本
	orderSourceTakeoutText := t.base.GetOrderSourceTakeoutText(order.GetOrderSourceTakeoutText())

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
		serialNoText := order.SerialNo
		if order.DeskUuid > 0 {
			spacing := 60
			if t.base.IsMyText(serialNoText) { // 判断文字是否包含缅甸语
				spacing = 80
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
			printer.RestoreDefaultLineSpacing()
		} else if order.IsTakeoutBill() {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText)
		} else {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText + mealNumStr)
		}

		// 整单备注
		if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
			printer.LineFeed(1)
			printer.AppendText("\n------------------------\n")
			printer.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
			printer.AppendText("\n------------------------\n")
		}

		printer.LineFeed(3)

		// 分批类型
		if batchTagText != "" && !order.IsTakeoutBill() {
			printer.SetAlignment(pkg.AlignCenter)
			printer.AppendText(batchTagText)
			printer.LineFeed(2)
		}

		// 设置对齐方式和列
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{490, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)

		// 遍历订单中的产品
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
			// 处理自助餐标记
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
				// 设置行间距
				printer.SetLineSpacing(60)
				printer.SetCharacterSize(2, 2)
				//
				productNum := "x" + t.base.FloatToString(num)
				if len(productNum) >= 3 {
					w := 20 - (len(productNum) - 3)
					printer.AppendText(t.base.PrintText(
						productName, "", productNum,
						w, w, 0, 0, 2,
					))
					printer.LineFeed()
				} else {
					printer.PrintInColumns(productName, productNum)
				}
				//
				printer.SetCharacterSize(1, 2)
				printer.RestoreDefaultLineSpacing()
				printer.LineFeed()

				// 处理产品属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					printer.AppendText(attr.GetLocale(t.base.Lang))
					printer.SetLineSpacing(22)
					printer.LineFeed(2)
					printer.RestoreDefaultLineSpacing()
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
					printer.SetLineSpacing(60)
					printer.SetCharacterSize(2, 2)
					printer.AppendText(remarkText)
					printer.SetCharacterSize(1, 1)
					printer.SetLineSpacing(22)
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
		serialNoText := order.SerialNo
		if order.IsOrderSourceTakeout() {
			serialNoText = t.base.Translate("外卖") + " " + serialNoText
		}
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 60
			if t.base.IsMyText(serialNoText) {
				spacing = 80
			}
			printer.SetLineSpacing(spacing)
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("桌号") + ": " + serialNoText + mealNumStr)
			printer.RestoreDefaultLineSpacing()
		} else if order.IsTakeoutBill() {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("外送") + ": " + serialNoText)
		} else {
			printer.AppendText(orderSourceTakeoutText + t.base.Translate("取单号") + ": " + serialNoText + mealNumStr)
		}

		// 整单备注
		if orderRemark := order.GetLatestOrderRemarkRes(); orderRemark != nil {
			printer.LineFeed(1)
			printer.AppendText("\n------------------------\n")
			printer.SetLineSpacing(120)
			printer.AppendText(t.base.Translate("整单备注") + ": " + orderRemark.Remark.GetLocale(t.base.Lang))
			printer.RestoreDefaultLineSpacing()
			printer.AppendText("\n------------------------\n")
		} else {
			printer.LineFeed(2)
		}

		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.AppendText(updateTime)

		// 分批类型
		if batchTagText != "" && !order.IsTakeoutBill() {
			printer.SetCharacterSize(2, 2)
			printer.SetPrintModes(true, true, false)
			printer.LineFeed(2)
			printer.SetAlignment(pkg.AlignCenter)
			printer.AppendText(batchTagText)
		}

		printer.LineFeed(3)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{490, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)

		isPrinter = false

		// 遍历订单中的产品
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
				// 设置行间距
				printer.SetLineSpacing(60)
				printer.SetCharacterSize(2, 2)
				//
				productNum := "x" + t.base.FloatToString(num)
				if len(productNum) >= 3 {
					w := 20 - (len(productNum) - 3)
					printer.AppendText(t.base.PrintText(
						productName, "", productNum,
						w, w, 0, 0, 2,
					))
					printer.LineFeed()
				} else {
					printer.PrintInColumns(productName, productNum)
				}
				//
				printer.SetCharacterSize(1, 2)
				printer.RestoreDefaultLineSpacing()
				printer.LineFeed()

				// 处理产品属性
				for _, attr := range utils.IfSlice(isShowSku == 0, product.ProductSauceNamesList, product.ProductAttrList) {
					printer.AppendText(attr.GetLocale(t.base.Lang))
					printer.SetLineSpacing(22)
					printer.LineFeed(2)
					printer.RestoreDefaultLineSpacing()
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
					printer.SetCharacterSize(2, 2)
					printer.AppendText(remarkText)
					printer.SetCharacterSize(1, 1)
					printer.SetLineSpacing(22)
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
	printer.CutPaper(printerItem.Printer.IsEnableSound())

	// 返回打印结果
	return printer.GetOrderData()
}
