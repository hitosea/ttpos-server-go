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

// BatchProductHandler 分批类型处理程序
type BatchProductHandler struct {
	productSrv service.IProductSrv // 商品服务
}

// GetBatchTagList 获取分批类型列表
// @Summary 获取分批类型列表
// @Description 获取分批类型列表
// @Tags 商家端.分批类型
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.BatchTagList "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/tag/list [get]
func (h *BatchProductHandler) GetBatchTagList(c *gin.Context) {
	ctx := helper.GetContext(c)
	listReq := req.BatchTagListReq{}

	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}
	res, err := h.productSrv.GetBatchTagList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetProductBatchType 获取分批类型详情
// @Summary 获取分批类型详情
// @Description 获取分批类型详情
// @Tags 商家端.分批类型
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "分批类型UUID"
// @Success 200 {object} product_resp.BatchTagDetail "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/tag [get]
func (h *BatchProductHandler) GetBatchTag(c *gin.Context) {
	ctx := helper.GetContext(c)
	detailReq := req.BatchTagReq{}
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}
	res, err := h.productSrv.GetBatchTag(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddProductBatchType 添加分批类型
// @Summary 添加分批类型
// @Description 添加分批类型
// @Tags 商家端.分批类型
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.BatchTagAddReq true "分批类型添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/tag/add [post]
func (h *BatchProductHandler) AddBatchTag(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.BatchTagAddReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.AddBatchTag(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "添加成功")
}

// EditProductBatchType 编辑分批类型
// @Summary 编辑分批类型
// @Description 编辑分批类型
// @Tags 商家端.分批类型
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.BatchTagEditReq true "分批类型编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/tag/edit [post]
func (h *BatchProductHandler) EditBatchTag(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.BatchTagEditReq{}
	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.EditBatchTag(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "保存成功")
}

// DeleteProductBatchType 删除分批类型
// @Summary 删除分批类型
// @Description 删除分批类型
// @Tags 商家端.分批类型
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.BatchTagDeleteReq true "分批类型删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/tag [delete]
func (h *BatchProductHandler) DeleteBatchTag(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.BatchTagDeleteReq{}
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.DeleteBatchTag(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "删除成功")
}

// SortProductBatchType 排序分批类型
// @Summary 排序分批类型
// @Description 排序分批类型
// @Tags 商家端.分批类型
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.BatchTagSortReq true "分批类型排序请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/tag/sort [post]
func (h *BatchProductHandler) SortBatchTag(c *gin.Context) {
	ctx := helper.GetContext(c)
	sortReq := req.BatchTagSortReq{}
	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.SortBatchTag(ctx, sortReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "保存成功")
}

// GetProductBatchTypeColorUsage 获取色块被选择情况
// @Summary 获取色块被选择情况
// @Description 获取色块被选择情况
// @Tags 商家端.分批类型
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} []product_resp.BatchTagColorUsageList "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/tag/color_usage [get]
func (h *BatchProductHandler) GetBatchTagColorUsage(c *gin.Context) {
	ctx := helper.GetContext(c)

	res, err := h.productSrv.GetBatchTagColorUsage(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// SaveBatchProduct 保存分批商品
// @Summary 保存分批商品
// @Description 保存分批商品
// @Tags 商家端.分批商品
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.SaveBatchProductReq true "分批商品保存请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/batch/product/save [post]
func (h *BatchProductHandler) SaveBatchProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	saveReq := req.SaveBatchProductReq{}
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.productSrv.SaveBatchProduct(ctx, saveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "保存成功")
}

func RegisterBatchProductHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	batchHandler := BatchProductHandler{
		productSrv: service.NewProductSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
			settingSrv,             // 设置服务
			cache,
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 分批类型相关接口
		privateApi.GET("/batch/tag/list", batchHandler.GetBatchTagList)              // 获取分批类型列表
		privateApi.GET("/batch/tag", batchHandler.GetBatchTag)                       // 获取分批类型详情
		privateApi.POST("/batch/tag/add", batchHandler.AddBatchTag)                  // 添加分批类型
		privateApi.POST("/batch/tag/edit", batchHandler.EditBatchTag)                // 编辑分批类型
		privateApi.DELETE("/batch/tag", batchHandler.DeleteBatchTag)                 // 删除分批类型
		privateApi.POST("/batch/tag/sort", batchHandler.SortBatchTag)                // 排序分批类型
		privateApi.GET("/batch/tag/color_usage", batchHandler.GetBatchTagColorUsage) // 获取色块被选择情况
		privateApi.POST("/batch/product/save", batchHandler.SaveBatchProduct)        // 保存分批商品
	}
}
