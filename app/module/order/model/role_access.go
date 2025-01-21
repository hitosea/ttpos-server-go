package model

import (
	"time"

	"gorm.io/gorm"
)

// RoleAccess 角色权限表
type RoleAccess struct {
	ID         int   `gorm:"primaryKey;column:id;comment:角色权限唯一标识符"`
	RoleID     int   `gorm:"column:role_id;comment:角色ID"`
	AccessID   int   `gorm:"column:access_id;comment:权限ID"`
	CreateTime int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *RoleAccess) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *RoleAccess) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
