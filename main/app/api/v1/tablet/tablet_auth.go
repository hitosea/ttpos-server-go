package tablet

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

// AuthHandler 认证鉴权控制器
type AuthHandler struct {
	authSrv service.IAuthSrv
	deskSrv service.IDeskSrv
}

// Login 登录
// @Summary 登录
// @Description 登录
// @Tags 平板端.认证鉴权
// @Accept json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.LoginReq true "登录参数"
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /tablet/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := helper.GetContext(c)
	var loginReq req.LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.HandleValidationError(c, err, loginReq, req.LoginRequestMessage)
		return
	}
	loginReq.Source = constant.SourceTablet
	loginResp, err := h.authSrv.Login(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, loginResp)
}

// RefreshToken 刷新token
// @Summary 刷新token
// @Description 刷新token
// @Tags 平板端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /tablet/refresh_token [get]
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
// @Tags 平板端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /tablet/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := helper.GetContext(c)
	err := h.authSrv.Logout(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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

	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	deskSrv := service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv, mustPlanSrv)

	wrapper := &AuthHandler{
		authSrv: authSrv,
		deskSrv: deskSrv,
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
