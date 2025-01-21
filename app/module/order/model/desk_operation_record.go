package model

import (
	"time"

	"gorm.io/gorm"
)

// DeskOperationRecord 桌台操作记录表
type DeskOperationRecord struct {
	ID            int    `gorm:"primaryKey;column:id;comment:桌台操作记录唯一标识符"`
	Client        string `gorm:"column:client;comment:客户端信息"`
	Message       string `gorm:"column:message;comment:消息内容"`
	TableID       int    `gorm:"column:table_id;comment:桌子ID"`
	OperatorName  string `gorm:"column:operator_name;comment:操作员名称"`
	OperatorEmail string `gorm:"column:operator_email;comment:操作员邮箱"`
	CreateTime    int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime    int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime    int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *DeskOperationRecord) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *DeskOperationRecord) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
