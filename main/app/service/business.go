package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IBusinessSrv 定义收银服务接口
type IBusinessSrv interface {
	Printer(ctx context.Context, req req.BusinessDataPrinterReq) (*resp.PrinterData, error)                                                                               // 打印
	CountBusiness(ctx context.Context, req req.BusinessDataCountReq, opts ...func(o *CountBusinessOption)) (*business_data_resp.BusinessDataAll, error)                   // 统计营业数据
	CountPaymentMethod(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataPaymentMethod, error)                                          // 统计支付方式
	CountProductCategory(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProductCategory, error)                                      // 统计商品分类
	CountProduct(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProduct, error)                                                      // 统计商品
	CountArea(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataArea, error)                                                            // 统计区域
	CountProductSales(ctx context.Context, req req.BusinessDataCountProductSalesReq) (*business_data_resp.BusinessDataCountProductSalesPagination, error)                 // 统计商品列表
	Count7Days(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataCount7Days, error)                                                     // 统计7天
	CountExport(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataExport, error)                                                        // 统计导出
	RankProduct(ctx context.Context, req req.BusinessDataRankProductReq) (*business_data_resp.BusinessDataProductRank, error)                                             // 统计商品排行
	CountShiftRefundAmount(ctx context.Context, req req.BusinessDataCountReq) *business_data_resp.BusinessDataShiftRefundAmount                                           // 统计班次退款金额
	CountHome(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataHome, error)                                                            // 统计首页
	CountKitchenEfficiencyAnalysis(ctx context.Context, req req.KitchenEfficiencyAnalysisReq) (*business_data_resp.BusinessDataKitchenEfficiencyAnalysis, error)          // 统计后厨效率分析
	CountKitchenEfficiencyAnalysisAvg(ctx context.Context, req req.KitchenEfficiencyAnalysisAvgReq) (*business_data_resp.BusinessDataKitchenEfficiencyAnalysisAvg, error) // 统计后厨效率分析平均时长
	ExportKitchenEfficiencyAnalysis(ctx context.Context, req req.KitchenEfficiencyAnalysisReq) error                                                                      // 导出后厨效率分析
	KitchenProductionDetail(ctx context.Context, req req.KitchenProductionDetailReq) (*business_data_resp.KitchenProductionDetail, error)                                 // 统计后厨菜品出品明细
	KitchenProductionDetailCount(ctx context.Context, req req.KitchenProductionDetailReq) (int64, error)                                                                  // 统计后厨菜品出品明细数量
	ExportKitchenProductionDetail(ctx context.Context, req req.KitchenProductionDetailReq) error                                                                          // 导出后厨菜品出品明细
	StatsKitchenEfficiencyAnalysis(ctx context.Context) (string, error)                                                                                                   // 统计后厨效率分析
	CountBusinessTimePeriod(ctx context.Context, req req.BusinessTimePeriodReq) business_data_resp.BusinessTimePeriod                                                     // 统计营业时段数据
	CountBusinessComprehensiveOperations(ctx context.Context, req req.StatisticsComprehensiveOperationsReq) business_data_resp.StatisticsComprehensiveOperations          // 统计综合运用数据
	CountBusinessPaymentMethod(ctx context.Context, req req.StatisticsPaymentMethodReq) business_data_resp.StatisticsPaymentMethod                                        // 统计支付方式数据
}

// businessSrv 收银服务结构体
type businessSrv struct {
	statisticsSrv IStatisticsSrv
	uploadFileSrv IUploadFileSrv
}

// NewBusinessSrv 创建新的收银产品类别服务
func NewBusinessSrv(statisticsSrv IStatisticsSrv, uploadFileSrv IUploadFileSrv) IBusinessSrv {
	return NewBusinessSrvImpl(statisticsSrv, uploadFileSrv)
}

// NewBusinessSrvImpl 创建新的收银服务实现
func NewBusinessSrvImpl(statisticsSrv IStatisticsSrv, uploadFileSrv IUploadFileSrv) IBusinessSrv {
	return &businessSrv{
		statisticsSrv: statisticsSrv,
		uploadFileSrv: uploadFileSrv,
	}
}

// Printer 打印
func (s *businessSrv) Printer(ctx context.Context, printerReq req.BusinessDataPrinterReq) (*resp.PrinterData, error) {
	setting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	storeSetting, err := setting.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店设置失败", zap.Error(err))
		fmt.Println("获取门店设置失败", zap.Error(err))
	}

	// 获取打印机设置
	printerSetting, err := setting.GetPrinterSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取打印机设置失败", zap.Error(err))
		fmt.Println("获取打印机设置失败", zap.Error(err))
	}

	// 获取门店业务设置
	businessSetting, err := setting.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店业务设置失败", zap.Error(err))
		fmt.Println("获取门店业务设置失败", zap.Error(err))
	}

	// 设置语言
	ctx.SetLanguage(printerSetting.DefaultLanguage)

	var printerData *resp.PrinterData

	// Initialize the pointer to avoid nil dereference
	reqPrinterData := &template.PrintingBusinessData{}
	// 获取参数
	printerParam := printerReq.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)
	// 统计类型
	if printerReq.StatisticsType <= 0 {
		// 销售数据
		saleData := s.statisticsSrv.CountSale(ctx, CountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
			CategoryType:   printerParam.CategoryType,
		})
		// 支付数据
		_, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req.BusinessDataCountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
			CategoryType:   printerParam.CategoryType,
			NotQueryFree:   true,
		})
		// 会员数量
		memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
		})
		// 未结订单
		unpaidOrderData := s.statisticsSrv.CountUnpaidOrder(ctx, CountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
		})

		reqPrinterData.All = &business_data_resp.BusinessDataAll{
			TotalSales:              saleData.TotalSaleAmount,
			TotalReceivedPrice:      saleData.TotalReceivedAmount,
			TotalPayPrice:           saleData.TotalBusinessAmount,
			TotalProductPrice:       saleData.TotalProductOriginPrice,
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
			TotalProductNum:         saleData.TotalProductNum,
			TotalTableNum:           int(saleData.TotalDeskNum),
			TotalCancelOrderNum:     int(saleData.TotalCancelOrderNum),
			TotalCancelOrderAmount:  saleData.TotalCancelOrderAmount,
			TotalGiveProductPrice:   saleData.TotalGiftAmount,
			TotalGiveProductNum:     saleData.TotalGiftNum,
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
			UnclosedTotalOrderNum:   int(unpaidOrderData.TotalOrderNum),
			UnclosedTotalPrice:      unpaidOrderData.TotalAmount,
			PaymentMethodIncomes:    paymentMethodIncomes,
			AbnormalData: func() business_data_resp.AbnormalData {
				AbnormalData, err := repository.NewOrderAbnormalRecordRepo(ctx.GetDB()).GetRecordInfo(
					0,
					ctx.GetStaff().DutyNo,
				)
				if err != nil {
					return business_data_resp.AbnormalData{}
				}
				return *AbnormalData
			}(),
			MemberData: func() business_data_resp.MemberData {
				memberData := s.statisticsSrv.CountMember(ctx, CountReq{
					QueryStartTime: int64(printerParam.QueryStartTime),
					QueryEndTime:   int64(printerParam.QueryEndTime),
					CategoryType:   printerParam.CategoryType,
				})
				return business_data_resp.MemberData{
					RechargeAmount: memberData.TotalRechargeAmount,
					GiftMoney:      memberData.TotalGiveAmount,
					GiftPoints:     memberData.TotalGivePoint,
					UserCount:      int(memberNum),
				}
			}(),
			PeakHourList: func() []business_data_resp.PeakHour {
				peakHours, err := repository.NewSaleOrderPeakTimeRepo(ctx.GetDB()).GetMaxRecord(
					storeSetting.TimeZone,
					printerParam.QueryStartTime,
					printerParam.QueryEndTime,
					ctx.GetStaffUuid(),
				)
				if err != nil {
					return []business_data_resp.PeakHour{}
				}
				return peakHours
			}(),
			CategoryList: func() []business_data_resp.Category {
				_, categoryList := s.BuildCategoryList(ctx, req.BusinessDataCountReq{
					QueryStartTime: int64(printerParam.QueryStartTime),
					QueryEndTime:   int64(printerParam.QueryEndTime),
					CategoryType:   printerParam.CategoryType,
				})
				return categoryList
			}(),
			PercentageList: func() []business_data_resp.Percentage {
				taxData := s.statisticsSrv.CountTax(ctx, CountReq{
					QueryStartTime: int64(printerParam.QueryStartTime),
					QueryEndTime:   int64(printerParam.QueryEndTime),
					CategoryType:   printerParam.CategoryType,
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
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
			CategoryType:   printerParam.CategoryType,
		})
		reqPrinterData.PaymentMethod = &business_data_resp.BusinessDataPaymentMethod{
			TotalReceivedPrice:   paymentData.TotalReceivedAmount,
			PaymentMethodIncomes: paymentMethodIncomes,
		}
	}

	if printerReq.StatisticsType == 2 {
		categoryData, categoryList := s.BuildCategoryList(ctx, req.BusinessDataCountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
			CategoryType:   printerParam.CategoryType,
			SortType:       printerParam.SortType,
			SortDirection:  printerParam.SortDirection,
		})

		paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req.BusinessDataCountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
			CategoryType:   printerParam.CategoryType,
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
		productList := s.statisticsSrv.CountProduct(ctx, CountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
			CategoryType:   printerParam.CategoryType,
		})
		// 将产品列表分成每100条一组，每组记录批次号
		const batchSize = 100
		type ProductBatch struct {
			BatchRange string                       `json:"batch_range"` // 批次范围 (如: "1-100", "101-200")
			Products   []business_data_resp.Product `json:"products"`    // 该批次的商品列表
		}
		var batches []ProductBatch
		for i := 0; i < len(productList); i += batchSize {
			end := i + batchSize
			if end > len(productList) {
				end = len(productList)
			}
			// 生成批次范围（从1开始计数）
			batchStart := i + 1
			batchEnd := end
			batchRange := fmt.Sprintf("%d-%d", batchStart, batchEnd)
			// 处理当前批次的产品数据
			batchProducts := make([]business_data_resp.Product, 0, end-i)
			for _, product := range productList[i:end] {
				batchProducts = append(batchProducts, business_data_resp.Product{
					Name:     product.ProductName,
					SalesNum: product.SaleNum,
					Price:    product.SalePrice,
					Subtotal: product.SaleAmount,
				})
			}
			// 添加到批次列表
			batches = append(batches, ProductBatch{
				BatchRange: batchRange,
				Products:   batchProducts,
			})
		}
		for i, v := range batches {
			reqPrinterData.Product = &business_data_resp.BusinessDataProduct{
				Products:   v.Products,
				BatchRange: utils.IfString(len(batches) > 1, v.BatchRange, ""),
			}
			// 打印
			printerContent, err := printer.NewPrinterRepo(ctx).PrintingBusinessData(
				reqPrinterData,
				int64(printerParam.QueryStartTime),
				int64(printerParam.QueryEndTime),
				utils.IfInt(i == 0, 1, 0),
			)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			if i == 0 {
				printerData = printerContent
			}
		}
	} else {
		// 打印
		printerData, err = printer.NewPrinterRepo(ctx).PrintingBusinessData(
			reqPrinterData,
			int64(printerParam.QueryStartTime),
			int64(printerParam.QueryEndTime),
			1,
		)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	return printerData, nil
}

