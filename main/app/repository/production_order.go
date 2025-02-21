package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductionOrderRepo interface {
	GetProductionOrder(opts ...DBOption) (*model.ProductionOrder, error)
	CreateProductionOrder(order *model.ProductionOrder) error
}

type productionOrderRepoImpl struct {
	db *gorm.DB
}

func NewProductionOrderRepo(db *gorm.DB) IProductionOrderRepo {
	return &productionOrderRepoImpl{db: db}
}

func (r *productionOrderRepoImpl) GetProductionOrder(opts ...DBOption) (*model.ProductionOrder, error) {
	var productionOrder model.ProductionOrder
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productionOrder)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productionOrder, nil
}

// 创建ProductionOrder记录及它管理的表记录
func (r *productionOrderRepoImpl) CreateProductionOrder(order *model.ProductionOrder) error {
	if err := r.db.Model(order).Error; err != nil {
		return err
	}
	return nil
}
