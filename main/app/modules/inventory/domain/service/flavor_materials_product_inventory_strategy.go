package service

import (
	"math"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"github.com/shopspring/decimal"
)

// flavorMaterialsProductInventoryStrategy 规格材料商品库存计算策略
// 直接通过关联的材料（FlavorMaterials）进行计算库存
type flavorMaterialsProductInventoryStrategy struct{}

// NewFlavorMaterialsProductInventoryStrategy 创建规格材料商品库存计算策略
func NewFlavorMaterialsProductInventoryStrategy() IInventoryStrategy {
	return &flavorMaterialsProductInventoryStrategy{}
}

// CalculateInventory 计算规格材料商品的库存
// 遍历 FlavorMaterials，对每个材料计算可生产数量，然后取最小值
func (s *flavorMaterialsProductInventoryStrategy) CalculateInventory(
	ctx context.Context,
	productBom interface{},
) (float64, error) {
	// 类型断言（使用现有 Model）
	bom, ok := productBom.(*model.ProductBom)
	if !ok {
		return 0, errors.New("商品BOM类型错误")
	}

	// 1. 检查是否有 FlavorMaterials
	if len(bom.FlavorMaterials) == 0 {
		return constant.ProductBomInfiniteStock, errors.New("规格材料未加载")
	}

	// 2. 判断是否标记售罄
	if bom.IsSoldOut == constant.ProductStatusSaleOut {
		return 0, nil
	}

	// 3. 遍历所有 FlavorMaterials，计算每个材料的可生产数量，取最小值
	// 参考 ProductBom.GetStockNum 中的逻辑：使用 material.Material.GetStockNum() 和 material.GetDecreaseNum(1)
	var minExpectedProductionNum float64 = constant.ProductBomInfiniteStock
	hasValidMaterial := false

	for _, material := range bom.FlavorMaterials {
		// 检查材料是否已加载
		if material.Material == nil {
			continue
		}

		// 获取材料库存数量
		stockNum := material.Material.GetStockNum()
		if stockNum <= 0 {
			continue
		}

		// 获取生产1个商品需要的材料数量
		num := material.GetDecreaseNum(1)
		if num <= 0 {
			continue
		}

		// 计算可生产数量：材料库存数量 / 生产1个商品需要的材料数量
		expectedProductionNum := decimal.NewFromFloat(stockNum).Div(decimal.NewFromFloat(num)).Truncate(4).InexactFloat64()
		if expectedProductionNum < minExpectedProductionNum {
			minExpectedProductionNum = expectedProductionNum
			hasValidMaterial = true
		}
	}

	// 4. 如果没有有效的材料，返回无限库存
	if !hasValidMaterial {
		return constant.ProductBomInfiniteStock, nil
	}

	return math.Max(0, minExpectedProductionNum), nil
}
