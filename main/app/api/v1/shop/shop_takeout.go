package shop

import (
	"strconv"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
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
	// 外卖菜单应用服务
	takeoutMenuAppSrv application.ITakeoutMenuAppService
	// 外卖服务
	takeoutSrv service.ITakeoutSrv
	// 外卖商品服务
	productTakeoutSrv service.IProductTakeoutSrv
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
		takeoutAppSrv:     takeoutAppSrv,
		takeoutSrv:        takeoutSrv,
		productTakeoutSrv: productTakeoutSrv,
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

// GetImportProgress 获取进度
// @Summary 获取外卖菜单进度
// @Description 查询指定平台最新的菜单进度信息
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param platform query string true "外卖平台"
// @Success 200 {object} response.ImportProgressResponse "成功"
// @Router /shop/takeout/menu/import/progress [get]
func (h *TakeoutHandler) GetImportProgress(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("平台参数不能为空"))
		return
	}
	ctx := helper.GetContext(c)
	progressResp, err := h.takeoutAppSrv.GetImportProgress(ctx, platform)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取导入进度失败"))
		return
	}

	helper.Success(c, progressResp)
}

// GetImportLogs 获取日志列表
// @Summary 获取外卖菜单历史日志
// @Description 分页查询菜单的历史日志，支持按平台、类型、状态筛选
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param query body request.GetImportLogsRequest true "获取导入日志列表请求"
// @Success 200 {object} response.ImportLogListResponse "成功"
// @Router /shop/takeout/menu/import/logs [get]
func (h *TakeoutHandler) GetImportLogs(c *gin.Context) {
	var reqData request.GetImportLogsRequest
	if err := c.ShouldBindQuery(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}

	ctx := helper.GetContext(c)
	logsResp, err := h.takeoutAppSrv.GetImportLogs(ctx, reqData)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取导入日志失败"))
		return
	}

	helper.Success(c, logsResp)
}

// PushTakeoutMenu 推送菜单数据
// @Summary 推送菜单数据到外卖平台
// @Description 推送菜单数据到外卖平台
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param platform body request.PushTakeoutMenuRequest true "外卖平台"
// @Success 200 {object} nil "成功"
// @Router /shop/takeout/menu/push [post]
func (h *TakeoutHandler) PushTakeoutMenu(c *gin.Context) {
	var req request.PushTakeoutMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	platform := req.Platform
	if platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("平台参数不能为空"))
		return
	}
	ctx := helper.GetContext(c)
	err := h.takeoutSrv.PushMenuToPlatform(ctx, platform)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "推送菜单数据失败"))
		return
	}
	helper.Success(c, nil)
}

// ReimportTakeoutMenu 重新导入菜单
// @Summary 重新导入失败的菜单
// @Description 基于失败的导入日志重新导入菜单到TTPOS
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param log_uuid query uint64 true "导入日志UUID"
// @Success 200 {object} resp.GrabMenuImportResp "成功"
// @Router /shop/takeout/menu/reimport [post]
func (h *TakeoutHandler) ReimportTakeoutMenu(c *gin.Context) {
	logUuidStr := c.Query("log_uuid")
	if logUuidStr == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("日志UUID不能为空"))
		return
	}

	logUuid, err := strconv.ParseUint(logUuidStr, 10, 64)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("日志UUID格式错误"))
		return
	}

	ctx := helper.GetContext(c)
	result, err := h.takeoutSrv.ReimportMenuToTTPOS(ctx, logUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "重新导入菜单失败"))
		return
	}

	helper.Success(c, result)
}

// GetProductCount 获取外卖商品统计
// @Summary 获取外卖商品统计
// @Description 获取指定平台或所有平台的外卖商品总数
// @Tags 商家端.外卖管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param platform query string false "外卖平台(grab/lineman等,不传则统计所有平台)"
// @Success 200 {object} object{total=int64} "成功"
// @Router /shop/takeout/products/count [get]
func (h *TakeoutHandler) GetProductCount(c *gin.Context) {
	// 获取参数
	platform := c.Query("platform")

	ctx := helper.GetContext(c)
	companyUuid := ctx.GetCompanyUuid()

	// 调用Service
	total, err := h.productTakeoutSrv.GetProductCount(ctx, companyUuid, platform, true)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "查询商品统计失败"))
		return
	}

	// 返回响应
	helper.Success(c, gin.H{
		"total": total,
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
		privateApi.POST("/takeout/menu/push", takeoutHandler.PushTakeoutMenu)             // 推送菜单数据 (platform参数)
		privateApi.GET("/takeout/menu/get", takeoutHandler.GetTakeoutMenu)                // 获取菜单数据 (platform参数)
		privateApi.GET("/takeout/menu/import/progress", takeoutHandler.GetImportProgress) // 获取导入进度
		privateApi.GET("/takeout/menu/import/logs", takeoutHandler.GetImportLogs)         // 获取导入日志列表
		privateApi.POST("/takeout/menu/reimport", takeoutHandler.ReimportTakeoutMenu)     // 重新导入菜单 (log_uuid参数)

		// 外卖商品统计
		privateApi.GET("/takeout/products/count", takeoutHandler.GetProductCount) // 获取外卖商品统计

	}
}
