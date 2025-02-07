package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// 产品包
type ProductPackageRepoInterface interface {
	GetProductPackageList() ([]model.ProductPackage, error)                  // 获取产品包列表
	UpdateProductPackage(id uint, productPackage model.ProductPackage) error // 更新产品包
	CreateProductPackage(productPackage model.ProductPackage) (uint, error)  // 创建产品包
	DeleteProductPackage(id uint) error                                      // 软删除产品包
}

func NewProductPackageRepo(db *gorm.DB) ProductPackageRepoInterface {
	return NewProductPackageRepoImpl(db)
}

// 创建新的产品包仓库实现
func NewProductPackageRepoImpl(db *gorm.DB) *ProductPackageRepoImpl {
	return &ProductPackageRepoImpl{db: db}
}

type ProductPackageRepoImpl struct {
	db *gorm.DB
}

// 获取产品包列表，排除逻辑删除的产品包
func (r *ProductPackageRepoImpl) GetProductPackageList() ([]model.ProductPackage, error) {
	var productPackages []model.ProductPackage
	err := r.db.Model(&model.ProductPackage{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productPackages).Error
	return productPackages, err
}

// 更新产品包
func (r *ProductPackageRepoImpl) UpdateProductPackage(id uint, productPackage model.ProductPackage) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ProductPackage{}).Where("id = ?", id).Updates(productPackage).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return err
	}

	if err := tx.Model(&productPackage.MultiLanguageName).Where("id = ?", productPackage.MultiLanguageNameUuid).Updates(productPackage.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return err
	}

	return tx.Commit().Error // 提交事务
}

// 创建产品包
func (r *ProductPackageRepoImpl) CreateProductPackage(productPackage model.ProductPackage) (uint, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&productPackage.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, err
	}

	// 创建产品包
	if err := tx.Create(&productPackage).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, err
	}

	return productPackage.Uuid, tx.Commit().Error // 提交事务
}

// 软删除产品包
func (r *ProductPackageRepoImpl) DeleteProductPackage(id uint) error {
	return r.db.Model(&model.ProductPackage{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
