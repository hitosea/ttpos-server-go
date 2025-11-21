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

// FullReductionActivityHandler 满减活动处理器
type FullReductionActivityHandler struct {
	fullReductionActivitySrv service.IFullReductionActivitySrv
}

// Create 创建满减活动
// @Summary 创建满减活动
// @Description 创建满减活动
// @Tags 商家端.满减活动
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.FullReductionActivityCreateReq true "创建活动"
// @Success 200 {object} dto.Response{data=resp.FullReductionActivityResp}
// @Router /shop/full_reduction_activity/create [post]
func (h *FullReductionActivityHandler) Create(c *gin.Context) {
	ctx := helper.GetContext(c)

	var createReq req.FullReductionActivityCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	result, err := h.fullReductionActivitySrv.Create(ctx, &createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result, "创建成功")
}

// Update 更新满减活动
// @Summary 更新满减活动
// @Description 更新满减活动
// @Tags 商家端.满减活动
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.FullReductionActivityUpdateReq true "更新活动"
// @Success 200 {object} dto.Response
// @Router /shop/full_reduction_activity/update [post]
func (h *FullReductionActivityHandler) Update(c *gin.Context) {
	ctx := helper.GetContext(c)

	var updateReq req.FullReductionActivityUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	if err := h.fullReductionActivitySrv.Update(ctx, &updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "更新成功")
}

// GetByUuid 获取满减活动详情
// @Summary 获取满减活动详情
// @Description 根据UUID获取满减活动详情
// @Tags 商家端.满减活动
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.FullReductionActivityGetReq true "获取活动"
// @Success 200 {object} dto.Response{data=resp.FullReductionActivityResp}
// @Router /shop/full_reduction_activity/get [post]
func (h *FullReductionActivityHandler) GetByUuid(c *gin.Context) {
	ctx := helper.GetContext(c)

	var getReq req.FullReductionActivityGetReq
	if err := c.ShouldBindJSON(&getReq); err != nil {
		helper.HandleValidationError(c, err, getReq, nil)
		return
	}

	result, err := h.fullReductionActivitySrv.GetByUuid(ctx, getReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result, "获取成功")
}

// GetList 获取满减活动列表
// @Summary 获取满减活动列表
// @Description 获取满减活动列表
// @Tags 商家端.满减活动
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.FullReductionActivityListReq true "获取列表"
// @Success 200 {object} dto.Response{data=resp.FullReductionActivityListResp}
// @Router /shop/full_reduction_activity/list [post]
func (h *FullReductionActivityHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)

	var listReq req.FullReductionActivityListReq
	if err := c.ShouldBindJSON(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	result, err := h.fullReductionActivitySrv.GetList(ctx, &listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result, "获取成功")
}

// Delete 删除满减活动
// @Summary 删除满减活动
// @Description 删除满减活动
// @Tags 商家端.满减活动
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.FullReductionActivityDeleteReq true "删除活动"
// @Success 200 {object} dto.Response
// @Router /shop/full_reduction_activity/delete [post]
func (h *FullReductionActivityHandler) Delete(c *gin.Context) {
	ctx := helper.GetContext(c)

	var deleteReq req.FullReductionActivityDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	if err := h.fullReductionActivitySrv.Delete(ctx, deleteReq.Uuid); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "删除成功")
}

// Disable 失效满减活动
// @Summary 失效满减活动
// @Description 失效满减活动
// @Tags 商家端.满减活动
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.FullReductionActivityDisableReq true "失效活动"
// @Success 200 {object} dto.Response
// @Router /shop/full_reduction_activity/disable [post]
func (h *FullReductionActivityHandler) Disable(c *gin.Context) {
	ctx := helper.GetContext(c)

	var disableReq req.FullReductionActivityDisableReq
	if err := c.ShouldBindJSON(&disableReq); err != nil {
		helper.HandleValidationError(c, err, disableReq, nil)
		return
	}

	if err := h.fullReductionActivitySrv.Disable(ctx, disableReq.Uuid); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "失效成功")
}

// RegisterFullReductionActivityRoutes 注册满减活动路由
func RegisterFullReductionActivityRoutes(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	fullReductionActivitySrv := service.NewFullReductionActivitySrv(dbm)
	handler := &FullReductionActivityHandler{
		fullReductionActivitySrv: fullReductionActivitySrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/full_reduction_activity/create", handler.Create)
		privateApi.POST("/full_reduction_activity/update", handler.Update)
		privateApi.POST("/full_reduction_activity/get", handler.GetByUuid)
		privateApi.POST("/full_reduction_activity/list", handler.GetList)
		privateApi.POST("/full_reduction_activity/delete", handler.Delete)
		privateApi.POST("/full_reduction_activity/disable", handler.Disable)
	}
}

