// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/config"

	"github.com/shopspring/decimal"
)

// statementOrderCodesoftTemplate Codesoft订单打印模板
type statementOrderCodesoftTemplate struct {
	base *printerTemplate
}

// NewStatementOrderCodesoftTemplate 创建新的Codesoft订单打印模板
func NewStatementOrderCodesoftTemplate(
	base *printerTemplate,
) *statementOrderCodesoftTemplate {
	return &statementOrderCodesoftTemplate{
		base: base,
	}
}

// GetPrintnrContent 获取打印内容
func (t *statementOrderCodesoftTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	printType int, // 来源 - 发票或其他
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
	brandName := config.Server.BrandName
	// 日历
	payTime := t.base.FormatUnixTimeDefault(saleOrder.FinishTime)

	// 就餐人数
	mealNumStr := ""
	if saleBill.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", saleBill.MealNum, name)
	}

	// 订单名称
	orderName := saleOrder.GetOrderName()

	// 订单来源为外卖的文本
	orderSourceTakeoutText := t.base.GetOrderSourceTakeoutText(saleBill.GetOrderSourceTakeoutText())

	// 宽度
	width := 48
	// 左侧宽度
	leftWidth := 28

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	if temp != 3 && temp != 4 && temp != 5 {
		printer.SetAlignment(pkg.AlignLeft)
		var title string
		if printType == printerConst.PrinterTemplateInvoice {
			title = t.base.Translate("发票")
		} else if printType == printerConst.PrinterTemplatePreBilling {
			title = t.base.Translate("预结账单")
		} else {
			title = t.base.Translate("结账单")
		}
		if saleBill.IsOrderSourceTakeout() {
			title += "(" + t.base.Translate("外卖") + ")"
		}
		printer.AppendText(title)
		printer.LineFeed(2)
	}

	printer.SetAlignment(pkg.AlignCenter)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 1)
	printer.AppendText(t.base.StoreSetting.Name + "\n")
	printer.SetLineSpacing(20)
	printer.LineFeed(1)
	printer.SetLineSpacing(90)
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)
	/* *
	* 模版1
	 */
	if temp == 1 {
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetLineSpacing(90)
		printer.SetCharacterSize(1, 1)
		printer.AppendText("\x1D\x21\x01")
		printer.SetPrintModes(true, true, false)
		serialNoText := saleBill.SerialNo
		if saleBill.DeskUuid > 0 {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s%s", orderSourceTakeoutText, t.base.Translate("桌号"), serialNoText, orderName, mealNumStr))
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("取单号"), serialNoText, orderName))
		}
		if t.base.Lang != "th" {
			printer.SetLineSpacing(50)
			printer.LineFeed(1)
			printer.SetLineSpacing(90)
		}
		//
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("订单号"), "", saleOrder.OrderNo, width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("收银员"), "", saleOrder.CashierName, width))
		printer.LineFeed(1)
		if payTime != "" {
			printer.AppendText(t.base.PrintText(t.base.Translate("时间"), "", payTime+"\n", width))
			printer.LineFeed()
		}
		printer.SetLineSpacing(60)
		printer.LineFeed()
		printer.SetLineSpacing(90)
	} else if temp == 2 {
		printer.AppendText(t.base.Translate("非常感谢您今天的到来，我们期待您的再次光临"))
		printer.LineFeed()
		printer.SetLineSpacing(30)
		if payTime != "" {
			printer.LineFeed()
			printer.AppendText(payTime + "\n")
		}
		//
		printer.LineFeed()
		if t.base.Lang != "th" {
			printer.SetLineSpacing(90)
		} else {
			printer.SetLineSpacing(50)
		}
		printer.SetCharacterSize(1, 1)
		printer.AppendText("\x1D\x21\x01")
		printer.SetPrintModes(true, true, false)
		serialNoText := saleBill.SerialNo
		if saleBill.DeskUuid > 0 {
			printer.SetLineSpacing(50)
			printer.AppendText(fmt.Sprintf("%s%s: %s%s%s", orderSourceTakeoutText, t.base.Translate("桌号"), serialNoText, orderName, mealNumStr))
			printer.SetLineSpacing(50)
			printer.LineFeed()
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("取单号"), serialNoText, orderName))
			printer.LineFeed()
		}
		printer.SetCharacterSize(1, 1)
		//
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("订单号"), "", saleOrder.OrderNo, width))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("收银员"), "", saleOrder.CashierName, width))
		printer.LineFeed()
		printer.SetLineSpacing(60)
		printer.LineFeed()
		printer.SetLineSpacing(90)
	} else if temp == 3 || temp == 4 || temp == 5 {
		//
		printer.SetCharacterSize(2, 2)
		printer.SetPrintModes(true, true, false)
		var title string
		if printType == printerConst.PrinterTemplateInvoice {
			title = t.base.Translate("发票")
		} else if printType == printerConst.PrinterTemplatePreBilling {
			title = t.base.Translate("预结账单")
		} else {
			title = t.base.Translate("结账单")
		}
		printer.AppendText(title)
		printer.SetPrintModes(false, false, false)
		printer.SetCharacterSize(1, 1)
		printer.LineFeed()
		//
		if t.base.Lang != "th" {
			printer.SetLineSpacing(20)
			printer.LineFeed(2)
			printer.SetLineSpacing(90)
		}
		// 公司名称
		if company != "" {
			printer.AppendText(t.base.Translate("公司名称") + ": " + company)
			printer.SetLineSpacing(20)
			printer.LineFeed(2)
		}
		if chainNumber != "" {
			printer.AppendText(t.base.Translate("连锁店编号") + ": " + chainNumber)
			printer.SetLineSpacing(20)
			printer.LineFeed(2)
		}
		if address != "" {
			printer.AppendText(t.base.Translate("商家地址") + ": " + address)
			printer.SetLineSpacing(40)
			printer.LineFeed(2)
		}
		if phone != "" {
			printer.AppendText(t.base.Translate("电话") + ": " + phone)
			printer.SetLineSpacing(40)
			printer.LineFeed(2)
		}
		if taxNumber != "" {
			printer.AppendText(t.base.Translate("税号") + ": " + taxNumber)
			printer.SetLineSpacing(40)
			printer.LineFeed(2)
		}
		if temp == 5 && printType == printerConst.PrinterTemplateBilling {
			if cashierSn := t.base.GetCashierSn(settingPrinterInfo.PrinterCashierDeviceSn); cashierSn != "" {
				printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("收银机SN"), cashierSn))
				printer.SetLineSpacing(40)
				printer.LineFeed(2)
			}
			if printerSn := settingPrinterInfo.PrinterSn; printerSn != "" {
				printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("打印机SN"), printerSn))
				printer.SetLineSpacing(40)
				printer.LineFeed(2)
			}
		}
		// 发票信息
		if printType == printerConst.PrinterTemplateInvoice {
			invoiceInfo := saleOrder.InvoiceInfo
			if invoiceInfo != nil && invoiceInfo.HasContent() {
				printer.SetLineSpacing(48)
				printer.AppendText("------------------------------------------------\n")
				if !t.base.IsThText(invoiceInfo.CompanyName) {
					printer.LineFeed()
				}
				if invoiceInfo.CompanyName != "" {
					printer.AppendText(t.base.Translate("公司名称") + ": " + invoiceInfo.CompanyName)
					printer.SetLineSpacing(20)
					printer.LineFeed(2)
				}
				if invoiceInfo.CompanyAddr != "" {
					printer.AppendText(t.base.Translate("公司地址") + ": " + invoiceInfo.CompanyAddr)
					printer.SetLineSpacing(20)
					printer.LineFeed(2)
				}
				if invoiceInfo.CompanyTaxNumber != "" {
					if !t.base.IsThText(invoiceInfo.CompanyTaxNumber) {
						printer.LineFeed(1)
					}
					printer.AppendText(t.base.Translate("公司税号") + ": " + invoiceInfo.CompanyTaxNumber)
					printer.SetLineSpacing(20)
					printer.LineFeed(2)
				}
				if invoiceInfo.CompanyPhone != "" {
					if !t.base.IsThText(invoiceInfo.CompanyPhone) {
						printer.LineFeed(1)
					}
					printer.AppendText(t.base.Translate("公司电话") + ": " + invoiceInfo.CompanyPhone)
					printer.SetLineSpacing(20)
					printer.LineFeed(2)
				}
			}
		}
		//
		if t.base.Lang != "th" {
			printer.SetLineSpacing(90)
		} else {
			printer.SetLineSpacing(10)
		}
		//
		printer.AppendText("------------------------------------------------\n")
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetCharacterSize(2, 2)
		printer.SetPrintModes(true, true, false)
		serialNoText := saleBill.SerialNo
		if saleBill.DeskUuid > 0 {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s%s", orderSourceTakeoutText, t.base.Translate("桌号"), serialNoText, orderName, mealNumStr))
			printer.LineFeed()
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("取单号"), serialNoText, orderName))
			printer.LineFeed()
		}
		if t.base.Lang != "th" {
			printer.SetLineSpacing(45)
			printer.LineFeed()
			printer.SetLineSpacing(90)
		}
		//
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		// 桌台备注
		if saleBill.Remark != "" {
			printer.AppendText(saleBill.Remark)
			printer.LineFeed(1)
		}
		printer.AppendText(t.base.Translate("收银员") + ": " + saleOrder.CashierName)
		printer.LineFeed()
		if payTime != "" {
			printer.AppendText(t.base.Translate("时间") + ": " + payTime)
			printer.LineFeed()
		}
		printer.AppendText(t.base.Translate("订单号") + ": " + saleOrder.OrderNo)
		printer.LineFeed()
	}
	//
	printer.RestoreDefaultLineSpacing()
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignLeft)
	//
	leftWidth = 25
	centerWidth := 16
	rightWidth := 16
	if temp == 3 || temp == 4 || temp == 5 {
		printer.AppendText("------------------------------------------------\n")
		printer.SetLineSpacing(25)
		printer.LineFeed()
		printer.SetLineSpacing(90)
	} else {
		printer.SetPrintModes(true, false, false)
		var hLeftWidth int
		if t.base.Lang == "en" {
			hLeftWidth = leftWidth - 10
		} else if t.base.Lang == "th" {
			hLeftWidth = leftWidth - 8
		} else {
			hLeftWidth = leftWidth
		}
		printer.AppendText(t.base.PrintText(t.base.Translate("商品"), t.base.Translate("单价")+"|"+t.base.Translate("数量"), t.base.Translate("小计"), width, hLeftWidth))
		printer.SetPrintModes(false, false, false)
		printer.AppendText("\n------------------------------------------------\n")
	}
	// 商品数量
	productNum := decimal.NewFromFloat(0)
	printer.SetLineSpacing(70)
	// 自助餐顾客类型
	for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		productNum = productNum.Add(decimal.NewFromFloat(float64(orderBuffetCustomer.Num)).Round(3))
		buffetNameText := orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNameByLang(t.base.Lang)
		// Requirement: story-main-buffet-customer-type-name-snapshot-fix
		customerTypeLocaleName := orderBuffetCustomer.GetLocaleName()
		if customerTypeLocaleName.GetLocale(t.base.Lang) != "" {
			buffetNameText += "\n(" + customerTypeLocaleName.GetLocale(t.base.Lang) + ")"
		}
		discountPrice := orderBuffetCustomer.GetOriginPrice()
		printer.AppendText(t.base.PrintText(
			buffetNameText,
			fmt.Sprintf("%s*%d", t.base.Amount(orderBuffetCustomer.SalePrice), orderBuffetCustomer.Num),
			t.base.GetPriceAndUnit(discountPrice),
			width,
			leftWidth,
			centerWidth,
			rightWidth,
		))
		printer.LineFeed()
		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.SetLineSpacing(70)
	}
	// 添加加钟商品
	buffetDelayProducts, num := t.base.MergeSaleOrderBuffetDelayProducts(saleOrder)
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, delay := range buffetDelayProducts {
		printer.AppendText(t.base.PrintText(
			delay.DelayName,
			fmt.Sprintf("%s*%d", t.base.Amount(delay.DelayPrice), delay.DelayNum),
			t.base.GetPriceAndUnit(delay.DelayTotalPrice),
			width,
			leftWidth,
			centerWidth,
			rightWidth,
		))
		printer.LineFeed()
		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.SetLineSpacing(70)
	}
	// 商品列表
	products, num := t.base.MergeSaleOrderProduct(MergeSaleOrderProductOptions{
		saleBill:   saleBill,
		saleOrder:  saleOrder,
		IsShowSku:  temp != 4,
		IsShowWrap: true,
	})
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, product := range products {
		printer.AppendText(t.base.PrintText(
			product.ProductName,
			fmt.Sprintf("%s*%v", t.base.Amount(product.ProductPrice), product.ProductNum),
			t.base.GetPriceAndUnit(product.ProductTotalPrice),
			width,
			leftWidth,
			centerWidth,
			rightWidth,
		))

		// 套餐子商品
		for k, subProduct := range product.SubProducts {
			printer.AppendText(t.base.PrintText(
				subProduct.ProductName,
				fmt.Sprintf("%v", subProduct.ProductNum),
				"",
				width,
				leftWidth,
				centerWidth,
				rightWidth,
			))
			if k != len(product.SubProducts)-1 {
				printer.LineFeed()
			}
		}

		printer.LineFeed()
		printer.SetLineSpacing(30)
		printer.LineFeed()
		printer.SetLineSpacing(70)
	}

	// 商品金额 = 订单总价 - 赠品金额
	printer.AppendText("------------------------------------------------\n")
	printer.SetLineSpacing(10)
	printer.LineFeed()
	printer.SetLineSpacing(90)
	printer.SetAlignment(pkg.AlignRight)
	if temp == 3 || temp == 4 || temp == 5 {
		printer.AppendText(t.base.PrintText(
			t.base.Translate("商品数量")+": "+t.base.FloatToString(productNum.Round(3).InexactFloat64()),
			"",
			t.base.Translate("商品金额")+": "+t.base.GetPriceAndUnit(saleOrder.ProductOriginalAmount),
			width,
		))
		printer.LineFeed()
	} else {
		printer.AppendText(t.base.Translate("商品数量") + ": " + t.base.FloatToString(productNum.Round(3).InexactFloat64()))
		printer.LineFeed()
		printer.AppendText(t.base.Translate("商品金额") + ": " + t.base.GetPriceAndUnit(saleOrder.ProductOriginalAmount))
		printer.LineFeed()
	}
	if saleOrder.ServiceFee > 0 {
		printer.AppendText(t.base.Translate("服务费") + ": " + t.base.GetPriceAndUnit(saleOrder.ServiceFee))
		printer.LineFeed()
	}

	// 税费 - 商品未含税
	if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 1 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 3) {
		for _, percentage := range saleOrder.GetPercentageList() {
			taxRate := percentage["TaxRate"]
			taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
			if t.base.Lang == "ja" {
				printer.AppendText(fmt.Sprintf("%s%% %s: %s", taxRate, t.base.Translate("的对象消费税"), t.base.GetPriceAndUnit(taxFee)))
			} else {
				if t.base.Lang != "en" && !t.base.IsThText(t.base.CurrencyUnit) {
					printer.SetLineSpacing(20)
					printer.LineFeed()
					printer.SetLineSpacing(90)
				}
				printer.AppendText(fmt.Sprintf("VAT (%s%%): %s", taxRate, t.base.GetPriceAndUnit(taxFee)))
			}
			printer.LineFeed()
		}
	}

	// 未免单 - 优惠折扣
	if !saleOrder.IsFreeSaleOrder() && saleOrder.CustomDiscountFee != 0 {
		if saleOrder.CustomDiscountFee != 0 {
			ratio := ""
			if temp == 3 || temp == 4 || temp == 5 {
				// 计算折扣率：折扣金额 / 原始金额 * 100
				discountRate := decimal.NewFromFloat(saleOrder.CustomDiscountFee).Div(decimal.NewFromFloat(saleOrder.ProductOriginalAmount)).Mul(decimal.NewFromInt(100))
				ratio = fmt.Sprintf(" (%s%% OFF)", t.base.Number(discountRate.InexactFloat64()))
			}
			//
			printer.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("优惠折扣"), t.base.GetPriceAndUnit(saleOrder.CustomDiscountFee), ratio))
			printer.LineFeed(1)
		}
	}

	// 会员优惠
	if saleOrder.MemberDiscountFee != 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("会员优惠"), t.base.GetPriceAndUnit(saleOrder.MemberDiscountFee)))
		printer.LineFeed(1)
		// 会员折扣
		oldGradeEquity := float64(100)
		oldCardDiscount := float64(100)
		gradeEquity := float64(100)
		cardDiscount := float64(100)
		if temp == 3 || temp == 4 || temp == 5 {
			if saleOrder.MemberDiscountRate != 0 {
				gradeEquity = saleOrder.MemberDiscountRate * 100
				oldGradeEquity = gradeEquity
			}
			if saleOrder.MemberCardDiscountRate != 0 {
				cardDiscount = saleOrder.MemberCardDiscountRate * 100
				oldCardDiscount = cardDiscount
			}
		}
		// 中文/繁体中文
		unit := "%"
		if t.base.Lang == "zh" || t.base.Lang == "zhtw" {
			unit = "折"
			gradeEquity /= 10
			cardDiscount /= 10
		}
		if oldGradeEquity != 100 && gradeEquity > 0 {
			printer.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("会员折扣"), t.base.Number(gradeEquity), unit))
			printer.LineFeed(1)
		}
		if oldCardDiscount != 100 && cardDiscount > 0 {
			printer.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("会员卡折扣"), t.base.Number(cardDiscount), unit))
			printer.LineFeed(1)
		}
	}

	// 会员积分抵扣
	if saleOrder.PayPointsAmount > 0 && saleOrder.PayPoints > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("会员积分抵扣"), t.base.GetPriceAndUnit(saleOrder.PayPointsAmount)))
		printer.LineFeed(1)
	}

	// 优惠券抵扣
	if couponExchangeAmount := saleOrder.CalcCouponExchangeAmount(); couponExchangeAmount > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("优惠券抵扣"), t.base.GetPriceAndUnit(couponExchangeAmount)))
		printer.LineFeed(1)
	}

	// 活动抵扣
	if saleOrder.ActivityAmount > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("活动抵扣"), t.base.GetPriceAndUnit(saleOrder.ActivityAmount)))
		printer.LineFeed(1)
	}

	// 抹零
	if checkOutZeroFee := saleOrder.GetCheckOutZeroFee(); checkOutZeroFee > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("手动抹零"), t.base.GetPriceAndUnit(checkOutZeroFee)))
		printer.LineFeed(1)
	}
	// 退款金额
	if returnAmount := saleOrder.GetReturnAmount(); returnAmount > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("退款金额"), t.base.GetPriceAndUnit(returnAmount)))
		printer.LineFeed(1)
	}
	// 支付手续费
	if saleOrder.PaymentCommissionFee > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("支付手续费"), t.base.GetPriceAndUnit(saleOrder.PaymentCommissionFee)))
		printer.LineFeed(1)
	}
	// 免单金额
	if saleOrder.IsFreeSaleOrder() {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("免单金额"), t.base.GetPriceAndUnit(saleOrder.GetAmount())))
		printer.LineFeed(1)
	}

	// 分隔
	if t.base.Lang == "th" {
		printer.SetLineSpacing(10)
	}
	if temp == 3 || temp == 4 || temp == 5 {
		printer.AppendText("------------------------------------------------")
	}

	// 应收
	printer.AppendText("\x1D\x21\x01\x01")
	printer.SetPrintModes(true, true, false)
	finalPrice := saleOrder.GetPrintReceivablePrice()
	printer.AppendText(t.base.PrintText(t.base.Translate("合计应收"), "", t.base.GetPriceAndUnit(finalPrice), width, 34))
	if t.base.Lang != "th" {
		printer.SetLineSpacing(30)
		printer.LineFeed()
	}

	// 恢复
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetPrintModes(false, false, false)
	printer.AppendText("\x1D\x21\x00\x00")

	// 税费 - 商品已含税
	if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 2 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 2) {
		printer.LineFeed()
		printer.AppendText("------------------------------------------------\n")
		printer.SetLineSpacing(5)
		printer.LineFeed()
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("合计 (其中VAT)"))
		printer.SetLineSpacing(90)
		printer.LineFeed(1)
		for _, percentage := range saleOrder.GetPercentageList() {
			taxRate := percentage["TaxRate"]
			taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
			totalPrice, _ := strconv.ParseFloat(percentage["TotalPrice"], 64)
			if t.base.Lang == "ja" {
				printer.AppendText(t.base.PrintText(fmt.Sprintf("%s%% %s", taxRate, t.base.Translate("的对象")), "", t.base.Amount(totalPrice)+" ("+t.base.GetPriceAndUnit(taxFee)+")", width, 34))
			} else {
				if t.base.Lang != "en" && !t.base.IsThText(t.base.CurrencyUnit) {
					printer.SetLineSpacing(20)
					printer.LineFeed()
					printer.SetLineSpacing(90)
				}
				printer.AppendText(t.base.PrintText(fmt.Sprintf("VAT (%s%%)", taxRate), "", t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")", width, 34))
			}
		}
	}

	// 支付方式
	if saleOrder.Status == constant.SaleOrderStatusFinish {
		if saleOrder.IsFreeSaleOrder() {
			printer.LineFeed()
			printer.SetLineSpacing(90)
			printer.AppendText("------------------------------------------------\n")
			printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), "", t.base.Translate("免单"), width, 20, 0, 28))
			printer.LineFeed()
			printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(0), width, 34))
			printer.SetLineSpacing(30)
			printer.LineFeed()
		}
		if len(saleOrder.PaymentOrders) > 0 {
			printer.LineFeed()
			printer.SetLineSpacing(90)
			printer.AppendText("------------------------------------------------\n")
			for _, paymentOrder := range saleOrder.PaymentOrders {
				printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), "", paymentOrder.PaymentMethod.GetName(), width, 20, 0, 28))
				printer.LineFeed()
				printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(paymentOrder.Amount), width, 34))
				if saleOrder.ChangeAmount > 0 && paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
					printer.AppendText(t.base.PrintText(t.base.Translate("找零"), "", t.base.Amount(saleOrder.ChangeAmount), width, 34))
				}
			}
			printer.SetLineSpacing(30)
			printer.LineFeed()
		}
	}

	// 会员信息
	if saleOrder.Member != nil {
		printer.LineFeed()
		printer.SetLineSpacing(70)
		printer.AppendText("------------------------------------------------\n")
		// 获取订单的积分发放规则
		var rule settingResp.PointsRule
		if !saleOrder.IsPaid() {
			pointsSetting, err := t.base.Setting.GetPointsSetting(t.base.Ctx)
			if err == nil {
				rule = pointsSetting.GetPointsGiftRule(saleBill.IsBuffetSaleBill(), saleOrder.Member.MemberLevelUuid)
			}
		}
		// 计算本单获取的积分
		point := saleOrder.GetMemberSurplusPoints(int(saleBill.MealNum), rule)
		balance := saleOrder.GetMemberSurplusBalance()
		printer.AppendText(t.base.PrintText(t.base.Translate("会员剩余余额"), "", t.base.GetPriceAndUnit(balance), width, 34) + "\n")
		printer.AppendText(t.base.PrintText(t.base.Translate("本次积分"), "", t.base.Number(point), width, 34))
		printer.SetLineSpacing(30)
		printer.LineFeed(2)
		printer.SetLineSpacing(90)
	}

	// 技术支持方
	printer.SetLineSpacing(90)
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignCenter)
	if t.base.Lang == "th" {
		printer.AppendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " + brandName)
	} else {
		printer.AppendText(t.base.Translate("感谢您的光临！本店由") + " " + brandName + " " + t.base.Translate("系统提供支持。"))
	}

	// Print and exit page mode
	printer.RestoreDefaultLineSpacing()
	printer.LineFeed()
	printer.PrintAndExitPageMode()
	printer.LineFeed(6)
	printer.CutPaper(settingPrinterInfo.IsEnableSound())

	// 返回打印数据
	return printer.GetOrderData()
}
