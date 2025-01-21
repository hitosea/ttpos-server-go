package model

import (
	"time"

	"gorm.io/gorm"
)

// MemberRechargeOrder 会员充值订单表
type MemberRechargeOrder struct {
	ID              int     `gorm:"primaryKey;column:id;comment:会员充值订单唯一标识符"`
	OrderNo         string  `gorm:"column:order_no;comment:订单编号"`
	MemberID        int     `gorm:"column:member_id;comment:会员ID"`
	Amount          float64 `gorm:"column:amount;comment:充值金额"`
	RewardAmount    float64 `gorm:"column:reward_amount;comment:奖励金额"`
	PaymentAmount   float64 `gorm:"column:payment_amount;comment:支付金额"`
	Status          string  `gorm:"column:status;comment:状态"`
	PaymentTypeID   int     `gorm:"column:payment_type_id;comment:支付类型ID"`
	PaymentTypeName string  `gorm:"column:payment_type_name;comment:支付类型名称"`
	OperatorID      int     `gorm:"column:operator_id;comment:操作员ID"`
	OperatorName    string  `gorm:"column:operator_name;comment:操作员名称"`
	CreateTime      int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime      int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime      int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *MemberRechargeOrder) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *MemberRechargeOrder) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
