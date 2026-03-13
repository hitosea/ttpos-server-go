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
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type statisticsHandler struct {
	businessSrv service.IBusinessSrv
	settingSrv  setting.ISrv
	dbm         *database.DBManager
}

// switchCompanyContext 切换到指定门店的数据库上下文
// 如果 companyUuid 为空（0），返回原 ctx；否则校验权限并切换数据库
func (h *statisticsHandler) switchCompanyContext(ctx context.Context, companyUuid uint64) (context.Context, error) {
	if companyUuid == 0 {
		return ctx, nil
	}

	// 权限校验：检查用户是否有权限访问该门店
	companyListResp, err := h.businessSrv.GetCompanyList(ctx)
	if err != nil {
		return nil, errors.New("获取门店列表失败")
	}

	hasPermission := false
	for _, c := range companyListResp.List {
		if c.CompanyUuid == companyUuid {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, errors.New("无权限访问该门店")
	}

	// 切换数据库连接
	shopDB := h.dbm.GetDB(companyUuid)
	if shopDB == nil {
		return nil, errors.New("门店数据库连接失败")
	}

	// 创建新的 context 并设置数据库连接和门店UUID
	newCtx := ctx.Copy()
	newCtx.SetDB(shopDB)
	newCtx.SetCompanyUuid(companyUuid)

	// 重新加载目标门店的 CompanySetting
	companySetting, err := h.settingSrv.GetCompanySetting(newCtx)
	if err != nil {
		return nil, errors.New("获取门店设置失败")
	}
	newCtx.SetCompanySetting(companySetting)

	return newCtx, nil
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

	// 切换门店上下文（如果指定了 company_uuid）
	targetCtx, err := h.switchCompanyContext(ctx, countReq.CompanyUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 判断数据管理功能是否开启
	companySetting := targetCtx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(targetCtx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	businessData, err := h.businessSrv.CountBusiness(targetCtx, countReq)
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
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
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
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
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

	// 切换门店上下文（如果指定了 company_uuid）
	targetCtx, err := h.switchCompanyContext(ctx, countReq.CompanyUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 判断数据管理功能是否开启
	companySetting := targetCtx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(targetCtx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	areaData, err := h.businessSrv.CountArea(targetCtx, countReq)
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

	// 切换门店上下文（如果指定了 company_uuid）
	targetCtx, err := h.switchCompanyContext(ctx, countReq.CompanyUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 判断数据管理功能是否开启
	companySetting := targetCtx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(targetCtx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	productRankData, err := h.businessSrv.RankProduct(targetCtx, countReq)
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
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
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

// CountShiftPeakHour 统计班次高峰期数据
// @Summary 统计班次高峰期数据
// @Description 统计班次高峰期数据
// @Tags 商家端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataShiftPeakHour} "统计数据"
// @Router /shop/statistics/shift_peak_hour [get]
func (h *statisticsHandler) CountShiftPeakHour(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	peakHourData := h.businessSrv.CountShiftPeakHour(ctx, countReq)

	helper.Success(c, peakHourData)
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
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
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
// @Tags 商家端.报表
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
// @Tags 商家端.报表
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
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KitchenProductionDetailReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.KitchenProductionDetail} "统计数据"
// @Router /shop/statistics/kitchen/production_detail [get]
func (h *statisticsHandler) CountKitchenProductionDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.KitchenProductionDetailReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	kitchenProductionDetailData, err := h.businessSrv.KitchenProductionDetail(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, kitchenProductionDetailData)
}

// ExportKitchenProductionDetail 导出后厨菜品出品明细
// @Summary 导出后厨菜品出品明细
// @Description 导出后厨菜品出品明细
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KitchenProductionDetailReq true "统计参数"
// @Success 200 {object} dto.Response{data=resp.FileExportResp} "统计数据"
// @Router /shop/statistics/kitchen/production_detail/export [get]
func (h *statisticsHandler) ExportKitchenProductionDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.KitchenProductionDetailReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	err := h.businessSrv.ExportKitchenProductionDetail(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// ExportProductSales 导出商品销售统计
// @Summary 导出商品销售统计
// @Description 导出商品销售统计数据
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountProductSalesReq true "统计参数"
// @Success 200 {object} dto.Response "导出成功"
// @Router /shop/statistics/product_sales/export [get]
func (h *statisticsHandler) ExportProductSales(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountProductSalesReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	err := h.businessSrv.ExportProductSales(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// CountBusinessTimePeriod 统计营业时段数据
// @Summary 统计营业时段数据
// @Description 移动端-报表-营业报表-时段营业统计
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessTimePeriodReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessTimePeriod} "统计数据"
// @Router /shop/statistics/business/time_period [get]
func (h *statisticsHandler) CountBusinessTimePeriod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.BusinessTimePeriodReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	req.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	businessTimePeriodData := h.businessSrv.CountBusinessTimePeriod(ctx, req)
	helper.Success(c, businessTimePeriodData)
}

// CountBusinessComprehensive 统计综合运营数据
// @Summary 统计综合运营数据
// @Description 移动端-报表-营业报表-综合运营统计
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.StatisticsSummaryReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.StatisticsSummary} "统计数据"
// @Router /shop/statistics/business/summary [get]
func (h *statisticsHandler) CountBusinessSummary(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.StatisticsSummaryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	req.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	businessSummary := h.businessSrv.CountBusinessSummary(ctx, req)
	helper.Success(c, businessSummary)
}

// CountBusinessPaymentMethod 统计营业收款数据
// @Summary 统计营业收款数据
// @Description 统计营业收款数据
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.StatisticsPaymentMethodReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.StatisticsPaymentMethod} "统计数据"
// @Router /shop/statistics/business/payment_method [get]
func (h *statisticsHandler) CountBusinessPaymentMethod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.StatisticsPaymentMethodReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	req.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	businessPaymentMethodData := h.businessSrv.CountBusinessPaymentMethod(ctx, req)
	helper.Success(c, businessPaymentMethodData)
}

// ExportKitchenEfficiencyAnalysis 导出后厨效率分析
// @Summary 导出后厨效率分析
// @Description 导出后厨效率分析
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KitchenEfficiencyAnalysisReq true "统计参数"
// @Success 200 {object} dto.Response{data=resp.FileExportResp} "统计数据"
// @Router /shop/statistics/kitchen/efficiency_analysis/export [get]
func (h *statisticsHandler) ExportKitchenEfficiencyAnalysis(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.KitchenEfficiencyAnalysisReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	err := h.businessSrv.ExportKitchenEfficiencyAnalysis(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// ExportBusinessTimePeriod 导出营业时段数据
// @Summary 导出营业时段数据
// @Description 移动端-报表-营业报表-导出时段营业统计
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessTimePeriodReq true "统计参数"
// @Success 200 {object} dto.Response "统计数据"
// @Router /shop/statistics/business/time_period/export [get]
func (h *statisticsHandler) ExportBusinessTimePeriod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessTimePeriodReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	err := h.businessSrv.ExportBusinessTimePeriod(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// ExportBusinessSummary 导出综合运营统计
// @Summary 导出综合运营统计
// @Description 导出综合运营统计
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.StatisticsSummaryReq true "统计参数"
// @Success 200 {object} dto.Response "统计数据"
// @Router /shop/statistics/business/summary/export [get]
func (h *statisticsHandler) ExportBusinessSummary(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.StatisticsSummaryReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	err := h.businessSrv.ExportBusinessSummary(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// ExportBusinessPaymentMethod 导出营业收款统计
// @Summary 导出营业收款统计
// @Description 导出营业收款统计
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.StatisticsPaymentMethodReq true "统计参数"
// @Success 200 {object} dto.Response "统计数据"
// @Router /shop/statistics/business/payment_method/export [get]
func (h *statisticsHandler) ExportBusinessPaymentMethod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.StatisticsPaymentMethodReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	err := h.businessSrv.ExportBusinessPaymentMethod(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// ChannelSales 渠道营业统计查询
// @Summary 渠道营业统计查询
// @Description 渠道营业统计查询
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param start_time query int64 false "开始时间戳（Unix秒）"
// @param end_time query int64 false "结束时间戳（Unix秒）"
// @Success 200 {object} dto.Response{data=resp.ChannelSalesResp} "统计数据"
// @Router /shop/statistics/channel_sales [get]
func (h *statisticsHandler) ChannelSales(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.ChannelSalesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	req.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	resp, err := h.businessSrv.CountChannelSales(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp)
}

// ExportChannelSales 导出渠道营业统计
// @Summary 导出渠道营业统计
// @Description 导出渠道营业统计
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param start_time query int64 false "开始时间戳（Unix秒）"
// @param end_time query int64 false "结束时间戳（Unix秒）"
// @Success 200 {object} dto.Response "导出任务已创建"
// @Router /shop/statistics/channel_sales/export [get]
func (h *statisticsHandler) ExportChannelSales(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.ChannelSalesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	req.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	err := h.businessSrv.ExportChannelSales(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// UserAnalysis 用户分析统计查询
// @Summary 用户分析统计查询
// @Description 用户分析统计查询
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param start_time query int64 false "开始时间戳（Unix秒）"
// @param end_time query int64 false "结束时间戳（Unix秒）"
// @Success 200 {object} dto.Response{data=resp.UserAnalysisResp} "统计数据"
// @Router /shop/statistics/user_analysis [get]
func (h *statisticsHandler) UserAnalysis(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.UserAnalysisReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	req.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	resp, err := h.businessSrv.CountUserAnalysis(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp)
}

// ExportUserAnalysis 导出用户分析统计
// @Summary 导出用户分析统计
// @Description 导出用户分析统计
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param start_time query int64 false "开始时间戳（Unix秒）"
// @param end_time query int64 false "结束时间戳（Unix秒）"
// @Success 200 {object} dto.Response "导出任务已创建"
// @Router /shop/statistics/user_analysis/export [get]
func (h *statisticsHandler) ExportUserAnalysis(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.UserAnalysisReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 判断数据管理功能是否开启
	companySetting := ctx.GetCompanySetting()
	dataSetting := h.settingSrv.GetDataManageSetting(ctx)
	req.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	err := h.businessSrv.ExportUserAnalysis(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// GetCompanyList 获取门店列表
// @Summary 获取门店列表
// @Description 获取门店汇总统计可选择的门店列表（总店返回本店及下级所有子店，子店返回本店及已授权的其他门店）
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CompanySummaryListResp} "门店列表"
// @Router /shop/statistics/company_list [get]
func (h *statisticsHandler) GetCompanyList(c *gin.Context) {
	ctx := helper.GetContext(c)
	resp, err := h.businessSrv.GetCompanyList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp)
}

// GetCompanyPaymentMethods 获取门店支付方式列表
// @Summary 获取门店支付方式列表
// @Description 获取有权限的所有门店的支付方式，汇总去重后返回
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CompanyPaymentMethodListResp} "支付方式列表"
// @Router /shop/statistics/company/payment_methods [get]
func (h *statisticsHandler) GetCompanyPaymentMethods(c *gin.Context) {
	ctx := helper.GetContext(c)
	resp, err := h.businessSrv.GetCompanyPaymentMethods(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, resp)
}

// CountCompanyBusinessSummary 获取门店汇总统计
// @Summary 获取门店汇总统计
// @Description 新管理端-报表-门店汇总统计。支持营业数据汇总（indicator_type=business）、支付方式汇总（indicator_type=payment_method）和退款金额汇总（indicator_type=refund）
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.StatisticsCompanySummaryReq true "统计参数"
// @Success 200 {object} dto.Response{data=resp.CompanyBusinessSummaryResp} "统计数据。indicator_type=business 返回 CompanyBusinessSummaryResp"
// @Success 201 {object} dto.Response{data=resp.CompanyPaymentMethodSummaryResp} "统计数据。indicator_type=payment_method 返回 CompanyPaymentMethodSummaryResp"
// @Success 202 {object} dto.Response{data=resp.CompanyRefundSummaryResp} "统计数据。indicator_type=refund 返回 CompanyRefundSummaryResp"
// @Router /shop/statistics/company/business/summary [get]
func (h *statisticsHandler) CountCompanyBusinessSummary(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.StatisticsCompanySummaryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	businessSummary, err := h.businessSrv.CountCompanyBusinessSummary(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, businessSummary)
}

// ExportCompanyBusinessSummary 导出门店汇总统计
// @Summary 导出门店汇总统计
// @Description 新管理端-报表-门店汇总统计导出。支持营业数据汇总（indicator_type=business）、支付方式汇总（indicator_type=payment_method）和退款金额汇总（indicator_type=refund）
// @Tags 商家端.报表
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.StatisticsCompanySummaryReq true "统计参数"
// @Success 200 {object} dto.Response "导出任务已创建"
// @Router /shop/statistics/company/business/summary/export [get]
func (h *statisticsHandler) ExportCompanyBusinessSummary(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req req.StatisticsCompanySummaryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	err := h.businessSrv.ExportCompanyBusinessSummary(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
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
	uploadFileSrv := service.NewUploadFileSrv(dbm)
	businessSrv := service.NewBusinessSrv(statisticsSrv, uploadFileSrv)

	wrapper := &statisticsHandler{
		businessSrv: businessSrv,
		settingSrv:  settingSrv,
		dbm:         dbm,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/statistics/payment_method", wrapper.CountPaymentMethod)                                  // 统计支付方式
		privateApi.GET("/statistics/product_category", wrapper.CountProductCategory)                              // 统计商品分类
		privateApi.GET("/statistics/product", wrapper.CountProduct)                                               // 统计商品
		privateApi.GET("/statistics/product_sales", wrapper.CountProductSales)                                    // 统计商品销售
		privateApi.GET("/statistics/product_sales/export", wrapper.ExportProductSales)                            // 导出商品销售统计
		privateApi.GET("/statistics/7days", wrapper.Count7Days)                                                   // 统计7天
		privateApi.GET("/statistics/export", wrapper.CountExport)                                                 // 统计导出
		privateApi.GET("/statistics/shift_refund_amount", wrapper.CountShiftRefundAmount)                         // 统计班次退款金额
		privateApi.GET("/statistics/shift_peak_hour", wrapper.CountShiftPeakHour)                                 // 统计班次高峰期
		privateApi.GET("/statistics/home", wrapper.CountHome)                                                     // 统计首页
		privateApi.GET("/statistics/kitchen/efficiency_analysis", wrapper.CountKitchenEfficiencyAnalysis)         // 统计后厨效率分析
		privateApi.GET("/statistics/kitchen/efficiency_analysis/avg", wrapper.CountKitchenEfficiencyAnalysisAvg)  // 统计后厨效率分析平均时长
		privateApi.GET("/statistics/kitchen/production_detail", wrapper.CountKitchenProductionDetail)             // 后厨菜品出品明细
		privateApi.GET("/statistics/kitchen/production_detail/export", wrapper.ExportKitchenProductionDetail)     // 导出后厨菜品出品明细
		privateApi.GET("/statistics/kitchen/efficiency_analysis/export", wrapper.ExportKitchenEfficiencyAnalysis) // 导出后厨效率分析
		privateApi.GET("/statistics/business/time_period", wrapper.CountBusinessTimePeriod)                       // 统计营业时段数据，移动端-报表-营业报表-时段营业统计
		privateApi.GET("/statistics/business/summary", wrapper.CountBusinessSummary)                              // 综合运营统计, 移动端-报表-营业报表-综合运营统计
		privateApi.GET("/statistics/business/payment_method", wrapper.CountBusinessPaymentMethod)                 // 统计支付方式, 移动端-报表-营业报表-支付方式统计
		privateApi.GET("/statistics/business/time_period/export", wrapper.ExportBusinessTimePeriod)               // 导出营业时段数据, 移动端-报表-营业报表-时段营业统计
		privateApi.GET("/statistics/business/summary/export", wrapper.ExportBusinessSummary)                      // 导出综合运营统计, 移动端-报表-营业报表-综合运营统计
		privateApi.GET("/statistics/business/payment_method/export", wrapper.ExportBusinessPaymentMethod)         // 导出营业收款统计, 移动端-报表-营业报表-支付方式统计
		privateApi.GET("/statistics/channel_sales", wrapper.ChannelSales)                                         // 渠道营业统计查询
		privateApi.GET("/statistics/channel_sales/export", wrapper.ExportChannelSales)                            // 导出渠道营业统计
		privateApi.GET("/statistics/user_analysis", wrapper.UserAnalysis)                                         // 用户分析统计查询
		privateApi.GET("/statistics/user_analysis/export", wrapper.ExportUserAnalysis)                            // 导出用户分析统计
		privateApi.GET("/statistics/company_list", wrapper.GetCompanyList)                                        // 获取门店汇总统计可选择的门店列表
		privateApi.GET("/statistics/company/payment_methods", wrapper.GetCompanyPaymentMethods)                   // 获取门店支付方式列表（汇总去重）
		privateApi.GET("/statistics/company/business/summary", wrapper.CountCompanyBusinessSummary)               // 获取门店汇总统计（营业数据汇总、支付方式汇总、退款金额汇总）
		privateApi.GET("/statistics/company/business/summary/export", wrapper.ExportCompanyBusinessSummary)       // 导出门店汇总统计（营业数据汇总、支付方式汇总、退款金额汇总）
	}

	// 需要认证 + 版本检查
	versionApi := router.Group("", middleware.MinVersionCheck(settingSrv, middleware.TypeStatistics), middleware.Auth(authSrv, dbm))
	{
		versionApi.GET("/statistics/business", wrapper.CountBusiness)        // 统计营业数据，移动管理端首页-店内概况
		versionApi.GET("/statistics/area", wrapper.CountArea)                // 统计区域，移动管理端首页-区域数据
		versionApi.GET("/statistics/product_rank", wrapper.CountProductRank) // 统计商品排行，移动管理端首页-销量、销售额排行
	}
}