type CountBusinessOption struct {
	IsShop bool // 是否是商家后台的首页统计的使用场景
}

func WithIsShop() func(o *CountBusinessOption) {
	return func(o *CountBusinessOption) {
		o.IsShop = true
	}
}

// CountBusiness 统计营业数据
func (s *businessSrv) CountBusiness(ctx context.Context, req req.BusinessDataCountReq, opts ...func(o *CountBusinessOption)) (*business_data_resp.BusinessDataAll, error) {
	option := &CountBusinessOption{}
	for _, opt := range opts {
		opt(option)
	}
	setting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	businessSetting, err := setting.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
		fmt.Println("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)
	// 销售数据
	saleData := s.statisticsSrv.CountSale(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})

	storeSetting, err := setting.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店设置失败", zap.Error(err))
		fmt.Println("获取门店设置失败", zap.Error(err))
	}

	// 支付数据
	_, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req)

	// 会员数量
	memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
	})
	// 未结订单
	unpaidOrderData := s.statisticsSrv.CountUnpaidOrder(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})
	// 营业时间
	openingHours := ""
	if req.TimeType == 5 {
		openingHours = businessSetting.OpeningHours
	}
	// 营业数据
	var businessData = business_data_resp.BusinessDataAll{
		TotalSales:                 saleData.TotalSaleAmount,
		TotalReceivedPrice:         saleData.TotalReceivedAmount,
		TotalPayPrice:              saleData.TotalBusinessAmount,
		TotalProductPrice:          saleData.TotalProductOriginPrice,
		TotalPayFeeMoney:           saleData.TotalPaymentFee,
		TotalServiceMoney:          saleData.TotalServiceFee,
		TotalTaxMoney:              saleData.TotalTax,
		TotalUserDiscountMoney:     saleData.TotalDiscountMember,
		TotalDiscountMoney:         saleData.TotalDiscount,
		TotalDiscountRatio:         saleData.TotalDiscountRatio,
		TotalFreeOrderPrice:        saleData.TotalFreeAmount,
		TotalFreeOrderNum:          int(saleData.TotalFreeNum),
		TotalGiveProductPrice:      saleData.TotalGiftAmount,
		TotalGiveProductNum:        saleData.TotalGiftNum,
		TotalRefundMoney:           saleData.TotalRefundAmount,
		TotalOrderNum:              int(saleData.TotalOrderNum),
		TotalPeopleNum:             int(saleData.TotalMealNum),
		TotalProductNum:            saleData.TotalProductNum,
		TotalTableNum:              int(saleData.TotalDeskNum),
		TotalCancelOrderNum:        int(saleData.TotalCancelOrderNum),
		TotalCancelOrderAmount:     saleData.TotalCancelOrderAmount,
		TotalTakeoutSaleAmount:     saleData.TotalTakeoutSaleAmount,
		TotalTakeoutBusinessAmount: saleData.TotalTakeoutBusinessAmount,
		TotalTakeoutRefundAmount:   saleData.TotalTakeoutRefundAmount,
		TotalTakeoutDeliveryFee:    saleData.TotalTakeoutDeliveryFee,
		AvgOrderPrice:              saleData.AvgOrderAmount,
		MinOrderPrice:              saleData.MinOrderAmount,
		MaxOrderPrice:              saleData.MaxOrderAmount,
		AllTableOrderNum:           int(saleData.TotalDeskNum),
		AllTablePeopleNum:          int(saleData.TotalMealNum),
		AllTableAvgOrderPrice:      saleData.AvgDeskOrderAmount,
		AllTableMinOrderPrice:      saleData.MinDeskOrderAmount,
		AllTableMaxOrderPrice:      saleData.MaxDeskOrderAmount,
		AllTablePeopleAvg:          saleData.AvgDeskPeopleOrderAmount,
		AllCashierOrderNum:         int(saleData.TotalInstantOrderNum),
		AllCashierMinOrderPrice:    saleData.MinInstantOrderAmount,
		AllCashierMaxOrderPrice:    saleData.MaxInstantOrderAmount,
		AllCashierAvgOrderPrice:    saleData.AvgInstantOrderAmount,
		AllTakeoutOrderNum:         int(saleData.TotalTakeoutOrderNum),
		AllTakeoutMinOrderPrice:    saleData.MinTakeoutOrderAmount,
		AllTakeoutMaxOrderPrice:    saleData.MaxTakeoutOrderAmount,
		AllTakeoutAvgOrderPrice:    saleData.AvgTakeoutOrderAmount,
		UnclosedTotalOrderNum:      int(unpaidOrderData.TotalOrderNum),
		UnclosedTotalPrice:         unpaidOrderData.TotalAmount,
		PaymentMethodIncomes:       paymentMethodIncomes,
		AbnormalData: func() business_data_resp.AbnormalData {
			AbnormalData, err := repository.NewOrderAbnormalRecordRepo(ctx.GetDB()).GetRecordInfo(
				0,
				utils.IfString(req.DutyNo != "", req.DutyNo, ctx.GetStaff().DutyNo),
			)
			if err != nil {
				return business_data_resp.AbnormalData{}
			}
			return *AbnormalData
		}(),
		MemberData: func() business_data_resp.MemberData {
			memberData := s.statisticsSrv.CountMember(ctx, CountReq{
				TimeType:       req.TimeType,
				QueryStartTime: req.QueryStartTime,
				QueryEndTime:   req.QueryEndTime,
				CategoryType:   req.CategoryType,
			})
			return business_data_resp.MemberData{
				RechargeAmount: memberData.TotalRechargeAmount,
				GiftMoney:      memberData.TotalGiveAmount,
				GiftPoints:     memberData.TotalGivePoint,
				UserCount:      int(memberNum),
			}
		}(),
		PeakHourList: func() []business_data_resp.PeakHour {
			peakHours, err := repository.NewSaleOrderPeakTimeRepo(ctx.GetDB()).GetMaxRecord(
				storeSetting.TimeZone,
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
			// 商家后台首页统计不展示分类列表。不查询该数据可以提升页面响应速度1s
			if option.IsShop {
				return nil
			}
			_, categoryList := s.BuildCategoryList(ctx, req)
			return categoryList
		}(),
		PercentageList: func() []business_data_resp.Percentage {
			taxData := s.statisticsSrv.CountTax(ctx, CountReq{
				TimeType:       req.TimeType,
				QueryStartTime: req.QueryStartTime,
				QueryEndTime:   req.QueryEndTime,
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
		OpeningHours: openingHours,
	}

	return &businessData, nil
}

// CountPaymentMethod 统计支付方式
func (s *businessSrv) CountPaymentMethod(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataPaymentMethod, error) {
	businessSetting, err := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global).GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
		fmt.Println("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)

	paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req)

	var paymentMethodData = business_data_resp.BusinessDataPaymentMethod{
		TotalReceivedPrice:   paymentData.TotalReceivedAmount,
		PaymentMethodIncomes: paymentMethodIncomes,
		OpeningHours:         businessSetting.OpeningHours,
	}

	return &paymentMethodData, nil
}

// CountProductCategory 统计商品分类
func (s *businessSrv) CountProductCategory(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProductCategory, error) {
	businessSetting, err := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global).GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
		fmt.Println("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)

	paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req)
	categoryData, categoryList := s.BuildCategoryList(ctx, req)
	var productCategoryData = business_data_resp.BusinessDataProductCategory{
		SalesNum:             int(categoryData.TotalSaleNum),
		TotalRefundMoney:     paymentData.TotalRefundAmount,
		TotalReceivedPrice:   paymentData.TotalReceivedAmount,
		CategoryList:         categoryList,
		PaymentMethodIncomes: paymentMethodIncomes,
		OpeningHours:         businessSetting.OpeningHours,
	}

	return &productCategoryData, nil
}

// CountProduct 统计商品
func (s *businessSrv) CountProduct(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataProduct, error) {
	businessSetting, err := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global).GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
		fmt.Println("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)
	var productData = business_data_resp.BusinessDataProduct{
		Products: func() []business_data_resp.Product {
			productList := s.statisticsSrv.CountProduct(ctx, CountReq{
				TimeType:       req.TimeType,
				QueryStartTime: req.QueryStartTime,
				QueryEndTime:   req.QueryEndTime,
				CategoryType:   req.CategoryType,
				DutyNo:         req.DutyNo,
			})
			list := make([]business_data_resp.Product, 0, len(productList))
			for _, product := range productList {
				list = append(list, business_data_resp.Product{
					Name:     product.ProductName,
					SalesNum: product.SaleNum,
					Price:    product.SalePrice,
					Subtotal: product.SaleAmount,
				})
			}
			return list
		}(),
		OpeningHours: businessSetting.OpeningHours,
	}

	return &productData, nil
}

// CountArea 统计区域
func (s *businessSrv) CountArea(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataArea, error) {
	areaData := s.statisticsSrv.CountArea(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})

	var areaList = []business_data_resp.Area{}
	for _, area := range areaData {
		areaList = append(areaList, business_data_resp.Area{
			Name:               area.AreaName,
			TotalSales:         area.AreaSaleAmount,
			TotalReceivedPrice: area.AreaBusinessAmount,
			TotalProductNum:    area.AreaProductNum,
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
				QueryStartTime: req.QueryStartTime,
				QueryEndTime:   req.QueryEndTime,
			})
			list := make([]business_data_resp.ProductRank, 0, len(productRankList))
			for _, productRank := range productRankList {
				list = append(list, business_data_resp.ProductRank{
					ProductName: productRank.ProductName,
					SalesNum:    productRank.SaleNum,
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
	productSalesData := s.statisticsSrv.CountProductSale(ctx, CountReq{
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		RankType:       req.SortType,
		RankDirection:  req.SortDirection,
		PageNo:         req.PageNo,
		PageSize:       req.PageSize,
		AreaUuid:       req.AreaUuid,
		CategoryUuid:   req.CategoryUuid,
		ProductName:    req.ProductName,
	})

	var list = []business_data_resp.BusinessDataCountProductSalesItem{}
	for _, productSale := range productSalesData.Data {
		list = append(list, business_data_resp.BusinessDataCountProductSalesItem{
			ProductName:        productSale.ProductName,
			SalesNum:           productSale.TotalSaleNum,
			SalesPrice:         productSale.TotalBusinessAmount,
			CategoryName:       productSale.CategoryName,
			OriginalSalesPrice: productSale.TotalOriginSaleAmount,
			TotalPayPrice:      productSale.TotalActualSaleAmount,
			GiveProductNum:     productSale.TotalGiveNum,
		})
	}

	var productListData = business_data_resp.BusinessDataCountProductSalesPagination{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    productSalesData.Total,
		},
	}

	return &productListData, nil
}

// Count7Days 统计7天
func (s *businessSrv) Count7Days(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataCount7Days, error) {
	sevenDaysData := s.statisticsSrv.Count7Days(ctx, CountReq{
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		Timezone:       ctx.GetCompany().CompanySetting.Timezone,
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

	// 统计免单支付
	var freePaymentData CountFreePaymentResp
	if !req.NotQueryFree {
		freePaymentData = s.statisticsSrv.CountFreePayment(ctx, CountReq{
			DutyNo:         req.DutyNo,
			QueryStartTime: req.QueryStartTime,
			QueryEndTime:   req.QueryEndTime,
			TimeType:       req.TimeType,
			CategoryType:   req.CategoryType,
		})
	}

	paymentMethodIncomes := make([]business_data_resp.PaymentMethodIncome, 0)
	for _, payment := range paymentData.PaymentList {
		if payment.PaymentCode == 0 {
			freePaymentData.PaymentName = payment.PaymentName
			freePaymentData.TotalOrderNum = freePaymentData.TotalOrderNum + payment.TotalOrderNum
			freePaymentData.TotalPaymentAmount = decimal.NewFromFloat(freePaymentData.TotalPaymentAmount).Add(decimal.NewFromFloat(payment.TotalPaymentAmount)).InexactFloat64()
			continue
		} else {
			paymentMethodIncomes = append(paymentMethodIncomes, business_data_resp.PaymentMethodIncome{
				Name:     payment.PaymentName,
				OrderNum: int(payment.TotalOrderNum),
				Amount:   payment.TotalPaymentAmount,
				Code:     payment.PaymentCode,
			})
		}
	}

	if !req.NotQueryFree && freePaymentData.TotalOrderNum > 0 {
		paymentMethodIncomes = append(paymentMethodIncomes, business_data_resp.PaymentMethodIncome{
			Name:     freePaymentData.PaymentName,
			OrderNum: int(freePaymentData.TotalOrderNum),
			Amount:   freePaymentData.TotalPaymentAmount,
			Code:     freePaymentData.PaymentCode,
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
			SalesNum: category.SaleNum,
			Prices:   category.SaleAmount,
		})
	}

	// 排序
	if req.SortType == 1 {
		if req.SortDirection == 1 {
			sort.Slice(list, func(i, j int) bool {
				return list[i].SalesNum < list[j].SalesNum
			})
		} else if req.SortDirection == 2 {
			sort.Slice(list, func(i, j int) bool {
				return list[i].SalesNum > list[j].SalesNum
			})
		}
	} else if req.SortType == 2 {
		if req.SortDirection == 1 {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Prices < list[j].Prices
			})
		} else if req.SortDirection == 2 {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Prices > list[j].Prices
			})
		}
	}

	return categoryData, list
}

// CountExport 统计导出
func (s *businessSrv) CountExport(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataExport, error) {
	exportData, err := s.statisticsSrv.CountExport(ctx, CountReq{
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		Timezone:       ctx.GetCompany().CompanySetting.Timezone,
	})
	if err != nil {
		return nil, err
	}

	exportDataList := make([]business_data_resp.BusinessDataExportItem, 0, len(exportData.Data))
	for _, export := range exportData.Data {
		areaList := make([]business_data_resp.BusinessDataExportArea, 0, len(export.AreaList))
		for _, area := range export.AreaList {
			areaList = append(areaList, business_data_resp.BusinessDataExportArea{
				AreaID:             area.AreaID,
				AreaName:           area.AreaName,
				AreaSaleAmount:     area.AreaSaleAmount,
				AreaBusinessAmount: area.AreaBusinessAmount,
				AreaProductNum:     area.AreaProductNum,
			})
		}

		paymentList := make([]business_data_resp.BusinessDataExportPayment, 0, len(export.PaymentList))
		for _, payment := range export.PaymentList {
			paymentList = append(paymentList, business_data_resp.BusinessDataExportPayment{
				ID:                 payment.ID,
				Sort:               payment.Sort,
				CreateTime:         payment.CreateTime,
				PaymentName:        payment.PaymentName,
				PaymentCode:        payment.PaymentCode,
				TotalOrderNum:      payment.TotalOrderNum,
				TotalPaymentAmount: payment.TotalPaymentAmount,
			})
		}

		memberData := s.statisticsSrv.CountMember(ctx, CountReq{
			QueryStartTime: req.QueryStartTime,
			QueryEndTime:   req.QueryEndTime,
			Timezone:       ctx.GetCompany().CompanySetting.Timezone,
		})

		exportDataList = append(exportDataList, business_data_resp.BusinessDataExportItem{
			Day:                        export.Day,
			TotalSaleAmount:            export.TotalSaleAmount,
			TotalBusinessAmount:        export.TotalBusinessAmount,
			TotalServiceFee:            export.TotalServiceFee,
			TotalPaymentFee:            export.TotalPaymentFee,
			TotalTax:                   export.TotalTax,
			TotalProductNum:            export.TotalProductNum,
			TotalMemberNum:             export.TotalMemberNum,
			TotalDiscountMember:        export.TotalDiscountMember,
			TotalDiscount:              export.TotalDiscount,
			TotalDiscountRatio:         export.TotalDiscountRatio,
			TotalRefundAmount:          export.TotalRefundAmount,
			TotalGiftAmount:            export.TotalGiftAmount,
			TotalGiftNum:               export.TotalGiftNum,
			TotalFreeAmount:            export.TotalFreeAmount,
			TotalFreeNum:               export.TotalFreeNum,
			TotalReceivedAmount:        export.TotalReceivedAmount,
			TotalTakeoutSaleAmount:     export.TotalTakeoutSaleAmount,
			TotalTakeoutBusinessAmount: export.TotalTakeoutBusinessAmount,
			TotalTakeoutRefundAmount:   export.TotalTakeoutRefundAmount,
			TotalTakeoutDeliveryFee:    export.TotalTakeoutDeliveryFee,
			TotalOrderNum:              export.TotalOrderNum,
			MinOrderAmount:             export.MinOrderAmount,
			MaxOrderAmount:             export.MaxOrderAmount,
			AvgOrderAmount:             export.AvgOrderAmount,
			TotalDeskNum:               export.TotalDeskNum,
			TotalMealNum:               export.TotalMealNum,
			MinDeskOrderAmount:         export.MinDeskOrderAmount,
			MaxDeskOrderAmount:         export.MaxDeskOrderAmount,
			AvgDeskOrderAmount:         export.AvgDeskOrderAmount,
			TotalInstantOrderNum:       export.TotalInstantOrderNum,
			MinInstantOrderAmount:      export.MinInstantOrderAmount,
			MaxInstantOrderAmount:      export.MaxInstantOrderAmount,
			AvgInstantOrderAmount:      export.AvgInstantOrderAmount,
			TotalTakeoutOrderNum:       export.TotalTakeoutOrderNum,
			MinTakeoutOrderAmount:      export.MinTakeoutOrderAmount,
			MaxTakeoutOrderAmount:      export.MaxTakeoutOrderAmount,
			AvgTakeoutOrderAmount:      export.AvgTakeoutOrderAmount,
			AreaList:                   areaList,
			PaymentList:                paymentList,
			TotalRechargeAmount:        memberData.TotalRechargeAmount,
		})
	}

	return &business_data_resp.BusinessDataExport{
		Days: exportData.Days,
		Data: exportDataList,
	}, nil
}

// CountShiftRefundAmount 统计班次退款金额
func (s *businessSrv) CountShiftRefundAmount(ctx context.Context, req req.BusinessDataCountReq) *business_data_resp.BusinessDataShiftRefundAmount {
	refundAmount := s.statisticsSrv.CountShiftRefundAmount(ctx, CountReq{DutyNo: req.DutyNo})

	return &business_data_resp.BusinessDataShiftRefundAmount{
		RefundAmount: refundAmount,
	}
}

// CountHome 统计首页
func (s *businessSrv) CountHome(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataHome, error) {
	// 销售数据
	saleData := s.statisticsSrv.CountSale(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
		CategoryType:   req.CategoryType,
		DutyNo:         req.DutyNo,
	})

	// 会员数量
	memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
		TimeType:       req.TimeType,
		QueryStartTime: req.QueryStartTime,
		QueryEndTime:   req.QueryEndTime,
	})

	// 首页数据
	var businessData = business_data_resp.BusinessDataHome{
		TotalReceivedPrice:     saleData.TotalReceivedAmount,
		TotalUserDiscountMoney: saleData.TotalDiscountMember,
		TotalDiscountMoney:     saleData.TotalDiscount,
		TotalRefundMoney:       saleData.TotalRefundAmount,
		TotalOrderNum:          int(saleData.TotalOrderNum),
		MemberData: func() business_data_resp.MemberData {
			memberData := s.statisticsSrv.CountMember(ctx, CountReq{
				TimeType:       req.TimeType,
				QueryStartTime: req.QueryStartTime,
				QueryEndTime:   req.QueryEndTime,
				CategoryType:   req.CategoryType,
			})
			return business_data_resp.MemberData{
				RechargeAmount: memberData.TotalRechargeAmount,
				GiftMoney:      memberData.TotalGiveAmount,
				GiftPoints:     memberData.TotalGivePoint,
				UserCount:      int(memberNum),
			}
		}(),
	}

	return &businessData, nil
}

