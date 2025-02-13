package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

// IOrderProductBomRepo 定义销售订单商品BOM仓库接口
type IOrderProductBomRepo interface {
	CreateBatch(products []model.SaleOrderProductBom) error // 批量创建
}

// orderProductBomRepo 销售订单商品BOM仓库
type orderProductBomRepo struct {
	db *gorm.DB
}

// NewOrderProductBomRepo 实例化销售订单商品BOM仓库
func NewOrderProductBomRepo(db *gorm.DB) IOrderProductBomRepo {
	return NewOrderProductBomRepoImpl(db)
}

// NewOrderProductBomRepoImpl 实例化销售订单商品BOM仓库实现
func NewOrderProductBomRepoImpl(db *gorm.DB) IOrderProductBomRepo {
	return &orderProductBomRepo{db: db}
}

// CreateBatch 批量创建
func (o *orderProductBomRepo) CreateBatch(products []model.SaleOrderProductBom) error {
	return o.db.Create(&products).Error
}
