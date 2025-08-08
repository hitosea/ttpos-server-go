package repository

import (
	"database/sql"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductRepo 定义商品仓库接口
type IProductRepo interface {
	IProductQueryRepo
	WithMultiLanguageName() DBOption                                                              // 预加载多语言名称
	WithProductUnit() DBOption                                                                    // 预加载产品单位
	WithProductUnitMultiLanguageName() DBOption                                                   // 预加载产品单位多语言名称
	WithProductBoms() DBOption                                                                    // 预加载产品Boms
	WithProductBomsProductFlavor() DBOption                                                       // 预加载产品Boms产品口味
	WithProductBomsProductFlavorMultiLanguageName() DBOption                                      // 预加载产品Boms产品口味多语言名称
	WithProductBomsProductSauce() DBOption                                                        // 预加载产品Boms产品酱料
	WithProductBomsProductSauceMultiLanguageName() DBOption                                       // 预加载产品Boms产品酱料多语言名称
	WithProductPackageAttributeGroup() DBOption                                                   // 预加载产品包装属性组
	WithProductPackageAttributeGroupProductAttributeGroup() DBOption                              // 预加载产品包装属性组产品属性组
	WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName() DBOption             // 预加载产品包装属性组产品属性组多语言名称
	WithProductPackageAttributeGroupProductPackageAttributes() DBOption                           // 预加载产品包装属性组产品包装属性
	WithProductPackageAttributeGroupProductPackageAttributesAttribute() DBOption                  // 预加载产品包装属性组产品包装属性属性
	WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName() DBOption // 预加载产品包装属性组产品包装属性属性多语言名称
	WithProductPackageImageFile() DBOption                                                        // 预加载产品包的图片信息
	WithProductCategory() DBOption                                                                // 预加载分类
	WithDineTax() DBOption                                                                        // 预加载堂食税
	WithTakeoutTax() DBOption                                                                     // 预加载外卖税

	WithProductPackage() DBOption                                                                           // 沽清 预加载产品
	WithProductPackageMultiLanguageName() DBOption                                                          // 沽清 预加载产品多语言
	WithProductFlavor() DBOption                                                                            // 沽清 预加载规格名称
	WithProductFlavorMultiLanguageName() DBOption                                                           // 沽清 预加载规格名称多语言
	GetSoldOutWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductBom, int64, error) // 沽清 分页获取沽清商品列表
	WhereBomUuid(uuid uint64) DBOption                                                                      // 沽清 产品 uuid 查询条件
	WhereBomIsSoldOut() DBOption                                                                            // 沽清 产品是否售罄
	UpdateProductBomSoldOut(opts []DBOption, vars map[string]any) error                                     // 沽清 更新产品售罄状态

	WhereUuid(uuid uint64) DBOption             // 查询条件 产品单位 uuid
	WhereUuidIn(uuids []uint64) DBOption        // 查询条件 产品单位 uuid 列表
	WhereCategoryKey(key string) DBOption       // 查询条件 产品分类key
	WhereByIsSpecial(isSpecial uint) DBOption   // 查询条件 产品分类是否特殊
	WhereParentUuid(parentUuid uint64) DBOption // 查询条件 产品分类父级uuid

	WhereCategoryUuid(categoryUuid uint64) DBOption               // 查询条件 产品分类uuid
	WhereSpecialCategoryUuid(specialCategoryUuid uint64) DBOption // 查询条件 特色分类uuid

	WithProductPackages() DBOption                  // 预加载产品单位关联的商品
	WithProductPackagesMultiLanguageName() DBOption // 预加载产品单位关联的商品多语言名称

	WithActiveProductBoms() DBOption                                 // 预加载商品BOM
	WithActiveProductBomsProductPackages() DBOption                  // 预加载商品BOM关联的商品包
	WithActiveProductBomsProductPackagesMultiLanguageName() DBOption // 预加载商品BOM关联的商品包多语言名称
}

