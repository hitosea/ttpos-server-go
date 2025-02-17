package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductMustPlanRepo 产品必点方案
type IProductMustPlanRepo interface {
	// 通过uuid列表获取产品必点方案列表
	GetProductMustPlanListByUuids(uuids []uint64) ([]model.ProductMustPlan, error)
	GetProductMustPlanByRegionUuid(regionUuid uint64) ([]model.ProductMustPlan, error)
}

func NewProductMustPlanRepo(db *gorm.DB) IProductMustPlanRepo {
	return NewProductMustPlanRepoImpl(db)
}

// NewProductMustPlanRepoImpl 创建新的产品必点方案仓库实现
func NewProductMustPlanRepoImpl(db *gorm.DB) *ProductMustPlanRepoImpl {
	return &ProductMustPlanRepoImpl{db: db}
}

type ProductMustPlanRepoImpl struct {
	db *gorm.DB
}

// GetProductMustPlanListByUuids 通过uuid列表获取产品必点方案列表
func (r *ProductMustPlanRepoImpl) GetProductMustPlanListByUuids(uuids []uint64) ([]model.ProductMustPlan, error) {
	var productMustPlans []model.ProductMustPlan
	err := r.db.Model(&model.ProductMustPlan{}).Where("uuid IN ? AND delete_time = ?", uuids, constant.NotDeleted).Find(&productMustPlans).Error
	return productMustPlans, err
}

// GetProductMustPlanByRegionUuid 通过区域uuid获取产品必点方案
func (r *ProductMustPlanRepoImpl) GetProductMustPlanByRegionUuid(regionUuid uint64) ([]model.ProductMustPlan, error) {
	var productMustPlanRegions []model.ProductMustPlanRegion
	err := r.db.Model(&model.ProductMustPlanRegion{}).Preload("ProductMustPlan", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ? AND delete_time = ?", constant.ProductMustPlanStatusOn, constant.NotDeleted)
	}).Preload("ProductMustPlan.ProductMustPlanItem").Where("desk_region_uuid = ? AND delete_time = ?", regionUuid, constant.NotDeleted).Find(&productMustPlanRegions).Error
	if err != nil {
		return nil, err
	}

	var productMustPlans []model.ProductMustPlan
	for _, productMustPlanRegion := range productMustPlanRegions {
		productMustPlans = append(productMustPlans, productMustPlanRegion.ProductMustPlan)
	}

	return productMustPlans, nil
}
