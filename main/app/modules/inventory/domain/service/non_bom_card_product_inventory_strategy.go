package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

// nonBomCardProductInventoryStrategy 无成本卡商品库存计算策略
type nonBomCardProductInventoryStrategy struct{}

// NewNonBomCardProductInventoryStrategy 创建无成本卡商品库存计算策略
func NewNonBomCardProductInventoryStrategy() IInventoryStrategy {
	return &nonBomCardProductInventoryStrategy{}
}

// CalculateInventory 计算无成本卡商品的库存
func (s *nonBomCardProductInventoryStrategy) CalculateInventory(
	ctx context.Context,
	productBom interface{},
) (float64, error) {
	// 类型断言（使用现有 Model）
	bom, ok := productBom.(*model.ProductBom)
	if !ok {
		return 0, errors.New("商品BOM类型错误")
	}

	// 1. 判断是否标记售罄
	if bom.IsSoldOut == constant.ProductStatusSaleOut {
		return 0, nil
	}

	// 2. 判断是否设置可售量
	if bom.SellableQuantity > 0 {
		return bom.SellableQuantity, nil
	}

	// 3. 返回无限库存
	return constant.ProductBomInfiniteStock, nil
}

