package service

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/shopspring/decimal"
)

// IStatisticsSrv 统计服务接口
type IStatisticsSrv interface {
	CountSale(ctx context.Context, req CountReq) CountSaleResp               // 统计销售
	CountPayment(ctx context.Context, req CountReq) CountPaymentResp         // 统计支付
	CountTax(ctx context.Context, req CountReq) []CountTaxResp               // 统计税类
	CountCategory(ctx context.Context, req CountReq) CountCategoryResp       // 统计分类
	CountProduct(ctx context.Context, req CountReq) []CountProductResp       // 统计商品
	CountArea(ctx context.Context, req CountReq) []CountAreaResp             // 统计区域
	Count7Days(ctx context.Context, req CountReq) Count7DaysResp             // 统计销售天数
	CountMemberNum(ctx context.Context, req CountReq) int64                  // 统计会员数量
	CountMember(ctx context.Context, req CountReq) CountMemberResp           // 统计会员
	CountMemberPayment(ctx context.Context, req CountReq) CountPaymentResp   // 统计会员支付
	CountUnpaidOrder(ctx context.Context, req CountReq) CountUnpaidOrderResp // 统计未结订单
	CountProductSale(ctx context.Context, req CountReq) CountProductSaleResp // 统计商品销售
	RankProduct(ctx context.Context, req CountReq) []CountProductRankResp    // 统计商品排行
	SaveSale(ctx context.Context, req SaveSaleReq) error                     // 保存销售
	SaveMember(ctx context.Context, req SaveMemberReq) error                 // 保存会员
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
func (s *statisticsSrv) CountSale(ctx context.Context, req CountReq) CountSaleResp {
	db := ctx.GetDB()
	opts := s.buildCountOpts(req)
	saleData := repository.NewStatisticsRepo(db).CountSale(opts...)
	memberData := s.CountMember(ctx, req)

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

	totalSaleAmount := decimal.NewFromFloat(saleData.TotalSaleAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalSaleAmount))
	totalReceivedAmount := decimal.NewFromFloat(saleData.TotalReceivedAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalPaymentAmount))
	totalPaymentFee := decimal.NewFromFloat(saleData.TotalPaymentFee.Float64).Add(decimal.NewFromFloat(memberData.TotalPaymentFee))
	totalRefundAmount := decimal.NewFromFloat(saleData.TotalRefundAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalRefundAmount))
	totalBusinessAmount := decimal.NewFromFloat(saleData.TotalBusinessAmount.Float64).Add(decimal.NewFromFloat(memberData.TotalPaymentFee))

	return CountSaleResp{
		TotalSaleAmount:          totalSaleAmount.Round(2).InexactFloat64(),
		TotalReceivedAmount:      totalReceivedAmount.Round(2).InexactFloat64(),
		TotalProductPrice:        saleData.TotalProductPrice.Float64,
		TotalProductNum:          saleData.TotalProductNum.Int64,
		TotalDiscountMember:      saleData.TotalDiscountMember.Float64,
		TotalBusinessAmount:      totalBusinessAmount.Round(2).InexactFloat64(),
		TotalServiceFee:          saleData.TotalServiceFee.Float64,
		TotalPaymentFee:          totalPaymentFee.Round(2).InexactFloat64(),
		TotalTax:                 saleData.TotalTax.Float64,
		TotalRefundAmount:        totalRefundAmount.Round(2).InexactFloat64(),
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
func (s *statisticsSrv) CountPayment(ctx context.Context, req CountReq) CountPaymentResp {
	opts := s.buildCountOpts(req)
	paymentData := repository.NewStatisticsRepo(ctx.GetDB()).CountPayment(opts...)
	memberPaymentData := s.CountMemberPayment(ctx, req)

	var (
		totalReceivedAmount decimal.Decimal
		totalRefundAmount   decimal.Decimal
		list                = make([]CountPaymentRespList, 0)
	)

	for _, payment := range paymentData {
		item, ok := slice.Find(list, func(index int, item CountPaymentRespList) bool {
			return item.PaymentCode == payment.PaymentCode
		})
		if !ok {
			list = append(list, CountPaymentRespList{
				PaymentName:        payment.PaymentName,
				PaymentCode:        payment.PaymentCode,
				TotalOrderNum:      payment.TotalOrderNum.Int64,
				TotalPaymentAmount: payment.TotalPaymentAmount.Float64,
			})
		} else {
			item.TotalOrderNum += payment.TotalOrderNum.Int64
			item.TotalPaymentAmount += payment.TotalPaymentAmount.Float64
		}
		if payment.PaymentCode != 10 {
			totalReceivedAmount = totalReceivedAmount.Add(decimal.NewFromFloat(payment.TotalPaymentAmount.Float64))
		}
		totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(payment.TotalRefundAmount.Float64))
	}

	for _, memberPayment := range memberPaymentData.PaymentList {
		item, ok := slice.Find(list, func(index int, item CountPaymentRespList) bool {
			return item.PaymentCode == memberPayment.PaymentCode
		})
		if !ok {
			list = append(list, CountPaymentRespList{
				PaymentName:        memberPayment.PaymentName,
				PaymentCode:        memberPayment.PaymentCode,
				TotalOrderNum:      memberPayment.TotalOrderNum,
				TotalPaymentAmount: memberPayment.TotalPaymentAmount,
			})
		} else {
			item.TotalOrderNum += memberPayment.TotalOrderNum
			item.TotalPaymentAmount += memberPayment.TotalPaymentAmount
		}
	}

	totalReceivedAmount = totalReceivedAmount.Add(decimal.NewFromFloat(memberPaymentData.TotalReceivedAmount))
	totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(memberPaymentData.TotalRefundAmount))

	return CountPaymentResp{
		TotalReceivedAmount: totalReceivedAmount.Round(2).InexactFloat64(),
		TotalRefundAmount:   totalRefundAmount.Round(2).InexactFloat64(),
		PaymentList:         list,
	}
}

