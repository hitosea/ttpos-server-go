package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductMustPlanItemRepo 定义必点方案商品仓库接口
type IProductMustPlanItemRepo interface {
	GetProductMustPlanItem(opts ...DBOption) (model.ProductMustPlanItem, error)                  // 查询单个必点方案商品
	GetProductMustPlanItemList(opts ...DBOption) ([]model.ProductMustPlanItem, error)            // 查询必点方案商品列表
	GetProductMustPlanItemByPackageUuid(packageUuid uint64) ([]model.ProductMustPlanItem, error) // 通过商品包uuid查询必点方案商品
}

// ProductMustPlanItemRepo 定义必点方案商品仓库结构
type ProductMustPlanItemRepo struct {
	db *gorm.DB
}

// NewProductMustPlanItemRepo 实例化必点方案商品仓库
func NewProductMustPlanItemRepo(db *gorm.DB) IProductMustPlanItemRepo {
	return NewProductMustPlanItemRepoImpl(db)
}

// NewProductMustPlanItemRepoImpl 实例化必点方案商品仓库实现
func NewProductMustPlanItemRepoImpl(db *gorm.DB) IProductMustPlanItemRepo {
	return &ProductMustPlanItemRepo{
		db: db,
	}
}

// GetProductMustPlanItem 查询单个必点方案商品
func (r *ProductMustPlanItemRepo) GetProductMustPlanItem(opts ...DBOption) (model.ProductMustPlanItem, error) {
	var item model.ProductMustPlanItem
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&item).Error

	return item, errors.WithMessage(err)
}

// GetProductMustPlanItemList 查询必点方案商品列表
func (r *ProductMustPlanItemRepo) GetProductMustPlanItemList(opts ...DBOption) ([]model.ProductMustPlanItem, error) {
	var items []model.ProductMustPlanItem
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Find(&items).Error

	return items, errors.WithMessage(err)
}

// GetProductMustPlanItemByPackageUuid 通过商品包uuid查询必点方案商品
func (r *ProductMustPlanItemRepo) GetProductMustPlanItemByPackageUuid(packageUuid uint64) ([]model.ProductMustPlanItem, error) {
	var items []model.ProductMustPlanItem
	items, err := r.GetProductMustPlanItemList(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByProductPackageUuid(packageUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductMustPlan",
			},
		),
	)

	return items, errors.WithMessage(err)
}
