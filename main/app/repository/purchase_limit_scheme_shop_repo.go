package repository

import (
	"time"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseLimitSchemeShopRepo 限购方案门店配置数据访问接口
type IPurchaseLimitSchemeShopRepo interface {
	// Create 创建门店配置
	Create(shop *model.PurchaseLimitSchemeShop) error

	// BatchCreate 批量创建门店配置
	BatchCreate(shops []*model.PurchaseLimitSchemeShop) error

	// GetBySchemeUuid 根据方案UUID查询所有门店配置
	GetBySchemeUuid(schemeUuid uint64, options ...DBOption) ([]*model.PurchaseLimitSchemeShop, error)

	// DeleteBySchemeUuid 根据方案UUID软删除所有门店配置
	DeleteBySchemeUuid(schemeUuid uint64) error

	// 选项方法
	WhereSchemeUuid(schemeUuid uint64) DBOption
	WhereCompanyUuid(companyUuid uint64) DBOption
}

type purchaseLimitSchemeShopRepoImpl struct {
	db *gorm.DB // ✅ 只持有 db 实例
}

// NewPurchaseLimitSchemeShopRepo 创建门店配置仓储实例
func NewPurchaseLimitSchemeShopRepo(db *gorm.DB) IPurchaseLimitSchemeShopRepo {
	return &purchaseLimitSchemeShopRepoImpl{db: db}
}

// Create 创建门店配置
func (r *purchaseLimitSchemeShopRepoImpl) Create(shop *model.PurchaseLimitSchemeShop) error {
	return r.db.Create(shop).Error
}

// BatchCreate 批量创建门店配置
func (r *purchaseLimitSchemeShopRepoImpl) BatchCreate(shops []*model.PurchaseLimitSchemeShop) error {
	if len(shops) == 0 {
		return nil
	}
	return r.db.Create(&shops).Error
}

// GetBySchemeUuid 根据方案UUID查询所有门店配置
func (r *purchaseLimitSchemeShopRepoImpl) GetBySchemeUuid(schemeUuid uint64, options ...DBOption) ([]*model.PurchaseLimitSchemeShop, error) {
	var shops []*model.PurchaseLimitSchemeShop
	db := r.db.Where("delete_time = ?", 0).Where("scheme_uuid = ?", schemeUuid)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Find(&shops).Error; err != nil {
		return nil, err
	}
	return shops, nil
}

// DeleteBySchemeUuid 根据方案UUID软删除所有门店配置
func (r *purchaseLimitSchemeShopRepoImpl) DeleteBySchemeUuid(schemeUuid uint64) error {
	return r.db.Model(&model.PurchaseLimitSchemeShop{}).
		Where("scheme_uuid = ?", schemeUuid).
		Where("delete_time = ?", 0).
		Update("delete_time", time.Now().Unix()).Error
}

// 选项方法
func (r *purchaseLimitSchemeShopRepoImpl) WhereSchemeUuid(schemeUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("scheme_uuid = ?", schemeUuid)
	}
}

func (r *purchaseLimitSchemeShopRepoImpl) WhereCompanyUuid(companyUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("company_uuid = ?", companyUuid)
	}
}
