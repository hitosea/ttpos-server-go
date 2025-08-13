// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// handoverImg58mmTemplate 图片订单打印模板 - 58mm适配版
type handoverImg58mmTemplate struct {
	base *printerTemplate
}

// NewHandoverImg58mmTemplate 创建新的图片订单打印模板 - 58mm版本
func NewHandoverImg58mmTemplate(
	base *printerTemplate,
) *handoverImg58mmTemplate {
	return &handoverImg58mmTemplate{
		base: base,
	}
}

// GetPrintContent58mm 图片打印 - 58mm版本
func (t *handoverImg58mmTemplate) GetPrintContent58mm(
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

	//  创建打印机实例 - 适配58mm打印纸
	img := pkg.NewImgFont(384, 50, 0) // 使用384像素宽度确保字节对齐
	// 调整58mm打印纸的文本间距和行高
	img.SetTextSpacing(1.1) // 稍微增加字符间距
	// 确保方向为0，避免旋转问题
	img.Direction = 0
	img.SetTextLineHeight(50)
	img.SetImagePadding(0)
	img.SetAlignment(pkg.AlignCenter)
	img.AppendText(t.base.StoreSetting.Name)
	img.LineFeed(1, 58)
	// 模版二
	if temp == 2 || temp == 3 {
		img.SetFontSize(28)
		img.AppendText(t.base.Translate("交班单"))
		img.SetFontSize(20)
		img.LineFeed(1)
		img.AppendText(fmt.Sprintf("%s %s %s", startTime, t.base.Translate("至"), endTime))
		img.LineFeed(1)
		img.LineFeed(1, 24)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("当班编号"), Width: 200, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: log.ShiftNo, Width: 0, Align: pkg.AlignRight},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("交班人"), Width: 230, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: cashierName, Width: 0, Align: pkg.AlignRight},
		)
		// 营业数据
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("总销售额"), Width: 230, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalSales), Width: 0, Align: pkg.AlignRight},
		)
		img.SetTextLineHeight(40)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 230, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalReceivedPrice), Width: 0, Align: pkg.AlignRight},
		)
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		// 支付方式
		numberWidth := 190
		numberWidth = utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 88, 122), utils.IfInt(t.base.Lang == "my", 122, 81))
		// 支付方式
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("支付方式"), Width: utils.IfInt(t.base.Lang == "sv", 135, utils.IfInt(isTrThEn, 132, 203)), Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: "", Width: utils.IfInt(isTrThEn, 72, utils.IfInt(t.base.Lang == "my", 10, utils.IfInt(t.base.Lang == "sv", 88, 1))), Align: pkg.AlignCenter, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: utils.IfInt(t.base.Lang == "sv", 88, numberWidth), Align: pkg.AlignCenter, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("金额"), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		totalPayPrice := decimal.NewFromFloat(0)
		for key, income := range businessData.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				if key == len(businessData.PaymentMethodIncomes)-1 {
					img.SetTextLineHeight(36)
				}
				img.PrintInColumns(
					pkg.ColumnConfig{Text: income.Name, Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
					pkg.ColumnConfig{Text: fmt.Sprintf("%d", income.OrderNum), Width: 80, Align: pkg.AlignCenter, FontWeight: 1},
					pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(income.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
				)
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			img.LineFeed(1, 10)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("总金额"), Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: "", Width: 80, Align: pkg.AlignCenter, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		// 其他费用
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("原商品金额"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalProductPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("支付手续费"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalPayFeeMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("服务费"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalServiceMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.SetTextLineHeight(36)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("税费"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalTaxMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 优惠折扣
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		if !(isOpenBalance || businessData.TotalUserDiscountMoney > 0) {
			img.SetTextLineHeight(36)
		}
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("优惠折扣"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalDiscountMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		if isOpenBalance || businessData.TotalUserDiscountMoney > 0 {
			img.LineFeed(1, 12)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("会员折扣"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalUserDiscountMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		img.LineFeed(1, 12)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("赠菜金额"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalGiveProductPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.LineFeed(1, 12)
		img.SetTextLineHeight(36)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("免单金额"), Width: 257, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalFreeOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 退款
		img.SetTextLineHeight(42)
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(36)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("退款金额"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalRefundMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.SetTextLineHeight(38)
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		// 异常信息
		if temp == 3 {
			if t.base.Lang == "my" {
				img.SetTextLineHeight(52)
			}
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("退菜次数"), Width: 325, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.RefundProductTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("退款次数"), Width: 325, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.RefundTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("反结账次数"), Width: 325, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.ReverseSettleTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("赠菜次数"), Width: 325, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.ProductFreeTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("免单次数"), Width: 325, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.FreeOrderTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("转菜次数"), Width: 325, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.ProductMoveTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("单品改价次数"), Width: 365, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.ChangePriceTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("整单改价次数"), Width: 365, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.ChangeOrderPriceTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("整单折扣次数"), Width: 338, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.DiscountOrderTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.SetTextLineHeight(38)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("整单抹零次数"), Width: 338, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.AbnormalData.RoundOrderTimes)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.AppendText("-------------------------------------------")
			img.LineFeed(1)
			img.SetTextLineHeight(50)
		}
		// 会员充值
		if isOpenBalance || businessData.MemberData.RechargeAmount > 0 {
			img.SetAlignment(pkg.AlignCenter)
			img.SetFontWeight(2)
			img.AppendText(t.base.Translate("会员数据"))
			img.SetFontWeight(1)
			img.LineFeed(1)
			img.SetAlignment(pkg.AlignLeft)
			//
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("充值金额"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.MemberData.RechargeAmount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("赠送金额"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.MemberData.GiftMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.SetTextLineHeight(38)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("赠送积分"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(businessData.MemberData.GiftPoints)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
			img.AppendText("-------------------------------------------")
			img.LineFeed(1)
			img.SetTextLineHeight(50)
		}
		//
		img.SetTextLineHeight(50)
		// 合计
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("所有订单数"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.TotalOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("取消订单数"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.TotalCancelOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("人数"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.TotalPeopleNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.SetTextLineHeight(36)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("取消订单金额"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalCancelOrderAmount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("平均订单金额"), Width: 216, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.AvgOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 高峰时间
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)

		// 高峰时间
		peakHoursWidth := utils.IfInt(isTrThEn, 162, 203)
		numberWidth = utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 149, 122), utils.IfInt(t.base.Lang == "my", 122, 81))
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("高峰时间"), Width: utils.IfInt(t.base.Lang == "sv", 135, peakHoursWidth), Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: utils.IfInt(t.base.Lang == "sv", 149, numberWidth), Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: utils.IfString(t.base.Lang == "en", "Amount", t.base.Translate("订单金额")), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		for key, peak := range businessData.PeakHourList {
			if key == len(businessData.PeakHourList)-1 {
				img.SetTextLineHeight(36)
			}
			img.PrintInColumns(
				pkg.ColumnConfig{Text: peak.TimePeriod, Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.Amount(float64(peak.OrderNum)), Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(peak.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		// 分类列表
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("分类"), Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("数量"), Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("小计"), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		for key, category := range businessData.CategoryList {
			if key == len(businessData.CategoryList)-1 {
				img.SetTextLineHeight(36)
			}
			img.PrintInColumns(
				pkg.ColumnConfig{Text: category.Name, Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.FloatToString(category.SalesNum), Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(category.Prices), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		// 汇总
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("上一班遗留备用金"), Width: 257, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.PreviousShiftCash), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("中途存入现金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.DepositCash), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("中途取出现金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.WithdrawCash), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("本班取出现金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.CashTakenOut), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("本班遗留备用金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.CashLeft), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
	} else {
		// 模版 一
		img.SetFontSize(28)
		img.AppendText(t.base.Translate("交班单"))
		img.SetFontSize(20)
		img.LineFeed(1)
		img.LineFeed(1, 24)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("当班编号"), Width: 200, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: log.ShiftNo, Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("交班人"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: cashierName, Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.AppendText(t.base.Translate("当班时间"))
		img.SetAlignment(pkg.AlignRight)
		img.AppendText(startTime + " " + t.base.Translate("至"))
		img.LineFeed(1, 32)
		img.AppendText(endTime)
		img.LineFeed(1)
		// 营业数据
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("总销售额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalSales), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("原商品金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalProductPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("支付手续费"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalPayFeeMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("服务费"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalServiceMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("税费"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalTaxMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("商品数量"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: utils.FormatFloat(businessData.TotalProductNum), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("优惠折扣"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalDiscountMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		if isOpenBalance || businessData.TotalUserDiscountMoney > 0 {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("会员折扣"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalUserDiscountMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("退款金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalRefundMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("赠菜金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalGiveProductPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("免单金额"), Width: 257, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalFreeOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.LineFeed(1, 10)
		if t.base.Lang == "my" {
			img.SetTextLineHeight(58)
		}
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 28},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalReceivedPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 28},
		)
		// 税收百分比对象列表
		if len(businessData.PercentageList) > 0 {
			img.AppendText("-------------------------------------------")
			img.SetTextLineHeight(48)
			for _, percentage := range businessData.PercentageList {
				img.SetAlignment(pkg.AlignLeft)
				img.SetFontWeight(2)
				if t.base.Lang == "ja" {
					img.AppendText(fmt.Sprintf("%s%s", t.base.Amount(percentage.TaxRate), "%"))
				} else {
					img.AppendText(fmt.Sprintf("VAT (%s%%)", t.base.Amount(percentage.TaxRate)))
				}
				img.SetFontWeight(1)
				img.LineFeed(1)
				img.PrintInColumns(
					pkg.ColumnConfig{Text: t.base.Translate("合计"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
					pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(percentage.TotalPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
				)
				img.SetAlignment(pkg.AlignRight)
				if t.base.Lang == "ja" {
					img.AppendText(fmt.Sprintf("(%s%s)", t.base.Translate("其中消费税"), t.base.GetPriceAndUnit(percentage.ConsumptionTax)))
				} else {
					img.AppendText(fmt.Sprintf("(%s%s)", t.base.Translate("其中VAT"), t.base.GetPriceAndUnit(percentage.ConsumptionTax)))
				}
				img.LineFeed(1)
			}
		}
		// 合计
		img.AppendText("-------------------------------------------") // 交班单58mm专用分割线，41个短横线
		img.SetTextLineHeight(50)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("合计"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignLeft)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("所有订单数"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.TotalOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("取消订单数"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.TotalCancelOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("桌数"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.TotalTableNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("人数"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.TotalPeopleNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("最小/大订单金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.MinOrderPrice), t.base.GetPriceAndUnit(businessData.MaxOrderPrice)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("取消订单金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalCancelOrderAmount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("平均订单金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.AvgOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 桌台方式
		img.LineFeed(1, 10)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("桌台方式"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单数（桌数）"), Width: 270, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.AllTableOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("人数"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.AllTablePeopleNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("最小/大订单金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.AllTableMinOrderPrice), t.base.GetPriceAndUnit(businessData.AllTableMaxOrderPrice)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("平均订单金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.AllTableAvgOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("人均"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.AllTablePeopleAvg), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 点餐方式
		img.LineFeed(1, 10)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("点餐方式"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.AllCashierOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("最小/大订单金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.AllCashierMinOrderPrice), t.base.GetPriceAndUnit(businessData.AllCashierMaxOrderPrice)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("平均订单金额"), Width: 230, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.AllCashierAvgOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 支付方式
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		numberWidth := 190
		numberWidth = utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 88, 122), utils.IfInt(t.base.Lang == "my", 122, 81))
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("支付方式"), Width: utils.IfInt(t.base.Lang == "sv", 135, utils.IfInt(isTrThEn, 132, 203)), Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: "", Width: utils.IfInt(isTrThEn, 72, utils.IfInt(t.base.Lang == "my", 10, utils.IfInt(t.base.Lang == "sv", 88, 1))), Align: pkg.AlignCenter, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: utils.IfInt(t.base.Lang == "sv", 88, numberWidth), Align: pkg.AlignCenter, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("金额"), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		// 支付方式
		totalPayPrice := decimal.NewFromFloat(0)
		for key, income := range businessData.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				if key == len(businessData.PaymentMethodIncomes)-1 {
					img.SetTextLineHeight(36)
				}
				img.PrintInColumns(
					pkg.ColumnConfig{Text: income.Name, Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
					pkg.ColumnConfig{Text: fmt.Sprintf("%d", income.OrderNum), Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
					pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(income.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
				)
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			img.LineFeed(1, 10)
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("总金额"), Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: "", Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		// 高峰时间
		peakHoursWidth := utils.IfInt(isTrThEn, 162, 203)
		numberWidth = utils.IfInt(isTrThEn, utils.IfInt(t.base.Lang == "en", 149, 122), utils.IfInt(t.base.Lang == "my", 122, 81))
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("高峰时间"), Width: utils.IfInt(t.base.Lang == "sv", 135, peakHoursWidth), Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: utils.IfInt(t.base.Lang == "sv", 149, numberWidth), Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: utils.IfString(t.base.Lang == "en", "Amount", t.base.Translate("订单金额")), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		for key, peak := range businessData.PeakHourList {
			if key == len(businessData.PeakHourList)-1 {
				img.SetTextLineHeight(36)
			}
			img.PrintInColumns(
				pkg.ColumnConfig{Text: peak.TimePeriod, Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: fmt.Sprintf("%d", peak.OrderNum), Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(peak.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		// 分类列表
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("分类"), Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("数量"), Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("小计"), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		for key, category := range businessData.CategoryList {
			if key == len(businessData.CategoryList)-1 {
				img.SetTextLineHeight(36)
			}
			img.PrintInColumns(
				pkg.ColumnConfig{Text: category.Name, Width: 190, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.FloatToString(category.SalesNum), Width: 81, Align: pkg.AlignCenter, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(category.Prices), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		// 汇总
		img.AppendText("-------------------------------------------")
		img.LineFeed(1)
		img.SetTextLineHeight(50)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("本班实收金额"), Width: 257, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.TotalReceivedPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("上一班遗留备用金"), Width: 257, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.PreviousShiftCash), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("中途存入现金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.DepositCash), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("中途取出现金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.WithdrawCash), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("本班取出现金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.CashTakenOut), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("本班遗留备用金"), Width: 243, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(log.CashLeft), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
	}
	//
	img.LineFeed(3)
	//
	openMoneyboxInt := 0
	if openMoneybox {
		openMoneyboxInt = utils.IfInt(t.base.IsSunMi, 2, 1)
	}
	return img.Save("", !t.base.IsSunMi, openMoneyboxInt)
}
