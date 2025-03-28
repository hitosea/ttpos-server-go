package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
)

// IBusinessSrv 定义收银服务接口
type IBusinessSrv interface {
	Printer(ctx context.Context, req req.BusinessDataPrinterReq) (*resp.PrinterData, error)                                                               // 打印
	CountBusiness(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataAll, error)                                         // 统计营业数据
	CountPaymentMethod(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataPaymentMethod, error)                          // 统计支付方式
	CountProductCategory(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProductCategory, error)                      // 统计商品分类
	CountProduct(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProduct, error)                                      // 统计商品
	CountArea(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataArea, error)                                            // 统计区域
	CountProductSales(ctx context.Context, req req.BusinessDataCountProductSalesReq) (*business_data_resp.BusinessDataCountProductSalesPagination, error) // 统计商品列表
	Count7Days(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataCount7Days, error)                                     // 统计7天
	RankProduct(ctx context.Context, req req.BusinessDataRankProductReq) (*business_data_resp.BusinessDataProductRank, error)                             // 统计商品排行
}

// businessSrv 收银服务结构体
type businessSrv struct {
	statisticsSrv IStatisticsSrv
}

// NewBusinessSrv 创建新的收银产品类别服务
func NewBusinessSrv(statisticsSrv IStatisticsSrv) IBusinessSrv {
	return NewBusinessSrvImpl(statisticsSrv)
}

// NewBusinessSrvImpl 创建新的收银服务实现
func NewBusinessSrvImpl(statisticsSrv IStatisticsSrv) IBusinessSrv {
	return &businessSrv{
		statisticsSrv: statisticsSrv,
	}
}

