package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanyStaffRepo interface {
	GetById(Id uint, withs ...Where) model.CompanyStaff
	WithCompany() Where
	GetByUsername(username string, withs ...Where) model.CompanyStaff
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

func (r *CompanyStaffRepo) GetById(Id uint, withs ...Where) model.CompanyStaff {
	var user model.CompanyStaff
	r.handleWiths(r.db, withs).First(&user, Id)
	return user
}

func (r *CompanyStaffRepo) WithCompany() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *CompanyStaffRepo) GetByUsername(username string, withs ...Where) model.CompanyStaff {
	var user model.CompanyStaff
	r.handleWiths(r.db, withs).Where("username = ?", username).Debug().First(&user)
	return user
}

func (r *CompanyStaffRepo) handleWiths(db *gorm.DB, withs []Where) *gorm.DB {
	if len(withs) == 0 {
		return db
	}
	for _, with := range withs {
		db = with(db)
	}
	return db
}
