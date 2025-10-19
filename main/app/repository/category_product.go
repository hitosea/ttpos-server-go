package repository

import (
	"time"
	"ttpos-server-go/app/errors"
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
	return categories, errors.WithMessage(err)
}

// UpdateProductCategory 更新商品类别
func (r *ProductCategoryRepoImpl) UpdateProductCategory(id uint, productCategory model.ProductCategory) error {
	// 实现更新商品类别的逻辑
	return r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Updates(map[string]any{
		"name":                     productCategory.Name,
		"multi_language_name_uuid": productCategory.MultiLanguageNameUuid,
		"status":                   productCategory.Status,
		"parent_uuid":              productCategory.ParentUuid,
		"is_special":               productCategory.IsSpecial,
		"sort":                     productCategory.Sort,
		"code":                     productCategory.Code,
		"update_time":              uint(productCategory.UpdateTime),
	}).Error
}

// CreateProductCategory 创建商品类别
func (r *ProductCategoryRepoImpl) CreateProductCategory(productCategory model.ProductCategory) (uint64, error) {
	// 实现创建商品类别的逻辑
	err := r.db.Model(&model.ProductCategory{}).Create(&productCategory).Error
	return productCategory.Uuid, errors.WithMessage(err)
}

// DeleteProductCategory 软删除商品类别
func (r *ProductCategoryRepoImpl) DeleteProductCategory(id uint) error {
	// 实现软删除商品类别的逻辑
	return r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

func (r *ProductCategoryRepoImpl) GetProductCategoryByIdWithMultiLanguageName(id uint) (*model.ProductCategory, error) {
	var productCategory model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Where("id = ?", id).Preload("MultiLanguageName").First(&productCategory).Error
	return &productCategory, errors.WithMessage(err)
}

func (r *ProductCategoryRepoImpl) GetProductCategoryListWithMultiLanguageName() ([]model.ProductCategory, error) {
	var productCategories []model.ProductCategory
	err := r.db.Model(&model.ProductCategory{}).Preload("MultiLanguageName").Find(&productCategories).Error
	return productCategories, errors.WithMessage(err)
}
