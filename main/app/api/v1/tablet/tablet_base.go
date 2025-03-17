package tablet

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础相关控制器
type BaseHandler struct {
	authSrv    service.IAuthSrv
	settingSrv setting.ISrv
	deskSrv    service.IDeskSrv
}

// VerifyAdvancedPassword 验证高级密码
// @Summary 验证高级密码
// @Description 验证高级密码
// @Tags 平板端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.VerifyPasswordReq true "验证密码参数"
// @Success 200 {object} dto.Response
// @Router /tablet/verify_advanced_password [post]
func (h *BaseHandler) VerifyAdvancedPassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var passwordReq req.VerifyPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, nil)
		return
	}
	verified := h.settingSrv.VerifyPassword(ctx, constant.SourceTablet, constant.PasswordTypeAdvanced, passwordReq.Password)
	if verified {
		helper.Success(c, gin.H{}, "验证成功")
	} else {
		helper.Fail(c, constant.CodeFail, "验证失败")
	}
}

// CheckUpdate 检查更新
// @Summary 检查更新
// @Description 检查更新
// @Tags 平板端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @param brand query string true "品牌参数"
// @Success 200 {object} dto.Response
// @Router /tablet/check_update [get]
func (h *BaseHandler) CheckUpdate(c *gin.Context) {
	ctx := helper.GetContext(c)
	updateInfo, err := h.settingSrv.CheckUpdate(ctx, constant.AppTypeTablet, c.Query("brand"), i18n.GetAcceptLanguage(c))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, updateInfo)
}

// EditSetting 修改设置
// @Summary 修改设置
// @Description 修改设置
// @Tags 平板端.设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.EditSettingReq true "修改设置参数"
// @Success 200 {object} dto.Response
// @Router /tablet/setting [post]
func (h *BaseHandler) EditSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var settingReq req.EditSettingReq
	if err := c.ShouldBindJSON(&settingReq); err != nil {
		helper.HandleValidationError(c, err, settingReq, req.UpdateAcceptOrderSettingMessage)
		return
	}
	err := h.deskSrv.ChangeBindDesk(ctx, settingReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetBase 基本信息
// @Summary 基本信息
// @Description 基本信息
// @Tags 平板端.基础信息
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.TabletBase}
// @Router /tablet/base [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.TabletBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, info)
}

func RegisterBaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv)

	deskSrv := service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv)

	wrapper := &BaseHandler{
		authSrv:    authSrv,
		settingSrv: settingSrv,
		deskSrv:    deskSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/base", wrapper.GetBase)                                     // 获取基本信息
		privateApi.POST("/verify_advanced_password", wrapper.VerifyAdvancedPassword) // 验证高级密码
		privateApi.GET("/check_update", wrapper.CheckUpdate)                         // 检查更新
		privateApi.POST("/setting", wrapper.EditSetting)                             // 修改设置(修改机器备注和换绑桌台)
	}
}
