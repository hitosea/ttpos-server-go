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

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// InstantHandler 收银点餐处理程序
type InstantHandler struct {
	orderService service.IOrderSrv // 订单服务
}

func (h *InstantHandler) CreateInstantOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 创建订单
	res, err := h.orderService.CreateInstantOrder(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("创建订单失败"))
		return
	}
	helper.Success(c, res)
}

// CancelOrder 处理取消点餐订单
// @Summary 取消点餐订单
// @Description 取消点餐订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消点餐订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cancel [post]
func (h *InstantHandler) CancelOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderCancelReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.orderService.CancelOrder(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// HideOrder 处理隐藏点餐订单（挂单）
// @Summary 隐藏（挂单）
// @Description 隐藏（挂单）
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @Success 200 {object} nil "隐藏点餐订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/hide [post]
func (h *InstantHandler) HideOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderDeleteReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	shopCart, err := h.orderService.HideOrder(ctx, req.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// ShowOrder 处理显示点餐订单（取单）
// @Summary 显示点餐订单（取单）
// @Description 显示点餐订单（取单）
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderShowReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/show [post]
func (h *InstantHandler) ShowOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderShowReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	shopCart, err := h.orderService.ShowOrder(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// OrderList 处理显示点餐订单列表（取单列表）
// @Summary 显示点餐订单列表（取单列表）
// @Description 显示点餐订单列表（取单列表）
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.HideSaleBillListReq true "列表参数"
// @Success 200 {object} dto.Response{data=resp.InstantHideOrderListResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/list [get]
func (h *InstantHandler) OrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var listReq req.HideSaleBillListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	resp, err := h.orderService.InstantHideOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, resp)
}

// OrderProductDelete 处理删除点餐订单商品
// @Summary 删除点餐订单商品
// @Description 删除点餐订单商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductDeleteReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/delete [delete]
func (h *InstantHandler) OrderProductDelete(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	staff := helper.GetStaff(c)
	source := helper.GetSource(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductDeleteReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	shopCart, err := h.orderService.OrderProductDelete(ctx, companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// OrderProductChangePrice 处理删除点餐订单商品改价
// @Summary 点餐订单商品改价
// @Description 点餐订单商品改价
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductChangePriceReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/price [post]
func (h *InstantHandler) OrderProductChangePrice(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductChangePriceReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderService.OrderProductChangePrice(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderChangePopulation 处理点餐订单修改人数
// @Summary 点餐订单修改人数
// @Description 点餐订单修改人数
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangePopulationReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/population [post]
func (h *InstantHandler) OrderChangePopulation(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangePopulationReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderService.OrderChangePopulation(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderProductRemark 处理点餐订单商品备注
// @Summary 点餐订单商品备注
// @Description 点餐订单商品备注
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/remark [post]
func (h *InstantHandler) OrderProductRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderService.OrderProductRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderCartInfo 查询点餐购物车信息
// @Summary 查询点餐购物车信息
// @Description 查询点餐购物车信息
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/info [get]
func (h *InstantHandler) OrderCartInfo(c *gin.Context) {
	// 点餐桌台查询时没有销售账单ID，只能通过判断请求来自哪个收银机判断是哪个销售账单。
	// 一个收银机只一个未挂单的点餐销售账单
	ctx := helper.GetContext(c)
	deviceSn := ctx.GetDeviceSn()
	ctx.Log().Debug("查询点餐销售账单", zap.Any("deviceSn", deviceSn))
	if deviceSn == "" {
		helper.ResponseFail(c, constant.CodeFail, errors.ErrNoDeviceSn)
		return
	}
	// 通过收银机sn获取收银机设备ID，通过设备ID查询属于该收银机的未挂单点餐账单。有0个或1个账单
	res, err := h.orderService.GetOrderCartInfoByDeviceSn(ctx, deviceSn)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if res == nil {
		// 没有查询到属于该收银机的未挂单销售账单
		helper.Success(c, resp.ShopCart{SaleOrderList: make([]resp.SaleOrder, 0)})
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductAdd 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/add [post]
func (h *InstantHandler) OrderCartProductAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 添加商品。 若没有点餐账单则新建一个
	res, err := h.orderService.InstantOrderCartProductAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductNum 修改购物车商品数量
// @Summary 修改购物车某个商品的数量
// @Description 修改购物车商品数量
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProductNumReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/num [post]
func (h *InstantHandler) OrderCartProductNum(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面修改购物车商品数量接口请求")
	// 绑定请求参数
	params := req.OrderCartProductNumReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("点餐页面修改购物车商品数量接口请求", zap.Any("params", params))
	// 修改购物车商品数量
	res, err := h.orderService.InstantOrderCartProductNum(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("修改商品数量成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCooking 送厨购物车商品
// @Summary 送厨购物车商品
// @Description 送厨购物车商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProductCookingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/cooking [post]
func (h *InstantHandler) OrderCartProductCooking(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面送厨购物车商品接口请求")
	// 绑定请求参数
	params := req.OrderCartProductCookingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("点餐页面送厨购物车商品接口请求", zap.Any("params", params))
	// 送厨购物车商品
	res, err := h.orderService.InstantOrderCartProductCooking(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("送厨购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductReturning 退菜购物车商品
// @Summary 退菜购物车商品
// @Description 退菜购物车商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProductReturningReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/returning [post]
func (h *InstantHandler) OrderCartProductReturning(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductReturningReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderService.InstantOrderCartProductReturning(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

func (h *InstantHandler) OrderMustPlan(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面必点方案接口请求")
	deviceSn := ctx.GetDeviceSn()
	ctx.Log().Debug("点餐页面必点方案接口", zap.Any("deviceSn", deviceSn))
	if deviceSn == "" {
		helper.ResponseFail(c, constant.CodeFail, errors.ErrNoDeviceSn)
		return
	}

	// 获取必点方案信息
	res, err := h.orderService.InstantOrderMustPlan(ctx, deviceSn)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("获取必点方案信息成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentInfo 获取结账页面信息
// @Summary 获取结账页面信息
// @Description 获取结账页面信息
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query string true "销售账单UUID"
// @param sale_order_uuid query string true "销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/payment/info [get]
func (h *InstantHandler) OrderPaymentInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面结账页面信息接口请求")

	params := &req.InstantOrderPaymentInfoReq{}
	if err := params.Parse(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Info("查询销售订单收银机结账页面信息", zap.Any("params", params))
	// 获取销售订单的付款信息
	res, err := h.orderService.InstantOrderPaymentInfo(ctx, params.SaleBillUuid, params.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("获取结账页面信息成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCreate 创建一个支付单
// @Summary 创建一个支付单
// @Description 创建一个支付单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCreateReq true "创建一个支付单参数"
// @Success 200 {object} dto.Response{data=resp.RechargeOrder}
// @Router /cashier/instant/order/payment/create [post]
func (h *InstantHandler) OrderPaymentCreate(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面结账页面信息接口请求")

	params := req.InstantOrderPaymentCreateReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("创建一个支付单", zap.Any("params", params))
	// 创建一个支付单
	res, err := h.orderService.InstantOrderPaymentCreate(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("创建一个支付单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderSaleOrderCreate 创建一个销售订单
// @Summary 创建一个销售订单
// @Description 创建一个销售订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderSaleOrderCreateReq true "创建一个销售订单参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/instant/order/sale_order/create [post]
func (h *InstantHandler) OrderSaleOrderCreate(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面创建一个销售订单接口请求")

	params := req.InstantOrderSaleOrderCreateReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("创建一个销售订单", zap.Any("params", params))
	// 创建一个销售订单
	res, err := h.orderService.InstantOrderSaleOrderCreate(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("创建一个销售订单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderSaleOrderMoveProduct 从一个销售订单移动商品到另一个销售订单
// @Summary 从一个销售订单移动商品到另一个销售订单
// @Description 从一个销售订单移动商品到另一个销售订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.InstantOrderSaleOrderMoveProductReq true "从一个销售订单移动商品到另一个销售订单参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/instant/order/sale_order/move_product [post]
func (h *InstantHandler) OrderSaleOrderMoveProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面从一个销售订单移动商品到另一个销售订单接口请求")

	params := req.InstantOrderSaleOrderMoveProductReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("从一个销售订单移动商品到另一个销售订单", zap.Any("params", params))
	// 从一个销售订单移动商品到另一个销售订单
	res, err := h.orderService.InstantOrderSaleOrderMoveProduct(ctx, params, false)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("从一个销售订单移动商品到另一个销售订单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderMustPlanConfirm 确认必点商品
// @Summary 确认必点商品
// @Description 确认必点商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.InstantOrderMustPlanConfirmReq true "确认必点商品参数"
// @Success 200 {object} dto.Response{}
// @Router /cashier/instant/order/must_plan/confirm [post]
func (h *InstantHandler) OrderMustPlanConfirm(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面确认必点商品接口请求")

	params := req.InstantOrderMustPlanConfirmReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("确认必点商品", zap.Any("params", params))
	// 确认必点商品
	res, err := h.orderService.InstantOrderMustPlanConfirm(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("确认必点商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderSaleOrderDelete 删除一个销售订单(删除拆单)
// @Summary 删除一个销售订单(删除拆单)
// @Description 删除一个销售订单(删除拆单)
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.InstantOrderSaleOrderDeleteReq true "删除一个销售订单(删除拆单)参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/instant/order/sale_order/delete [delete]
func (h *InstantHandler) OrderSaleOrderDelete(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面删除一个销售订单(删除拆单)接口请求")

	params := req.InstantOrderSaleOrderDeleteReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("删除一个销售订单(删除拆单)", zap.Any("params", params))
	// 删除一个销售订单(删除拆单)
	res, err := h.orderService.InstantOrderSaleOrderDelete(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("删除一个销售订单(删除拆单)成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderSaleOrderDeleteAll 删除所有子销售订单(撤销拆单)
// @Summary 删除所有子销售订单(撤销拆单)
// @Description 删除所有子销售订单(撤销拆单)
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.InstantOrderSaleOrderDeleteAllReq true "删除所有子销售订单(撤销拆单)参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/instant/order/sale_order/delete_all [delete]
func (h *InstantHandler) OrderSaleOrderDeleteAll(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面删除所有子销售订单(撤销拆单)接口请求")

	params := req.InstantOrderSaleOrderDeleteAllReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("删除所有子销售订单(撤销拆单)", zap.Any("params", params))
	// 删除所有子销售订单(撤销拆单)
	res, err := h.orderService.InstantOrderSaleOrderDeleteAll(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("删除所有子销售订单(撤销拆单)成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// RegisterInstantHandlers 注册收银订单路由
func RegisterInstantHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv)

	// 创建收银产品处理程序
	wrapper := InstantHandler{
		orderService: orderSrv, // 订单服务
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.POST("/instant/order/create", wrapper.CreateInstantOrder)                         // 创建点餐订单。 废弃，点餐点餐由系统自动创建
		privateApi.POST("/instant/order/cancel", wrapper.CancelOrder)                                // 取消点餐订单
		privateApi.POST("/instant/order/hide", wrapper.HideOrder)                                    // 隐藏点餐订单（挂单）
		privateApi.POST("/instant/order/show", wrapper.ShowOrder)                                    // 显示点餐订单（取单）
		privateApi.GET("/instant/order/list", wrapper.OrderList)                                     // 显示点餐订单列表（取单列表）
		privateApi.DELETE("/instant/order/product/delete", wrapper.OrderProductDelete)               // 删除点餐订单商品
		privateApi.POST("/instant/order/product/price", wrapper.OrderProductChangePrice)             // 点餐订单商品改价
		privateApi.POST("/instant/order/population", wrapper.OrderChangePopulation)                  // 点餐订单修改人数
		privateApi.POST("/instant/order/product/remark", wrapper.OrderProductRemark)                 // 点餐订单商品备注
		privateApi.GET("/instant/order/cart/info", wrapper.OrderCartInfo)                            // 查询点餐购物车信息
		privateApi.POST("/instant/order/cart/product/add", wrapper.OrderCartProductAdd)              // 向购物车添加商品
		privateApi.POST("/instant/order/cart/product/num", wrapper.OrderCartProductNum)              // 修改购物车商品数量
		privateApi.POST("/instant/order/cart/cooking", wrapper.OrderCartProductCooking)              // 送厨购物车商品
		privateApi.POST("/instant/order/cart/returning", wrapper.OrderCartProductReturning)          // 退菜购物车商品
		privateApi.GET("/instant/order/must_plan", wrapper.OrderMustPlan)                            // 获取点餐必点方案。废弃
		privateApi.POST("/instant/order/must_plan/confirm", wrapper.OrderMustPlanConfirm)            // 确认必点商品
		privateApi.GET("/instant/order/payment/info", wrapper.OrderPaymentInfo)                      // 获取结账页面信息
		privateApi.POST("/instant/order/payment/create", wrapper.OrderPaymentCreate)                 // 创建一个支付单
		privateApi.POST("/instant/order/sale_order/create", wrapper.OrderSaleOrderCreate)            // 创建一个销售订单
		privateApi.POST("/instant/order/sale_order/move_product", wrapper.OrderSaleOrderMoveProduct) // 从一个销售订单移动商品到另一个销售订单
		privateApi.DELETE("/instant/order/sale_order/delete", wrapper.OrderSaleOrderDelete)          // 删除一个销售订单(删除拆单)
		privateApi.DELETE("/instant/order/sale_order/delete_all", wrapper.OrderSaleOrderDeleteAll)   // 删除所有子销售订单(撤销拆单)
	}
}
