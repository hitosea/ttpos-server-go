package resp

import "ttpos-server-go/app/modules/agent/engine/tool"

// ChatResp 对话响应
type ChatResp struct {
	Content      string               `json:"content"`                 // AI 回复内容
	ToolCalls    []tool.ExecuteResult `json:"tool_calls,omitempty"`    // 工具调用记录
	FinishReason string               `json:"finish_reason,omitempty"` // 结束原因
	Usage        *UsageInfo           `json:"usage,omitempty"`         // Token 使用情况
}

// UsageInfo Token 使用信息
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
