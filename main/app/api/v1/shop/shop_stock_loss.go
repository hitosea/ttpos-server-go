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

// StockLossHandler 报损单控制器
type StockLossHandler struct {
	stockLossSrv service.IStockLossSrv
}

// GetStockLossList 获取报损单列表
// @Summary 获取报损单列表
// @Description 分页获取报损单列表，支持多仓库筛选、关键字搜索、创建时间范围筛选
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页数量"
// @Param warehouse_uuids query []int false "仓库UUID列表"
// @Param keyword query string false "关键字（搜索单据编号和ERP单据编号）"
// @Param start_create_time query int false "创建开始时间（时间戳）"
// @Param end_create_time query int false "创建结束时间（时间戳）"
// @Param status_in query []int false "状态列表 0:已保存 1:已提交 2:已审核通过 3:已驳回"
// @Param loss_type query int false "报损类型 1:物品损坏 2:物品报废 3:物品过期"
// @Success 200 {object} dto.Response{data=resp.StockLossListResp} "成功"
// @Router /shop/stock_loss/list [get]
func (h *StockLossHandler) GetStockLossList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.StockLossListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.stockLossSrv.GetStockLossList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetStockLossDetail 获取报损单详情
// @Summary 获取报损单详情
// @Description 根据UUID获取报损单详情，包含物品明细、附件、批注
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "报损单UUID"
// @Success 200 {object} dto.Response{data=resp.StockLossDetailResp} "成功"
// @Router /shop/stock_loss/detail [get]
func (h *StockLossHandler) GetStockLossDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.StockLossDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.stockLossSrv.GetStockLossDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// SaveStockLoss 保存报损单
// @Summary 保存报损单
// @Description 保存报损单信息（新建或修改）
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockLossSaveReq true "保存报损单请求参数"
// @Success 200 {object} dto.Response{data=resp.StockLossUuidResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_loss/save [post]
func (h *StockLossHandler) SaveStockLoss(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockLossSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	stockLossUuid, err := h.stockLossSrv.SaveStockLoss(ctx, saveReq)

	retData := resp.StockLossUuidResp{
		Uuid: stockLossUuid,
	}

	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, retData, errors.WithMessage(err))
		return
	}

	helper.Success(c, retData, "保存成功")
}

// DeleteStockLoss 删除报损单
// @Summary 删除报损单
// @Description 删除报损单（软删除，仅已保存状态可删除）
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockLossDeleteReq true "删除报损单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_loss/delete [delete]
func (h *StockLossHandler) DeleteStockLoss(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.StockLossDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.stockLossSrv.DeleteStockLoss(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{}, "删除成功")
}

// SubmitStockLoss 提交报损单
// @Summary 提交报损单
// @Description 提交报损单进行审核
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockLossSaveReq true "提交报损单请求参数（is_submit设为true）"
// @Success 200 {object} dto.Response{data=resp.StockLossUuidResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_loss/submit [post]
func (h *StockLossHandler) SubmitStockLoss(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockLossSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	// 设置为提交模式
	saveReq.IsSubmit = true

	stockLossUuid, err := h.stockLossSrv.SaveStockLoss(ctx, saveReq)
	retData := resp.StockLossUuidResp{
		Uuid: stockLossUuid,
	}
	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, retData, errors.WithMessage(err))
		return
	}

	helper.Success(c, retData, "提交成功")
}

// ApproveStockLoss 审核通过报损单
// @Summary 审核通过报损单
// @Description 审核通过报损单，扣减仓库库存
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockLossApproveReq true "审核报损单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_loss/approve [post]
func (h *StockLossHandler) ApproveStockLoss(c *gin.Context) {
	ctx := helper.GetContext(c)
	var approveReq req.StockLossApproveReq
	if err := c.ShouldBindJSON(&approveReq); err != nil {
		helper.HandleValidationError(c, err, approveReq, nil)
		return
	}

	err := h.stockLossSrv.ApproveStockLoss(ctx, approveReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{}, "审核成功")
}

// RejectStockLoss 驳回报损单
// @Summary 驳回报损单
// @Description 驳回报损单
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockLossRejectReq true "驳回报损单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_loss/reject [post]
func (h *StockLossHandler) RejectStockLoss(c *gin.Context) {
	ctx := helper.GetContext(c)
	var rejectReq req.StockLossRejectReq
	if err := c.ShouldBindJSON(&rejectReq); err != nil {
		helper.HandleValidationError(c, err, rejectReq, nil)
		return
	}

	err := h.stockLossSrv.RejectStockLoss(ctx, rejectReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{}, "驳回成功")
}

// ResubmitStockLoss 重新提交报损单
// @Summary 重新提交报损单
// @Description 将已驳回的报损单重新提交审核，支持修改报损单信息后提交
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.StockLossSaveReq true "重新提交报损单请求参数"
// @Success 200 {object} dto.Response{data=resp.StockLossUuidResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/stock_loss/resubmit [post]
func (h *StockLossHandler) ResubmitStockLoss(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.StockLossSaveReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.HandleValidationError(c, err, saveReq, nil)
		return
	}

	// 强制设置为重新提交模式
	saveReq.SetIsResubmit(true)

	stockLossUuid, err := h.stockLossSrv.SaveStockLoss(ctx, saveReq)
	retData := resp.StockLossUuidResp{
		Uuid: stockLossUuid,
	}
	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, retData, errors.WithMessage(err))
		return
	}

	helper.Success(c, retData, "重新提交成功")
}

// GetStockLossAnnotationList 获取报损单批注列表
// @Summary 获取报损单批注列表
// @Description 根据报损单UUID获取批注列表
// @Tags 商家端.报损管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param stock_loss_uuid query int true "报损单UUID"
// @Success 200 {object} dto.Response{data=resp.StockLossAnnotationListResp} "成功"
// @Router /shop/stock_loss/annotation_list [get]
func (h *StockLossHandler) GetStockLossAnnotationList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var annotationReq req.StockLossAnnotationListReq
	if err := c.ShouldBindQuery(&annotationReq); err != nil {
		helper.HandleValidationError(c, err, annotationReq, nil)
		return
	}

	resp, err := h.stockLossSrv.GetStockLossAnnotationList(ctx, annotationReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// RegisterStockLossHandlers 注册报损单相关路由
func RegisterStockLossHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	stockLossSrv := service.NewStockLossSrv(dbm)

	wrapper := &StockLossHandler{
		stockLossSrv: stockLossSrv,
	}

	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/stock_loss/list", wrapper.GetStockLossList)                      // 获取报损单列表
		privateApi.GET("/stock_loss/detail", wrapper.GetStockLossDetail)                  // 获取报损单详情
		privateApi.GET("/stock_loss/annotation_list", wrapper.GetStockLossAnnotationList) // 获取批注列表
		privateApi.POST("/stock_loss/save", wrapper.SaveStockLoss)                        // 保存报损单
		privateApi.DELETE("/stock_loss/delete", wrapper.DeleteStockLoss)                  // 删除报损单
		privateApi.POST("/stock_loss/submit", wrapper.SubmitStockLoss)                    // 提交报损单
		privateApi.POST("/stock_loss/approve", wrapper.ApproveStockLoss)                  // 审核通过报损单
		privateApi.POST("/stock_loss/reject", wrapper.RejectStockLoss)                    // 驳回报损单
		privateApi.POST("/stock_loss/resubmit", wrapper.ResubmitStockLoss)                // 重新提交报损单
	}
}
