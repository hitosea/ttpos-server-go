package repository

import (
	sale "jjjshop-server-go/app/module/order/model"

	"gorm.io/gorm"
)

// 基础的CRUD方法
type SaleBillModel struct {
	db *gorm.DB
}

func NewSaleBillModel(db *gorm.DB) *SaleBillModel {
	return &SaleBillModel{db: db}
}

// Create 创建记录
func (m *SaleBillModel) Create(data *sale.SaleBill) error {
	return m.db.Create(data).Error
}
