// Package template 提供打印模板相关功能
package template

import (
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/pkg/utils"
)

// rechargeSunmiPrinterTemplate 图片订单打印模板
type rechargeSunmiPrinterTemplate struct {
	base *printerTemplate
}

// NewRechargeSunmiPrinterTemplate 创建新的图片订单打印模板
func NewRechargeSunmiTemplate(
	base *printerTemplate,
) *rechargeSunmiPrinterTemplate {
	return &rechargeSunmiPrinterTemplate{
		base: base,
	}
}

// GetPrintContent 图片打印
func (t *rechargeSunmiPrinterTemplate) GetPrintContent(
	settingPrinterInfo settingResp.PrinterInfo,
	order model.MemberRechargeOrder,
) string {
	printerType := settingPrinterInfo.PrinterType
	// 日历
	payTime := t.base.FormatUnixTimeDefault(order.PaymentTime)

	// 收银员名称
	cashierName := utils.IfString(order.Staff.RealName != "", order.Staff.RealName, order.Staff.Username)

	// 是否自己打印
	isOneself := printerType != constant.PrinterTypeSunmiLan && printerType != constant.PrinterTypeSunmiCloud

	/* *
	* 打印模版
	 */
	printer := pkg.NewPrinter(567)
	printer.SetLineSpacing(20)
	printer.SetAlignment(pkg.AlignLeft)
	printer.AppendText(t.base.Translate("充值单"))
	printer.LineFeed(2)
	printer.SetAlignment(pkg.AlignCenter)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(t.base.StoreSetting.Name)
	printer.SetLineSpacing(30)
	printer.LineFeed(1)
	//
	printer.SetLineSpacing(45)
	printer.SetPrintModes(false, false, false)
	printer.SetCharacterSize(1, 1)
	printer.LineFeed(1)
	//
	printer.SetupColumns(
		[]int{320, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	printer.PrintInColumns(t.base.Translate("收银员"), cashierName)
	printer.PrintInColumns(t.base.Translate("时间"), payTime)
	printer.PrintInColumns(t.base.Translate("充值前"), t.base.GetPriceAndUnit(order.Balance))
	printer.PrintInColumns(t.base.Translate("本次充值"), t.base.GetPriceAndUnit(order.RechargeAmount))
	printer.PrintInColumns(t.base.Translate("赠送金额"), t.base.GetPriceAndUnit(order.GiftAmount))
	printer.PrintInColumns(t.base.Translate("赠送积分"), t.base.Amount(order.GiftPoint))
	printer.PrintInColumns(t.base.Translate("充值后"), t.base.GetPriceAndUnit(order.BalanceRecharged))
	printer.SetLineSpacing(35)
	// 退款
	if order.RefundMoney > 0 {
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("退款"), t.base.GetPriceAndUnit(order.RefundMoney))
	}
	// 合计應收：
	printer.AppendText("------------------------------------------------")
	printer.LineFeed(1)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 1)
	printer.SetLineSpacing(60)
	printer.SetupColumns(
		[]int{380, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	printer.PrintInColumns(t.base.Translate("合计应收："), t.base.GetPriceAndUnit(order.GetReceivableAmount()))
	printer.SetCharacterSize(1, 1)
	printer.SetPrintModes(false, false, false)
	// 支付方式
	printer.SetupColumns(
		[]int{320, pkg.AlignLeft, 0},
		[]int{0, pkg.AlignRight, 0},
	)
	printer.AppendText("------------------------------------------------")
	if isOneself {
		printer.SetLineSpacing(20)
	} else {
		printer.SetLineSpacing(40)
	}
	printer.LineFeed(1)
	printer.SetLineSpacing(45)
	for _, paymentOrder := range order.PaymentOrders {
		printer.PrintInColumns(t.base.Translate("支付方式"), paymentOrder.PaymentMethod.GetName())
		printer.PrintInColumns(t.base.Translate("实收金额"), t.base.GetPriceAndUnit(paymentOrder.Amount))
		if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			printer.PrintInColumns(t.base.Translate("找零"), t.base.Amount(order.ChargeDue))
		}
	}

	// Print and exit page mode
	printer.RestoreDefaultLineSpacing()
	printer.LineFeed()
	printer.PrintAndExitPageMode()
	printer.LineFeed(5)
	printer.CutPaper(false)

	//
	return printer.GetOrderData()
}
