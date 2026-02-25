package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/modules/agent/dto/req"
	"ttpos-server-go/app/modules/agent/service"
)

// AgentApi Agent API 处理器
type AgentApi struct {
	agentSrv service.IAgentSrv
}

// NewAgentApi 创建 Agent API
func NewAgentApi(agentSrv service.IAgentSrv) *AgentApi {
	return &AgentApi{agentSrv: agentSrv}
}

// Chat 对话接口
// @Summary AI 对话
// @Description 与 AI 助手进行对话，AI 会根据需要调用工具完成任务
// @Tags AI Agent
// @Accept json
// @Produce json
// @Param request body req.ChatReq true "对话请求"
// @Success 200 {object} resp.ChatResp
// @Router /api/v1/ai/chat [post]
func (a *AgentApi) Chat(c *gin.Context) {
	var chatReq req.ChatReq
	if err := c.ShouldBindJSON(&chatReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    0,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	result, err := a.agentSrv.Chat(c.Request.Context(), chatReq.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": "chat error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "success",
		"data":    result,
	})
}

// Health 健康检查
func (a *AgentApi) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "ok",
	})
}
