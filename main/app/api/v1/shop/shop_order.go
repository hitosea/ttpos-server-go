package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"errors"

	"github.com/gin-gonic/gin"
)

// OrderHandler 商家端处理程序
type OrderHandler struct {
	service     service.IOrderSrv // 订单服务
	deskService service.IDeskSrv  // 桌台服务
}

// GetShopOrderList 处理获取订单列表
// @Summary 获取订单列表
// @Description 获取订单列表
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderListReq true "列表参数"
// @Success 200 {object} resp.OrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/list [get]
func (h *OrderHandler) GetShopOrderList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	source := helper.GetSource(c)
	staff := helper.GetStaff(c)
	// 绑定请求参数
	orderListReq := req.OrderListReq{}
	if err := c.ShouldBindQuery(&orderListReq); err != nil {
		helper.HandleValidationError(c, err, orderListReq, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.service.GetOrderLists(companyUuid, staff, source, orderListReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetOrderInfo 处理获取订单详情
// @Summary 获取订单详情
// @Description 获取订单详情
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderInfoReq true "详情参数"
// @Success 200 {object} resp.OrderInfosResp "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/info [get]
func (h *OrderHandler) GetOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	orderInfoReq := req.OrderInfoReq{}
	if err := c.ShouldBindQuery(&orderInfoReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 获取收银产品列表
	res, err := h.service.GetOrderInfos(ctx, orderInfoReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CancelOrder 处理取消订单
// @Summary 取消订单
// @Description 取消订单
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	orderCancelReq := req.OrderCancelReq{}
	if err := c.ShouldBindJSON(&orderCancelReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.service.CancelOrder(ctx, orderCancelReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// DeleteOrder 处理删除订单
// @Summary 删除订单
// @Description 删除订单
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDeleteReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/delete [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	orderDeleteReq := req.OrderDeleteReq{}
	if err := c.ShouldBindJSON(&orderDeleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.service.DeleteOrder(companyUuid, orderDeleteReq.SaleBillUuid, orderDeleteReq.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// IsCellClose 判断订单是否可关闭
// @Summary 判断订单是否可关闭
// @Description 判断订单是否可关闭
// @Tags 商家端.订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderIsCellCloseReq true "详情参数"
// @Failure 404 {object} nil "未找到"
// @Router /shop/order/is_cell_close [get]
func (h *OrderHandler) IsCellClose(c *gin.Context) {
	ctx := helper.GetContext(c)
	//
	params := req.OrderIsCellCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	var err error
	if params.DeskUuid > 0 {
		_, err = h.deskService.IsCellCloseDesk(ctx, params.DeskUuid)
	} else if params.SaleBillUuid > 0 {
		_, err = h.service.IsCellCancelOrder(ctx, params.SaleBillUuid)
	} else {
		err = errors.New("参数错误")
	}
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 返回结果
	helper.Success(c, gin.H{})
}

// RegisterOrderHandlers 注册收银订单路由
func RegisterOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv, mustPlanSrv)

	// 初始化处理器
	wrapper := OrderHandler{
		service: orderSrv,
		deskService: service.NewDeskSrv( // 订单服务
			dbm,
			service.NewLocaleSrv(),
			orderSrv,
			settingSrv,
			deviceSrv,
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/order/list", wrapper.GetShopOrderList)
		privateApi.GET("/order/info", wrapper.GetOrderInfo)
		privateApi.POST("/order/cancel", wrapper.CancelOrder)
		privateApi.DELETE("/order/delete", wrapper.DeleteOrder)
		privateApi.GET("/order/is_cell_close", wrapper.IsCellClose)
	}
}
