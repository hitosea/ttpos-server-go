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

// StockReconciliationHandler 盘点单控制器
type StockReconciliationHandler struct {
	stockReconciliationSrv service.IStockReconciliationSrv
}

// GetStockReconciliationList 获取盘点单列表
// @Summary 获取盘点单列表
// @Description 分页获取盘点单列表，支持多仓库筛选、关键字搜索、创建时间范围筛选
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页数量"
// @Param warehouse_uuids query []int false "仓库UUID列表"
// @Param keyword query string false "关键字（搜索单据编号和ERP盘点单号）"
// @Param start_create_time query int false "创建开始时间（时间戳）"
// @Param end_create_time query int false "创建结束时间（时间戳）"
// @Param status_in query []int false "状态列表"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationListResp} "成功"
// @Router /shop/stock_reconciliation/list [get]
func (h *StockReconciliationHandler) GetStockReconciliationList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.StockReconciliationListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.stockReconciliationSrv.GetStockReconciliationList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetStockReconciliationDetail 获取盘点单详情
// @Summary 获取盘点单详情
// @Description 根据UUID获取盘点单详情
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "盘点单UUID"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationDetailResp} "成功"
// @Router /shop/stock_reconciliation/detail [get]
func (h *StockReconciliationHandler) GetStockReconciliationDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.StockReconciliationDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.stockReconciliationSrv.GetStockReconciliationDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// CreateStockReconciliation 创建盘点单
// @Summary 创建盘点单
// @Description 创建新的盘点单
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationCreateReq true "创建盘点单请求参数"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationCreateResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/create [post]
func (h *StockReconciliationHandler) CreateStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.StockReconciliationCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	err := h.stockReconciliationSrv.CreateStockReconciliation(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// SaveStockReconciliation 保存盘点单
// @Summary 保存盘点单
// @Description 保存盘点单信息
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationSaveReq true "保存盘点单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/save [post]
func (h *StockReconciliationHandler) SaveStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockReconciliationSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	err := h.stockReconciliationSrv.SaveStockReconciliation(ctx, saveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// DeleteStockReconciliation 删除盘点单
// @Summary 删除盘点单
// @Description 删除盘点单（软删除）
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationDeleteReq true "删除盘点单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/delete [delete]
func (h *StockReconciliationHandler) DeleteStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.StockReconciliationDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.stockReconciliationSrv.DeleteStockReconciliation(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// SubmitStockReconciliation 提交盘点单
// @Summary 提交盘点单
// @Description 提交盘点单进行审核
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationSaveReq true "提交盘点单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/submit [post]
func (h *StockReconciliationHandler) SubmitStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockReconciliationSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	err := h.stockReconciliationSrv.SaveStockReconciliation(ctx, saveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// ApproveStockReconciliation 审核盘点单
// @Summary 审核盘点单
// @Description 审核通过盘点单
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationApproveReq true "审核盘点单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/approve [post]
func (h *StockReconciliationHandler) ApproveStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var approveReq req.StockReconciliationApproveReq
	if err := c.ShouldBindJSON(&approveReq); err != nil {
		helper.HandleValidationError(c, err, approveReq, nil)
		return
	}

	err := h.stockReconciliationSrv.ApproveStockReconciliation(ctx, approveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// RejectStockReconciliation 驳回盘点单
// @Summary 驳回盘点单
// @Description 驳回盘点单
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationRejectReq true "驳回盘点单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/reject [post]
func (h *StockReconciliationHandler) RejectStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var rejectReq req.StockReconciliationRejectReq
	if err := c.ShouldBindJSON(&rejectReq); err != nil {
		helper.HandleValidationError(c, err, rejectReq, nil)
		return
	}

	err := h.stockReconciliationSrv.RejectStockReconciliation(ctx, rejectReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, nil)
}

// RegisterStockReconciliationHandlers 注册盘点单相关路由
func RegisterStockReconciliationHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	stockReconciliationSrv := service.NewStockReconciliationSrv(dbm)

	wrapper := &StockReconciliationHandler{
		stockReconciliationSrv: stockReconciliationSrv,
	}

	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/stock_reconciliation/list", wrapper.GetStockReconciliationList)
		privateApi.GET("/stock_reconciliation/detail", wrapper.GetStockReconciliationDetail)
		privateApi.POST("/stock_reconciliation/create", wrapper.CreateStockReconciliation)
		privateApi.POST("/stock_reconciliation/save", wrapper.SaveStockReconciliation)
		privateApi.DELETE("/stock_reconciliation/delete", wrapper.DeleteStockReconciliation)
		privateApi.POST("/stock_reconciliation/submit", wrapper.SubmitStockReconciliation)
		privateApi.POST("/stock_reconciliation/approve", wrapper.ApproveStockReconciliation)
		privateApi.POST("/stock_reconciliation/reject", wrapper.RejectStockReconciliation)
	}
}
