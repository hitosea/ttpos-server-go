package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type IReturnOrderRepo interface {
	CreateReturnOrder(order model.ReturnOrder) (model.ReturnOrder, error) // 创建退货单
	CreateReturnOrderAmount(amounts []model.ReturnOrderAmount) error      // 创建退货金额
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

func (r *returnOrderRepo) CreateReturnOrder(order model.ReturnOrder) (model.ReturnOrder, error) {
	err := r.db.Model(&model.ReturnOrder{}).Create(&order).Error
	return order, errors.WithMessage(err)
}

func (r *returnOrderRepo) CreateReturnOrderAmount(amounts []model.ReturnOrderAmount) error {
	return r.db.Model(&model.ReturnOrderAmount{}).Create(&amounts).Error
}
