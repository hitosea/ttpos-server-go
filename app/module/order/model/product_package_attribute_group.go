package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductPackageAttributeGroup 产品包属性组表
type ProductPackageAttributeGroup struct {
	ID               int   `gorm:"primaryKey;column:id;comment:记录唯一标识符"`
	IsMust           bool  `gorm:"column:is_must;comment:是否必选"`
	MaxSelection     int   `gorm:"column:max_selection;comment:最大选择数量"`
	ProductPackageID int   `gorm:"column:product_package_id;comment:产品包ID"`
	CreateTime       int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime       int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime       int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *ProductPackageAttributeGroup) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *ProductPackageAttributeGroup) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
