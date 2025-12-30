package cashier

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

// RechargeOrderHandler 充值订单处理程序
type RechargeOrderHandler struct {
	rechargeOrderSrv service.IRechargeOrderSrv
}

// GetRechargeOrderList 处理获取充值订单列表
// @Summary 获取充值订单列表
// @Description 获取充值订单列表
// @Tags 收银端.充值订单
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
// @Router /cashier/recharge_order/list [get]
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
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query int true "充值订单Uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeOrderInfo}
// @Router /cashier/recharge_order/info [get]
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

// CancelRechargeOrder 取消充值订单
// @Summary 取消充值订单
// @Description 取消充值订单
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderUuidReq true "取消充值订单参数"
// @Success 200 {object} dto.Response
// @Router /cashier/recharge_order/cancel [post]
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

// Print 充值订单-打印小票
// @Summary 充值订单-打印小票
// @Description 充值订单-打印小票
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.PrintRechargeOrderReq true "充值订单打印小票参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData}
// @Router /cashier/recharge_order/print [post]
func (h *RechargeOrderHandler) PrintTicket(c *gin.Context) {
	var printRechargeOrderReq req.PrintRechargeOrderReq
	if err := c.ShouldBindJSON(&printRechargeOrderReq); err != nil {
		helper.HandleValidationError(c, err, printRechargeOrderReq, nil)
		return
	}
	order, err := h.rechargeOrderSrv.PrintTicket(helper.GetContext(c), printRechargeOrderReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, order, "发送成功")
}

// GetRechargeOrderRefundInfo 获取充值订单退款信息
// @Summary 获取充值订单退款信息
// @Description 获取充值订单退款信息
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query int true "充值订单Uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeOrderRefundInfo}
// @Router /cashier/recharge_order/refund [get]
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

// CheckRechargeOrderReverseSettle 检查充值订单反结账信息
// @Summary 检查充值订单反结账信息
// @Description 检查充值订单反结账信息
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query int true "充值订单Uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeOrderReverseSettleInfo}
// @Router /cashier/recharge_order/check_reverse_settle [get]
func (h *RechargeOrderHandler) CheckRechargeOrderReverseSettle(c *gin.Context) {
	uuid, err := strconv.ParseUint(c.Query("uuid"), 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeParamError, "参数错误")
		return
	}
	res, err := h.rechargeOrderSrv.CheckRechargeOrderReverseSettle(helper.GetContext(c), uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// RechargeOrderReverseSettle 反结账充值订单
// @Summary 反结账充值订单
// @Description 反结账充值订单
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderUuidReq true "反结账充值订单参数"
// @Success 200 {object} dto.Response
// @Router /cashier/recharge_order/reverse_settle [post]
func (h *RechargeOrderHandler) RechargeOrderReverseSettle(c *gin.Context) {
	var rechargeOrderUuidReq req.RechargeOrderUuidReq
	if err := c.ShouldBindJSON(&rechargeOrderUuidReq); err != nil {
		helper.HandleValidationError(c, err, rechargeOrderUuidReq, req.LoginRequestMessage)
		return
	}
	err := h.rechargeOrderSrv.RechargeOrderReverseSettle(helper.GetContext(c), rechargeOrderUuidReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// RechargeOrderRefund 充值订单退款
// @Summary 充值订单退款
// @Description 充值订单退款
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderRefundReq true "退款充值订单参数"
// @Success 200 {object} dto.Response
// @Router /cashier/recharge_order/refund [post]
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
				helper.Success(c, gin.H{"is_open_cash_box": true})
				return
			}
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// ReturnOrder 处理退款订单
// @Summary 重新退款
// @Description 重新退款
// @Tags 收银端.充值订单
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderReReturnReq true "详情参数"
// @Success 200 {object} nil "退款订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/recharge_order/re_return [post]
func (h *RechargeOrderHandler) ReReturnOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.RechargeOrderReReturnReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	//
	err := h.rechargeOrderSrv.RechargeOrderReReturnOrder(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// RegisterRechargeOrderHandlers 注册收银充值订单路由
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
		privateApi.POST("/recharge_order/cancel", wrapper.CancelRechargeOrder)
		privateApi.POST("/recharge_order/print", wrapper.PrintTicket)
		privateApi.GET("/recharge_order/refund", wrapper.GetRechargeOrderRefundInfo)
		privateApi.POST("/recharge_order/refund", wrapper.RechargeOrderRefund)
		privateApi.POST("/recharge_order/re_return", wrapper.ReReturnOrder)
		privateApi.GET("/recharge_order/check_reverse_settle", wrapper.CheckRechargeOrderReverseSettle)
		privateApi.POST("/recharge_order/reverse_settle", wrapper.RechargeOrderReverseSettle)
	}
}
