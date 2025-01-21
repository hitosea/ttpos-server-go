package model

import (
	"time"

	"gorm.io/gorm"
)

// DeskType 餐桌类型表
type DeskType struct {
	ID         int    `gorm:"primaryKey;column:id;comment:餐桌类型唯一标识符"`
	Name       string `gorm:"column:name;comment:餐桌类型名称"`
	OrderBy    int    `gorm:"column:order_by;comment:排序序号"`
	RangeMin   int    `gorm:"column:range_min;comment:最少人数"`
	RangeMax   int    `gorm:"column:range_max;comment:最多人数"`
	CreateTime int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *DeskType) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *DeskType) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
