// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/app/printer/pkg/images"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"
)

// invoiceImgTemplate 图片订单打印模板
type invoiceImgTemplate struct {
	base *printerTemplate
}

// NewInvoiceImgTemplate 创建新的图片订单打印模板
func NewInvoiceImgTemplate(
	base *printerTemplate,
) *invoiceImgTemplate {
	return &invoiceImgTemplate{
		base: base,
	}
}

// ImgPrint 图片打印
func (t *invoiceImgTemplate) GetPrintContent(
	settingPrinterInfo respSetting.PrinterInfo,
	tmpInfo model.PrinterTemplate,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	temp := tmpInfo.Template
	isShowSku := tmpInfo.IsShowSku

	/* *
	 * 模版2
	 */
	if temp == 2 {
		return NewStatementOrderImgTemplate(t.base).GetPrintContent(
			settingPrinterInfo,
			constant.PrinterTemplateInvoice,
			utils.IfInt(isShowSku == 0, 4, 3),
			saleBill,
			saleOrder,
			0,
		)
	}

	// 店铺设置
	company := t.base.StoreSetting.Company
	address := t.base.StoreSetting.Address
	phone := t.base.StoreSetting.Phone
	taxNumber := t.base.StoreSetting.TaxNumber
	// 品牌
	brandName := config.Server.BrandName
	// 日历
	payTime := t.base.FormatUnixTimeDefault(saleBill.FinishTime)

	// 合计应收
	finalPrice := saleOrder.GetPrintReceivablePrice()

	//  创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetTextLineHeight(50)
	img.SetImagePadding(0)
	img.SetAlignment(pkg.AlignCenter)
	img.SetFontWeight(2)
	img.SetFontSize(24)
	img.AppendText(t.base.StoreSetting.Name)
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.LineFeed(1)
	img.LineFeed(1, 10)
	img.AppendText(t.base.Translate("非常感谢您今天的到来，我们期待您的再次光临"))
	img.LineFeed(1, 60)
	img.SetFontSize(28)
	img.AppendText(t.base.Translate("发票"))
	img.SetFontSize(20)
	img.LineFeed(1)
	if payTime != "" {
		img.AppendText(payTime)
		img.LineFeed(1)
	}
	if t.base.Lang == "ja" {
		img.SetAlignment(pkg.AlignRight)
		img.AppendText(t.base.Translate("先生/小姐"))
		img.LineFeed(1, 20)
		img.SetAlignment(pkg.AlignLeft)
	}
	img.AppendSplitLine()
	img.LineFeed(1)
	// 合计应收
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("合计"), Width: 260, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 28},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(finalPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 28},
	)
	img.LineFeed(1, 12)
	// 服务费
	img.PrintInColumns(
		pkg.ColumnConfig{Text: "(" + t.base.Translate("其中服务费"), Width: 350, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(saleOrder.ServiceFee) + ")", Width: 0, Align: pkg.AlignRight},
	)
	// 消费税
	if t.base.ConsumptionTax != 4 {
		img.PrintInColumns(
			pkg.ColumnConfig{Text: "(" + t.base.Translate("其中VAT"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(saleOrder.TaxFee) + ")", Width: 0, Align: pkg.AlignRight},
		)
	}

	// 分割线
	img.AppendSplitLine()

	// 未含税
	if t.base.ConsumptionTax != 4 {
		img.AppendText(t.base.Translate("仅作为餐饮费收取以上金额"))
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignRight)
		img.AppendText(t.base.Translate("合计 (其中VAT)"))
		img.LineFeed(1)
		for _, percentage := range saleOrder.GetPercentageList() {
			taxRate := percentage["TaxRate"]
			taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
			totalPrice, _ := strconv.ParseFloat(percentage["TotalPrice"], 64)
			if t.base.Lang == "ja" {
				img.PrintInColumns(
					pkg.ColumnConfig{
						Text:  fmt.Sprintf("%s%s%s", taxRate, "%", t.base.Translate("的对象")),
						Width: 350,
						Align: pkg.AlignLeft,
					},
					pkg.ColumnConfig{
						Text:  fmt.Sprintf("%s (%s)", t.base.Amount(totalPrice), t.base.Amount(taxFee)),
						Width: 0,
						Align: pkg.AlignRight,
					},
				)
			} else {
				img.PrintInColumns(
					pkg.ColumnConfig{
						Text:  fmt.Sprintf("VAT (%s)", taxRate+"%"),
						Width: 350,
						Align: pkg.AlignLeft,
					},
					pkg.ColumnConfig{
						Text:  fmt.Sprintf("%s (%s)", t.base.Amount(totalPrice), t.base.Amount(taxFee)),
						Width: 0,
						Align: pkg.AlignRight,
					},
				)
			}
		}
	}

	// 不包含退款金额
	img.SetAlignment(pkg.AlignLeft)
	if returnAmount := saleOrder.GetReturnAmount(); returnAmount > 0 {
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("不包含退款金额"), t.base.GetPriceAndUnit(returnAmount)))
		img.LineFeed(1, 40)
	}

	// 支付方式
	img.AppendSplitLine()
	img.LineFeed(1, 40)
	img.LineFeed(1, 5)
	if saleOrder.IsFreeSaleOrder() {
		img.SetTextLineHeight(40)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("免单"), Width: 320, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 20},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(0), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 20},
		)
	} else {
		for key, paymentOrder := range saleOrder.PaymentOrders {
			img.SetTextLineHeight(40)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: paymentOrder.PaymentMethod.GetName(), Width: 320, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 20},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(paymentOrder.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 20},
			)
			if key < len(saleOrder.PaymentOrders)-1 {
				img.LineFeed(1, 10)
			}
			img.SetTextLineHeight(50)
		}
	}

	// 发票信息
	img.SetAlignment(pkg.AlignLeft)
	invoiceInfo := saleOrder.InvoiceInfo
	if invoiceInfo != nil && invoiceInfo.HasContent() {
		img.AppendSplitLine()
		img.AppendText(t.base.Translate("发票信息"))
		img.LineFeed(1)
		if invoiceInfo.CompanyName != "" {
			img.AppendText(t.base.Translate("公司名称") + ": " + invoiceInfo.CompanyName)
			lineFeedHeight := 40
			if invoiceInfo.CompanyAddr != "" || invoiceInfo.CompanyTaxNumber != "" || invoiceInfo.CompanyPhone != "" {
				lineFeedHeight = 50
			}
			img.LineFeed(1, lineFeedHeight)
		}
		if invoiceInfo.CompanyAddr != "" {
			img.AppendText(t.base.Translate("公司地址") + ": " + invoiceInfo.CompanyAddr)
			lineFeedHeight := 40
			if invoiceInfo.CompanyTaxNumber != "" || invoiceInfo.CompanyPhone != "" {
				lineFeedHeight = 50
			}
			img.LineFeed(1, lineFeedHeight)
		}
		if invoiceInfo.CompanyTaxNumber != "" {
			img.AppendText(t.base.Translate("税号") + ": " + invoiceInfo.CompanyTaxNumber)
			lineFeedHeight := 40
			if invoiceInfo.CompanyPhone != "" {
				lineFeedHeight = 50
			}
			img.LineFeed(1, lineFeedHeight)
		}
		if invoiceInfo.CompanyPhone != "" {
			img.AppendText(t.base.Translate("联系电话") + ": " + invoiceInfo.CompanyPhone)
			img.LineFeed(1, 40)
		}
	}

	// 信息
	img.SetTextLineHeight(50)
	img.AppendSplitLine()
	img.LineFeed(1, 40)
	img.AppendText(t.base.Translate("收银员") + ": " + saleOrder.CashierName)
	img.LineFeed(1)
	img.AppendText(t.base.Translate("订单号") + ": " + saleOrder.OrderNo)
	img.LineFeed(1)
	img.AppendText(t.base.Translate("打印次数") + ": " + strconv.Itoa(invoiceInfo.PrintNum))
	img.LineFeed(1)

	// 公司名称
	if company != "" {
		img.SetTextLineHeight(40)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("公司名称"), company))
		img.RecoverDefaultTextLineHeight()
		img.LineFeed(1)
	}
	if address != "" {
		img.SetTextLineHeight(40)
		if t.base.Lang == "th" {
			img.AppendText(fmt.Sprintf("%s : %s", t.base.Translate("地址"), address))
		} else {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("地址"), address))
		}
		img.RecoverDefaultTextLineHeight()
		img.LineFeed(1)
	}
	if phone != "" {
		img.SetTextLineHeight(40)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("电话"), phone))
		img.RecoverDefaultTextLineHeight()
		img.LineFeed(1)
	}
	if taxNumber != "" {
		img.SetTextLineHeight(40)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("税号"), taxNumber))
		img.RecoverDefaultTextLineHeight()
		img.LineFeed(1)
	}
	img.LineFeed(1, 40)
	// 担当者：
	if t.base.Lang == "ja" {
		// img.SetTextLineHeight(50)
		// img.SetAlignment(pkg.AlignLeft)
		// img.AppendText("                                   担当者")
		img.SetTextLineHeight(10)
		img.SetAlignment(pkg.AlignRight)
		img.AppendEmbeddedImg(images.PersonInChargeImg, 317, false, -40) // -53
		img.AppendSplitLine()
		img.RecoverDefaultTextLineHeight()
		img.LineFeed(1)
	}
	//
	img.SetAlignment(pkg.AlignLeft)
	img.AppendText("*" + t.base.Translate("保管注意事项"))
	img.LineFeed(1)
	img.AppendText(t.base.Translate("如需保管时请将印刷页面朝内折叠"))

	// 技术支持方
	img.LineFeed(1)
	img.AppendSplitLine()
	img.LineFeed(1)
	img.SetAlignment(pkg.AlignCenter)
	if t.base.Lang == "tr" {
		img.AppendText("Ziyaretiniz için teşekkür ederiz! Bu mağaza")
		img.LineFeed(1)
		img.AppendText("tarafından: " + brandName + " Sistem destek sağlar.")
	} else if t.base.Lang == "th" {
		img.AppendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " + brandName)
	} else {
		img.AppendText(t.base.Translate("感谢您的光临！本店由") + " " + brandName + " " + t.base.Translate("系统提供支持。"))
	}
	//
	img.LineFeed(4)
	//
	img.SetSegmentationHeight(utils.IfInt(settingPrinterInfo.IsCashierPrinter, 2200, 200))
	//
	return img.Save("", !t.base.IsSunMi, 0)
}
