package model

import (
	"time"

	"gorm.io/gorm"
)

// Supplier 供应商信息表
type Supplier struct {
	ID          int    `gorm:"primaryKey;column:id;comment:供应商唯一标识符"`
	Name        string `gorm:"column:name;comment:供应商名称"`
	Contact     string `gorm:"column:contact;comment:联系人"`
	Phone       string `gorm:"column:phone;comment:联系电话"`
	Address     string `gorm:"column:address;comment:地址"`
	Email       string `gorm:"column:email;comment:邮箱"`
	Status      string `gorm:"column:status;comment:状态,enable启用、disable禁用"`
	Description string `gorm:"column:description;comment:描述"`
	CreateTime  int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime  int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime  int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *Supplier) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Supplier) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
