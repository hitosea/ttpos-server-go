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

	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionMemberOrderCreate).
		WithPath(c.Request.URL.Path)
	// 创建会员端订单
	res, checkRes, err := h.orderSrv.CreateMemberOrder(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
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

// CreateDineInOrder 创建会员端堂食订单
// @Summary 创建会员端堂食订单
// @Description 创建会员端堂食订单，与收银机即时点餐价格一致，适用于到店自取/堂食场景
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CreateMemberDineInOrderReq true "详情参数"
// @Success 200 {object} resp.CreateInstantOrderResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/create [post]
func (h *OrderHandler) CreateDineInOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.CreateMemberDineInOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("创建会员端堂食订单", zap.Any("params", params))

	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionMemberOrderCreate).
		WithPath(c.Request.URL.Path)
	// 创建会员端堂食订单
	res, err := h.orderSrv.CreateMemberDineInOrder(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// SubmitDineInOrder 先下单后付模式：提交堂食订单到收银机
// @Summary 提交堂食订单到收银机
// @Description 先下单后付模式下，会员端创建订单后可多次加购，最终通过此接口提交生成 H5 订单
// @Tags 会员端-堂食订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.SubmitMemberDineInOrderReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/submit [post]
func (h *OrderHandler) SubmitDineInOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.SubmitMemberDineInOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	ctx.Log().Debug("先下单后付模式：提交堂食订单", zap.Any("params", params))

	// 提交堂食订单
	err := h.orderSrv.SubmitMemberDineInOrder(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

// GetMemberOrderFormInfo 获取订单提交表单信息
// @Summary 获取订单提交表单信息
// @Description 获取订单提交表单信息
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.GetMemberOrderFormInfoReq true "详情参数"
// @Success 200 {object} resp.CreateMemberOrderResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/form/info [get]
func (h *OrderHandler) GetMemberOrderFormInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetMemberOrderFormInfoReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("获取订单提交表单信息", zap.Any("params", params))

	// 获取订单提交表单信息
	res, err := h.orderSrv.GetMemberOrderFormInfo(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDineInOrderFormInfo 获取堂食订单提交表单信息
// @Summary 获取堂食订单提交表单信息
// @Description 获取堂食订单提交表单信息，包含商品列表、金额信息、支付方式等
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.GetDineInOrderFormInfoReq true "详情参数"
// @Success 200 {object} resp.DineInOrderFormResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/form_info [get]
func (h *OrderHandler) GetDineInOrderFormInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetDineInOrderFormInfoReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("获取堂食订单提交表单信息", zap.Any("params", params))

	// 获取堂食订单提交表单信息
	res, err := h.orderSrv.GetDineInOrderFormInfo(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// SetDineInOrderDiningMethod 设置堂食订单用餐方式
// @Summary 设置堂食订单用餐方式
// @Description 设置堂食订单用餐方式（堂食/打包）
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.SetDineInOrderDiningMethodReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/dining_method [post]
func (h *OrderHandler) SetDineInOrderDiningMethod(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.SetDineInOrderDiningMethodReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("设置堂食订单用餐方式", zap.Any("params", params))

	// 设置用餐方式
	tracker := helper.StartTrack(ctx, constant.ActionMemberDineInSetDiningMethod).
		WithBill(params.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	err := h.orderSrv.SetDineInOrderDiningMethod(ctx, params)
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// PayDineInOrder 堂食订单提交支付
// @Summary 堂食订单提交支付
// @Description 堂食订单提交支付
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PayDineInOrderReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/pay [post]
func (h *OrderHandler) PayDineInOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.PayDineInOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("堂食订单提交支付", zap.Any("params", params))

	// 提交支付
	tracker := helper.StartTrack(ctx, constant.ActionMemberDineInPay).
		WithBill(params.SaleBillUuid, params.SaleOrderUuid).
		WithPath(c.Request.URL.Path)
	err := h.orderSrv.PayDineInOrder(ctx, params)
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// GetDineInOrderPayInfo 获取堂食订单支付信息
// @Summary 获取堂食订单支付信息
// @Description 获取堂食订单支付信息，返回支付二维码或跳转链接
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.GetDineInOrderPayInfoReq true "详情参数"
// @Success 200 {object} resp.DineInOrderPaymentInfoResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/pay/info [get]
func (h *OrderHandler) GetDineInOrderPayInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetDineInOrderPayInfoReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取支付信息
	res, err := h.orderSrv.GetDineInOrderPayInfo(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDineInOrderPayStatus 获取堂食订单支付状态
// @Summary 获取堂食订单支付状态
// @Description 获取堂食订单支付状态
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.GetDineInOrderPayStatusReq true "详情参数"
// @Success 200 {object} resp.DineInOrderPaymentStatusResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/pay/status [get]
func (h *OrderHandler) GetDineInOrderPayStatus(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetDineInOrderPayStatusReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取支付状态
	res, err := h.orderSrv.GetDineInOrderPayStatus(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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

	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionMemberOrderAddress).
		WithPath(c.Request.URL.Path)
	// 设置订单地址
	res, err := h.orderSrv.SetMemberOrderAddress(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
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

	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionMemberOrderPay).
		WithPath(c.Request.URL.Path)
	// 提交支付
	err := h.orderSrv.PayMemberOrder(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
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
// @Param data query member_req.GetMemberOrderPayInfoReq true "详情参数"
// @Success 200 {object} resp.MemberOrderPaymentInfoResp "成功"
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
// @Success 200 {object} resp.MemberOrderPaymentStatusResp "成功"
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

// GetMemberOrderList 获取会员端订单列表
// @Summary 获取会员端订单列表
// @Description 获取会员端订单列表
// @Tags 会员端-订单列表
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.MemberOrderListReq true "详情参数"
// @Success 200 {object} resp.GetMemberOrderListResp "成功"
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

// GetMemberOrderDetail 获取会员端订单详情
// @Summary 获取会员端订单详情
// @Description 获取会员端订单详情
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query string true "会员端销售订单UUID"
// @Success 200 {object} resp.GetMemberOrderDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/detail [get]
func (h *OrderHandler) GetMemberOrderDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetMemberOrderDetailReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取会员端订单详情
	res, err := h.orderSrv.GetMemberOrderDetail(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetMemberOrderPaymentMethodList 获取会员端订单支付方式列表
// @Summary 获取会员端订单支付方式列表
// @Description 获取会员端订单支付方式列表
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query string true "会员端销售订单UUID"
// @Success 200 {object} resp.GetMemberOrderPaymentMethodListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/payment/method/list [get]
func (h *OrderHandler) GetMemberOrderPaymentMethodList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetMemberOrderDetailReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取会员端订单支付方式列表
	res, err := h.orderSrv.GetMemberOrderPaymentMethodList(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 取消订单
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.CancelOrderReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.CancelOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionMemberOrderCancel).
		WithPath(c.Request.URL.Path)
	// 取消订单
	err := h.orderSrv.MemberOrderCancel(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetRiderInfo 获取骑手信息
// @Summary 获取骑手信息
// @Description 获取骑手信息
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query string true "会员端销售订单UUID"
// @Success 200 {object} resp.MemberOrderCoordinates "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/rider [get]
func (h *OrderHandler) GetRiderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.GetRiderInfoReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("获取骑手信息", zap.Any("params", params))

	// 获取骑手信息
	res, err := h.orderSrv.GetRiderInfo(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// SendAuthCode 发送认证验证码
// @Summary 发送认证验证码
// @Description 发送认证验证码
// @Tags 会员端-订单
// @Accept json
// @Produce json
// @param data body req.MemberOrderSendAuthCodeReq true "详情参数"
// @Success 200 {object} dto.Response{}
// @Router /member/order/send_auth_code [post]
func (h *OrderHandler) SendAuthCode(c *gin.Context) {
	ctx := helper.GetContext(c)
	sendCodeReq := req.MemberOrderSendAuthCodeReq{}
	if err := c.ShouldBindJSON(&sendCodeReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.orderSrv.SendAuthCode(ctx, sendCodeReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{
		"uuid": sendCodeReq.MemberSaleOrderUuid,
	})
}

// CheckDineInOrder 会员端堂食订单结算前检查
// @Summary 会员端堂食订单结算前检查
// @Description 在会员端点击"去结算"时调用，校验消费税变动、服务费变动、库存、价格变动、规格、商品变动（被删除、下架）、限购等
// @Tags 会员端-堂食订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.CheckDineInOrderReq true "详情参数"
// @Success 200 {object} nil "检查通过"
// @Failure 400 {object} resp.OrderCheckRes "检查不通过，返回具体变动信息"
// @Router /member/order/dine_in/check [get]
func (h *OrderHandler) CheckDineInOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.CheckDineInOrderReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("会员端堂食订单结算前检查", zap.Any("params", params))

	// 执行结算前检查
	checkRes, err := h.orderSrv.CheckDineInOrder(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("堂食订单结算前检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	// 检查通过
	helper.Success(c, gin.H{})
}

// GetMemberDineInOrderList 获取会员端堂食订单列表
// @Summary 获取会员端堂食订单列表
// @Description 获取会员端堂食订单列表，支持状态过滤（全部、待支付、进行中、已完成、已取消）
// @Tags 会员端-堂食订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.MemberDineInOrderListReq true "详情参数"
// @Success 200 {object} resp.GetMemberDineInOrderListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/list [get]
func (h *OrderHandler) GetMemberDineInOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.MemberDineInOrderListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取会员端堂食订单列表
	res, err := h.orderSrv.GetMemberDineInOrderList(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetMemberDineInOrderDetail 获取会员端堂食订单详情
// @Summary 获取会员端堂食订单详情
// @Description 获取会员端堂食订单详情，包含商品列表、金额信息、状态信息等
// @Tags 会员端-堂食订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.GetMemberDineInOrderDetailReq true "详情参数"
// @Success 200 {object} resp.GetMemberDineInOrderDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/detail [get]
func (h *OrderHandler) GetMemberDineInOrderDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetMemberDineInOrderDetailReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取会员端堂食订单详情
	res, err := h.orderSrv.GetMemberDineInOrderDetail(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// MockPayDineInOrderCallback 模拟堂食订单支付完成回调
// @Summary 模拟堂食订单支付完成回调（仅测试用）
// @Description 该接口仅在调试/测试模式下可用，用于模拟支付完成后的回调流程
// @Tags 会员端-堂食订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MockPayDineInOrderCallbackReq true "请求参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/pay/mock_callback [post]
func (h *OrderHandler) MockPayDineInOrderCallback(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.MockPayDineInOrderCallbackReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 执行模拟支付回调
	err := h.orderSrv.MockPayDineInOrderCallback(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

// CancelDineInOrder 取消会员端堂食订单
// @Summary 取消会员端堂食订单
// @Description 取消未付款的会员端堂食订单。仅支持未付款状态的订单取消。
// @Tags 会员端-堂食订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CancelMemberDineInOrderReq true "请求参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /member/order/dine_in/cancel [post]
func (h *OrderHandler) CancelDineInOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.CancelMemberDineInOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 取消堂食订单
	tracker := helper.StartTrack(ctx, constant.ActionMemberDineInCancel).
		WithBill(params.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	err := h.orderSrv.CancelMemberDineInOrder(ctx, params)
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
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
	orderSrv := service.NewOrderSrv(
		dbm,
		service.NewLocaleSrv(),
		settingSrv,
		service.NewMustPlanSrv(dbm),
		service.NewPaymentMethodSrv(dbm, settingSrv),
		service.NewMemberSrv(dbm, cache),
		service.NewCashBoxSrv(dbm),
		service.WithSmsSrv(dbm),
	)
	wrapper := &OrderHandler{
		orderSrv: orderSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.POST("/order/create", wrapper.CreateOrder)                                   // 创建外送订单
		privateApi.POST("/order/dine_in/create", wrapper.CreateDineInOrder)                     // 创建堂食订单
		privateApi.POST("/order/dine_in/submit", wrapper.SubmitDineInOrder)                     // 先下单后付模式：提交堂食订单
		privateApi.GET("/order/form/info", wrapper.GetMemberOrderFormInfo)                      // 获取外送订单提交表单信息
		privateApi.GET("/order/dine_in/check", wrapper.CheckDineInOrder)                        // 堂食订单结算前检查
		privateApi.GET("/order/dine_in/form_info", wrapper.GetDineInOrderFormInfo)              // 获取堂食订单提交表单信息
		privateApi.POST("/order/dine_in/dining_method", wrapper.SetDineInOrderDiningMethod)     // 设置堂食订单用餐方式
		privateApi.POST("/order/address", wrapper.SetOrderAddress)                              // 设置订单地址
		privateApi.POST("/order/pay", wrapper.PayOrder)                                         // 外送订单提交支付
		privateApi.GET("/order/pay/info", wrapper.PayOrderInfo)                                 // 外送订单获取支付信息
		privateApi.GET("/order/pay/status", wrapper.PayOrderStatus)                             // 外送订单获取支付状态
		privateApi.POST("/order/dine_in/pay", wrapper.PayDineInOrder)                           // 堂食订单提交支付
		privateApi.GET("/order/dine_in/pay/info", wrapper.GetDineInOrderPayInfo)                // 堂食订单获取支付信息
		privateApi.GET("/order/dine_in/pay/status", wrapper.GetDineInOrderPayStatus)            // 堂食订单获取支付状态
		privateApi.GET("/order/dine_in/list", wrapper.GetMemberDineInOrderList)                 // 获取堂食订单列表
		privateApi.GET("/order/dine_in/detail", wrapper.GetMemberDineInOrderDetail)             // 获取堂食订单详情
		privateApi.POST("/order/dine_in/cancel", wrapper.CancelDineInOrder)                     // 取消堂食订单
		privateApi.GET("/order/list", wrapper.GetMemberOrderList)                               // 获取会员端订单列表
		privateApi.GET("/order/detail", wrapper.GetMemberOrderDetail)                           // 获取会员端订单详情
		privateApi.GET("/order/payment/method/list", wrapper.GetMemberOrderPaymentMethodList)   // 获取会员端订单支付方式列表
		privateApi.POST("/order/cancel", wrapper.CancelOrder)                                   // 取消订单
		privateApi.POST("/order/send_auth_code", wrapper.SendAuthCode)                          // 发送认证验证码
		privateApi.GET("/order/rider", wrapper.GetRiderInfo)                                    // 获取骑手信息，以及商家、会员坐标
		privateApi.POST("/order/dine_in/pay/mock_callback", wrapper.MockPayDineInOrderCallback) // 模拟堂食订单支付回调（仅测试用）
	}
}
