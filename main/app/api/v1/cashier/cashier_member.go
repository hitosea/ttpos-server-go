package cashier

import (
	"go.uber.org/zap"
	"strconv"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MemberHandler 会员相关控制器
type MemberHandler struct {
	memberSrv        service.IMemberSrv
	rechargeOrderSrv service.IRechargeOrderSrv
}

// GetMemberLevels 会员等级列表
// @Summary 会员等级列表
// @Description 会员等级列表
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.MemberLevelList}
// @Router /cashier/member/levels [get]
func (h *MemberHandler) GetMemberLevels(c *gin.Context) {
	helper.Success(c, h.memberSrv.GetLevels(helper.GetCompanyUuid(c)))
}

// SearchMember 模糊搜索会员
// @Summary 模糊搜索会员
// @Description 模糊搜索会员
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param keyword query string false "关键字搜索：uuid\phone，前端处理前后空格"
// @Success 200 {object} dto.Response{data=resp.SearchMemberList}
// @Router /cashier/member/search [get]
func (h *MemberHandler) SearchMember(c *gin.Context) {
	helper.Success(c, h.memberSrv.SearchMember(helper.GetCompanyUuid(c), c.Query("keyword")))
}

// RechargeMember 充值会员信息
// @Summary 充值会员信息
// @Description 充值会员信息
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query integer true "uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeMember}
// @Router /cashier/member/recharge_member [get]
func (h *MemberHandler) RechargeMember(c *gin.Context) {
	uuid, err := strconv.ParseUint(c.Query("uuid"), 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeParamError, "参数错误")
	}
	info := h.memberSrv.GetRechargeMember(helper.GetCompanyUuid(c), uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
	}
	helper.Success(c, info)
}

// AddMember 添加会员
// @Summary 添加会员
// @Description 添加会员
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.AddMemberReq true "添加会员参数"
// @Success 200 {object} dto.Response
// @Router /cashier/member/add [post]
func (h *MemberHandler) AddMember(c *gin.Context) {
	ctx := helper.GetContext(c)
	var addMemberReq req.AddMemberReq
	if err := c.ShouldBindJSON(&addMemberReq); err != nil {
		helper.HandleValidationError(c, err, addMemberReq, req.AddMemberReqMessage)
		return
	}
	if err := h.memberSrv.AddMember(ctx, addMemberReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "添加会员成功")
}

// GetPendingRechargeOrder 获取进行中的充值订单
// @Summary 获取进行中的充值订单
// @Description 获取进行中的充值订单
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.RechargeOrder}
// @Router /cashier/member/recharge_order_in_progress [get]
func (h *MemberHandler) GetPendingRechargeOrder(c *gin.Context) {
	order := h.rechargeOrderSrv.GetPendingRechargeOrder(helper.GetCompanyUuid(c))
	if order.Uuid == 0 {
		helper.Success(c, gin.H{})
	} else {
		helper.Success(c, order)
	}
}

