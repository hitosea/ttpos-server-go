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

// WarehouseHandler 仓库控制器
type WarehouseHandler struct {
	authSrv      service.IAuthSrv
	warehouseSrv service.IWarehouseSrv
}

// GetWarehouseList 获取仓库列表
// @Summary 获取仓库列表
// @Description 获取仓库列表，支持分页和筛选
// @Tags 商家端.仓库管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20"
// @Param keyword query string false "搜索关键字：仓库编码或名称"
// @Param status query int false "仓库状态:0-禁用；1-启用"
// @Param type query string false "仓库类型：normal-普通；transit-在途"
// @Success 200 {object} dto.Response{data=resp.WarehouseListResp} "成功"
// @Router /shop/warehouse/list [get]
func (h *WarehouseHandler) GetWarehouseList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var request req.WarehouseListReq
	if err := c.ShouldBindQuery(&request); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	result, err := h.warehouseSrv.GetWarehouseList(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// CreateWarehouse 创建仓库
// @Summary 创建仓库
// @Description 创建新的仓库
// @Tags 商家端.仓库管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.CreateWarehouseReq true "创建仓库请求"
// @Success 200 {object} dto.Response "成功"
// @Router /shop/warehouse/create [post]
func (h *WarehouseHandler) CreateWarehouse(c *gin.Context) {
	ctx := helper.GetContext(c)

	var request req.CreateWarehouseReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	err := h.warehouseSrv.CreateWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// UpdateWarehouse 更新仓库
// @Summary 更新仓库
// @Description 更新仓库信息
// @Tags 商家端.仓库管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.UpdateWarehouseReq true "更新仓库请求"
// @Success 200 {object} dto.Response "成功"
// @Router /shop/warehouse/update [post]
func (h *WarehouseHandler) UpdateWarehouse(c *gin.Context) {
	ctx := helper.GetContext(c)

	var request req.UpdateWarehouseReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	err := h.warehouseSrv.UpdateWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// SetDefaultWarehouse 设置默认仓库
// @Summary 更新仓库
// @Description 更新仓库信息
// @Tags 商家端.仓库管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.SetDefaultWarehouseReq true "设置默认仓库请求"
// @Success 200 {object} dto.Response "成功"
// @Router /shop/warehouse/set_default [post]
func (h *WarehouseHandler) SetDefaultWarehouse(c *gin.Context) {
	ctx := helper.GetContext(c)
	var request req.SetDefaultWarehouseReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.warehouseSrv.SetDefaultWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetWarehouseInOutList 获取仓库出入库明细列表
// @Summary 获取仓库出入库明细列表
// @Description 获取仓库出入库明细列表
// @Tags 商家端.仓库管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.GetWarehouseInOutListReq true "获取仓库出入库明细列表请求"
// @Success 200 {object} dto.Response "成功"
// @Router /shop/warehouse/in_out/list [get]
func (h *WarehouseHandler) GetWarehouseInOutList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var request req.GetWarehouseInOutListReq
	if err := c.ShouldBindQuery(&request); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	result, err := h.warehouseSrv.GetWarehouseInOutList(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result)
}

// DeleteWarehouse 删除仓库
// @Summary 删除仓库
// @Description 删除仓库
// @Tags 商家端.仓库管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.DeleteWarehouseReq true "删除仓库请求"
// @Success 200 {object} dto.Response "成功"
// @Router /shop/warehouse/delete [delete]
func (h *WarehouseHandler) DeleteWarehouse(c *gin.Context) {
	ctx := helper.GetContext(c)
	var request req.DeleteWarehouseReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	err := h.warehouseSrv.DeleteWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// RegisterWarehouseHandlers 注册仓库相关路由
func RegisterWarehouseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	warehouseSrv := service.NewWarehouseSrv(dbm)

	// 初始化控制器
	warehouseHandler := &WarehouseHandler{
		authSrv:      authSrv,
		warehouseSrv: warehouseSrv,
	}

	// 需要认证的路由
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{

		privateApi.GET("/warehouse/list", warehouseHandler.GetWarehouseList)            // 仓库列表
		privateApi.POST("/warehouse/create", warehouseHandler.CreateWarehouse)          // 创建仓库
		privateApi.POST("/warehouse/update", warehouseHandler.UpdateWarehouse)          // 更新仓库
		privateApi.DELETE("/warehouse/delete", warehouseHandler.DeleteWarehouse)        // 删除仓库
		privateApi.POST("/warehouse/set_default", warehouseHandler.SetDefaultWarehouse) // 设置默认

		privateApi.GET("/warehouse/in_out/list", warehouseHandler.GetWarehouseInOutList) // 出入库明细列表
	}
}
