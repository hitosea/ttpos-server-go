package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/member_service"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证鉴权控制器
type AuthHandler struct {
	loginSrv member_service.ILoginSrv
}

// Login 获取登陆信息
// @Summary 获取登陆信息
// @Description 获取登陆信息
// @Tags 会员端.认证
// @Accept json
// @Produce json
// @param data query member_req.MemberLoginInfoReq true "详情参数"
// @Success 200 {object} dto.Response{data=member_resp.MemberLoginInfoResp}
// @Router /member/login_info [get]
func (h *AuthHandler) LoginInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	loginReq := member_req.MemberLoginInfoReq{}
	if err := c.ShouldBindQuery(&loginReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	loginResp, err := h.loginSrv.GetLoginInfo(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}
	helper.Success(c, loginResp)
}

// SendCode 发送验证码
// @Summary 发送验证码
// @Description 发送验证码
// @Tags 会员端.认证
// @Accept json
// @Produce json
// @param data body member_req.MemberSendCodeReq true "详情参数"
// @Success 200 {object} dto.Response{}
// @Router /member/send_code [post]
func (h *AuthHandler) SendCode(c *gin.Context) {
	ctx := helper.GetContext(c)
	sendCodeReq := member_req.MemberSendCodeReq{}
	if err := c.ShouldBindJSON(&sendCodeReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.loginSrv.SendCode(ctx, sendCodeReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, nil)
}

func RegisterAuthHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	loginSrv := member_service.NewLoginSrv(dbm, cache, service.NewSMSSrv(dbm))

	wrapper := &AuthHandler{
		loginSrv: loginSrv,
	}

	publicApi := router.Group("")
	{
		publicApi.GET("/login_info", wrapper.LoginInfo)
		publicApi.POST("/send_code", wrapper.SendCode)
	}

	// // 需要认证
	// privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	// {
	// 	privateApi.GET("/refresh_token", wrapper.RefreshToken) // 刷新token
	// 	privateApi.POST("/logout", wrapper.Logout)             // 退出登录
	// }
}