// CreateRechargeOrder 创建充值订单
// @Summary 创建充值订单
// @Description 创建充值订单
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeReq true "创建充值订单参数"
// @Success 200 {object} dto.Response{data=resp.RechargeOrder}
// @Router /cashier/member/create_recharge_order [post]
func (h *MemberHandler) CreateRechargeOrder(c *gin.Context) {
	var (
		rechargeReq req.RechargeReq
		err         error
		order       resp.RechargeOrder
	)
	if err = c.ShouldBindJSON(&rechargeReq); err != nil {
		helper.HandleValidationError(c, err, rechargeReq, req.CreateRechargeOrderReqMessage)
		return
	}
	ctx := helper.GetContext(c)
	if order, err = h.rechargeOrderSrv.CreateRechargeOrder(ctx, rechargeReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if order.Uuid == 0 {
		helper.Success(c, gin.H{})
	} else {
		helper.Success(c, order)
	}
}

// AddPaymentMethod 充值订单添加支付方式
// @Summary 充值订单添加支付方式
// @Description 充值订单添加支付方式
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderAddPaymentMethodReq true "充值订单添加支付方式参数"
// @Success 200 {object} dto.Response{data=resp.RechargeOrder}
// @Router /cashier/member/recharge_order_add_payment_method [post]
func (h *MemberHandler) AddPaymentMethod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var (
		addPaymentMethodReq req.RechargeOrderAddPaymentMethodReq
		err                 error
		order               resp.RechargeOrder
	)
	if err = c.ShouldBindJSON(&addPaymentMethodReq); err != nil {
		helper.HandleValidationError(c, err, addPaymentMethodReq, req.RechargeOrderAddPaymentMethodReqMessage)
		return
	}

	addPaymentMethodReq.CompanySetting = helper.GetCompanySetting(c)
	if order, err = h.rechargeOrderSrv.AddPaymentMethod(ctx, addPaymentMethodReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if order.Uuid == 0 {
		helper.Success(c, gin.H{})
	} else {
		helper.Success(c, order)
	}
}

// CancelPaymentMethod 充值订单撤销支付方式
// @Summary 充值订单撤销支付方式
// @Description 充值订单撤销支付方式
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RechargeOrderCancelPaymentMethodReq true "充值订单撤销支付方式参数"
// @Success 200 {object} dto.Response{data=resp.RechargeOrder}
// @Router /cashier/member/recharge_order_cancel_payment_method [post]
func (h *MemberHandler) CancelPaymentMethod(c *gin.Context) {
	ctx := helper.GetContext(c)
	var (
		cancelPaymentMethodReq req.RechargeOrderCancelPaymentMethodReq
		err                    error
		order                  resp.RechargeOrder
	)
	if err = c.ShouldBindJSON(&cancelPaymentMethodReq); err != nil {
		helper.HandleValidationError(c, err, cancelPaymentMethodReq, nil)
		return
	}
	if order, err = h.rechargeOrderSrv.CancelPaymentMethod(ctx, cancelPaymentMethodReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if order.Uuid == 0 {
		helper.Success(c, gin.H{})
	} else {
		helper.Success(c, order)
	}
}

// ConfirmRechargeOrder 确认充值订单
// @Summary 确认充值订单
// @Description 确认充值订单
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ConfirmRechargeOrder true "确认充值订单参数"
// @Success 200 {object} dto.Response{data=resp.ConfirmRechargeOrder}
// @Router /cashier/member/confirm_recharge_order [post]
func (h *MemberHandler) ConfirmRechargeOrder(c *gin.Context) {
	var confirmRechargeOrder req.ConfirmRechargeOrder
	if err := c.ShouldBindJSON(&confirmRechargeOrder); err != nil {
		helper.HandleValidationError(c, err, confirmRechargeOrder, nil)
		return
	}
	ctx := helper.GetContext(c)
	order, err := h.rechargeOrderSrv.ConfirmRechargeOrder(ctx, confirmRechargeOrder)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, order)
}

// PrintRechargeOrder 打印充值订单
// @Summary 打印充值订单
// @Description 打印充值订单
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.PrintRechargeOrderReq true "打印充值订单参数"
// @Success 200 {object} dto.Response{data=resp.PrinterLogData}
// @Router /cashier/member/print_recharge_order [post]
func (h *MemberHandler) PrintRechargeOrder(c *gin.Context) {
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
	helper.Success(c, order)
}

// GetMemberDiscount 获取会员优惠
// @Summary 获取会员优惠
// @Description 获取会员优惠
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @param member_uuid query integer true "会员Uuid"
// @Success 200 {object} dto.Response{data=resp.MemberDiscountResp}
// @Router /cashier/member/order_discount [get]
func (h *MemberHandler) GetMemberDiscount(c *gin.Context) {
	var discountReq req.GetMemberDiscountReq
	if err := c.ShouldBindQuery(&discountReq); err != nil {
		helper.HandleValidationError(c, err, discountReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	ctx.Log().Info("获取会员优惠", zap.Any("params", discountReq))
	order, err := h.memberSrv.GetMemberDiscount(ctx, discountReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, order)
}

// CheckPassword 使用会员优惠验证密码
// @Summary 使用会员优惠验证密码
// @Description 使用会员优惠验证密码
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CheckMemberPasswordReq true "使用会员优惠验证密码"
// @Success 200 {object} dto.Response{data=resp.PrinterLogData}
// @Router /cashier/member/check_password [get]
func (h *MemberHandler) CheckPassword(c *gin.Context) {
	var passwordReq req.CheckMemberPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, req.CheckMemberPasswordMessage)
		return
	}
	ctx := helper.GetContext(c)
	err := h.memberSrv.CheckMemberPassword(ctx, passwordReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "密码正确")
}

func RegisterMemberHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	printerLogSrv := service.NewPrinterLogSrv(dbm, settingSrv)
	rechargePrintSrv := service.NewRechargePrinterSrv(dbm, settingSrv, printerLogSrv)
	memberSrv := service.NewMemberSrv(dbm)
	rechargeOrderSrv := service.NewRechargeOrderSrv(dbm, cache, paymentMethodSrv, settingSrv, cashBoxSrv, rechargePrintSrv, memberSrv)
	staffOperationLogSrv := service.NewStaffOperationLogSrv(dbm, authSrv)

	wrapper := &MemberHandler{
		memberSrv:        memberSrv,
		rechargeOrderSrv: rechargeOrderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv), middleware.OperationLog(staffOperationLogSrv))
	{
		privateApi.GET("/member/levels", wrapper.GetMemberLevels)                                    // 获取会员等级列表
		privateApi.POST("/member/add", wrapper.AddMember)                                            // 添加会员
		privateApi.GET("/member/search", wrapper.SearchMember)                                       // 模糊搜索会员
		privateApi.GET("/member/recharge_member", wrapper.RechargeMember)                            // 充值会员信息
		privateApi.GET("/member/order_discount", wrapper.GetMemberDiscount)                          // 获取会员优惠
		privateApi.GET("/member/check_password", wrapper.CheckPassword)                              // 使用会员优惠验证密码
		privateApi.GET("/member/recharge_order_in_progress", wrapper.GetPendingRechargeOrder)        // 获取进行中的充值订单
		privateApi.POST("/member/create_recharge_order", wrapper.CreateRechargeOrder)                // 创建充值订单
		privateApi.POST("/member/recharge_order_add_payment_method", wrapper.AddPaymentMethod)       // 充值订单添加支付方式
		privateApi.POST("/member/recharge_order_cancel_payment_method", wrapper.CancelPaymentMethod) // 充值订单撤销支付方式
		privateApi.POST("/member/confirm_recharge_order", wrapper.ConfirmRechargeOrder)              // 确认充值订单
		privateApi.POST("/member/print_recharge_order", wrapper.PrintRechargeOrder)                  // 打印充值订单
	}
}
