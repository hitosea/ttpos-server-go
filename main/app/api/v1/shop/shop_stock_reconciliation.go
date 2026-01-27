package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
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

// SaveStockReconciliation 保存盘点单
// @Summary 保存盘点单
// @Description 保存盘点单信息
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationSaveReq true "保存盘点单请求参数"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationUuidResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/save [post]
func (h *StockReconciliationHandler) SaveStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockReconciliationSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	stockReconciliationUuid, err := h.stockReconciliationSrv.SaveStockReconciliation(ctx, saveReq)

	retData := resp.StockReconciliationUuidResp{
		Uuid: stockReconciliationUuid,
	}

	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, retData, errors.WithMessage(err))
		return
	}

	helper.Success(c, retData, "保存成功")
}

// CheckMaterials 检查物品
// @Summary 检查物品
// @Description 检查物品
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationCheckMaterialsReq true "检查物品请求参数"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationCheckMaterialsListResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/check_materials [post]
func (h *StockReconciliationHandler) CheckMaterials(c *gin.Context) {
	ctx := helper.GetContext(c)
	var checkMaterialsReq req.StockReconciliationCheckMaterialsReq
	if err := c.ShouldBindJSON(&checkMaterialsReq); err != nil {
		helper.HandleValidationError(c, err, checkMaterialsReq, nil)
		return
	}
	materials, err := h.stockReconciliationSrv.CheckMaterials(ctx, checkMaterialsReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, materials, "检查成功")
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

	helper.Success(c, gin.H{}, "删除成功")
}

// SubmitStockReconciliation 提交盘点单
// @Summary 提交盘点单
// @Description 提交盘点单进行审核
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationSaveReq true "提交盘点单请求参数"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationUuidResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/submit [post]
func (h *StockReconciliationHandler) SubmitStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockReconciliationSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}
	stockReconciliationUuid, err := h.stockReconciliationSrv.SaveStockReconciliation(ctx, saveReq)
	retData := resp.StockReconciliationUuidResp{
		Uuid: stockReconciliationUuid,
	}
	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, retData, errors.WithMessage(err))
		return
	}

	helper.Success(c, retData, "提交成功")
}

// ApproveStockReconciliation 审核盘点单
// @Summary 审核盘点单
// @Description 审核通过盘点单
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationApproveReq true "审核盘点单请求参数"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationApproveResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/approve [post]
func (h *StockReconciliationHandler) ApproveStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var approveReq req.StockReconciliationApproveReq
	if err := c.ShouldBindJSON(&approveReq); err != nil {
		helper.HandleValidationError(c, err, approveReq, nil)
		return
	}

	disabledMaterials, err := h.stockReconciliationSrv.ApproveStockReconciliation(ctx, approveReq)
	if err != nil {
		if len(disabledMaterials) > 0 {
			helper.ErrorWithData(c, constant.CodeMaterialDisabled, resp.StockReconciliationApproveResp{
				List: disabledMaterials,
			}, err)
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{}, "审核成功")
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

	helper.Success(c, gin.H{}, "驳回成功")
}

// ResubmitStockReconciliation 重新提交盘点单
// @Summary 重新提交盘点单
// @Description 将已驳回的盘点单重新提交审核，支持修改盘点单信息后提交
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockReconciliationSaveReq true "重新提交盘点单请求参数（is_resubmit必须为true）"
// @Success 200 {object} dto.Response{data=resp.StockReconciliationUuidResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_reconciliation/resubmit [post]
func (h *StockReconciliationHandler) ResubmitStockReconciliation(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockReconciliationSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	// 强制设置为重新提交模式
	saveReq.SetIsResubmit(true)

	stockReconciliationUuid, err := h.stockReconciliationSrv.SaveStockReconciliation(ctx, saveReq)
	retData := resp.StockReconciliationUuidResp{
		Uuid: stockReconciliationUuid,
	}
	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, retData, errors.WithMessage(err))
		return
	}

	helper.Success(c, retData, "重新提交成功")
}

// GetStockReconciliationTemplate 获取盘点单模板
// @Summary 获取盘点单模板
// @Description 获取盘点单模板，包含日盘、周盘、月盘的物品编号列表
// @Tags 商家端.盘点管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.StockReconciliationTemplateResp} "成功"
// @Router /shop/stock_reconciliation/template [get]
func (h *StockReconciliationHandler) GetStockReconciliationTemplate(c *gin.Context) {
	ctx := helper.GetContext(c)

	resp, err := h.stockReconciliationSrv.GetStockReconciliationTemplate(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
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
	localeSrv := service.NewLocaleSrv()
	translateSrv := service.NewTranslateSrv(dbm, cache)
	productSrv := service.NewProductSrv(dbm, localeSrv, settingSrv, cache, translateSrv)
	stockReconciliationSrv := service.NewStockReconciliationSrv(dbm, productSrv)

	wrapper := &StockReconciliationHandler{
		stockReconciliationSrv: stockReconciliationSrv,
	}

	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/stock_reconciliation/list", wrapper.GetStockReconciliationList)         // 获取盘点单列表
		privateApi.GET("/stock_reconciliation/detail", wrapper.GetStockReconciliationDetail)     // 获取盘点单详情（包含批注列表）
		privateApi.GET("/stock_reconciliation/template", wrapper.GetStockReconciliationTemplate) // 获取盘点单模板
		privateApi.POST("/stock_reconciliation/save", wrapper.SaveStockReconciliation)           // 保存盘点单
		privateApi.DELETE("/stock_reconciliation/delete", wrapper.DeleteStockReconciliation)     // 删除盘点单
		privateApi.POST("/stock_reconciliation/submit", wrapper.SubmitStockReconciliation)       // 提交盘点单
		privateApi.POST("/stock_reconciliation/approve", wrapper.ApproveStockReconciliation)     // 审核盘点单
		privateApi.POST("/stock_reconciliation/reject", wrapper.RejectStockReconciliation)       // 驳回盘点单
		privateApi.POST("/stock_reconciliation/resubmit", wrapper.ResubmitStockReconciliation)   // 重新提交盘点单
		privateApi.POST("/stock_reconciliation/check_materials", wrapper.CheckMaterials)         // 检查物品
	}
}
