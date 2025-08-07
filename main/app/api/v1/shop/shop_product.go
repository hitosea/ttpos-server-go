package shop

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

// ProductHandler 商品处理程序
type ProductHandler struct {
	productSrv service.IProductSrv // 商品服务
}

// GetProductCategoryList 获取商品分类列表
// @Summary 获取商品分类列表
// @Description 获取商品分类列表
// @Tags 商家端.商品分类
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category/list [get]
func (h *ProductHandler) GetProductCategoryList(c *gin.Context) {
	// 获取商品分类列表
	res, err := h.productSrv.GetProductCategoryList(helper.GetCompanyUuid(c))
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetProductUnitList 获取商品单位列表
// @Summary 获取商品单位列表
// @Description 获取商品单位列表
// @Tags 商家端.商品单位
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} product_resp.ProductUnitListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/unit/list [get]
func (h *ProductHandler) GetProductUnitList(c *gin.Context) {
	ctx := helper.GetContext(c)
	listReq := req.ProductUnitListReq{}

	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	// 获取商品单位列表
	res, err := h.productSrv.GetProductUnitList(ctx, listReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetProductUnitList 获取商品单位详情
// @Summary 获取商品单位详情
// @Description 获取商品单位详情
// @Tags 商家端.商品单位
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "商品单位UUID"
// @Success 200 {object} product_resp.ProductUnitDetail "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/unit [get]
func (h *ProductHandler) GetProductUnit(c *gin.Context) {
	ctx := helper.GetContext(c)
	unitReq := req.ProductUnitReq{}

	if err := c.ShouldBindQuery(&unitReq); err != nil {
		helper.HandleValidationError(c, err, unitReq, dto.PageReqMessage)
		return
	}
	// 获取商品单位列表
	res, err := h.productSrv.GetProductUnit(ctx, unitReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// AddProductUnit 添加商品单位
// @Summary 添加商品单位
// @Description 添加商品单位
// @Tags 商家端.商品单位
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductUnitAddReq true "商品单位添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/unit/add [post]
func (h *ProductHandler) AddProductUnit(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductUnitAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, dto.PageReqMessage)
		return
	}
	// 添加商品单位
	err := h.productSrv.AddProductUnit(ctx, addReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

// EditProductUnit 编辑商品单位
// @Summary 编辑商品单位
// @Description 编辑商品单位
// @Tags 商家端.商品单位
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductUnitEditReq true "商品单位编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/unit/edit [post]
func (h *ProductHandler) EditProductUnit(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductUnitEditReq{}

	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, dto.PageReqMessage)
		return
	}
	// 编辑商品单位
	err := h.productSrv.EditProductUnit(ctx, editReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

// DeleteProductUnit 删除商品单位
// @Summary 删除商品单位
// @Description 删除商品单位
// @Tags 商家端.商品单位
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductUnitReq true "商品单位删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/unit [delete]
func (h *ProductHandler) DeleteProductUnit(c *gin.Context) {
	ctx := helper.GetContext(c)
	unitReq := req.ProductUnitReq{}

	if err := c.ShouldBindJSON(&unitReq); err != nil {
		helper.HandleValidationError(c, err, unitReq, dto.PageReqMessage)
		return
	}
	// 删除商品单位
	err := h.productSrv.DeleteProductUnit(ctx, unitReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

// SortProductUnit 排序商品单位
// @Summary 排序商品单位
// @Description 排序商品单位
// @Tags 商家端.商品单位
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductUnitSortReq true "商品单位排序请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/unit/sort [post]
func (h *ProductHandler) SortProductUnit(c *gin.Context) {
	ctx := helper.GetContext(c)
	sortReq := req.ProductUnitSortReq{}

	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.HandleValidationError(c, err, sortReq, dto.PageReqMessage)
		return
	}
	// 排序商品单位
	err := h.productSrv.SortProductUnit(ctx, sortReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
}

// GetProductSourceList 获取商品加料列表
// @Summary 获取商品加料列表
// @Description 获取商品加料列表
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} product_resp.ProductSourceListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/source/list [get]
func (h *ProductHandler) GetProductSourceList(c *gin.Context) {
	ctx := helper.GetContext(c)
	sourceListReq := req.ProductSourceListReq{}

	if err := c.ShouldBindQuery(&sourceListReq); err != nil {
		helper.HandleValidationError(c, err, sourceListReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductSourceList(ctx, sourceListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductSource 获取商品加料详情
// @Summary 获取商品加料详情
// @Description 获取商品加料详情
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "商品加料UUID"
// @Success 200 {object} product_resp.ProductSourceDetail "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/source [get]
func (h *ProductHandler) GetProductSource(c *gin.Context) {
	ctx := helper.GetContext(c)
	sourceReq := req.ProductSourceReq{}
	if err := c.ShouldBindQuery(&sourceReq); err != nil {
		helper.HandleValidationError(c, err, sourceReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductSource(ctx, sourceReq)
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

	// 创建收银产品处理程序
	wrapper := ProductHandler{
		productSrv: service.NewProductSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
			settingSrv,
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList) // 获取商品分类列表
		privateApi.GET("/product/unit/list", wrapper.GetProductUnitList)         // 获取商品单位列表
		privateApi.GET("/product/unit", wrapper.GetProductUnit)                  // 获取商品单位详情
		privateApi.POST("/product/unit/add", wrapper.AddProductUnit)             // 添加商品单位
		privateApi.POST("/product/unit/edit", wrapper.EditProductUnit)           // 编辑商品单位
		privateApi.DELETE("/product/unit", wrapper.DeleteProductUnit)            // 删除商品单位
		privateApi.POST("/product/unit/sort", wrapper.SortProductUnit)           // 排序商品单位

		privateApi.GET("/product/source/list", wrapper.GetProductSourceList) // 获取商品加料列表
		privateApi.GET("/product/source", wrapper.GetProductSource)          // 获取商品加料详情
	}
}
