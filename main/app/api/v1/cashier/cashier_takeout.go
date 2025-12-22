package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	appService "ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// TakeoutHandler 外卖处理器
type TakeoutHandler struct {
	settingsSrv service.ITakeoutSettingsSrv
	orderAppSrv application.ITakeoutOrderAppService
	orderSrv    service.ITakeoutOrderSrv
}

// GetSettings 获取外卖配置
// @Summary 获取外卖配置
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param platform query string true "平台: grab,foodpanda,lineman"
// @Success 200 {object} response.TakeoutSettingsResp
// @Router /cashier/takeout/settings [get]
func (h *TakeoutHandler) GetSettings(c *gin.Context) {
	ctx := helper.GetContext(c)

	platform := c.Query("platform")
	if platform == "" {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.New("platform参数不能为空"))
		return
	}

	settings, err := h.settingsSrv.GetSettings(ctx, platform)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, settings)
}

// SaveSettings 保存外卖配置
// @Summary 保存外卖配置
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param request body request.TakeoutSettingsSaveReq true "配置参数"
// @Success 200 {object} nil "保存成功"
// @Router /cashier/takeout/settings [post]
func (h *TakeoutHandler) SaveSettings(c *gin.Context) {
	ctx := helper.GetContext(c)

	var req request.TakeoutSettingsSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	if err := h.settingsSrv.SaveSettings(ctx, &req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// GetList 获取外卖订单列表
// @Summary 获取外卖订单列表
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param page_no query int true "页码"
// @Param page_size query int true "每页数量"
// @Param platform query string false "平台筛选: grab,foodpanda,lineman"
// @Param status query int false "状态: 0=全部,1=待接单,2=已接单,3=制作中,4=已完成,5=已拒单"
// @Param start_time query int64 false "开始时间"
// @Param end_time query int64 false "结束时间"
// @Param search query string false "搜索关键词"
// @Success 200 {object} dto.Response{data=response.TakeoutOrderListResp} "外卖订单列表"
// @Router /cashier/takeout/order/list [get]
func (h *TakeoutHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)

	var req request.TakeoutOrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	list, err := h.orderSrv.GetList(ctx, &req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, list)
}

// GetDetail 获取外卖订单详情
// @Summary 获取外卖订单详情
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param order_uuid query uint64 true "订单UUID"
// @Success 200 {object} dto.Response{data=response.TakeoutOrderResp} "外卖订单详情"
// @Router /cashier/takeout/order/detail [get]
func (h *TakeoutHandler) GetDetail(c *gin.Context) {
	ctx := helper.GetContext(c)

	var req request.TakeoutOrderDetailReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	order, err := h.orderSrv.GetByUuid(ctx, req.OrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, order)
}

// AcceptOrder 接单
// @Summary 接单
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param request body request.TakeoutOrderAcceptReq true "接单参数"
// @Success 200 {object} nil "接单成功"
// @Router /cashier/takeout/order/accept [post]
func (h *TakeoutHandler) AcceptOrder(c *gin.Context) {
	ctx := helper.GetContext(c)

	var req request.TakeoutOrderAcceptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	if err := h.orderSrv.AcceptOrder(ctx, &req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// RejectOrder 拒单
// @Summary 拒单
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param request body request.TakeoutOrderRejectReq true "拒单参数"
// @Success 200 {object} nil "拒单成功"
// @Router /cashier/takeout/order/reject [post]
func (h *TakeoutHandler) RejectOrder(c *gin.Context) {
	ctx := helper.GetContext(c)

	var req request.TakeoutOrderRejectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	if err := h.orderSrv.RejectOrder(ctx, &req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// SyncOrder 同步订单-模拟接收新订单
// @Summary 同步订单-模拟接收新订单
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param request body request.TakeoutOrderSyncReq true "同步订单参数"
// @Success 200 {object} nil "同步订单成功"
// @Router /cashier/takeout/order/sync [post]
func (h *TakeoutHandler) SyncOrder(c *gin.Context) {
	ctx := helper.GetContext(c)

	var req request.TakeoutOrderSyncReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	if err := h.orderAppSrv.SyncNewOrder(ctx, req.Platform, "req.TakeoutOrderUuid", req.RawData); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// HandlePushOrderState 处理订单状态变更-模拟接收新订单
// @Summary 处理订单状态变更-模拟接收新订单
// @Tags 收银端.外卖管理
// @Accept json
// @Produce json
// @Param request body request.TakeoutOrderEvent true "订单状态变更参数"
// @Success 200 {object} nil "处理订单状态变更成功"
// @Router /cashier/takeout/order/push-state [post]
func (h *TakeoutHandler) HandlePushOrderState(c *gin.Context) {
	ctx := helper.GetContext(c)
	var req request.TakeoutOrderEvent
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}

	if err := h.orderAppSrv.HandlePushOrderState(ctx, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// RegisterTakeoutHandlers 注册外卖路由
func RegisterTakeoutHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := appService.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := appService.NewRoleAccessSrv(dbm)
	deviceSrv := appService.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := appService.NewCashBoxSrv(dbm)
	statisticsSrv := appService.NewStatisticsSrv()
	staffShiftSrv := appService.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := appService.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	orderAppSrv := application.NewTakeoutOrderAppService(dbm)
	orderSrv := service.NewTakeoutOrderSrv(dbm)
	settingsSrv := service.NewTakeoutSettingsSrv(dbm)
	wrapper := &TakeoutHandler{
		settingsSrv: settingsSrv,
		orderAppSrv: orderAppSrv,
		orderSrv:    orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		// 配置管理
		privateApi.GET("/takeout/settings", wrapper.GetSettings)   // 获取配置
		privateApi.POST("/takeout/settings", wrapper.SaveSettings) // 保存配置
		// 订单管理
		privateApi.GET("/takeout/order/list", wrapper.GetList)        // 获取订单列表
		privateApi.GET("/takeout/order/detail", wrapper.GetDetail)    // 获取订单详情
		privateApi.POST("/takeout/order/accept", wrapper.AcceptOrder) // 接单
		privateApi.POST("/takeout/order/reject", wrapper.RejectOrder) // 拒单
		// 模拟接收新订单
		privateApi.POST("/takeout/order/sync", wrapper.SyncOrder)                  // 同步订单
		privateApi.POST("/takeout/order/push-state", wrapper.HandlePushOrderState) // 处理订单状态变更
	}
}
