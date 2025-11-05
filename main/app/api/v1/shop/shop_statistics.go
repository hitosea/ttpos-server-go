package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type statisticsHandler struct {
	businessSrv service.IBusinessSrv
}

// CountBusiness 统计营业数据，移动管理端首页-店内概况
// @Summary 统计营业数据，移动管理端首页-店内概况
// @Description 统计营业数据，移动管理端首页-店内概况
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataAll} "统计数据"
// @Router /shop/statistics/business [get]
func (h *statisticsHandler) CountBusiness(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	businessData, err := h.businessSrv.CountBusiness(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, businessData)
}

// CountPaymentMethod 统计支付方式
// @Summary 统计支付方式
// @Description 统计支付方式
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataPaymentMethod} "统计数据"
// @Router /shop/statistics/payment_method [get]
func (h *statisticsHandler) CountPaymentMethod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	paymentMethodData, err := h.businessSrv.CountPaymentMethod(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, paymentMethodData)
}

// CountProductCategory 统计商品分类
// @Summary 统计商品分类
// @Description 统计商品分类
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataProductCategory} "统计数据"
// @Router /shop/statistics/product_category [get]
func (h *statisticsHandler) CountProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	productCategoryData, err := h.businessSrv.CountProductCategory(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, productCategoryData)
}

// CountProduct 统计商品
// @Summary 统计商品
// @Description 统计商品
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataProduct} "统计数据"
// @Router /shop/statistics/product [get]
func (h *statisticsHandler) CountProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	productData, err := h.businessSrv.CountProduct(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, productData)
}

// CountArea 统计区域，移动管理端首页-区域数据
// @Summary 统计区域，移动管理端首页-区域数据
// @Description 统计区域，移动管理端首页-区域数据
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataArea} "统计数据"
// @Router /shop/statistics/area [get]
func (h *statisticsHandler) CountArea(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	areaData, err := h.businessSrv.CountArea(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, areaData)
}

// CountProductRank 统计商品排行，移动管理端首页-销量、销售额排行
// @Summary 统计商品排行，移动管理端首页-销量、销售额排行
// @Description 统计商品排行，移动管理端首页-销量、销售额排行
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataRankProductReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataProductRank} "统计数据"
// @Router /shop/statistics/product_rank [get]
func (h *statisticsHandler) CountProductRank(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataRankProductReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	productRankData, err := h.businessSrv.RankProduct(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, productRankData)
}

// CountProductSales 统计商品销售
// @Summary 统计商品销售
// @Description 统计商品销售
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountProductSalesReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataCountProductSalesPagination} "统计数据"
// @Router /shop/statistics/product_sales [get]
func (h *statisticsHandler) CountProductSales(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountProductSalesReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	productSalesData, err := h.businessSrv.CountProductSales(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, productSalesData)
}

// Count7Days 统计7天
// @Summary 统计7天
// @Description 统计7天
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataCount7Days} "统计数据"
// @Router /shop/statistics/7days [get]
func (h *statisticsHandler) Count7Days(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	sevenDaysData, err := h.businessSrv.Count7Days(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, sevenDaysData)
}

// CountExport 统计导出
// @Summary 统计导出
// @Description 统计导出
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataExport} "统计数据"
// @Router /shop/statistics/export [get]
func (h *statisticsHandler) CountExport(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	exportData, err := h.businessSrv.CountExport(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, exportData)
}

// CountShiftRefundAmount 统计班次退款金额
// @Summary 统计班次退款金额
// @Description 统计班次退款金额
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataShiftRefundAmount} "统计数据"
// @Router /shop/statistics/shift_refund_amount [get]
func (h *statisticsHandler) CountShiftRefundAmount(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	refundAmount := h.businessSrv.CountShiftRefundAmount(ctx, countReq)

	helper.Success(c, refundAmount)
}

