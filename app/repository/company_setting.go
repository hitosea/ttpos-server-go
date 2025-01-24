package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type CompanySettingRepository struct {
	dbm *database.DBManager
}

func NewCompanySettingRepository(dbm *database.DBManager) *CompanySettingRepository {
	return &CompanySettingRepository{dbm: dbm}
}

func (r *CompanySettingRepository) GetById(id uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.dbm.GetDB(id).Model(&model.CompanySetting{}).First(&companySetting, id)
	return companySetting
}

func (r *CompanySettingRepository) GetByCompanyId(companyId uint) model.CompanySetting {
	var companySetting model.CompanySetting
	r.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("company_id = ?", companyId).First(&companySetting)
	return companySetting
}

func (r *CompanySettingRepository) Update(companySetting model.CompanySetting) error {
	if err := r.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("company_id = ?", companySetting.CompanyId).Updates(companySetting).Error; err != nil {
		return err
	}
	return nil
}

func (r *CompanySettingRepository) Delete(companyId uint) error {
	return r.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("company_id = ?", companyId).Update("delete_time", time.Now().Unix()).Error
}
