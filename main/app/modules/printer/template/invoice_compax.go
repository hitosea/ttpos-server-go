// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"
)

// invoiceCompaxTemplate Compax发票打印模板
type invoiceCompaxTemplate struct {
	base *printerTemplate
}

// NewInvoiceCompaxTemplate 创建新的Compax发票打印模板
func NewInvoiceCompaxTemplate(
	base *printerTemplate,
) *invoiceCompaxTemplate {
	return &invoiceCompaxTemplate{
		base: base,
	}
}

// invoiceCompaxTemplate 获取打印内容
func (t *invoiceCompaxTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	tmpInfo model.PrinterTemplate,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
	isCashierPrinter bool,
) string {
	temp := tmpInfo.Template
	isShowSku := tmpInfo.IsShowSku

	/* *
	 * 模版2
	 */
	if temp == 2 {
		return NewStatementOrderCompaxTemplate(t.base).GetPrintContent(
			settingPrinterInfo,
			printerConst.PrinterTemplateInvoice,
			utils.IfInt(isShowSku == 0, 4, 3),
			saleBill,
			saleOrder,
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

	// 宽度
	width := 48
	leftWidth := 28
	lineSpacing := "\x1B\x33\x32"

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	// 重置行间距
	printer.RestoreDefaultLineSpacing()
	// 货币宽度
	currencyWidth := width
	if t.base.Lang == "th" {
		currencyWidth = currencyWidth - 1
	} else if t.base.CurrencyUnit == "฿" || t.base.CurrencyUnit == "¥" {
		currencyWidth = currencyWidth - 1
	}
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetCharacterSize(2, 1)
	printer.AppendText(t.base.StoreSetting.Name)
	printer.SetCharacterSize(1, 1)
	if isCashierPrinter {
		printer.SetLineSpacing(60)
	} else {
		printer.SetLineSpacing(80)
	}
	printer.LineFeed(1)
	printer.AppendText("\x1B\x33\x28")
	printer.AppendText(t.base.Translate("非常感谢您今天的到来，我们期待您的再次光临"))
	printer.AppendText(lineSpacing)
	printer.LineFeed(1)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.Translate("发票"))
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(1)
	printer.AppendText(payTime)
	printer.SetLineSpacing(60)
	printer.LineFeed(1)
	//
	if t.base.Lang == "ja" {
		printer.SetLineSpacing(34)
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("先生/小姐"))
		printer.LineFeed()
	}
	//
	printer.SetLineSpacing(30)
	printer.AppendText("------------------------------------------------\n")
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.PrintText(t.base.Translate("合计"), "", t.base.GetPriceAndUnit(finalPrice), currencyWidth-24))
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(2)
	// 服务费
	printer.AppendText(t.base.PrintText("("+t.base.Translate("其中服务费"), "", t.base.GetPriceAndUnit(saleOrder.ServiceFee)+")", currencyWidth, leftWidth))
	printer.LineFeed(2)
	// 消费税
	if t.base.ConsumptionTax != 4 {
		printer.AppendText(t.base.PrintText("("+t.base.Translate("其中VAT"), "", t.base.GetPriceAndUnit(saleOrder.TaxFee)+")", currencyWidth, leftWidth))
		printer.LineFeed(2)
	}
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignLeft)

	// 未含税
	if t.base.ConsumptionTax != 4 {
		printer.SetLineSpacing(12)
		printer.LineFeed()
		printer.AppendText(lineSpacing)
		printer.AppendText(t.base.Translate("仅作为餐饮费收取以上金额"))
		printer.LineFeed()
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("合计 (其中VAT)"))
		printer.LineFeed(1)
		printer.AppendText(lineSpacing)
		printer.SetAlignment(pkg.AlignLeft)
		for _, percentage := range saleOrder.GetPercentageList() {
			taxRate := percentage["TaxRate"]
			taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
			totalPrice, _ := strconv.ParseFloat(percentage["TotalPrice"], 64)
			if t.base.Lang == "ja" {
				printer.AppendText(t.base.PrintText(taxRate+"%"+t.base.Translate("的对象"), "", t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")", currencyWidth, leftWidth))
			} else {
				printer.AppendText(t.base.PrintText("VAT ("+taxRate+"%)", "", t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")", currencyWidth, leftWidth))
			}
		}
	}

	// 不包含退款金额
	if returnAmount := saleOrder.GetReturnAmount(); returnAmount > 0 {
		printer.SetLineSpacing(12)
		printer.LineFeed()
		printer.AppendText(lineSpacing)
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("不包含退款金额"), t.base.GetPriceAndUnit(returnAmount)))
		printer.LineFeed(1, 40)
	}

	// 支付方式
	printer.SetLineSpacing(40)
	printer.SetPrintModes(true, false, false)
	printer.AppendText("------------------------------------------------")
	printer.SetLineSpacing(40)
	printer.LineFeed()
	printer.SetPrintModes(true, false, false)
	if saleOrder.IsFreeSaleOrder() {
		printer.SetLineSpacing(50)
		printer.AppendText(t.base.PrintText(t.base.Translate("免单"), "", t.base.GetPriceAndUnit(0), currencyWidth))
	} else {
		for key, paymentOrder := range saleOrder.PaymentOrders {
			printer.SetLineSpacing(50)
			printer.AppendText(t.base.PrintText(paymentOrder.PaymentMethod.GetName(), "", t.base.GetPriceAndUnit(paymentOrder.Amount), currencyWidth))
			if key < len(saleOrder.PaymentOrders)-1 {
				printer.LineFeed(1)
			}
		}
	}
	//
	printer.SetPrintModes(false, false, false)
	printer.SetLineSpacing(40)
	printer.LineFeed()
	printer.SetLineSpacing(10)

	// 发票信息
	invoiceInfo := saleOrder.InvoiceInfo
	if invoiceInfo != nil && invoiceInfo.HasContent() {
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		printer.SetLineSpacing(40)
		printer.LineFeed()
		printer.AppendText(lineSpacing)
		printer.AppendText(t.base.Translate("发票信息"))
		printer.LineFeed(1)
		if invoiceInfo.CompanyName != "" {
			printer.AppendText(t.base.Translate("公司名称") + ": " + invoiceInfo.CompanyName)
			printer.LineFeed(1)
		}
		if invoiceInfo.CompanyAddr != "" {
			printer.AppendText(t.base.Translate("公司地址") + ": " + invoiceInfo.CompanyAddr)
			printer.LineFeed(1)
		}
		if invoiceInfo.CompanyTaxNumber != "" {
			printer.AppendText(t.base.Translate("税号") + ": " + invoiceInfo.CompanyTaxNumber)
			printer.LineFeed(1)
		}
		if invoiceInfo.CompanyPhone != "" {
			printer.AppendText(t.base.Translate("联系电话") + ": " + invoiceInfo.CompanyPhone)
			printer.LineFeed(1)
		}
	}

	//
	printer.AppendText("------------------------------------------------\n")
	printer.SetLineSpacing(40)
	printer.AppendText(lineSpacing)
	printer.AppendText(t.base.Translate("收银员") + ": " + saleOrder.CashierName)
	printer.LineFeed()
	printer.AppendText(t.base.Translate("订单号") + ": " + saleOrder.OrderNo)
	printer.LineFeed()
	printer.AppendText(t.base.Translate("打印次数") + ": " + strconv.Itoa(invoiceInfo.PrintNum))
	printer.LineFeed()
	if company != "" {
		printer.AppendText("\x1B\x33\x28")
		printer.AppendText(t.base.Translate("公司名称") + ": " + company)
		printer.AppendText(lineSpacing)
		printer.LineFeed()
	}
	if address != "" {
		printer.AppendText("\x1B\x33\x28")
		printer.AppendText(t.base.Translate("地址") + ": " + address)
		printer.AppendText(lineSpacing)
		printer.LineFeed()
	}
	if taxNumber != "" {
		printer.AppendText(t.base.Translate("税号") + ": " + taxNumber)
		printer.LineFeed()
	}
	if phone != "" {
		printer.AppendText(t.base.Translate("电话") + ": " + phone)
		printer.LineFeed()
	}

	// 保管注意事项
	printer.AppendText("\x1B\x33\x28")
	printer.LineFeed()
	printer.AppendText(lineSpacing)
	printer.AppendText("*" + t.base.Translate("保管注意事项"))
	printer.LineFeed()
	printer.AppendText(t.base.Translate("如需保管时请将印刷页面朝内折叠"))

	// 技术支持方
	printer.AppendText("\n------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignCenter)
	if t.base.Lang == "th" {
		printer.AppendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " + brandName)
	} else {
		printer.AppendText(t.base.Translate("感谢您的光临！本店由") + " " + brandName + " " + t.base.Translate("系统提供支持。"))
	}

	// Print and exit page mode
	printer.PrintAndExitPageMode()
	printer.LineFeed(4)
	printer.CutPaper(settingPrinterInfo.IsEnableSound())

	// 返回打印数据
	return printer.GetOrderData()
}
