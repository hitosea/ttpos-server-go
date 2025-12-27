package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	printerService "ttpos-server-go/app/printer/service"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// ProductHandler 商品处理程序
type ProductHandler struct {
	productSrv        service.IProductSrv        // 商品服务
	productTakeoutSrv service.IProductTakeoutSrv // 外卖商品服务
	uploadFileSrv     service.IUploadFileSrv     // 文件上传服务
	pinterSrv         printerService.IPrinterSrv // 打印服务
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

// GetProductCategory 获取商品分类详情
// @Summary 获取商品分类详情
// @Description 获取商品分类详情
// @Tags 商家端.商品分类
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "商品分类UUID"
// @Success 200 {object} product_resp.ProductShopCategoryDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category [get]
func (h *ProductHandler) GetProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	detailReq := req.ProductShopCategoryReq{}
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductShopCategory(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
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
	helper.Success(c, nil, "保存成功")
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
	_, err := h.productSrv.AddProductShopCategory(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "添加成功")
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
	_, err := h.productSrv.EditProductShopCategory(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "保存成功")
}

// DeleteProductCategory 删除商品分类
// @Summary 删除商品分类
// @Description 删除商品分类
// @Tags 商家端.商品分类
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopCategoryDeleteReq true "商品分类删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/category [delete]
func (h *ProductHandler) DeleteProductCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductShopCategoryDeleteReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.DeleteProductShopCategory(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "删除成功")
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
	helper.Success(c, gin.H{}, "保存成功")
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
	helper.Success(c, gin.H{}, "保存成功")
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
	helper.Success(c, gin.H{}, "删除成功")
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
	helper.Success(c, gin.H{}, "保存成功")
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
	helper.Success(c, gin.H{}, "保存成功")
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
	helper.Success(c, gin.H{}, "保存成功")
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
	helper.Success(c, gin.H{}, "删除成功")
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
	helper.Success(c, gin.H{}, "保存成功")
}

