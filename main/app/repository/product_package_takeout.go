package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPackageTakeoutRepo interface {
	IProductPackageTakeoutQueryRepo
	CreateProductPackageTakeout(productPackageTakeout *model.ProductPackageTakeout) error
	UpdateProductPackageTakeout(data map[string]any, opts ...DBOption) error
	DestroyProductPackageTakeout(opts ...DBOption) error
	ForceDestroyProductPackageTakeout(opts ...DBOption) error
}

type IProductPackageTakeoutQueryRepo interface {
	GetProductPackageTakeout(opts ...DBOption) (*model.ProductPackageTakeout, error)
	GetProductPackageTakeoutList(opts ...DBOption) ([]*model.ProductPackageTakeout, error)
	GetProductPackageTakeoutByProductPackageUuid(productPackageUuid uint64, takeoutType uint) (*model.ProductPackageTakeout, error)
	GetProductPackageTakeoutBySourceProductId(source string, sourceProductId string) (*model.ProductPackageTakeout, error) // 根据来源平台和商品ID查询
	CheckProductPackageTakeoutExist(productPackageUuid uint64, takeoutType uint) bool
	GetProductPackageTakeoutIncludeSoftDelete(productPackageUuid uint64, takeoutType uint) (*model.ProductPackageTakeout, error) // 获取外卖商品（包括软删除的记录）
	GetProductPackageTakeoutCount(opts ...DBOption) (int64, error)                                                               // 统计外卖商品数量

	WithProductPackage(opts ...DBOption) DBOption                  // 预加载商品包
	WithProductPackageMultiLanguageName(opts ...DBOption) DBOption // 预加载商品包多语言名称
	WithMultiLanguageName(opts ...DBOption) DBOption               // 预加载多语言名称
	WithDescribeMultiLanguageName(opts ...DBOption) DBOption       // 预加载卖点多语言
	WithProductCategory(opts ...DBOption) DBOption                 // 预加载外卖分类
	WithProductSpecialCategory(opts ...DBOption) DBOption          // 预加载外卖特色分类
	WithImageFile(opts ...DBOption) DBOption                       // 预加载图片

	WhereByUuid(uuid uint64) DBOption                                      // 根据UUID查询
	WhereByProductPackageUuid(productPackageUuid uint64) DBOption          // 根据商品包UUID查询
	WhereByProductPackageUuids(productPackageUuids []uint64) DBOption      // 根据商品包UUID列表查询
	WithProductBomTakeouts(opts ...DBOption) DBOption                      // 预加载外卖规格价格列表
	WithProductPackageAttributeTakeouts(opts ...DBOption) DBOption         // 预加载外卖属性价格列表
	WithProductPackageGroupItemTakeouts(opts ...DBOption) DBOption         // 预加载外卖套餐子商品价格列表
	WhereByTakeoutType(takeoutType uint) DBOption                          // 根据外卖类型查询
	WhereByTakeoutTypes(takeoutTypes []uint) DBOption                      // 根据外卖类型列表查询
	WhereBySource(source string) DBOption                                  // 根据来源平台查询
	WhereBySourceProductId(sourceProductId string) DBOption                // 根据来源商品ID查询
	WhereByCategoryUuid(categoryUuid uint64) DBOption                      // 根据分类UUID查询
	WhereBySpecialCategoryUuid(specialCategoryUuid uint64) DBOption        // 根据特殊分类UUID查询
	WhereByCategoryUuidOrSpecialCategoryUuid(categoryUuid uint64) DBOption // 根据分类UUID或特殊分类UUID查询（OR条件）
}

type productPackageTakeoutRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageTakeoutRepo(db *gorm.DB) IProductPackageTakeoutRepo {
	return &productPackageTakeoutRepoImpl{db: db}
}

func (r *productPackageTakeoutRepoImpl) GetProductPackageTakeoutList(opts ...DBOption) ([]*model.ProductPackageTakeout, error) {
	var list []*model.ProductPackageTakeout
	db := r.db

	db = db.Model(&model.ProductPackageTakeout{}).Debug()

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&list)
	if result.Error != nil {
		return nil, result.Error
	}

	return list, nil
}

func (r *productPackageTakeoutRepoImpl) GetProductPackageTakeout(opts ...DBOption) (*model.ProductPackageTakeout, error) {
	var productPackageTakeout model.ProductPackageTakeout
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productPackageTakeout)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productPackageTakeout, nil
}

func (r *productPackageTakeoutRepoImpl) GetProductPackageTakeoutByProductPackageUuid(productPackageUuid uint64, takeoutType uint) (*model.ProductPackageTakeout, error) {
	return r.GetProductPackageTakeout(
		CommonRepo.WhereByProductPackageUuid(productPackageUuid),
		r.WhereByTakeoutType(takeoutType),
		CommonRepo.WhereBySoftDelete(),
	)
}

// GetProductPackageTakeoutBySourceProductId 根据来源平台和商品ID查询外卖商品
func (r *productPackageTakeoutRepoImpl) GetProductPackageTakeoutBySourceProductId(source string, sourceProductId string) (*model.ProductPackageTakeout, error) {
	return r.GetProductPackageTakeout(
		r.WhereBySource(source),
		r.WhereBySourceProductId(sourceProductId),
		CommonRepo.WhereBySoftDelete(),
	)
}

func (r *productPackageTakeoutRepoImpl) CheckProductPackageTakeoutExist(productPackageUuid uint64, takeoutType uint) bool {
	_, err := r.GetProductPackageTakeoutByProductPackageUuid(productPackageUuid, takeoutType)
	return err == nil
}

