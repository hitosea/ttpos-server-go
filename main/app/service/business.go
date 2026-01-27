package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer"
	"ttpos-server-go/app/modules/printer/template"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	pkgLanguage "ttpos-server-go/pkg/language"
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
	ExportProductSales(ctx context.Context, req req.BusinessDataCountProductSalesReq) error                                                                               // 导出商品销售统计
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
	CountBusinessSummary(ctx context.Context, req req.StatisticsSummaryReq) business_data_resp.StatisticsSummary                                                          // 统计综合运用数据
	CountBusinessPaymentMethod(ctx context.Context, req req.StatisticsPaymentMethodReq) business_data_resp.StatisticsPaymentMethod                                        // 统计收款数据
	ExportBusinessTimePeriod(ctx context.Context, req req.BusinessTimePeriodReq) error                                                                                    // 导出时段营业统计数据
	ExportBusinessSummary(ctx context.Context, req req.StatisticsSummaryReq) error                                                                                        // 导出综合运营统计数据
	ExportBusinessPaymentMethod(ctx context.Context, req req.StatisticsPaymentMethodReq) error                                                                            // 导出收款统计数据
	CountChannelSales(ctx context.Context, req req.ChannelSalesReq) (*resp.ChannelSalesResp, error)                                                                       // 统计渠道营业数据
	ExportChannelSales(ctx context.Context, req req.ChannelSalesReq) error                                                                                                // 导出渠道营业统计数据
	CountCompanyBusinessSummary(ctx context.Context, req req.StatisticsCompanySummaryReq) (interface{}, error)                                                            // 获取门店汇总统计（营业数据汇总、支付方式汇总、退款金额汇总）
	ExportCompanyBusinessSummary(ctx context.Context, req req.StatisticsCompanySummaryReq) error                                                                          // 导出门店汇总统计
	GetCompanyPaymentMethods(ctx context.Context) (*resp.CompanyPaymentMethodListResp, error)                                                                             // 获取有权限的所有门店的支付方式（汇总去重）
	CountUserAnalysis(ctx context.Context, req req.UserAnalysisReq) (*resp.UserAnalysisResp, error)                                                                       // 统计用户分析数据
	ExportUserAnalysis(ctx context.Context, req req.UserAnalysisReq) error                                                                                                // 导出用户分析统计数据
	GetCompanyList(ctx context.Context) (*resp.CompanySummaryListResp, error)                                                                                             // 获取门店汇总统计可选择的门店列表
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
		// 业务逻辑错误后继续执行（可能导致 nil 指针）-- 不会导致，storeSetting是结构体，不会为 nil, printerSetting \ businessSetting 同理
	}

	// 获取打印机设置
	printerSetting, err := setting.GetPrinterSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取打印机设置失败", zap.Error(err))
	}

	// 获取门店业务设置
	businessSetting, err := setting.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店业务设置失败", zap.Error(err))
	}

	// 获取数据管理设置
	companySetting := ctx.GetCompanySetting()
	dataSetting := setting.GetDataManageSetting(ctx)
	excludeDataManage := companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage

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
			QueryStartTime:    int64(printerParam.QueryStartTime),
			QueryEndTime:      int64(printerParam.QueryEndTime),
			CategoryType:      printerParam.CategoryType,
			ExcludeDataManage: excludeDataManage,
		})
		// 支付数据
		_, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req.BusinessDataCountReq{
			QueryStartTime:    int64(printerParam.QueryStartTime),
			QueryEndTime:      int64(printerParam.QueryEndTime),
			CategoryType:      printerParam.CategoryType,
			NotQueryFree:      true,
			ExcludeDataManage: excludeDataManage,
		})
		// 会员数量
		memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
			QueryStartTime: int64(printerParam.QueryStartTime),
			QueryEndTime:   int64(printerParam.QueryEndTime),
		})
		// 未结订单
		unpaidOrderData := s.statisticsSrv.CountUnpaidOrder(ctx, CountReq{
			QueryStartTime:    int64(printerParam.QueryStartTime),
			QueryEndTime:      int64(printerParam.QueryEndTime),
			ExcludeDataManage: excludeDataManage,
		})

		reqPrinterData.All = &business_data_resp.BusinessDataAll{
			TotalSales:             saleData.TotalSaleAmount,
			TotalReceivedPrice:     saleData.TotalReceivedAmount,
			TotalPayPrice:          saleData.TotalBusinessAmount,
			TotalProductPrice:      saleData.TotalProductOriginPrice,
			TotalPayFeeMoney:       saleData.TotalPaymentFee,
			TotalServiceMoney:      saleData.TotalServiceFee,
			TotalTaxMoney:          saleData.TotalTax,
			TotalUserDiscountMoney: saleData.TotalDiscountMember,
			TotalDiscountMoney:     saleData.TotalDiscount,
			TotalDiscountRatio:     saleData.TotalDiscountRatio,
			TotalFreeOrderPrice:    saleData.TotalFreeAmount,
			TotalRefundMoney:       saleData.TotalRefundAmount,
			TotalOrderNum:          int(saleData.TotalOrderNum),
			TotalPeopleNum:         int(saleData.TotalMealNum),
			TotalProductNum:        saleData.TotalProductNum,
			TotalTableNum:          int(saleData.TotalDeskNum),
			TotalCancelOrderNum:    int(saleData.TotalCancelOrderNum),
			TotalCancelOrderAmount: saleData.TotalCancelOrderAmount,
			TotalGiveProductPrice:  saleData.TotalGiftAmount,
			TotalGiveProductNum:    saleData.TotalGiftNum,
			AvgOrderPrice:          saleData.AvgOrderAmount,
			MinOrderPrice:          saleData.MinOrderAmount,
			MaxOrderPrice:          saleData.MaxOrderAmount,
			// 桌台方式
			AllTableOrderNum:      int(saleData.TotalDeskNum),
			AllTablePeopleNum:     int(saleData.TotalMealNum),
			AllTableAvgOrderPrice: saleData.AvgDeskOrderAmount,
			AllTableMinOrderPrice: saleData.MinDeskOrderAmount,
			AllTableMaxOrderPrice: saleData.MaxDeskOrderAmount,
			AllTablePeopleAvg:     saleData.AvgDeskPeopleOrderAmount,
			// 收银方式-店内
			AllCashierOrderNum:      int(saleData.TotalInstantOrderNum),
			AllCashierMinOrderPrice: saleData.MinInstantOrderAmount,
			AllCashierMaxOrderPrice: saleData.MaxInstantOrderAmount,
			AllCashierAvgOrderPrice: saleData.AvgInstantOrderAmount,
			// 收银方式-外卖
			AllTakeawayOrderNum:      int(saleData.TotalInstantOrderTakeawayNum),
			AllTakeawayMinOrderPrice: saleData.MinInstantOrderTakeawayAmount,
			AllTakeawayMaxOrderPrice: saleData.MaxInstantOrderTakeawayAmount,
			AllTakeawayAvgOrderPrice: saleData.AvgInstantOrderTakeawayAmount,
			// 未结账数据
			UnclosedTotalOrderNum: int(unpaidOrderData.TotalOrderNum),
			UnclosedTotalPrice:    unpaidOrderData.TotalAmount,
			PaymentMethodIncomes:  paymentMethodIncomes,
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
					excludeDataManage,
				)
				if err != nil {
					return []business_data_resp.PeakHour{}
				}
				return peakHours
			}(),
			CategoryList: func() []business_data_resp.Category {
				_, categoryList := s.BuildCategoryList(ctx, req.BusinessDataCountReq{
					QueryStartTime:    int64(printerParam.QueryStartTime),
					QueryEndTime:      int64(printerParam.QueryEndTime),
					CategoryType:      printerParam.CategoryType,
					ExcludeDataManage: excludeDataManage,
				})
				return categoryList
			}(),
			PercentageList: func() []business_data_resp.Percentage {
				taxData := s.statisticsSrv.CountTax(ctx, CountReq{
					QueryStartTime:    int64(printerParam.QueryStartTime),
					QueryEndTime:      int64(printerParam.QueryEndTime),
					CategoryType:      printerParam.CategoryType,
					ExcludeDataManage: excludeDataManage,
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
			QueryStartTime:    int64(printerParam.QueryStartTime),
			QueryEndTime:      int64(printerParam.QueryEndTime),
			CategoryType:      printerParam.CategoryType,
			ExcludeDataManage: excludeDataManage,
		})
		reqPrinterData.PaymentMethod = &business_data_resp.BusinessDataPaymentMethod{
			TotalReceivedPrice:   paymentData.TotalReceivedAmount,
			PaymentMethodIncomes: paymentMethodIncomes,
		}
	}

	if printerReq.StatisticsType == 2 {
		categoryData, categoryList := s.BuildCategoryList(ctx, req.BusinessDataCountReq{
			QueryStartTime:    int64(printerParam.QueryStartTime),
			QueryEndTime:      int64(printerParam.QueryEndTime),
			CategoryType:      printerParam.CategoryType,
			SortType:          printerParam.SortType,
			SortDirection:     printerParam.SortDirection,
			ExcludeDataManage: excludeDataManage,
		})

		paymentData, paymentMethodIncomes := s.BuildPaymentMethodIncome(ctx, req.BusinessDataCountReq{
			QueryStartTime:    int64(printerParam.QueryStartTime),
			QueryEndTime:      int64(printerParam.QueryEndTime),
			CategoryType:      printerParam.CategoryType,
			ExcludeDataManage: excludeDataManage,
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
			QueryStartTime:    int64(printerParam.QueryStartTime),
			QueryEndTime:      int64(printerParam.QueryEndTime),
			CategoryType:      printerParam.CategoryType,
			ExcludeDataManage: excludeDataManage,
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
		TimeType:          req.TimeType,
		QueryStartTime:    req.QueryStartTime,
		QueryEndTime:      req.QueryEndTime,
		CategoryType:      req.CategoryType,
		DutyNo:            req.DutyNo,
		ExcludeDataManage: req.ExcludeDataManage,
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
		TimeType:          req.TimeType,
		QueryStartTime:    req.QueryStartTime,
		QueryEndTime:      req.QueryEndTime,
		CategoryType:      req.CategoryType,
		DutyNo:            req.DutyNo,
		ExcludeDataManage: req.ExcludeDataManage,
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
		AllTakeawayOrderNum:        int(saleData.TotalInstantOrderTakeawayNum),
		AllTakeawayMinOrderPrice:   saleData.MinInstantOrderTakeawayAmount,
		AllTakeawayMaxOrderPrice:   saleData.MaxInstantOrderTakeawayAmount,
		AllTakeawayAvgOrderPrice:   saleData.AvgInstantOrderTakeawayAmount,
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
				DutyNo:         req.DutyNo,
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
				req.ExcludeDataManage,
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
				TimeType:          req.TimeType,
				QueryStartTime:    req.QueryStartTime,
				QueryEndTime:      req.QueryEndTime,
				CategoryType:      req.CategoryType,
				ExcludeDataManage: req.ExcludeDataManage,
				DutyNo:            req.DutyNo,
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
				TimeType:          req.TimeType,
				QueryStartTime:    req.QueryStartTime,
				QueryEndTime:      req.QueryEndTime,
				CategoryType:      req.CategoryType,
				DutyNo:            req.DutyNo,
				ExcludeDataManage: req.ExcludeDataManage,
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
	// 使用 GetParam 方法处理时间参数（包括日期时间字符串）
	settingSrv := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	businessSetting, err := settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)

	areaData := s.statisticsSrv.CountArea(ctx, CountReq{
		TimeType:          req.TimeType,
		QueryStartTime:    req.QueryStartTime,
		QueryEndTime:      req.QueryEndTime,
		CategoryType:      req.CategoryType,
		DutyNo:            req.DutyNo,
		ExcludeDataManage: req.ExcludeDataManage, // 传递过滤参数
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
	// 处理日期时间字符串参数（优先级：QueryStartTime/QueryEndTime > QueryStartDate/QueryEndDate）
	queryStartTime := req.QueryStartTime
	queryEndTime := req.QueryEndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && queryStartTime == 0 && queryEndTime == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		startTime, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			queryStartTime = startTime
		}
		endTime, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			queryEndTime = endTime
		}
	}

	var productRankData = business_data_resp.BusinessDataProductRank{
		Ranks: func() []business_data_resp.ProductRank {
			productRankList := s.statisticsSrv.RankProduct(ctx, CountReq{
				RankType:          req.RankType,
				QueryStartTime:    queryStartTime,
				QueryEndTime:      queryEndTime,
				ExcludeDataManage: req.ExcludeDataManage, // 传递过滤参数
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
	// 解析分类UUID字符串（如"uuid1,uuid2,,,,"）转换为数组
	var categoryUuids []uint64
	if req.CategoryUuids != "" {
		categoryUuidStrs := strings.Split(req.CategoryUuids, ",")
		for _, categoryUuidStr := range categoryUuidStrs {
			categoryUuidStr = strings.TrimSpace(categoryUuidStr)
			if categoryUuidStr == "" {
				continue
			}
			categoryUuid, err := strconv.ParseUint(categoryUuidStr, 10, 64)
			if err == nil && categoryUuid > 0 {
				categoryUuids = append(categoryUuids, categoryUuid)
			}
		}
	}
	// 处理向后兼容：如果传入单个 CategoryUuid，转换为 CategoryUuids 数组
	if len(categoryUuids) == 0 && req.CategoryUuid > 0 {
		categoryUuids = []uint64{req.CategoryUuid}
	}

	// 解析订单类型字符串（如"1,2,3"）转换为数组
	var orderTypes []uint
	if req.OrderType != "" {
		orderTypeStrs := strings.Split(req.OrderType, ",")
		for _, orderTypeStr := range orderTypeStrs {
			orderTypeStr = strings.TrimSpace(orderTypeStr)
			if orderTypeStr == "" {
				continue
			}
			orderType, err := strconv.ParseUint(orderTypeStr, 10, 32)
			if err == nil && orderType >= 1 && orderType <= 3 {
				orderTypes = append(orderTypes, uint(orderType))
			}
		}
	}

	// 处理日期时间字符串参数（优先级：TimeType > QueryStartTime/QueryEndTime > QueryStartDate/QueryEndDate）
	queryStartTime := req.QueryStartTime
	queryEndTime := req.QueryEndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && queryStartTime == 0 && queryEndTime == 0 && req.TimeType == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		startTime, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			queryStartTime = startTime
		}
		endTime, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			queryEndTime = endTime
		}
	}

	productSalesData := s.statisticsSrv.CountProductSale(ctx, CountReq{
		TimeType:          req.TimeType,
		QueryStartTime:    queryStartTime,
		QueryEndTime:      queryEndTime,
		RankType:          req.SortType,
		RankDirection:     req.SortDirection,
		PageNo:            req.PageNo,
		PageSize:          req.PageSize,
		AreaUuid:          req.AreaUuid,
		CategoryUuid:      req.CategoryUuid,
		CategoryUuids:     categoryUuids,
		ProductName:       req.ProductName,
		OrderTypes:        orderTypes,
		OrderSource:       req.OrderSource,
		ExcludeDataManage: req.ExcludeDataManage,
	})

	var list = []business_data_resp.BusinessDataCountProductSalesItem{}
	for _, productSale := range productSalesData.Data {
		list = append(list, business_data_resp.BusinessDataCountProductSalesItem{
			ProductName:        productSale.ProductName,
			SalesNum:           productSale.TotalSaleNum,
			SalesPrice:         productSale.TotalActualSaleAmount,
			CategoryName:       productSale.CategoryName,
			OriginalSalesPrice: productSale.TotalOriginSaleAmount,
			TotalPayPrice:      productSale.TotalBusinessAmount,
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

// ExportProductSales 导出商品销售统计
func (s *businessSrv) ExportProductSales(ctx context.Context, req req.BusinessDataCountProductSalesReq) error {
	// 检查数据量
	req.PageNo = 1
	req.PageSize = 1000
	result, err := s.CountProductSales(ctx, req)
	if err != nil {
		return err
	}
	if result.Meta.Total > 1000 {
		return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
	}
	if result.Meta.Total == 0 {
		return errors.WithMessage(errors.New("没有数据需要导出"))
	}

	// 创建导出任务
	params, err := json.Marshal(req)
	if err != nil {
		return err
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "Product Sales Statistics",
		ZhName:   "商品销售统计",
		ZhTwName: "商品銷售統計",
		ThName:   "สถิติการขายสินค้า",
		MyName:   "ကုန်ပစ္စည်းရောင်းအားစာရင်းဇယား",
		JaName:   "商品販売統計",
		KoName:   "상품 판매 통계",
		TrName:   "Ürün Satış İstatistikleri",
		SvName:   "Produktförsäljningsstatistik",
	}

	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeProductSales)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}

	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeProductSales,
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
		oldRecord, err := repository.NewExportRecordRepo(tx).GetUnfinishedExportRecord(model.ExportTypeProductSales)
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
		recordUuid := record.Uuid
		record, err := repository.NewExportRecordRepo(db).GetByUuid(recordUuid)
		if err != nil {
			logger.Logger.Error("获取导出ExportProductSales失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			if err := repository.NewExportRecordRepo(db).Update(recordUuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportProductSales失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			}
			return
		}
		if record == nil {
			logger.Logger.Error("导出记录不存在", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			if err := repository.NewExportRecordRepo(db).Update(recordUuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": "导出记录不存在",
			}); err != nil {
				logger.Logger.Error("导出ExportProductSales失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			}
			return
		}
		res, err := s.ExportProductSalesTask(ctx, req, record)
		if err != nil {
			logger.Logger.Error("导出ExportProductSalesTask失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportProductSalesTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			}
			return
		}
		if err := repository.NewExportRecordRepo(db).Update(record.Uuid, map[string]any{
			"status":    model.ExportStatusSuccess,
			"file_uuid": res.FileUuid,
		}); err != nil {
			logger.Logger.Error("导出ExportProductSalesTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			return
		}
	})

	return nil
}

// ExportProductSalesTask 导出商品销售统计任务处理
func (s *businessSrv) ExportProductSalesTask(ctx context.Context, req req.BusinessDataCountProductSalesReq, record *model.ExportRecord) (*resp.FileExportResp, error) {
	req.PageNo = 1
	req.PageSize = 1000
	result, err := s.CountProductSales(ctx, req)
	if err != nil {
		return nil, err
	}

	headerMap := map[string][]string{
		"zh":    {"商品名称", "商品分类", "销售数量", "原价销售额", "赠菜", "实际销售额", "营业收入"},
		"th":    {"ชื่อสินค้า", "หมวดหมู่สินค้า", "จำนวนการขาย", "ยอดขายราคาเดิม", "ของแถม", "ยอดขายจริง", "รายได้จากการดำเนินงาน"},
		"en":    {"Product Name", "Category", "Sales Quantity", "Original Sales Amount", "Gift", "Actual Sales Amount", "Business Revenue"},
		"zhtw":  {"商品名稱", "商品分類", "銷售數量", "原價銷售額", "贈菜", "實際銷售額", "營業收入"},
		"zh_tw": {"商品名稱", "商品分類", "銷售數量", "原價銷售額", "贈菜", "實際銷售額", "營業收入"},
		"ja":    {"商品名", "商品分類", "販売数量", "原価販売額", "おまけ", "実際販売額", "営業収入"},
		"ko":    {"상품명", "상품 분류", "판매 수량", "원가 판매액", "사은품", "실제 판매액", "영업 수입"},
		"my":    {"ကုန်ပစ္စည်းအမည်", "ကုန်ပစ္စည်းအမျိုးအစား", "ရောင်းအားအရေအတွက်", "မူလရောင်းအားငွေ", "လက်ဆောင်ပေးခြင်း", "အမှန်တကယ်ရောင်းအားငွေ", "လုပ်ငန်းဝင်ငွေ"},
		"tr":    {"Ürün Adı", "Kategori", "Satış Miktarı", "Orijinal Satış Tutarı", "Hediye", "Gerçek Satış Tutarı", "İşletme Geliri"},
		"sv":    {"Produktnamn", "Kategori", "Försäljningskvantitet", "Originalt försäljningsbelopp", "Gåva", "Faktiskt försäljningsbelopp", "Företagsintäkter"},
	}

	// 使用 record 中已生成的文件名
	fileName := record.ExportName
	xlsxFile := excelize.NewFile()
	sheetName := "Sheet1"
	index, err := xlsxFile.NewSheet(sheetName)
	if err != nil {
		logger.Logger.Error("创建Excel工作表失败", zap.Error(err))
		return nil, errors.WithMessage(err)
	}
	xlsxFile.SetActiveSheet(index)

	lang := ctx.GetLanguage()
	headers := func() []string {
		headers := headerMap[lang]
		if headers == nil {
			headers = headerMap["en"]
		}
		return headers
	}()

	// 写入表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheetName, cell, cell, style)
	}

	// 写入数据（无序号列，第一列从商品名称开始）
	for rowIdx, item := range result.List {
		offsetRow := rowIdx + 2
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", offsetRow), item.ProductName)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", offsetRow), item.CategoryName)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", offsetRow), item.SalesNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", offsetRow), item.OriginalSalesPrice)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", offsetRow), item.GiveProductNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("F%d", offsetRow), item.SalesPrice)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("G%d", offsetRow), item.TotalPayPrice)
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

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil
}

// Count7Days 统计7天
func (s *businessSrv) Count7Days(ctx context.Context, req req.BusinessDataCountReq) (*business_data_resp.BusinessDataCount7Days, error) {
	// 使用 GetParam 方法处理时间参数（包括日期时间字符串）
	settingSrv := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	businessSetting, err := settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)

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
		DutyNo:            req.DutyNo,
		QueryStartTime:    req.QueryStartTime,
		QueryEndTime:      req.QueryEndTime,
		TimeType:          req.TimeType,
		CategoryType:      req.CategoryType,
		ExcludeDataManage: req.ExcludeDataManage,
	})

	// 统计免单支付
	var freePaymentData CountFreePaymentResp
	if !req.NotQueryFree {
		freePaymentData = s.statisticsSrv.CountFreePayment(ctx, CountReq{
			DutyNo:            req.DutyNo,
			QueryStartTime:    req.QueryStartTime,
			QueryEndTime:      req.QueryEndTime,
			TimeType:          req.TimeType,
			CategoryType:      req.CategoryType,
			ExcludeDataManage: req.ExcludeDataManage,
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
		TimeType:          req.TimeType,
		QueryStartTime:    req.QueryStartTime,
		QueryEndTime:      req.QueryEndTime,
		CategoryType:      req.CategoryType,
		DutyNo:            req.DutyNo,
		ExcludeDataManage: req.ExcludeDataManage,
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
	// 使用 GetParam 方法处理时间参数（包括日期时间字符串）
	settingSrv := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	businessSetting, err := settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)

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
			TotalGrabOrderNum:          export.TotalGrabOrderNum,
			MinGrabOrderAmount:         export.MinGrabOrderAmount,
			MaxGrabOrderAmount:         export.MaxGrabOrderAmount,
			AvgGrabOrderAmount:         export.AvgGrabOrderAmount,
			AreaList:                   areaList,
			PaymentList:                paymentList,
			TotalRechargeAmount:        export.TotalRechargeAmount,
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
	// 使用 GetParam 方法处理时间参数（包括日期时间字符串）
	settingSrv := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	businessSetting, err := settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取营业设置失败", zap.Error(err))
	}
	req = req.GetParam(ctx.GetCompanySetting().Timezone, businessSetting.OpeningHours)

	// 销售数据
	saleData := s.statisticsSrv.CountSale(ctx, CountReq{
		TimeType:          req.TimeType,
		QueryStartTime:    req.QueryStartTime,
		QueryEndTime:      req.QueryEndTime,
		CategoryType:      req.CategoryType,
		DutyNo:            req.DutyNo,
		ExcludeDataManage: req.ExcludeDataManage,
	})

	// 会员数量
	memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
		TimeType:          req.TimeType,
		QueryStartTime:    req.QueryStartTime,
		QueryEndTime:      req.QueryEndTime,
		ExcludeDataManage: req.ExcludeDataManage,
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
				TimeType:          req.TimeType,
				QueryStartTime:    req.QueryStartTime,
				QueryEndTime:      req.QueryEndTime,
				CategoryType:      req.CategoryType,
				ExcludeDataManage: req.ExcludeDataManage,
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
	for index, product := range productList {
		efficiencyAnalysisDataList[product.Uuid] = &business_data_resp.KitchenEfficiencyAnalysisItem{
			ProductPackageUuid: product.Uuid,
			ProductName:        product.MultiLanguageName.GetNames(),
			CategoryName:       product.ProductCategory.MultiLanguageName.GetNames(),
		}
		efficiencyAnalysisDataList[product.Uuid].SetIndex(index)
		productPackageUuids = append(productPackageUuids, product.Uuid)
	}

	// 处理日期时间字符串参数（优先级：StartTime/EndTime > QueryStartDate/QueryEndDate）
	startTime := req.StartTime
	endTime := req.EndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && startTime == 0 && endTime == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		start, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			startTime = start
		}
		end, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			endTime = end
		}
	}

	efficiencyAnalysisResults, err := repository.NewKitchenEfficiencyAnalysisRepo(ctx.GetDB()).GetKitchenEfficiencyAnalysisByProductPackageUuid(productPackageUuids, startTime, endTime)
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

	// 排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].GetIndex() > list[j].GetIndex()
	})

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
	// 处理日期时间字符串参数（优先级：StartTime/EndTime > QueryStartDate/QueryEndDate）
	startTime := req.StartTime
	endTime := req.EndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && startTime == 0 && endTime == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		start, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			startTime = start
		}
		end, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			endTime = end
		}
	}

	avg, err := repository.NewKitchenEfficiencyAnalysisRepo(ctx.GetDB()).GetKitchenEfficiencyAnalysisAvg(startTime, endTime)
	if err != nil {
		return nil, err
	}
	return &business_data_resp.BusinessDataKitchenEfficiencyAnalysisAvg{
		Avg: avg,
	}, nil
}

