// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/config"

	"github.com/shopspring/decimal"
)

// statementOrderSunmiTemplate Sunmi订单打印模板
type statementOrderSunmiTemplate struct {
	base *printerTemplate
}

// NewStatementOrderSunmiTemplate 创建新的Sunmi订单打印模板
func NewStatementOrderSunmiTemplate(
	base *printerTemplate,
) *statementOrderSunmiTemplate {
	return &statementOrderSunmiTemplate{
		base: base,
	}
}

// GetPrintnrContent 获取打印内容
func (t *statementOrderSunmiTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	printType int,
	temp int,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	printerType := settingPrinterInfo.PrinterType
	//
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

	// 是否自己打印
	isOneself := printerType != printerConst.PrinterTypeSunmiLan && printerType != printerConst.PrinterTypeSunmiCloud

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	//
	if temp != 3 && temp != 4 && temp != 5 {
		printer.SetAlignment(pkg.AlignLeft)
		if printType == printerConst.PrinterTemplateInvoice {
			printer.AppendText(t.base.Translate("发票"))
		} else if printType == printerConst.PrinterTemplatePreBilling {
			printer.AppendText(t.base.Translate("预结账单"))
		} else {
			printer.AppendText(t.base.Translate("结账单"))
		}
		printer.LineFeed(1)
		if isOneself {
			printer.SetLineSpacing(20)
		} else {
			printer.SetLineSpacing(40)
		}
	} else {
		if isOneself {
			printer.SetLineSpacing(20)
		} else {
			printer.SetLineSpacing(25)
		}
	}

	printer.SetAlignment(pkg.AlignCenter)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 1)
	printer.SetLineSpacing(30)
	printer.AppendText(t.base.StoreSetting.Name + "\n")
	printer.LineFeed(1)
	printer.SetCharacterSize(1, 1)
	/* *
	* 模版1
	 */
	if temp == 1 {
		printer.SetAlignment(pkg.AlignLeft)
		if isOneself {
			printer.SetLineSpacing(30)
		} else {
			printer.SetLineSpacing(70)
		}
		printer.SetPrintModes(true, true, false)
		if t.base.Lang == "my" || t.base.IsMy(saleBill.SerialNo) {
			printer.SetLineSpacing(80)
		} else {
			printer.SetLineSpacing(70)
		}
		if saleBill.DeskUuid > 0 {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s%s", orderSourceTakeoutText, t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			printer.LineFeed(1)
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("取单号"), saleBill.SerialNo, orderName))
			printer.LineFeed(1)
		}
		//
		printer.RestoreDefaultLineSpacing()
		printer.SetLineSpacing(40)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{260, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("订单号"), saleOrder.OrderNo)
		printer.PrintInColumns(t.base.Translate("收银员"), saleOrder.CashierName)
		if payTime != "" {
			printer.PrintInColumns(t.base.Translate("时间"), payTime)
			printer.SetLineSpacing(20)
		}
		printer.LineFeed()
		printer.SetLineSpacing(40)
	} else if temp == 2 {
		printer.SetLineSpacing(50)
		printer.SetPrintModes(false, false, false)
		printer.SetCharacterSize(1, 1)
		printer.AppendText(t.base.Translate("非常感谢您今天的到来，我们期待您的再次光临"))
		printer.LineFeed()
		//
		if payTime != "" {
			printer.AppendText(payTime + "\n")
			if isOneself {
				printer.SetLineSpacing(20)
				printer.LineFeed()
			}
		}
		//
		if isOneself {
			printer.SetLineSpacing(20)
		} else {
			printer.SetLineSpacing(40)
		}
		printer.SetPrintModes(true, true, false)
		if saleBill.DeskUuid > 0 {
			if isOneself {
				printer.SetLineSpacing(30)
			} else {
				printer.SetLineSpacing(70)
			}
			if t.base.Lang == "my" || t.base.IsMy(saleBill.SerialNo) {
				if isOneself {
					printer.LineFeed()
				}
				printer.SetLineSpacing(80)
			} else {
				printer.SetLineSpacing(70)
			}
			printer.AppendText(fmt.Sprintf("%s%s: %s%s%s", orderSourceTakeoutText, t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			printer.LineFeed()
			if isOneself {
				printer.SetLineSpacing(20)
			} else {
				printer.SetLineSpacing(40)
			}
		} else if saleBill.SerialNo != "" {
			if t.base.Lang == "my" || t.base.IsMy(saleBill.SerialNo) {
				printer.SetLineSpacing(80)
			} else {
				printer.SetLineSpacing(70)
			}
			printer.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("取单号"), saleBill.SerialNo, orderName))
			printer.LineFeed()
		}
		//
		printer.RestoreDefaultLineSpacing()
		printer.SetLineSpacing(40)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{260, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("订单号"), saleOrder.OrderNo)
		printer.PrintInColumns(t.base.Translate("收银员"), saleOrder.CashierName)
		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.SetLineSpacing(40)
	} else if temp == 3 || temp == 4 || temp == 5 {
		//
		printer.SetCharacterSize(2, 1)
		printer.SetPrintModes(true, true, false)
		if printType == printerConst.PrinterTemplateInvoice {
			printer.AppendText(t.base.Translate("发票"))
		} else if printType == printerConst.PrinterTemplatePreBilling {
			printer.AppendText(t.base.Translate("预结账单"))
		} else {
			printer.AppendText(t.base.Translate("结账单"))
		}
		printer.SetCharacterSize(1, 1)
		printer.LineFeed()
		printer.SetLineSpacing(25)
		printer.LineFeed()
		printer.SetLineSpacing(25)
		//
		printer.SetPrintModes(false, false, false)
		printer.SetLineSpacing(45)
		// 公司名称
		if company != "" {
			printer.AppendText(t.base.Translate("公司名称") + ": " + company)
			printer.LineFeed(1)
		}
		if chainNumber != "" {
			printer.AppendText(t.base.Translate("连锁店编号") + ": " + chainNumber)
			printer.LineFeed(1)
		}
		if address != "" {
			printer.AppendText(t.base.Translate("商家地址") + ": " + address)
			printer.LineFeed(1)
		}
		if phone != "" {
			printer.AppendText(t.base.Translate("电话") + ": " + phone)
			printer.LineFeed(1)
		}
		if taxNumber != "" {
			printer.AppendText(t.base.Translate("税号") + ": " + taxNumber)
			printer.LineFeed(1)
		}
		if temp == 5 && printType == printerConst.PrinterTemplateBilling {
			if cashierSn := t.base.GetCashierSn(settingPrinterInfo.PrinterCashierDeviceSn); cashierSn != "" {
				printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("收银机SN"), cashierSn))
				printer.LineFeed(1)
			}
			if printerSn := settingPrinterInfo.PrinterSn; printerSn != "" {
				printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("打印机SN"), printerSn))
				printer.LineFeed(1)
			}
		}
		//
		printer.AppendText("------------------------------------------------\n")
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetCharacterSize(2, 1)
		printer.SetPrintModes(true, true, false)
		if t.base.Lang == "my" || t.base.IsMy(saleBill.SerialNo) {
			printer.SetLineSpacing(80)
		} else {
			printer.SetLineSpacing(70)
		}
		if saleBill.DeskUuid > 0 {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s%s", orderSourceTakeoutText, t.base.Translate("桌号"), saleBill.SerialNo, orderName, mealNumStr))
			printer.LineFeed()
		} else if saleBill.SerialNo != "" {
			printer.AppendText(fmt.Sprintf("%s%s: %s%s", orderSourceTakeoutText, t.base.Translate("取单号"), saleBill.SerialNo, orderName))
			printer.LineFeed()
		}
		//
		printer.SetLineSpacing(45)
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
	printer.AppendText("\x1B\x33\x28")
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignLeft)
	var productWidth, priceQtyWidth int
	if t.base.Lang == "en" || t.base.Lang == "th" || t.base.Lang == "tr" || t.base.Lang == "my" {
		productWidth = 220
		priceQtyWidth = 230
		if t.base.Lang == "my" {
			productWidth = 200
			priceQtyWidth = 240
		}
	} else {
		productWidth = 320
		priceQtyWidth = 120
	}
	printer.SetupColumns(
		[]int{productWidth, pkg.AlignLeft, 0},
		[]int{priceQtyWidth, pkg.AlignCenter, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	if temp != 3 && temp != 4 && temp != 5 {
		printer.PrintInColumns(t.base.Translate("商品"), t.base.Translate("单价")+"|"+t.base.Translate("数量"), t.base.Translate("小计"))
	}
	printer.AppendText("------------------------------------------------\n")

	// 商品数量
	productNum := decimal.NewFromFloat(0)
	printer.SetupColumns(
		[]int{320, pkg.AlignLeft, 0},
		[]int{120, pkg.AlignCenter, 0},
		[]int{0, pkg.AlignRight, 0},
	)
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
		printer.PrintInColumns(
			buffetNameText,
			fmt.Sprintf("%s*%d", t.base.Amount(orderBuffetCustomer.SalePrice), orderBuffetCustomer.Num),
			t.base.GetPriceAndUnit(discountPrice),
		)
		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.SetLineSpacing(40)
	}
	// 添加加钟商品
	buffetDelayProducts, num := t.base.MergeSaleOrderBuffetDelayProducts(saleOrder)
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for _, delay := range buffetDelayProducts {
		printer.PrintInColumns(
			delay.DelayName,
			fmt.Sprintf("%s*%d", t.base.Amount(delay.DelayPrice), delay.DelayNum),
			t.base.GetPriceAndUnit(delay.DelayTotalPrice),
		)
		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.SetLineSpacing(40)
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
		printer.PrintInColumns(
			product.ProductName,
			fmt.Sprintf("%s*%v", t.base.Amount(product.ProductPrice), product.ProductNum),
			t.base.GetPriceAndUnit(product.ProductTotalPrice),
		)

		// 套餐子商品
		for _, subProduct := range product.SubProducts {
			printer.PrintInColumns(
				subProduct.ProductName,
				fmt.Sprintf("%v", subProduct.ProductNum),
				"",
			)
		}

		printer.SetLineSpacing(20)
		printer.LineFeed()
		printer.SetLineSpacing(40)
	}
	// 分割线
	printer.AppendText("------------------------------------------------\n")
	printer.SetLineSpacing(45)
	printer.SetAlignment(pkg.AlignRight)
	if temp == 3 || temp == 4 || temp == 5 {
		printer.SetupColumns(
			[]int{250, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(
			t.base.Translate("商品数量")+": "+t.base.FloatToString(productNum.Round(3).InexactFloat64()),
			t.base.Translate("商品金额")+": "+t.base.GetPriceAndUnit(saleOrder.ProductOriginalAmount),
		)
	} else {
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("商品数量") + ": " + t.base.FloatToString(productNum.Round(3).InexactFloat64()))
		printer.LineFeed()
		printer.AppendText(t.base.Translate("商品金额") + ": " + t.base.GetPriceAndUnit(saleOrder.ProductOriginalAmount))
		printer.LineFeed()
	}
	printer.SetAlignment(pkg.AlignRight)
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
				// 防止除零错误
				if saleOrder.ProductOriginalAmount > 0 {
					// 计算折扣率：折扣金额 / 原始金额 * 100
					discountRate := decimal.NewFromFloat(saleOrder.CustomDiscountFee).Div(decimal.NewFromFloat(saleOrder.ProductOriginalAmount)).Mul(decimal.NewFromInt(100))
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
	if saleOrder.IsFreeSaleOrder() && saleOrder.GetAmount() > 0 {
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("免单金额"), t.base.GetPriceAndUnit(saleOrder.GetAmount())))
		printer.LineFeed(1)
	}

	// 分隔
	if temp == 3 || temp == 4 || temp == 5 {
		printer.AppendText("------------------------------------------------\n")
	}
	if isOneself {
		printer.SetLineSpacing(12)
		printer.LineFeed()
		printer.SetLineSpacing(40)
	}
	printer.SetupColumns(
		[]int{280, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	// 应收
	printer.AppendText("\x1D\x21\x01\x01")
	printer.SetPrintModes(true, true, false)
	printer.PrintInColumns(
		t.base.Translate("合计应收"),
		t.base.GetPriceAndUnit(saleOrder.GetPrintReceivablePrice()),
	)
	printer.SetPrintModes(false, false, false)
	printer.SetLineSpacing(20)

	// 税费 - 商品已含税
	if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 2 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 2) {
		printer.LineFeed()
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(t.base.Translate("合计 (其中VAT)"))
		printer.LineFeed(2)
		percentageList := saleOrder.GetPercentageList()
		for key, percentage := range percentageList {
			taxRate := percentage["TaxRate"]
			taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
			totalPrice, _ := strconv.ParseFloat(percentage["TotalPrice"], 64)
			if t.base.Lang == "ja" {
				printer.PrintInColumns(
					fmt.Sprintf("%s%% %s", taxRate, t.base.Translate("的对象")),
					fmt.Sprintf("%s (%s)", t.base.Amount(totalPrice), t.base.Amount(taxFee)),
				)
			} else {
				printer.PrintInColumns(
					fmt.Sprintf("VAT (%s%%)", taxRate),
					t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")",
				)
			}
			if key != len(percentageList)-1 {
				printer.LineFeed()
			}
		}
	}
	if !isOneself {
		printer.LineFeed()
	}
	// 支付方式
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetLineSpacing(45)
	if saleOrder.Status == constant.SaleOrderStatusFinish {
		printer.AppendText("------------------------------------------------\n")
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		if saleOrder.IsFreeSaleOrder() {
			printer.PrintInColumns(
				t.base.Translate("支付方式"),
				t.base.Translate("免单"),
			)
			printer.PrintInColumns(
				t.base.Translate("实收金额"),
				t.base.GetPriceAndUnit(0),
			)
		}
		if len(saleOrder.PaymentOrders) > 0 {
			for _, paymentOrder := range saleOrder.PaymentOrders {
				printer.PrintInColumns(
					t.base.Translate("支付方式"),
					paymentOrder.PaymentMethod.GetName(),
				)
				printer.PrintInColumns(
					t.base.Translate("实收金额"),
					t.base.GetPriceAndUnit(paymentOrder.Amount),
				)
				if saleOrder.ChangeAmount > 0 && paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
					printer.PrintInColumns(
						t.base.Translate("找零"),
						t.base.Amount(saleOrder.ChangeAmount),
					)
				}
			}
		}
	}

	// 会员信息
	if saleOrder.Member != nil {
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
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
		printer.PrintInColumns(
			t.base.Translate("会员剩余余额"),
			t.base.GetPriceAndUnit(balance),
		)
		printer.PrintInColumns(
			t.base.Translate("本次积分"),
			t.base.Number(point),
		)
	}

	// 发票信息
	if printType == printerConst.PrinterTemplateInvoice {
		invoiceInfo := saleOrder.InvoiceInfo
		if invoiceInfo != nil && invoiceInfo.HasContent() {
			printer.SetLineSpacing(10)
			printer.LineFeed()
			printer.AppendText("------------------------------------------------\n")
			if isOneself {
				printer.LineFeed(1)
			} else {
				printer.LineFeed(2)
			}
			printer.AppendText(t.base.Translate("发票信息"))
			printer.SetLineSpacing(24)
			if isOneself {
				printer.LineFeed(1)
			} else {
				printer.LineFeed(2)
			}
			if invoiceInfo.CompanyName != "" {
				if t.base.IsMy(invoiceInfo.CompanyName) && !isOneself {
					printer.SetLineSpacing(50)
				} else {
					printer.SetLineSpacing(40)
				}
				printer.PrintInColumns(
					t.base.Translate("公司名称"),
					invoiceInfo.CompanyName,
				)
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
				if t.base.IsMy(invoiceInfo.CompanyName) && !isOneself {
					printer.SetLineSpacing(50)
				} else {
					printer.SetLineSpacing(40)
				}
				printer.PrintInColumns(
					t.base.Translate("公司地址"),
					invoiceInfo.CompanyAddr,
				)
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
				printer.PrintInColumns(
					t.base.Translate("公司税号"),
					invoiceInfo.CompanyTaxNumber,
				)
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
				printer.PrintInColumns(
					t.base.Translate("公司电话"),
					invoiceInfo.CompanyPhone,
				)
				if !isOneself {
					printer.SetLineSpacing(20)
					printer.LineFeed(2)
				} else {
					printer.LineFeed(1)
					printer.SetLineSpacing(40)
				}
			}
		} else {
			printer.SetLineSpacing(10)
			printer.LineFeed()
		}
	}

	// 技术支持方
	printer.SetLineSpacing(45)
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignCenter)
	if t.base.Lang == "th" {
		printer.AppendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " + brandName)
	} else {
		printer.AppendText(t.base.Translate("感谢您的光临！本店由") + " " + brandName + " " + t.base.Translate("系统提供支持。"))
	}

	// Print and exit page mode
	printer.LineFeed()
	printer.PrintAndExitPageMode()
	printer.LineFeed(3)
	printer.CutPaper(false)

	// 返回打印数据
	return printer.GetOrderData()
}
