package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

// sauceNonBomCardProductInventoryStrategy 无成本卡无材料小料商品库存计算策略
// 用于处理既没有成本卡也没有关联材料的小料
type sauceNonBomCardProductInventoryStrategy struct{}

// NewSauceNonBomCardProductInventoryStrategy 创建无成本卡无材料小料商品库存计算策略
func NewSauceNonBomCardProductInventoryStrategy() IInventoryStrategy {
	return &sauceNonBomCardProductInventoryStrategy{}
}

// CalculateInventory 计算无成本卡无材料小料商品的库存
// 参考 non_bom_card 策略，但专门用于小料，检查 ProductSauce.SauceMaterials
func (s *sauceNonBomCardProductInventoryStrategy) CalculateInventory(
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

	// 3. 确认没有成本卡和关联材料（这是该策略的前提条件）
	if bom.ProductSauce.HasProductBomCard() {
		return constant.ProductBomInfiniteStock, errors.New("小料有成本卡，应使用 sauce_bom_card 策略")
	}
	if len(bom.ProductSauce.SauceMaterials) > 0 {
		return constant.ProductBomInfiniteStock, errors.New("小料有关联材料，应使用 sauce_materials 策略")
	}

	// 4. 判断是否标记售罄
	if bom.IsSoldOut == constant.ProductStatusSaleOut {
		return 0, nil
	}

	// 5. 判断是否设置可售量
	if bom.IsOpenStockBool() {
		return bom.StockNum, nil
	}

	// 6. 返回无限库存
	return constant.ProductBomInfiniteStock, nil
}