func (s *businessSrv) KitchenProductionDetail(ctx context.Context, req req.KitchenProductionDetailReq) (*business_data_resp.KitchenProductionDetail, error) {
	// 处理日期时间字符串参数（优先级：StartTime/EndTime > QueryStartDate/QueryEndDate）
	startTime := req.StartTime
	endTime := req.EndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && startTime == 0 && endTime == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		start, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			startTime = start
		}
		end, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			endTime = end
		}
	}

	db := ctx.GetDB()
	productionRepo := repository.NewProductionRepo(db)
	// 根据过滤条件,获取商品bom_uuid列表
	productBomUuids, err := s.getKitchenProductionDetailProductBomUuids(ctx, req)
	if err != nil {
		return nil, err
	}

	// 如果没有商品bom_uuid且有查询条件时(关键词不为空、分类不为空),则返回空数据
	if len(productBomUuids) == 0 && (req.Keyword != "" || len(req.CategoryUuids) > 0) {
		return &business_data_resp.KitchenProductionDetail{
			List: make([]business_data_resp.KitchenProductionDetailItem, 0),
			Meta: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    0,
			},
		}, nil
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
	productionOrderProducts, err := productionRepo.GetProductionOrderList(req.PageNo, req.PageSize, productBomUuids, startTime, endTime, opts...)
	if err != nil {
		return nil, err
	}

	count, err := productionRepo.GetProductionOrderListCount(productBomUuids, startTime, endTime)
	if err != nil {
		return nil, err
	}

	productionDataList := make([]business_data_resp.KitchenProductionDetailItem, 0)
	for _, productionOrderProduct := range productionOrderProducts {
		if productionOrderProduct.SaleOrderProduct.IsPackageProduct() {
			continue // 跳过套餐商品
		}
		item := business_data_resp.KitchenProductionDetailItem{
			ProductName: func() dto.LocaleResponse {
				if productionOrderProduct.IsTakeoutOrder() {
					return *pkgLanguage.JsonToLocaleResponse(productionOrderProduct.Name)
				}
				return productionOrderProduct.SaleOrderProduct.MultiLanguageName.GetNames()
			}(),
			FlavorName: func() dto.LocaleResponse {
				if productionOrderProduct.IsTakeoutOrder() {
					return *pkgLanguage.JsonToLocaleResponse(productionOrderProduct.FlavorName)
				}
				return productionOrderProduct.SaleOrderProduct.GetFlavorName()
			}(),
			CategoryName: productionOrderProduct.ProductCategory.MultiLanguageName.GetNames(),
			Number:       productionOrderProduct.Num,
			CreateTime:   productionOrderProduct.CreateTime,
			MakeFinishTime: func() int64 {
				if productionOrderProduct.SendDuration == 0 { // 如果上菜时间大于0,则返回0,表示关闭了智能后厨
					return 0
				}
				return productionOrderProduct.MadeTime
			}(),
			MakeDuration: func() int64 {
				if productionOrderProduct.SendDuration == 0 { // 如果上菜时间大于0,则返回0,表示关闭了智能后厨
					return 0
				}
				return productionOrderProduct.MakeDuration
			}(),
			SendFinishTime: func() int64 {
				if productionOrderProduct.SendDuration > 0 { // 如果上菜时间大于0,则返回上菜完成时间
					return int64(productionOrderProduct.FinishedTime)
				}
				return 0
			}(),
			SendDuration: productionOrderProduct.SendDuration,
			FinishTime:   int64(productionOrderProduct.FinishedTime),
			AllDuration:  productionOrderProduct.AllDuration,
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
		productBomUuidsByName, err := repository.NewProductRepo(db).GetProductBomUuidsByKeyword(req.Keyword, ctx.GetLanguage())
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
	// 处理日期时间字符串参数（优先级：StartTime/EndTime > QueryStartDate/QueryEndDate）
	startTime := req.StartTime
	endTime := req.EndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && startTime == 0 && endTime == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		start, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			startTime = start
		}
		end, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			endTime = end
		}
	}

	db := ctx.GetDB()
	productionRepo := repository.NewProductionRepo(db)
	// 根据过滤条件,获取商品bom_uuid列表
	productBomUuids, err := s.getKitchenProductionDetailProductBomUuids(ctx, req)
	if err != nil {
		return 0, err
	}

	count, err := productionRepo.GetProductionOrderListCount(productBomUuids, startTime, endTime)
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
		return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "Details of Dish Presentation",
		ZhName:   "菜品出品详情",
		ZhTwName: "菜品出品詳情",
		ThName:   "รายละเอียดเมนูอาหาร",
		MyName:   "ဟင်းလျာအသေးစိတ်",
		JaName:   "料理の提供詳細",
		KoName:   "메뉴 서비스 정보",
		TrName:   "Yemek Sunumu Detayları",
		SvName:   "Maträttsdetaljer",
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeKitchenEfficiencyAnalysis)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
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
		recordUuid := record.Uuid
		record, err := repository.NewExportRecordRepo(db).GetByUuid(recordUuid)
		if err != nil {
			logger.Logger.Error("获取导出ExportKitchenProductionDetail失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			if err := repository.NewExportRecordRepo(db).Update(recordUuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
				return
			}
			return
		}
		if record == nil {
			logger.Logger.Error("导出记录不存在", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			if err := repository.NewExportRecordRepo(db).Update(recordUuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": "导出记录不存在",
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
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
		if res, err := s.ExportKitchenEfficiencyAnalysisTask(ctx, *params, record); err != nil {
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
		return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "Detailed List of Dish Output",
		ZhName:   "菜品出品明细",
		ZhTwName: "菜品出品明細",
		ThName:   "รายละเอียดการออกอาหาร",
		MyName:   "ဟင်းလျာထုတ်လုပ်မှု အသေးစိတ်စာရင်း",
		JaName:   "料理の提供明細",
		KoName:   "메뉴 출품 상세 내역",
		TrName:   "Yemek Çıkış Ayrıntıları",
		SvName:   "Detaljerad lista över maträttsutbud",
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeKitchenProductionDetail)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
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
		recordUuid := record.Uuid
		record, err := repository.NewExportRecordRepo(db).GetByUuid(recordUuid)
		if err != nil {
			logger.Logger.Error("获取导出ExportKitchenProductionDetail失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			if err := repository.NewExportRecordRepo(db).Update(recordUuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
				return
			}
			return
		}
		if record == nil {
			logger.Logger.Error("导出记录不存在", zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
			if err := repository.NewExportRecordRepo(db).Update(recordUuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": "导出记录不存在",
			}); err != nil {
				logger.Logger.Error("导出ExportKitchenProductionDetailTask失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", recordUuid))
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
		if res, err := s.ExportKitchenProductionDetailTask(ctx, *params, record); err != nil {
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

	// 查询所有套餐
	productPackageRepo := repository.NewProductPackageRepo(db)
	productPackageUuids, err := productPackageRepo.GetProductPackageUuidsByIsPackage()
	if err != nil {
		return "", err
	}
	packageUuidMap := make(map[uint64]bool) // 套餐uuid列表
	for _, productPackageUuid := range productPackageUuids {
		packageUuidMap[productPackageUuid] = true
	}

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

				IsPackage: func() int { // 是否是套餐
					if _, ok := packageUuidMap[item.ProductPackageUuid]; ok {
						return 1
					}
					return 0
				}(),
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
	for _, period := range businessTimePeriodData.BusinessTimePeriodList {
		businessTimePeriodList = append(businessTimePeriodList, business_data_resp.BusinessTimePeriodItem{
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

// CountBusinessSummary 统计综合运营
func (s *businessSrv) CountBusinessSummary(ctx context.Context, req req.StatisticsSummaryReq) business_data_resp.StatisticsSummary {
	// 调用统计服务
	businessSummaryData := s.statisticsSrv.CountBusinessSummary(ctx, req)

	// 构建返回列表
	businessSummaryList := make([]business_data_resp.StatisticsSummaryItem, 0, len(businessSummaryData.StatisticsComprehensiveList))
	for _, item := range businessSummaryData.StatisticsComprehensiveList {
		businessSummaryList = append(businessSummaryList, business_data_resp.StatisticsSummaryItem{
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

	return business_data_resp.StatisticsSummary{
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    businessSummaryData.TotalStatisticsComprehensiveNum,
		},
		List: businessSummaryList,
	}
}

// CountBusinessPaymentMethod 统计收款数据
func (s *businessSrv) CountBusinessPaymentMethod(ctx context.Context, req req.StatisticsPaymentMethodReq) business_data_resp.StatisticsPaymentMethod {
	// 调用统计服务
	businessPaymentMethodData := s.statisticsSrv.CountBusinessPaymentMethod(ctx, req)

	// 构建返回列表
	businessPaymentMethodList := make([]business_data_resp.StatisticsPaymentMethodItem, 0, len(businessPaymentMethodData.StatisticsPaymentMethodList))
	for _, item := range businessPaymentMethodData.StatisticsPaymentMethodList {
		businessPaymentMethodList = append(businessPaymentMethodList, business_data_resp.StatisticsPaymentMethodItem{
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
func (s *businessSrv) ExportKitchenProductionDetailTask(ctx context.Context, req req.KitchenProductionDetailReq, record *model.ExportRecord) (*resp.FileExportResp, error) { // 修改返回类型
	req.PageNo = 1
	req.PageSize = 1000 // 最多导出1000条数据
	productionDetail, err := s.KitchenProductionDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	// 商家时区
	timezone := ctx.GetCompanySetting().Timezone
	timezoneUtil := utils.SetTimezone(timezone)
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
	// 使用 record 中已生成的文件名
	fileName := record.ExportName
	xlsxFile := excelize.NewFile()
	sheetName := "Sheet1"
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
func (s *businessSrv) ExportKitchenEfficiencyAnalysisTask(ctx context.Context, req req.KitchenEfficiencyAnalysisReq, record *model.ExportRecord) (*resp.FileExportResp, error) { // 修改返回类型
	req.PageNo = 1
	req.PageSize = 1000 // 最多导出1000条数据
	result, err := s.CountKitchenEfficiencyAnalysis(ctx, req)
	if err != nil {
		return nil, err
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
	// 使用 record 中已生成的文件名
	fileName := record.ExportName
	xlsxFile := excelize.NewFile() // 修改这里，直接使用 NewFile()
	sheetName := "Sheet1"
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

// generateExportFileName 生成导出文件名
// 格式: 报表名YYYY-MM-DD.xlsx 或 报表名YYYY-MM-DD（序号）.xlsx
// 同一天多次导出同名报表时，自动添加序号避免冲突
func (s *businessSrv) generateExportFileName(
	ctx context.Context,
	reportName string, // 报表名称（多语言）
	exportType uint8, // 导出类型
) (string, error) {
	// 1. 获取商户时区
	timezone := ctx.GetCompanySetting().Timezone
	timezoneUtils := utils.SetTimezone(timezone)
	dateString := timezoneUtils.FormatUnixTime(time.Now().Unix(), "2006-01-02")

	// 2. 查询同一天已导出的同名报表（数据库连接已包含商户隔离）
	db := ctx.GetDB()
	exportRecordRepo := repository.NewExportRecordRepo(db)

	// 获取当天的开始和结束时间戳（商户时区）
	startTime, endTime := timezoneUtils.TodayStartEndUnix()

	records, err := exportRecordRepo.GetByDateAndType(
		exportType,
		startTime,
		endTime,
	)
	if err != nil {
		return "", errors.WithMessage(err, "查询导出记录失败")
	}

	// 3. 计算序号
	suffix := ""
	if len(records) > 0 {
		suffix = fmt.Sprintf("（%d）", len(records))
	}

	// 4. 生成文件名（包含 .xlsx 后缀）
	return fmt.Sprintf("%s%s%s.xlsx", reportName, dateString, suffix), nil
}

// 导出时段营业统计数据
func (s *businessSrv) ExportBusinessTimePeriod(ctx context.Context, request req.BusinessTimePeriodReq) error {
	db := ctx.GetDB()
	// 判断是否还有正在导出的任务
	oldRecord, err := repository.NewExportRecordRepo(db).GetUnfinishedExportRecord(model.ExportTypeBusinessData)
	if err != nil {
		return err
	}
	if oldRecord != nil {
		return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
	}

	result := s.CountBusinessTimePeriod(ctx, request)
	if result.Meta.Total == 0 {
		return errors.WithMessage(errors.New("没有数据需要导出"))
	}
	if result.Meta.Total > 1000 {
		return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "Business Time Period Statistics",
		ZhName:   "时段营业统计",
		ZhTwName: "時段營業統計",
		ThName:   "สถิติยอดขายตามช่วงเวลา",
		MyName:   "Business Time Period Statistics",
		JaName:   "時間帯別売上統計",
		KoName:   "영업시간대별 통계",
		TrName:   "Zaman Dilimi Satış İstatistikleri",
		SvName:   "Försäljningsstatistik per tidsperiod",
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeBusinessData)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeBusinessData,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(params),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	err = repository.NewExportRecordRepo(db).Create(record)
	if err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		_, err := s.ExportBusinessTimePeriodTask(ctx, ExportBusinessTimePeriodTaskParams{
			Record:      *record,
			FillNameMul: fileNameMul,
			Request:     request,
		})
		if err != nil {
			if err := repository.NewExportRecordRepo(ctx.GetDB()).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出时段营业统计数据失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			}
			return
		}
	})

	return nil
}

// 导出时段营业统计数据参数
type ExportBusinessTimePeriodTaskParams struct {
	Request     req.BusinessTimePeriodReq // 请求参数
	Record      model.ExportRecord        // 导出记录
	FillNameMul model.MultiLanguageName   // 多语言名称
}

// 导出时段营业统计数据
func (s *businessSrv) ExportBusinessTimePeriodTask(ctx context.Context, params ExportBusinessTimePeriodTaskParams) (*resp.FileExportResp, error) {
	db := ctx.GetDB()

	params.Request.PageNo = 1
	params.Request.PageSize = 1000
	result := s.CountBusinessTimePeriod(ctx, params.Request)

	// 根据语言获取表头
	headerMap := map[string][]string{
		"zh": { // 中文
			"时段", "订单金额", "实付金额", "订单量", "用餐人数", "订单金额人均", "实付金额人均",
		},
		"en": { // 英文
			"Time Period", "Order Amount", "Paid Amount", "Order Count", "Number of Diners", "Order Amount Per Person", "Paid Amount Per Person",
		},
		"th": { // 泰语
			"ช่วงเวลา", "ยอดคำสั่งซื้อ", "ยอดชำระเงิน", "จำนวนออเดอร์", "จำนวนลูกค้า", "ยอดคำสั่งซื้อต่อคน", "ยอดชำระเงินต่อคน",
		},
		"zhtw": { // 繁体中文
			"時段", "訂單金額", "實付金額", "訂單量", "用餐人數", "訂單金額人均", "實付金額人均",
		},
		"ja": { // 日语
			"時間帯", "注文金額", "支払い金額", "注文数", "来店人数", "一人当たり注文金額", "一人当たり支払い金額",
		},
		"ko": { // 韩语
			"시간대", "주문 금액", "실제 결제 금액", "주문 건수", "식사 인원", "인당 주문 금액", "인당 실제 결제 금액",
		},
		"my": { // 缅甸语
			"အချိန်အပိုင်းအခြား", "အော်ဒါအလုံးစုံပမာဏ", "ပေးစာငွေ", "အော်ဒါအရေအတွက်", "စားသုံးသူအရေအတွက်", "တစ်ဦးလျှင်အော်ဒါပမာဏ", "တစ်ဦးလျှင်ပေးစာငွေ",
		},
		"tr": { // 土耳其语
			"Zaman Dilimi", "Sipariş Tutarı", "Ödenen Tutar", "Sipariş Sayısı", "Yemek Yiyen Kişi Sayısı", "Kişi Başı Sipariş Tutarı", "Kişi Başı Ödenen Tutar",
		},
		"sv": { // 瑞典语
			"Tidsperiod", "Orderbelopp", "Betalt Belopp", "Antal Order", "Antal Gäster", "Orderbelopp per Person", "Betalt Belopp per Person",
		},
	}
	xlsxFile := excelize.NewFile() // 修改这里，直接使用 NewFile()
	sheetName := "Sheet1"
	// 创建一个新的工作表
	index, err := xlsxFile.NewSheet(sheetName)
	if err != nil {
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
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", offsetRow), item.TimePeriod)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", offsetRow), item.OrderAmount)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", offsetRow), item.PayAmount)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", offsetRow), item.OrderNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", offsetRow), item.MealNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("F%d", offsetRow), item.OrderAmountMealAvg)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("G%d", offsetRow), item.PayAmountMealAvg)
	}

	// 自动调整列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheetName, colName, colName, 20)
	}

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, params.Record.ExportName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.NewExportRecordRepo(db).Update(params.Record.Uuid, map[string]any{
		"file_uuid": res.Uuid,
		"status":    model.ExportStatusSuccess,
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil
}

// 导出综合运营统计数据
func (s *businessSrv) ExportBusinessSummary(ctx context.Context, request req.StatisticsSummaryReq) error {
	db := ctx.GetDB()
	// 判断是否还有正在导出的任务
	oldRecord, err := repository.NewExportRecordRepo(db).GetUnfinishedExportRecord(model.ExportTypeBusinessDataSummary)
	if err != nil {
		return err
	}
	if oldRecord != nil {
		return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
	}

	result := s.CountBusinessSummary(ctx, request)
	if result.Meta.Total == 0 {
		return errors.WithMessage(errors.New("没有数据需要导出"))
	}
	if result.Meta.Total > 1000 {
		return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "Business Comprehensive Operations Statistics",
		ZhName:   "综合运营统计",
		ZhTwName: "綜合營運統計",
		ThName:   "สถิติการดำเนินงานแบบองค์รวม",
		MyName:   "စုပေါင်းလုပ်ငန်းဆောင်ရွက်မှုစာရင်း",
		JaName:   "総合運営統計",
		KoName:   "종합 운영 통계",
		TrName:   "Kapsamlı İşletme İstatistikleri",
		SvName:   "Statistik för omfattande verksamhet",
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeBusinessDataSummary)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeBusinessDataSummary,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(params),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	err = repository.NewExportRecordRepo(db).Create(record)
	if err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		_, err := s.ExportBusinessSummaryTask(ctx, ExportBusinessSummaryTaskParams{
			Record:      *record,
			FillNameMul: fileNameMul,
			Request:     request,
		})
		if err != nil {
			if err := repository.NewExportRecordRepo(ctx.GetDB()).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出综合运营统计数据失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			}
			return
		}
	})

	return nil
}

// 导出综合运营统计数据参数
type ExportBusinessSummaryTaskParams struct {
	Request     req.StatisticsSummaryReq // 请求参数
	Record      model.ExportRecord       // 导出记录
	FillNameMul model.MultiLanguageName  // 多语言名称
}

// 导出综合运营统计数据
func (s *businessSrv) ExportBusinessSummaryTask(ctx context.Context, params ExportBusinessSummaryTaskParams) (*resp.FileExportResp, error) {
	db := ctx.GetDB()

	params.Request.PageNo = 1
	params.Request.PageSize = 1000
	result := s.CountBusinessSummary(ctx, params.Request)

	// 根据语言获取表头
	headerMap := map[string][]string{
		"zh": { // 中文
			"营业日", "订单金额", "实付金额", "订单量", "用餐人数", "消费桌数", "订单金额人均", "实付金额人均", "订单金额单均", "实付金额单均", "点餐订单金额", "桌台订单金额", "外送订单金额",
		},
		"en": { // 英文 (根据中文翻译)
			"Business Day", "Order Amount", "Paid Amount", "Order Count", "Number of Diners", "Number of Tables Consumed", "Order Amount Per Person", "Paid Amount Per Person", "Order Amount Per Order", "Paid Amount Per Order", "Meal Order Amount", "Table Order Amount", "Takeout Order Amount",
		},
		"th": { // 泰语 (根据中文翻译)
			"วันดำเนินธุรกิจ", "ยอดคำสั่งซื้อ", "ยอดชำระเงิน", "จำนวนออเดอร์", "จำนวนลูกค้า", "จำนวนโต๊ะที่ใช้", "ยอดคำสั่งซื้อต่อคน", "ยอดชำระเงินต่อคน", "ยอดคำสั่งซื้อต่อบิล", "ยอดชำระเงินต่อบิล", "ยอดคำสั่งซื้อรับประทานอาหาร", "ยอดคำสั่งซื้อโต๊ะ", "ยอดคำสั่งซื้อนำกลับบ้าน",
		},
		"zhtw": { // 繁体中文 (根据中文翻译)
			"營業日", "訂單金額", "實付金額", "訂單量", "用餐人數", "消費桌數", "訂單金額人均", "實付金額人均", "訂單金額單均", "實付金額單均", "點餐訂單金額", "桌台訂單金額", "外送訂單金額",
		},
		"ja": { // 日语 (根据中文翻译)
			"営業日", "注文金額", "支払い金額", "注文数", "来店人数", "利用テーブル数", "一人当たり注文金額", "一人当たり支払い金額", "一件当たり注文金額", "一件当たり支払い金額", "食事注文金額", "テーブル注文金額", "テイクアウト注文金額",
		},
		"ko": { // 韩语 (根据中文翻译)
			"영업일", "주문 금액", "실제 결제 금액", "주문 건수", "식사 인원", "소비 테이블 수", "인당 주문 금액", "인당 실제 결제 금액", "주문당 주문 금액", "주문당 실제 결제 금액", "식사 주문 금액", "테이블 주문 금액", "포장 주문 금액",
		},
		"my": { // 缅甸语 (根据中文翻译)
			"လုပ်ငန်းသက်တမ်းနေ့", "အော်ဒါစုစုပေါင်းပမာဏ", "ပေးသွင်းငွေ", "အော်ဒါအရေအတွက်", "စားသုံးသူအရေအတွက်", "စားသုံးထားသောစားပွဲအရေအတွက်", "တစ်ဦးလျှင်အော်ဒါပမာဏ", "တစ်ဦးလျှင်ပေးသွင်းငွေ", "တစ်ဦးလျှင်အော်ဒါပမာဏ", "တစ်ဦးလျှင်ပေးသွင်းငွေ", "အစားအသောက်အော်ဒါပမာဏ", "စားပွဲအော်ဒါပမာဏ", "ယူဆောင်အော်ဒါပမာဏ",
		},
		"tr": { // 土耳其语 (根据中文翻译)
			"İşletme Günü", "Sipariş Tutarı", "Ödenen Tutar", "Sipariş Sayısı", "Yemek Yiyen Kişi Sayısı", "Tüketilen Masa Sayısı", "Kişi Başına Sipariş Tutarı", "Kişi Başına Ödenen Tutar", "Sipariş Başına Sipariş Tutarı", "Sipariş Başına Ödenen Tutar", "Yemek Sipariş Tutarı", "Masa Sipariş Tutarı", "Paket Sipariş Tutarı",
		},
		"sv": { // 瑞典语 (根据中文翻译)
			"Affärsdag", "Orderbelopp", "Betalt Belopp", "Antal Order", "Antal Gäster", "Antal Konsumerade Bord", "Orderbelopp per Person", "Betalt Belopp per Person", "Orderbelopp per Order", "Betalt Belopp per Order", "Matbeställningsbelopp", "Bordsorderbelopp", "Takeaway Orderbelopp",
		},
	}
	xlsxFile := excelize.NewFile() // 修改这里，直接使用 NewFile()
	sheetName := "Sheet1"
	// 创建一个新的工作表
	index, err := xlsxFile.NewSheet(sheetName)
	if err != nil {
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
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", offsetRow), item.OrderAmount)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", offsetRow), item.PayAmount)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", offsetRow), item.OrderNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", offsetRow), item.MealNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("F%d", offsetRow), item.DeskNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("G%d", offsetRow), item.OrderAmountMealAvg)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("H%d", offsetRow), item.PayAmountMealAvg)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("I%d", offsetRow), item.OrderAmountAvg)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("J%d", offsetRow), item.PayAmountAvg)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("K%d", offsetRow), item.InstantOrderAmount)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("L%d", offsetRow), item.DeskOrderAmount)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("M%d", offsetRow), item.TakeoutOrderAmount)
	}

	// 自动调整列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheetName, colName, colName, 20)
	}

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, params.Record.ExportName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.NewExportRecordRepo(db).Update(params.Record.Uuid, map[string]any{
		"file_uuid": res.Uuid,
		"status":    model.ExportStatusSuccess,
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil
}

// 导出收款数据
func (s *businessSrv) ExportBusinessPaymentMethod(ctx context.Context, request req.StatisticsPaymentMethodReq) error {
	db := ctx.GetDB()
	// 判断是否还有正在导出的任务
	oldRecord, err := repository.NewExportRecordRepo(db).GetUnfinishedExportRecord(model.ExportTypeBusinessDataPaymentMethod)
	if err != nil {
		return err
	}
	if oldRecord != nil {
		return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
	}

	result := s.CountBusinessPaymentMethod(ctx, request)
	if result.Meta.Total == 0 {
		return errors.WithMessage(errors.New("没有数据需要导出"))
	}
	if result.Meta.Total > 1000 {
		return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "Business Payment Statistics",
		ZhName:   "营业收款统计",
		ZhTwName: "營業收款統計",
		ThName:   "สถิติการชำระเงินของธุรกิจ",
		MyName:   "လုပ်ငန်းငွေလက်ခံမှုစာရင်း",
		JaName:   "営業入金統計",
		KoName:   "영업 수금 통계",
		TrName:   "İşletme Tahsilat İstatistikleri",
		SvName:   "Företagets betalningsstatistik",
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeBusinessDataPaymentMethod)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeBusinessDataPaymentMethod,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(params),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	err = repository.NewExportRecordRepo(db).Create(record)
	if err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		_, err := s.ExportBusinessPaymentMethodTask(ctx, ExportBusinessPaymentMethodTaskParams{
			Record:      *record,
			FillNameMul: fileNameMul,
			Request:     request,
		})
		if err != nil {
			if err := repository.NewExportRecordRepo(ctx.GetDB()).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出营业收款统计数据失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			}
			return
		}
	})

	return nil
}

