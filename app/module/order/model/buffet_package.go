package model

import (
	"time"

	"gorm.io/gorm"
)

// BuffetPackage 自助餐套餐信息表
type BuffetPackage struct {
	ID                  int    `gorm:"primaryKey;column:id;comment:自助餐套餐唯一标识符"`
	Name                string `gorm:"column:name;comment:自助餐套餐名称"`
	MultiLanguageNameID int    `gorm:"column:multi_language_name_id;comment:多语言名称ID"`
	OrderBy             int    `gorm:"column:order_by;comment:排序顺序"`
	TaxID               int    `gorm:"column:tax_id;comment:税收ID"`
	IsLimitTime         bool   `gorm:"column:is_limit_time;comment:是否限时"`
	LimitTime           int    `gorm:"column:limit_time;comment:限时时间（分钟）"`
	CanCombined         bool   `gorm:"column:can_combined;comment:是否可合并"`
	NonOrderingTime     int    `gorm:"column:non_ordering_time;comment:不可下单时间（分钟）"`
	ReminderOrderTime   int    `gorm:"column:reminder_order_time;comment:提醒下单时间（分钟）"`
	CreateTime          int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime          int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime          int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *BuffetPackage) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *BuffetPackage) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
