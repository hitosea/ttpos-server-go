package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderRepo 定义订单仓库接口
type IOrderRepo interface {
	CreateSaleBill(model model.SaleBill) (model.SaleBill, error)    // 创建销售单
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)           // 获取销售单
	CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error) // 创建订单
}

// orderRepo 订单仓库
type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 创建新的订单仓库
func NewOrderRepo(db *gorm.DB) IOrderRepo {
	return NewOrderRepoImpl(db)
}

// NewOrderRepoImpl 创建新的订单仓库实现
func NewOrderRepoImpl(db *gorm.DB) IOrderRepo {
	return &orderRepo{db: db}
}

// CreateSaleBill 创建销售单
func (r *orderRepo) CreateSaleBill(model model.SaleBill) (model.SaleBill, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, err
	}

	return model, nil
}

// GetSaleBill 获取销售单
func (r *orderRepo) GetSaleBill(opts ...DBOption) (model.SaleBill, error) {
	var model model.SaleBill
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&model)
	if result.Error != nil {
		return model, result.Error
	}

	return model, nil
}

// CreateSaleOrder 创建销售订单
func (r *orderRepo) CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, err
	}
	return model, nil
}