// 导出收款数据参数
type ExportBusinessPaymentMethodTaskParams struct {
	Request     req.StatisticsPaymentMethodReq // 请求参数
	Record      model.ExportRecord             // 导出记录
	FillNameMul model.MultiLanguageName        // 多语言名称
}

// 导出营业收款统计数据
func (s *businessSrv) ExportBusinessPaymentMethodTask(ctx context.Context, params ExportBusinessPaymentMethodTaskParams) (*resp.FileExportResp, error) {
	db := ctx.GetDB()

	params.Request.PageNo = 1
	params.Request.PageSize = 1000
	result := s.CountBusinessPaymentMethod(ctx, params.Request)

	// 根据语言获取表头
	headerMap := map[string][]string{
		"zh": { // 中文
			"营业日", "支付方式", "次数", "收款小记",
		},
		"en": { // 英文（根据中文翻译）
			"Business Day", "Payment Method", "Count", "Summary of Receipts",
		},
		"th": { // 泰语（根据中文翻译）
			"วันดำเนินธุรกิจ", "วิธีการชำระเงิน", "จำนวนครั้ง", "สรุปการรับเงิน",
		},
		"zhtw": { // 繁體中文（根据中文翻译）
			"營業日", "支付方式", "次數", "收款小記",
		},
		"ja": { // 日语（根据中文翻译）
			"営業日", "支払方法", "回数", "入金概要",
		},
		"ko": { // 韩语（根据中文翻译）
			"영업일", "결제 방식", "횟수", "수납 요약",
		},
		"my": { // 缅甸语（根据中文翻译）
			"လုပ်ငန်းသက်တမ်းနေ့", "ငွေပေးချေမှုနည်းလမ်း", "အကြိမ်အရေအတွက်", "ငွေလက်ခံမှုအကျဉ်းချုပ်",
		},
		"tr": { // 土耳其语（根据中文翻译）
			"İşletme Günü", "Ödeme Yöntemi", "Adet", "Tahsilat Özeti",
		},
		"sv": { // 瑞典语（根据中文翻译）
			"Affärsdag", "Betalningsmetod", "Antal", "Sammanfattning av mottagna betalningar",
		},
	}
	xlsxFile := excelize.NewFile() // 修改这里，直接使用 NewFile()
	sheetNameMul := model.MultiLanguageName{
		EnName:   "Report",     // 英文
		ZhName:   "报表",         // 中文
		ZhTwName: "報表",         // 繁体中文
		ThName:   "รายงาน",     // 泰语
		MyName:   "အစီရင်ခံစာ", // 缅甸语
		JaName:   "レポート",       // 日语
		KoName:   "보고서",        // 韩语
		TrName:   "Rapor",      // 土耳其语
		SvName:   "Rapport",    // 瑞典语
	}
	sheetName := sheetNameMul.GetNameByLang(ctx.GetLanguage())
	// 创建一个新的工作表
	index, err := xlsxFile.NewSheet(sheetName)
	if err != nil {
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
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", offsetRow), item.PaymentName)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", offsetRow), item.PaymentNum)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", offsetRow), item.PaymentAmount)
	}

	// 自动调整列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheetName, colName, colName, 20)
	}

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, params.Record.ExportName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.NewExportRecordRepo(db).Update(params.Record.Uuid, map[string]any{
		"file_uuid": res.Uuid,
		"status":    model.ExportStatusSuccess,
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil
}

// CountChannelSales 统计渠道营业数据
func (s *businessSrv) CountChannelSales(ctx context.Context, req req.ChannelSalesReq) (*resp.ChannelSalesResp, error) {
	db := ctx.GetDB()
	statisticsRepo := repository.NewStatisticsRepo(db)

	// 处理日期时间字符串参数（优先级：StartTime/EndTime > QueryStartDate/QueryEndDate）
	startTime := req.StartTime
	endTime := req.EndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && startTime == 0 && endTime == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		start, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			startTime = start
		}
		end, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			endTime = end
		}
	}

	// 处理默认时间：如果未传时间，使用今日范围
	if startTime == 0 || endTime == 0 {
		timezone := ctx.GetCompanySetting().Timezone
		timezoneUtil := utils.SetTimezone(timezone)
		startTime, endTime = timezoneUtil.TodayStartEndUnix()
	}

	// 参数校验
	if startTime > endTime {
		return nil, errors.WithMessage(errors.New("开始时间不能大于结束时间"))
	}

	// 调用 Repository 获取渠道统计数据
	channelData, err := statisticsRepo.CountChannelSale(startTime, endTime, req.ExcludeDataManage)
	if err != nil {
		return nil, errors.WithMessage(err, "统计渠道营业数据失败")
	}

	// 转换为响应格式
	convertToBlock := func(data *model.ChannelSaleRepoResult) *resp.ChannelSalesBlock {
		if data == nil {
			return &resp.ChannelSalesBlock{
				TotalOrderNum:      0,
				MinOrderAmount:     0,
				MaxOrderAmount:     0,
				AvgOrderAmount:     0,
				TotalDeskNum:       0,
				TotalMealNum:       0,
				OrderAmountMealAvg: 0,
			}
		}
		block := &resp.ChannelSalesBlock{
			TotalOrderNum: data.TotalOrderNum.Int64,
		}
		// 转换 NullFloat64 为 float64，无效时返回 0
		if data.MinOrderAmount.Valid {
			block.MinOrderAmount = data.MinOrderAmount.Float64
		}
		if data.MaxOrderAmount.Valid {
			block.MaxOrderAmount = data.MaxOrderAmount.Float64
		}
		if data.AvgOrderAmount.Valid {
			block.AvgOrderAmount = data.AvgOrderAmount.Float64
		}
		// 转换 NullInt64 为 int64，无效时返回 0
		if data.TotalDeskNum.Valid {
			block.TotalDeskNum = data.TotalDeskNum.Int64
		}
		if data.TotalMealNum.Valid {
			block.TotalMealNum = data.TotalMealNum.Int64
		}
		// 转换人均订单金额，使用 decimal 确保精度，保留两位小数
		if data.OrderAmountMealAvg.Valid {
			orderAmountMealAvgDec := decimal.NewFromFloat(data.OrderAmountMealAvg.Float64)
			block.OrderAmountMealAvg = orderAmountMealAvgDec.Round(2).InexactFloat64()
		}
		return block
	}

	return &resp.ChannelSalesResp{
		Summary:         convertToBlock(channelData["summary"]),
		Table:           convertToBlock(channelData["table"]),
		DineIn:          convertToBlock(channelData["dine_in"]),
		TakeoutShop:     convertToBlock(channelData["takeout_shop"]),
		TakeoutDelivery: convertToBlock(channelData["takeout_delivery"]),
		DineInStore:     convertToBlock(channelData["dine_in_store"]),
		Takeaway:        convertToBlock(channelData["takeaway"]),
		Grab:            convertToBlock(channelData["grab"]),
		Lineman:         convertToBlock(channelData["lineman"]),
		Takeout:         convertToBlock(channelData["takeout"]),
		Meta: &resp.ChannelSalesMeta{
			StartTime: startTime,
			EndTime:   endTime,
		},
	}, nil
}

// ExportChannelSales 导出渠道营业统计数据
func (s *businessSrv) ExportChannelSales(ctx context.Context, req req.ChannelSalesReq) error {
	db := ctx.GetDB()
	// 判断是否还有正在导出的任务
	oldRecord, err := repository.NewExportRecordRepo(db).GetUnfinishedExportRecord(model.ExportTypeChannelSales)
	if err != nil {
		return err
	}
	if oldRecord != nil {
		return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
	}

	// 获取统计数据
	result, err := s.CountChannelSales(ctx, req)
	if err != nil {
		return err
	}

	// 检查是否有数据
	hasData := result.Summary != nil && result.Summary.TotalOrderNum > 0
	if !hasData {
		return errors.WithMessage(errors.New("没有数据需要导出"))
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "Channel Sales Statistics",
		ZhName:   "渠道营业统计",
		ZhTwName: "渠道營業統計",
		ThName:   "สถิติการขายช่องทาง",
		MyName:   "လမ်းကြောင်းရောင်းအားစာရင်း",
		JaName:   "チャネル売上統計",
		KoName:   "채널 매출 통계",
		TrName:   "Kanal Satış İstatistikleri",
		SvName:   "Kanalförsäljningsstatistik",
	}

	// 创建导出任务
	params, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeChannelSales)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeChannelSales,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(params),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	err = repository.NewExportRecordRepo(db).Create(record)
	if err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		_, err := s.ExportChannelSalesTask(ctx, ExportChannelSalesTaskParams{
			Record:      *record,
			FillNameMul: fileNameMul,
			Request:     req,
			Result:      result,
		})
		if err != nil {
			if err := repository.NewExportRecordRepo(ctx.GetDB()).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出渠道营业统计数据失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			}
			return
		}
	})

	return nil
}

// ExportChannelSalesTaskParams 导出渠道营业统计数据参数
type ExportChannelSalesTaskParams struct {
	Request     req.ChannelSalesReq     // 请求参数
	Record      model.ExportRecord      // 导出记录
	FillNameMul model.MultiLanguageName // 多语言名称
	Result      *resp.ChannelSalesResp  // 统计数据
}

