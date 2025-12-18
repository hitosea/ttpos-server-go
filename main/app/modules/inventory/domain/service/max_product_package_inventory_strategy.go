package service

import (
	"math"
	"ttpos-server-go/pkg/context"
)

// maxProductPackageInventoryStrategy 商品包库存最大值策略
// 商品包库存等于该商品包下所有商品BOM库存中的最大值
type maxProductPackageInventoryStrategy struct{}

// NewMaxProductPackageInventoryStrategy 创建最大值策略
func NewMaxProductPackageInventoryStrategy() IProductPackageInventoryStrategy {
	return &maxProductPackageInventoryStrategy{}
}

// CalculatePackageInventory 计算商品包库存（取最大值）
// 注意：会忽略0值库存，只计算非零库存的最大值
// 如果所有库存都是0，则返回0
func (s *maxProductPackageInventoryStrategy) CalculatePackageInventory(
	ctx context.Context,
	inventories []float64,
) (float64, error) {
	maxInventory := 0.0
	for _, inventory := range inventories {
		if inventory != 0 {
			maxInventory = math.Max(maxInventory, inventory)
		}
	}
	return maxInventory, nil
}
