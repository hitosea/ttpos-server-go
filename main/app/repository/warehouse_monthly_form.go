package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IWarehouseMonthlyFormRepo interface {
	CreateWarehouseMonthlyForm(warehouseMonthlyForm model.WarehouseMonthlyForm) error
}

type warehouseMonthlyFormRepoImpl struct {
	db *gorm.DB
}

func NewWarehouseMonthlyFormRepo(db *gorm.DB) IWarehouseMonthlyFormRepo {
	return &warehouseMonthlyFormRepoImpl{db: db}
}

func (r *warehouseMonthlyFormRepoImpl) CreateWarehouseMonthlyForm(warehouseMonthlyForm model.WarehouseMonthlyForm) error {
	return r.db.Create(&warehouseMonthlyForm).Error
}
