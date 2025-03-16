package cashier

import (
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
	service     service.IOrderSrv // 订单服务
	deskService service.IDeskSrv  // 桌台服务
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
	res, err := h.service.GetOrderLists(ctx, req)
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
	res, err := h.service.GetOrderInfos(ctx, req)
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
	err := h.service.CancelOrder(ctx, req)
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
	res, err := h.service.GetReturnOrderInfo(ctx, req)
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
	err := h.service.ReturnOrder(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
	res, err := h.service.GetReverseSettleInfo(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
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
	err := h.service.ReverseSettle(ctx, req)
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
	err := h.service.DeleteOrder(ctx, companyUuid, req.SaleBillUuid, req.SaleOrderUuid)
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
		_, productList, err = h.deskService.IsCellCloseDesk(ctx, params.DeskUuid)
		if productList != nil {
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err.Error())
			return
		}
	} else if params.SaleBillUuid > 0 {
		productList, err = h.deskService.IsCellCloseInstant(ctx, params.SaleBillUuid)
		if productList != nil {
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err.Error())
			return
		}
	} else {
		err = errors.New("参数错误")
	}
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
// @Router /cashier/order/print [post]
func (h *OrderHandler) OrderPrint(c *gin.Context) {
	var printReq req.OrderPrintReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.service.OrderPrint(ctx, printReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
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
	res, err := h.service.OrderPrintInvoice(ctx, printReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
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
// @Success 200 {object} dto.Response{data=resp.PrinterData} "发票信息"
// @Router /cashier/order/invoice [get]
func (h *OrderHandler) OrderInvoiceInfo(c *gin.Context) {
	var invoiceReq req.OrderInvoiceInfoReq
	if err := c.ShouldBindQuery(&invoiceReq); err != nil {
		helper.HandleValidationError(c, err, invoiceReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res := h.service.OrderPrintInvoiceInfo(ctx, invoiceReq)
	helper.Success(c, res)
}

// RegisterOrderHandlers 注册收银订单路由
func RegisterOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, mustPlanSrv)

	// 初始化处理器
	wrapper := OrderHandler{
		service:     orderSrv,
		deskService: service.NewDeskSrv(dbm, service.NewLocaleSrv(), orderSrv, settingSrv, deviceSrv),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order/list", wrapper.GetCashierOrderList)         // 获取订单列表
		privateApi.GET("/order/info", wrapper.GetOrderInfo)                // 获取订单详情
		privateApi.POST("/order/cancel", wrapper.CancelOrder)              // 取消订单
		privateApi.GET("/order/return", wrapper.ReturnOrderInfo)           // 获取退款弹窗信息
		privateApi.POST("/order/return", wrapper.ReturnOrder)              // 整单退款或部分退款
		privateApi.GET("/order/reverse_settle", wrapper.ReverseSettleInfo) // 获取反结账弹窗信息
		privateApi.POST("/order/reverse_settle", wrapper.ReverseSettle)    // 反结账
		privateApi.DELETE("/order/delete", wrapper.DeleteOrder)            // 删除订单
		privateApi.GET("/order/is_cell_close", wrapper.IsCellClose)        // 判断订单是否可关闭
		privateApi.POST("/order/print", wrapper.OrderPrint)                // 打印
		privateApi.POST("/order/print/invoice", wrapper.OrderPrintInvoice) // 打印发票
		privateApi.GET("/order/invoice", wrapper.OrderInvoiceInfo)         // 获取发票信息
	}
}
