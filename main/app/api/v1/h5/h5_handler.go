package h5

import (
	"fmt"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// Handler 商家H5端处理程序
type Handler struct {
	h5Srv      service.IH5Srv      // h5扫码服务
	deskSrv    service.IDeskSrv    // 桌台服务
	buffetSrv  service.IBuffetSrv  // 自助餐服务
	productSrv service.IProductSrv // 产品服务
	callSrv    service.ICallSrv    // 呼叫服务
	orderSrv   service.IOrderSrv   // 订单服务
}

// BaseInfo 获取桌码基础信息
// @Summary 桌码基础信息
// @Description 获取桌码基础信息，整个h5应用需要的基础信息
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.H5BaseInfo{}
// @Router /h5/base [get]
func (h *Handler) BaseInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	deskUuid := ctx.GetDeskUuid()
	ctx.Log().Info("GetBaseInfo", zap.Uint64("deskUuid", deskUuid))
	info, err := h.h5Srv.GetBaseInfo(ctx, deskUuid)
	if err != nil {
		ctx.Log().Info("获取桌台基本信息失败", zap.String("error", err.Error()))
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// GetBuffetList 获取自助餐套餐列表
// @Summary 自助餐套餐信息
// @Description 获取自助餐套餐列表
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.BuffetListPaginationResp "自助餐列表"
// @Router /h5/buffet/list [get]
func (h *Handler) GetBuffetList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 获取自助餐列表
	res, err := h.buffetSrv.GetBuffetList(companyUuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OpenDesk 处理创建开台
// @Summary 开台
// @Description 开台
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "开台参数"
// @Success 200 {object} resp.CreateDeskOrderResp "开台成功"
// @Failure 404 {object} nil "未找到"
// @Router /h5/desk/open [post]
func (h *Handler) OpenDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取桌台uuid
	params.DeskUuid = ctx.GetDeskUuid()
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

// GetProductCategoryList 获取收银产品类别列表
// @Summary 获取收银产品类别列表
// @Description 获取收银产品类别列表
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} product_resp.ProductCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /h5/product/category/list [get]
func (h *Handler) GetProductCategoryList(c *gin.Context) {
	// 获取收银产品类别列表
	res, err := h.productSrv.GetProductCategoryList(helper.GetCompanyUuid(c))

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// GetProductList 获取收银产品列表
// @Summary 获取收银产品列表
// @Description 获取收银产品列表
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页条数"
// @Success 200 {object} product_resp.ProductListWithPaginationResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /h5/product/list [get]
func (h *Handler) GetProductList(c *gin.Context) {
	// 绑定请求参数
	productListReq := req.ProductListReq{}
	if err := c.ShouldBindQuery(&productListReq); err != nil {
		helper.HandleValidationError(c, err, productListReq, dto.PageReqMessage)
		return
	}

	// 获取收银产品列表
	res, err := h.productSrv.GetProductList(helper.GetContext(c), productListReq)

	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// Call 发起呼叫
// @Summary 发起呼叫
// @Description 发起呼叫
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CallReq true "呼叫请求"
// @Success 200 {object} dto.Response
// @Router /h5/call [post]
func (h *Handler) Call(c *gin.Context) {
	ctx := helper.GetContext(c)
	var callReq req.CallReq
	if err := c.ShouldBindJSON(&callReq); err != nil {
		helper.HandleValidationError(c, err, callReq, nil)
		return
	}
	err := h.callSrv.Call(ctx, callReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "已呼叫服务员，请稍等")
}

// OrderProductRemark 处理桌台订单商品备注
// @Summary 桌台订单商品备注
// @Description 桌台订单商品备注
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.H5DeskPing}
// @Failure 404 {object} nil "未找到"
// @Router /h5/remark [post]
func (h *Handler) OrderProductRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderSrv.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 都是加购到第一个子单中
	params.SaleOrderUuid = saleOrderUuid
	params.SaleBillUuid = saleBillUuid
	shopCart, err := h.orderSrv.OrderProductRemark(ctx, params, repository.WithUnorderedH5Product())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	res, err := h.deskSrv.GetH5DeskPing(helper.GetContext(c), shopCart.Desk.Uuid, shopCart)
	// 处理错误
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
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.H5DeskPing}
// @Failure 404 {object} nil "未找到"
// @Router /h5/order/cart/product/add [post]
func (h *Handler) OrderCartProductAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderSrv.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 都是加购到第一个子单中
	params.SaleOrderUuid = saleOrderUuid
	params.SaleBillUuid = saleBillUuid
	if params.SaleOrderUuid == 0 || params.SaleBillUuid == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(errors.New("没有桌台账单")))
		return
	}
	params.SetIsH5Product() // 设置为H5商品
	// 添加商品。 若没有点餐账单则新建一个
	shopCart, err := h.orderSrv.InstantOrderCartProductAdd(ctx, params, repository.WithUnorderedH5Product())
	if err != nil {
		if strings.Contains(err.Error(), errors.ErrProductPriceChanged.Error()) { //  [app/service/order.go:4264]: [app/service/order.go:4562]: [app/service/order_action.go:271]: [app/service/order_action.go:586]: 商品超过限购
			fmt.Println("InstantOrderCartProductAdd 1111", err)
			res := resp.H5DeskPing{
				Product: shopCart.Product,
			}
			helper.ErrorWithData(c, constant.CodeOrderCheckProductPriceChanged, res, fmt.Errorf("%s", i18n.Translate(ctx.GetLanguage(), err.Error())))
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	_, err = h.orderSrv.GetUnOrderedH5ProductList(ctx, saleBillUuid, shopCart, repository.WithUnorderedH5Product())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	res, err := h.deskSrv.GetH5DeskPing(helper.GetContext(c), shopCart.Desk.Uuid, shopCart)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductPackageAdd 向购物车添加套餐
// @Summary 向购物车添加套餐
// @Description 向购物车添加套餐
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductPackageAddReq true "套餐参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /h5/order/cart/product_package/add [post]
func (h *Handler) OrderCartProductPackageAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductPackageAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}

	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderSrv.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 都是加购到第一个子单中
	params.SaleOrderUuid = saleOrderUuid
	params.SaleBillUuid = saleBillUuid
	if params.SaleOrderUuid == 0 || params.SaleBillUuid == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(errors.New("没有桌台账单")))
		return
	}
	params.SetIsH5Product() // 设置为H5商品

	// 向购物车添加套餐
	res, err := h.orderSrv.OrderCartProductPackageAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetOrderCartProductUnordered 查询未下单商品
