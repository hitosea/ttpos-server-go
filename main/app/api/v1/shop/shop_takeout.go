package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/types/request"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// TakeoutHandler 外卖处理程序
type TakeoutHandler struct {
	// 外卖应用服务（统一服务）
	takeoutAppSrv application.ITakeoutAppService
	// 外卖服务
	takeoutSrv service.ITakeoutSrv
}

// NewTakeoutHandler 创建外卖 Handler
func NewTakeoutHandler(
	dbm *database.DBManager,
	cache cache.Cache,
	productSrv service.IProductSrv,
	productTakeoutSrv service.IProductTakeoutSrv,
	translateSrv service.ITranslateSrv,
	settingSrv setting.ISrv,
	takeoutAppSrv application.ITakeoutAppService,
) *TakeoutHandler {
	takeoutSrv := service.NewTakeoutSrv(dbm, cache, productSrv, productTakeoutSrv, translateSrv, settingSrv)
	return &TakeoutHandler{
		takeoutAppSrv: takeoutAppSrv,
		takeoutSrv:    takeoutSrv,
	}
}

// NewTakeoutStatusHandler 创建外卖状态 Handler（向后兼容）
func NewTakeoutStatusHandler(dbm *database.DBManager, cache cache.Cache) *TakeoutHandler {
	return &TakeoutHandler{
		takeoutAppSrv: nil, // 暂时设为nil，后面完善
		takeoutSrv:    nil,
	}
}

// GetTakeoutStatus 获取指定平台外卖状态
// @Summary 获取外卖平台状态
// @Description 获取指定外卖平台的状态信息（是否开启、菜单数据等）
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param platform query string true "外卖平台(grab/lineman等)"
// @Success 200 {object} response.TakeoutStatusResponse "成功"
// @Router /shop/takeout/status/info [get]
func (h *TakeoutHandler) GetTakeoutStatus(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("平台参数不能为空"))
		return
	}

	ctx := helper.GetContext(c)

	status, err := h.takeoutAppSrv.GetTakeoutStatus(ctx, platform)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, status)
}

// ToggleTakeoutStatus 切换指定平台外卖状态
// @Summary 切换外卖平台状态
// @Description 开启或关闭指定外卖平台的功能
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body request.ToggleTakeoutStatusRequest true "切换状态请求"
// @Success 200 {object} nil "成功"
// @Router /shop/takeout/status/toggle [post]
func (h *TakeoutHandler) ToggleTakeoutStatus(c *gin.Context) {
	var req request.ToggleTakeoutStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	ctx := helper.GetContext(c)

	status, err := h.takeoutAppSrv.ToggleTakeoutStatus(ctx, req.Platform, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, status)
}

// GetBindingLink 获取绑定链接
// @Summary 获取外卖平台绑定链接
// @Description 获取指定外卖平台的绑定链接，用户可跳转到平台页面完成配置
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param platform query string true "外卖平台(grab/lineman等)"
// @Success 200 {object} nil "成功"
// @Router /shop/takeout/binding-link [get]
func (h *TakeoutHandler) GetBindingLink(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("平台参数不能为空"))
		return
	}

	ctx := helper.GetContext(c)

	// 调用应用服务
	result, err := h.takeoutAppSrv.GetBindingLink(ctx, platform)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取绑定链接失败"))
		return
	}

	helper.Success(c, result)
}

// CheckBindingStatus 检查绑定状态
// @Summary 检查外卖平台绑定状态
// @Description 验证是否已经绑定指定的外卖平台，前端可以定时查询
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param platform query int true "外卖平台UUID"
// @Success 200 {object} nil "成功"
// @Router /shop/takeout/binding-status [get]
func (h *TakeoutHandler) CheckBindingStatus(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("平台参数不能为空"))
		return
	}

	ctx := helper.GetContext(c)

	// 调用应用服务
	result, err := h.takeoutAppSrv.CheckBindingStatus(ctx, platform)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "检查绑定状态失败"))
		return
	}

	helper.Success(c, result)
}

// GetTakeoutMenu 获取外卖菜单
// @Summary 获取外卖平台商品菜单
// @Description 从指定外卖平台API获取商品菜单数据
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param platform query string true "外卖平台(grab/lineman等)"
// @Success 200 {object} nil "成功"
// @Router /shop/takeout/menu/get [get]
func (h *TakeoutHandler) GetTakeoutMenu(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("平台参数不能为空"))
		return
	}

	ctx := helper.GetContext(c)

	// 这里应该根据平台类型调用不同的服务
	// 暂时只支持Grab，后续可以扩展
	result, err := h.takeoutAppSrv.GetGrabMenu(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取外卖菜单失败"))
		return
	}

	helper.Success(c, result)
}

// ImportMenu 导入外卖菜单
// @Summary 导入外卖菜单
// @Description 按规则导入外卖菜单（分类/商品/规格/属性/单位映射或创建）
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param body body req.TakeoutMenuImportReq true "外卖菜单 JSON"
// @Success 200 {object} nil "成功"
// @Router /shop/takeout/menu/import [post]
func (h *TakeoutHandler) ImportMenu(c *gin.Context) {
	var importReq req.TakeoutMenuImportReq
	if err := c.ShouldBindJSON(&importReq); err != nil {
		helper.HandleValidationError(c, err, importReq, nil)
		return
	}

	ctx := helper.GetContext(c)
	result, err := h.takeoutSrv.ImportMenu(ctx, importReq.Platform, importReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "导入外卖菜单失败"))
		return
	}

	helper.Success(c, resp.GrabMenuImportResp{
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
		CreatedItems: result.CreatedItems,
		UpdatedItems: result.UpdatedItems,
		Failures:     result.Failures,
	})
}

// RegisterTakeoutHandlers 注册外卖路由
func RegisterTakeoutHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	productSrv := service.NewProductSrv(dbm, service.NewLocaleSrv(), settingSrv, cache, service.NewTranslateSrv(dbm, cache))
	translateSrv := service.NewTranslateSrv(dbm, cache)
	takeoutAppSrv := application.NewTakeoutAppService(dbm)
	productTakeoutSrv := service.NewProductTakeoutSrv(dbm, service.NewLocaleSrv(), settingSrv, cache, translateSrv)
	takeoutHandler := NewTakeoutHandler(dbm, cache, productSrv, productTakeoutSrv, translateSrv, settingSrv, takeoutAppSrv)

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 外卖状态管理路由
		privateApi.GET("/takeout/status/info", takeoutHandler.GetTakeoutStatus)       // 根据平台获取状态
		privateApi.POST("/takeout/status/toggle", takeoutHandler.ToggleTakeoutStatus) // 切换指定平台状态

		// 外卖绑定相关路由
		privateApi.GET("/takeout/binding-link", takeoutHandler.GetBindingLink)       // 获取绑定链接 (platform参数)
		privateApi.GET("/takeout/binding-status", takeoutHandler.CheckBindingStatus) // 检查绑定状态 (platform参数)

		// 外卖菜单相关路由
		privateApi.GET("/takeout/menu/get", takeoutHandler.GetTakeoutMenu) // 获取菜单数据 (platform参数)
		privateApi.POST("/takeout/menu/import", takeoutHandler.ImportMenu) // 导入菜单数据 (platform参数)

	}
}