func (s *businessSrv) CountKitchenEfficiencyAnalysis(ctx context.Context, req req.KitchenEfficiencyAnalysisReq) (*business_data_resp.BusinessDataKitchenEfficiencyAnalysis, error) {
	// 根据名称进行模糊搜索
	// 根据分类
	// 获取商品列表
	productList, err := repository.NewProductRepo(ctx.GetDB()).GetProductListByKeywordAndCategory(req.Keyword, req.CategoryUuids)
	if err != nil {
		return nil, err
	}
	productPackageUuids := make([]uint64, 0)
	efficiencyAnalysisDataList := make(map[uint64]*business_data_resp.KitchenEfficiencyAnalysisItem)
	for _, product := range productList {
		efficiencyAnalysisDataList[product.Uuid] = &business_data_resp.KitchenEfficiencyAnalysisItem{
			ProductPackageUuid: product.Uuid,
			ProductName:        product.MultiLanguageName.GetNames(),
			CategoryName:       product.ProductCategory.MultiLanguageName.GetNames(),
		}
		productPackageUuids = append(productPackageUuids, product.Uuid)
	}

	efficiencyAnalysisResults, err := repository.NewKitchenEfficiencyAnalysisRepo(ctx.GetDB()).GetKitchenEfficiencyAnalysisByProductPackageUuid(productPackageUuids, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	if len(efficiencyAnalysisResults) > 0 {
		for _, result := range efficiencyAnalysisResults {
			efficiencyAnalysisData, ok := efficiencyAnalysisDataList[result.ProductPackageUuid]
			if ok {
				efficiencyAnalysisData.Min = result.Min
				efficiencyAnalysisData.Max = result.Max
				efficiencyAnalysisData.Avg = result.Avg
				efficiencyAnalysisData.SetExist(true)
			}
		}
	} else {
		efficiencyAnalysisDataList = make(map[uint64]*business_data_resp.KitchenEfficiencyAnalysisItem)
	}

	list := make([]business_data_resp.KitchenEfficiencyAnalysisItem, 0)
	for _, efficiencyAnalysisData := range efficiencyAnalysisDataList {
		if efficiencyAnalysisData.GetExist() {
			list = append(list, *efficiencyAnalysisData)
		}
	}

	// 分页返回
	pageNo := req.PageNo
	pageSize := req.PageSize
	total := len(list)
	start := (pageNo - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	list = list[start:end]
	return &business_data_resp.BusinessDataKitchenEfficiencyAnalysis{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    int64(total),
		},
	}, nil
}

func (s *businessSrv) CountKitchenEfficiencyAnalysisAvg(ctx context.Context, req req.KitchenEfficiencyAnalysisAvgReq) (*business_data_resp.BusinessDataKitchenEfficiencyAnalysisAvg, error) {
	avg, err := repository.NewKitchenEfficiencyAnalysisRepo(ctx.GetDB()).GetKitchenEfficiencyAnalysisAvg(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	return &business_data_resp.BusinessDataKitchenEfficiencyAnalysisAvg{
		Avg: avg,
	}, nil
}

func (s *businessSrv) KitchenProductionDetail(ctx context.Context, req req.KitchenProductionDetailReq) (*business_data_resp.KitchenProductionDetail, error) {
	// 参数默认值
	if req.StartTime == 0 {
		dayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		req.StartTime = dayStart.Unix()
	}
	if req.EndTime == 0 {
		dayEnd := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.Local)
		req.EndTime = dayEnd.Unix()
	}
	db := ctx.GetDB()
	productionRepo := repository.NewProductionRepo(db)
	// 根据过滤条件,获取商品bom_uuid列表
	productBomUuids, err := s.getKitchenProductionDetailProductBomUuids(ctx, req)
	if err != nil {
		return nil, err
	}

	opts := []repository.DBOption{
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "SaleOrderProduct.MultiLanguageName",
				Args:  []any{},
			},
			repository.WithPreload{
				Query: "SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
				Args:  []any{},
			},
			repository.WithPreload{
				Query: "ProductCategory.MultiLanguageName",
				Args:  []any{},
			},
		),
	}
	productionOrderProducts, err := productionRepo.GetProductionOrderList(req.PageNo, req.PageSize, productBomUuids, req.StartTime, req.EndTime, opts...)
	if err != nil {
		return nil, err
	}

	count, err := productionRepo.GetProductionOrderListCount(productBomUuids, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	productionDataList := make([]business_data_resp.KitchenProductionDetailItem, 0)
	for _, productionOrderProduct := range productionOrderProducts {
		item := business_data_resp.KitchenProductionDetailItem{
			ProductName:    productionOrderProduct.SaleOrderProduct.MultiLanguageName.GetNames(),
			FlavorName:     productionOrderProduct.SaleOrderProduct.GetFlavorName(),
			CategoryName:   productionOrderProduct.ProductCategory.MultiLanguageName.GetNames(),
			Number:         productionOrderProduct.Num,
			CreateTime:     productionOrderProduct.CreateTime,
			MakeFinishTime: productionOrderProduct.MadeTime,
			MakeDuration:   productionOrderProduct.MakeDuration,
			SendFinishTime: int64(productionOrderProduct.FinishedTime),
			SendDuration:   productionOrderProduct.SendDuration,
			FinishTime:     int64(productionOrderProduct.FinishedTime),
			AllDuration:    productionOrderProduct.AllDuration,
		}
		productionDataList = append(productionDataList, item)
	}

	return &business_data_resp.KitchenProductionDetail{
		List: productionDataList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    count,
		},
	}, nil
}

