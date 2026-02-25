package service

import (
	"context"
	"fmt"

	"ttpos-server-go/app/modules/agent/dto/resp"
	"ttpos-server-go/app/modules/agent/engine/tool"
	"ttpos-server-go/pkg/claude"
)

const (
	// MaxToolIterations 最大工具调用轮次，防止无限循环
	MaxToolIterations = 10
)

// SystemPrompt 系统提示词
const SystemPrompt = `你是 TTPOS 餐饮系统的智能助手。你可以帮助用户：
- 查询订单信息
- 获取当前时间
- 进行数学计算

请根据用户的需求，调用合适的工具完成任务。回复时使用中文，保持专业友好。
如果用户的请求不在你的能力范围内，请友好地告知。
当你调用工具获取到结果后，请用自然语言总结结果并回复用户。`

// IAgentSrv Agent 服务接口
type IAgentSrv interface {
	// Chat 执行一次对话
	Chat(ctx context.Context, message string) (*resp.ChatResp, error)
}

// AgentSrv Agent 服务实现
type AgentSrv struct {
	client   *claude.Client
	registry *tool.Registry
	executor *tool.Executor
}

// NewAgentSrv 创建 Agent 服务
func NewAgentSrv(client *claude.Client, registry *tool.Registry) *AgentSrv {
	return &AgentSrv{
		client:   client,
		registry: registry,
		executor: tool.NewExecutor(registry),
	}
}

// Chat 执行一次对话（支持多轮工具调用）
func (s *AgentSrv) Chat(ctx context.Context, message string) (*resp.ChatResp, error) {
	result := &resp.ChatResp{
		ToolCalls: make([]tool.ExecuteResult, 0),
	}

	// 初始化消息列表
	messages := []claude.Message{
		claude.NewTextMessage("user", message),
	}

	// 准备请求
	req := &claude.MessageRequest{
		System:   SystemPrompt,
		Messages: messages,
		Tools:    s.registry.ToClaudeTools(),
	}

	var totalInputTokens, totalOutputTokens int

	// 循环处理，直到 AI 完成回复或达到最大轮次
	for i := 0; i < MaxToolIterations; i++ {
		// 调用 Claude API
		response, err := s.client.CreateMessage(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("call claude api: %w", err)
		}

		totalInputTokens += response.Usage.InputTokens
		totalOutputTokens += response.Usage.OutputTokens

		// 如果是正常结束，返回结果
		if response.StopReason == "end_turn" {
			result.Content = response.GetTextContent()
			result.FinishReason = response.StopReason
			result.Usage = &resp.UsageInfo{
				InputTokens:  totalInputTokens,
				OutputTokens: totalOutputTokens,
				TotalTokens:  totalInputTokens + totalOutputTokens,
			}
			return result, nil
		}

		// 如果需要调用工具
		if response.StopReason == "tool_use" {
			// 将 assistant 的响应加入消息历史
			req.Messages = append(req.Messages, claude.Message{
				Role:    "assistant",
				Content: response.Content,
			})

			// 执行所有工具调用
			toolResults := make([]claude.ContentBlock, 0)
			toolUseBlocks := response.GetToolUseBlocks()

			for _, block := range toolUseBlocks {
				// 执行工具
				execResult := s.executor.Execute(ctx, block.Name, block.Input)
				result.ToolCalls = append(result.ToolCalls, *execResult)

				// 构建 tool_result 返回给 Claude
				toolResults = append(toolResults, claude.ContentBlock{
					Type:      "tool_result",
					ToolUseId: block.Id,
					Content:   tool.FormatResultForClaude(execResult),
					IsError:   !execResult.Success,
				})
			}

			// 将工具结果加入消息历史
			req.Messages = append(req.Messages, claude.Message{
				Role:    "user",
				Content: toolResults,
			})

			continue
		}

		// 其他停止原因（如 max_tokens）
		result.Content = response.GetTextContent()
		result.FinishReason = response.StopReason
		result.Usage = &resp.UsageInfo{
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
			TotalTokens:  totalInputTokens + totalOutputTokens,
		}
		return result, nil
	}

	return nil, fmt.Errorf("exceeded max tool iterations (%d)", MaxToolIterations)
}
