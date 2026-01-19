package cashier

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/controller"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

// InstantHandler 收银点餐处理程序
type InstantHandler struct {
	orderSrv   service.IOrderSrv   // 订单服务
	memberSrv  service.IMemberSrv  // 会员服务
	otherSrv   service.IOtherSrv   // 其他服务
	productSrv service.IProductSrv // 产品服务
}

// InvalidateSaleBillSettingCache 失效销售单设置缓存（辅助函数）
// 参数：
//   - ctx: 上下文, 用于提取 companyUuid
func (h *InstantHandler) InvalidateSaleBillSettingCache(ctx context.Context) {
	if !adapter.IsObjectStorageCacheEnabled(ctx.GetCompanyUuid()) {
		return
	}
	if err := controller.GetSaleBillSettingController().Invalidate(ctx, persistence.GlobalObjectUuid); err != nil {
		logger.Error("失效销售单设置缓存失败", zap.Error(err))
	}
}

func (h *InstantHandler) CreateInstantOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 创建订单
	res, err := h.orderSrv.CreateInstantOrder(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "创建订单失败"))
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
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	err := h.orderSrv.CancelOrder(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 点餐桌台查询时没有销售账单ID，只能通过判断请求来自哪个收银机判断是哪个销售账单。
	// 一个收银机只一个未挂单的点餐销售账单
	deviceSn := ctx.GetDeviceSn()
	if deviceSn == "" {
		err := errors.NewWithCode(constant.CodeParamError, "DeviceSn should not be an empty string")
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, ""))
		return
	}
	// 通过收银机sn获取收银机设备ID，通过设备ID查询属于该收银机的未挂单点餐账单。有0个或1个账单
	res, err := h.orderSrv.GetOrderCartInfoByDeviceSn(ctx, deviceSn)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "h.orderSrv.GetOrderCartInfoByDeviceSn failed", "deviceSn:", deviceSn))
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
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	shopCart, err := h.orderSrv.HideOrder(ctx, req.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
	shopCart, err := h.orderSrv.ShowOrder(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// SetOrderSource 设置点餐订单来源
// @Summary 设置点餐订单来源
// @Description 将点餐订单标记为某个外卖渠道
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderSetOrderSourceReq true "设置参数"
// @Success 200 {object} dto.Response{data=object} "操作成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/set_order_source [post]
func (h *InstantHandler) SetOrderSource(c *gin.Context) {
	ctx := helper.GetContext(c)
	payload := req.OrderSetOrderSourceReq{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		helper.HandleValidationError(c, err, payload, req.OrderReqMessage)
		return
	}
	shopCart, err := h.orderSrv.SetOrderSource(ctx, payload.SaleBillUuid, payload.OrderSourceUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	helper.Success(c, shopCart)
}

// SetNationality 设置点餐订单国籍
// @Summary 设置点餐订单国籍
// @Description 设置当前点餐订单的国籍信息
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderSetNationalityReq true "设置参数"
// @Success 200 {object} dto.Response{data=object} "操作成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/set_nationality [post]
func (h *InstantHandler) SetNationality(c *gin.Context) {
	ctx := helper.GetContext(c)
	payload := req.OrderSetNationalityReq{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		helper.HandleValidationError(c, err, payload, req.OrderReqMessage)
		return
	}
	shopCart, err := h.orderSrv.SetNationality(ctx, payload.SaleBillUuid, payload.NationalityUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
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
	resp, err := h.orderSrv.InstantHideOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, resp)
}

// OrderTakeout 处理打包
// @Summary 打包
// @Description 打包
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderTakeoutReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/takeout [post]
func (h *InstantHandler) OrderTakeout(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderTakeoutReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	res, err := h.orderSrv.OrderTakeout(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
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
	shopCart, err := h.orderSrv.OrderProductDelete(ctx, companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
	info, err := h.orderSrv.OrderProductChangePrice(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderDiscount 处理点餐订单打折
// @Summary 点餐订单打折
// @Description 点餐订单打折
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDiscountMethodReq true "打折参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/discount [post]
func (h *InstantHandler) OrderDiscount(c *gin.Context) {
	ctx := helper.GetContext(c)
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	// 绑定请求参数
	params := req.OrderDiscountMethodReq{}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	//
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

// OrderDiscountCancel 取消点餐订单所有优惠折扣，包括改价、打折、抹零
// @Summary 取消点餐订单所有优惠折扣，包括改价、打折、抹零
// @Description 取消点餐订单所有优惠折扣，包括改价、打折、抹零
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDiscountCancelReq true "取消优惠折扣参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/discount/cancel [post]
func (h *InstantHandler) OrderDiscountCancel(c *gin.Context) {
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
	info, err := h.orderSrv.OrderProductRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderRemark 处理点餐订单整单备注
// @Summary 点餐订单整单备注
// @Description 点餐订单整单备注
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderRemarkReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/remark [post]
func (h *InstantHandler) OrderRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	//
	info, err := h.orderSrv.OrderRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderRemarkList 处理获取整单备注列表
// @Summary 获取整单备注列表
// @Description 获取整单备注列表
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.OrderRemarkResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/remark/list [get]
func (h *InstantHandler) OrderRemarkList(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.otherSrv.GetOrderRemarkList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderItemRemarkList 处理获取单品备注列表
// @Summary 获取单品备注列表
// @Description 获取单品备注列表
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.OrderItemRemarkResp}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/item/remark/list [get]
func (h *InstantHandler) OrderItemRemarkList(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.otherSrv.GetOrderItemRemarkList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
		err := errors.NewWithCode(constant.CodeParamError, "DeviceSn should not be an empty string")
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, ""))
		return
	}
	// 通过收银机sn获取收银机设备ID，通过设备ID查询属于该收银机的未挂单点餐账单。有0个或1个账单
	res, err := h.orderSrv.GetOrderCartInfoByDeviceSn(ctx, deviceSn)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "h.orderSrv.GetOrderCartInfoByDeviceSn failed", "deviceSn:", deviceSn))
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
// @Security JwtToken
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
	res, err := h.orderSrv.InstantOrderCartProductAdd(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), errors.ErrProductPriceChanged.Error()) {
			helper.ErrorWithData(c, constant.CodeOrderCheckProductPriceChanged, res, fmt.Errorf("%s", i18n.Translate(ctx.GetLanguage(), err.Error())))
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductPackageAdd 向购物车添加套餐
// @Summary 向购物车添加套餐
// @Description 向购物车添加套餐
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductPackageAddReq true "套餐参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product_package/add [post]
func (h *InstantHandler) OrderCartProductPackageAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductPackageAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 向购物车添加套餐
	res, err := h.orderSrv.OrderCartProductPackageAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductFlavorAndAttribute 查询购物车商品“规格/属性”
// @Summary 查询购物车商品“规格/属性”
// @Description 查询购物车商品“规格/属性”
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderCartProductFlavorAndAttributeReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ProductFlavorAndAttributeRes}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/flavor_and_attribute [get]
func (h *InstantHandler) OrderCartProductFlavorAndAttribute(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductFlavorAndAttributeReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 查询购物车商品“规格/属性”
	res, err := h.orderSrv.OrderCartProductFlavorAndAttribute(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductFlavorAndAttributeChange 修改购物车商品“规格/属性”
// @Summary 修改购物车商品“规格/属性”
// @Description 修改购物车商品“规格/属性”
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductFlavorAndAttributeChangeReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/flavor_and_attribute [post]
func (h *InstantHandler) OrderCartProductFlavorAndAttributeChange(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductFlavorAndAttributeChangeReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 修改购物车商品“规格/属性”
	res, err := h.orderSrv.OrderCartProductFlavorAndAttributeChange(ctx, params)
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
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
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
	res, err := h.orderSrv.OrderCartProductNum(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCooking 送厨购物车商品
// @Summary 送厨购物车商品
// @Description 送厨购物车商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
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
	res, errRes, err := h.orderSrv.InstantOrderCartProductCooking(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if errRes != nil {
		helper.FailWithData(c, errRes.Code, errRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(errRes.Code))
		return
	}
	// 返回结果
	// 返回结果
	code := res.GetCode()
	helper.FailWithData(c, code, res, nil, constant.ParseCodeOrderCheck(code))
	// helper.Success(c, res)
}

// OrderCartProductReturning 退菜购物车商品
// @Summary 退菜购物车商品
// @Description 退菜购物车商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductReturningReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/returning [post]
func (h *InstantHandler) OrderCartProductReturning(c *gin.Context) {
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
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/cancel_returning [post]
func (h *InstantHandler) OrderCartProductCancelReturning(c *gin.Context) {
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

// OrderCartProductGiving 赠菜购物车商品
// @Summary 赠菜购物车商品
// @Description 赠菜购物车商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductGivingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/giving [post]
func (h *InstantHandler) OrderCartProductGiving(c *gin.Context) {
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
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/cancel_giving [post]
func (h *InstantHandler) OrderCartProductCancelGiving(c *gin.Context) {
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

// OrderPaymentInfo 获取结账页面信息
// @Summary 获取结账页面信息
// @Description 获取结账页面信息
// @Tags 收银端.点餐.结账
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
	params := &req.InstantOrderPaymentInfoReq{}
	if err := params.Parse(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 获取销售订单的付款信息
	res, err := h.orderSrv.InstantOrderPaymentInfo(ctx, nil, params.SaleBillUuid, params.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCoupon 选择或取消优惠券
// @Summary 选择或取消优惠券
// @Description 选择或取消优惠券
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCouponReq true "选择或取消优惠券参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/payment/coupon [post]
func (h *InstantHandler) OrderPaymentCoupon(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面选择或取消优惠券接口请求")

	var couponReq req.InstantOrderPaymentCouponReq
	if err := c.ShouldBindJSON(&couponReq); err != nil {
		helper.HandleValidationError(c, err, couponReq, nil)
		return
	}
	ctx.Log().Info("选择或取消优惠券", zap.Any("params", couponReq))
	// 选择或取消优惠券
	res, err := h.orderSrv.OrderPaymentCoupon(ctx, couponReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentQrcode 获取支付方式的二维码信息
// @Summary 获取支付方式的二维码信息
// @Description 获取支付方式的二维码信息
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.InstantOrderPaymentQrcodeReq true "获取支付方式的二维码信息参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentQrcodeInfoResp}
// @Router /cashier/instant/order/payment/qrcode [get]
func (h *InstantHandler) OrderPaymentQrcodeInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	params := req.InstantOrderPaymentQrcodeReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取支付二维码
	res, err := h.orderSrv.InstantOrderPaymentQrcode(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentInfo 设置订单的抵扣积分数量
// @Summary 设置订单的抵扣积分数量
// @Description 设置订单的抵扣积分数量
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentPointsReq true "设置订单的抵扣积分数量参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/payment/points [post]
func (h *InstantHandler) OrderPaymentPoints(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面设置订单的抵扣积分数量接口请求")

	params := req.InstantOrderPaymentPointsReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Info("设置订单的抵扣积分数量", zap.Any("params", params))
	// 设置订单的抵扣积分数量
	res, err := h.orderSrv.OrderPaymentPoints(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCreate 创建一个支付单
// @Summary 创建一个支付单
// @Description 创建一个支付单
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCreateReq true "创建一个支付单参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
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
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCancelReq true "撤销一个支付单参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
// @Router /cashier/instant/order/payment/cancel [post]
func (h *InstantHandler) OrderPaymentCancel(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面撤销一个支付单接口请求")

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
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentFinishReq true "完成销售订单的付款结账参数"
// @Success 200 {object} dto.Response{data=resp.OrderFinishResp}
// @Router /cashier/instant/order/payment/finish [post]
func (h *InstantHandler) OrderPaymentFinish(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面销售订单的付款结账接口请求")

	params := req.InstantOrderPaymentFinishReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("销售订单的付款结账", zap.Any("params", params))
	// 销售订单的付款结账
	res, err := h.orderSrv.InstantOrderPaymentFinish(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "请刷新优惠券列表") {
			// 获取销售订单的付款信息
			res, err := h.orderSrv.InstantOrderPaymentInfo(ctx, nil, params.SaleBillUuid, params.SaleOrderUuid)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
				return
			}
			// 返回结果
			helper.ErrorWithData(c, constant.CodeCouponInvalid, res, fmt.Errorf("%s", i18n.Translate(ctx.GetLanguage(), "优惠券信息变化，请重新确认。")))
			return
		}
		if strings.Contains(err.Error(), "物品库存不足") {
			ctx.Log().Error("桌台销售订单的付款结账失败", zap.Any("err", err))
			itemCode := ""
			re := regexp.MustCompile(`物品库存不足,(WPR\d+)`)
			matches := re.FindStringSubmatch(err.Error())
			if len(matches) > 1 {
				itemCode = matches[1]
			}
			productInfos, err := h.orderSrv.GetProductNameByItemCode(ctx, itemCode, params.SaleOrderUuid)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
				return
			}
			productList := make([]resp.Product, 0)
			for _, productInfo := range productInfos {
				productList = append(productList, resp.Product{
					LocaleName: productInfo.ProductName,
				})
			}
			orderCheckRes := &resp.OrderCheckRes{
				Products: &resp.CartProductList{
					List: productList,
				},
			}
			helper.FailWithData(c, constant.CodeOrderCheckProductStockZero, orderCheckRes, nil, i18n.Translate(ctx.GetLanguage(), "以下商品库存不足，请删除后再下单"))
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx.Log().Debug("销售订单的付款结账成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderFree 免单
// @Summary 免单
// @Description 免单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderFreeReq true "免单参数"
// @Success 200 {object} dto.Response{data=resp.OrderFinishResp}
// @Router /cashier/instant/order/free [post]
func (h *InstantHandler) OrderFree(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面免单接口请求")

	params := req.InstantOrderFreeReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("免单", zap.Any("params", params))
	// 免单
	res, err := h.orderSrv.InstantOrderFree(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "物品库存不足") {
			ctx.Log().Error("桌台销售订单的付款结账失败", zap.Any("err", err))
			itemCode := ""
			re := regexp.MustCompile(`物品库存不足,(WPR\d+)`)
			matches := re.FindStringSubmatch(err.Error())
			if len(matches) > 1 {
				itemCode = matches[1]
			}
			productInfos, err := h.orderSrv.GetProductNameByItemCode(ctx, itemCode, params.SaleOrderUuid)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
				return
			}
			productList := make([]resp.Product, 0)
			for _, productInfo := range productInfos {
				productList = append(productList, resp.Product{
					LocaleName: productInfo.ProductName,
				})
			}
			orderCheckRes := &resp.OrderCheckRes{
				Products: &resp.CartProductList{
					List: productList,
				},
			}
			helper.FailWithData(c, constant.CodeOrderCheckProductStockZero, orderCheckRes, nil, i18n.Translate(ctx.GetLanguage(), "以下商品库存不足，请删除后再下单"))
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("免单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentZeroRule 设置结账抹零规则
// @Summary 设置结账抹零规则
// @Description 设置结账抹零规则
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentZeroRuleReq true "设置结账抹零规则参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
// @Router /cashier/instant/order/payment/zero_rule [post]
func (h *InstantHandler) OrderPaymentZeroRule(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面设置结账抹零规则接口请求")

	params := req.InstantOrderPaymentZeroRuleReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	if err := params.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Info("设置结账抹零规则", zap.Any("params", params))
	// 设置结账抹零规则
	res, err := h.orderSrv.InstantOrderPaymentZeroRule(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("设置结账抹零规则成功", zap.Any("res", res))
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
	params := req.InstantOrderSaleOrderCreateReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 创建一个销售订单
	res, err := h.orderSrv.InstantOrderSaleOrderCreate(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderSaleOrderMoveProduct 从一个销售订单移动商品到另一个销售订单
// @Summary 从一个销售订单移动商品到另一个销售订单
// @Description 从一个销售订单移动商品到另一个销售订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
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
	res, err := h.orderSrv.SaleOrderMoveProduct(ctx, params, false)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("从一个销售订单移动商品到另一个销售订单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res, "拆单成功")
}

// OrderMustPlanConfirm 确认必点商品
// @Summary 确认必点商品
// @Description 确认必点商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
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
	res, mustPlan, err := h.orderSrv.InstantOrderMustPlanConfirm(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if !res {
		helper.ErrorWithDetail(c, constant.CodeOrderCheckProductMust, errors.New(fmt.Sprintf("【%s】%s", mustPlan.Name, i18n.Translate(ctx.GetLanguage(), errors.ErrMustPlanNotComplete.Error()))))
		return
	}
	ctx.Log().Debug("确认必点商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderCheck 订单检查
// @Summary 订单检查
// @Description 订单检查。场景：1、点击结账按钮时，检查订单是否可以结账
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @Success 200 {object} dto.Response{data=resp.OrderCheckRes}
// @Router /cashier/instant/order/check [get]
func (h *InstantHandler) OrderCheck(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面订单检查接口请求")

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
		// 如果开启缓存,要失效sale_bill_setting的缓存
		h.InvalidateSaleBillSettingCache(ctx)
		ctx.Log().Debug("送厨检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	ctx.Log().Debug("订单检查成功")
	// 返回结果
	helper.Success(c, resp.OrderCheckRes{})
}

// OrderSaleOrderDelete 删除一个销售订单(删除拆单)
// @Summary 删除一个销售订单(删除拆单)
// @Description 删除一个销售订单(删除拆单)
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
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
	res, err := h.orderSrv.InstantOrderSaleOrderDelete(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
// @Security JwtToken
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
	res, err := h.orderSrv.InstantOrderSaleOrderDeleteAll(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("删除所有子销售订单(撤销拆单)成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// GetMemberDiscount 获取订单会员优惠
// @Summary 获取订单会员优惠
// @Description 获取订单会员优惠
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @param member_uuid query integer true "会员Uuid"
// @Success 200 {object} dto.Response{data=resp.MemberDiscountResp}
// @Router /cashier/instant/order/member/discount [get]
func (h *InstantHandler) GetMemberDiscount(c *gin.Context) {
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
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CheckMemberPasswordReq true "确认使用会员优惠并验证密码"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Router /cashier/instant/order/member/confirm [post]
func (h *InstantHandler) OrderUseMember(c *gin.Context) {
	var passwordReq req.CheckMemberPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, req.CheckMemberPasswordMessage)
		return
	}
	ctx := helper.GetContext(c)
	res, isCustomAmountAndZero, err := h.orderSrv.OrderUseMember(ctx, passwordReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if isCustomAmountAndZero {
		helper.FailWithData(c, constant.CodeSuccess, res, nil, "改价/抹零已失效，请重新进行改价/抹零操作")
		return
	}
	helper.Success(c, res)
}

// OrderMemberCancel 不使用此会员
// @Summary 不使用此会员
// @Description 不使用此会员
// @Tags 收银端.点餐.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderMemberCancelReq true "不使用此会员"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Router /cashier/instant/order/member/cancel [delete]
func (h *InstantHandler) OrderMemberCancel(c *gin.Context) {
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
// @Summary 打印小票
// @Description 打印小票
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/instant/order/print [post]
func (h *InstantHandler) OrderPrint(c *gin.Context) {
	var printReq req.OrderPrintReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderPrint(ctx, printReq, true)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res, "发送成功")
}

// OrderPrint 打印发票
// @Summary 打印发票
// @Description 打印发票
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintInvoiceReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/instant/order/print/invoice [post]
func (h *InstantHandler) OrderPrintInvoice(c *gin.Context) {
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

// OrderUnlock 订单解锁
// @Summary 订单解锁
// @Description 订单解锁
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderUnlockReq true "订单解锁"
// @Success 200 {object} dto.Response
// @Router /cashier/instant/order/unlock [post]
func (h *InstantHandler) OrderUnlock(c *gin.Context) {
	var unlockReq req.OrderUnlockReq
	if err := c.ShouldBindJSON(&unlockReq); err != nil {
		helper.HandleValidationError(c, err, unlockReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	err := h.orderSrv.OrderUnlock(ctx, unlockReq.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetOrderMemberList 获取订单会员列表
// @Summary 获取订单会员列表
// @Description 获取订单会员列表
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.InstantOrderMemberList}
// @Router /cashier/instant/order/member/list [get]
func (h *InstantHandler) GetOrderMemberList(c *gin.Context) {
	var req req.OrderGetOrderMemberListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.GetOrderMemberList(ctx, req.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// OrderCartProductBatchCooking 获取分批送厨弹框的销售订单商品列表
// @Summary 获取分批送厨弹框的销售订单商品列表
// @Description 获取分批送厨弹框的销售订单商品列表
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.GetOrderCartProductBatchCookingListReq true "获取分批送厨弹框的销售订单商品列表"
// @Success 200 {object} dto.Response{data=resp.OrderCartProductBatchCookingRes}
// @Router /cashier/instant/order/cart/batch/cooking [get]
func (h *InstantHandler) OrderCartProductBatchCookingList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetOrderCartProductBatchCookingListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 获取分批送厨弹框的销售订单商品列表
	res, err := h.orderSrv.GetOrderCartProductBatchCookingList(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductBatchCooking 分批送厨弹框的销售订单商品列表
// @Summary 分批送厨弹框的销售订单商品列表
// @Description 分批送厨弹框的销售订单商品列表
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductBatchCookingReq true "分批送厨弹框的销售订单商品列表"
// @Success 200 {object} dto.Response{data=resp.OrderCartProductBatchCooking}
// @Router /cashier/instant/order/cart/batch/cooking [post]
func (h *InstantHandler) OrderCartProductBatchCooking(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductBatchCookingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 分批送厨弹框的销售订单商品列表
	res, err := h.orderSrv.OrderCartProductBatchCooking(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetBatchTagList 获取分批类型列表
// @Summary 获取分批类型列表
// @Description 获取分批类型列表，按 sort 排序，优先级高的在前
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=product_resp.BatchTagList}
// @Router /cashier/instant/batch_tag/list [get]
func (h *InstantHandler) GetBatchTagList(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 调用 Service 层获取分批类型列表
	batchTagList, err := h.productSrv.GetBatchTagList(ctx, req.BatchTagListReq{})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, batchTagList)
}

// ChangeBatchTag 更换分批类型（前置模式）
// @Summary 更换分批类型
// @Description 更换未送厨商品的分批类型（前置模式）
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ChangeBatchTagReq true "更换分批类型请求"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/instant/order/cart/batch/change_tag [post]
func (h *InstantHandler) ChangeBatchTag(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.ChangeBatchTagReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 更换分批类型
	res, err := h.orderSrv.ChangeBatchTag(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentActivity 选择或取消满减活动
// @Summary 选择或取消满减活动
// @Description 选择或取消满减活动
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentActivityReq true "选择或取消满减活动参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/payment/activity [post]
func (h *InstantHandler) OrderPaymentActivity(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到点餐页面选择或取消满减活动接口请求")

	var activityReq req.InstantOrderPaymentActivityReq
	if err := c.ShouldBindJSON(&activityReq); err != nil {
		helper.HandleValidationError(c, err, activityReq, nil)
		return
	}
	ctx.Log().Info("选择或取消满减活动", zap.Any("params", activityReq))
	// 选择或取消满减活动
	res, err := h.orderSrv.OrderPaymentActivity(ctx, activityReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterInstantHandlers 注册收银订单路由
func RegisterInstantHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	otherSrv := service.NewOtherSrv(dbm, cache, settingSrv)
	translateSrv := service.NewTranslateSrv(dbm, cache)
	productSrv := service.NewProductSrv(dbm, localeSrv, settingSrv, cache, translateSrv)
	// 创建收银产品处理程序
	wrapper := InstantHandler{
		orderSrv:   orderSrv,   // 订单服务
		memberSrv:  memberSrv,  // 会员服务
		otherSrv:   otherSrv,   // 其他服务
		productSrv: productSrv, // 产品服务
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/instant/order/create", wrapper.CreateInstantOrder)                                                  // 创建点餐订单。 废弃，点餐点餐由系统自动创建
		privateApi.POST("/instant/order/cancel", wrapper.CancelOrder)                                                         // 取消点餐订单
		privateApi.POST("/instant/order/hide", wrapper.HideOrder)                                                             // 隐藏点餐订单（挂单）
		privateApi.POST("/instant/order/show", wrapper.ShowOrder)                                                             // 显示点餐订单（取单）
		privateApi.POST("/instant/set_order_source", wrapper.SetOrderSource)                                                  // 设置点餐订单来源
		privateApi.POST("/instant/set_nationality", wrapper.SetNationality)                                                   // 设置点餐订单国籍
		privateApi.GET("/instant/order/list", wrapper.OrderList)                                                              // 显示点餐订单列表（取单列表）
		privateApi.POST("/instant/order/takeout", wrapper.OrderTakeout)                                                       // 打包
		privateApi.DELETE("/instant/order/product/delete", wrapper.OrderProductDelete)                                        // 删除点餐订单商品
		privateApi.POST("/instant/order/product/price", wrapper.OrderProductChangePrice)                                      // 点餐订单商品改价
		privateApi.POST("/instant/order/discount", wrapper.OrderDiscount)                                                     // 点餐订单打折
		privateApi.POST("/instant/order/discount/cancel", wrapper.OrderDiscountCancel)                                        // 取消点餐订单所有优惠折扣，包括改价、打折、抹零（撤销优惠折扣）
		privateApi.POST("/instant/order/product/remark", wrapper.OrderProductRemark)                                          // 点餐订单商品备注
		privateApi.POST("/instant/order/remark", wrapper.OrderRemark)                                                         // 整单备注
		privateApi.GET("/instant/order/remark/list", wrapper.OrderRemarkList)                                                 // 获取整单备注列表
		privateApi.GET("/instant/order/item/remark/list", wrapper.OrderItemRemarkList)                                        // 获取单品备注列表
		privateApi.GET("/instant/order/cart/info", wrapper.OrderCartInfo)                                                     // 查询点餐购物车信息
		privateApi.POST("/instant/order/cart/product/add", wrapper.OrderCartProductAdd)                                       // 向购物车添加商品
		privateApi.POST("/instant/order/cart/product_package/add", wrapper.OrderCartProductPackageAdd)                        // 向购物车添加套餐
		privateApi.GET("/instant/order/cart/product/flavor_and_attribute", wrapper.OrderCartProductFlavorAndAttribute)        // 查询购物车商品“规格/属性”
		privateApi.POST("/instant/order/cart/product/flavor_and_attribute", wrapper.OrderCartProductFlavorAndAttributeChange) // 修改购物车商品“规格/属性”
		privateApi.POST("/instant/order/cart/product/num", wrapper.OrderCartProductNum)                                       // 修改购物车商品数量
		privateApi.POST("/instant/order/cart/cooking", wrapper.OrderCartProductCooking)                                       // 送厨购物车商品
		privateApi.POST("/instant/order/cart/product/returning", wrapper.OrderCartProductReturning)                           // 退菜购物车商品
		privateApi.POST("/instant/order/cart/product/cancel_returning", wrapper.OrderCartProductCancelReturning)              // 取消退菜购物车商品
		privateApi.POST("/instant/order/cart/product/giving", wrapper.OrderCartProductGiving)                                 // 赠菜购物车商品
		privateApi.POST("/instant/order/cart/product/cancel_giving", wrapper.OrderCartProductCancelGiving)                    // 取消赠菜购物车商品
		privateApi.POST("/instant/order/must_plan/confirm", wrapper.OrderMustPlanConfirm)                                     // 确认必点商品
		privateApi.GET("/instant/order/check", wrapper.OrderCheck)                                                            // 订单检查。场景：1、点击结账按钮时，检查订单是否可以结账
		privateApi.GET("/instant/order/payment/info", wrapper.OrderPaymentInfo)                                               // 获取结账页面信息
		privateApi.POST("/instant/order/payment/coupon", wrapper.OrderPaymentCoupon)                                          // 选择或取消优惠券
		privateApi.POST("/instant/order/payment/points", wrapper.OrderPaymentPoints)                                          // 设置订单的抵扣积分数量
		privateApi.POST("/instant/order/payment/create", wrapper.OrderPaymentCreate)                                          // 创建一个支付单
		privateApi.POST("/instant/order/payment/cancel", wrapper.OrderPaymentCancel)                                          // 撤销一个支付单
		privateApi.POST("/instant/order/payment/finish", wrapper.OrderPaymentFinish)                                          // 完成销售订单的付款结账
		privateApi.POST("/instant/order/free", wrapper.OrderFree)                                                             // 免单
		privateApi.POST("/instant/order/payment/zero_rule", wrapper.OrderPaymentZeroRule)                                     // 设置结账抹零规则
		privateApi.POST("/instant/order/sale_order/create", wrapper.OrderSaleOrderCreate)                                     // 创建一个销售订单
		privateApi.POST("/instant/order/sale_order/move_product", wrapper.OrderSaleOrderMoveProduct)                          // 从一个销售订单移动商品到另一个销售订单
		privateApi.DELETE("/instant/order/sale_order/delete", wrapper.OrderSaleOrderDelete)                                   // 删除一个销售订单(删除拆单)
		privateApi.DELETE("/instant/order/sale_order/delete_all", wrapper.OrderSaleOrderDeleteAll)                            // 删除所有子销售订单(撤销拆单)
		privateApi.GET("/instant/order/member/discount", wrapper.GetMemberDiscount)                                           // 获取订单会员优惠
		privateApi.POST("/instant/order/member/confirm", wrapper.OrderUseMember)                                              // 确认使用会员优惠并验证密码
		privateApi.DELETE("/instant/order/member/cancel", wrapper.OrderMemberCancel)                                          // 不使用此会员
		privateApi.POST("/instant/order/print", wrapper.OrderPrint)                                                           // 打印
		privateApi.POST("/instant/order/print/invoice", wrapper.OrderPrintInvoice)                                            // 打印发票
		privateApi.POST("/instant/order/unlock", wrapper.OrderUnlock)                                                         // 订单解锁
		privateApi.GET("/instant/order/payment/qrcode", wrapper.OrderPaymentQrcodeInfo)                                       // 获取支付方式的二维码信息
		privateApi.GET("/instant/order/member/list", wrapper.GetOrderMemberList)                                              // 获取订单会员列表
		privateApi.GET("/instant/order/cart/batch/cooking", wrapper.OrderCartProductBatchCookingList)                         // 获取分批送厨弹框的销售订单商品列表
		privateApi.POST("/instant/order/cart/batch/cooking", wrapper.OrderCartProductBatchCooking)                            // 分批送厨
		privateApi.GET("/instant/batch_tag/list", wrapper.GetBatchTagList)                                                    // 获取分批类型列表
		privateApi.POST("/instant/order/cart/batch/change_tag", wrapper.ChangeBatchTag)                                       // 更换分批类型
		privateApi.POST("/instant/order/payment/activity", wrapper.OrderPaymentActivity)                                      // 选择或取消满减活动
	}
}
