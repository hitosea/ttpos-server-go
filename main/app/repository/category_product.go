package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ProductCategoryRepoImpl 实现 IProductCategoryRepo
type ProductCategoryRepoImpl struct {
	db *gorm.DB
}

func NewProductCategoryRepoImpl(db *gorm.DB) *ProductCategoryRepoImpl {
	return &ProductCategoryRepoImpl{db: db}
}

// GetProductCategoryList 获取商品类别列表
func (r *ProductCategoryRepoImpl) GetProductCategoryList() ([]model.ProductCategory, error) {
	// 实现获取商品类别列表的逻辑
	var categories []model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Find(&categories).Error
	return categories, err
}

// UpdateProductCategory 更新商品类别
func (r *ProductCategoryRepoImpl) UpdateProductCategory(id uint, productCategory model.ProductCategory) error {
	// 实现更新商品类别的逻辑
	return r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Updates(productCategory).Error
}

// CreateProductCategory 创建商品类别
func (r *ProductCategoryRepoImpl) CreateProductCategory(productCategory model.ProductCategory) (uint, error) {
	// 实现创建商品类别的逻辑
	err := r.db.Create(&productCategory).Error
	return productCategory.ID, err
}

// DeleteProductCategory 软删除商品类别
func (r *ProductCategoryRepoImpl) DeleteProductCategory(id uint) error {
	// 实现软删除商品类别的逻辑
	return r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

func (r *ProductCategoryRepoImpl) GetProductCategoryByIdWithMultiLanguageName(id uint) (*model.ProductCategory, error) {
	var productCategory model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Preload("MultiLanguageName").First(&productCategory).Error
	return &productCategory, err
}

func (r *ProductCategoryRepoImpl) GetProductCategoryListWithMultiLanguageName() ([]model.ProductCategory, error) {
	var productCategories []model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Preload("MultiLanguageName").Find(&productCategories).Error
	return productCategories, err
}