// ExportChannelSalesTask 导出渠道营业统计数据任务
func (s *businessSrv) ExportChannelSalesTask(ctx context.Context, params ExportChannelSalesTaskParams) (*resp.FileExportResp, error) {
	db := ctx.GetDB()
	lang := ctx.GetLanguage()

	// 根据语言获取渠道名称
	channelNameMap := map[string]map[string]string{
		"zh": {
			"summary":          "合计",
			"table":            "桌台",
			"dine_in":          "点餐-店内",
			"takeout_shop":     "点餐-外卖",
			"takeout_delivery": "外送",
			"dine_in_store":    "堂食",
			"takeaway":         "外带",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "外卖",
		},
		"en": {
			"summary":          "Total",
			"table":            "Table",
			"dine_in":          "Dine-in Order",
			"takeout_shop":     "Takeaway Order",
			"takeout_delivery": "Delivery",
			"dine_in_store":    "Dine-in",
			"takeaway":         "Takeaway",
			"grab":             "Grab",
			"lineman":          "LINE MAN",
			"takeout":          "Takeout",
		},
		"th": {
			"summary":          "รวม",
			"table":            "โต๊ะ",
			"dine_in":          "สั่งอาหารในร้าน",
			"takeout_shop":     "สั่งอาหารกลับบ้าน",
			"takeout_delivery": "จัดส่ง",
			"dine_in_store":    "สั่งอาหารในร้าน",
			"takeaway":         "นำอาหารกลับบ้าน",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "สั่งอาหารกลับบ้าน",
		},
		"zhtw": {
			"summary":          "合計",
			"table":            "桌台",
			"dine_in":          "點餐-店內",
			"takeout_shop":     "點餐-外賣",
			"takeout_delivery": "外送",
			"dine_in_store":    "店內點餐",
			"takeaway":         "外帶",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "外送",
		},
		"ja": {
			"summary":          "合計",
			"table":            "テーブル",
			"dine_in":          "店内注文",
			"takeout_shop":     "テイクアウト注文",
			"takeout_delivery": "配達",
			"dine_in_store":    "店内点餐",
			"takeaway":         "外帶",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "外送",
		},
		"ko": {
			"summary":          "합계",
			"table":            "테이블",
			"dine_in":          "매장 주문",
			"takeout_shop":     "테이크아웃 주문",
			"takeout_delivery": "배달",
			"dine_in_store":    "店内点餐",
			"takeaway":         "外帶",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "外送",
		},
		"my": {
			"summary":          "စုစုပေါင်း",
			"table":            "စားပွဲ",
			"dine_in":          "ဆိုင်တွင်မှာယူ",
			"takeout_shop":     "အိမ်သို့ယူ",
			"takeout_delivery": "ပို့ဆောင်မှု",
			"dine_in_store":    "ဆိုင်တွင်မှာယူ",
			"takeaway":         "အိမ်သို့ယူ",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "အိမ်သို့ယူ",
		},
		"tr": {
			"summary":          "Toplam",
			"table":            "Masa",
			"dine_in":          "Restoranda Sipariş",
			"takeout_shop":     "Paket Sipariş",
			"takeout_delivery": "Teslimat",
			"dine_in_store":    "Restoranda Sipariş",
			"takeaway":         "Paket Sipariş",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "Teslimat",
		},
		"sv": {
			"summary":          "Totalt",
			"table":            "Bord",
			"dine_in":          "Beställning på restaurang",
			"takeout_shop":     "Takeaway-beställning",
			"takeout_delivery": "Leverans",
			"dine_in_store":    "Restoranda Sipariş",
			"takeaway":         "Paket Sipariş",
			"grab":             "Grab",
			// "lineman":          "LINE MAN",
			"takeout": "Teslimat",
		},
	}
	channelNames := channelNameMap[lang]
	if channelNames == nil {
		channelNames = channelNameMap["en"]
	}

	xlsxFile := excelize.NewFile()

	// 根据语言获取指标名称
	labelMap := map[string]map[string]string{
		"zh": {
			"order_count":    "所有订单数",
			"min_amount":     "最小订单金额",
			"max_amount":     "最大订单金额",
			"avg_amount":     "平均订单金额",
			"table_count":    "桌数",
			"guest_count":    "人数",
			"order_meal_avg": "人均",
		},
		"en": {
			"order_count":    "Total Orders",
			"min_amount":     "Min Order Amount",
			"max_amount":     "Max Order Amount",
			"avg_amount":     "Avg Order Amount",
			"table_count":    "Table Count",
			"guest_count":    "Guest Count",
			"order_meal_avg": "Per Person",
		},
		"th": {
			"order_count":    "จำนวนคำสั่งซื้อทั้งหมด",
			"min_amount":     "จำนวนเงินคำสั่งซื้อขั้นต่ำ",
			"max_amount":     "จำนวนเงินคำสั่งซื้อสูงสุด",
			"avg_amount":     "จำนวนเงินคำสั่งซื้อเฉลี่ย",
			"table_count":    "จำนวนโต๊ะ",
			"guest_count":    "จำนวนคน",
			"order_meal_avg": "ต่อคน",
		},
		"zhtw": {
			"order_count":    "所有訂單數",
			"min_amount":     "最小訂單金額",
			"max_amount":     "最大訂單金額",
			"avg_amount":     "平均訂單金額",
			"table_count":    "桌數",
			"guest_count":    "人數",
			"order_meal_avg": "人均",
		},
		"ja": {
			"order_count":    "全注文数",
			"min_amount":     "最小注文金額",
			"max_amount":     "最大注文金額",
			"avg_amount":     "平均注文金額",
			"table_count":    "テーブル数",
			"guest_count":    "人数",
			"order_meal_avg": "一人あたり",
		},
		"ko": {
			"order_count":    "전체 주문 수",
			"min_amount":     "최소 주문 금액",
			"max_amount":     "최대 주문 금액",
			"avg_amount":     "평균 주문 금액",
			"table_count":    "테이블 수",
			"guest_count":    "인원 수",
			"order_meal_avg": "인당",
		},
		"my": {
			"order_count":    "အော်ဒါအရေအတွက်စုစုပေါင်း",
			"min_amount":     "အော်ဒါငွေပမာဏအနည်းဆုံး",
			"max_amount":     "အော်ဒါငွေပမာဏအများဆုံး",
			"avg_amount":     "အော်ဒါငွေပမာဏပျမ်းမျှ",
			"table_count":    "စားပွဲအရေအတွက်",
			"guest_count":    "လူအရေအတွက်",
			"order_meal_avg": "တစ်ဦးလျှင်",
		},
		"tr": {
			"order_count":    "Toplam Sipariş",
			"min_amount":     "Min Sipariş Tutarı",
			"max_amount":     "Max Sipariş Tutarı",
			"avg_amount":     "Ort Sipariş Tutarı",
			"table_count":    "Masa Sayısı",
			"guest_count":    "Kişi Sayısı",
			"order_meal_avg": "Kişi başı",
		},
		"sv": {
			"order_count":    "Totalt antal beställningar",
			"min_amount":     "Min beställningsbelopp",
			"max_amount":     "Max beställningsbelopp",
			"avg_amount":     "Genomsnittligt beställningsbelopp",
			"table_count":    "Antal bord",
			"guest_count":    "Antal gäster",
			"order_meal_avg": "Per person",
		},
	}
	labels := labelMap[lang]
	if labels == nil {
		labels = labelMap["en"]
	}

	// 辅助函数：写入渠道数据
	writeChannelData := func(sheetName string, rowIdx *int, channelName string, block *resp.ChannelSalesBlock, includeTableInfo bool) (int, int) {
		startRow := *rowIdx
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", *rowIdx), channelName)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", *rowIdx), labels["order_count"])
		if block != nil {
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", *rowIdx), block.TotalOrderNum)
		}
		*rowIdx++

		if includeTableInfo && block != nil && block.TotalDeskNum > 0 {
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", *rowIdx), labels["table_count"])
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", *rowIdx), block.TotalDeskNum)
			*rowIdx++

			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", *rowIdx), labels["guest_count"])
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", *rowIdx), block.TotalMealNum)
			*rowIdx++

			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", *rowIdx), labels["order_meal_avg"])
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", *rowIdx), block.OrderAmountMealAvg)
			*rowIdx++
		}

		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", *rowIdx), labels["min_amount"])
		if block != nil {
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", *rowIdx), block.MinOrderAmount)
		}
		*rowIdx++

		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", *rowIdx), labels["max_amount"])
		if block != nil {
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", *rowIdx), block.MaxOrderAmount)
		}
		*rowIdx++

		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", *rowIdx), labels["avg_amount"])
		if block != nil {
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", *rowIdx), block.AvgOrderAmount)
		}
		endRow := *rowIdx
		*rowIdx++

		// 合并A列单元格
		if err := xlsxFile.MergeCell(sheetName, fmt.Sprintf("A%d", startRow), fmt.Sprintf("A%d", endRow)); err != nil {
			return startRow, endRow
		}
		return startRow, endRow
	}

	// 创建样式
	aColumnStyle, _ := xlsxFile.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	bColumnStyle, _ := xlsxFile.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	cColumnStyle, _ := xlsxFile.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Sheet1：合计、堂食、外带、外送、外卖
	sheet1Name := "Sheet1"
	index1, err := xlsxFile.NewSheet(sheet1Name)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	xlsxFile.SetActiveSheet(index1)

	rowIdx1 := 1
	writeChannelData(sheet1Name, &rowIdx1, channelNames["summary"], params.Result.Summary, false)
	writeChannelData(sheet1Name, &rowIdx1, channelNames["dine_in_store"], params.Result.DineInStore, false)
	writeChannelData(sheet1Name, &rowIdx1, channelNames["takeaway"], params.Result.Takeaway, false)
	writeChannelData(sheet1Name, &rowIdx1, channelNames["takeout_delivery"], params.Result.TakeoutDelivery, false)
	_, sheet1EndRow := writeChannelData(sheet1Name, &rowIdx1, channelNames["takeout"], params.Result.Takeout, false)

	// 为 Sheet1 设置样式和列宽
	for row := 1; row <= sheet1EndRow; row++ {
		cellA, _ := excelize.CoordinatesToCellName(1, row)
		xlsxFile.SetCellStyle(sheet1Name, cellA, cellA, aColumnStyle)
		cellB, _ := excelize.CoordinatesToCellName(2, row)
		xlsxFile.SetCellStyle(sheet1Name, cellB, cellB, bColumnStyle)
		cellC, _ := excelize.CoordinatesToCellName(3, row)
		xlsxFile.SetCellStyle(sheet1Name, cellC, cellC, cColumnStyle)
	}
	xlsxFile.SetColWidth(sheet1Name, "A", "A", 15)
	xlsxFile.SetColWidth(sheet1Name, "B", "B", 20)
	xlsxFile.SetColWidth(sheet1Name, "C", "C", 15)

	// Sheet2：合计、桌台、点餐-店内、点餐-外卖、外送、Grab、LINE MAN
	sheet2Name := "Sheet2"
	index2, err := xlsxFile.NewSheet(sheet2Name)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	xlsxFile.SetActiveSheet(index2)

	rowIdx2 := 1
	writeChannelData(sheet2Name, &rowIdx2, channelNames["summary"], params.Result.Summary, false)
	writeChannelData(sheet2Name, &rowIdx2, channelNames["table"], params.Result.Table, true)
	writeChannelData(sheet2Name, &rowIdx2, channelNames["dine_in"], params.Result.DineIn, false)
	writeChannelData(sheet2Name, &rowIdx2, channelNames["takeout_shop"], params.Result.TakeoutShop, false)
	writeChannelData(sheet2Name, &rowIdx2, channelNames["takeout_delivery"], params.Result.TakeoutDelivery, false)
	_, sheet2EndRow := writeChannelData(sheet2Name, &rowIdx2, channelNames["grab"], params.Result.Grab, false)
	// _, sheet2EndRow := writeChannelData(sheet2Name, &rowIdx2, channelNames["lineman"], params.Result.Lineman, false)

	// 为 Sheet2 设置样式和列宽
	for row := 1; row <= sheet2EndRow; row++ {
		cellA, _ := excelize.CoordinatesToCellName(1, row)
		xlsxFile.SetCellStyle(sheet2Name, cellA, cellA, aColumnStyle)
		cellB, _ := excelize.CoordinatesToCellName(2, row)
		xlsxFile.SetCellStyle(sheet2Name, cellB, cellB, bColumnStyle)
		cellC, _ := excelize.CoordinatesToCellName(3, row)
		xlsxFile.SetCellStyle(sheet2Name, cellC, cellC, cColumnStyle)
	}
	xlsxFile.SetColWidth(sheet2Name, "A", "A", 15)
	xlsxFile.SetColWidth(sheet2Name, "B", "B", 20)
	xlsxFile.SetColWidth(sheet2Name, "C", "C", 15)

	// 删除默认的 Sheet
	xlsxFile.DeleteSheet("Sheet")

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, params.Record.ExportName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.NewExportRecordRepo(db).Update(params.Record.Uuid, map[string]any{
		"file_uuid": res.Uuid,
		"status":    model.ExportStatusSuccess,
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil
}

// CountUserAnalysis 统计用户分析数据
func (s *businessSrv) CountUserAnalysis(ctx context.Context, req req.UserAnalysisReq) (*resp.UserAnalysisResp, error) {
	db := ctx.GetDB()
	statisticsRepo := repository.NewStatisticsRepo(db)

	// 处理日期时间字符串参数（优先级：StartTime/EndTime > QueryStartDate/QueryEndDate）
	startTime := req.StartTime
	endTime := req.EndTime
	if req.QueryStartDate != "" && req.QueryEndDate != "" && startTime == 0 && endTime == 0 {
		timeUtil := utils.SetTimezone(ctx.GetCompanySetting().Timezone)
		start, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			startTime = start
		}
		end, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			endTime = end
		}
	}

	// 处理默认时间：如果未传时间，使用今日范围
	if startTime == 0 || endTime == 0 {
		timezone := ctx.GetCompanySetting().Timezone
		timezoneUtil := utils.SetTimezone(timezone)
		startTime, endTime = timezoneUtil.TodayStartEndUnix()
	}

	// 参数校验
	if startTime > endTime {
		return nil, errors.WithMessage(errors.New("开始时间不能大于结束时间"))
	}

	// 获取语言
	language := ctx.GetLanguage()

	// 获取门店业务设置，检查是否开启国籍功能
	settingSrv := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	businessSetting, err := settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		ctx.Log().Error("获取门店业务设置失败", zap.Error(err))
		// 如果获取设置失败，默认不开启国籍统计
		businessSetting.EnableNationality = "0"
	}
	enableNationality := businessSetting.EnableNationality == "1"

	// 获取收银机设置，检查是否开启点餐和桌台功能
	cashierSetting, err := settingSrv.GetCashierSetting(ctx, nil)
	if err != nil {
		ctx.Log().Error("获取收银机设置失败", zap.Error(err))
		// 如果获取设置失败，默认不开启点餐和桌台统计
		cashierSetting.OrderMethod.IsCashierOrder = "0"
		cashierSetting.OrderMethod.IsTableOrder = "0"
	}
	enableCashierOrder := cashierSetting.OrderMethod.IsCashierOrder == "1"
	enableTableOrder := cashierSetting.OrderMethod.IsTableOrder == "1"

	// 调用 Repository 获取统计数据
	repoResult, err := statisticsRepo.CountUserAnalysis(startTime, endTime, language, enableNationality, enableCashierOrder, enableTableOrder, req.ExcludeDataManage)
	if err != nil {
		return nil, errors.WithMessage(err, "统计用户分析数据失败")
	}

	// 转换为响应格式
	convertToItems := func(items []model.UserAnalysisItemRepo) []resp.UserAnalysisItem {
		result := make([]resp.UserAnalysisItem, 0, len(items))
		for _, item := range items {
			result = append(result, resp.UserAnalysisItem{
				Name:       i18n.Translate(language, item.Name),
				OrderCount: item.OrderCount,
				Percentage: item.Percentage.InexactFloat64(),
			})
		}
		return result
	}

	return &resp.UserAnalysisResp{
		Nationality:   convertToItems(repoResult.Nationality),
		OrderSource:   convertToItems(repoResult.OrderSource),
		DeskSource:    convertToItems(repoResult.DeskSource),
		DiningMethod:  convertToItems(repoResult.DiningMethod),
		TakeoutMethod: convertToItems(repoResult.TakeoutMethod),
	}, nil
}

// ExportUserAnalysis 导出用户分析统计数据
func (s *businessSrv) ExportUserAnalysis(ctx context.Context, req req.UserAnalysisReq) error {
	db := ctx.GetDB()
	// 判断是否还有正在导出的任务
	oldRecord, err := repository.NewExportRecordRepo(db).GetUnfinishedExportRecord(model.ExportTypeUserAnalysis)
	if err != nil {
		return err
	}
	if oldRecord != nil {
		return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
	}

	// 获取统计数据
	result, err := s.CountUserAnalysis(ctx, req)
	if err != nil {
		return err
	}

	// 检查是否有数据
	hasData := len(result.Nationality) > 0 || len(result.OrderSource) > 0 || len(result.DeskSource) > 0 || len(result.DiningMethod) > 0
	if !hasData {
		return errors.WithMessage(errors.New("没有数据需要导出"))
	}

	fileNameMul := model.MultiLanguageName{
		EnName:   "User Analysis Statistics",
		ZhName:   "用户分析统计",
		ZhTwName: "用戶分析統計",
		ThName:   "สถิติการวิเคราะห์ผู้ใช้",
		MyName:   "အသုံးပြုသူခွဲခြမ်းစိတ်ဖြာစာရင်း",
		JaName:   "ユーザー分析統計",
		KoName:   "사용자 분석 통계",
		TrName:   "Kullanıcı Analiz İstatistikleri",
		SvName:   "Användaranalysstatistik",
	}

	// 创建导出任务
	params := map[string]interface{}{
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
	}
	paramsJson, err := json.Marshal(params)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeUserAnalysis)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeUserAnalysis,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(paramsJson),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	err = repository.NewExportRecordRepo(db).Create(record)
	if err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		_, err := s.ExportUserAnalysisTask(ctx, ExportUserAnalysisTaskParams{
			Record:      *record,
			FillNameMul: fileNameMul,
			Result:      result,
		})
		if err != nil {
			if err := repository.NewExportRecordRepo(ctx.GetDB()).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出用户分析统计数据失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			}
			return
		}
	})

	return nil
}

// ExportUserAnalysisTaskParams 导出用户分析统计数据参数
type ExportUserAnalysisTaskParams struct {
	Record      model.ExportRecord      // 导出记录
	FillNameMul model.MultiLanguageName // 多语言名称
	Result      *resp.UserAnalysisResp  // 统计数据
}

// ExportUserAnalysisTask 导出用户分析统计数据任务
func (s *businessSrv) ExportUserAnalysisTask(ctx context.Context, params ExportUserAnalysisTaskParams) (*resp.FileExportResp, error) {
	db := ctx.GetDB()
	lang := ctx.GetLanguage()

	// 创建 Excel 文件
	xlsxFile := excelize.NewFile()
	defer xlsxFile.Close()

	// 定义表头映射（使用 i18n 翻译）
	headers := map[string]string{
		"name":        i18n.Translate(lang, "名称"),
		"order_count": i18n.Translate(lang, "订单数"),
		"percentage":  i18n.Translate(lang, "占比%"),
	}

	// 写入四个统计维度
	writeSheet := func(sheetName string, items []resp.UserAnalysisItem) error {
		sheetIndex, err := xlsxFile.NewSheet(sheetName)
		if err != nil {
			return err
		}
		xlsxFile.SetActiveSheet(sheetIndex)

		// 写入表头
		xlsxFile.SetCellValue(sheetName, "A1", headers["name"])
		xlsxFile.SetCellValue(sheetName, "B1", headers["order_count"])
		xlsxFile.SetCellValue(sheetName, "C1", headers["percentage"])

		// 写入数据
		for i, item := range items {
			row := i + 2
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.Name)
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.OrderCount)
			xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Percentage)
		}

		return nil
	}

	// 写入各个维度（使用 i18n 翻译）
	sheetNames := map[string]string{
		"nationality":    i18n.Translate(lang, "国籍"),
		"order_source":   i18n.Translate(lang, "点餐方式来源"),
		"desk_source":    i18n.Translate(lang, "桌台方式来源"),
		"dining_method":  i18n.Translate(lang, "用餐"),
		"takeout_method": i18n.Translate(lang, "外卖方式"),
	}

	if err := writeSheet(sheetNames["nationality"], params.Result.Nationality); err != nil {
		return nil, errors.WithMessage(err, "写入国籍统计失败")
	}
	if err := writeSheet(sheetNames["order_source"], params.Result.OrderSource); err != nil {
		return nil, errors.WithMessage(err, "写入点餐方式来源统计失败")
	}
	if err := writeSheet(sheetNames["desk_source"], params.Result.DeskSource); err != nil {
		return nil, errors.WithMessage(err, "写入桌台方式来源统计失败")
	}
	if err := writeSheet(sheetNames["dining_method"], params.Result.DiningMethod); err != nil {
		return nil, errors.WithMessage(err, "写入用餐方式统计失败")
	}
	if err := writeSheet(sheetNames["takeout_method"], params.Result.TakeoutMethod); err != nil {
		return nil, errors.WithMessage(err, "写入外卖方式统计失败")
	}

	// 删除默认的 Sheet1
	xlsxFile.DeleteSheet("Sheet1")

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, params.Record.ExportName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if err := repository.NewExportRecordRepo(db).Update(params.Record.Uuid, map[string]any{
		"file_uuid": res.Uuid,
		"status":    model.ExportStatusSuccess,
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil
}

// GetCompanyList 获取门店汇总统计可选择的门店列表
func (s *businessSrv) GetCompanyList(ctx context.Context) (*resp.CompanySummaryListResp, error) {
	// 获取当前门店和设置
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()

	var companyList []*resp.CompanySummaryItem

	// 获取数据库管理器
	dbm := database.GetDBManager(config.Database)
	saasDB := dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)

	// 总店：使用 GetVisibleCompanyList 获取本店及下级所有子店
	if companySetting.IsHeadquarter() {
		visibleCompanies, err := companyRepo.GetVisibleCompanyList(company.Uuid)
		if err != nil {
			return nil, errors.WithMessage(err, "获取可见门店列表失败")
		}

		// 转换为 CompanySummaryItem 格式
		companyList = make([]*resp.CompanySummaryItem, 0, len(visibleCompanies))
		for _, c := range visibleCompanies {
			companyList = append(companyList, &resp.CompanySummaryItem{
				CompanyUuid: c.Uuid,
				CompanyName: c.Name,
			})
		}
	} else {
		// 子店：使用 AuthService.GetCompanyList 获取本店及已授权的其他门店
		// 创建必要的服务实例
		captchaSrv := NewCaptchaSrv(nil) // 这里不需要验证码服务，传 nil
		roleAccessSrv := NewRoleAccessSrv(dbm)
		settingSrv := setting.NewSrv(dbm, cache.Global)
		deviceSrv := NewDeviceSrv(settingSrv, dbm)
		cashBoxSrv := NewCashBoxSrv(dbm)
		statisticsSrv := NewStatisticsSrv()
		staffShiftSrv := NewStaffShiftSrv(cache.Global, dbm, cashBoxSrv, statisticsSrv)
		authSrv := NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

		// 获取子店的门店列表
		authCompanyList := authSrv.GetCompanyList(ctx)
		// 转换为 CompanySummaryItem 格式
		companyList = make([]*resp.CompanySummaryItem, 0, len(authCompanyList))
		for _, c := range authCompanyList {
			companyList = append(companyList, &resp.CompanySummaryItem{
				CompanyUuid: c.CompanyUuid,
				CompanyName: c.CompanyName,
			})
		}
	}

	return &resp.CompanySummaryListResp{
		List: companyList,
	}, nil
}

