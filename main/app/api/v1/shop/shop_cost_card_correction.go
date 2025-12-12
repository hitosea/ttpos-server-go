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

// CostCardCorrectionHandler 成本卡修正处理程序
type CostCardCorrectionHandler struct {
	costCardCorrectionSrv service.ICostCardCorrectionSrv
}

// NewCostCardCorrectionHandler 创建成本卡修正处理程序
func NewCostCardCorrectionHandler(costCardCorrectionSrv service.ICostCardCorrectionSrv) *CostCardCorrectionHandler {
	return &CostCardCorrectionHandler{
		costCardCorrectionSrv: costCardCorrectionSrv,
	}
}

// PreviewCorrection 预览修正影响
// @Summary 预览修正影响
// @Description 预览成本卡修正对订单、材料库存的影响
// @Tags 成本卡修正
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.CostCardCorrectionPreviewReq true "预览请求"
// @Success 200 {object} dto.Response{data=resp.CostCardCorrectionPreviewResp} "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /api/v1/cost_card_correction/preview [post]
func (h *CostCardCorrectionHandler) PreviewCorrection(c *gin.Context) {
	// 绑定请求参数
	var req req.CostCardCorrectionPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}

	// 验证参数
	if err := req.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}

	// 获取 context
	ctx := helper.GetContext(c)

	// 调用服务
	res, err := h.costCardCorrectionSrv.PreviewCorrection(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// ExecuteCorrection 执行修正操作
// @Summary 执行修正操作
// @Description 执行成本卡修正，包括退回材料、重新计算消耗、重新生成出库记录等
// @Tags 成本卡修正
// @Accept json
// @Produce json
// @Security JwtToken
// @Param request body req.CostCardCorrectionReq true "修正请求"
// @Success 200 {object} dto.Response{data=resp.CostCardCorrectionResp} "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /api/v1/cost_card_correction/execute [post]
func (h *CostCardCorrectionHandler) ExecuteCorrection(c *gin.Context) {
	// 绑定请求参数
	var req req.CostCardCorrectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}

	// 验证参数
	if err := req.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}

	// 获取 context
	ctx := helper.GetContext(c)

	// 调用服务
	res, err := h.costCardCorrectionSrv.ExecuteCorrection(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// GetCorrectionLogs 查询修正日志
// @Summary 查询修正日志
// @Description 查询成本卡修正操作日志
// @Tags 成本卡修正
// @Accept json
// @Produce json
// @Security JwtToken
// @Param correction_uuid query uint64 false "修正操作UUID"
// @Param order_uuid query uint64 false "订单UUID"
// @Param page_no query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} dto.Response{data=resp.CostCardCorrectionLogsResp} "成功"
// @Failure 400 {object} dto.Response "错误请求"
// @Router /api/v1/cost_card_correction/logs [get]
func (h *CostCardCorrectionHandler) GetCorrectionLogs(c *gin.Context) {
	// 绑定请求参数
	var req req.CostCardCorrectionLogsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}

	// 验证参数
	if err := req.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, err)
		return
	}

	// 获取 context
	ctx := helper.GetContext(c)

	// 调用服务
	res, err := h.costCardCorrectionSrv.GetCorrectionLogs(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// RegisterCostCardCorrectionHandlers 注册成本卡修正路由
func RegisterCostCardCorrectionHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 创建处理程序
	costCardCorrectionHandler := NewCostCardCorrectionHandler(service.NewCostCardCorrectionSrv(dbm))

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/cost_card_correction/preview", costCardCorrectionHandler.PreviewCorrection)
		privateApi.POST("/cost_card_correction/execute", costCardCorrectionHandler.ExecuteCorrection)
		privateApi.GET("/cost_card_correction/logs", costCardCorrectionHandler.GetCorrectionLogs)
	}
}
