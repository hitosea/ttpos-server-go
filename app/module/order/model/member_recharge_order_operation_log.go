package model

import (
	"time"

	"gorm.io/gorm"
)

// MemberRechargeOrderOperationLog 会员充值订单操作日志表
type MemberRechargeOrderOperationLog struct {
	ID                    int    `gorm:"primaryKey;column:id;comment:会员充值订单操作日志唯一标识符"`
	MemberRechargeOrderID int    `gorm:"column:member_recharge_order_id;comment:会员充值订单ID"`
	OperationType         string `gorm:"column:operation_type;comment:操作类型"`
	OperatorID            int    `gorm:"column:operator_id;comment:操作员ID"`
	OperatorName          string `gorm:"column:operator_name;comment:操作员名称"`
	Remark                string `gorm:"column:remark;comment:备注"`
	CreateTime            int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime            int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime            int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *MemberRechargeOrderOperationLog) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *MemberRechargeOrderOperationLog) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
