// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"
)

// invoiceXprinterTemplate xprinter发票打印模板
type invoiceXprinterTemplate struct {
	base *printerTemplate
}

// NewInvoiceXprinterTemplate 创建新的xprinter发票打印模板
func NewInvoiceXprinterTemplate(
	base *printerTemplate,
) *invoiceXprinterTemplate {
	return &invoiceXprinterTemplate{
		base: base,
	}
}

// invoiceXprinterTemplate 获取打印内容
func (t *invoiceXprinterTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	tmpInfo model.PrinterTemplate,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
	isCashierPrinter bool,
) string {
	printerType := settingPrinterInfo.PrinterType
	temp := tmpInfo.Template
	isShowSku := tmpInfo.IsShowSku

	/* *
	 * 模版2
	 */
	if temp == 2 {
		return NewStatementOrderXprinterTemplate(t.base).GetPrintContent(
			settingPrinterInfo,
			constant.PrinterTemplateInvoice,
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
	isXinyeWifi := printerType == constant.PrinterTypeXPrinterWifi || printerType == constant.PrinterTypeXPrinterLan
	lineSpacing := "\x1B\x33\x32"
	if isXinyeWifi {
		lineSpacing = "\x1B\x33\x38"
	}

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.RestoreDefaultLineSpacing()
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
	if isXinyeWifi {
		printer.LineFeed(1)
	}
	printer.AppendText("\x1B\x33\x48")
	printer.AppendText(t.base.Translate("非常感谢您今天的到来，我们期待您的再次光临"))
	printer.AppendText(lineSpacing)
	printer.LineFeed(2)
	printer.SetCharacterSize(2, 2)
	invoiceTitle := t.base.Translate("发票")
	if saleBill.IsOrderSourceTakeout() {
		invoiceTitle += "(" + t.base.Translate("外卖") + ")"
	}
	printer.AppendText(invoiceTitle)
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(1)
	printer.AppendText(lineSpacing)
	if isXinyeWifi {
		printer.AppendText("\x1B\x33\x38")
	} else {
		printer.AppendText("\x1B\x33\x18")
	}
	printer.LineFeed(1)
	printer.AppendText(lineSpacing)
	printer.AppendText(payTime)
	printer.SetLineSpacing(60)
	printer.LineFeed(2)
	//
	if t.base.Lang == "ja" {
		printer.SetLineSpacing(52)
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("先生/小姐"))
		printer.LineFeed()
	}
	//
	printer.AppendText("------------------------------------------------\n")
	printer.SetLineSpacing(18)
	if isXinyeWifi {
		printer.SetLineSpacing(40)
	}
	printer.LineFeed()
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.PrintText(t.base.Translate("合计"), "", t.base.GetPriceAndUnit(finalPrice), width-24))
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(2)
	// 服务费
	printer.AppendText(t.base.PrintText("("+t.base.Translate("其中服务费"), "", t.base.GetPriceAndUnit(saleOrder.ServiceFee)+")", width, leftWidth))
	printer.LineFeed(2)
	// 消费税
	if t.base.ConsumptionTax != 4 {
		printer.AppendText(t.base.PrintText("("+t.base.Translate("其中VAT"), "", t.base.GetPriceAndUnit(saleOrder.TaxFee)+")", width, leftWidth))
		printer.LineFeed(2)
	}
	if isXinyeWifi {
		printer.SetLineSpacing(60)
	}
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignLeft)

	// 未含税
	if t.base.ConsumptionTax != 4 {
		if isXinyeWifi {
			printer.SetLineSpacing(24)
		} else {
			printer.SetLineSpacing(12)
		}
		printer.LineFeed()
		printer.AppendText(lineSpacing)
		printer.AppendText(t.base.Translate("仅作为餐饮费收取以上金额"))
		printer.LineFeed(2)
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
				printer.AppendText(t.base.PrintText(taxRate+"%"+t.base.Translate("的对象"), "", t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")", width))
			} else {
				printer.AppendText(t.base.PrintText("VAT ("+taxRate+"%)", "", t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")", width))
			}
			printer.LineFeed(1)
			if isXinyeWifi {
				printer.SetLineSpacing(32)
			} else {
				printer.SetLineSpacing(22)
			}
			printer.LineFeed()
			printer.AppendText(lineSpacing)
		}
	}

	// 不包含退款金额
	if returnAmount := saleOrder.GetReturnAmount(); returnAmount > 0 {
		printer.LineFeed()
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("不包含退款金额"), t.base.GetPriceAndUnit(returnAmount)))
		printer.LineFeed(1)
	}

	// 支付方式
	printer.LineFeed()
	if isXinyeWifi {
		printer.SetLineSpacing(24)
	} else {
		printer.SetLineSpacing(22)
	}
	printer.AppendText("------------------------------------------------")
	printer.LineFeed()
	printer.SetCharacterSize(1, 1)
	printer.SetPrintModes(true, false, false)
	if saleOrder.IsFreeSaleOrder() {
		printer.AppendText(t.base.PrintText(t.base.Translate("免单"), "", t.base.GetPriceAndUnit(0), width))
		printer.LineFeed(2)
	} else {
		for key, paymentOrder := range saleOrder.PaymentOrders {
			printer.AppendText(t.base.PrintText(paymentOrder.PaymentMethod.GetName(), "", t.base.GetPriceAndUnit(paymentOrder.Amount), width))
			printer.LineFeed()
			if key < len(saleOrder.PaymentOrders)-1 {
				printer.LineFeed(2)
			}
		}
	}

	//
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)
	if isXinyeWifi {
		printer.SetLineSpacing(48)
	} else {
		printer.SetLineSpacing(20)
	}

	// 发票信息
	invoiceInfo := saleOrder.InvoiceInfo
	if invoiceInfo != nil && invoiceInfo.HasContent() {
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		printer.AppendText(t.base.Translate("发票信息"))
		printer.LineFeed(2)
		if invoiceInfo.CompanyName != "" {
			if isXinyeWifi {
				printer.SetLineSpacing(80)
			}
			printer.AppendText(t.base.Translate("公司名称") + ": " + invoiceInfo.CompanyName)
			if isXinyeWifi {
				printer.SetLineSpacing(48)
			} else {
				printer.SetLineSpacing(20)
			}
			printer.LineFeed(2)
		}
		if invoiceInfo.CompanyAddr != "" {
			if isXinyeWifi {
				printer.SetLineSpacing(80)
			}
			printer.AppendText(t.base.Translate("公司地址") + ": " + invoiceInfo.CompanyAddr)
			if isXinyeWifi {
				printer.SetLineSpacing(48)
			} else {
				printer.SetLineSpacing(20)
			}
			printer.LineFeed(2)
		}
		if invoiceInfo.CompanyTaxNumber != "" {
			if isXinyeWifi {
				printer.SetLineSpacing(80)
			}
			printer.AppendText(t.base.Translate("税号") + ": " + invoiceInfo.CompanyTaxNumber)
			if isXinyeWifi {
				printer.SetLineSpacing(48)
			} else {
				printer.SetLineSpacing(20)
			}
			printer.LineFeed(2)
		}
		if invoiceInfo.CompanyPhone != "" {
			if isXinyeWifi {
				printer.SetLineSpacing(80)
			}
			printer.AppendText(t.base.Translate("联系电话") + ": " + invoiceInfo.CompanyPhone)
			if isXinyeWifi {
				printer.SetLineSpacing(48)
			} else {
				printer.SetLineSpacing(20)
			}
			printer.LineFeed(2)
		}
	}

	// 收银员
	if isXinyeWifi {
		printer.SetLineSpacing(48)
	} else {
		printer.SetLineSpacing(20)
	}
	printer.AppendText("------------------------------------------------\n")
	printer.LineFeed()
	printer.AppendText(t.base.Translate("收银员") + ": " + saleOrder.CashierName)
	printer.LineFeed(2)
	printer.AppendText(t.base.Translate("订单号") + ": " + saleOrder.OrderNo)
	printer.LineFeed(2)
	printer.AppendText(t.base.Translate("打印次数") + ": " + strconv.Itoa(invoiceInfo.PrintNum))
	printer.LineFeed(2)
	if company != "" {
		if isXinyeWifi {
			printer.SetLineSpacing(80)
		}
		printer.AppendText(t.base.Translate("公司名称") + ": " + company)
		if isXinyeWifi {
			printer.SetLineSpacing(48)
		} else {
			printer.SetLineSpacing(20)
		}
		printer.LineFeed(2)
	}
	if address != "" {
		if isXinyeWifi {
			printer.SetLineSpacing(80)
		}
		printer.AppendText(t.base.Translate("地址") + ": " + address)
		printer.SetLineSpacing(20)
		printer.LineFeed(2)
	}
	if taxNumber != "" {
		if isXinyeWifi {
			printer.SetLineSpacing(80)
		}
		printer.AppendText(t.base.Translate("税号") + ": " + taxNumber)
		if isXinyeWifi {
			printer.SetLineSpacing(48)
		} else {
			printer.SetLineSpacing(20)
		}
		printer.LineFeed(2)
	}
	if phone != "" {
		if isXinyeWifi {
			printer.SetLineSpacing(80)
		}
		printer.AppendText(t.base.Translate("电话") + ": " + phone)
		if isXinyeWifi {
			printer.SetLineSpacing(48)
		} else {
			printer.SetLineSpacing(20)
		}
		printer.LineFeed(2)
	}

	// 保管注意事项
	printer.SetLineSpacing(80)
	printer.LineFeed()
	printer.AppendText("*" + t.base.Translate("保管注意事项"))
	if isXinyeWifi {
		printer.SetLineSpacing(48)
	} else {
		printer.SetLineSpacing(20)
	}
	printer.LineFeed(2)
	printer.SetLineSpacing(80)
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
	printer.LineFeed(4)
	printer.PrintAndExitPageMode()
	printer.LineFeed(4)
	printer.CutPaper(settingPrinterInfo.IsEnableSound())

	// 返回打印数据
	return printer.GetOrderData()
}
