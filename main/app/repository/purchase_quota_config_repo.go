package repository

import (
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseQuotaConfigRepo 限购配置数据访问接口
type IPurchaseQuotaConfigRepo interface {
	// GetByMaterialCodeAndShop 根据物品编码和门店UUID查询限购配置
	GetByMaterialCodeAndShop(materialCode string, shopUuid uint64, options ...DBOption) (*model.PurchaseQuotaConfig, error)

	// GetList 查询限购配置列表
	GetList(options ...DBOption) ([]*model.PurchaseQuotaConfig, int64, error)

	// Create 创建限购配置（主表）
	Create(config *model.PurchaseQuotaConfig) error

	// Update 更新限购配置（主表）
	Update(config *model.PurchaseQuotaConfig) error

	// Delete 软删除限购配置（主表和关联表）
	Delete(uuid uint64) error

	// BatchCreateShops 批量创建门店关联
	BatchCreateShops(configUuid uint64, shopUuids []uint64) error

	// DeleteShops 删除门店关联（软删除）
	DeleteShops(configUuid uint64) error

	// 选项方法
	WhereStatus(status uint8) DBOption
	WhereMaterialCode(materialCode string) DBOption
	WhereUuid(uuid uint64) DBOption
}

type purchaseQuotaConfigRepoImpl struct {
	db *gorm.DB // ✅ 只持有 db 实例
}

// NewPurchaseQuotaConfigRepo 创建限购配置仓储实例
func NewPurchaseQuotaConfigRepo(db *gorm.DB) IPurchaseQuotaConfigRepo {
	return &purchaseQuotaConfigRepoImpl{db: db}
}

// GetByMaterialCodeAndShop 根据物品编码和门店UUID查询限购配置
// 查询逻辑：apply_to_all_shops=1 OR 存在关联表记录
func (r *purchaseQuotaConfigRepoImpl) GetByMaterialCodeAndShop(
	materialCode string,
	companyUuid uint64,
	options ...DBOption,
) (*model.PurchaseQuotaConfig, error) {
	var config model.PurchaseQuotaConfig
	db := r.db.Where("ttpos_purchase_quota_config.delete_time = ?", 0).
		Where("ttpos_purchase_quota_config.status = ?", constant.PurchaseQuotaConfigStatusEnabled).
		Where("ttpos_purchase_quota_config.material_code = ?", materialCode).
		Where(`(
			ttpos_purchase_quota_config.apply_to_all_shops = 1 
			OR EXISTS (
				SELECT 1 FROM ttpos_purchase_quota_config_shop 
				WHERE ttpos_purchase_quota_config_shop.config_uuid = ttpos_purchase_quota_config.uuid
				AND ttpos_purchase_quota_config_shop.company_uuid = ?
				AND ttpos_purchase_quota_config_shop.delete_time = 0
			)
		)`, companyUuid)

	for _, option := range options {
		db = option(db)
	}

	if err := db.First(&config).Error; err != nil {
		return nil, err
	}

	return &config, nil
}

// Create 创建限购配置
func (r *purchaseQuotaConfigRepoImpl) Create(config *model.PurchaseQuotaConfig) error {
	return r.db.Create(config).Error
}

// Update 更新限购配置
func (r *purchaseQuotaConfigRepoImpl) Update(config *model.PurchaseQuotaConfig) error {
	return r.db.Save(config).Error
}

// Delete 软删除限购配置（同时软删除关联表记录）
func (r *purchaseQuotaConfigRepoImpl) Delete(uuid uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 软删除主表
		if err := tx.Model(&model.PurchaseQuotaConfig{}).
			Where("uuid = ?", uuid).
			Update("delete_time", time.Now().Unix()).Error; err != nil {
			return err
		}

		// 2. 软删除关联表
		if err := tx.Model(&model.PurchaseQuotaConfigShop{}).
			Where("config_uuid = ?", uuid).
			Update("delete_time", time.Now().Unix()).Error; err != nil {
			return err
		}

		return nil
	})
}

// BatchCreateShops 批量创建门店关联
func (r *purchaseQuotaConfigRepoImpl) BatchCreateShops(configUuid uint64, companyUuids []uint64) error {
	if len(companyUuids) == 0 {
		return nil
	}

	shops := make([]model.PurchaseQuotaConfigShop, 0, len(companyUuids))
	now := time.Now().Unix()

	for _, companyUuid := range companyUuids {
		shops = append(shops, model.PurchaseQuotaConfigShop{
			ConfigUuid:  configUuid,
			CompanyUuid: companyUuid,
			CreateTime:  now,
		})
	}

	return r.db.Create(&shops).Error
}

// DeleteShops 删除指定配置的所有门店关联（软删除）
func (r *purchaseQuotaConfigRepoImpl) DeleteShops(configUuid uint64) error {
	return r.db.Model(&model.PurchaseQuotaConfigShop{}).
		Where("config_uuid = ?", configUuid).
		Update("delete_time", time.Now().Unix()).Error
}

// GetList 查询限购配置列表
func (r *purchaseQuotaConfigRepoImpl) GetList(
	options ...DBOption,
) ([]*model.PurchaseQuotaConfig, int64, error) {
	var list []*model.PurchaseQuotaConfig
	var total int64

	db := r.db.Where("delete_time = ?", 0)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Model(&model.PurchaseQuotaConfig{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 选项方法
func (r *purchaseQuotaConfigRepoImpl) WhereStatus(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *purchaseQuotaConfigRepoImpl) WhereMaterialCode(materialCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_code = ?", materialCode)
	}
}

func (r *purchaseQuotaConfigRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}
