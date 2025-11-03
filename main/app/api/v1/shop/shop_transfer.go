package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/app/service/transfer_order"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// TransferOrderHandler 调拨单控制器
type TransferOrderHandler struct {
	authSrv          service.IAuthSrv
	transferOrderSrv transfer_order.ITransferOrderSrv
}

// GetTransferOrderList 获取调拨单列表
// @Summary 获取调拨单列表
// @Description 分页获取调拨单列表
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderListReq true "调拨单列表请求参数"
// @Success 200 {object} dto.Response{data=resp.TransferOrderListResp} "成功"
// @Router /shop/transfer/order/list [get]
func (h *TransferOrderHandler) GetTransferOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.TransferOrderListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.transferOrderSrv.GetTransferOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetTransferOrderDetail 获取调拨单详情
// @Summary 获取调拨单详情
// @Description 根据UUID获取调拨单详情
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "调拨单UUID"
// @Success 200 {object} dto.Response{data=resp.TransferOrderDetailResp} "成功"
// @Router /shop/transfer/order/detail [get]
func (h *TransferOrderHandler) GetTransferOrderDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.TransferOrderDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.transferOrderSrv.GetTransferOrderDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// CreateTransferOrder 创建调拨单
// @Summary 创建调拨单
// @Description 创建新的调拨单
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderCreateReq true "创建调拨单请求参数"
// @Success 200 {object} dto.Response{data=resp.TransferOrderCreateResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/transfer/order/create [post]
func (h *TransferOrderHandler) CreateTransferOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.TransferOrderCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	resp, err := h.transferOrderSrv.CreateTransferOrder(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// UpdateTransferOrder 更新调拨单
// @Summary 更新调拨单
// @Description 更新调拨单信息（仅待提交状态可更新）
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderUpdateReq true "更新调拨单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/transfer/order/update [post]
func (h *TransferOrderHandler) UpdateTransferOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.TransferOrderUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	err := h.transferOrderSrv.UpdateTransferOrder(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// DeleteTransferOrder 删除调拨单
// @Summary 删除调拨单
// @Description 删除调拨单（仅待提交状态可删除）
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderDeleteReq true "删除调拨单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/transfer/order/delete [delete]
func (h *TransferOrderHandler) DeleteTransferOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.TransferOrderDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.transferOrderSrv.DeleteTransferOrder(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// SubmitTransferOrder 提交调拨单
// @Summary 提交调拨单
// @Description 提交调拨单进入审批流程
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderSubmitReq true "提交调拨单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/transfer/order/submit [post]
func (h *TransferOrderHandler) SubmitTransferOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var submitReq req.TransferOrderSubmitReq
	if err := c.ShouldBindJSON(&submitReq); err != nil {
		helper.HandleValidationError(c, err, submitReq, nil)
		return
	}

	err := h.transferOrderSrv.SubmitTransferOrder(ctx, submitReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// ApproveTransferOrder 审批通过调拨单
// @Summary 审批通过调拨单
// @Description 审批通过调拨单
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderApproveReq true "审批调拨单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/transfer/order/approve [post]
func (h *TransferOrderHandler) ApproveTransferOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var approveReq req.TransferOrderApproveReq
	if err := c.ShouldBindJSON(&approveReq); err != nil {
		helper.HandleValidationError(c, err, approveReq, nil)
		return
	}

	err := h.transferOrderSrv.ApproveTransferOrder(ctx, approveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// RejectTransferOrder 驳回调拨单
// @Summary 驳回调拨单
// @Description 驳回调拨单
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderRejectReq true "驳回调拨单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/transfer/order/reject [post]
func (h *TransferOrderHandler) RejectTransferOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var rejectReq req.TransferOrderRejectReq
	if err := c.ShouldBindJSON(&rejectReq); err != nil {
		helper.HandleValidationError(c, err, rejectReq, nil)
		return
	}

	err := h.transferOrderSrv.RejectTransferOrder(ctx, rejectReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// ReceiveTransferOrder 收货调拨单
// @Summary 收货调拨单
// @Description 确认收货完成调拨
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderReceiveReq true "收货调拨单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/transfer/order/receive [post]
func (h *TransferOrderHandler) ReceiveTransferOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var receiveReq req.TransferOrderReceiveReq
	if err := c.ShouldBindJSON(&receiveReq); err != nil {
		helper.HandleValidationError(c, err, receiveReq, nil)
		return
	}

	err := h.transferOrderSrv.ReceiveTransferOrder(ctx, receiveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// GetTransferOrderApprovalList 获取调拨单审批流程列表
// @Summary 获取调拨单审批流程列表
// @Description 获取调拨单的审批流程记录
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param transfer_order_uuid query int true "调拨单UUID"
// @Success 200 {object} dto.Response{data=resp.TransferOrderApprovalListResp} "成功"
// @Router /shop/transfer/order/approval/list [get]
func (h *TransferOrderHandler) GetTransferOrderApprovalList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.TransferOrderApprovalListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.transferOrderSrv.GetTransferOrderApprovalList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetTransferOrderLogList 获取调拨单操作日志列表
// @Summary 获取调拨单操作日志列表
// @Description 分页获取调拨单的操作日志
// @Tags 商家端.调拨单管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.TransferOrderLogListReq true "操作日志列表请求参数"
// @Success 200 {object} dto.Response{data=resp.TransferOrderLogListResp} "成功"
// @Router /shop/transfer/order/log/list [get]
func (h *TransferOrderHandler) GetTransferOrderLogList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.TransferOrderLogListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.transferOrderSrv.GetTransferOrderLogList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// RegisterTransferOrderHandlers 注册调拨单相关路由
func RegisterTransferOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 调拨单服务
	transferOrderSrv := transfer_order.NewTransferOrderSrv(dbm)

	wrapper := &TransferOrderHandler{
		authSrv:          authSrv,
		transferOrderSrv: transferOrderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.MinVersionCheck("2.9.0"), middleware.Auth(authSrv, dbm))
	{
		// 调拨单管理
		privateApi.GET("/transfer/order/list", wrapper.GetTransferOrderList)
		privateApi.GET("/transfer/order/detail", wrapper.GetTransferOrderDetail)
		privateApi.POST("/transfer/order/create", wrapper.CreateTransferOrder)
		privateApi.POST("/transfer/order/update", wrapper.UpdateTransferOrder)
		privateApi.DELETE("/transfer/order/delete", wrapper.DeleteTransferOrder)
		privateApi.POST("/transfer/order/submit", wrapper.SubmitTransferOrder)
		privateApi.POST("/transfer/order/approve", wrapper.ApproveTransferOrder)
		privateApi.POST("/transfer/order/reject", wrapper.RejectTransferOrder)
		privateApi.POST("/transfer/order/receive", wrapper.ReceiveTransferOrder)

		// 审批流程和操作日志
		privateApi.GET("/transfer/order/approval/list", wrapper.GetTransferOrderApprovalList)
		privateApi.GET("/transfer/order/log/list", wrapper.GetTransferOrderLogList)
	}
}
