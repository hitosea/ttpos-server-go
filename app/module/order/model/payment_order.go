package model

import (
	"time"

	"gorm.io/gorm"
)

// PaymentOrder 支付记录表
type PaymentOrder struct {
	ID                int     `gorm:"primaryKey;column:id;comment:支付记录唯一标识符"`
	PaymentTypeName   string  `gorm:"column:payment_type_name;comment:支付类型名称"`
	PaymentTypeID     int     `gorm:"column:payment_type_id;comment:支付类型ID"`
	PaymentFeePercent float64 `gorm:"column:payment_fee_percent;comment:支付手续费百分比"`
	SaleOrderID       int     `gorm:"column:sale_order_id;comment:销售订单ID"`
	CurrencyUnit      string  `gorm:"column:currency_unit;comment:货币单位"`
	PaymentAmount     float64 `gorm:"column:payment_amount;comment:支付金额"`
	Amount            float64 `gorm:"column:amount;comment:金额"`
	TransactionNumber string  `gorm:"column:transaction_number;comment:交易号"`
	Status            string  `gorm:"column:status;comment:支付状态"`
	CreateTime        int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime        int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime        int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *PaymentOrder) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *PaymentOrder) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
