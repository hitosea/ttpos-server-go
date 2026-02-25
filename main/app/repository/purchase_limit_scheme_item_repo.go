package repository

import (
	"time"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseLimitSchemeItemRepo 限购方案物品配置数据访问接口
type IPurchaseLimitSchemeItemRepo interface {
	// Create 创建物品配置
	Create(item *model.PurchaseLimitSchemeItem) error

	// BatchCreate 批量创建物品配置
	BatchCreate(items []*model.PurchaseLimitSchemeItem) error

	// GetBySchemeUuid 根据方案UUID查询所有物品配置
	GetBySchemeUuid(schemeUuid uint64, options ...DBOption) ([]*model.PurchaseLimitSchemeItem, error)

	// DeleteBySchemeUuid 根据方案UUID软删除所有物品配置
	DeleteBySchemeUuid(schemeUuid uint64) error

	// HardDeleteBySchemeUuid 根据方案UUID物理删除所有物品配置
	HardDeleteBySchemeUuid(schemeUuid uint64) error

	// 选项方法
	WhereSchemeUuid(schemeUuid uint64) DBOption
	WhereMaterialCode(materialCode string) DBOption
}

type purchaseLimitSchemeItemRepoImpl struct {
	db *gorm.DB // ✅ 只持有 db 实例
}

// NewPurchaseLimitSchemeItemRepo 创建物品配置仓储实例
func NewPurchaseLimitSchemeItemRepo(db *gorm.DB) IPurchaseLimitSchemeItemRepo {
	return &purchaseLimitSchemeItemRepoImpl{db: db}
}

// Create 创建物品配置
func (r *purchaseLimitSchemeItemRepoImpl) Create(item *model.PurchaseLimitSchemeItem) error {
	return r.db.Create(item).Error
}

// BatchCreate 批量创建物品配置
func (r *purchaseLimitSchemeItemRepoImpl) BatchCreate(items []*model.PurchaseLimitSchemeItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

// GetBySchemeUuid 根据方案UUID查询所有物品配置
func (r *purchaseLimitSchemeItemRepoImpl) GetBySchemeUuid(schemeUuid uint64, options ...DBOption) ([]*model.PurchaseLimitSchemeItem, error) {
	var items []*model.PurchaseLimitSchemeItem
	db := r.db.Where("delete_time = ?", 0).Where("scheme_uuid = ?", schemeUuid)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteBySchemeUuid 根据方案UUID软删除所有物品配置
func (r *purchaseLimitSchemeItemRepoImpl) DeleteBySchemeUuid(schemeUuid uint64) error {
	return r.db.Model(&model.PurchaseLimitSchemeItem{}).
		Where("scheme_uuid = ?", schemeUuid).
		Where("delete_time = ?", 0).
		Update("delete_time", time.Now().Unix()).Error
}

// HardDeleteBySchemeUuid 根据方案UUID物理删除所有物品配置
func (r *purchaseLimitSchemeItemRepoImpl) HardDeleteBySchemeUuid(schemeUuid uint64) error {
	return r.db.Unscoped().Where("scheme_uuid = ?", schemeUuid).Delete(&model.PurchaseLimitSchemeItem{}).Error
}

// 选项方法
func (r *purchaseLimitSchemeItemRepoImpl) WhereSchemeUuid(schemeUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("scheme_uuid = ?", schemeUuid)
	}
}

func (r *purchaseLimitSchemeItemRepoImpl) WhereMaterialCode(materialCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_code = ?", materialCode)
	}
}
