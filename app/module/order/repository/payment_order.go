package repository

import (
	"ttpos-server-go/app/module/order/model"

	"gorm.io/gorm"
)

type PaymentOrderModel struct {
	db *gorm.DB
}

func NewPaymentOrderModel(db *gorm.DB) *PaymentOrderModel {
	return &PaymentOrderModel{db: db}
}

// Create 创建记录
func (m *PaymentOrderModel) Create(data *model.PaymentOrder) error {
	return m.db.Create(data).Error
}
