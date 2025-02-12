package assistant

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
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
	var loginRequest req.LoginReq
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		helper.HandleValidationError(c, err, loginRequest, req.LoginRequestMessage)
		return
	}
	loginRequest.Source = constant.SourceAssistant
	token, err := h.authSrv.Login(loginRequest, c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, gin.H{"token": token})
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
	err := h.authSrv.Logout(c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, "退出成功")
}

// BindCashier 点餐助手绑定收银机
// @Summary 点餐助手绑定收银机
// @Description 点餐助手绑定收银机
// @Tags 点餐助手端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.LoginReq true "登录参数"
// @Success 200 {object} dto.Response
// @Router /assistant/bind_cashier [post]
func (h *AuthHandler) BindCashier(c *gin.Context) {
	var bindReq req.BindCashierReq
	if err := c.ShouldBindJSON(&bindReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	token, err := h.authSrv.BindCashier(bindReq, c.Copy())
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
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
// @Success 200 {object} dto.Response{list=[]resp.OnlineCashier}
// @Router /assistant/online_cashiers [get]
func (h *AuthHandler) GetOnlineCashiers(c *gin.Context) {
	helper.Success(c, gin.H{"list": h.authSrv.GetOnlineCashiers(c.GetUint64(jwt.CompanyUuid))})
}

// GetAssistantBase 点餐助手端信息
// @Summary 点餐助手端信息
// @Description 点餐助手端信息
// @Tags 点餐助手端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.AssistantBase}
// @Router /cashier/base [get]
func (h *AuthHandler) GetAssistantBase(c *gin.Context) {
	info, err := h.authSrv.AssistantBase(c.Copy())
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
		publicApi.POST("/login", wrapper.Login) // 登录
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/online_cashiers", wrapper.GetOnlineCashiers) // 获取在线的收银机
		privateApi.POST("/bind_cashier", wrapper.BindCashier)         // 绑定收银机
		privateApi.GET("/base", wrapper.GetAssistantBase)             // 获取基本信息
		privateApi.POST("/logout", wrapper.Logout)                    // 获取基本信息
	}
}
