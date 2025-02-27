package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderProductAttributeRepo 销售订单商品属性仓库接口
type IOrderProductAttributeRepo interface {
	CreateBatch(attributes []*model.SaleOrderProductAttribute) error // 批量创建
}

// orderProductAttributeRepo 销售订单商品属性仓库
type orderProductAttributeRepo struct {
	db *gorm.DB
}

// NewOrderProductAttributeRepo 创建销售订单商品属性仓库
func NewOrderProductAttributeRepo(db *gorm.DB) IOrderProductAttributeRepo {
	return NewOrderProductAttributeRepoImpl(db)
}

// NewOrderProductAttributeRepoImpl 创建销售订单商品属性仓库实现
func NewOrderProductAttributeRepoImpl(db *gorm.DB) IOrderProductAttributeRepo {
	return &orderProductAttributeRepo{db: db}
}

// CreateBatch 批量创建销售订单商品属性
func (o *orderProductAttributeRepo) CreateBatch(attributes []*model.SaleOrderProductAttribute) error {
	return o.db.Create(&attributes).Error
}
