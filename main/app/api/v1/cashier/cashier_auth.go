package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 结构体
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
// @Success 200 {object} dto.Response
// @Router /cashier/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest req.LoginReq
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		helper.HandleValidationError(c, err, loginRequest, req.LoginRequestMessage)
		return
	}
	loginRequest.Source = constant.SourceCashier
	token, err := h.authSrv.Login(loginRequest, c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, gin.H{"token": token})
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
	err := h.authSrv.Logout(c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, "退出成功")
}

// GetCashierBase 收银端信息
// @Summary 收银端信息
// @Description 收银端信息
// @Tags 收银端.认证
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.CashierBase}
// @Router /cashier/base [get]
func (h *AuthHandler) GetCashierBase(c *gin.Context) {
	info, err := h.authSrv.CashierBase(c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, info)
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
		privateApi.GET("/base", wrapper.GetCashierBase) // 获取基本信息
		privateApi.POST("/logout", wrapper.Logout)      // 退出登录
	}
}
