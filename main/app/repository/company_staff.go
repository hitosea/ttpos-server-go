package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanyStaffRepo interface {
	WithCompanySetting() With
	WithCompany() With
	GetByUsername(username string, withs ...With) model.CompanyStaff
}

func NewCompanyStaffRepo(db *gorm.DB) ICompanyStaffRepo {
	return NewCompanyStaffRepoImpl(db)
}

type CompanyStaffRepo struct {
	db *gorm.DB
}

func NewCompanyStaffRepoImpl(db *gorm.DB) *CompanyStaffRepo {
	return &CompanyStaffRepo{db: db}
}

func (r *CompanyStaffRepo) WithCompany() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *CompanyStaffRepo) WithCompanySetting() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("CompanySetting")
	}
}

func (r *CompanyStaffRepo) GetByUsername(username string, withs ...With) model.CompanyStaff {
	var companyStaff model.CompanyStaff
	handleWiths(r.db, withs).Where("BINARY username = ? OR phone = ?", username, username).Debug().First(&companyStaff)
	return companyStaff
}
