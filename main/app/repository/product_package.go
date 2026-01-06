package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

type IProductPackageRepo interface {
	IProductPackageQueryRepo
	CreateProductPackage(productPackage *model.ProductPackage) error
	CreateProductPackages(productPackages []model.ProductPackage) error
	UpdateProductPackage(data map[string]any, opts ...DBOption) error
	AddActualSaleNum(productPackageUuid uint64, saleNum float64) error
	SubActualSaleNum(productPackageUuid uint64, saleNum float64) error
	DestroyProductPackage(opts ...DBOption) error
}

type IProductPackageQueryRepo interface {
	GetProductPackage(opts ...DBOption) (*model.ProductPackage, error)
	GetProductPackageList(opts ...DBOption) ([]*model.ProductPackage, error)
	GetProductPackageBoms(productPackageUuid uint64) (*model.ProductPackage, error) // 获取商品包的库存信息
	GetProductPackageBaseInfoByBomUuid(companyUuid uint64, flavorBomUuid uint64) (*model.ProductBom, error)
	GetProductPackageListByUuids(uuids []uint64) ([]*model.ProductPackage, error)
	GetProductPackageBatchTagCount() ([]uint64, error)                                                       // 获取分批商品数量
	SetProductPackageBatch(uuids []uint64, isBatch uint) error                                               // 将is_batch设置为1或0
	GetProductPackageListByUuidsAndIsBatch(uuids []uint64, isBatch uint) ([]*model.ProductPackage, error)    // 根据uuid列表和is_batch查询商品包列表
	GetProductPackageListByIsBatch(isBatch uint) ([]*model.ProductPackage, error)                            // 根据is_batch查询商品包列表
	GetProductPackageUuidsByIsPackage() ([]uint64, error)                                                    // 查询所有套餐的uuid列表
	GetProductPackageByUuidWithAssociations(productPackageUuid uint64) (*model.ProductPackage, error)        // 通过UUID查询商品包信息，包括MultiLanguageName、DineTax、TakeoutTax、ProductUnit、ProductUnit.MultiLanguageName、ProductCategory
	GetProductPackagesByUuidsWithAssociations(productPackageUuids []uint64) ([]*model.ProductPackage, error) // 批量通过UUID列表查询商品包信息，包括MultiLanguageName、DineTax、TakeoutTax、ProductUnit、ProductUnit.MultiLanguageName、ProductCategory

	WithMultiLanguageName(opts ...DBOption) DBOption                      // 预加载多语言名称
	WithProductBoms(opts ...DBOption) DBOption                            // 预加载商品bom
	WithProductBomsProductFlavor(opts ...DBOption) DBOption               // 预加载商品bom产品-规格
	WithProductCategory(opts ...DBOption) DBOption                        // 预加载商品分类
	WithProductCategoryMultiLanguageName(opts ...DBOption) DBOption       // 预加载商品分类多语言名称
	WithProductPackageAttributeGroups(opts ...DBOption) DBOption          // 预加载产品包装属性组
	WithProductPackageAttributeGroupAttributes(opts ...DBOption) DBOption // 预加载产品包装属性组产品属性
	WithProductPackageGroups(opts ...DBOption) DBOption                   // 预加载商品套餐组
	WithProductPackageGroupItems(opts ...DBOption) DBOption               // 预加载商品套餐组商品
	WithProductPackageGroupMultiLanguageName(opts ...DBOption) DBOption   // 预加载商品套餐组多语言名称
	WithProductLabel(opts ...DBOption) DBOption                           // 预加载商品标签
}

type productPackageRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageRepo(db *gorm.DB) IProductPackageRepo {
	return &productPackageRepoImpl{db: db}
}

func (r *productPackageRepoImpl) GetProductPackageList(opts ...DBOption) ([]*model.ProductPackage, error) {
	var productPackages []*model.ProductPackage
	db := r.db

	db = db.Model(&model.ProductPackage{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productPackages)
	if result.Error != nil {
		return nil, result.Error
	}

	return productPackages, nil
}

func (r *productPackageRepoImpl) GetProductPackage(opts ...DBOption) (*model.ProductPackage, error) {
	var productPackage model.ProductPackage
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productPackage)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productPackage, nil
}

