// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"ttpos-server-go/app/constant"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
)

// statementOrderImgTemplate 图片订单打印模板
type statementOrderImgTemplate struct {
	base *printerTemplate
}

// NewStatementOrderImgTemplate 创建新的图片订单打印模板
func NewStatementOrderImgTemplate(
	ctx context.Context,
	setting *setting.Srv,
	storeSetting *respSetting.Store,
	printerSetting *respSetting.Printer,
	currencySetting *respSetting.Currency,
) *statementOrderImgTemplate {
	return &statementOrderImgTemplate{
		base: NewPrinterTemplate(ctx, setting, storeSetting, printerSetting, currencySetting, false),
	}
}

// ImgPrint 图片打印
func (t *statementOrderImgTemplate) ImgPrint(
	printType int,
	temp int,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	name := t.base.Translate("人")
	// 店铺设置
	company := t.base.StoreSetting.Company
	address := t.base.StoreSetting.Address
	phone := t.base.StoreSetting.Phone
	taxNumber := t.base.StoreSetting.TaxNumber
	chainNumber := t.base.StoreSetting.ChainNumber
	// 品牌
	// brandName := config.Server.BrandName
	// 日历
	payTime := t.base.FormatUnixTimeDefault(saleBill.FinishTime)

	// 就餐人数
	mealNumStr := ""
	if saleBill.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", saleBill.MealNum, name)
	}

	// 订单名称
	orderName := fmt.Sprintf("%d", saleOrder.Index)
	if orderName != "" {
		orderName = "-" + orderName
	}

	//  创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetTextLineHeight(50)
	img.SetImagePadding(0)
	if t.base.Lang == "my" {
		img.SetImagePadding(3)
	}
	if temp != 3 {
		img.SetAlignment(pkg.AlignLeft)
		if printType == constant.PrinterTemplateInvoice {
			img.AppendText(t.base.Translate("发票"))
		} else if printType == constant.PrinterTemplatePreBilling {
			img.AppendText(t.base.Translate("预结账单"))
		} else {
			img.AppendText(t.base.Translate("结账单"))
		}
		img.LineFeed(1, 60)
	}
	img.SetAlignment(pkg.AlignCenter)
	img.SetFontSize(26)
	img.SetFontWeight(2)
	img.AppendText(t.base.StoreSetting.Name)
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.LineFeed(1, 60)
	//
	if temp == 1 {
		img.LineFeed(1, 10)
		img.RecoverDefaultTextLineHeight()
		img.SetAlignment(pkg.AlignLeft)
		if saleBill.DeskUuid > 0 {
			img.SetFontWeight(2)
			img.SetFontSize(28)
			spacing := 50
			if t.base.IsMyText(saleBill.SerialNo) {
				spacing = 68
			}
			img.SetTextLineHeight(spacing)
			img.AppendText(fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			img.RecoverDefaultTextLineHeight()
			img.SetFontSize(20)
			img.LineFeed(1)
		} else if saleBill.SerialNo != "" {
			img.SetFontWeight(2)
			img.SetFontSize(28)
			img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName))
			img.SetFontSize(20)
			img.LineFeed(1)
		}
		//
		img.LineFeed(1, 12)
		img.SetFontWeight(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单号"), Width: 300, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: saleOrder.OrderNo, Width: 0, Align: pkg.AlignRight},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("收银员"), Width: 300, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: "saleOrder.CashierUuid", Width: 0, Align: pkg.AlignRight},
		)
		if saleOrder.FinishTime > 0 {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("时间"), Width: 300, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: t.base.FormatUnixTimeDefault(saleOrder.FinishTime), Width: 0, Align: pkg.AlignRight},
			)
		}
		img.LineFeed(1)
	} else if temp == 2 {
		img.AppendText(t.base.Translate("非常感谢您今天的到来，我们期待您的再次光临"))
		img.LineFeed(1, 60)
		//
		if saleOrder.FinishTime > 0 {
			img.AppendText(t.base.FormatUnixTimeDefault(saleOrder.FinishTime))
			img.LineFeed(1)
		}
		//
		img.SetFontWeight(2)
		img.SetFontSize(28)
		if saleBill.DeskUuid > 0 {
			spacing := 50
			if t.base.IsMy(saleBill.SerialNo) {
				spacing = 68
			}
			img.SetTextLineHeight(spacing)
			img.AppendText(fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			img.RecoverDefaultTextLineHeight()
		} else {
			img.SetFontWeight(2)
			img.SetFontSize(28)
			img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName))
			img.SetFontSize(20)
			img.LineFeed(1)
		}
		//
		img.SetFontSize(20)
		img.RecoverDefaultTextLineHeight()
		img.SetFontWeight(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单号"), Width: 300, Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: saleOrder.OrderNo, Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		// if saleOrder.CashierName != "" {
		// 	img.PrintInColumns(
		// 		pkg.ColumnConfig{Text: t.base.Translate("收银员"), Width: 300, Align: pkg.AlignLeft, FontWeight: 2},
		// 		pkg.ColumnConfig{Text: saleOrder.CashierName, Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		// 	)
		// }
		img.LineFeed(1)
	} else if temp == 3 {
		// 打印logo
		// if t.base.setting != nil {
		// 	storeSetting := t.base.StoreSetting
		// 	img.SetTextLineHeight(25)
		// 	img.SetAlignment(ImgFont.AlignCenter)
		// 	// whiteBackgroundWithBlackTextLogoPath := Supplier.GetWhiteBackgroundWithBlackTextLogoPath(saleOrder.AppId, "http://nginx"+ImgHelp.RemoveImageDomain(storeSetting["logoUrl"]))
		// 	img.AppendImg(whiteBackgroundWithBlackTextLogoPath, 150, false, -25)
		// 	img.LineFeed(1)
		// }
		//
		img.LineFeed(1, 10)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.SetFontSize(26)
		img.SetTextLineHeight(50)
		if printType == constant.PrinterTemplateInvoice {
			img.AppendText(t.base.Translate("发票"))
		} else if printType == constant.PrinterTemplatePreBilling {
			img.AppendText(t.base.Translate("预结账单"))
		} else {
			img.AppendText(t.base.Translate("结账单"))
		}
		img.RecoverDefaultTextLineHeight()
		img.SetFontSize(20)
		img.SetFontWeight(1)
		img.LineFeed(1)
		// 公司名称
		if company != "" {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("公司名称"), company))
			img.LineFeed(1)
		}
		// 连锁店编号
		if chainNumber != "" {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("连锁店编号"), chainNumber))
			img.LineFeed(1)
		}
		if address != "" {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("商家地址"), address))
			img.LineFeed(1)
		}
		if phone != "" {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("电话"), phone))
			img.LineFeed(1)
		}
		if taxNumber != "" {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("税号"), taxNumber))
			img.LineFeed(1)
		}
		// 发票信息
		// if saleOrder.Template == 3 && saleOrder.InvoiceInfo && (saleOrder.InvoiceInfo.CompanyName || saleOrder.InvoiceInfo.CompanyAddr || saleOrder.InvoiceInfo.CompanyTaxNumber || saleOrder.InvoiceInfo.CompanyPhone) {
		// 	img.AppendSplitLine(true, 40)
		// 	if saleOrder.InvoiceInfo.CompanyName != "" {
		// 		img.AppendText(saleOrder.InvoiceInfo.CompanyName)
		// 		img.LineFeed(1, (saleOrder.InvoiceInfo.CompanyAddr || saleOrder.InvoiceInfo.CompanyTaxNumber || saleOrder.InvoiceInfo.CompanyPhone) ? 50 : 40)
		// 	}
		// 	if saleOrder.InvoiceInfo.CompanyAddr != "" {
		// 		img.AppendText(saleOrder.InvoiceInfo.CompanyAddr)
		// 		img.LineFeed(1, (saleOrder.InvoiceInfo.CompanyTaxNumber || saleOrder.InvoiceInfo.CompanyPhone) ? 50 : 40)
		// 	}
		// 	if saleOrder.InvoiceInfo.CompanyTaxNumber != "" {
		// 		img.AppendText(saleOrder.InvoiceInfo.CompanyTaxNumber)
		// 		img.LineFeed(1, (saleOrder.InvoiceInfo.CompanyPhone) ? 50 : 40)
		// 	}
		// 	if saleOrder.InvoiceInfo.CompanyPhone != "" {
		// 		img.AppendText(saleOrder.InvoiceInfo.CompanyPhone)
		// 		img.LineFeed(1, 40)
		// 	}
		// }
		//
		img.AppendSplitLine()
		img.RecoverDefaultTextLineHeight()
		img.SetAlignment(pkg.AlignLeft)
		if saleBill.DeskUuid > 0 {
			img.SetFontWeight(2)
			img.SetFontSize(28)
			if t.base.IsMy(saleBill.SerialNo) {
				img.SetTextLineHeight(68)
			} else {
				img.SetTextLineHeight(50)
			}
			img.AppendText(fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			img.RecoverDefaultTextLineHeight()
			img.LineFeed(1)
		} else {
			img.SetFontWeight(2)
			img.SetFontSize(28)
			img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName))
			img.RecoverDefaultTextLineHeight()
			img.LineFeed(1)
		}
		img.SetFontWeight(1)
		img.SetFontSize(20)
		img.SetAlignment(pkg.AlignLeft)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("收银员"), "saleOrder.CashierName"))
		img.LineFeed(1)
		if payTime != "" {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("时间"), payTime))
			img.LineFeed(1)
		}
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("订单号"), saleOrder.OrderNo))
		img.LineFeed(1)
	}
	// 计算宽度
	var productWidth, priceQtyWidth int
	if t.base.Lang == "en" || t.base.Lang == "th" || t.base.Lang == "tr" || t.base.Lang == "my" {
		productWidth = 220
		priceQtyWidth = 230
	} else {
		productWidth = 120
		priceQtyWidth = 120
	}

	if temp != 3 {
		img.SetTextLineHeight(30)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("商品"), Width: productWidth, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.Translate("单价") + "|" + t.base.Translate("数量"), Width: priceQtyWidth, Align: pkg.AlignCenter},
			pkg.ColumnConfig{Text: t.base.Translate("小计"), Width: 0, Align: pkg.AlignRight},
		)
	}
	img.AppendSplitLine()
	img.LineFeed(1, 40)
	img.SetTextLineHeight(50)
	// 赠品金额
	freeMoney := float64(0)
	// 商品列表
	productNum := uint(0)
	// 商品列表
	for key, item := range saleOrder.SaleOrderProducts {
		if item.IsUnCookingProduct() {
			continue
		}
		productNum += item.Num
		productTotalPrice := item.TotalPrice
		// 赠品
		var gift string
		if item.IsGiftBool() {
			gift = "(" + t.base.Translate("赠") + ") "
			freeMoney += item.TotalPrice
			productTotalPrice = 0
		}
		productName := gift + item.MultiLanguageName.GetNameByLang(t.base.Lang) + "\n" + item.GetAttributeNamesByLang(t.base.Lang)
		//
		img.SetTextLineHeight(45)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: productName, Width: productWidth, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s*%d", t.base.Amount(item.Price), item.Num), Width: priceQtyWidth, Align: pkg.AlignCenter},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(productTotalPrice), Width: 0, Align: pkg.AlignRight},
		)
		img.RecoverDefaultTextLineHeight()
		if key != len(saleOrder.SaleOrderProducts)-1 {
			img.LineFeed(1, 10)
		}
	}
	//

	//
	return img.Save("", !t.base.IsSunMi, false)
}
