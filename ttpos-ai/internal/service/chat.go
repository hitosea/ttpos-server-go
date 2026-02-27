package service

import (
	"context"
	"fmt"
	"ttpos-ai/pkg/llm"
	"ttpos-ai/pkg/logger"

	"go.uber.org/zap"
)

// IChatService 聊天服务接口
type IChatService interface {
	Chat(ctx context.Context, messages []ChatMessage) (*ChatResult, error)
	ChatStream(ctx context.Context, messages []ChatMessage) (<-chan *StreamChunk, error)
	SimpleChat(ctx context.Context, prompt string) (string, error)
	SimpleChatWithSystem(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// ChatService 聊天服务实现
type ChatService struct {
	client *llm.Client
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // user/assistant/system
	Content string `json:"content"` // 消息内容
}

// ChatResult 聊天结果
type ChatResult struct {
	Content      string     `json:"content"`       // 响应内容
	FinishReason string     `json:"finish_reason"` // 结束原因
	Usage        *ChatUsage `json:"usage"`         // token使用情况
}

// ChatUsage token使用统计
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 输入token数
	CompletionTokens int `json:"completion_tokens"` // 输出token数
	TotalTokens      int `json:"total_tokens"`      // 总token数
}

// StreamChunk 流式响应块
type StreamChunk struct {
	Content string `json:"content,omitempty"` // 内容片段
	Done    bool   `json:"done"`              // 是否结束
	Error   string `json:"error,omitempty"`   // 错误信息
}

// NewChatService 创建聊天服务
func NewChatService() (IChatService, error) {
	if !llm.IsEnabled() {
		return nil, fmt.Errorf("LLM功能未启用")
	}

	client, err := llm.NewClient()
	if err != nil {
		return nil, err
	}

	return &ChatService{
		client: client,
	}, nil
}

// Chat 执行多轮对话
func (s *ChatService) Chat(ctx context.Context, messages []ChatMessage) (*ChatResult, error) {
	logger.Logger.Info("执行LLM对话", zap.Int("message_count", len(messages)))

	llmMessages := make([]llm.Message, 0, len(messages))
	for _, msg := range messages {
		llmMessages = append(llmMessages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	resp, err := s.client.Chat(ctx, &llm.ChatRequest{
		Messages: llmMessages,
	})
	if err != nil {
		logger.Logger.Error("LLM对话失败", zap.Error(err))
		return nil, fmt.Errorf("对话失败: %w", err)
	}

	result := &ChatResult{
		Content:      resp.Content,
		FinishReason: resp.FinishReason,
	}

	if resp.Usage != nil {
		result.Usage = &ChatUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
		logger.Logger.Info("LLM对话完成", zap.Int("total_tokens", resp.Usage.TotalTokens))
	}

	return result, nil
}

// ChatStream 执行流式对话
func (s *ChatService) ChatStream(ctx context.Context, messages []ChatMessage) (<-chan *StreamChunk, error) {
	logger.Logger.Info("执行流式LLM对话", zap.Int("message_count", len(messages)))

	llmMessages := make([]llm.Message, 0, len(messages))
	for _, msg := range messages {
		llmMessages = append(llmMessages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	llmStream, err := s.client.ChatStream(ctx, &llm.ChatRequest{
		Messages: llmMessages,
	})
	if err != nil {
		logger.Logger.Error("流式LLM对话失败", zap.Error(err))
		return nil, fmt.Errorf("流式对话失败: %w", err)
	}

	// 转换为服务层的StreamChunk
	chunkChan := make(chan *StreamChunk)
	go func() {
		defer close(chunkChan)
		for chunk := range llmStream {
			sc := &StreamChunk{
				Content: chunk.Content,
				Done:    chunk.Done,
			}
			if chunk.Error != nil {
				sc.Error = chunk.Error.Error()
			}
			chunkChan <- sc
		}
	}()

	return chunkChan, nil
}

// SimpleChat 简单对话
func (s *ChatService) SimpleChat(ctx context.Context, prompt string) (string, error) {
	logger.Logger.Info("执行简单LLM对话")

	content, err := s.client.SimpleChat(ctx, prompt)
	if err != nil {
		logger.Logger.Error("简单LLM对话失败", zap.Error(err))
		return "", fmt.Errorf("对话失败: %w", err)
	}

	return content, nil
}

// SimpleChatWithSystem 带系统提示的简单对话
func (s *ChatService) SimpleChatWithSystem(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger.Logger.Info("执行带系统提示的LLM对话")

	content, err := s.client.SimpleChatWithSystem(ctx, systemPrompt, userPrompt)
	if err != nil {
		logger.Logger.Error("带系统提示的LLM对话失败", zap.Error(err))
		return "", fmt.Errorf("对话失败: %w", err)
	}

	return content, nil
}
