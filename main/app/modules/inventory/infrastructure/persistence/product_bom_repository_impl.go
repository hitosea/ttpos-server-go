package inventory

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	appRepo "ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// ProductBomRepositoryImpl 商品BOM仓储实现
type ProductBomRepositoryImpl struct {
	dbm *database.DBManager
}

// NewProductBomRepository 创建商品BOM仓储
func NewProductBomRepository(dbm *database.DBManager) repository.IProductBomRepository {
	return &ProductBomRepositoryImpl{
		dbm: dbm,
	}
}

// FindByUuid 根据UUID查找商品BOM
// 性能优化：
// 1. 使用 GetFlavorProductBomByUuid 方法，它已经包含了必要的预加载
// 2. 预加载 ProductBomCard.RelatedMaterials.Material.WarehouseItems，避免 N+1 查询
// 3. 使用索引 uuid 字段进行快速查询
func (r *ProductBomRepositoryImpl) FindByUuid(
	ctx context.Context,
	uuid uint64,
) (interface{}, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewProductBomRepo(db)

	// 使用现有的 GetFlavorProductBomByUuid 方法，它已经包含了必要的预加载
	// 预加载链：ProductBomCard -> RelatedMaterials -> Material -> WarehouseItems
	// 这样可以一次性加载所有需要的数据，避免 N+1 查询问题
	productBom, err := repo.GetFlavorProductBomByUuid(uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询商品BOM失败")
	}

	// 如果是小料，需要额外预加载 ProductSauce 的相关数据（成本卡和关联材料）
	if productBom != nil && productBom.IsSauce() {
		db := r.getDB(ctx)
		err = db.Model(productBom).
			Preload("ProductSauce.ProductBomCard.RelatedMaterials.Material.WarehouseItems").
			Preload("ProductSauce.SauceMaterials", func(db *gorm.DB) *gorm.DB {
				return db.Where("delete_time = ?", 0)
			}).
			Preload("ProductSauce.SauceMaterials.Material.WarehouseItems").
			First(productBom, productBom.Uuid).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, errors.WithMessage(err, "预加载小料数据失败")
		}
	}

	return productBom, nil
}

// FindByProductPackageUuid 根据商品包UUID查找BOM列表
// 性能优化：
// 1. 使用索引 product_package_uuid 字段进行快速查询
// 2. 预加载 ProductBomCard.RelatedMaterials.Material.WarehouseItems，避免 N+1 查询
// 3. 批量查询，减少数据库往返次数
func (r *ProductBomRepositoryImpl) FindByProductPackageUuid(
	ctx context.Context,
	productPackageUuid uint64,
) ([]interface{}, error) {
	db := r.getDB(ctx)
	repo := appRepo.NewProductBomRepo(db)

	// 创建 DBOption 函数
	whereProductPackageUuid := func(db *gorm.DB) *gorm.DB {
		return db.Where("product_package_uuid = ?", productPackageUuid)
	}

	// 查询商品包下的所有BOM
	// 预加载关联数据，避免后续查询时的 N+1 问题
	productBoms, err := repo.GetProductBoms(
		whereProductPackageUuid,
		appRepo.CommonRepo.WhereBySoftDelete(),
		appRepo.CommonRepo.Preload(
			appRepo.WithPreload{
				Query: "ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			appRepo.WithPreload{
				Query: "FlavorMaterials",
				Args: []interface{}{
					appRepo.CommonRepo.DBOption(appRepo.CommonRepo.WhereBySoftDelete()),
				},
			},
			appRepo.WithPreload{
				Query: "FlavorMaterials.Material.WarehouseItems",
			},
			appRepo.WithPreload{
				Query: "ProductSauce.ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			appRepo.WithPreload{
				Query: "ProductSauce.SauceMaterials",
				Args: []interface{}{
					appRepo.CommonRepo.DBOption(appRepo.CommonRepo.WhereBySoftDelete()),
				},
			},
			appRepo.WithPreload{
				Query: "ProductSauce.SauceMaterials.Material.WarehouseItems",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询商品BOM列表失败")
	}

	// 转换为 interface{} 切片
	result := make([]interface{}, len(productBoms))
	for i, bom := range productBoms {
		result[i] = bom
	}

	return result, nil
}

// getDB 获取数据库连接
func (r *ProductBomRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	return ctx.GetDB()
}
