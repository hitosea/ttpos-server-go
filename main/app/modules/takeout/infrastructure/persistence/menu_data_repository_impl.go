package persistence

import (
	"ttpos-server-go/app/model"
	inventoryApp "ttpos-server-go/app/modules/inventory/application"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/menu/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// menuDataRepositoryImpl 菜单数据仓储实现
type menuDataRepositoryImpl struct {
	dbm *database.DBManager
}

// NewMenuDataRepository 创建菜单数据仓储
func NewMenuDataRepository(dbm *database.DBManager) menuRepo.IMenuDataRepository {
	return &menuDataRepositoryImpl{
		dbm: dbm,
	}
}

// GetTakeoutCategories 获取外卖分类列表
func (r *menuDataRepositoryImpl) GetTakeoutCategories(ctx context.Context, companyUuid uint64, categoryIDs []uint64) ([]*model.ProductCategory, error) {
	db := ctx.GetDB()
	var categories []*model.ProductCategory

	query := db.Model(&model.ProductCategory{}).
		Where("is_display_in_takeout = ?", 1).
		Where("delete_time = ?", 0).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Order("sort ASC, id ASC")

	// 如果指定了分类ID，则只查询指定分类
	if len(categoryIDs) > 0 {
		query = query.Where("uuid IN ?", categoryIDs)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// GetTakeoutProducts 获取指定分类下的外卖商品
func (r *menuDataRepositoryImpl) GetTakeoutProducts(ctx context.Context, companyUuid uint64, categoryUuid uint64) ([]*model.ProductPackageTakeout, error) {
	db := ctx.GetDB()
	var products []*model.ProductPackageTakeout

	err := db.Model(&model.ProductPackageTakeout{}).
		Where("category_uuid = ?", categoryUuid).
		Preload("ProductPackage", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Preload("MultiLanguageName", "delete_time = ?", 0).
				Preload("DescribeMultiLanguageName", "delete_time = ?", 0).
				Preload("ProductBoms", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0).
						Preload("ProductFlavor.MultiLanguageName", "delete_time = ?", 0).
						Preload("ProductSauce.MultiLanguageName", "delete_time = ?", 0)
				}).
				Preload("ProductPackageAttributeGroups", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0).
						Preload("ProductAttributeGroup.MultiLanguageName", "delete_time = ?", 0).
						Preload("ProductPackageAttributes", func(db *gorm.DB) *gorm.DB {
							return db.Where("delete_time = ?", 0).
								Preload("Attribute.MultiLanguageName", "delete_time = ?", 0)
						})
				}).
				Preload("ProductPackageGroups", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0).
						Preload("MultiLanguageName", "delete_time = ?", 0).
						Preload("ProductPackageGroupItems", func(db *gorm.DB) *gorm.DB {
							return db.Where("delete_time = ?", 0).
								Preload("ProductPackage", func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ?", 0).
										Preload("MultiLanguageName", "delete_time = ?", 0)
								}).
								Preload("ProductBom", func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ?", 0).
										Preload("ProductFlavor.MultiLanguageName", "delete_time = ?", 0)
								})
						})
				})
		}).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Preload("ImageFile").
		Preload("ProductBomTakeouts", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Preload("ProductBom", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				}).
				Preload("ProductBom.ProductFlavor.MultiLanguageName", "delete_time = ?", 0).
				Preload("ProductBom.ProductSauce.MultiLanguageName", "delete_time = ?", 0)
		}).
		Preload("ProductPackageAttributeTakeouts", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Preload("ProductPackageAttribute", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				}).
				Preload("ProductPackageAttribute.Attribute.MultiLanguageName", "delete_time = ?", 0)
		}).
		Order("id ASC").
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	// 注入库存
	r.InjectStockNum(ctx, products)

	return products, nil
}

// InjectStockNum 使用 ProductInventoryAppService 批量查询库存
// 返回库存映射表，key 为 BomUuid，value 为库存数量
func (r *menuDataRepositoryImpl) InjectStockNum(ctx context.Context, takeoutProducts []*model.ProductPackageTakeout) {
	if len(takeoutProducts) == 0 {
		return
	}

	// 使用工厂方法创建库存应用服务实例
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(r.dbm, cache.Global)

	// Step 1: 收集所有需要查询库存的 BOM UUID 及其对应的对象引用
	bomUuids := make([]uint64, 0)
	bomMap := make(map[uint64]*model.ProductBom) // 用于快速定位 BOM 对象

	for _, takeoutProduct := range takeoutProducts {
		// 使用索引遍历以获取可修改的引用
		for i := range takeoutProduct.ProductPackage.ProductBoms {
			bom := &takeoutProduct.ProductPackage.ProductBoms[i]
			if bom.IsDelete() || bom.IsDown() || !bom.IsSauce() || bom.IsFlavor() {
				continue
			}
			bomUuids = append(bomUuids, bom.Uuid)
			bomMap[bom.Uuid] = bom
		}
	}

	// Step 2: 如果没有需要查询的 BOM，直接返回
	if len(bomUuids) == 0 {
		return
	}

	// Step 3: 批量查询所有 BOM 的库存
	inventoryMap, err := appService.GetProductInventoriesBatch(ctx, bomUuids)
	if err != nil {
		logger.Logger.Error("批量查询商品规格/小料库存失败", zap.Error(err), zap.Int("bom_count", len(bomUuids)))
		// 查询失败，设置所有 BOM 为无限库存
		for _, bom := range bomMap {
			bom.StockNum = 99999999
		}
		return
	}

	// Step 4: 将库存值注入到对应的 BOM 对象中
	for bomUuid, bom := range bomMap {
		if inventory, ok := inventoryMap[bomUuid]; ok {
			bom.StockNum = inventory
		} else {
			// 如果某个 BOM 没有返回库存数据，设置为无限库存
			logger.Logger.Warn("未查询到商品规格/小料库存，设置为无限库存", zap.Uint64("bom_uuid", bomUuid))
			bom.StockNum = 99999999
		}
	}
}
