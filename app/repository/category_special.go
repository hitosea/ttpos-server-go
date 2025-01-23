package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// 商品特殊类别
type ProductSpecialCategoryRepository struct {
	db *gorm.DB
}

func NewProductSpecialCategoryRepositoryImpl(db *gorm.DB) *ProductSpecialCategoryRepository {
	return &ProductSpecialCategoryRepository{db: db}
}

// 获取商品特殊类别列表
func (r *ProductSpecialCategoryRepository) GetProductSpecialCategoryList() ([]model.ProductSpecialCategory, error) {
	var categories []model.ProductSpecialCategory
	err := r.db.Model(&model.ProductSpecialCategory{}).Find(&categories).Error
	return categories, err
}

// 更新商品特殊类别
func (r *ProductSpecialCategoryRepository) UpdateProductSpecialCategory(id uint, productSpecialCategory model.ProductSpecialCategory) error {
	return r.db.Model(&model.ProductSpecialCategory{}).Where("id = ?", id).Updates(productSpecialCategory).Error
}

// 创建商品特殊类别
func (r *ProductSpecialCategoryRepository) CreateProductSpecialCategory(productSpecialCategory model.ProductSpecialCategory) error {
	return r.db.Create(&productSpecialCategory).Error
}

// 软删除商品特殊类别
func (r *ProductSpecialCategoryRepository) DeleteProductSpecialCategory(id uint) error {
	return r.db.Model(&model.ProductSpecialCategory{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
