// Package template 提供打印模板相关功能
package template

import (
	"fmt"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	settingResp "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// businessDataSunmiTemplate 商米打印模板
type businessDataSunmiTemplate struct {
	base *printerTemplate
}

// NewBusinessDataSunmiTemplate 创建新的商米打印模板
func NewBusinessDataSunmiTemplate(
	base *printerTemplate,
) *businessDataSunmiTemplate {
	return &businessDataSunmiTemplate{
		base: base,
	}
}

// GetPrintContent 获取内容
func (t *businessDataSunmiTemplate) GetPrintContent(
	printerInfo settingResp.PrinterInfo,
	businessData *PrintingBusinessData,
	startTime int64,
	endTime int64,
) string {
	printerType := printerInfo.PrinterType
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
	isTrMyEn := t.base.Lang == "tr" || t.base.Lang == "my" || t.base.Lang == "en"

	// 是否自己打印
	isOneself := printerType != constant.PrinterTypeSunmiLan && printerType != constant.PrinterTypeSunmiCloud

	//  创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.LineFeed(1)
	printer.SetAlignment(pkg.AlignCenter)
	printer.AppendText(fmt.Sprintf("%s\n", t.base.StoreSetting.Name))
	printer.LineFeed(1)
	printer.SetLineSpacing(45)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(fmt.Sprintf("%s\n", t.base.Translate("营业数据")))
	printer.SetCharacterSize(1, 1)
	if isOneself {
		printer.SetLineSpacing(20)
	}
	printer.LineFeed(1)
	printer.SetLineSpacing(45)
	//
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignCenter)
	printer.AppendText(fmt.Sprintf("%s %s %s\n", startTimeStr, t.base.Translate("至"), endTimeStr))
	printer.SetLineSpacing(utils.IfInt(isOneself, 25, 40))
	printer.LineFeed(1)
	printer.SetLineSpacing(45)
	//
	printer.RestoreDefaultLineSpacing()
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignLeft)
	if businessData.PaymentMethod != nil {
		printer.SetupColumns(
			[]int{360, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(t.base.Translate("实收金额"), t.base.GetPriceAndUnit(businessData.PaymentMethod.TotalReceivedPrice))
		printer.SetPrintModes(false, false, false)
		printer.AppendText("------------------------------------------------\n")
		printer.LineFeed(1)
		for _, income := range businessData.PaymentMethod.PaymentMethodIncomes {
			if income.Code == -1 {
				income.Name = t.base.Translate("免单金额")
			}
			printer.PrintInColumns(income.Name, t.base.GetPriceAndUnit(income.Amount))
			printer.LineFeed(1)
		}
	} else if businessData.ProductCategory != nil {
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{96, pkg.AlignCenter, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(t.base.Translate("分类"), t.base.Translate("数量"), t.base.Translate("小计"))
		printer.SetPrintModes(false, false, false)
		printer.AppendText("------------------------------------------------\n")
		for _, category := range businessData.ProductCategory.CategoryList {
			printer.PrintInColumns(category.Name, t.base.FloatToString(category.SalesNum), t.base.GetPriceAndUnit(category.Prices))
			printer.LineFeed(1)
		}
		printer.AppendText("------------------------------------------------\n")
		//
		printer.SetupColumns(
			[]int{360, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		if businessData.ProductCategory.SalesNum > 0 {
			printer.PrintInColumns(t.base.Translate("销售笔数"), fmt.Sprintf("%v", businessData.ProductCategory.SalesNum))
			printer.LineFeed(1)
		}
		//
		for _, income := range businessData.ProductCategory.PaymentMethodIncomes {
			if income.Code == -1 {
				income.Name = t.base.Translate("免单金额")
			}
			printer.PrintInColumns(income.Name, t.base.GetPriceAndUnit(income.Amount))
			printer.LineFeed(1)
		}
		//
		if businessData.ProductCategory.TotalRefundMoney > 0 {
			printer.PrintInColumns(t.base.Translate("退款金额"), t.base.GetPriceAndUnit(businessData.ProductCategory.TotalRefundMoney))
			printer.LineFeed(1)
		}
		printer.PrintInColumns(t.base.Translate("实收金额"), t.base.GetPriceAndUnit(businessData.ProductCategory.TotalReceivedPrice))
		printer.LineFeed(1)
	} else if businessData.Product != nil {
		// 批次号
		if businessData.Product.BatchRange != "" {
			printer.SetAlignment(pkg.AlignLeft)
			printer.AppendText(t.base.Translate("数量") + "：" + businessData.Product.BatchRange)
			printer.SetLineSpacing(utils.IfInt(isOneself, 25, 40))
			printer.LineFeed(1)
			printer.SetLineSpacing(45)
		}
		// 按商品
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{120, pkg.AlignCenter, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(t.base.Translate("商品名称"), t.base.Translate("销量"), t.base.Translate("小计"))
		printer.SetPrintModes(false, false, false)
		printer.AppendText("------------------------------------------------\n")
		for _, product := range businessData.Product.Products {
			printer.PrintInColumns(product.Name, fmt.Sprintf("%s*%v", t.base.Number(product.Price), product.SalesNum), t.base.GetPriceAndUnit(product.Subtotal))
			printer.LineFeed(1)
		}
	} else if businessData.All != nil {
		// 全部
		printer.SetLineSpacing(utils.IfInt(isOneself, 25, 20))
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("总销售额"), t.base.GetPriceAndUnit(businessData.All.TotalSales))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("原商品金额"), t.base.GetPriceAndUnit(businessData.All.TotalProductPrice))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("服务费"), t.base.GetPriceAndUnit(businessData.All.TotalServiceMoney))
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{400, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("支付手续费"), t.base.GetPriceAndUnit(businessData.All.TotalPayFeeMoney))
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("税费"), t.base.GetPriceAndUnit(businessData.All.TotalTaxMoney))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("商品数量"), utils.FormatFloat(businessData.All.TotalProductNum))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("优惠折扣"), t.base.GetPriceAndUnit(businessData.All.TotalDiscountMoney))
		printer.LineFeed(1)
		if isOpenBalance || businessData.All.TotalUserDiscountMoney > 0 {
			printer.PrintInColumns(t.base.Translate("会员折扣"), t.base.GetPriceAndUnit(businessData.All.TotalUserDiscountMoney))
			printer.LineFeed(1)
		}
		printer.PrintInColumns(t.base.Translate("退款金额"), t.base.GetPriceAndUnit(businessData.All.TotalRefundMoney))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("赠菜金额"), t.base.GetPriceAndUnit(businessData.All.TotalGiveProductPrice))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("免单金额"), t.base.GetPriceAndUnit(businessData.All.TotalFreeOrderPrice))
		printer.LineFeed(1)
		printer.SetPrintModes(true, true, false)
		printer.SetCharacterSize(2, 1)
		printer.PrintInColumns(t.base.Translate("实收金额"), t.base.GetPriceAndUnit(businessData.All.TotalReceivedPrice))
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		// 税收百分比对象列表
		for _, percentage := range businessData.All.PercentageList {
			printer.SetAlignment(pkg.AlignLeft)
			printer.SetPrintModes(true, false, false)
			if t.base.Lang == "ja" {
				printer.AppendText(fmt.Sprintf("%s%s%s", t.base.Amount(percentage.TaxRate), "%", t.base.Translate("的对象")))
			} else {
				printer.AppendText(fmt.Sprintf("VAT (%s%%)", t.base.Amount(percentage.TaxRate)))
			}
			printer.SetPrintModes(false, false, false)
			printer.LineFeed(2)
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
		// 会员充值
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("会员数据"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.PrintInColumns(t.base.Translate("充值金额"), t.base.GetPriceAndUnit(businessData.All.MemberData.RechargeAmount))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("赠送金额"), t.base.GetPriceAndUnit(businessData.All.MemberData.GiftMoney))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("赠送积分"), strconv.FormatFloat(businessData.All.MemberData.GiftPoints, 'f', -1, 64))
		printer.LineFeed(1)
		// 未结账相关
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("未结账数据"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignLeft)
		printer.PrintInColumns(t.base.Translate("订单数"), fmt.Sprintf("%.0f", float64(businessData.All.UnclosedTotalOrderNum)))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("金额"), t.base.GetPriceAndUnit(businessData.All.UnclosedTotalPrice))
		printer.LineFeed(1)
		// 合计
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("合计"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignLeft)
		printer.PrintInColumns(t.base.Translate("所有订单数"), fmt.Sprintf("%.0f", float64(businessData.All.TotalOrderNum)))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("取消订单数"), fmt.Sprintf("%.0f", float64(businessData.All.TotalCancelOrderNum)))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("桌数"), fmt.Sprintf("%.0f", float64(businessData.All.TotalTableNum)))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("人数"), fmt.Sprintf("%.0f", float64(businessData.All.TotalPeopleNum)))
		printer.LineFeed(1)
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("最小/大订单金额"), fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.MinOrderPrice), t.base.GetPriceAndUnit(businessData.All.MaxOrderPrice)))
		if t.base.Lang == "my" {
			printer.SetLineSpacing(utils.IfInt(isOneself, 25, 20))
		}
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("取消订单金额"), t.base.GetPriceAndUnit(businessData.All.TotalCancelOrderAmount))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.All.AvgOrderPrice))
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
		printer.PrintInColumns(t.base.Translate("订单数（桌数）"), fmt.Sprintf("%.0f", float64(businessData.All.AllTableOrderNum)))
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{320, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(t.base.Translate("人数"), fmt.Sprintf("%.0f", float64(businessData.All.AllTablePeopleNum)))
		printer.LineFeed(1)
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("最小/大订单金额"), fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.AllTableMinOrderPrice), t.base.GetPriceAndUnit(businessData.All.AllTableMaxOrderPrice)))
		if t.base.Lang == "my" {
			printer.SetLineSpacing(utils.IfInt(isOneself, 25, 20))
		}
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.All.AllTableAvgOrderPrice))
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("人均"), t.base.GetPriceAndUnit(businessData.All.AllTablePeopleAvg))
		printer.LineFeed(2)
		// 收银方式-店内
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("点餐方式-店内"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(2)
		printer.SetAlignment(pkg.AlignLeft)
		printer.PrintInColumns(t.base.Translate("订单数"), fmt.Sprintf("%.0f", float64(businessData.All.AllCashierOrderNum)))
		printer.LineFeed(1)
		if t.base.Lang == "my" {
			printer.SetLineSpacing(50)
		}
		printer.PrintInColumns(t.base.Translate("最小/大订单金额"), fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.AllCashierMinOrderPrice), t.base.GetPriceAndUnit(businessData.All.AllCashierMaxOrderPrice)))
		if t.base.Lang == "my" {
			printer.SetLineSpacing(utils.IfInt(isOneself, 25, 20))
		}
		printer.LineFeed(1)
		printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.All.AllCashierAvgOrderPrice))
		printer.LineFeed(1)
		// 收银方式-外卖
		if businessData.All.AllTakeawayOrderNum > 0 {
			printer.SetAlignment(pkg.AlignCenter)
			printer.SetPrintModes(true, false, false)
			printer.AppendText(t.base.Translate("点餐方式-外卖"))
			printer.SetPrintModes(false, false, false)
			printer.LineFeed(2)
			printer.SetAlignment(pkg.AlignLeft)
			printer.PrintInColumns(t.base.Translate("订单数"), fmt.Sprintf("%.0f", float64(businessData.All.AllTakeawayOrderNum)))
			printer.LineFeed(1)
			if t.base.Lang == "my" {
				printer.SetLineSpacing(50)
			}
			printer.PrintInColumns(t.base.Translate("最小/大订单金额"), fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.AllTakeawayMinOrderPrice), t.base.GetPriceAndUnit(businessData.All.AllTakeawayMaxOrderPrice)))
			if t.base.Lang == "my" {
				printer.SetLineSpacing(utils.IfInt(isOneself, 25, 20))
			}
			printer.LineFeed(1)
			printer.PrintInColumns(t.base.Translate("平均订单金额"), t.base.GetPriceAndUnit(businessData.All.AllTakeawayAvgOrderPrice))
			printer.LineFeed(1)
		}
		// 支付方式
		printer.AppendText("------------------------------------------------")
		printer.LineFeed(2)
		printer.SetupColumns(
			[]int{utils.IfInt(isTrThEn, 220, 270), pkg.AlignLeft, 0},
			[]int{utils.IfInt(isTrMyEn, 200, 180), pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.SetPrintModes(true, false, false)
		printer.PrintInColumns(
			t.base.Translate("支付方式"),
			t.base.Translate("订单数"),
			utils.IfString(t.base.Lang == "my", "ငွေပမာဏ ", t.base.Translate("金额")),
		)
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{40, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		totalPayPrice := decimal.NewFromFloat(0)
		for _, income := range businessData.All.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				printer.PrintInColumns(income.Name, fmt.Sprintf("%v", income.OrderNum), t.base.GetPriceAndUnit(income.Amount))
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
			[]int{utils.IfInt(isTrThEn, 220, 270), pkg.AlignLeft, 0},
			[]int{utils.IfInt(isTrMyEn, 200, 180), pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		printer.PrintInColumns(
			t.base.Translate("高峰时间"),
			t.base.Translate("订单数"),
			utils.IfString(t.base.Lang == "my", "ငွေပမာဏ ", t.base.Translate("金额")),
		)
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetupColumns(
			[]int{300, pkg.AlignLeft, 0},
			[]int{40, pkg.AlignLeft, 0},
			[]int{0, pkg.AlignRight, 0},
		)
		for _, peak := range businessData.All.PeakHourList {
			printer.PrintInColumns(peak.TimePeriod, fmt.Sprintf("%v", peak.OrderNum), t.base.GetPriceAndUnit(peak.Amount))
			printer.LineFeed(1)
		}
		printer.LineFeed(4)
	}
	// Print and exit page mode
	printer.PrintAndExitPageMode()
	printer.LineFeed(4)
	printer.CutPaper(false)
	//
	return printer.GetOrderData()
}
