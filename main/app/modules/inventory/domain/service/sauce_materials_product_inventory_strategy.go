package service

import (
	"math"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"github.com/shopspring/decimal"
)

// sauceMaterialsProductInventoryStrategy 小料材料商品库存计算策略
// 直接通过关联的材料（SauceMaterials）进行计算库存
type sauceMaterialsProductInventoryStrategy struct{}

// NewSauceMaterialsProductInventoryStrategy 创建小料材料商品库存计算策略
func NewSauceMaterialsProductInventoryStrategy() IInventoryStrategy {
	return &sauceMaterialsProductInventoryStrategy{}
}

// CalculateInventory 计算小料材料商品的库存
// 遍历 ProductSauce.SauceMaterials，对每个材料计算可生产数量，然后取最小值
func (s *sauceMaterialsProductInventoryStrategy) CalculateInventory(
	ctx context.Context,
	productBom interface{},
) (float64, error) {
	// 类型断言（使用现有 Model）
	bom, ok := productBom.(*model.ProductBom)
	if !ok {
		return 0, errors.New("商品BOM类型错误")
	}

	// 1. 检查是否为小料
	if !bom.IsSauce() {
		return 0, errors.New("不是小料商品")
	}

	// 2. 检查 ProductSauce 是否已加载
	if bom.ProductSauce.Uuid == 0 {
		return constant.ProductBomInfiniteStock, errors.New("小料信息未加载")
	}

	// 3. 检查是否有 SauceMaterials
	if len(bom.ProductSauce.SauceMaterials) == 0 {
		return constant.ProductBomInfiniteStock, errors.New("小料材料未加载")
	}

	// 4. 判断是否标记售罄
	if bom.IsSoldOut == constant.ProductStatusSaleOut {
		return 0, nil
	}

	// 5. 遍历所有 SauceMaterials，计算每个材料的可生产数量，取最小值
	// 参考 ProductBom.GetStockNum 中的逻辑：使用 material.Material.GetStockNum() 和 material.GetDecreaseNum(1)
	var minExpectedProductionNum float64 = constant.ProductBomInfiniteStock
	hasValidMaterial := false

	for _, material := range bom.ProductSauce.SauceMaterials {
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

	// 6. 如果没有有效的材料，返回无限库存
	if !hasValidMaterial {
		return constant.ProductBomInfiniteStock, nil
	}

	return math.Max(0, minExpectedProductionNum), nil
}