// CountHome 统计首页
// @Summary 统计首页
// @Description 统计首页
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataHome} "统计数据"
// @Router /shop/statistics/home [get]
func (h *statisticsHandler) CountHome(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	homeData, err := h.businessSrv.CountHome(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, homeData)
}

// CountKitchenEfficiencyAnalysis 统计后厨效率分析
// @Summary 统计后厨效率分析
// @Description 统计后厨效率分析
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KitchenEfficiencyAnalysisReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataKitchenEfficiencyAnalysis} "统计数据"
// @Router /shop/statistics/kitchen/efficiency_analysis [get]
func (h *statisticsHandler) CountKitchenEfficiencyAnalysis(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.KitchenEfficiencyAnalysisReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	kitchenEfficiencyAnalysisData, err := h.businessSrv.CountKitchenEfficiencyAnalysis(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, kitchenEfficiencyAnalysisData)
}

// CountKitchenEfficiencyAnalysisAvg 统计后厨效率分析平均时长
// @Summary 统计后厨效率分析平均时长
// @Description 统计后厨效率分析平均时长
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KitchenEfficiencyAnalysisAvgReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataKitchenEfficiencyAnalysisAvg} "统计数据"
// @Router /shop/statistics/kitchen/efficiency_analysis/avg [get]
func (h *statisticsHandler) CountKitchenEfficiencyAnalysisAvg(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.KitchenEfficiencyAnalysisAvgReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	kitchenEfficiencyAnalysisAvgData, err := h.businessSrv.CountKitchenEfficiencyAnalysisAvg(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, kitchenEfficiencyAnalysisAvgData)
}

// CountKitchenProductionDetail 统计后厨菜品出品明细
// @Summary 统计后厨菜品出品明细
// @Description 统计后厨菜品出品明细
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KitchenProductionDetailReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataKitchenProductionDetail} "统计数据"
// @Router /shop/statistics/kitchen/production_detail [get]
func (h *statisticsHandler) CountKitchenProductionDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.KitchenProductionDetailReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	kitchenProductionDetailData, err := h.businessSrv.CountKitchenProductionDetail(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, kitchenProductionDetailData)
}

func RegisterStatisticsHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	businessSrv := service.NewBusinessSrv(statisticsSrv)

	wrapper := &statisticsHandler{
		businessSrv: businessSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/statistics/business", wrapper.CountBusiness)                                            // 统计营业数据，移动管理端首页-店内概况
		privateApi.GET("/statistics/payment_method", wrapper.CountPaymentMethod)                                 // 统计支付方式
		privateApi.GET("/statistics/product_category", wrapper.CountProductCategory)                             // 统计商品分类
		privateApi.GET("/statistics/product", wrapper.CountProduct)                                              // 统计商品
		privateApi.GET("/statistics/area", wrapper.CountArea)                                                    // 统计区域，移动管理端首页-区域数据
		privateApi.GET("/statistics/product_rank", wrapper.CountProductRank)                                     // 统计商品排行，移动管理端首页-销量、销售额排行
		privateApi.GET("/statistics/product_sales", wrapper.CountProductSales)                                   // 统计商品销售
		privateApi.GET("/statistics/7days", wrapper.Count7Days)                                                  // 统计7天
		privateApi.GET("/statistics/export", wrapper.CountExport)                                                // 统计导出
		privateApi.GET("/statistics/shift_refund_amount", wrapper.CountShiftRefundAmount)                        // 统计班次退款金额
		privateApi.GET("/statistics/home", wrapper.CountHome)                                                    // 统计首页
		privateApi.GET("/statistics/kitchen/efficiency_analysis", wrapper.CountKitchenEfficiencyAnalysis)        // 统计后厨效率分析
		privateApi.GET("/statistics/kitchen/efficiency_analysis/avg", wrapper.CountKitchenEfficiencyAnalysisAvg) // 统计后厨效率分析平均时长
		privateApi.GET("/statistics/kitchen/production_detail", wrapper.CountKitchenProductionDetail)            // 后厨菜品出品明细
	}
}
