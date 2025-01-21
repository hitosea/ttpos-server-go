package model

import (
	"time"

	"gorm.io/gorm"
)

// SaleOrderDiscountStrategy 销售订单折扣策略表
type SaleOrderDiscountStrategy struct {
	ID                 int     `gorm:"primaryKey;column:id;comment:销售订单折扣策略唯一标识符"`
	SaleOrderID        int     `gorm:"column:sale_order_id;comment:销售订单ID"`
	DiscountStrategyID int     `gorm:"column:discount_strategy_id;comment:折扣策略ID"`
	DiscountAmount     float64 `gorm:"column:discount_amount;comment:折扣金额"`
	CreateTime         int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime         int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime         int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *SaleOrderDiscountStrategy) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *SaleOrderDiscountStrategy) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