// GetCompanyPaymentMethods 获取有权限的所有门店的支付方式（汇总去重）
func (s *businessSrv) GetCompanyPaymentMethods(ctx context.Context) (*resp.CompanyPaymentMethodListResp, error) {
	// 1. 获取有权限的门店列表
	companyListResp, err := s.GetCompanyList(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取门店列表失败")
	}

	if len(companyListResp.List) == 0 {
		return &resp.CompanyPaymentMethodListResp{
			List: []resp.CompanyPaymentMethodItem{},
		}, nil
	}

	// 2. 获取数据库管理器
	dbm := database.GetDBManager(config.Database)

	// 3. 使用 goroutine 并发查询各个门店，提高性能
	// 使用带缓冲的 channel 控制并发数量，最多10个协程，避免资源消耗过大
	maxConcurrent := 10
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// 使用 channel 收集结果，避免竞态条件
	type paymentMethodInfo struct {
		Name        string
		PaymentName string
		Sort        int
		CreateTime  int64
		ID          uint
	}
	type resultItem struct {
		paymentMethods []paymentMethodInfo
		err            error
	}
	resultChan := make(chan resultItem, len(companyListResp.List))

	// 并发查询各个门店
	for _, companyItem := range companyListResp.List {
		wg.Add(1)
		utils.Go(func() {
			func(item *resp.CompanySummaryItem) {
				defer wg.Done()

				// 获取信号量，控制并发数
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				companyUuid := item.CompanyUuid

				// 获取门店数据库连接
				shopDB := dbm.GetDB(companyUuid)
				if shopDB == nil {
					logger.Logger.Warn("获取门店数据库连接失败", zap.Uint64("company_uuid", companyUuid))
					resultChan <- resultItem{paymentMethods: nil, err: errors.New("获取门店数据库连接失败")}
					return
				}

				// 获取支付方式列表（只获取启用的支付方式）
				paymentMethodRepo := repository.NewPaymentMethodRepo(shopDB)
				paymentRepo := NewPaymentRepo(ctx, dbm)
				paymentMethods := paymentMethodRepo.GetPaymentMethodList(
					paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable),
				)
				lianLianPayAvailable := true
				err := paymentRepo.ValidateConfigError(companyUuid)
				if err != nil {
					lianLianPayAvailable = false
				}

				// 收集支付方式信息（包含名称、排序、创建时间、ID）
				paymentMethodInfos := make([]paymentMethodInfo, 0, len(paymentMethods))
				for _, method := range paymentMethods {
					if method != nil {
						if method.PaymentName != "" && paymentMethodRepo.FilterPaymentMethod(*method, lianLianPayAvailable) {
							paymentMethodInfos = append(paymentMethodInfos, paymentMethodInfo{
								Name:        method.Name,
								PaymentName: method.PaymentName,
								Sort:        method.Sort,
								CreateTime:  method.CreateTime,
								ID:          method.ID,
							})
						}
					}
				}

				resultChan <- resultItem{paymentMethods: paymentMethodInfos, err: nil}
			}(companyItem)
		})
	}

	// 等待所有 goroutine 完成
	utils.Go(func() {
		wg.Wait()
		close(resultChan)
	})

	// 4. 收集所有门店的支付方式（用于去重）
	// 使用 map 存储支付方式信息，key 为支付方式名称
	// 如果支付方式名称相同，保留 Sort 最小、CreateTime 最大、ID 最大的那个
	paymentMethodMap := make(map[string]paymentMethodInfo)
	for result := range resultChan {
		if result.err != nil {
			logger.Logger.Warn("查询门店支付方式失败", zap.Error(result.err))
			continue
		}
		for _, pmInfo := range result.paymentMethods {
			existing, exists := paymentMethodMap[pmInfo.Name]
			if !exists {
				// 如果不存在，直接添加
				paymentMethodMap[pmInfo.Name] = pmInfo
			} else {
				// 如果存在，比较 Sort、CreateTime 和 ID
				// 优先按 Sort 升序，如果 Sort 相同则按 CreateTime 倒序，如果都相同则按 ID 倒序
				if pmInfo.Sort < existing.Sort {
					paymentMethodMap[pmInfo.Name] = pmInfo
				} else if pmInfo.Sort == existing.Sort {
					if pmInfo.CreateTime > existing.CreateTime {
						paymentMethodMap[pmInfo.Name] = pmInfo
					} else if pmInfo.CreateTime == existing.CreateTime && pmInfo.ID > existing.ID {
						paymentMethodMap[pmInfo.Name] = pmInfo
					}
				}
			}
		}
	}

	// 5. 转换为响应格式（去重后的支付方式列表）
	paymentMethodList := make([]resp.CompanyPaymentMethodItem, 0, len(paymentMethodMap))
	for _, pmInfo := range paymentMethodMap {
		paymentMethodList = append(paymentMethodList, resp.CompanyPaymentMethodItem{
			PaymentName: pmInfo.Name,
		})
	}

	// 6. 按 Sort 升序、CreateTime 倒序、ID 倒序排序
	sort.Slice(paymentMethodList, func(i, j int) bool {
		pmI := paymentMethodMap[paymentMethodList[i].PaymentName]
		pmJ := paymentMethodMap[paymentMethodList[j].PaymentName]
		if pmI.Sort != pmJ.Sort {
			return pmI.Sort < pmJ.Sort // Sort 升序
		}
		if pmI.CreateTime != pmJ.CreateTime {
			return pmI.CreateTime > pmJ.CreateTime // CreateTime 倒序
		}
		return pmI.ID > pmJ.ID // ID 倒序
	})

	return &resp.CompanyPaymentMethodListResp{
		List: paymentMethodList,
	}, nil
}

// CountCompanyBusinessSummary 获取门店汇总统计（营业数据汇总、支付方式汇总、退款金额汇总）
func (s *businessSrv) CountCompanyBusinessSummary(ctx context.Context, request req.StatisticsCompanySummaryReq) (interface{}, error) {
	// 根据指标类型调用不同的处理逻辑
	if request.IndicatorType == "payment_method" {
		return s.countCompanyPaymentMethodSummary(ctx, request)
	}
	if request.IndicatorType == "refund" {
		return s.countCompanyRefundSummary(ctx, request)
	}

	// 默认处理营业数据汇总
	if request.IndicatorType != "business" && request.IndicatorType != "" {
		return nil, errors.New("当前仅支持营业数据汇总、支付方式汇总和退款金额汇总指标类型")
	}

	// 获取当前用户时区
	timezone := ctx.GetCompanySetting().Timezone
	timeUtil := utils.SetTimezone(timezone)

	// 调用 GetCompanyList 获取所有可见商家列表（包含 CompanyUuid 和 CompanyName）
	companyListResp, err := s.GetCompanyList(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取门店列表失败")
	}

	// 处理 CompanyUuids：如果为空，使用所有可见门店；否则过滤出匹配的门店
	var companyList []*resp.CompanySummaryItem
	if len(request.CompanyUuids) == 0 {
		// CompanyUuids 为空时，使用所有可见门店
		companyList = companyListResp.List
	} else {
		// CompanyUuids 不为空时，过滤出匹配的门店（保持顺序）
		companyUuidSet := make(map[uint64]bool)
		for _, uuid := range request.CompanyUuids {
			companyUuidSet[uuid] = true
		}
		companyList = make([]*resp.CompanySummaryItem, 0, len(request.CompanyUuids))
		for _, c := range companyListResp.List {
			if companyUuidSet[c.CompanyUuid] {
				companyList = append(companyList, c)
			}
		}
	}

	// 处理默认值：QueryStartDate 为空时，默认为今日开始日期：YYYY-MM-DD 00:00:00
	queryStartDate := request.QueryStartDate
	if queryStartDate == "" {
		startTime, _ := timeUtil.TodayStartEnd()
		queryStartDate = startTime.Format("2006-01-02 15:04:05")
	}

	// 处理默认值：QueryEndDate 为空时，默认为今日结束日期：YYYY-MM-DD 23:59:59
	queryEndDate := request.QueryEndDate
	if queryEndDate == "" {
		_, endTime := timeUtil.TodayStartEnd()
		queryEndDate = endTime.Format("2006-01-02 15:04:05")
	}

	// 获取数据库管理器
	dbm := database.GetDBManager(config.Database)
	settingSrv := setting.NewSrv(dbm, cache.Global)
	saasDB := dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)

	// 使用 goroutine 并发查询各个门店，提高性能
	// 使用带缓冲的 channel 控制并发数量，最多10个协程，避免资源消耗过大
	maxConcurrent := 10
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// 使用 channel 收集结果，避免竞态条件
	type resultItem struct {
		items []resp.CompanyBusinessSummaryItem
		err   error
	}
	resultChan := make(chan resultItem, len(companyList))

	// 并发查询各个门店
	for _, companyItem := range companyList {
		wg.Add(1)
		utils.Go(func() {
			func(item *resp.CompanySummaryItem) {
				defer wg.Done()

				// 获取信号量，控制并发数
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				companyUuid := item.CompanyUuid
				companyName := item.CompanyName

				// 获取完整的 Company 信息（用于 SetCompany）
				company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
				if err != nil {
					logger.Logger.Warn("获取门店信息失败", zap.Uint64("company_uuid", companyUuid), zap.Error(err))
					resultChan <- resultItem{items: nil, err: err}
					return
				}

				// 获取门店数据库连接
				shopDB := dbm.GetDB(companyUuid)
				if shopDB == nil {
					logger.Logger.Warn("获取门店数据库连接失败", zap.Uint64("company_uuid", companyUuid))
					resultChan <- resultItem{items: nil, err: errors.New("获取门店数据库连接失败")}
					return
				}

				// 创建门店 context
				shopCtx := ctx.Copy()
				shopCtx.SetCompanyUuid(companyUuid)
				shopCtx.SetCompany(*company)
				shopCtx.SetDB(shopDB)

				// 获取门店设置
				companySettingRepo := repository.NewCompanySettingRepo(shopDB)
				shopCompanySetting, err := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
					return db.Where("company_uuid = ?", companyUuid)
				})
				if err != nil {
					logger.Logger.Warn("获取门店设置失败", zap.Uint64("company_uuid", companyUuid), zap.Error(err))
					resultChan <- resultItem{items: nil, err: err}
					return
				}
				shopCtx.SetCompanySetting(shopCompanySetting)
				dataManageSetting := settingSrv.GetDataManageSetting(shopCtx)
				excludeDataManage := shopCompanySetting.IsOpenDataManagement() && dataManageSetting.IsEnableDataManage

				// 调用统计服务获取门店数据
				statisticsReq := req.StatisticsSummaryReq{
					PageReq: dto.PageReq{
						PageNo:   1,
						PageSize: 1000, // 获取所有数据，不分页
					},
					QueryStartDate:    queryStartDate,
					QueryEndDate:      queryEndDate,
					Cycle:             request.Cycle, // 使用请求参数中的周期设置
					ExcludeDataManage: excludeDataManage,
				}

				statisticsData := s.statisticsSrv.CountBusinessSummary(shopCtx, statisticsReq)

				// 转换为响应格式
				items := make([]resp.CompanyBusinessSummaryItem, 0, len(statisticsData.StatisticsComprehensiveList))
				for _, statItem := range statisticsData.StatisticsComprehensiveList {
					items = append(items, resp.CompanyBusinessSummaryItem{
						Date:               statItem.Date,
						CompanyName:        companyName, // 使用 GetCompanyList 返回的 CompanyName，避免重复查询
						OrderAmount:        statItem.OrderAmount,
						PayAmount:          statItem.PayAmount,
						OrderNum:           statItem.OrderNum,
						MealNum:            statItem.MealNum,
						DeskNum:            statItem.DeskNum,
						AvgCustomerPrice:   statItem.PayAmountMealAvg,
						OrderAmountMealAvg: statItem.OrderAmountMealAvg,
						OrderAmountAvg:     statItem.OrderAmountAvg,
						PayAmountAvg:       statItem.PayAmountAvg,
						InstantOrderAmount: statItem.InstantOrderAmount,
						DeskOrderAmount:    statItem.DeskOrderAmount,
						TakeoutOrderAmount: statItem.TakeoutOrderAmount,
					})
				}

				resultChan <- resultItem{items: items, err: nil}
			}(companyItem)
		})
	}

	// 等待所有 goroutine 完成
	utils.Go(func() {
		wg.Wait()
		close(resultChan)
	})

	// 收集所有门店的统计数据
	allItems := make([]resp.CompanyBusinessSummaryItem, 0)
	for result := range resultChan {
		if result.err != nil {
			// 记录错误但继续处理其他门店
			continue
		}
		allItems = append(allItems, result.items...)
	}

	// 根据 report 参数决定返回明细表还是汇总表
	// report = 0: 明细表，返回每个营业日每个商家的数据，不包括汇总行（SummaryRow 返回默认值）
	// report = 1: 汇总表，返回每个营业日每个商家数据总和，包括汇总行
	var finalList []resp.CompanyBusinessSummaryItem
	var summaryRow resp.CompanyBusinessSummaryItem

	if request.Report == 1 {
		// 汇总表：按日期分组汇总
		// 记录每个日期对应的商家名称集合（用于去重）
		dateCompanyNamesMap := make(map[string]map[string]bool)

		// 使用 decimal 进行金额累加，避免精度错误
		type dateSummaryDecimal struct {
			OrderAmount        decimal.Decimal
			PayAmount          decimal.Decimal
			InstantOrderAmount decimal.Decimal
			DeskOrderAmount    decimal.Decimal
			TakeoutOrderAmount decimal.Decimal
			OrderNum           int64
			MealNum            int64
			DeskNum            int64
		}
		dateDecimalMap := make(map[string]*dateSummaryDecimal)

		for _, item := range allItems {
			if dateDecimal, exists := dateDecimalMap[item.Date]; exists {
				// 累加同一天的数据（使用 decimal）
				dateDecimal.OrderAmount = dateDecimal.OrderAmount.Add(decimal.NewFromFloat(item.OrderAmount))
				dateDecimal.PayAmount = dateDecimal.PayAmount.Add(decimal.NewFromFloat(item.PayAmount))
				dateDecimal.InstantOrderAmount = dateDecimal.InstantOrderAmount.Add(decimal.NewFromFloat(item.InstantOrderAmount))
				dateDecimal.DeskOrderAmount = dateDecimal.DeskOrderAmount.Add(decimal.NewFromFloat(item.DeskOrderAmount))
				dateDecimal.TakeoutOrderAmount = dateDecimal.TakeoutOrderAmount.Add(decimal.NewFromFloat(item.TakeoutOrderAmount))
				dateDecimal.OrderNum += item.OrderNum
				dateDecimal.MealNum += item.MealNum
				dateDecimal.DeskNum += item.DeskNum
				// 收集商家名称
				if dateCompanyNamesMap[item.Date] == nil {
					dateCompanyNamesMap[item.Date] = make(map[string]bool)
				}
				dateCompanyNamesMap[item.Date][item.CompanyName] = true
			} else {
				// 创建新的日期汇总项（使用 decimal）
				dateDecimalMap[item.Date] = &dateSummaryDecimal{
					OrderAmount:        decimal.NewFromFloat(item.OrderAmount),
					PayAmount:          decimal.NewFromFloat(item.PayAmount),
					InstantOrderAmount: decimal.NewFromFloat(item.InstantOrderAmount),
					DeskOrderAmount:    decimal.NewFromFloat(item.DeskOrderAmount),
					TakeoutOrderAmount: decimal.NewFromFloat(item.TakeoutOrderAmount),
					OrderNum:           item.OrderNum,
					MealNum:            item.MealNum,
					DeskNum:            item.DeskNum,
				}
				// 初始化商家名称集合
				if dateCompanyNamesMap[item.Date] == nil {
					dateCompanyNamesMap[item.Date] = make(map[string]bool)
				}
				dateCompanyNamesMap[item.Date][item.CompanyName] = true
			}
		}

		// 转换为列表并计算人均和单均，设置商家名称
		finalList = make([]resp.CompanyBusinessSummaryItem, 0, len(dateDecimalMap))
		for date, dateDecimal := range dateDecimalMap {
			dateItem := resp.CompanyBusinessSummaryItem{
				Date:               date,
				CompanyName:        "", // 稍后设置
				OrderAmount:        dateDecimal.OrderAmount.InexactFloat64(),
				PayAmount:          dateDecimal.PayAmount.InexactFloat64(),
				OrderNum:           dateDecimal.OrderNum,
				MealNum:            dateDecimal.MealNum,
				DeskNum:            dateDecimal.DeskNum,
				InstantOrderAmount: dateDecimal.InstantOrderAmount.InexactFloat64(),
				DeskOrderAmount:    dateDecimal.DeskOrderAmount.InexactFloat64(),
				TakeoutOrderAmount: dateDecimal.TakeoutOrderAmount.InexactFloat64(),
			}

			// 计算人均和单均（使用 decimal）
			if dateDecimal.MealNum > 0 {
				avgCustomerPrice := dateDecimal.PayAmount.Div(decimal.NewFromInt(dateDecimal.MealNum))
				orderAmountMealAvg := dateDecimal.OrderAmount.Div(decimal.NewFromInt(dateDecimal.MealNum))
				dateItem.AvgCustomerPrice = utils.Round(avgCustomerPrice.InexactFloat64(), 2)
				dateItem.OrderAmountMealAvg = utils.Round(orderAmountMealAvg.InexactFloat64(), 2)
			}
			if dateDecimal.OrderNum > 0 {
				orderAmountAvg := dateDecimal.OrderAmount.Div(decimal.NewFromInt(dateDecimal.OrderNum))
				payAmountAvg := dateDecimal.PayAmount.Div(decimal.NewFromInt(dateDecimal.OrderNum))
				dateItem.OrderAmountAvg = utils.Round(orderAmountAvg.InexactFloat64(), 2)
				dateItem.PayAmountAvg = utils.Round(payAmountAvg.InexactFloat64(), 2)
			}

			// 设置商家名称：将所有商家名称用、符号连接
			companyNames := make([]string, 0, len(dateCompanyNamesMap[date]))
			for name := range dateCompanyNamesMap[date] {
				companyNames = append(companyNames, name)
			}
			sort.Strings(companyNames) // 排序以保证一致性
			dateItem.CompanyName = strings.Join(companyNames, "、")
			finalList = append(finalList, dateItem)
		}

		// 按日期排序
		sort.Slice(finalList, func(i, j int) bool {
			return finalList[i].Date < finalList[j].Date
		})

		// 计算总汇总行（使用 decimal）
		var totalOrderAmount, totalPayAmount, totalInstantOrderAmount, totalDeskOrderAmount, totalTakeoutOrderAmount decimal.Decimal
		var totalOrderNum, totalMealNum, totalDeskNum int64

		for _, item := range finalList {
			totalOrderAmount = totalOrderAmount.Add(decimal.NewFromFloat(item.OrderAmount))
			totalPayAmount = totalPayAmount.Add(decimal.NewFromFloat(item.PayAmount))
			totalOrderNum += item.OrderNum
			totalMealNum += item.MealNum
			totalDeskNum += item.DeskNum
			totalInstantOrderAmount = totalInstantOrderAmount.Add(decimal.NewFromFloat(item.InstantOrderAmount))
			totalDeskOrderAmount = totalDeskOrderAmount.Add(decimal.NewFromFloat(item.DeskOrderAmount))
			totalTakeoutOrderAmount = totalTakeoutOrderAmount.Add(decimal.NewFromFloat(item.TakeoutOrderAmount))
		}

		// 计算总汇总行的人均和单均（使用 decimal）
		var avgCustomerPrice, orderAmountMealAvg, orderAmountAvg, payAmountAvg decimal.Decimal

		if totalMealNum > 0 {
			avgCustomerPrice = totalPayAmount.Div(decimal.NewFromInt(totalMealNum))
			orderAmountMealAvg = totalOrderAmount.Div(decimal.NewFromInt(totalMealNum))
		}
		if totalOrderNum > 0 {
			orderAmountAvg = totalOrderAmount.Div(decimal.NewFromInt(totalOrderNum))
			payAmountAvg = totalPayAmount.Div(decimal.NewFromInt(totalOrderNum))
		}

		// 收集所有商家名称（用于汇总行）
		allCompanyNamesSet := make(map[string]bool)
		for _, namesMap := range dateCompanyNamesMap {
			for name := range namesMap {
				allCompanyNamesSet[name] = true
			}
		}
		allCompanyNames := make([]string, 0, len(allCompanyNamesSet))
		for name := range allCompanyNamesSet {
			allCompanyNames = append(allCompanyNames, name)
		}
		sort.Strings(allCompanyNames) // 排序以保证一致性

		summaryRow = resp.CompanyBusinessSummaryItem{
			Date:               "汇总",
			CompanyName:        strings.Join(allCompanyNames, "、"), // 所有商家名称用、符号连接
			OrderAmount:        utils.Round(totalOrderAmount.InexactFloat64(), 2),
			PayAmount:          utils.Round(totalPayAmount.InexactFloat64(), 2),
			OrderNum:           totalOrderNum,
			MealNum:            totalMealNum,
			DeskNum:            totalDeskNum,
			AvgCustomerPrice:   utils.Round(avgCustomerPrice.InexactFloat64(), 2),
			OrderAmountMealAvg: utils.Round(orderAmountMealAvg.InexactFloat64(), 2),
			OrderAmountAvg:     utils.Round(orderAmountAvg.InexactFloat64(), 2),
			PayAmountAvg:       utils.Round(payAmountAvg.InexactFloat64(), 2),
			InstantOrderAmount: utils.Round(totalInstantOrderAmount.InexactFloat64(), 2),
			DeskOrderAmount:    utils.Round(totalDeskOrderAmount.InexactFloat64(), 2),
			TakeoutOrderAmount: utils.Round(totalTakeoutOrderAmount.InexactFloat64(), 2),
		}
	} else {
		// 明细表：返回每个营业日每个商家的数据，不包括汇总行（SummaryRow 返回默认值）
		// 按日期排序
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].Date < allItems[j].Date
		})
		finalList = allItems
		// 明细表返回默认值
		summaryRow = resp.CompanyBusinessSummaryItem{}
	}

	// 分页处理
	total := int64(len(finalList))
	pageNo := utils.IfInt(request.PageNo > 0, request.PageNo, 1)
	pageSize := utils.IfInt(request.PageSize > 0, request.PageSize, 20)
	start := (pageNo - 1) * pageSize
	end := start + pageSize

	if start > len(finalList) {
		start = len(finalList)
	}
	if end > len(finalList) {
		end = len(finalList)
	}

	var list []resp.CompanyBusinessSummaryItem
	if start < len(finalList) {
		list = finalList[start:end]
	} else {
		list = make([]resp.CompanyBusinessSummaryItem, 0)
	}

	return &resp.CompanyBusinessSummaryResp{
		Meta: dto.PageResponse{
			PageNo:   pageNo,
			PageSize: pageSize,
			Total:    total,
		},
		List:       list,
		SummaryRow: summaryRow,
	}, nil
}

