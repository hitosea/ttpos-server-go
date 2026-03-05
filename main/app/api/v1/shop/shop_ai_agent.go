package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/ai_agent"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	purchaseorder "ttpos-server-go/app/service/purchase_order"
	message "ttpos-server-go/app/service/rpc/message"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AIAgentHandler handles AI procurement agent endpoints.
type AIAgentHandler struct {
	authSrv service.IAuthSrv
	agent   *ai_agent.Agent
}

// RunAnalysisReq request for running procurement analysis.
type RunAnalysisReq struct {
	WarehouseUuid uint64 `json:"warehouse_uuid" form:"warehouse_uuid" binding:"omitempty"`
	ForecastDays  int    `json:"forecast_days" form:"forecast_days" binding:"omitempty,min=1,max=30"`
}

// SubmitReviewReq request for submitting a review decision.
type SubmitReviewReq struct {
	SessionID string `json:"session_id" binding:"required"`
	Decision  string `json:"decision" binding:"required,oneof=approved rejected"`
	Comment   string `json:"comment" binding:"omitempty,max=500"`
}

// GetSessionReq request for getting session status.
type GetSessionReq struct {
	SessionID string `form:"session_id" binding:"required"`
}

// RunAnalysis starts the procurement analysis workflow.
// @Summary AI 智能采购分析
// @Description 分析仓库库存、预测需求、生成采购建议。返回分析结果和采购提案，等待审核。
// @Tags 商家端.AI采购
// @Accept json
// @Produce json
// @Param warehouse_uuid body uint64 false "仓库UUID（不传则使用默认仓库）"
// @Param forecast_days body int false "预测天数(默认3天)" default(3)
// @Security JwtToken
// @Success 200 {object} map[string]any "分析结果"
// @Router /shop/ai/procurement/analysis [post]
func (h *AIAgentHandler) RunAnalysis(c *gin.Context) {
	var request RunAnalysisReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.HandleValidationError(c, err, request, nil)
		return
	}

	ctx := helper.GetContext(c)

	// 未传 warehouse_uuid 时自动使用默认仓库
	warehouseUuid := request.WarehouseUuid
	if warehouseUuid == 0 {
		wh, err := repository.NewWarehouseRepo(ctx.GetDB()).GetDefaultWarehouse()
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("获取默认仓库失败"))
			return
		}
		warehouseUuid = wh.Uuid
	}

	sessionID := uuid.New().String()
	state := h.agent.RunAnalysis(ctx, sessionID, warehouseUuid, request.ForecastDays)

	helper.Success(c, map[string]any{
		"session_id": sessionID,
		"status":     state.Status,
		"forecasts":  state.Forecasts,
		"proposals":  state.Proposals,
		"anomalies":  state.Anomalies,
		"step_log":   state.StepLog,
		"error":      state.Error,
	})
}

// SubmitReview submits a review decision (approve/reject) for a procurement proposal.
// @Summary 审核采购提案
// @Description 对AI生成的采购提案进行审核。通过后自动创建采购订单。
// @Tags 商家端.AI采购
// @Accept json
// @Produce json
// @Param session_id body string true "会话ID"
// @Param decision body string true "审核决定: approved/rejected"
// @Param comment body string false "审核意见"
// @Security JwtToken
// @Success 200 {object} map[string]any "审核结果"
// @Router /shop/ai/procurement/review [post]
func (h *AIAgentHandler) SubmitReview(c *gin.Context) {
	var request SubmitReviewReq
	if err := c.ShouldBindJSON(&request); err != nil {
		helper.HandleValidationError(c, err, request, nil)
		return
	}

	ctx := helper.GetContext(c)
	state := h.agent.SubmitReview(ctx, request.SessionID, request.Decision, request.Comment)

	if state.Error != "" {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New(state.Error))
		return
	}

	helper.Success(c, map[string]any{
		"session_id":     request.SessionID,
		"status":         state.Status,
		"decision":       state.ReviewDecision,
		"created_orders": state.CreatedOrders,
		"step_log":       state.StepLog,
	})
}

// GetSession retrieves the current state of a procurement analysis session.
// @Summary 查询分析状态
// @Description 查询AI采购分析会话的当前状态。
// @Tags 商家端.AI采购
// @Accept json
// @Produce json
// @Param session_id query string true "会话ID"
// @Security JwtToken
// @Success 200 {object} map[string]any "会话状态"
// @Router /shop/ai/procurement/session [get]
func (h *AIAgentHandler) GetSession(c *gin.Context) {
	var request GetSessionReq
	if err := c.ShouldBindQuery(&request); err != nil {
		helper.HandleValidationError(c, err, request, nil)
		return
	}

	state, ok := h.agent.GetSession(request.SessionID)
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("session not found"))
		return
	}

	helper.Success(c, map[string]any{
		"session_id":      request.SessionID,
		"status":          state.Status,
		"needs_purchase":  state.NeedsPurchase,
		"forecasts":       state.Forecasts,
		"proposals":       state.Proposals,
		"anomalies":       state.Anomalies,
		"created_orders":  state.CreatedOrders,
		"review_decision": state.ReviewDecision,
		"step_log":        state.StepLog,
		"error":           state.Error,
	})
}

// RegisterAIAgentHandlers registers AI procurement agent endpoints.
func RegisterAIAgentHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// Initialize services
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// Services needed for AI agent
	localeSrv := service.NewLocaleSrv()
	translateSrv := service.NewTranslateSrv(dbm, cache)
	messageSrv := message.NewIMessageSrv(dbm)
	materialSrv := service.NewMaterialSrv(dbm, localeSrv, settingSrv, translateSrv, messageSrv)
	warehouseSrv := service.NewWarehouseSrv(dbm, settingSrv, materialSrv, translateSrv)
	purchaseOrderSrv := purchaseorder.NewPurchaseOrderSrv(dbm, settingSrv)

	// AI agent dependencies
	cfg := ai_agent.DefaultConfig()
	llmClient := ai_agent.NewLLMClient(cfg)

	// Build service dependencies for the agent
	deps := &ai_agent.NodeDeps{
		DBM:              dbm,
		MaterialSrv:      materialSrv,
		WarehouseSrv:     warehouseSrv,
		SupplierSrv:      service.NewSupplierSrv(dbm),
		PurchaseOrderSrv: purchaseOrderSrv,
		StatisticsSrv:    statisticsSrv,
		LLM:              llmClient,
		Config:           cfg,
	}

	agent := ai_agent.NewAgent(deps)

	wrapper := &AIAgentHandler{
		authSrv: authSrv,
		agent:   agent,
	}

	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/ai/procurement/analysis", wrapper.RunAnalysis)
		privateApi.POST("/ai/procurement/review", wrapper.SubmitReview)
		privateApi.GET("/ai/procurement/session", wrapper.GetSession)
	}
}
