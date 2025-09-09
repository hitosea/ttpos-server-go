package service

import (
	"sort"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/shopspring/decimal"
)

// IStatisticsSrv 统计服务接口
type IStatisticsSrv interface {
	CountSale(ctx context.Context, req CountReq) CountSaleResp                                        // 统计销售
	CountSaleDays(ctx context.Context, req CountReq, days []string) []CountSaleDaysResp               // 统计销售天数
	CountPayment(ctx context.Context, req CountReq) CountPaymentResp                                  // 统计支付
	CountPaymentDays(ctx context.Context, req CountReq, days []string) []CountPaymentDaysResp         // 统计支付天数
	CountTax(ctx context.Context, req CountReq) []CountTaxResp                                        // 统计税类
	CountCategory(ctx context.Context, req CountReq) CountCategoryResp                                // 统计分类
	CountProduct(ctx context.Context, req CountReq) []CountProductResp                                // 统计商品
	CountArea(ctx context.Context, req CountReq) []CountAreaResp                                      // 统计区域
	CountAreaDays(ctx context.Context, req CountReq, days []string) []CountAreaDaysResp               // 统计区域天数
	Count7Days(ctx context.Context, req CountReq) Count7DaysResp                                      // 统计销售天数
	CountMemberNum(ctx context.Context, req CountReq) int64                                           // 统计会员数量
	CountMemberNumDays(ctx context.Context, req CountReq, days []string) []CountMemberNumDaysResp     // 统计会员数量天数
	CountMember(ctx context.Context, req CountReq) CountMemberResp                                    // 统计会员
	CountMemberPayment(ctx context.Context, req CountReq) CountPaymentResp                            // 统计会员支付
	CountMemberPaymentDays(ctx context.Context, req CountReq, days []string) []CountPaymentDaysResp   // 统计会员支付天数
	CountUnpaidOrder(ctx context.Context, req CountReq) CountUnpaidOrderResp                          // 统计未结订单
	CountProductSale(ctx context.Context, req CountReq) CountProductSaleResp                          // 统计商品销售
	CountFreePayment(ctx context.Context, req CountReq) CountFreePaymentResp                          // 统计免单支付
	CountFreePaymentDays(ctx context.Context, req CountReq, days []string) []CountFreePaymentDaysResp // 统计免单支付天数
	CountExport(ctx context.Context, req CountReq) (CountExportResp, error)                           // 统计导出
	CountShiftRefundAmount(ctx context.Context, req CountReq) float64                                 // 统计班次退款金额
	CountCancelOrder(ctx context.Context, req CountReq) CountCancelOrderResp                          // 统计取消订单
	RankProduct(ctx context.Context, req CountReq) []CountProductRankResp                             // 统计商品排行
	SaveSale(ctx context.Context, req SaveSaleReq) error                                              // 保存销售
	SaveMember(ctx context.Context, req SaveMemberReq) error                                          // 保存会员
}

// statisticsSrv 统计服务实现
type statisticsSrv struct {
}

// NewStatisticsSrv 创建统计服务
func NewStatisticsSrv() IStatisticsSrv {
	return &statisticsSrv{}
}

// NewStatisticsSrvImpl 创建统计服务实现
func NewStatisticsSrvImpl() IStatisticsSrv {
	return &statisticsSrv{}
}

// CountSaleResp 统计销售响应
type CountSaleResp struct {
	TotalSaleAmount            float64 `json:"total_sale_amount"`             // 总销售额
	TotalReceivedAmount        float64 `json:"total_received_amount"`         // 总实收金额
	TotalProductPrice          float64 `json:"total_product_price"`           // 总商品原价
	TotalProductOriginPrice    float64 `json:"total_product_origin_price"`    // 总原商品金额
	TotalProductNum            float64 `json:"total_product_num"`             // 总商品数量
	TotalDiscountMember        float64 `json:"total_discount_member"`         // 总会员折扣
	TotalBusinessAmount        float64 `json:"total_business_amount"`         // 总营业收入
	TotalServiceFee            float64 `json:"total_service_fee"`             // 总服务费
	TotalPaymentFee            float64 `json:"total_payment_fee"`             // 总支付手续费
	TotalTax                   float64 `json:"total_tax"`                     // 总税额
	TotalRefundAmount          float64 `json:"total_refund_amount"`           // 总退款金额
	TotalDiscount              float64 `json:"total_discount"`                // 总优惠折扣
	TotalDiscountRatio         float64 `json:"total_discount_ratio"`          // 总优惠折扣率
	TotalGiftAmount            float64 `json:"total_gift_amount"`             // 总赠菜金额
	TotalGiftNum               float64 `json:"total_gift_num"`                // 总赠菜数量
	TotalFreeAmount            float64 `json:"total_free_amount"`             // 总免单金额
	TotalFreeNum               float64 `json:"total_free_num"`                // 总免单数量
	TotalOrderNum              int64   `json:"total_order_num"`               // 总订单数量
	TotalTakeoutSaleAmount     float64 `json:"total_takeout_sale_amount"`     // 总外送销售
	TotalTakeoutBusinessAmount float64 `json:"total_takeout_business_amount"` // 总外送营收
	TotalTakeoutRefundAmount   float64 `json:"total_takeout_refund_amount"`   // 总外送退款金额
	TotalTakeoutDeliveryFee    float64 `json:"total_takeout_delivery_fee"`    // 总外送配送费
	TotalDeskNum               int64   `json:"total_desk_num"`                // 总桌台数量
	TotalMealNum               int64   `json:"total_meal_num"`                // 总用餐人数
	TotalCancelOrderNum        int64   `json:"total_cancel_order_num"`        // 总取消订单数
	TotalCancelOrderAmount     float64 `json:"total_cancel_order_amount"`     // 总取消订单金额
	TotalInstantOrderNum       int64   `json:"total_instant_order_num"`       // 总即时订单数量
	TotalInstantOrderAmount    float64 `json:"total_instant_order_amount"`    // 总即时订单金额
	TotalTakeoutOrderNum       int64   `json:"total_takeout_order_num"`       // 总外送订单数
	TotalTakeoutOrderAmount    float64 `json:"total_takeout_order_amount"`    // 总外送订单金额
	MinOrderAmount             float64 `json:"min_order_amount"`              // 最小订单金额
	MaxOrderAmount             float64 `json:"max_order_amount"`              // 最大订单金额
	AvgOrderAmount             float64 `json:"avg_order_amount"`              // 平均订单金额
	MinDeskOrderAmount         float64 `json:"min_desk_order_amount"`         // 最小桌台订单金额
	MaxDeskOrderAmount         float64 `json:"max_desk_order_amount"`         // 最大桌台订单金额
	AvgDeskOrderAmount         float64 `json:"avg_desk_order_amount"`         // 平均桌台订单金额
	AvgDeskPeopleOrderAmount   float64 `json:"avg_desk_people_order_amount"`  // 平均桌台每人订单金额
	MinInstantOrderAmount      float64 `json:"min_instant_order_amount"`      // 最小即时订单金额
	MaxInstantOrderAmount      float64 `json:"max_instant_order_amount"`      // 最大即时订单金额
	AvgInstantOrderAmount      float64 `json:"avg_instant_order_amount"`      // 平均即时订单金额
	MinTakeoutOrderAmount      float64 `json:"min_takeout_order_amount"`      // 总外送最小订单金额
	MaxTakeoutOrderAmount      float64 `json:"max_takeout_order_amount"`      // 总外送最大订单金额
	AvgTakeoutOrderAmount      float64 `json:"avg_takeout_order_amount"`      // 总外送平均订单金额
}

// CountSale 统计销售
func (s *statisticsSrv) CountSale(ctx context.Context, req CountReq) CountSaleResp {
	db := ctx.GetDB()
	opts := s.buildCountOpts(ctx, req)
	saleData := repository.NewStatisticsRepo(db).CountSale(opts...)
	memberData := s.CountMember(ctx, req)
	cancelOrderData := s.CountCancelOrder(ctx, req)

	// 总优惠折扣率 = 总优惠折扣 / 总销售额
	var discountRatio decimal.Decimal
	if saleData.TotalSaleAmount.Float64 > 0 {
		discountRatio = decimal.NewFromFloat(saleData.TotalDiscount.Float64).Div(decimal.NewFromFloat(saleData.TotalSaleAmount.Float64)).Mul(decimal.NewFromInt(100))
	}

	// 平均桌台每人订单金额 = 总桌台订单金额 / 总桌台数量 / 总用餐人数
	var avgDeskPeopleOrderAmount decimal.Decimal
	if saleData.TotalMealNum.Int64 > 0 {
		avgDeskPeopleOrderAmount = decimal.NewFromFloat(saleData.TotalDeskOrderAmount.Float64).Div(decimal.NewFromInt(saleData.TotalMealNum.Int64))
	}

	totalSaleAmount := decimal.NewFromFloat(saleData.TotalSaleAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalSaleAmount))
	totalReceivedAmount := decimal.NewFromFloat(saleData.TotalReceivedAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalPaymentAmount))
	totalPaymentFee := decimal.NewFromFloat(saleData.TotalPaymentFee.Float64).Add(decimal.NewFromFloat(memberData.TotalPaymentFee))
	totalRefundAmount := decimal.NewFromFloat(saleData.TotalRefundAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalRefundAmount))
	totalBusinessAmount := decimal.NewFromFloat(saleData.TotalBusinessAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalPaymentAmount))

	return CountSaleResp{
		TotalSaleAmount:            totalSaleAmount.Round(2).InexactFloat64(),
		TotalReceivedAmount:        totalReceivedAmount.Round(2).InexactFloat64(),
		TotalProductPrice:          saleData.TotalProductPrice.Float64,
		TotalProductOriginPrice:    saleData.TotalProductOriginPrice.Float64,
		TotalProductNum:            saleData.TotalProductNum.Float64,
		TotalDiscountMember:        saleData.TotalDiscountMember.Float64,
		TotalBusinessAmount:        totalBusinessAmount.Round(2).InexactFloat64(),
		TotalServiceFee:            saleData.TotalServiceFee.Float64,
		TotalPaymentFee:            totalPaymentFee.Round(2).InexactFloat64(),
		TotalTax:                   saleData.TotalTax.Float64,
		TotalRefundAmount:          totalRefundAmount.Round(2).InexactFloat64(),
		TotalDiscount:              saleData.TotalDiscount.Float64,
		TotalDiscountRatio:         discountRatio.Round(2).InexactFloat64(),
		TotalGiftAmount:            saleData.TotalGiftAmount.Float64,
		TotalGiftNum:               saleData.TotalGiftNum.Float64,
		TotalFreeAmount:            saleData.TotalFreeAmount.Float64,
		TotalFreeNum:               saleData.TotalFreeNum.Float64,
		TotalOrderNum:              saleData.TotalOrderNum.Int64,
		TotalTakeoutSaleAmount:     saleData.TotalTakeoutSaleAmount.Float64,
		TotalTakeoutBusinessAmount: saleData.TotalTakeoutBusinessAmount.Float64,
		TotalTakeoutRefundAmount:   saleData.TotalTakeoutRefundAmount.Float64,
		TotalTakeoutDeliveryFee:    saleData.TotalTakeoutDeliveryFee.Float64,
		TotalDeskNum:               saleData.TotalDeskNum.Int64,
		TotalMealNum:               saleData.TotalMealNum.Int64,
		TotalCancelOrderNum:        cancelOrderData.TotalCancelOrderNum,
		TotalCancelOrderAmount:     cancelOrderData.TotalCancelOrderAmount,
		TotalInstantOrderNum:       saleData.TotalInstantOrderNum.Int64,
		TotalInstantOrderAmount:    saleData.TotalInstantOrderAmount.Float64,
		TotalTakeoutOrderNum:       saleData.TotalTakeoutOrderNum.Int64,
		TotalTakeoutOrderAmount:    saleData.TotalTakeoutOrderAmount.Float64,
		MinOrderAmount:             saleData.MinOrderAmount.Float64,
		MaxOrderAmount:             saleData.MaxOrderAmount.Float64,
		AvgOrderAmount:             saleData.AvgOrderAmount.Float64,
		MinDeskOrderAmount:         saleData.MinDeskOrderAmount.Float64,
		MaxDeskOrderAmount:         saleData.MaxDeskOrderAmount.Float64,
		AvgDeskOrderAmount:         saleData.AvgDeskOrderAmount.Float64,
		AvgDeskPeopleOrderAmount:   avgDeskPeopleOrderAmount.Round(2).InexactFloat64(),
		MinInstantOrderAmount:      saleData.MinInstantOrderAmount.Float64,
		MaxInstantOrderAmount:      saleData.MaxInstantOrderAmount.Float64,
		AvgInstantOrderAmount:      saleData.AvgInstantOrderAmount.Float64,
		MinTakeoutOrderAmount:      saleData.MinTakeoutOrderAmount.Float64,
		MaxTakeoutOrderAmount:      saleData.MaxTakeoutOrderAmount.Float64,
		AvgTakeoutOrderAmount:      saleData.AvgTakeoutOrderAmount.Float64,
	}
}

