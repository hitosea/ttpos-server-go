package repository

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductionOrderRepo interface {
	GetProductionOrder(opts ...DBOption) (*model.ProductionOrder, error)                                                     // 获取生产订单
	GetProduct(opts ...DBOption) (*model.ProductionOrderProduct, error)                                                      // 获取生产订单商品
	CreateProductionOrder(order *model.ProductionOrder) error                                                                // 创建生产订单
	GetLimitedProducts(column string, pageNo, pageSize int, opts ...DBOption) ([]model.ProductionOrderProduct, int64, error) // 分页获取账单ID、分类ID
	GetLimitedHistoryProducts(opts ...DBOption) ([]model.ProductionOrderProduct, error)                                      // 历史获取销售账单Uuid
	GetProducts(opts ...DBOption) ([]model.ProductionOrderProduct, error)                                                    // 获取生产订单商品
	GetFinishedLimitProducts(limit int) ([]model.ProductionOrderProduct, error)                                              // 获取完成的生产订单商品
	WhereProductStatus(status uint) DBOption                                                                                 // 生产商品状态
	WhereProductFinishedTime(finishedTime int64) DBOption                                                                    // 生产商品完成时间条件
	WhereProductUuid(uuid uint64) DBOption                                                                                   // 生产商品Uuid条件
	WhereProductSaleBillUuidIn(uuids []uint64) DBOption                                                                      // 生产商品销售账单uuid条件
	WhereProductFirstCategoryUuidIn(uuids []uint64) DBOption                                                                 // 生产商品分类Uuid条件
	WhereProductHistoryCondition() DBOption                                                                                  // 历史上菜条件
	WithSaleBill() DBOption                                                                                                  // 关联销售账单
	WithProductCategory() DBOption                                                                                           // 关联商品分类
	WithProductCategoryMultiLanguageName() DBOption                                                                          // 关联商品分类多语言
	UpdateProduct(uuid uint64, vars map[string]any) error                                                                    // 更新送厨商品
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

	result := db.First(&productionOrder)
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

func (r *productionRepo) GetProducts(opts ...DBOption) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	db := r.db.Model(&model.ProductionOrderProduct{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	result := db.Order("create_time asc").Find(&productionOrderProducts)
	if result.Error != nil {
		return nil, result.Error
	}
	return productionOrderProducts, nil
}

func (r *productionRepo) GetFinishedLimitProducts(limit int) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	result := r.db.Model(&model.ProductionOrderProduct{}).Preload("SaleBill").Scopes(NotDeleted).
		Where("status = ?", constant.ProductionOrderProductStatusFinished).
		Order("finished_time desc").Limit(limit).Find(&productionOrderProducts)
	if result.Error != nil {
		return nil, result.Error
	}
	return productionOrderProducts, nil
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
	db := r.db.Model(&model.ProductionOrderProduct{}).Session(&gorm.Session{})
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
func (r *productionRepo) GetLimitedHistoryProducts(opts ...DBOption) ([]model.ProductionOrderProduct, error) {
	var productionOrderProducts []model.ProductionOrderProduct
	db := r.db.Model(&model.ProductionOrderProduct{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Model(&model.ProductionOrderProduct{}).Select("DISTINCT sale_order_uuid").Order("finished_time desc").Find(&productionOrderProducts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productionOrderProducts, nil
}

// WhereProductHistoryCondition 历史上菜条件
func (r *productionRepo) WhereProductHistoryCondition() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_order_id in (?)", r.db.Model(&model.ProductionOrderProduct{}).Select("DISTINCT sale_order_uuid").Order("finished_time desc"))
	}
}

// WhereProductStatus 状态条件
func (r *productionRepo) WhereProductStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereProductFinishedTime 生产商品完成时间条件
func (r *productionRepo) WhereProductFinishedTime(finishedTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("finished_time = ?", finishedTime)
	}
}

// WhereProductUuid 生产商品Uuid条件
func (r *productionRepo) WhereProductUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereProductSaleBillUuidIn 生产商品销售账单uuid条件
func (r *productionRepo) WhereProductSaleBillUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid in (?)", uuids)
	}
}

// WhereProductFirstCategoryUuidIn 生产商品分类Uuid条件
func (r *productionRepo) WhereProductFirstCategoryUuidIn(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("first_category_uuid in (?)", uuids)
	}
}

// WithSaleBill 关联销售账单
func (r *productionRepo) WithSaleBill() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleBill")
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
func (r *productionRepo) UpdateProduct(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.ProductionOrderProduct{}).Where("uuid = ?", uuid).Updates(vars).Error
}
