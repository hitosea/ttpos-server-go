package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanySettingRepo interface {
	WhereErpnextCompanyAbbr(erpnextCompanyAbbr string) DBOption
	WhereErpnextCompanyAbbrNotEmpty() DBOption
	WhereSiteCode(siteCode string) DBOption

	GetOne(opts ...DBOption) (model.CompanySetting, error)
	Get() model.CompanySetting
	GetAllByHeadquarterUuid(headquarterUuid uint64) ([]model.CompanySetting, error) // 获取总部下所有公司的设置
	UpdateSmsQuota(companyUuid uint64, quota int) error                             // 扣减公司的短信余额

	GetErpnextCompanyAbbrUuidMap(opts ...DBOption) (map[string]uint64, error)
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

func (r *companySettingRepo) GetAllByHeadquarterUuid(headquarterUuid uint64) ([]model.CompanySetting, error) {
	var companySettings []model.CompanySetting
	err := r.db.Model(&model.CompanySetting{}).Scopes(NotDeleted).Where("headquarter_uuid = ? or (company_uuid = ? and headquarter_uuid = 0)", headquarterUuid, headquarterUuid).Find(&companySettings).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return companySettings, nil
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

func (r *companySettingRepo) GetErpnextCompanyAbbrUuidMap(opts ...DBOption) (map[string]uint64, error) {
	var companySettings []model.CompanySetting
	db := r.db.Model(&model.CompanySetting{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&companySettings).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	erpCompanyAbbrUuidMap := make(map[string]uint64)
	for _, companySetting := range companySettings {
		erpCompanyAbbrUuidMap[companySetting.ErpnextCompanyAbbr] = companySetting.CompanyUuid
	}
	return erpCompanyAbbrUuidMap, nil
}

func (r *companySettingRepo) WhereErpnextCompanyAbbrNotEmpty() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("erpnext_company_abbr != ''")
	}
}

func (r *companySettingRepo) WhereSiteCode(siteCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("erpnext_site_code = ?", siteCode)
	}
}