// CountSaleDaysResp 统计销售天数响应
type CountSaleDaysResp struct {
	CountSaleResp
	Day string `json:"day"` // 日期
}

// CountSaleDays 统计销售天数
func (s *statisticsSrv) CountSaleDays(ctx context.Context, req CountReq, days []string) []CountSaleDaysResp {
	repo := repository.NewStatisticsRepo(ctx.GetDB())
	opts := s.buildCountOpts(ctx, req)
	saleData := repo.CountSaleDays(opts...)
	memberData := repo.CountMemberDays(opts...)

	list := make([]CountSaleDaysResp, 0, len(days))
	for _, day := range days {
		var (
			totalSaleAmount            decimal.Decimal
			totalReceivedAmount        decimal.Decimal
			totalProductPrice          decimal.Decimal
			totalDiscountMember        decimal.Decimal
			totalBusinessAmount        decimal.Decimal
			totalServiceFee            decimal.Decimal
			totalPaymentFee            decimal.Decimal
			totalTax                   decimal.Decimal
			totalRefundAmount          decimal.Decimal
			totalDiscount              decimal.Decimal
			totalDiscountRatio         decimal.Decimal
			totalGiveAmount            decimal.Decimal
			totalFreeAmount            decimal.Decimal
			totalTakeoutSaleAmount     decimal.Decimal
			totalTakeoutBusinessAmount decimal.Decimal
			totalTakeoutRefundAmount   decimal.Decimal
			totalTakeoutDeliveryFee    decimal.Decimal
			totalInstantOrderAmount    decimal.Decimal
			minOrderAmount             decimal.Decimal
			maxOrderAmount             decimal.Decimal
			avgOrderAmount             decimal.Decimal
			minDeskOrderAmount         decimal.Decimal
			maxDeskOrderAmount         decimal.Decimal
			avgDeskOrderAmount         decimal.Decimal
			avgDeskPeopleOrderAmount   decimal.Decimal
			minInstantOrderAmount      decimal.Decimal
			maxInstantOrderAmount      decimal.Decimal
			avgInstantOrderAmount      decimal.Decimal
			minTakeoutOrderAmount      decimal.Decimal
			maxTakeoutOrderAmount      decimal.Decimal
			avgTakeoutOrderAmount      decimal.Decimal
			totalProductNum            decimal.Decimal
			totalGiveNum               decimal.Decimal
			totalFreeNum               decimal.Decimal
			totalOrderNum              int64
			totalDeskNum               int64
			totalMealNum               int64
			totalInstantOrderNum       int64
			totalTakeoutOrderNum       int64
		)

		saleResult, ok := slice.FindBy(saleData, func(index int, dayData model.StatisticsSaleDaysData) bool {
			return dayData.Day.String == day
		})
		if ok {
			totalSaleAmount = decimal.NewFromFloat(saleResult.TotalSaleAmount.Float64)
			totalReceivedAmount = decimal.NewFromFloat(saleResult.TotalReceivedAmount.Float64)
			totalProductPrice = decimal.NewFromFloat(saleResult.TotalProductPrice.Float64)
			totalDiscountMember = decimal.NewFromFloat(saleResult.TotalDiscountMember.Float64)
			totalBusinessAmount = decimal.NewFromFloat(saleResult.TotalBusinessAmount.Float64)
			totalServiceFee = decimal.NewFromFloat(saleResult.TotalServiceFee.Float64)
			totalPaymentFee = decimal.NewFromFloat(saleResult.TotalPaymentFee.Float64)
			totalTax = decimal.NewFromFloat(saleResult.TotalTax.Float64)
			totalRefundAmount = decimal.NewFromFloat(saleResult.TotalRefundAmount.Float64)
			totalDiscount = decimal.NewFromFloat(saleResult.TotalDiscount.Float64)
			if totalSaleAmount.GreaterThan(decimal.Zero) {
				totalDiscountRatio = totalDiscount.Div(totalSaleAmount).Mul(decimal.NewFromFloat(100)).Round(2)
			}
			totalGiveAmount = decimal.NewFromFloat(saleResult.TotalGiftAmount.Float64)
			totalFreeAmount = decimal.NewFromFloat(saleResult.TotalFreeAmount.Float64)
			totalTakeoutSaleAmount = decimal.NewFromFloat(saleResult.TotalTakeoutSaleAmount.Float64)
			totalTakeoutBusinessAmount = decimal.NewFromFloat(saleResult.TotalTakeoutBusinessAmount.Float64)
			totalTakeoutRefundAmount = decimal.NewFromFloat(saleResult.TotalTakeoutRefundAmount.Float64)
			totalTakeoutDeliveryFee = decimal.NewFromFloat(saleResult.TotalTakeoutDeliveryFee.Float64)
			totalInstantOrderAmount = decimal.NewFromFloat(saleResult.TotalInstantOrderAmount.Float64)
			minOrderAmount = decimal.NewFromFloat(saleResult.MinOrderAmount.Float64).Round(2)
			maxOrderAmount = decimal.NewFromFloat(saleResult.MaxOrderAmount.Float64).Round(2)
			avgOrderAmount = decimal.NewFromFloat(saleResult.AvgOrderAmount.Float64).Round(2)
			minDeskOrderAmount = decimal.NewFromFloat(saleResult.MinDeskOrderAmount.Float64).Round(2)
			maxDeskOrderAmount = decimal.NewFromFloat(saleResult.MaxDeskOrderAmount.Float64).Round(2)
			avgDeskOrderAmount = decimal.NewFromFloat(saleResult.AvgDeskOrderAmount.Float64).Round(2)
			if saleResult.TotalMealNum.Int64 > 0 {
				avgDeskPeopleOrderAmount = decimal.NewFromFloat(saleResult.TotalDeskOrderAmount.Float64).Div(decimal.NewFromInt(saleResult.TotalMealNum.Int64)).Round(2)
			}
			minInstantOrderAmount = decimal.NewFromFloat(saleResult.MinInstantOrderAmount.Float64).Round(2)
			maxInstantOrderAmount = decimal.NewFromFloat(saleResult.MaxInstantOrderAmount.Float64).Round(2)
			avgInstantOrderAmount = decimal.NewFromFloat(saleResult.AvgInstantOrderAmount.Float64).Round(2)
			minTakeoutOrderAmount = decimal.NewFromFloat(saleResult.MinTakeoutOrderAmount.Float64).Round(2)
			maxTakeoutOrderAmount = decimal.NewFromFloat(saleResult.MaxTakeoutOrderAmount.Float64).Round(2)
			avgTakeoutOrderAmount = decimal.NewFromFloat(saleResult.AvgTakeoutOrderAmount.Float64).Round(2)
			totalProductNum = decimal.NewFromFloat(saleResult.TotalProductNum.Float64).Round(2)
			totalGiveNum = decimal.NewFromFloat(saleResult.TotalGiftNum.Float64).Round(2)
			totalFreeNum = decimal.NewFromFloat(saleResult.TotalFreeNum.Float64).Round(2)
			totalOrderNum = saleResult.TotalOrderNum.Int64
			totalDeskNum = saleResult.TotalDeskNum.Int64
			totalMealNum = saleResult.TotalMealNum.Int64
			totalInstantOrderNum = saleResult.TotalInstantOrderNum.Int64
			totalTakeoutOrderNum = saleResult.TotalTakeoutOrderNum.Int64
		}
		memberResult, ok := slice.FindBy(memberData, func(index int, dayData model.StatisticsMemberDaysData) bool {
			return dayData.Day.String == day
		})
		if ok {
			totalSaleAmount = totalSaleAmount.Add(decimal.NewFromFloat(memberResult.TotalSaleAmount.Float64))
			totalReceivedAmount = totalReceivedAmount.Add(decimal.NewFromFloat(memberResult.TotalPaymentAmount.Float64))
			totalBusinessAmount = totalBusinessAmount.Add(decimal.NewFromFloat(memberResult.TotalPaymentAmount.Float64))
			totalPaymentFee = totalPaymentFee.Add(decimal.NewFromFloat(memberResult.TotalPaymentFee.Float64))
			totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(memberResult.TotalRefundAmount.Float64))
		}
		list = append(list, CountSaleDaysResp{
			CountSaleResp: CountSaleResp{
				TotalSaleAmount:            totalSaleAmount.InexactFloat64(),
				TotalReceivedAmount:        totalReceivedAmount.InexactFloat64(),
				TotalProductPrice:          totalProductPrice.InexactFloat64(),
				TotalProductNum:            totalProductNum.InexactFloat64(),
				TotalDiscountMember:        totalDiscountMember.InexactFloat64(),
				TotalBusinessAmount:        totalBusinessAmount.InexactFloat64(),
				TotalServiceFee:            totalServiceFee.InexactFloat64(),
				TotalPaymentFee:            totalPaymentFee.InexactFloat64(),
				TotalTax:                   totalTax.InexactFloat64(),
				TotalRefundAmount:          totalRefundAmount.InexactFloat64(),
				TotalDiscount:              totalDiscount.InexactFloat64(),
				TotalDiscountRatio:         totalDiscountRatio.InexactFloat64(),
				TotalGiftAmount:            totalGiveAmount.InexactFloat64(),
				TotalGiftNum:               totalGiveNum.InexactFloat64(),
				TotalFreeAmount:            totalFreeAmount.InexactFloat64(),
				TotalFreeNum:               totalFreeNum.InexactFloat64(),
				TotalOrderNum:              totalOrderNum,
				TotalDeskNum:               totalDeskNum,
				TotalMealNum:               totalMealNum,
				TotalInstantOrderNum:       totalInstantOrderNum,
				TotalInstantOrderAmount:    totalInstantOrderAmount.InexactFloat64(),
				TotalTakeoutOrderNum:       totalTakeoutOrderNum,
				TotalTakeoutSaleAmount:     totalTakeoutSaleAmount.InexactFloat64(),
				TotalTakeoutBusinessAmount: totalTakeoutBusinessAmount.InexactFloat64(),
				TotalTakeoutRefundAmount:   totalTakeoutRefundAmount.InexactFloat64(),
				TotalTakeoutDeliveryFee:    totalTakeoutDeliveryFee.InexactFloat64(),
				MinOrderAmount:             minOrderAmount.InexactFloat64(),
				MaxOrderAmount:             maxOrderAmount.InexactFloat64(),
				AvgOrderAmount:             avgOrderAmount.InexactFloat64(),
				MinDeskOrderAmount:         minDeskOrderAmount.InexactFloat64(),
				MaxDeskOrderAmount:         maxDeskOrderAmount.InexactFloat64(),
				AvgDeskOrderAmount:         avgDeskOrderAmount.InexactFloat64(),
				AvgDeskPeopleOrderAmount:   avgDeskPeopleOrderAmount.InexactFloat64(),
				MinInstantOrderAmount:      minInstantOrderAmount.InexactFloat64(),
				MaxInstantOrderAmount:      maxInstantOrderAmount.InexactFloat64(),
				AvgInstantOrderAmount:      avgInstantOrderAmount.InexactFloat64(),
				MinTakeoutOrderAmount:      minTakeoutOrderAmount.InexactFloat64(),
				MaxTakeoutOrderAmount:      maxTakeoutOrderAmount.InexactFloat64(),
				AvgTakeoutOrderAmount:      avgTakeoutOrderAmount.InexactFloat64(),
			},
			Day: day,
		})
	}
	return list
}

