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

// NationalityHandler 国籍处理器
//
// 任务: story-order-source-nationality Phase 2.6
// 需求: R2.1-R2.7
//
// @version v2.10.0
type NationalityHandler struct {
	nationalitySrv service.INationalitySrv
}

// GetList 获取国籍列表
// @Summary 获取国籍列表
// @Description 获取所有国籍配置
// @Tags 商家端.国籍
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.NationalityListResp}
// @Router /shop/nationality/list [get]
func (h *NationalityHandler) GetList(c *gin.Context) {
	ctx := helper.GetContext(c)

	list, err := h.nationalitySrv.GetList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, list, "获取成功")
}

// Create 创建国籍
// @Summary 创建国籍
// @Description 创建新的国籍配置
// @Tags 商家端.国籍
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.NationalityCreateReq true "国籍数据"
// @Success 200 {object} dto.Response{data=resp.NationalityCreateResp}
// @Router /shop/nationality/create [post]
func (h *NationalityHandler) Create(c *gin.Context) {
	ctx := helper.GetContext(c)

	var createReq req.NationalityCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	result, err := h.nationalitySrv.Create(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result, "创建成功")
}

// Update 更新国籍
// @Summary 更新国籍
// @Description 更新国籍配置
// @Tags 商家端.国籍
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.NationalityUpdateReq true "国籍数据"
// @Success 200 {object} dto.Response
// @Router /shop/nationality/update [post]
func (h *NationalityHandler) Update(c *gin.Context) {
	ctx := helper.GetContext(c)

	var updateReq req.NationalityUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	if err := h.nationalitySrv.Update(ctx, updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "更新成功")
}

// Delete 删除国籍
// @Summary 删除国籍
// @Description 删除国籍配置
// @Tags 商家端.国籍
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.NationalityDeleteReq true "删除参数"
// @Success 200 {object} dto.Response
// @Router /shop/nationality/delete [post]
func (h *NationalityHandler) Delete(c *gin.Context) {
	ctx := helper.GetContext(c)

	var deleteReq req.NationalityDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	if err := h.nationalitySrv.Delete(ctx, deleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil, "删除成功")
}

// RegisterNationalityRoutes 注册国籍路由
func RegisterNationalityRoutes(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	nationalitySrv := service.NewNationalitySrv(dbm)
	handler := &NationalityHandler{
		nationalitySrv: nationalitySrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/nationality/list", handler.GetList)
		privateApi.POST("/nationality/create", handler.Create)
		privateApi.POST("/nationality/update", handler.Update)
		privateApi.POST("/nationality/delete", handler.Delete)
	}
}