func (s *businessSrv) getKitchenProductionDetailProductBomUuids(ctx context.Context, req req.KitchenProductionDetailReq) ([]uint64, error) {
	db := ctx.GetDB()
	productBomUuids := []uint64{}
	// 根据Keyword(商品名称)查询所有商品的product_bom_uuid
	if req.Keyword != "" {
		productBomUuidsByName, err := repository.NewProductRepo(db).GetProductBomUuidsByKeyword(req.Keyword)
		if err != nil {
			return nil, err
		}
		if len(productBomUuidsByName) > 0 {
			productBomUuids = append(productBomUuids, productBomUuidsByName...)
		}
	}
	// 根据Keyword(内部编码)查询所有商品的product_bom_uuid
	if req.Keyword != "" {
		productBomUuidsByInternalCode, err := repository.NewProductRepo(db).GetProductBomUuidsByInternalCode(req.Keyword)
		if err != nil {
			return nil, err
		}
		if len(productBomUuidsByInternalCode) > 0 {
			productBomUuids = append(productBomUuids, productBomUuidsByInternalCode...)
		}
	}
	// 根据CategoryUuids查询所有商品的product_bom_uuid
	if len(req.CategoryUuids) > 0 {
		productBomUuidsByCategoryUuids, err := repository.NewProductRepo(db).GetProductBomUuidsByCategoryUuids(req.CategoryUuids)
		if err != nil {
			return nil, err
		}
		if len(productBomUuidsByCategoryUuids) > 0 {
			productBomUuids = append(productBomUuids, productBomUuidsByCategoryUuids...)
		}
	}
	return productBomUuids, nil
}

