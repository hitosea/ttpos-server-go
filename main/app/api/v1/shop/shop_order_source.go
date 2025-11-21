package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// OrderSourceHandler 外卖来源处理器
//
// 任务: story-order-source-nationality Phase 2.5
// 需求: R1.1-R1.6
//
// @version v2.10.0
type OrderSourceHandler struct {
	orderSourceSrv service.IOrderSourceSrv
}

// GetList 获取外卖来源列表
// @Summary 获取外卖来源列表
// @Description 获取所有外卖来源配置
// @Tags 商家端.外卖来源
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.OrderSourceListResp}
// @Router /shop/order_source/list [get]
func (h *OrderSourceHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)

	list, err := h.orderSourceSrv.GetList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, list, "获取成功")
}

// Create 创建外卖来源
// @Summary 创建外卖来源
// @Description 创建新的外卖来源配置
// @Tags 商家端.外卖来源
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.OrderSourceCreateReq true "外卖来源数据"
// @Success 200 {object} dto.Response{data=resp.OrderSourceCreateResp}
// @Router /shop/order_source/create [post]
func (h *OrderSourceHandler) Create(c *gin.Context) {
	ctx := helper.GetContext(c)

	var createReq req.OrderSourceCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	result, err := h.orderSourceSrv.Create(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result, "创建成功")
}

// Update 更新外卖来源
// @Summary 更新外卖来源
// @Description 更新外卖来源配置
// @Tags 商家端.外卖来源
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.OrderSourceUpdateReq true "外卖来源数据"
// @Success 200 {object} dto.Response
// @Router /shop/order_source/update [post]
func (h *OrderSourceHandler) Update(c *gin.Context) {
	ctx := helper.GetContext(c)

	var updateReq req.OrderSourceUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	if err := h.orderSourceSrv.Update(ctx, updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "更新成功")
}

// Delete 删除外卖来源
// @Summary 删除外卖来源
// @Description 删除外卖来源配置
// @Tags 商家端.外卖来源
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.OrderSourceDeleteReq true "删除参数"
// @Success 200 {object} dto.Response
// @Router /shop/order_source/delete [post]
func (h *OrderSourceHandler) Delete(c *gin.Context) {
	ctx := helper.GetContext(c)

	var deleteReq req.OrderSourceDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	if err := h.orderSourceSrv.Delete(ctx, deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "删除成功")
}

// RegisterOrderSourceRoutes 注册外卖来源路由
func RegisterOrderSourceRoutes(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	orderSourceSrv := service.NewOrderSourceSrv(dbm)
	handler := &OrderSourceHandler{
		orderSourceSrv: orderSourceSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/order_source/list", handler.GetList)
		privateApi.POST("/order_source/create", handler.Create)
		privateApi.POST("/order_source/update", handler.Update)
		privateApi.POST("/order_source/delete", handler.Delete)
	}
}
