package repository

import (
	"time"

	"gorm.io/gorm"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

// IProductFlavorRepo 定义商品规格仓库接口
type IProductFlavorRepo interface {
	GetProductFlavor(opts ...DBOption) (model.ProductFlavor, error)  // 查询单个商品规格
	WithMultiLanguageName() DBOption                                 // 预加载多语言名称
	CreateProductFlavor(productFlavor model.ProductFlavor) error     // 创建商品规格
	UpdateProductFlavor(data map[string]any, opts ...DBOption) error // 更新商品规格
	DeleteProductFlavor(opts ...DBOption) error                      // 删除商品规格
}

// ProductFlavorRepo 定义商品规格仓库结构
type ProductFlavorRepo struct {
	db *gorm.DB
}

// NewProductFlavorRepo 实例化商品规格仓库
func NewProductFlavorRepo(db *gorm.DB) IProductFlavorRepo {
	return NewProductFlavorRepoImpl(db)
}

// NewProductFlavorRepoImpl 实例化商品规格仓库实现
func NewProductFlavorRepoImpl(db *gorm.DB) IProductFlavorRepo {
	return &ProductFlavorRepo{
		db: db,
	}
}

// GetProductFlavor 查询单个商品规格
func (r *ProductFlavorRepo) GetProductFlavor(opts ...DBOption) (model.ProductFlavor, error) {
	var flavor model.ProductFlavor
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&flavor).Error

	return flavor, errors.WithMessage(err)
}

// WithMultiLanguageName 预加载多语言名称
func (r *ProductFlavorRepo) WithMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MultiLanguageName")
	}
}

// CreateProductFlavor 创建商品规格
func (r *ProductFlavorRepo) CreateProductFlavor(productFlavor model.ProductFlavor) error {
	err := r.db.Create(&productFlavor).Error
	return errors.WithMessage(err)
}

// UpdateProductFlavor 更新商品规格
func (r *ProductFlavorRepo) UpdateProductFlavor(data map[string]any, opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Model(&model.ProductFlavor{}).Updates(data).Error
	return errors.WithMessage(err)
}

// DeleteProductFlavor 删除商品规格
func (r *ProductFlavorRepo) DeleteProductFlavor(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Model(&model.ProductFlavor{}).Update("delete_time", time.Now().Unix()).Error
	return errors.WithMessage(err)
}
