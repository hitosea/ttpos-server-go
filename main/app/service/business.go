package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/pkg/context"
)

// IBusinessSrv 定义收银服务接口
type IBusinessSrv interface {
	Printer(ctx context.Context, req req.BusinessDataPrinterReq) (*resp.PrinterData, error)                                          // 打印
	CountBusiness(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataAll, error)                    // 统计营业数据
	CountPaymentMethod(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataPaymentMethod, error)     // 统计支付方式
	CountProductCategory(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProductCategory, error) // 统计商品分类
	CountProduct(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProduct, error)                 // 统计商品
	CountArea(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataArea, error)                       // 统计区域
	RankProduct(ctx context.Context, req req.BusinessDataRankProductReq) (*business_data_resp.BusinessDataProductRank, error)        // 统计商品排行
}

// businessSrv 收银服务结构体
type businessSrv struct {
}

// NewBusinessSrv 创建新的收银产品类别服务
func NewBusinessSrv() IBusinessSrv {
	return NewBusinessSrvImpl()
}

// NewBusinessSrvImpl 创建新的收银服务实现
func NewBusinessSrvImpl() IBusinessSrv {
	return &businessSrv{}
}

// Printer 打印
func (s *businessSrv) Printer(ctx context.Context, req req.BusinessDataPrinterReq) (*resp.PrinterData, error) {
	// Initialize the pointer to avoid nil dereference
	reqPrinterData := &template.PrintingBusinessData{}
	//
	if req.StatisticsType == 0 {
		reqPrinterData.All = &business_data_resp.BusinessDataAll{
			TotalSales:              2131231,
			TotalReceivedPrice:      43124,
			TotalPayPrice:           21230,
			TotalPayFeeMoney:        2110,
			TotalServiceMoney:       120,
			TotalTaxMoney:           10124,
			TotalUserDiscountMoney:  120,
			TotalDiscountMoney:      120,
			TotalFreeOrderPrice:     120,
			TotalRefundMoney:        10,
			TotalOrderNum:           1230,
			TotalPeopleNum:          120,
			TotalProductNum:         320,
			TotalTableNum:           120,
			AvgOrderPrice:           620,
			MinOrderPrice:           120,
			MaxOrderPrice:           1200,
			AllTableOrderNum:        1230,
			AllTablePeopleNum:       120,
			AllTableAvgOrderPrice:   620,
			AllTableMinOrderPrice:   120,
			AllTableMaxOrderPrice:   1200,
			AllTablePeopleAvg:       10,
			AllCashierOrderNum:      1230,
			AllCashierMinOrderPrice: 120,
			AllCashierMaxOrderPrice: 1200,
			AllCashierAvgOrderPrice: 620,
			PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
				{
					Name:     "现金",
					OrderNum: 1,
					Amount:   123213,
					Code:     40,
				},
				{
					Name:     "支付宝",
					OrderNum: 1,
					Amount:   24121,
					Code:     41,
				},
				{
					Name:     "微信支付",
					OrderNum: 1,
					Amount:   123213,
					Code:     42,
				},
			},
			AbnormalData: business_data_resp.AbnormalData{},
			MemberData: business_data_resp.MemberData{
				RechargeAmount: 120,
				GiftMoney:      120,
				GiftPoints:     120,
			},
			PeakHourList: []business_data_resp.PeakHour{
				{
					TimePeriod: "12",
					OrderNum:   120,
					Amount:     120,
				},
				{
					TimePeriod: "121232",
					OrderNum:   120,
					Amount:     120,
				},
			},
			CategoryList: []business_data_resp.Category{
				{
					Name:     "12",
					SalesNum: 1,
					Prices:   323,
				},
				{
					Name:     "121232",
					SalesNum: 2,
					Prices:   23,
				},
			},
			PercentageList: []business_data_resp.Percentage{
				{
					TaxRate:        120,
					ConsumptionTax: 120,
				},
				{
					TaxRate:        110,
					ConsumptionTax: 2120,
				},
			},
		}
	}

	if req.StatisticsType == 1 {
		reqPrinterData.PaymentMethod = &business_data_resp.BusinessDataPaymentMethod{
			TotalReceivedPrice: 1121,
			PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
				{
					Name:     "现金",
					OrderNum: 1,
					Amount:   123213,
					Code:     40,
				},
				{
					Name:     "支付宝",
					OrderNum: 1,
					Amount:   24121,
					Code:     41,
				},
				{
					Name:     "微信支付",
					OrderNum: 1,
					Amount:   123213,
					Code:     42,
				},
			},
		}
	}

	if req.StatisticsType == 2 {
		reqPrinterData.ProductCategory = &business_data_resp.BusinessDataProductCategory{
			SalesNum: 120,
			CategoryList: []business_data_resp.Category{
				{
					Name:     "12",
					SalesNum: 1,
					Prices:   323,
				},
				{
					Name:     "121232",
					SalesNum: 2,
					Prices:   23,
				},
			},
			PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
				{
					Name:     "现金",
					OrderNum: 1,
					Amount:   123213,
					Code:     40,
				},
				{
					Name:     "支付宝",
					OrderNum: 1,
					Amount:   24121,
					Code:     41,
				},
				{
					Name:     "微信支付",
					OrderNum: 1,
					Amount:   123213,
					Code:     42,
				},
			},
		}
	}

	if req.StatisticsType == 3 {
		reqPrinterData.Product = &business_data_resp.BusinessDataProduct{
			Products: []business_data_resp.Product{
				{
					Name:     "12",
					SalesNum: 1,
					Price:    323,
					Subtotal: 323,
				},
				{
					Name:     "121232",
					SalesNum: 2,
					Price:    23,
					Subtotal: 46,
				},
			},
		}
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx).PrintingBusinessData(reqPrinterData, int64(req.QueryStartTime), int64(req.QueryEndTime))
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerData, nil
}

