package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OrderHandler 认证鉴权控制器
type OrderHandler struct {
	orderSrv service.IOrderSrv
}

// CreateOrder 创建会员端订单
// @Summary 创建会员端订单
// @Description 创建会员端订单
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CreateMemberOrderReq true "详情参数"
// @Success 200 {object} resp.CreateMemberOrderResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/create [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.CreateMemberOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("创建会员端订单", zap.Any("params", params))

	// 创建会员端订单
	res, checkRes, err := h.orderSrv.CreateMemberOrder(ctx, params)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("提交订单检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// SetOrderAddress 设置订单地址
// @Summary 设置订单地址
// @Description 设置订单地址
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.SetMemberOrderAddressReq true "详情参数"
// @Success 200 {object} resp.CreateMemberOrderResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/address [post]
func (h *OrderHandler) SetOrderAddress(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.SetMemberOrderAddressReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("设置订单地址", zap.Any("params", params))

	// 设置订单地址
	res, err := h.orderSrv.SetMemberOrderAddress(ctx, params)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// PayOrder 提交支付
// @Summary 提交支付
// @Description 提交支付
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.PayMemberOrderReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/pay [post]
func (h *OrderHandler) PayOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.PayMemberOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("提交支付", zap.Any("params", params))

	// 提交支付
	err := h.orderSrv.PayMemberOrder(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// PayOrderStatus 获取支付信息
// @Summary 获取支付信息
// @Description 获取支付信息
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query string true "会员端销售订单UUID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/pay/info [get]
func (h *OrderHandler) PayOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.GetMemberOrderPayInfoReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取支付信息
	res, err := h.orderSrv.GetMemberOrderPayInfo(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// PayOrderStatus 获取支付状态
// @Summary 获取支付状态
// @Description 获取支付状态
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query string true "会员端销售订单UUID"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/pay/status [get]
func (h *OrderHandler) PayOrderStatus(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.GetMemberOrderPayStatusReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取支付信息
	res, err := h.orderSrv.GetMemberOrderPayStatus(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// PaidOrder 支付成功
// @Summary 支付成功
// @Description 支付成功
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.PaidMemberOrderReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/xie-test/order/paid [post]
func (h *OrderHandler) PaidOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.PaidMemberOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("支付成功", zap.Any("params", params))

	// 支付成功
	if err := h.orderSrv.PaidMemberOrder(ctx, params); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

// GetMemberOrderList 获取会员端订单列表
// @Summary 获取会员端订单列表
// @Description 获取会员端订单列表
// @Tags 会员端-订单列表
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MemberOrderListReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/list [get]
func (h *OrderHandler) GetMemberOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.MemberOrderListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取会员端订单列表
	res, err := h.orderSrv.GetMemberOrderList(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

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
	// 初始化处理器
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, service.NewMustPlanSrv(dbm), service.NewPaymentMethodSrv(dbm, settingSrv), service.NewMemberSrv(dbm, cache), service.NewCashBoxSrv(dbm))
	wrapper := &OrderHandler{
		orderSrv: orderSrv,
	}
	// 需要认证
	// privateApi := router.Group("")
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.POST("/order/create", wrapper.CreateOrder)       // 创建订单
		privateApi.POST("/order/address", wrapper.SetOrderAddress)  // 设置订单地址
		privateApi.POST("/order/pay", wrapper.PayOrder)             // 提交支付
		privateApi.GET("/order/pay/info", wrapper.PayOrderInfo)     // 获取支付信息
		privateApi.GET("/order/pay/status", wrapper.PayOrderStatus) // 获取支付状态
		privateApi.GET("/order/list", wrapper.GetMemberOrderList)   // 获取会员端订单列表
		privateApi.POST("/xie-test/order/paid", wrapper.PaidOrder)  // 支付成功 TODO 上线前删除改接口
	}
}
