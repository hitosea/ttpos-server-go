package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanyStaffRepo interface {
	WhereUsername(username string) DBOption
	GetCompanyStaff(opts ...DBOption) model.CompanyStaff
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

func (r *CompanyStaffRepo) WhereUsername(username string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("BINARY username = ? OR phone = ?", username, username)
	}
}

func (r *CompanyStaffRepo) GetCompanyStaff(opts ...DBOption) model.CompanyStaff {
	var companyStaff model.CompanyStaff
	db := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Debug().First(&companyStaff)
	return companyStaff
}
