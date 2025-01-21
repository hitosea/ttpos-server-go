package repository

import (
	"jjjshop-server-go/app/module/order/model"

	"gorm.io/gorm"
)

type SaleOrderProductModel struct {
	db *gorm.DB
}

func NewSaleOrderProductModel(db *gorm.DB) *SaleOrderProductModel {
	return &SaleOrderProductModel{db: db}
}

// Create 创建记录
func (m *SaleOrderProductModel) Create(data *model.SaleOrderProduct) error {
	return m.db.Create(data).Error
}
