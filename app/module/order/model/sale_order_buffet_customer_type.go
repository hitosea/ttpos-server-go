package model

import (
	"time"

	"gorm.io/gorm"
)

// SaleOrderBuffetCustomerType 销售订单顾客类型表
type SaleOrderBuffetCustomerType struct {
	ID                   int   `gorm:"primaryKey;column:id;comment:销售订单顾客类型唯一标识符"`
	SaleOrderID          int   `gorm:"column:sale_order_id;comment:销售订单ID"`
	BuffetPackageID      int   `gorm:"column:buffet_package_id;comment:自助餐套餐ID"`
	BuffetCustomerTypeID int   `gorm:"column:buffet_customer_type_id;comment:自助餐客户类型ID"`
	Num                  int   `gorm:"column:num;comment:人数"`
	CreateTime           int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime           int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime           int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *SaleOrderBuffetCustomerType) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *SaleOrderBuffetCustomerType) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
