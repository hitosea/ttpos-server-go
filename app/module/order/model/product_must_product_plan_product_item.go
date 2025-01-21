package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductMustProductPlanProductItem 产品必选产品计划产品明细表
type ProductMustProductPlanProductItem struct {
	ID                       int   `gorm:"primaryKey;column:id;comment:产品必选产品计划产品明细唯一标识符"`
	ProductMustProductPlanID int   `gorm:"column:product_must_product_plan_id;comment:产品必选产品计划ID"`
	ProductPackageID         int   `gorm:"column:product_package_id;comment:产品包ID"`
	CreateTime               int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime               int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime               int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *ProductMustProductPlanProductItem) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *ProductMustProductPlanProductItem) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
