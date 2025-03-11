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

// 打印机类型枚举
const (
	PrinterTypeEnumXPrinterWifi = "xprinter_wifi"
)

// printerTemplate Codesoft菜品打印模板
type dishesCodesoftTemplate struct {
	base *printerTemplate
}

// NewdishesCodesoftTemplate 创建新的Codesoft菜品打印模板
func NewDishesCodesoftTemplate(
	ctx context.Context,
	setting *setting.Srv,
	storeSetting *respSetting.Store,
	printerSetting *respSetting.Printer,
	currencySetting *respSetting.Currency,
) *dishesCodesoftTemplate {
	return &dishesCodesoftTemplate{
		base: NewPrinterTemplate(ctx, setting, storeSetting, printerSetting, currencySetting, false),
	}
}

// ExtractLanguage 提取语言
// 这是一个参考 PHP 的 extractLanguage 函数的 Go 实现
func ExtractLanguage(input string) string {
	// TODO: 在这里实现更复杂的提取语言逻辑
	return input
}

// CompleteOrder 整单模版
func (t *dishesCodesoftTemplate) CompleteOrder(
	printerItem *model.ProductPrinterItem,
	order model.SaleBill,
	products printer_model.Products,
) string {
	// 人的翻译
	name := t.base.Translate("人")
	// 获取模板 ID，这里简化处理，实际应该从数据库或配置中获取
	templateID := 1
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

	// 创建打印机实例
	printer := pkg.NewPrinter(567, "", "")
	printer.LineFeed(1)

	/**
	 * 模版一
	 */
	if templateID == 1 {
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
		printer.SetLineSpacing(40)

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
			[]int{490, pkg.AlignLeft, 0},
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
			totalNum := fmt.Sprintf("%d", product.TotalNum)
			// 打印产品名称和数量
			printer.PrintInColumns(productName, totalNum)

			// 设置字符大小和行间距
			printer.SetCharacterSize(1, 1)
			printer.SetLineSpacing(45)
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
			printer.SetLineSpacing(spacing)
			printer.AppendText(t.base.Translate("桌号") + ": " + order.SerialNo + mealNumStr)
			printer.RestoreDefaultLineSpacing()
			printer.LineFeed()
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
			// 产品名称
			productName := buffetText + product.ProductName.GetLocale(t.base.Lang)
			// 打印产品名称和数量
			totalNum := fmt.Sprintf("%d", product.TotalNum)
			// 设置行间距
			if t.base.IsMyText(productName) {
				printer.SetLineSpacing(80)
			} else {
				printer.SetLineSpacing(68)
			}

			// 打印产品名称和数量
			printer.PrintInColumns(productName, totalNum)
			// 设置字符大小和行间距
			printer.SetCharacterSize(1, 1)
			printer.SetLineSpacing(50)
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
// func (t *dishesCodesoftTemplate) OneDishOneOrder(
// 	printerConfig map[string]interface{},
// 	printerItem map[string]interface{},
// 	order map[string]interface{},
// 	products interface{},
// ) string {
// 	// 人的翻译
// 	name := t.base.Translate("人")

// 	// 获取模板ID，这里简化处理
// 	templateID := 1

// 	// 获取打印机类型
// 	printerType := ""
// 	if printer, ok := printerItem["printer"].(map[string]interface{}); ok {
// 		if printerTypeInfo, ok := printer["printer_type"].(map[string]interface{}); ok {
// 			if value, ok := printerTypeInfo["value"].(string); ok {
// 				printerType = value
// 			}
// 		}
// 	}

// 	// 自助餐标记开关
// 	buffetSignOpen := 1
// 	if value, ok := printerConfig["buffet_sign_open"].(string); ok {
// 		if intValue, err := strconv.Atoi(value); err == nil {
// 			buffetSignOpen = intValue
// 		}
// 	} else if value, ok := printerConfig["buffet_sign_open"].(int); ok {
// 		buffetSignOpen = value
// 	}

// 	// 格式化时间
// 	updateTime := ""
// 	if updateTimeStr, ok := order["update_time"].(string); ok {
// 		// 转换时间格式
// 		if parsedTime, err := time.Parse(time.RFC3339, updateTimeStr); err == nil {
// 			updateTime = parsedTime.Format("2006/01/02 15:04:05")
// 		} else {
// 			// 如果解析失败，直接使用原始字符串
// 			updateTime = updateTimeStr
// 		}
// 	} else {
// 		// 使用当前时间
// 		updateTime = time.Now().Format("2006/01/02 15:04:05")
// 	}

// 	// 佛历判断（简化处理）
// 	if t.base.DefaultCalendar == 3 {
// 		// 这里应该调用转换佛历的函数，这里简化处理
// 		// updateTime = DateHelp.ChangeBuddhistCalendar(updateTime)
// 	}

// 	// 创建打印机实例
// 	printer := pkg.NewPrinter(567)
// 	printer.LineFeed()
// 	printer.RestoreDefaultLineSpacing()

// 	// 是否有打印内容
// 	isPrinter := false

// 	/**
// 	 * 模版二
// 	 */
// 	if templateID == 2 {
// 		printer.SetAlignment(pkg.AlignCenter)
// 		printer.SetPrintModes(true, true, false)
// 		printer.SetCharacterSize(2, 2)

// 		// 处理桌号或取单号
// 		if tableNo, ok := order["table_no"].(string); ok && tableNo != "" {
// 			// 设置行间距
// 			spacing := 60
// 			if t.base.IsMyText(tableNo) {
// 				spacing = 80
// 			}

// 			printer.SetLineSpacing(spacing)

// 			// 处理就餐人数
// 			mealNumStr := ""
// 			if mealNum, ok := order["meal_num"].(string); ok && mealNum != "" {
// 				mealNumStr = fmt.Sprintf(" (%s%s)", mealNum, name)
// 			} else if mealNum, ok := order["meal_num"].(int); ok && mealNum > 0 {
// 				mealNumStr = fmt.Sprintf(" (%d%s)", mealNum, name)
// 			}

// 			printer.AppendText("桌号" + ": " + tableNo + mealNumStr)
// 			printer.RestoreDefaultLineSpacing()
// 		} else if callNo, ok := order["call_no"].(string); ok && callNo != "" {
// 			// 处理就餐人数
// 			mealNumStr := ""
// 			if mealNum, ok := order["meal_num"].(string); ok && mealNum != "" {
// 				mealNumStr = fmt.Sprintf(" (%s%s)", mealNum, name)
// 			} else if mealNum, ok := order["meal_num"].(int); ok && mealNum > 0 {
// 				mealNumStr = fmt.Sprintf(" (%d%s)", mealNum, name)
// 			}

// 			printer.AppendText("取单号" + ": " + callNo + mealNumStr)
// 		}

// 		printer.LineFeed()
// 		printer.LineFeed()
// 		printer.LineFeed()

// 		// 设置对齐方式和列
// 		printer.SetAlignment(pkg.AlignLeft)
// 		printer.SetupColumns(
// 			[]int{490, pkg.AlignLeft, 0},
// 			[]int{0, pkg.AlignRight, 0},
// 		)
// 		// 遍历订单中的产品
// 		if orderProducts, ok := order["product"].([]interface{}); ok {
// 			for _, product := range orderProducts {

// 				// 处理自助餐标记
// 				buffetText := ""
// 				if buffetSignOpen == 1 {
// 					if product.IsBuffet {
// 						buffetText = "自助餐 - "
// 					}
// 				}

// 				// 获取产品名称
// 				productNameText := ""
// 				if nameText, ok := productDetail["product_name_text"].(string); ok {
// 					productNameText = nameText
// 				}
// 				productName := buffetText + productNameText

// 				// 处理产品属性
// 				productAttr := ""
// 				if attrStr, ok := product["product_attr"].(string); ok {
// 					// 去除前后分号
// 					productAttr = strings.Trim(attrStr, ";")
// 				}

// 				// 分割产品属性
// 				productAttrs := []string{}
// 				if productAttr != "" {
// 					// 替换并分割
// 					productAttr = strings.ReplaceAll(productAttr, "};{", "}|-*-|-*-|{")
// 					productAttrs = strings.Split(productAttr, "|-*-|-*-|")
// 				}

// 				// 定义产品导出函数
// 				exportation := func(num int) {
// 					// 设置行间距
// 					printer.SetLineSpacing(60)
// 					printer.SetCharacterSize(2, 2)
// 					printer.PrintInColumns(productName, strconv.Itoa(num))
// 					printer.SetCharacterSize(1, 2)

// 					printer.RestoreDefaultLineSpacing()
// 					printer.LineFeed()

// 					// 处理产品属性
// 					for _, productAttr := range productAttrs {
// 						printer.AppendText(ExtractLanguage(productAttr))
// 						if printerType == PrinterTypeEnumXPrinterWifi {
// 							printer.SetLineSpacing(40)
// 						} else {
// 							printer.SetLineSpacing(22)
// 						}
// 						printer.LineFeed(2)
// 						printer.RestoreDefaultLineSpacing()
// 					}

// 					// 处理备注
// 					if remark, ok := product["remark"].(string); ok && remark != "" {
// 						if printerType == PrinterTypeEnumXPrinterWifi {
// 							printer.SetLineSpacing(120)
// 						} else {
// 							printer.SetLineSpacing(60)
// 						}
// 						printer.SetCharacterSize(2, 2)
// 						printer.AppendText(remark)
// 						printer.SetCharacterSize(1, 1)
// 						if printerType == PrinterTypeEnumXPrinterWifi {
// 							printer.SetLineSpacing(40)
// 						} else {
// 							printer.SetLineSpacing(22)
// 						}
// 						printer.LineFeed(2)
// 						printer.RestoreDefaultLineSpacing()
// 					}
// 					printer.LineFeed()
// 				}

// 				// 处理打印选择
// 				printSelect := 1
// 				if val, ok := printerItem["print_select"].(int); ok {
// 					printSelect = val
// 				} else if valStr, ok := printerItem["print_select"].(string); ok {
// 					if intVal, err := strconv.Atoi(valStr); err == nil {
// 						printSelect = intVal
// 					}
// 				}

// 				// 获取产品数量
// 				totalNum := 0
// 				if num, ok := product["total_num"].(int); ok {
// 					totalNum = num
// 				} else if numStr, ok := product["total_num"].(string); ok {
// 					if intNum, err := strconv.Atoi(numStr); err == nil {
// 						totalNum = intNum
// 					}
// 				}

// 				// 根据打印选择执行打印
// 				if printSelect == 2 {
// 					for i := 0; i < totalNum; i++ {
// 						exportation(1)
// 					}
// 				} else {
// 					exportation(totalNum)
// 				}

// 				// 标记有内容被打印
// 				isPrinter = true
// 			}
// 		}
// 		// 设置字符大小和打印模式
// 		printer.SetCharacterSize(1, 1)
// 		printer.SetPrintModes(false, false, false)
// 		printer.LineFeed()
// 		printer.AppendText("------------------------------------------------\n")
// 		printer.SetAlignment(pkg.AlignCenter)
// 		printer.AppendText(updateTime)
// 	} else { // 模版一
// 		printer.SetAlignment(pkg.AlignLeft)
// 		printer.SetPrintModes(true, true, false)
// 		printer.SetCharacterSize(2, 2)

// 		// 处理桌号或取单号
// 		if tableNo, ok := order["table_no"].(string); ok && tableNo != "" {
// 			// 设置行间距
// 			spacing := 60
// 			if t.base.IsMyText(tableNo) {
// 				spacing = 80
// 			}

// 			printer.SetLineSpacing(spacing)

// 			// 处理就餐人数
// 			mealNumStr := ""
// 			if mealNum, ok := order["meal_num"].(string); ok && mealNum != "" {
// 				mealNumStr = fmt.Sprintf(" (%s%s)", mealNum, name)
// 			} else if mealNum, ok := order["meal_num"].(int); ok && mealNum > 0 {
// 				mealNumStr = fmt.Sprintf(" (%d%s)", mealNum, name)
// 			}

// 			printer.AppendText("桌号" + ": " + tableNo + mealNumStr)
// 			printer.RestoreDefaultLineSpacing()
// 		} else if callNo, ok := order["call_no"].(string); ok && callNo != "" {
// 			// 处理就餐人数
// 			mealNumStr := ""
// 			if mealNum, ok := order["meal_num"].(string); ok && mealNum != "" {
// 				mealNumStr = fmt.Sprintf(" (%s%s)", mealNum, name)
// 			} else if mealNum, ok := order["meal_num"].(int); ok && mealNum > 0 {
// 				mealNumStr = fmt.Sprintf(" (%d%s)", mealNum, name)
// 			}

// 			printer.AppendText("取单号" + ": " + callNo + mealNumStr)
// 		}
// 		printer.LineFeed()
// 		printer.LineFeed()
// 		printer.SetCharacterSize(1, 1)
// 		printer.SetPrintModes(false, false, false)
// 		printer.AppendText(updateTime)
// 		printer.LineFeed()
// 		printer.LineFeed()
// 		printer.LineFeed()
// 		printer.SetAlignment(pkg.AlignLeft)
// 		printer.SetupColumns(
// 			[]int{490, pkg.AlignLeft, 0},
// 			[]int{0, pkg.AlignRight, 0},
// 		)

// 		isPrinter = false

// 		// 遍历订单中的产品
// 		if orderProducts, ok := order["product"].([]interface{}); ok {
// 			for _, productInterface := range orderProducts {
// 				product, ok := productInterface.(map[string]interface{})
// 				if !ok {
// 					continue
// 				}

// 				// 如果指定了特定产品，跳过不匹配的产品
// 				if products != nil {
// 					// 使用 MD5 哈希比较两个 JSON 对象
// 					productsBytes, _ := json.Marshal(products)
// 					productBytes, _ := json.Marshal(product)

// 					productsHash := md5.Sum(productsBytes)
// 					productHash := md5.Sum(productBytes)

// 					if hex.EncodeToString(productsHash[:]) != hex.EncodeToString(productHash[:]) {
// 						continue
// 					}
// 				}
// 				// 处理自助餐标记
// 				buffetText := ""
// 				if buffetSignOpen == 1 {
// 					if isBuffet, ok := product["is_buffet_product"].(int); ok && isBuffet == 1 {
// 						buffetText = "自助餐 - "
// 					} else if isBuffetStr, ok := product["is_buffet_product"].(string); ok && isBuffetStr == "1" {
// 						buffetText = "自助餐 - "
// 					}
// 				}

// 				// 获取产品名称
// 				productNameText := ""
// 				if nameText, ok := productDetail["product_name_text"].(string); ok {
// 					productNameText = nameText
// 				}
// 				productName := buffetText + productNameText

// 				// 处理产品属性
// 				productAttr := ""
// 				if attrStr, ok := product["product_attr"].(string); ok {
// 					// 去除前后分号
// 					productAttr = strings.Trim(attrStr, ";")
// 				}

// 				// 分割产品属性
// 				productAttrs := []string{}
// 				if productAttr != "" {
// 					// 替换并分割
// 					productAttr = strings.ReplaceAll(productAttr, "};{", "}|-*-|-*-|{")
// 					productAttrs = strings.Split(productAttr, "|-*-|-*-|")
// 				}
// 				// 定义产品导出函数
// 				exportation := func(num int) {
// 					// 设置行间距
// 					if printerType == PrinterTypeEnumXPrinterWifi {
// 						printer.SetLineSpacing(120)
// 					} else {
// 						printer.SetLineSpacing(60)
// 					}
// 					printer.SetCharacterSize(2, 2)
// 					printer.PrintInColumns(productName, strconv.Itoa(num))
// 					printer.SetCharacterSize(1, 2)

// 					printer.RestoreDefaultLineSpacing()
// 					printer.LineFeed()

// 					// 处理产品属性
// 					for _, attrText := range productAttrs {
// 						printer.AppendText(ExtractLanguage(attrText))
// 						if printerType == PrinterTypeEnumXPrinterWifi {
// 							printer.SetLineSpacing(40)
// 						} else {
// 							printer.SetLineSpacing(22)
// 						}
// 						printer.LineFeed(2)
// 						printer.RestoreDefaultLineSpacing()
// 					}

// 					// 处理备注
// 					if remark, ok := product["remark"].(string); ok && remark != "" {
// 						printer.SetCharacterSize(2, 2)
// 						printer.AppendText(remark)
// 						printer.SetCharacterSize(1, 1)
// 						if printerType == PrinterTypeEnumXPrinterWifi {
// 							printer.SetLineSpacing(40)
// 						} else {
// 							printer.SetLineSpacing(22)
// 						}
// 						printer.LineFeed(2)
// 						printer.RestoreDefaultLineSpacing()
// 					}
// 					printer.LineFeed()
// 				}

// 				// 获取打印选择
// 				printSelect := 1
// 				if val, ok := printerItem["print_select"].(int); ok {
// 					printSelect = val
// 				} else if valStr, ok := printerItem["print_select"].(string); ok {
// 					if intVal, err := strconv.Atoi(valStr); err == nil {
// 						printSelect = intVal
// 					}
// 				}

// 				// 获取产品数量
// 				totalNum := 0
// 				if num, ok := product["total_num"].(int); ok {
// 					totalNum = num
// 				} else if numStr, ok := product["total_num"].(string); ok {
// 					if intNum, err := strconv.Atoi(numStr); err == nil {
// 						totalNum = intNum
// 					}
// 				}

// 				// 根据打印选择执行打印
// 				if printSelect == 2 {
// 					for i := 0; i < totalNum; i++ {
// 						exportation(1)
// 					}
// 				} else {
// 					exportation(totalNum)
// 				}
// 				// 标记有内容被打印
// 				isPrinter = true
// 			}
// 		}
// 	}
// 	// 如果没有内容被打印则返回空字符串
// 	if !isPrinter {
// 		return ""
// 	}

// 	printer.LineFeed()
// 	// 打印并退出页面模式
// 	printer.PrintAndExitPageMode()
// 	printer.LineFeed(6)
// 	printer.CutPaper(false)

// 	// 返回打印结果
// 	return printer.GetOrderData()
// }
