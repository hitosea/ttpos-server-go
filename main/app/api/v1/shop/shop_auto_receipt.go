package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/purchase_order"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AutoReceiptHandler 自动收货规则控制器
type AutoReceiptHandler struct {
	autoReceiptSrv   service.IAutoReceiptSrv
	purchaseOrderSrv purchase_order.IPurchaseOrderSrv
}

// CreateAutoReceiptRule 创建自动收货规则
// @Summary 创建自动收货规则
// @Description 创建自动收货规则，关联门店列表。同一门店在同一仓库内仅允许配置一次
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.CreateAutoReceiptRuleReq true "创建自动收货规则请求"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/rule/create [post]
func (h *AutoReceiptHandler) CreateAutoReceiptRule(c *gin.Context) {
	ctx := helper.GetContext(c)
	var r req.CreateAutoReceiptRuleReq
	if err := c.ShouldBindJSON(&r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.autoReceiptSrv.CreateRule(ctx, r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

// UpdateAutoReceiptRule 更新自动收货规则
// @Summary 更新自动收货规则
// @Description 全量更新自动收货规则，包括名称、仓库、延迟天数、状态和门店列表。门店列表为全量替换语义，自动 diff 增删
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.UpdateAutoReceiptRuleReq true "更新自动收货规则请求"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/rule/update [post]
func (h *AutoReceiptHandler) UpdateAutoReceiptRule(c *gin.Context) {
	ctx := helper.GetContext(c)
	var r req.UpdateAutoReceiptRuleReq
	if err := c.ShouldBindJSON(&r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.autoReceiptSrv.UpdateRule(ctx, r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

// DeleteAutoReceiptRule 删除自动收货规则
// @Summary 删除自动收货规则
// @Description 批量删除自动收货规则，级联软删除关联的门店记录
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeleteAutoReceiptRuleReq true "删除自动收货规则请求"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/rule/delete [delete]
func (h *AutoReceiptHandler) DeleteAutoReceiptRule(c *gin.Context) {
	ctx := helper.GetContext(c)
	var r req.DeleteAutoReceiptRuleReq
	if err := c.ShouldBindJSON(&r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.autoReceiptSrv.DeleteRule(ctx, r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

// GetAutoReceiptRuleList 获取规则列表
// @Summary 获取自动收货规则列表
// @Description 获取当前总部下所有自动收货规则，返回规则及关联门店
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.AutoReceiptRuleListResp "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/rule/list [get]
func (h *AutoReceiptHandler) GetAutoReceiptRuleList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var r req.AutoReceiptRuleListReq
	if err := c.ShouldBindQuery(&r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	data, err := h.autoReceiptSrv.GetRuleList(ctx, r)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, data)
}

// GetAutoReceiptShopList 获取可选门店列表
// @Summary 获取可选门店列表
// @Description 获取当前总部下所有子门店，已在同仓库其他规则中配置的门店标记为 disabled
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Param warehouse_erp_code query string true "仓库ERP编码"
// @Success 200 {object} resp.AutoReceiptShopListResp "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/shop_list [get]
func (h *AutoReceiptHandler) GetAutoReceiptShopList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var r req.AutoReceiptShopListReq
	if err := c.ShouldBindQuery(&r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	data, err := h.autoReceiptSrv.GetShopList(ctx, r)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, data)
}

// GetAutoReceiptLogList 获取自动收货记录列表
// @Summary 获取自动收货记录列表
// @Description 分页查询自动收货执行日志，支持按门店和收货时间范围筛选。前端传入日期时间字符串，后端根据总店时区解析为时间戳
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param shop_company_uuid query int false "门店UUID"
// @Param start_time query string false "收货时间范围-开始（格式: 2006-01-02 15:04:05）"
// @Param end_time query string false "收货时间范围-结束（格式: 2006-01-02 15:04:05）"
// @Success 200 {object} resp.AutoReceiptLogListResp "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/log/list [get]
func (h *AutoReceiptHandler) GetAutoReceiptLogList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var r req.AutoReceiptLogListReq
	if err := c.ShouldBindQuery(&r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	data, err := h.autoReceiptSrv.GetLogList(ctx, r)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, data)
}

// GetAutoReceiptLogDetail 获取自动收货记录详情
// @Summary 获取自动收货记录详情
// @Description 根据日志UUID查询对应门店的收货单详情，跨门店库查询
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "记录UUID"
// @Success 200 {object} resp.PurchaseReceiptOrderDetailResp "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/log/detail [get]
func (h *AutoReceiptHandler) GetAutoReceiptLogDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var r req.AutoReceiptLogDetailReq
	if err := c.ShouldBindQuery(&r); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 查日志并构建门店上下文
	shopCtx, receiptOrderUuid, err := h.autoReceiptSrv.GetLogDetail(ctx, r)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 复用收货单详情接口
	detail, err := h.purchaseOrderSrv.GetPurchaseReceiptOrderDetail(shopCtx, req.PurchaseReceiptOrderDetailReq{
		Uuid: receiptOrderUuid,
	})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, detail)
}

// GetAutoReceiptWarehouseList 获取发货仓库列表
// @Summary 获取发货仓库列表
// @Description 获取当前总部下所有有ERP编码的仓库，用于创建/编辑规则时选择仓库
// @Tags 商家端.自动收货配置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.AutoReceiptWarehouseListResp "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /shop/auto_receipt/warehouse/list [get]
func (h *AutoReceiptHandler) GetAutoReceiptWarehouseList(c *gin.Context) {
	ctx := helper.GetContext(c)
	data, err := h.autoReceiptSrv.GetWarehouseList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, data)
}

// RegisterAutoReceiptHandlers 注册自动收货路由
func RegisterAutoReceiptHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	sSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(sSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, sSrv)

	poSrv := purchase_order.NewPurchaseOrderSrv(dbm, sSrv)

	wrapper := &AutoReceiptHandler{
		autoReceiptSrv:   service.NewAutoReceiptSrv(dbm, cache),
		purchaseOrderSrv: poSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/auto_receipt/rule/create", wrapper.CreateAutoReceiptRule)         // 创建自动收货规则
		privateApi.POST("/auto_receipt/rule/update", wrapper.UpdateAutoReceiptRule)         // 更新自动收货规则
		privateApi.DELETE("/auto_receipt/rule/delete", wrapper.DeleteAutoReceiptRule)       // 删除自动收货规则
		privateApi.GET("/auto_receipt/rule/list", wrapper.GetAutoReceiptRuleList)           // 规则列表
		privateApi.GET("/auto_receipt/shop_list", wrapper.GetAutoReceiptShopList)           // 可选门店列表
		privateApi.GET("/auto_receipt/warehouse/list", wrapper.GetAutoReceiptWarehouseList) // 发货仓库列表
		privateApi.GET("/auto_receipt/log/list", wrapper.GetAutoReceiptLogList)             // 自动收货记录
		privateApi.GET("/auto_receipt/log/detail", wrapper.GetAutoReceiptLogDetail)         // 收货单详情
	}
}
