// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// handoverCompaxTemplate 图片订单打印模板
type handoverCompaxTemplate struct {
	base *printerTemplate
}

// NewHandoverCompaxTemplate 创建新的图片订单打印模板
func NewHandoverCompaxTemplate(
	base *printerTemplate,
) *handoverCompaxTemplate {
	return &handoverCompaxTemplate{
		base: base,
	}
}

// GetPrintContent 图片打印
func (t *handoverCompaxTemplate) GetPrintContent(
	printerInfo settingResp.PrinterInfo,
	temp int,
	log *model.StaffShiftLog,
	businessData *business_data_resp.BusinessDataAll,
	openMoneybox bool,
) string {
	// 店铺设置
	companySetting, _ := t.base.Setting.GetCompanySetting(t.base.Ctx)
	paymentSetting, _ := t.base.Setting.GetPaymentSetting(t.base.Ctx, companySetting)
	isOpenBalance := paymentSetting.IsBalance == "1"
	// 日历
	startTime := t.base.FormatUnixTimeDefault(log.ShiftStartTime)
	endTime := t.base.FormatUnixTimeDefault(log.ShiftEndTime)
	if log.ShiftEndTime == 0 {
		endTime = t.base.FormatUnixTimeDefault(time.Now().Unix())
	}
	// 判断是否是土耳其语、泰语、英语
	isTrThEn := t.base.Lang == "tr" || t.base.Lang == "th" || t.base.Lang == "en"
	// 收银员名称
	cashierName := utils.IfString(log.Staff.RealName != "", log.Staff.RealName, log.Staff.Username)

	// 宽度
	width := 48
	leftWidth := 27
	centerWidth := 12
	rightWidth := 19
	differenceWidth := 0
	if t.base.CurrencyUnit == "¥" || t.base.CurrencyUnit == "\xC2\xA5" {
		differenceWidth = 1
	}

	//  创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.SetAlignment(pkg.AlignCenter)
	printer.AppendText(t.base.StoreSetting.Name)
	printer.LineFeed(2)

	// 模版二
	if temp == 2 || temp == 3 {
		printer.SetLineSpacing(28)
		printer.SetCharacterSize(2, 2)
		printer.AppendText(t.base.Translate("交班单"))
		printer.SetCharacterSize(1, 1)
		printer.LineFeed(2)
		printer.AppendText(fmt.Sprintf("%s %s %s", startTime, t.base.Translate("至"), endTime))
		printer.LineFeed(3)
		printer.SetLineSpacing(42)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("当班编号"), "", log.ShiftNo, width))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("交班人"), "", cashierName, width))
		printer.LineFeed()
		// 营业数据
		printer.AppendText(t.base.PrintText(t.base.Translate("总销售额"), "", t.base.GetPriceAndUnit(businessData.TotalSales), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(businessData.TotalReceivedPrice), width-differenceWidth))
		printer.LineFeed(1)
		// 支付方式
		printer.AppendText("------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), t.base.Translate("订单数"), t.base.Translate("金额"), width, 26-utils.IfInt(isTrThEn, 4, 0), 20, 16))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		totalPayPrice := decimal.NewFromFloat(0)
		for key, income := range businessData.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				printer.AppendText(t.base.PrintText(income.Name, fmt.Sprintf("%d", income.OrderNum), t.base.GetPriceAndUnit(income.Amount), width-differenceWidth, 26, 10, 18))
				if key != len(businessData.PaymentMethodIncomes)-1 {
					printer.LineFeed(1)
				}
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			printer.AppendText(t.base.PrintText(t.base.Translate("总金额"), "", t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()), width-differenceWidth, 26, 10, 18))
			printer.LineFeed(1)
		}
		// 其他费用
		printer.AppendText("\n------------------------------------------------")
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("原商品金额"), "", t.base.GetPriceAndUnit(businessData.TotalProductPrice), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("支付手续费"), "", t.base.GetPriceAndUnit(businessData.TotalPayFeeMoney), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("服务费"), "", t.base.GetPriceAndUnit(businessData.TotalServiceMoney), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("税费"), "", t.base.GetPriceAndUnit(businessData.TotalTaxMoney), width-differenceWidth))
		printer.LineFeed(1)
		// 优惠折扣
		printer.AppendText("------------------------------------------------")
		printer.AppendText(t.base.PrintText(t.base.Translate("优惠折扣"), "", t.base.GetPriceAndUnit(businessData.TotalDiscountMoney), width-differenceWidth))
		printer.LineFeed(1)
		if isOpenBalance || businessData.TotalUserDiscountMoney > 0 {
			printer.AppendText(t.base.PrintText(t.base.Translate("会员折扣"), "", t.base.GetPriceAndUnit(businessData.TotalUserDiscountMoney), width-differenceWidth))
			printer.LineFeed(1)
		}
		printer.AppendText(t.base.PrintText(t.base.Translate("赠菜金额"), "", t.base.GetPriceAndUnit(businessData.TotalGiveProductPrice), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("免单金额"), "", t.base.GetPriceAndUnit(businessData.TotalFreeOrderPrice), width-differenceWidth))
		printer.LineFeed(1)
		// 退款
		printer.AppendText("------------------------------------------------")
		printer.AppendText(t.base.PrintText(t.base.Translate("退款金额"), "", t.base.GetPriceAndUnit(businessData.TotalRefundMoney), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText("------------------------------------------------")
		// 异常信息
		if temp == 3 {
			printer.AppendText(t.base.PrintText(t.base.Translate("退菜次数"), "", t.base.Amount(float64(businessData.AbnormalData.RefundProductTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("退款次数"), "", t.base.Amount(float64(businessData.AbnormalData.RefundTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("反结账次数"), "", t.base.Amount(float64(businessData.AbnormalData.ReverseSettleTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("赠菜次数"), "", t.base.Amount(float64(businessData.AbnormalData.ProductFreeTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("免单次数"), "", t.base.Amount(float64(businessData.AbnormalData.FreeOrderTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("转菜次数"), "", t.base.Amount(float64(businessData.AbnormalData.ProductMoveTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("单品改价次数"), "", t.base.Amount(float64(businessData.AbnormalData.ChangePriceTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("整单改价次数"), "", t.base.Amount(float64(businessData.AbnormalData.ChangeOrderPriceTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("整单折扣次数"), "", t.base.Amount(float64(businessData.AbnormalData.DiscountOrderTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("整单抹零次数"), "", t.base.Amount(float64(businessData.AbnormalData.RoundOrderTimes)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText("------------------------------------------------")
		}
		// 会员充值
		if isOpenBalance || businessData.MemberData.RechargeAmount > 0 {
			printer.SetAlignment(pkg.AlignCenter)
			printer.SetPrintModes(true, false, false)
			printer.AppendText(t.base.Translate("会员数据"))
			printer.SetPrintModes(false, false, false)
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText(t.base.PrintText(t.base.Translate("充值金额"), "", t.base.GetPriceAndUnit(businessData.MemberData.RechargeAmount), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("赠送金额"), "", t.base.GetPriceAndUnit(businessData.MemberData.GiftMoney), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText(t.base.PrintText(t.base.Translate("赠送积分"), "", t.base.Amount(float64(businessData.MemberData.GiftPoints)), width-differenceWidth))
			printer.LineFeed(1)
			printer.AppendText("------------------------------------------------")
		}
		// 合计
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("所有订单数"), "", float64(businessData.TotalOrderNum), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("取消订单数"), "", fmt.Sprintf("%.0f", float64(businessData.TotalCancelOrderNum)), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("人数"), "", float64(businessData.TotalPeopleNum), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("取消订单金额"), "", t.base.GetPriceAndUnit(businessData.TotalCancelOrderAmount), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.AvgOrderPrice), width-differenceWidth))
		printer.LineFeed(1)
		// 高峰时间
		printer.AppendText("------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("高峰时间"), t.base.Translate("订单数"), t.base.Translate("订单金额"), width, 24-utils.IfInt(isTrThEn, 4, 0), 20, 18))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed()
		for key, peak := range businessData.PeakHourList {
			if key == len(businessData.PeakHourList)-1 {
				printer.SetPrintModes(true, false, false)
			}
			printer.AppendText(t.base.PrintText(peak.TimePeriod, fmt.Sprintf("%d", peak.OrderNum), t.base.GetPriceAndUnit(peak.Amount), width, 26, 10, 18))
		}
		// 分类列表
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText("\n------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("分类"), t.base.Translate("数量"), t.base.Translate("小计"), width, 26-utils.IfInt(t.base.Lang == "tr", 2, 0), 20, 14))
		printer.SetPrintModes(false, false, false)
		printer.SetLineSpacing(40)
		for _, category := range businessData.CategoryList {
			printer.AppendText(t.base.PrintText(category.Name, t.base.FloatToString(category.SalesNum), t.base.GetPriceAndUnit(category.Prices), width-differenceWidth, leftWidth, centerWidth, rightWidth))
			printer.LineFeed()
			printer.SetLineSpacing(45)
		}
		// 汇总
		printer.AppendText("------------------------------------------------\n")
		printer.AppendText(t.base.PrintText(t.base.Translate("上一班遗留备用金"), "", t.base.GetPriceAndUnit(log.PreviousShiftCash), width-differenceWidth, leftWidth+6))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("中途存入现金"), "", t.base.GetPriceAndUnit(log.DepositCash), width-differenceWidth, leftWidth))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("中途取出现金"), "", t.base.GetPriceAndUnit(log.WithdrawCash), width-differenceWidth, leftWidth))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("本班取出现金"), "", t.base.GetPriceAndUnit(log.CashTakenOut), width-differenceWidth, leftWidth+6))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("本班遗留备用金"), "", t.base.GetPriceAndUnit(log.CashLeft), width-differenceWidth, leftWidth+6))
	} else {
		// 模版 一
		printer.SetLineSpacing(42)
		printer.SetCharacterSize(2, 2)
		printer.AppendText(t.base.Translate("交班单"))
		printer.SetCharacterSize(1, 1)
		printer.LineFeed(2)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("当班编号"), "", log.ShiftNo, width))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("交班人"), "", cashierName, width))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("当班时间"), "", startTime+" "+t.base.Translate("至"), width))
		printer.LineFeed()
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(endTime)
		printer.LineFeed()
		// 营业数据
		printer.AppendText(t.base.PrintText(t.base.Translate("总销售额"), "", t.base.GetPriceAndUnit(businessData.TotalSales), width-differenceWidth, leftWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("原商品金额"), "", t.base.GetPriceAndUnit(businessData.TotalProductPrice), width-differenceWidth, leftWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("支付手续费"), "", t.base.GetPriceAndUnit(businessData.TotalPayFeeMoney), width-differenceWidth, leftWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("服务费"), "", t.base.GetPriceAndUnit(businessData.TotalServiceMoney), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("税费"), "", t.base.GetPriceAndUnit(businessData.TotalTaxMoney), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("商品数量"), "", utils.FormatFloat(businessData.TotalProductNum), width-utils.IfInt(t.base.Lang == "th", 1, 0)))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("优惠折扣"), "", t.base.GetPriceAndUnit(businessData.TotalDiscountMoney), width-differenceWidth))
		printer.LineFeed(1)
		if isOpenBalance || businessData.TotalUserDiscountMoney > 0 {
			printer.AppendText(t.base.PrintText(t.base.Translate("会员折扣"), "", t.base.GetPriceAndUnit(businessData.TotalUserDiscountMoney), width-differenceWidth))
			printer.LineFeed(1)
		}
		printer.AppendText(t.base.PrintText(t.base.Translate("退款金额"), "", t.base.GetPriceAndUnit(businessData.TotalRefundMoney), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("免单金额"), "", t.base.GetPriceAndUnit(businessData.TotalFreeOrderPrice), width-differenceWidth))
		printer.LineFeed(1)
		printer.SetPrintModes(true, true, false)
		printer.SetCharacterSize(2, 1)
		printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(businessData.TotalReceivedPrice), width-differenceWidth))
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.AppendText("\n------------------------------------------------")
		printer.LineFeed(1)
		// 税收百分比对象列表
		for _, percentage := range businessData.PercentageList {
			printer.SetAlignment(pkg.AlignLeft)
			printer.SetPrintModes(true, false, false)
			if t.base.Lang == "ja" {
				printer.AppendText(fmt.Sprintf("%s%s%s", t.base.Amount(percentage.TaxRate), "%", t.base.Translate("的对象")))
			} else {
				printer.AppendText(fmt.Sprintf("VAT (%s%%)", t.base.Amount(percentage.TaxRate)))
			}
			printer.SetPrintModes(false, false, false)
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText(t.base.PrintText(t.base.Translate("合计"), "", t.base.GetPriceAndUnit(percentage.TotalPrice), width))
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignRight)
			if t.base.Lang == "ja" {
				printer.AppendText(fmt.Sprintf("(%s%s)", t.base.Translate("其中消费税"), t.base.GetPriceAndUnit(percentage.ConsumptionTax)))
			} else {
				printer.AppendText(fmt.Sprintf("(%s%s)", t.base.Translate("其中VAT"), t.base.GetPriceAndUnit(percentage.ConsumptionTax)))
			}
			printer.LineFeed(1)
		}
		printer.AppendText("------------------------------------------------")
		// 合计
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("合计"))
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("所有订单数"), "", float64(businessData.TotalOrderNum), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("取消订单数"), "", fmt.Sprintf("%.0f", float64(businessData.TotalCancelOrderNum)), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("桌数"), "", float64(businessData.TotalTableNum), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("人数"), "", float64(businessData.TotalPeopleNum), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("最小/大订单金额"), "", fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.MinOrderPrice), t.base.GetPriceAndUnit(businessData.MaxOrderPrice)), width-differenceWidth*2))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("取消订单金额"), "", t.base.GetPriceAndUnit(businessData.TotalCancelOrderAmount), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.AvgOrderPrice), width-differenceWidth))
		printer.LineFeed(1)
		// 桌台方式
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("桌台方式"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("订单数（桌数）"), "", fmt.Sprintf("%.0f", float64(businessData.AllTableOrderNum)), width))
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("人数"), "", fmt.Sprintf("%.0f", float64(businessData.AllTablePeopleNum)), width))
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("最小/大订单金额"), "", fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.AllTableMinOrderPrice), t.base.GetPriceAndUnit(businessData.AllTableMaxOrderPrice)), width-differenceWidth*2))
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.AllTableAvgOrderPrice), width-differenceWidth))
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("人均"), "", t.base.GetPriceAndUnit(businessData.AllTablePeopleAvg), width-differenceWidth))
		printer.LineFeed(1)
		// 收银方式-店内
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("点餐方式-店内"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("订单数"), "", fmt.Sprintf("%.0f", float64(businessData.AllCashierOrderNum)), width))
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("最小/大订单金额"), "", fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.AllCashierMinOrderPrice), t.base.GetPriceAndUnit(businessData.AllCashierMaxOrderPrice)), width-differenceWidth*2))
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.AllCashierAvgOrderPrice), width-differenceWidth))
		printer.LineFeed(1)
		// 收银方式-外卖
		if businessData.AllTakeawayOrderNum > 0 {
			printer.SetAlignment(pkg.AlignCenter)
			printer.SetPrintModes(true, false, false)
			printer.AppendText(t.base.Translate("点餐方式-外卖"))
			printer.SetPrintModes(false, false, false)
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText(t.base.PrintText(t.base.Translate("订单数"), "", fmt.Sprintf("%.0f", float64(businessData.AllTakeawayOrderNum)), width))
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText(t.base.PrintText(t.base.Translate("最小/大订单金额"), "", fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.AllTakeawayMinOrderPrice), t.base.GetPriceAndUnit(businessData.AllTakeawayMaxOrderPrice)), width-differenceWidth*2))
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.AllTakeawayAvgOrderPrice), width-differenceWidth))
			printer.LineFeed(1)
		}
		// 支付方式
		printer.AppendText("------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), t.base.Translate("订单数"), t.base.Translate("金额"), width, 26-utils.IfInt(isTrThEn, 4, 0), 20, 14))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		totalPayPrice := decimal.NewFromFloat(0)
		for key, income := range businessData.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				printer.AppendText(t.base.PrintText(income.Name, fmt.Sprintf("%d", income.OrderNum), t.base.GetPriceAndUnit(income.Amount), width-differenceWidth, 26, 10, 18))
				if key != len(businessData.PaymentMethodIncomes)-1 {
					printer.LineFeed(1)
				}
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			printer.AppendText(t.base.PrintText(t.base.Translate("总金额"), "", t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()), width-differenceWidth, 26, 10, 18))
			printer.LineFeed(1)
		}
		// 高峰时间
		printer.AppendText("\n------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("高峰时间"), t.base.Translate("订单数"), t.base.Translate("订单金额"), width, 24-utils.IfInt(isTrThEn, 4, 0), 20, 18))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		for _, peak := range businessData.PeakHourList {
			printer.AppendText(t.base.PrintText(peak.TimePeriod, fmt.Sprintf("%d", peak.OrderNum), t.base.GetPriceAndUnit(peak.Amount), width-differenceWidth, 26, 10, 18))
		}
		// 分类列表
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText("\n------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("分类"), t.base.Translate("数量"), t.base.Translate("小计"), width, 26-utils.IfInt(t.base.Lang == "en", 4, 0), 20, 14))
		printer.SetPrintModes(false, false, false)
		printer.SetLineSpacing(40)
		printer.LineFeed()
		for _, category := range businessData.CategoryList {
			printer.AppendText(t.base.PrintText(category.Name, t.base.FloatToString(category.SalesNum), t.base.GetPriceAndUnit(category.Prices), width-differenceWidth, leftWidth, centerWidth, rightWidth))
			printer.LineFeed()
			printer.SetLineSpacing(45)
		}
		// 汇总
		printer.AppendText("------------------------------------------------\n")
		printer.AppendText(t.base.PrintText(t.base.Translate("本班实收金额"), "", t.base.GetPriceAndUnit(businessData.TotalReceivedPrice), width-differenceWidth, leftWidth+6))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("上一班遗留备用金"), "", t.base.GetPriceAndUnit(log.PreviousShiftCash), width-differenceWidth, leftWidth+6))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("中途存入现金"), "", t.base.GetPriceAndUnit(log.DepositCash), width-differenceWidth, leftWidth))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("中途取出现金"), "", t.base.GetPriceAndUnit(log.WithdrawCash), width-differenceWidth, leftWidth))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("本班取出现金"), "", t.base.GetPriceAndUnit(log.CashTakenOut), width-differenceWidth, leftWidth+6))
		printer.LineFeed()
		printer.AppendText(t.base.PrintText(t.base.Translate("本班遗留备用金"), "", t.base.GetPriceAndUnit(log.CashLeft), width-differenceWidth, leftWidth+6))
	}

	//
	printer.LineFeed()
	printer.PrintAndExitPageMode()
	printer.LineFeed(4)
	printer.CutPaper(printerInfo.IsEnableSound())
	// 打开钱箱
	if openMoneybox {
		printer.AppendText("\x10\x14\x01\x00\x01")
	}
	//
	return printer.GetOrderData()
}
