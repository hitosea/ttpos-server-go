package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IWarehouseMonthlyFormRepo 月度报表仓库接口
type IWarehouseMonthlyFormRepo interface {
	CreateWarehouseMonthlyForm(warehouseMonthlyForm model.WarehouseMonthlyForm) error                               // 创建月度报表
	CreateWarehouseMonthlyProductBomForm(warehouseMonthlyProductBomForm model.WarehouseMonthlyProductBomForm) error // 创建月度商品bom报表
	CreateWarehouseMonthlyMaterialForm(warehouseMonthlyMaterialForm model.WarehouseMonthlyMaterialForm) error       // 创建月度材料报表
}

// warehouseMonthlyFormRepoImpl 月度报表仓库实现
type warehouseMonthlyFormRepoImpl struct {
	db *gorm.DB
}

// NewWarehouseMonthlyFormRepo 创建月度报表仓库
func NewWarehouseMonthlyFormRepo(db *gorm.DB) IWarehouseMonthlyFormRepo {
	return &warehouseMonthlyFormRepoImpl{db: db}
}

// CreateWarehouseMonthlyForm 创建月度报表
func (r *warehouseMonthlyFormRepoImpl) CreateWarehouseMonthlyForm(warehouseMonthlyForm model.WarehouseMonthlyForm) error {
	return r.db.Create(&warehouseMonthlyForm).Error
}

// CreateWarehouseMonthlyProductBomForm 创建月度商品bom报表
func (r *warehouseMonthlyFormRepoImpl) CreateWarehouseMonthlyProductBomForm(warehouseMonthlyProductBomForm model.WarehouseMonthlyProductBomForm) error {
	return r.db.Create(&warehouseMonthlyProductBomForm).Error
}

// CreateWarehouseMonthlyMaterialForm 创建月度材料报表
func (r *warehouseMonthlyFormRepoImpl) CreateWarehouseMonthlyMaterialForm(warehouseMonthlyMaterialForm model.WarehouseMonthlyMaterialForm) error {
	return r.db.Create(&warehouseMonthlyMaterialForm).Error
}
