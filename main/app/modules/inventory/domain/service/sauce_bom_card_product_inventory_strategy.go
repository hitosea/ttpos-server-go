package service

import (
	"math"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

// sauceBomCardProductInventoryStrategy 小料成本卡商品库存计算策略
// 通过 ProductSauce 的成本卡（ProductBomCard）进行计算库存
type sauceBomCardProductInventoryStrategy struct{}

// NewSauceBomCardProductInventoryStrategy 创建小料成本卡商品库存计算策略
func NewSauceBomCardProductInventoryStrategy() IInventoryStrategy {
	return &sauceBomCardProductInventoryStrategy{}
}

// CalculateInventory 计算小料成本卡商品的库存
// 使用 ProductSauce.ProductBomCard.CalculateExpectedProductionNum() 方法计算库存
func (s *sauceBomCardProductInventoryStrategy) CalculateInventory(
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

	// 3. 检查是否有成本卡
	if !bom.ProductSauce.HasProductBomCard() {
		return constant.ProductBomInfiniteStock, errors.New("小料成本卡未加载")
	}

	// 4. 检查成本卡是否已加载
	if bom.ProductSauce.ProductBomCard == nil {
		return constant.ProductBomInfiniteStock, errors.New("小料成本卡未加载")
	}

	// 5. 判断是否标记售罄
	if bom.IsSoldOut == constant.ProductStatusSaleOut {
		return 0, nil
	}

	// 6. 使用成本卡的 CalculateExpectedProductionNum 方法计算库存
	inventory := bom.ProductSauce.ProductBomCard.CalculateExpectedProductionNum()
	// 确保返回非负数
	return math.Max(0, inventory), nil
}
