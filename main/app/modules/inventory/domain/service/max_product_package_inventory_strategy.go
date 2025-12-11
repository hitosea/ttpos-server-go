package service

import (
	"math"
	"ttpos-server-go/app/errors"
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
func (s *maxProductPackageInventoryStrategy) CalculatePackageInventory(
	ctx context.Context,
	inventories []float64,
) (float64, error) {
	if len(inventories) == 0 {
		return 0, errors.New("库存列表为空")
	}

	var maxInventory float64 = math.SmallestNonzeroFloat64
	var hasValidInventory bool

	for _, inventory := range inventories {
		if inventory > maxInventory {
			maxInventory = inventory
			hasValidInventory = true
		}
	}

	if !hasValidInventory {
		return 0, errors.New("无法计算商品包库存：所有BOM库存查询失败")
	}

	return maxInventory, nil
}
