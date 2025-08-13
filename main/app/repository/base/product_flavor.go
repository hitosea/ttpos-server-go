package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductFlavorRepo 商品规格
type IProductFlavorRepo interface {
	GetProductFlavorList() ([]model.ProductFlavor, error)
	UpdateProductFlavor(id uint, productFlavor model.ProductFlavor) error
	CreateProductFlavor(productFlavor model.ProductFlavor) (uint64, error)
	DeleteProductFlavor(id uint) error
	GetProductFlavorUuidByNameOptimized(name string) (uint64, error)
}

func NewProductFlavorRepo(db *gorm.DB) IProductFlavorRepo {
	return NewProductFlavorRepoImpl(db)
}

// NewProductFlavorRepoImpl 创建新的商品规格仓库实现
func NewProductFlavorRepoImpl(db *gorm.DB) *ProductFlavorRepoImpl {
	return &ProductFlavorRepoImpl{db: db}
}

type ProductFlavorRepoImpl struct {
	db *gorm.DB
}

// GetProductFlavorList 获取商品规格列表，排除逻辑删除的规格
func (r *ProductFlavorRepoImpl) GetProductFlavorList() ([]model.ProductFlavor, error) {
	var productFlavors []model.ProductFlavor
	err := r.db.Model(&model.ProductFlavor{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productFlavors).Error
	return productFlavors, errors.WithMessage(err)
}

// UpdateProductFlavor 更新商品规格
func (r *ProductFlavorRepoImpl) UpdateProductFlavor(id uint, productFlavor model.ProductFlavor) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ProductFlavor{}).Where("id = ?", id).Updates(productFlavor).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&productFlavor.MultiLanguageName).Where("id = ?", productFlavor.MultiLanguageNameUuid).Updates(productFlavor.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// CreateProductFlavor 创建商品规格
func (r *ProductFlavorRepoImpl) CreateProductFlavor(productFlavor model.ProductFlavor) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&productFlavor.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	// 创建商品规格
	if err := tx.Create(&productFlavor).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	return productFlavor.Uuid, tx.Commit().Error // 提交事务
}

// DeleteProductFlavor 软删除商品规格
func (r *ProductFlavorRepoImpl) DeleteProductFlavor(id uint) error {
	return r.db.Model(&model.ProductFlavor{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

// GetProductFlavorUuidByNameOptimized 获取商品规格UUID
func (r *ProductFlavorRepoImpl) GetProductFlavorUuidByNameOptimized(name string) (uint64, error) {
	var productFlavors []model.ProductFlavor
	err := r.db.Model(&model.ProductFlavor{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productFlavors).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	// 在内存中查找匹配的商品规格
	for _, productFlavor := range productFlavors {
		if productFlavor.MultiLanguageName.Uuid != 0 {
			names := productFlavor.MultiLanguageName.GetNames()
			if names.ZH == name || names.ZHTW == name || names.EN == name ||
				names.TH == name || names.MY == name || names.JA == name ||
				names.KO == name || names.TR == name || names.SV == name {
				return productFlavor.Uuid, nil
			}
		}
	}
	return 0, nil
}
