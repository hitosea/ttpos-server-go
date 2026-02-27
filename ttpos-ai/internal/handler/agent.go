package handler

import (
	"net/http"
	"ttpos-ai/internal/service"
	"ttpos-ai/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentHandler Agent处理器
type AgentHandler struct {
	agentService service.IAgentService
}

// NewAgentHandler 创建Agent处理器
func NewAgentHandler(agentService service.IAgentService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

// QueryRequest 查询请求
type QueryRequest struct {
	Question string                 `json:"question" binding:"required"` // 问题
	Messages []service.ChatMessage `json:"messages"`                     // 历史消息（可选）
}

// QueryResponse 查询响应
type QueryResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Query 处理查询请求
// @Summary 自然语言查询数据库
// @Description 将自然语言问题转换为SQL查询，并返回结果
// @Tags Agent
// @Accept json
// @Produce json
// @Param request body QueryRequest true "查询请求"
// @Success 200 {object} QueryResponse
// @Router /api/v1/agent/query [post]
func (h *AgentHandler) Query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, QueryResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	var result *service.AgentResult
	var err error

	if len(req.Messages) > 0 {
		// 带历史消息的查询
		result, err = h.agentService.QueryWithHistory(c.Request.Context(), req.Messages)
	} else {
		// 单轮查询
		result, err = h.agentService.Query(c.Request.Context(), req.Question)
	}

	if err != nil {
		logger.Logger.Error("Agent查询失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, QueryResponse{
			Code:    500,
			Message: "查询失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, QueryResponse{
		Code:    1,
		Message: "success",
		Data:    result,
	})
}
