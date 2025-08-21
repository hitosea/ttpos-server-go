package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

// IProductRepo 定义商品仓库接口
type IProductRepo interface {
	IProductQueryRepo
	WithMultiLanguageName() DBOption                                                              // 预加载多语言名称
	WithProductUnit() DBOption                                                                    // 预加载产品单位
	WithProductUnitMultiLanguageName() DBOption                                                   // 预加载产品单位多语言名称
	WithProductBoms(opts ...DBOption) DBOption                                                    // 预加载产品Boms
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
	WithProductCategoryMultiLanguageName() DBOption                                               // 预加载分类多语言名称
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

	WhereUuid(uuid uint64) DBOption              // 查询条件 产品单位 uuid
	WhereUuidIn(uuids []uint64) DBOption         // 查询条件 产品单位 uuid 列表
	WhereCategoryKey(key string) DBOption        // 查询条件 产品分类key
	WhereByIsSpecial(isSpecial uint) DBOption    // 查询条件 产品分类是否特殊
	WhereParentUuid(parentUuid uint64) DBOption  // 查询条件 产品分类父级uuid
	WhereProductSauceUuid(uuid uint64) DBOption  // 查询条件 产品加料uuid
	WhereProductType(productType uint8) DBOption // 查询条件 产品类型

	WhereCategoryUuid(categoryUuid uint64) DBOption               // 查询条件 产品分类uuid
	WhereSpecialCategoryUuid(specialCategoryUuid uint64) DBOption // 查询条件 特色分类uuid
	WhereHasMultipleSpec() DBOption                               // 查询条件 是否多规格
	WhereHasAttribute() DBOption                                  // 查询条件 是否属性
	WhereHasSauce() DBOption                                      // 查询条件 是否加料

	WithProductPackages() DBOption                  // 预加载产品单位关联的商品
	WithProductPackagesMultiLanguageName() DBOption // 预加载产品单位关联的商品多语言名称

	WithActiveProductBoms(opts ...DBOption) DBOption                 // 预加载商品BOM
	WithActiveProductBomsProductPackages() DBOption                  // 预加载商品BOM关联的商品包
	WithActiveProductBomsProductPackagesMultiLanguageName() DBOption // 预加载商品BOM关联的商品包多语言名称

	WithProductAttributes() DBOption                                                                                    // 预加载商品属性
	WithProductAttributesMultiLanguageName() DBOption                                                                   // 预加载商品属性多语言名称
	WithProductAttributesProductPackageAttributes() DBOption                                                            // 预加载商品属性关联的产品包属性
	WithProductAttributesProductPackageAttributesProductPackageAttributeGroup() DBOption                                // 预加载商品属性关联的产品包属性组
	WithProductAttributesProductPackageAttributesProductPackageAttributeGroupProductPackage() DBOption                  // 预加载商品属性关联的产品包属性组关联的产品包
	WithProductAttributesProductPackageAttributesProductPackageAttributeGroupProductPackageMultiLanguageName() DBOption // 预加载商品属性关联的产品包属性组关联的产品包多语言名称

	WhereAttributeGroupUuid(uuid uint64) DBOption        // 查询条件 商品属性分组uuid
	WhereProductAttributeGroupUuid(uuid uint64) DBOption // 查询条件 商品属性分组uuid
	WithProductPackageAttributes() DBOption              // 预加载商品属性关联的产品包属性
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
	GetProductDetail(uuid uint64) (*model.ProductPackage, error)                                                    // 获取商品详情
	GetProductCount(opts ...DBOption) (int64, error)                                                                // 获取商品数量
	GetProductFlavor(opts ...DBOption) (model.ProductFlavor, error)                                                 // 获取商品口味详情
	GetProductFlavorList(opts ...DBOption) ([]model.ProductFlavor, error)                                           // 获取商品口味列表
	GetProductFlavorCount(opts ...DBOption) (int64, error)                                                          // 获取商品规格数量
	GetProductFlavorMaxSort(opts ...DBOption) (int64, error)                                                        // 获取商品规格最大排序
	GetProductBom(opts ...DBOption) (model.ProductBom, error)                                                       // 获取商品BOM详情
	GetProductBomCount(opts ...DBOption) (int64, error)                                                             // 获取商品BOM数量

	PaginateGetProductUnitList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductUnit, int64, error) // 分页获取产品单位列表
	GetProductUnitList(opts ...DBOption) ([]model.ProductUnit, error)                                          // 获取产品单位列表
	GetProductUnit(opts ...DBOption) (model.ProductUnit, error)                                                // 获取产品单位详情
	GetProductUnitCount(opts ...DBOption) (int64, error)                                                       // 获取产品单位数量
	GetProductUnitByUnitUuid(unitUuid uint64) (*model.ProductUnit, error)                                      // 获取产品单位详情

	PaginateGetProductSauceList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductSauce, int64, error) // 分页获取商品加料列表
	GetProductSauceList(opts ...DBOption) ([]model.ProductSauce, error)                                          // 获取商品加料列表
	GetProductSauce(opts ...DBOption) (model.ProductSauce, error)                                                // 获取商品加料详情
	GetProductSauceCount(opts ...DBOption) (int64, error)                                                        // 获取商品加料数量

	PaginateGetProductAttributeGroupList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductAttributeGroup, int64, error) // 分页获取商品属性分组列表
	GetProductAttributeGroups(opts ...DBOption) ([]model.ProductAttributeGroup, error)                                             // 获取商品属性分组列表
	GetProductAttributeGroup(opts ...DBOption) (model.ProductAttributeGroup, error)                                                // 获取商品属性分组详情
	GetProductAttributes(opts ...DBOption) ([]model.ProductAttribute, error)                                                       // 获取商品属性列表
	GetProductAttribute(opts ...DBOption) (model.ProductAttribute, error)                                                          // 获取商品属性详情
	GetProductPackageAttributeGroups(opts ...DBOption) ([]model.ProductPackageAttributeGroup, error)                               // 获取商品包属性组列表

	PaginateGetProductFlavorList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductFlavor, int64, error) // 分页获取商品规格列表
	CheckMultiLanguageNameExist(localeResponse dto.LocaleResponse) dto.LocaleResponse                              // 检查多语言名称是否存在
	CheckBarcodeExist(barcode string, uuid uint64) bool                                                            // 检查条形码是否存在
	CheckBarcodeFormat(barcode string) bool                                                                        // 检查条形码格式
	CheckPrice(price, minPrice, maxPrice float64, places int) bool                                                 // 检查价格范围

	PaginateGetProductShopList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) // 分页获取商品列表（商家端）
	GetProductShopList(opts ...DBOption) ([]model.ProductPackage, error)                                          // 获取商品列表（商家端）
	GetProductShopMaxSort(opts ...DBOption) (int64, error)                                                        // 获取商品最大排序

	BatchUpdateSort(table any, sorts map[uint64]int) error // 批量更新排序
}