// Printer 打印
func (s *businessSrv) Printer(ctx context.Context, printerReq req.BusinessDataPrinterReq) (*resp.PrinterData, error) {
	// Initialize the pointer to avoid nil dereference
	reqPrinterData := &template.PrintingBusinessData{}
	//
	if printerReq.StatisticsType == 0 {
		// 销售数据
		saleData := s.statisticsSrv.CountSale(ctx, CountReq{
			TimeType:       printerReq.TimeType,
			QueryStartTime: int64(printerReq.QueryStartTime),
			QueryEndTime:   int64(printerReq.QueryEndTime),
			CategoryType:   printerReq.CategoryType,
		})
		// 支付数据
		_, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req.BusinessDataCountReq{
			QueryStartTime: int64(printerReq.QueryStartTime),
			QueryEndTime:   int64(printerReq.QueryEndTime),
			TimeType:       printerReq.TimeType,
			CategoryType:   printerReq.CategoryType,
		})
		// 会员数量
		memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
			TimeType:       printerReq.TimeType,
			QueryStartTime: int64(printerReq.QueryStartTime),
			QueryEndTime:   int64(printerReq.QueryEndTime),
			CategoryType:   printerReq.CategoryType,
		})

		reqPrinterData.All = &business_data_resp.BusinessDataAll{
			TotalSales:              saleData.TotalSaleAmount,
			TotalReceivedPrice:      saleData.TotalReceivedAmount,
			TotalPayPrice:           saleData.TotalBusinessAmount,
			TotalProductPrice:       saleData.TotalProductPrice,
			TotalPayFeeMoney:        saleData.TotalPaymentFee,
			TotalServiceMoney:       saleData.TotalServiceFee,
			TotalTaxMoney:           saleData.TotalTax,
			TotalUserDiscountMoney:  saleData.TotalDiscountMember,
			TotalDiscountMoney:      saleData.TotalDiscount,
			TotalDiscountRatio:      saleData.TotalDiscountRatio,
			TotalFreeOrderPrice:     saleData.TotalFreeAmount,
			TotalRefundMoney:        saleData.TotalRefundAmount,
			TotalOrderNum:           int(saleData.TotalOrderNum),
			TotalPeopleNum:          int(saleData.TotalMealNum),
			TotalProductNum:         int(saleData.TotalProductNum),
			TotalTableNum:           int(saleData.TotalDeskNum),
			AvgOrderPrice:           saleData.AvgOrderAmount,
			MinOrderPrice:           saleData.MinOrderAmount,
			MaxOrderPrice:           saleData.MaxOrderAmount,
			AllTableOrderNum:        int(saleData.TotalDeskNum),
			AllTablePeopleNum:       int(saleData.TotalMealNum),
			AllTableAvgOrderPrice:   saleData.AvgDeskOrderAmount,
			AllTableMinOrderPrice:   saleData.MinDeskOrderAmount,
			AllTableMaxOrderPrice:   saleData.MaxDeskOrderAmount,
			AllTablePeopleAvg:       saleData.AvgDeskPeopleOrderAmount,
			AllCashierOrderNum:      int(saleData.TotalInstantOrderNum),
			AllCashierMinOrderPrice: saleData.MinInstantOrderAmount,
			AllCashierMaxOrderPrice: saleData.MaxInstantOrderAmount,
			AllCashierAvgOrderPrice: saleData.AvgInstantOrderAmount,
			PaymentMethodIncomes:    paymentMethodIncomes,
			AbnormalData: func() business_data_resp.AbnormalData {
				AbnormalData, err := repository.NewOrderAbnormalRecordRepo(ctx.GetDB()).GetRecordInfo(
					ctx.GetStaffUuid(),
					ctx.GetStaff().DutyNo,
				)
				if err != nil {
					return business_data_resp.AbnormalData{}
				}
				return *AbnormalData
			}(),
			MemberData: business_data_resp.MemberData{
				RechargeAmount: 120,
				GiftMoney:      120,
				GiftPoints:     120,
				UserCount:      int(memberNum),
			},
			PeakHourList: func() []business_data_resp.PeakHour {
				peakHours, err := repository.NewSaleOrderPeakTimeRepo(ctx.GetDB()).GetMaxRecord(
					uint(printerReq.QueryStartTime),
					uint(printerReq.QueryEndTime),
					ctx.GetStaffUuid(),
				)
				if err != nil {
					return []business_data_resp.PeakHour{}
				}
				return peakHours
			}(),
			CategoryList: func() []business_data_resp.Category {
				_, categoryList := s.BuildCategoryList(ctx, req.BusinessDataCountReq{
					TimeType:       printerReq.TimeType,
					QueryStartTime: int64(printerReq.QueryStartTime),
					QueryEndTime:   int64(printerReq.QueryEndTime),
					CategoryType:   printerReq.CategoryType,
				})
				return categoryList
			}(),
			PercentageList: func() []business_data_resp.Percentage {
				taxData := s.statisticsSrv.CountTax(ctx, CountReq{
					TimeType:       printerReq.TimeType,
					QueryStartTime: int64(printerReq.QueryStartTime),
					QueryEndTime:   int64(printerReq.QueryEndTime),
					CategoryType:   printerReq.CategoryType,
				})
				list := make([]business_data_resp.Percentage, 0, len(taxData))
				for _, tax := range taxData {
					list = append(list, business_data_resp.Percentage{
						TaxRate:        tax.TaxRate,
						ConsumptionTax: tax.TotalTaxFee,
						TotalPrice:     tax.TotalProductAmount,
					})
				}
				return list
			}(),
		}
	}

	if printerReq.StatisticsType == 1 {
		// 支付数据
		paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req.BusinessDataCountReq{
			QueryStartTime: int64(printerReq.QueryStartTime),
			QueryEndTime:   int64(printerReq.QueryEndTime),
			TimeType:       printerReq.TimeType,
			CategoryType:   printerReq.CategoryType,
		})
		reqPrinterData.PaymentMethod = &business_data_resp.BusinessDataPaymentMethod{
			TotalReceivedPrice:   paymentData.TotalReceivedAmount,
			PaymentMethodIncomes: paymentMethodIncomes,
		}
	}

	if printerReq.StatisticsType == 2 {
		categoryData, categoryList := s.BuildCategoryList(ctx, req.BusinessDataCountReq{
			TimeType:       printerReq.TimeType,
			QueryStartTime: int64(printerReq.QueryStartTime),
			QueryEndTime:   int64(printerReq.QueryEndTime),
			CategoryType:   printerReq.CategoryType,
		})

		paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req.BusinessDataCountReq{
			QueryStartTime: int64(printerReq.QueryStartTime),
			QueryEndTime:   int64(printerReq.QueryEndTime),
			TimeType:       printerReq.TimeType,
			CategoryType:   printerReq.CategoryType,
		})

		reqPrinterData.ProductCategory = &business_data_resp.BusinessDataProductCategory{
			SalesNum:             int(categoryData.TotalSaleNum),
			TotalRefundMoney:     paymentData.TotalRefundAmount,
			TotalReceivedPrice:   paymentData.TotalReceivedAmount,
			CategoryList:         categoryList,
			PaymentMethodIncomes: paymentMethodIncomes,
		}
	}

	if printerReq.StatisticsType == 3 {
		reqPrinterData.Product = &business_data_resp.BusinessDataProduct{
			Products: func() []business_data_resp.Product {
				productList := s.statisticsSrv.CountProduct(ctx, CountReq{
					TimeType:       printerReq.TimeType,
					QueryStartTime: int64(printerReq.QueryStartTime),
					QueryEndTime:   int64(printerReq.QueryEndTime),
					CategoryType:   printerReq.CategoryType,
				})
				list := make([]business_data_resp.Product, 0, len(productList))
				for _, product := range productList {
					list = append(list, business_data_resp.Product{
						Name:     product.ProductName,
						SalesNum: int(product.SaleNum),
						Price:    product.SalePrice,
						Subtotal: product.SaleAmount,
					})
				}
				return list
			}(),
		}
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx).PrintingBusinessData(reqPrinterData, int64(printerReq.QueryStartTime), int64(printerReq.QueryEndTime))
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerData, nil
}

