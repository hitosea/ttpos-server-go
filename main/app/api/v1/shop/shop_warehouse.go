package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc/message"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// WarehouseHandler 仓库控制器
type WarehouseHandler struct {
	authSrv                    service.IAuthSrv
	warehouseSrv               service.IWarehouseSrv
	salesOutboundSummarySrv    service.ISalesOutboundSummarySrv
}

// GetWarehouseList 获取对方机构列表
// @Summary 获取对方机构列表
// @Description 获取对方机构列表
// @Tags 商家端.仓库档案
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.OtherOrgListResp} "成功"
// @Router /shop/warehouse/org/list [get]
func (h *WarehouseHandler) GetOtherOrgList(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.warehouseSrv.GetOtherOrgList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// GetWarehouseList 获取仓库列表
// @Summary 获取仓库列表
// @Description 获取仓库列表，支持分页和筛选
// @Tags 商家端.仓库档案
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
		helper.HandleValidationError(c, err, request, nil)
		return
	}
	result, err := h.warehouseSrv.GetWarehouseList(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// GetWarehouseList 获取总部仓库列表
// @Summary 获取总部仓库列表
// @Description 获取仓库列表，支持分页和筛选
// @Tags 商家端.仓库档案
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.WarehouseListResp} "成功"
// @Router /shop/warehouse/headquarter/list [get]
func (h *WarehouseHandler) GetHeadquarterWarehouseList(c *gin.Context) {
	ctx := helper.GetContext(c)

	result, err := h.warehouseSrv.GetHeadquarterWarehouseList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// GetWarehouse 获取仓库
// @Summary 获取仓库
// @Description 获取仓库
// @Tags 商家端.仓库档案
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string false "仓库ID"
// @Success 200 {object} dto.Response{data=resp.WarehouseResp} "成功"
// @Router /shop/warehouse [get]
func (h *WarehouseHandler) GetWarehouse(c *gin.Context) {
	ctx := helper.GetContext(c)
	var request req.WarehouseReq
	if err := c.ShouldBindQuery(&request); err != nil {
		helper.HandleValidationError(c, err, request, nil)
		return
	}
	result, err := h.warehouseSrv.GetWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, result)
}

// CreateWarehouse 创建仓库
// @Summary 创建仓库
// @Description 创建新的仓库
// @Tags 商家端.仓库档案
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
		helper.HandleValidationError(c, err, request, nil)
		return
	}
	err := h.warehouseSrv.CreateWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{}, "保存成功")
}

// UpdateWarehouse 更新仓库
// @Summary 更新仓库
// @Description 更新仓库信息
// @Tags 商家端.仓库档案
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
		helper.HandleValidationError(c, err, request, nil)
		return
	}

	err := h.warehouseSrv.UpdateWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{}, "保存成功")
}

