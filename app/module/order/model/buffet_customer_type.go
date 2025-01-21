package model

import (
	"time"

	"gorm.io/gorm"
)

// BuffetCustomerType 自助餐客户类型表
type BuffetCustomerType struct {
	ID                  int    `gorm:"primaryKey;column:id;comment:自助餐客户类型唯一标识符"`
	Name                string `gorm:"column:name;comment:自助餐客户类型名称"`
	MultiLanguageNameID int    `gorm:"column:multi_language_name_id;comment:多语言名称ID"`
	CreateTime          int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime          int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime          int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *BuffetCustomerType) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *BuffetCustomerType) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
