package repository

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

// IProductRepo 定义商品仓库接口
type IProductRepo interface {
	GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) // 分页获取商品列表
	GetProductPackageListByUuids(uuids []uint64) ([]model.ProductPackage, error)                                    // 通过uuid列表获取商品列表
	GetProductCategoryList(opts ...DBOption) ([]model.ProductCategory, error)                                       // 获取产品类别列表
	GetProduct(opts ...DBOption) (model.ProductPackage, error)                                                      // 获取商品详情
	GetProductFlavor(opts ...DBOption) (model.ProductFlavor, error)                                                 // 获取商品口味详情
	GetProductBom(opts ...DBOption) (model.ProductBom, error)                                                       // 获取商品BOM详情
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
	WithProductPackageImageFile() DBOption                                                                          // 预加载产品包的图片信息
	WithProductCategory() DBOption                                                                                  // 预加载分类
	WithDineTax() DBOption                                                                                          // 预加载堂食税
	WithTakeoutTax() DBOption                                                                                       // 预加载外卖税

	WithProductPackage() DBOption                                                                           // 沽清 预加载产品
	WithProductPackageMultiLanguageName() DBOption                                                          // 沽清 预加载产品多语言
	WithProductFlavor() DBOption                                                                            // 沽清 预加载规格名称
	WithProductFlavorMultiLanguageName() DBOption                                                           // 沽清 预加载规格名称多语言
	GetSoldOutWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductBom, int64, error) // 沽清 分页获取沽清商品列表
	WhereBomUuid(uuid uint64) DBOption                                                                      // 沽清 产品 uuid 查询条件
	WhereBomIsSoldOut() DBOption                                                                            // 沽清 产品是否售罄
	UpdateProductBomSoldOut(opts []DBOption, vars map[string]any) error                                     // 沽清 更新产品售罄状态
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

// 默认关联对象
func (r *productRepo) defaultPreload() []DBOption {
	return []DBOption{
		r.WithMultiLanguageName(),
		r.WithProductUnit(),
		r.WithProductUnitMultiLanguageName(),
		r.WithProductBoms(),
		r.WithProductBomsProductFlavor(),
		r.WithProductBomsProductFlavorMultiLanguageName(),
		r.WithProductBomsProductSauce(),
		r.WithProductBomsProductSauceMultiLanguageName(),
		r.WithProductPackageAttributeGroup(),
		r.WithProductPackageAttributeGroupProductAttributeGroup(),
		r.WithProductPackageAttributeGroupProductPackageAttributes(),
		r.WithProductPackageAttributeGroupProductPackageAttributesAttribute(),
		r.WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName(),
		r.WithProductPackageAttributeGroupProductAttributeGroup(),
		r.WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName(),
		r.WithProductPackageAttributeGroupProductPackageAttributes(),
		r.WithProductPackageAttributeGroupProductPackageAttributesAttribute(),
		r.WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName(),
		r.WithProductPackageImageFile(),
		r.WithProductCategory(),
	}
}

// GetProductListWithPagination 分页获取商品列表
func (r *productRepo) GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) {
	var products []model.ProductPackage
	var total int64
	var prefix = config.Database.TablePrefix
	var productPackage = prefix + "product_package"
	var productBom = prefix + "product_bom"

	db := r.db.Model(&model.ProductPackage{}).Session(&gorm.Session{})

	opts = append(r.defaultPreload(), opts...)

	for _, opt := range opts {
		db = opt(db)
	}

	db = db.Joins(fmt.Sprintf("LEFT JOIN %s ON %s.uuid = %s.product_package_uuid", productBom, productPackage, productBom))
	db = db.Where(fmt.Sprintf("%s.status = ?", productPackage), 1).
		Where(fmt.Sprintf("%s.delete_time = ?", productPackage), 0).
		Where(fmt.Sprintf("%s.delete_time = ?", productBom), 0).
		Order(fmt.Sprintf("%s.sort ASC", productPackage)).
		Order(fmt.Sprintf("%s.id DESC", productPackage))

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	// 获取列表
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&products).Error

	return products, total, errors.WithMessage(err)
}

// GetSoldOutWithPagination 分页获取沽清商品列表
func (r *productRepo) GetSoldOutWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductBom, int64, error) {
	var productBom []model.ProductBom
	var total int64
	db := r.db.Model(&model.ProductBom{}).Session(&gorm.Session{}).Where("is_sold_out = 1")
	for _, opt := range opts {
		db = opt(db)
	}
	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	// 获取列表
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&productBom).Error
	return productBom, total, errors.WithMessage(err)
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
		return nil, errors.WithMessage(err)
	}

	return categories, nil
}

// GetProduct 获取商品详情
func (r *productRepo) GetProduct(opts ...DBOption) (model.ProductPackage, error) {
	var product model.ProductPackage

	db := r.db.Model(&model.ProductPackage{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&product).Error

	return product, errors.WithMessage(err)
}

// GetProductFlavor 获取商品规格详情
func (r *productRepo) GetProductFlavor(opts ...DBOption) (model.ProductFlavor, error) {
	var productFlavor model.ProductFlavor

	db := r.db.Model(&model.ProductFlavor{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&productFlavor).Error

	return productFlavor, errors.WithMessage(err)
}

// GetProductBom 获取商品BOM
func (r *productRepo) GetProductBom(opts ...DBOption) (model.ProductBom, error) {
	var productBom model.ProductBom

	db := r.db.Model(&model.ProductBom{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&productBom).Error

	return productBom, errors.WithMessage(err)
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
		return db.Preload("ProductPackageAttributeGroups")
	}
}

// WithProductPackageAttributeGroupProductAttributeGroup 预加载产品包装属性组产品属性组
func (r *productRepo) WithProductPackageAttributeGroupProductAttributeGroup() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductAttributeGroup")
	}
}

// WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName 预加载产品包装属性组产品属性组多语言名称
func (r *productRepo) WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductAttributeGroup.MultiLanguageName")
	}
}

// WithProductPackageAttributeGroupProductPackageAttributes 预加载产品包装属性组产品包装属性
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributes() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductPackageAttributes")
	}
}

// WithProductPackageAttributeGroupProductPackageAttributesAttribute 预加载产品包装属性组产品包装属性属性
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributesAttribute() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductPackageAttributes.Attribute")
	}
}

// WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName 预加载产品包装属性组产品包装属性属性多语言名称
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductPackageAttributes.Attribute.MultiLanguageName")
	}
}

func (r *productRepo) WithProductPackageImageFile() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ImageFile")
	}
}

func (r *productRepo) WithProductCategory() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory")
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
	return db.Updates(vars).Error
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

// WithDineTax 预加载堂食税
func (r *productRepo) WithDineTax() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("DineTax")
	}
}

// WithTakeoutTax 预加载外卖税
func (r *productRepo) WithTakeoutTax() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutTax")
	}
}

// GetProductPackageListByUuids 通过uuid列表获取商品列表
func (r *productRepo) GetProductPackageListByUuids(uuids []uint64) ([]model.ProductPackage, error) {
	var productPackages []model.ProductPackage
	err := r.db.Model(&model.ProductPackage{}).Where("uuid IN ? AND delete_time = ?", uuids, constant.NotDeleted).Find(&productPackages).Error
	return productPackages, errors.WithMessage(err)
}