// @Summary 查询未下单商品
// @Description 查询未下单商品
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.UnsentKitchen}
// @Failure 404 {object} nil "未找到"
// @Router /h5/order/cart/product/unordered/list [get]
func (h *Handler) GetOrderCartProductUnordered(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderSrv.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if saleBillUuid == 0 || saleOrderUuid == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(errors.New("没有桌台账单")))
		return
	}
	res, err := h.orderSrv.GetUnOrderedH5ProductList(ctx, saleBillUuid, nil, repository.WithUnorderedH5Product())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetOrderCartProductOrdered 查询已下单商品
// @Summary 查询已下单商品
// @Description 查询已下单商品
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.H5CartSendProduct}
// @Failure 404 {object} nil "未找到"
// @Router /h5/order/cart/product/ordered/list [get]
func (h *Handler) GetOrderCartProductOrdered(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderSrv.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if saleBillUuid == 0 || saleOrderUuid == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(errors.New("没有桌台账单")))
		return
	}
	// // 如果关闭h5接单功能，则拒单所有h5待接单订单。
	// {
	// 	companySetting := ctx.GetCompanySetting()
	// 	if !companySetting.GetIsOpenH5Order() {
	// 		h.orderSrv.RejectAllH5Order(ctx, saleBillUuid)
	// 	}
	// }
	res, err := h.orderSrv.GetOrderedH5ProductList(ctx, saleBillUuid, nil, repository.WithOrderedH5ProductWithReject())
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
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductNumReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.UnsentKitchen}
// @Failure 404 {object} nil "未找到"
// @Router /h5/order/cart/product/num [post]
func (h *Handler) OrderCartProductNum(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面修改购物车商品数量接口请求")
	// 绑定请求参数
	params := req.OrderCartProductNumReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderSrv.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if saleBillUuid == 0 || saleOrderUuid == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(errors.New("没有桌台账单")))
		return
	}
	params.SaleOrderUuid = saleOrderUuid
	params.SaleBillUuid = saleBillUuid
	ctx.Log().Debug("扫码点餐页面修改购物车商品数量接口请求", zap.Any("params", params))
	// 修改购物车商品数量
	shopCart, err := h.orderSrv.OrderCartProductNum(ctx, params, repository.WithUnorderedH5Product())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	_, err = h.orderSrv.GetUnOrderedH5ProductList(ctx, saleBillUuid, shopCart, repository.WithUnorderedH5Product())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	res, err := h.deskSrv.GetH5DeskPing(helper.GetContext(c), shopCart.Desk.Uuid, shopCart)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// ConfirmOrder 确认下单
// @Summary 确认下单
// @Description 确认下单
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.H5CartSendProduct}
// @Failure 404 {object} nil "未找到"
// @Router /h5/order/cart/confirm [post]
func (h *Handler) ConfirmOrder(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderSrv.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if saleBillUuid == 0 || saleOrderUuid == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(errors.New("没有桌台账单")))
		return
	}

	// 绑定请求参数
	params := req.H5ConfirmOrderReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	if failedData, err := h.orderSrv.ConfirmH5Order(ctx, saleBillUuid, saleOrderUuid, params.IgnoreMust); err != nil {
		helper.ErrorWithData(c, constant.CodeFail, failedData, err)
		return
	}
	res, err := h.orderSrv.GetOrderedH5ProductList(ctx, saleBillUuid, nil, repository.WithOrderedH5ProductWithReject())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskBuffetProductList 处理获取自助餐商品列表
