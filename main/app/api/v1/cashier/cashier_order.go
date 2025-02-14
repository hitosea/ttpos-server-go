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

	"errors"

	"github.com/gin-gonic/gin"
)

// CashierOrderHandler 收银点餐处理程序
type CashierOrderHandler struct {
	service     service.IOrderSrv // 订单服务
	deskService service.IDeskSrv  // 桌台服务
}

// GetCashierOrderList 处理获取订单列表
// @Summary 获取订单列表
// @Description 获取订单列表
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @param data query req.OrderListReq true "列表参数"
// @Success 200 {object} resp.OrderListPaginationResp "订单列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/list [get]
func (h *CashierOrderHandler) GetCashierOrderList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	source := helper.GetSource(c)
	staff := helper.GetStaff(c)
	// 绑定请求参数
	req := req.OrderListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.service.GetOrderLists(companyUuid, staff, source, req)
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
// @Success 200 {object} resp.OrderInfosResp "订单详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/info [get]
func (h *CashierOrderHandler) GetOrderInfo(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	source := helper.GetSource(c)
	staff := helper.GetStaff(c)
	// 绑定请求参数
	req := req.OrderInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 获取收银产品列表
	res, err := h.service.GetOrderInfos(companyUuid, staff, source, req)
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
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/cancel [post]
func (h *CashierOrderHandler) CancelOrder(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	source := helper.GetSource(c)
	staff := helper.GetStaff(c)
	// 绑定请求参数
	req := req.OrderCancelReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.service.CancelOrder(companyUuid, staff, source, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// CancelOrder 处理删除订单
// @Summary 删除订单
// @Description 删除订单
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @param data body req.OrderDeleteReq true "详情参数"
// @Success 200 {object} nil "取消订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/delete [delete]
func (h *CashierOrderHandler) DeleteOrder(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	req := req.OrderDeleteReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.service.DeleteOrder(companyUuid, req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// IsCellCloseDesk 判断订单是否可关闭
// @Summary 判断订单是否可关闭
// @Description 判断订单是否可关闭
// @Tags 收银端.订单
// @Accept json
// @Produce json
// @param data query req.OrderIsCellCloseReq true "详情参数"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/order/is_cell_close [get]
func (h *CashierOrderHandler) IsCellClose(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	//
	params := req.OrderIsCellCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	var err error
	if params.DeskUuid > 0 {
		_, err = h.deskService.IsCellCloseDesk(companyUuid, params.DeskUuid)
	} else if params.SaleBillUuid > 0 {
		_, err = h.service.IsCellCancelOrder(companyUuid, params.SaleBillUuid)
	} else {
		err = errors.New("参数错误")
	}
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// todo 获取已经送厨的商品 - 等王总写完拿来用

	// 返回结果
	helper.Success(c, gin.H{})
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
	orderSrv := service.NewOrderSrv(dbm, service.NewLocaleSrv(), settingSrv)

	// 初始化处理器
	wrapper := CashierOrderHandler{
		service: orderSrv,
		deskService: service.NewDeskSrv( // 订单服务
			dbm,
			service.NewLocaleSrv(),
			orderSrv,
			settingSrv,
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/order/list", wrapper.GetCashierOrderList)
		privateApi.GET("/order/info", wrapper.GetOrderInfo)
		privateApi.POST("/order/cancel", wrapper.CancelOrder)
		privateApi.DELETE("/order/delete", wrapper.DeleteOrder)
		privateApi.GET("/order/is_cell_close", wrapper.IsCellClose)
	}
}
