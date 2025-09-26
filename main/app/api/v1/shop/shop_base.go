package shop

import (
	"strconv"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BaseHandler struct {
	authSrv       service.IAuthSrv
	staffShiftSrv service.IStaffShiftSrv
	smsSrv        service.ISmsSrv
	settingSrv    setting.ISrv
}

// SubmitShift 提交交班
// @Summary 提交交班
// @Description 提交交班
// @Tags 商家端.交班
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SubmitShiftReq true "提交交班参数"
// @Success 200 {object} dto.Response{data=resp.ShiftSubmit}
// @Router /shop/shift [post]
func (h *BaseHandler) SubmitShift(c *gin.Context) {
	ctx := helper.GetContext(c)
	var submitReq req.SubmitShiftReq
	if err := c.ShouldBindJSON(&submitReq); err != nil {
		helper.HandleValidationError(c, err, submitReq, nil)
		return
	}

	info, err := h.staffShiftSrv.SubmitShift(ctx, submitReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info, "交班成功")
}

// GetRemainingSmsQuota 获取商家剩余短信额度
func (h *BaseHandler) GetRemainingSmsQuota(c *gin.Context) {
	cid := c.Query("cid")
	if cid == "" {
		helper.Fail(c, constant.CodeFail, "店铺ID参数错误")
		return
	}
	companyUuid, err := strconv.ParseUint(cid, 10, 64)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "店铺ID参数值错误")
		return
	}
	ctx := context.NewContext(
		context.WithCompanyUuid(companyUuid),
	)
	companySetting, err := h.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "获取店铺信息错误")
		return
	}
	helper.Success(c, gin.H{"remaining_sms_quota": companySetting.SmsQuota}, "ok")
}

// SendMemberRechargeSMS 发送会员充值短信
// @Summary 发送会员充值短信
// @Description 发送会员充值短信
// @Tags 商家端.发送会员充值短信
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SendMemberRechargeSMS true "提交发送会员充值短信参数"
// @Success 200 {object} dto.Response
// @Router /shop/sms/member-recharge [post]
func (h *BaseHandler) SendMemberRechargeSMS(c *gin.Context) {
	ctx := helper.GetContext(c)
	var sendReq req.SendMemberRechargeSMS
	if err := c.ShouldBindJSON(&sendReq); err != nil {
		helper.HandleValidationError(c, err, sendReq, nil)
		return
	}
	go func() {
		err := h.smsSrv.SendMemberRechargeSMS(ctx, sendReq.Phone, &sms.MemberRechargeRequest{
			Company:       sendReq.Company,
			Recharge:      sendReq.Recharge,
			BonusMoney:    sendReq.BonusMoney,
			BonusPoints:   sendReq.BonusPoints,
			Balance:       sendReq.Balance,
			PointsBalance: sendReq.PointsBalance,
		})
		if err != nil {
			logger.Logger.Error("SHOP_SendMemberRechargeSMS error", zap.Any("error", err))
		}
	}()
	helper.Success(c, nil, "发送成功")
}

// GetBase 基本信息
// @Summary 基本信息
// @Description 基本信息
// @Tags 商家端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.ShopBase}
// @Router /shop/base [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.ShopBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// CheckUpdate 检查更新
// @Summary 检查更新
// @Description 检查更新
// @Tags 商家端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param brand query string true "品牌参数"
// @Success 200 {object} dto.Response
// @Router /shop/check_update [get]
func (h *BaseHandler) CheckUpdate(c *gin.Context) {
	ctx := helper.GetContext(c)
	updateInfo, err := h.settingSrv.CheckUpdate(ctx, constant.AppTypeShop, c.Query("brand"), i18n.GetAcceptLanguage(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, updateInfo)
}

// RegisterOrderHandlers 注册商家订单路由
func RegisterBaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	smsSrv := service.NewSMSSrvImpl(dbm)

	// 初始化处理器
	wrapper := BaseHandler{
		authSrv:       authSrv,
		staffShiftSrv: staffShiftSrv,
		smsSrv:        smsSrv,
		settingSrv:    settingSrv,
	}

	// 不需要认证，获取商家剩余短信额度
	publicApi := router.Group("")
	{
		publicApi.GET("remaining_sms_quota", wrapper.GetRemainingSmsQuota)
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/shift", wrapper.SubmitShift)
		privateApi.POST("/sms/member-recharge", wrapper.SendMemberRechargeSMS)

		privateApi.GET("/base", wrapper.GetBase)             // 获取基础信息
		privateApi.GET("/check_update", wrapper.CheckUpdate) // 检查更新
	}
}
