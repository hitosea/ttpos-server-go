package model

import (
	"time"

	"gorm.io/gorm"
)

// StockInRecord 入库记录表
type StockInRecord struct {
	ID              int     `gorm:"primaryKey;column:id;comment:入库记录唯一标识符"`
	PurchaseOrderID int     `gorm:"column:purchase_order_id;comment:采购订单ID"`
	MaterialID      int     `gorm:"column:material_id;comment:原料ID"`
	Quantity        int     `gorm:"column:quantity;comment:数量"`
	UnitPrice       float64 `gorm:"column:unit_price;comment:单价"`
	Amount          float64 `gorm:"column:amount;comment:金额"`
	OperatorID      int     `gorm:"column:operator_id;comment:操作员ID"`
	OperatorName    string  `gorm:"column:operator_name;comment:操作员名称"`
	Remark          string  `gorm:"column:remark;comment:备注"`
	CreateTime      int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime      int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime      int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *StockInRecord) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *StockInRecord) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
