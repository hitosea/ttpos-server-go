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

// PurchaseHandler 采购控制器
type PurchaseHandler struct {
	authSrv          service.IAuthSrv
	purchaseOrderSrv service.IPurchaseOrderSrv
}

// GetPurchaseOrderList 获取采购订单列表
// @Summary 获取采购订单列表
// @Description 分页获取采购订单列表
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderListReq true "采购订单列表请求参数"
// @Success 200 {object} dto.Response{data=resp.PurchaseOrderListResp} "成功"
// @Router /shop/purchase/order/list [get]
func (h *PurchaseHandler) GetPurchaseOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.PurchaseOrderListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetPurchaseOrderDetail 获取采购订单详情
// @Summary 获取采购订单详情
// @Description 根据UUID获取采购订单详情
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "采购订单UUID"
// @Success 200 {object} dto.Response{data=resp.PurchaseOrderDetailResp} "成功"
// @Router /shop/purchase/order/detail [get]
func (h *PurchaseHandler) GetPurchaseOrderDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.PurchaseOrderDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseOrderDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// CreatePurchaseOrder 创建采购订单
// @Summary 创建采购订单
// @Description 创建新的采购订单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderCreateReq true "创建采购订单请求参数"
// @Success 200 {object} dto.Response{data=resp.PurchaseOrderCreateResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/create [post]
func (h *PurchaseHandler) CreatePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.PurchaseOrderCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.CreatePurchaseOrder(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// UpdatePurchaseOrder 更新采购订单
// @Summary 更新采购订单
// @Description 更新采购订单信息
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderUpdateReq true "更新采购订单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/purchase/order/update [post]
func (h *PurchaseHandler) UpdatePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.PurchaseOrderUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	err := h.purchaseOrderSrv.UpdatePurchaseOrder(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// DeletePurchaseOrder 删除采购订单
// @Summary 删除采购订单
// @Description 删除采购订单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderDeleteReq true "删除采购订单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/delete [delete]
func (h *PurchaseHandler) DeletePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.PurchaseOrderDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.purchaseOrderSrv.DeletePurchaseOrder(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// SubmitPurchaseOrder 提交采购订单
// @Summary 提交采购订单
// @Description 提交采购订单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderSubmitReq true "提交采购订单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/submit [post]
func (h *PurchaseHandler) SubmitPurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var statusReq req.PurchaseOrderSubmitReq
	if err := c.ShouldBindJSON(&statusReq); err != nil {
		helper.HandleValidationError(c, err, statusReq, nil)
		return
	}

	err := h.purchaseOrderSrv.SubmitPurchaseOrder(ctx, statusReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// ApprovePurchaseOrder 审核采购订单
// @Summary 审核采购订单
// @Description 审核采购订单（通过或驳回）
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderApproveReq true "审核采购订单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/approve [post]
func (h *PurchaseHandler) ApprovePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var approveReq req.PurchaseOrderApproveReq
	if err := c.ShouldBindJSON(&approveReq); err != nil {
		helper.HandleValidationError(c, err, approveReq, nil)
		return
	}

	err := h.purchaseOrderSrv.ApprovePurchaseOrder(ctx, approveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// CreatePurchaseReceipt 创建收货记录
// @Summary 创建收货记录
// @Description 创建采购收货记录
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseReceiptCreateReq true "创建收货记录请求参数"
// @Success 200 {object} dto.Response{data=resp.PurchaseReceiptCreateResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/create [post]
func (h *PurchaseHandler) CreatePurchaseReceipt(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.PurchaseReceiptCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.CreatePurchaseReceiptOrder(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// UpdatePurchaseReceipt 更新收货记录
// @Summary 更新收货记录
// @Description 更新收货记录
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseReceiptOrderUpdateReq true "更新收货记录请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/update [post]
func (h *PurchaseHandler) UpdatePurchaseReceipt(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.PurchaseReceiptOrderUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	err := h.purchaseOrderSrv.UpdatePurchaseReceiptOrder(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// GetPurchaseReceiptList 获取收货记录列表
// @Summary 获取收货记录列表
// @Description 分页获取收货记录列表
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param order_no query string false "订单号"
// @Param status query int false "状态"
// @Success 200 {object} dto.Response{data=resp.PurchaseReceiptOrderListResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/list [get]
func (h *PurchaseHandler) GetPurchaseReceiptList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.PurchaseReceiptOrderListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseReceiptOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetPurchaseReceiptDetail 获取收货记录详情
// @Summary 获取收货记录详情
// @Description 根据UUID获取收货记录详情
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "收货记录UUID"
// @Success 200 {object} dto.Response{data=resp.PurchaseReceiptOrderDetailResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/detail [get]
func (h *PurchaseHandler) GetPurchaseReceiptDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.PurchaseReceiptOrderDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseReceiptOrderDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// CancelPurchaseReceipt 取消收货单
// @Summary 取消收货单
// @Description 取消收货单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseReceiptOrderCancelReq true "取消收货单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/cancel [post]
func (h *PurchaseHandler) CancelPurchaseReceipt(c *gin.Context) {
	ctx := helper.GetContext(c)
	var cancelReq req.PurchaseReceiptOrderCancelReq
	if err := c.ShouldBindJSON(&cancelReq); err != nil {
		helper.HandleValidationError(c, err, cancelReq, nil)
		return
	}

	err := h.purchaseOrderSrv.CancelPurchaseReceiptOrder(ctx, cancelReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

func RegisterPurchaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 采购服务
	purchaseOrderSrv := service.NewPurchaseOrderSrv(dbm)

	wrapper := &PurchaseHandler{
		authSrv:          authSrv,
		purchaseOrderSrv: purchaseOrderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 采购订单管理
		privateApi.GET("/purchase/order/list", wrapper.GetPurchaseOrderList)
		privateApi.GET("/purchase/order/detail", wrapper.GetPurchaseOrderDetail)
		privateApi.POST("/purchase/order/create", wrapper.CreatePurchaseOrder)
		privateApi.POST("/purchase/order/update", wrapper.UpdatePurchaseOrder)
		privateApi.DELETE("/purchase/order/delete", wrapper.DeletePurchaseOrder)
		privateApi.POST("/purchase/order/approve", wrapper.ApprovePurchaseOrder)
		privateApi.POST("/purchase/order/submit", wrapper.SubmitPurchaseOrder)

		// 收货管理
		privateApi.POST("/purchase/receipt/create", wrapper.CreatePurchaseReceipt)
		privateApi.POST("/purchase/receipt/update", wrapper.UpdatePurchaseReceipt)
		privateApi.GET("/purchase/receipt/list", wrapper.GetPurchaseReceiptList)
		privateApi.GET("/purchase/receipt/detail", wrapper.GetPurchaseReceiptDetail)
	}
}
