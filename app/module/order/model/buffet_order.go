package model

import (
	"time"

	"gorm.io/gorm"
)

// BuffetOrder 自助餐订单表
type BuffetOrder struct {
	ID              int   `gorm:"primaryKey;column:id;comment:自助餐订单唯一标识符"`
	SaleBillID      int   `gorm:"column:sale_bill_id;comment:销售账单ID"`
	BuffetPackageID int   `gorm:"column:buffet_package_id;comment:自助餐套餐ID"`
	CreateTime      int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime      int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime      int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *BuffetOrder) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *BuffetOrder) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
