package repository

import (
	"time"

	"gorm.io/gorm"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

// IProductFlavorRepo 定义商品规格仓库接口
type IProductFlavorRepo interface {
	GetProductFlavor(opts ...DBOption) (model.ProductFlavor, error)      // 查询单个商品规格
	WithMultiLanguageName() DBOption                                     // 预加载多语言名称
	CreateProductFlavor(productFlavor model.ProductFlavor) error         // 创建商品规格
	CreateProductFlavorList(productFlavors []model.ProductFlavor) error  // 创建商品规格列表
	UpdateProductFlavor(data map[string]any, opts ...DBOption) error     // 更新商品规格
	DeleteProductFlavor(opts ...DBOption) error                          // 删除商品规格
	GetProductFlavorList(uuids ...uint64) ([]model.ProductFlavor, error) // 获取商品规格列表
	DestroyProductFlavor(opts ...DBOption) error                         // 销毁商品规格
	UpdateNameByMultiLanguageNameUuid(multiLanguageNameUuid uint64, name string) error
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

// GetProductFlavorList 获取商品规格列表
func (r *ProductFlavorRepo) GetProductFlavorList(uuids ...uint64) ([]model.ProductFlavor, error) {
	if len(uuids) == 0 {
		return []model.ProductFlavor{}, nil
	}
	var flavors []model.ProductFlavor
	db := r.db.Model(&model.ProductFlavor{}).Where("delete_time = ?", 0).Where("uuid IN (?)", uuids)
	err := db.Find(&flavors).Error
	return flavors, errors.WithMessage(err)
}

// DestroyProductFlavor 销毁商品规格
func (r *ProductFlavorRepo) DestroyProductFlavor(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Delete(&model.ProductFlavor{}).Error
	return errors.WithMessage(err)
}

// CreateProductFlavorList 创建商品规格列表
func (r *ProductFlavorRepo) CreateProductFlavorList(productFlavors []model.ProductFlavor) error {
	err := r.db.Create(&productFlavors).Error
	return errors.WithMessage(err)
}

// UpdateNameByMultiLanguageNameUuid 根据多语言名称UUID更新name字段
func (r *ProductFlavorRepo) UpdateNameByMultiLanguageNameUuid(multiLanguageNameUuid uint64, name string) error {
	return r.db.Model(&model.ProductFlavor{}).Where("multi_language_name_uuid = ?", multiLanguageNameUuid).Update("name", name).Error
}
