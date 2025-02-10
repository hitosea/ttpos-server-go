package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanySettingRepo interface {
	GetById(companyId uint64) model.CompanySetting
	GetByCompanyId(companyId uint64) model.CompanySetting
	GetByCompanyIdFromCompanyDB() model.CompanySetting
	Update(companySetting model.CompanySetting) error
	Delete(companyId uint64) error
}

func NewCompanySettingRepo(db *gorm.DB) ICompanySettingRepo {
	return NewCompanySettingRepoImpl(db)
}

type CompanySettingRepo struct {
	db *gorm.DB
}

func NewCompanySettingRepoImpl(db *gorm.DB) *CompanySettingRepo {
	return &CompanySettingRepo{db: db}
}

func (r *CompanySettingRepo) GetById(companyId uint64) model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.Model(&model.CompanySetting{}).First(&companySetting, companyId)
	return companySetting
}

func (r *CompanySettingRepo) GetByCompanyId(companyId uint64) model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.Model(&model.CompanySetting{}).Where("company_id = ?", companyId).First(&companySetting)
	return companySetting
}

func (r *CompanySettingRepo) GetByCompanyIdFromCompanyDB() model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.First(&companySetting)
	return companySetting
}

func (r *CompanySettingRepo) Update(companySetting model.CompanySetting) error {
	if err := r.db.Model(&model.CompanySetting{}).Where("company_id = ?", companySetting.CompanyUuid).Updates(companySetting).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompanySettingRepo) Delete(companyId uint64) error {
	return r.db.Model(&model.CompanySetting{}).Where("company_id = ?", companyId).Update("delete_time", time.Now().Unix()).Error
}
