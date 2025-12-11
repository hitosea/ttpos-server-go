package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// statisticsHandler 营业数据相关控制器
type statisticsHandler struct {
	businessSrv service.IBusinessSrv
}

// Printer 打印
// @Summary 打印
// @Description 打印
// @Tags 收银端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataPrinterReq true "打印参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/statistics/printer [post]
func (h *statisticsHandler) Printer(c *gin.Context) {
	ctx := helper.GetContext(c)
	var printerReq req.BusinessDataPrinterReq
	if err := c.ShouldBindJSON(&printerReq); err != nil {
		helper.HandleValidationError(c, err, printerReq, nil)
		return
	}
	printerData, err := h.businessSrv.Printer(ctx, printerReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, printerData, "发送成功")
}

// CountBusiness 统计营业数据
// @Summary 统计营业数据
// @Description 统计营业数据
// @Tags 收银端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataAll} "统计数据"
// @Router /cashier/statistics/business [get]
func (h *statisticsHandler) CountBusiness(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}

	companySetting := ctx.GetCompanySetting()
	dataSetting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global).GetDataManageSetting(ctx)

	countReq.NotQueryFree = true
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
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
// @Tags 收银端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataPaymentMethod} "统计数据"
// @Router /cashier/statistics/payment_method [get]
func (h *statisticsHandler) CountPaymentMethod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}

	companySetting := ctx.GetCompanySetting()
	dataSetting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global).GetDataManageSetting(ctx)
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
// @Tags 收银端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataProductCategory} "统计数据"
// @Router /cashier/statistics/product_category [get]
func (h *statisticsHandler) CountProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}

	companySetting := ctx.GetCompanySetting()
	dataSetting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global).GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
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
// @Tags 收银端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataCountReq true "统计参数"
// @Success 200 {object} dto.Response{data=business_data_resp.BusinessDataProduct} "统计数据"
// @Router /cashier/statistics/product [get]
func (h *statisticsHandler) CountProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	var countReq req.BusinessDataCountReq
	if err := c.ShouldBindQuery(&countReq); err != nil {
		helper.HandleValidationError(c, err, countReq, nil)
		return
	}

	companySetting := ctx.GetCompanySetting()
	dataSetting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global).GetDataManageSetting(ctx)
	countReq.ExcludeDataManage = companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
	productData, err := h.businessSrv.CountProduct(ctx, countReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, productData)
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
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/statistics/printer", wrapper.Printer)                      // 打印
		privateApi.GET("/statistics/business", wrapper.CountBusiness)                // 统计营业数据
		privateApi.GET("/statistics/payment_method", wrapper.CountPaymentMethod)     // 统计支付方式
		privateApi.GET("/statistics/product_category", wrapper.CountProductCategory) // 统计商品分类
		privateApi.GET("/statistics/product", wrapper.CountProduct)                  // 统计商品
	}
}