// GetProductPackageTakeoutIncludeSoftDelete 获取外卖商品（包括软删除的记录）
func (r *productPackageTakeoutRepoImpl) GetProductPackageTakeoutIncludeSoftDelete(productPackageUuid uint64, takeoutType uint) (*model.ProductPackageTakeout, error) {
	return r.GetProductPackageTakeout(
		CommonRepo.WhereByProductPackageUuid(productPackageUuid),
		r.WhereByTakeoutType(takeoutType),
	)
}

// GetProductPackageTakeoutCount 统计外卖商品数量
func (r *productPackageTakeoutRepoImpl) GetProductPackageTakeoutCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.ProductPackageTakeout{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	return total, err
}

func (r *productPackageTakeoutRepoImpl) CreateProductPackageTakeout(productPackageTakeout *model.ProductPackageTakeout) error {
	return r.db.Create(productPackageTakeout).Error
}

func (r *productPackageTakeoutRepoImpl) UpdateProductPackageTakeout(data map[string]any, opts ...DBOption) error {
	db := r.db.Model(&model.ProductPackageTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Updates(data).Error
}

func (r *productPackageTakeoutRepoImpl) DestroyProductPackageTakeout(opts ...DBOption) error {
	db := r.db.Model(&model.ProductPackageTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Update("delete_time", time.Now().Unix()).Error
}

func (r *productPackageTakeoutRepoImpl) ForceDestroyProductPackageTakeout(opts ...DBOption) error {
	db := r.db.Model(&model.ProductPackageTakeout{})

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Delete(&model.ProductPackageTakeout{}).Error
}

// WhereByUuid 根据UUID查询
func (r *productPackageTakeoutRepoImpl) WhereByUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereByProductPackageUuid 根据商品包UUID查询
func (r *productPackageTakeoutRepoImpl) WhereByProductPackageUuid(productPackageUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_package_uuid = ?", productPackageUuid)
	}
}

// WhereByProductPackageUuids 根据商品包UUID列表查询
func (r *productPackageTakeoutRepoImpl) WhereByProductPackageUuids(productPackageUuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(productPackageUuids) == 0 {
			return db
		}
		return db.Where("product_package_uuid IN ?", productPackageUuids)
	}
}

// WhereByTakeoutType 根据外卖类型查询
func (r *productPackageTakeoutRepoImpl) WhereByTakeoutType(takeoutType uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("takeout_type = ?", takeoutType)
	}
}

// WhereByTakeoutTypes 根据外卖类型列表查询
func (r *productPackageTakeoutRepoImpl) WhereByTakeoutTypes(takeoutTypes []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("takeout_type IN (?)", takeoutTypes)
	}
}

// WhereBySource 根据来源平台查询
func (r *productPackageTakeoutRepoImpl) WhereBySource(source string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source = ?", source)
	}
}

// WhereBySourceProductId 根据来源商品ID查询
func (r *productPackageTakeoutRepoImpl) WhereBySourceProductId(sourceProductId string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source_product_id = ?", sourceProductId)
	}
}

// WhereByCategoryUuid 根据分类UUID查询
func (r *productPackageTakeoutRepoImpl) WhereByCategoryUuid(categoryUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("category_uuid = ?", categoryUuid)
	}
}

// WhereBySpecialCategoryUuid 根据特殊分类UUID查询
func (r *productPackageTakeoutRepoImpl) WhereBySpecialCategoryUuid(specialCategoryUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("special_category_uuid = ?", specialCategoryUuid)
	}
}

// WhereByCategoryUuidOrSpecialCategoryUuid 根据分类UUID或特殊分类UUID查询（OR条件）
func (r *productPackageTakeoutRepoImpl) WhereByCategoryUuidOrSpecialCategoryUuid(categoryUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("category_uuid = ? OR special_category_uuid = ?", categoryUuid, categoryUuid)
	}
}

// WithProductPackage 预加载商品包
func (r *productPackageTakeoutRepoImpl) WithProductPackage(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackage", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageMultiLanguageName 预加载商品包多语言名称
func (r *productPackageTakeoutRepoImpl) WithProductPackageMultiLanguageName(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackage.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithMultiLanguageName 预加载多语言名称
func (r *productPackageTakeoutRepoImpl) WithMultiLanguageName(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithDescribeMultiLanguageName 预加载卖点多语言
func (r *productPackageTakeoutRepoImpl) WithDescribeMultiLanguageName(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("DescribeMultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductCategory 预加载外卖分类
func (r *productPackageTakeoutRepoImpl) WithProductCategory(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		}).Preload("ProductCategory.MultiLanguageName")
	}
}

// WithProductSpecialCategory 预加载外卖特色分类
func (r *productPackageTakeoutRepoImpl) WithProductSpecialCategory(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductSpecialCategory", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		}).Preload("ProductSpecialCategory.MultiLanguageName")
	}
}

// WithImageFile 预加载图片
func (r *productPackageTakeoutRepoImpl) WithImageFile(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ImageFile", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductBomTakeouts 预加载外卖规格价格列表
func (r *productPackageTakeoutRepoImpl) WithProductBomTakeouts(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBomTakeouts", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageAttributeTakeouts 预加载外卖属性价格列表
func (r *productPackageTakeoutRepoImpl) WithProductPackageAttributeTakeouts(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeTakeouts", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageGroupItemTakeouts 预加载外卖套餐子商品价格列表
func (r *productPackageTakeoutRepoImpl) WithProductPackageGroupItemTakeouts(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageGroupItemTakeouts", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}
