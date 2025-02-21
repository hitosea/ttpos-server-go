package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanySettingRepo interface {
	Get() model.CompanySetting
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
