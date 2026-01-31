// Package grab 定义 GrabFood API 交互的数据结构
package grab

import (
	"ttpos-api/ttpos-takeout/message"
)

// ListOrdersResponse Grab ListOrders API 响应
// 使用 TakeoutOrder 以支持 Price.Total 字段（SDK 缺失该字段）
type ListOrdersResponse struct {
	// Orders 订单列表
	// 使用 message.TakeoutOrder 以获取完整的 Price 结构（包含 Total）
	Orders []message.TakeoutOrder `json:"orders"`

	// More 是否有更多数据
	More bool `json:"more"`
}
