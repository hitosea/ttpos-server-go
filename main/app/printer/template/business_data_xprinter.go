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

// businessDataXprinterTemplate 商米打印模板
type businessDataXprinterTemplate struct {
	base *printerTemplate
}

// NewBusinessDataXprinterTemplate 创建新的商米打印模板
func NewBusinessDataXprinterTemplate(
	base *printerTemplate,
) *businessDataXprinterTemplate {
	return &businessDataXprinterTemplate{
		base: base,
	}
}

// GetPrintContent 获取内容
func (t *businessDataXprinterTemplate) GetPrintContent(
	printerType string,
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

	// 宽度
	width := 48 - utils.IfInt(printerType == constant.BrandA11510P, 1, 0)
	differenceWidth := 0
	if printerType == constant.BrandA11510P && (t.base.CurrencyUnit == "￥" || t.base.CurrencyUnit == "¥" || t.base.CurrencyUnit == "\xC2\xA5") {
		differenceWidth = 1
	}

	//  创建打印机实例
	printer := pkg.NewPrinter(567)
	printer.SetAlignment(pkg.AlignCenter)
	printer.AppendText(fmt.Sprintf("%s\n", t.base.StoreSetting.Name))
	printer.LineFeed(1)
	printer.SetLineSpacing(40)
	printer.SetPrintModes(true, true, false)
	printer.SetCharacterSize(2, 2)
	printer.AppendText(fmt.Sprintf("%s\n", t.base.Translate("营业数据")))
	printer.LineFeed(1)
	printer.SetLineSpacing(25)
	printer.LineFeed()
	if printerType == constant.PrinterTypeXPrinterWifi {
		printer.LineFeed(2)
	}
	printer.SetLineSpacing(70)
	printer.SetCharacterSize(1, 1)
	printer.RestoreDefaultLineSpacing()
	//
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignCenter)
	printer.AppendText(fmt.Sprintf("%s %s %s\n", startTimeStr, t.base.Translate("至"), endTimeStr))
	printer.LineFeed(2)
	//
	printer.RestoreDefaultLineSpacing()
	printer.SetPrintModes(false, false, false)
	printer.SetAlignment(pkg.AlignLeft)
	if businessData.PaymentMethod != nil {
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(businessData.PaymentMethod.TotalReceivedPrice), width))
		printer.SetPrintModes(false, false, false)
		printer.AppendText("\n------------------------------------------------")
		printer.LineFeed(1)
		for _, income := range businessData.PaymentMethod.PaymentMethodIncomes {
			if income.Code == -1 {
				income.Name = t.base.Translate("免单金额")
			}
			printer.AppendText(t.base.PrintText(income.Name, "", t.base.GetPriceAndUnit(income.Amount), width, 30, 1, 24))
			printer.LineFeed(2)
		}
	} else if businessData.ProductCategory != nil {
		// 按商品分类
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("分类"), t.base.Translate("数量"), t.base.Translate("小计"), width, 29))
		printer.SetPrintModes(false, false, false)
		printer.AppendText("\n------------------------------------------------\n")
		printer.LineFeed()
		for key, category := range businessData.ProductCategory.CategoryList {
			printer.AppendText(t.base.PrintText(category.Name, category.SalesNum, t.base.GetPriceAndUnit(category.Prices), width, 29))
			printer.LineFeed()
			if key != len(businessData.ProductCategory.CategoryList)-1 {
				printer.LineFeed()
			}
		}
		printer.AppendText("\n------------------------------------------------\n")
		//
		printer.AppendText(t.base.PrintText(t.base.Translate("销售笔数"), "", fmt.Sprintf("%v", businessData.ProductCategory.SalesNum), width))
		printer.LineFeed()
		printer.LineFeed()
		//
		for _, income := range businessData.ProductCategory.PaymentMethodIncomes {
			if income.Code == -1 {
				income.Name = t.base.Translate("免单金额")
			}
			printer.AppendText(t.base.PrintText(income.Name, "", t.base.GetPriceAndUnit(income.Amount), width, 30, 1, 24))
			printer.LineFeed()
			printer.LineFeed()
		}
		//
		if businessData.ProductCategory.TotalRefundMoney > 0 {
			printer.AppendText(t.base.PrintText(t.base.Translate("退款金额"), "", t.base.GetPriceAndUnit(businessData.ProductCategory.TotalRefundMoney), width))
			printer.LineFeed()
			printer.LineFeed()
		}
		printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(businessData.ProductCategory.TotalReceivedPrice), width))
	} else if businessData.Product != nil {
		// 按商品
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("商品名称"), t.base.Translate("销量"), t.base.Translate("小计"), width, 26))
		printer.SetPrintModes(false, false, false)
		printer.AppendText("\n------------------------------------------------")
		for _, product := range businessData.Product.Products {
			printer.AppendText(t.base.PrintText(product.Name, fmt.Sprintf("%s*%v", t.base.Number(product.Price), product.SalesNum), t.base.GetPriceAndUnit(product.Subtotal), width-differenceWidth, 26, 16, 16))
			printer.LineFeed()
			printer.LineFeed()
		}
	} else if businessData.All != nil {
		// 全部
		printer.SetLineSpacing(utils.IfInt(printerType == constant.BrandA11510P, 40, 90))
		printer.AppendText(t.base.PrintText(t.base.Translate("总销售额"), "", t.base.GetPriceAndUnit(businessData.All.TotalSales), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("原商品金额"), "", t.base.GetPriceAndUnit(businessData.All.TotalProductPrice), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("服务费"), "", t.base.GetPriceAndUnit(businessData.All.TotalServiceMoney), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("支付手续费"), "", t.base.GetPriceAndUnit(businessData.All.TotalPayFeeMoney), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("税费"), "", t.base.GetPriceAndUnit(businessData.All.TotalTaxMoney), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("商品数量"), "", utils.FormatFloat(businessData.All.TotalProductNum), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("优惠折扣"), "", t.base.GetPriceAndUnit(businessData.All.TotalDiscountMoney), width))
		printer.LineFeed(1)
		if isOpenBalance || businessData.All.TotalUserDiscountMoney > 0 {
			printer.AppendText(t.base.PrintText(t.base.Translate("会员折扣"), "", t.base.GetPriceAndUnit(businessData.All.TotalUserDiscountMoney), width))
			printer.LineFeed(1)
		}
		printer.AppendText(t.base.PrintText(t.base.Translate("退款金额"), "", t.base.GetPriceAndUnit(businessData.All.TotalRefundMoney), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("赠菜金额"), "", t.base.GetPriceAndUnit(businessData.All.TotalGiveProductPrice), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("免单金额"), "", t.base.GetPriceAndUnit(businessData.All.TotalFreeOrderPrice), width))
		printer.LineFeed(1)
		printer.SetPrintModes(true, true, false)
		printer.SetCharacterSize(2, 1)
		printer.AppendText(t.base.PrintText(t.base.Translate("实收金额"), "", t.base.GetPriceAndUnit(businessData.All.TotalReceivedPrice), width))
		printer.SetCharacterSize(1, 1)
		printer.SetPrintModes(false, false, false)
		printer.AppendText("\n------------------------------------------------")
		printer.LineFeed(1)
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
		// 会员充值
		printer.AppendText("------------------------------------------------")
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("会员数据"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("充值金额"), "", t.base.GetPriceAndUnit(businessData.All.MemberData.RechargeAmount), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("赠送金额"), "", t.base.GetPriceAndUnit(businessData.All.MemberData.GiftMoney), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("赠送积分"), "", strconv.FormatFloat(businessData.All.MemberData.GiftPoints, 'f', -1, 64), width))
		printer.LineFeed(1)
		// 未结账相关
		printer.AppendText("------------------------------------------------")
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("未结账数据"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("订单数"), "", fmt.Sprintf("%.0f", float64(businessData.All.UnclosedTotalOrderNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("金额"), "", t.base.GetPriceAndUnit(businessData.All.UnclosedTotalPrice), width))
		printer.LineFeed(1)
		// 合计
		printer.AppendText("------------------------------------------------")
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("合计"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("所有订单数"), "", fmt.Sprintf("%.0f", float64(businessData.All.TotalOrderNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("取消订单数"), "", fmt.Sprintf("%.0f", float64(businessData.All.TotalCancelOrderNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("桌数"), "", fmt.Sprintf("%.0f", float64(businessData.All.TotalTableNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("人数"), "", fmt.Sprintf("%.0f", float64(businessData.All.TotalPeopleNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("最小/大订单金额"), "", fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.MinOrderPrice), t.base.GetPriceAndUnit(businessData.All.MaxOrderPrice)), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("取消订单金额"), "", t.base.GetPriceAndUnit(businessData.All.TotalCancelOrderAmount), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.All.AvgOrderPrice), width))
		printer.LineFeed(1)
		// 桌台方式
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("桌台方式"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("订单数（桌数）"), "", fmt.Sprintf("%.0f", float64(businessData.All.AllTableOrderNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("人数"), "", fmt.Sprintf("%.0f", float64(businessData.All.AllTablePeopleNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("最小/大订单金额"), "", fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.AllTableMinOrderPrice), t.base.GetPriceAndUnit(businessData.All.AllTableMaxOrderPrice)), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.All.AllTableAvgOrderPrice), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("人均"), "", t.base.GetPriceAndUnit(businessData.All.AllTablePeopleAvg), width))
		printer.LineFeed(1)
		// 收银方式
		printer.SetAlignment(pkg.AlignCenter)
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.Translate("点餐方式"))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		printer.SetAlignment(pkg.AlignLeft)
		printer.AppendText(t.base.PrintText(t.base.Translate("订单数"), "", fmt.Sprintf("%.0f", float64(businessData.All.AllCashierOrderNum)), width))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("最小/大订单金额"), "", fmt.Sprintf("%s/%s", t.base.GetPriceAndUnit(businessData.All.AllCashierMinOrderPrice), t.base.GetPriceAndUnit(businessData.All.AllCashierMaxOrderPrice)), width-differenceWidth))
		printer.LineFeed(1)
		printer.AppendText(t.base.PrintText(t.base.Translate("平均订单金额"), "", t.base.GetPriceAndUnit(businessData.All.AllCashierAvgOrderPrice), width))
		printer.LineFeed(1)
		// 支付方式
		printer.AppendText("------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("支付方式"), t.base.Translate("订单数"), t.base.Translate("金额"), width, 24-utils.IfInt(isTrThEn, 4, 0), 20, 16))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		totalPayPrice := decimal.NewFromFloat(0)
		for _, income := range businessData.All.PaymentMethodIncomes {
			if income.Code != constant.PaymentMethodCodeFreePay {
				printer.AppendText(t.base.PrintText(income.Name, fmt.Sprintf("%v", income.OrderNum), t.base.GetPriceAndUnit(income.Amount), width, 26, 10, 18))
				printer.LineFeed()
				totalPayPrice = totalPayPrice.Add(decimal.NewFromFloat(income.Amount).Round(2))
			}
		}
		if totalPayPrice.GreaterThan(decimal.NewFromFloat(0)) {
			printer.AppendText(t.base.PrintText(t.base.Translate("总金额"), "", t.base.GetPriceAndUnit(totalPayPrice.Round(2).InexactFloat64()), width, 26, 10, 18))
			printer.LineFeed(1)
		}
		// 高峰时间
		printer.AppendText("------------------------------------------------")
		printer.SetPrintModes(true, false, false)
		printer.AppendText(t.base.PrintText(t.base.Translate("高峰时间"), t.base.Translate("订单数"), t.base.Translate("订单金额"), width, 24-utils.IfInt(isTrThEn, 4, 0), 20, 16))
		printer.SetPrintModes(false, false, false)
		printer.LineFeed(1)
		for _, peak := range businessData.All.PeakHourList {
			printer.AppendText(t.base.PrintText(peak.TimePeriod, fmt.Sprintf("%v", peak.OrderNum), t.base.GetPriceAndUnit(peak.Amount), width, 26, 10, 18))
			printer.LineFeed(1)
		}
	}
	//
	printer.LineFeed(3)
	// Print and exit page mode
	printer.PrintAndExitPageMode()
	printer.LineFeed(4)
	printer.CutPaper(true)
	//
	return printer.GetOrderData()
}
