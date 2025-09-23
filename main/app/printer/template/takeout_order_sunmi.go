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

	"github.com/shopspring/decimal"
)

// takeoutOrderSunmiTemplate Sunmi订单打印模板
type takeoutOrderSunmiTemplate struct {
	base *printerTemplate
}

// NewTakeoutOrderSunmiTemplate 创建新的Sunmi订单打印模板
func NewTakeoutOrderSunmiTemplate(
	base *printerTemplate,
) *takeoutOrderSunmiTemplate {
	return &takeoutOrderSunmiTemplate{
		base: base,
	}
}

// GetPrintnrContent 获取打印内容
func (t *takeoutOrderSunmiTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	temp int,
	memberSaleOrder *model.MemberSaleOrder,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	printerType := settingPrinterInfo.PrinterType
	// 品牌
	brandName := config.Server.BrandName
	// 日历
	payTime := t.base.FormatUnixTimeDefault(memberSaleOrder.PayTime)
	// 是否自己打印
	isOneself := printerType != constant.PrinterTypeSunmiLan && printerType != constant.PrinterTypeSunmiCloud

	// 宽度
	width := 48
	// 默认行间距
	defaultLineSpacing := 45

	// 创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 1)
	printer.AppendText(t.base.StoreSetting.Name + "\n")
	printer.SetLineSpacing(20)
	printer.LineFeed(2)
	printer.SetLineSpacing(defaultLineSpacing)
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
	printer.LineFeed()
	printer.SetLineSpacing(defaultLineSpacing)
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
	printer.PrintInColumns(t.base.Translate("商品"), t.base.Translate("单价")+"|"+t.base.Translate("数量"), t.base.Translate("小计"))
	printer.AppendText("------------------------------------------------\n")

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
	// 商品金额 = 订单总价 - 赠品金额
	printer.AppendText("------------------------------------------------\n")
	printer.SetLineSpacing(45)
	printer.SetAlignment(pkg.AlignRight)
	printer.SetAlignment(pkg.AlignRight)
	printer.AppendText(t.base.Translate("商品数量") + ": " + t.base.FloatToString(productNum.Round(3).InexactFloat64()))
	printer.LineFeed()
	printer.AppendText(t.base.Translate("商品金额") + ": " + t.base.GetPriceAndUnit(saleOrder.ProductOriginalAmount))
	printer.LineFeed()

	// 应收金额
	printer.SetupColumns(
		[]int{280, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	printer.AppendText("\x1D\x21\x01\x01")
	printer.SetPrintModes(true, true, false)
	printer.PrintInColumns(
		t.base.Translate("合计应收"),
		t.base.GetPriceAndUnit(saleOrder.GetPrintReceivablePrice()),
	)
	printer.SetPrintModes(false, false, false)
	printer.SetLineSpacing(20)

	// 税费 - 商品已含税
	if saleOrder.TaxFee > 0 {
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

	printer.SetLineSpacing(defaultLineSpacing)
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetupColumns(
		[]int{320, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)

	// 分割线
	if saleOrder.MemberDiscountFee != 0 || len(saleOrder.PaymentOrders) > 0 {
		printer.AppendText("------------------------------------------------\n")
		printer.SetLineSpacing(defaultLineSpacing)
	}

	// 会员折扣
	if memberSaleOrder.MemberDiscountFee != 0 {
		printer.PrintInColumns(
			t.base.Translate("会员折扣"),
			fmt.Sprintf("%s%v", "-", t.base.GetPriceAndUnit(memberSaleOrder.MemberDiscountFee)),
		)
	}

	// 支付方式
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
	printer.CutPaper(settingPrinterInfo.IsEnableSound())

	// 返回打印数据
	return printer.GetOrderData()
}
