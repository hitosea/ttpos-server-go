package model

import (
	"ttpos-server-go/config"
)

// CustomerCall 客户呼叫记录表 ttpos_customer_call
type CustomerCall struct {
	BaseModel
	DeskUuid uint64 `gorm:"column:desk_uuid;type:bigint(20) unsigned;default:0;comment:桌台ID;NOT NULL" json:"desk_uuid"`
	DeskNo   string `gorm:"column:desk_no;type:varchar(255);comment:桌台编号,不随后台改变;NOT NULL" json:"desk_no"`
	CallType uint8  `gorm:"column:call_type;type:tinyint(4) unsigned;default:1;comment:呼叫类型(1服务员,2结账)" json:"call_type"`
	Status   int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态,0-unhandled未处理 1-handled已处理;NOT NULL" json:"status"`
	IsSend   int    `gorm:"column:is_send;type:tinyint(1);default:0;comment:消息发送状态 0-否 1-是;NOT NULL" json:"is_send"`
}

// TableName 设置表名
func (CustomerCall) TableName() string {
	return config.Database.TablePrefix + "customer_call"
}
