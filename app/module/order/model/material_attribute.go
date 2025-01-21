package model

import (
	"time"

	"gorm.io/gorm"
)

// MaterialAttribute 原料扩展属性表
type MaterialAttribute struct {
	ID                           int   `gorm:"primaryKey;column:id;comment:记录唯一标识符"`
	MaterialID                   int   `gorm:"column:material_id;comment:原料ID"`
	HistoricalPurchaseQuantity   int   `gorm:"column:historical_purchase_quantity;comment:历史采购数量"`
	HistoricalLossReportQuantity int   `gorm:"column:historical_loss_report_quantity;comment:历史报损数量"`
	HistoricalSaleQuantity       int   `gorm:"column:historical_sale_quantity;comment:历史销售数量"`
	CreateTime                   int64 `gorm:"column:create_time;comment:创建时间"`
	UpdateTime                   int64 `gorm:"column:update_time;comment:更新时间"`
	DeleteTime                   int64 `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *MaterialAttribute) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *MaterialAttribute) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
