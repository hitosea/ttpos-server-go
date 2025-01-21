package model

import (
	"time"

	"gorm.io/gorm"
)

// BuffetProduct 自助餐产品表
type BuffetProduct struct {
	ID                      int   `gorm:"primaryKey;column:id;comment:显示记录唯一标识符"`
	ProductPackageID        int   `gorm:"column:product_package_id;comment:产品包ID"`
	DisplayCashier          bool  `gorm:"column:display_cashier;comment:是否在收银台显示"`
	DisplayTable            bool  `gorm:"column:display_table;comment:是否在桌面显示"`
	DisplayKitchen          bool  `gorm:"column:display_kitchen;comment:是否在厨房显示"`
	DisplayAssistant        bool  `gorm:"column:display_assistant;comment:是否在助手显示"`
	LimitedPurchaseQuantity int   `gorm:"column:limited_purchase_quantity;comment:限购数量"`
	CreateTime              int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime              int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime              int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *BuffetProduct) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *BuffetProduct) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
