package repository

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductionOrderRepo interface {
	IProductionOrderQueryRepo
	CreateProductionOrder(order *model.ProductionOrder) error // 创建生产订单
	WhereProductStatus(status uint) DBOption                  // 生产商品状态
	WhereUuid(uuid uint64) DBOption                           // Uuid 条件
	WhereProductPackageUuidIn(uuids []uint64) DBOption        // 商品ID条件
	WhereProductFinishedTime(finishedTime int64) DBOption     // 生产商品完成时间条件
	WhereProductMadeTime(madeTime int64) DBOption             // 生产商品制作时间条件
	WhereProductMakeStatus(finishStatus []uint) DBOption      // 制作状态
	WhereProductUuid(uuid uint64) DBOption                    // 生产商品Uuid条件
	WhereSaleOrderProductUuid(uuid uint64) DBOption           // 生产商品销售订单uuid条件
	WhereSaleBillUuidIn(uuids []uint64) DBOption              // 销售账单uuid条件
	WhereSaleBillUuid(uuid uint64) DBOption                   // 销售账单uuid条件
	WhereSource(source string) DBOption                       // 来源条件
	WhereProductFirstCategoryUuidIn(uuids []uint64) DBOption  // 生产商品分类Uuid条件
	WhereProductNumGT0() DBOption                             // 送厨商品数量大于0

	SaleBillUuidOpt() DBOption                                // 历史上菜条件
	WithSaleOrderProductAll() DBOption                        // 关联销售订单商品
	WithProductCategory() DBOption                            // 关联商品分类
	WithProductCategoryMultiLanguageName() DBOption           // 关联商品分类多语言
	UpdateProduct(opts []DBOption, vars map[string]any) error // 更新送厨商品
	UpdateOrder(opts []DBOption, vars map[string]any) error   // 更新送厨单
}

// IProductionOrderQueryRepo 生产订单查询仓库接口
type IProductionOrderQueryRepo interface {
	GetProductionOrder(opts ...DBOption) (*model.ProductionOrder, error)                                                          // 获取生产订单
	GetProduct(opts ...DBOption) (*model.ProductionOrderProduct, error)                                                           // 获取生产订单商品
	GetLimitedProducts(column string, pageNo, pageSize int, opts ...DBOption) ([]model.ProductionOrderProduct, int64, error)      // 分页获取账单ID、分类ID
	GetLimitedHistoryProducts(orderField string, opts ...DBOption) ([]model.ProductionOrderProduct, error)                        // 历史获取销售账单Uuid
	GetProducts(limit int, orderBy string, statusOpt DBOption, opts ...DBOption) (float64, []model.ProductionOrderProduct, error) // 获取生产订单商品

	GetProductsByPackageUuid(packageUuid uint64) ([]model.ProductionOrderProduct, error) // 根据套餐uuid获取套餐下所有子商品
}

type productionRepo struct {
	db *gorm.DB
}

func NewProductionRepo(db *gorm.DB) IProductionOrderRepo {
	return NewProductionOrderRepoImpl(db)
}

func NewProductionOrderRepoImpl(db *gorm.DB) IProductionOrderRepo {
	return &productionRepo{db: db}
}

func (r *productionRepo) GetProductionOrder(opts ...DBOption) (*model.ProductionOrder, error) {
	var productionOrder model.ProductionOrder
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Order("create_time desc").First(&productionOrder)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productionOrder, nil
}

func (r *productionRepo) GetProduct(opts ...DBOption) (*model.ProductionOrderProduct, error) {
	var product model.ProductionOrderProduct
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&product)
	if result.Error != nil {
		return nil, result.Error
	}

	return &product, nil
}

const (
	CreateTimeAsc    string = "create_time asc"
	FinishedTimeDesc string = "finished_time desc"
	MadeTimeDesc     string = "made_time desc"
)

func (r *productionRepo) GetProducts(limit int, orderBy string, statusOpt DBOption, opts ...DBOption) (float64, []model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	if statusOpt != nil {
		db = statusOpt(db).Session(&gorm.Session{})
	} else {
		db = db.Session(&gorm.Session{})
	}

	var total float64
	// 统计商品数量总和
	db.Select("IFNULL(SUM(num), 0) as total").Scan(&total)

	for _, opt := range opts {
		db = opt(db)
	}
	db.Preload("SaleBill").
		Preload("SaleOrderProduct").
		Preload("SaleOrderProduct.MultiLanguageName").
		Preload("SaleOrderProduct.SaleOrderProductBoms", NotDeleted).
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce").
		Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName").
		Preload("SaleOrderProduct.SaleOrderProductAttributes", NotDeleted).
		Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute").
		Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName").Order(orderBy)
	if limit > 0 {
		db.Limit(limit)
	}
	result := db.Find(&productionOrderProducts)
	if result.Error != nil {
		return total, nil, result.Error
	}
	return total, productionOrderProducts, nil
}

