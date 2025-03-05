package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductAttributeRepo 产品属性仓库接口
type IProductAttributeRepo interface {
	GetProductAttributeList() ([]model.ProductAttribute, error)                     // 获取产品属性列表
	UpdateProductAttribute(id uint, productAttribute model.ProductAttribute) error  // 更新产品属性
	CreateProductAttribute(productAttribute model.ProductAttribute) (uint64, error) // 创建产品属性
	DeleteProductAttribute(id uint) error                                           // 删除产品属性
}

// NewProductAttributeRepo 创建新的产品属性仓库
func NewProductAttributeRepo(db *gorm.DB) IProductAttributeRepo {
	return NewProductAttributeRepoImpl(db)
}

// NewProductAttributeRepoImpl 创建新的产品属性仓库实现
func NewProductAttributeRepoImpl(db *gorm.DB) *ProductAttributeRepoImpl {
	return &ProductAttributeRepoImpl{db: db}
}

type ProductAttributeRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetProductAttributeList 获取产品属性列表，排除逻辑删除的产品属性
func (r *ProductAttributeRepoImpl) GetProductAttributeList() ([]model.ProductAttribute, error) {
	var productAttributes []model.ProductAttribute
	err := r.db.Model(&model.ProductAttribute{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productAttributes).Error
	return productAttributes, err
}

// UpdateProductAttribute 更新产品属性
func (r *ProductAttributeRepoImpl) UpdateProductAttribute(id uint, productAttribute model.ProductAttribute) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ProductAttribute{}).Where("id = ?", id).Updates(productAttribute).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&productAttribute.MultiLanguageName).Where("id = ?", productAttribute.MultiLanguageNameUuid).Updates(productAttribute.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// CreateProductAttribute 创建产品属性
func (r *ProductAttributeRepoImpl) CreateProductAttribute(productAttribute model.ProductAttribute) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&productAttribute.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, err
	}

	// 将多语言名称ID存入产品属性
	productAttribute.MultiLanguageNameUuid = productAttribute.MultiLanguageName.Uuid

	// 创建产品属性
	if err := tx.Create(&productAttribute).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, err
	}

	return productAttribute.Uuid, tx.Commit().Error // 提交事务
}

// DeleteProductAttribute 软删除产品属性
func (r *ProductAttributeRepoImpl) DeleteProductAttribute(id uint) error {
	return r.db.Model(&model.ProductAttribute{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