// productRepo 商品仓库
type productRepo struct {
	db *gorm.DB
}

// getTableName 获取带前缀的表名
func (r *productRepo) getTableName(tableName string) string {
	return config.Database.TablePrefix + tableName
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

// GetProductDetail 获取商品详情
func (r *productRepo) GetProductDetail(uuid uint64) (*model.ProductPackage, error) {
	product, err := r.GetProduct(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
			WithPreload{
				Query: "ProductCategory.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductUnit.MultiLanguageName",
			},
			WithPreload{
				Query: "DineTax",
			},
			WithPreload{
				Query: "TakeoutTax",
			},
			WithPreload{
				Query: "ImageFile",
			},
			WithPreload{
				Query: "ProductBoms.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductBoms.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductPackageAttributeGroups",
			},
			WithPreload{
				Query: "ProductPackageAttributeGroups.ProductAttributeGroup",
			},
			WithPreload{
				Query: "ProductPackageAttributeGroups.ProductAttributeGroup.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductPackageAttributeGroups.ProductPackageAttributes.Attribute.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductPackageGroups.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductPackageGroups.ProductPackageGroupItems.ProductBom.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductPackageGroups.ProductPackageGroupItems.ProductPackage.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品失败")
	}

	return &product, nil
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

// GetProductFlavorCount 获取商品规格数量
func (r *productRepo) GetProductFlavorCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.ProductFlavor{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	return total, errors.WithMessage(err)
}

// GetProductFlavorMaxSort 获取商品规格最大排序
func (r *productRepo) GetProductFlavorMaxSort(opts ...DBOption) (int64, error) {
	var sort sql.NullInt64
	db := r.db.Model(&model.ProductFlavor{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Select("MAX(sort) as sort").Find(&sort).Error
	return sort.Int64, errors.WithMessage(err)
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

// GetProductBomCount 获取商品BOM数量
func (r *productRepo) GetProductBomCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.ProductBom{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	return total, errors.WithMessage(err)
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
func (r *productRepo) WithProductBoms(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
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
	productUnitTable := r.getTableName("product_unit")
	productPackageTable := r.getTableName("product_package")
	db := r.db.Model(&model.ProductUnit{}).Where(productUnitTable+".delete_time = ?", constant.NotDeleted)
	db = db.Joins("LEFT JOIN "+productPackageTable+" ON "+productPackageTable+".unit_uuid = "+productUnitTable+".uuid AND "+productPackageTable+".delete_time = ?", constant.NotDeleted)
	db = db.Select(productUnitTable + ".*, COUNT(" + productPackageTable + ".uuid) as product_package_count")
	db = db.Group(productUnitTable + ".uuid") // 分组统计关联商品数量

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

// GetProductUnitList 获取产品单位列表
func (r *productRepo) GetProductUnitList(opts ...DBOption) ([]model.ProductUnit, error) {
	var units []model.ProductUnit
	db := r.db.Model(&model.ProductUnit{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc, create_time asc").Find(&units).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return units, nil
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

// WhereProductSauceUuid 根据加料uuid查询
func (r *productRepo) WhereProductSauceUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_sauce_uuid = ?", uuid)
	}
}

// WhereProductType 根据产品类型查询
func (r *productRepo) WhereProductType(productType uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_type = ?", productType)
	}
}

// WhereIsMultipleSpec 根据是否多规格查询
func (r *productRepo) WhereHasMultipleSpec() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		subQuery := r.db.Select("COUNT(*)").Model(&model.ProductBom{}).Where("product_package_uuid = ttpos_product_package.uuid AND product_flavor_uuid > 0 AND delete_time = ?", constant.NotDeleted)
		return db.Where("(?) > 1", subQuery)
	}
}

// WhereIsAttribute 根据是否属性查询
func (r *productRepo) WhereHasAttribute() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		subQuery := r.db.Select("COUNT(*)").Model(&model.ProductPackageAttributeGroup{}).Where("product_package_uuid = ttpos_product_package.uuid AND delete_time = ?", constant.NotDeleted)
		return db.Where("(?) > 0", subQuery)
	}
}

// WhereIsSauce 根据是否加料查询
func (r *productRepo) WhereHasSauce() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		subQuery := r.db.Select("COUNT(*)").Model(&model.ProductBom{}).Where("product_package_uuid = ttpos_product_package.uuid AND product_sauce_uuid > 0 AND delete_time = ?", constant.NotDeleted)
		return db.Where("(?) > 0", subQuery)
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

// GetProductUnitByUnitUuid 根据单位uuid获取产品单位详情
func (r *productRepo) GetProductUnitByUnitUuid(unitUuid uint64) (*model.ProductUnit, error) {
	unit, err := r.GetProductUnit(
		CommonRepo.WhereByUuid(unitUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &unit, nil
}

// PaginateGetProductSauceList 分页获取商品加料列表
func (r *productRepo) PaginateGetProductSauceList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductSauce, int64, error) {
	var sauces []model.ProductSauce
	var total int64
	productSauceTable := r.getTableName("product_sauce")
	productBomTable := r.getTableName("product_bom")
	db := r.db.Model(&model.ProductSauce{}).Where(productSauceTable+".delete_time = ?", constant.NotDeleted)
	db = db.Joins("LEFT JOIN "+productBomTable+" ON "+productBomTable+".product_sauce_uuid = "+productSauceTable+".uuid AND "+productBomTable+".delete_time = ?", constant.NotDeleted)
	db = db.Select(productSauceTable + ".*, COUNT(" + productBomTable + ".uuid) as product_package_count")
	db = db.Group(productSauceTable + ".uuid") // 分组统计关联商品数量
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	err = db.Offset((pageNo - 1) * pageSize).Order(productSauceTable + ".sort asc, " + productSauceTable + ".create_time desc").Limit(pageSize).Find(&sauces).Error
	return sauces, total, errors.WithMessage(err)
}

func (r *productRepo) GetProductSauceList(opts ...DBOption) ([]model.ProductSauce, error) {
	var sauces []model.ProductSauce
	db := r.db.Model(&model.ProductSauce{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc, create_time asc").Find(&sauces).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return sauces, nil
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

func (r *productRepo) WithActiveProductBoms(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductBoms", func(db *gorm.DB) *gorm.DB {
			db = db.Scopes(NotDeleted)
			for _, opt := range opts {
				db = opt(db)
			}
			return db
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

func (r *productRepo) PaginateGetProductAttributeGroupList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductAttributeGroup, int64, error) {
	var attributeGroups []model.ProductAttributeGroup
	var total int64
	db := r.db.Model(&model.ProductAttributeGroup{}).Where("delete_time = ?", constant.NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	err = db.Offset((pageNo - 1) * pageSize).Order("sort asc, create_time asc").Limit(pageSize).Find(&attributeGroups).Error
	return attributeGroups, total, errors.WithMessage(err)
}

func (r *productRepo) WithProductAttributes() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductAttributes", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc, create_time asc").Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) WithProductAttributesMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductAttributes.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) GetProductAttributeGroup(opts ...DBOption) (model.ProductAttributeGroup, error) {
	var attributeGroup model.ProductAttributeGroup
	db := r.db.Model(&model.ProductAttributeGroup{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&attributeGroup).Error
	return attributeGroup, errors.WithMessage(err)
}

func (r *productRepo) WithProductAttributesProductPackageAttributes() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductAttributes.ProductPackageAttributes", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) WithProductAttributesProductPackageAttributesProductPackageAttributeGroup() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductAttributes.ProductPackageAttributes.ProductPackageAttributeGroup", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) WithProductAttributesProductPackageAttributesProductPackageAttributeGroupProductPackage() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductAttributes.ProductPackageAttributes.ProductPackageAttributeGroup.ProductPackage", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) WithProductAttributesProductPackageAttributesProductPackageAttributeGroupProductPackageMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductAttributes.ProductPackageAttributes.ProductPackageAttributeGroup.ProductPackage.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

// GetProductAttributes 获取商品属性列表
func (r *productRepo) GetProductAttributes(opts ...DBOption) ([]model.ProductAttribute, error) {
	var attributes []model.ProductAttribute
	db := r.db.Model(&model.ProductAttribute{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc, create_time asc").Find(&attributes).Error
	return attributes, errors.WithMessage(err)
}

// GetProductAttribute 获取商品属性详情
func (r *productRepo) GetProductAttribute(opts ...DBOption) (model.ProductAttribute, error) {
	var attribute model.ProductAttribute
	db := r.db.Model(&model.ProductAttribute{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&attribute).Error
	return attribute, errors.WithMessage(err)
}

// GetProductPackageAttributeGroups 获取产品包属性组列表
func (r *productRepo) GetProductPackageAttributeGroups(opts ...DBOption) ([]model.ProductPackageAttributeGroup, error) {
	var attributeGroups []model.ProductPackageAttributeGroup
	db := r.db.Model(&model.ProductPackageAttributeGroup{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&attributeGroups).Error
	return attributeGroups, errors.WithMessage(err)
}

func (r *productRepo) WhereProductAttributeGroupUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_attribute_group_uuid = ?", uuid)
	}
}

func (r *productRepo) WithProductPackageAttributes() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageAttributes", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

func (r *productRepo) GetProductAttributeGroups(opts ...DBOption) ([]model.ProductAttributeGroup, error) {
	var attributeGroups []model.ProductAttributeGroup
	db := r.db.Model(&model.ProductAttributeGroup{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc, create_time asc").Find(&attributeGroups).Error
	return attributeGroups, errors.WithMessage(err)
}

// PaginateGetProductFlavorList 分页获取商品规格列表
func (r *productRepo) PaginateGetProductFlavorList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductFlavor, int64, error) {
	var flavors []model.ProductFlavor
	var total int64
	productFlavorTable := r.getTableName("product_flavor")
	productBomTable := r.getTableName("product_bom")
	db := r.db.Model(&model.ProductFlavor{}).Where(productFlavorTable+".delete_time = ?", constant.NotDeleted)
	db = db.Joins("LEFT JOIN "+productBomTable+" ON "+productBomTable+".product_flavor_uuid = "+productFlavorTable+".uuid AND "+productBomTable+".product_sauce_uuid = 0 AND "+productBomTable+".delete_time = ?", constant.NotDeleted)
	db = db.Select(productFlavorTable + ".*, COUNT(" + productBomTable + ".uuid) as product_package_count")
	db = db.Group(productFlavorTable + ".uuid") // 分组统计关联商品数量
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Order(productFlavorTable + ".sort asc, " + productFlavorTable + ".id desc").Find(&flavors).Error
	return flavors, total, errors.WithMessage(err)
}

// CheckMultiLanguageNameExist 检查多语言名称是否存在，返回存在的语言名称
func (r *productRepo) CheckMultiLanguageNameExist(localeResponse dto.LocaleResponse) dto.LocaleResponse {
	var result dto.LocaleResponse
	productPackageTable := r.getTableName("product_package")
	multiLanguageNameTable := r.getTableName("multi_language_name")

	// 定义语言字段映射
	languageFields := map[string]string{
		"zh":   "zh_name",
		"th":   "th_name",
		"en":   "en_name",
		"zhtw": "zh_tw_name",
		"ja":   "ja_name",
		"ko":   "ko_name",
		"my":   "my_name",
		"tr":   "tr_name",
		"sv":   "sv_name",
	}

	// 构建动态查询条件
	var conditions []string
	var args []interface{}

	for langKey, columnName := range languageFields {
		value := localeResponse.GetLocale(langKey)
		if value != "" {
			conditions = append(conditions, multiLanguageNameTable+"."+columnName+" = ?")
			args = append(args, value)
		}
	}

	// 如果没有任何条件，直接返回空结果
	if len(conditions) == 0 {
		return result
	}

	// 查询匹配的多语言名称记录
	var matchedRecords []model.MultiLanguageName
	err := r.db.Model(&model.ProductPackage{}).
		Select(multiLanguageNameTable+".*").
		Joins("JOIN "+multiLanguageNameTable+" ON "+productPackageTable+".multi_language_name_uuid = "+multiLanguageNameTable+".uuid").
		Where(productPackageTable+".delete_time = ?", constant.NotDeleted).
		Where("("+strings.Join(conditions, " OR ")+")", args...).
		Find(&matchedRecords).Error

	if err != nil {
		return result
	}

	// 检查每个匹配的记录，设置存在的语言名称
	for _, record := range matchedRecords {
		for langKey := range languageFields {
			inputValue := localeResponse.GetLocale(langKey)
			recordValue := record.GetNameByLang(langKey)

			// 如果输入的名称与数据库中的名称匹配，则标记为存在
			if inputValue != "" && inputValue == recordValue {
				result.SetLocale(langKey, inputValue)
			}
		}
	}

	return result
}

// CheckBarcodeExist 检查条形码是否存在
func (r *productRepo) CheckBarcodeExist(barcode string, uuid uint64) bool {
	db := r.db.Model(&model.ProductBom{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("barcode_value = ?", barcode).
		Where("barcode_value <> ?", "")
	if uuid != 0 {
		db = db.Where("uuid <> ?", uuid)
	}
	return db.First(&model.ProductBom{}).Error == nil
}

// CheckBarcodeFormat 检查条形码格式
func (r *productRepo) CheckBarcodeFormat(barcode string) bool {
	return regexp.MustCompile(`^[0-9]{1,13}$`).MatchString(barcode)
}

// CheckPrice 检查价格范围
func (r *productRepo) CheckPrice(price, minPrice, maxPrice float64, places int) bool {
	if price < minPrice || price > maxPrice {
		return false
	}
	if places > 0 {
		priceStr := fmt.Sprintf("%.2f", price)
		parts := strings.Split(priceStr, ".")
		if len(parts) == 2 && len(parts[1]) > places {
			return false
		}
	}

	return true
}

// PaginateGetProductShopList 分页获取商品列表（商家端）
func (r *productRepo) PaginateGetProductShopList(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) {
	var products []model.ProductPackage
	var total int64

	db := r.db.Model(&model.ProductPackage{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&products).Error
	return products, total, errors.WithMessage(err)
}

// WithProductCategoryMultiLanguageName 预加载分类多语言名称
func (r *productRepo) WithProductCategoryMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory.MultiLanguageName", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

// GetProductShopList 获取商品列表（商家端）
func (r *productRepo) GetProductShopList(opts ...DBOption) ([]model.ProductPackage, error) {
	var products []model.ProductPackage
	db := r.db.Model(&model.ProductPackage{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&products).Error
	return products, errors.WithMessage(err)
}

func (r *productRepo) WhereAttributeGroupUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("attribute_group_uuid = ?", uuid)
	}
}

// GetProductShopMaxSort 获取商品最大排序
func (r *productRepo) GetProductShopMaxSort(opts ...DBOption) (int64, error) {
	var sort sql.NullInt64
	db := r.db.Model(&model.ProductPackage{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Select("MAX(sort) as sort").Find(&sort).Error
	return sort.Int64, errors.WithMessage(err)
}

func (r *productRepo) BatchUpdateSort(table any, sorts map[uint64]int) error {
	// 检查是否有数据需要更新
	if len(sorts) == 0 {
		return nil
	}
	// 构建 CASE WHEN 语句
	caseWhenSQL := "CASE uuid"
	var args []any
	var uuids []uint64

	for uuid, sort := range sorts {
		caseWhenSQL += " WHEN ? THEN ?"
		args = append(args, uuid, sort)
		uuids = append(uuids, uuid)
	}
	caseWhenSQL += " END"

	// 根据传入的模型类型确定错误消息
	var errorMessage string
	switch table.(type) {
	case *model.ProductUnit, *model.ProductAttributeGroup, *model.ProductAttribute, *model.ProductSauce, *model.ProductFlavor, *model.ProductCategory:
		// 无需处理
	default:
		return errors.New("更新排序失败")
	}
	// 一条SQL语句批量更新排序
	err := r.db.Model(table).
		Where("uuid IN ?", uuids).Debug().
		Update("sort", gorm.Expr(caseWhenSQL, args...)).Error

	if err != nil {
		return errors.WithMessage(errors.New("更新排序失败"), errorMessage)
	}
	return nil
}

func (r *productRepo) GetProductFlavorList(opts ...DBOption) ([]model.ProductFlavor, error) {
	var flavors []model.ProductFlavor
	db := r.db.Model(&model.ProductFlavor{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc, create_time asc").Find(&flavors).Error
	return flavors, errors.WithMessage(err)
}
