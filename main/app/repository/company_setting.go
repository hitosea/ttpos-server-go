package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanySettingRepo interface {
	WhereErpnextCompanyAbbr(erpnextCompanyAbbr string) DBOption

	GetOne(opts ...DBOption) (model.CompanySetting, error)
	Get() model.CompanySetting
	UpdateSmsQuota(companyUuid uint64, quota int) error // 扣减公司的短信余额
}

func NewCompanySettingRepo(db *gorm.DB) ICompanySettingRepo {
	return NewCompanySettingRepoImpl(db)
}

type companySettingRepo struct {
	db *gorm.DB
}

func NewCompanySettingRepoImpl(db *gorm.DB) ICompanySettingRepo {
	return &companySettingRepo{db: db}
}

func (r *companySettingRepo) Get() model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.Model(&model.CompanySetting{}).First(&companySetting)
	return companySetting
}

func (r *companySettingRepo) UpdateSmsQuota(companyUuid uint64, quota int) error {
	if err := r.db.Model(&model.CompanySetting{}).Where("company_uuid = ?", companyUuid).Update("sms_quota", gorm.Expr("sms_quota - ?", quota)).Error; err != nil {
		return errors.WithMessage(err, "failed to update SMS quota")
	}
	return nil
}

func (r *companySettingRepo) WhereErpnextCompanyAbbr(erpCompanyAbbr string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("erpnext_company_abbr = ?", erpCompanyAbbr)
	}
}

func (r *companySettingRepo) GetOne(opts ...DBOption) (model.CompanySetting, error) {
	var companySetting model.CompanySetting
	db := r.db.Model(&model.CompanySetting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Take(&companySetting).Error
	return companySetting, err
}