// CountBusiness 统计营业数据
func (s *businessSrv) CountBusiness(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataAll, error) {
	// 营业数据
	var businessData = business_data_resp.BusinessDataAll{
		TotalSales:              2131231,
		TotalReceivedPrice:      43124,
		TotalPayPrice:           21230,
		TotalPayFeeMoney:        2110,
		TotalServiceMoney:       120,
		TotalTaxMoney:           10124,
		TotalUserDiscountMoney:  120,
		TotalDiscountMoney:      120,
		TotalFreeOrderPrice:     120,
		TotalRefundMoney:        10,
		TotalOrderNum:           1230,
		TotalPeopleNum:          120,
		TotalProductNum:         320,
		TotalTableNum:           120,
		AvgOrderPrice:           620,
		MinOrderPrice:           120,
		MaxOrderPrice:           1200,
		AllTableOrderNum:        1230,
		AllTablePeopleNum:       120,
		AllTableAvgOrderPrice:   620,
		AllTableMinOrderPrice:   120,
		AllTableMaxOrderPrice:   1200,
		AllTablePeopleAvg:       10,
		AllCashierOrderNum:      1230,
		AllCashierMinOrderPrice: 120,
		AllCashierMaxOrderPrice: 1200,
		AllCashierAvgOrderPrice: 620,
		PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
			{
				Name:     "现金",
				OrderNum: 1,
				Amount:   123213,
				Code:     40,
			},
			{
				Name:     "支付宝",
				OrderNum: 1,
				Amount:   24121,
				Code:     41,
			},
			{
				Name:     "微信支付",
				OrderNum: 1,
				Amount:   123213,
				Code:     42,
			},
		},
		AbnormalData: business_data_resp.AbnormalData{},
		MemberData: business_data_resp.MemberData{
			RechargeAmount: 120,
			GiftMoney:      120,
			GiftPoints:     120,
		},
		PeakHourList: []business_data_resp.PeakHour{
			{
				TimePeriod: "12",
				OrderNum:   120,
				Amount:     120,
			},
			{
				TimePeriod: "121232",
				OrderNum:   120,
				Amount:     120,
			},
		},
		CategoryList: []business_data_resp.Category{
			{
				Name:     "12",
				SalesNum: 1,
				Prices:   323,
			},
			{
				Name:     "121232",
				SalesNum: 2,
				Prices:   23,
			},
		},
		PercentageList: []business_data_resp.Percentage{
			{
				TaxRate:        120,
				ConsumptionTax: 120,
			},
			{
				TaxRate:        110,
				ConsumptionTax: 2120,
			},
		},
	}

	return &businessData, nil
}

