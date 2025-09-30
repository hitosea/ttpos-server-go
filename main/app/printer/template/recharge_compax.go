// Package template 提供打印模板相关功能
package template

import (
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/pkg/utils"
)

// rechargeCompaxTemplate 图片订单打印模板
type rechargeCompaxTemplate struct {
	base *printerTemplate
}

// NewRechargeCompaxTemplate 创建新的图片订单打印模板
func NewRechargeCompaxTemplate(
	base *printerTemplate,
) *rechargeCompaxTemplate {
	return &rechargeCompaxTemplate{
		base: base,
	}
}

// GetPrintContent 获取打印内容
func (t *rechargeCompaxTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	order model.MemberRechargeOrder,
) string {
	// 日历
	payTime := t.base.FormatUnixTimeDefault(order.PaymentTime)

	// 收银员名称
	cashierName := utils.IfString(order.Staff.RealName != "", order.Staff.RealName, order.Staff.Username)

	/* *
	* 打印模版
	 */
	width := 48
	printer := pkg.NewPrinter(567)
	printer.SetLineSpacing(35)
	printer.SetAlignment(pkg.AlignLeft)
	printer.AppendText(t.base.Translate("充值单"))
	printer.LineFeed(2)
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.StoreSetting.Name)
	printer.LineFeed(3)

	//
	printer.SetLineSpacing(45)
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(1)
	printer.AppendText(t.base.PrintText(t.base.Translate("收银员"), "", cashierName, width))
	printer.AppendText(t.base.PrintText(t.base.Translate("时间"), "", payTime, width))
	printer.AppendText(t.base.PrintText(t.base.Translate("充值前"), "", t.base.GetPriceAndUnit(order.Balance), width))
	printer.AppendText(t.base.PrintText(t.base.Translate("本次充值"), "", t.base.GetPriceAndUnit(order.RechargeAmount), width))
	printer.AppendText(t.base.PrintText(t.base.Translate("赠送金额"), "", t.base.GetPriceAndUnit(order.GiftAmount), width))
	printer.AppendText(t.base.PrintText(t.base.Translate("赠送积分"), "", t.base.Amount(order.GiftPoint), width))
	printer.AppendText(t.base.PrintText(t.base.Translate("充值后"), "", t.base.GetPriceAndUnit(order.BalanceRecharged), width))
	printer.SetLineSpacing(35)
	// 退款
	if order.RefundMoney > 0 {
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("退款"), "", t.base.GetPriceAndUnit(order.RefundMoney), width))
	}
	// 合计應收：
	printer.AppendText("------------------------------------------------\n")
	printer.SetAlignment(pkg.AlignLeft)
	printer.SetCharacterSize(2, 1)
	printer.SetPrintModes(true, true, false)
	printer.AppendText(t.base.PrintText(t.base.Translate("合计应收："), "", t.base.GetPriceAndUnit(order.GetReceivableAmount()), width))
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)
	printer.AppendText("\n------------------------------------------------\n")
	printer.SetLineSpacing(45)
	// 支付方式
	for _, paymentOrder := range order.PaymentOrders {
		printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), "", paymentOrder.PaymentMethod.GetName(), width, 20, 0, 28) + "\n")
		printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(paymentOrder.Amount), width, 34))
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			printer.AppendText(t.base.PrintText(t.base.Translate("找零"), "", t.base.Amount(order.ChargeDue), width, 34))
		}
	}

	// Print and exit page mode
	printer.RestoreDefaultLineSpacing()
	printer.LineFeed()
	printer.PrintAndExitPageMode()
	printer.LineFeed(5)
	printer.CutPaper(settingPrinterInfo.IsEnableSound())

	//
	return printer.GetOrderData()
}
