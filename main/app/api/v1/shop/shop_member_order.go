package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"ttpos-server-go/app/dto/req/member_req"

	"github.com/gin-gonic/gin"
)

type MemberOrderHandler struct {
	memberOrderSrv service.IMemberOrderSrv
}

// GetMemberOrderList 获取外送订单管理页面，订单列表
// @Summary 获取外送订单管理页面，订单列表
// @Description 获取外送订单管理页面，订单列表
// @Tags 商家端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param status query string true "状态: unaccept-待接单, accept-备餐中, undelivery-待配送, delivery-配送中, completed-已完成, cancel-已取消"
// @Param order_no query string true "订单编号"
// @Param serial_no query string true "订单序号"
// @Param date_range query int true "日期类型 -1=全都、 0=今天、 1=昨天、 2=本周"
// @Param time_type query int true "时间类型  1=下单时间、 2=支付时间"
// @Param query_start_time query int true "查询开始时间戳"
// @Param query_end_time query int true "查询结束时间戳"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderManageListResp}
// @Router /shop/member_order/list [get]
func (h *MemberOrderHandler) GetMemberOrderList(c *gin.Context) {
	var memberOrderListReq req.MemberOrderManageListReq
	if err := c.ShouldBindQuery(&memberOrderListReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderListReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberOrderManageList(helper.GetContext(c), memberOrderListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetMemberOrderDetail 获取外送订单管理页面，订单详情
// @Summary 获取外送订单管理页面，订单详情
// @Description 获取外送订单管理页面，订单详情
// @Tags 商家端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query string true "订单UUID"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderManageDetailResp}
// @Router /shop/member_order/detail [get]
func (h *MemberOrderHandler) GetMemberOrderDetail(c *gin.Context) {
	var memberOrderDetailReq req.GetMemberOrderManageDetailReq
	if err := c.ShouldBindQuery(&memberOrderDetailReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderDetailReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberOrderManageDetail(helper.GetContext(c), memberOrderDetailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// RejectOrder 拒单
// @Summary 拒单
// @Description 拒单
// @Tags 商家端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param RejectOrderReq body req.RejectOrderReq true "拒绝接单"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderDetailResp}
// @Router /shop/member_order/reject [post]
func (h *MemberOrderHandler) RejectMemberOrder(c *gin.Context) {
	var rejectOrderReq req.RejectOrderReq
	if err := c.ShouldBindJSON(&rejectOrderReq); err != nil {
		helper.HandleValidationError(c, err, rejectOrderReq, nil)
		return
	}
	err := h.memberOrderSrv.RejectMemberSaleOrder(helper.GetContext(c), rejectOrderReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 取消订单
// @Tags 商家端.外送订单管理相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.CancelOrderReq true "详情参数"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/member_order/cancel [post]
func (h *MemberOrderHandler) CancelMemberOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := member_req.CancelOrderReq{}

	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 取消订单
	err := h.memberOrderSrv.MemberOrderCancel(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// RegisterDeliveryOrderHandlers 注册商家充值订单路由
func RegisterMemberOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	// 初始化处理器
	wrapper := MemberOrderHandler{
		memberOrderSrv: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/member_order/list", wrapper.GetMemberOrderList)
		privateApi.GET("/member_order/info", wrapper.GetMemberOrderDetail)
		privateApi.POST("/member_order/reject", wrapper.RejectMemberOrder)
		privateApi.POST("/member_order/cancel", wrapper.CancelMemberOrder)
	}
}
