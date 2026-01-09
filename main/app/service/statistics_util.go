package service

import (
	"ttpos-server-go/app/model"

	"github.com/shopspring/decimal"
)

// IStatisticsUtilSrv 统计工具服务接口
type IStatisticsUtilSrv interface {
	// MergeTakeoutStatistics 合并外卖订单统计到已有统计数据
	// 将外卖订单统计与已有统计结合，重新计算准确的统计值
	// saleData: 已有销售统计数据（从 repository.CountSale 获取）
	// takeoutData: 外卖订单统计数据（从 repository.CountTakeout 获取）
	// cancelOrderNum: 原有取消订单数（从 CountCancelOrder 获取，可选，默认0）
	// cancelOrderAmount: 原有取消订单金额（从 CountCancelOrder 获取，可选，默认0）
	MergeTakeoutStatistics(saleData model.StatisticsSaleData, takeoutData model.StatisticsTakeoutSaleData, cancelOrderData CountCancelOrderResp) CountSaleResp
}

// statisticsUtilSrv 统计工具服务实现
type statisticsUtilSrv struct {
}

// NewStatisticsUtilSrv 创建统计工具服务实例
func NewStatisticsUtilSrv() IStatisticsUtilSrv {
	return &statisticsUtilSrv{}
}

// MergeTakeoutStatistics 合并外卖订单统计到已有统计数据
// 将外卖订单统计与已有统计结合，重新计算准确的统计值
// 注意：不影响原有统计逻辑，只是将外卖订单统计合并到最终结果中
func (s *statisticsUtilSrv) MergeTakeoutStatistics(saleData model.StatisticsSaleData, takeoutData model.StatisticsTakeoutSaleData, cancelOrderData CountCancelOrderResp) CountSaleResp {
	// 使用 decimal 进行精确计算，避免浮点数精度问题

	// 1. TotalSaleAmount（总销售额）= 原有销售额 + 外卖订单销售额
	totalSaleAmount := decimal.NewFromFloat(saleData.TotalSaleAmount.Float64).
		Add(decimal.NewFromFloat(takeoutData.TotalSaleAmount))

	// 2. TotalReceivedAmount（总实收金额）= 原有实收金额 + 外卖订单实付金额
	totalReceivedAmount := decimal.NewFromFloat(saleData.TotalReceivedAmount.Float64).
		Add(decimal.NewFromFloat(takeoutData.TotalPayAmount))

	// 3. TotalProductNum（总商品数量）= 原有商品数量 + 外卖订单商品数量
	totalProductNum := decimal.NewFromFloat(saleData.TotalProductNum.Float64).
		Add(decimal.NewFromInt(takeoutData.TotalProductNum))

	// 4. TotalBusinessAmount（总营业收入）= 原有营业收入 + 外卖订单营业收入
	totalBusinessAmount := decimal.NewFromFloat(saleData.TotalBusinessAmount.Float64).
		Add(decimal.NewFromFloat(takeoutData.TotalBusinessAmount))

	// 5. TotalTax（总税费）= 原有税费 + 外卖订单税费
	totalTax := decimal.NewFromFloat(saleData.TotalTax.Float64).
		Add(decimal.NewFromFloat(takeoutData.TotalTax))

	// 6. TotalRefundAmount（总退款金额）= 原有退款金额 + 外卖订单退款金额
	totalRefundAmount := decimal.NewFromFloat(saleData.TotalRefundAmount.Float64).
		Add(decimal.NewFromFloat(takeoutData.TotalRefundAmount))

	// 7. TotalDiscount（总优惠折扣）= 原有优惠折扣 + 外卖订单优惠折扣
	totalDiscount := decimal.NewFromFloat(saleData.TotalDiscount.Float64).
		Add(decimal.NewFromFloat(takeoutData.TotalDiscount))

	// 8. TotalDiscountRatio（总优惠折扣率）= 总优惠折扣 / 总销售额 * 100
	var totalDiscountRatio decimal.Decimal
	if totalSaleAmount.GreaterThan(decimal.Zero) {
		totalDiscountRatio = totalDiscount.Div(totalSaleAmount).Mul(decimal.NewFromInt(100))
	}

	// 9. TotalOrderNum（总订单数）= 原有订单数 + 外卖订单数
	totalOrderNum := saleData.TotalOrderNum.Int64 + takeoutData.TotalOrderNum

	// 10. TotalCancelOrderNum（总取消订单数）= 原有取消订单数 + 外卖取消订单数（拒单）
	totalCancelOrderNum := cancelOrderData.TotalCancelOrderNum + takeoutData.CancelOrderNum

	// 11. TotalCancelOrderAmount（总取消订单金额）= 原有取消订单金额 + 外卖取消订单金额（拒单）
	totalCancelOrderAmount := decimal.NewFromFloat(cancelOrderData.TotalCancelOrderAmount).
		Add(decimal.NewFromFloat(takeoutData.CancelOrderAmount))

	// 12. MinOrderAmount（最小订单金额）= 取原有最小值和外卖最小值的较小值
	minOrderAmount := saleData.MinOrderAmount.Float64
	if takeoutData.MinOrderAmount < minOrderAmount {
		minOrderAmount = takeoutData.MinOrderAmount
	}

	// 13. MaxOrderAmount（最大订单金额）= 取原有最大值和外卖最大值的较大值
	maxOrderAmount := saleData.MaxOrderAmount.Float64
	if takeoutData.MaxOrderAmount > maxOrderAmount {
		maxOrderAmount = takeoutData.MaxOrderAmount
	}

	// 14. AvgOrderAmount（平均订单金额）= 总订单金额 / 总订单数
	// 如果外卖订单数为0，则使用原有平均订单金额
	// 如果外卖订单数不为0，则使用总订单金额 / 总订单数
	var avgOrderAmount decimal.Decimal
	if takeoutData.TotalOrderNum > 0 {
		totalOrderAmount := decimal.NewFromFloat(saleData.TotalOrderAmount.Float64).
			Add(decimal.NewFromFloat(takeoutData.TotalOrderAmount)).
			Sub(decimal.NewFromFloat(takeoutData.TotalRefundAmount))
		curTotalOrderNum := totalOrderNum - takeoutData.CancelOrderNum
		avgOrderAmount = totalOrderAmount.Div(decimal.NewFromInt(curTotalOrderNum))
	} else {
		avgOrderAmount = decimal.NewFromFloat(saleData.AvgOrderAmount.Float64)
	}

	// 15. TotalProductPrice（总原商品金额）= 原有原商品金额 + 外卖原商品金额
	totalProductOriginPrice := decimal.NewFromFloat(saleData.TotalProductOriginPrice.Float64).
		Add(decimal.NewFromFloat(takeoutData.TotalProductOriginPrice))

	// 构建返回结果
	// 注意：这里只返回合并后的字段，其他字段需要调用方从原有统计中获取并填充
	return CountSaleResp{
		TotalSaleAmount:         totalSaleAmount.Round(2).InexactFloat64(),
		TotalReceivedAmount:     totalReceivedAmount.Round(2).InexactFloat64(),
		TotalProductNum:         totalProductNum.Round(2).InexactFloat64(),
		TotalBusinessAmount:     totalBusinessAmount.Round(2).InexactFloat64(),
		TotalTax:                totalTax.Round(2).InexactFloat64(),
		TotalRefundAmount:       totalRefundAmount.Round(2).InexactFloat64(),
		TotalDiscount:           totalDiscount.Round(2).InexactFloat64(),
		TotalDiscountRatio:      totalDiscountRatio.Round(2).InexactFloat64(),
		TotalOrderNum:           totalOrderNum,
		TotalCancelOrderNum:     totalCancelOrderNum,
		TotalCancelOrderAmount:  totalCancelOrderAmount.Round(2).InexactFloat64(),
		MinOrderAmount:          minOrderAmount,
		MaxOrderAmount:          maxOrderAmount,
		AvgOrderAmount:          avgOrderAmount.Round(2).InexactFloat64(),
		TotalProductOriginPrice: totalProductOriginPrice.Round(2).InexactFloat64(),
	}
}
