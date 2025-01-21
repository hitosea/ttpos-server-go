package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductPackageAttribute 产品包属性表
type ProductPackageAttribute struct {
	ID                int   `gorm:"primaryKey;column:id;comment:记录唯一标识符"`
	AttributeGroupID  int   `gorm:"column:attribute_group_id;comment:产品包属性组ID"`
	AttributeID       int   `gorm:"column:attribute_id;comment:产品属性ID"`
	IsDefaultSelected bool  `gorm:"column:is_default_selected;comment:是否默认选中"`
	CreateTime        int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime        int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime        int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *ProductPackageAttribute) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *ProductPackageAttribute) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
