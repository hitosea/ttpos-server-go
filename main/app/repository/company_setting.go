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

type CompanySettingRepo struct {
	db *gorm.DB
}

func NewCompanySettingRepoImpl(db *gorm.DB) *CompanySettingRepo {
	return &CompanySettingRepo{db: db}
}

func (r *CompanySettingRepo) Get() model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.First(&companySetting)
	return companySetting
}
