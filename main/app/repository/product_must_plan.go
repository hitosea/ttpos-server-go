package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// IProductMustPlanRepo 产品必点方案
type IProductMustPlanRepo interface {
	IProductMustPlanQueryRepo
	CreateProductMustPlan(plan *model.ProductMustPlan) (*model.ProductMustPlan, error)
	CreateProductMustPlanItem(planItems []model.ProductMustPlanItem) error
	CreateProductMustPlanRegion(planRegions []model.ProductMustPlanRegion) error
}

type IProductMustPlanQueryRepo interface {
	GetProductMustPlanList(ctx context.Context, opts ...DBOption) ([]*model.ProductMustPlan, error)
	GetProductMustPlanListAllInfos(ctx context.Context) ([]*model.ProductMustPlan, error)
	GetProductMustPlanListByUuids(uuids []uint64) ([]model.ProductMustPlan, error) //  通过uuid列表获取产品必点方案列表
	GetProductMustPlanByRegionUuid(regionUuid uint64) ([]model.ProductMustPlan, error)
	GetProductMustPlanListDeskInfos(ctx context.Context) ([]*model.ProductMustPlan, error)
	HasActiveProductMustPlan(ctx context.Context) (bool, error) // 判断商户是否有生效的必点方案
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
	return productMustPlans, errors.WithMessage(err)
}

// GetProductMustPlanByRegionUuid 通过区域uuid获取产品必点方案
func (r *ProductMustPlanRepoImpl) GetProductMustPlanByRegionUuid(regionUuid uint64) ([]model.ProductMustPlan, error) {
	var productMustPlanRegions []model.ProductMustPlanRegion
	err := r.db.Model(&model.ProductMustPlanRegion{}).Preload("ProductMustPlan", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ? AND delete_time = ?", constant.ProductMustPlanStatusOn, constant.NotDeleted)
	}).Preload("ProductMustPlan.ProductMustPlanItems").Where("desk_region_uuid = ? AND delete_time = ?", regionUuid, constant.NotDeleted).Find(&productMustPlanRegions).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	var productMustPlans []model.ProductMustPlan
	for _, productMustPlanRegion := range productMustPlanRegions {
		productMustPlans = append(productMustPlans, productMustPlanRegion.ProductMustPlan)
	}

	return productMustPlans, nil
}

// GetProductMustPlanList 获取产品必点方案列表
func (r *ProductMustPlanRepoImpl) GetProductMustPlanList(ctx context.Context, opts ...DBOption) ([]*model.ProductMustPlan, error) {
	productMustPlans := make([]*model.ProductMustPlan, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productMustPlans)
	if result.Error != nil {
		return nil, result.Error
	}

	return productMustPlans, nil
}

// CreateProductMustPlan 创建商品必点方案
func (r *ProductMustPlanRepoImpl) CreateProductMustPlan(plan *model.ProductMustPlan) (*model.ProductMustPlan, error) {
	err := r.db.Create(plan).Error
	return plan, err
}

// CreateProductMustPlanItem 创建商品必点方案商品
func (r *ProductMustPlanRepoImpl) CreateProductMustPlanItem(planItems []model.ProductMustPlanItem) error {
	err := r.db.Create(&planItems).Error
	return err
}

// CreateProductMustPlanRegion 创建商品必点方案区域
func (r *ProductMustPlanRepoImpl) CreateProductMustPlanRegion(planRegions []model.ProductMustPlanRegion) error {
	err := r.db.Create(&planRegions).Error
	return err
}

// GetProductMustPlanListAllInfos 获取搜索必点商品方案列表的数据信息
func (r *ProductMustPlanRepoImpl) GetProductMustPlanListAllInfos(ctx context.Context) ([]*model.ProductMustPlan, error) {
	productMustPlans, err := r.GetProductMustPlanList(ctx,
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByStatus(constant.ProductMustPlanStatusOn),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductUnit.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ImageFile",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductBoms",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductBoms.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductBoms.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductPackageAttributeGroups.ProductAttributeGroup.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductPackageAttributeGroups.ProductPackageAttributes.Attribute.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return productMustPlans, nil
}

// GetProductMustPlanListDeskInfos 获取搜索必点商品方案列表的数据信息
func (r *ProductMustPlanRepoImpl) GetProductMustPlanListDeskInfos(ctx context.Context) ([]*model.ProductMustPlan, error) {
	productMustPlans, err := r.GetProductMustPlanList(ctx,
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByStatus(constant.ProductMustPlanStatusOn),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductUnit.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ImageFile",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductBoms",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductBoms.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductBoms.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductPackageAttributeGroups",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductPackageAttributeGroups.ProductAttributeGroup.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanItems.ProductPackage.ProductPackageAttributeGroups.ProductPackageAttributes.Attribute.MultiLanguageName",
			},
			WithPreload{
				Query: "ProductMustPlanRegions.DeskRegion.Desks",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return productMustPlans, nil
}

// HasActiveProductMustPlan 判断商户是否有生效的必点方案
// 生效条件：
// 1. 必点方案未删除（delete_time = 0）
// 2. 必点方案状态为开启（status = 1）
func (r *ProductMustPlanRepoImpl) HasActiveProductMustPlan(ctx context.Context) (bool, error) {
	// 获取商户UUID
	companyUuid := GetCompanyUuid(r.db)
	if companyUuid == 0 {
		// 如果无法获取商户UUID，直接查询数据库
		return r.queryHasActiveProductMustPlan()
	}

	// 检查是否启用对象存储缓存
	if !adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 未启用缓存，直接查询数据库
		return r.queryHasActiveProductMustPlan()
	}

	// 使用对象存储模块缓存查询
	return r.hasActiveProductMustPlanWithCache(companyUuid)
}

// queryHasActiveProductMustPlan 查询商户是否有生效的必点方案（数据库查询）
func (r *ProductMustPlanRepoImpl) queryHasActiveProductMustPlan() (bool, error) {
	var count int64
	err := r.db.Model(&model.ProductMustPlan{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("status = ?", constant.ProductMustPlanStatusOn).
		Count(&count).Error
	if err != nil {
		return false, errors.WithMessage(err)
	}
	return count > 0, nil
}

// hasActiveProductMustPlanWithCache 使用对象存储模块缓存查询商户是否有生效的必点方案
func (r *ProductMustPlanRepoImpl) hasActiveProductMustPlanWithCache(companyUuid uint64) (bool, error) {
	// 构建缓存 key（使用固定的 uuid=0 表示商户级别的查询）
	key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeProductMustPlanActive, 0)

	// 定义 bool 包装类型用于缓存
	type BoolValue struct {
		Value bool `json:"value"`
	}

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*BoolValue](cache.Global, 5*time.Minute)

	// 使用缓存查询
	result, err := cacheLayer.GET(key, func() (*BoolValue, error) {
		// 缓存未命中时，从数据库查询
		hasActive, err := r.queryHasActiveProductMustPlan()
		if err != nil {
			return &BoolValue{Value: false}, errors.WithMessage(err)
		}
		return &BoolValue{Value: hasActive}, nil
	})

	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.queryHasActiveProductMustPlan()
	}

	return result.Value, nil
}
