package kitchen

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productSrv    service.IProductSrv    // 产品服务
	productionSrv service.IProductionSrv // 送厨商品服务
}

// GetProductCategoryList 获取产品类别列表
// @Summary 获取产品类别列表
// @Description 获取产品类别列表
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kitchen/product/category/list [get]
func (h *ProductHandler) GetProductCategoryList(c *gin.Context) {
	res, err := h.productSrv.GetProductCategoryList(helper.GetCompanyUuid(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductListByOrder 根据订单获取送厨商品列表
// @Summary 根据订单获取送厨商品列表
// @Description 根据订单获取送厨商品列表
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.ProductionListReq true "分页参数"
// @Success 200 {object} dto.Response{data=resp.ProductionListWithPagination}
// @Router /kitchen/product/list_by_order [get]
func (h *ProductHandler) GetProductListByOrder(c *gin.Context) {
	var listReq req.ProductionListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	res, err := h.productionSrv.GetProductListByOrder(helper.GetContext(c), listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductListByCategory 根据分类获取送厨商品列表
// @Summary 根据分类获取送厨商品列表
// @Description 根据分类获取送厨商品列表
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.ProductionListByCategoryReq true "分页参数"
// @Success 200 {object} dto.Response{data=resp.ProductionListWithPagination}
// @Router /kitchen/product/list_by_category [get]
func (h *ProductHandler) GetProductListByCategory(c *gin.Context) {
	var listReq req.ProductionListByCategoryReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	res, err := h.productionSrv.GetProductListByCategory(helper.GetContext(c), listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetHistory 获取上菜历史
// @Summary 获取上菜历史
// @Description 获取上菜历史
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ProductionHistory}
// @Router /kitchen/product/history [get]
func (h *ProductHandler) GetHistory(c *gin.Context) {
	res, err := h.productionSrv.GetHistory(helper.GetContext(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// Finish 制作完成
// @Summary 制作完成
// @Description 制作完成
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ProductUuid true "送厨商品Uuid参数"
// @Router /kitchen/product/finish [post]
func (h *ProductHandler) Finish(c *gin.Context) {
	var productionReq req.ProductUuid
	if err := c.ShouldBindJSON(&productionReq); err != nil {
		helper.HandleValidationError(c, err, productionReq, nil)
		return
	}
	err := h.productionSrv.Finish(helper.GetContext(c), productionReq.ProductUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// Recovery 恢复制作
// @Summary 恢复制作
// @Description 恢复制作
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ProductUuid true "送厨商品Uuid参数"
// @Router /kitchen/product/recovery [post]
func (h *ProductHandler) Recovery(c *gin.Context) {
	var productionReq req.ProductUuid
	if err := c.ShouldBindJSON(&productionReq); err != nil {
		helper.HandleValidationError(c, err, productionReq, nil)
		return
	}
	err := h.productionSrv.Recovery(helper.GetContext(c), productionReq.ProductUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// ConfirmReturn 厨显端确认退菜
// @Summary 厨显端确认退菜
// @Description 厨显端确认退菜
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ProductUuid true "送厨商品Uuid参数"
// @Router /kitchen/product/confirm_return [post]
func (h *ProductHandler) ConfirmReturn(c *gin.Context) {
	var productionReq req.ProductUuid
	if err := c.ShouldBindJSON(&productionReq); err != nil {
		helper.HandleValidationError(c, err, productionReq, nil)
		return
	}
	err := h.productionSrv.ConfirmReturn(helper.GetContext(c), productionReq.ProductUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// ConfirmReturnAll 厨显端确认退菜整单
// @Summary 厨显端确认退菜整单
// @Description 厨显端确认退菜整单
// @Tags 厨显端.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SaleBillUuid true "按订单查看送厨商品，确认整单取消时传递销售账单Uuid"
// @Router /kitchen/product/confirm_return_all [post]
func (h *ProductHandler) ConfirmReturnAll(c *gin.Context) {
	var returnAllReq req.SaleBillUuid
	if err := c.ShouldBindJSON(&returnAllReq); err != nil {
		helper.HandleValidationError(c, err, returnAllReq, nil)
		return
	}
	err := h.productionSrv.ConfirmReturnAll(helper.GetContext(c), returnAllReq.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

func RegisterProductHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 创建产品处理程序
	wrapper := ProductHandler{
		productSrv: service.NewProductSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
			settingSrv,
			cache,
		),
		productionSrv: service.NewProductionSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList)      // 获取产品类别列表
		privateApi.GET("/product/list_by_order", wrapper.GetProductListByOrder)       // 按照订单获取送厨商品
		privateApi.GET("/product/list_by_category", wrapper.GetProductListByCategory) // 按照分类获取送厨商品
		privateApi.GET("/product/history", wrapper.GetHistory)                        // 获取上菜历史
		privateApi.POST("/product/finish", wrapper.Finish)                            // 制作完成
		privateApi.POST("/product/recovery", wrapper.Recovery)                        // 恢复制作
		privateApi.POST("/product/confirm_return", wrapper.ConfirmReturn)             // 厨显端确认退菜
		privateApi.POST("/product/confirm_return_all", wrapper.ConfirmReturnAll)      // 厨显端确认退菜整单
	}
}
