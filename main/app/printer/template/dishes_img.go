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

// dishesImgTemplate 图片菜品打印模板
type dishesImgTemplate struct {
	base *printerTemplate
}

// NewDishesImgTemplate 创建新的图片菜品打印模板
func NewDishesImgTemplate(
	ctx context.Context,
	setting *setting.Srv,
	storeSetting *respSetting.Store,
	printerSetting *respSetting.Printer,
	currencySetting *respSetting.Currency,
) *dishesImgTemplate {
	return &dishesImgTemplate{
		base: NewPrinterTemplate(ctx, setting, storeSetting, printerSetting, currencySetting, false),
	}
}

// CompleteOrder 整单模版
func (t *dishesImgTemplate) CompleteOrder(
	tmp int,
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

	// 创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetAlignment(pkg.AlignCenter)
	img.SetImagePadding(0) // 确保没有填充
	img.SetFontWeight(5)
	img.SetFontSize(30)

	// 桌号
	if order.DeskUuid > 0 {
		// 判断文字是否包含缅甸语
		spacing := 50
		if t.base.IsMyText(order.SerialNo) {
			spacing = 68
		}
		img.SetTextLineHeight(spacing)
		img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("桌号"), order.SerialNo, mealNumStr))
		img.SetTextLineHeight(45)
		img.LineFeed(1, 20)
	} else {
		img.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
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

	// 商品和数量
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("商品"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.Translate("数量"), Width: 0, Align: pkg.AlignRight},
	)
	img.AppendSplitLine()
	img.LineFeed(1)

	if tmp == 2 {
		img.SetTextLineHeight(50)
	} else {
		img.SetTextLineHeight(40)
	}
	isPrinter := false
	for _, product := range products {
		// 处理自助餐文本
		buffetText := ""
		if buffetSignOpen == "1" {
			if product.IsBuffet {
				buffetText = t.base.Translate("自助餐") + "-"
			}
		}

		if tmp == 2 {
			if t.base.Lang == "my" {
				img.SetTextLineHeight(75)
			} else {
				img.SetTextLineHeight(60)
			}
			img.SetTextLineHeight(60)
		} else {
			img.SetTextLineHeight(50)
		}
		img.LineFeed(1, 12)

		// 产品名称
		productName := buffetText + product.ProductName.GetLocale(t.base.Lang)
		// 打印产品名称和数量
		totalNum := "x" + fmt.Sprintf("%d", product.TotalNum)
		if tmp == 2 {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: productName, Width: 500, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
				pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 42},
			)
		} else {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: productName, Width: 500, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 20},
				pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 20, LineHeight: 42},
			)
		}
		if t.base.Lang == "my" {
			img.LineFeed(1, 12)
		}

		// 分割处理属性
		for _, attr := range product.ProductAttrList {
			if t.base.Lang == "my" {
				img.SetTextLineHeight(50)
			} else {
				img.SetTextLineHeight(40)
			}
			img.AppendText(attr.GetLocale(t.base.Lang))
			img.LineFeed(1, 40)
		}

		if product.Remark != "" {
			if t.base.IsMyText(product.Remark) {
				img.SetTextLineHeight(68)
			} else {
				img.SetTextLineHeight(50)
			}
			img.LineFeed(1, 12)
			img.SetFontSize(28)
			img.AppendText(product.Remark)
			img.LineFeed(1, 50)
			img.SetFontSize(20)
		}

		img.LineFeed(1, 12)
		img.SetTextLineHeight(50)
		isPrinter = true
	}

	// 返回打印结果
	if !isPrinter {
		return ""
	}
	//
	img.LineFeed(3, 110)
	//
	return img.Save("", !t.base.IsSunMi, false)
}