// CountMemberPayment 统计会员支付
func (s *statisticsSrv) CountMemberPayment(ctx context.Context, req CountReq) CountPaymentResp {
	opts := s.buildCountOpts(req)
	paymentData := repository.NewStatisticsRepo(ctx.GetDB()).CountMemberPayment(opts...)

	var (
		totalReceivedAmount decimal.Decimal
		totalRefundAmount   decimal.Decimal
		list                = make([]CountPaymentRespList, 0)
	)
	for _, payment := range paymentData {
		list = append(list, CountPaymentRespList{
			PaymentName:        payment.PaymentName,
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

// CountTaxResp 统计税类响应
type CountTaxResp struct {
	TaxRate            float64 `json:"tax_rate"`             // 税类
	TotalTaxFee        float64 `json:"total_tax_fee"`        // 总税费
	TotalProductAmount float64 `json:"total_product_amount"` // 总商品金额: 含税
}

// CountTax 统计税类
func (s *statisticsSrv) CountTax(ctx context.Context, req CountReq) []CountTaxResp {
	opts := s.buildCountOpts(req)
	opts = append(opts, repository.NewCommonRepo().WhereByRefundTime(0))
	taxData := repository.NewStatisticsRepo(ctx.GetDB()).CountTax(opts...)

	var list []CountTaxResp
	for _, tax := range taxData {
		list = append(list, CountTaxResp{
			TaxRate:            tax.TaxRate.Float64,
			TotalTaxFee:        tax.TotalTaxFee.Float64,
			TotalProductAmount: tax.TotalProductAmount.Float64,
		})
	}

	return list
}

type CountCategoryResp struct {
	TotalSaleNum int64                   `json:"total_sale_num"` // 总销售数量
	CategoryList []CountCategoryListResp `json:"category_list"`  // 分类列表
}

// CountCategoryResp 统计分类响应
type CountCategoryListResp struct {
	CategoryName string  `json:"category_name"` // 分类名称
	SaleNum      int64   `json:"sale_num"`      // 销售数量
	SaleAmount   float64 `json:"sale_amount"`   // 销售金额
}

// CountCategory 统计分类
func (s *statisticsSrv) CountCategory(ctx context.Context, req CountReq) CountCategoryResp {
	var (
		SaleNum int64
		list    []CountCategoryListResp
	)

	opts := s.buildCountOpts(req)
	categoryData := repository.NewStatisticsRepo(ctx.GetDB()).CountCategory(req.CategoryType, ctx.GetLanguage(), opts...)

	for _, category := range categoryData {
		categoryName := category.CategoryParentName.String
		if category.CategoryName.String != "" {
			categoryName = categoryName + "-" + category.CategoryName.String
		}
		list = append(list, CountCategoryListResp{
			CategoryName: categoryName,
			SaleNum:      category.SaleNum.Int64,
			SaleAmount:   category.SaleAmount.Float64,
		})
		SaleNum += category.SaleNum.Int64
	}

	return CountCategoryResp{
		TotalSaleNum: SaleNum,
		CategoryList: list,
	}
}

// CountProductResp 统计商品响应
type CountProductResp struct {
	ProductName string  `json:"product_name"` // 商品名称
	SalePrice   float64 `json:"sale_price"`   // 销售单价
	SaleNum     int64   `json:"sale_num"`     // 销售数量
	SaleAmount  float64 `json:"sale_amount"`  // 销售金额
}

// CountProduct 统计商品
func (s *statisticsSrv) CountProduct(ctx context.Context, req CountReq) []CountProductResp {
	opts := s.buildCountOpts(req)
	productData := repository.NewStatisticsRepo(ctx.GetDB()).CountProduct(ctx.GetLanguage(), opts...)

	var list []CountProductResp
	for _, product := range productData {
		list = append(list, CountProductResp{
			ProductName: product.ProductName.String + "（" + product.FlavorName.String + "）",
			SalePrice:   product.SalePrice.Float64,
			SaleNum:     product.SaleNum.Int64,
			SaleAmount:  product.SaleAmount.Float64,
		})
	}
	return list
}

type CountAreaResp struct {
	AreaName           string  `json:"area_name"`            // 区域名称
	AreaSaleAmount     float64 `json:"area_sale_amount"`     // 区域销售额
	AreaBusinessAmount float64 `json:"area_business_amount"` // 区域营业收入
	AreaProductNum     int64   `json:"area_product_num"`     // 区域商品数量
}

// CountArea 统计区域
func (s *statisticsSrv) CountArea(ctx context.Context, req CountReq) []CountAreaResp {
	opts := s.buildCountOpts(req)
	areaData := repository.NewStatisticsRepo(ctx.GetDB()).CountArea(opts...)

	var list []CountAreaResp
	for _, area := range areaData {
		list = append(list, CountAreaResp{
			AreaName:           area.AreaName.String,
			AreaSaleAmount:     area.AreaSaleAmount.Float64,
			AreaBusinessAmount: area.AreaBusinessAmount.Float64,
			AreaProductNum:     area.AreaProductNum.Int64,
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
	opts := s.buildCountOpts(req)
	sevenDayData := repository.NewStatisticsRepo(ctx.GetDB()).Count7Days(opts...)

	days := s.buildDays(req)
	sevenDayList := make([]Count7DaysDataResp, 0, len(sevenDayData))
	for _, day := range days {
		oneDayData := Count7DaysDataResp{
			Day:        day,
			TotalNum:   0,
			TotalMoney: 0,
		}
		result, ok := slice.FindBy(sevenDayData, func(index int, dayData model.Statistics7DaysData) bool {
			return dayData.Day.String == day
		})
		if ok {
			oneDayData.TotalNum = result.TotalOrderNum.Int64
			oneDayData.TotalMoney = result.TotalReceivedAmount.Float64
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
	req.IsCreateTime = true
	opts := s.buildCountOpts(req)
	return repository.NewStatisticsRepo(ctx.GetDB()).CountMemberNum(opts...)
}

type CountUnpaidOrderResp struct {
	TotalOrderNum int64   `json:"total_order_num"` // 总订单数
	TotalAmount   float64 `json:"total_amount"`    // 总金额
}

// CountUnpaidOrder 统计未结订单
func (s *statisticsSrv) CountUnpaidOrder(ctx context.Context, req CountReq) CountUnpaidOrderResp {
	req.IsCreateTime = true
	opts := s.buildCountOpts(req)
	unpaidOrderData := repository.NewStatisticsRepo(ctx.GetDB()).CountUnpaidOrder(opts...)

	return CountUnpaidOrderResp{
		TotalOrderNum: unpaidOrderData.TotalOrderNum.Int64,
		TotalAmount:   unpaidOrderData.TotalAmount.Float64,
	}
}

// CountProductRankResp 统计商品排行响应
type CountProductRankResp struct {
	ProductName string  `json:"product_name"` // 商品名称
	SaleNum     int64   `json:"sale_num"`     // 销售数量
	SaleAmount  float64 `json:"sale_amount"`  // 销售金额
}

// CountProductRank 统计商品排行
func (s *statisticsSrv) RankProduct(ctx context.Context, req CountReq) []CountProductRankResp {
	opts := s.buildCountOpts(req)
	productData := repository.NewStatisticsRepo(ctx.GetDB()).RankProduct(req.RankType, ctx.GetLanguage(), opts...)
	var list []CountProductRankResp
	for _, product := range productData {
		list = append(list, CountProductRankResp{
			ProductName: product.ProductName.String + "（" + product.FlavorName.String + "）",
			SaleNum:     product.SaleNum.Int64,
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
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(ctx.GetCompanyUuid())
	statisticsRepo := repository.NewStatisticsRepo(db)

	// 先删除
	statisticsRepo.DeleteSale(req.SaleBillUuid)
	statisticsRepo.DeletePayment(req.SaleBillUuid)
	statisticsRepo.DeleteProduct(req.SaleBillUuid)

	if req.OnlyDelete {
		return nil
	}

	// 查询销售账单详情
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	saleBill.CalcAll()
	if err != nil {
		return nil
	}

	var (
		sales    []model.StatisticsSale
		payments []model.StatisticsPayment
		products []model.StatisticsProduct
	)
	// 销售订单
	for _, saleOrder := range saleBill.SaleOrders {
		var (
			productNum           int
			giveNum              int
			freeNum              int
			refundNum            int
			productPrice         decimal.Decimal
			productSalePrice     decimal.Decimal
			productTax           decimal.Decimal
			serviceFee           decimal.Decimal
			serviceTax           decimal.Decimal
			freeAmount           decimal.Decimal
			refundAmount         decimal.Decimal
			paymentBalance       decimal.Decimal
			refundTax            decimal.Decimal
			refundServiceFee     decimal.Decimal
			refundDiscount       decimal.Decimal
			refundDiscountMember decimal.Decimal
			refundFee            decimal.Decimal
		)
		// 销售商品
		for _, saleProduct := range saleOrder.SaleOrderProducts {
			if saleProduct.CancelTime == 0 {
				productNum += int(saleProduct.Num)
				productPrice = productPrice.Add(decimal.NewFromFloat(saleProduct.GetUnitPriceNoneTax()))
				productSalePrice = productSalePrice.Add(decimal.NewFromFloat(saleProduct.SalePrice))
				productTax = productTax.Add(decimal.NewFromFloat(saleProduct.TaxFee))
				productGiveNum := 0
				productFreeNum := 0
				serviceFee = serviceFee.Add(decimal.NewFromFloat(saleProduct.ServiceFee))
				serviceTax = serviceTax.Add(decimal.NewFromFloat(saleProduct.ServiceTaxFee))
				if saleProduct.GiftTime > 0 {
					giveNum += int(saleProduct.Num)
					productGiveNum = int(saleProduct.Num)
				}
				if saleOrder.IsFree > 0 {
					productFreeNum = int(saleProduct.Num)
				}
				var bomUuid uint64
				for _, productBom := range saleProduct.SaleOrderProductBoms {
					if productBom.IsFlavorBom == 1 {
						bomUuid = productBom.ProductBomUuid
					}
				}

				for _, refundProduct := range saleProduct.ReturnOrderProducts {
					refundNum += int(refundProduct.Num)
					refundTax = refundTax.Add(decimal.NewFromFloat(saleProduct.TaxFee).Mul(decimal.NewFromFloat(float64(refundProduct.Num))))
					refundServiceFee = refundServiceFee.Add(decimal.NewFromFloat(saleProduct.ServiceFee).Mul(decimal.NewFromFloat(float64(refundProduct.Num))))
					refundDiscount = refundDiscount.Add(decimal.NewFromFloat(saleProduct.DiscountFee).Mul(decimal.NewFromFloat(float64(refundProduct.Num))))
					refundDiscountMember = refundDiscountMember.Add(decimal.NewFromFloat(saleProduct.MemberDiscountFee).Mul(decimal.NewFromFloat(float64(refundProduct.Num))))
				}

				products = append(products, model.StatisticsProduct{
					SaleBillUuid:       saleBill.Uuid,
					SaleOrderUuid:      saleOrder.Uuid,
					DutyNo:             saleBill.DutyNo,
					DeskUuid:           saleBill.DeskUuid,
					ProductPackageUuid: saleProduct.ProductPackageUuid,
					ProductBomUuid:     bomUuid,
					ProductPrice:       saleProduct.GetUnitPriceNoneTax(),
					ProductSalePrice:   saleProduct.ProductPrice,
					ProductFinalPrice:  saleProduct.Price,
					ProductNum:         int(saleProduct.Num),
					TaxRate:            saleProduct.TaxRate,
					TaxFee:             saleProduct.TaxFee,
					ServiceFee:         saleProduct.ServiceFee,
					ServiceTax:         saleProduct.ServiceTaxFee,
					GiveNum:            productGiveNum,
					FreeNum:            productFreeNum,
					CompleteTime:       saleBill.FinishTime,
					RefundNum:          refundNum,
				})
			}
		}
		if saleOrder.IsFree > 0 {
			freeNum = 1
			freeAmount = decimal.NewFromFloat(saleOrder.Amount)
		}
		if saleOrder.GetCanReturnAmount() == 0 {
			refundFee = decimal.NewFromFloat(saleOrder.PaymentCommissionFee)
		}
		// 支付订单
		for _, salePayment := range saleOrder.PaymentOrders {
			var paymentRefundAmount decimal.Decimal
			for _, refundOrderAmount := range salePayment.ReturnOrderAmounts {
				refundAmount = refundAmount.Add(decimal.NewFromFloat(refundOrderAmount.Amount))
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
			ProductPrice:         productPrice.Round(2).InexactFloat64(),
			ProductSalePrice:     productSalePrice.Round(2).InexactFloat64(),
			ProductNum:           productNum,
			ProductTax:           productTax.Round(2).InexactFloat64(),
			ServiceFee:           serviceFee.Round(2).InexactFloat64(),
			ServiceTax:           serviceTax.Round(2).InexactFloat64(),
			Discount:             saleOrder.CustomDiscountFee,
			DiscountMember:       saleOrder.MemberDiscountFee,
			GiftAmount:           saleOrder.GiftAmount,
			GiftNum:              giveNum,
			FreeAmount:           freeAmount.Round(2).InexactFloat64(),
			FreeNum:              freeNum,
			PaymentAmount:        saleOrder.Amount,
			PaymentFee:           saleOrder.PaymentCommissionFee,
			PaymentBalance:       paymentBalance.Round(2).InexactFloat64(),
			RefundAmount:         refundAmount.Round(2).InexactFloat64(),
			RefundTax:            refundTax.Round(2).InexactFloat64(),
			RefundServiceFee:     refundServiceFee.Round(2).InexactFloat64(),
			RefundDiscount:       refundDiscount.Round(2).InexactFloat64(),
			RefundDiscountMember: refundDiscountMember.Round(2).InexactFloat64(),
			RefundFee:            refundFee.Round(2).InexactFloat64(),
			CompleteTime:         saleBill.FinishTime,
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
	PageNo         int    `json:"page_no"`          // 页码
	PageSize       int    `json:"page_size"`        // 每页大小
	AreaUuid       uint64 `json:"area_uuid"`        // 区域UUID -1=全都
	CategoryUuid   uint64 `json:"category_uuid"`    // 分类UUID -1=全都
}

// buildCountOpts 构建统计选项
func (s *statisticsSrv) buildCountOpts(req CountReq) []repository.DBOption {
	var (
		opts           []repository.DBOption
		commonRepo     = repository.NewCommonRepo()
		queryStartTime int64
		queryEndTime   int64
	)
	// 处理时间范围
	if req.TimeType > 0 && req.TimeType < 5 {
		now := time.Now()
		var startTime, endTime time.Time
		switch req.TimeType {
		case 1: // 今天
			startTime = now.Truncate(24 * time.Hour)
			endTime = startTime.Add(24*time.Hour - time.Second)
		case 2: // 昨天
			startTime = now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
			endTime = startTime.Add(24*time.Hour - time.Second)
		case 3: // 本周
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startTime = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			endTime = startTime.AddDate(0, 0, 7).Add(-time.Second)
		case 4: // 本月
			startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endTime = startTime.AddDate(0, 1, 0).Add(-time.Second)
		}
		queryStartTime = startTime.Unix()
		queryEndTime = endTime.Unix()
	}
	if req.QueryStartTime > 0 && req.QueryEndTime > 0 {
		queryStartTime = req.QueryStartTime
		queryEndTime = req.QueryEndTime
	}
	if queryStartTime > 0 && queryEndTime > 0 {
		if req.IsCreateTime {
			opts = append(opts, commonRepo.WhereBetweenByCreateTime(queryStartTime, queryEndTime))
		} else {
			opts = append(opts, commonRepo.WhereBetweenByCompleteTime(queryStartTime, queryEndTime))
		}
	}
	if req.DutyNo != "" {
		opts = append(opts, commonRepo.WhereByDutyNo(req.DutyNo))
	}
	// logger.Logger.Info("buildCountOptsReq", zap.Any("req", req))
	// if req.AreaUuid > 0 {
	// 	prefix := config.Database.TablePrefix
	// 	opts = append(opts, commonRepo.WhereSubQueryByRegionUuid(prefix+"sale_bill", req.AreaUuid))
	// }
	return opts
}

// buildDays 构建日期
func (s *statisticsSrv) buildDays(req CountReq) []string {
	var (
		days      []string
		format    = "2006-01-02"
		startTime = time.Unix(req.QueryStartTime, 0)
		endTime   = time.Unix(req.QueryEndTime, 0)
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

	// 先删除
	statisticsRepo.DeleteMember(req.MemberRechargeOrderUuid)
	statisticsRepo.DeleteMemberPayment(req.MemberRechargeOrderUuid)

	if req.OnlyDelete {
		return nil
	}

	memberRechargeOrderRepo := repository.NewMemberRechargeOrderRepo(db)
	memberRechargeOrder := memberRechargeOrderRepo.GetRechargeOrderAllInfo(req.MemberRechargeOrderUuid)
	if memberRechargeOrder.Uuid == 0 {
		return nil
	}

	var payments []model.StatisticsMemberPayment
	for _, paymentOrder := range memberRechargeOrder.PaymentOrders {
		var paymentRefundAmount decimal.Decimal
		for _, refundOrderAmount := range paymentOrder.ReturnOrderAmounts {
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

	memberData := statisticsRepo.CountMember(s.buildCountOpts(req)...)

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
	TotalSaleNum          int64   `json:"total_sale_num"`
	TotalOriginSaleAmount float64 `json:"total_origin_sale_amount"`
	TotalActualSaleAmount float64 `json:"total_actual_sale_amount"`
	TotalGiveNum          int64   `json:"total_give_num"`
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
	}, s.buildCountOpts(req)...)

	var data []CountProductSale
	for _, productSale := range productSaleData {
		categoryName := productSale.CategoryName.String
		if productSale.CategoryParentName.String != "" {
			categoryName = productSale.CategoryParentName.String + "-" + categoryName
		}
		data = append(data, CountProductSale{
			ProductName:           productSale.ProductName.String,
			CategoryName:          categoryName,
			TotalSaleNum:          productSale.SaleNum.Int64,
			TotalOriginSaleAmount: productSale.OriginSaleAmount.Float64,
			TotalActualSaleAmount: productSale.ActualSaleAmount.Float64,
			TotalGiveNum:          productSale.GiveNum.Int64,
			TotalBusinessAmount:   productSale.BusinessAmount.Float64,
		})
	}

	return CountProductSaleResp{
		Data:  data,
		Total: total,
	}

}
