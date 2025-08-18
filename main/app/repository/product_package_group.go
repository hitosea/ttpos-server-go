package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductPackageGroupRepo 商品套餐组仓库接口
type IProductPackageGroupRepo interface {
	IProductPackageGroupQueryRepo
	CreateProductPackageGroup(productPackageGroup *model.ProductPackageGroup) error // 创建商品套餐组
	UpdateProductPackageGroup(data map[string]any, opts ...DBOption) error          // 更新商品套餐组
	DeleteProductPackageGroup(opts ...DBOption) error                               // 删除商品套餐组

	CreateProductPackageGroupItem(productPackageGroupItem *model.ProductPackageGroupItem) error // 创建商品套餐组商品
	UpdateProductPackageGroupItem(data map[string]any, opts ...DBOption) error                  // 更新商品套餐组商品
	DeleteProductPackageGroupItem(opts ...DBOption) error                                       // 删除商品套餐组商品
}

// IProductPackageGroupQueryRepo 商品套餐组查询仓库接口
type IProductPackageGroupQueryRepo interface {
	GetProductPackageGroup(opts ...DBOption) (*model.ProductPackageGroup, error)         // 获取商品套餐组
	GetProductPackageGroupItem(opts ...DBOption) (*model.ProductPackageGroupItem, error) // 获取商品套餐组商品
}

// productPackageGroupRepoImpl 商品套餐组仓库
type productPackageGroupRepoImpl struct {
	db *gorm.DB
}

// NewProductPackageGroupRepo 创建商品套餐组仓库
func NewProductPackageGroupRepo(db *gorm.DB) IProductPackageGroupRepo {
	return &productPackageGroupRepoImpl{db: db}
}

// CreateProductPackageGroup 创建商品套餐组
func (r *productPackageGroupRepoImpl) CreateProductPackageGroup(productPackageGroup *model.ProductPackageGroup) error {
	return r.db.Create(productPackageGroup).Error
}

// UpdateProductPackageGroup 更新商品套餐组
func (r *productPackageGroupRepoImpl) UpdateProductPackageGroup(data map[string]any, opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageGroup{}).Updates(data).Error
}

// DeleteProductPackageGroup 删除商品套餐组
func (r *productPackageGroupRepoImpl) DeleteProductPackageGroup(opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageGroup{}).Update("delete_time", time.Now().Unix()).Error
}

// CreateProductPackageGroupItem 创建商品套餐组商品
func (r *productPackageGroupRepoImpl) CreateProductPackageGroupItem(productPackageGroupItem *model.ProductPackageGroupItem) error {
	return r.db.Create(productPackageGroupItem).Error
}

// UpdateProductPackageGroupItem 更新商品套餐组商品
func (r *productPackageGroupRepoImpl) UpdateProductPackageGroupItem(data map[string]any, opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageGroupItem{}).Updates(data).Error
}

// DeleteProductPackageGroupItem 删除商品套餐组商品
func (r *productPackageGroupRepoImpl) DeleteProductPackageGroupItem(opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageGroupItem{}).Update("delete_time", time.Now().Unix()).Error
}

// GetProductPackageGroup 获取商品套餐组
func (r *productPackageGroupRepoImpl) GetProductPackageGroup(opts ...DBOption) (*model.ProductPackageGroup, error) {
	var productPackageGroup model.ProductPackageGroup
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&productPackageGroup).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &productPackageGroup, nil
}

// GetProductPackageGroupItem 获取商品套餐组商品
func (r *productPackageGroupRepoImpl) GetProductPackageGroupItem(opts ...DBOption) (*model.ProductPackageGroupItem, error) {
	var productPackageGroupItem model.ProductPackageGroupItem
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&productPackageGroupItem).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &productPackageGroupItem, nil
}
