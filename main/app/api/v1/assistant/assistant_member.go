package assistant

import (
	"strconv"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
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

// MemberHandler 会员相关控制器
type MemberHandler struct {
	memberSrv        service.IMemberSrv
	rechargeOrderSrv service.IRechargeOrderSrv
}

// RechargeMember 充值会员信息
// @Summary 充值会员信息
// @Description 充值会员信息
// @Tags 点餐助手端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param uuid query integer true "uuid"
// @Success 200 {object} dto.Response{data=resp.RechargeMember}
// @Router /assistant/member/recharge_member [get]
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
// @Tags 点餐助手端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.AddMemberReq true "添加会员参数"
// @Success 200 {object} dto.Response
// @Router /assistant/member/add [post]
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

// GetMemberLevels 会员等级列表
// @Summary 会员等级列表
// @Description 会员等级列表
// @Tags 点餐助手端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.MemberLevelList}
// @Router /assistant/member/levels [get]
func (h *MemberHandler) GetMemberLevels(c *gin.Context) {
	helper.Success(c, h.memberSrv.GetLevels(helper.GetCompanyUuid(c)))
}

// GetMemberCardTypes 会员卡类型列表
// @Summary 会员卡类型列表
// @Description 会员卡类型列表
// @Tags 点餐助手端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.MemberCardTypeList}
// @Router /assistant/member/card_types [get]
func (h *MemberHandler) GetMemberCardTypes(c *gin.Context) {
	helper.Success(c, h.memberSrv.GetCardTypes(helper.GetCompanyUuid(c)))
}

// GetMemberDiscount 获取会员优惠
// @Summary 获取会员优惠
// @Description 获取会员优惠
// @Tags 点餐助手端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @param member_uuid query integer true "会员Uuid"
// @Success 200 {object} dto.Response
// @Router /assistant/member/order_discount [get]
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

// SearchMember 模糊搜索会员
// @Summary 模糊搜索会员
// @Description 模糊搜索会员
// @Tags 点餐助手端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param keyword query string false "关键字搜索"
// @Success 200 {object} dto.Response{data=resp.SearchMemberList}
// @Router /assistant/member/search [get]
func (h *MemberHandler) SearchMember(c *gin.Context) {
	helper.Success(c, h.memberSrv.SearchMember(helper.GetCompanyUuid(c), c.Query("keyword")))
}

// CheckPassword 使用会员优惠验证密码
// @Summary 使用会员优惠验证密码
// @Description 使用会员优惠验证密码
// @Tags 点餐助手端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CheckMemberPasswordReq true "使用会员优惠验证密码"
// @Success 200 {object} dto.Response
// @Router /assistant/member/check_password [get]
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
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	smsSrv := service.NewSMSSrv(dbm)
	rechargeOrderSrv := service.NewRechargeOrderSrv(dbm, cache, paymentMethodSrv, settingSrv, cashBoxSrv, memberSrv, smsSrv, staffShiftSrv)

	wrapper := &MemberHandler{
		memberSrv:        memberSrv,
		rechargeOrderSrv: rechargeOrderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/member/recharge_member", wrapper.RechargeMember)   // 充值会员信息
		privateApi.POST("/member/add", wrapper.AddMember)                   // 添加会员
		privateApi.GET("/member/levels", wrapper.GetMemberLevels)           // 获取会员等级列表
		privateApi.GET("/member/card_types", wrapper.GetMemberCardTypes)    // 获取会员卡类型列表
		privateApi.GET("/member/order_discount", wrapper.GetMemberDiscount) // 获取会员优惠
		privateApi.GET("/member/search", wrapper.SearchMember)              // 模糊搜索会员
		privateApi.GET("/member/check_password", wrapper.CheckPassword)     // 使用会员优惠验证密码
	}
}