// countCompanyPaymentMethodSummary 获取门店支付方式汇总统计
func (s *businessSrv) countCompanyPaymentMethodSummary(ctx context.Context, request req.StatisticsCompanySummaryReq) (*resp.CompanyPaymentMethodSummaryResp, error) {
	// 获取当前用户时区
	timezone := ctx.GetCompanySetting().Timezone
	timeUtil := utils.SetTimezone(timezone)

	// 调用 GetCompanyList 获取所有可见商家列表（包含 CompanyUuid 和 CompanyName）
	companyListResp, err := s.GetCompanyList(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取门店列表失败")
	}

	// 处理 CompanyUuids：如果为空，使用所有可见门店；否则过滤出匹配的门店
	var companyList []*resp.CompanySummaryItem
	if len(request.CompanyUuids) == 0 {
		// CompanyUuids 为空时，使用所有可见门店
		companyList = companyListResp.List
	} else {
		// CompanyUuids 不为空时，过滤出匹配的门店（保持顺序）
		companyUuidSet := make(map[uint64]bool)
		for _, uuid := range request.CompanyUuids {
			companyUuidSet[uuid] = true
		}
		companyList = make([]*resp.CompanySummaryItem, 0, len(request.CompanyUuids))
		for _, c := range companyListResp.List {
			if companyUuidSet[c.CompanyUuid] {
				companyList = append(companyList, c)
			}
		}
	}

	// 处理默认值：QueryStartDate 为空时，默认为今日开始日期：YYYY-MM-DD 00:00:00
	queryStartDate := request.QueryStartDate
	if queryStartDate == "" {
		startTime, _ := timeUtil.TodayStartEnd()
		queryStartDate = startTime.Format("2006-01-02 15:04:05")
	}

	// 处理默认值：QueryEndDate 为空时，默认为今日结束日期：YYYY-MM-DD 23:59:59
	queryEndDate := request.QueryEndDate
	if queryEndDate == "" {
		_, endTime := timeUtil.TodayStartEnd()
		queryEndDate = endTime.Format("2006-01-02 15:04:05")
	}

	// 获取数据库管理器
	dbm := database.GetDBManager(config.Database)
	settingSrv := setting.NewSrv(dbm, cache.Global)
	saasDB := dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)

	// 使用 goroutine 并发查询各个门店，提高性能
	// 使用带缓冲的 channel 控制并发数量，最多10个协程，避免资源消耗过大
	maxConcurrent := 10
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// 使用 channel 收集结果，避免竞态条件
	type resultItem struct {
		items []resp.CompanyPaymentMethodSummaryItem
		err   error
	}
	resultChan := make(chan resultItem, len(companyList))

	// 并发查询各个门店
	for _, companyItem := range companyList {
		wg.Add(1)
		utils.Go(func() {
			func(item *resp.CompanySummaryItem) {
				defer wg.Done()

				// 获取信号量，控制并发数
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				companyUuid := item.CompanyUuid
				companyName := item.CompanyName

				// 获取完整的 Company 信息（用于 SetCompany）
				company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
				if err != nil {
					logger.Logger.Warn("获取门店信息失败", zap.Uint64("company_uuid", companyUuid), zap.Error(err))
					resultChan <- resultItem{items: nil, err: err}
					return
				}

				// 获取门店数据库连接
				shopDB := dbm.GetDB(companyUuid)
				if shopDB == nil {
					logger.Logger.Warn("获取门店数据库连接失败", zap.Uint64("company_uuid", companyUuid))
					resultChan <- resultItem{items: nil, err: errors.New("获取门店数据库连接失败")}
					return
				}

				// 创建门店 context
				shopCtx := ctx.Copy()
				shopCtx.SetCompanyUuid(companyUuid)
				shopCtx.SetCompany(*company)
				shopCtx.SetDB(shopDB)

				// 获取门店设置
				companySettingRepo := repository.NewCompanySettingRepo(shopDB)
				shopCompanySetting, err := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
					return db.Where("company_uuid = ?", companyUuid)
				})
				if err != nil {
					logger.Logger.Warn("获取门店设置失败", zap.Uint64("company_uuid", companyUuid), zap.Error(err))
					resultChan <- resultItem{items: nil, err: err}
					return
				}
				shopCtx.SetCompanySetting(shopCompanySetting)
				dataManageSetting := settingSrv.GetDataManageSetting(shopCtx)
				excludeDataManage := shopCompanySetting.IsOpenDataManagement() && dataManageSetting.IsEnableDataManage

				// 调用统计服务获取门店支付方式数据
				statisticsReq := req.StatisticsPaymentMethodReq{
					PageReq: dto.PageReq{
						PageNo:   1,
						PageSize: 1000, // 获取所有数据，不分页
					},
					QueryStartDate:    queryStartDate,
					QueryEndDate:      queryEndDate,
					Cycle:             request.Cycle, // 使用请求参数中的周期设置
					ExcludeDataManage: excludeDataManage,
					OrderDelivery:     1,
					Source:            1,
				}

				// 处理支付方式筛选：使用支付方式名称（因为不同商家的同一支付方式UUID可能不同）
				if len(request.PaymentMethodNames) > 0 {
					statisticsReq.PaymentMethodNames = request.PaymentMethodNames
					statisticsReq.OrderDelivery = 0
				}

				statisticsData := s.statisticsSrv.CountBusinessPaymentMethod(shopCtx, statisticsReq)

				// 转换为响应格式
				items := make([]resp.CompanyPaymentMethodSummaryItem, 0, len(statisticsData.StatisticsPaymentMethodList))
				for _, statItem := range statisticsData.StatisticsPaymentMethodList {
					items = append(items, resp.CompanyPaymentMethodSummaryItem{
						Date:          statItem.Date,
						CompanyName:   companyName, // 使用 GetCompanyList 返回的 CompanyName，避免重复查询
						PaymentName:   statItem.PaymentName,
						PaymentAmount: statItem.PaymentAmount,
						PaymentNum:    statItem.PaymentNum,
						PaymentRatio:  0.0, // 稍后计算占比
					})
				}

				resultChan <- resultItem{items: items, err: nil}
			}(companyItem)
		})
	}

	// 等待所有 goroutine 完成
	utils.Go(func() {
		wg.Wait()
		close(resultChan)
	})

	// 收集所有门店的统计数据
	allItems := make([]resp.CompanyPaymentMethodSummaryItem, 0)
	for result := range resultChan {
		if result.err != nil {
			// 记录错误但继续处理其他门店
			continue
		}
		allItems = append(allItems, result.items...)
	}

	// 计算支付占比：需要先计算每个日期/门店的总支付金额（使用 decimal）
	// 按日期和门店分组，计算总支付金额
	type dateCompanyKey struct {
		Date        string
		CompanyName string
	}
	dateCompanyTotalMap := make(map[dateCompanyKey]decimal.Decimal)
	for _, item := range allItems {
		key := dateCompanyKey{
			Date:        item.Date,
			CompanyName: item.CompanyName,
		}
		if total, exists := dateCompanyTotalMap[key]; exists {
			dateCompanyTotalMap[key] = total.Add(decimal.NewFromFloat(item.PaymentAmount))
		} else {
			dateCompanyTotalMap[key] = decimal.NewFromFloat(item.PaymentAmount)
		}
	}

	// 计算占比（使用 decimal）
	// 支付占比 = 某个支付方式的支付金额 / 该日期该门店所有支付方式的总金额（实付金额）
	for i := range allItems {
		key := dateCompanyKey{
			Date:        allItems[i].Date,
			CompanyName: allItems[i].CompanyName,
		}
		totalAmount := dateCompanyTotalMap[key] // 该日期该门店所有支付方式的总金额（实付金额）
		if totalAmount.GreaterThan(decimal.Zero) {
			paymentAmount := decimal.NewFromFloat(allItems[i].PaymentAmount) // 某个支付方式的支付金额
			ratio := paymentAmount.Div(totalAmount).Mul(decimal.NewFromInt(100))
			allItems[i].PaymentRatio = utils.Round(ratio.InexactFloat64(), 2)
		}
	}

	// 根据 report 参数决定返回明细表还是汇总表
	// report = 0: 明细表，返回每个营业日每个商家每个支付方式的数据，不包括汇总行（SummaryRow 返回空数组）
	// report = 1: 汇总表，返回每个营业日每个支付方式的数据总和（跨门店），包括汇总行（按支付方式分组）
	var finalList []resp.CompanyPaymentMethodSummaryItem
	var summaryRow []resp.CompanyPaymentMethodSummaryItem

	if request.Report == 1 {
		// 汇总表：按日期、支付方式分组汇总，跨门店汇总
		// 同一营业日，同一支付方式，不同门店的数据合并为一条记录
		type datePaymentKey struct {
			Date        string
			PaymentName string
		}
		// 记录每个日期+支付方式对应的所有门店名称集合
		datePaymentCompanyNamesMap := make(map[datePaymentKey]map[string]bool)

		// 使用 decimal 进行金额累加，避免精度错误
		type datePaymentDecimal struct {
			PaymentAmount decimal.Decimal
			PaymentNum    int64
		}
		datePaymentDecimalMap := make(map[datePaymentKey]*datePaymentDecimal)

		for _, item := range allItems {
			key := datePaymentKey{
				Date:        item.Date,
				PaymentName: item.PaymentName,
			}

			if decimalItem, exists := datePaymentDecimalMap[key]; exists {
				// 累加同一日期、同一支付方式的数据（跨门店汇总，使用 decimal）
				decimalItem.PaymentAmount = decimalItem.PaymentAmount.Add(decimal.NewFromFloat(item.PaymentAmount))
				decimalItem.PaymentNum += item.PaymentNum
			} else {
				// 创建新的日期支付方式汇总项（使用 decimal）
				datePaymentDecimalMap[key] = &datePaymentDecimal{
					PaymentAmount: decimal.NewFromFloat(item.PaymentAmount),
					PaymentNum:    item.PaymentNum,
				}
			}

			// 收集每个日期+支付方式对应的所有门店名称（用于显示）
			if datePaymentCompanyNamesMap[key] == nil {
				datePaymentCompanyNamesMap[key] = make(map[string]bool)
			}
			datePaymentCompanyNamesMap[key][item.CompanyName] = true
		}

		// 计算每个日期的总支付金额（用于计算占比，使用 decimal）
		dateTotalMap := make(map[string]decimal.Decimal)
		for key, decimalItem := range datePaymentDecimalMap {
			if total, exists := dateTotalMap[key.Date]; exists {
				dateTotalMap[key.Date] = total.Add(decimalItem.PaymentAmount)
			} else {
				dateTotalMap[key.Date] = decimalItem.PaymentAmount
			}
		}

		// 转换为列表并计算占比
		finalList = make([]resp.CompanyPaymentMethodSummaryItem, 0, len(datePaymentDecimalMap))
		for key, decimalItem := range datePaymentDecimalMap {
			// 收集该日期+支付方式涉及的所有门店名称
			companyNames := make([]string, 0, len(datePaymentCompanyNamesMap[key]))
			for name := range datePaymentCompanyNamesMap[key] {
				companyNames = append(companyNames, name)
			}
			sort.Strings(companyNames) // 排序以保证一致性
			companyNamesStr := strings.Join(companyNames, "、")

			datePaymentItem := resp.CompanyPaymentMethodSummaryItem{
				Date:          key.Date,
				CompanyName:   companyNamesStr, // 合并所有涉及的门店名称
				PaymentName:   key.PaymentName,
				PaymentAmount: decimalItem.PaymentAmount.InexactFloat64(),
				PaymentNum:    decimalItem.PaymentNum,
				PaymentRatio:  0.0, // 稍后计算
			}

			// 计算占比：某个支付方式的支付金额 / 该日期所有门店所有支付方式的总金额（实付金额）（使用 decimal）
			totalAmount := dateTotalMap[key.Date] // 该日期所有门店所有支付方式的总金额（实付金额）
			if totalAmount.GreaterThan(decimal.Zero) {
				ratio := decimalItem.PaymentAmount.Div(totalAmount).Mul(decimal.NewFromInt(100)) // 某个支付方式的支付金额 / 实付金额
				datePaymentItem.PaymentRatio = utils.Round(ratio.InexactFloat64(), 2)
			}
			datePaymentItem.PaymentAmount = utils.Round(datePaymentItem.PaymentAmount, 2)
			finalList = append(finalList, datePaymentItem)
		}

		// 按日期、支付方式排序
		sort.Slice(finalList, func(i, j int) bool {
			if finalList[i].Date != finalList[j].Date {
				return finalList[i].Date < finalList[j].Date
			}
			return finalList[i].PaymentName < finalList[j].PaymentName
		})

		// 计算汇总行：按支付方式分组，跨所有门店汇总，每个支付方式一条汇总记录
		// 汇总所有日期的数据，按支付方式分组
		// 按支付方式汇总所有日期的数据（使用 decimal）
		type paymentSummaryDecimal struct {
			PaymentAmount decimal.Decimal
			PaymentNum    int64
			CompanyNames  map[string]bool // 记录该支付方式涉及的所有门店名称
		}
		paymentSummaryDecimalMap := make(map[string]*paymentSummaryDecimal)

		for _, item := range finalList {
			if paymentDecimal, exists := paymentSummaryDecimalMap[item.PaymentName]; exists {
				paymentDecimal.PaymentAmount = paymentDecimal.PaymentAmount.Add(decimal.NewFromFloat(item.PaymentAmount))
				paymentDecimal.PaymentNum += item.PaymentNum
				// 拆分门店名称（可能包含多个门店，用"、"分隔），逐个添加到 map 中自动去重
				companyNames := strings.Split(item.CompanyName, "、")
				for _, name := range companyNames {
					name = strings.TrimSpace(name)
					if name != "" {
						paymentDecimal.CompanyNames[name] = true
					}
				}
			} else {
				// 拆分门店名称（可能包含多个门店，用"、"分隔），逐个添加到 map 中自动去重
				companyNamesMap := make(map[string]bool)
				companyNames := strings.Split(item.CompanyName, "、")
				for _, name := range companyNames {
					name = strings.TrimSpace(name)
					if name != "" {
						companyNamesMap[name] = true
					}
				}
				paymentSummaryDecimalMap[item.PaymentName] = &paymentSummaryDecimal{
					PaymentAmount: decimal.NewFromFloat(item.PaymentAmount),
					PaymentNum:    item.PaymentNum,
					CompanyNames:  companyNamesMap,
				}
			}
		}

		// 计算总支付金额（用于计算占比，使用 decimal）
		// 汇总所有支付方式的总金额，作为实付金额（所有支付方式的总和）
		var totalPaymentAmount decimal.Decimal // 所有支付方式的总金额（实付金额）
		for _, paymentDecimal := range paymentSummaryDecimalMap {
			totalPaymentAmount = totalPaymentAmount.Add(paymentDecimal.PaymentAmount)
		}

		// 转换为汇总行列表并计算占比
		summaryRow = make([]resp.CompanyPaymentMethodSummaryItem, 0, len(paymentSummaryDecimalMap))
		for paymentName, paymentDecimal := range paymentSummaryDecimalMap {
			// 收集该支付方式涉及的所有门店名称
			companyNames := make([]string, 0, len(paymentDecimal.CompanyNames))
			for name := range paymentDecimal.CompanyNames {
				companyNames = append(companyNames, name)
			}
			sort.Strings(companyNames) // 排序以保证一致性
			companyNamesStr := strings.Join(companyNames, "、")

			paymentSummaryItem := resp.CompanyPaymentMethodSummaryItem{
				Date:          "汇总",
				CompanyName:   companyNamesStr, // 合并所有涉及的门店名称
				PaymentName:   paymentName,
				PaymentAmount: paymentDecimal.PaymentAmount.InexactFloat64(),
				PaymentNum:    paymentDecimal.PaymentNum,
				PaymentRatio:  0.0, // 稍后计算
			}

			// 计算占比（使用 decimal）
			// 支付占比 = 某个支付方式的总支付金额 / 所有支付方式的总支付金额（实付金额）
			if totalPaymentAmount.GreaterThan(decimal.Zero) {
				ratio := paymentDecimal.PaymentAmount.Div(totalPaymentAmount).Mul(decimal.NewFromInt(100)) // 某个支付方式的支付金额 / 实付金额
				paymentSummaryItem.PaymentRatio = utils.Round(ratio.InexactFloat64(), 2)
			}
			paymentSummaryItem.PaymentAmount = utils.Round(paymentSummaryItem.PaymentAmount, 2)
			summaryRow = append(summaryRow, paymentSummaryItem)
		}

		// 按支付方式名称排序汇总行
		sort.Slice(summaryRow, func(i, j int) bool {
			return summaryRow[i].PaymentName < summaryRow[j].PaymentName
		})
	} else {
		// 明细表：返回每个营业日每个商家每个支付方式的数据，不包括汇总行（SummaryRow 返回空数组）
		// 按日期、商家名称、支付方式排序
		sort.Slice(allItems, func(i, j int) bool {
			if allItems[i].Date != allItems[j].Date {
				return allItems[i].Date < allItems[j].Date
			}
			if allItems[i].CompanyName != allItems[j].CompanyName {
				return allItems[i].CompanyName < allItems[j].CompanyName
			}
			return allItems[i].PaymentName < allItems[j].PaymentName
		})
		finalList = allItems
		// 明细表返回空数组
		summaryRow = make([]resp.CompanyPaymentMethodSummaryItem, 0)
	}

	// 分页处理
	total := int64(len(finalList))
	pageNo := utils.IfInt(request.PageNo > 0, request.PageNo, 1)
	pageSize := utils.IfInt(request.PageSize > 0, request.PageSize, 20)
	start := (pageNo - 1) * pageSize
	end := start + pageSize

	if start > len(finalList) {
		start = len(finalList)
	}
	if end > len(finalList) {
		end = len(finalList)
	}

	var list []resp.CompanyPaymentMethodSummaryItem
	if start < len(finalList) {
		list = finalList[start:end]
	} else {
		list = make([]resp.CompanyPaymentMethodSummaryItem, 0)
	}

	return &resp.CompanyPaymentMethodSummaryResp{
		Meta: dto.PageResponse{
			PageNo:   pageNo,
			PageSize: pageSize,
			Total:    total,
		},
		List:       list,
		SummaryRow: summaryRow,
	}, nil
}

