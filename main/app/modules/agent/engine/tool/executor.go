package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

// Executor 工具执行器
type Executor struct {
	registry *Registry
}

// NewExecutor 创建工具执行器
func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

// Execute 执行工具
func (e *Executor) Execute(ctx context.Context, toolName string, inputJSON json.RawMessage) *ExecuteResult {
	result := &ExecuteResult{
		ToolName: toolName,
		Success:  false,
	}

	// 获取工具
	tool, ok := e.registry.Get(toolName)
	if !ok {
		result.Error = fmt.Sprintf("tool %s not found", toolName)
		return result
	}

	// 解析输入参数
	var input map[string]any
	if len(inputJSON) > 0 {
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			result.Error = fmt.Sprintf("parse input: %v", err)
			return result
		}
	} else {
		input = make(map[string]any)
	}
	result.Input = input

	// 执行工具
	output, err := tool.Execute(ctx, input)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Output = output
	result.Success = true
	return result
}

// ExecuteBatch 批量执行工具
func (e *Executor) ExecuteBatch(ctx context.Context, calls []ToolCall) []*ExecuteResult {
	results := make([]*ExecuteResult, len(calls))
	for i, call := range calls {
		results[i] = e.Execute(ctx, call.Name, call.Input)
	}
	return results
}

// ToolCall 工具调用信息
type ToolCall struct {
	Id    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// FormatResultForClaude 将执行结果格式化为返回给 Claude 的字符串
func FormatResultForClaude(result *ExecuteResult) string {
	if !result.Success {
		return fmt.Sprintf("Error: %s", result.Error)
	}

	output, err := json.Marshal(result.Output)
	if err != nil {
		return fmt.Sprintf("Error serializing output: %v", err)
	}
	return string(output)
}
