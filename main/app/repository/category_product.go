package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ProductCategoryRepository 实现 ProductCategoryRepositoryInterface
type ProductCategoryRepository struct {
	db *gorm.DB
}

func NewProductCategoryRepositoryImpl(db *gorm.DB) *ProductCategoryRepository {
	return &ProductCategoryRepository{db: db}
}

// 获取商品类别列表
func (r *ProductCategoryRepository) GetProductCategoryList() ([]model.ProductCategory, error) {
	// 实现获取商品类别列表的逻辑
	var categories []model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Find(&categories).Error
	return categories, err
}

// 更新商品类别
func (r *ProductCategoryRepository) UpdateProductCategory(id uint, productCategory model.ProductCategory) error {
	// 实现更新商品类别的逻辑
	return r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Updates(productCategory).Error
}

// 创建商品类别
func (r *ProductCategoryRepository) CreateProductCategory(productCategory model.ProductCategory) (uint, error) {
	// 实现创建商品类别的逻辑
	err := r.db.Create(&productCategory).Error
	return productCategory.Id, err
}

// 软删除商品类别
func (r *ProductCategoryRepository) DeleteProductCategory(id uint) error {
	// 实现软删除商品类别的逻辑
	return r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

func (r *ProductCategoryRepository) GetProductCategoryByIdWithMultiLanguageName(id uint) (*model.ProductCategory, error) {
	var productCategory model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Preload("MultiLanguageName").First(&productCategory).Error
	return &productCategory, err
}

func (r *ProductCategoryRepository) GetProductCategoryListWithMultiLanguageName() ([]model.ProductCategory, error) {
	var productCategories []model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Preload("MultiLanguageName").Find(&productCategories).Error
	return productCategories, err
}