// IProductQueryRepo 商品查询仓库接口
type IProductQueryRepo interface {
	GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) // 分页获取商品列表
	GetProductPackageListByUuids(uuids []uint64) ([]model.ProductPackage, error)                                    // 通过uuid列表获取商品列表
	GetProductCategoryList(opts ...DBOption) ([]model.ProductCategory, error)                                       // 获取产品类别列表
	GetProductCategory(opts ...DBOption) (model.ProductCategory, error)                                             // 获取产品分类详情
	GetProductCategoryCount(opts ...DBOption) (int64, error)                                                        // 获取产品分类数量
	GetProductCategoryMaxSort(opts ...DBOption) (int64, error)                                                      // 获取产品分类最大排序
	GetProduct(opts ...DBOption) (model.ProductPackage, error)                                                      // 获取商品详情
	GetProductCount(opts ...DBOption) (int64, error)                                                                // 获取商品数量
	GetProductFlavor(opts ...DBOption) (model.ProductFlavor, error)                                                 // 获取商品口味详情
	GetProductBom(opts ...DBOption) (model.ProductBom, error)                                                       // 获取商品BOM详情
	PaginateGetProductUnitList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductUnit, int64, error)      // 分页获取产品单位列表
	GetProductUnit(opts ...DBOption) (model.ProductUnit, error)                                                     // 获取产品单位详情
	GetProductUnitCount(opts ...DBOption) (int64, error)                                                            // 获取产品单位数量

	PaginateGetProductSauceList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductSauce, int64, error) // 分页获取商品加料列表
	GetProductSauce(opts ...DBOption) (model.ProductSauce, error)                                                // 获取商品加料详情
	GetProductSauceCount(opts ...DBOption) (int64, error)                                                        // 获取商品加料数量
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
func (r *productRepo) defaultPreload(hasPackage bool) []DBOption {

	preloads := []DBOption{
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

	// 如果有商品套餐，则添加商品套餐预加载
	if hasPackage {
		// 预加载商品套餐
		packagePreloads := []DBOption{
			// 预加载商品套餐
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups",
					Args: []any{
						CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					},
				},
			)),
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups.MultiLanguageName",
				},
			)),
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups.ProductPackageGroupItems",
					Args: []any{
						CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					},
				},
			)),
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups.ProductPackageGroupItems.ProductBom.ProductFlavor.MultiLanguageName",
				},
			)),
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups.ProductPackageGroupItems.ProductPackage.MultiLanguageName",
				},
			)),
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups.ProductPackageGroupItems.ProductPackage.ProductUnit.MultiLanguageName",
				},
			)),
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups.ProductPackageGroupItems.ProductPackage.ProductPackageAttributeGroups.ProductAttributeGroup.MultiLanguageName",
				},
			)),
			CommonRepo.DBOption(CommonRepo.Preload(
				WithPreload{
					Query: "ProductPackageGroups.ProductPackageGroupItems.ProductPackage.ProductPackageAttributeGroups.ProductPackageAttributes.Attribute.MultiLanguageName",
				},
			)),
		}
		preloads = append(preloads, packagePreloads...)
	}

	return preloads
}

// 查询商家是否存在商品套餐
func (r *productRepo) HasProductPackage() (bool, error) {
	var productPackage model.ProductPackage
	db := r.db.Model(&model.ProductPackage{}).Session(&gorm.Session{}).Where("product_type = ?", constant.ProductTypePackage)
	err := db.First(&productPackage).Error
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return false, nil
		}
		return false, errors.WithMessage(err)
	}

	return true, nil
}

// GetProductListWithPagination 分页获取商品列表
func (r *productRepo) GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) {
	var products []model.ProductPackage
	var total int64

	db := r.db.Model(&model.ProductPackage{}).Session(&gorm.Session{})

	// 查询商家是否存在商品套餐
	hasPackage, errPackage := r.HasProductPackage()
	if errPackage != nil {
		return nil, 0, errors.WithMessage(errPackage)
	}
	opts = append(r.defaultPreload(hasPackage), opts...)

	for _, opt := range opts {
		db = opt(db)
	}

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
	db := r.db.Model(&model.ProductBom{}).Scopes(NotDeleted).
		Where("product_package_uuid not in (?)", r.db.Model(&model.ProductPackage{}).Select("uuid").Where("delete_time > 0")).Session(&gorm.Session{}).Where("is_sold_out = 1")
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

// GetProductCategoryCount 获取产品分类数量
func (r *productRepo) GetProductCategoryCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.ProductCategory{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	return total, errors.WithMessage(err)
}

// GetProductCategoryMaxSort 获取产品分类最大排序
func (r *productRepo) GetProductCategoryMaxSort(opts ...DBOption) (int64, error) {
	var sort sql.NullInt64
	db := r.db.Model(&model.ProductCategory{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Select("MAX(sort) as sort").Find(&sort).Error
	return sort.Int64, errors.WithMessage(err)
}

// GetProductCategory 获取产品分类详情
func (r *productRepo) GetProductCategory(opts ...DBOption) (model.ProductCategory, error) {
	var productCategory model.ProductCategory
	db := r.db.Model(&model.ProductCategory{})
	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&productCategory).Error
	return productCategory, errors.WithMessage(err)
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

// GetProductCount 获取商品数量
func (r *productRepo) GetProductCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.ProductPackage{})
	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Count(&total).Error
	return total, errors.WithMessage(err)
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
		return db.Preload("ProductPackageAttributeGroups", "delete_time = ?", constant.NotDeleted)
	}
}

// WithProductPackageAttributeGroupProductAttributeGroup 预加载产品包装属性组产品属性组
func (r *productRepo) WithProductPackageAttributeGroupProductAttributeGroup() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductAttributeGroup", "delete_time = ?", constant.NotDeleted)
	}
}

// WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName 预加载产品包装属性组产品属性组多语言名称
func (r *productRepo) WithProductPackageAttributeGroupProductAttributeGroupMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductAttributeGroup.MultiLanguageName", "delete_time = ?", constant.NotDeleted)
	}
}

