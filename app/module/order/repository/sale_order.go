package repository

import (
	"jjjshop-server-go/app/module/order/model"

	"gorm.io/gorm"
)

type SaleOrderModel struct {
	db *gorm.DB
}

func NewSaleOrderModel(db *gorm.DB) *SaleOrderModel {
	return &SaleOrderModel{db: db}
}

// Create 创建记录
func (m *SaleOrderModel) Create(data *model.SaleOrder) error {
	return m.db.Create(data).Error
}