// CountPaymentReq 统计支付请求
type CountPaymentReq struct {
	ShiftNo        string `json:"shift_no"`         // 当班编号
	QueryStartTime int64  `json:"query_start_time"` // 查询开始时间
	QueryEndTime   int64  `json:"query_end_time"`   // 查询结束时间
}

// CountPaymentResp 统计支付响应
type CountPaymentResp struct {
	TotalReceivedAmount float64                `json:"total_received_amount"` // 总实收金额
	TotalRefundAmount   float64                `json:"total_refund_amount"`   // 总退款金额
	PaymentList         []CountPaymentRespList `json:"payment_list"`          // 支付方式列表
}

type CountPaymentRespList struct {
	ID                 uint64  `json:"id"`                   // 支付方式ID
	Sort               int     `json:"sort"`                 // 支付方式排序
	CreateTime         int64   `json:"create_time"`          // 支付方式创建时间
	PaymentName        string  `json:"payment_name"`         // 支付方式名称
	PaymentCode        int     `json:"payment_code"`         // 支付方式编码
	TotalOrderNum      int64   `json:"total_order_num"`      // 总订单数量
	TotalPaymentAmount float64 `json:"total_payment_amount"` // 总支付金额
}

// CountPayment 统计支付
func (s *statisticsSrv) CountPayment(ctx context.Context, req CountReq) CountPaymentResp {
	opts := s.buildCountOpts(ctx, req)
	paymentData := repository.NewStatisticsRepo(ctx.GetDB()).CountPayment(opts...)
	memberPaymentData := s.CountMemberPayment(ctx, req)

	var (
		totalReceivedAmount decimal.Decimal
		totalRefundAmount   decimal.Decimal
		list                = make([]CountPaymentRespList, 0)
	)

	i := 0
	for _, payment := range paymentData {
		item, ok := slice.Find(list, func(index int, item CountPaymentRespList) bool {
			i = index
			return item.PaymentCode == payment.PaymentCode
		})
		if !ok {
			list = append(list, CountPaymentRespList{
				ID:         payment.ID,
				Sort:       payment.Sort,
				CreateTime: payment.CreateTime,
				PaymentName: func() string {
					if payment.PaymentCode == 0 {
						return i18n.Translate(ctx.GetLanguage(), "免单")
					}
					if payment.PaymentName == "" {
						return "-"
					}
					return payment.PaymentName
				}(),
				PaymentCode:        payment.PaymentCode,
				TotalOrderNum:      payment.TotalOrderNum.Int64,
				TotalPaymentAmount: payment.TotalPaymentAmount.Float64,
			})
		} else {
			item.TotalOrderNum += payment.TotalOrderNum.Int64
			item.TotalPaymentAmount += payment.TotalPaymentAmount.Float64
			list[i] = *item
		}
		if payment.PaymentCode != 10 {
			totalReceivedAmount = totalReceivedAmount.Add(decimal.NewFromFloat(payment.TotalPaymentAmount.Float64))
		}
		totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(payment.TotalRefundAmount.Float64))
	}

	i = 0
	for _, memberPayment := range memberPaymentData.PaymentList {
		item, ok := slice.Find(list, func(index int, item CountPaymentRespList) bool {
			i = index
			return item.PaymentCode == memberPayment.PaymentCode
		})
		if !ok {
			list = append(list, CountPaymentRespList{
				ID:         memberPayment.ID,
				Sort:       memberPayment.Sort,
				CreateTime: memberPayment.CreateTime,
				PaymentName: func() string {
					if memberPayment.PaymentName == "" {
						return "-"
					}
					return memberPayment.PaymentName
				}(),
				PaymentCode:        memberPayment.PaymentCode,
				TotalOrderNum:      memberPayment.TotalOrderNum,
				TotalPaymentAmount: memberPayment.TotalPaymentAmount,
			})
		} else {
			item.TotalOrderNum += memberPayment.TotalOrderNum
			item.TotalPaymentAmount += memberPayment.TotalPaymentAmount
			list[i] = *item
		}
	}

	// 先按Sort升序排序，再按CreateTime降序排序， ID降序排序
	if len(list) > 0 {
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Sort == list[j].Sort {
				if list[i].CreateTime == list[j].CreateTime {
					return list[i].ID < list[j].ID
				}
				return list[i].CreateTime > list[j].CreateTime
			}
			return list[i].Sort < list[j].Sort
		})
	}

	totalReceivedAmount = totalReceivedAmount.Add(decimal.NewFromFloat(memberPaymentData.TotalReceivedAmount))
	totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(memberPaymentData.TotalRefundAmount))

	return CountPaymentResp{
		TotalReceivedAmount: totalReceivedAmount.Round(2).InexactFloat64(),
		TotalRefundAmount:   totalRefundAmount.Round(2).InexactFloat64(),
		PaymentList:         list,
	}
}

// CountPaymentDaysResp 统计支付天数响应
type CountPaymentDaysResp struct {
	PaymentList []CountPaymentRespList `json:"payment_list"` // 支付方式列表
	Day         string                 `json:"day"`          // 日期
}

// CountPaymentDays 统计支付天数
func (s *statisticsSrv) CountPaymentDays(ctx context.Context, req CountReq, days []string) []CountPaymentDaysResp {
	opts := s.buildCountOpts(ctx, req)
	paymentData := repository.NewStatisticsRepo(ctx.GetDB()).CountPaymentDays(opts...)

	list := make([]CountPaymentDaysResp, 0)
	for _, day := range days {
		paymentList := make([]CountPaymentRespList, 0)
		for _, payment := range paymentData {
			if payment.Day.String == day {
				paymentList = append(paymentList, CountPaymentRespList{
					ID:         payment.ID,
					Sort:       payment.Sort,
					CreateTime: payment.CreateTime,
					PaymentName: func() string {
						if payment.PaymentName == "" {
							return "-"
						}
						return payment.PaymentName
					}(),
					PaymentCode:        payment.PaymentCode,
					TotalOrderNum:      payment.TotalOrderNum.Int64,
					TotalPaymentAmount: payment.TotalPaymentAmount.Float64,
				})
			}
		}
		list = append(list, CountPaymentDaysResp{
			PaymentList: paymentList,
			Day:         day,
		})
	}

	return list
}

// CountMemberPayment 统计会员支付
func (s *statisticsSrv) CountMemberPayment(ctx context.Context, req CountReq) CountPaymentResp {
	opts := s.buildCountOpts(ctx, req)
	paymentData := repository.NewStatisticsRepo(ctx.GetDB()).CountMemberPayment(opts...)

	var (
		totalReceivedAmount decimal.Decimal
		totalRefundAmount   decimal.Decimal
		list                = make([]CountPaymentRespList, 0)
	)
	for _, payment := range paymentData {
		list = append(list, CountPaymentRespList{
			ID:         payment.ID,
			Sort:       payment.Sort,
			CreateTime: payment.CreateTime,
			PaymentName: func() string {
				if payment.PaymentName == "" {
					return "-"
				}
				return payment.PaymentName
			}(),
			PaymentCode:        payment.PaymentCode,
			TotalOrderNum:      payment.TotalOrderNum.Int64,
			TotalPaymentAmount: payment.TotalPaymentAmount.Float64,
		})

		totalReceivedAmount = totalReceivedAmount.Add(decimal.NewFromFloat(payment.TotalPaymentAmount.Float64))
		totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(payment.TotalRefundAmount.Float64))
	}

	return CountPaymentResp{
		TotalReceivedAmount: totalReceivedAmount.Round(2).InexactFloat64(),
		TotalRefundAmount:   totalRefundAmount.Round(2).InexactFloat64(),
		PaymentList:         list,
	}
}

// CountMemberPaymentDays 统计会员支付天数
func (s *statisticsSrv) CountMemberPaymentDays(ctx context.Context, req CountReq, days []string) []CountPaymentDaysResp {
	opts := s.buildCountOpts(ctx, req)
	paymentData := repository.NewStatisticsRepo(ctx.GetDB()).CountMemberPaymentDays(opts...)

	list := make([]CountPaymentDaysResp, 0)
	for _, day := range days {
		paymentList := make([]CountPaymentRespList, 0)
		for _, payment := range paymentData {
			if payment.Day.String == day {
				paymentList = append(paymentList, CountPaymentRespList{
					ID:         payment.ID,
					Sort:       payment.Sort,
					CreateTime: payment.CreateTime,
					PaymentName: func() string {
						if payment.PaymentName == "" {
							return "-"
						}
						return payment.PaymentName
					}(),
					PaymentCode:        payment.PaymentCode,
					TotalOrderNum:      payment.TotalOrderNum.Int64,
					TotalPaymentAmount: payment.TotalPaymentAmount.Float64,
				})
			}
		}
		list = append(list, CountPaymentDaysResp{
			PaymentList: paymentList,
			Day:         day,
		})
	}
	return list
}

// CountTaxResp 统计税类响应
type CountTaxResp struct {
	TaxRate            float64 `json:"tax_rate"`             // 税类
	TotalTaxFee        float64 `json:"total_tax_fee"`        // 总税费
	TotalProductAmount float64 `json:"total_product_amount"` // 总商品金额: 含税
}

