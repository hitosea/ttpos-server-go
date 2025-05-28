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
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BaseHandler struct {
	staffShiftSrv service.IStaffShiftSrv
	smsSrv        service.ISmsSrv
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
		staffShiftSrv: staffShiftSrv,
		smsSrv:        smsSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/shift", wrapper.SubmitShift)
		privateApi.POST("/sms/member-recharge", wrapper.SendMemberRechargeSMS)
	}
}