// CountPaymentMethod 统计支付方式
func (s *businessSrv) CountPaymentMethod(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataPaymentMethod, error) {
	var paymentMethodData = business_data_resp.BusinessDataPaymentMethod{
		TotalReceivedPrice: 120,
		PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
			{
				Name:     "现金",
				OrderNum: 1,
				Amount:   123213,
				Code:     40,
			},
			{
				Name:     "支付宝",
				OrderNum: 1,
				Amount:   24121,
				Code:     41,
			},
			{
				Name:     "微信支付",
				OrderNum: 1,
				Amount:   123213,
				Code:     42,
			},
		},
	}

	return &paymentMethodData, nil
}

// CountProductCategory 统计商品分类
func (s *businessSrv) CountProductCategory(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProductCategory, error) {
	var productCategoryData = business_data_resp.BusinessDataProductCategory{
		SalesNum:           120,
		TotalRefundMoney:   120,
		TotalReceivedPrice: 120,
		CategoryList: []business_data_resp.Category{
			{
				Name:     "12",
				SalesNum: 1,
				Prices:   323,
			},
			{
				Name:     "121232",
				SalesNum: 2,
				Prices:   23,
			},
		},
		PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
			{
				Name:     "现金",
				OrderNum: 1,
				Amount:   123213,
				Code:     40,
			},
			{
				Name:     "支付宝",
				OrderNum: 1,
				Amount:   24121,
				Code:     41,
			},
			{
				Name:     "微信支付",
				OrderNum: 1,
				Amount:   123213,
				Code:     42,
			},
		},
	}

	return &productCategoryData, nil
}

// CountProduct 统计商品
func (s *businessSrv) CountProduct(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProduct, error) {
	var productData = business_data_resp.BusinessDataProduct{
		Products: []business_data_resp.Product{
			{
				Name:     "12",
				SalesNum: 1,
				Price:    323,
				Subtotal: 323,
			},
			{
				Name:     "121232",
				SalesNum: 2,
				Price:    23,
				Subtotal: 46,
			},
		},
	}

	return &productData, nil
}

// CountArea 统计区域
func (s *businessSrv) CountArea(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataArea, error) {
	var areaData = business_data_resp.BusinessDataArea{
		Areas: []business_data_resp.Area{
			{
				Name:               "12",
				TotalSales:         120,
				TotalReceivedPrice: 120,
				TotalProductNum:    120,
			},
			{
				Name:               "121232",
				TotalSales:         120,
				TotalReceivedPrice: 120,
				TotalProductNum:    120,
			},
		},
	}

	return &areaData, nil
}

// RankProduct 统计商品排行
func (s *businessSrv) RankProduct(ctx context.Context, req req.BusinessDataRankProductReq) (*business_data_resp.BusinessDataProductRank, error) {
	var productRankData = business_data_resp.BusinessDataProductRank{
		Ranks: []business_data_resp.ProductRank{
			{
				ProductName: "12",
				SalesNum:    1,
				SalesPrice:  323,
			},
			{
				ProductName: "121232",
				SalesNum:    2,
				SalesPrice:  23,
			},
		},
	}

	return &productRankData, nil
}