// CountTax 统计税类
func (s *statisticsSrv) CountTax(ctx context.Context, req CountReq) []CountTaxResp {
	opts := s.buildCountOpts(ctx, req)
	opts = append(opts, repository.NewCommonRepo().WhereByRefundTime(0))
	taxData := repository.NewStatisticsRepo(ctx.GetDB()).CountTax(opts...)
	buffetTaxData := repository.NewStatisticsRepo(ctx.GetDB()).CountBuffetTax(opts...)
	buffetDelayTaxData := repository.NewStatisticsRepo(ctx.GetDB()).CountBuffetDelayTax(opts...)

	var list []CountTaxResp
	for _, tax := range taxData {
		rate := decimal.NewFromFloat(tax.TaxRate.Float64).Mul(decimal.NewFromInt(100)).InexactFloat64()
		list = append(list, CountTaxResp{
			TaxRate:            rate,
			TotalTaxFee:        tax.TotalTaxFee.Float64,
			TotalProductAmount: tax.TotalProductAmount.Float64,
		})
	}

	for _, tax := range buffetTaxData {
		if !slice.ContainBy(list, func(item CountTaxResp) bool {
			rate := decimal.NewFromFloat(tax.TaxRate.Float64).Mul(decimal.NewFromInt(100)).InexactFloat64()
			return item.TaxRate == rate
		}) {
			rate := decimal.NewFromFloat(tax.TaxRate.Float64).Mul(decimal.NewFromInt(100)).InexactFloat64()
			list = append(list, CountTaxResp{
				TaxRate:            rate,
				TotalTaxFee:        tax.TotalTaxFee.Float64,
				TotalProductAmount: tax.TotalProductAmount.Float64,
			})
		} else {
			for i, item := range list {
				rate := decimal.NewFromFloat(tax.TaxRate.Float64).Mul(decimal.NewFromInt(100)).InexactFloat64()
				if item.TaxRate == rate {
					incTaxFee := decimal.NewFromFloat(tax.TotalTaxFee.Float64)
					incProductAmount := decimal.NewFromFloat(tax.TotalProductAmount.Float64)
					item.TotalTaxFee = decimal.NewFromFloat(item.TotalTaxFee).Add(incTaxFee).InexactFloat64()
					item.TotalProductAmount = decimal.NewFromFloat(item.TotalProductAmount).Add(incProductAmount).InexactFloat64()
					list[i] = item
				}
			}
		}
	}

	for _, tax := range buffetDelayTaxData {
		if !slice.ContainBy(list, func(item CountTaxResp) bool {
			rate := decimal.NewFromFloat(tax.TaxRate.Float64).Mul(decimal.NewFromInt(100)).InexactFloat64()
			return item.TaxRate == rate
		}) {
			rate := decimal.NewFromFloat(tax.TaxRate.Float64).Mul(decimal.NewFromInt(100)).InexactFloat64()
			list = append(list, CountTaxResp{
				TaxRate:            rate,
				TotalTaxFee:        tax.TotalTaxFee.Float64,
				TotalProductAmount: tax.TotalProductAmount.Float64,
			})
		} else {
			for i, item := range list {
				rate := decimal.NewFromFloat(tax.TaxRate.Float64).Mul(decimal.NewFromInt(100)).InexactFloat64()
				if item.TaxRate == rate {
					incTaxFee := decimal.NewFromFloat(tax.TotalTaxFee.Float64)
					incProductAmount := decimal.NewFromFloat(tax.TotalProductAmount.Float64)
					item.TotalTaxFee = decimal.NewFromFloat(item.TotalTaxFee).Add(incTaxFee).InexactFloat64()
					item.TotalProductAmount = decimal.NewFromFloat(item.TotalProductAmount).Add(incProductAmount).InexactFloat64()
					list[i] = item
				}
			}
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].TaxRate < list[j].TaxRate
	})

	return list
}

type CountCategoryResp struct {
	TotalSaleNum int64                   `json:"total_sale_num"` // 总销售数量
	CategoryList []CountCategoryListResp `json:"category_list"`  // 分类列表
}

// CountCategoryListResp 统计分类响应
type CountCategoryListResp struct {
	CategoryName string  `json:"category_name"` // 分类名称
	SaleNum      float64 `json:"sale_num"`      // 销售数量
	SaleAmount   float64 `json:"sale_amount"`   // 销售金额
}

// CountCategory 统计分类
func (s *statisticsSrv) CountCategory(ctx context.Context, req CountReq) CountCategoryResp {
	var (
		list []CountCategoryListResp
	)

	opts := s.buildCountOpts(ctx, req)
	orderNum, categoryData := repository.NewStatisticsRepo(ctx.GetDB()).CountCategory(req.CategoryType, ctx.GetLanguage(), opts...)

	for _, category := range categoryData {
		categoryName := category.CategoryParentName.String
		if category.CategoryName.String != "" {
			categoryName = categoryName + "-" + category.CategoryName.String
		}
		list = append(list, CountCategoryListResp{
			CategoryName: categoryName,
			SaleNum:      category.SaleNum.Float64,
			SaleAmount:   category.SaleAmount.Float64,
		})
	}

	return CountCategoryResp{
		TotalSaleNum: orderNum,
		CategoryList: list,
	}
}

// CountProductResp 统计商品响应
type CountProductResp struct {
	ProductName string  `json:"product_name"` // 商品名称
	SalePrice   float64 `json:"sale_price"`   // 销售单价
	SaleNum     float64 `json:"sale_num"`     // 销售数量
	SaleAmount  float64 `json:"sale_amount"`  // 销售金额
}

// CountProduct 统计商品
func (s *statisticsSrv) CountProduct(ctx context.Context, req CountReq) []CountProductResp {
	opts := s.buildCountOpts(ctx, req)
	productData := repository.NewStatisticsRepo(ctx.GetDB()).CountProduct(ctx.GetLanguage(), opts...)

	var list []CountProductResp
	for _, product := range productData {
		productName := product.ProductName.String
		if product.ProductType.Int64 != constant.ProductTypePackage {
			productName = product.ProductName.String + "（" + product.FlavorName.String + "）"
		}
		list = append(list, CountProductResp{
			ProductName: productName,
			SalePrice:   product.SalePrice.Float64,
			SaleNum:     product.SaleNum.Float64,
			SaleAmount:  product.SaleAmount.Float64,
		})
	}
	return list
}

type CountAreaResp struct {
	AreaID             int64   `json:"area_id"`              // 区域id
	AreaName           string  `json:"area_name"`            // 区域名称
	AreaSaleAmount     float64 `json:"area_sale_amount"`     // 区域销售额
	AreaBusinessAmount float64 `json:"area_business_amount"` // 区域营业收入
	AreaProductNum     float64 `json:"area_product_num"`     // 区域商品数量
}

// CountArea 统计区域
func (s *statisticsSrv) CountArea(ctx context.Context, req CountReq) []CountAreaResp {
	opts := s.buildCountOpts(ctx, req)
	areaData := repository.NewStatisticsRepo(ctx.GetDB()).CountArea(opts...)

	var list []CountAreaResp
	for _, area := range areaData {
		list = append(list, CountAreaResp{
			AreaName:           area.AreaName.String,
			AreaSaleAmount:     area.AreaSaleAmount.Float64,
			AreaBusinessAmount: area.AreaBusinessAmount.Float64,
			AreaProductNum:     area.AreaProductNum.Float64,
		})
	}
	return list
}

// CountAreaDaysResp 统计区域天数响应
type CountAreaDaysResp struct {
	AreaList []CountAreaResp `json:"area_list"` // 区域列表
	Day      string          `json:"day"`       // 日期
}

// CountAreaDays 统计区域天数
func (s *statisticsSrv) CountAreaDays(ctx context.Context, req CountReq, days []string) []CountAreaDaysResp {
	opts := s.buildCountOpts(ctx, req)
	areaData := repository.NewStatisticsRepo(ctx.GetDB()).CountAreaDays(opts...)

	var list []CountAreaDaysResp

	for _, day := range days {
		var areaList []CountAreaResp
		for _, area := range areaData {
			if area.Day.String == day {
				areaList = append(areaList, CountAreaResp{
					AreaID:             area.AreaID.Int64,
					AreaName:           area.AreaName.String,
					AreaSaleAmount:     area.AreaSaleAmount.Float64,
					AreaBusinessAmount: area.AreaBusinessAmount.Float64,
					AreaProductNum:     area.AreaProductNum.Float64,
				})
			}
		}

		list = append(list, CountAreaDaysResp{
			AreaList: areaList,
			Day:      day,
		})
	}
	return list
}

type Count7DaysResp struct {
	Days []string             `json:"days"` // 7天
	Data []Count7DaysDataResp `json:"data"` // 7天数据
}

type Count7DaysDataResp struct {
	Day        string  `json:"day"`         // 日期
	TotalNum   int64   `json:"total_num"`   // 总订单数量
	TotalMoney float64 `json:"total_money"` // 总实收金额
}

// Count7Days 统计7天
func (s *statisticsSrv) Count7Days(ctx context.Context, req CountReq) Count7DaysResp {
	days := s.buildDays(req)
	sevenDayList := make([]Count7DaysDataResp, 0, len(days))
	for _, day := range days {
		timezone := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		startTime, _ := timezone.FormatTimeToUnix(day)
		endTime := startTime + 86399
		sevenDayData := repository.NewStatisticsRepo(ctx.GetDB()).Count7Days(
			repository.NewCommonRepoImpl().WhereBetweenByCompleteTime(startTime, endTime),
		)
		oneDayData := Count7DaysDataResp{
			Day:        day,
			TotalNum:   sevenDayData.TotalOrderNum.Int64,
			TotalMoney: sevenDayData.TotalReceivedAmount.Float64,
		}
		sevenDayList = append(sevenDayList, oneDayData)
	}

	return Count7DaysResp{
		Days: days,
		Data: sevenDayList,
	}
}

// CountMemberNum 统计会员数量
func (s *statisticsSrv) CountMemberNum(ctx context.Context, req CountReq) int64 {
	commonRepo := repository.NewCommonRepoImpl()
	req.IsCreateTime = true
	opts := s.buildCountOpts(ctx, req)
	opts = append(opts, commonRepo.WhereBySoftDelete())
	opts = append(opts, commonRepo.WhereByIsVisitor(0))
	return repository.NewStatisticsRepo(ctx.GetDB()).CountMemberNum(opts...)
}

type CountMemberNumDaysResp struct {
	Day       string `json:"day"`        // 日期
	MemberNum int64  `json:"member_num"` // 会员数量
}

// CountMemberNumDays 统计会员数量天数
func (s *statisticsSrv) CountMemberNumDays(ctx context.Context, req CountReq, days []string) []CountMemberNumDaysResp {
	commonRepo := repository.NewCommonRepoImpl()
	req.IsCreateTime = true
	opts := s.buildCountOpts(ctx, req)
	opts = append(opts, commonRepo.WhereBySoftDelete())
	opts = append(opts, commonRepo.WhereByIsVisitor(0))
	memberNumData := repository.NewStatisticsRepo(ctx.GetDB()).CountMemberNumDays(opts...)

	var list []CountMemberNumDaysResp
	for _, day := range days {
		result, ok := slice.FindBy(memberNumData, func(index int, dayData model.CountMemberNumDaysResp) bool {
			return dayData.Day.String == day
		})
		if ok {
			list = append(list, CountMemberNumDaysResp{
				Day:       day,
				MemberNum: result.MemberNum.Int64,
			})
		} else {
			list = append(list, CountMemberNumDaysResp{
				Day:       day,
				MemberNum: 0,
			})
		}
	}
	return list
}

type CountUnpaidOrderResp struct {
	TotalOrderNum int64   `json:"total_order_num"` // 总订单数
	TotalAmount   float64 `json:"total_amount"`    // 总金额
}

// CountUnpaidOrder 统计未结订单
func (s *statisticsSrv) CountUnpaidOrder(ctx context.Context, req CountReq) CountUnpaidOrderResp {
	req.IsCreateTime = true
	opts := s.buildCountOpts(ctx, req)
	unpaidOrderData := repository.NewStatisticsRepo(ctx.GetDB()).CountUnpaidOrder(opts...)

	return CountUnpaidOrderResp{
		TotalOrderNum: unpaidOrderData.TotalOrderNum.Int64,
		TotalAmount:   unpaidOrderData.TotalAmount.Float64,
	}
}

// CountProductRankResp 统计商品排行响应
type CountProductRankResp struct {
	ProductName string  `json:"product_name"` // 商品名称
	SaleNum     float64 `json:"sale_num"`     // 销售数量
	SaleAmount  float64 `json:"sale_amount"`  // 销售金额
}

// RankProduct 统计商品排行
func (s *statisticsSrv) RankProduct(ctx context.Context, req CountReq) []CountProductRankResp {
	opts := s.buildCountOpts(ctx, req)
	productData := repository.NewStatisticsRepo(ctx.GetDB()).RankProduct(req.RankType, ctx.GetLanguage(), opts...)
	var list []CountProductRankResp
	for _, product := range productData {
		list = append(list, CountProductRankResp{
			ProductName: product.ProductName.String,
			SaleNum:     product.SaleNum.Float64,
			SaleAmount:  product.SaleAmount.Float64,
		})
	}
	return list
}

// SaveSaleReq 保存销售请求
type SaveSaleReq struct {
	SaleBillUuid uint64
	OnlyDelete   bool
}

