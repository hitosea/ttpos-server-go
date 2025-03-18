package h5

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// H5Handler 商家H5端处理程序
type H5Handler struct {
	service        service.IH5Srv      // h5扫码服务
	deskSrv        service.IDeskSrv    // 桌台服务
	buffetSrv      service.IBuffetSrv  // 自助餐服务
	productService service.IProductSrv // 产品服务
	callSrv        service.ICallSrv    // 呼叫服务
	orderService   service.IOrderSrv   // 订单服务
}

// BaseInfo 获取桌码基础信息
// @Summary 桌码基础信息
// @Description 获取桌码基础信息，整个h5应用需要的基础信息
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.H5BaseInfo{}
// @Router /h5/base/info [get]
func (h H5Handler) BaseInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	deskUuid := ctx.GetDeskUuid()
	ctx.Log().Info("GetBaseInfo", zap.Uint64("deskUuid", deskUuid))
	info, err := h.service.GetBaseInfo(ctx, deskUuid)
	if err != nil {
		ctx.Log().Info("获取桌台基本信息失败", zap.String("error", err.Error()))
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// GetBaseInfo 获取自助餐套餐列表
// @Summary 自助餐套餐信息
// @Description 获取自助餐套餐列表
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.BuffetListPaginationResp "自助餐列表"
// @Router /h5/buffet/list [get]
func (h *H5Handler) GetBuffetList(c *gin.Context) {
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

// CreateDeskOrder 处理创建开台
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
func (h *H5Handler) OpenDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取桌台uuid
	params.DeskUuid = ctx.GetDeskUuid()
	// todo 判断门店设置扫码H5能否开台
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
func (h *H5Handler) GetProductCategoryList(c *gin.Context) {
	// 获取收银产品类别列表
	res, err := h.productService.GetProductCategoryList(helper.GetCompanyUuid(c))

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
func (h *H5Handler) GetProductList(c *gin.Context) {
	// 绑定请求参数
	req := req.ProductListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}

	// 获取收银产品列表
	res, err := h.productService.GetProductList(helper.GetContext(c), req)

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
func (h *H5Handler) Call(c *gin.Context) {
	ctx := helper.GetContext(c)
	var callReq req.CallReq
	if err := c.ShouldBindJSON(&callReq); err != nil {
		helper.HandleValidationError(c, err, callReq, nil)
		return
	}
	callReq.DeskUuid = ctx.GetDeskUuid() // 设置桌台uuid
	err := h.callSrv.Call(ctx, callReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// OrderProductRemark 处理桌台订单商品备注
// @Summary 桌台订单商品备注
// @Description 桌台订单商品备注
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /h5/remark [post]
func (h *H5Handler) OrderProductRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderService.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 都是加购到第一个子单中
	params.SaleOrderUuid = saleOrderUuid
	params.SaleBillUuid = saleBillUuid
	info, err := h.orderService.OrderProductRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderCartProductAdd 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /h5/order/cart/product/add [post]
func (h *H5Handler) OrderCartProductAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取桌台的账单uuid和第一子单的uuid
	saleBillUuid, saleOrderUuid, err := h.orderService.GetSaleBillUuidAndSaleOrderUuid(ctx, ctx.GetDeskUuid())
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
	res, err := h.orderService.InstantOrderCartProductAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterH5Handlers 注册扫码h5路由
func RegisterH5Handlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	deskSrv := service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv)
	buffetSrv := service.NewBuffetSrv(dbm, localeSrv)
	h5Srv := service.NewH5Srv(dbm, deskSrv, orderSrv, buffetSrv, settingSrv)
	productService := service.NewProductSrv(dbm, localeSrv)
	callSrv := service.NewCallSrv(dbm)
	// 初始化处理器
	wrapper := H5Handler{
		service:        h5Srv,
		deskSrv:        deskSrv,
		buffetSrv:      buffetSrv,
		productService: productService,
		callSrv:        callSrv,
		orderService:   orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.DeskAuth(authSrv, dbm))
	{
		privateApi.GET("/base/info", wrapper.BaseInfo)                           // 获取桌码基础信息
		privateApi.GET("/buffet/list", wrapper.GetBuffetList)                    // 获取自助餐套餐列表
		privateApi.POST("/desk/open", wrapper.OpenDesk)                          // 开台
		privateApi.GET("/product/category/list", wrapper.GetProductCategoryList) // 获取收银产品类别列表
		privateApi.GET("/product/list", wrapper.GetProductList)                  // 获取收银产品列表
		privateApi.POST("/call", wrapper.Call)                                   // 发起呼叫
		privateApi.POST("/remark", wrapper.OrderProductRemark)                   // 给商品添加备注
		privateApi.POST("/order/cart/product/add", wrapper.OrderCartProductAdd)  // 向购物车添加商品
	}
}
