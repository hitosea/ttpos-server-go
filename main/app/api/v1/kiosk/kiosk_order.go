package kiosk

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/controller"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"go.uber.org/zap"
)

// OrderHandler 订单相关控制器
type OrderHandler struct {
	orderSrv service.IOrderSrv
}

// InvalidateSaleBillSettingCache 失效销售单设置缓存（辅助函数）
// 参数：
//   - ctx: 上下文, 用于提取 companyUuid
func (h *OrderHandler) InvalidateSaleBillSettingCache(ctx context.Context) {
	if !adapter.IsObjectStorageCacheEnabled(ctx.GetCompanyUuid()) {
		return
	}
	if err := controller.GetSaleBillSettingController().Invalidate(ctx, persistence.GlobalObjectUuid); err != nil {
		logger.Error("失效销售单设置缓存失败", zap.Error(err))
	}
}

// GetCartInfo 查询购物车信息
// @Summary 查询购物车信息
// @Description 查询购物车信息
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query uint64 false "销售账单UUID"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/cart/info [get]
func (h *OrderHandler) GetCartInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartInfoReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("查询购物车信息", zap.Any("params", params))

	var res *resp.ShopCart
	var err error

	// 如果提供了 sale_bill_uuid，则通过 sale_bill_uuid 查询
	if params.SaleBillUuid > 0 {
		res, err = h.orderSrv.GetOrderCartInfo(ctx, params.SaleBillUuid)
	} else {
		// 否则通过设备SN查询（类似 POS 端即时点餐）
		deviceSn := ctx.GetDeviceSn()
		if deviceSn == "" {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.NewWithCode(constant.CodeParamError, "sale_bill_uuid 或 deviceSn 必须提供一个"))
			return
		}
		res, err = h.orderSrv.GetOrderCartInfoByDeviceSn(ctx, deviceSn)
		if res == nil {
			// 没有查询到属于该设备的未挂单销售账单
			helper.Success(c, resp.ShopCart{SaleOrderList: make([]resp.SaleOrder, 0)})
			return
		}
	}

	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// AddProduct 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/cart/product/add [post]
