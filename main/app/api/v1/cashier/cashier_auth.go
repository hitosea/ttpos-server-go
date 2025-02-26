package cashier

import (
	"github.com/gin-gonic/gin"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
)

// AuthHandler 认证鉴权控制器
type AuthHandler struct {
	authSrv service.IAuthSrv
}

// Login 收银端登录
// @Summary 收银端登录
// @Description 收银端登录
// @Tags 收银端.认证
// @Accept json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.LoginReq true "登录参数"
// @Success 200 {object} dto.Response{data=resp.CashierLoginResp}
// @Router /cashier/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := helper.GetContext(c)
	var loginReq req.LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.HandleValidationError(c, err, loginReq, req.LoginRequestMessage)
		return
	}
	loginReq.Source = constant.SourceCashier
	loginResp, err := h.authSrv.Login(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}
	helper.Success(c, resp.CashierLoginResp{
		Token:        loginResp.Token,
		IsFirstLogin: loginResp.CashierIsFirstLogin,
	})
}

// Logout 收银端退出登录
// @Summary 收银端退出登录
// @Description 收银端退出登录
// @Tags 收银端.认证
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /cashier/logout [post]
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
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	wrapper := &AuthHandler{
		authSrv: authSrv,
	}

	publicApi := router.Group("")
	{
		publicApi.POST("/login", wrapper.Login)
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.POST("/logout", wrapper.Logout) // 退出登录
	}
}
