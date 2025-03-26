package service

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/shopspring/decimal"
)

// IStatisticsSrv 统计服务接口
type IStatisticsSrv interface {
	CountSale(ctx context.Context, req CountSaleReq) CountSaleResp          // 统计销售
	CountPayment(ctx context.Context, req CountPaymentReq) CountPaymentResp // 统计支付
	SaveSale(ctx context.Context, req SaveSaleReq) error                    // 保存销售
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

// CountSaleReq 统计销售请求
type CountSaleReq struct {
	ShiftNo        string `json:"shift_no"`         // 当班编号
	QueryStartTime int64  `json:"query_start_time"` // 查询开始时间
	QueryEndTime   int64  `json:"query_end_time"`   // 查询结束时间
}

// CountSaleResp 统计销售响应
type CountSaleResp struct {
	TotalSaleAmount          float64 `json:"total_sale_amount"`            // 总销售额
	TotalReceivedAmount      float64 `json:"total_received_amount"`        // 总实收金额
	TotalProductPrice        float64 `json:"total_product_price"`          // 总商品原价
	TotalProductNum          int64   `json:"total_product_num"`            // 总商品数量
	TotalDiscountMember      float64 `json:"total_discount_member"`        // 总会员折扣
	TotalBusinessAmount      float64 `json:"total_business_amount"`        // 总营业收入
	TotalServiceFee          float64 `json:"total_service_fee"`            // 总服务费
	TotalPaymentFee          float64 `json:"total_payment_fee"`            // 总支付手续费
	TotalTax                 float64 `json:"total_tax"`                    // 总税额
	TotalRefundAmount        float64 `json:"total_refund_amount"`          // 总退款金额
	TotalDiscount            float64 `json:"total_discount"`               // 总优惠折扣
	TotalDiscountRatio       float64 `json:"total_discount_ratio"`         // 总优惠折扣率
	TotalGiftAmount          float64 `json:"total_gift_amount"`            // 总赠菜金额
	TotalGiftNum             int64   `json:"total_gift_num"`               // 总赠菜数量
	TotalFreeAmount          float64 `json:"total_free_amount"`            // 总免单金额
	TotalFreeNum             int64   `json:"total_free_num"`               // 总免单数量
	TotalOrderNum            int64   `json:"total_order_num"`              // 总订单数量
	TotalDeskNum             int64   `json:"total_desk_num"`               // 总桌台数量
	TotalMealNum             int64   `json:"total_meal_num"`               // 总用餐人数
	TotalInstantOrderNum     int64   `json:"total_instant_order_num"`      // 总即时订单数量
	TotalInstantOrderAmount  float64 `json:"total_instant_order_amount"`   // 总即时订单金额
	MinOrderAmount           float64 `json:"min_order_amount"`             // 最小订单金额
	MaxOrderAmount           float64 `json:"max_order_amount"`             // 最大订单金额
	AvgOrderAmount           float64 `json:"avg_order_amount"`             // 平均订单金额
	MinDeskOrderAmount       float64 `json:"min_desk_order_amount"`        // 最小桌台订单金额
	MaxDeskOrderAmount       float64 `json:"max_desk_order_amount"`        // 最大桌台订单金额
	AvgDeskOrderAmount       float64 `json:"avg_desk_order_amount"`        // 平均桌台订单金额
	AvgDeskPeopleOrderAmount float64 `json:"avg_desk_people_order_amount"` // 平均桌台每人订单金额
	MinInstantOrderAmount    float64 `json:"min_instant_order_amount"`     // 最小即时订单金额
	MaxInstantOrderAmount    float64 `json:"max_instant_order_amount"`     // 最大即时订单金额
	AvgInstantOrderAmount    float64 `json:"avg_instant_order_amount"`     // 平均即时订单金额
}

// CountSale 统计销售
func (s *statisticsSrv) CountSale(ctx context.Context, req CountSaleReq) CountSaleResp {
	var (
		opts           []repository.DBOption
		statisticsRepo = repository.NewStatisticsRepo(ctx.GetDB())
		commonRepo     = repository.NewCommonRepo()
	)

	if req.ShiftNo != "" {
		opts = append(opts, commonRepo.WhereByShiftNo(req.ShiftNo))
	}
	if req.QueryStartTime > 0 && req.QueryEndTime > 0 {
		opts = append(opts, commonRepo.WhereBetweenByCompleteTime(req.QueryStartTime, req.QueryEndTime))
	}

	saleData := statisticsRepo.CountSale(opts...)

	// 总优惠折扣率 = 总优惠折扣 / 总销售额
	var discountRatio decimal.Decimal
	if saleData.TotalSaleAmount.Float64 > 0 {
		discountRatio = decimal.NewFromFloat(saleData.TotalDiscount.Float64).Div(decimal.NewFromFloat(saleData.TotalSaleAmount.Float64))
	}

	// 平均桌台每人订单金额 = 总桌台订单金额 / 总桌台数量 / 总用餐人数
	var avgDeskPeopleOrderAmount decimal.Decimal
	if saleData.TotalMealNum.Int64 > 0 {
		avgDeskPeopleOrderAmount = decimal.NewFromFloat(saleData.TotalDeskOrderAmount.Float64).Div(decimal.NewFromInt(saleData.TotalMealNum.Int64))
	}

	return CountSaleResp{
		TotalSaleAmount:          saleData.TotalSaleAmount.Float64,
		TotalReceivedAmount:      saleData.TotalReceivedAmount.Float64,
		TotalProductPrice:        saleData.TotalProductPrice.Float64,
		TotalProductNum:          saleData.TotalProductNum.Int64,
		TotalDiscountMember:      saleData.TotalDiscountMember.Float64,
		TotalBusinessAmount:      saleData.TotalBusinessAmount.Float64,
		TotalServiceFee:          saleData.TotalServiceFee.Float64,
		TotalPaymentFee:          saleData.TotalPaymentFee.Float64,
		TotalTax:                 saleData.TotalTax.Float64,
		TotalRefundAmount:        saleData.TotalRefundAmount.Float64,
		TotalDiscount:            saleData.TotalDiscount.Float64,
		TotalDiscountRatio:       discountRatio.Round(2).InexactFloat64(),
		TotalGiftAmount:          saleData.TotalGiftAmount.Float64,
		TotalGiftNum:             saleData.TotalGiftNum.Int64,
		TotalFreeAmount:          saleData.TotalFreeAmount.Float64,
		TotalFreeNum:             saleData.TotalFreeNum.Int64,
		TotalOrderNum:            saleData.TotalOrderNum.Int64,
		TotalDeskNum:             saleData.TotalDeskNum.Int64,
		TotalMealNum:             saleData.TotalMealNum.Int64,
		TotalInstantOrderNum:     saleData.TotalInstantOrderNum.Int64,
		TotalInstantOrderAmount:  saleData.TotalInstantOrderAmount.Float64,
		MinOrderAmount:           saleData.MinOrderAmount.Float64,
		MaxOrderAmount:           saleData.MaxOrderAmount.Float64,
		AvgOrderAmount:           saleData.AvgOrderAmount.Float64,
		MinDeskOrderAmount:       saleData.MinDeskOrderAmount.Float64,
		MaxDeskOrderAmount:       saleData.MaxDeskOrderAmount.Float64,
		AvgDeskOrderAmount:       saleData.AvgDeskOrderAmount.Float64,
		AvgDeskPeopleOrderAmount: avgDeskPeopleOrderAmount.Round(2).InexactFloat64(),
		MinInstantOrderAmount:    saleData.MinInstantOrderAmount.Float64,
		MaxInstantOrderAmount:    saleData.MaxInstantOrderAmount.Float64,
		AvgInstantOrderAmount:    saleData.AvgInstantOrderAmount.Float64,
	}
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
	PaymentName        string  `json:"payment_name"`         // 支付方式名称
	PaymentCode        int     `json:"payment_code"`         // 支付方式编码
	TotalOrderNum      int64   `json:"total_order_num"`      // 总订单数量
	TotalPaymentAmount float64 `json:"total_payment_amount"` // 总支付金额
}

// CountPayment 统计支付
func (s *statisticsSrv) CountPayment(ctx context.Context, req CountPaymentReq) CountPaymentResp {
	var (
		opts           []repository.DBOption
		list           []CountPaymentRespList
		statisticsRepo = repository.NewStatisticsRepo(ctx.GetDB())
		commonRepo     = repository.NewCommonRepo()
	)

	if req.ShiftNo != "" {
		opts = append(opts, commonRepo.WhereByShiftNo(req.ShiftNo))
	}
	if req.QueryStartTime > 0 && req.QueryEndTime > 0 {
		opts = append(opts, commonRepo.WhereBetweenByCompleteTime(req.QueryStartTime, req.QueryEndTime))
	}

	paymentData := statisticsRepo.CountPayment(opts...)

	var (
		totalReceivedAmount decimal.Decimal
		totalRefundAmount   decimal.Decimal
	)
	for _, payment := range paymentData {
		list = append(list, CountPaymentRespList{
			PaymentName:        payment.PaymentName,
			PaymentCode:        payment.PaymentCode,
			TotalOrderNum:      payment.TotalOrderNum.Int64,
			TotalPaymentAmount: payment.TotalPaymentAmount.Float64,
		})
		if payment.PaymentCode != 10 {
			totalReceivedAmount = totalReceivedAmount.Add(decimal.NewFromFloat(payment.TotalPaymentAmount.Float64))
		}
		totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(payment.TotalRefundAmount.Float64))
	}

	return CountPaymentResp{
		TotalReceivedAmount: totalReceivedAmount.Round(2).InexactFloat64(),
		TotalRefundAmount:   totalRefundAmount.Round(2).InexactFloat64(),
		PaymentList:         list,
	}
}