// SetDefaultWarehouse 设置默认仓库
// @Summary 更新仓库
// @Description 更新仓库信息
// @Tags 商家端.仓库档案
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
		helper.HandleValidationError(c, err, request, nil)
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
// @Param request query req.GetWarehouseInOutListReq true "获取仓库出入库明细列表请求"
// @Success 200 {object} dto.Response{data=resp.WarehouseInOutListResp} "成功"
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
// @Tags 商家端.仓库档案
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
		helper.HandleValidationError(c, err, request, nil)
		return
	}
	err := h.warehouseSrv.DeleteWarehouse(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// CheckCodeExists 检查仓库编码是否存在
// @Summary 检查仓库编码是否存在
// @Tags 商家端.仓库档案
// @Accept json
// @Produce json
// @Security JwtToken
// @Param code query string true "仓库编码"
// @Param uuid query string false "仓库UUID"
// @Success 200 {object} dto.Response{data=resp.CheckNameCodeExistsResp} "成功"
// @Router /shop/warehouse/code_exists [get]
func (h *WarehouseHandler) CheckCodeExists(c *gin.Context) {
	var checkReq req.CheckCodeExistsReq
	if err := c.ShouldBindQuery(&checkReq); err != nil {
		helper.HandleValidationError(c, err, checkReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.warehouseSrv.CheckCodeExists(ctx, checkReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetWarehouseMaterialList 获取仓库物品列表
// @Summary 获取仓库物品列表
// @Description 根据仓库ID获取仓库物品列表，包含物品多语言名称、库存数量、所有单位及转换率
// @Tags 商家端.仓库档案
// @Accept json
// @Produce json
// @Security JwtToken
// @Param warehouse_uuid query int true "仓库UUID"
// @Param page_no query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20"
// @Success 200 {object} dto.Response{data=resp.WarehouseMaterialListResp} "成功"
// @Router /shop/warehouse/material/list [get]
func (h *WarehouseHandler) GetWarehouseMaterialList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var request req.WarehouseMaterialListReq
	if err := c.ShouldBindQuery(&request); err != nil {
		helper.HandleValidationError(c, err, request, nil)
		return
	}

	result, err := h.warehouseSrv.GetWarehouseMaterialList(ctx, request)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result)
}

// RegenerateSalesOutboundSummary 重新生成销售出库汇总记录
// @Summary 重新生成销售出库汇总记录
// @Description 删除指定日期的旧销售出库汇总记录，并重新生成新的汇总记录
// @Tags 商家端.仓库管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.RegenerateSalesOutboundSummaryReq true "重新生成销售出库汇总记录请求"
// @Success 200 {object} dto.Response{data=resp.RegenerateSalesOutboundSummaryResp} "成功"
// @Router /shop/inventory/regenerate-sales-outbound-summary [post]
func (h *WarehouseHandler) RegenerateSalesOutboundSummary(c *gin.Context) {
	ctx := helper.GetContext(c)
	companyUuid := ctx.GetCompanyUuid()

	var request req.RegenerateSalesOutboundSummaryReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.HandleValidationError(c, err, request, nil)
		return
	}

	// 使用当前门店的 UUID（如果请求中没有指定）
	if request.CompanyUuid == 0 {
		request.CompanyUuid = companyUuid
	}

	// 权限校验：只能操作自己门店的数据
	if request.CompanyUuid != companyUuid {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("无权操作其他门店的数据"))
		return
	}

	result, err := h.salesOutboundSummarySrv.RegenerateSalesOutboundSummary(c, request.CompanyUuid, request.Date)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, result, "重新生成成功")
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
	translateSrv := service.NewTranslateSrv(dbm, cache)
	messageSrv := message.NewIMessageSrv(dbm)
	materialSrv := service.NewMaterialSrv(dbm, service.NewLocaleSrv(), settingSrv, translateSrv, messageSrv)
	warehouseSrv := service.NewWarehouseSrv(dbm, settingSrv, materialSrv, translateSrv)
	salesOutboundSummarySrv := service.NewSalesOutboundSummarySrv(dbm, settingSrv, cache)

	// 初始化控制器
	warehouseHandler := &WarehouseHandler{
		authSrv:                 authSrv,
		warehouseSrv:            warehouseSrv,
		salesOutboundSummarySrv: salesOutboundSummarySrv,
	}

	// 需要认证的路由
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/warehouse/list", warehouseHandler.GetWarehouseList)                        // 仓库列表
		privateApi.GET("/warehouse/headquarter/list", warehouseHandler.GetHeadquarterWarehouseList) // 总部仓库列表
		privateApi.GET("/warehouse", warehouseHandler.GetWarehouse)                                 // 获取仓库
		privateApi.POST("/warehouse/create", warehouseHandler.CreateWarehouse)                      // 创建仓库
		privateApi.POST("/warehouse/update", warehouseHandler.UpdateWarehouse)                      // 更新仓库
		privateApi.DELETE("/warehouse/delete", warehouseHandler.DeleteWarehouse)                    // 删除仓库
		privateApi.POST("/warehouse/set_default", warehouseHandler.SetDefaultWarehouse)             // 设置默认
		privateApi.GET("/warehouse/code_exists", warehouseHandler.CheckCodeExists)                  // 检查仓库编码是否存在

		privateApi.GET("/warehouse/in_out/list", warehouseHandler.GetWarehouseInOutList) // 出入库明细列表

		privateApi.GET("/warehouse/org/list", warehouseHandler.GetOtherOrgList) // 对方机构列表

		privateApi.GET("/warehouse/material/list", warehouseHandler.GetWarehouseMaterialList) // 仓库物品列表

		privateApi.POST("/inventory/regenerate-sales-outbound-summary", warehouseHandler.RegenerateSalesOutboundSummary) // 重新生成销售出库汇总记录
	}
}
