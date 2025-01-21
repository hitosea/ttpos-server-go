package model

import (
	"time"

	"gorm.io/gorm"
)

// PurchaseOrder 采购订单表
type PurchaseOrder struct {
	ID            int     `gorm:"primaryKey;column:id;comment:采购订单唯一标识符"`
	OrderNumber   string  `gorm:"column:order_number;comment:订单编号"`
	SupplierID    int     `gorm:"column:supplier_id;comment:供应商ID"`
	TotalAmount   float64 `gorm:"column:total_amount;comment:总金额"`
	Status        string  `gorm:"column:status;comment:状态,pending待处理、completed已完成、cancelled已取消"`
	PurchaserID   int     `gorm:"column:purchaser_id;comment:采购员ID"`
	PurchaserName string  `gorm:"column:purchaser_name;comment:采购员名称"`
	Remark        string  `gorm:"column:remark;comment:备注"`
	CreateTime    int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime    int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime    int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *PurchaseOrder) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
