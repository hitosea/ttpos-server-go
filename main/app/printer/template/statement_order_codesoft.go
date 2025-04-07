// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
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
	payTime := t.base.FormatUnixTimeDefault(saleBill.FinishTime)

	// 就餐人数
	mealNumStr := ""
	if saleBill.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", saleBill.MealNum, name)
	}

	// 订单名称
	orderName := saleOrder.GetOrderName()

	// 宽度
	width := 48
	// 左侧宽度
	leftWidth := 28

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	if temp != 3 {
		printer.SetAlignment(pkg.AlignLeft)
		if printType == constant.PrinterTemplateInvoice {
			printer.AppendText(t.base.Translate("发票"))
		} else if printType == constant.PrinterTemplatePreBilling {
			printer.AppendText(t.base.Translate("预结账单"))
		} else {
			printer.AppendText(t.base.Translate("结账单"))
		}
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
		if saleBill.DeskUuid > 0 {
			printer.AppendText(fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName))
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
		if saleBill.DeskUuid > 0 {
			printer.SetLineSpacing(50)
			printer.AppendText(fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			printer.SetLineSpacing(50)
			printer.LineFeed()
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName))
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
	} else if temp == 3 {
		//
		printer.SetCharacterSize(2, 2)
		printer.SetPrintModes(true, true, false)
		if printType == constant.PrinterTemplateInvoice {
			printer.AppendText(t.base.Translate("发票"))
		} else if printType == constant.PrinterTemplatePreBilling {
			printer.AppendText(t.base.Translate("预结账单"))
		} else {
			printer.AppendText(t.base.Translate("结账单"))
		}
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
		// 发票信息
		if printType == constant.PrinterTemplateInvoice {
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
		if saleBill.DeskUuid > 0 {
			printer.AppendText(fmt.Sprintf("%s: %s%s%s", t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			printer.LineFeed()
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName))
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
	if temp == 3 {
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
	// 赠品金额 / 商品数量
	freeMoney := float64(0)
	productNum := uint(0)
	printer.SetLineSpacing(70)
	// 自助餐顾客类型
	for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
		if orderBuffetCustomer.IsDelete() {
			continue
		}
		productNum += orderBuffetCustomer.Num
		buffetNameText := orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNameByLang(t.base.Lang)
		if orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name != "" {
			buffetNameText += "\n(" + orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name + ")"
		}
		discountPrice := orderBuffetCustomer.GetDiscountPrice()
		printer.AppendText(t.base.PrintText(
			buffetNameText,
			fmt.Sprintf("%s*%d", t.base.Amount(orderBuffetCustomer.Price), orderBuffetCustomer.Num),
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
	for _, delay := range saleOrder.SaleOrderBuffetDelayProducts {
		if delay.IsDelete() {
			continue
		}
		productNum += delay.Num
		discountPrice := delay.GetAmount()
		printer.AppendText(t.base.PrintText(
			delay.Name,
			fmt.Sprintf("%s*%d", t.base.Amount(delay.Price), delay.Num),
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
	// 商品列表
	for _, item := range saleOrder.SaleOrderProducts {
		if item.IsDelete() || item.IsUnCookingProduct() || item.IsCancelProduct() {
			continue
		}
		if item.IsBuffetProduct() && item.GetPrice() <= 0 {
			continue
		}
		//
		productNum += item.Num
		productTotalPrice := item.GetPrice()
		// 赠品
		var gift string
		if item.IsGiftBool() {
			gift = "(" + t.base.Translate("赠") + ") "
			freeMoney += item.GetPrice()
			productTotalPrice = 0
		}
		productName := gift + item.MultiLanguageName.GetNameByLang(t.base.Lang) + "\n" + item.GetAttributeNamesByLang(t.base.Lang)
		//
		printer.AppendText(t.base.PrintText(
			productName,
			fmt.Sprintf("%s*%d", t.base.Amount(item.Price), item.Num),
			t.base.GetPriceAndUnit(productTotalPrice),
			width,
			leftWidth,
			centerWidth,
			rightWidth,
		))
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
	if temp == 3 {
		printer.AppendText(t.base.PrintText(
			t.base.Translate("商品数量")+": "+fmt.Sprintf("%d", productNum),
			"",
			t.base.Translate("商品金额")+": "+t.base.GetPriceAndUnit(saleOrder.ProductAmount),
			width,
		))
		printer.LineFeed()
	} else {
		printer.AppendText(t.base.Translate("商品数量") + ": " + fmt.Sprintf("%d", productNum))
		printer.LineFeed()
		printer.AppendText(t.base.Translate("商品金额") + ": " + t.base.GetPriceAndUnit(saleOrder.ProductAmount))
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
			if temp == 3 {
				if saleOrder.CustomDiscountFee <= 0 {
					ratio = " (0% OFF)"
				} else {
					// 计算折扣率：折扣金额 / 原始金额 * 100
					discountRate := decimal.NewFromFloat(saleOrder.CustomDiscountFee).Div(decimal.NewFromFloat(float64(saleOrder.GetOriginAmount()))).Mul(decimal.NewFromInt(100))
					ratio = fmt.Sprintf(" (%s%% OFF)", t.base.Number(discountRate.InexactFloat64()))
				}
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
		gradeEquity := float64(100)
		cardDiscount := float64(100)
		if temp == 3 {
			if saleOrder.MemberDiscountRate != 0 {
				gradeEquity = saleOrder.MemberDiscountRate
			}
			if saleOrder.MemberCardDiscountRate != 0 {
				cardDiscount = saleOrder.MemberCardDiscountRate
			}
		}
		// 中文/繁体中文
		unit := "%"
		if t.base.Lang == "zh" || t.base.Lang == "zhtw" {
			unit = "折"
		}
		if gradeEquity != 100 && gradeEquity > 0 {
			printer.AppendText(fmt.Sprintf("%s: %.1f%s", t.base.Translate("会员折扣"), float64(gradeEquity/10), unit))
			printer.LineFeed(1)
		}
		if cardDiscount != 100 && cardDiscount > 0 {
			printer.AppendText(fmt.Sprintf("%s: %.1f%s", t.base.Translate("会员卡折扣"), float64(cardDiscount/10), unit))
			printer.LineFeed(1)
		}
	}

	// 抹零
	if saleOrder.ZeroFee > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("手动抹零"), t.base.GetPriceAndUnit(saleOrder.ZeroFee)))
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
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("免单金额"), t.base.GetPriceAndUnit(saleOrder.Amount)))
		printer.LineFeed(1)
	}

	// 分隔
	if t.base.Lang == "th" {
		printer.SetLineSpacing(10)
	}
	if temp == 3 {
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
				printer.AppendText(t.base.PrintText(fmt.Sprintf("%s%% %s", taxRate, t.base.Translate("的对象")), "", t.base.GetPriceAndUnit(totalPrice)+" ("+t.base.GetPriceAndUnit(taxFee)+")", width, 34))
			} else {
				if t.base.Lang != "en" && !t.base.IsThText(t.base.CurrencyUnit) {
					printer.SetLineSpacing(20)
					printer.LineFeed()
					printer.SetLineSpacing(90)
				}
				printer.AppendText(t.base.PrintText(fmt.Sprintf("VAT (%s%%)", taxRate), "", t.base.GetPriceAndUnit(totalPrice)+" ("+t.base.GetPriceAndUnit(taxFee)+")", width, 34))
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
				printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), "", paymentOrder.PaymentMethodName, width, 20, 0, 28))
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
		// 获取商家当前的积分赠送比例
		var giftRatio float64 = 0
		if !saleOrder.IsPaid() {
			pointsSetting, err := t.base.Setting.GetPointsSetting(t.base.Ctx)
			if err == nil {
				giftRatio = pointsSetting.GetGiftRatio()
			}
		}
		// 计算本单获取的积分
		point := saleOrder.GetMemberSurplusPoints(giftRatio)
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
	printer.CutPaper(true)

	// 返回打印数据
	return printer.GetOrderData()
}