// CountBusiness 统计营业数据
func (s *businessSrv) CountBusiness(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataAll, error) {
	// 销售数据
	saleData := s.statisticsSrv.CountSale(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: int64(req.QueryStartTime),
		QueryEndTime:   int64(req.QueryEndTime),
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})
	// 支付数据
	_, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req)
	// 会员数量
	memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: int64(req.QueryStartTime),
		QueryEndTime:   int64(req.QueryEndTime),
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})

	// 营业数据
	var businessData = business_data_resp.BusinessDataAll{
		TotalSales:              saleData.TotalSaleAmount,
		TotalReceivedPrice:      saleData.TotalReceivedAmount,
		TotalPayPrice:           saleData.TotalBusinessAmount,
		TotalProductPrice:       saleData.TotalProductPrice,
		TotalPayFeeMoney:        saleData.TotalPaymentFee,
		TotalServiceMoney:       saleData.TotalServiceFee,
		TotalTaxMoney:           saleData.TotalTax,
		TotalUserDiscountMoney:  saleData.TotalDiscountMember,
		TotalDiscountMoney:      saleData.TotalDiscount,
		TotalDiscountRatio:      saleData.TotalDiscountRatio,
		TotalFreeOrderPrice:     saleData.TotalFreeAmount,
		TotalRefundMoney:        saleData.TotalRefundAmount,
		TotalOrderNum:           int(saleData.TotalOrderNum),
		TotalPeopleNum:          int(saleData.TotalMealNum),
		TotalProductNum:         int(saleData.TotalProductNum),
		TotalTableNum:           int(saleData.TotalDeskNum),
		AvgOrderPrice:           saleData.AvgOrderAmount,
		MinOrderPrice:           saleData.MinOrderAmount,
		MaxOrderPrice:           saleData.MaxOrderAmount,
		AllTableOrderNum:        int(saleData.TotalDeskNum),
		AllTablePeopleNum:       int(saleData.TotalMealNum),
		AllTableAvgOrderPrice:   saleData.AvgDeskOrderAmount,
		AllTableMinOrderPrice:   saleData.MinDeskOrderAmount,
		AllTableMaxOrderPrice:   saleData.MaxDeskOrderAmount,
		AllTablePeopleAvg:       saleData.AvgDeskPeopleOrderAmount,
		AllCashierOrderNum:      int(saleData.TotalInstantOrderNum),
		AllCashierMinOrderPrice: saleData.MinInstantOrderAmount,
		AllCashierMaxOrderPrice: saleData.MaxInstantOrderAmount,
		AllCashierAvgOrderPrice: saleData.AvgInstantOrderAmount,
		PaymentMethodIncomes:    paymentMethodIncomes,
		AbnormalData: func() business_data_resp.AbnormalData {
			AbnormalData, err := repository.NewOrderAbnormalRecordRepo(ctx.GetDB()).GetRecordInfo(
				ctx.GetStaffUuid(),
				ctx.GetStaff().DutyNo,
			)
			if err != nil {
				return business_data_resp.AbnormalData{}
			}
			return *AbnormalData
		}(),
		MemberData: business_data_resp.MemberData{
			RechargeAmount: 120,
			GiftMoney:      120,
			GiftPoints:     120,
			UserCount:      int(memberNum),
		},
		PeakHourList: func() []business_data_resp.PeakHour {
			peakHours, err := repository.NewSaleOrderPeakTimeRepo(ctx.GetDB()).GetMaxRecord(
				uint(req.QueryStartTime),
				uint(req.QueryEndTime),
				ctx.GetStaffUuid(),
			)
			if err != nil {
				// 如果出错，返回空切片
				return []business_data_resp.PeakHour{}
			}
			return peakHours
		}(),
		CategoryList: func() []business_data_resp.Category {
			_, categoryList := s.BuildCategoryList(ctx, req)
			return categoryList
		}(),
		PercentageList: func() []business_data_resp.Percentage {
			taxData := s.statisticsSrv.CountTax(ctx, CountReq{
				TimeType:       req.TimeType,
				QueryStartTime: int64(req.QueryStartTime),
				QueryEndTime:   int64(req.QueryEndTime),
				CategoryType:   req.CategoryType,
			})
			list := make([]business_data_resp.Percentage, 0, len(taxData))
			for _, tax := range taxData {
				list = append(list, business_data_resp.Percentage{
					TaxRate:        tax.TaxRate,
					ConsumptionTax: tax.TotalTaxFee,
					TotalPrice:     tax.TotalProductAmount,
				})
			}
			return list
		}(),
	}

	return &businessData, nil
}

