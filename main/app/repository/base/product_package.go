package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ProductPackageRepoInterface 产品包
type ProductPackageRepoInterface interface {
	GetProductPackageList() ([]model.ProductPackage, error)                      // 获取产品包列表
	UpdateProductPackage(id uint, productPackage model.ProductPackage) error     // 更新产品包
	UpdateProductPackageActualSaleNum(productPackage model.ProductPackage) error // 更新产品包的实际销量
	CreateProductPackage(productPackage model.ProductPackage) (uint64, error)    // 创建产品包
	DeleteProductPackage(id uint) error                                          // 软删除产品包
}

func NewProductPackageRepo(db *gorm.DB) ProductPackageRepoInterface {
	return NewProductPackageRepoImpl(db)
}

// NewProductPackageRepoImpl 创建新的产品包仓库实现
func NewProductPackageRepoImpl(db *gorm.DB) *ProductPackageRepoImpl {
	return &ProductPackageRepoImpl{db: db}
}

type ProductPackageRepoImpl struct {
	db *gorm.DB
}

// GetProductPackageList 获取产品包列表，排除逻辑删除的产品包
func (r *ProductPackageRepoImpl) GetProductPackageList() ([]model.ProductPackage, error) {
	var productPackages []model.ProductPackage
	err := r.db.Model(&model.ProductPackage{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productPackages).Error
	return productPackages, errors.WithMessage(err)
}

// UpdateProductPackage 更新产品包
func (r *ProductPackageRepoImpl) UpdateProductPackage(id uint, productPackage model.ProductPackage) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ProductPackage{}).Where("id = ?", id).Updates(productPackage).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&productPackage.MultiLanguageName).Where("id = ?", productPackage.MultiLanguageNameUuid).Updates(productPackage.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// UpdateProductPackageActualSaleNum 更新产品包的实际销量
func (r *ProductPackageRepoImpl) UpdateProductPackageActualSaleNum(productPackage model.ProductPackage) error {
	return r.db.Model(&model.ProductPackage{}).Where("uuid = ?", productPackage.Uuid).Update("actual_sale_num", productPackage.ActualSaleNum).Error
}

// CreateProductPackage 创建产品包
func (r *ProductPackageRepoImpl) CreateProductPackage(productPackage model.ProductPackage) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&productPackage.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	// 创建产品包
	if err := tx.Create(&productPackage).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	return productPackage.Uuid, tx.Commit().Error // 提交事务
}

// DeleteProductPackage 软删除产品包
func (r *ProductPackageRepoImpl) DeleteProductPackage(id uint) error {
	return r.db.Model(&model.ProductPackage{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
