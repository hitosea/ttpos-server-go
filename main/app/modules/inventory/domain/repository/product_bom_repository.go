package repository

import (
	"ttpos-server-go/pkg/context"
)

// IProductBomRepository 商品BOM仓储接口
// 注意：由于现有代码使用 Model 层（main/app/model/product.go），
// 本接口返回 *model.ProductBom，而不是 Domain Entity。
// 在 Infrastructure 层实现时，直接使用现有 Model。
type IProductBomRepository interface {
	// FindByUuid 根据UUID查找商品BOM
	// 需要预加载关联数据：ProductBomCard.RelatedMaterials.Material
	FindByUuid(ctx context.Context, uuid uint64) (interface{}, error)

	// FindByUuids 根据UUID列表批量查找商品BOM列表
	// 需要预加载关联数据：ProductBomCard.RelatedMaterials.Material
	FindByUuids(ctx context.Context, uuids []uint64) ([]interface{}, error)

	// FindByProductPackageUuid 根据商品包UUID查找BOM列表
	// 用于批量查询某个商品包下的所有BOM
	FindByProductPackageUuid(ctx context.Context, productPackageUuid uint64) ([]interface{}, error)
}
