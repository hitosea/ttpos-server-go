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

// statementOrderImgTemplate 图片订单打印模板
type statementOrderImgTemplate struct {
	base *printerTemplate
}

// NewStatementOrderImgTemplate 创建新的图片订单打印模板
func NewStatementOrderImgTemplate(
	base *printerTemplate,
) *statementOrderImgTemplate {
	return &statementOrderImgTemplate{
		base: base,
	}
}

// ImgPrint 图片打印
func (t *statementOrderImgTemplate) GetPrintContent(
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
	brandName := config.Server.BrandName
	// 日历
	payTime := t.base.FormatUnixTimeDefault(saleBill.FinishTime)

	// 就餐人数
	mealNumStr := ""
	if saleBill.MealNum > 0 {
		mealNumStr = fmt.Sprintf(" (%d%s)", saleBill.MealNum, name)
	}

	// 订单名称
	orderName := fmt.Sprintf("%d", saleOrder.Index)
	if orderName != "" && saleOrder.Index > 0 {
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
		img.LineFeed(1, 50)
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
			pkg.ColumnConfig{Text: saleOrder.CashierName, Width: 0, Align: pkg.AlignRight},
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
			img.LineFeed(1, 80)
			img.RecoverDefaultTextLineHeight()
		} else {
			img.SetFontWeight(2)
			img.SetFontSize(28)
			img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("取单号"), saleBill.SerialNo, orderName))
			img.SetFontSize(20)
			img.LineFeed(1, 80)
		}
		//
		img.SetFontSize(20)
		img.RecoverDefaultTextLineHeight()
		img.SetFontWeight(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单号"), Width: 300, Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: saleOrder.OrderNo, Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		if saleOrder.CashierName != "" {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("收银员"), Width: 300, Align: pkg.AlignLeft, FontWeight: 2},
				pkg.ColumnConfig{Text: saleOrder.CashierName, Width: 0, Align: pkg.AlignRight, FontWeight: 2},
			)
		}
		img.LineFeed(1)
	} else if temp == 3 {
		// 打印logo
		if logoAddr := t.base.GetLogoAddr(); logoAddr != "" {
			img.SetTextLineHeight(25)
			img.SetAlignment(pkg.AlignCenter)
			img.AppendImg(logoAddr, 150, false, 0)
		}
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
		img.LineFeed(1, 60)
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
		if printType == constant.PrinterTemplateInvoice {
			invoiceInfo := saleOrder.InvoiceInfo
			if invoiceInfo.HasContent() {
				img.AppendSplitLine()
				if invoiceInfo.CompanyName != "" {
					img.AppendText(invoiceInfo.CompanyName)
					lineFeedHeight := 40
					if invoiceInfo.CompanyAddr != "" || invoiceInfo.CompanyTaxNumber != "" || invoiceInfo.CompanyPhone != "" {
						lineFeedHeight = 50
					}
					img.LineFeed(1, lineFeedHeight)
				}
				if invoiceInfo.CompanyAddr != "" {
					img.AppendText(invoiceInfo.CompanyAddr)
					lineFeedHeight := 40
					if invoiceInfo.CompanyTaxNumber != "" || invoiceInfo.CompanyPhone != "" {
						lineFeedHeight = 50
					}
					img.LineFeed(1, lineFeedHeight)
				}
				if invoiceInfo.CompanyTaxNumber != "" {
					img.AppendText(invoiceInfo.CompanyTaxNumber)
					lineFeedHeight := 40
					if invoiceInfo.CompanyPhone != "" {
						lineFeedHeight = 50
					}
					img.LineFeed(1, lineFeedHeight)
				}
				if invoiceInfo.CompanyPhone != "" {
					img.AppendText(invoiceInfo.CompanyPhone)
					img.LineFeed(1, 40)
				}
			}
		}
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
			img.LineFeed(1, 60)
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
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("收银员"), saleOrder.CashierName))
		img.LineFeed(1)
		if payTime != "" {
			img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("时间"), payTime))
			img.LineFeed(1)
		}
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("订单号"), saleOrder.OrderNo))
		img.LineFeed(1)
	}

	// 商品列表 - 标题
	if temp != 3 {
		var productWidth, priceQtyWidth int
		if t.base.Lang == "en" || t.base.Lang == "th" || t.base.Lang == "tr" || t.base.Lang == "my" {
			productWidth = 210
			priceQtyWidth = 230
			if t.base.Lang == "th" {
				productWidth = 200
			}
		} else {
			productWidth = 310
			priceQtyWidth = 130
		}
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

	// 赠品金额 / 商品数量
	freeMoney := float64(0)
	productNum := uint(0)
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
		img.PrintInColumns(
			pkg.ColumnConfig{Text: buffetNameText, Width: 310, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s*%d", t.base.Amount(orderBuffetCustomer.Price), orderBuffetCustomer.Num), Width: 120, Align: pkg.AlignCenter},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(discountPrice), Width: 0, Align: pkg.AlignRight},
		)
	}
	// 添加加钟商品
	for _, delay := range saleOrder.SaleOrderBuffetDelayProducts {
		if delay.IsDelete() {
			continue
		}
		productNum += delay.Num
		discountPrice := delay.GetAmount()
		img.PrintInColumns(
			pkg.ColumnConfig{Text: delay.Name, Width: 310, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s*%d", t.base.Amount(delay.Price), saleBill.MealNum), Width: 120, Align: pkg.AlignCenter},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(discountPrice), Width: 0, Align: pkg.AlignRight},
		)
	}
	// 商品列表
	for key, item := range saleOrder.SaleOrderProducts {
		if item.IsDelete() || item.IsUnCookingProduct() || item.IsCancelProduct() {
			continue
		}
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
		img.SetTextLineHeight(45)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: productName, Width: 310, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s*%d", t.base.Amount(item.Price), item.Num), Width: 120, Align: pkg.AlignCenter},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(productTotalPrice), Width: 0, Align: pkg.AlignRight},
		)
		img.RecoverDefaultTextLineHeight()
		if key != len(saleOrder.SaleOrderProducts)-1 {
			img.LineFeed(1, 10)
		}
	}

	// 商品数量
	img.AppendSplitLine()
	img.LineFeed(1)
	img.SetAlignment(pkg.AlignRight)
	// 商品金额 = 订单总价 - 赠品金额
	totalProductPrice := saleOrder.Amount - freeMoney
	if temp == 3 {
		img.PrintInColumns(
			pkg.ColumnConfig{Text: fmt.Sprintf("%s: %d", t.base.Translate("商品数量"), productNum), Width: 250, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s: %s", t.base.Translate("商品金额"), t.base.GetPriceAndUnit(totalProductPrice)), Width: 0, Align: pkg.AlignRight},
		)
	} else {
		img.SetAlignment(pkg.AlignRight)
		img.AppendText(fmt.Sprintf("%s: %d", t.base.Translate("商品数量"), productNum))
		img.LineFeed(1)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("商品金额"), t.base.GetPriceAndUnit(totalProductPrice)))
		img.LineFeed(1)
	}

	img.SetAlignment(pkg.AlignRight)

	// 服务费
	if saleOrder.ServiceFee > 0 {
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("服务费"), t.base.GetPriceAndUnit(saleOrder.ServiceFee)))
		img.LineFeed(1)
	}

	// 税费 - 商品未含税
	if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 2 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 3) {
		for _, percentage := range saleOrder.GetPercentageList() {
			taxRate := percentage["TaxRate"]
			taxFee, _ := strconv.ParseFloat(percentage["TaxFee"], 64)
			if t.base.Lang == "ja" {
				img.AppendText(fmt.Sprintf("%s: %s", taxRate+"%"+t.base.Translate("的对象消费税"), t.base.GetPriceAndUnit(taxFee)))
			} else {
				img.AppendText(fmt.Sprintf("%s: %s", "VAT ("+taxRate+"%)"+t.base.Translate("的对象消费税"), t.base.GetPriceAndUnit(taxFee)))
			}
			img.LineFeed(1)
		}
	}

	// 未免单 - 优惠折扣
	if saleOrder.IsFree == 0 && saleOrder.CustomDiscountFee != 0 {
		if saleOrder.CustomDiscountFee != 0 {
			ratio := ""
			if temp == 3 {
				if saleOrder.CustomDiscountFee <= 0 {
					ratio = " (0% OFF)"
				} else {
					// 计算折扣率：折扣金额 / 原始金额 * 100
					discountRate := decimal.NewFromFloat(saleOrder.CustomDiscountFee).Div(decimal.NewFromFloat(float64(saleOrder.GetOriginAmount()))).Mul(decimal.NewFromInt(100))
					ratio = fmt.Sprintf(" (%.0f%% OFF)", discountRate.InexactFloat64())
				}
			}
			//
			img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("优惠折扣"), t.base.GetPriceAndUnit(saleOrder.CustomDiscountFee), ratio))
			img.LineFeed(1)
		}
	}

	// 会员优惠
	if saleOrder.MemberDiscountFee != 0 {
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("会员优惠"), t.base.GetPriceAndUnit(saleOrder.MemberDiscountFee)))
		img.LineFeed(1)
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
			img.AppendText(fmt.Sprintf("%s: %.1f%s", t.base.Translate("会员折扣"), float64(gradeEquity/10), unit))
			img.LineFeed(1)
		}
		if cardDiscount != 100 && cardDiscount > 0 {
			img.AppendText(fmt.Sprintf("%s: %.1f%s", t.base.Translate("会员卡折扣"), float64(cardDiscount/10), unit))
			img.LineFeed(1)
		}
	}

	// 抹零
	if saleOrder.ZeroFee > 0 {
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("手动抹零"), t.base.GetPriceAndUnit(saleOrder.ZeroFee)))
		img.LineFeed(1)
	}
	if returnAmount := saleOrder.GetReturnAmount(); returnAmount > 0 {
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("退款金额"), t.base.GetPriceAndUnit(returnAmount)))
		img.LineFeed(1)
	}
	if saleOrder.PaymentCommissionFee > 0 {
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("支付手续费"), t.base.GetPriceAndUnit(saleOrder.PaymentCommissionFee)))
		img.LineFeed(1)
	}
	if saleOrder.IsFreeSaleOrder() {
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("免单金额"), t.base.GetPriceAndUnit(saleOrder.Amount)))
		img.LineFeed(1)
	}

	// 分割线
	if temp == 3 {
		img.AppendSplitLine()
		img.LineFeed(1)
	} else {
		img.LineFeed(1, 10)
	}

	// 合计应收
	img.SetFontSize(26)
	img.SetFontWeight(2)
	finalPrice := saleOrder.GetPrintReceivablePrice()
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("合计应收"), Width: 280, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(finalPrice), Width: 0, Align: pkg.AlignRight},
	)
	img.SetFontSize(20)
	img.SetFontWeight(1)

	// 税费 - 商品已含税
	if saleOrder.TaxFee > 0 && saleBill.SaleBillSetting.TaxFeeType == 1 && (t.base.ConsumptionTax == 1 || t.base.ConsumptionTax == 2) {
		img.AppendSplitLine()
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
					pkg.ColumnConfig{Text: taxRate + "%" + t.base.Translate("的对象"), Width: 280, Align: pkg.AlignLeft},
					pkg.ColumnConfig{Text: fmt.Sprintf("%s (%s)", t.base.GetPriceAndUnit(totalPrice), t.base.Amount(taxFee)), Width: 0, Align: pkg.AlignRight},
				)
			} else {
				img.PrintInColumns(
					pkg.ColumnConfig{Text: fmt.Sprintf("VAT (%s)", taxRate+"%"), Width: 280, Align: pkg.AlignLeft},
					pkg.ColumnConfig{Text: fmt.Sprintf("%s (%s)", t.base.GetPriceAndUnit(totalPrice), t.base.Amount(taxFee)), Width: 0, Align: pkg.AlignRight},
				)
			}
		}
	}

	// 支付方式
	if saleOrder.Status == constant.SaleOrderStatusFinish {
		if saleOrder.IsFreeSaleOrder() {
			img.AppendSplitLine()
			img.LineFeed(1)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("支付方式"), Width: 280, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: t.base.Translate("免单"), Width: 0, Align: pkg.AlignRight},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 280, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(0), Width: 0, Align: pkg.AlignRight},
			)
		}
		if len(saleOrder.PaymentOrders) > 0 {
			img.AppendSplitLine()
			img.LineFeed(1)
			for _, paymentOrder := range saleOrder.PaymentOrders {
				img.PrintInColumns(
					pkg.ColumnConfig{Text: t.base.Translate("支付方式"), Width: 280, Align: pkg.AlignLeft},
					pkg.ColumnConfig{Text: paymentOrder.PaymentMethodName, Width: 0, Align: pkg.AlignRight},
				)
				img.PrintInColumns(
					pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 280, Align: pkg.AlignLeft},
					pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(paymentOrder.Amount), Width: 0, Align: pkg.AlignRight},
				)
				if saleOrder.ChangeAmount > 0 && paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
					img.PrintInColumns(
						pkg.ColumnConfig{Text: t.base.Translate("找零"), Width: 280, Align: pkg.AlignLeft},
						pkg.ColumnConfig{Text: t.base.Amount(saleOrder.ChangeAmount), Width: 0, Align: pkg.AlignRight},
					)
				}
			}
		}
	}

	// 会员信息
	if saleOrder.Member != nil {
		img.AppendSplitLine()
		img.LineFeed(1)
		point := saleOrder.GetMemberSurplusPoints()
		balance := saleOrder.GetMemberSurplusBalance()
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("会员剩余余额"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(balance), Width: 0, Align: pkg.AlignRight},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("本次积分"), Width: 280, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.Amount(point), Width: 0, Align: pkg.AlignRight},
		)
	}

	// 打印支付二维码
	// if saleOrder.PaymentOrders != nil && saleOrder.PaymentOrders[0].PaymentMethod.QrcodeFile.GetUrl() != nil && saleOrder.PaymentOrders[0].PayType.Source != 0 && saleOrder.PaymentOrders[0].PayType.Value > 0 {
	// 	payValue := saleOrder.PaymentOrders[0].PayType.Value
	// 	paymentOrderId := saleOrder.PaymentOrders[0].Id
	// 	payType := (new PayType([], t.Ctx.AppId)).Where("source", "<>", 0).Where("value", payValue).Find()
	// 	paymentOrder := payType.Source == 2 ? (new PaymentOrder([], t.Ctx.AppId)).Where("id", paymentOrderId).Find() : nil
	// 	if payType != nil && (payType.Source != 2 || (payType.Source == 2 && paymentOrder != nil)) {
	// 		img.AppendSplitLine()
	// 		img.LineFeed(1)
	// 		img.SetAlignment(pkg.AlignCenter)
	// 		img.SetTextLineHeight(35)
	// 		img.AppendText(t.base.Translate("请用") + " " + payType.Name + " " + t.base.Translate("扫一扫支付"))
	// 		img.LineFeed(1)
	// 		// 1-自行添加 2-LianLianPay
	// 		if payType.Source == 1 {
	// 			$printer->appendImg('http://nginx' . $payType->qrcode);
	// 		} else {
	// 			if ($paymentOrder->link_url) {
	// 				$printer->appendQrcode($paymentOrder->link_url);
	// 			} else {
	// 				$printer->appendQrcode($paymentOrder->qr_code, 280, 10, true);
	// 			}
	// 		}
	// 		$printer->setTextLineHeight(50);
	// 	}
	// }

	// img.AppendSplitLine()
	// img.LineFeed(1)
	// img.SetAlignment(pkg.AlignCenter)
	// img.SetTextLineHeight(35)
	// img.AppendText(t.base.Translate("请用") + " " + "payType.Name" + " " + t.base.Translate("扫一扫支付"))
	// img.LineFeed(1)
	// img.AppendQrcode("logoAddr", 280, 0, false)

	// 技术支持方
	img.AppendSplitLine()
	img.LineFeed(1, 50)
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

	return img.Save("", !t.base.IsSunMi, false)
}