// CountPaymentMethod 统计支付方式
func (s *businessSrv) CountPaymentMethod(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataPaymentMethod, error) {
	paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req)

	var paymentMethodData = business_data_resp.BusinessDataPaymentMethod{
		TotalReceivedPrice:   paymentData.TotalReceivedAmount,
		PaymentMethodIncomes: paymentMethodIncomes,
	}

	return &paymentMethodData, nil
}

// CountProductCategory 统计商品分类
func (s *businessSrv) CountProductCategory(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProductCategory, error) {
	paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req)
	categoryData, categoryList := s.BuildCategoryList(ctx, req)
	var productCategoryData = business_data_resp.BusinessDataProductCategory{
		SalesNum:             int(categoryData.TotalSaleNum),
		TotalRefundMoney:     paymentData.TotalRefundAmount,
		TotalReceivedPrice:   paymentData.TotalReceivedAmount,
		CategoryList:         categoryList,
		PaymentMethodIncomes: paymentMethodIncomes,
	}

	return &productCategoryData, nil
}

// CountProduct 统计商品
func (s *businessSrv) CountProduct(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProduct, error) {
	var productData = business_data_resp.BusinessDataProduct{
		Products: func() []business_data_resp.Product {
			productList := s.statisticsSrv.CountProduct(ctx, CountReq{
				TimeType:       req.TimeType,
				QueryStartTime: int64(req.QueryStartTime),
				QueryEndTime:   int64(req.QueryEndTime),
				CategoryType:   req.CategoryType,
				DutyNo:         req.DutyNo,
			})
			list := make([]business_data_resp.Product, 0, len(productList))
			for _, product := range productList {
				list = append(list, business_data_resp.Product{
					Name:     product.ProductName,
					SalesNum: int(product.SaleNum),
					Price:    product.SalePrice,
					Subtotal: product.SaleAmount,
				})
			}
			return list
		}(),
	}

	return &productData, nil
}

// CountArea 统计区域
func (s *businessSrv) CountArea(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataArea, error) {
	areaData := s.statisticsSrv.CountArea(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: int64(req.QueryStartTime),
		QueryEndTime:   int64(req.QueryEndTime),
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})

	var areaList = []business_data_resp.Area{}
	for _, area := range areaData {
		areaList = append(areaList, business_data_resp.Area{
			Name:               area.AreaName,
			TotalSales:         area.AreaSaleAmount,
			TotalReceivedPrice: area.AreaBusinessAmount,
			TotalProductNum:    int(area.AreaProductNum),
		})
	}

	return &business_data_resp.BusinessDataArea{
		Areas: areaList,
	}, nil
}

