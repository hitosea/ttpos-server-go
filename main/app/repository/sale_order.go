package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleOrderRepo interface {
	GetSaleOrder(opts ...DBOption) (model.SaleOrder, error)
	GetSaleOrderByUuid(uuid uint64) (*model.SaleOrder, error)
	UpdateSaleOrder(model *model.SaleOrder) error
}

type saleOrderRepo struct {
	db *gorm.DB
}

func NewSaleOrderRepo(db *gorm.DB) ISaleOrderRepo {
	return &saleOrderRepo{db: db}
}

func (r *saleOrderRepo) GetSaleOrder(opts ...DBOption) (model.SaleOrder, error) {
	var saleOrder model.SaleOrder
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&saleOrder)
	if result.Error != nil {
		return saleOrder, result.Error
	}

	return saleOrder, nil
}

func (r *saleOrderRepo) GetSaleOrderByUuid(uuid uint64) (*model.SaleOrder, error) {
	order, err := r.GetSaleOrder(CommonRepo.WhereByUuid(uuid))
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *saleOrderRepo) UpdateSaleOrder(model *model.SaleOrder) error {
	return r.db.Model(model).Save(model).Error
}
