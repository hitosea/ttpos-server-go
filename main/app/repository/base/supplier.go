package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// 供应商
type SupplierRepoInterface interface {
	GetSupplierList() ([]model.Supplier, error)            // 获取供应商列表
	UpdateSupplier(id uint, supplier model.Supplier) error // 更新供应商
	CreateSupplier(supplier model.Supplier) (uint, error)  // 创建供应商
	DeleteSupplier(id uint) error                          // 软删除供应商
}

func NewSupplierRepo(db *gorm.DB) SupplierRepoInterface {
	return NewSupplierRepoImpl(db)
}

// 创建新的打印机标签仓库实现
func NewSupplierRepoImpl(db *gorm.DB) *SupplierRepoImpl {
	return &SupplierRepoImpl{db: db}
}

type SupplierRepoImpl struct {
	db *gorm.DB
}

// 获取供应商列表，排除逻辑删除的供应商
func (r *SupplierRepoImpl) GetSupplierList() ([]model.Supplier, error) {
	var suppliers []model.Supplier
	err := r.db.Model(&model.Supplier{}).Where("delete_time = ?", 0).Find(&suppliers).Error
	return suppliers, err
}

// 更新供应商
func (r *SupplierRepoImpl) UpdateSupplier(id uint, supplier model.Supplier) error {
	return r.db.Model(&model.Supplier{}).Where("id = ?", id).Updates(supplier).Error
}

// 创建供应商
func (r *SupplierRepoImpl) CreateSupplier(supplier model.Supplier) (uint, error) {
	return supplier.Id, r.db.Create(&supplier).Error
}

// 软删除供应商
func (r *SupplierRepoImpl) DeleteSupplier(id uint) error {
	return r.db.Model(&model.Supplier{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
