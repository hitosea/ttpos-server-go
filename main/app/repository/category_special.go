package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ProductSpecialCategoryRepoImpl 商品特殊类别
type ProductSpecialCategoryRepoImpl struct {
	db *gorm.DB
}

func NewProductSpecialCategoryRepoImpl(db *gorm.DB) *ProductSpecialCategoryRepoImpl {
	return &ProductSpecialCategoryRepoImpl{db: db}
}

// GetProductSpecialCategoryList 获取商品特殊类别列表
func (r *ProductSpecialCategoryRepoImpl) GetProductSpecialCategoryList() ([]model.ProductSpecialCategory, error) {
	var categories []model.ProductSpecialCategory
	err := r.db.Model(&model.ProductSpecialCategory{}).Find(&categories).Error
	return categories, err
}

// UpdateProductSpecialCategory 更新商品特殊类别
func (r *ProductSpecialCategoryRepoImpl) UpdateProductSpecialCategory(id uint, productSpecialCategory model.ProductSpecialCategory) error {
	return r.db.Model(&model.ProductSpecialCategory{}).Where("id = ?", id).Updates(productSpecialCategory).Error
}

// CreateProductSpecialCategory 创建商品特殊类别
func (r *ProductSpecialCategoryRepoImpl) CreateProductSpecialCategory(productSpecialCategory model.ProductSpecialCategory) (uint, error) {
	err := r.db.Create(&productSpecialCategory).Error
	return productSpecialCategory.Id, err
}

// DeleteProductSpecialCategory 软删除商品特殊类别
func (r *ProductSpecialCategoryRepoImpl) DeleteProductSpecialCategory(id uint) error {
	return r.db.Model(&model.ProductSpecialCategory{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

func (r *ProductSpecialCategoryRepoImpl) GetProductSpecialCategoryByIdWithMultiLanguageName(id uint) (*model.ProductSpecialCategory, error) {
	var productSpecialCategory model.ProductSpecialCategory
	err := r.db.Model(&model.ProductSpecialCategory{}).Where("id = ?", id).Preload("MultiLanguageName").First(&productSpecialCategory).Error
	return &productSpecialCategory, err
}

func (r *ProductSpecialCategoryRepoImpl) GetProductSpecialCategoryListWithMultiLanguageName() ([]model.ProductSpecialCategory, error) {
	var productSpecialCategories []model.ProductSpecialCategory
	err := r.db.Model(&model.ProductSpecialCategory{}).Preload("MultiLanguageName").Find(&productSpecialCategories).Error
	return productSpecialCategories, err
}