// KitchenProductionDetailCount 统计后厨菜品出品明细数量
func (s *businessSrv) KitchenProductionDetailCount(ctx context.Context, req req.KitchenProductionDetailReq) (int64, error) {
	// 参数默认值
	if req.StartTime == 0 {
		dayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		req.StartTime = dayStart.Unix()
	}
	if req.EndTime == 0 {
		dayEnd := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.Local)
		req.EndTime = dayEnd.Unix()
	}
	db := ctx.GetDB()
	productionRepo := repository.NewProductionRepo(db)
	// 根据过滤条件,获取商品bom_uuid列表
	productBomUuids, err := s.getKitchenProductionDetailProductBomUuids(ctx, req)
	if err != nil {
		return 0, err
	}

	count, err := productionRepo.GetProductionOrderListCount(productBomUuids, req.StartTime, req.EndTime)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 导出后厨效率分析数据
func (s *businessSrv) ExportKitchenEfficiencyAnalysis(ctx context.Context, request req.KitchenEfficiencyAnalysisReq) error {
	request.PageNo = 1
	request.PageSize = 1000 // 最多导出1000条数据
	result, err := s.CountKitchenEfficiencyAnalysis(ctx, request)
	if err != nil {
		return err
	}
	if result.Meta.Total > 1000 {
		return errors.WithMessage(errors.New("最多导出1000条数据"))
	}

	fileNameMap := map[string]string{
		"zh":    "菜品出品详情%s.xlsx",                       // 中文
		"th":    "รายละเอียดเมนูอาหาร%s.xlsx",          // 泰文
		"en":    "Details of Dish Presentation%s.xlsx", // 英文
		"zhtw":  "菜品出品詳情%s.xlsx",                       // 繁体中文
		"zh_tw": "菜品出品詳情%s.xlsx",                       // 繁体中文
		"ja":    "料理の提供詳細%s.xlsx",                      // 日文
		"ko":    "메뉴 세부 정보%s.xlsx",              // 韩文
		"my":    "ဟင်းလျာအသေးစိတ်%s.xlsx",              // 缅甸文
		"tr":    "Yemek Sunumu Detayları%s.xlsx",       // 土耳其文
		"sv":    "Maträttsdetaljer%s.xlsx",            // 瑞典文
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	timezoneUtils := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
	dateString := timezoneUtils.FormatUnixTime(time.Now().Unix(), "2006-01-02")
	fileName := fmt.Sprintf(fileNameMap[ctx.GetLanguage()], dateString)
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeKitchenEfficiencyAnalysis,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(params),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	db := ctx.GetDB()
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 判断是否还有正在导出的任务
		oldRecord, err := repository.NewExportRecordRepo(tx).GetUnfinishedExportRecord(model.ExportTypeKitchenEfficiencyAnalysis)
		if err != nil {
			return err
		}
		if oldRecord != nil {
			return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
		}

		err = repository.NewExportRecordRepo(tx).Create(record)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		db := ctx.GetDB()
		record, err := repository.NewExportRecordRepo(db).GetByUuid(record.Uuid)
		if err != nil {
			logger.Logger.Error("获取导出ExportKitchenProductionDetail失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		}
		if record == nil {
			logger.Logger.Error("导出记录不存在", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": "导出记录不存在",
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		}
		params := &req.KitchenEfficiencyAnalysisReq{}
		if err := json.Unmarshal([]byte(record.ExportParams), params); err != nil {
			logger.Logger.Error("获取导出ExportKitchenProductionDetail失败,解析参数失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		}
		if res, err := s.ExportKitchenEfficiencyAnalysisTask(ctx, *params); err != nil {
			logger.Logger.Error("导出ExportKitchenProductionDetailTask失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		} else {
			record.FileUuid = res.FileUuid
			record.Status = model.ExportStatusSuccess
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"file_uuid": record.FileUuid,
				"status":    record.Status,
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
		}
	})

	return nil
}

// 导出厨房出品明细数据到谷歌桶
func (s *businessSrv) ExportKitchenProductionDetail(ctx context.Context, request req.KitchenProductionDetailReq) error { // 修改返回类型
	request.PageNo = 1
	request.PageSize = 1000 // 最多导出1000条数据
	count, err := s.KitchenProductionDetailCount(ctx, request)
	if err != nil {
		return err
	}
	if count > 1000 {
		return errors.WithMessage(errors.New("最多导出1000条数据"))
	}

	fileNameMap := map[string]string{
		"zh":    "菜品出品明细%s.xlsx",                                // 中文
		"th":    "รายละเอียดการออกอาหาร%s.xlsx",                 // 泰文
		"en":    "Detailed List of Dish Output%s.xlsx",          // 英文
		"zhtw":  "菜品出品明細%s.xlsx",                                // 繁体中文
		"zh_tw": "菜品出品明細%s.xlsx",                                // 繁体中文
		"ja":    "料理の提供明細%s.xlsx",                               // 日文
		"ko":    "메뉴 출품 상세 내역%s.xlsx",               // 韩文
		"my":    "ဟင်းလျာထုတ်လုပ်မှု အသေးစိတ်စာရင်း%s.xlsx",     // 缅甸文
		"tr":    "Yemek Çıkış Ayrıntıları%s.xlsx",             // 土耳其文
		"sv":    "Detaljerad lista över maträttsutbud%s.xlsx", // 瑞典文
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	timezoneUtils := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
	dateString := timezoneUtils.FormatUnixTime(time.Now().Unix(), "2006-01-02")
	fileName := fmt.Sprintf(fileNameMap[ctx.GetLanguage()], dateString)
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeKitchenProductionDetail,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(params),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	db := ctx.GetDB()
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 判断是否还有正在导出的任务
		oldRecord, err := repository.NewExportRecordRepo(tx).GetUnfinishedExportRecord(model.ExportTypeKitchenProductionDetail)
		if err != nil {
			return err
		}
		if oldRecord != nil {
			return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
		}

		err = repository.NewExportRecordRepo(tx).Create(record)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		db := ctx.GetDB()
		record, err := repository.NewExportRecordRepo(db).GetByUuid(record.Uuid)
		if err != nil {
			logger.Logger.Error("获取导出ExportKitchenProductionDetail失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		}
		if record == nil {
			logger.Logger.Error("导出记录不存在", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": "导出记录不存在",
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		}
		params := &req.KitchenProductionDetailReq{}
		if err := json.Unmarshal([]byte(record.ExportParams), params); err != nil {
			logger.Logger.Error("获取导出ExportKitchenProductionDetail失败,解析参数失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		}
		if res, err := s.ExportKitchenProductionDetailTask(ctx, *params); err != nil {
			logger.Logger.Error("导出ExportKitchenProductionDetailTask失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
			return
		} else {
			record.FileUuid = res.FileUuid
			record.Status = model.ExportStatusSuccess
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"file_uuid": record.FileUuid,
				"status":    record.Status,
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
				return
			}
		}
	})

	return nil
}

// 统计某商家当天的商品后厨效率
func (s *businessSrv) StatsKitchenEfficiencyAnalysis(ctx context.Context) (string, error) {
	db := ctx.GetDB()
	timezone := ctx.GetCompanySetting().Timezone
	timezoneUtils := utils.SetTimezone(timezone)
	startTime, endTime := timezoneUtils.TodayStartEndUnix()
	dateString := timezoneUtils.FormatUnixTime(startTime, "2006-01-02")

	logger.Logger.Info("统计当天后厨效率分析数据", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.String("timezone", timezone), zap.String("dateString", dateString), zap.Int64("startTime", startTime), zap.Int64("endTime", endTime))

	// 统计当天各个商品的后厨效率分析数据
	kitchenEfficiencyAnalysisRepo := repository.NewKitchenEfficiencyAnalysisRepo(db)
	kitchenEfficiencyAnalysis, err := kitchenEfficiencyAnalysisRepo.CalculateKitchenEfficiencyAnalysis(startTime, endTime)
	if err != nil {
		return "", err
	}

	// 查询效率表中是否存在当天数据
	list, err := kitchenEfficiencyAnalysisRepo.KitchenEfficiencyAnalysisDay(startTime, endTime)
	if err != nil {
		return "", err
	}

	existItemMap := make(map[uint64]*model.KitchenEfficiencyAnalysis)

	if len(list) > 0 {
		for _, item := range list {
			if result, ok := existItemMap[item.ProductPackageUuid]; ok { // 正常数据是没有重复的,如果重复了,则说明数据有问题
				logger.Logger.Warn("商品后厨效率分析数据已存在", zap.Any("product_package_uuid", item.ProductPackageUuid), zap.Any("result", *result))
				continue
			}
			existItemMap[item.ProductPackageUuid] = item
		}
	}

	createList := make([]*model.KitchenEfficiencyAnalysis, 0)
	for _, item := range kitchenEfficiencyAnalysis {
		if result, ok := existItemMap[item.ProductPackageUuid]; ok {
			// 存在商品,则更新分析数据
			result.Min = item.Min
			result.Max = item.Max
			result.Avg = item.Avg
			result.Total = item.Total
			result.Count = item.Count
			//
			result.Date = startTime
			result.DateString = dateString
			result.Timezone = timezone
			result.SetUpdate()
		} else {
			item := &model.KitchenEfficiencyAnalysis{
				ProductPackageUuid: item.ProductPackageUuid,
				Min:                item.Min,
				Max:                item.Max,
				Avg:                item.Avg,
				Total:              item.Total,
				Count:              item.Count,
				Date:               startTime,
				DateString:         dateString,
				Timezone:           timezone,
			}
			createList = append(createList, item)
		}
	}

	updateList := make([]*model.KitchenEfficiencyAnalysis, 0)
	for _, item := range list {
		if item, ok := existItemMap[item.ProductPackageUuid]; ok {
			if item.GetUpdate() {
				updateList = append(updateList, item)
			}
		}
	}

	//
	if err := db.Transaction(func(tx *gorm.DB) error {
		if len(updateList) > 0 {
			for _, item := range updateList {
				item.UpdateTime = time.Now().Unix()
				if err := tx.Model(&model.KitchenEfficiencyAnalysis{}).Where("product_package_uuid = ?", item.ProductPackageUuid).Where("date = ?", item.Date).Updates(item).Error; err != nil {
					return err
				}
			}
		}
		if len(createList) > 0 {
			for _, item := range createList {
				item.CreateTime = time.Now().Unix()
				item.UpdateTime = time.Now().Unix()
				if err := repository.NewKitchenEfficiencyAnalysisRepo(tx).CreateKitchenEfficiencyAnalysis(*item); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	return "success", nil
}

// CountBusinessTimePeriod 统计营业时段
func (s *businessSrv) CountBusinessTimePeriod(ctx context.Context, req req.BusinessTimePeriodReq) business_data_resp.BusinessTimePeriod {
	// 调用统计服务
	businessTimePeriodData := s.statisticsSrv.CountBusinessTimePeriod(ctx, req)

	// 构建返回列表
	businessTimePeriodList := make([]business_data_resp.BusinessTimePeriodItem, 0, len(businessTimePeriodData.BusinessTimePeriodList))
	for i, period := range businessTimePeriodData.BusinessTimePeriodList {
		no := (req.PageNo-1)*req.PageSize + i + 1
		businessTimePeriodList = append(businessTimePeriodList, business_data_resp.BusinessTimePeriodItem{
			No:                 no,
			TimePeriod:         period.TimePeriod,
			OrderAmount:        period.OrderAmount,
			OrderAmountMealAvg: period.OrderAmountMealAvg,
			PayAmount:          period.PayAmount,
			PayAmountMealAvg:   period.PayAmountMealAvg,
			OrderNum:           period.OrderNum,
			MealNum:            period.MealNum,
		})
	}

	return business_data_resp.BusinessTimePeriod{
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    businessTimePeriodData.TotalBusinessTimePeriodNum,
		},
		List: businessTimePeriodList,
	}
}

// CountBusinessComprehensiveOperations 统计综合运用
func (s *businessSrv) CountBusinessComprehensiveOperations(ctx context.Context, req req.StatisticsComprehensiveOperationsReq) business_data_resp.StatisticsComprehensiveOperations {
	// 调用统计服务
	businessComprehensiveOperationsData := s.statisticsSrv.CountBusinessComprehensiveOperations(ctx, req)

	// 构建返回列表
	businessComprehensiveOperationsList := make([]business_data_resp.StatisticsComprehensiveOperationsItem, 0, len(businessComprehensiveOperationsData.StatisticsComprehensiveList))
	for i, item := range businessComprehensiveOperationsData.StatisticsComprehensiveList {
		no := (req.PageNo-1)*req.PageSize + i + 1
		businessComprehensiveOperationsList = append(businessComprehensiveOperationsList, business_data_resp.StatisticsComprehensiveOperationsItem{
			No:                 no,
			Date:               item.Date,
			OrderAmount:        item.OrderAmount,
			PayAmount:          item.PayAmount,
			OrderNum:           item.OrderNum,
			MealNum:            item.MealNum,
			DeskNum:            item.DeskNum,
			OrderAmountMealAvg: item.OrderAmountMealAvg,
			PayAmountMealAvg:   item.PayAmountMealAvg,
			OrderAmountAvg:     item.OrderAmountAvg,
			PayAmountAvg:       item.PayAmountAvg,
			InstantOrderAmount: item.InstantOrderAmount,
			DeskOrderAmount:    item.DeskOrderAmount,
			TakeoutOrderAmount: item.TakeoutOrderAmount,
		})
	}

	return business_data_resp.StatisticsComprehensiveOperations{
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    businessComprehensiveOperationsData.TotalStatisticsComprehensiveNum,
		},
		List: businessComprehensiveOperationsList,
	}
}

// CountBusinessPaymentMethod 统计支付方式
func (s *businessSrv) CountBusinessPaymentMethod(ctx context.Context, req req.StatisticsPaymentMethodReq) business_data_resp.StatisticsPaymentMethod {
	// 调用统计服务
	businessPaymentMethodData := s.statisticsSrv.CountBusinessPaymentMethod(ctx, req)

	// 构建返回列表
	businessPaymentMethodList := make([]business_data_resp.StatisticsPaymentMethodItem, 0, len(businessPaymentMethodData.StatisticsPaymentMethodList))
	for i, item := range businessPaymentMethodData.StatisticsPaymentMethodList {
		no := (req.PageNo-1)*req.PageSize + i + 1
		businessPaymentMethodList = append(businessPaymentMethodList, business_data_resp.StatisticsPaymentMethodItem{
			No:            no,
			Date:          item.Date,
			PaymentName:   item.PaymentName,
			PaymentNum:    int(item.PaymentNum),
			PaymentAmount: item.PaymentAmount,
		})
	}

	return business_data_resp.StatisticsPaymentMethod{
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    businessPaymentMethodData.TotalStatisticsPaymentMethodNum,
		},
		List: businessPaymentMethodList,
	}
}

// 导出厨房出品明细数据到谷歌桶
func (s *businessSrv) ExportKitchenProductionDetailTask(ctx context.Context, req req.KitchenProductionDetailReq) (*resp.FileExportResp, error) { // 修改返回类型
	req.PageNo = 1
	req.PageSize = 1000 // 最多导出1000条数据
	productionDetail, err := s.KitchenProductionDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	// 商家时区
	timezone := ctx.GetCompanySetting().Timezone
	timezoneUtil := utils.SetTimezone(timezone)
	// 文件名多语言
	fileNameMul := model.MultiLanguageName{
		EnName:   "Detailed List of Dish Output",          // 英文
		ZhName:   "菜品出品明细",                                // 中文
		ZhTwName: "菜品出品明細",                                // 繁体中文
		ThName:   "รายละเอียดการออกอาหาร",                 // 泰语
		MyName:   "ဟင်းလျာထုတ်လုပ်မှု အသေးစိတ်စာရင်း",     // 缅甸语
		JaName:   "料理の提供明細",                               // 日语
		KoName:   "메뉴 출품 상세 내역",               // 韩语
		TrName:   "Yemek Çıkış Ayrıntıları",             // 土耳其语
		SvName:   "Detaljerad lista över maträttsutbud", // 瑞典语
	}
	headerMap := map[string][]string{
		"zh": { // 中文
			"名称", "规格", "分类", "完成数量", "下单时间", "制作完成时间", "制作总耗时", "传菜完成时间", "传菜总耗时", "完成时间", "总耗时",
		},
		"th": { // 泰语
			"ชื่อ", "ข้อกำหนด", "ประเภท", "จำนวนที่เสร็จสิ้น", "เวลาสั่งซื้อ", "เวลาเสร็จสิ้นการผลิต", "เวลาที่ใช้ในการผลิตทั้งหมด", "เวลาเสร็จสิ้นการส่งอาหาร", "เวลาที่ใช้ในการส่งอาหารทั้งหมด", "เวลาเสร็จสิ้น", "เวลาที่ใช้ทั้งหมด",
		},
		"en": { // 英文
			"Name", "Specification", "Category", "Quantity Completed", "Order Placement Time", "Production Completion Time", "Total Production Time", "Food Delivery Completion Time", "Total Food Delivery Time", "Completion Time", "Total Time Consumed",
		},
		"zhtw": { // 繁体中文
			"名稱", "規格", "分類", "完成數量", "下單時間", "製作完成時間", "製作總耗時", "傳菜完成時間", "傳菜總耗時", "完成時間", "總耗時",
		},
		"zh_tw": { //
			"名稱", "規格", "分類", "完成數量", "下單時間", "製作完成時間", "製作總耗時", "傳菜完成時間", "傳菜總耗時", "完成時間", "總耗時",
		},
		"ja": { // 日语
			"名称", "規格", "分類", "完了数量", "発注時刻", "製作完了時刻", "製作総時間", "伝菜完了時刻", "伝菜総時間", "完了時刻", "総時間",
		},
		"ko": { // 韩语
			"이름", "규격", "카테고리", "완료 수량", "주문 시간", "제작 완료 시간", "제작 총 소요 시간", "전송 완료 시간", "전송 총 소요 시간", "완료 시간", "총 소요 시간",
		},
		"my": { // 缅甸语
			"အမည်", "အရွယ်အစား", "အမျိုးအစား", "ပြီးစီးအရေအတွက်", "မှာယူချိန်", "ပြုလုပ်ပြီးစီးချိန်", "ပြုလုပ်ရန်အချိန်ယူမှုစုစုပေါင်း", "စားပွဲတင်ပြီးစီးချိန်", "စားပွဲတင်အချိန်ယူမှုစုစုပေါင်း", "ပြီးစီးချိန်", "စုစုပေါင်းအချိန်ယူမှု",
		},
		"tr": { // 土耳其语
			"Ad", "özellikler", "Kategori", "Tamamlanan Miktar", "Sipariş Verme Zamanı", "Üretim Tamamlama Zamanı", "Toplam Üretim Süresi", "Yemek Servis Tamamlama Zamanı", "Toplam Yemek Servis Süresi", "Tamamlama Zamanı", "Toplam Süre",
		},
		"sv": { // 瑞典语
			"Namn", "specifikation", "Kategori", "Antal Färdigställda", "Beställningstid", "Produktionsfärdigställningstid", "Total Produktionstid", "Matleveransfärdigställningstid", "Total Matleveranstid", "Färdigställningstid", "Total Tid",
		},
	}
	fileName := fmt.Sprintf("%s_%d.xlsx", fileNameMul.GetNameByLang(ctx.GetLanguage()), time.Now().Unix()) // 文件名,格式
	xlsxFile := excelize.NewFile()                                                                         // 修改这里，直接使用 NewFile()
	sheetName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	// 创建一个新的工作表
	index, err := xlsxFile.NewSheet(sheetName)
	if err != nil {
		logger.Logger.Error("创建Excel工作表失败", zap.Error(err))
		return nil, errors.WithMessage(err)
	}
	xlsxFile.SetActiveSheet(index)

	lang := ctx.GetLanguage()
	// 设置表头
	headers := func() []string {
		headers := headerMap[lang]
		if headers == nil {
			headers = headerMap["en"]
		}
		return headers
	}() // 根据语言获取表头

	// 写入表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1) // 从第一行开始
		xlsxFile.SetCellValue(sheetName, cell, header)
		// 设置样式，加粗表头
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheetName, cell, cell, style)
	}

	// 写入数据
	for rowIdx, item := range productionDetail.List {
		// 数据从第二行开始写入
		offsetRow := rowIdx + 2 // +1 for 0-based index, +1 for header row
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", offsetRow), item.ProductName.GetLocale(lang))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", offsetRow), item.FlavorName.GetLocale(lang))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", offsetRow), item.CategoryName.GetLocale(lang))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", offsetRow), item.Number)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", offsetRow), timezoneUtil.FormatUnixTime(item.CreateTime, "2006-01-02 15:04:05"))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("F%d", offsetRow), timezoneUtil.FormatUnixTime(item.MakeFinishTime, "2006-01-02 15:04:05"))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("G%d", offsetRow), utils.FormatIntToTime(item.MakeDuration))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("H%d", offsetRow), timezoneUtil.FormatUnixTime(item.SendFinishTime, "2006-01-02 15:04:05"))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("I%d", offsetRow), utils.FormatIntToTime(item.SendDuration))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("J%d", offsetRow), timezoneUtil.FormatUnixTime(item.FinishTime, "2006-01-02 15:04:05"))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("K%d", offsetRow), utils.FormatIntToTime(item.AllDuration))
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheetName, colName, colName, 20)
	}

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		logger.Logger.Error("写入Excel文件到内存失败", zap.Error(err))
		return nil, errors.WithMessage(err)
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, fileName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil // 返回文件UUID
}

// 导出后厨菜品出品效率分析到谷歌桶
func (s *businessSrv) ExportKitchenEfficiencyAnalysisTask(ctx context.Context, req req.KitchenEfficiencyAnalysisReq) (*resp.FileExportResp, error) { // 修改返回类型
	req.PageNo = 1
	req.PageSize = 1000 // 最多导出1000条数据
	result, err := s.CountKitchenEfficiencyAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	// 文件名多语言
	fileNameMul := model.MultiLanguageName{
		EnName:   "Details of Dish Presentation",      // 英文
		ZhName:   "菜品出品明细",                            // 中文
		ZhTwName: "菜品出品詳情",                            // 繁体中文
		ThName:   "รายละเอียดเมนูอาหาร",               // 泰语
		MyName:   "ဟင်းလျာထုတ်လုပ်မှု အသေးစိတ်စာရင်း", // 缅甸语
		JaName:   "料理の提供詳細",                           // 日语
		KoName:   "메뉴 세부 정보",                   // 韩语
		TrName:   "Yemek Sunumu Detayları",            // 土耳其语
		SvName:   "Maträttsdetaljer",                 // 瑞典语
	}
	headerMap := map[string][]string{
		"zh": { // 中文
			"名称", "分类", "最短出品时长", "最最长出品时长值", "平均出品时长",
		},
		"th": { // 泰语
			"ชื่อ", "ประเภท", "ประเภท", "ระยะเวลาในการผลิตที่สั้นที่สุด", "ระยะเวลาในการผลิตที่ยาวที่สุด", "ระยะเวลาในการผลิตเฉลี่ย",
		},
		"en": { // 英文
			"Name", "Category", "Shortest Production Time", "Longest Production Time", "Average Production Time",
		},
		"zhtw": { // 繁体中文
			"名稱", "分類", "最短出品時長", "最長出品時長", "平均出品時長",
		},
		"zh_tw": { // 繁体中文
			"名稱", "分類", "最短出品時長", "最長出品時長", "平均出品時長",
		},
		"ja": { // 日语
			"名称", "分類", "最短の製造時間", "最長の製造時間", "平均製造時間",
		},
		"ko": { // 韩语
			"명칭", "규격", "분류", "최단 제작 시간", "최단 제작 시간", "평균 제작 시간",
		},
		"my": { // 缅甸语
			"အမည်", "အမျိုးအစား", "ထုတ်လုပ်ချိန်အတိုဆုံး", "ထုတ်လုပ်ချိန်အရှည်ဆုံး", "ပျမ်းမျှထုတ်လုပ်ချိန်",
		},
		"tr": { // 土耳其语
			"İsim", "Kategori", "En Kısa Üretim Süresi", "En Uzun Üretim Süresi", "Ortalama Üretim Süresi",
		},
		"sv": { // 瑞典语
			"Namn", "Kategori", "Kortaste Produktionstid", "Längsta Produktionstid", "Genomsnittlig Produktionstid",
		},
	}
	fileName := fmt.Sprintf("%s_%d.xlsx", fileNameMul.GetNameByLang(ctx.GetLanguage()), time.Now().Unix()) // 文件名,格式
	xlsxFile := excelize.NewFile()                                                                         // 修改这里，直接使用 NewFile()
	sheetName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	// 创建一个新的工作表
	index, err := xlsxFile.NewSheet(sheetName)
	if err != nil {
		logger.Logger.Error("创建Excel工作表失败", zap.Error(err))
		return nil, errors.WithMessage(err)
	}
	xlsxFile.SetActiveSheet(index)

	lang := ctx.GetLanguage()
	// 设置表头
	headers := func() []string {
		headers := headerMap[lang]
		if headers == nil {
			headers = headerMap["en"]
		}
		return headers
	}() // 根据语言获取表头

	// 写入表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1) // 从第一行开始
		xlsxFile.SetCellValue(sheetName, cell, header)
		// 设置样式，加粗表头
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheetName, cell, cell, style)
	}

	// 写入数据
	for rowIdx, item := range result.List {
		// 数据从第二行开始写入
		offsetRow := rowIdx + 2 // +1 for 0-based index, +1 for header row
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", offsetRow), item.ProductName.GetLocale(lang))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", offsetRow), item.CategoryName.GetLocale(lang))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", offsetRow), utils.FormatIntToTime(int64(item.Min)))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", offsetRow), utils.FormatIntToTime(int64(item.Max)))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", offsetRow), utils.FormatIntToTime(int64(item.Avg)))
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheetName, colName, colName, 20)
	}

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		logger.Logger.Error("写入Excel文件到内存失败", zap.Error(err))
		return nil, errors.WithMessage(err)
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, fileName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil // 返回文件UUID
}
