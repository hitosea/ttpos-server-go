package shop

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

// OrderHandler 商家端处理程序
type OrderHandler struct {
	service     service.IOrderSrv // 订单服务
	deskService service.IDeskSrv  // 桌台服务
}

// GetShopOrderList 处理获取订单列表
// @Summary 获取订单列表
// @Description 获取订单列表
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderListReq true "列表参数"
// @Success 200 {object} resp.OrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/list [get]
func (h *OrderHandler) GetShopOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	orderListReq := req.OrderListReq{}
	if err := c.ShouldBindQuery(&orderListReq); err != nil {
		helper.HandleValidationError(c, err, orderListReq, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.service.GetOrderLists(ctx, orderListReq)
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
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderInfoReq true "详情参数"
// @Success 200 {object} resp.OrderInfosResp "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/info [get]
func (h *OrderHandler) GetOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	orderInfoReq := req.OrderInfoReq{}
	if err := c.ShouldBindQuery(&orderInfoReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 获取收银产品列表
	res, err := h.service.GetOrderInfos(ctx, orderInfoReq)
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
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	orderCancelReq := req.OrderCancelReq{}
	if err := c.ShouldBindJSON(&orderCancelReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionShopOrderCancel).
		WithPath(c.Request.URL.Path)
	//
	err := h.service.CancelOrder(ctx, orderCancelReq)
	// 记录耗时
	tracker.End(ctx, err)
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
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDeleteReq true "详情参数"
// @Success 200 {object} nil "删除订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/delete [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	orderDeleteReq := req.OrderDeleteReq{}
	if err := c.ShouldBindJSON(&orderDeleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionShopOrderDelete).
		WithBill(orderDeleteReq.SaleBillUuid, orderDeleteReq.SaleOrderUuid).
		WithPath(c.Request.URL.Path)
	//
	err := h.service.DeleteOrder(ctx, companyUuid, orderDeleteReq.SaleBillUuid, orderDeleteReq.SaleOrderUuid)
	// 记录耗时
	tracker.End(ctx, err)
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
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderIsCellCloseReq true "详情参数"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/is_cell_close [get]
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
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err)
			return
		}
	} else if params.SaleBillUuid > 0 {
		productList, err = h.deskService.IsCellCloseInstant(ctx, params.SaleBillUuid)
		if productList != nil {
			helper.FailWithData(c, constant.CodeOrderCheckProductCooking, &productList, err)
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

// ReturnOrderInfo 获取退款信息
// @Summary 获取退款信息
// @Description 获取退款信息
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderReturnInfoReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.OrderReturnInfoResp}
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/return [get]
func (h *OrderHandler) ReturnOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReturnInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	if ctx.GetCompany().IsOpenErp() && ctx.GetStaff().CashierOnline == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("当前不可进行退款"))
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
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/return [post]
func (h *OrderHandler) ReturnOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReturnReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	if ctx.GetCompany().IsOpenErp() && ctx.GetStaff().CashierOnline == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("当前不可进行退款"))
		return
	}
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionShopOrderReturn).
		WithPath(c.Request.URL.Path)
	//
	err, codeFail := h.service.ReturnOrder(ctx, req)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// ReturnOrder 处理退款订单
// @Summary 重新退款
// @Description 重新退款
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderReReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/re_return [post]
func (h *OrderHandler) ReReturnOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderReReturnReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionShopOrderReReturn).
		WithPath(c.Request.URL.Path)
	//
	err, codeFail := h.service.ReReturnOrder(ctx, req)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// RejectAllH5Order 拒单商家的所有待接单h5订单
// @Summary 拒单商家的所有待接单h5订单
// @Description 拒单商家的所有待接单h5订单
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{}
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/reject_all [post]
func (h *OrderHandler) RejectAllH5Order(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionShopOrderRejectAll).
		WithPath(c.Request.URL.Path)

	err := h.service.RejectAllH5OrderInShop(ctx)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, gin.H{})
}

// GetShopOrderList 处理获取订单列表
// @Summary 获取订单列表
// @Description 获取订单列表
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderListReq true "列表参数"
// @Success 200 {object} resp.OrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/export [get]
func (h *OrderHandler) ExportShopOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	orderListReq := req.OrderListReq{}
	if err := c.ShouldBindQuery(&orderListReq); err != nil {
		helper.HandleValidationError(c, err, orderListReq, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.service.ExportOrderLists(ctx, orderListReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetPaymentAmount 处理获取实付金额
// @Summary 获取实付金额
// @Description 获取实付金额
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPaymentAmountReq true "详情参数"
// @Success 200 {object} resp.GetPaymentAmountResp "实付金额"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/payment_amount [post]
func (h *OrderHandler) GetPaymentAmount(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderPaymentAmountReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 获取实付金额
	res := h.service.GetPaymentAmount(ctx, req)
	// 返回结果
	helper.Success(c, res)
}

// RegisterOrderHandlers 注册商家订单路由
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
		service:     orderSrv,
		deskService: service.NewDeskSrv(dbm, service.NewLocaleSrv(), orderSrv, settingSrv, deviceSrv, mustPlanSrv),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order/list", wrapper.GetShopOrderList)
		privateApi.GET("/order/info", wrapper.GetOrderInfo)
		privateApi.GET("/order/export", wrapper.ExportShopOrderList)
		privateApi.GET("/order/is_cell_close", wrapper.IsCellClose)
		privateApi.GET("/order/return", wrapper.ReturnOrderInfo)
		privateApi.POST("/order/cancel", wrapper.CancelOrder)
		privateApi.POST("/order/return", wrapper.ReturnOrder)
		privateApi.DELETE("/order/delete", wrapper.DeleteOrder)
		privateApi.POST("/order/re_return", wrapper.ReReturnOrder)
		privateApi.POST("/order/reject_all", wrapper.RejectAllH5Order)
		privateApi.POST("/order/payment_amount", wrapper.GetPaymentAmount)
	}
}
