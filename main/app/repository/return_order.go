package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type IReturnOrderRepo interface {
	CreateReturnOrder(order model.ReturnOrder) (model.ReturnOrder, error) // 创建退货单
	CreateReturnOrderRecord(order model.ReturnOrder) (uint64, error)      // 创建退货单
	CreateReturnOrderAmount(amounts []model.ReturnOrderAmount) error      // 创建退货金额
	CreateReturnOrderProduct(products []*model.ReturnOrderProduct) error  // 创建退货商品
}

func NewReturnOrderRepo(db *gorm.DB) IReturnOrderRepo {
	return NewReturnOrderRepoImpl(db)
}

type returnOrderRepo struct {
	db *gorm.DB
}

func NewReturnOrderRepoImpl(db *gorm.DB) IReturnOrderRepo {
	return &returnOrderRepo{db: db}
}

func (r *returnOrderRepo) CreateReturnOrderRecord(order model.ReturnOrder) (uint64, error) {
	order.SetNil()
	err := r.db.Model(&model.ReturnOrder{}).Create(&order).Error
	return order.Uuid, errors.WithMessage(err)
}

func (r *returnOrderRepo) CreateReturnOrder(order model.ReturnOrder) (model.ReturnOrder, error) {
	err := r.db.Model(&model.ReturnOrder{}).Create(&order).Error
	return order, errors.WithMessage(err)
}

func (r *returnOrderRepo) CreateReturnOrderAmount(amounts []model.ReturnOrderAmount) error {
	for _, amount := range amounts {
		amount.SetNil()
	}
	return r.db.Model(&model.ReturnOrderAmount{}).Create(&amounts).Error
}

func (r *returnOrderRepo) CreateReturnOrderProduct(products []*model.ReturnOrderProduct) error {
	if len(products) == 0 {
		return nil // 避免gorm报错empty slice found
	}
	for _, product := range products {
		product.SetNil()
	}
	return r.db.Model(&model.ReturnOrderProduct{}).Create(products).Error
}
