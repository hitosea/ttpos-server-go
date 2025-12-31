package cashier

import (
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// OrderHandler 收银点餐处理程序
type OrderHandler struct {
	orderSrv     service.IOrderSrv     // 订单服务
	deskSrv      service.IDeskSrv      // 桌台服务
	saasStaffSrv service.ISaasStaffSrv // 统一账号服务
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
// @Router /cashier/order/list [get]
func (h *OrderHandler) GetCashierOrderList(c *gin.Context) {
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
// @Router /cashier/order/info [get]
func (h *OrderHandler) GetOrderInfo(c *gin.Context) {
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
// @Router /cashier/order/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
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
// @Router /cashier/order/return [get]
func (h *OrderHandler) ReturnOrderInfo(c *gin.Context) {
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
// @Router /cashier/order/return [post]
func (h *OrderHandler) ReturnOrder(c *gin.Context) {
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
	if codeFail == constant.CodeSuccessOpenCashBox {
		helper.Success(c, gin.H{"is_open_cash_box": true})
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// ReturnOrder 处理退款订单
// @Summary 重新退款
// @Description 重新退款
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderReReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/re_return [post]
func (h *OrderHandler) ReReturnOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReReturnReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	err, codeFail := h.orderSrv.ReReturnOrder(ctx, req)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// ReverseSettleInfo 获取反结账弹窗信息
// @Summary 获取反结账弹窗信息
// @Description 获取反结账弹窗信息
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderReverseSettleInfoReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.OrderReverseSettleInfoResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/reverse_settle [get]
func (h *OrderHandler) ReverseSettleInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReverseSettleInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	res, err := h.orderSrv.GetReverseSettleInfo(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CheckAuthorization 检查授权
// @Summary 检查授权
// @Description 检查当前员工是否有权限进行折扣/退款操作
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CheckAuthorizationReq true "检查授权参数"
// @Success 200 {object} dto.Response{data=resp.CheckAuthorizationResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/check_authorization [post]
func (h *OrderHandler) CheckAuthorization(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.CheckAuthorizationReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 调用 Service 检查授权
	hasPermission, err := h.orderSrv.CheckAuthorization(ctx, req.OperationType)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, resp.CheckAuthorizationResp{
		HasPermission: hasPermission,
	})
}

// VerifyPassword 密码验证
// @Summary 密码验证
// @Description 验证授权员工账号和密码
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordForSensitiveOperationReq true "密码验证参数"
// @Success 200 {object} dto.Response{data=resp.VerifyPasswordResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/verify_password [post]
func (h *OrderHandler) VerifyPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.VerifyPasswordForSensitiveOperationReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 调用 Service 验证密码
	verified, err := h.orderSrv.VerifyPassword(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, resp.VerifyPasswordResp{
		Verified: verified,
	})
}

// QueryStaffByContact 根据邮箱或手机号查询员工
// @Summary 根据邮箱或手机号查询员工
// @Description 根据邮箱或手机号查询员工，支持模糊搜索，返回员工基本信息，用于下拉列表展示。根据操作类型（operation_type）区分优惠折扣场景和退款场景，不同场景返回对应授权列表中的员工。
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param operation_type query string true "操作类型，discount-折扣操作，refund-退款操作"
// @Param keyword query string false "搜索关键词（邮箱或手机号，支持模糊匹配）"
// @Success 200 {object} dto.Response{data=resp.QueryStaffByContactResp}
// @Router /cashier/order/query_staff [get]
func (h *OrderHandler) QueryStaffByContact(c *gin.Context) {
	ctx := helper.GetContext(c)
	var queryReq req.QueryStaffByContactReq
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		helper.HandleValidationError(c, err, queryReq, nil)
		return
	}

	res, err := h.saasStaffSrv.QueryStaffByContact(ctx, queryReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}

	helper.Success(c, res)
}

// ReverseSettle 处理反结账
// @Summary 反结账
// @Description 反结账
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderReverseSettleReq true "详情参数"
// @Success 200 {object} nil "反结账成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/reverse_settle [post]
func (h *OrderHandler) ReverseSettle(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReverseSettleReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	err := h.orderSrv.ReverseSettle(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
// @Router /cashier/order/delete [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
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

// IsCellClose 判断订单是否可关闭
// @Summary 判断订单是否可关闭
// @Description 判断订单是否可关闭
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderIsCellCloseReq true "详情参数"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/is_cell_close [get]
func (h *OrderHandler) IsCellClose(c *gin.Context) {
	ctx := helper.GetContext(c)
	//
	params := req.OrderIsCellCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	var err error
	var productList *resp.CartProductList
	if params.DeskUuid > 0 {
		_, productList, err = h.deskSrv.IsCellCloseDesk(ctx, params.DeskUuid)
		if productList != nil {
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err)
			return
		}
	} else if params.SaleBillUuid > 0 {
		productList, err = h.deskSrv.IsCellCloseInstant(ctx, params.SaleBillUuid)
		if productList != nil {
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err)
			return
		}
	} else {
		err = errors.New("参数错误")
	}
	if err != nil {
		if strings.Contains(err.Error(), "订单已结账") {
			err = errors.New("当前订单已被部分支付，无法整单取消")
		}
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
// @Router /cashier/order/print [post]
func (h *OrderHandler) OrderPrint(c *gin.Context) {
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
// @Router /cashier/order/print/invoice [post]
func (h *OrderHandler) OrderPrintInvoice(c *gin.Context) {
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
// @Router /cashier/order/invoice [get]
func (h *OrderHandler) OrderInvoiceInfo(c *gin.Context) {
	var invoiceReq req.OrderInvoiceInfoReq
	if err := c.ShouldBindQuery(&invoiceReq); err != nil {
		helper.HandleValidationError(c, err, invoiceReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res := h.orderSrv.OrderPrintInvoiceInfo(ctx, invoiceReq)
	helper.Success(c, res)
}

// RegisterOrderHandlers 注册收银订单路由
func RegisterOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	// 初始化处理器
	wrapper := OrderHandler{
		orderSrv:     orderSrv,
		deskSrv:      service.NewDeskSrv(dbm, service.NewLocaleSrv(), orderSrv, settingSrv, deviceSrv, mustPlanSrv),
		saasStaffSrv: service.NewSaasStaffSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order/list", wrapper.GetCashierOrderList)                // 获取订单列表
		privateApi.GET("/order/info", wrapper.GetOrderInfo)                       // 获取订单详情
		privateApi.POST("/order/cancel", wrapper.CancelOrder)                     // 取消订单
		privateApi.GET("/order/return", wrapper.ReturnOrderInfo)                  // 获取退款弹窗信息
		privateApi.POST("/order/return", wrapper.ReturnOrder)                     // 整单退款或部分退款
		privateApi.POST("/order/re_return", wrapper.ReReturnOrder)                // 重新退款
		privateApi.GET("/order/reverse_settle", wrapper.ReverseSettleInfo)        // 获取反结账弹窗信息
		privateApi.POST("/order/reverse_settle", wrapper.ReverseSettle)           // 反结账
		privateApi.DELETE("/order/delete", wrapper.DeleteOrder)                   // 删除订单
		privateApi.GET("/order/is_cell_close", wrapper.IsCellClose)               // 判断订单是否可关闭
		privateApi.POST("/order/print", wrapper.OrderPrint)                       // 打印
		privateApi.POST("/order/print/invoice", wrapper.OrderPrintInvoice)        // 打印发票
		privateApi.GET("/order/invoice", wrapper.OrderInvoiceInfo)                // 获取发票信息
		privateApi.POST("/order/check_authorization", wrapper.CheckAuthorization) // 检查授权
		privateApi.POST("/order/verify_password", wrapper.VerifyPassword)         // 密码验证
		privateApi.GET("/order/query_staff", wrapper.QueryStaffByContact)         // 根据邮箱或手机号查询员工
	}
}
