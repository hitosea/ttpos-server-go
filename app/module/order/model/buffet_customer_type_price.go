package model

import (
	"time"

	"gorm.io/gorm"
)

// BuffetCustomerTypePrice 自助餐顾客类型价格表
type BuffetCustomerTypePrice struct {
	ID              int     `gorm:"primaryKey;column:id;comment:自助餐顾客类型价格唯一标识符"`
	BuffetPackageID int     `gorm:"column:buffet_package_id;comment:自助餐套餐ID"`
	CustomerTypeID  int     `gorm:"column:customer_type_id;comment:客户类型ID"`
	Price           float64 `gorm:"column:price;comment:价格"`
	CreateTime      int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime      int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime      int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *BuffetCustomerTypePrice) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *BuffetCustomerTypePrice) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
