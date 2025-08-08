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
// @Success 200 {object} product_resp.ProductShopCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category/list [get]
func (h *ProductHandler) GetProductCategoryList(c *gin.Context) {
	// 获取商品分类列表
	ctx := helper.GetContext(c)
	listReq := req.ProductShopCategoryListReq{}

	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductShopCategoryList(ctx, listReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// SortProductCategory 排序商品分类
// @Summary 排序商品分类
// @Description 排序商品分类
// @Tags 商家端.商品分类
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopCategorySortReq true "商品分类排序请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category/sort [post]
func (h *ProductHandler) SortProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	sortReq := req.ProductShopCategorySortReq{}

	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 排序商品单位
	err := h.productSrv.SortProductShopCategory(ctx, sortReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// AddProductCategory 添加商品分类
// @Summary 添加商品分类
// @Description 添加商品分类
// @Tags 商家端.商品分类
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopCategoryAddReq true "商品分类添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category/add [post]
func (h *ProductHandler) AddProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductShopCategoryAddReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.AddProductShopCategory(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// EditProductCategory 编辑商品分类
// @Summary 编辑商品分类
// @Description 编辑商品分类
// @Tags 商家端.商品分类
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopCategoryEditReq true "商品分类编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category/edit [post]
func (h *ProductHandler) EditProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductShopCategoryEditReq{}
	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.EditProductShopCategory(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// DeleteProductCategory 删除商品分类
// @Summary 删除商品分类
// @Description 删除商品分类
// @Tags 商家端.商品分类
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopCategoryReq true "商品分类删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category [delete]
func (h *ProductHandler) DeleteProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductShopCategoryReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.DeleteProductShopCategory(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
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

// GetProductSauceList 获取商品加料列表
// @Summary 获取商品加料列表
// @Description 获取商品加料列表
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} product_resp.ProductSauceListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/sauce/list [get]
func (h *ProductHandler) GetProductSauceList(c *gin.Context) {
	ctx := helper.GetContext(c)
	sauceListReq := req.ProductSauceListReq{}

	if err := c.ShouldBindQuery(&sauceListReq); err != nil {
		helper.HandleValidationError(c, err, sauceListReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductSauceList(ctx, sauceListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductSauce 获取商品加料详情
// @Summary 获取商品加料详情
// @Description 获取商品加料详情
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "商品加料UUID"
// @Success 200 {object} product_resp.ProductSauceDetail "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/sauce [get]
func (h *ProductHandler) GetProductSauce(c *gin.Context) {
	ctx := helper.GetContext(c)
	sauceReq := req.ProductSauceReq{}
	if err := c.ShouldBindQuery(&sauceReq); err != nil {
		helper.HandleValidationError(c, err, sauceReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductSauce(ctx, sauceReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddProductSauce 添加商品加料
// @Summary 添加商品加料
// @Description 添加商品加料
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductSauceAddReq true "商品加料添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/sauce/add [post]
func (h *ProductHandler) AddProductSauce(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductSauceAddReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.AddProductSauce(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// EditProductSauce 编辑商品加料
// @Summary 编辑商品加料
// @Description 编辑商品加料
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductSauceEditReq true "商品加料编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/sauce/edit [post]
func (h *ProductHandler) EditProductSauce(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductSauceEditReq{}
	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.EditProductSauce(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// DeleteProductSauce 删除商品加料
// @Summary 删除商品加料
// @Description 删除商品加料
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductSauceReq true "商品加料删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/sauce [delete]
func (h *ProductHandler) DeleteProductSauce(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductSauceReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.DeleteProductSauce(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// SortProductSauce 排序商品加料
// @Summary 排序商品加料
// @Description 排序商品加料
// @Tags 商家端.商品加料
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductSauceSortReq true "商品加料排序请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/sauce/sort [post]
func (h *ProductHandler) SortProductSauce(c *gin.Context) {
	ctx := helper.GetContext(c)
	sortReq := req.ProductSauceSortReq{}
	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.HandleValidationError(c, err, sortReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.SortProductSauce(ctx, sortReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
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
		privateApi.POST("/product/category/sort", wrapper.SortProductCategory)   // 排序商品分类
		privateApi.POST("/product/category/add", wrapper.AddProductCategory)     // 添加商品分类
		privateApi.POST("/product/category/edit", wrapper.EditProductCategory)   // 编辑商品分类
		privateApi.DELETE("/product/category", wrapper.DeleteProductCategory)    // 删除商品分类

		privateApi.GET("/product/unit/list", wrapper.GetProductUnitList) // 获取商品单位列表
		privateApi.GET("/product/unit", wrapper.GetProductUnit)          // 获取商品单位详情
		privateApi.POST("/product/unit/add", wrapper.AddProductUnit)     // 添加商品单位
		privateApi.POST("/product/unit/edit", wrapper.EditProductUnit)   // 编辑商品单位
		privateApi.DELETE("/product/unit", wrapper.DeleteProductUnit)    // 删除商品单位
		privateApi.POST("/product/unit/sort", wrapper.SortProductUnit)   // 排序商品单位

		privateApi.GET("/product/sauce/list", wrapper.GetProductSauceList) // 获取商品加料列表
		privateApi.GET("/product/sauce", wrapper.GetProductSauce)          // 获取商品加料详情
		privateApi.POST("/product/sauce/add", wrapper.AddProductSauce)     // 添加商品加料
		privateApi.POST("/product/sauce/edit", wrapper.EditProductSauce)   // 编辑商品加料
		privateApi.DELETE("/product/sauce", wrapper.DeleteProductSauce)    // 删除商品加料
		privateApi.POST("/product/sauce/sort", wrapper.SortProductSauce)   // 排序商品加料
	}
}
