package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPaymentOrderRepo 定义仓库接口
type IPaymentOrderRepo interface {
	WhereSaleOrderUuid(uuid uint64) DBOption
	WhereSaleOrderUuids(uuids []uint64) DBOption
	GetPaymentOrder(opts ...DBOption) (model.PaymentOrder, error) // 获取详细信息
	GetPaymentOrderList(opts ...DBOption) ([]model.PaymentOrder, error)
}

// paymentOrderRepo 仓库
type paymentOrderRepo struct {
	db *gorm.DB
}

// NewPaymentOrderRepo 创建新仓库
func NewPaymentOrderRepo(db *gorm.DB) IPaymentOrderRepo {
	return NewPaymentOrderRepoImpl(db)
}

// NewPaymentOrderRepoImpl 创建新仓库实现
func NewPaymentOrderRepoImpl(db *gorm.DB) IPaymentOrderRepo {
	return &paymentOrderRepo{db: db}
}

// GetPaymentOrder 获取详细信息
func (r *paymentOrderRepo) GetPaymentOrder(opts ...DBOption) (model.PaymentOrder, error) {
	var paymentOrder model.PaymentOrder
	for _, w := range opts {
		r.db = w(r.db)
	}
	r.db = CommonRepo.WhereBySoftDelete()(r.db)
	err := r.db.Order("id asc").First(&paymentOrder).Error
	return paymentOrder, err
}

// GetPaymentOrderList 获取列表
func (r *paymentOrderRepo) GetPaymentOrderList(opts ...DBOption) ([]model.PaymentOrder, error) {
	var paymentOrders []model.PaymentOrder
	for _, w := range opts {
		r.db = w(r.db)
	}
	r.db = CommonRepo.WhereBySoftDelete()(r.db)
	err := r.db.Order("id asc").Find(&paymentOrders).Error
	return paymentOrders, err
}

// WhereSaleOrderUuid 根据sale_order_uuid查询
func (r *paymentOrderRepo) WhereSaleOrderUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_order_uuid = ?", uuid)
	}
}

// WhereSaleOrderUuids 根据sale_order_uuids查询
func (r *paymentOrderRepo) WhereSaleOrderUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_order_uuid in (?)", uuids)
	}
}
