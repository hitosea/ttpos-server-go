// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// takeoutOrderImgTemplate 图片外送单打印模板
type takeoutOrderImgTemplate struct {
	base *printerTemplate
}

// NewTakeoutOrderImgTemplate 创建新的图片外送单打印模板
func NewTakeoutOrderImgTemplate(
	base *printerTemplate,
) *takeoutOrderImgTemplate {
	return &takeoutOrderImgTemplate{
		base: base,
	}
}

// ImgPrint 图片打印
func (t *takeoutOrderImgTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	temp int,
	memberSaleOrder *model.MemberSaleOrder,
	saleBill *model.SaleBill,
	saleOrder *model.SaleOrder,
) string {
	// 品牌
	brandName := config.Server.BrandName
	// 日历
	payTime := t.base.FormatUnixTimeDefault(memberSaleOrder.PayTime)

	// 订单名称
	orderName := saleOrder.GetOrderName()

	//  创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetTextLineHeight(50)
	img.SetImagePadding(0)
	if t.base.Lang == "my" {
		img.SetImagePadding(3)
	}
	img.SetAlignment(pkg.AlignCenter)
	img.SetFontSize(26)
	img.SetFontWeight(2)
	img.AppendText(t.base.StoreSetting.Name)
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.LineFeed(1, 60)
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
	img.AppendText(fmt.Sprintf("%s: %s%s", t.base.Translate("外送"), saleBill.SerialNo, orderName))
	img.RecoverDefaultTextLineHeight()
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.LineFeed(3, 70)
	//
	img.RecoverDefaultTextLineHeight()
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("订单号"), Width: 200, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: saleOrder.OrderNo, Width: 0, Align: pkg.AlignRight},
	)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("支付时间"), Width: 200, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: payTime, Width: 0, Align: pkg.AlignRight},
	)

	// 商品列表 - 标题
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
	img.AppendSplitLine()
	img.LineFeed(1, 40)
	img.SetTextLineHeight(50)

	// 商品列表
	productNum := decimal.NewFromFloat(0)
	products, num := t.base.MergeSaleOrderProduct(MergeSaleOrderProductOptions{
		saleBill:   saleBill,
		saleOrder:  saleOrder,
		IsShowSku:  temp != 4,
		IsShowWrap: false,
	})
	productNum = productNum.Add(decimal.NewFromFloat(num).Round(3))
	for key, product := range products {
		img.SetTextLineHeight(40)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: product.ProductName, Width: 300, RightPadding: 40, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s*%v", t.base.Amount(product.ProductPrice), product.ProductNum), Width: 120, RightPadding: 15, Align: pkg.AlignCenter},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(product.ProductTotalPrice), Width: 0, Align: pkg.AlignRight},
		)
		// 套餐子商品
		for _, subProduct := range product.SubProducts {
			img.SetTextLineHeight(40)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: subProduct.ProductName, Width: 300, RightPadding: 40, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: fmt.Sprintf("%v", subProduct.ProductNum), Width: 120, RightPadding: 15, Align: pkg.AlignCenter},
				pkg.ColumnConfig{Text: "", Width: 0, Align: pkg.AlignRight},
			)
		}
		img.RecoverDefaultTextLineHeight()
		if key != len(products)-1 {
			img.LineFeed(1, 10)
		}
	}

	// 商品数量
	img.AppendSplitLine()
	img.LineFeed(1)
	img.SetAlignment(pkg.AlignRight)
	img.AppendText(fmt.Sprintf("%s: %v", t.base.Translate("商品数量"), productNum))
	img.LineFeed(1)
	img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("商品金额"), t.base.GetPriceAndUnit(saleOrder.ProductOriginalAmount)))
	img.LineFeed(1)

	// 分割线
	img.LineFeed(1, 10)

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

	// 税费
	if saleOrder.TaxFee > 0 {
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
					pkg.ColumnConfig{Text: fmt.Sprintf("%s (%s)", t.base.Amount(totalPrice), t.base.Amount(taxFee)), Width: 0, Align: pkg.AlignRight},
				)
			} else {
				img.PrintInColumns(
					pkg.ColumnConfig{Text: fmt.Sprintf("VAT (%s)", taxRate+"%"), Width: 280, Align: pkg.AlignLeft},
					pkg.ColumnConfig{Text: fmt.Sprintf("%s (%s)", t.base.Amount(totalPrice), t.base.Amount(taxFee)), Width: 0, Align: pkg.AlignRight},
				)
			}
		}
	}

	// 分割线
	if saleOrder.MemberDiscountFee != 0 || len(saleOrder.PaymentOrders) > 0 {
		img.AppendSplitLine()
		img.LineFeed(1)
	}

	// 会员折扣
	if memberSaleOrder.MemberDiscountFee != 0 {
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("会员折扣"), Width: 280, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s%v", "-", t.base.GetPriceAndUnit(memberSaleOrder.MemberDiscountFee)), Width: 0, Align: pkg.AlignRight},
		)
	}

	// 支付方式
	if len(saleOrder.PaymentOrders) > 0 {
		for _, paymentOrder := range saleOrder.PaymentOrders {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("支付方式"), Width: 280, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: paymentOrder.PaymentMethod.GetName(), Width: 0, Align: pkg.AlignRight},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 280, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(paymentOrder.Amount), Width: 0, Align: pkg.AlignRight},
			)
		}
	}

	// 订单备注
	if memberSaleOrder.Remark != "" {
		img.AppendSplitLine()
		img.LineFeed(1)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("顾客备注"), memberSaleOrder.Remark))
		img.LineFeed(1)
	}

	// 订单地址
	if memberSaleOrder.Address != nil {
		img.AppendSplitLine()
		img.LineFeed(1)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("联系人"), memberSaleOrder.ContactName))
		img.LineFeed(1)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("手机号码"), memberSaleOrder.GetContactPhoneMask()))
		img.LineFeed(1)
		img.AppendText(fmt.Sprintf("%s: %s", t.base.Translate("收货地址"), memberSaleOrder.ContactAddress+" "+memberSaleOrder.ContactAddressDetail))
		img.LineFeed(1)
	}

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
		img.AppendText(t.base.Translate("祝您用餐愉快！本店由") + " " + brandName + " " + t.base.Translate("系统提供支持。"))
	}
	//
	img.LineFeed(4)
	//
	img.SetSegmentationHeight(utils.IfInt(settingPrinterInfo.IsCashierPrinter, 2200, 200))
	//
	return img.Save("", !t.base.IsSunMi && settingPrinterInfo.IsEnableSound(), 0)
}
