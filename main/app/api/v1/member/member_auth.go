package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/member_service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
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
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
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
	helper.Success(c, gin.H{
		"uuid": sendCodeReq.Phone,
	})
}

// Login 登录
// @Summary 登录
// @Description 登录
// @Tags 会员端.认证
// @Accept json
// @Produce json
// @param data body member_req.MemberLoginReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /member/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := helper.GetContext(c)
	loginReq := member_req.MemberLoginReq{}
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	loginResp, err := h.loginSrv.Login(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}
	helper.Success(c, loginResp)
}

// Login 游客登录
// @Summary 游客登录
// @Description 游客登录，如果游客不存在则自动创建
// @Tags 会员端.游客
// @Accept json
// @Produce json
// @param data body req.VisitorLoginReq true "游客登录参数"
// @Success 200 {object} dto.Response{data=resp.VisitorInfoResp}
// @Router /member/visitor/login [post]
func (h *AuthHandler) VisitorLogin(c *gin.Context) {
	ctx := helper.GetContext(c)
	loginReq := req.VisitorLoginReq{}
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	if err := loginReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	visitorInfo, err := h.loginSrv.VisitorLogin(ctx, loginReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, visitorInfo)
}

// Register 注册
// @Summary 注册
// @Description 注册
// @Tags 会员端.认证
// @Accept json
// @Produce json
// @param data body member_req.MemberRegisterReq true "详情参数"
// @Success 200 {object} dto.Response{data=member_resp.LoginResp}
// @Router /member/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := helper.GetContext(c)
	registerReq := member_req.MemberRegisterReq{}
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	registerResp, err := h.loginSrv.Register(ctx, registerReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, registerResp)
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

	// 初始化处理器
	wrapper := &AuthHandler{
		loginSrv: member_service.NewLoginSrv(
			dbm, cache,
			service.NewSMSSrv(dbm),
			setting.NewSrvImpl(dbm, cache),
		),
	}

	publicApi := router.Group("")
	{
		publicApi.GET("/login_info", wrapper.LoginInfo)
		publicApi.POST("/send_code", wrapper.SendCode)
		publicApi.POST("/login", wrapper.Login)
		publicApi.POST("/visitor/login", wrapper.VisitorLogin)
	}

	// 需要认证
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.POST("/register", wrapper.Register)
	}
}
