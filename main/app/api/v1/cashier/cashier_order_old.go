package cashier

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

// OrderOldHandler 收银点餐处理程序
type OrderOldHandler struct {
	orderSrv service.IOrderSrv // 订单服务
	deskSrv  service.IDeskSrv  // 桌台服务
}

// GetCashierOrderList 处理获取订单列表
// @Summary 获取订单列表
// @Description 获取订单列表
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderListReq true "列表参数"
// @Success 200 {object} resp.OrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/old/list [get]
func (h *OrderOldHandler) GetCashierOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.orderSrv.GetOrderLists(ctx, req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetOrderInfo 处理获取订单详情
// @Summary 获取订单详情
// @Description 获取订单详情
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderInfoReq true "详情参数"
// @Success 200 {object} resp.OrderInfosResp "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/info [get]
func (h *OrderOldHandler) GetOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 获取收银产品列表
	res, err := h.orderSrv.GetOrderInfos(ctx, req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CancelOrder 处理取消订单
// @Summary 取消订单
// @Description 取消订单
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/cancel [post]
func (h *OrderOldHandler) CancelOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderCancelReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 订单列表中取消订单不需要密码
	req.NotNeedPassword = true
	err := h.orderSrv.CancelOrder(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// ReturnOrderInfo 获取退款信息
// @Summary 获取退款信息
// @Description 获取退款信息
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderReturnInfoReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.OrderReturnInfoResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/return [get]
func (h *OrderOldHandler) ReturnOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReturnInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	res, err := h.orderSrv.GetReturnOrderInfo(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// ReturnOrder 处理退款订单
// @Summary 退款订单
// @Description 退款订单
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/return [post]
func (h *OrderOldHandler) ReturnOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReturnReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	err, codeFail := h.orderSrv.ReturnOrder(ctx, req)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// DeleteOrder 处理删除订单
// @Summary 删除订单
// @Description 删除订单
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDeleteReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/old/order/delete [delete]
func (h *OrderOldHandler) DeleteOrder(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderDeleteReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	err := h.orderSrv.DeleteOrder(ctx, companyUuid, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderPrint 打印小票
// @Summary 打印小票
// @Description 打印小票
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/old/order/print [post]
func (h *OrderOldHandler) OrderPrint(c *gin.Context) {
	var printReq req.OrderPrintReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderPrint(ctx, printReq, false)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res, "发送成功")
}

// OrderPrint 打印发票
// @Summary 打印发票
// @Description 打印发票
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintInvoiceReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/old/order/print/invoice [post]
func (h *OrderOldHandler) OrderPrintInvoice(c *gin.Context) {
	var printReq req.OrderPrintInvoiceReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderPrintInvoice(ctx, printReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res, "发送成功")
}

// OrderInvoiceInfo 获取发票信息
// @Summary 获取发票信息
// @Description 获取发票信息
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @Success 200 {object} dto.Response{data=resp.SaleOrderInvoiceInfo} "发票信息"
// @Router /cashier/old/order/invoice [get]
func (h *OrderOldHandler) OrderInvoiceInfo(c *gin.Context) {
	var invoiceReq req.OrderInvoiceInfoReq
	if err := c.ShouldBindQuery(&invoiceReq); err != nil {
		helper.HandleValidationError(c, err, invoiceReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res := h.orderSrv.OrderPrintInvoiceInfo(ctx, invoiceReq)
	helper.Success(c, res)
}

// RegisterOrderOldHandler 注册收银订单路由
func RegisterOrderOldHandler(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv)

	// 初始化处理器
	wrapper := OrderOldHandler{
		orderSrv: orderSrv,
		deskSrv:  service.NewDeskSrv(dbm, service.NewLocaleSrv(), orderSrv, settingSrv, deviceSrv, mustPlanSrv),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/old/order/list", wrapper.GetCashierOrderList)         // 获取订单列表
		privateApi.GET("/old/order/info", wrapper.GetOrderInfo)                // 获取订单详情
		privateApi.POST("/old/order/cancel", wrapper.CancelOrder)              // 取消订单
		privateApi.GET("/old/order/return", wrapper.ReturnOrderInfo)           // 获取退款弹窗信息
		privateApi.POST("/old/order/return", wrapper.ReturnOrder)              // 整单退款或部分退款
		privateApi.DELETE("/old/order/delete", wrapper.DeleteOrder)            // 删除订单
		privateApi.POST("/old/order/print", wrapper.OrderPrint)                // 打印
		privateApi.POST("/old/order/print/invoice", wrapper.OrderPrintInvoice) // 打印发票
		privateApi.GET("/old/order/invoice", wrapper.OrderInvoiceInfo)         // 获取发票信息
	}
}
