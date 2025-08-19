package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ISaleOrderProductAttributeRepo 定义销售订单商品属性仓库接口
type ISaleOrderProductAttributeRepo interface {
	CreateSaleOrderProductAttributeRecord(model model.SaleOrderProductAttribute) error
	UpdateSaleOrderProductAttributeRecord(model model.SaleOrderProductAttribute) error
	UpdateOrCreateSaleOrderProductAttributeRecord(model model.SaleOrderProductAttribute) error
}

// saleOrderProductAttributeRepo 销售订单商品属性仓库
type saleOrderProductAttributeRepo struct {
	db *gorm.DB
}

// NewSaleOrderProductAttributeRepo 实例化销售订单商品属性仓库
func NewSaleOrderProductAttributeRepo(db *gorm.DB) ISaleOrderProductAttributeRepo {
	return NewSaleOrderProductAttributeRepoImpl(db)
}

// NewSaleOrderProductAttributeRepoImpl 实例化销售订单商品属性仓库实现
func NewSaleOrderProductAttributeRepoImpl(db *gorm.DB) ISaleOrderProductAttributeRepo {
	return &saleOrderProductAttributeRepo{db: db}
}

// CreateSaleOrderProductAttributeRecord 创建销售订单商品属性记录
func (r *saleOrderProductAttributeRepo) CreateSaleOrderProductAttributeRecord(obj model.SaleOrderProductAttribute) error {
	obj.SetNil()
	return r.db.Model(&model.SaleOrderProductAttribute{}).Create(&obj).Error
}

// UpdateSaleOrderProductAttributeRecord 更新销售订单商品属性记录
func (r *saleOrderProductAttributeRepo) UpdateSaleOrderProductAttributeRecord(obj model.SaleOrderProductAttribute) error {
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return errors.WithMessage(errors.New("销售订单商品属性UUID或ID不能为0"))
	}
	return r.db.Model(&model.SaleOrderProductAttribute{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
}

// UpdateOrCreateSaleOrderProductAttributeRecord 修改或新建销售订单商品属性记录
// 如果没有uuid则新建，否则修改
func (r *saleOrderProductAttributeRepo) UpdateOrCreateSaleOrderProductAttributeRecord(obj model.SaleOrderProductAttribute) error {
	// 如果主键id为0或uuid为0则create，否则update
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return r.CreateSaleOrderProductAttributeRecord(obj)
	}
	return r.UpdateSaleOrderProductAttributeRecord(obj)
}
