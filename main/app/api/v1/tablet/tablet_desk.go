package tablet

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

// DeskHandler 桌台相关控制器
type DeskHandler struct {
	deskSrv  service.IDeskSrv
	orderSrv service.IOrderSrv
}

// GetDeskList 处理获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskListReq true "列表参数"
// @Success 200 {array} resp.DeskListWithPaginationResp "桌台列表"
// @Failure 404 {object} nil "未找到"
// @Router /tablet/desk/list [get]
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
		helper.ErrorWithDetail(c, constant.CodeFail, errors.ErrInternal)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// BindDesk 绑定桌台
// @Summary 绑定桌台
// @Description 绑定桌台
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BindDeskReq true "绑定桌台请求参数"
// @Success 200 {object} resp.DeskInfoResp "桌台详情"
// @Router /tablet/desk/bind [post]
func (h *DeskHandler) BindDesk(c *gin.Context) {
	var bindDeskReq req.BindDeskReq
	if err := c.ShouldBindJSON(&bindDeskReq); err != nil {
		helper.HandleValidationError(c, err, bindDeskReq, req.LoginRequestMessage)
		return
	}
	data, err := h.deskSrv.BindDesk(helper.GetContext(c), bindDeskReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, data, "绑定桌台成功")
}

// CreateDeskOrder 处理创建开台
// @Summary 开台
// @Description 开台
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "开台参数"
// @Success 200 {object} resp.CreateDeskOrderResp "开台成功"
// @Failure 404 {object} nil "未找到"
// @Router /tablet/desk/open [post]
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

// GetDeskInfo 处理获取收银台列表
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskInfoResp "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /tablet/desk/info [get]
func (h *DeskHandler) GetDeskInfo(c *gin.Context) {
	// 获取收银产品列表
	res, err := h.deskSrv.GetDeskInfo(helper.GetCompanyUuid(c), helper.GetDeskUuid(c))
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskPing 获取桌台详情-用于定时轮询
// @Summary 获取桌台详情-用于定时轮询
// @Description 获取桌台详情-用于定时轮询
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskPing "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /tablet/desk/ping [get]
func (h *DeskHandler) GetDeskPing(c *gin.Context) {
	// 获取收银产品列表
	res, err := h.deskSrv.GetDeskPing(helper.GetContext(c), helper.GetDeskUuid(c))
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetSentKitchen 获取已送厨商品
// @Summary 获取已送厨商品
// @Description 获取已送厨商品
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.GetProductListReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.SentKitchen}
// @Router /tablet/desk/order/sent_kitchen [get]
func (h *DeskHandler) GetSentKitchen(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetProductListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.GetSentKitchen(ctx, params.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// GetDeskBuffetProductList 处理获取自助餐商品列表
// @Summary 获取自助餐商品列表
// @Description 获取自助餐商品列表
// @Tags 平板端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderChangeBuffetProductListReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.BuffetProductList}
// @Failure 404 {object} nil "未找到"
// @Router /tablet/desk/order/buffet/product/list [get]
func (h *DeskHandler) GetDeskBuffetProductList(c *gin.Context) {
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

func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv)

	deskSrv := service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv)

	wrapper := &DeskHandler{
		deskSrv:  deskSrv,
		orderSrv: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/desk/list", wrapper.GetDeskList) // 获取桌台列表
		privateApi.POST("/desk/bind", wrapper.BindDesk)   // 绑定桌台

		privateApi.POST("/desk/open", wrapper.CreateDeskOrder)                              // 创建桌台订单(开桌)
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)                                   // 获取桌台详情
		privateApi.GET("/desk/ping", wrapper.GetDeskPing)                                   // 获取桌台详情-用于定时轮询
		privateApi.POST("/desk/order/cart/product/add", nil)                                // todo 向购物车添加商品并送厨
		privateApi.GET("/desk/order/sent_kitchen", wrapper.GetSentKitchen)                  // 获取已送厨商品
		privateApi.GET("/desk/order/buffet/product/list", wrapper.GetDeskBuffetProductList) // 获取自助餐商品列表
	}
}
