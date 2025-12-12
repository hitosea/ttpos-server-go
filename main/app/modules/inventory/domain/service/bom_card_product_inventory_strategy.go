package service

import (
	"math"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

// bomCardProductInventoryStrategy 有成本卡商品库存计算策略
type bomCardProductInventoryStrategy struct{}

// NewBomCardProductInventoryStrategy 创建有成本卡商品库存计算策略
func NewBomCardProductInventoryStrategy() IInventoryStrategy {
	return &bomCardProductInventoryStrategy{}
}

// CalculateInventory 计算有成本卡商品的库存
func (s *bomCardProductInventoryStrategy) CalculateInventory(
	ctx context.Context,
	productBom interface{},
) (float64, error) {
	// 类型断言（使用现有 Model）
	bom, ok := productBom.(*model.ProductBom)
	if !ok {
		return 0, errors.New("商品BOM类型错误")
	}

	// 1. 判断是否开启成本卡控制
	if bom.UseBomCardStock == constant.Yes {
		// 2. 根据成本卡计算材料用量得到库存
		if bom.ProductBomCard == nil {
			return constant.ProductBomInfiniteStock, errors.WithMessage(errors.New("成本卡未加载"))
		}

		// 使用现有的 CalculateExpectedProductionNum 方法
		inventory := bom.ProductBomCard.CalculateExpectedProductionNum()
		// 确保返回非负数
		return math.Max(0, inventory), nil
	}

	// 3. 成本卡控制未开启，执行无成本卡商品的逻辑
	return s.calculateNonBomCardInventory(bom)
}

// calculateNonBomCardInventory 计算无成本卡商品的库存（内部方法）
func (s *bomCardProductInventoryStrategy) calculateNonBomCardInventory(
	bom *model.ProductBom,
) (float64, error) {
	// 判断是否标记售罄
	if bom.IsSoldOut == constant.ProductStatusSaleOut {
		return 0, nil
	}

	// 特殊需求: 商品绑定了成本卡,当成本卡中的材料不允许负库存时,一定要求材料库存不能负. 所以只要成本卡中的材料有一个不允许负库存,则返回成本卡计算的库存值
	if bom.HasProductBomCard() && bom.ProductBomCard != nil {
		for _, material := range bom.ProductBomCard.RelatedMaterials {
			if material.Material.AllowNegativeStock == constant.No {
				return bom.ProductBomCard.CalculateExpectedProductionNum(), nil
			}
		}
	}

	// 判断是否设置可售量
	if bom.IsOpenStockBool() {
		return bom.StockNum, nil
	}

	// 返回无限库存
	return constant.ProductBomInfiniteStock, nil
}