// SaveSaleReq 保存销售请求
type SaveSaleReq struct {
	SaleBill *model.SaleBill
}

// SaveSale 保存销售
func (s *statisticsSrv) SaveSale(ctx context.Context, req SaveSaleReq) error {
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)

	// 先删除
	statisticsRepo.DeleteSale(req.SaleBill.Uuid)
	statisticsRepo.DeletePayment(req.SaleBill.Uuid)

	var (
		sales    []model.StatisticsSale
		payments []model.StatisticsPayment
	)
	// 销售订单
	for _, saleOrder := range req.SaleBill.SaleOrders {
		var (
			productNum       int
			giveNum          int
			freeNum          int
			productPrice     decimal.Decimal
			productSalePrice decimal.Decimal
			productTax       decimal.Decimal
			serviceFee       decimal.Decimal
			serviceTax       decimal.Decimal
			freeAmount       decimal.Decimal
			refundAmount     decimal.Decimal
		)
		// 销售商品
		for _, saleProduct := range saleOrder.SaleOrderProducts {
			productNum += int(saleProduct.Num)
			productPrice = productPrice.Add(decimal.NewFromFloat(saleProduct.GetUnitPriceNoneTax()))
			productSalePrice = productSalePrice.Add(decimal.NewFromFloat(saleProduct.SalePrice))
			productTax = productTax.Add(decimal.NewFromFloat(saleProduct.TaxFee))
			serviceFee = serviceFee.Add(decimal.NewFromFloat(saleProduct.ServiceFee))
			serviceTax = serviceTax.Add(decimal.NewFromFloat(saleProduct.ServiceTaxFee))
			if saleProduct.GiftTime > 0 {
				giveNum += int(saleProduct.Num)
			}
		}
		if saleOrder.IsFree > 0 {
			freeNum = 1
			freeAmount = decimal.NewFromFloat(saleOrder.Amount)
		}
		// 支付订单
		for _, salePayment := range saleOrder.PaymentOrders {
			var paymentRefundAmount decimal.Decimal
			for _, refundOrder := range salePayment.ReturnOrderAmounts {
				if refundOrder.RefundStatus == 1 {
					refundAmount = refundAmount.Add(decimal.NewFromFloat(refundOrder.Amount))
					paymentRefundAmount = paymentRefundAmount.Add(decimal.NewFromFloat(refundOrder.Amount))
				}
			}
			payments = append(payments, model.StatisticsPayment{
				SaleBillUuid:      req.SaleBill.Uuid,
				DutyNo:            req.SaleBill.DutyNo,
				DeskUuid:          req.SaleBill.DeskUuid,
				SaleOrderUuid:     saleOrder.Uuid,
				PaymentMethodUuid: salePayment.PaymentMethodUuid,
				PaymentAmount:     salePayment.Amount,
				RefundAmount:      paymentRefundAmount.Round(2).InexactFloat64(),
				CompleteTime:      req.SaleBill.FinishTime,
			})
		}
		sale := model.StatisticsSale{
			SaleBillUuid:     req.SaleBill.Uuid,
			DutyNo:           req.SaleBill.DutyNo,
			DeskUuid:         req.SaleBill.DeskUuid,
			SaleOrderUuid:    saleOrder.Uuid,
			MealNum:          int(req.SaleBill.MealNum),
			ProductPrice:     productPrice.Round(2).InexactFloat64(),
			ProductSalePrice: productSalePrice.Round(2).InexactFloat64(),
			ProductNum:       productNum,
			ProductTax:       productTax.Round(2).InexactFloat64(),
			ServiceFee:       serviceFee.Round(2).InexactFloat64(),
			ServiceTax:       serviceTax.Round(2).InexactFloat64(),
			Discount:         saleOrder.CustomDiscountFee,
			DiscountMember:   saleOrder.MemberDiscountFee,
			GiftAmount:       saleOrder.GiftAmount,
			GiftNum:          giveNum,
			FreeAmount:       freeAmount.Round(2).InexactFloat64(),
			FreeNum:          freeNum,
			PaymentAmount:    saleOrder.Amount,
			PaymentFee:       saleOrder.PaymentCommissionFee,
			PaymentBalance:   saleOrder.MemberBalance,
			RefundAmount:     refundAmount.Round(2).InexactFloat64(),
			CompleteTime:     req.SaleBill.FinishTime,
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
	return nil
}