// RankProduct 统计商品排行
func (s *businessSrv) RankProduct(ctx context.Context, req req.BusinessDataRankProductReq) (*business_data_resp.BusinessDataProductRank, error) {
	var productRankData = business_data_resp.BusinessDataProductRank{
		Ranks: func() []business_data_resp.ProductRank {
			productRankList := s.statisticsSrv.RankProduct(ctx, CountReq{
				RankType:       req.RankType,
				QueryStartTime: int64(req.QueryStartTime),
				QueryEndTime:   int64(req.QueryEndTime),
			})
			list := make([]business_data_resp.ProductRank, 0, len(productRankList))
			for _, productRank := range productRankList {
				list = append(list, business_data_resp.ProductRank{
					ProductName: productRank.ProductName,
					SalesNum:    int(productRank.SaleNum),
					SalesPrice:  productRank.SaleAmount,
				})
			}
			return list
		}(),
	}

	return &productRankData, nil
}

// CountProductSales 统计商品列表
func (s *businessSrv) CountProductSales(ctx context.Context, req req.BusinessDataCountProductSalesReq) (*business_data_resp.BusinessDataCountProductSalesPagination, error) {
	var list = []business_data_resp.BusinessDataCountProductSalesItem{
		{
			ProductName:        "12",
			SalesNum:           1,
			SalesPrice:         323,
			CategoryName:       "12",
			OriginalSalesPrice: 120,
			TotalPayPrice:      120,
			GiveProductNum:     120,
		},
		{
			ProductName:        "121232",
			SalesNum:           2,
			SalesPrice:         23,
			CategoryName:       "121232",
			OriginalSalesPrice: 120,
			TotalPayPrice:      120,
			GiveProductNum:     120,
		},
		{
			ProductName:        "121232",
			SalesNum:           2,
			SalesPrice:         23,
			CategoryName:       "121232",
			OriginalSalesPrice: 120,
			TotalPayPrice:      120,
			GiveProductNum:     120,
		},
	}
	var productListData = business_data_resp.BusinessDataCountProductSalesPagination{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   1,
			PageSize: 10,
			Total:    int64(len(list)),
		},
	}

	return &productListData, nil
}

// Count7Days 统计7天
func (s *businessSrv) Count7Days(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataCount7Days, error) {
	sevenDaysData := s.statisticsSrv.Count7Days(ctx, CountReq{
		QueryStartTime: int64(req.QueryStartTime),
		QueryEndTime:   int64(req.QueryEndTime),
	})

	sevenDaysDataList := make([]business_data_resp.BusinessDataCount7DaysItem, 0, len(sevenDaysData.Data))
	for _, day := range sevenDaysData.Data {
		sevenDaysDataList = append(sevenDaysDataList, business_data_resp.BusinessDataCount7DaysItem{
			Day:        day.Day,
			TotalNum:   day.TotalNum,
			TotalMoney: day.TotalMoney,
		})
	}

	return &business_data_resp.BusinessDataCount7Days{
		Days: sevenDaysData.Days,
		Data: sevenDaysDataList,
	}, nil
}

// BuildPaymentMethodIncome 构建支付方式收入
func (s *businessSrv) BuildPaymentMethodIncome(ctx context.Context, req req.BusinessDataCountReq) (CountPaymentResp, []business_data_resp.PaymentMethodIncome) {
	paymentData := s.statisticsSrv.CountPayment(ctx, CountReq{
		DutyNo:         req.DutyNo,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		TimeType:       req.TimeType,
		CategoryType:   req.CategoryType,
	})

	paymentMethodIncomes := make([]business_data_resp.PaymentMethodIncome, 0)
	for _, payment := range paymentData.PaymentList {
		paymentMethodIncomes = append(paymentMethodIncomes, business_data_resp.PaymentMethodIncome{
			Name:     payment.PaymentName,
			OrderNum: int(payment.TotalOrderNum),
			Amount:   payment.TotalPaymentAmount,
			Code:     payment.PaymentCode,
		})
	}

	return paymentData, paymentMethodIncomes
}

// BuildCategoryList 构建分类列表
func (s *businessSrv) BuildCategoryList(ctx context.Context, req req.BusinessDataCountReq) (CountCategoryResp, []business_data_resp.Category) {
	categoryData := s.statisticsSrv.CountCategory(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})
	list := make([]business_data_resp.Category, 0, len(categoryData.CategoryList))
	for _, category := range categoryData.CategoryList {
		list = append(list, business_data_resp.Category{
			Name:     category.CategoryName,
			SalesNum: int(category.SaleNum),
			Prices:   category.SaleAmount,
		})
	}

	return categoryData, list
}
