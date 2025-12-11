package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// PaymentMethodHandler 支付方式管理控制器
type PaymentMethodHandler struct {
	paymentMethodSrv service.IPaymentMethodSrv
}

// NewPaymentMethodHandler 创建支付方式管理控制器
func NewPaymentMethodHandler(paymentMethodSrv service.IPaymentMethodSrv) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		paymentMethodSrv: paymentMethodSrv,
	}
}

// GetList 获取支付方式列表
// @Summary 获取支付方式管理列表
// @Description 获取支付方式管理列表（分页）
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页大小"
// @Success 200 {object} dto.Response{data=resp.PaymentMethodListResp}
// @Router /shop/payment_method/list [get]
func (h *PaymentMethodHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.PaymentMethodManagementListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	resp, err := h.paymentMethodSrv.GetManagementList(ctx, &listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, resp)
}

// GetDetail 获取支付方式详情
// @Summary 获取支付方式详情
// @Description 获取支付方式详情
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query uint64 true "支付方式UUID"
// @Success 200 {object} dto.Response{data=resp.PaymentMethodDetailResp}
// @Router /shop/payment_method/detail [get]
func (h *PaymentMethodHandler) GetDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var getReq req.PaymentMethodGetReq
	if err := c.ShouldBindQuery(&getReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	resp, err := h.paymentMethodSrv.GetDetail(ctx, &getReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, gin.H{"payment_method": resp})
}

// Create 创建支付方式
// @Summary 创建支付方式
// @Description 创建支付方式
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PaymentMethodCreateReq true "创建支付方式参数"
// @Success 200 {object} dto.Response
// @Router /shop/payment_method/create [post]
func (h *PaymentMethodHandler) Create(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.PaymentMethodCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := h.paymentMethodSrv.Create(ctx, &createReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, "创建成功")
}

// Update 更新支付方式
// @Summary 更新支付方式
// @Description 更新支付方式（包括状态）
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PaymentMethodUpdateReq true "更新支付方式参数"
// @Success 200 {object} dto.Response
// @Router /shop/payment_method/update [put]
func (h *PaymentMethodHandler) Update(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.PaymentMethodUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := h.paymentMethodSrv.Update(ctx, &updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, "更新成功")
}

// Delete 删除支付方式
// @Summary 删除支付方式
// @Description 删除支付方式（软删除）
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PaymentMethodDeleteReq true "删除支付方式参数"
// @Success 200 {object} dto.Response
// @Router /shop/payment_method/delete [delete]
func (h *PaymentMethodHandler) Delete(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.PaymentMethodDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := h.paymentMethodSrv.Delete(ctx, &deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, "删除成功")
}

// UpdateSort 批量更新支付方式排序
// @Summary 批量更新支付方式排序
// @Description 批量更新支付方式排序，确保排序值连续
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PaymentMethodSortUpdateReq true "排序更新参数"
// @Success 200 {object} dto.Response
// @Router /shop/payment_method/update_sort [put]
func (h *PaymentMethodHandler) UpdateSort(c *gin.Context) {
	ctx := helper.GetContext(c)
	var sortReq req.PaymentMethodSortUpdateReq
	if err := c.ShouldBindJSON(&sortReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := h.paymentMethodSrv.UpdateSort(ctx, &sortReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, "更新排序成功")
}

// GetLianlianPayConfig 获取 LianlianPay 配置
// @Summary 获取 LianlianPay 配置
// @Description 获取 LianlianPay 配置（敏感字段返回占位符）
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.LianlianPayConfigResp}
// @Router /shop/payment_method/lianlianpay_config [get]
func (h *PaymentMethodHandler) GetLianlianPayConfig(c *gin.Context) {
	ctx := helper.GetContext(c)

	resp, err := h.paymentMethodSrv.GetLianlianPayConfig(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, resp)
}

// UpdateLianlianPayConfig 更新 LianlianPay 配置
// @Summary 更新 LianlianPay 配置
// @Description 更新 LianlianPay 配置（敏感字段加密存储）
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.LianlianPayConfigUpdateReq true "LianlianPay 配置参数"
// @Success 200 {object} dto.Response
// @Router /shop/payment_method/lianlianpay_config [put]
func (h *PaymentMethodHandler) UpdateLianlianPayConfig(c *gin.Context) {
	ctx := helper.GetContext(c)
	var configReq req.LianlianPayConfigUpdateReq
	if err := c.ShouldBindJSON(&configReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := h.paymentMethodSrv.UpdateLianlianPayConfig(ctx, &configReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, "更新配置成功")
}

// RegisterPaymentMethodHandlers 注册支付方式管理路由
func RegisterPaymentMethodHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	settingSrvInstance := settingSrv.NewSrv(dbm, cache)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrvInstance)
	handler := NewPaymentMethodHandler(paymentMethodSrv)

	// 初始化认证服务（复用 RegisterSettingHandlers 中的逻辑）
	captchaSrv := service.NewCaptchaSrv(cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrvInstance, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrvInstance)

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/payment_method/list", handler.GetList)                               // 获取支付方式列表
		privateApi.GET("/payment_method/detail", handler.GetDetail)                           // 获取支付方式详情
		privateApi.POST("/payment_method/create", handler.Create)                             // 创建支付方式
		privateApi.PUT("/payment_method/update", handler.Update)                              // 更新支付方式
		privateApi.DELETE("/payment_method/delete", handler.Delete)                           // 删除支付方式
		privateApi.PUT("/payment_method/update_sort", handler.UpdateSort)                     // 批量更新排序
		privateApi.GET("/payment_method/lianlianpay_config", handler.GetLianlianPayConfig)    // 获取 LianlianPay 配置
		privateApi.PUT("/payment_method/lianlianpay_config", handler.UpdateLianlianPayConfig) // 更新 LianlianPay 配置
	}
}
