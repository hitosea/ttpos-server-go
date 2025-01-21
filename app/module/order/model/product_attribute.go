package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductAttribute 产品属性记录表
type ProductAttribute struct {
	ID                 int   `gorm:"primaryKey;column:id;comment:产品属性记录唯一标识符"`
	SaleOrderProductID int   `gorm:"column:sale_order_product_id;comment:销售订单产品ID"`
	AttributeID        int   `gorm:"column:attribute_id;comment:商品属性ID"`
	CreateTime         int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime         int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime         int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *ProductAttribute) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *ProductAttribute) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
