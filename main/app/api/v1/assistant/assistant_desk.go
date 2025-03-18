package assistant

import (
	"bytes"
	"io"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeskHandler 桌台处理程序
type DeskHandler struct {
	deskSrv   service.IDeskSrv
	orderSrv  service.IOrderSrv
	memberSrv service.IMemberSrv
}

// GetDeskRegionAndType 处理获取桌台的区域和类型
// @Summary 获取桌台的区域和类型
// @Description 获取桌台的区域和类型
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskRegionAndTypeListWithPaginationResp "桌台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/region_and_type [get]
func (h *DeskHandler) GetDeskRegionAndType(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 处理获取桌台的区域和类型的逻辑
	res, err := h.deskSrv.GetDeskRegionAndTypeList(companyId)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskList 处理获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskListReq true "列表参数"
// @Success 200 {array} resp.DeskListWithPaginationResp "桌台列表"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/list [get]
func (h *DeskHandler) GetDeskList(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var deskListReq req.DeskListReq
	if err := c.ShouldBindQuery(&deskListReq); err != nil {
		helper.HandleValidationError(c, err, deskListReq, dto.PageReqMessage)
		return
	}
	res, err := h.deskSrv.GetDeskList(ctx, companyId, deskListReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, apperrors.ErrInternal)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskInfo 处理获取桌台详情
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskInfoReq true "详情参数"
// @Success 200 {object} resp.DeskInfoResp "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/info [get]
func (h *DeskHandler) GetDeskInfo(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 绑定请求参数
	var deskInfoReq req.DeskInfoReq
	if err := c.ShouldBindQuery(&deskInfoReq); err != nil {
		helper.HandleValidationError(c, err, deskInfoReq, nil)
		return
	}
	res, err := h.deskSrv.GetDeskInfo(companyId, deskInfoReq.Uuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, apperrors.ErrInternal)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CreateDeskOrder 处理创建开台
// @Summary 开台
// @Description 开台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "开台参数"
// @Success 200 {object} resp.CreateDeskOrderResp "开台成功"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/open [post]
func (h *DeskHandler) CreateDeskOrder(c *gin.Context) {

	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	// 创建桌台订单
	res, err := h.deskSrv.CreateDeskOrder(ctx, params)
	// 处理错误
	if err != nil {
		ctx.Log().Error("创建桌台订单失败", zap.Error(err))
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// OrderProductRemark 处理桌台订单商品备注
// @Summary 桌台订单商品备注
// @Description 桌台订单商品备注
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/product/remark [post]
func (h *DeskHandler) OrderProductRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderProductRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderCheck 订单检查
// @Summary 订单检查
// @Description 订单检查。场景：1、点击结账按钮时，检查订单是否可以结账
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @Success 200 {object} dto.Response{data=resp.OrderCheckRes}
// @Router /assistant/desk/order/check [get]
func (h *DeskHandler) OrderCheck(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面订单检查接口请求")

	params := req.InstantOrderCheckReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("订单检查", zap.Any("params", params))
	// 订单检查
	checkRes, err := h.orderSrv.OrderCheck(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("送厨检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	ctx.Log().Debug("订单检查成功")
	// 返回结果
	helper.Success(c, resp.OrderCheckRes{})
}

// OrderPaymentInfo 获取结账页面信息
// @Summary 获取结账页面信息
// @Description 获取结账页面信息
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query string true "销售账单UUID"
// @param sale_order_uuid query string true "销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/payment/info [get]
func (h *DeskHandler) OrderPaymentInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面结账页面信息接口请求")

	params := &req.InstantOrderPaymentInfoReq{}
	if err := params.Parse(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Info("查询销售订单收银机结账页面信息", zap.Any("params", params))
	// 获取销售订单的付款信息
	res, err := h.orderSrv.InstantOrderPaymentInfo(ctx, params.SaleBillUuid, params.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("获取结账页面信息成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCreate 创建一个支付单
// @Summary 创建一个支付单
// @Description 创建一个支付单
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCreateReq true "创建一个支付单参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
// @Router /assistant/desk/order/payment/create [post]
func (h *DeskHandler) OrderPaymentCreate(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面创建一个支付单接口请求")

	params := req.InstantOrderPaymentCreateReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("创建一个支付单", zap.Any("params", params))
	// 创建一个支付单
	res, err := h.orderSrv.InstantOrderPaymentCreate(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("创建一个支付单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCancel 撤销一个支付单
// @Summary 撤销一个支付单
// @Description 撤销一个支付单
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCancelReq true "撤销一个支付单参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
// @Router /assistant/desk/order/payment/cancel [post]
func (h *DeskHandler) OrderPaymentCancel(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面撤销一个支付单接口请求")

	params := req.InstantOrderPaymentCancelReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("撤销一个支付单", zap.Any("params", params))
	// 撤销一个支付单
	res, err := h.orderSrv.InstantOrderPaymentCancel(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("撤销一个支付单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentFinish 完成销售订单的付款结账
// @Summary 完成销售订单的付款结账
// @Description 完成销售订单的付款结账
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentFinishReq true "完成销售订单的付款结账参数"
// @Success 200 {object} dto.Response{data=resp.OrderFinishResp}
// @Router /assistant/desk/order/payment/finish [post]
func (h *DeskHandler) OrderPaymentFinish(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面销售订单的付款结账接口请求")

	params := req.InstantOrderPaymentFinishReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("桌台销售订单的付款结账", zap.Any("params", params))
	// 桌台销售订单的付款结账
	res, err := h.orderSrv.InstantOrderPaymentFinish(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("桌台销售订单的付款结账成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductAdd 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/add [post]
func (h *DeskHandler) OrderCartProductAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 添加商品。 若没有点餐账单则新建一个
	res, err := h.orderSrv.InstantOrderCartProductAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductNum 修改购物车商品数量
// @Summary 修改购物车某个商品的数量
// @Description 修改购物车商品数量
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductNumReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/num [post]
func (h *DeskHandler) OrderCartProductNum(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面修改购物车商品数量接口请求")
	// 绑定请求参数
	params := req.OrderCartProductNumReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("桌台页面修改购物车商品数量接口请求", zap.Any("params", params))
	// 修改购物车商品数量
	res, err := h.orderSrv.OrderCartProductNum(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("修改商品数量成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCooking 送厨购物车商品
// @Summary 送厨购物车商品
// @Description 送厨购物车商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductCookingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/cooking [post]
func (h *DeskHandler) OrderCartProductCooking(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductCookingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("桌台页面送厨购物车商品接口请求", zap.Any("params", params))
	// 送厨购物车商品
	res, checkRes, err := h.orderSrv.InstantOrderCartProductCooking(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("送厨检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	ctx.Log().Debug("送厨购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// ChangeDesk 处理切换桌台
// @Summary 切换桌台
// @Description 切换桌台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ChangeDeskReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/change [post]
func (h *DeskHandler) ChangeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.ChangeDeskReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	info, err := h.deskSrv.ChangeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// CompleteDesk 处理清台
// @Summary 清台
// @Description 清台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskInfoReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/complete [post]
func (h *DeskHandler) CompleteDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskJsonUuidReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	err := h.deskSrv.CompleteDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// MergeDesk 处理合并桌台
// @Summary 合并桌台
// @Description 合并桌台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.MergeDeskReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/merge [post]
func (h *DeskHandler) MergeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.MergeDeskReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	info, deskMergeCheckResp, err := h.deskSrv.MergeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, deskMergeCheckResp, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderDiscount 处理桌台订单打折
// @Summary 桌台订单打折
// @Description 桌台订单打折
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDiscountMethodReq true "打折参数，根据discount_method值(1:改价,2:打折,3:抹零)提供对应的额外字段"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/discount [post]
func (h *DeskHandler) OrderDiscount(c *gin.Context) {
	ctx := helper.GetContext(c)
	bodyBytes, _ := io.ReadAll(c.Request.Body) // Body只能读取一次，之后想再次从body中读取数据需要重新往body中写入数据
	// 绑定请求参数
	params := req.OrderDiscountMethodReq{}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新写入数据
	// 从body中读取数据
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新写入数据

	var shopCart *resp.ShopCart
	var err error
	// 改价
	if params.DiscountMethod == 1 {
		amountChangeReq := req.OrderAmountChangeReq{}
		if err := c.ShouldBindJSON(&amountChangeReq); err != nil {
			helper.HandleValidationError(c, err, amountChangeReq, req.OrderReqMessage)
			return
		}
		shopCart, err = h.orderSrv.OrderAmountChange(ctx, amountChangeReq)
	}
	// 打折
	if params.DiscountMethod == 2 {
		discountReq := req.OrderDiscountReq{}
		if err := c.ShouldBindJSON(&discountReq); err != nil {
			helper.HandleValidationError(c, err, discountReq, req.OrderReqMessage)
			return
		}
		shopCart, err = h.orderSrv.OrderDiscount(ctx, discountReq)
	}
	// 抹零
	if params.DiscountMethod == 3 {
		zeroRuleReq := req.OrderZeroRuleReq{}
		if err := c.ShouldBindJSON(&zeroRuleReq); err != nil {
			helper.HandleValidationError(c, err, zeroRuleReq, req.OrderReqMessage)
			return
		}
		shopCart, err = h.orderSrv.OrderZeroRule(ctx, zeroRuleReq)
	}
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// OrderDiscountCancel 处理桌台订单取消打折
// @Summary 取消桌台订单所有优惠折扣，包括改价、打折、抹零
// @Description 取消桌台订单所有优惠折扣，包括改价、打折、抹零
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDiscountCancelReq true "取消优惠折扣参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/discount/cancel [post]
func (h *DeskHandler) OrderDiscountCancel(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderDiscountCancelReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderDiscountCancel(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info, "操作成功")
}

// OrderProductChangePrice 处理桌台订单商品改价
// @Summary 桌台订单商品改价
// @Description 桌台订单商品改价
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductChangePriceReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/product/price [post]
func (h *DeskHandler) OrderProductChangePrice(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductChangePriceReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderProductChangePrice(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderCartProductReturning 退菜购物车商品
// @Summary 退菜购物车商品
// @Description 退菜购物车商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/returning [post]
func (h *DeskHandler) OrderCartProductReturning(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductReturningReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductReturning(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCancelReturning 取消退菜购物车商品
// @Summary 取消退菜购物车商品
// @Description 取消退菜购物车商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/cancel_returning [post]
func (h *DeskHandler) OrderCartProductCancelReturning(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProduct{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductCancelReturning(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductChangeDesk 转菜购物车商品
// @Summary 转菜购物车商品
// @Description 转菜购物车商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductChangeDeskReq true "转菜购物车商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/change_desk [post]
func (h *DeskHandler) OrderCartProductChangeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductChangeDeskReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 转菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductChangeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("转菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductGiving 赠菜购物车商品
// @Summary 赠菜购物车商品
// @Description 赠菜购物车商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductGivingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/giving [post]
func (h *DeskHandler) OrderCartProductGiving(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductGivingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductGiving(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCancelGiving 取消赠菜购物车商品
// @Summary 取消赠菜购物车商品
// @Description 取消赠菜购物车商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/cancel_giving [post]
func (h *DeskHandler) OrderCartProductCancelGiving(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProduct{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductCancelGiving(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// GetMemberDiscount 获取订单会员优惠
// @Summary 获取订单会员优惠
// @Description 获取订单会员优惠
// @Tags 点餐助手端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @param member_uuid query integer true "会员Uuid"
// @Success 200 {object} dto.Response{data=resp.MemberDiscountResp}
// @Router /assistant/desk/order/member/discount [get]
func (h *DeskHandler) GetMemberDiscount(c *gin.Context) {
	var discountReq req.GetMemberDiscountReq
	if err := c.ShouldBindQuery(&discountReq); err != nil {
		helper.HandleValidationError(c, err, discountReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	ctx.Log().Info("获取会员优惠", zap.Any("params", discountReq))
	order, err := h.memberSrv.GetMemberDiscount(ctx, discountReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, order)
}

// OrderUseMember 确认使用会员优惠并验证密码
// @Summary 确认使用会员优惠并验证密码
// @Description 确认使用会员优惠并验证密码
// @Tags 点餐助手端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CheckMemberPasswordReq true "确认使用会员优惠并验证密码"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Router /assistant/desk/order/member/confirm [post]
func (h *DeskHandler) OrderUseMember(c *gin.Context) {
	var passwordReq req.CheckMemberPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, req.CheckMemberPasswordMessage)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderUseMember(ctx, passwordReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// OrderMemberCancel 不使用此会员
// @Summary 不使用此会员
// @Description 不使用此会员
// @Tags 点餐助手端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderMemberCancelReq true "不使用此会员"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Router /assistant/desk/order/member/cancel [delete]
func (h *DeskHandler) OrderMemberCancel(c *gin.Context) {
	var passwordReq req.OrderMemberCancelReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, req.CheckMemberPasswordMessage)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderMemberCancel(ctx, passwordReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// OrderPrint 打印小票
// @Summary 桌台订单打印小票
// @Description 桌台订单打印小票
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /assistant/desk/order/print [post]
func (h *DeskHandler) OrderPrint(c *gin.Context) {
	var printReq req.OrderPrintReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderPrint(ctx, printReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// OrderPrintInvoice 打印发票
// @Summary 桌台订单打印发票
// @Description 桌台订单打印发票
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintInvoiceReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /assistant/desk/order/print/invoice [post]
func (h *DeskHandler) OrderPrintInvoice(c *gin.Context) {
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
	helper.Success(c, res)
}

// OrderProductDelete 处理删除桌台订单商品
// @Summary 删除桌台订单商品
// @Description 删除桌台订单商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductDeleteReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/product/delete [delete]
func (h *DeskHandler) OrderProductDelete(c *gin.Context) {
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
	shopCart, err := h.orderSrv.OrderProductDelete(ctx, companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// OrderChangeBuffet 处理桌台订单调整自助餐
// @Summary 桌台订单调整自助餐
// @Description 桌台订单调整自助餐
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangeBuffetReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/buffet [post]
func (h *DeskHandler) OrderChangeBuffet(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangeBuffetReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderChangeBuffet(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// GetUnSendKitchen 获取未送厨商品
// @Summary 获取未送厨商品
// @Description 获取未送厨商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.GetProductListReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.UnSendKitchen}
// @Router /assistant/desk/order/get_un_send_kitchen [get]
func (h *DeskHandler) GetUnSendKitchen(c *gin.Context) {
	ctx := helper.GetContext(c)
	params := req.GetProductListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	info, err := h.orderSrv.GetUnSendKitchen(ctx, params.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// GetSendKitchen 获取已送厨商品
// @Summary 获取已送厨商品
// @Description 获取已送厨商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.GetProductListReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.SendKitchen}
// @Router /assistant/desk/order/get_send_kitchen [get]
func (h *DeskHandler) GetSendKitchen(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetProductListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.GetSendKitchen(ctx, params.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderChangePopulation 处理桌台订单修改人数
// @Summary 桌台订单修改人数
// @Description 桌台订单修改人数
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangePopulationReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/population [post]
func (h *DeskHandler) OrderChangePopulation(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangePopulationReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderChangePopulation(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv)

	// 创建处理程序
	wrapper := DeskHandler{
		deskSrv:   service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv),
		orderSrv:  orderSrv,
		memberSrv: service.NewMemberSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)                                 // 获取桌台的区域和类型
		privateApi.GET("/desk/list", wrapper.GetDeskList)                                                     // 获取桌台列表
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)                                                     // 获取桌台详情
		privateApi.POST("/desk/open", wrapper.CreateDeskOrder)                                                // 创建桌台订单(开桌)
		privateApi.POST("/desk/order/cart/product/add", wrapper.OrderCartProductAdd)                          // 向购物车添加商品
		privateApi.DELETE("/desk/order/product/delete", wrapper.OrderProductDelete)                           // 删除桌台订单商品
		privateApi.POST("/desk/order/cart/product/num", wrapper.OrderCartProductNum)                          // 修改购物车商品数量
		privateApi.POST("/desk/order/cart/cooking", wrapper.OrderCartProductCooking)                          // 送厨购物车商品
		privateApi.GET("/desk/order/member/discount", wrapper.GetMemberDiscount)                              // 获取订单会员优惠
		privateApi.POST("/desk/order/member/confirm", wrapper.OrderUseMember)                                 // 确认使用会员优惠并验证密码
		privateApi.DELETE("/desk/order/member/cancel", wrapper.OrderMemberCancel)                             // 不使用此会员
		privateApi.POST("/desk/order/print", wrapper.OrderPrint)                                              // 打印小票
		privateApi.POST("/desk/order/print/invoice", wrapper.OrderPrintInvoice)                               // 打印发票
		privateApi.POST("/desk/change", wrapper.ChangeDesk)                                                   // 切换桌台（转台）
		privateApi.POST("/desk/complete", wrapper.CompleteDesk)                                               // 完成桌台（清台）
		privateApi.POST("/desk/merge", wrapper.MergeDesk)                                                     // 合并桌台
		privateApi.POST("/desk/order/discount", wrapper.OrderDiscount)                                        // 桌台订单打折
		privateApi.POST("/desk/order/discount/cancel", wrapper.OrderDiscountCancel)                           // 取消桌台订单所有优惠折扣，包括改价、打折、抹零
		privateApi.POST("/desk/order/product/price", wrapper.OrderProductChangePrice)                         // 桌台订单商品改价
		privateApi.POST("/desk/order/cart/product/returning", wrapper.OrderCartProductReturning)              // 退菜购物车商品
		privateApi.POST("/desk/order/cart/product/cancel_returning", wrapper.OrderCartProductCancelReturning) // 取消退菜购物车商品
		privateApi.POST("/desk/order/product/remark", wrapper.OrderProductRemark)                             // 桌台订单商品备注
		privateApi.GET("/desk/order/check", wrapper.OrderCheck)                                               // 订单检查。场景：1、点击结账按钮时，检查订单是否可以结账
		privateApi.GET("/desk/order/payment/info", wrapper.OrderPaymentInfo)                                  // 获取结账页面信息
		privateApi.POST("/desk/order/payment/create", wrapper.OrderPaymentCreate)                             // 创建一个支付单
		privateApi.POST("/desk/order/payment/cancel", wrapper.OrderPaymentCancel)                             // 撤销一个支付单
		privateApi.POST("/desk/order/payment/finish", wrapper.OrderPaymentFinish)                             // 完成销售订单的付款结账
		privateApi.POST("/desk/order/cart/product/change_desk", wrapper.OrderCartProductChangeDesk)           // 转菜
		privateApi.POST("/desk/order/cart/product/giving", wrapper.OrderCartProductGiving)                    // 赠菜购物车商品
		privateApi.POST("/desk/order/cart/product/cancel_giving", wrapper.OrderCartProductCancelGiving)       // 取消赠菜购物车商品
		privateApi.POST("/desk/order/population", wrapper.OrderChangePopulation)                              // 桌台订单修改人数
		privateApi.POST("/desk/order/buffet", wrapper.OrderChangeBuffet)                                      // 桌台订单调整自助餐
		privateApi.GET("/desk/order/get_un_send_kitchen", wrapper.GetUnSendKitchen)                           // 获取未送厨商品
		privateApi.GET("/desk/order/get_send_kitchen", wrapper.GetSendKitchen)                                // 获取已送厨商品
	}
}
