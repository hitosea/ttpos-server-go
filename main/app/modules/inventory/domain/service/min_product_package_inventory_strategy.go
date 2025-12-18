package service

import (
	"math"
	"ttpos-server-go/pkg/context"
)

// minProductPackageInventoryStrategy 商品包库存最小值策略
// 商品包库存等于该商品包下所有商品BOM库存中的最小值
type minProductPackageInventoryStrategy struct{}

// NewMinProductPackageInventoryStrategy 创建最小值策略
func NewMinProductPackageInventoryStrategy() IProductPackageInventoryStrategy {
	return &minProductPackageInventoryStrategy{}
}

// CalculatePackageInventory 计算商品包库存（取最小值）
func (s *minProductPackageInventoryStrategy) CalculatePackageInventory(
	ctx context.Context,
	inventories []float64,
) (float64, error) {
	if len(inventories) == 0 {
		return 0, nil
	}

	minInventory := inventories[0]
	for _, inventory := range inventories[1:] {
		minInventory = math.Min(minInventory, inventory)
	}
	return minInventory, nil
}
