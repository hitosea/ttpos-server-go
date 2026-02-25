package tools

import (
	"context"
	"fmt"
	"time"

	"ttpos-server-go/app/modules/agent/engine/tool"
)

// RegisterDemoTools 注册演示用的工具
func RegisterDemoTools(registry *tool.Registry) {
	registry.MustRegister(&GetCurrentTimeTool{})
	registry.MustRegister(&CalculatorTool{})
	registry.MustRegister(&QueryOrderTool{})
}

// =============================================================================
// GetCurrentTimeTool 获取当前时间
// =============================================================================

type GetCurrentTimeTool struct{}

func (t *GetCurrentTimeTool) Name() string {
	return "get_current_time"
}

func (t *GetCurrentTimeTool) Description() string {
	return "获取当前服务器时间，可指定时区。默认返回 Asia/Shanghai 时区的时间。"
}

func (t *GetCurrentTimeTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"timezone": map[string]any{
				"type":        "string",
				"description": "时区名称，例如 Asia/Shanghai、America/New_York",
				"default":     "Asia/Shanghai",
			},
			"format": map[string]any{
				"type":        "string",
				"description": "时间格式：datetime（完整日期时间）、date（仅日期）、time（仅时间）、unix（Unix时间戳）",
				"enum":        []string{"datetime", "date", "time", "unix"},
				"default":     "datetime",
			},
		},
	}
}

func (t *GetCurrentTimeTool) Execute(ctx context.Context, input map[string]any) (any, error) {
	timezone := "Asia/Shanghai"
	if tz, ok := input["timezone"].(string); ok && tz != "" {
		timezone = tz
	}

	format := "datetime"
	if f, ok := input["format"].(string); ok && f != "" {
		format = f
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %s", timezone)
	}

	now := time.Now().In(loc)

	result := map[string]any{
		"timezone": timezone,
	}

	switch format {
	case "date":
		result["date"] = now.Format("2006-01-02")
	case "time":
		result["time"] = now.Format("15:04:05")
	case "unix":
		result["unix"] = now.Unix()
	default:
		result["datetime"] = now.Format("2006-01-02 15:04:05")
	}

	return result, nil
}

// =============================================================================
// CalculatorTool 计算器
// =============================================================================

type CalculatorTool struct{}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "执行基本数学运算（加、减、乘、除）。用于需要精确计算的场景。"
}

func (t *CalculatorTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"description": "运算类型",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
			},
			"a": map[string]any{
				"type":        "number",
				"description": "第一个操作数",
			},
			"b": map[string]any{
				"type":        "number",
				"description": "第二个操作数",
			},
		},
		"required": []string{"operation", "a", "b"},
	}
}

func (t *CalculatorTool) Execute(ctx context.Context, input map[string]any) (any, error) {
	operation, _ := input["operation"].(string)
	a, _ := input["a"].(float64)
	b, _ := input["b"].(float64)

	var result float64
	var expr string

	switch operation {
	case "add":
		result = a + b
		expr = fmt.Sprintf("%.2f + %.2f", a, b)
	case "subtract":
		result = a - b
		expr = fmt.Sprintf("%.2f - %.2f", a, b)
	case "multiply":
		result = a * b
		expr = fmt.Sprintf("%.2f × %.2f", a, b)
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
		expr = fmt.Sprintf("%.2f ÷ %.2f", a, b)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return map[string]any{
		"expression": expr,
		"result":     result,
	}, nil
}

// =============================================================================
// QueryOrderTool 查询订单（模拟）
// =============================================================================

type QueryOrderTool struct{}

func (t *QueryOrderTool) Name() string {
	return "query_order"
}

func (t *QueryOrderTool) Description() string {
	return `查询订单信息。可以根据订单号查询单个订单，或根据时间范围、状态等条件查询订单列表。

支持的查询条件：
- order_no: 订单号（精确匹配，查询单个订单）
- status: 订单状态（pending-待支付, paid-已支付, completed-已完成, cancelled-已取消）
- limit: 返回数量限制，默认10，最大50

如果不提供任何条件，返回最近的订单列表。`
}

func (t *QueryOrderTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"order_no": map[string]any{
				"type":        "string",
				"description": "订单号，精确匹配查询单个订单",
			},
			"status": map[string]any{
				"type":        "string",
				"description": "订单状态筛选",
				"enum":        []string{"pending", "paid", "completed", "cancelled"},
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "返回数量限制",
				"default":     10,
				"minimum":     1,
				"maximum":     50,
			},
		},
	}
}

func (t *QueryOrderTool) Execute(ctx context.Context, input map[string]any) (any, error) {
	// 模拟订单数据
	mockOrders := []map[string]any{
		{
			"order_no":    "ORD20240225001",
			"status":      "completed",
			"total_amount": 128.50,
			"items_count": 3,
			"create_time": "2024-02-25 10:30:00",
			"customer":    "张三",
		},
		{
			"order_no":    "ORD20240225002",
			"status":      "paid",
			"total_amount": 256.00,
			"items_count": 5,
			"create_time": "2024-02-25 11:15:00",
			"customer":    "李四",
		},
		{
			"order_no":    "ORD20240225003",
			"status":      "pending",
			"total_amount": 88.00,
			"items_count": 2,
			"create_time": "2024-02-25 12:00:00",
			"customer":    "王五",
		},
		{
			"order_no":    "ORD20240225004",
			"status":      "completed",
			"total_amount": 512.80,
			"items_count": 8,
			"create_time": "2024-02-25 12:30:00",
			"customer":    "赵六",
		},
		{
			"order_no":    "ORD20240225005",
			"status":      "cancelled",
			"total_amount": 66.00,
			"items_count": 1,
			"create_time": "2024-02-25 13:00:00",
			"customer":    "钱七",
		},
	}

	// 如果指定了订单号，精确查询
	if orderNo, ok := input["order_no"].(string); ok && orderNo != "" {
		for _, order := range mockOrders {
			if order["order_no"] == orderNo {
				return map[string]any{
					"found": true,
					"order": order,
				}, nil
			}
		}
		return map[string]any{
			"found":   false,
			"message": fmt.Sprintf("订单 %s 不存在", orderNo),
		}, nil
	}

	// 按状态筛选
	var filtered []map[string]any
	status, hasStatus := input["status"].(string)

	for _, order := range mockOrders {
		if hasStatus && status != "" && order["status"] != status {
			continue
		}
		filtered = append(filtered, order)
	}

	// 限制数量
	limit := 10
	if l, ok := input["limit"].(float64); ok && l > 0 {
		limit = int(l)
		if limit > 50 {
			limit = 50
		}
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// 计算统计信息
	var totalAmount float64
	for _, order := range filtered {
		if amount, ok := order["total_amount"].(float64); ok {
			totalAmount += amount
		}
	}

	return map[string]any{
		"total_count":  len(filtered),
		"total_amount": totalAmount,
		"orders":       filtered,
	}, nil
}
