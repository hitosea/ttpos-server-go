package model

import (
	"time"

	"gorm.io/gorm"
)

// WarehouseOutFormItem 出库单明细表
type WarehouseOutFormItem struct {
	ID                 int     `gorm:"primaryKey;column:id;comment:出库单明细唯一标识符"`
	WarehouseOutFormID int     `gorm:"column:warehouse_out_form_id;comment:出库单ID"`
	MaterialID         int     `gorm:"column:material_id;comment:原料ID"`
	Quantity           int     `gorm:"column:quantity;comment:数量"`
	UnitPrice          float64 `gorm:"column:unit_price;comment:单价"`
	Amount             float64 `gorm:"column:amount;comment:金额"`
	CreateTime         int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime         int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime         int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *WarehouseOutFormItem) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *WarehouseOutFormItem) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