// countCompanyRefundSummary 获取门店退款金额汇总统计
func (s *businessSrv) countCompanyRefundSummary(ctx context.Context, request req.StatisticsCompanySummaryReq) (*resp.CompanyRefundSummaryResp, error) {
	// 获取当前用户时区
	timezone := ctx.GetCompanySetting().Timezone
	timeUtil := utils.SetTimezone(timezone)

	// 调用 GetCompanyList 获取所有可见商家列表（包含 CompanyUuid 和 CompanyName）
	companyListResp, err := s.GetCompanyList(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取门店列表失败")
	}

	// 处理 CompanyUuids：如果为空，使用所有可见门店；否则过滤出匹配的门店
	var companyList []*resp.CompanySummaryItem
	if len(request.CompanyUuids) == 0 {
		// CompanyUuids 为空时，使用所有可见门店
		companyList = companyListResp.List
	} else {
		// CompanyUuids 不为空时，过滤出匹配的门店（保持顺序）
		companyUuidSet := make(map[uint64]bool)
		for _, uuid := range request.CompanyUuids {
			companyUuidSet[uuid] = true
		}
		companyList = make([]*resp.CompanySummaryItem, 0, len(request.CompanyUuids))
		for _, c := range companyListResp.List {
			if companyUuidSet[c.CompanyUuid] {
				companyList = append(companyList, c)
			}
		}
	}

	// 处理默认值：QueryStartDate 为空时，默认为今日开始日期：YYYY-MM-DD 00:00:00
	queryStartDate := request.QueryStartDate
	if queryStartDate == "" {
		startTime, _ := timeUtil.TodayStartEnd()
		queryStartDate = startTime.Format("2006-01-02 15:04:05")
	}

	// 处理默认值：QueryEndDate 为空时，默认为今日结束日期：YYYY-MM-DD 23:59:59
	queryEndDate := request.QueryEndDate
	if queryEndDate == "" {
		_, endTime := timeUtil.TodayStartEnd()
		queryEndDate = endTime.Format("2006-01-02 15:04:05")
	}

	// 获取数据库管理器
	dbm := database.GetDBManager(config.Database)
	settingSrv := setting.NewSrv(dbm, cache.Global)
	saasDB := dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)

	// 使用 goroutine 并发查询各个门店，提高性能
	// 使用带缓冲的 channel 控制并发数量，最多10个协程，避免资源消耗过大
	maxConcurrent := 10
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// 使用 channel 收集结果，避免竞态条件
	type resultItem struct {
		items     []resp.CompanyRefundSummaryItem
		orderNums map[string]int64 // 日期 -> 订单数，用于汇总时计算退款率
		err       error
	}
	resultChan := make(chan resultItem, len(companyList))

	// 并发查询各个门店
	for _, companyItem := range companyList {
		wg.Add(1)
		utils.Go(func() {
			func(item *resp.CompanySummaryItem) {
				defer wg.Done()

				// 获取信号量，控制并发数
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				companyUuid := item.CompanyUuid
				companyName := item.CompanyName

				// 获取完整的 Company 信息（用于 SetCompany）
				company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
				if err != nil {
					logger.Logger.Warn("获取门店信息失败", zap.Uint64("company_uuid", companyUuid), zap.Error(err))
					resultChan <- resultItem{items: nil, err: err}
					return
				}

				// 获取门店数据库连接
				shopDB := dbm.GetDB(companyUuid)
				if shopDB == nil {
					logger.Logger.Warn("获取门店数据库连接失败", zap.Uint64("company_uuid", companyUuid))
					resultChan <- resultItem{items: nil, err: errors.New("获取门店数据库连接失败")}
					return
				}

				// 创建门店 context
				shopCtx := ctx.Copy()
				shopCtx.SetCompanyUuid(companyUuid)
				shopCtx.SetCompany(*company)
				shopCtx.SetDB(shopDB)

				// 获取门店设置
				companySettingRepo := repository.NewCompanySettingRepo(shopDB)
				shopCompanySetting, err := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
					return db.Where("company_uuid = ?", companyUuid)
				})
				if err != nil {
					logger.Logger.Warn("获取门店设置失败", zap.Uint64("company_uuid", companyUuid), zap.Error(err))
					resultChan <- resultItem{items: nil, err: err}
					return
				}
				shopCtx.SetCompanySetting(shopCompanySetting)
				dataManageSetting := settingSrv.GetDataManageSetting(shopCtx)
				excludeDataManage := shopCompanySetting.IsOpenDataManagement() && dataManageSetting.IsEnableDataManage

				// 调用统计服务获取门店退款数据
				statisticsReq := req.StatisticsSummaryReq{
					PageReq: dto.PageReq{
						PageNo:   1,
						PageSize: 1000, // 获取所有数据，不分页
					},
					QueryStartDate:    queryStartDate,
					QueryEndDate:      queryEndDate,
					Cycle:             request.Cycle, // 使用请求参数中的周期设置
					ExcludeDataManage: excludeDataManage,
				}

				statisticsData := s.statisticsSrv.CountRefundSummary(shopCtx, statisticsReq)

				// 转换为响应格式，同时保存订单数信息
				items := make([]resp.CompanyRefundSummaryItem, 0, len(statisticsData.StatisticsRefundSummaryList))
				orderNums := make(map[string]int64) // 日期 -> 订单数
				for _, statItem := range statisticsData.StatisticsRefundSummaryList {
					items = append(items, resp.CompanyRefundSummaryItem{
						Date:                statItem.Date,
						CompanyName:         companyName, // 使用 GetCompanyList 返回的 CompanyName，避免重复查询
						RefundAmount:        statItem.RefundAmount,
						RefundNum:           statItem.RefundNum,
						RefundRate:          statItem.RefundRate,
						PartialRefundAmount: statItem.PartialRefundAmount,
						PartialRefundNum:    statItem.PartialRefundNum,
						FullRefundAmount:    statItem.FullRefundAmount,
						FullRefundNum:       statItem.FullRefundNum,
					})
					// 保存订单数（按日期累加，因为同一日期可能有多个商家）
					orderNums[statItem.Date] += statItem.OrderNum
				}

				resultChan <- resultItem{items: items, orderNums: orderNums, err: nil}
			}(companyItem)
		})
	}

	// 等待所有 goroutine 完成
	utils.Go(func() {
		wg.Wait()
		close(resultChan)
	})

	// 收集所有门店的统计数据，同时收集订单数信息
	allItems := make([]resp.CompanyRefundSummaryItem, 0)
	dateOrderNumMap := make(map[string]int64) // 日期 -> 订单数总和（所有商家累加）
	for result := range resultChan {
		if result.err != nil {
			// 记录错误但继续处理其他门店
			continue
		}
		allItems = append(allItems, result.items...)
		// 累加订单数
		for date, orderNum := range result.orderNums {
			dateOrderNumMap[date] += orderNum
		}
	}

	// 根据 report 参数决定返回明细表还是汇总表
	// report = 0: 明细表，返回每个营业日每个商家的数据，不包括汇总行（SummaryRow 返回默认值）
	// report = 1: 汇总表，返回每个营业日每个商家数据总和，包括汇总行
	var finalList []resp.CompanyRefundSummaryItem
	var summaryRow resp.CompanyRefundSummaryItem

	if request.Report == 1 {
		// 汇总表：按日期分组汇总
		// 记录每个日期对应的商家名称集合（用于去重）
		dateCompanyNamesMap := make(map[string]map[string]bool)

		// 使用 decimal 进行金额累加，避免精度错误
		type dateSummaryDecimal struct {
			RefundAmount        decimal.Decimal
			RefundNum           int64
			PartialRefundAmount decimal.Decimal
			PartialRefundNum    int64
			FullRefundAmount    decimal.Decimal
			FullRefundNum       int64
			OrderNum            int64 // 用于计算退款率
		}
		dateDecimalMap := make(map[string]*dateSummaryDecimal)

		for _, item := range allItems {
			if dateDecimal, exists := dateDecimalMap[item.Date]; exists {
				// 累加同一天的数据（使用 decimal）
				dateDecimal.RefundAmount = dateDecimal.RefundAmount.Add(decimal.NewFromFloat(item.RefundAmount))
				dateDecimal.RefundNum += item.RefundNum
				dateDecimal.PartialRefundAmount = dateDecimal.PartialRefundAmount.Add(decimal.NewFromFloat(item.PartialRefundAmount))
				dateDecimal.PartialRefundNum += item.PartialRefundNum
				dateDecimal.FullRefundAmount = dateDecimal.FullRefundAmount.Add(decimal.NewFromFloat(item.FullRefundAmount))
				dateDecimal.FullRefundNum += item.FullRefundNum
				// 收集商家名称
				if dateCompanyNamesMap[item.Date] == nil {
					dateCompanyNamesMap[item.Date] = make(map[string]bool)
				}
				dateCompanyNamesMap[item.Date][item.CompanyName] = true
			} else {
				// 创建新的日期汇总项（使用 decimal）
				dateDecimalMap[item.Date] = &dateSummaryDecimal{
					RefundAmount:        decimal.NewFromFloat(item.RefundAmount),
					RefundNum:           item.RefundNum,
					PartialRefundAmount: decimal.NewFromFloat(item.PartialRefundAmount),
					PartialRefundNum:    item.PartialRefundNum,
					FullRefundAmount:    decimal.NewFromFloat(item.FullRefundAmount),
					FullRefundNum:       item.FullRefundNum,
					OrderNum:            dateOrderNumMap[item.Date], // 从订单数 map 中获取
				}
				// 初始化商家名称集合
				if dateCompanyNamesMap[item.Date] == nil {
					dateCompanyNamesMap[item.Date] = make(map[string]bool)
				}
				dateCompanyNamesMap[item.Date][item.CompanyName] = true
			}
		}

		// 转换为列表并计算退款率，设置商家名称
		finalList = make([]resp.CompanyRefundSummaryItem, 0, len(dateDecimalMap))
		for date, dateDecimal := range dateDecimalMap {
			// 计算退款率：退款笔数 / 订单数量 * 100
			var refundRate decimal.Decimal
			if dateDecimal.OrderNum > 0 {
				refundRate = decimal.NewFromInt(dateDecimal.RefundNum).Div(decimal.NewFromInt(dateDecimal.OrderNum)).Mul(decimal.NewFromInt(100))
			}

			dateItem := resp.CompanyRefundSummaryItem{
				Date:                date,
				CompanyName:         "", // 稍后设置
				RefundAmount:        utils.Round(dateDecimal.RefundAmount.InexactFloat64(), 2),
				RefundNum:           dateDecimal.RefundNum,
				RefundRate:          utils.Round(refundRate.InexactFloat64(), 2),
				PartialRefundAmount: utils.Round(dateDecimal.PartialRefundAmount.InexactFloat64(), 2),
				PartialRefundNum:    dateDecimal.PartialRefundNum,
				FullRefundAmount:    utils.Round(dateDecimal.FullRefundAmount.InexactFloat64(), 2),
				FullRefundNum:       dateDecimal.FullRefundNum,
			}

			// 设置商家名称：将所有商家名称用、符号连接
			companyNames := make([]string, 0, len(dateCompanyNamesMap[date]))
			for name := range dateCompanyNamesMap[date] {
				companyNames = append(companyNames, name)
			}
			sort.Strings(companyNames) // 排序以保证一致性
			dateItem.CompanyName = strings.Join(companyNames, "、")
			finalList = append(finalList, dateItem)
		}

		// 按日期排序
		sort.Slice(finalList, func(i, j int) bool {
			return finalList[i].Date < finalList[j].Date
		})

		// 计算总汇总行（使用 decimal）
		var totalRefundAmount, totalPartialRefundAmount, totalFullRefundAmount decimal.Decimal
		var totalRefundNum, totalPartialRefundNum, totalFullRefundNum, totalOrderNum int64

		// 累加所有日期的订单数
		for _, orderNum := range dateOrderNumMap {
			totalOrderNum += orderNum
		}

		for _, item := range finalList {
			totalRefundAmount = totalRefundAmount.Add(decimal.NewFromFloat(item.RefundAmount))
			totalRefundNum += item.RefundNum
			totalPartialRefundAmount = totalPartialRefundAmount.Add(decimal.NewFromFloat(item.PartialRefundAmount))
			totalPartialRefundNum += item.PartialRefundNum
			totalFullRefundAmount = totalFullRefundAmount.Add(decimal.NewFromFloat(item.FullRefundAmount))
			totalFullRefundNum += item.FullRefundNum
		}

		// 计算总退款率：汇总退款订单数量 ÷ 汇总总订单数量 × 100%
		var totalRefundRate decimal.Decimal
		if totalOrderNum > 0 {
			totalRefundRate = decimal.NewFromInt(totalRefundNum).Div(decimal.NewFromInt(totalOrderNum)).Mul(decimal.NewFromInt(100))
		}

		// 收集所有商家名称（用于汇总行）
		allCompanyNamesSet := make(map[string]bool)
		for _, namesMap := range dateCompanyNamesMap {
			for name := range namesMap {
				allCompanyNamesSet[name] = true
			}
		}
		allCompanyNames := make([]string, 0, len(allCompanyNamesSet))
		for name := range allCompanyNamesSet {
			allCompanyNames = append(allCompanyNames, name)
		}
		sort.Strings(allCompanyNames) // 排序以保证一致性

		summaryRow = resp.CompanyRefundSummaryItem{
			Date:                "汇总",
			CompanyName:         strings.Join(allCompanyNames, "、"), // 所有商家名称用、符号连接
			RefundAmount:        utils.Round(totalRefundAmount.InexactFloat64(), 2),
			RefundNum:           totalRefundNum,
			RefundRate:          utils.Round(totalRefundRate.InexactFloat64(), 2),
			PartialRefundAmount: utils.Round(totalPartialRefundAmount.InexactFloat64(), 2),
			PartialRefundNum:    totalPartialRefundNum,
			FullRefundAmount:    utils.Round(totalFullRefundAmount.InexactFloat64(), 2),
			FullRefundNum:       totalFullRefundNum,
		}
	} else {
		// 明细表：返回每个营业日每个商家的数据，不包括汇总行（SummaryRow 返回默认值）
		// 按日期排序
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].Date < allItems[j].Date
		})
		finalList = allItems
		// 明细表返回默认值
		summaryRow = resp.CompanyRefundSummaryItem{}
	}

	// 分页处理
	total := int64(len(finalList))
	pageNo := utils.IfInt(request.PageNo > 0, request.PageNo, 1)
	pageSize := utils.IfInt(request.PageSize > 0, request.PageSize, 20)
	start := (pageNo - 1) * pageSize
	end := start + pageSize

	if start > len(finalList) {
		start = len(finalList)
	}
	if end > len(finalList) {
		end = len(finalList)
	}

	var list []resp.CompanyRefundSummaryItem
	if start < len(finalList) {
		list = finalList[start:end]
	} else {
		list = make([]resp.CompanyRefundSummaryItem, 0)
	}

	return &resp.CompanyRefundSummaryResp{
		Meta: dto.PageResponse{
			PageNo:   pageNo,
			PageSize: pageSize,
			Total:    total,
		},
		List:       list,
		SummaryRow: summaryRow,
	}, nil
}

// ExportCompanyBusinessSummary 导出门店汇总统计
func (s *businessSrv) ExportCompanyBusinessSummary(ctx context.Context, request req.StatisticsCompanySummaryReq) error {
	db := ctx.GetDB()
	// 判断是否还有正在导出的任务
	oldRecord, err := repository.NewExportRecordRepo(db).GetUnfinishedExportRecord(model.ExportTypeCompanyBusinessSummary)
	if err != nil {
		return err
	}
	if oldRecord != nil {
		return errors.WithMessage(errors.New("正在导出,请稍后再操作"))
	}

	// 获取统计数据，检查是否有数据
	result, err := s.CountCompanyBusinessSummary(ctx, request)
	if err != nil {
		return err
	}

	// 根据指标类型判断数据量
	var total int64
	switch request.IndicatorType {
	case "business":
		if resp, ok := result.(*resp.CompanyBusinessSummaryResp); ok {
			total = resp.Meta.Total
		}
	case "payment_method":
		if resp, ok := result.(*resp.CompanyPaymentMethodSummaryResp); ok {
			total = resp.Meta.Total
		}
	case "refund":
		if resp, ok := result.(*resp.CompanyRefundSummaryResp); ok {
			total = resp.Meta.Total
		}
	}

	if total == 0 {
		return errors.WithMessage(errors.New("没有数据需要导出"))
	}
	if total > 1000 {
		return errors.WithMessage(errors.New("请选择具体时间段，最多可导出1000条以下的数据"))
	}

	// 根据指标类型设置文件名
	var fileNameMul model.MultiLanguageName
	switch request.IndicatorType {
	case "business":
		fileNameMul = model.MultiLanguageName{
			EnName:   "Business Data Summary",
			ZhName:   "营业数据汇总",
			ZhTwName: "營業數據匯總",
			ThName:   "สรุปข้อมูลธุรกิจ",
			MyName:   "လုပ်ငန်းဒေတာစုပေါင်းစာရင်း",
			JaName:   "営業データ集計",
			KoName:   "영업 데이터 요약",
			TrName:   "İşletme Verileri Özeti",
			SvName:   "Affärsdata Sammanfattning",
		}
	case "payment_method":
		fileNameMul = model.MultiLanguageName{
			EnName:   "Payment Method Summary",
			ZhName:   "支付方式汇总",
			ZhTwName: "支付方式匯總",
			ThName:   "สรุปวิธีการชำระเงิน",
			MyName:   "ငွေပေးချေမှုနည်းလမ်းစုပေါင်းစာရင်း",
			JaName:   "支払方法集計",
			KoName:   "결제 방식 요약",
			TrName:   "Ödeme Yöntemi Özeti",
			SvName:   "Betalningsmetod Sammanfattning",
		}
	case "refund":
		fileNameMul = model.MultiLanguageName{
			EnName:   "Refund Amount Summary",
			ZhName:   "退款金额汇总",
			ZhTwName: "退款金額匯總",
			ThName:   "สรุปจำนวนเงินคืน",
			MyName:   "ငွေပြန်လည်ပေးအပ်မှုစုပေါင်းစာရင်း",
			JaName:   "返金金額集計",
			KoName:   "환불 금액 요약",
			TrName:   "İade Tutarı Özeti",
			SvName:   "Återbetalningsbelopp Sammanfattning",
		}
	default:
		fileNameMul = model.MultiLanguageName{
			EnName:   "Store Summary Statistics",
			ZhName:   "门店汇总统计",
			ZhTwName: "門店匯總統計",
			ThName:   "สถิติสรุปของร้าน",
			MyName:   "ဆိုင်စုပေါင်းစာရင်း",
			JaName:   "店舗集計統計",
			KoName:   "매장 요약 통계",
			TrName:   "Mağaza Özet İstatistikleri",
			SvName:   "Butikssammanfattningsstatistik",
		}
	}

	// 创建导出任务
	params, err := json.Marshal(request)
	if err != nil {
		return err
	}
	reportName := fileNameMul.GetNameByLang(ctx.GetLanguage())
	fileName, err := s.generateExportFileName(ctx, reportName, model.ExportTypeCompanyBusinessSummary)
	if err != nil {
		return errors.WithMessage(err, "生成文件名失败")
	}
	uuid, _ := utils.GetID()
	record := &model.ExportRecord{
		BaseModel:    model.BaseModel{Uuid: uuid},
		ExportType:   model.ExportTypeCompanyBusinessSummary,
		ExportName:   fileName,
		FileUuid:     0,
		Status:       model.ExportStatusPending,
		ErrorMsg:     "",
		ExportParams: string(params),
		StaffUuid:    ctx.GetStaffUuid(),
	}

	err = repository.NewExportRecordRepo(db).Create(record)
	if err != nil {
		return err
	}

	// 异步处理导出文件的任务
	utils.Go(func() {
		_, err := s.ExportCompanyBusinessSummaryTask(ctx, ExportCompanyBusinessSummaryTaskParams{
			Record:      *record,
			FillNameMul: fileNameMul,
			Request:     request,
		})
		if err != nil {
			if err := repository.NewExportRecordRepo(ctx.GetDB()).Update(record.Uuid, map[string]any{
				"status":    model.ExportStatusFailed,
				"error_msg": err.Error(),
			}); err != nil {
				logger.Logger.Error("导出门店汇总统计数据失败,更新导出记录失败", zap.Error(err), zap.Any("company_uuid", ctx.GetCompanySetting().CompanyUuid), zap.Any("record_uuid", record.Uuid))
			}
			return
		}
	})

	return nil
}

// ExportCompanyBusinessSummaryTaskParams 导出门店汇总统计数据参数
type ExportCompanyBusinessSummaryTaskParams struct {
	Request     req.StatisticsCompanySummaryReq // 请求参数
	Record      model.ExportRecord              // 导出记录
	FillNameMul model.MultiLanguageName         // 多语言名称
}

// ExportCompanyBusinessSummaryTask 导出门店汇总统计数据任务
func (s *businessSrv) ExportCompanyBusinessSummaryTask(ctx context.Context, params ExportCompanyBusinessSummaryTaskParams) (*resp.FileExportResp, error) {
	// 获取统计数据（明细表和汇总表）
	// 明细表：report=0，获取所有数据（不分页）
	detailRequest := params.Request
	detailRequest.Report = 0 // 明细表
	detailRequest.PageNo = 1
	detailRequest.PageSize = 10000 // 设置一个很大的值，确保获取所有数据
	detailResult, err := s.CountCompanyBusinessSummary(ctx, detailRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "获取明细表数据失败")
	}

	// 汇总表：report=1，获取所有数据（不分页）
	summaryRequest := params.Request
	summaryRequest.Report = 1 // 汇总表
	summaryRequest.PageNo = 1
	summaryRequest.PageSize = 10000 // 设置一个很大的值，确保获取所有数据
	summaryResult, err := s.CountCompanyBusinessSummary(ctx, summaryRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "获取汇总表数据失败")
	}

	// 创建Excel文件
	xlsxFile := excelize.NewFile()
	defer xlsxFile.Close()

	// 删除默认的Sheet1（excelize.NewFile()会自动创建一个默认Sheet1，我们需要删除它）
	defaultSheetName := xlsxFile.GetSheetName(0)
	if defaultSheetName != "" {
		xlsxFile.DeleteSheet(defaultSheetName)
	}

	lang := ctx.GetLanguage()

	// 根据指标类型处理导出
	switch params.Request.IndicatorType {
	case "business":
		err = s.exportBusinessSummaryToExcel(xlsxFile, detailResult, summaryResult, lang)
	case "payment_method":
		err = s.exportPaymentMethodSummaryToExcel(xlsxFile, detailResult, summaryResult, lang)
	case "refund":
		err = s.exportRefundSummaryToExcel(xlsxFile, detailResult, summaryResult, lang)
	default:
		err = s.exportBusinessSummaryToExcel(xlsxFile, detailResult, summaryResult, lang)
	}

	if err != nil {
		return nil, errors.WithMessage(err, "生成Excel文件失败")
	}

	// 将 Excel 文件写入内存
	var b bytes.Buffer
	if err := xlsxFile.Write(&b); err != nil {
		return nil, errors.WithMessage(err, "写入Excel文件到内存失败")
	}

	// 上传文档
	res, err := s.uploadFileSrv.UploadDocument(ctx, &b, params.Record.ExportName, int64(b.Len()), 0)
	if err != nil {
		return nil, errors.WithMessage(err, "上传文件失败")
	}

	if err := repository.NewExportRecordRepo(ctx.GetDB()).Update(params.Record.Uuid, map[string]any{
		"file_uuid": res.Uuid,
		"status":    model.ExportStatusSuccess,
	}); err != nil {
		return nil, errors.WithMessage(err, "更新导出记录失败")
	}

	return &resp.FileExportResp{FileUuid: res.Uuid}, nil
}

