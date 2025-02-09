package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanySettingRepo interface {
	GetById(companyId uint) model.CompanySetting
	GetByCompanyId(companyId uint) model.CompanySetting
	GetByCompanyIdFromCompanyDB(companyId uint) model.CompanySetting
	Update(companySetting model.CompanySetting) error
	Delete(companyId uint) error
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

func (r *CompanySettingRepo) GetById(companyId uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.Model(&model.CompanySetting{}).First(&companySetting, companyId)
	return companySetting
}

func (r *CompanySettingRepo) GetByCompanyId(companyId uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.Model(&model.CompanySetting{}).Where("company_id = ?", companyId).First(&companySetting)
	return companySetting
}

func (r *CompanySettingRepo) GetByCompanyIdFromCompanyDB(companyId uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.First(&companySetting)
	return companySetting
}

func (r *CompanySettingRepo) Update(companySetting model.CompanySetting) error {
	if err := r.db.Model(&model.CompanySetting{}).Where("company_id = ?", companySetting.CompanyId).Updates(companySetting).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompanySettingRepo) Delete(companyId uint) error {
	return r.db.Model(&model.CompanySetting{}).Where("company_id = ?", companyId).Update("delete_time", time.Now().Unix()).Error
}
