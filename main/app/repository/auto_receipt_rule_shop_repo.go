package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IAutoReceiptRuleShopRepo interface {
	BatchCreate(shops []model.AutoReceiptRuleShop) error
	SoftDeleteByUuids(uuids []uint64, headquarterUuid uint64) error
	SoftDeleteByRuleUuids(ruleUuids []uint64, headquarterUuid uint64) error
	GetByUuids(uuids []uint64, headquarterUuid uint64) ([]model.AutoReceiptRuleShop, error)
	GetByRuleUuids(ruleUuids []uint64) ([]model.AutoReceiptRuleShop, error)
	CountByRuleUuid(ruleUuid uint64) (int64, error)
	GetConfiguredShopUuids(headquarterUuid uint64, warehouseErpCode string, excludeRuleUuid uint64) ([]uint64, error)
	GetAllEnabledWithRules() ([]RuleShopJoinResult, error)
}

// RuleShopJoinResult 规则+门店 JOIN 查询结果（定时任务用）
type RuleShopJoinResult struct {
	RuleUuid         uint64 `gorm:"column:rule_uuid"`
	WarehouseErpCode string `gorm:"column:warehouse_erp_code"`
	DelayDays        int    `gorm:"column:delay_days"`
	ShopCompanyUuid  uint64 `gorm:"column:shop_company_uuid"`
	HeadquarterUuid  uint64 `gorm:"column:headquarter_company_uuid"`
}

func NewAutoReceiptRuleShopRepo(db *gorm.DB) IAutoReceiptRuleShopRepo {
	return &autoReceiptRuleShopRepo{db: db}
}

type autoReceiptRuleShopRepo struct {
	db *gorm.DB
}

func (r *autoReceiptRuleShopRepo) BatchCreate(shops []model.AutoReceiptRuleShop) error {
	if len(shops) == 0 {
		return nil
	}
	err := r.db.Create(&shops).Error
	return errors.WithMessage(err)
}

func (r *autoReceiptRuleShopRepo) SoftDeleteByUuids(uuids []uint64, headquarterUuid uint64) error {
	// 通过 JOIN 规则主表确保租户隔离
	err := r.db.Exec(
		`UPDATE ttpos_auto_receipt_rule_shop s
		 JOIN ttpos_auto_receipt_rule r ON r.uuid = s.rule_uuid
		 SET s.delete_time = UNIX_TIMESTAMP()
		 WHERE s.uuid IN ? AND s.delete_time = 0
		 AND r.headquarter_company_uuid = ? AND r.delete_time = 0`,
		uuids, headquarterUuid).Error
	return errors.WithMessage(err)
}

func (r *autoReceiptRuleShopRepo) SoftDeleteByRuleUuids(ruleUuids []uint64, headquarterUuid uint64) error {
	// 通过 JOIN 规则主表确保租户隔离
	err := r.db.Exec(
		`UPDATE ttpos_auto_receipt_rule_shop s
		 JOIN ttpos_auto_receipt_rule r ON r.uuid = s.rule_uuid
		 SET s.delete_time = UNIX_TIMESTAMP()
		 WHERE s.rule_uuid IN ? AND s.delete_time = 0
		 AND r.headquarter_company_uuid = ? AND r.delete_time = 0`,
		ruleUuids, headquarterUuid).Error
	return errors.WithMessage(err)
}

func (r *autoReceiptRuleShopRepo) GetByUuids(uuids []uint64, headquarterUuid uint64) ([]model.AutoReceiptRuleShop, error) {
	var shops []model.AutoReceiptRuleShop
	if len(uuids) == 0 {
		return shops, nil
	}
	// 通过 JOIN 规则主表确保租户隔离
	err := r.db.Table("ttpos_auto_receipt_rule_shop AS s").
		Select("s.*").
		Joins("JOIN ttpos_auto_receipt_rule AS r ON r.uuid = s.rule_uuid").
		Where("s.uuid IN ? AND s.delete_time = 0 AND r.headquarter_company_uuid = ? AND r.delete_time = 0",
			uuids, headquarterUuid).
		Find(&shops).Error
	return shops, errors.WithMessage(err)
}

func (r *autoReceiptRuleShopRepo) GetByRuleUuids(ruleUuids []uint64) ([]model.AutoReceiptRuleShop, error) {
	var shops []model.AutoReceiptRuleShop
	if len(ruleUuids) == 0 {
		return shops, nil
	}
	err := r.db.Where("rule_uuid IN ? AND delete_time = 0", ruleUuids).Find(&shops).Error
	return shops, errors.WithMessage(err)
}

func (r *autoReceiptRuleShopRepo) CountByRuleUuid(ruleUuid uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.AutoReceiptRuleShop{}).
		Where("rule_uuid = ? AND delete_time = 0", ruleUuid).
		Count(&count).Error
	return count, errors.WithMessage(err)
}

// GetConfiguredShopUuids 查询同仓库下已配置的门店UUID列表（同一门店在同一仓库内只允许配置一次）
func (r *autoReceiptRuleShopRepo) GetConfiguredShopUuids(headquarterUuid uint64, warehouseErpCode string, excludeRuleUuid uint64) ([]uint64, error) {
	var shopUuids []uint64
	db := r.db.Model(&model.AutoReceiptRuleShop{}).
		Select("s.shop_company_uuid").
		Table("ttpos_auto_receipt_rule_shop AS s").
		Joins("JOIN ttpos_auto_receipt_rule AS r ON r.uuid = s.rule_uuid").
		Where("r.headquarter_company_uuid = ? AND r.warehouse_erp_code = ? AND r.delete_time = 0 AND s.delete_time = 0",
			headquarterUuid, warehouseErpCode)
	if excludeRuleUuid > 0 {
		db = db.Where("r.uuid != ?", excludeRuleUuid)
	}
	err := db.Pluck("s.shop_company_uuid", &shopUuids).Error
	return shopUuids, errors.WithMessage(err)
}

// GetAllEnabledWithRules 获取所有启用规则的门店（JOIN查询，定时任务用）
func (r *autoReceiptRuleShopRepo) GetAllEnabledWithRules() ([]RuleShopJoinResult, error) {
	var results []RuleShopJoinResult
	err := r.db.Table("ttpos_auto_receipt_rule_shop AS s").
		Select("s.rule_uuid, r.warehouse_erp_code, r.delay_days, r.headquarter_company_uuid, s.shop_company_uuid").
		Joins("JOIN ttpos_auto_receipt_rule AS r ON r.uuid = s.rule_uuid").
		Where("r.status = 1 AND r.delete_time = 0 AND s.delete_time = 0").
		Find(&results).Error
	return results, errors.WithMessage(err)
}
