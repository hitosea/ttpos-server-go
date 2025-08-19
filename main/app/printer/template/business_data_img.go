// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/printer/pkg"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// businessDataImgTemplate 图片订单打印模板
type businessDataImgTemplate struct {
	base *printerTemplate
}

// NewBusinessDataImgTemplate 创建新的图片订单打印模板
func NewBusinessDataImgTemplate(
	base *printerTemplate,
) *businessDataImgTemplate {
	return &businessDataImgTemplate{
		base: base,
	}
}

// GetPrintContent 获取内容
func (t *businessDataImgTemplate) GetPrintContent(
	businessData *PrintingBusinessData,
	startTime int64,
	endTime int64,
) string {
	// 店铺设置
	companySetting, _ := t.base.Setting.GetCompanySetting(t.base.Ctx)
	paymentSetting, _ := t.base.Setting.GetPaymentSetting(t.base.Ctx, companySetting)
	isOpenBalance := paymentSetting.IsBalance == "1"
	// 日历
	startTimeStr := t.base.FormatUnixTimeDefault(startTime)
	endTimeStr := t.base.FormatUnixTimeDefault(endTime)
	if endTime == 0 {
		endTimeStr = t.base.FormatUnixTimeDefault(time.Now().Unix())
	}
	// 判断是否是土耳其语、泰语、英语
	isTrThEn := t.base.Lang == "tr" || t.base.Lang == "th" || t.base.Lang == "en"

	//  创建打印机实例
	img := pkg.NewImgFont(568, 0, 0)
	img.SetAlignment(pkg.AlignCenter)
	img.AppendText(t.base.StoreSetting.Name)
	img.LineFeed(1)
	img.LineFeed(1, 20)
	img.SetFontSize(28)
	img.AppendText(t.base.Translate("营业数据"))
	img.LineFeed(1, 58)
	img.SetFontSize(20)
	img.AppendText(fmt.Sprintf("%s %s %s", startTimeStr, t.base.Translate("至"), endTimeStr))
	img.LineFeed(1, 80)
	img.RestoreDefault()
	img.SetImagePadding(0)
	img.SetAlignment(pkg.AlignLeft)
	// 按支付方式
	if businessData.PaymentMethod != nil {
		img.SetTextLineHeight(40)
		img.AppendText(t.base.Translate("实收金额"), pkg.WithFixedWidth(320))
		img.SetAlignment(pkg.AlignRight)
		img.AppendText(t.base.GetPriceAndUnit(businessData.PaymentMethod.TotalReceivedPrice), pkg.WithFixedWidth(0), pkg.WithDeviationWidth(20))
		img.SetAlignment(pkg.AlignLeft)
		img.LineFeed(1)
		img.AppendSplitLine()
		img.RecoverDefaultTextLineHeight()
		for _, income := range businessData.PaymentMethod.PaymentMethodIncomes {
			if income.Code == -1 {
				income.Name = t.base.Translate("免单金额")
			}
			img.AppendText(income.Name, pkg.WithFixedWidth(320))
			img.SetAlignment(pkg.AlignRight)
			img.AppendText(t.base.GetPriceAndUnit(income.Amount), pkg.WithFixedWidth(0), pkg.WithDeviationWidth(20))
			img.SetAlignment(pkg.AlignLeft)
			img.LineFeed(1)
		}
	} else if businessData.ProductCategory != nil {
		// 按商品分类
		img.SetTextLineHeight(40)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("分类"), Width: 300, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("数量"), Width: 96, Align: pkg.AlignRight, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.Translate("小计"), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.AppendSplitLine(pkg.WithLineFeed(true))
		img.RecoverDefaultTextLineHeight()
		for _, category := range businessData.ProductCategory.CategoryList {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: category.Name, Width: 300, Align: pkg.AlignLeft, FontWeight: 2},
				pkg.ColumnConfig{Text: t.base.FloatToString(category.SalesNum), Width: 96, Align: pkg.AlignRight, FontWeight: 2},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(category.Prices), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
			)
		}
		//
		img.LineFeed(1, 10)
		img.AppendSplitLine(pkg.WithLineFeed(true))
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("销售笔数"), Width: 300, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%v", businessData.ProductCategory.SalesNum), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.RecoverDefaultTextLineHeight()
		for _, income := range businessData.ProductCategory.PaymentMethodIncomes {
			if income.Code == -1 {
				income.Name = t.base.Translate("免单金额")
			}
			img.PrintInColumns(
				pkg.ColumnConfig{Text: income.Name, Width: 300, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(income.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		if businessData.ProductCategory.TotalRefundMoney > 0 {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("退款金额"), Width: 300, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.ProductCategory.TotalRefundMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 300, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 28},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.ProductCategory.TotalReceivedPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 2, FontSize: 28},
		)
		img.SetTextLineHeight(40)
	} else if businessData.Product != nil {
		// 按商品
		img.SetTextLineHeight(40)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("商品名称"), Width: 300, Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("销量"), Width: 120, Align: pkg.AlignRight, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("小计"), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		img.AppendSplitLine(pkg.WithLineFeed(true))
		img.RecoverDefaultTextLineHeight()
		for _, product := range businessData.Product.Products {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: product.Name, Width: 300, Align: pkg.AlignLeft},
				pkg.ColumnConfig{Text: fmt.Sprintf("%s*%v", t.base.Number(product.Price), product.SalesNum), Width: 120, Align: pkg.AlignRight},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(product.Subtotal), Width: 0, Align: pkg.AlignRight},
			)
		}
	} else if businessData.All != nil {
		// 全部
		img.RecoverDefaultTextLineHeight()
		// 营业数据
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("总销售额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalSales), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("原商品金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalProductPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("服务费"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalServiceMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("支付手续费"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalPayFeeMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("税费"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalTaxMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("商品数量"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: utils.FormatFloat(businessData.All.TotalProductNum), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("优惠折扣"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalDiscountMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		if isOpenBalance || businessData.All.TotalUserDiscountMoney > 0 {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("会员折扣"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalUserDiscountMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
			)
		}
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("退款金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalRefundMoney), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("赠菜金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalGiveProductPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("免单金额"), Width: 380, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalFreeOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("实收金额"), Width: 400, Align: pkg.AlignLeft, FontWeight: 2, FontSize: 22},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalReceivedPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		// 税收百分比对象列表
		if len(businessData.All.PercentageList) > 0 {
			img.AppendSplitLine()
			img.SetTextLineHeight(45)
			for _, percentage := range businessData.All.PercentageList {
				img.SetAlignment(pkg.AlignLeft)
				img.SetFontWeight(2)
				if t.base.Lang == "ja" {
					img.AppendText(fmt.Sprintf("%s%s%s", t.base.Amount(percentage.TaxRate), "%", t.base.Translate("的对象")))
				} else {
					img.AppendText(fmt.Sprintf("VAT (%s%%)", t.base.Amount(percentage.TaxRate)))
				}
				img.SetFontWeight(1)
				img.LineFeed(1)
				img.PrintInColumns(
					pkg.ColumnConfig{Text: t.base.Translate("合计"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
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
		// 充值相关
		img.AppendSplitLine()
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("会员数据"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignLeft)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("充值金额"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.MemberData.RechargeAmount), Width: 0, Align: pkg.AlignRight},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("赠送金额"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.MemberData.GiftMoney), Width: 0, Align: pkg.AlignRight},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("赠送积分"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: strconv.FormatFloat(businessData.All.MemberData.GiftPoints, 'f', -1, 64), Width: 0, Align: pkg.AlignRight},
		)
		// 未结账相关
		img.AppendSplitLine()
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("未结账数据"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignLeft)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.UnclosedTotalOrderNum)), Width: 0, Align: pkg.AlignRight},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("金额"), Width: 350, Align: pkg.AlignLeft},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.UnclosedTotalPrice), Width: 0, Align: pkg.AlignRight},
		)
		// 合计
		img.AppendSplitLine()
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("合计"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.SetAlignment(pkg.AlignLeft)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("所有订单数"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.TotalOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("取消订单数"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.TotalCancelOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("桌数"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.TotalTableNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("人数"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.TotalPeopleNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("最小/大订单金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.MinOrderPrice), t.base.GetPriceAndUnit(businessData.All.MaxOrderPrice)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("取消订单金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.TotalCancelOrderAmount), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("平均订单金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.AvgOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 桌台方式
		img.LineFeed(1, 12)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("桌台方式"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单数（桌数）"), Width: 400, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.AllTableOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("人数"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.AllTablePeopleNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("最小/大订单金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.AllTableMinOrderPrice), t.base.GetPriceAndUnit(businessData.All.AllTableMaxOrderPrice)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("平均订单金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.AllTableAvgOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("人均"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.AllTablePeopleAvg), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 点餐方式
		img.LineFeed(1, 12)
		img.SetAlignment(pkg.AlignCenter)
		img.SetFontWeight(2)
		img.AppendText(t.base.Translate("点餐方式"))
		img.SetFontWeight(1)
		img.LineFeed(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%.0f", float64(businessData.All.AllCashierOrderNum)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("最小/大订单金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.AllCashierMinOrderPrice), t.base.GetPriceAndUnit(businessData.All.AllCashierMaxOrderPrice)), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("平均订单金额"), Width: 350, Align: pkg.AlignLeft, FontWeight: 1},
			pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(businessData.All.AllCashierAvgOrderPrice), Width: 0, Align: pkg.AlignRight, FontWeight: 1},
		)
		// 支付方式
		img.AppendSplitLine()
		img.LineFeed(1)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("支付方式"), Width: utils.IfInt(isTrThEn || t.base.Lang == "sv", 230, 270), Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: utils.IfInt(t.base.Lang == "en" || t.base.Lang == "sv", 220, 180), Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("金额"), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		totalPayPrice := decimal.NewFromFloat(0)
		for _, income := range businessData.All.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				img.PrintInColumns(
					pkg.ColumnConfig{Text: income.Name, Width: utils.IfInt(isTrThEn, 230, 270), Align: pkg.AlignLeft, FontWeight: 2},
					pkg.ColumnConfig{Text: fmt.Sprintf("%v", income.OrderNum), Width: 96, Align: pkg.AlignLeft, FontWeight: 2},
					pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(income.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
				)
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: t.base.Translate("总金额"), Width: utils.IfInt(isTrThEn, 230, 270), Align: pkg.AlignLeft, FontWeight: 2},
				pkg.ColumnConfig{Text: "", Width: 96, Align: pkg.AlignLeft, FontWeight: 2},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
			)
		}
		img.AppendSplitLine()
		img.LineFeed(1)
		img.SetTextLineHeight(45)
		// 高峰时间
		peakHoursWidth := utils.IfInt(isTrThEn, 230, 270)
		numberWidth := utils.IfInt(t.base.Lang == "en" || t.base.Lang == "sv", 220, 180)
		img.PrintInColumns(
			pkg.ColumnConfig{Text: t.base.Translate("高峰时间"), Width: utils.IfInt(t.base.Lang == "sv", 200, peakHoursWidth), Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: t.base.Translate("订单数"), Width: utils.IfInt(t.base.Lang == "sv", 220, numberWidth), Align: pkg.AlignLeft, FontWeight: 2},
			pkg.ColumnConfig{Text: utils.IfString(t.base.Lang == "en", "Amount", t.base.Translate("订单金额")), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
		)
		for _, peak := range businessData.All.PeakHourList {
			img.PrintInColumns(
				pkg.ColumnConfig{Text: peak.TimePeriod, Width: utils.IfInt(isTrThEn, 230, 270), Align: pkg.AlignLeft, FontWeight: 2},
				pkg.ColumnConfig{Text: fmt.Sprintf("%v", peak.OrderNum), Width: 96, Align: pkg.AlignLeft, FontWeight: 2},
				pkg.ColumnConfig{Text: t.base.GetPriceAndUnit(peak.Amount), Width: 0, Align: pkg.AlignRight, FontWeight: 2},
			)
		}
		img.LineFeed(1)
	}
	//
	img.LineFeed(2)
	//
	return img.Save("", !t.base.IsSunMi, 0)
}