// SaveSale 保存销售
func (s *statisticsSrv) SaveSale(ctx context.Context, req SaveSaleReq) error {
	db := database.GetDBManager(config.Database).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)

	// 如果销售单号为空，直接返回
	if req.SaleBillUuid == 0 {
		return nil
	}

	// 先删除
	statisticsRepo.DeleteSale(req.SaleBillUuid)
	statisticsRepo.DeletePayment(req.SaleBillUuid)
	statisticsRepo.DeleteProduct(req.SaleBillUuid)
	statisticsRepo.DeleteCustomerType(req.SaleBillUuid)
	statisticsRepo.DeleteDelay(req.SaleBillUuid)

	if req.OnlyDelete {
		return nil
	}

	// 查询销售账单详情, 如果销售账单不存在、已删除、未完成，则不统计
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil || saleBill == nil || saleBill.ID == 0 || saleBill.IsDelete() || !saleBill.IsFinish() {
		return nil
	}

	saleBill.CalcAll()

	var (
		sales            []model.StatisticsSale
		payments         []model.StatisticsPayment
		products         []model.StatisticsProduct
		customerTypes    []model.StatisticsCustomerType
		delays           []model.StatisticsDelay
		isTakeout        bool            // 是否外送订单
		orderDeliveryFee decimal.Decimal // 配送费
	)

	// 会员销售订单
	if saleBill.BillType == constant.SaleBillTypeTakeout {
		memberSaleOrder, _ := repository.NewMemberSaleOrderRepoImpl(db).GetMemberSaleOrder(
			repository.NewCommonRepoImpl().WhereByUuid(saleBill.MemberSaleOrderUuid),
			repository.NewCommonRepoImpl().WhereBySoftDelete(),
		)
		// 如果会员销售订单不存在，则不统计
		if memberSaleOrder == nil || memberSaleOrder.ID == 0 {
			return nil
		}
		// 如果会员销售订单状态不是已完成，则不统计
		if memberSaleOrder.Status != constant.MemberSaleOrderStatusCompleted || memberSaleOrder.FinishTime == 0 {
			return nil
		}
		isTakeout = true
		orderDeliveryFee = decimal.NewFromFloat(memberSaleOrder.DeliveryFeeAmount) // 配送费
	}

	// 销售订单
	for _, saleOrder := range saleBill.SaleOrders {
		// 如果订单已删除、未结算，则不统计
		if saleOrder.IsDelete() || !saleOrder.IsSettled() {
			continue
		}
		var (
			orderProductNum           float64
			orderGiveNum              float64
			orderFreeNum              float64
			orderRefundNum            float64
			orderProductPrice         decimal.Decimal
			orderProductOriginPrice   decimal.Decimal
			orderProductSalePrice     decimal.Decimal
			orderProductTax           decimal.Decimal
			orderServiceFee           decimal.Decimal
			orderServiceTax           decimal.Decimal
			orderFreeAmount           decimal.Decimal
			orderGiveAmount           decimal.Decimal
			paymentBalance            decimal.Decimal
			orderDiscount             decimal.Decimal
			orderRefundAmount         decimal.Decimal
			orderRefundPaymentBalance decimal.Decimal
			orderRefundTax            decimal.Decimal
			noOrderRefundTax          decimal.Decimal
			orderRefundServiceFee     decimal.Decimal
			orderRefundDiscount       decimal.Decimal
			orderRefundDiscountMember decimal.Decimal
			orderRefundFee            decimal.Decimal

			isFree          bool = saleOrder.IsFree > 0
			isStatFree      bool = saleBill.SaleBillSetting.IsStatFree == 1
			isSateGive      bool = saleBill.SaleBillSetting.IsStatGift == 1
			isFeeType       bool = saleBill.SaleBillSetting.TaxFeeType == 2
			isFixServiceFee bool = saleBill.SaleBillSetting.ServiceFeeType == 1
		)

		if isFree {
			if isStatFree {
				isSateGive = true
			} else {
				isSateGive = false
			}
		}

		orderGiveAmount = decimal.NewFromFloat(saleOrder.CalcGiftAmount(saleOrder.SaleOrderProducts))

		// 统计订单免单
		if isFree {
			orderFreeNum = 1
			orderFreeAmount = decimal.NewFromFloat(saleOrder.GetAmount()).Round(2)
			if isStatFree {
				if saleOrder.ZeroFee > 0 {
					orderDiscount = orderDiscount.Add(decimal.NewFromFloat(saleOrder.OriginAmount))
				} else if saleOrder.CustomDiscountRate == 1 {
					orderDiscount = orderDiscount.Add(decimal.NewFromFloat(saleOrder.Amount))
				}
				orderServiceFee = decimal.NewFromFloat(saleOrder.ServiceFee)
			}
			// 赠菜计入优惠折扣
			if isSateGive {
				orderDiscount = orderDiscount.Add(orderGiveAmount)
			}
		} else {
			// 统计自定义优惠折扣
			orderDiscount = decimal.NewFromFloat(saleOrder.CustomDiscountFee)
			orderDiscount = orderDiscount.Add(decimal.NewFromFloat(saleOrder.ZeroCheckoutFee))
			orderServiceFee = decimal.NewFromFloat(saleOrder.ServiceFee)
			// 赠菜不计入优惠折扣
			if !isSateGive && saleOrder.CustomAmount == constant.SaleOrderCustomAmountCancel {
				orderDiscount = orderDiscount.Sub(orderGiveAmount)
			}
			if isSateGive && saleOrder.CustomAmount != constant.SaleOrderCustomAmountCancel {
				orderDiscount = orderDiscount.Add(orderGiveAmount)
			}
		}

		// 销售商品
		for _, saleProduct := range saleOrder.SaleOrderProducts {
			// 如果商品已删除、已取消、未接单、是套餐子商品, 则不统计
			if saleProduct.IsDelete() || saleProduct.IsCancelProduct() || !saleProduct.IsAcceptOrderProduct() || saleProduct.IsPackageSubProduct() {
				continue
			}
			// 统计商品数量
			productNum := saleProduct.Num
			productNumDec := decimal.NewFromFloat(productNum)
			orderProductNum += productNum

			productFinalPrice := decimal.NewFromFloat(saleProduct.TotalPrice)

			// 统计赠送商品数量
			productGiveNum := 0.0
			if saleProduct.GiftTime > 0 {
				productGiveNum = productNum
				orderGiveNum += productGiveNum
				productFinalPrice = decimal.Zero
			}

			// 统计免单商品数据
			productFreeNum := 0.0
			if isFree {
				productFreeNum = productNum
				productFinalPrice = decimal.Zero
			}

			// 统计: 商品定价(折扣前)、商品税、服务费、服务费税
			productPrice := decimal.NewFromFloat(saleProduct.SalePrice)
			if isFeeType {
				productPrice = productPrice.Sub(decimal.NewFromFloat(saleProduct.TaxFee))
			}
			saleProductNoTax := decimal.NewFromFloat(saleProduct.SalePriceNoTax).Round(2)
			orderProductOriginPrice = orderProductOriginPrice.Add(saleProductNoTax.Mul(productNumDec))
			productTax := decimal.NewFromFloat(saleProduct.TaxFee)
			productServiceFee := decimal.NewFromFloat(saleProduct.ServiceFee)
			productServiceTax := decimal.NewFromFloat(saleProduct.ServiceTaxFee)
			if saleProduct.GiftTime > 0 {
				productTax = decimal.NewFromFloat(0)
				productServiceFee = decimal.NewFromFloat(0)
				productServiceTax = decimal.NewFromFloat(0)
			}
			if isFree {
				if isStatFree {
					orderProductPrice = orderProductPrice.Add(productPrice.Mul(productNumDec))
					orderProductTax = orderProductTax.Add(productTax.Mul(productNumDec))
					orderServiceTax = orderServiceTax.Add(productServiceTax.Mul(productNumDec))
					if saleOrder.CustomDiscountRate != 1 {
						orderDiscount = orderDiscount.Add(productPrice.Add(productTax).Add(productServiceFee).Add(productServiceTax).Mul(productNumDec))
					}
				}
			} else {
				if saleProduct.GiftTime > 0 {
					if isSateGive {
						orderProductPrice = orderProductPrice.Add(productPrice.Mul(productNumDec))
						orderProductTax = orderProductTax.Add(productTax.Mul(productNumDec))
						orderServiceTax = orderServiceTax.Add(productServiceTax.Mul(productNumDec))
					}
				} else {
					orderProductPrice = orderProductPrice.Add(productPrice.Mul(productNumDec))
					orderProductTax = orderProductTax.Add(productTax.Mul(productNumDec))
					orderServiceTax = orderServiceTax.Add(productServiceTax.Mul(productNumDec))
				}
			}

			// 统计商品销售价
			productSalePrice := decimal.NewFromFloat(saleProduct.SalePrice)
			orderProductSalePrice = orderProductSalePrice.Add(productSalePrice.Mul(productNumDec))

			// 保存商品BOM UUID
			var productBomUuid uint64
			for _, productBom := range saleProduct.SaleOrderProductBoms {
				if productBom.IsDelete() {
					continue
				}
				if productBom.IsFlavorBom == 1 {
					productBomUuid = productBom.ProductBomUuid
				}
			}

			productRefundNum := 0.0
			for _, refundProduct := range saleProduct.ReturnOrderProducts {
				productRefundNum += refundProduct.Num
				orderRefundTax = orderRefundTax.Add(
					decimal.NewFromFloat(saleProduct.TaxFee).Add(
						decimal.NewFromFloat(saleProduct.ServiceTaxFee),
					).Mul(decimal.NewFromFloat(refundProduct.Num)),
				)
				if isFeeType {
					noOrderRefundTax = noOrderRefundTax.Add(
						decimal.NewFromFloat(saleProduct.TaxFee).Mul(decimal.NewFromFloat(refundProduct.Num)),
					)
				}
				if !isFixServiceFee {
					orderRefundServiceFee = orderRefundServiceFee.Add(
						decimal.NewFromFloat(saleProduct.ServiceFee).Mul(decimal.NewFromFloat(refundProduct.Num)),
					)
				}
			}

			orderRefundNum += productRefundNum

			products = append(products, model.StatisticsProduct{
				SaleBillUuid:            saleBill.Uuid,
				SaleOrderUuid:           saleOrder.Uuid,
				DutyNo:                  saleBill.DutyNo,
				DeskUuid:                saleBill.DeskUuid,
				ProductPackageUuid:      saleProduct.ProductPackageUuid,
				ProductBomUuid:          productBomUuid,
				ProductPrice:            productPrice.InexactFloat64(),
				ProductSalePrice:        productSalePrice.InexactFloat64(),
				ProductFinalPrice:       productFinalPrice.InexactFloat64(),
				FlavorPrice:             saleProduct.FlavorPrice,
				SaucePrice:              saleProduct.SaucePrice,
				ProductNum:              productNum,
				TaxRate:                 saleProduct.TaxRate,
				TaxFee:                  productTax.InexactFloat64(),
				ServiceFee:              productServiceFee.InexactFloat64(),
				ServiceTax:              productServiceTax.InexactFloat64(),
				GiveNum:                 productGiveNum,
				FreeNum:                 productFreeNum,
				CompleteTime:            saleBill.FinishTime,
				RefundNum:               productRefundNum,
				IsTakeout:               utils.IfInt(isTakeout, 1, 0),
				MemberOrderDiscountRate: saleProduct.MemberOrderDiscountRate,
			})
		}

		// 统计自助餐
		for _, saleBuffetCustomerType := range saleOrder.SaleOrderBuffetCustomerTypes {
			if saleBuffetCustomerType.IsDelete() {
				continue
			}
			// 统计商品数量
			productNum := float64(saleBuffetCustomerType.Num)
			productNumDec := decimal.NewFromFloat(productNum)
			orderProductNum += productNum

			// 统计: 商品定价(折扣前)、商品税、服务费、服务费税
			productPrice := decimal.NewFromFloat(saleBuffetCustomerType.SalePrice)
			if isFeeType {
				productPrice = productPrice.Sub(decimal.NewFromFloat(saleBuffetCustomerType.TaxFee))
			}
			saleProductNoTax := decimal.NewFromFloat(saleBuffetCustomerType.SalePriceNoTax).Round(2)
			orderProductOriginPrice = orderProductOriginPrice.Add(saleProductNoTax.Mul(productNumDec))
			productTax := decimal.NewFromFloat(saleBuffetCustomerType.TaxFee)
			productServiceFee := decimal.NewFromFloat(saleBuffetCustomerType.ServiceFee)
			productServiceTax := decimal.NewFromFloat(saleBuffetCustomerType.ServiceTaxFee)

			freeNum := 0.0
			if isFree {
				freeNum = float64(saleBuffetCustomerType.Num)
				if isStatFree {
					orderProductPrice = orderProductPrice.Add(productPrice.Mul(productNumDec))
					orderProductTax = orderProductTax.Add(productTax.Mul(productNumDec))
					orderServiceTax = orderServiceTax.Add(productServiceTax.Mul(productNumDec))
					if saleOrder.CustomDiscountRate != 1 {
						orderDiscount = orderDiscount.Add(productPrice.Add(productTax).Add(productServiceFee).Add(productServiceTax).Mul(productNumDec))
					}
				}
			} else {
				orderProductPrice = orderProductPrice.Add(productPrice.Mul(productNumDec))
				orderProductTax = orderProductTax.Add(productTax.Mul(productNumDec))
				orderServiceTax = orderServiceTax.Add(productServiceTax.Mul(productNumDec))
			}

			// 统计商品销售价
			productSalePrice := decimal.NewFromFloat(saleBuffetCustomerType.SalePrice)
			orderProductSalePrice = orderProductSalePrice.Add(productSalePrice.Mul(productNumDec))

			productRefundNum := 0.0
			for _, refundProduct := range saleBuffetCustomerType.ReturnOrderProducts {
				productRefundNum += refundProduct.Num
				orderRefundTax = orderRefundTax.Add(
					decimal.NewFromFloat(saleBuffetCustomerType.TaxFee).Add(
						decimal.NewFromFloat(saleBuffetCustomerType.ServiceTaxFee),
					).Mul(decimal.NewFromFloat(refundProduct.Num)),
				)
				if isFeeType {
					noOrderRefundTax = noOrderRefundTax.Add(
						decimal.NewFromFloat(saleBuffetCustomerType.TaxFee).Mul(decimal.NewFromFloat(refundProduct.Num)),
					)
				}
				if !isFixServiceFee {
					orderRefundServiceFee = orderRefundServiceFee.Add(
						decimal.NewFromFloat(saleBuffetCustomerType.ServiceFee).Mul(decimal.NewFromFloat(refundProduct.Num)),
					)
				}
			}

			orderRefundNum += productRefundNum

			customerTypes = append(customerTypes, model.StatisticsCustomerType{
				SaleBillUuid:                saleBill.Uuid,
				SaleOrderUuid:               saleOrder.Uuid,
				DutyNo:                      saleBill.DutyNo,
				DeskUuid:                    saleBill.DeskUuid,
				BuffetPackageUuid:           saleBuffetCustomerType.BuffetPackageUuid,
				BuffetCustomerTypePriceUuid: saleBuffetCustomerType.BuffetCustomerTypePriceUuid,
				ProductPrice:                productPrice.InexactFloat64(),
				ProductSalePrice:            productSalePrice.InexactFloat64(),
				ProductNum:                  productNum,
				TaxRate:                     saleBuffetCustomerType.TaxRate,
				TaxFee:                      productTax.InexactFloat64(),
				ServiceFee:                  saleBuffetCustomerType.ServiceFee,
				ServiceTax:                  saleBuffetCustomerType.ServiceTaxFee,
				FreeNum:                     freeNum,
				RefundNum:                   productRefundNum,
				CompleteTime:                saleBill.FinishTime,
			})
		}

		// 统计加钟
		for _, saleBuffetDelayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
			if saleBuffetDelayProduct.IsDelete() {
				continue
			}
			// 统计商品数量
			productNum := float64(saleBuffetDelayProduct.Num)
			productNumDec := decimal.NewFromFloat(productNum)
			orderProductNum += productNum

			// 统计: 商品定价(折扣前)、商品税、服务费、服务费税
			productPrice := decimal.NewFromFloat(saleBuffetDelayProduct.Price)

			saleProductNoTax := productPrice
			orderProductOriginPrice = orderProductOriginPrice.Add(saleProductNoTax.Mul(productNumDec))

			freeNum := 0.0
			if isFree {
				freeNum = float64(saleBuffetDelayProduct.Num)
				if isStatFree {
					orderProductPrice = orderProductPrice.Add(productPrice.Mul(productNumDec))
					if saleOrder.CustomDiscountRate != 1 {
						orderDiscount = orderDiscount.Add(productPrice.Mul(productNumDec))
					}
				}
			} else {
				orderProductPrice = orderProductPrice.Add(productPrice.Mul(productNumDec))
			}

			// 统计商品销售价
			productSalePrice := decimal.NewFromFloat(saleBuffetDelayProduct.Price)
			orderProductSalePrice = orderProductSalePrice.Add(productSalePrice.Mul(productNumDec))

			productRefundNum := 0.0
			for _, refundProduct := range saleBuffetDelayProduct.ReturnOrderProducts {
				productRefundNum += refundProduct.Num
			}

			orderRefundNum += productRefundNum

			delays = append(delays, model.StatisticsDelay{
				SaleBillUuid:    saleBill.Uuid,
				SaleOrderUuid:   saleOrder.Uuid,
				DutyNo:          saleBill.DutyNo,
				DeskUuid:        saleBill.DeskUuid,
				BuffetDelayUuid: saleBuffetDelayProduct.BuffetDelayUuid,
				ProductPrice:    productPrice.InexactFloat64(),
				ProductNum:      productNum,
				FreeNum:         freeNum,
				RefundNum:       productRefundNum,
				CompleteTime:    saleBill.FinishTime,
			})

		}

		if !isFree && saleOrder.GetCanReturnAmount() == 0 {
			orderRefundFee = decimal.NewFromFloat(saleOrder.PaymentCommissionFee)
			// orderRefundDiscount = decimal.NewFromFloat(saleOrder.CustomDiscountFee).Add(decimal.NewFromFloat(saleOrder.ZeroCheckoutFee))
			// if !isSateGive {
			// 	orderRefundDiscount = orderRefundDiscount.Sub(orderGiveAmount)
			// }
			// orderRefundDiscountMember = decimal.NewFromFloat(saleOrder.MemberDiscountFee)
			if isFixServiceFee {
				orderRefundServiceFee = decimal.NewFromFloat(saleOrder.ServiceFee)
			}
		}
		// 支付订单
		for _, salePayment := range saleOrder.PaymentOrders {
			var paymentRefundAmount decimal.Decimal
			for _, refundOrderAmount := range salePayment.ReturnOrderAmounts {
				if refundOrderAmount.PaymentMethod != nil && refundOrderAmount.PaymentMethod.Code == 10 {
					orderRefundPaymentBalance = orderRefundPaymentBalance.Add(decimal.NewFromFloat(refundOrderAmount.Amount))
				} else {
					orderRefundAmount = orderRefundAmount.Add(decimal.NewFromFloat(refundOrderAmount.Amount))
				}
				paymentRefundAmount = paymentRefundAmount.Add(decimal.NewFromFloat(refundOrderAmount.Amount))
			}
			if salePayment.PaymentMethod != nil && salePayment.PaymentMethod.Code == 10 {
				paymentBalance = decimal.NewFromFloat(salePayment.Amount)
			}
			payments = append(payments, model.StatisticsPayment{
				SaleBillUuid:      saleBill.Uuid,
				DutyNo:            saleBill.DutyNo,
				DeskUuid:          saleBill.DeskUuid,
				SaleOrderUuid:     saleOrder.Uuid,
				PaymentMethodUuid: salePayment.PaymentMethodUuid,
				PaymentAmount:     salePayment.Amount,
				RefundAmount:      paymentRefundAmount.Round(2).InexactFloat64(),
				CompleteTime:      saleBill.FinishTime,
			})
		}

		sale := model.StatisticsSale{
			SaleBillUuid:         saleBill.Uuid,
			DutyNo:               saleBill.DutyNo,
			DeskUuid:             saleBill.DeskUuid,
			SaleOrderUuid:        saleOrder.Uuid,
			MealNum:              int(saleBill.MealNum),
			ProductPrice:         orderProductPrice.InexactFloat64(),
			ProductOriginPrice:   orderProductOriginPrice.InexactFloat64(),
			ProductSalePrice:     orderProductSalePrice.InexactFloat64(),
			ProductNum:           orderProductNum,
			ProductTax:           orderProductTax.InexactFloat64(),
			ServiceFee:           orderServiceFee.InexactFloat64(),
			ServiceTax:           orderServiceTax.InexactFloat64(),
			Discount:             orderDiscount.InexactFloat64(),
			DiscountMember:       saleOrder.MemberDiscountFee,
			GiftAmount:           orderGiveAmount.InexactFloat64(),
			GiftNum:              orderGiveNum,
			FreeAmount:           orderFreeAmount.InexactFloat64(),
			FreeNum:              orderFreeNum,
			PaymentAmount:        saleOrder.PaymentAmount,
			PaymentFee:           saleOrder.PaymentCommissionFee,
			PaymentBalance:       paymentBalance.InexactFloat64(),
			RefundAmount:         orderRefundAmount.InexactFloat64(),
			RefundPaymentBalance: orderRefundPaymentBalance.InexactFloat64(),
			RefundTax:            orderRefundTax.InexactFloat64(),
			NoRefundTax:          noOrderRefundTax.InexactFloat64(),
			RefundServiceFee:     orderRefundServiceFee.InexactFloat64(),
			RefundDiscount:       orderRefundDiscount.InexactFloat64(),
			RefundDiscountMember: orderRefundDiscountMember.InexactFloat64(),
			RefundFee:            orderRefundFee.InexactFloat64(),
			CompleteTime:         saleBill.FinishTime,
			IsTakeout:            utils.IfInt(isTakeout, 1, 0),
			DeliveryFee:          orderDeliveryFee.InexactFloat64(),
		}
		sales = append(sales, sale)
	}

	if len(sales) > 0 {
		err := statisticsRepo.SaveSale(sales)
		if err != nil {
			return err
		}
	}

	if len(payments) > 0 {
		err := statisticsRepo.SavePayment(payments)
		if err != nil {
			return err
		}
	}

	if len(products) > 0 {
		err := statisticsRepo.SaveProduct(products)
		if err != nil {
			return err
		}
	}

	if len(customerTypes) > 0 {
		err := statisticsRepo.SaveCustomerType(customerTypes)
		if err != nil {
			return err
		}
	}

	if len(delays) > 0 {
		err := statisticsRepo.SaveDelay(delays)
		if err != nil {
			return err
		}
	}

	return nil
}

