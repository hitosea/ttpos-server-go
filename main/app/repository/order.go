package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderRepo 定义订单仓库接口
type IOrderRepo interface {
	CreateSaleBill(model model.SaleBill) (uint64, error)   // 创建销售单
	CreateSaleOrder(model model.SaleOrder) (uint64, error) // 创建订单
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
func (r *orderRepo) CreateSaleBill(model model.SaleBill) (uint64, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return 0, err
	}
	return model.Uuid, nil
}

// CreateSaleOrder 创建销售订单
func (r *orderRepo) CreateSaleOrder(model model.SaleOrder) (uint64, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return 0, err
	}
	return model.Uuid, nil
}
