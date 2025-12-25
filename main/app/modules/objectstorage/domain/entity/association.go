package entity

import "context"

// Association 关联配置实体
type Association struct {
	// Path 关联路径，支持嵌套，如 "SaleBillSetting"、"SaleOrders.SaleOrderProducts.ProductPackage"
	Path string

	// ObjectType 对象类型，用于构建缓存 key
	ObjectType string

	// GetUUID 从对象中提取 UUID 的函数
	GetUUID func(obj interface{}) uint64

	// QueryFunc 单个对象查询函数
	QueryFunc func(ctx context.Context, uuid uint64) (interface{}, error)

	// BatchQueryFunc 批量查询函数（可选，用于性能优化）
	// 返回 map[uint64]interface{}，key 为 UUID，value 为对象
	BatchQueryFunc func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error)
}
