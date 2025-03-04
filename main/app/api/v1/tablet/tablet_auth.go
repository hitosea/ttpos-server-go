package tablet

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

// AuthHandler 认证鉴权控制器
type AuthHandler struct {
	authSrv service.IAuthSrv
	deskSrv service.IDeskSrv
}

// Login 平板端登录
// @Summary 平板端登录
// @Description 平板端登录
// @Tags 平板端.认证鉴权
// @Accept json
// @Produce json
// @Param X-SIGN header string true "验证码sign"
// @param data body req.LoginReq true "登录参数"
// @Success 200 {object} dto.Response
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
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{"token": loginResp.Token})
}

// Logout 平板端退出登录
// @Summary 平板端退出登录
// @Description 平板端退出登录
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
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{}, "退出成功")
}

// GetBase 平板端信息
// @Summary 平板端信息
// @Description 平板端信息
// @Tags 平板端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.TabletBase}
// @Router /tablet/base [get]
func (h *AuthHandler) GetBase(c *gin.Context) {
	ctx := helper.GetContext(c)
	info, err := h.authSrv.TabletBase(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, info)
}

// GetDeskList 获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 平板端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.TabletDeskList}
// @Router /tablet/desk/list [get]
func (h *AuthHandler) GetDeskList(c *gin.Context) {
	ctx := helper.GetContext(c)
	list, err := h.deskSrv.GetTabletDeskList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, list)
}

// BindDesk 绑定/换绑桌台
// @Summary 绑定/换绑桌台
// @Description 绑定/换绑桌台
// @Tags 平板端.认证鉴权
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BindDeskReq true "绑定/换绑桌台请求参数"
// @Success 200 {object} dto.Response
// @Router /tablet/bind_desk [post]
func (h *AuthHandler) BindDesk(c *gin.Context) {
	var bindDeskReq req.BindDeskReq
	if err := c.ShouldBindJSON(&bindDeskReq); err != nil {
		helper.HandleValidationError(c, err, bindDeskReq, req.LoginRequestMessage)
		return
	}
	err := h.deskSrv.BindDesk(helper.GetContext(c), bindDeskReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{}, "绑定桌台成功")
}

func RegisterAuthHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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

	wrapper := &AuthHandler{
		authSrv: authSrv,
		deskSrv: deskSrv,
	}

	publicApi := router.Group("")
	{
		publicApi.POST("/login", wrapper.Login) // 登录
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/desk/list", wrapper.GetDeskList) // 获取可绑定的桌台
		privateApi.POST("/bind_desk", wrapper.BindDesk)   // 绑定桌台
		privateApi.GET("/base", wrapper.GetBase)          // 获取基本信息
		privateApi.POST("/logout", wrapper.Logout)        // 退出登录
	}
}