// GetProductAttributeGroupList 获取商品属性分组列表
// @Summary 获取商品属性分组列表
// @Description 获取商品属性分组列表
// @Tags 商家端.商品属性分组
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} product_resp.ProductAttributeGroupListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/attribute/group/list [get]
func (h *ProductHandler) GetProductAttributeGroupList(c *gin.Context) {
	ctx := helper.GetContext(c)
	attributeGroupListReq := req.ProductAttributeGroupListReq{}
	if err := c.ShouldBindQuery(&attributeGroupListReq); err != nil {
		helper.HandleValidationError(c, err, attributeGroupListReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductAttributeGroupList(ctx, attributeGroupListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductAttributeGroup 获取商品属性分组详情
// @Summary 获取商品属性分组详情
// @Description 获取商品属性分组详情
// @Tags 商家端.商品属性分组
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "商品属性分组UUID"
// @Success 200 {object} product_resp.ProductAttributeGroupDetail "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/attribute/group [get]
func (h *ProductHandler) GetProductAttributeGroup(c *gin.Context) {
	ctx := helper.GetContext(c)
	attributeGroupReq := req.ProductAttributeGroupReq{}
	if err := c.ShouldBindQuery(&attributeGroupReq); err != nil {
		helper.HandleValidationError(c, err, attributeGroupReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductAttributeGroup(ctx, attributeGroupReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddProductAttributeGroup 添加商品属性分组，商品属性一起添加
// @Summary 添加商品属性分组，商品属性一起添加
// @Description 添加商品属性分组，商品属性一起添加
// @Tags 商家端.商品属性分组
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductAttributeGroupAddReq true "商品属性分组添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/attribute/group/add [post]
func (h *ProductHandler) AddProductAttributeGroup(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductAttributeGroupAddReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, dto.PageReqMessage)
		return
	}
	_, err := h.productSrv.AddProductAttributeGroup(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// GetProductFlavor 获取商品规格详情
// @Summary 获取商品规格详情
// @Description 获取商品规格详情
// @Tags 商家端.商品规格
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "商品规格UUID"
// @Success 200 {object} product_resp.ProductFlavorDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/flavor [get]
func (h *ProductHandler) GetProductFlavor(c *gin.Context) {
	ctx := helper.GetContext(c)
	flavorReq := req.ProductFlavorReq{}
	if err := c.ShouldBindQuery(&flavorReq); err != nil {
		helper.HandleValidationError(c, err, flavorReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductFlavor(ctx, flavorReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddProductFlavor 添加商品规格
// @Summary 添加商品规格
// @Description 添加商品规格
// @Tags 商家端.商品规格
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductFlavorAddReq true "商品规格添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/flavor/add [post]
func (h *ProductHandler) AddProductFlavor(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductFlavorAddReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, dto.PageReqMessage)
		return
	}
	_, err := h.productSrv.AddProductFlavor(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// GetProductFlavorList 获取商品规格列表
// @Summary 获取商品规格列表
// @Description 获取商品规格列表
// @Tags 商家端.商品规格
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} product_resp.ProductFlavorListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/flavor/list [get]
func (h *ProductHandler) GetProductFlavorList(c *gin.Context) {
	ctx := helper.GetContext(c)
	flavorListReq := req.ProductFlavorListReq{}
	if err := c.ShouldBindQuery(&flavorListReq); err != nil {
		helper.HandleValidationError(c, err, flavorListReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductFlavorList(ctx, flavorListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// EditProductAttributeGroup 编辑商品属性分组，商品属性一起编辑
// @Summary 编辑商品属性分组，商品属性一起编辑
// @Description 编辑商品属性分组，商品属性一起编辑
// @Tags 商家端.商品属性分组
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductAttributeGroupEditReq true "商品属性分组编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/attribute/group/edit [post]
func (h *ProductHandler) EditProductAttributeGroup(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductAttributeGroupEditReq{}
	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.EditProductAttributeGroup(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// DeleteProductAttributeGroup 删除商品属性分组
// @Summary 删除商品属性分组
// @Description 删除商品属性分组
// @Tags 商家端.商品属性分组
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductAttributeGroupReq true "商品属性分组删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/attribute/group [delete]
func (h *ProductHandler) DeleteProductAttributeGroup(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductAttributeGroupReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.DeleteProductAttributeGroup(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "删除成功")
}

// DeleteProductAttribute 删除商品属性值
// @Summary 删除商品属性值
// @Description 删除商品属性值
// @Tags 商家端.商品属性值
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductAttributeDeleteReq true "商品属性值删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/attribute [delete]
func (h *ProductHandler) DeleteProductAttribute(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductAttributeDeleteReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.DeleteProductAttribute(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "删除成功")
}

// DeleteProductFlavor 删除商品规格
// @Summary 删除商品规格
// @Description 删除商品规格
// @Tags 商家端.商品规格
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductFlavorDeleteReq true "商品规格删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/flavor [delete]
func (h *ProductHandler) DeleteProductFlavor(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductFlavorDeleteReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.DeleteProductFlavor(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "删除成功")
}

// SortProductAttributeGroup 排序商品属性分组
// @Summary 排序商品属性分组
// @Description 排序商品属性分组
// @Tags 商家端.商品属性分组
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductAttributeGroupSortReq true "商品属性分组排序请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/attribute/group/sort [post]
func (h *ProductHandler) SortProductAttributeGroup(c *gin.Context) {
	ctx := helper.GetContext(c)
	sortReq := req.ProductAttributeGroupSortReq{}
	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.HandleValidationError(c, err, sortReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.SortProductAttributeGroup(ctx, sortReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// SortProductFlavor 排序商品规格
// @Summary 排序商品规格
// @Description 排序商品规格
// @Tags 商家端.商品规格
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductFlavorSortReq true "商品规格排序请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/flavor/sort [post]
func (h *ProductHandler) SortProductFlavor(c *gin.Context) {
	ctx := helper.GetContext(c)
	sortReq := req.ProductFlavorSortReq{}
	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.HandleValidationError(c, err, sortReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.SortProductFlavor(ctx, sortReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// // SortProductAttribute 排序商品属性
// // @Summary 排序商品属性
// // @Description 排序商品属性
// // @Tags 商家端.商品属性
// // @Accept json
// // @Produce json
// // @Security JwtToken
// // @Param data body req.ProductAttributeSortReq true "商品属性排序请求"
// // @Success 200 {object} nil "成功"
// // @Failure 400 {object} nil "错误请求"
// // @Router /shop/product/attribute/sort [post]
// func (h *ProductHandler) SortProductAttribute(c *gin.Context) {
// 	ctx := helper.GetContext(c)
// 	sortReq := req.ProductAttributeSortReq{}
// 	if err := c.ShouldBindJSON(&sortReq); err != nil {
// 		helper.HandleValidationError(c, err, sortReq, dto.PageReqMessage)
// 		return
// 	}
// 	err := h.productSrv.SortProductAttribute(ctx, sortReq)
// 	if err != nil {
// 		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
// 		return
// 	}
// 	helper.Success(c, gin.H{}, "保存成功")
// }

// EditProductFlavor 编辑商品规格
// @Summary 编辑商品规格
// @Description 编辑商品规格
// @Tags 商家端.商品规格
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductFlavorEditReq true "商品规格编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/flavor/edit [post]
func (h *ProductHandler) EditProductFlavor(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductFlavorEditReq{}
	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.EditProductFlavor(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{}, "保存成功")
}

// ImportProductList 导入商品-获取导入商品列表
// @Summary 导入商品-获取导入商品列表
// @Description 导入商品-获取导入商品列表
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductImportListReq true "导入商品列表请求"
// @Success 200 {object} product_resp.ProductImportResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/import/list [post]
func (h *ProductHandler) ImportProductList(c *gin.Context) {
	ctx := helper.GetContext(c)
	importReq := req.ProductImportListReq{}
	if err := c.ShouldBindJSON(&importReq); err != nil {
		helper.HandleValidationError(c, err, importReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.ImportProductList(ctx, importReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// ImportProduct 导入商品
// @Summary 导入商品
// @Description 导入商品
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductImportReq true "导入商品请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/import [post]
func (h *ProductHandler) ImportProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	importReq := req.ProductImportReq{}
	if err := c.ShouldBindJSON(&importReq); err != nil {
		helper.HandleValidationError(c, err, importReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.ImportProduct(ctx, importReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "导入成功")
}

// GetProductSingleList 获取单规格商品列表
// @Summary 获取单规格商品列表
// @Description 获取单规格商品列表
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.ProductSingleListReq true "单规格商品列表请求"
// @Success 200 {object} product_resp.ProductSingleListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/single/list [get]
func (h *ProductHandler) GetProductSingleList(c *gin.Context) {
	ctx := helper.GetContext(c)
	singleListReq := req.ProductSingleListReq{}
	if err := c.ShouldBindQuery(&singleListReq); err != nil {
		helper.HandleValidationError(c, err, singleListReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductSingleList(ctx, singleListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductShopList 获取商品列表
// @Summary 获取商品列表
// @Description 获取商品列表
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.ProductShopListReq true "商品列表请求"
// @Success 200 {object} product_resp.ProductShopListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/list [get]
func (h *ProductHandler) GetProductShopList(c *gin.Context) {
	ctx := helper.GetContext(c)
	productShopListReq := req.ProductShopListReq{}
	if err := c.ShouldBindQuery(&productShopListReq); err != nil {
		helper.HandleValidationError(c, err, productShopListReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductShopList(ctx, productShopListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductPrinterList 获取打印档口列表
// @Summary 获取打印档口列表
// @Description 获取打印档口列表
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ProductPrinterList}
// @Router /shop/product/printer_list [get]
func (h *ProductHandler) GetProductPrinterList(c *gin.Context) {
	ctx := helper.GetContext(c)
	data, err := h.pinterSrv.GetProductPrinterList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, data)
}

// SortProductShopList 排序商品列表
// @Summary 排序商品列表
// @Description 排序商品列表
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.SortProductShopListReq true "商品排序请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/sort [post]
func (h *ProductHandler) SortProductShopList(c *gin.Context) {
	ctx := helper.GetContext(c)
	sortReq := req.SortProductShopListReq{}
	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.HandleValidationError(c, err, sortReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.SortProductShopList(ctx, sortReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "保存成功")
}

// GetProductDetail 获取商品详情
// @Summary 获取商品详情
// @Description 获取商品详情
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.ProductDetailReq true "商品详情请求"
// @Success 200 {object} product_resp.ProductDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/detail [get]
func (h *ProductHandler) GetProductDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	detailReq := req.ProductDetailReq{}
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, dto.PageReqMessage)
		return
	}
	res, err := h.productSrv.GetProductDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// ProductShopStatus 修改商品状态
// @Summary 修改商品状态
// @Description 修改商品状态
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopStatusReq true "商品状态请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/status [post]
func (h *ProductHandler) ProductShopStatus(c *gin.Context) {
	ctx := helper.GetContext(c)
	statusReq := req.ProductShopStatusReq{}
	if err := c.ShouldBindJSON(&statusReq); err != nil {
		helper.HandleValidationError(c, err, statusReq, dto.PageReqMessage)
		return
	}
	err := h.productSrv.ProductShopStatus(ctx, statusReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "设置成功")
}

// ProductShopAdd 添加商品
// @Summary 添加商品
// @Description 添加商品
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopAddReq true "商品添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/add [post]
func (h *ProductHandler) ProductShopAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductShopAddReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, nil)
		return
	}
	uuid, err := h.productSrv.AddProductShop(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, map[string]interface{}{
		"uuid": uuid,
	}, "保存成功")
}

// ProductShopEdit 编辑商品
// @Summary 编辑商品
// @Description 编辑商品
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopEditReq true "商品编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/edit [post]
func (h *ProductHandler) ProductShopEdit(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductShopEditReq{}
	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, nil)
		return
	}
	if err := editReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	data, replace, err := h.productSrv.EditProductShop(ctx, editReq)
	if err != nil {
		if data != nil {
			appErr := errors.AppError{
				Code:    constant.CodeProductEditCanNotDeletePackage,
				Message: err.Error(),
				Replace: replace,
			}
			helper.ErrorWithData(c, constant.CodeFail, data, appErr)
		} else {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		}
		return
	}
	helper.Success(c, nil, "保存成功")
}

// ProductShopDelete 删除商品
// @Summary 删除商品
// @Description 删除商品
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopDeleteReq true "商品删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/delete [delete]
func (h *ProductHandler) ProductShopDelete(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductShopDeleteReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}
	data, err := h.productSrv.DeleteProductShop(ctx, deleteReq)
	if err != nil {
		if data != nil {
			appErr := errors.AppError{
				Code:    constant.CodeProductDeleteCanNotDeletePackage,
				Message: "商品已关联如下套餐，暂时无法删除，请先修改套餐",
				Replace: []string{},
			}
			helper.ErrorWithData(c, constant.CodeProductDeleteCanNotDeletePackage, data, appErr)
		} else {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		}
		return
	}
	helper.Success(c, nil, "删除成功")
}

// ProductTakeoutShopAdd 添加外卖商品
// @Summary 添加外卖商品
// @Description 添加外卖商品配置
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductTakeoutShopAddReq true "外卖商品添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/takeout/add [post]
func (h *ProductHandler) ProductTakeoutShopAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductTakeoutShopAddReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, nil)
		return
	}
	productPackageTakeout, err := h.productTakeoutSrv.AddProductTakeoutShop(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{"uuid": productPackageTakeout.Uuid}, "保存成功")
}

// ProductTakeoutShopEdit 编辑外卖商品
// @Summary 编辑外卖商品
// @Description 编辑外卖商品配置
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductTakeoutShopEditReq true "外卖商品编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/takeout/edit [post]
func (h *ProductHandler) ProductTakeoutShopEdit(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductTakeoutShopEditReq{}
	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, nil)
		return
	}
	if err := editReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	err := h.productTakeoutSrv.EditProductTakeoutShop(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "保存成功")
}

// ProductTakeoutShopDetail 获取外卖商品详情
// @Summary 获取外卖商品详情
// @Description 获取外卖商品详情
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.ProductTakeoutShopDetailReq true "外卖商品详情请求"
// @Success 200 {object} product_resp.ProductTakeoutShopDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/takeout/detail [get]
func (h *ProductHandler) ProductTakeoutShopDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	detailReq := req.ProductTakeoutShopDetailReq{}
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, dto.PageReqMessage)
		return
	}
	res, err := h.productTakeoutSrv.GetProductTakeoutShopDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// ProductTakeoutShopDelete 删除外卖商品
// @Summary 删除外卖商品
// @Description 删除外卖商品（软删除），再次添加时会自动还原
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductTakeoutShopDeleteReq true "外卖商品删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/takeout/delete [post]
func (h *ProductHandler) ProductTakeoutShopDelete(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductTakeoutShopDeleteReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}
	err := h.productTakeoutSrv.DeleteProductTakeoutShop(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "删除成功")
}

// ProductTaxList 获取商品税类列表
// @Summary 获取商品税类列表
// @Description 获取商品税类列表
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductTaxListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/tax/list [get]
func (h *ProductHandler) ProductTaxList(c *gin.Context) {
	ctx := helper.GetContext(c)
	res := h.productSrv.ProductTaxList(ctx)
	helper.Success(c, res)
}

// UploadProductImage 上传商品图片
// @Summary 上传商品图片
// @Description 上传商品图片
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @param file formData file true "上传商品图片"
// @Success 200 {object} dto.Response
// @Router /shop/product/upload_image [post]
func (h *ProductHandler) UploadProductImage(c *gin.Context) {
	ctx := helper.GetContext(c)
	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	fileReader, err := file.Open()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	uploadFileResp, err := h.uploadFileSrv.UploadImage(ctx, fileReader, file.Filename, file.Size, 0, "productImage")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, uploadFileResp)
}

// ProductShopChangePrice 商品改价

// @Summary 商品改价
// @Description 商品改价
// @Tags 商家端.商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductShopChangePriceReq true "商品改价请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/change_price [post]
func (h *ProductHandler) ProductShopChangePrice(c *gin.Context) {
	ctx := helper.GetContext(c)
	changePriceReq := req.ProductShopChangePriceReq{}
	if err := c.ShouldBindJSON(&changePriceReq); err != nil {
		helper.HandleValidationError(c, err, changePriceReq, nil)
		return
	}
	err := h.productSrv.ProductShopChangePrice(ctx, changePriceReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "改价成功")
}

// TakeoutBatchCreate 批量创建外卖商品
// @Summary 批量创建外卖商品
// @Description 批量创建外卖商品映射关系
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TakeoutBatchCreateReq true "批量创建请求"
// @Success 200 {object} product_resp.TakeoutBatchResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/takeout/products/batch_create [post]
func (h *ProductHandler) TakeoutBatchCreate(c *gin.Context) {
	ctx := helper.GetContext(c)
	batchReq := req.TakeoutBatchCreateReq{}
	if err := c.ShouldBindJSON(&batchReq); err != nil {
		helper.HandleValidationError(c, err, batchReq, nil)
		return
	}
	if err := batchReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	result, err := h.productTakeoutSrv.BatchCreateProducts(ctx, batchReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result, "批量创建完成")
}

// TakeoutBatchOnline 批量上架外卖商品
// @Summary 批量上架外卖商品
// @Description 批量上架外卖商品
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TakeoutBatchOnlineReq true "批量上架请求"
// @Success 200 {object} product_resp.TakeoutBatchResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/takeout/products/batch_online [post]
func (h *ProductHandler) TakeoutBatchOnline(c *gin.Context) {
	ctx := helper.GetContext(c)
	batchReq := req.TakeoutBatchOnlineReq{}
	if err := c.ShouldBindJSON(&batchReq); err != nil {
		helper.HandleValidationError(c, err, batchReq, nil)
		return
	}
	if err := batchReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	result, err := h.productTakeoutSrv.BatchOnlineProducts(ctx, batchReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result, "批量上架完成")
}

// TakeoutBatchOffline 批量下架外卖商品
// @Summary 批量下架外卖商品
// @Description 批量下架外卖商品
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TakeoutBatchOfflineReq true "批量下架请求"
// @Success 200 {object} product_resp.TakeoutBatchResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/takeout/products/batch_offline [post]
func (h *ProductHandler) TakeoutBatchOffline(c *gin.Context) {
	ctx := helper.GetContext(c)
	batchReq := req.TakeoutBatchOfflineReq{}
	if err := c.ShouldBindJSON(&batchReq); err != nil {
		helper.HandleValidationError(c, err, batchReq, nil)
		return
	}
	if err := batchReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	result, err := h.productTakeoutSrv.BatchOfflineProducts(ctx, batchReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result, "批量下架完成")
}

// TakeoutBatchDelete 批量删除外卖商品
// @Summary 批量删除外卖商品
// @Description 批量删除外卖商品映射关系
// @Tags 商家端.外卖商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TakeoutBatchDeleteReq true "批量删除请求"
// @Success 200 {object} product_resp.TakeoutBatchResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/takeout/products/batch_delete [post]
func (h *ProductHandler) TakeoutBatchDelete(c *gin.Context) {
	ctx := helper.GetContext(c)
	batchReq := req.TakeoutBatchDeleteReq{}
	if err := c.ShouldBindJSON(&batchReq); err != nil {
		helper.HandleValidationError(c, err, batchReq, nil)
		return
	}
	if err := batchReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	result, err := h.productTakeoutSrv.BatchDeleteProducts(ctx, batchReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result, "批量删除完成")
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

	// 创建收银产品处理程序
	localeSrv := service.NewLocaleSrv()
	wrapper := ProductHandler{
		productSrv: service.NewProductSrv(
			dbm,        // 数据库管理器
			localeSrv,  // 多语言服务
			settingSrv, // 设置服务
			cache,
			translateSrv,
		),
		productTakeoutSrv: service.NewProductTakeoutSrv(dbm, localeSrv, settingSrv, cache, translateSrv),
		uploadFileSrv:     service.NewUploadFileSrv(dbm),
		pinterSrv:         printerService.NewPrinterSrv(dbm, cache),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList) // 获取商品分类列表
		privateApi.GET("/product/category", wrapper.GetProductCategory)          // 获取商品分类详情
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

		privateApi.GET("/product/attribute/group/list", wrapper.GetProductAttributeGroupList) // 获取属性组列表
		privateApi.GET("/product/attribute/group", wrapper.GetProductAttributeGroup)          // 获取属性组详情
		privateApi.POST("/product/attribute/group/add", wrapper.AddProductAttributeGroup)     // 添加属性组，属性值一起添加
		privateApi.POST("/product/attribute/group/edit", wrapper.EditProductAttributeGroup)   // 编辑属性组，属性值一起编辑
		privateApi.DELETE("/product/attribute/group", wrapper.DeleteProductAttributeGroup)    // 删除属性组
		privateApi.POST("/product/attribute/group/sort", wrapper.SortProductAttributeGroup)   // 排序属性组
		privateApi.DELETE("/product/attribute", wrapper.DeleteProductAttribute)               // 删除属性值

		privateApi.GET("/product/flavor/list", wrapper.GetProductFlavorList) // 获取商品规格列表
		privateApi.GET("/product/flavor", wrapper.GetProductFlavor)          // 获取商品规格详情
		privateApi.POST("/product/flavor/add", wrapper.AddProductFlavor)     // 添加商品规格
		privateApi.POST("/product/flavor/edit", wrapper.EditProductFlavor)   // 编辑商品规格
		privateApi.DELETE("/product/flavor", wrapper.DeleteProductFlavor)    // 删除商品规格
		privateApi.POST("/product/flavor/sort", wrapper.SortProductFlavor)   // 排序商品规格

		privateApi.POST("/product/import/list", wrapper.ImportProductList) // 获取导入商品列表
		privateApi.POST("/product/import", wrapper.ImportProduct)          // 导入商品

		// 获取单规格商品列表
		privateApi.GET("/product/single/list", wrapper.GetProductSingleList) // 获取单规格商品列表

		privateApi.GET("/product/list", wrapper.GetProductShopList)              // 获取商品列表
		privateApi.GET("/product/printer_list", wrapper.GetProductPrinterList)   // 获取商品打印机列表
		privateApi.POST("/product/sort", wrapper.SortProductShopList)            // 排序商品列表
		privateApi.GET("/product/detail", wrapper.GetProductDetail)              // 获取商品详情
		privateApi.POST("/product/status", wrapper.ProductShopStatus)            // 修改商品状态
		privateApi.POST("/product/add", wrapper.ProductShopAdd)                  // 添加商品
		privateApi.POST("/product/edit", wrapper.ProductShopEdit)                // 编辑商品
		privateApi.DELETE("/product/delete", wrapper.ProductShopDelete)          // 删除商品
		privateApi.POST("/product/change_price", wrapper.ProductShopChangePrice) // 商品改价
		privateApi.GET("/product/tax/list", wrapper.ProductTaxList)              // 获取商品税类列表
		privateApi.POST("/product/upload_image", wrapper.UploadProductImage)     // 上传商品图片

		// 外卖商品
		privateApi.POST("/product/takeout/add", wrapper.ProductTakeoutShopAdd)       // 添加外卖商品
		privateApi.POST("/product/takeout/edit", wrapper.ProductTakeoutShopEdit)     // 编辑外卖商品
		privateApi.GET("/product/takeout/detail", wrapper.ProductTakeoutShopDetail)  // 获取外卖商品详情
		privateApi.POST("/product/takeout/delete", wrapper.ProductTakeoutShopDelete) // 删除外卖商品

		// 外卖商品批量操作
		privateApi.POST("/takeout/products/batch_create", wrapper.TakeoutBatchCreate)   // 批量创建外卖商品
		privateApi.POST("/takeout/products/batch_online", wrapper.TakeoutBatchOnline)   // 批量上架外卖商品
		privateApi.POST("/takeout/products/batch_offline", wrapper.TakeoutBatchOffline) // 批量下架外卖商品
		privateApi.POST("/takeout/products/batch_delete", wrapper.TakeoutBatchDelete)   // 批量删除外卖商品
	}
}
