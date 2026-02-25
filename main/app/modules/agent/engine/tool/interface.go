package tool

import "context"

// Tool 定义了一个可被 AI 调用的工具接口
type Tool interface {
	// Name 返回工具名称，用于 Claude API 注册
	Name() string

	// Description 返回工具描述，帮助 AI 理解何时使用此工具
	Description() string

	// InputSchema 返回 JSON Schema，定义工具的输入参数
	InputSchema() map[string]any

	// Execute 执行工具逻辑
	// ctx: 包含用户信息、权限、数据库连接等
	// input: AI 传入的参数（已解析的 map）
	// 返回: 执行结果（会序列化后返回给 AI）和错误
	Execute(ctx context.Context, input map[string]any) (any, error)
}

// ExecuteResult 工具执行结果
type ExecuteResult struct {
	ToolName string `json:"tool_name"`
	Input    any    `json:"input"`
	Output   any    `json:"output"`
	Error    string `json:"error,omitempty"`
	Success  bool   `json:"success"`
}
