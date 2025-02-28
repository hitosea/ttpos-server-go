package assistant

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

// AuthHandler 认证鉴权
type AuthHandler struct {
	authSrv service.IAuthSrv
}

// Login 点餐助手登录
// @Summary 点餐助手登录
// @Description 点餐助手登录
// @Tags 点餐助手端.认证鉴权
// @Accept json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.LoginReq true "登录参数"
// @Success 200 {object} dto.Response
// @Router /assistant/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := helper.GetContext(c)
	var loginReq req.LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.HandleValidationError(c, err, loginReq, req.LoginRequestMessage)
		return
	}
	loginReq.Source = constant.SourceAssistant
	loginResp, err := h.authSrv.Login(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}
	helper.Success(c, gin.H{"token": loginResp.Token})
}

// Logout 点餐助手退出登录
// @Summary 点餐助手退出登录
// @Description 点餐助手退出登录
// @Tags 点餐助手端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /assistant/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := helper.GetContext(c)
	err := h.authSrv.Logout(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{}, "退出成功")
}

// BindCashier 点餐助手绑定收银机
// @Summary 点餐助手绑定收银机
// @Description 点餐助手绑定收银机
// @Tags 点餐助手端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BindCashierReq true "绑定收银机参数"
// @Success 200 {object} dto.Response
// @Router /assistant/bind_cashier [post]
func (h *AuthHandler) BindCashier(c *gin.Context) {
	ctx := helper.GetContext(c)
	var bindReq req.BindCashierReq
	if err := c.ShouldBindJSON(&bindReq); err != nil {
		helper.HandleValidationError(c, err, bindReq, nil)
		return
	}
	token, err := h.authSrv.BindCashier(ctx, bindReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{"token": token})
}

// GetOnlineCashiers 获取在线收银机
// @Summary 获取在线收银机
// @Description 获取在线收银机
// @Tags 点餐助手端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.OnlineCashierList}
// @Router /assistant/online_cashiers [get]
func (h *AuthHandler) GetOnlineCashiers(c *gin.Context) {
	helper.Success(c, h.authSrv.GetOnlineCashiers(helper.GetCompanyUuid(c)))
}

// GetAssistantBase 点餐助手端信息
// @Summary 点餐助手端信息
// @Description 点餐助手端信息
// @Tags 点餐助手端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.AssistantBase}
// @Router /assistant/base [get]
func (h *AuthHandler) GetAssistantBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.AssistantBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, info)
}

func RegisterAuthHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	wrapper := &AuthHandler{
		authSrv: authSrv,
	}

	publicApi := router.Group("")
	{
		publicApi.POST("/login", wrapper.Login) // 登录
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/online_cashiers", wrapper.GetOnlineCashiers) // 获取在线的收银机
		privateApi.POST("/bind_cashier", wrapper.BindCashier)         // 绑定收银机
		privateApi.GET("/base", wrapper.GetAssistantBase)             // 获取基本信息
		privateApi.POST("/logout", wrapper.Logout)                    // 退出登录
	}
}
