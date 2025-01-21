package model

import (
	"time"

	"gorm.io/gorm"
)

// MaterialCategory 原料类别表
type MaterialCategory struct {
	ID                  int    `gorm:"primaryKey;column:id;comment:记录唯一标识符"`
	Name                string `gorm:"column:name;comment:名称"`
	MultiLanguageNameID int    `gorm:"column:multi_language_name_id;comment:多语言名称ID"`
	Status              string `gorm:"column:status;comment:状态, open开启、close关闭"`
	Level               int    `gorm:"column:level;comment:层级"`
	ParentID            int    `gorm:"column:parent_id;comment:父级ID"`
	CategoryKey         string `gorm:"column:category_key;comment:关键字"`
	OrderBy             int    `gorm:"column:order_by;comment:排序"`
	RefCount            int    `gorm:"column:ref_count;comment:关联数量"`
	CreateTime          int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime          int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime          int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *MaterialCategory) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *MaterialCategory) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