// WithProductPackageAttributeGroupProductPackageAttributes 预加载产品包装属性组产品包装属性
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributes() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductPackageAttributes", "delete_time = ?", constant.NotDeleted)
	}
}

// WithProductPackageAttributeGroupProductPackageAttributesAttribute 预加载产品包装属性组产品包装属性属性
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributesAttribute() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductPackageAttributes.Attribute", "delete_time = ?", constant.NotDeleted)
	}
}

// WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName 预加载产品包装属性组产品包装属性属性多语言名称
func (r *productRepo) WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributeGroups.ProductPackageAttributes.Attribute.MultiLanguageName", "delete_time = ?", constant.NotDeleted)
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

// GetProductUnitList 获取产品单位列表
func (r *productRepo) PaginateGetProductUnitList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductUnit, int64, error) {
	var units []model.ProductUnit
	var total int64

	// 关联了商品，要连表join获取关联商品数量
	db := r.db.Model(&model.ProductUnit{}).Where("ttpos_product_unit.delete_time = ?", constant.NotDeleted)
	db = db.Joins("LEFT JOIN ttpos_product_package ON ttpos_product_package.unit_uuid = ttpos_product_unit.uuid AND ttpos_product_package.delete_time = ?", constant.NotDeleted)
	db = db.Select("ttpos_product_unit.*, COUNT(ttpos_product_package.uuid) as product_package_count")
	db = db.Group("ttpos_product_unit.uuid") // 分组统计关联商品数量

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("sort asc, create_time asc").Find(&units).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	return units, total, nil
}

// WithProductPackages 预加载产品单位关联的商品
func (r *productRepo) WithProductPackages() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackages", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

// WithProductPackagesMultiLanguageName 预加载产品单位关联的商品多语言名称
func (r *productRepo) WithProductPackagesMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackages.MultiLanguageName")
	}
}

// GetProductUnit 获取产品单位详情
func (r *productRepo) GetProductUnit(opts ...DBOption) (model.ProductUnit, error) {
	var unit model.ProductUnit

	db := r.db.Model(&model.ProductUnit{}).Scopes(NotDeleted)

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&unit).Error

	return unit, errors.WithMessage(err)
}

// WhereUuid 根据uuid查询
func (r *productRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereUuidIn 根据uuid列表查询
func (r *productRepo) WhereUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid IN (?)", uuids)
	}
}

// WhereCategoryKey 根据分类key查询
func (r *productRepo) WhereCategoryKey(key string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("category_key = ?", key)
	}
}

// WhereByIsSpecial 根据是否特殊查询
func (r *productRepo) WhereByIsSpecial(isSpecial uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_special = ?", isSpecial)
	}
}

// WhereParentUuid 根据父级分类uuid查询
func (r *productRepo) WhereParentUuid(parentUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("parent_uuid = ?", parentUuid)
	}
}

// WhereCategoryUuid 根据分类uuid查询
func (r *productRepo) WhereCategoryUuid(categoryUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("category_uuid = ?", categoryUuid)
	}
}

// WhereSpecialCategoryUuid 根据特色分类uuid查询
func (r *productRepo) WhereSpecialCategoryUuid(specialCategoryUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("special_category_uuid = ?", specialCategoryUuid)
	}
}

// GetProductUnitCount 获取产品单位数量
func (r *productRepo) GetProductUnitCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.ProductUnit{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	return total, errors.WithMessage(err)
}

// PaginateGetProductSauceList 分页获取商品加料列表
func (r *productRepo) PaginateGetProductSauceList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductSauce, int64, error) {
	var sauces []model.ProductSauce
	var total int64
	db := r.db.Model(&model.ProductSauce{}).Where("ttpos_product_sauce.delete_time = ?", constant.NotDeleted)
	db = db.Joins("LEFT JOIN ttpos_product_bom ON ttpos_product_bom.product_sauce_uuid = ttpos_product_sauce.uuid AND ttpos_product_bom.delete_time = ?", constant.NotDeleted)
	db = db.Select("ttpos_product_sauce.*, COUNT(ttpos_product_bom.uuid) as product_package_count")
	db = db.Group("ttpos_product_sauce.uuid") // 分组统计关联商品数量
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&sauces).Error
	return sauces, total, errors.WithMessage(err)
}

// GetProductSauce 获取商品加料详情
func (r *productRepo) GetProductSauce(opts ...DBOption) (model.ProductSauce, error) {
	var sauce model.ProductSauce
	db := r.db.Model(&model.ProductSauce{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&sauce).Error
	return sauce, errors.WithMessage(err)
}

// GetProductSauceCount 获取商品加料数量
func (r *productRepo) GetProductSauceCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.ProductSauce{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	return total, errors.WithMessage(err)
}

func (r *productRepo) WithActiveProductBoms() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) WithActiveProductBomsProductPackages() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms.ProductPackage", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) WithActiveProductBomsProductPackagesMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms.ProductPackage.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}
