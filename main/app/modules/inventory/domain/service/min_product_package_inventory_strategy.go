package service

import (
	"math"
	"ttpos-server-go/app/errors"
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
		return 0, errors.New("库存列表为空")
	}

	var minInventory float64 = math.MaxFloat64
	var hasValidInventory bool

	for _, inventory := range inventories {
		if inventory < minInventory {
			minInventory = inventory
			hasValidInventory = true
		}
	}

	if !hasValidInventory {
		return 0, errors.New("无法计算商品包库存：所有BOM库存查询失败")
	}

	return minInventory, nil
}
