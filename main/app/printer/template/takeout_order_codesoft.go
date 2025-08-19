// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/config"

	"github.com/shopspring/decimal"
)

// takeoutOrderCodesoftTemplate Codesoft订单打印模板
type takeoutOrderCodesoftTemplate struct {
	base *printerTemplate
}

// NewTakeoutOrderCodesoftTemplate 创建新的Codesoft订单打印模板
func NewTakeoutOrderCodesoftTemplate(
	base *printerTemplate,
) *takeoutOrderCodesoftTemplate {
	return &takeoutOrderCodesoftTemplate{
		base: base,
	}
}

// GetPrintnrContent 获取打印内容
func (t *takeoutOrderCodesoftTemplate) GetPrintContent(
	printerType string,
	temp int,
	memberSaleOrder *model.MemberSaleOrder,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	// 品牌
	brandName := config.Server.BrandName
	// 日历
	payTime := t.base.FormatUnixTimeDefault(memberSaleOrder.PayTime)

	// 宽度
	width := 48
	// 左侧宽度
	leftWidth := 28

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(2, 1)
	printer.SetLineSpacing(30)
	printer.AppendText(t.base.StoreSetting.Name + "\n")
	printer.LineFeed(1)
	printer.SetLineSpacing(45)
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)
	/* *
	* 模版1
	 */
	printer.SetCharacterSize(2, 2)
	printer.SetPrintModes(true, true, false)
	printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("外送"), saleBill.SerialNo))
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)
	printer.LineFeed()
	printer.SetLineSpacing(90)
	printer.SetCharacterSize(1, 1)
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignLeft)
	// 订单信息
	printer.AppendText(t.base.PrintText(t.base.Translate("订单号"), "", saleOrder.OrderNo, width, 20, 0, 28))
	printer.LineFeed()
	printer.AppendText(t.base.PrintText(t.base.Translate("支付时间"), "", payTime, width, 20, 0, 28))
	printer.LineFeed()
	printer.SetAlignment(pkg.AlignLeft)
	printer.RestoreDefaultLineSpacing()
	printer.SetPrintModes(false, false, false)
	// 商品列表表头
	leftWidth = 25
	centerWidth := 16
	rightWidth := 16
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
	// 商品列表
	productNum := decimal.NewFromFloat(0)
	products, num := t.base.MergeSaleOrderProduct(MergeSaleOrderProductOptions{
		saleBill:   saleBill,
		saleOrder:  saleOrder,
		IsShowSku:  temp != 4,
		IsShowWrap: false,
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
		printer.SetLineSpacing(50)
		printer.LineFeed()
		printer.SetLineSpacing(60)
	}

	// 商品金额 = 订单总价 - 赠品金额
	printer.AppendText("------------------------------------------------\n")
	printer.SetLineSpacing(25)
	printer.LineFeed()
	printer.SetLineSpacing(90)
	printer.SetAlignment(pkg.AlignRight)
	printer.AppendText(t.base.Translate("商品数量") + ": " + t.base.FloatToString(productNum.Round(3).InexactFloat64()))
	printer.LineFeed()
	printer.AppendText(t.base.Translate("商品金额") + ": " + t.base.GetPriceAndUnit(saleOrder.ProductOriginalAmount))
	printer.LineFeed()
	// 应收金额
	printer.AppendText("\x1D\x21\x01\x01")
	printer.SetPrintModes(true, true, false)
	finalPrice := saleOrder.GetPrintReceivablePrice()
	printer.AppendText(t.base.PrintText(t.base.Translate("合计应收"), "", t.base.GetPriceAndUnit(finalPrice), width, 34))
	if printerType != PrinterTypeXPrinterLan {
		printer.SetLineSpacing(30)
		printer.LineFeed()
	}
	// 恢复
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetPrintModes(false, false, false)
	printer.AppendText("\x1D\x21\x00\x00")

	// 税费 - 商品已含税
	if saleOrder.TaxFee > 0 {
		printer.LineFeed()
		printer.AppendText("------------------------------------------------\n")
		printer.SetLineSpacing(20)
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
				printer.AppendText(t.base.PrintText(fmt.Sprintf("%s%% %s", taxRate, t.base.Translate("的对象")), "", t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")", width, 34))
			} else {
				printer.AppendText(t.base.PrintText(fmt.Sprintf("VAT (%s%%)", taxRate), "", t.base.Amount(totalPrice)+" ("+t.base.Amount(taxFee)+")", width, 34))
			}
		}
	}

	printer.SetLineSpacing(90)
	printer.SetAlignment(pkg.AlignLeft)

	// 分割线
	if saleOrder.MemberDiscountFee != 0 || len(saleOrder.PaymentOrders) > 0 {
		printer.AppendText("------------------------------------------------\n")
	}

	// 会员折扣
	if memberSaleOrder.MemberDiscountFee != 0 {
		printer.AppendText(t.base.PrintText(t.base.Translate("会员折扣"), "", fmt.Sprintf("%s%v", "-", t.base.GetPriceAndUnit(memberSaleOrder.MemberDiscountFee)), width, 20, 0, 28))
		printer.LineFeed()
	}

	// 支付方式
	if len(saleOrder.PaymentOrders) > 0 {
		for _, paymentOrder := range saleOrder.PaymentOrders {
			printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), "", paymentOrder.PaymentMethod.GetName(), width, 20, 0, 28))
			printer.LineFeed()
			printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(paymentOrder.Amount), width, 34))
			printer.LineFeed()
		}
	}

	// 订单备注
	if memberSaleOrder.Remark != "" {
		printer.AppendText("------------------------------------------------\n")
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("顾客备注"), memberSaleOrder.Remark))
		printer.LineFeed()
	}

	// 订单地址
	if memberSaleOrder.Address != nil {
		printer.AppendText("------------------------------------------------\n")
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("联系人"), memberSaleOrder.ContactName))
		printer.LineFeed()
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("手机号码"), memberSaleOrder.GetContactPhoneMask()))
		printer.LineFeed()
		printer.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("收货地址"), memberSaleOrder.ContactAddress+" "+memberSaleOrder.ContactAddressDetail))
		printer.LineFeed()
	}

	// 技术支持方
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignCenter)
	if t.base.Lang == "th" {
		printer.AppendText("ขอบคุณที่แวะมาหากัน!สนับสนุนโดย " + brandName)
	} else {
		printer.AppendText(t.base.Translate("祝您用餐愉快！本店由") + " " + brandName + " " + t.base.Translate("系统提供支持。"))
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
