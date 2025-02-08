package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductRepo 定义商品仓库接口
type IProductRepo interface {
	GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) // 分页获取商品列表
	GetProductCategoryList(opts ...DBOption) ([]model.ProductCategory, error)                                       // 获取产品类别列表
}

// productRepo 商品仓库
type productRepo struct {
	db *gorm.DB
}

// NewProductRepo 创建新的商品仓库
func NewProductRepo(db *gorm.DB) IProductRepo {
	return NewProductRepoImpl(db)
}

// NewProductRepoImpl 创建新的商品仓库实现
func NewProductRepoImpl(db *gorm.DB) IProductRepo {
	return &productRepo{db: db}
}

// GetProductListWithPagination 分页获取商品列表
func (r *productRepo) GetProductListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.ProductPackage, int64, error) {
	var products []model.ProductPackage
	var total int64

	db := r.db.Model(&model.ProductPackage{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&products).Error

	return products, total, err
}

// GetProductCategoryList 获取产品类别列表
func (r *productRepo) GetProductCategoryList(opts ...DBOption) ([]model.ProductCategory, error) {
	var categories []model.ProductCategory

	db := r.db.Model(&model.ProductCategory{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}
