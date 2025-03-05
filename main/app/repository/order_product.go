package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IOrderProductRepo 定义商品仓库接口
type IOrderProductRepo interface {
	WhereSaleBillUuids(uuids []uint64) DBOption
	WhereSaleOrderUuids(uuids []uint64) DBOption
	GetProductList(opts ...DBOption) ([]model.SaleOrderProduct, error) // 获取商品列表
	GetProductInfoByUuid(uuid uint64) (model.SaleOrderProduct, error)  // 通过UUID获取商品详情
	GetProductInfo(opts ...DBOption) (*model.SaleOrderProduct, error)  // 获取商品详情
	Delete(uuid uint64) error
	Create(model *model.SaleOrderProduct) (*model.SaleOrderProduct, error) // 创建订单商品
	Update(data map[string]interface{}, opts ...DBOption) error            // 更新订单商品
}

// orderProductRepo 商品仓库
type orderProductRepo struct {
	db *gorm.DB
}

// NewOrderProductRepo 创建新的商品仓库
func NewOrderProductRepo(db *gorm.DB) IOrderProductRepo {
	return NewOrderProductRepoImpl(db)
}

// NewOrderProductRepoImpl 创建新的商品仓库实现
func NewOrderProductRepoImpl(db *gorm.DB) IOrderProductRepo {
	return &orderProductRepo{db: db}
}

// WhereSaleBillUuids 根据sale_bill_uuids查询
func (r *orderProductRepo) WhereSaleBillUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_bill_uuid in (?)", uuids)
	}
}

// WhereSaleOrderUuids 根据sale_order_uuids查询
func (r *orderProductRepo) WhereSaleOrderUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_order_uuid in (?)", uuids)
	}
}

// GetProductList 获取商品列表
func (r *orderProductRepo) GetProductList(opts ...DBOption) ([]model.SaleOrderProduct, error) {
	var products []model.SaleOrderProduct

	db := r.db.Model(&model.SaleOrderProduct{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	// 获取列表
	err := db.Find(&products).Error

	return products, errors.WithMessage(err)
}

// Delete 软删除产品
func (r *orderProductRepo) Delete(uuid uint64) error {
	return r.db.Model(&model.SaleOrderProduct{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// Create 创建订单产品
func (r *orderProductRepo) Create(model *model.SaleOrderProduct) (*model.SaleOrderProduct, error) {
	err := r.db.Create(model).Error

	return model, errors.WithMessage(err)
}

// GetProductInfoByUuid 获取商品信息
func (r *orderProductRepo) GetProductInfoByUuid(uuid uint64) (model.SaleOrderProduct, error) {
	var product model.SaleOrderProduct

	err := r.db.Where("uuid = ?", uuid).
		Preload("MultiLanguageName").
		Preload("SaleOrderProductAttributes").
		Preload("SaleOrderProductBoms").
		First(&product).Error

	return product, errors.WithMessage(err)
}

// GetProductInfo 获取商品详情
func (r *orderProductRepo) GetProductInfo(opts ...DBOption) (*model.SaleOrderProduct, error) {
	var product model.SaleOrderProduct

	db := r.db.Model(&model.SaleOrderProduct{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&product).Error

	if err != nil && !utils.IsNotFoundRecord(err) {
		return &product, errors.WithMessage(err)
	}

	return &product, nil
}

// Update 更新
func (r *orderProductRepo) Update(data map[string]interface{}, opts ...DBOption) error {
	db := r.db.Model(&model.SaleOrderProduct{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Updates(data).Error

	return errors.WithMessage(err)
}
