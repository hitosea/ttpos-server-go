package cashier

import (
	"strconv"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// DeskHandler 桌台处理程序
type DeskHandler struct {
	service      service.IDeskSrv // 主服务
	orderService service.IOrderSrv
}

// GetDeskRegionAndType 处理获取桌台的区域和类型
// @Summary 获取桌台的区域和类型
// @Description 获取桌台的区域和类型
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskRegionAndTypeListWithPaginationResp "桌台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/region_and_type [get]
func (h *DeskHandler) GetDeskRegionAndType(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 处理获取桌台的区域和类型的逻辑
	res, err := h.service.GetDeskRegionAndTypeList(companyId)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskList 处理获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskListReq true "列表参数"
// @Success 200 {object} resp.DeskListWithPaginationResp "收银台列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/list [get]
func (h *DeskHandler) GetDeskList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var listReq req.DeskListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	// 获取收银产品列表
	res, err := h.service.GetDeskList(ctx, companyUuid, listReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskInfo 处理获取收银台列表
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskInfoReq true "详情参数"
// @Success 200 {object} resp.DeskInfoResp "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/info [get]
func (h *DeskHandler) GetDeskInfo(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	var infoReq req.DeskInfoReq
	if err := c.ShouldBindQuery(&infoReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 获取收银产品列表
	res, err := h.service.GetDeskInfo(companyUuid, infoReq.Uuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CreateDeskOrder 处理创建桌台订单
// @Summary 创建桌台订单
// @Description 创建桌台订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "创建桌台订单参数"
// @Success 200 {object} resp.CreateDeskOrderResp "创建桌台订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/open [post]
func (h *DeskHandler) CreateDeskOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	// 创建桌台订单
	res, err := h.service.CreateDeskOrder(ctx, params)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CloseDesk 处理关闭桌台
// @Summary 关闭桌台
// @Description 关闭桌台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskCloseReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/close [post]
func (h *DeskHandler) CloseDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	err := h.service.CloseDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// CompleteDesk 处理清台
// @Summary 清台
// @Description 清台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskInfoReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/complete [post]
func (h *DeskHandler) CompleteDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskJsonUuidReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	err := h.service.CompleteDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// CancelDeskOrder 处理取消桌台订单
// @Summary 取消桌台订单
// @Description 取消桌台订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cancel [post]
func (h *DeskHandler) CancelDeskOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var cancelReq req.OrderCancelReq
	if err := c.ShouldBindJSON(&cancelReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.orderService.CancelOrder(ctx, cancelReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductDelete 处理删除桌台订单商品
// @Summary 删除桌台订单商品
// @Description 删除桌台订单商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductDeleteReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/delete [delete]
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
	shopCart, err := h.orderService.OrderProductDelete(ctx, companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// OrderProductChangePrice 处理桌台订单商品改价
// @Summary 桌台订单商品改价
// @Description 桌台订单商品改价
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductChangePriceReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/price [post]
func (h *DeskHandler) OrderProductChangePrice(c *gin.Context) {
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

// OrderAmountChange 处理桌台订单改价
// @Summary 桌台订单改价
// @Description 桌台订单改价
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderAmountChangeReq true "改价参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/amount/change [post]
func (h *DeskHandler) OrderAmountChange(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderAmountChangeReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderService.OrderAmountChange(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderDiscount 处理桌台订单打折
// @Summary 桌台订单打折
// @Description 桌台订单打折
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDiscountReq true "打折参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/discount [post]
func (h *DeskHandler) OrderDiscount(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderDiscountReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderService.OrderDiscount(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderChangePopulation 处理桌台订单修改人数
// @Summary 桌台订单修改人数
// @Description 桌台订单修改人数
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangePopulationReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/population [post]
func (h *DeskHandler) OrderChangePopulation(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangePopulationReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderChangePopulation(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductRemark 处理桌台订单商品备注
// @Summary 桌台订单商品备注
// @Description 桌台订单商品备注
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/remark [post]
func (h *DeskHandler) OrderProductRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderProductRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderCartInfo 处理查询点餐购物车信息
// @Summary 查询点餐购物车信息
// @Description 查询点餐购物车信息
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Param sale_bill_uuid path string true "账单ID"
// @Param sale_order_uuid path string true "销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/info [get]
func (h *DeskHandler) OrderCartInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	saleBillUuid, err := strconv.ParseUint(c.Query("sale_bill_uuid"), 10, 64)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	res, err := h.orderService.GetOrderCartInfo(ctx, saleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductAdd 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/add [post]
func (h *DeskHandler) OrderCartProductAdd(c *gin.Context) {
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
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.OrderCartProductNumReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/num [post]
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
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.OrderCartProductCookingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/cooking [post]
func (h *DeskHandler) OrderCartProductCooking(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面送厨购物车商品接口请求")
	// 绑定请求参数
	params := req.OrderCartProductCookingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("桌台页面送厨购物车商品接口请求", zap.Any("params", params))
	// 送厨购物车商品
	res, errRes, err := h.orderService.InstantOrderCartProductCooking(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if errRes != nil {
		ctx.Log().Debug("送厨检查不通过", zap.Any("res", errRes))
		helper.FailWithData(c, errRes.Code, errRes.OrderCheckRes)
		return
	}
	ctx.Log().Debug("送厨购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductReturning 退菜购物车商品
// @Summary 退菜购物车商品
// @Description 退菜购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/returning [post]
func (h *DeskHandler) OrderCartProductReturning(c *gin.Context) {
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

// OrderCartProductCancelReturning 取消退菜购物车商品
// @Summary 取消退菜购物车商品
// @Description 取消退菜购物车商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/cancel_returning [post]
func (h *DeskHandler) OrderCartProductCancelReturning(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProduct{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderService.InstantOrderCartProductCancelReturning(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductGiving 赠菜购物车商品
// @Summary 取赠菜购物车商品
// @Description 赠菜购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.OrderCartProductGivingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/giving [post]
func (h *DeskHandler) OrderCartProductGiving(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductGivingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderService.InstantOrderCartProductGiving(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCancelGiving 取消赠菜购物车商品
// @Summary 取消赠菜购物车商品
// @Description 取消赠菜购物车商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/cancel_giving [post]
func (h *DeskHandler) OrderCartProductCancelGiving(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProduct{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderService.InstantOrderCartProductCancelGiving(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderMustPlanConfirm 确认必点商品
// @Summary 确认必点商品
// @Description 确认必点商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.InstantOrderMustPlanConfirmReq true "确认必点商品参数"
// @Success 200 {object} dto.Response{}
// @Router /cashier/desk/order/must_plan/confirm [post]
func (h *DeskHandler) OrderMustPlanConfirm(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面确认必点商品接口请求")

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

// OrderPaymentInfo 获取结账页面信息
// @Summary 获取结账页面信息
// @Description 获取结账页面信息
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query string true "销售账单UUID"
// @param sale_order_uuid query string true "销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/payment/info [get]
func (h *DeskHandler) OrderPaymentInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面结账页面信息接口请求")

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

// OrderSaleOrderCreate 创建一个销售订单
// @Summary 创建一个销售订单
// @Description 创建一个销售订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderSaleOrderCreateReq true "创建一个销售订单参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/create [post]
func (h *DeskHandler) OrderSaleOrderCreate(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面创建一个销售订单接口请求")

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
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.InstantOrderSaleOrderMoveProductReq true "从一个销售订单移动商品到另一个销售订单参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/move_product [post]
func (h *DeskHandler) OrderSaleOrderMoveProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面从一个销售订单移动商品到另一个销售订单接口请求")

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
	helper.Success(c, res, "拆单成功")
}

// OrderSaleOrderDelete 删除一个销售订单(删除拆单)
// @Summary 删除一个销售订单(删除拆单)
// @Description 删除一个销售订单(删除拆单)
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.InstantOrderSaleOrderDeleteReq true "删除一个销售订单(删除拆单)参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/delete [delete]
func (h *DeskHandler) OrderSaleOrderDelete(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面删除一个销售订单(删除拆单)接口请求")

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
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.InstantOrderSaleOrderDeleteAllReq true "删除所有子销售订单(撤销拆单)参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/delete_all [delete]
func (h *DeskHandler) OrderSaleOrderDeleteAll(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面删除所有子销售订单(撤销拆单)接口请求")

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

// RegisterDeskHandlers 注册收银产品路由
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

	// 初始化处理器
	wrapper := DeskHandler{
		service: service.NewDeskSrv(
			dbm,        // 数据库管理器
			localeSrv,  // 多语言服务
			orderSrv,   // 订单服务
			settingSrv, // 设置服务
			deviceSrv,  // 设备服务
		),
		orderService: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)                                 // 获取桌台的区域和类型
		privateApi.GET("/desk/list", wrapper.GetDeskList)                                                     // 获取桌台列表
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)                                                     // 获取桌台详情
		privateApi.POST("/desk/close", wrapper.CloseDesk)                                                     // 关闭桌台
		privateApi.POST("/desk/complete", wrapper.CompleteDesk)                                               // 完成桌台
		privateApi.POST("/desk/open", wrapper.CreateDeskOrder)                                                // 创建桌台订单(开桌)
		privateApi.POST("/desk/order/cancel", wrapper.CancelDeskOrder)                                        // 取消桌台订单
		privateApi.DELETE("/desk/order/product/delete", wrapper.OrderProductDelete)                           // 删除桌台订单商品
		privateApi.POST("/desk/order/product/price", wrapper.OrderProductChangePrice)                         // 桌台订单商品改价
		privateApi.POST("/desk/order/amount/change", wrapper.OrderAmountChange)                               // 桌台订单改价
		privateApi.POST("/desk/order/discount", wrapper.OrderDiscount)                                        // 桌台订单打折
		privateApi.POST("/desk/order/population", wrapper.OrderChangePopulation)                              // 桌台订单修改人数
		privateApi.POST("/desk/order/product/remark", wrapper.OrderProductRemark)                             // 桌台订单商品备注
		privateApi.GET("/desk/order/cart/info", wrapper.OrderCartInfo)                                        // 查询点餐购物车信息
		privateApi.POST("/desk/order/cart/product/add", wrapper.OrderCartProductAdd)                          // 向购物车添加商品
		privateApi.POST("/desk/order/cart/product/num", wrapper.OrderCartProductNum)                          // 修改购物车商品数量
		privateApi.POST("/desk/order/cart/cooking", wrapper.OrderCartProductCooking)                          // 送厨购物车商品
		privateApi.POST("/desk/order/cart/product/returning", wrapper.OrderCartProductReturning)              // 退菜购物车商品
		privateApi.POST("/desk/order/cart/product/cancel_returning", wrapper.OrderCartProductCancelReturning) // 取消退菜购物车商品
		privateApi.POST("/desk/order/cart/product/giving", wrapper.OrderCartProductGiving)                    // 赠菜购物车商品
		privateApi.POST("/desk/order/cart/product/cancel_giving", wrapper.OrderCartProductCancelGiving)       // 取消赠菜购物车商品
		privateApi.POST("/desk/order/must_plan/confirm", wrapper.OrderMustPlanConfirm)                        // 确认必点商品
		privateApi.GET("/desk/order/payment/info", wrapper.OrderPaymentInfo)                                  // 获取结账页面信息
		privateApi.POST("/desk/order/sale_order/create", wrapper.OrderSaleOrderCreate)                        // 创建一个销售订单
		privateApi.POST("/desk/order/sale_order/move_product", wrapper.OrderSaleOrderMoveProduct)             // 从一个销售订单移动商品到另一个销售订单
		privateApi.DELETE("/desk/order/sale_order/delete", wrapper.OrderSaleOrderDelete)                      // 删除一个销售订单(删除拆单)
		privateApi.DELETE("/desk/order/sale_order/delete_all", wrapper.OrderSaleOrderDeleteAll)               // 删除所有子销售订单(撤销拆单)
	}
}
