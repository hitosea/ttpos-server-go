// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// handoverSunmiTemplate 图片订单打印模板
type handoverSunmiTemplate struct {
	base *printerTemplate
}

// NewHandoverSunmiTemplate 创建新的图片订单打印模板
func NewHandoverSunmiTemplate(
	base *printerTemplate,
) *handoverSunmiTemplate {
	return &handoverSunmiTemplate{
		base: base,
	}
}

// GetPrintContent 图片打印
func (t *handoverSunmiTemplate) GetPrintContent(
	printerInfo settingResp.PrinterInfo,
	temp int,
	log *model.StaffShiftLog,
	businessData *business_data_resp.BusinessDataAll,
	openMoneybox bool,
) string {
	printerType := printerInfo.PrinterType
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
	// 是否自己打印
	isOneself := printerType != PrinterTypeSunmiLan && printerType != PrinterTypeSunmiCloud
	lineSpacing := utils.IfInt(!isOneself, 20, 20)

	// 收银员名称
	cashierName := utils.IfString(log.Staff.RealName != "", log.Staff.RealName, log.Staff.Username)

	//  创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.LineFeed()
	printer.SetAlignment(pkg.AlignCenter)
	printer.AppendText(t.base.StoreSetting.Name + "\n")

	// 模版二
	if temp == 2 || temp == 3 {
		printer.SetPrintModes(true, true, false)
		printer.SetLineSpacing(34)
		printer.LineFeed()
		printer.AppendText(t.base.Translate("交班单"))
		printer.LineFeed(2)
		printer.SetPrintModes(false, false, false)
		printer.AppendText(fmt.Sprintf("%s %s %s", startTime, t.base.Translate("至"), endTime))
		printer.LineFeed(2)
		printer.SetLineSpacing(lineSpacing)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{220, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("当班编号"), log.ShiftNo)
		printer.LineFeed()
		printer.PrintInColumns(t.base.Translate("交班人"), cashierName)
		printer.LineFeed()
		// 营业数据
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("总销售额"), t.base.GetPriceAndUnit(businessData.TotalSales))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("实收金额"), t.base.GetPriceAndUnit(businessData.TotalReceivedPrice))
		printer.LineFeed(1)
		// 支付方式
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetupColumns(
			[]int{utils.IfInt(isTrThEn, 240, 280), pkg.AlignLeft, 0},
			[]int{utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 200, 180), utils.IfInt(t.base.Lang == "my", 180, 120)), pkg.AlignLeft, 0},
			[]int{0, utils.IfInt(t.base.Lang == "my", pkg.AlignLeft, pkg.AlignRight), 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(t.base.Translate("支付方式"), t.base.Translate("订单数"), t.base.Translate("金额"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{120, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		totalPayPrice := decimal.NewFromFloat(0)
		for _, income := range businessData.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				printer.PrintInColumns(income.Name, fmt.Sprintf("%d", income.OrderNum), t.base.GetPriceAndUnit(income.Amount))
				printer.LineFeed()
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			printer.PrintInColumns(t.base.Translate("总金额"), "", t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()))
			printer.LineFeed(1)
		}
		// 其他费用
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetupColumns(
			[]int{utils.IfInt(t.base.Lang == "my", 400, 320), pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("原商品金额"), t.base.GetPriceAndUnit(businessData.TotalProductPrice))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("支付手续费"), t.base.GetPriceAndUnit(businessData.TotalPayFeeMoney))
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("服务费"), t.base.GetPriceAndUnit(businessData.TotalServiceMoney))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("税费"), t.base.GetPriceAndUnit(businessData.TotalTaxMoney))
		printer.LineFeed(1)
		// 优惠折扣
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.PrintInColumns(t.base.Translate("优惠折扣"), t.base.GetPriceAndUnit(businessData.TotalDiscountMoney))
		printer.LineFeed(1)
		if isOpenBalance || businessData.TotalUserDiscountMoney > 0 {
			printer.PrintInColumns(t.base.Translate("会员折扣"), t.base.GetPriceAndUnit(businessData.TotalUserDiscountMoney))
			printer.LineFeed(1)
		}
		printer.PrintInColumns(t.base.Translate("赠菜金额"), t.base.GetPriceAndUnit(businessData.TotalGiveProductPrice))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("免单金额"), t.base.GetPriceAndUnit(businessData.TotalFreeOrderPrice))
		printer.LineFeed(1)
		// 退款
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.PrintInColumns(t.base.Translate("退款金额"), t.base.GetPriceAndUnit(businessData.TotalRefundMoney))
		printer.LineFeed(1)
		printer.AppendText("------------------------------------------------")
		// 异常信息
		if temp == 3 {
			printer.SetupColumns(
				[]int{510, pkg.AlignLeft, 0},
				[]int{0, pkg.AlignRight, 0},
			)
			printer.LineFeed(2)
			printer.PrintInColumns(t.base.Translate("退菜次数"), t.base.Amount(float64(businessData.AbnormalData.RefundProductTimes)))
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("退款次数"), t.base.Amount(float64(businessData.AbnormalData.RefundTimes)))
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("反结账次数"), t.base.Amount(float64(businessData.AbnormalData.ReverseSettleTimes)))
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("赠菜次数"), t.base.Amount(float64(businessData.AbnormalData.ProductFreeTimes)))
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("免单次数"), t.base.Amount(float64(businessData.AbnormalData.FreeOrderTimes)))
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("转菜次数"), t.base.Amount(float64(businessData.AbnormalData.ProductMoveTimes)))
			if t.base.Lang != "my" || !isOneself {
				printer.LineFeed(1)
			}
			if t.base.Lang == "my" {
				printer.SetLineSpacing(50)
			}
			printer.PrintInColumns(t.base.Translate("单品改价次数"), t.base.Amount(float64(businessData.AbnormalData.ChangePriceTimes)))
			printer.SetLineSpacing(lineSpacing)
			if t.base.Lang != "my" || !isOneself {
				printer.LineFeed(1)
			}
			if t.base.Lang == "my" {
				printer.SetLineSpacing(50)
			}
			printer.PrintInColumns(t.base.Translate("整单改价次数"), t.base.Amount(float64(businessData.AbnormalData.ChangeOrderPriceTimes)))
			printer.SetLineSpacing(lineSpacing)
			if t.base.Lang != "my" || !isOneself {
				printer.LineFeed(1)
			}
			if t.base.Lang == "my" {
				printer.SetLineSpacing(50)
			}
			printer.PrintInColumns(t.base.Translate("整单折扣次数"), t.base.Amount(float64(businessData.AbnormalData.DiscountOrderTimes)))
			printer.SetLineSpacing(lineSpacing)
			if t.base.Lang != "my" || !isOneself {
				printer.LineFeed(1)
			}
			if t.base.Lang == "my" {
				printer.SetLineSpacing(50)
			}
			printer.PrintInColumns(t.base.Translate("整单抹零次数"), t.base.Amount(float64(businessData.AbnormalData.RoundOrderTimes)))
			printer.SetLineSpacing(lineSpacing)
			printer.LineFeed(1)
			printer.SetLineSpacing(lineSpacing)
			printer.AppendText("------------------------------------------------")
		}
		//
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		// 会员充值
		if isOpenBalance || businessData.MemberData.RechargeAmount > 0 {
			printer.LineFeed(2)
			printer.SetAlignment(pkg.AlignCenter)
			printer.SetPrintModes(true, false, false)
			printer.PrintInColumns(t.base.Translate("会员数据"))
			printer.SetPrintModes(false, false, false)
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignLeft)
			printer.PrintInColumns(t.base.Translate("充值金额"), t.base.GetPriceAndUnit(businessData.MemberData.RechargeAmount))
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("赠送金额"), t.base.GetPriceAndUnit(businessData.MemberData.GiftMoney))
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("赠送积分"), t.base.Amount(float64(businessData.MemberData.GiftPoints)))
			printer.LineFeed(1)
			printer.AppendText("------------------------------------------------")
		}
		// 合计
		printer.LineFeed(2)
		printer.PrintInColumns(t.base.Translate("所有订单数"), t.base.Amount(float64(businessData.TotalOrderNum)))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("取消订单数"), t.base.Amount(float64(businessData.TotalCancelOrderNum)))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("人数"), t.base.Amount(float64(businessData.TotalPeopleNum)))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("取消订单金额"), t.base.GetPriceAndUnit(businessData.TotalCancelOrderAmount))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.AvgOrderPrice))
		printer.LineFeed(1)
		// 高峰时间
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetPrintModes(true, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "th", 180, 240), 280), pkg.AlignLeft, 0},
			[]int{utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 200, 180), utils.IfInt(t.base.Lang == "my", 180, 120)), pkg.AlignLeft, 0},
			[]int{0, utils.IfInt(t.base.Lang == "my", pkg.AlignLeft, pkg.AlignRight), 0},
		)
		printer.PrintInColumns(t.base.Translate("高峰时间"), t.base.Translate("订单数"), t.base.Translate("订单金额"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{270, pkg.AlignLeft, 0},
			[]int{96, pkg.AlignCenter, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		for _, peak := range businessData.PeakHourList {
			printer.PrintInColumns(peak.TimePeriod, fmt.Sprintf("%d", peak.OrderNum), t.base.GetPriceAndUnit(peak.Amount))
			printer.LineFeed(1)
		}
		// 分类列表
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{96, pkg.AlignCenter, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(t.base.Translate("分类"), t.base.Translate("数量"), t.base.Translate("小计"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed()
		for _, category := range businessData.CategoryList {
			printer.PrintInColumns(category.Name, t.base.FloatToString(category.SalesNum), t.base.GetPriceAndUnit(category.Prices))
			printer.LineFeed()
		}
		// 汇总
		printer.SetupColumns(
			[]int{utils.IfInt(t.base.Lang == "en", 430, 400), pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("上一班遗留备用金"), t.base.GetPriceAndUnit(log.PreviousShiftCash))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("中途存入现金"), t.base.GetPriceAndUnit(log.DepositCash))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("中途取出现金"), t.base.GetPriceAndUnit(log.WithdrawCash))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("本班取出现金"), t.base.GetPriceAndUnit(log.CashTakenOut))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("本班遗留备用金"), t.base.GetPriceAndUnit(log.CashLeft))
		printer.SetLineSpacing(lineSpacing)
	} else {
		// 模版 一
		printer.LineFeed()
		printer.SetPrintModes(true, true, false)
		printer.SetLineSpacing(50)
		printer.AppendText(t.base.Translate("交班单"))
		printer.LineFeed(1)
		// 交班信息
		printer.SetLineSpacing(lineSpacing)
		printer.SetPrintModes(false, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{220, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("当班编号"), log.ShiftNo)
		printer.LineFeed()
		printer.PrintInColumns(t.base.Translate("交班人"), cashierName)
		printer.LineFeed()
		printer.PrintInColumns(t.base.Translate("当班时间"), startTime+" "+t.base.Translate("至"))
		printer.SetLineSpacing(lineSpacing - 6)
		printer.LineFeed()
		printer.SetAlignment(pkg.AlignRight)
		printer.AppendText(endTime)
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed(2)
		// 营业数据
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetAlignment(pkg.AlignLeft)
		printer.PrintInColumns(t.base.Translate("总销售额"), t.base.GetPriceAndUnit(businessData.TotalSales))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("原商品金额"), t.base.GetPriceAndUnit(businessData.TotalProductPrice))
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{utils.IfInt(t.base.Lang == "my", 400, 320), pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("支付手续费"), t.base.GetPriceAndUnit(businessData.TotalPayFeeMoney))
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("服务费"), t.base.GetPriceAndUnit(businessData.TotalServiceMoney))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("税费"), t.base.GetPriceAndUnit(businessData.TotalTaxMoney))
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.PrintInColumns(t.base.Translate("商品数量"), utils.FormatFloat(businessData.TotalProductNum))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("优惠折扣"), t.base.GetPriceAndUnit(businessData.TotalDiscountMoney))
		printer.LineFeed(1)
		if isOpenBalance || businessData.TotalUserDiscountMoney > 0 {
			printer.PrintInColumns(t.base.Translate("会员折扣"), t.base.GetPriceAndUnit(businessData.TotalUserDiscountMoney))
			printer.LineFeed(1)
		}
		printer.PrintInColumns(t.base.Translate("退款金额"), t.base.GetPriceAndUnit(businessData.TotalRefundMoney))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("赠菜金额"), t.base.GetPriceAndUnit(businessData.TotalGiveProductPrice))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("免单金额"), t.base.GetPriceAndUnit(businessData.TotalFreeOrderPrice))
		printer.LineFeed(1)
		printer.SetPrintModes(true, true, false)
		printer.SetCharacterSize(2, 1)
		printer.PrintInColumns(t.base.Translate("实收金额"), t.base.GetPriceAndUnit(businessData.TotalReceivedPrice))
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
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
			printer.PrintInColumns(t.base.Translate("合计"), t.base.GetPriceAndUnit(percentage.TotalPrice))
			printer.LineFeed(1)
			printer.SetAlignment(pkg.AlignRight)
			if t.base.Lang == "ja" {
				printer.AppendText(fmt.Sprintf("(%s%s)", t.base.Translate("其中消费税"), t.base.GetPriceAndUnit(percentage.ConsumptionTax)))
			} else {
				printer.AppendText(fmt.Sprintf("(%s%s)", t.base.Translate("其中VAT"), t.base.GetPriceAndUnit(percentage.ConsumptionTax)))
			}
			printer.LineFeed(2)
		}
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		// 合计
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("合计"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(2)
		printer.PrintInColumns(t.base.Translate("所有订单数"), strconv.FormatInt(int64(businessData.TotalOrderNum), 10))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("取消订单数"), strconv.FormatInt(int64(businessData.TotalCancelOrderNum), 10))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("桌数"), strconv.FormatInt(int64(businessData.TotalTableNum), 10))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("人数"), strconv.FormatInt(int64(businessData.TotalPeopleNum), 10))
		printer.LineFeed(1)
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("最小/大订单金额"), fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.MinOrderPrice), t.base.GetPriceAndUnit(businessData.MaxOrderPrice)))
		if t.base.Lang == "my" {
			printer.SetLineSpacing(lineSpacing)
		}
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("取消订单金额"), t.base.GetPriceAndUnit(businessData.TotalCancelOrderAmount))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.AvgOrderPrice))
		printer.LineFeed(2)
		// 桌台方式
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("桌台方式"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{380, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("订单数（桌数）"), fmt.Sprintf("%.0f", float64(businessData.AllTableOrderNum)))
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("人数"), fmt.Sprintf("%.0f", float64(businessData.AllTablePeopleNum)))
		printer.LineFeed(1)
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("最小/大订单金额"), fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.AllTableMinOrderPrice), t.base.GetPriceAndUnit(businessData.AllTableMaxOrderPrice)))
		if t.base.Lang == "my" {
			printer.SetLineSpacing(lineSpacing)
		}
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.AllTableAvgOrderPrice))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("人均"), t.base.GetPriceAndUnit(businessData.AllTablePeopleAvg))
		printer.LineFeed(2)
		// 收银方式
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("点餐方式"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignLeft)
		printer.PrintInColumns(t.base.Translate("订单数"), fmt.Sprintf("%.0f", float64(businessData.AllCashierOrderNum)))
		printer.LineFeed(1)
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("最小/大订单金额"), fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.AllCashierMinOrderPrice), t.base.GetPriceAndUnit(businessData.AllCashierMaxOrderPrice)))
		if t.base.Lang == "my" {
			printer.SetLineSpacing(lineSpacing)
		}
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.AllCashierAvgOrderPrice))
		printer.LineFeed(1)
		// 支付方式
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetupColumns(
			[]int{utils.IfInt(isTrThEn, 240, 280), pkg.AlignLeft, 0},
			[]int{utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 200, 180), utils.IfInt(t.base.Lang == "my", 180, 120)), pkg.AlignLeft, 0},
			[]int{0, utils.IfInt(t.base.Lang == "my", pkg.AlignLeft, pkg.AlignRight), 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(t.base.Translate("支付方式"), t.base.Translate("订单数"), t.base.Translate("金额"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{120, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		totalPayPrice := decimal.NewFromFloat(0)
		for _, income := range businessData.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				printer.PrintInColumns(income.Name, fmt.Sprintf("%d", income.OrderNum), t.base.GetPriceAndUnit(income.Amount))
				printer.LineFeed()
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			printer.PrintInColumns(t.base.Translate("总金额"), "", t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()))
			printer.LineFeed(1)
		}
		// 高峰时间
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetPrintModes(true, false, false)
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "th", 180, 240), 280), pkg.AlignLeft, 0},
			[]int{utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 200, 180), utils.IfInt(t.base.Lang == "my", 180, 120)), pkg.AlignLeft, 0},
			[]int{0, utils.IfInt(t.base.Lang == "my", pkg.AlignLeft, pkg.AlignRight), 0},
		)
		printer.PrintInColumns(t.base.Translate("高峰时间"), t.base.Translate("订单数"), t.base.Translate("订单金额"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{270, pkg.AlignLeft, 0},
			[]int{96, pkg.AlignCenter, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		for _, peak := range businessData.PeakHourList {
			printer.PrintInColumns(peak.TimePeriod, fmt.Sprintf("%d", peak.OrderNum), t.base.GetPriceAndUnit(peak.Amount))
			printer.LineFeed(1)
		}
		// 分类列表
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		printer.SetAlignment(pkg.AlignLeft)
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{96, pkg.AlignCenter, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(t.base.Translate("分类"), t.base.Translate("数量"), t.base.Translate("小计"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		for _, category := range businessData.CategoryList {
			printer.PrintInColumns(category.Name, t.base.FloatToString(category.SalesNum), t.base.GetPriceAndUnit(category.Prices))
			printer.LineFeed(1)
		}
		// 汇总
		printer.SetupColumns(
			[]int{400, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(
			utils.IfString(t.base.Lang == "en", "Actual amount received this shift", t.base.Translate("本班实收金额")),
			t.base.GetPriceAndUnit(businessData.TotalReceivedPrice),
		)
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("上一班遗留备用金"), t.base.GetPriceAndUnit(log.PreviousShiftCash))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("中途存入现金"), t.base.GetPriceAndUnit(log.DepositCash))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("中途取出现金"), t.base.GetPriceAndUnit(log.WithdrawCash))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("本班取出现金"), t.base.GetPriceAndUnit(log.CashTakenOut))
		printer.SetLineSpacing(lineSpacing)
		printer.LineFeed()
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("本班遗留备用金"), t.base.GetPriceAndUnit(log.CashLeft))
	}

	//
	printer.LineFeed(4)
	printer.PrintAndExitPageMode()
	printer.LineFeed(4)
	printer.CutPaper(false)
	// 打开钱箱
	if openMoneybox {
		printer.AppendText("\x1B\x70\x00\x19\xFA")
	}
	//
	return printer.GetOrderData()
}
