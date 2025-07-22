package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MemberOrderHandler 接单相关控制器
type MemberOrderHandler struct {
	memberOrderSrv service.IMemberOrderSrv
}

// GetMemberOrderList 外送订单接单列表
// @Summary 外送订单接单列表
// @Description 外送订单接单列表
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param status query string true "状态: unaccept-待接单, accept-备餐中, undelivery-待配送, delivery-配送中, completed-已完成, cancel-已取消"
// @Success 200 {object} dto.Response{data=resp.GetMemberCashierOrderListResp}
// @Router /cashier/member_order/list [get]
func (h *MemberOrderHandler) GetMemberOrderList(c *gin.Context) {
	var memberOrderListReq req.MemberOrderListReq
	if err := c.ShouldBindQuery(&memberOrderListReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderListReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberCashierOrderList(helper.GetContext(c), memberOrderListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetMemberOrderDetail 外送订单详情
// @Summary 外送订单详情
// @Description 外送订单详情
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query int true "会员端销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderDetailResp}
// @Router /cashier/member_order/detail [get]
func (h *MemberOrderHandler) GetMemberOrderDetail(c *gin.Context) {
	var memberOrderDetailReq req.GetMemberOrderDetailReq
	if err := c.ShouldBindQuery(&memberOrderDetailReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderDetailReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberCashierOrderDetail(helper.GetContext(c), memberOrderDetailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AcceptOrder 接单
// @Summary 接单
// @Description 接单
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param AcceptOrderReq body req.AcceptOrderReq true "接受接单"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderDetailResp}
// @Router /cashier/member_order/accept [post]
func (h *MemberOrderHandler) AcceptOrder(c *gin.Context) {
	var acceptOrderReq req.AcceptOrderReq
	if err := c.ShouldBindJSON(&acceptOrderReq); err != nil {
		helper.HandleValidationError(c, err, acceptOrderReq, nil)
		return
	}
	err := h.memberOrderSrv.AcceptMemberSaleOrder(helper.GetContext(c), acceptOrderReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// RejectOrder 拒单
// @Summary 拒单
// @Description 拒单
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param RejectOrderReq body req.RejectOrderReq true "拒绝接单"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderDetailResp}
// @Router /cashier/member_order/reject [post]
func (h *MemberOrderHandler) RejectOrder(c *gin.Context) {
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

// CookFinish 备餐完成
// @Summary 备餐完成
// @Description 备餐完成
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param CookFinishOrderReq body req.CookFinishOrderReq true "备餐完成"
// @Success 200 {object} dto.Response{data=resp.GetMemberOrderDetailResp}
// @Router /cashier/member_order/cook_finish [post]
func (h *MemberOrderHandler) CookFinish(c *gin.Context) {
	var cookFinishOrderReq req.CookFinishOrderReq
	if err := c.ShouldBindJSON(&cookFinishOrderReq); err != nil {
		helper.HandleValidationError(c, err, cookFinishOrderReq, nil)
		return
	}
	err := h.memberOrderSrv.CookFinishMemberSaleOrder(helper.GetContext(c), cookFinishOrderReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetMemberOrderSearch 搜索订单列表通过关键词
// @Summary 搜索订单列表通过关键词
// @Description 搜索订单列表通过关键词
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param keyword query string true "关键字"
// @Success 200 {object} dto.Response{data=resp.GetMemberCashierOrderSearchResp}
// @Router /cashier/member_order/search [get]
func (h *MemberOrderHandler) GetMemberOrderSearch(c *gin.Context) {
	var memberOrderSearchReq req.MemberOrderSearchReq
	if err := c.ShouldBindQuery(&memberOrderSearchReq); err != nil {
		helper.HandleValidationError(c, err, memberOrderSearchReq, nil)
		return
	}
	res, err := h.memberOrderSrv.GetMemberCashierOrderSearch(helper.GetContext(c), memberOrderSearchReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetMemberOrderReturnInfo 获取外送订单退款弹窗信息
// @Summary 获取外送订单退款弹窗信息
// @Description 获取外送订单退款弹窗信息
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param member_sale_order_uuid query int true "会员端销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.MemberOrderReturnInfoResp}
// @Router /cashier/member_order/return_info [get]
func (h *MemberOrderHandler) GetMemberOrderReturnInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	reqParam := member_req.MemberOrderReturnInfoReq{}
	if err := c.ShouldBindQuery(&reqParam); err != nil {
		helper.HandleValidationError(c, err, reqParam, nil)
		return
	}
	//
	res, err := h.memberOrderSrv.GetMemberOrderReturnInfo(ctx, reqParam)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// MemberOrderReturn 外送订单退款/部分退款
// @Summary 外送订单退款/部分退款
// @Description 外送订单退款/部分退款
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.MemberOrderReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/member_order/return [post]
func (h *MemberOrderHandler) MemberOrderReturn(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	reqParam := req.OrderReturnReq{}
	if err := c.ShouldBindJSON(&reqParam); err != nil {
		helper.HandleValidationError(c, err, reqParam, nil)
		return
	}
	//
	err, codeFail := h.memberOrderSrv.MemberOrderReturn(ctx, reqParam)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// MemberOrderReReturn 外送订单重新退款
// @Summary 外送订单重新退款
// @Description 外送订单重新退款
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body member_req.MemberOrderReReturnReq true "详情参数"
// @Success 200 {object} nil "重新退款成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/member_order/re_return [post]
func (h *MemberOrderHandler) MemberOrderReReturn(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	reqParam := req.OrderReReturnReq{}
	if err := c.ShouldBindJSON(&reqParam); err != nil {
		helper.HandleValidationError(c, err, reqParam, nil)
		return
	}
	//
	err, codeFail := h.memberOrderSrv.MemberOrderReReturn(ctx, reqParam)
	if err != nil {
		helper.ErrorWithMessage(c, codeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 取消订单
// @Tags 收银端.外送接单相关
// @Accept json
// @Produce json
// @Security JwtToken
// @Param CancelOrderReq body member_req.CancelOrderReq true "取消订单请求"
// @Success 200 {object} dto.Response
// @Router /cashier/member_order/cancel [post]
func (h *MemberOrderHandler) CancelOrder(c *gin.Context) {
	var cancelOrderReq member_req.CancelOrderReq
	if err := c.ShouldBindJSON(&cancelOrderReq); err != nil {
		helper.HandleValidationError(c, err, cancelOrderReq, nil)
		return
	}
	err := h.memberOrderSrv.MemberOrderCancelInCashier(helper.GetContext(c), cancelOrderReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

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
		privateApi.GET("/member_order/list", wrapper.GetMemberOrderList)              // 获取外送订单接单列表
		privateApi.GET("/member_order/detail", wrapper.GetMemberOrderDetail)          // 获取外送订单接单详情
		privateApi.POST("/member_order/accept", wrapper.AcceptOrder)                  // 接单
		privateApi.POST("/member_order/reject", wrapper.RejectOrder)                  // 拒单
		privateApi.POST("/member_order/cook_finish", wrapper.CookFinish)              // 备餐完成
		privateApi.GET("/member_order/search", wrapper.GetMemberOrderSearch)          // 搜索订单列表通过关键词
		privateApi.GET("/member_order/return_info", wrapper.GetMemberOrderReturnInfo) // 获取外送订单退款弹窗信息
		privateApi.POST("/member_order/return", wrapper.MemberOrderReturn)            // 外送订单退款/部分退款
		privateApi.POST("/member_order/re_return", wrapper.MemberOrderReReturn)       // 外送订单重新退款
		privateApi.POST("/member_order/cancel", wrapper.CancelOrder)                  // 取消订单
	}
}
