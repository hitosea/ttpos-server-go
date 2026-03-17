package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IAutoReceiptRuleRepo interface {
	Create(rule model.AutoReceiptRule) (model.AutoReceiptRule, error)
	Update(uuid uint64, headquarterUuid uint64, vars map[string]any) error
	SoftDelete(uuids []uint64, headquarterUuid uint64) error
	GetByUuid(uuid uint64, headquarterUuid uint64) (model.AutoReceiptRule, error)
	GetListByHeadquarter(headquarterUuid uint64, warehouseErpCode string) ([]model.AutoReceiptRule, error)
	GetAllEnabled() ([]model.AutoReceiptRule, error)
}

func NewAutoReceiptRuleRepo(db *gorm.DB) IAutoReceiptRuleRepo {
	return &autoReceiptRuleRepo{db: db}
}

type autoReceiptRuleRepo struct {
	db *gorm.DB
}

func (r *autoReceiptRuleRepo) Create(rule model.AutoReceiptRule) (model.AutoReceiptRule, error) {
	err := r.db.Create(&rule).Error
	return rule, errors.WithMessage(err)
}

func (r *autoReceiptRuleRepo) Update(uuid uint64, headquarterUuid uint64, vars map[string]any) error {
	err := r.db.Model(&model.AutoReceiptRule{}).
		Where("uuid = ? AND headquarter_company_uuid = ? AND delete_time = 0", uuid, headquarterUuid).
		Updates(vars).Error
	return errors.WithMessage(err)
}

func (r *autoReceiptRuleRepo) SoftDelete(uuids []uint64, headquarterUuid uint64) error {
	err := r.db.Model(&model.AutoReceiptRule{}).
		Where("uuid IN ? AND headquarter_company_uuid = ? AND delete_time = 0", uuids, headquarterUuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
	return errors.WithMessage(err)
}

func (r *autoReceiptRuleRepo) GetByUuid(uuid uint64, headquarterUuid uint64) (model.AutoReceiptRule, error) {
	var rule model.AutoReceiptRule
	err := r.db.Where("uuid = ? AND headquarter_company_uuid = ? AND delete_time = 0", uuid, headquarterUuid).
		First(&rule).Error
	return rule, errors.WithMessage(err)
}

func (r *autoReceiptRuleRepo) GetListByHeadquarter(headquarterUuid uint64, warehouseErpCode string) ([]model.AutoReceiptRule, error) {
	var rules []model.AutoReceiptRule
	db := r.db.Where("headquarter_company_uuid = ? AND delete_time = 0", headquarterUuid)
	if warehouseErpCode != "" {
		db = db.Where("warehouse_erp_code = ?", warehouseErpCode)
	}
	err := db.Order("create_time desc").Find(&rules).Error
	return rules, errors.WithMessage(err)
}

// GetAllEnabled 获取所有启用的规则（定时任务用）
func (r *autoReceiptRuleRepo) GetAllEnabled() ([]model.AutoReceiptRule, error) {
	var rules []model.AutoReceiptRule
	err := r.db.Where("status = 1 AND delete_time = 0").Find(&rules).Error
	return rules, errors.WithMessage(err)
}