// CreateProductionOrder 创建ProductionOrder记录及它管理的表记录
func (r *productionRepo) CreateProductionOrder(order *model.ProductionOrder) error {
	if err := r.db.Model(order).Create(order).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// GetLimitedProducts 分页获取账单ID、分类ID
func (r *productionRepo) GetLimitedProducts(column string, pageNo, pageSize int, opts ...DBOption) ([]model.ProductionOrderProduct, int64, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	var total int64
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted).Session(&gorm.Session{})
	for _, opt := range opts {
		db = opt(db)
	}
	// 获取总数
	err := db.Select(fmt.Sprintf("count(distinct `%s`) as total", column)).Scan(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	// 获取列表
	err = db.Select("DISTINCT " + column).Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("create_time asc").Find(&productionOrderProducts).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	return productionOrderProducts, total, nil
}

// GetLimitedHistoryProducts 历史获取销售账单Uuid
func (r *productionRepo) GetLimitedHistoryProducts(orderField string, opts ...DBOption) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Model(&model.ProductionOrderProduct{}).Select(fmt.Sprintf("sale_bill_uuid, MAX(%s) as finished_time", orderField)).Group("sale_bill_uuid").Order(orderField + " desc").Find(&productionOrderProducts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productionOrderProducts, nil
}

// SaleBillUuidOpt 历史上菜条件
func (r *productionRepo) SaleBillUuidOpt() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid in (?)", r.db.Model(&model.ProductionOrderProduct{}).Select("DISTINCT sale_bill_uuid"))
	}
}

// WhereProductStatus 状态条件
func (r *productionRepo) WhereProductStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereUuid uuid条件
func (r *productionRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereProductPackageUuidIn 商品ID条件
func (r *productionRepo) WhereProductPackageUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_package_uuid in (?)", uuids)
	}
}

// WhereProductFinishedTime 生产商品完成时间条件
func (r *productionRepo) WhereProductFinishedTime(finishedTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("finished_time > ?", finishedTime)
	}
}

// WhereProductMadeTime 生产商品制作时间条件
func (r *productionRepo) WhereProductMadeTime(madeTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("made_time > ?", madeTime)
	}
}

// WhereProductMakeStatus 制作状态条件
func (r *productionRepo) WhereProductMakeStatus(makeStatus []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("make_status in (?)", makeStatus)
	}
}

// WhereProductUuid 生产商品Uuid条件
func (r *productionRepo) WhereProductUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereSaleOrderProductUuid 生产商品销售订单商品Uuid条件
func (r *productionRepo) WhereSaleOrderProductUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_order_product_uuid = ?", uuid)
	}
}

// WhereSaleBillUuidIn 销售账单uuid条件
func (r *productionRepo) WhereSaleBillUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid in (?)", uuids)
	}
}

// WhereSaleBillUuid 销售账单uuid条件
func (r *productionRepo) WhereSaleBillUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid = ?", uuid)
	}
}

// WhereSource 来源
func (r *productionRepo) WhereSource(source string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source = ?", source)
	}
}

// WhereProductFirstCategoryUuidIn 生产商品分类Uuid条件
func (r *productionRepo) WhereProductFirstCategoryUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("first_category_uuid in (?)", uuids)
	}
}

// WhereProductNumGT0 送厨商品数量大于0
func (r *productionRepo) WhereProductNumGT0() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("num > 0")
	}
}

// WithSaleBill 关联销售账单
func (r *productionRepo) WithSaleBill() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleBill")
	}
}

// WithSaleOrderProductAll 关联销售订单商品
func (r *productionRepo) WithSaleOrderProductAll() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrderProduct.MultiLanguageName").
			Preload("SaleOrderProduct.SaleOrderProductBoms", NotDeleted).
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce").
			Preload("SaleOrderProduct.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName").
			Preload("SaleOrderProduct.SaleOrderProductAttributes", NotDeleted).
			Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute").
			Preload("SaleOrderProduct.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName")
	}
}

// WithProductCategory 关联商品分类
func (r *productionRepo) WithProductCategory() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory")
	}
}

// WithProductCategoryMultiLanguageName 关联商品分类多语言
func (r *productionRepo) WithProductCategoryMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductCategory.MultiLanguageName")
	}
}

// UpdateProduct 修改
func (r *productionRepo) UpdateProduct(opts []DBOption, vars map[string]any) error {
	db := r.db.Model(&model.ProductionOrderProduct{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

// UpdateOrder 修改
func (r *productionRepo) UpdateOrder(opts []DBOption, vars map[string]any) error {
	db := r.db.Model(&model.ProductionOrder{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

func (r *productionRepo) GetProductsByPackageUuid(packageUuid uint64) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	err := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted).
		Where("sale_order_product_uuid in (?)", r.db.Model(&model.SaleOrderProduct{}).Select("uuid").Where("package_uuid = ?", packageUuid).Scopes(NotDeleted)).
		Find(&productionOrderProducts).Error

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return productionOrderProducts, nil
}
