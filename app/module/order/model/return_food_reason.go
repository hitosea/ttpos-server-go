package model

import (
	"time"

	"gorm.io/gorm"
)

// ReturnFoodReason 退菜原因表
type ReturnFoodReason struct {
	ID                  int    `gorm:"primaryKey;column:id;comment:退菜原因唯一标识符"`
	Name                string `gorm:"column:name;comment:名称"`
	MultiLanguageNameID int    `gorm:"column:multi_language_name_id;comment:多语言名称ID"`
	CreateTime          int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime          int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime          int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *ReturnFoodReason) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *ReturnFoodReason) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
