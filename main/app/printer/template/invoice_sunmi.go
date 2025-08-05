// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"
)

// invoiceSunmiTemplate xprinter发票打印模板
type invoiceSunmiTemplate struct {
	base *printerTemplate
}

// NewInvoiceXprinterTemplate 创建新的xprinter发票打印模板
func NewInvoiceSunmiTemplate(
	base *printerTemplate,
) *invoiceSunmiTemplate {
	return &invoiceSunmiTemplate{
		base: base,
	}
}

// GetPrintContent 获取打印内容
func (t *invoiceSunmiTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	printerType string,
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
		return NewStatementOrderSunmiTemplate(t.base).GetPrintContent(
			settingPrinterInfo,
			printerType,
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

	// 是否自己打印
	isOneself := printerType != constant.PrinterTypeSunmiLan && printerType != constant.PrinterTypeSunmiCloud

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.RestoreDefaultLineSpacing()
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetLineSpacing(35)
	printer.SetCharacterSize(2, 1)
	printer.AppendText(t.base.StoreSetting.Name)
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(2)
	printer.AppendText(t.base.Translate("非常感谢您今天的到来，我们期待您的再次光临"))
	printer.LineFeed(2)
	printer.SetLineSpacing(28)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.Translate("发票"))
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(2)
	printer.AppendText(payTime)
	printer.LineFeed(2)
	//
	if t.base.Lang == "ja" {
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("先生/小姐"))
		printer.LineFeed(1)
	}
	//
	printer.AppendText("------------------------------------------------\n")
	printer.SetLineSpacing(10)
	printer.LineFeed()
	printer.SetCharacterSize(2, 2)
	printer.SetupColumns(
		[]int{260, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	printer.PrintInColumns(t.base.Translate("合计"), t.base.GetPriceAndUnit(finalPrice))
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(2)
	// 服务费
	printer.SetupColumns(
		[]int{350, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	printer.PrintInColumns("("+t.base.Translate("其中服务费"), t.base.GetPriceAndUnit(saleOrder.ServiceFee)+")")
	printer.LineFeed(2)
	// 消费税
	if t.base.ConsumptionTax != 4 {
		printer.PrintInColumns("("+t.base.Translate("其中VAT"), t.base.GetPriceAndUnit(saleOrder.TaxFee)+")")
		printer.LineFeed(2)
	}

	//
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetLineSpacing(12)
	printer.LineFeed()
	printer.SetLineSpacing(20)
	// 未含税
	if t.base.ConsumptionTax != 4 {
		printer.AppendText(t.base.Translate("仅作为餐饮费收取以上金额"))
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("合计 (其中VAT)"))
		printer.LineFeed(1)
		printer.SetLineSpacing(12)
		printer.LineFeed()
		printer.SetLineSpacing(20)
		printer.SetupColumns(
			[]int{200, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetAlignment(pkg.AlignLeft)
		percentageList := saleOrder.GetPercentageList()
		for key, percentage := range percentageList {
			taxRate := percentage["TaxRate"]
			taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
			totalPrice, _ := strconv.ParseFloat(percentage["TotalPrice"], 64)
			if t.base.Lang == "ja" {
				printer.PrintInColumns(
					taxRate+"%"+t.base.Translate("的对象"),
					t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")",
				)
			} else {
				printer.PrintInColumns(
					"VAT ("+taxRate+"%)",
					t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")",
				)
			}
			if key != len(percentageList)-1 {
				printer.LineFeed()
			}
		}
	}

	// 不包含退款金额
	if returnAmount := saleOrder.GetReturnAmount(); returnAmount > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("不包含退款金额"), t.base.GetPriceAndUnit(returnAmount)))
		printer.LineFeed(1)
	}
	printer.SetLineSpacing(12)
	printer.LineFeed()

	// 支付方式
	printer.AppendText("------------------------------------------------\n")
	printer.LineFeed()
	printer.SetupColumns(
		[]int{320, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	printer.SetCharacterSize(1, 1)
	printer.SetPrintModes(true, false, false)
	if saleOrder.IsFreeSaleOrder() {
		printer.PrintInColumns(t.base.Translate("免单"), t.base.GetPriceAndUnit(0))
		printer.LineFeed(2)
	} else {
		for key, paymentOrder := range saleOrder.PaymentOrders {
			printer.PrintInColumns(paymentOrder.PaymentMethod.GetName(), t.base.GetPriceAndUnit(paymentOrder.Amount))
			printer.LineFeed()
			if key < len(saleOrder.PaymentOrders)-1 {
				printer.LineFeed(2)
			}
		}
	}
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)

	// 发票信息
	invoiceInfo := saleOrder.InvoiceInfo
	if invoiceInfo != nil && invoiceInfo.HasContent() {
		printer.SetLineSpacing(10)
		printer.LineFeed()
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		printer.AppendText(t.base.Translate("发票信息"))
		printer.SetLineSpacing(20)
		if isOneself {
			printer.LineFeed(1)
		} else {
			printer.LineFeed(2)
		}
		if invoiceInfo.CompanyName != "" {
			printer.SetLineSpacing(40)
			if t.base.IsMy(invoiceInfo.CompanyName) && !isOneself {
				printer.SetLineSpacing(50)
			}
			printer.AppendText(t.base.Translate("公司名称") + ": " + invoiceInfo.CompanyName)
			if !isOneself {
				printer.SetLineSpacing(20)
				if invoiceInfo.CompanyAddr != "" || invoiceInfo.CompanyTaxNumber != "" || invoiceInfo.CompanyPhone != "" {
					printer.LineFeed(2)
				} else {
					printer.LineFeed(1)
				}
			} else {
				printer.LineFeed(1)
				printer.SetLineSpacing(40)
			}
		}
		if invoiceInfo.CompanyAddr != "" {
			printer.SetLineSpacing(40)
			if t.base.IsMy(invoiceInfo.CompanyAddr) && !isOneself {
				printer.SetLineSpacing(50)
			}
			printer.AppendText(t.base.Translate("公司地址") + ": " + invoiceInfo.CompanyAddr)
			if !isOneself {
				printer.SetLineSpacing(20)
				if invoiceInfo.CompanyTaxNumber != "" || invoiceInfo.CompanyPhone != "" {
					printer.LineFeed(2)
				} else {
					printer.LineFeed(1)
				}
			} else {
				printer.LineFeed(1)
				printer.SetLineSpacing(40)
			}
		}
		if invoiceInfo.CompanyTaxNumber != "" {
			printer.SetLineSpacing(40)
			if t.base.IsMy(invoiceInfo.CompanyTaxNumber) && !isOneself {
				printer.SetLineSpacing(50)
			}
			printer.AppendText(t.base.Translate("税号") + ": " + invoiceInfo.CompanyTaxNumber)
			if !isOneself {
				printer.SetLineSpacing(20)
				if invoiceInfo.CompanyPhone != "" {
					printer.LineFeed(2)
				} else {
					printer.LineFeed(1)
				}
			} else {
				printer.LineFeed(1)
				printer.SetLineSpacing(40)
			}
		}
		if invoiceInfo.CompanyPhone != "" {
			printer.SetLineSpacing(40)
			if t.base.IsMy(invoiceInfo.CompanyPhone) && !isOneself {
				printer.SetLineSpacing(50)
			}
			printer.AppendText(t.base.Translate("联系电话") + ": " + invoiceInfo.CompanyPhone)
			if !isOneself {
				printer.SetLineSpacing(20)
				printer.LineFeed(2)
			} else {
				printer.LineFeed(1)
				printer.SetLineSpacing(40)
			}
		}
	}

	// 收银员
	printer.AppendText("------------------------------------------------\n")
	if !isOneself {
		printer.SetLineSpacing(10)
		printer.LineFeed()
	}
	printer.SetLineSpacing(40)
	printer.AppendText(t.base.Translate("收银员") + ": " + saleOrder.CashierName)
	printer.LineFeed(1)
	printer.AppendText(t.base.Translate("订单号") + ": " + saleOrder.OrderNo)
	printer.LineFeed(1)
	printer.AppendText(t.base.Translate("打印次数") + ": " + strconv.Itoa(invoiceInfo.PrintNum))
	printer.LineFeed(1)
	if company != "" {
		printer.SetLineSpacing(40)
		if t.base.IsMy(company) && !isOneself {
			printer.SetLineSpacing(50)
		}
		printer.AppendText(t.base.Translate("公司名称") + ": " + company)
		if !isOneself {
			printer.SetLineSpacing(20)
			printer.LineFeed(2)
		} else {
			printer.LineFeed(1)
			printer.SetLineSpacing(40)
		}
	}
	if address != "" {
		printer.SetLineSpacing(40)
		if t.base.IsMy(address) && !isOneself {
			printer.SetLineSpacing(50)
		}
		printer.AppendText(t.base.Translate("地址") + ": " + address)
		if !isOneself {
			printer.SetLineSpacing(20)
			printer.LineFeed(2)
		} else {
			printer.LineFeed(1)
			printer.SetLineSpacing(40)
		}
	}
	if taxNumber != "" {
		printer.SetLineSpacing(40)
		if t.base.IsMy(taxNumber) && !isOneself {
			printer.SetLineSpacing(50)
		}
		printer.AppendText(t.base.Translate("税号") + ": " + taxNumber)
		if !isOneself {
			printer.SetLineSpacing(20)
			printer.LineFeed(2)
		} else {
			printer.LineFeed(1)
			printer.SetLineSpacing(40)
		}
	}
	if phone != "" {
		printer.SetLineSpacing(40)
		if t.base.IsMy(phone) && !isOneself {
			printer.SetLineSpacing(50)
		}
		printer.AppendText(t.base.Translate("电话") + ": " + phone)
		if !isOneself {
			printer.SetLineSpacing(20)
			printer.LineFeed(2)
		} else {
			printer.LineFeed(1)
			printer.SetLineSpacing(40)
		}
	}
	printer.LineFeed(1)

	// 保管注意事项
	printer.AppendText("*" + t.base.Translate("保管注意事项"))
	if !isOneself {
		printer.LineFeed(2)
	} else {
		printer.LineFeed(1)
	}
	printer.AppendText(t.base.Translate("如需保管时请将印刷页面朝内折叠"))

	// 技术支持方
	printer.LineFeed(1)
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetLineSpacing(50)
	if t.base.Lang == "th" {
		printer.AppendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " + brandName)
	} else {
		printer.AppendText(t.base.Translate("感谢您的光临！本店由") + " " + brandName + " " + t.base.Translate("系统提供支持。"))
	}

	// Print and exit page mode
	printer.LineFeed(6)
	printer.PrintAndExitPageMode()
	printer.CutPaper(false)

	// 返回打印数据
	return printer.GetOrderData()
}
