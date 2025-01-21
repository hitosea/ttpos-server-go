package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductSaleInventory 销售库存表
type ProductSaleInventory struct {
	ID               int    `gorm:"primaryKey;column:id;comment:销售库存唯一标识符"`
	ProductPackageID int    `gorm:"column:product_package_id;comment:产品包ID"`
	Num              int    `gorm:"column:num;comment:数量"`
	Status           string `gorm:"column:status;comment:状态,unclear未沽清、clear已沽清"`
	InventoryCount   int    `gorm:"column:inventory_count;comment:库存数量,实际库存量"`
	CreateTime       int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime       int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime       int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *ProductSaleInventory) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *ProductSaleInventory) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
