package cashier

import (
	"errors"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// CashierInstantHandler 收银点餐处理程序
type CashierInstantHandler struct {
	orderService service.IOrderSrv // 订单服务
}

// CreateInstantOrder 创建点餐订单
// @Summary 创建点餐订单
// @Description 创建点餐订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Success 200 {object} resp.CreateOrderResp
// @Router /cashier/instant/order/create [post]
func (h *CashierInstantHandler) CreateInstantOrder(c *gin.Context) {
	// 创建订单
	res, err := h.orderService.CreateInstantOrder(helper.GetCompanyUuid(c))
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
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消点餐订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cancel [post]
func (h *CashierInstantHandler) CancelOrder(c *gin.Context) {
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
	err := h.orderService.CancelOrder(companyUuid, staff, source, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// CloseOrder 处理关闭点餐订单
// @Summary 关闭点餐订单
// @Description 关闭点餐订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data query req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/close [post]
func (h *CashierInstantHandler) CloseOrder(c *gin.Context) {
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
	err := h.orderService.CancelOrder(companyUuid, staff, source, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductDelete 处理删除点餐订单商品
// @Summary 删除点餐订单商品
// @Description 删除点餐订单商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderProductDeleteReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/delete [delete]
func (h *CashierInstantHandler) OrderProductDelete(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	params := req.OrderProductDeleteReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderProductDelete(companyUuid, params.SaleBillUuid, params.SaleOrderUuid, params.OrderProductUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductDelete 处理删除点餐订单商品改价
// @Summary 删除点餐订单商品改价
// @Description 删除点餐订单商品改价
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderProductChangePriceReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/price [post]
func (h *CashierInstantHandler) OrderProductChangePrice(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	params := req.OrderProductChangePriceReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderProductChangePrice(companyUuid, params.SaleBillUuid, params.SaleOrderUuid, params.OrderProductUuid, params.Price)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
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
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv)

	// 创建收银产品处理程序
	wrapper := CashierInstantHandler{
		orderService: orderSrv, // 订单服务
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.POST("/instant/order/create", wrapper.CreateInstantOrder)
		privateApi.POST("/instant/order/cancel", wrapper.CancelOrder)
		privateApi.POST("/instant/order/close", wrapper.CloseOrder)
		privateApi.DELETE("/instant/order/product/delete", wrapper.OrderProductDelete)
		privateApi.POST("/instant/order/product/price", wrapper.OrderProductChangePrice)
	}
}
