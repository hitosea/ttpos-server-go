package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanyStaffRepo interface {
	WhereUsername(username string) DBOption
	WhereUsernameOrPhone(username, phone string) DBOption
	WhereNotUuid(uuid uint64) DBOption
	GetCompanyStaff(opts ...DBOption) model.CompanyStaff
	CreateCompanyStaff(companyStaff *model.CompanyStaff) *model.CompanyStaff
	UpdateCompanyStaff(uuid uint64, vars map[string]any) error
}

func NewCompanyStaffRepo(db *gorm.DB) ICompanyStaffRepo {
	return NewCompanyStaffRepoImpl(db)
}

type companyStaffRepo struct {
	db *gorm.DB
}

func NewCompanyStaffRepoImpl(db *gorm.DB) ICompanyStaffRepo {
	return &companyStaffRepo{db: db}
}

func (r *companyStaffRepo) WhereUsername(username string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("BINARY username = ? OR phone = ?", username, username)
	}
}

func (r *companyStaffRepo) GetCompanyStaff(opts ...DBOption) model.CompanyStaff {
	var companyStaff model.CompanyStaff
	db := r.db.Model(&model.CompanyStaff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&companyStaff)
	return companyStaff
}

func (r *companyStaffRepo) CreateCompanyStaff(companyStaff *model.CompanyStaff) *model.CompanyStaff {
	r.db.Model(&model.CompanyStaff{}).Create(companyStaff)
	return companyStaff
}

func (r *companyStaffRepo) UpdateCompanyStaff(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.CompanyStaff{}).Where("uuid = ?", uuid).Updates(vars).Error
}

func (r *companyStaffRepo) WhereNotUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid <> ?", uuid)
	}
}

func (r *companyStaffRepo) WhereUsernameOrPhone(username, phone string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("BINARY username = ? OR phone = ?", username, phone)
	}
}
