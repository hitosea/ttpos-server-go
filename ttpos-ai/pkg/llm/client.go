package llm

import (
	"context"
	"fmt"
	"ttpos-ai/config"
	"ttpos-ai/pkg/logger"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// Client LLM客户端封装
type Client struct {
	chatModel model.ChatModel
	config    *config.LLMConf
}

// NewClient 创建新的LLM客户端
func NewClient() (*Client, error) {
	cfg := &config.LLM
	if !cfg.Enabled {
		return nil, fmt.Errorf("LLM功能未启用")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("LLM API密钥未配置")
	}

	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		MaxTokens:   &cfg.MaxTokens,
		Temperature: &cfg.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("创建ChatModel失败: %w", err)
	}

	return &Client{
		chatModel: chatModel,
		config:    cfg,
	}, nil
}

// ChatRequest 对话请求
type ChatRequest struct {
	Messages    []Message // 对话消息列表
	MaxTokens   *int      // 可选：覆盖默认max_tokens
	Temperature *float32  // 可选：覆盖默认temperature
}

// Message 对话消息
type Message struct {
	Role    string // user/assistant/system
	Content string // 消息内容
}

// ChatResponse 对话响应
type ChatResponse struct {
	Content      string // 响应内容
	FinishReason string // 结束原因
	Usage        *Usage // token使用情况
}

// Usage token使用统计
type Usage struct {
	PromptTokens     int // 输入token数
	CompletionTokens int // 输出token数
	TotalTokens      int // 总token数
}

// Chat 执行对话（非流式）
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	messages := make([]*schema.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		var role schema.RoleType
		switch msg.Role {
		case "system":
			role = schema.System
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		default:
			role = schema.User
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	opts := make([]model.Option, 0)
	if req.MaxTokens != nil {
		opts = append(opts, model.WithMaxTokens(*req.MaxTokens))
	}
	if req.Temperature != nil {
		opts = append(opts, model.WithTemperature(*req.Temperature))
	}

	result, err := c.chatModel.Generate(ctx, messages, opts...)
	if err != nil {
		logger.Logger.Error("LLM对话失败", zap.Error(err))
		return nil, fmt.Errorf("LLM对话失败: %w", err)
	}

	resp := &ChatResponse{
		Content:      result.Content,
		FinishReason: string(result.ResponseMeta.FinishReason),
	}

	if result.ResponseMeta.Usage != nil {
		resp.Usage = &Usage{
			PromptTokens:     result.ResponseMeta.Usage.PromptTokens,
			CompletionTokens: result.ResponseMeta.Usage.CompletionTokens,
			TotalTokens:      result.ResponseMeta.Usage.TotalTokens,
		}
	}

	return resp, nil
}

// StreamChunk 流式响应块
type StreamChunk struct {
	Content string // 内容片段
	Done    bool   // 是否结束
	Error   error  // 错误信息
}

// ChatStream 执行流式对话
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	messages := make([]*schema.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		var role schema.RoleType
		switch msg.Role {
		case "system":
			role = schema.System
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		default:
			role = schema.User
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	opts := make([]model.Option, 0)
	if req.MaxTokens != nil {
		opts = append(opts, model.WithMaxTokens(*req.MaxTokens))
	}
	if req.Temperature != nil {
		opts = append(opts, model.WithTemperature(*req.Temperature))
	}

	stream, err := c.chatModel.Stream(ctx, messages, opts...)
	if err != nil {
		logger.Logger.Error("LLM流式对话失败", zap.Error(err))
		return nil, fmt.Errorf("LLM流式对话失败: %w", err)
	}

	chunkChan := make(chan *StreamChunk)
	go func() {
		defer close(chunkChan)
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err.Error() != "EOF" {
					chunkChan <- &StreamChunk{Error: err}
				}
				chunkChan <- &StreamChunk{Done: true}
				return
			}
			chunkChan <- &StreamChunk{Content: msg.Content}
		}
	}()

	return chunkChan, nil
}

// SimpleChat 简单对话（单轮）
func (c *Client) SimpleChat(ctx context.Context, prompt string) (string, error) {
	resp, err := c.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// SimpleChatWithSystem 带系统提示的简单对话
func (c *Client) SimpleChatWithSystem(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	resp, err := c.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GetModel 获取当前使用的模型名称
func (c *Client) GetModel() string {
	return c.config.Model
}

// IsEnabled 检查LLM功能是否启用
func IsEnabled() bool {
	return config.LLM.Enabled
}
