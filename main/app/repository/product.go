package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductRepo 定义商品仓库接口
type IProductRepo interface {
	GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) // 分页获取商品列表
	GetProductCategoryList(opts ...DBOption) ([]model.ProductCategory, error)                                       // 获取产品类别列表
	WithMultiLanguageName() DBOption                                                                                // 预加载多语言名称
	WithProductUnit() DBOption                                                                                      // 预加载产品单位
	WithProductUnitMultiLanguageName() DBOption                                                                     // 预加载产品单位多语言名称
	WithProductBoms() DBOption                                                                                      // 预加载产品Boms
	WithProductBomsProductFlavor() DBOption                                                                         // 预加载产品Boms产品口味
	WithProductBomsProductFlavorMultiLanguageName() DBOption                                                        // 预加载产品Boms产品口味多语言名称
	WithProductBomsProductSauce() DBOption                                                                          // 预加载产品Boms产品酱料
	WithProductBomsProductSauceMultiLanguageName() DBOption                                                         // 预加载产品Boms产品酱料多语言名称
	WithProductPackageAttributeGroup() DBOption                                                                     // 预加载产品包装属性组
	WithProductPackageAttributeGroupProductAttributeGroup() DBOption                                                // 预加载产品包装属性组产品属性组
	WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName() DBOption                               // 预加载产品包装属性组产品属性组多语言名称
	WithProductPackageAttributeGroupProductPackageAttributes() DBOption                                             // 预加载产品包装属性组产品包装属性
	WithProductPackageAttributeGroupProductPackageAttributesAttribute() DBOption                                    // 预加载产品包装属性组产品包装属性属性
	WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName() DBOption                   // 预加载产品包装属性组产品包装属性属性多语言名称

	WithProductPackage() DBOption                                                                           // bom关联包
	WithProductPackageMultiLanguageName() DBOption                                                          // bom关联包多语言
	WithProductFlavor() DBOption                                                                            // bom关联规格名称
	WithProductFlavorMultiLanguageName() DBOption                                                           // bom关联规格名称多语言
	GetSoldOutWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductBom, int64, error) // 分页获取沽清商品列表
	WhereBomUuid(uuid uint64) DBOption                                                                      // bom uuid 查询条件
	WhereBomIsSoldOut() DBOption                                                                            // 根据bom售罄查询
	UpdateProductBomSoldOut(opts []DBOption, vars map[string]any) error                                     // 更新bom售罄状态
}

// productRepo 商品仓库
type productRepo struct {
	db *gorm.DB
}

// NewProductRepo 创建新的商品仓库
func NewProductRepo(db *gorm.DB) IProductRepo {
	return NewProductRepoImpl(db)
}

// NewProductRepoImpl 创建新的商品仓库实现
func NewProductRepoImpl(db *gorm.DB) IProductRepo {
	return &productRepo{db: db}
}

// GetProductListWithPagination 分页获取商品列表
func (r *productRepo) GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) {
	var products []model.ProductPackage
	var total int64

	db := r.db.Model(&model.ProductPackage{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&products).Error

	return products, total, err
}

// GetSoldOutWithPagination 分页获取沽清商品列表
func (r *productRepo) GetSoldOutWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductBom, int64, error) {
	var productBom []model.ProductBom
	var total int64
	db := r.db.Model(&model.ProductBom{}).Session(&gorm.Session{})
	for _, opt := range opts {
		db = opt(db)
	}
	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 获取列表
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&productBom).Error
	return productBom, total, err
}

// GetProductCategoryList 获取产品类别列表
func (r *productRepo) GetProductCategoryList(opts ...DBOption) ([]model.ProductCategory, error) {
	var categories []model.ProductCategory

	db := r.db.Model(&model.ProductCategory{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// WithMultiLanguageName 预加载多语言名称
func (r *productRepo) WithMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MultiLanguageName")
	}
}

// WithProductUnit 预加载产品单位
func (r *productRepo) WithProductUnit() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductUnit")
	}
}

// WithProductUnitMultiLanguageName 预加载产品单位多语言名称
func (r *productRepo) WithProductUnitMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductUnit.MultiLanguageName")
	}
}

// WithProductBoms 预加载产品Boms
func (r *productRepo) WithProductBoms() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms")
	}
}

// WithProductBomsProductFlavor 预加载产品Boms产品口味
func (r *productRepo) WithProductBomsProductFlavor() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms.ProductFlavor")
	}
}

// WithProductBomsProductFlavorMultiLanguageName 预加载产品Boms产品口味多语言名称
func (r *productRepo) WithProductBomsProductFlavorMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms.ProductFlavor.MultiLanguageName")
	}
}

// WithProductBomsProductSauce 预加载产品Boms产品酱料
func (r *productRepo) WithProductBomsProductSauce() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms.ProductSauce")
	}
}

// WithProductBomsProductSauceMultiLanguageName 预加载产品Boms产品酱料多语言名称
func (r *productRepo) WithProductBomsProductSauceMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms.ProductSauce.MultiLanguageName")
	}
}

// WithProductPackageAttributeGroup 预加载产品包装属性组
func (r *productRepo) WithProductPackageAttributeGroup() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroup")
	}
}

// WithProductPackageAttributeGroupProductAttributeGroup 预加载产品包装属性组产品属性组
func (r *productRepo) WithProductPackageAttributeGroupProductAttributeGroup() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroup.ProductAttributeGroup")
	}
}

// WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName 预加载产品包装属性组产品属性组多语言名称
func (r *productRepo) WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroup.ProductAttributeGroup.MultiLanguageName")
	}
}

// WithProductPackageAttributeGroupProductPackageAttributes 预加载产品包装属性组产品包装属性
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributes() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroup.ProductPackageAttributes")
	}
}

// WithProductPackageAttributeGroupProductPackageAttributesAttribute 预加载产品包装属性组产品包装属性属性
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributesAttribute() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroup.ProductPackageAttributes.Attribute")
	}
}

// WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName 预加载产品包装属性组产品包装属性属性多语言名称
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroup.ProductPackageAttributes.Attribute.MultiLanguageName")
	}
}

// WithProductPackage 预加载产品包
func (r *productRepo) WithProductPackage() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackage")
	}
}

// WithProductPackageMultiLanguageName 预加载产品包多语言
func (r *productRepo) WithProductPackageMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackage.MultiLanguageName")
	}
}

// WithProductFlavor 预加载产品规格
func (r *productRepo) WithProductFlavor() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductFlavor")
	}
}

// WithProductFlavorMultiLanguageName 预加载产品规格多语言
func (r *productRepo) WithProductFlavorMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductFlavor.MultiLanguageName")
	}
}

// UpdateProductBomSoldOut 更新bom售罄状态
func (r *productRepo) UpdateProductBomSoldOut(opts []DBOption, vars map[string]any) error {
	db := r.db.Model(&model.ProductBom{})
	for _, opt := range opts {
		db = opt(db)
	}
	return r.db.Updates(vars).Error
}

// WhereBomUuid 根据bom UUID查询
func (r *productRepo) WhereBomUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereBomIsSoldOut 根据bom售罄查询
func (r *productRepo) WhereBomIsSoldOut() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_sold_out = 1")
	}
}
