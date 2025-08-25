// Package template 提供打印模板相关功能
package template

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/pkg/utils"
)

// rechargeImgTemplate 图片订单打印模板
type rechargeImgTemplate struct {
	base *printerTemplate
}

// NewRechargeImgTemplate 创建新的图片订单打印模板
func NewRechargeImgTemplate(
	base *printerTemplate,
) *rechargeImgTemplate {
	return &rechargeImgTemplate{
		base: base,
	}
}

// ImgPrint 图片打印
func (t *rechargeImgTemplate) GetPrintContent(
	order model.MemberRechargeOrder,
) string {
	// 日历
	payTime := t.base.FormatUnixTimeDefault(order.PaymentTime)

	// 收银员名称
	cashierName := utils.IfString(order.Staff.RealName != "", order.Staff.RealName, order.Staff.Username)

	//  创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetTextLineHeight(45)
	img.SetImagePadding(0)
	img.SetAlignment(pkg.AlignLeft)
	img.AppendText(t.base.Translate("充值单"))
	img.LineFeed(1)
	img.LineFeed(1, 24)
	img.SetAlignment(pkg.AlignCenter)
	img.SetFontWeight(2)
	img.SetFontSize(34)
	img.AppendText(t.base.StoreSetting.Name)
	img.SetFontSize(20)
	img.SetFontWeight(1)
	img.LineFeed(1)
	img.LineFeed(1, 24)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("收银员"), Width: 350, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: cashierName, Width: 0, Align: pkg.AlignRight},
	)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("时间"), Width: 300, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: payTime, Width: 0, Align: pkg.AlignRight},
	)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("充值前"), Width: 350, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(order.Balance), Width: 0, Align: pkg.AlignRight},
	)
	// 本次充值
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("本次充值"), Width: 350, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(order.RechargeAmount), Width: 0, Align: pkg.AlignRight},
	)
	// 赠送金额
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("赠送金额"), Width: 350, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(order.GiftAmount), Width: 0, Align: pkg.AlignRight},
	)
	// 赠送积分
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("赠送积分"), Width: 350, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.Amount(order.GiftPoint), Width: 0, Align: pkg.AlignRight},
	)
	img.SetTextLineHeight(35)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("充值后"), Width: 350, Align: pkg.AlignLeft},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(order.BalanceRecharged), Width: 0, Align: pkg.AlignRight},
	)
	// 退款
	if order.RefundMoney > 0 {
		img.AppendSplitLine()
		img.LineFeed(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("退款"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(order.RefundMoney), Width: 0, Align: pkg.AlignRight},
		)
	}
	// 合计應收：
	img.AppendSplitLine()
	img.LineFeed(1)
	img.LineFeed(1, 14)
	img.PrintInColumns(
		pkg.ColumnConfig{Text: t.base.Translate("合计应收："), Width: 350, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 28},
		pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(order.GetReceivableAmount()), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 28},
	)
	img.LineFeed(1, 14)
	// 支付方式
	img.AppendSplitLine()
	img.LineFeed(1)
	img.SetTextLineHeight(45)
	for _, paymentOrder := range order.PaymentOrders {
		additional := ""
		if t.base.Lang == "my" {
			additional = " "
		}
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("支付方式"), Width: 280, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: paymentOrder.PaymentMethod.GetName() + additional, Width: 0, Align: pkg.AlignRight},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 280, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(paymentOrder.Amount), Width: 0, Align: pkg.AlignRight},
		)
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("找零"), Width: 280, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: t.base.Amount(order.ChargeDue), Width: 0, Align: pkg.AlignRight},
			)
		}
	}
	//
	img.LineFeed(2)
	img.LineFeed(1, 20)
	//
	return img.Save("", !t.base.IsSunMi, 0)
}
