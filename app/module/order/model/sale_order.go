package model

import (
	"time"

	"gorm.io/gorm"
)

// SaleOrder 销售订单表
type SaleOrder struct {
	ID                    int     `gorm:"primaryKey;column:id;comment:销售订单唯一标识符"`
	OrderNo               string  `gorm:"column:order_no;comment:订单编号"`
	IsBuffet              bool    `gorm:"column:is_buffet;comment:是否自助餐"`
	Status                string  `gorm:"column:status;comment:订单状态"`
	ProductAmount         float64 `gorm:"column:product_amount;comment:商品金额"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;comment:商品原始金额"`
	ServiceFee            float64 `gorm:"column:service_fee;comment:服务费"`
	TaxFee                float64 `gorm:"column:tax_fee;comment:税费"`
	DiscountFee           float64 `gorm:"column:discount_fee;comment:折扣费用"`
	MemberDiscountFee     float64 `gorm:"column:member_discount_fee;comment:会员折扣费用"`
	Amount                float64 `gorm:"column:amount;comment:总金额"`
	IsGift                bool    `gorm:"column:is_gift;comment:是否免单"`
	ConsumerID            int     `gorm:"column:consumer_id;comment:消费者ID"`
	CashierID             int     `gorm:"column:cashier_id;comment:收银员ID"`
	SaleBillID            int     `gorm:"column:sale_bill_id;comment:账单ID"`
	FinishTime            int64   `gorm:"column:finish_time;comment:完成时间"`
	CreateTime            int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime            int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime            int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *SaleOrder) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *SaleOrder) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
