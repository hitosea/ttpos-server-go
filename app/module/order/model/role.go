package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色表
type Role struct {
	ID         int    `gorm:"primaryKey;column:id;comment:角色唯一标识符"`
	Name       string `gorm:"column:name;comment:角色名称"`
	CreateTime int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *Role) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Role) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