// @Summary 获取自助餐商品列表
// @Description 获取自助餐商品列表
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderChangeBuffetProductListReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.BuffetProductList}
// @Failure 404 {object} nil "未找到"
// @Router /h5/desk/order/buffet/product/list [get]
func (h *Handler) GetDeskBuffetProductList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangeBuffetProductListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("获取自助餐商品列表", zap.Any("params", params))
	//
	info, err := h.orderSrv.OrderDeskBuffetProductList(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderMustPlanConfirm 确认必点商品
// @Summary 确认必点商品
// @Description 确认必点商品
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderMustPlanConfirmReq true "确认必点商品参数"
// @Success 200 {object} dto.Response{}
// @Router /h5/desk/order/must_plan/confirm [post]
func (h *Handler) OrderMustPlanConfirm(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面确认必点商品接口请求")

	params := req.InstantOrderMustPlanConfirmReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("确认必点商品", zap.Any("params", params))
	// 确认必点商品
	res, mustPlan, err := h.orderSrv.InstantOrderMustPlanConfirm(ctx, params, service.WithIsH5Order())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if !res {
		helper.ErrorWithDetail(c, constant.CodeOrderCheckProductMust, errors.New(fmt.Sprintf("【%s】%s", mustPlan.Name, i18n.Translate(ctx.GetLanguage(), errors.ErrMustPlanNotComplete.Error()))))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// GetDeskPing 桌台详情-用于定时轮询
// @Summary 桌台详情-用于定时轮询
// @Description 桌台详情-用于定时轮询
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.H5DeskPing} "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /h5/desk/ping [get]
func (h *Handler) GetDeskPing(c *gin.Context) {
	ctx := helper.GetContext(c)
	deskUuid := ctx.GetDeskUuid()
	// // 绑定请求参数
	// var deskInfoReq req.DeskInfoReq
	// if err := c.ShouldBindQuery(&deskInfoReq); err != nil {
	// 	helper.HandleValidationError(c, err, deskInfoReq, nil)
	// 	return
	// }
	res, err := h.deskSrv.GetH5DeskPing(helper.GetContext(c), deskUuid, nil)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetProductPackageDetail 获取商品选购详情
// @Summary 获取商品选购详情
// @Description 获取商品选购详情
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.GetProductPackageDetailReq true "商品选购详情参数"
// @Success 200 {object} dto.Response{data=resp.ProductPackageDetailRes}
// @Failure 404 {object} nil "未找到"
// @Router /h5/product/package/detail [get]
func (h *Handler) GetProductPackageDetail(c *gin.Context) {
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

// RegisterH5Handlers 注册扫码h5路由
func RegisterH5Handlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	deskSrv := service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv, mustPlanSrv)
	buffetSrv := service.NewBuffetSrv(dbm)
	h5Srv := service.NewH5Srv(dbm, deskSrv, orderSrv, buffetSrv, settingSrv)
	productService := service.NewProductSrv(dbm, localeSrv, settingSrv)
	callSrv := service.NewCallSrv(dbm)
	// 初始化处理器
	wrapper := Handler{
		h5Srv:      h5Srv,
		deskSrv:    deskSrv,
		buffetSrv:  buffetSrv,
		productSrv: productService,
		callSrv:    callSrv,
		orderSrv:   orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.DeskAuth(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.BaseInfo)                                                   // 获取桌码基础信息
		privateApi.GET("/buffet/list", wrapper.GetBuffetList)                                       // 获取自助餐套餐列表
		privateApi.POST("/desk/open", wrapper.OpenDesk)                                             // 开台
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList)                    // 获取收银产品类别列表
		privateApi.GET("/product/list", wrapper.GetProductList)                                     // 获取收银产品列表
		privateApi.POST("/call", wrapper.Call)                                                      // 发起呼叫
		privateApi.POST("/remark", wrapper.OrderProductRemark)                                      // 给商品添加备注
		privateApi.POST("/order/cart/product/add", wrapper.OrderCartProductAdd)                     // 向购物车添加商品
		privateApi.POST("/desk/order/cart/product_package/add", wrapper.OrderCartProductPackageAdd) // 向购物车添加套餐
		privateApi.GET("/order/cart/product/unordered/list", wrapper.GetOrderCartProductUnordered)  // 查询购物车未下单商品列表
		privateApi.GET("/order/cart/product/ordered/list", wrapper.GetOrderCartProductOrdered)      // 查询购物车已下单商品列表
		privateApi.POST("/order/cart/product/num", wrapper.OrderCartProductNum)                     // 修改购物车商品数量
		privateApi.POST("/order/cart/confirm", wrapper.ConfirmOrder)                                // 确认下单
		privateApi.GET("/desk/order/buffet/product/list", wrapper.GetDeskBuffetProductList)         // 获取自助餐商品列表
		privateApi.POST("/desk/order/must_plan/confirm", wrapper.OrderMustPlanConfirm)              // 确认必点商品
		privateApi.GET("/desk/ping", wrapper.GetDeskPing)                                           // 定时获取桌台信息
		// 通过product_package_uuid获取该商品的选购详情，如某个规格属性加料的组合被选购了多少各
		privateApi.GET("/product/package/detail", wrapper.GetProductPackageDetail) // 获取商品选购详情
	}
}
