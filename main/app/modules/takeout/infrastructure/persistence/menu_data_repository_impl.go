package persistence

import (
	"context"
	"ttpos-server-go/app/model"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/menu/repository"
	"ttpos-server-go/pkg/database"

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
	db := r.dbm.GetDB(companyUuid)
	var categories []*model.ProductCategory

	query := db.Model(&model.ProductCategory{}).
		Where("is_display_in_takeout = ?", 1).
		Where("status = ?", 1).
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
	db := r.dbm.GetDB(companyUuid)
	var products []*model.ProductPackageTakeout

	err := db.Model(&model.ProductPackageTakeout{}).
		Where("category_uuid = ?", categoryUuid).
		Where("status = ?", 1).
		Where("delete_time = ?", 0).
		Preload("ProductPackage", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Preload("MultiLanguageName", "delete_time = ?", 0).
				Preload("DescribeMultiLanguageName", "delete_time = ?", 0).
				Preload("ProductBoms", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0).
						Where("status = ?", 1).
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
		Order("id ASC").
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	return products, nil
}
