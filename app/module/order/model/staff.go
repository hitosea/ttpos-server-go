package model

import (
	"time"

	"gorm.io/gorm"
)

// Staff 员工表
type Staff struct {
	ID         int    `gorm:"primaryKey;column:id;comment:员工唯一标识符"`
	Name       string `gorm:"column:name;comment:员工姓名"`
	Email      string `gorm:"column:email;comment:电子邮件地址"`
	Phone      string `gorm:"column:phone;comment:电话号码"`
	Password   string `gorm:"column:password;comment:密码"`
	Status     string `gorm:"column:status;comment:状态,active激活、inactive禁用"`
	CreateTime int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *Staff) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Staff) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
