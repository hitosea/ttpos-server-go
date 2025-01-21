package model

import (
	"time"

	"gorm.io/gorm"
)

// Access 权限表
type Access struct {
	ID           int    `gorm:"primaryKey;column:id;comment:权限唯一标识符"`
	Name         string `gorm:"column:name;comment:权限名称"`
	Path         string `gorm:"column:path;comment:路径"`
	APIPath      string `gorm:"column:api_path;comment:API路径"`
	ParentID     int    `gorm:"column:parent_id;comment:父权限ID"`
	OrderBy      int    `gorm:"column:order_by;comment:排序"`
	Icon         string `gorm:"column:icon;comment:图标"`
	RedirectName string `gorm:"column:redirect_name;comment:重定向名称"`
	IsRoute      bool   `gorm:"column:is_route;comment:是否是路由"`
	IsMenu       bool   `gorm:"column:is_menu;comment:是否是菜单"`
	IsShow       bool   `gorm:"column:is_show;comment:是否显示"`
	CreateTime   int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime   int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime   int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *Access) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Access) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
