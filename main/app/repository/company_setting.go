package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type ICompanySettingRepo interface {
	GetById(companyId uint) model.CompanySetting
	GetByCompanyId(companyId uint) model.CompanySetting
	GetByCompanyIdFromCompanyDB(companyId uint) model.CompanySetting
	Update(companySetting model.CompanySetting) error
	Delete(companyId uint) error
}

func NewCompanySettingRepo(dbm *database.DBManager) ICompanySettingRepo {
	return NewCompanySettingRepoImpl(dbm)
}

type CompanySettingRepo struct {
	dbm *database.DBManager
}

func NewCompanySettingRepoImpl(dbm *database.DBManager) *CompanySettingRepo {
	return &CompanySettingRepo{dbm: dbm}
}

func (r *CompanySettingRepo) GetById(companyId uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.dbm.GetDB(companyId).Model(&model.CompanySetting{}).First(&companySetting, companyId)
	return companySetting
}

func (r *CompanySettingRepo) GetByCompanyId(companyId uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("company_id = ?", companyId).First(&companySetting)
	return companySetting
}

func (r *CompanySettingRepo) GetByCompanyIdFromCompanyDB(companyId uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.dbm.GetDB(companyId).First(&companySetting)
	return companySetting
}

func (r *CompanySettingRepo) Update(companySetting model.CompanySetting) error {
	if err := r.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("company_id = ?", companySetting.CompanyId).Updates(companySetting).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompanySettingRepo) Delete(companyId uint) error {
	return r.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("company_id = ?", companyId).Update("delete_time", time.Now().Unix()).Error
}