func (h *OrderHandler) AddProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("向购物车添加商品", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionCartProductAdd).
		WithBill(params.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	// 添加商品。若没有点餐账单则新建一个
	res, err := h.orderSrv.InstantOrderCartProductAdd(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// AddProductPackage 向购物车添加套餐
// @Summary 向购物车添加套餐
// @Description 向购物车添加套餐
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductPackageAddReq true "套餐参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/cart/product_package/add [post]
func (h *OrderHandler) AddProductPackage(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductPackageAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("向购物车添加套餐", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionCartPackageAdd).
		WithBill(params.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	// 向购物车添加套餐
	res, err := h.orderSrv.OrderCartProductPackageAdd(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// UpdateProductNum 修改购物车商品数量
// @Summary 修改购物车商品数量
// @Description 修改购物车商品数量
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductNumReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/cart/product/num [post]
func (h *OrderHandler) UpdateProductNum(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("修改购物车商品数量")
	// 绑定请求参数
	params := req.OrderCartProductNumReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("修改购物车商品数量", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionCartProductNum).
		WithBill(params.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	// 修改购物车商品数量
	res, err := h.orderSrv.OrderCartProductNum(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetProductPackageDetail 获取商品选购详情
// @Summary 获取商品选购详情
// @Description 获取商品选购详情
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.GetProductPackageDetailReq true "商品选购详情参数"
// @Success 200 {object} dto.Response{data=resp.ProductPackageDetailRes}
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/product/package/detail [get]
func (h *OrderHandler) GetProductPackageDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetProductPackageDetailReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("获取商品选购详情", zap.Any("params", params))
	productPackage, err := h.orderSrv.GetProductPackageDetail(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, productPackage)
}

// DeleteProduct 删除购物车商品
// @Summary 删除购物车商品
// @Description 删除购物车商品
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductDeleteReq true "删除商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/cart/product/delete [delete]
func (h *OrderHandler) DeleteProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductDeleteReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("删除购物车商品", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionOrderProductDelete).
		WithBill(params.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	// 删除商品
	res, err := h.orderSrv.OrderProductDelete(ctx, ctx.GetDbId(), ctx.GetStaffUuid(), ctx.GetSource(), params)
	// 记录耗时
	tracker.End(ctx, err)
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
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCancelReq true "取消订单参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var cancelReq req.OrderCancelReq
	if err := c.ShouldBindJSON(&cancelReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消订单", zap.Any("params", cancelReq))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionCancelOrder).
		WithBill(cancelReq.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	// 取消订单
	err := h.orderSrv.CancelOrder(ctx, cancelReq)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderCheck 订单检查
// @Summary 订单检查
// @Description 订单检查。场景：1、点击结账按钮时，检查订单是否可以结账
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @Success 200 {object} dto.Response{data=resp.OrderCheckRes}
// @Router /kiosk/order/check [get]
func (h *OrderHandler) OrderCheck(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到kiosk订单检查接口请求")

	params := req.InstantOrderCheckReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("kiosk订单检查", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionOrderCheck).
		WithBill(params.SaleBillUuid, params.SaleOrderUuid).
		WithPath(c.Request.URL.Path)
	// 订单检查
	checkRes, err := h.orderSrv.OrderCheck(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		// 如果开启缓存,要失效sale_bill_setting的缓存
		h.InvalidateSaleBillSettingCache(ctx)

		ctx.Log().Debug("kiosk检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	ctx.Log().Debug("kiosk订单检查成功")
	// 返回结果
	helper.Success(c, resp.OrderCheckRes{})
}

// OrderTakeout 处理打包
// @Summary 打包
// @Description 打包
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderTakeoutReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/takeout [post]
func (h *OrderHandler) OrderTakeout(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderTakeoutReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("kiosk打包", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionOrderTakeout).
		WithBill(params.SaleBillUuid, 0).
		WithPath(c.Request.URL.Path)
	// 打包
	res, err := h.orderSrv.OrderTakeout(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentInfo 获取结账页面信息
// @Summary 获取结账页面信息
// @Description 获取结账页面信息
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query string true "销售账单UUID"
// @param sale_order_uuid query string true "销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /kiosk/order/payment/info [get]
func (h *OrderHandler) OrderPaymentInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到kiosk结账页面信息接口请求")

	params := &req.InstantOrderPaymentInfoReq{}
	if err := params.Parse(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Info("查询kiosk销售订单结账页面信息", zap.Any("params", params))
	// 获取销售订单的付款信息
	res, err := h.orderSrv.InstantOrderPaymentInfo(ctx, nil, params.SaleBillUuid, params.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// PayOrderInfo 获取支付信息
// @Summary 获取支付信息
// @Description 获取支付信息（创建支付订单，返回二维码和支付信息，可轮询此接口查询支付状态）
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query uint64 true "销售账单UUID"
// @param sale_order_uuid query uint64 true "销售订单UUID"
// @param payment_method_uuid query uint64 true "支付方式UUID"
// @param payment_amount query float64 true "支付金额"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentQrcodeInfoResp} "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kiosk/order/pay [post]
func (h *OrderHandler) PayOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.InstantOrderPaymentQrcodeReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("获取kiosk支付信息", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionPaymentCreate).
		WithBill(params.SaleBillUuid, params.SaleOrderUuid).
		WithPath(c.Request.URL.Path)
	// 获取支付信息
	res, err := h.orderSrv.InstantOrderPaymentQrcode(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// PayOrderStatus 获取支付状态
// @Summary 获取支付状态
// @Description 获取支付状态（查询订单的支付状态）
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query uint64 true "销售账单UUID"
// @param sale_order_uuid query uint64 true "销售订单UUID"
// @param payment_method_uuid query uint64 true "支付方式UUID"
// @param payment_amount query float64 true "支付金额"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentQrcodeInfoResp} "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kiosk/order/pay/status [get]
func (h *OrderHandler) PayOrderStatus(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.InstantOrderPaymentQrcodeReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Debug("获取kiosk支付状态", zap.Any("params", params))
	// 获取支付状态（复用 InstantOrderPaymentQrcode 方法，如果支付订单已存在会返回状态）
	res, err := h.orderSrv.InstantOrderPaymentQrcode(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// PayOrderConfirm Kiosk 支付完成确认
// @Summary Kiosk 支付完成确认
// @Description 用于 KBank 等需要主动上报支付结果的场景。Kiosk 调用此接口确认支付完成后，系统会自动执行送厨、打印送厨单和完成订单
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KioskPaymentConfirmReq true "支付确认参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kiosk/order/pay/confirm [post]
func (h *OrderHandler) PayOrderConfirm(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.KioskPaymentConfirmReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("kiosk支付确认", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionPaymentFinish).
		WithBill(params.SaleBillUuid, params.SaleOrderUuid).
		WithPath(c.Request.URL.Path)
	// 调用 service 层处理支付确认
	err := h.orderSrv.KioskPaymentConfirm(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, nil)
}

// OrderPrint 打印小票
// @Summary 订单打印小票
// @Description 订单打印小票
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /kiosk/order/print [post]
func (h *OrderHandler) OrderPrint(c *gin.Context) {
	var printReq req.OrderPrintReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionOrderPrint).
		WithBill(printReq.SaleBillUuid, printReq.SaleOrderUuid).
		WithPath(c.Request.URL.Path)
	res, err := h.orderSrv.OrderPrint(ctx, printReq, true)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res, "发送成功")
}

// FinishOrder Kiosk 订单完成
// @Summary Kiosk 订单完成
// @Description 检查订单实付是否符合完成支付条件，执行送厨和完成订单状态。用于客户端主动触发订单完成流程
// @Tags 自助点餐机.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.KioskOrderFinishReq true "订单完成参数"
// @Success 200 {object} dto.Response{data=resp.KioskOrderFinishResp} "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /kiosk/order/finish [post]
func (h *OrderHandler) FinishOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.KioskOrderFinishReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("kiosk订单完成", zap.Any("params", params))
	// 开始耗时跟踪
	tracker := helper.StartTrack(ctx, constant.ActionPaymentFinish).
		WithBill(params.SaleBillUuid, params.SaleOrderUuid).
		WithPath(c.Request.URL.Path)
	// 调用 service 层处理订单完成
	res, err := h.orderSrv.KioskOrderFinish(ctx, params)
	// 记录耗时
	tracker.End(ctx, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果（包含订单流水号）
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
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	wrapper := &OrderHandler{
		orderSrv: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order/cart/info", wrapper.GetCartInfo)                          // 查询购物车信息
		privateApi.POST("/order/cart/product/add", wrapper.AddProduct)                   // 向购物车添加商品
		privateApi.POST("/order/cart/product_package/add", wrapper.AddProductPackage)    // 向购物车添加套餐
		privateApi.POST("/order/cart/product/num", wrapper.UpdateProductNum)             // 修改购物车商品数量
		privateApi.GET("/order/product/package/detail", wrapper.GetProductPackageDetail) // 获取商品选购详情
		privateApi.DELETE("/order/cart/product/delete", wrapper.DeleteProduct)           // 删除购物车商品
		privateApi.POST("/order/cancel", wrapper.CancelOrder)                            // 取消订单
		privateApi.GET("/order/check", wrapper.OrderCheck)                               // 订单检查
		privateApi.POST("/order/takeout", wrapper.OrderTakeout)                          // 打包
		privateApi.GET("/order/payment/info", wrapper.OrderPaymentInfo)                  // 获取结账页面信息
		privateApi.POST("/order/pay", wrapper.PayOrder)                                  // 发起支付
		privateApi.GET("/order/pay/status", wrapper.PayOrderStatus)                      // 获取支付状态
		privateApi.POST("/order/pay/confirm", wrapper.PayOrderConfirm)                   // 支付完成确认（KBank 等主动上报场景）
		privateApi.POST("/order/finish", wrapper.FinishOrder)                            // 订单完成（检查支付、送厨、结账）
		privateApi.POST("/order/print", wrapper.OrderPrint)                              // 打印小票
	}
}
