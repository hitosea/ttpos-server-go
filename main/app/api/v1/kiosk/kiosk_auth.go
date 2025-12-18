package kiosk

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证鉴权控制器
type AuthHandler struct {
	authSrv service.IAuthSrv
}

// Login 登录
// @Summary 登录
// @Description 登录
// @Tags 自助点餐机.认证
// @Accept json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.LoginReq true "登录参数"
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /kiosk/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := helper.GetContext(c)
	var loginReq req.LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.HandleValidationError(c, err, loginReq, req.LoginRequestMessage)
		return
	}
	loginReq.Source = constant.SourceKiosk

	var loginResp resp.LoginResp
	var err error

	// 版本判断：如果版本号 >= 2.11.0，使用统一认证登录
	if ctx.Version(context.GTE, constant.ClientVersionV2110) {
		loginResp, err = h.authSrv.SaasLogin(ctx, loginReq)
	} else {
		// 旧版本兼容，使用原有登录方法
		loginResp, err = h.authSrv.Login(ctx, loginReq)
	}

	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, errors.WithMessage(err))
		return
	}
	helper.Success(c, loginResp)
}

// RefreshToken 刷新token
// @Summary 刷新token
// @Description 刷新token
// @Tags 自助点餐机.认证
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /kiosk/refresh_token [get]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	ctx := helper.GetContext(c)
	loginResp, err := h.authSrv.RefreshToken(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}
	helper.Success(c, loginResp)
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags 自助点餐机.认证
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /kiosk/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := helper.GetContext(c)
	err := h.authSrv.Logout(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, gin.H{}, "退出成功")
}

func RegisterAuthHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &AuthHandler{
		authSrv: authSrv,
	}

	publicApi := router.Group("")
	{
		publicApi.POST("/login", wrapper.Login) // 登录
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/refresh_token", wrapper.RefreshToken) // 刷新token
		privateApi.POST("/logout", wrapper.Logout)             // 退出登录
	}
}
