package shop

import (
	"strconv"
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

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/pkg/errors"
)

type RechargeOrderHandler struct {
	rechargeOrderSrv service.IRechargeOrderSrv
}

// GetRechargeOrderList 处理获取充值订单列表
// @Summary 获取充值订单列表
// @Description 获取充值订单列表
// @Tags 商家端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param order_no query string false "订单编号"
// @Param date_type query int false "日期类型 -1=全都、 0=今天、 1=昨天、 2=本周"
// @Param enable_create_time query boolean false "启用添加时间 false-不启用，true-启用"
// @Param enable_payment_time query boolean false "启用支付时间 false-不启用，true-启用"
// @Param query_start_time query int false "查询开始时间戳"
// @Param query_end_time query int false "查询结束时间戳"
// @Param status query int false "充值订单状态, -1=全都、 0=待付款、1=已完成、2=已取消"
// @Success 200 {object} dto.Response{data=resp.RechargeOrderList}
// @Router /shop/recharge_order/list [get]
func (h *RechargeOrderHandler) GetRechargeOrderList(c *gin.Context) {
	var rechargeOrderListReq req.RechargeOrderListReq
	if err := c.ShouldBindQuery(&rechargeOrderListReq); err != nil {
		helper.HandleValidationError(c, err, rechargeOrderListReq, dto.PageReqMessage)
		return
	}
	res, err := h.rechargeOrderSrv.GetRechargeOrderList(helper.GetContext(c), rechargeOrderListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetRechargeOrderInfo 获取充值订单详情
// @Summary 获取充值订单详情
// @Description 获取充值订单详情
// @Tags 商家端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query int true "充值订单Uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeOrderInfo}
// @Router /shop/recharge_order/info [get]
func (h *RechargeOrderHandler) GetRechargeOrderInfo(c *gin.Context) {
	uuid, err := strconv.ParseUint(c.Query("uuid"), 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeParamError, "参数错误")
		return
	}
	res, err := h.rechargeOrderSrv.GetRechargeOrderInfo(helper.GetContext(c), uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetRechargeOrderRefundInfo 获取充值订单退款信息
// @Summary 获取充值订单退款信息
// @Description 获取充值订单退款信息
// @Tags 商家端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query int true "充值订单Uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeOrderRefundInfo}
// @Router /shop/recharge_order/refund [get]
func (h *RechargeOrderHandler) GetRechargeOrderRefundInfo(c *gin.Context) {
	uuid, err := strconv.ParseUint(c.Query("uuid"), 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeParamError, "参数错误")
		return
	}
	res, err := h.rechargeOrderSrv.GetRechargeOrderRefundInfo(helper.GetContext(c), uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// CancelRechargeOrder 取消充值订单
// @Summary 取消充值订单
// @Description 取消充值订单
// @Tags 商家端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderUuidReq true "取消充值订单参数"
// @Success 200 {object} dto.Response
// @Router /shop/recharge_order/cancel [post]
func (h *RechargeOrderHandler) CancelRechargeOrder(c *gin.Context) {
	var rechargeOrderUuidReq req.RechargeOrderUuidReq
	if err := c.ShouldBindJSON(&rechargeOrderUuidReq); err != nil {
		helper.HandleValidationError(c, err, rechargeOrderUuidReq, req.LoginRequestMessage)
		return
	}
	err := h.rechargeOrderSrv.CancelRechargeOrder(helper.GetContext(c), rechargeOrderUuidReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "操作成功")
}

// RechargeOrderRefund 充值订单退款
// @Summary 充值订单退款
// @Description 充值订单退款
// @Tags 商家端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderRefundReq true "退款充值订单参数"
// @Success 200 {object} dto.Response
// @Router /shop/recharge_order/refund [post]
func (h *RechargeOrderHandler) RechargeOrderRefund(c *gin.Context) {
	var orderRefundReq req.RechargeOrderRefundReq
	if err := c.ShouldBindJSON(&orderRefundReq); err != nil {
		helper.HandleValidationError(c, err, orderRefundReq, req.LoginRequestMessage)
		return
	}
	err := h.rechargeOrderSrv.RechargeOrderRefund(helper.GetContext(c), orderRefundReq)
	if err != nil {
		appErr := errors.AppError{}
		if pkgerrors.As(err, &appErr) {
			if appErr.GetCode() == constant.CodeSuccessOpenCashBox {
				helper.Success(c, gin.H{})
				return
			}
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// RegisterRechargeOrderHandlers 注册商家充值订单路由
func RegisterRechargeOrderHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	smsSrv := service.NewSMSSrv(dbm)
	rechargeOrderSrv := service.NewRechargeOrderSrv(dbm, cache, paymentMethodSrv, settingSrv, cashBoxSrv, memberSrv, smsSrv, staffShiftSrv)
	// 初始化处理器
	wrapper := RechargeOrderHandler{
		rechargeOrderSrv: rechargeOrderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/recharge_order/list", wrapper.GetRechargeOrderList)
		privateApi.GET("/recharge_order/info", wrapper.GetRechargeOrderInfo)
		privateApi.GET("/recharge_order/refund", wrapper.GetRechargeOrderRefundInfo)
		privateApi.POST("/recharge_order/cancel", wrapper.CancelRechargeOrder)
		privateApi.POST("/recharge_order/refund", wrapper.RechargeOrderRefund)
	}
}
