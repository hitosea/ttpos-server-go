package service

import (
	"ttpos-server-go/pkg/context"
)

// IInventoryStrategy 库存计算策略接口
// 使用策略模式处理不同商品类型的库存计算逻辑
type IInventoryStrategy interface {
	// CalculateInventory 计算库存
	// productBom: 商品BOM（使用现有 Model，类型为 interface{}）
	// 返回: 库存数量（float64），无限库存返回 99999999
	CalculateInventory(ctx context.Context, productBom interface{}) (float64, error)
}