// OneDishOneOrder 一菜一单模版
func (t *dishesImgTemplate) OneDishOneOrder(
	tmp int,
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
	// 是否打印
	isPrinter := false
	// 创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	//
	if tmp == 2 {
		img.SetAlignment(pkg.AlignCenter)
		img.SetImagePadding(0) // 确保没有填充
		img.SetFontWeight(5)
		img.SetFontSize(30)

		// 桌号
		if order.DeskUuid > 0 {
			// 判断文字是否包含缅甸语
			spacing := 50
			if t.base.IsMyText(order.SerialNo) {
				spacing = 68
			}
			img.SetTextLineHeight(spacing)
			img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("桌号"), order.SerialNo, mealNumStr))
			img.SetTextLineHeight(45)
			img.LineFeed(1, 20)
		} else {
			img.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo)
		}

		// 设置边距
		img.SetFontSize(20)
		img.SetFontWeight(1)
		img.SetTextLineHeight(36)
		img.LineFeed(1, 40)
		img.LineFeed(2)
		img.SetTextLineHeight(50)

		// 商品和数量
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
				if t.base.Lang == "my" {
					img.SetTextLineHeight(90)
				} else {
					img.SetTextLineHeight(64)
				}
				img.LineFeed(1, 12)
				img.PrintInColumns(
					pkg.ColumnConfig{Text: productName, Width: 500, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
					pkg.ColumnConfig{Text: "x" + fmt.Sprintf("%d", num), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 50},
				)
				if t.base.Lang == "my" {
					img.LineFeed(1, 12)
				}

				// 分割处理属性
				for _, attr := range product.ProductAttrList {
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

				if product.Remark != "" {
					if t.base.IsMyText(product.Remark) {
						img.SetTextLineHeight(68)
					} else {
						img.SetTextLineHeight(50)
					}
					img.LineFeed(1, 12)
					img.SetFontSize(28)
					img.AppendText(product.Remark)
					img.LineFeed(1, 50)
					img.SetFontSize(20)
				}

				img.LineFeed(1, 12)
				img.SetTextLineHeight(50)
			}

			// 根据打印选择执行打印
			if productPrinter.PrintModeScene == 1 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1)
				}
			} else {
				exportation(int(product.TotalNum))
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
		if order.DeskUuid > 0 {
			spacing := 50
			if t.base.IsMyText(order.SerialNo) {
				spacing = 68
			}
			img.SetTextLineHeight(spacing)
			img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("桌号"), order.SerialNo, mealNumStr))
			img.SetTextLineHeight(45)
		} else {
			img.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo)
		}

		img.LineFeed(1, 60)
		img.SetFontSize(20)
		img.AppendText(updateTime)
		img.LineFeed(1)
		img.LineFeed(1, 24)

		// 商品和数量
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
				if t.base.Lang == "my" {
					img.SetTextLineHeight(90)
				} else {
					img.SetTextLineHeight(64)
				}
				img.LineFeed(1, 12)
				img.PrintInColumns(
					pkg.ColumnConfig{Text: productName, Width: 500, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
					pkg.ColumnConfig{Text: "x" + fmt.Sprintf("%d", num), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 50},
				)
				if t.base.Lang == "my" {
					img.LineFeed(1, 12)
				}

				// 分割处理属性
				for _, attr := range product.ProductAttrList {
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

				if product.Remark != "" {
					if t.base.IsMyText(product.Remark) {
						img.SetTextLineHeight(68)
					} else {
						img.SetTextLineHeight(50)
					}
					img.LineFeed(1, 12)
					img.SetFontSize(28)
					img.AppendText(product.Remark)
					img.LineFeed(1, 50)
					img.SetFontSize(20)
				}

				img.LineFeed(1, 12)
				img.SetTextLineHeight(50)
			}

			// 根据打印选择执行打印
			if productPrinter.PrintModeScene == 1 {
				for i := 0; i < int(product.TotalNum); i++ {
					exportation(1)
				}
			} else {
				exportation(int(product.TotalNum))
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
	return img.Save("", !t.base.IsSunMi, false)
}

// ReturnMenuTemplate 退菜单模版
func (t *dishesImgTemplate) ReturnMenuTemplate(
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
	// 是否有打印内容
	isPrinter := false

	// 创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetAlignment(pkg.AlignCenter)
	img.SetImagePadding(0) // 确保没有填充
	img.SetFontWeight(5)
	img.SetFontSize(30)
	img.AppendText(t.base.Translate("退菜单"))
	img.LineFeed(1, 68)

	// 桌号
	if order.DeskUuid > 0 {
		// 判断文字是否包含缅甸语
		spacing := 50
		if t.base.IsMyText(order.SerialNo) {
			spacing = 68
		}
		img.SetTextLineHeight(spacing)
		img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("桌号"), order.SerialNo, mealNumStr))
		img.SetTextLineHeight(45)
		img.LineFeed(1)
	} else {
		img.AppendText(t.base.Translate("取单号") + ": " + order.SerialNo + "\n")
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
	for _, product := range products {
		// 处理自助餐文本
		buffetText := ""
		if buffetSignOpen == "1" {
			if product.IsBuffet {
				buffetText = t.base.Translate("自助餐") + "-"
			}
		}
		// 产品名称
		productName := "(" + t.base.Translate("退") + ") " + buffetText + product.ProductName.GetLocale(t.base.Lang)
		if t.base.Lang == "my" {
			img.SetTextLineHeight(90)
		} else {
			img.SetTextLineHeight(64)
		}
		img.LineFeed(1, 12)

		// 打印产品名称和数量
		totalNum := "x" + fmt.Sprintf("%d", product.TotalNum)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: productName, Width: 500, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 30},
			pkg.ColumnConfig{Text: totalNum, Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 30, LineHeight: 50},
		)
		if t.base.Lang == "my" {
			img.LineFeed(1, 12)
		}

		// 分割处理属性
		for _, attr := range product.ProductAttrList {
			if t.base.Lang == "my" {
				img.SetTextLineHeight(50)
			} else {
				img.SetTextLineHeight(40)
			}
			img.AppendText(attr.GetLocale(t.base.Lang))
			img.LineFeed(1, 50)
		}

		if product.Remark != "" {
			if t.base.IsMyText(product.Remark) {
				img.SetTextLineHeight(50)
			} else {
				img.SetTextLineHeight(40)
			}
			img.AppendText(product.Remark)
			img.LineFeed(1, 50)
		}

		img.LineFeed(1, 12)
		img.SetTextLineHeight(50)

		// 标记已打印
		isPrinter = true

		// 退菜原因
		img.AppendSplitLine()
		img.LineFeed(1, 34)
		// 获取退菜原因文本
		reasonText := product.Reason.GetLocale(t.base.Lang)
		// 如果有自定义原因，则添加
		if product.CustomReason != "" {
			reasonText += "、" + product.CustomReason
		}
		img.AppendText(fmt.Sprintf(
			"%s： %s",
			t.base.Translate("退菜原因"),
			reasonText,
		))
	}

	if !isPrinter {
		return ""
	}

	// 设置行间距
	img.SetTextLineHeight(50)
	// 换行
	img.LineFeed(4)

	//
	return img.Save("", !t.base.IsSunMi, false)
}
