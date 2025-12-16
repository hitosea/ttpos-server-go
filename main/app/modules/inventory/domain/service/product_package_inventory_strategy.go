package service

import (
	"ttpos-server-go/pkg/context"
)

// IProductPackageInventoryStrategy 商品包库存计算策略接口
// 用于计算商品包下所有BOM库存的聚合方式（最小值或最大值）
type IProductPackageInventoryStrategy interface {
	// CalculatePackageInventory 计算商品包库存
	// inventories: 商品包下所有BOM的库存列表
	// 返回: 聚合后的库存数量（float64）
	CalculatePackageInventory(ctx context.Context, inventories []float64) (float64, error)
}