// CountReq 统计请求
type CountReq struct {
	TimeType       int    `json:"time_type"`        // 时间类型 (1 今天, 2 昨天, 3 本周, 4 本月)
	QueryStartTime int64  `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64  `json:"query_end_time"`   // 查询结束时间戳
	CategoryType   int    `json:"category_type"`    // 分类类型 (1 按一级分类, 2 按二级分类)
	DutyNo         string `json:"duty_no"`          // 班次编号
	RankType       int    `json:"rank_type"`        // 排行类型 (1 按销售数量, 2 按销售金额)
	RankDirection  int    `json:"rank_direction"`   // 排行方向 (1 升序, 2 降序)
	IsCreateTime   bool   `json:"is_create_time"`   // 是否是创建时间
	IsUpdateTime   bool   `json:"is_update_time"`   // 是否是更新时间
	PageNo         int    `json:"page_no"`          // 页码
	PageSize       int    `json:"page_size"`        // 每页大小
	AreaUuid       uint64 `json:"area_uuid"`        // 区域UUID -1=全都
	CategoryUuid   uint64 `json:"category_uuid"`    // 分类UUID -1=全都
	ProductName    string `json:"product_name"`     // 商品名称
	Timezone       string `json:"timezone"`         // 时区
	StaffUuid      uint64 `json:"staff_uuid"`       // 操作员UUID
}

// buildCountOpts 构建统计选项
func (s *statisticsSrv) buildCountOpts(ctx context.Context, req CountReq) []repository.DBOption {
	var (
		opts           []repository.DBOption
		commonRepo     = repository.NewCommonRepo()
		queryStartTime int64
		queryEndTime   int64
	)
	if req.Timezone == "" {
		req.Timezone = ctx.GetCompanySetting().Timezone
	}
	// 处理时间范围
	if req.TimeType > 0 && req.TimeType < 5 {
		switch req.TimeType {
		case 1: // 今天
			queryStartTime, queryEndTime = utils.SetTimezone(req.Timezone).TodayStartEndUnix()
		case 2: // 昨天
			queryStartTime, queryEndTime = utils.SetTimezone(req.Timezone).YesterdayStartEndUnix()
		case 3: // 本周
			queryStartTime, queryEndTime = utils.SetTimezone(req.Timezone).WeekStartEndUnix()
		case 4: // 本月
			queryStartTime, queryEndTime = utils.SetTimezone(req.Timezone).MonthStartEndUnix()
		}
	}
	if req.QueryStartTime > 0 && req.QueryEndTime > 0 {
		queryStartTime = req.QueryStartTime
		queryEndTime = req.QueryEndTime
	}
	if queryStartTime > 0 && queryEndTime > 0 {
		if req.IsCreateTime {
			opts = append(opts, commonRepo.WhereBetweenByCreateTime(queryStartTime, queryEndTime))
		} else if req.IsUpdateTime {
			opts = append(opts, commonRepo.WhereBetweenByUpdateTime(queryStartTime, queryEndTime))
		} else {
			opts = append(opts, commonRepo.WhereBetweenByCompleteTime(queryStartTime, queryEndTime))
		}
	}
	if req.DutyNo != "" {
		opts = append(opts, commonRepo.WhereByDutyNo(req.DutyNo))
	}

	return opts
}

// buildDays 构建日期
func (s *statisticsSrv) buildDays(req CountReq) []string {

	location, _ := time.LoadLocation(req.Timezone) // 或其他时区

	var (
		days      []string
		format    = "2006-01-02"
		startTime = time.Unix(req.QueryStartTime, 0).In(location)
		endTime   = time.Unix(req.QueryEndTime, 0).In(location)
	)

	for startTime.Before(endTime) {
		days = append(days, startTime.Format(format))
		startTime = startTime.AddDate(0, 0, 1)
	}

	return days
}

// SaveMemberReq 保存会员请求
type SaveMemberReq struct {
	MemberRechargeOrderUuid uint64
	OnlyDelete              bool
}

// SaveMember 保存会员
func (s *statisticsSrv) SaveMember(ctx context.Context, req SaveMemberReq) error {
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)

	if req.MemberRechargeOrderUuid == 0 {
		return nil
	}

	// 先删除
	statisticsRepo.DeleteMember(req.MemberRechargeOrderUuid)
	statisticsRepo.DeleteMemberPayment(req.MemberRechargeOrderUuid)

	if req.OnlyDelete {
		return nil
	}

	memberRechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	memberRechargeOrder := memberRechargeOrderRepo.GetRechargeOrderAllInfo(req.MemberRechargeOrderUuid)
	// 如果充值订单已删除、未结算，则不统计
	if memberRechargeOrder.Uuid == 0 || memberRechargeOrder.IsDelete() || memberRechargeOrder.Status != 1 {
		return nil
	}

	var payments []model.StatisticsMemberPayment
	for _, paymentOrder := range memberRechargeOrder.PaymentOrders {
		if paymentOrder.IsDelete() {
			continue
		}
		var paymentRefundAmount decimal.Decimal
		for _, refundOrderAmount := range paymentOrder.ReturnOrderAmounts {
			if refundOrderAmount.IsDelete() {
				continue
			}
			paymentRefundAmount = paymentRefundAmount.Add(decimal.NewFromFloat(refundOrderAmount.Amount))
		}
		payment := model.StatisticsMemberPayment{
			MemberRechargeOrderUuid: memberRechargeOrder.Uuid,
			DutyNo:                  memberRechargeOrder.DutyNo,
			PaymentMethodUuid:       paymentOrder.PaymentMethodUuid,
			PaymentAmount:           paymentOrder.Amount,
			RefundAmount:            paymentRefundAmount.Round(2).InexactFloat64(),
			CompleteTime:            memberRechargeOrder.PaymentTime,
		}
		payments = append(payments, payment)
	}

	paymentFee := decimal.NewFromFloat(memberRechargeOrder.Amount).Sub(decimal.NewFromFloat(memberRechargeOrder.RechargeAmount))
	var refundFee decimal.Decimal
	for _, returnOrder := range memberRechargeOrder.ReturnOrders {
		if returnOrder.IsDelete() {
			continue
		}
		if returnOrder.ReturnType == constant.ReturnOrderRefundTypeTotal {
			refundFee = paymentFee
		}
	}

	member := model.StatisticsMember{
		MemberRechargeOrderUuid: memberRechargeOrder.Uuid,
		DutyNo:                  memberRechargeOrder.DutyNo,
		RechargeAmount:          memberRechargeOrder.RechargeAmount,
		GiveAmount:              memberRechargeOrder.GiftAmount,
		GivePoint:               memberRechargeOrder.GiftPoint,
		PaymentAmount:           memberRechargeOrder.Amount,
		PaymentFee:              paymentFee.Round(2).InexactFloat64(),
		RefundAmount:            memberRechargeOrder.RefundAmount,
		RefundFee:               refundFee.Round(2).InexactFloat64(),
		CompleteTime:            memberRechargeOrder.PaymentTime,
	}
	err := statisticsRepo.SaveMember(member)
	if err != nil {
		return nil
	}

	if len(payments) > 0 {
		err := statisticsRepo.SaveMemberPayment(payments)
		if err != nil {
			return err
		}
	}

	return nil
}

// CountMemberResp 统计会员响应
type CountMemberResp struct {
	TotalSaleAmount     float64 `json:"total_sale_amount"`     // 总销售额
	TotalRechargeAmount float64 `json:"total_recharge_amount"` // 总充值金额
	TotalGiveAmount     float64 `json:"total_give_amount"`     // 总赠送金额
	TotalGivePoint      float64 `json:"total_give_point"`      // 总赠送积分
	TotalPaymentAmount  float64 `json:"total_payment_amount"`  // 总支付金额
	TotalPaymentFee     float64 `json:"total_payment_fee"`     // 总支付手续费
	TotalRefundAmount   float64 `json:"total_refund_amount"`   // 总退款金额
}

// CountMember 统计会员
func (s *statisticsSrv) CountMember(ctx context.Context, req CountReq) CountMemberResp {
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)
	memberData := statisticsRepo.CountMember(s.buildCountOpts(ctx, req)...)

	return CountMemberResp{
		TotalSaleAmount:     memberData.TotalSaleAmount.Float64,
		TotalRechargeAmount: memberData.TotalRechargeAmount.Float64,
		TotalGiveAmount:     memberData.TotalGiveAmount.Float64,
		TotalGivePoint:      memberData.TotalGivePoint.Float64,
		TotalPaymentAmount:  memberData.TotalPaymentAmount.Float64,
		TotalPaymentFee:     memberData.TotalPaymentFee.Float64,
		TotalRefundAmount:   memberData.TotalRefundAmount.Float64,
	}
}

// CountProductSaleResp 统计商品销售响应
type CountProductSaleResp struct {
	Data  []CountProductSale `json:"data"`
	Total int64              `json:"total"`
}

type CountProductSale struct {
	ProductName           string  `json:"product_name"`
	CategoryName          string  `json:"category_name"`
	TotalSaleNum          float64 `json:"total_sale_num"`
	TotalOriginSaleAmount float64 `json:"total_origin_sale_amount"`
	TotalActualSaleAmount float64 `json:"total_actual_sale_amount"`
	TotalGiveNum          float64 `json:"total_give_num"`
	TotalBusinessAmount   float64 `json:"total_business_amount"`
}

// CountProductSale 统计商品销售
func (s *statisticsSrv) CountProductSale(ctx context.Context, req CountReq) CountProductSaleResp {
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)
	productSaleData, total := statisticsRepo.CountProductSale(repository.CountProductSaleRepoReq{
		PageNo:        req.PageNo,
		PageSize:      req.PageSize,
		RankType:      req.RankType,
		RankDirection: req.RankDirection,
		Language:      ctx.GetLanguage(),
		AreaUuid:      req.AreaUuid,
		CategoryUuid:  req.CategoryUuid,
		ProductName:   req.ProductName,
	}, s.buildCountOpts(ctx, req)...)

	var data []CountProductSale
	for _, productSale := range productSaleData {
		categoryName := productSale.CategoryName.String
		if productSale.CategoryParentName.String != "" {
			categoryName = productSale.CategoryParentName.String + "-" + categoryName
		}
		data = append(data, CountProductSale{
			ProductName:           productSale.ProductName.String,
			CategoryName:          categoryName,
			TotalSaleNum:          productSale.SaleNum.Float64,
			TotalOriginSaleAmount: productSale.OriginSaleAmount.Float64,
			TotalActualSaleAmount: productSale.ActualSaleAmount.Float64,
			TotalGiveNum:          productSale.GiveNum.Float64,
			TotalBusinessAmount:   productSale.BusinessAmount.Float64,
		})
	}

	return CountProductSaleResp{
		Data:  data,
		Total: total,
	}

}

// CountFreePaymentResp 统计免单支付响应
type CountFreePaymentResp struct {
	PaymentName        string  `json:"payment_name"`
	PaymentCode        int     `json:"payment_code"`
	TotalOrderNum      int64   `json:"total_order_num"`
	TotalPaymentAmount float64 `json:"total_payment_amount"`
	TotalRefundAmount  float64 `json:"total_refund_amount"`
}

// CountFreePayment 统计免单支付
func (s *statisticsSrv) CountFreePayment(ctx context.Context, req CountReq) CountFreePaymentResp {
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)
	freePaymentData := statisticsRepo.CountFreePayment(s.buildCountOpts(ctx, req)...)

	return CountFreePaymentResp{
		PaymentName:        i18n.Translate(ctx.GetLanguage(), "免单"),
		PaymentCode:        0,
		TotalOrderNum:      freePaymentData.TotalOrderNum.Int64,
		TotalPaymentAmount: freePaymentData.TotalFreeAmount.Float64,
	}
}

// CountFreePaymentDaysResp 统计免单支付天数响应
type CountFreePaymentDaysResp struct {
	PaymentList []CountPaymentRespList `json:"payment_list"`
	Day         string                 `json:"day"`
}

// CountFreePaymentDays 统计免单支付天数
func (s *statisticsSrv) CountFreePaymentDays(ctx context.Context, req CountReq, days []string) []CountFreePaymentDaysResp {
	opts := s.buildCountOpts(ctx, req)
	paymentData := repository.NewStatisticsRepo(ctx.GetDB()).CountFreePaymentDays(opts...)

	list := make([]CountFreePaymentDaysResp, 0)
	for _, day := range days {
		paymentList := make([]CountPaymentRespList, 0)
		result, ok := slice.FindBy(paymentData, func(index int, item model.StatisticsFreePaymentDaysData) bool {
			return item.Day.String == day
		})
		if ok {
			if result.TotalFreeAmount.Float64 > 0 {
				paymentList = append(paymentList, CountPaymentRespList{
					PaymentName:        i18n.Translate(ctx.GetLanguage(), "免单"),
					PaymentCode:        0,
					TotalOrderNum:      result.TotalOrderNum.Int64,
					TotalPaymentAmount: result.TotalFreeAmount.Float64,
				})
			}
		}
		list = append(list, CountFreePaymentDaysResp{
			PaymentList: paymentList,
			Day:         day,
		})
	}
	return list
}

// CountExportResp 统计导出响应
type CountExportResp struct {
	Days []string          `json:"days"`
	Data []CountExportData `json:"data"`
}

// CountExportData 统计导出数据
type CountExportData struct {
	Day                        string                   `json:"day"`
	TotalSaleAmount            float64                  `json:"total_sale_amount"`
	TotalBusinessAmount        float64                  `json:"total_business_amount"`
	TotalServiceFee            float64                  `json:"total_service_fee"`
	TotalPaymentFee            float64                  `json:"total_payment_fee"`
	TotalTax                   float64                  `json:"total_tax"`
	TotalProductNum            float64                  `json:"total_product_num"`
	TotalMemberNum             int64                    `json:"total_member_num"`
	TotalDiscountMember        float64                  `json:"total_discount_member"`
	TotalDiscount              float64                  `json:"total_discount"`
	TotalDiscountRatio         float64                  `json:"total_discount_ratio"`
	TotalRefundAmount          float64                  `json:"total_refund_amount"`
	TotalGiftAmount            float64                  `json:"total_gift_amount"`
	TotalGiftNum               float64                  `json:"total_gift_num"`
	TotalFreeAmount            float64                  `json:"total_free_amount"`
	TotalFreeNum               float64                  `json:"total_free_num"`
	TotalReceivedAmount        float64                  `json:"total_received_amount"`
	TotalOrderNum              int64                    `json:"total_order_num"`
	TotalTakeoutSaleAmount     float64                  `json:"total_takeout_sale_amount"`
	TotalTakeoutBusinessAmount float64                  `json:"total_takeout_business_amount"`
	TotalTakeoutRefundAmount   float64                  `json:"total_takeout_refund_amount"`
	TotalTakeoutDeliveryFee    float64                  `json:"total_takeout_delivery_fee"`
	MinOrderAmount             float64                  `json:"min_order_amount"`
	MaxOrderAmount             float64                  `json:"max_order_amount"`
	AvgOrderAmount             float64                  `json:"avg_order_amount"`
	TotalDeskNum               int64                    `json:"total_desk_num"`
	TotalMealNum               int64                    `json:"total_meal_num"`
	MinDeskOrderAmount         float64                  `json:"min_desk_order_amount"`
	MaxDeskOrderAmount         float64                  `json:"max_desk_order_amount"`
	AvgDeskOrderAmount         float64                  `json:"avg_desk_order_amount"`
	TotalInstantOrderNum       int64                    `json:"total_instant_order_num"`
	MinInstantOrderAmount      float64                  `json:"min_instant_order_amount"`
	MaxInstantOrderAmount      float64                  `json:"max_instant_order_amount"`
	AvgInstantOrderAmount      float64                  `json:"avg_instant_order_amount"`
	TotalTakeoutOrderNum       int64                    `json:"total_takeout_order_num"`
	MinTakeoutOrderAmount      float64                  `json:"min_takeout_order_amount"`
	MaxTakeoutOrderAmount      float64                  `json:"max_takeout_order_amount"`
	AvgTakeoutOrderAmount      float64                  `json:"avg_takeout_order_amount"`
	AreaList                   []CountExportAreaData    `json:"area_list"`
	PaymentList                []CountExportPaymentData `json:"payment_list"`
}

// CountExportAreaData 统计导出区域数据
type CountExportAreaData struct {
	AreaID             int64   `json:"area_id"`              // 区域id
	AreaName           string  `json:"area_name"`            // 区域名称
	AreaSaleAmount     float64 `json:"area_sale_amount"`     // 区域销售额
	AreaBusinessAmount float64 `json:"area_business_amount"` // 区域营业收入
	AreaProductNum     float64 `json:"area_product_num"`     // 区域商品数量
}

// CountExportPaymentData 统计导出支付数据
type CountExportPaymentData struct {
	ID                 uint64  `json:"id"`
	Sort               int     `json:"sort"`
	CreateTime         int64   `json:"create_time"`
	PaymentName        string  `json:"payment_name"`
	PaymentCode        int     `json:"payment_code"`
	TotalOrderNum      int64   `json:"total_order_num"`
	TotalPaymentAmount float64 `json:"total_payment_amount"`
}

// CountExport 统计导出
func (s *statisticsSrv) CountExport(ctx context.Context, req CountReq) (CountExportResp, error) {
	days := s.buildDays(req)
	saleData := s.CountSaleDays(ctx, req, days)
	areaData := s.CountAreaDays(ctx, req, days)
	paymentData := s.CountPaymentDays(ctx, req, days)
	memberPaymentData := s.CountMemberPaymentDays(ctx, req, days)
	freePaymentData := s.CountFreePaymentDays(ctx, req, days)
	memberNumData := s.CountMemberNumDays(ctx, req, days)

	var data []CountExportData
	for _, sale := range saleData {
		// 统计区域数据
		areaList := make([]CountExportAreaData, 0)
		areaResult, ok := slice.FindBy(areaData, func(index int, item CountAreaDaysResp) bool {
			return item.Day == sale.Day
		})
		if ok {
			for _, area := range areaResult.AreaList {
				areaList = append(areaList, CountExportAreaData(area))
			}
		}

		// 统计支付数据
		paymentCode := make(map[int]bool)
		paymentList := make([]CountExportPaymentData, 0)
		paymentResult, ok := slice.FindBy(paymentData, func(index int, item CountPaymentDaysResp) bool {
			return item.Day == sale.Day
		})
		if ok {
			for _, payment := range paymentResult.PaymentList {
				paymentList = append(paymentList, CountExportPaymentData(payment))
				paymentCode[payment.PaymentCode] = true
			}
		}

		// 统计会员数据
		memberResult, ok := slice.FindBy(memberPaymentData, func(index int, item CountPaymentDaysResp) bool {
			return item.Day == sale.Day
		})
		if ok {
			for _, member := range memberResult.PaymentList {
				if _, ok := paymentCode[member.PaymentCode]; !ok {
					paymentList = append(paymentList, CountExportPaymentData(member))
					paymentCode[member.PaymentCode] = true
				} else {
					i := 0
					payment, ok := slice.FindBy(paymentList, func(index int, item CountExportPaymentData) bool {
						i = index
						return item.PaymentCode == member.PaymentCode
					})
					if ok {
						payment.TotalOrderNum += member.TotalOrderNum
						payment.TotalPaymentAmount += member.TotalPaymentAmount
					}
					paymentList[i] = payment
				}
			}
		}

		// 统计免单数据
		freeResult, ok := slice.FindBy(freePaymentData, func(index int, item CountFreePaymentDaysResp) bool {
			return item.Day == sale.Day
		})
		if ok {
			for _, free := range freeResult.PaymentList {
				paymentList = append(paymentList, CountExportPaymentData(free))
			}
		}

		// 统计会员数量
		var totalMemberNum int64
		memberNumResult, ok := slice.FindBy(memberNumData, func(index int, item CountMemberNumDaysResp) bool {
			return item.Day == sale.Day
		})
		if ok {
			totalMemberNum = memberNumResult.MemberNum
		}

		data = append(data, CountExportData{
			Day:                        sale.Day,
			TotalSaleAmount:            sale.TotalSaleAmount,
			TotalBusinessAmount:        sale.TotalBusinessAmount,
			TotalServiceFee:            sale.TotalServiceFee,
			TotalPaymentFee:            sale.TotalPaymentFee,
			TotalTax:                   sale.TotalTax,
			TotalProductNum:            sale.TotalProductNum,
			TotalMemberNum:             totalMemberNum,
			TotalDiscountMember:        sale.TotalDiscountMember,
			TotalDiscount:              sale.TotalDiscount,
			TotalDiscountRatio:         sale.TotalDiscountRatio,
			TotalRefundAmount:          sale.TotalRefundAmount,
			TotalGiftAmount:            sale.TotalGiftAmount,
			TotalGiftNum:               sale.TotalGiftNum,
			TotalFreeAmount:            sale.TotalFreeAmount,
			TotalFreeNum:               sale.TotalFreeNum,
			TotalTakeoutSaleAmount:     sale.TotalTakeoutSaleAmount,
			TotalTakeoutBusinessAmount: sale.TotalTakeoutBusinessAmount,
			TotalTakeoutRefundAmount:   sale.TotalTakeoutRefundAmount,
			TotalTakeoutDeliveryFee:    sale.TotalTakeoutDeliveryFee,
			TotalReceivedAmount:        sale.TotalReceivedAmount,
			TotalOrderNum:              sale.TotalOrderNum,
			MinOrderAmount:             sale.MinOrderAmount,
			MaxOrderAmount:             sale.MaxOrderAmount,
			AvgOrderAmount:             sale.AvgOrderAmount,
			TotalDeskNum:               sale.TotalDeskNum,
			TotalMealNum:               sale.TotalMealNum,
			MinDeskOrderAmount:         sale.MinDeskOrderAmount,
			MaxDeskOrderAmount:         sale.MaxDeskOrderAmount,
			AvgDeskOrderAmount:         sale.AvgDeskOrderAmount,
			TotalInstantOrderNum:       sale.TotalInstantOrderNum,
			MinInstantOrderAmount:      sale.MinInstantOrderAmount,
			MaxInstantOrderAmount:      sale.MaxInstantOrderAmount,
			AvgInstantOrderAmount:      sale.AvgInstantOrderAmount,
			TotalTakeoutOrderNum:       sale.TotalTakeoutOrderNum,
			MinTakeoutOrderAmount:      sale.MinTakeoutOrderAmount,
			MaxTakeoutOrderAmount:      sale.MaxTakeoutOrderAmount,
			AvgTakeoutOrderAmount:      sale.AvgTakeoutOrderAmount,
			AreaList:                   areaList,
			PaymentList:                paymentList,
		})
	}

	return CountExportResp{
		Days: days,
		Data: data,
	}, nil
}

// CountShiftRefundAmount 统计班次退款金额
func (s *statisticsSrv) CountShiftRefundAmount(ctx context.Context, req CountReq) float64 {
	commonRepo := repository.NewCommonRepoImpl()
	returnOrderRepo := repository.NewReturnOrderRepoImpl(ctx.GetDB())
	return returnOrderRepo.SumRefundAmount(
		commonRepo.WhereByDutyNo(req.DutyNo),
		commonRepo.WhereBySoftDelete(),
	)
}

// CountCancelOrderResp 统计取消订单响应
type CountCancelOrderResp struct {
	TotalCancelOrderNum    int64   `json:"total_cancel_order_num"`    // 总取消订单数
	TotalCancelOrderAmount float64 `json:"total_cancel_order_amount"` // 总取消订单金额
}

// CountCancelOrder 统计取消订单
func (s *statisticsSrv) CountCancelOrder(ctx context.Context, req CountReq) CountCancelOrderResp {
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)
	req.IsUpdateTime = true
	cancelOrderData := statisticsRepo.CountCancelOrder(s.buildCountOpts(ctx, req)...)

	return CountCancelOrderResp{
		TotalCancelOrderNum:    cancelOrderData.TotalCancelOrderNum.Int64,
		TotalCancelOrderAmount: cancelOrderData.TotalCancelOrderAmount.Float64,
	}
}
