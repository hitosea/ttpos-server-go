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

// ProductLabelHandler 商品标签处理程序
type ProductLabelHandler struct {
	productLabelSrv service.IProductLabelSrv // 商品标签服务
}

// GetProductLabelList 获取商品标签列表
// @Summary 获取商品标签列表
// @Description 获取商品标签列表，包含关联的商品信息
// @Tags 商家端.商品标签
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.ProductLabelListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/label/list [get]
func (h *ProductLabelHandler) GetProductLabelList(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 获取商品标签列表
	res, err := h.productLabelSrv.GetProductLabelList(ctx)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// AddProductLabel 添加商品标签
// @Summary 添加商品标签
// @Description 添加商品标签
// @Tags 商家端.商品标签
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductLabelAddReq true "商品标签添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/label/add [post]
func (h *ProductLabelHandler) AddProductLabel(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductLabelAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, nil)
		return
	}
	// 添加商品标签
	err := h.productLabelSrv.AddProductLabel(ctx, addReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{}, "保存成功")
}

// EditProductLabel 编辑商品标签
// @Summary 编辑商品标签
// @Description 编辑商品标签
// @Tags 商家端.商品标签
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductLabelEditReq true "商品标签编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/label/edit [post]
func (h *ProductLabelHandler) EditProductLabel(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.ProductLabelEditReq{}

	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, nil)
		return
	}
	// 编辑商品标签
	err := h.productLabelSrv.EditProductLabel(ctx, editReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{}, "保存成功")
}

// DeleteProductLabel 删除商品标签
// @Summary 删除商品标签
// @Description 删除商品标签
// @Tags 商家端.商品标签
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductLabelDeleteReq true "商品标签删除请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product/label/delete [post]
func (h *ProductLabelHandler) DeleteProductLabel(c *gin.Context) {
	ctx := helper.GetContext(c)
	deleteReq := req.ProductLabelDeleteReq{}

	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}
	// 删除商品标签
	err := h.productLabelSrv.DeleteProductLabel(ctx, deleteReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{}, "删除成功")
}

// RegisterProductLabelHandlers 注册商品标签路由
func RegisterProductLabelHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 创建商品标签处理程序
	handler := ProductLabelHandler{
		productLabelSrv: service.NewProductLabelSrv(dbm),
	}
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/product/label/list", handler.GetProductLabelList)   // 获取商品标签列表
		privateApi.POST("/product/label/add", handler.AddProductLabel)       // 添加商品标签
		privateApi.POST("/product/label/edit", handler.EditProductLabel)     // 编辑商品标签
		privateApi.POST("/product/label/delete", handler.DeleteProductLabel) // 删除商品标签
	}
}
