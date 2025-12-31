package kiosk

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
	productSrv service.IProductSrv // 产品服务
}

// GetProductCategoryList 获取产品类别列表
// @Summary 获取产品类别列表
// @Description 获取产品类别列表
// @Tags 自助点餐机.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kiosk/product/category/list [get]
func (h *ProductHandler) GetProductCategoryList(c *gin.Context) {
	ctx := helper.GetContext(c)
	res, err := h.productSrv.GetProductCategoryList(ctx.GetDbId())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductList 获取产品列表
// @Summary 获取产品列表
// @Description 获取产品列表
// @Tags 自助点餐机.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页条数"
// @Param category_uuid query uint64 false "分类UUID"
// @Success 200 {object} product_resp.ProductListWithPaginationResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kiosk/product/list [get]
func (h *ProductHandler) GetProductList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var productListReq req.ProductListReq
	if err := c.ShouldBindQuery(&productListReq); err != nil {
		helper.HandleValidationError(c, err, productListReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductList(ctx, productListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductDetail 获取商品详情
// @Summary 获取商品详情
// @Description 获取商品详细信息
// @Tags 自助点餐机.产品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query uint64 true "商品UUID"
// @Success 200 {object} product_resp.ProductDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kiosk/product/detail [get]
func (h *ProductHandler) GetProductDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var productDetailReq req.ProductDetailReq
	if err := c.ShouldBindQuery(&productDetailReq); err != nil {
		helper.HandleValidationError(c, err, productDetailReq, nil)
		return
	}
	res, err := h.productSrv.GetProductDetail(ctx, productDetailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
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
	translateSrv := service.NewTranslateSrv(dbm, cache)

	// 创建产品处理程序
	wrapper := ProductHandler{
		productSrv: service.NewProductSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
			settingSrv,
			cache,
			translateSrv,
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList) // 获取产品类别列表
		privateApi.GET("/product/list", wrapper.GetProductList)                  // 获取商品列表
		privateApi.GET("/product/detail", wrapper.GetProductDetail)              // 获取商品详情
	}
}
