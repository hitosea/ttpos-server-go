package model

import (
	"time"

	"gorm.io/gorm"
)

// SaleBill 销售账单表
type SaleBill struct {
	ID            int     `gorm:"primaryKey;column:id;comment:订单唯一标识符"`
	SN            string  `gorm:"column:sn;comment:订单编号"`
	BillType      string  `gorm:"column:bill_type;comment:账单类型"`
	DiningMethod  string  `gorm:"column:dining_method;comment:用餐方式"`
	IsBuffet      bool    `gorm:"column:is_buffet;comment:是否自助餐"`
	Status        string  `gorm:"column:status;comment:订单状态"`
	Reason        string  `gorm:"column:reason;comment:原因"`
	OrderAmount   float64 `gorm:"column:order_amount;comment:订单总金额"`
	ProductAmount float64 `gorm:"column:product_amount;comment:商品金额"`
	PaymentAmount float64 `gorm:"column:payment_amount;comment:支付金额"`
	ConsumerID    int     `gorm:"column:consumer_id;comment:消费者ID"`
	CashierID     int     `gorm:"column:cashier_id;comment:收银员ID"`
	BuffetOrderID int     `gorm:"column:buffet_order_id;comment:自助餐订单ID"`
	TableID       int     `gorm:"column:table_id;comment:餐桌ID"`
	HideBillTime  int64   `gorm:"column:hide_bill_time;comment:隐藏账单时间"`
	FinishTime    int64   `gorm:"column:finish_time;comment:完成时间"`
	CreateTime    int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime    int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime    int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *SaleBill) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *SaleBill) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
