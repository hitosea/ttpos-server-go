package handler

import (
	"io"
	"net/http"
	"ttpos-ai/internal/service"
	"ttpos-ai/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService service.IChatService
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(chatService service.IChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Messages []service.ChatMessage `json:"messages" binding:"required,min=1"`
	Stream   bool                  `json:"stream"` // 是否流式响应
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Chat 处理聊天请求
// @Summary 聊天对话
// @Description 支持多轮对话，可选流式响应
// @Tags AI
// @Accept json
// @Produce json
// @Param request body ChatRequest true "聊天请求"
// @Success 200 {object} ChatResponse
// @Router /api/v1/chat [post]
func (h *ChatHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	if req.Stream {
		h.handleStreamChat(c, req.Messages)
		return
	}

	h.handleChat(c, req.Messages)
}

// handleChat 处理普通聊天
func (h *ChatHandler) handleChat(c *gin.Context, messages []service.ChatMessage) {
	result, err := h.chatService.Chat(c.Request.Context(), messages)
	if err != nil {
		logger.Logger.Error("聊天失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Code:    500,
			Message: "对话失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Code:    1,
		Message: "success",
		Data:    result,
	})
}

// handleStreamChat 处理流式聊天 (SSE)
func (h *ChatHandler) handleStreamChat(c *gin.Context, messages []service.ChatMessage) {
	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	stream, err := h.chatService.ChatStream(c.Request.Context(), messages)
	if err != nil {
		logger.Logger.Error("流式聊天失败", zap.Error(err))
		c.SSEvent("error", gin.H{"error": err.Error()})
		return
	}

	c.Stream(func(w io.Writer) bool {
		chunk, ok := <-stream
		if !ok {
			return false
		}

		if chunk.Error != "" {
			c.SSEvent("error", gin.H{"error": chunk.Error})
			return false
		}

		if chunk.Done {
			c.SSEvent("done", gin.H{"done": true})
			return false
		}

		c.SSEvent("message", gin.H{"content": chunk.Content})
		return true
	})
}

// SimpleChat 简单聊天
// @Summary 简单聊天
// @Description 单轮对话
// @Tags AI
// @Accept json
// @Produce json
// @Param request body SimpleChatRequest true "简单聊天请求"
// @Success 200 {object} ChatResponse
// @Router /api/v1/chat/simple [post]
func (h *ChatHandler) SimpleChat(c *gin.Context) {
	var req SimpleChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	var content string
	var err error

	if req.SystemPrompt != "" {
		content, err = h.chatService.SimpleChatWithSystem(c.Request.Context(), req.SystemPrompt, req.Prompt)
	} else {
		content, err = h.chatService.SimpleChat(c.Request.Context(), req.Prompt)
	}

	if err != nil {
		logger.Logger.Error("简单聊天失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Code:    500,
			Message: "对话失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Code:    1,
		Message: "success",
		Data: gin.H{
			"content": content,
		},
	})
}

// SimpleChatRequest 简单聊天请求
type SimpleChatRequest struct {
	Prompt       string `json:"prompt" binding:"required"`
	SystemPrompt string `json:"system_prompt"` // 可选的系统提示
}
