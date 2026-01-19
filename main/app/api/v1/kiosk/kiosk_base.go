package kiosk

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

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础相关控制器
type BaseHandler struct {
	authSrv    service.IAuthSrv
	settingSrv setting.ISrv
	deviceSrv  service.IDeviceSrv
}

// GetBase 基本信息
// @Summary 基本信息
// @Description 获取自助点餐机首页基本信息
// @Tags 自助点餐机.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.KioskBase}
// @Router /kiosk/base [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.KioskBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

// VerifyAdvancedPassword 验证高级密码
// @Summary 验证高级密码
// @Description 验证高级密码
// @Tags 自助点餐机.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证密码参数"
// @Success 200 {object} dto.Response
// @Router /kiosk/verify_advanced_password [post]
func (h *BaseHandler) VerifyAdvancedPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceKiosk, constant.PasswordTypeAdvanced, passwordReq.Password)
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// SaveSetting 保存设置
// @Summary 保存设置
// @Description 保存设置，目前仅支持修改机器备注
// @Tags 自助点餐机.设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.EditDeviceRemarkReq true "修改设置参数"
// @Success 200 {object} dto.Response
// @Router /kiosk/setting [post]
func (h *BaseHandler) SaveSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var settingReq req.EditDeviceRemarkReq
	if err := c.ShouldBindJSON(&settingReq); err != nil {
		helper.HandleValidationError(c, err, settingReq, nil)
		return
	}
	err := h.deviceSrv.UpdateRemark(ctx, settingReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

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

	wrapper := &BaseHandler{
		authSrv:    authSrv,
		settingSrv: settingSrv,
		deviceSrv:  deviceSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.GetBase)                                     // 获取基本信息
		privateApi.POST("/verify_advanced_password", wrapper.VerifyAdvancedPassword) // 验证高级密码
		privateApi.POST("/setting", wrapper.SaveSetting)                             // 保存设置（机器备注）
	}
}
