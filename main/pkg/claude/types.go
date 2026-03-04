package claude

import "encoding/json"

// MessageRequest Claude API 请求
type MessageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
	Stream    bool      `json:"stream,omitempty"`
}

// Message 消息
type Message struct {
	Role    string         `json:"role"` // user/assistant
	Content []ContentBlock `json:"content"`
}

// NewTextMessage 创建文本消息
func NewTextMessage(role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentBlock{
			{Type: "text", Text: text},
		},
	}
}

// ContentBlock 内容块
type ContentBlock struct {
	Type string `json:"type"` // text/tool_use/tool_result
	Text string `json:"text,omitempty"`

	// tool_use 字段
	Id    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// tool_result 字段
	ToolUseId string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

// MessageResponse 响应
type MessageResponse struct {
	Id           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"` // end_turn/tool_use/max_tokens
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// GetTextContent 获取文本内容
func (r *MessageResponse) GetTextContent() string {
	for _, block := range r.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

// GetToolUseBlocks 获取工具调用块
func (r *MessageResponse) GetToolUseBlocks() []ContentBlock {
	var blocks []ContentBlock
	for _, block := range r.Content {
		if block.Type == "tool_use" {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

// Usage Token 使用情况
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// APIError API 错误响应
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}
