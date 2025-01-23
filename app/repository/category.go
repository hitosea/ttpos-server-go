package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// 商品类别
type ProductCategoryRepositoryInterface interface {
	GetProductCategoryList() ([]model.ProductCategory, error)
	UpdateProductCategory(id uint, productCategory model.ProductCategory) error
	CreateProductCategory(productCategory model.ProductCategory) error
	DeleteProductCategory(id uint) error
}

func NewProductCategoryRepository(db *gorm.DB) ProductCategoryRepositoryInterface {
	return NewProductCategoryRepositoryImpl(db)
}

// 原料类别
type MaterialCategoryRepositoryInterface interface {
	GetMaterialCategoryList() ([]model.MaterialCategory, error)
	UpdateMaterialCategory(id uint, materialCategory model.MaterialCategory) error
	CreateMaterialCategory(materialCategory model.MaterialCategory) error
	DeleteMaterialCategory(id uint) error
}

func NewMaterialCategoryRepository(db *gorm.DB) MaterialCategoryRepositoryInterface {
	return NewMaterialCategoryRepositoryImpl(db)
}

// 商品特殊类别
type ProductSpecialCategoryRepositoryInterface interface {
	GetProductSpecialCategoryList() ([]model.ProductSpecialCategory, error)
	UpdateProductSpecialCategory(id uint, productSpecialCategory model.ProductSpecialCategory) error
	CreateProductSpecialCategory(productSpecialCategory model.ProductSpecialCategory) error
	DeleteProductSpecialCategory(id uint) error
}

func NewProductSpecialCategoryRepository(db *gorm.DB) ProductSpecialCategoryRepositoryInterface {
	return NewProductSpecialCategoryRepositoryImpl(db)
}