// GetProductPackageBoms 查询商品包库存信息
func (r *productPackageRepoImpl) GetProductPackageBoms(productPackageUuid uint64) (*model.ProductPackage, error) {
	productPackage, err := r.GetProductPackage(
		CommonRepo.WhereByUuid(productPackageUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductBoms",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackage, nil
}

// GetProductPackageByUuidWithAssociations 通过UUID查询商品包信息，包括MultiLanguageName、DineTax、TakeoutTax、ProductUnit、ProductUnit.MultiLanguageName、ProductCategory
func (r *productPackageRepoImpl) GetProductPackageByUuidWithAssociations(productPackageUuid uint64) (*model.ProductPackage, error) {
	productPackage, err := r.GetProductPackage(
		CommonRepo.WhereByUuid(productPackageUuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
			WithPreload{
				Query: "DineTax",
			},
			WithPreload{
				Query: "TakeoutTax",
			},
			WithPreload{
				Query: "ProductUnit",
			},
			WithPreload{
				Query: "ProductUnit.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductCategory",
			},
			WithPreload{
				Query: "ProductBoms",
			},
			WithPreload{
				Query: "productPackageAttributeGroups",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackage, nil
}

// GetProductPackagesByUuidsWithAssociations 批量通过UUID列表查询商品包信息，包括MultiLanguageName、DineTax、TakeoutTax、ProductUnit、ProductUnit.MultiLanguageName、ProductCategory
func (r *productPackageRepoImpl) GetProductPackagesByUuidsWithAssociations(productPackageUuids []uint64) ([]*model.ProductPackage, error) {
	if len(productPackageUuids) == 0 {
		return []*model.ProductPackage{}, nil
	}
	productPackages, err := r.GetProductPackageList(
		CommonRepo.WhereInUuids(productPackageUuids),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
			WithPreload{
				Query: "DineTax",
			},
			WithPreload{
				Query: "TakeoutTax",
			},
			WithPreload{
				Query: "ProductUnit",
			},
			WithPreload{
				Query: "ProductUnit.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductCategory",
			},
			WithPreload{
				Query: "ProductBoms",
			},
			WithPreload{
				Query: "ProductPackageAttributeGroups",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackages, nil
}

// WithProductBomsProductFlavor 预加载商品bom产品-规格
func (r *productPackageRepoImpl) WithProductBomsProductFlavor(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms.ProductFlavor", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

func (r *productPackageRepoImpl) GetProductPackageBaseInfoByBomUuid(companyUuid uint64, flavorBomUuid uint64) (*model.ProductBom, error) {
	// 检查是否启用对象存储缓存
	var productBom *model.ProductBom
	var err error

	if adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 使用对象存储模块缓存查询
		productBom, err = r.getProductBomWithCache(companyUuid, flavorBomUuid)
	} else {
		// 直接查询数据库
		productBom, err = r.queryProductBomBaseInfo(flavorBomUuid)
	}

	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom, nil
}

// queryProductBomBaseInfo 查询 ProductBom 基础信息（包含预加载的关联数据）
// 这是一个私有方法，用于统一查询逻辑，避免代码重复
func (r *productPackageRepoImpl) queryProductBomBaseInfo(flavorBomUuid uint64) (*model.ProductBom, error) {
	// 先查询出ProductBom
	productBomRepo := NewProductBomRepo(r.db)
	productBom, err := productBomRepo.GetProductBom(
		CommonRepo.WhereByUuid(flavorBomUuid),
		CommonRepo.Preload(WithPreload{Query: "ProductPackage.TakeoutTax"}),
		CommonRepo.Preload(
			WithPreload{Query: "ProductPackage.MultiLanguageName"},
			WithPreload{Query: "ProductPackage.ProductCategory"},
			WithPreload{Query: "ProductPackage.DineTax"},
			WithPreload{Query: "ProductFlavor.MultiLanguageName"},
			WithPreload{Query: "ProductPackage.ProductPackageAttributeGroups"},
			WithPreload{Query: "ProductPackage.ProductPackageAttributeGroups.ProductPackageAttributes"},
			WithPreload{Query: "ProductPackage.ProductPackageAttributeGroups.ProductAttributeGroup"},
			WithPreload{Query: "ProductPackage.ProductPackageAttributeGroups.ProductAttributeGroup.MultiLanguageName"},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom, nil
}

// getProductBomWithCache 使用对象存储模块缓存查询 ProductBom（包含预加载的关联数据）
func (r *productPackageRepoImpl) getProductBomWithCache(companyUuid uint64, flavorBomUuid uint64) (*model.ProductBom, error) {
	if flavorBomUuid == 0 {
		return nil, errors.New("flavorBomUuid cannot be 0")
	}
	if companyUuid == 0 {
		return nil, errors.New("companyUuid cannot be 0")
	}
	// 构建缓存 key（使用 companyUuid 而不是 context）
	key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductBomBaseInfo, flavorBomUuid)

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*model.ProductBom](cache.Global, 5*time.Minute)

	// 使用缓存查询
	result, err := cacheLayer.GET(key, func() (*model.ProductBom, error) {
		// 缓存未命中时，从数据库查询（包含所有预加载）
		return r.queryProductBomBaseInfo(flavorBomUuid)
	})

	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.queryProductBomBaseInfo(flavorBomUuid)
	}

	return result, nil
}

func (r *productPackageRepoImpl) GetProductPackageListByUuids(uuids []uint64) ([]*model.ProductPackage, error) {
	var productPackages []*model.ProductPackage
	err := r.db.Model(&model.ProductPackage{}).Where("uuid IN ?", uuids).Find(&productPackages).Error
	return productPackages, errors.WithMessage(err)
}

func (r *productPackageRepoImpl) CreateProductPackage(productPackage *model.ProductPackage) error {
	// 创建product_package表数据
	productPkg := *productPackage
	productPkg.SetNil()
	if err := r.db.Model(&model.ProductPackage{}).Create(&productPkg).Error; err != nil {
		return errors.WithMessage(err)
	}

	// 创建multi_language_name表数据
	multiLanguageNameRepo := NewMultiLanguageNameRepo(r.db)
	if _, err := multiLanguageNameRepo.CreateMultiLanguageName(productPackage.MultiLanguageName); err != nil {
		return errors.WithMessage(err)
	}

	productPackageAttributeGroupRepo := NewProductPackageAttributeGroupRepo(r.db)
	if err := productPackageAttributeGroupRepo.CreateProductPackageAttributeGroups(productPackage.ProductPackageAttributeGroups); err != nil {
		return errors.WithMessage(err)
	}

	// 创建product_package_attribute表数据
	productPackageAttributeRepo := NewProductPackageAttributeRepo(r.db)
	for _, attributeGroup := range productPackage.ProductPackageAttributeGroups {
		if err := productPackageAttributeRepo.CreateProductPackageAttributes(attributeGroup.ProductPackageAttributes); err != nil {
			return errors.WithMessage(err)
		}
	}
	// 创建product_bom表数据
	productBomRepo := NewProductBomRepo(r.db)
	if err := productBomRepo.CreateProductBoms(productPackage.ProductBoms); err != nil {
		return errors.WithMessage(err)
	}

	// 创建related_material表数据
	relatedMaterialRepo := NewRelatedMaterialRepo(r.db)
	for _, bom := range productPackage.ProductBoms {
		if len(bom.FlavorMaterials) > 0 {
			relatedMaterials := make([]model.RelatedMaterial, 0)
			for _, material := range bom.FlavorMaterials {
				relatedMaterials = append(relatedMaterials, *material)
			}
			if err := relatedMaterialRepo.CreateRelatedMaterials(relatedMaterials); err != nil {
				return errors.WithMessage(err)
			}
		}
	}
	return nil
}

func (r *productPackageRepoImpl) CreateProductPackages(productPackages []model.ProductPackage) error {
	return r.db.Create(productPackages).Error
}

// UpdateProductPackage 更新产品包
func (r *productPackageRepoImpl) UpdateProductPackage(data map[string]any, opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Model(&model.ProductPackage{}).Updates(data).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productPackageRepoImpl) AddActualSaleNum(productPackageUuid uint64, saleNum float64) error {
	if err := r.db.Model(&model.ProductPackage{}).Where("uuid = ?", productPackageUuid).Update("actual_sale_num", gorm.Expr("actual_sale_num + ?", saleNum)).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productPackageRepoImpl) SubActualSaleNum(productPackageUuid uint64, saleNum float64) error {
	if err := r.db.Model(&model.ProductPackage{}).Where("uuid = ?", productPackageUuid).Update("actual_sale_num", gorm.Expr("actual_sale_num - ?", saleNum)).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// DestroyProductPackage 销毁商品包
func (r *productPackageRepoImpl) DestroyProductPackage(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ProductPackage{}).Error
}

// GetProductPackageBatchTagCount 获取分批商品数量
func (r *productPackageRepoImpl) GetProductPackageBatchTagCount() ([]uint64, error) {
	var productPackageBatchTagCounts []uint64
	err := r.db.Model(&model.ProductPackage{}).Where("is_batch <> ?", 0).Where("delete_time = ?", 0).Select("uuid").Scan(&productPackageBatchTagCounts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageBatchTagCounts, nil
}

// 根据uuids将is_batch设置为1或0
func (r *productPackageRepoImpl) SetProductPackageBatch(uuids []uint64, isBatch uint) error {
	err := r.db.Model(&model.ProductPackage{}).Where("uuid IN ?", uuids).Updates(map[string]any{"is_batch": isBatch}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// 根据uuid列表和is_batch查询商品包列表
func (r *productPackageRepoImpl) GetProductPackageListByUuidsAndIsBatch(uuids []uint64, isBatch uint) ([]*model.ProductPackage, error) {
	var productPackages []*model.ProductPackage
	err := r.db.Model(&model.ProductPackage{}).Where("uuid IN ?", uuids).Where("is_batch = ?", isBatch).Where("delete_time = ?", 0).Find(&productPackages).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackages, nil
}

// 根据is_batch查询商品包列表
func (r *productPackageRepoImpl) GetProductPackageListByIsBatch(isBatch uint) ([]*model.ProductPackage, error) {
	var productPackages []*model.ProductPackage
	err := r.db.Model(&model.ProductPackage{}).Where("is_batch = ?", isBatch).Where("delete_time = ?", 0).Find(&productPackages).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackages, nil
}

// WithMultiLanguageName 预加载多语言名称
func (r *productPackageRepoImpl) WithMultiLanguageName(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductBoms 预加载商品bom
func (r *productPackageRepoImpl) WithProductBoms(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageAttributeGroups 预加载产品包装属性组
func (r *productPackageRepoImpl) WithProductPackageAttributeGroups(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageGroups 预加载商品套餐组
func (r *productPackageRepoImpl) WithProductPackageGroups(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageGroups", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageGroupItems 预加载商品套餐组商品
func (r *productPackageRepoImpl) WithProductPackageGroupItems(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageGroups.ProductPackageGroupItems", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageGroupMultiLanguageName 预加载商品套餐组多语言名称
func (r *productPackageRepoImpl) WithProductPackageGroupMultiLanguageName(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageGroups.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductPackageAttributeGroupAttributes 预加载产品包装属性组产品属性
func (r *productPackageRepoImpl) WithProductPackageAttributeGroupAttributes(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductPackageAttributes", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

func (r *productPackageRepoImpl) WithProductCategory(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

func (r *productPackageRepoImpl) WithProductCategoryMultiLanguageName(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// WithProductLabel 预加载商品标签
func (r *productPackageRepoImpl) WithProductLabel(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductLabel", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}

// 查询所有套餐的uuid列表
func (r *productPackageRepoImpl) GetProductPackageUuidsByIsPackage() ([]uint64, error) {
	var productPackageUuids []uint64
	err := r.db.Model(&model.ProductPackage{}).Where("delete_time = ?", 0).Where("product_type = ?", 1).Select("uuid").Scan(&productPackageUuids).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageUuids, nil
}