// exportBusinessSummaryToExcel 导出营业数据汇总到Excel
func (s *businessSrv) exportBusinessSummaryToExcel(xlsxFile *excelize.File, detailResult, summaryResult interface{}, lang string) error {
	detailResp, ok1 := detailResult.(*resp.CompanyBusinessSummaryResp)
	summaryResp, ok2 := summaryResult.(*resp.CompanyBusinessSummaryResp)
	if !ok1 || !ok2 {
		return errors.New("数据类型错误")
	}

	// 表头映射
	headerMap := map[string][]string{
		"zh": { // 中文
			"营业日", "门店名称", "订单金额", "实付金额", "订单量", "用餐人数", "消费桌数", "平均客单价", "订单金额人均", "订单金额单均", "实付金额单均", "点餐订单金额", "桌台订单金额", "外送订单金额",
		},
		"en": { // 英文
			"Business Day", "Store Name", "Order Amount", "Paid Amount", "Order Count", "Number of Diners", "Number of Tables Consumed", "Average Customer Price", "Order Amount Per Person", "Order Amount Per Order", "Paid Amount Per Order", "Meal Order Amount", "Table Order Amount", "Takeout Order Amount",
		},
		"th": { // 泰语
			"วันดำเนินธุรกิจ", "ชื่อร้าน", "ยอดคำสั่งซื้อ", "ยอดชำระเงิน", "จำนวนออเดอร์", "จำนวนลูกค้า", "จำนวนโต๊ะที่ใช้", "ราคาเฉลี่ยต่อลูกค้า", "ยอดคำสั่งซื้อต่อคน", "ยอดคำสั่งซื้อต่อบิล", "ยอดชำระเงินต่อบิล", "ยอดคำสั่งซื้อรับประทานอาหาร", "ยอดคำสั่งซื้อโต๊ะ", "ยอดคำสั่งซื้อนำกลับบ้าน",
		},
		"zhtw": { // 繁体中文
			"營業日", "門店名稱", "訂單金額", "實付金額", "訂單量", "用餐人數", "消費桌數", "平均客單價", "訂單金額人均", "訂單金額單均", "實付金額單均", "點餐訂單金額", "桌台訂單金額", "外送訂單金額",
		},
		"ja": { // 日语
			"営業日", "店舗名", "注文金額", "支払い金額", "注文数", "来店人数", "利用テーブル数", "平均客単価", "一人当たり注文金額", "一件当たり注文金額", "一件当たり支払い金額", "食事注文金額", "テーブル注文金額", "テイクアウト注文金額",
		},
		"ko": { // 韩语
			"영업일", "매장명", "주문 금액", "실제 결제 금액", "주문 건수", "식사 인원", "소비 테이블 수", "평균 고객 단가", "인당 주문 금액", "주문당 주문 금액", "주문당 실제 결제 금액", "식사 주문 금액", "테이블 주문 금액", "포장 주문 금액",
		},
		"my": { // 缅甸语
			"လုပ်ငန်းသက်တမ်းနေ့", "ဆိုင်အမည်", "အော်ဒါစုစုပေါင်းပမာဏ", "ပေးသွင်းငွေ", "အော်ဒါအရေအတွက်", "စားသုံးသူအရေအတွက်", "စားသုံးထားသောစားပွဲအရေအတွက်", "ပျမ်းမျှဖောက်သည်စျေးနှုန်း", "တစ်ဦးလျှင်အော်ဒါပမာဏ", "တစ်ဦးလျှင်အော်ဒါပမာဏ", "တစ်ဦးလျှင်ပေးသွင်းငွေ", "အစားအသောက်အော်ဒါပမာဏ", "စားပွဲအော်ဒါပမာဏ", "ယူဆောင်အော်ဒါပမာဏ",
		},
		"tr": { // 土耳其语
			"İşletme Günü", "Mağaza Adı", "Sipariş Tutarı", "Ödenen Tutar", "Sipariş Sayısı", "Yemek Yiyen Kişi Sayısı", "Tüketilen Masa Sayısı", "Ortalama Müşteri Fiyatı", "Kişi Başına Sipariş Tutarı", "Sipariş Başına Sipariş Tutarı", "Sipariş Başına Ödenen Tutar", "Yemek Sipariş Tutarı", "Masa Sipariş Tutarı", "Paket Sipariş Tutarı",
		},
		"sv": { // 瑞典语
			"Affärsdag", "Butiksnamn", "Orderbelopp", "Betalt Belopp", "Antal Order", "Antal Gäster", "Antal Konsumerade Bord", "Genomsnittligt Kundpris", "Orderbelopp per Person", "Orderbelopp per Order", "Betalt Belopp per Order", "Matbeställningsbelopp", "Bordsorderbelopp", "Takeaway Orderbelopp",
		},
	}

	headers := func() []string {
		h := headerMap[lang]
		if h == nil {
			h = headerMap["en"]
		}
		return h
	}()

	// Sheet1: 明细表
	sheet1Name := "Sheet1"
	sheet1Index, err := xlsxFile.NewSheet(sheet1Name)
	if err != nil {
		return errors.WithMessage(err, "创建明细表Sheet失败")
	}
	xlsxFile.SetActiveSheet(sheet1Index)

	// 写入明细表表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheet1Name, cell, header)
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheet1Name, cell, cell, style)
	}

	// 写入明细表数据
	for rowIdx, item := range detailResp.List {
		offsetRow := rowIdx + 2
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("B%d", offsetRow), item.CompanyName)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("C%d", offsetRow), item.OrderAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("D%d", offsetRow), item.PayAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("E%d", offsetRow), item.OrderNum)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("F%d", offsetRow), item.MealNum)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("G%d", offsetRow), item.DeskNum)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("H%d", offsetRow), item.AvgCustomerPrice)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("I%d", offsetRow), item.OrderAmountMealAvg)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("J%d", offsetRow), item.OrderAmountAvg)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("K%d", offsetRow), item.PayAmountAvg)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("L%d", offsetRow), item.InstantOrderAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("M%d", offsetRow), item.DeskOrderAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("N%d", offsetRow), item.TakeoutOrderAmount)
	}

	// 自动调整明细表列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheet1Name, colName, colName, 20)
	}

	// Sheet2: 汇总表
	sheet2Name := "Sheet2"
	_, err = xlsxFile.NewSheet(sheet2Name)
	if err != nil {
		return errors.WithMessage(err, "创建汇总表Sheet失败")
	}

	// 写入汇总表表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheet2Name, cell, header)
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheet2Name, cell, cell, style)
	}

	// 写入汇总表数据
	for rowIdx, item := range summaryResp.List {
		offsetRow := rowIdx + 2
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("B%d", offsetRow), item.CompanyName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("C%d", offsetRow), item.OrderAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("D%d", offsetRow), item.PayAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("E%d", offsetRow), item.OrderNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("F%d", offsetRow), item.MealNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("G%d", offsetRow), item.DeskNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("H%d", offsetRow), item.AvgCustomerPrice)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("I%d", offsetRow), item.OrderAmountMealAvg)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("J%d", offsetRow), item.OrderAmountAvg)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("K%d", offsetRow), item.PayAmountAvg)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("L%d", offsetRow), item.InstantOrderAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("M%d", offsetRow), item.DeskOrderAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("N%d", offsetRow), item.TakeoutOrderAmount)
	}

	// 写入汇总行
	if summaryResp.SummaryRow.Date != "" || summaryResp.SummaryRow.CompanyName != "" {
		summaryRow := len(summaryResp.List) + 2
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("A%d", summaryRow), summaryResp.SummaryRow.Date)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("B%d", summaryRow), summaryResp.SummaryRow.CompanyName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("C%d", summaryRow), summaryResp.SummaryRow.OrderAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("D%d", summaryRow), summaryResp.SummaryRow.PayAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("E%d", summaryRow), summaryResp.SummaryRow.OrderNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("F%d", summaryRow), summaryResp.SummaryRow.MealNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("G%d", summaryRow), summaryResp.SummaryRow.DeskNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("H%d", summaryRow), summaryResp.SummaryRow.AvgCustomerPrice)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("I%d", summaryRow), summaryResp.SummaryRow.OrderAmountMealAvg)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("J%d", summaryRow), summaryResp.SummaryRow.OrderAmountAvg)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("K%d", summaryRow), summaryResp.SummaryRow.PayAmountAvg)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("L%d", summaryRow), summaryResp.SummaryRow.InstantOrderAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("M%d", summaryRow), summaryResp.SummaryRow.DeskOrderAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("N%d", summaryRow), summaryResp.SummaryRow.TakeoutOrderAmount)
	}

	// 自动调整汇总表列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheet2Name, colName, colName, 20)
	}

	return nil
}

// exportPaymentMethodSummaryToExcel 导出支付方式汇总到Excel
func (s *businessSrv) exportPaymentMethodSummaryToExcel(xlsxFile *excelize.File, detailResult, summaryResult interface{}, lang string) error {
	detailResp, ok1 := detailResult.(*resp.CompanyPaymentMethodSummaryResp)
	summaryResp, ok2 := summaryResult.(*resp.CompanyPaymentMethodSummaryResp)
	if !ok1 || !ok2 {
		return errors.New("数据类型错误")
	}

	// 表头映射
	headerMap := map[string][]string{
		"zh": { // 中文
			"营业日", "门店名称", "支付方式", "支付金额", "支付笔数", "支付占比",
		},
		"en": { // 英文
			"Business Day", "Store Name", "Payment Method", "Payment Amount", "Payment Count", "Payment Ratio",
		},
		"th": { // 泰语
			"วันดำเนินธุรกิจ", "ชื่อร้าน", "วิธีการชำระเงิน", "จำนวนเงินที่ชำระ", "จำนวนครั้งที่ชำระ", "สัดส่วนการชำระเงิน",
		},
		"zhtw": { // 繁体中文
			"營業日", "門店名稱", "支付方式", "支付金額", "支付筆數", "支付占比",
		},
		"ja": { // 日语
			"営業日", "店舗名", "支払方法", "支払金額", "支払回数", "支払比率",
		},
		"ko": { // 韩语
			"영업일", "매장명", "결제 방식", "결제 금액", "결제 건수", "결제 비율",
		},
		"my": { // 缅甸语
			"လုပ်ငန်းသက်တမ်းနေ့", "ဆိုင်အမည်", "ငွေပေးချေမှုနည်းလမ်း", "ငွေပေးချေမှုပမာဏ", "ငွေပေးချေမှုအကြိမ်အရေအတွက်", "ငွေပေးချေမှုရာခိုင်နှုန်း",
		},
		"tr": { // 土耳其语
			"İşletme Günü", "Mağaza Adı", "Ödeme Yöntemi", "Ödeme Tutarı", "Ödeme Sayısı", "Ödeme Oranı",
		},
		"sv": { // 瑞典语
			"Affärsdag", "Butiksnamn", "Betalningsmetod", "Betalningsbelopp", "Betalningsantal", "Betalningsandel",
		},
	}

	headers := func() []string {
		h := headerMap[lang]
		if h == nil {
			h = headerMap["en"]
		}
		return h
	}()

	// Sheet1: 明细表
	sheet1NameMul := model.MultiLanguageName{
		EnName:   "Details",
		ZhName:   "明细表",
		ZhTwName: "明細表",
		ThName:   "รายละเอียด",
		MyName:   "အသေးစား",
		JaName:   "明細表",
		KoName:   "상세",
		TrName:   "Detaylar",
		SvName:   "Detaljer",
	}
	sheet1Name := sheet1NameMul.GetNameByLang(lang)
	sheet1Index, err := xlsxFile.NewSheet(sheet1Name)
	if err != nil {
		return errors.WithMessage(err, "创建明细表Sheet失败")
	}
	xlsxFile.SetActiveSheet(sheet1Index)

	// 写入明细表表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheet1Name, cell, header)
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheet1Name, cell, cell, style)
	}

	// 写入明细表数据
	for rowIdx, item := range detailResp.List {
		offsetRow := rowIdx + 2
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("B%d", offsetRow), item.CompanyName)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("C%d", offsetRow), item.PaymentName)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("D%d", offsetRow), item.PaymentAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("E%d", offsetRow), item.PaymentNum)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("F%d", offsetRow), item.PaymentRatio)
	}

	// 自动调整明细表列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheet1Name, colName, colName, 20)
	}

	// Sheet2: 汇总表
	sheet2NameMul := model.MultiLanguageName{
		EnName:   "Summary",
		ZhName:   "汇总表",
		ZhTwName: "匯總表",
		ThName:   "สรุป",
		MyName:   "စုပေါင်းစာရင်း",
		JaName:   "集計表",
		KoName:   "요약",
		TrName:   "Özet",
		SvName:   "Sammanfattning",
	}
	sheet2Name := sheet2NameMul.GetNameByLang(lang)
	_, err = xlsxFile.NewSheet(sheet2Name)
	if err != nil {
		return errors.WithMessage(err, "创建汇总表Sheet失败")
	}

	// 写入汇总表表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheet2Name, cell, header)
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheet2Name, cell, cell, style)
	}

	// 写入汇总表数据
	for rowIdx, item := range summaryResp.List {
		offsetRow := rowIdx + 2
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("B%d", offsetRow), item.CompanyName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("C%d", offsetRow), item.PaymentName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("D%d", offsetRow), item.PaymentAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("E%d", offsetRow), item.PaymentNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("F%d", offsetRow), item.PaymentRatio)
	}

	// 写入汇总行（支付方式汇总的汇总行是数组）
	for rowIdx, item := range summaryResp.SummaryRow {
		summaryRow := len(summaryResp.List) + 2 + rowIdx
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("A%d", summaryRow), item.Date)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("B%d", summaryRow), item.CompanyName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("C%d", summaryRow), item.PaymentName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("D%d", summaryRow), item.PaymentAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("E%d", summaryRow), item.PaymentNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("F%d", summaryRow), item.PaymentRatio)
	}

	// 自动调整汇总表列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheet2Name, colName, colName, 20)
	}

	return nil
}

// exportRefundSummaryToExcel 导出退款金额汇总到Excel
func (s *businessSrv) exportRefundSummaryToExcel(xlsxFile *excelize.File, detailResult, summaryResult interface{}, lang string) error {
	detailResp, ok1 := detailResult.(*resp.CompanyRefundSummaryResp)
	summaryResp, ok2 := summaryResult.(*resp.CompanyRefundSummaryResp)
	if !ok1 || !ok2 {
		return errors.New("数据类型错误")
	}

	// 表头映射
	headerMap := map[string][]string{
		"zh": { // 中文
			"营业日", "门店名称", "退款金额", "退款笔数", "退款率", "部分退款金额", "部分退款笔数", "整单退款金额", "整单退款笔数",
		},
		"en": { // 英文
			"Business Day", "Store Name", "Refund Amount", "Refund Count", "Refund Rate", "Partial Refund Amount", "Partial Refund Count", "Full Refund Amount", "Full Refund Count",
		},
		"th": { // 泰语
			"วันดำเนินธุรกิจ", "ชื่อร้าน", "จำนวนเงินคืน", "จำนวนครั้งที่คืน", "อัตราการคืนเงิน", "จำนวนเงินคืนบางส่วน", "จำนวนครั้งที่คืนบางส่วน", "จำนวนเงินคืนเต็มจำนวน", "จำนวนครั้งที่คืนเต็มจำนวน",
		},
		"zhtw": { // 繁体中文
			"營業日", "門店名稱", "退款金額", "退款筆數", "退款率", "部分退款金額", "部分退款筆數", "整單退款金額", "整單退款筆數",
		},
		"ja": { // 日语
			"営業日", "店舗名", "返金金額", "返金回数", "返金率", "一部返金金額", "一部返金回数", "全額返金金額", "全額返金回数",
		},
		"ko": { // 韩语
			"영업일", "매장명", "환불 금액", "환불 건수", "환불률", "부분 환불 금액", "부분 환불 건수", "전액 환불 금액", "전액 환불 건수",
		},
		"my": { // 缅甸语
			"လုပ်ငန်းသက်တမ်းနေ့", "ဆိုင်အမည်", "ငွေပြန်လည်ပေးအပ်မှုပမာဏ", "ငွေပြန်လည်ပေးအပ်မှုအကြိမ်အရေအတွက်", "ငွေပြန်လည်ပေးအပ်မှုရာခိုင်နှုန်း", "အပိုင်းငွေပြန်လည်ပေးအပ်မှုပမာဏ", "အပိုင်းငွေပြန်လည်ပေးအပ်မှုအကြိမ်အရေအတွက်", "အပြည့်အဝငွေပြန်လည်ပေးအပ်မှုပမာဏ", "အပြည့်အဝငွေပြန်လည်ပေးအပ်မှုအကြိမ်အရေအတွက်",
		},
		"tr": { // 土耳其语
			"İşletme Günü", "Mağaza Adı", "İade Tutarı", "İade Sayısı", "İade Oranı", "Kısmi İade Tutarı", "Kısmi İade Sayısı", "Tam İade Tutarı", "Tam İade Sayısı",
		},
		"sv": { // 瑞典语
			"Affärsdag", "Butiksnamn", "Återbetalningsbelopp", "Återbetalningsantal", "Återbetalningsandel", "Delvis återbetalningsbelopp", "Delvis återbetalningsantal", "Fullt återbetalningsbelopp", "Fullt återbetalningsantal",
		},
	}

	headers := func() []string {
		h := headerMap[lang]
		if h == nil {
			h = headerMap["en"]
		}
		return h
	}()

	// Sheet1: 明细表
	sheet1Name := "Sheet1"
	sheet1Index, err := xlsxFile.NewSheet(sheet1Name)
	if err != nil {
		return errors.WithMessage(err, "创建明细表Sheet失败")
	}
	xlsxFile.SetActiveSheet(sheet1Index)

	// 写入明细表表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheet1Name, cell, header)
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheet1Name, cell, cell, style)
	}

	// 写入明细表数据
	for rowIdx, item := range detailResp.List {
		offsetRow := rowIdx + 2
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("B%d", offsetRow), item.CompanyName)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("C%d", offsetRow), item.RefundAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("D%d", offsetRow), item.RefundNum)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("E%d", offsetRow), item.RefundRate)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("F%d", offsetRow), item.PartialRefundAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("G%d", offsetRow), item.PartialRefundNum)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("H%d", offsetRow), item.FullRefundAmount)
		xlsxFile.SetCellValue(sheet1Name, fmt.Sprintf("I%d", offsetRow), item.FullRefundNum)
	}

	// 自动调整明细表列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheet1Name, colName, colName, 20)
	}

	// Sheet2: 汇总表
	sheet2Name := "Sheet2"
	_, err = xlsxFile.NewSheet(sheet2Name)
	if err != nil {
		return errors.WithMessage(err, "创建汇总表Sheet失败")
	}

	// 写入汇总表表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheet2Name, cell, header)
		style, _ := xlsxFile.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		xlsxFile.SetCellStyle(sheet2Name, cell, cell, style)
	}

	// 写入汇总表数据
	for rowIdx, item := range summaryResp.List {
		offsetRow := rowIdx + 2
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("A%d", offsetRow), item.Date)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("B%d", offsetRow), item.CompanyName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("C%d", offsetRow), item.RefundAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("D%d", offsetRow), item.RefundNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("E%d", offsetRow), item.RefundRate)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("F%d", offsetRow), item.PartialRefundAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("G%d", offsetRow), item.PartialRefundNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("H%d", offsetRow), item.FullRefundAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("I%d", offsetRow), item.FullRefundNum)
	}

	// 写入汇总行
	if summaryResp.SummaryRow.Date != "" || summaryResp.SummaryRow.CompanyName != "" {
		summaryRow := len(summaryResp.List) + 2
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("A%d", summaryRow), summaryResp.SummaryRow.Date)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("B%d", summaryRow), summaryResp.SummaryRow.CompanyName)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("C%d", summaryRow), summaryResp.SummaryRow.RefundAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("D%d", summaryRow), summaryResp.SummaryRow.RefundNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("E%d", summaryRow), summaryResp.SummaryRow.RefundRate)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("F%d", summaryRow), summaryResp.SummaryRow.PartialRefundAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("G%d", summaryRow), summaryResp.SummaryRow.PartialRefundNum)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("H%d", summaryRow), summaryResp.SummaryRow.FullRefundAmount)
		xlsxFile.SetCellValue(sheet2Name, fmt.Sprintf("I%d", summaryRow), summaryResp.SummaryRow.FullRefundNum)
	}

	// 自动调整汇总表列宽
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		xlsxFile.SetColWidth(sheet2Name, colName, colName, 20)
	}

	return nil
}
