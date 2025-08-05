package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证鉴权控制器
type AuthHandler struct {
	authSrv          service.IAuthSrv
	staffLoginLogSrv service.IStaffLoginLogSrv
}

// Login 登录
// @Summary 登录
// @Description 登录
// @Tags 商家端.认证
// @Accept json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.LoginReq true "登录参数"
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /shop/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := helper.GetContext(c)
	var loginReq req.LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.HandleValidationError(c, err, loginReq, req.LoginRequestMessage)
		return
	}
	loginReq.Source = constant.SourceShop
	loginResp, err := h.authSrv.Login(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}

	// 保存登录日志
	claims, _ := auth.ParseToken(loginResp.Token, config.JWT.Secret)
	h.staffLoginLogSrv.Add(ctx, claims.StaffUuid, loginReq.Username, c.ClientIP(), "登录成功")

	helper.Success(c, resp.LoginResp{
		Token:        loginResp.Token,
		RefreshToken: loginResp.RefreshToken,
	})
}

// RefreshToken 刷新token
// @Summary 刷新token
// @Description 刷新token
// @Tags 商家端.认证
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /shop/refresh_token [get]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	ctx := helper.GetContext(c)
	loginResp, err := h.authSrv.RefreshToken(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}
	helper.Success(c, resp.LoginResp{
		Token:        loginResp.Token,
		RefreshToken: loginResp.RefreshToken,
	})
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags 商家端.认证
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /shop/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := helper.GetContext(c)
	err := h.authSrv.Logout(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, gin.H{}, "退出成功")
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改密码
// @Tags 商家端.认证
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ChangePasswordReq true "修改密码参数"
// @Success 200 {object} dto.Response
// @Router /shop/change_password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	ctx := helper.GetContext(c)
	var changePasswordReq req.ChangePasswordReq
	if err := c.ShouldBindJSON(&changePasswordReq); err != nil {
		helper.HandleValidationError(c, err, changePasswordReq, req.ChangePasswordRequestMessage)
		return
	}
	err := h.authSrv.ChangePassword(ctx, changePasswordReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeSystemError, err)
		return
	}
	helper.Success(c, gin.H{}, "修改密码成功")
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
		authSrv:          authSrv,
		staffLoginLogSrv: service.NewStaffLoginLogSrv(dbm),
	}

	publicApi := router.Group("")
	{
		publicApi.POST("/login", wrapper.Login)
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 修改密码
		privateApi.POST("/change_password", wrapper.ChangePassword)
		privateApi.GET("/refresh_token", wrapper.RefreshToken) // 刷新token
		privateApi.POST("/logout", wrapper.Logout)             // 退出登录
	}
}
