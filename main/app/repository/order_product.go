package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderProductRepo 定义商品仓库接口
type IOrderProductRepo interface {
	WhereSaleBillUuids(uuids []uint64) DBOption
	WhereSaleOrderUuids(uuids []uint64) DBOption
	GetProductList(opts ...DBOption) ([]model.SaleOrderProduct, error) // 获取商品列表
	Delete(uuid uint64) error
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

	return products, err
}

// DeleteProductAttribute 软删除产品
func (r *orderProductRepo) Delete(uuid uint64) error {
	return r.db.Model(&model.SaleOrderProduct{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
