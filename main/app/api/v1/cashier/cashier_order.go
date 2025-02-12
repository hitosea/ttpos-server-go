package cashier

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

	"github.com/gin-gonic/gin"
)

// CashierOrderHandler 收银点餐处理程序
type CashierOrderHandler struct {
	orderService service.IOrderSrv // 订单服务
}

// GetCashierOrderList 处理获取收银订单列表
// @Summary 获取收银订单列表
// @Description 获取收银订单列表
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @param data query req.OrderListReq true "列表参数"
// @Success 200 {object} resp.CashierOrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/list [get]
func (h *CashierOrderHandler) GetCashierOrderList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	req := req.OrderListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.orderService.GetCashierOrderList(companyUuid, req)
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
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @param data query req.OrderInfoReq true "详情参数"
// @Success 200 {object} resp.CashierOrderInfoResp "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/info [get]
func (h *CashierOrderHandler) GetOrderInfo(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	req := req.OrderInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 获取收银产品列表
	res, err := h.orderService.GetCashierOrderInfo(companyUuid, req)
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
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @param data body req.OrderInfoReq true "详情参数"
// @Success 200 {object} resp.CashierOrderInfoResp "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/cancel [post]
func (h *CashierOrderHandler) CancelOrder(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	source := helper.GetSource(c)
	staff := helper.GetStaff(c)
	// 绑定请求参数
	req := req.OrderInfoReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.orderService.CancelOrder(companyUuid, staff, source, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, nil)
}

// RegisterOrderHandlers 注册收银订单路由
func RegisterOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	// 初始化处理器
	wrapper := CashierOrderHandler{
		orderService: service.NewOrderSrv(dbm, service.NewLocaleSrv(), cache), // 订单服务
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/order/list", wrapper.GetCashierOrderList)
		privateApi.GET("/order/info", wrapper.GetOrderInfo)
		privateApi.POST("/order/cancel", wrapper.CancelOrder)
	}
}
