package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type ICompanyStaffRepo interface {
	GetById(uint, ...Where) model.CompanyStaff
	WithCompany() Where
	GetByUsername(string, ...Where) model.CompanyStaff
}

func NewCompanyStaffRepo(db *database.DBManager) ICompanyStaffRepo {
	return NewCompanyStaffRepoImpl(db)
}

type CompanyStaffRepo struct {
	dbm *database.DBManager
}

func NewCompanyStaffRepoImpl(db *database.DBManager) *CompanyStaffRepo {
	return &CompanyStaffRepo{dbm: db}
}

func (r *CompanyStaffRepo) GetById(Id uint, withs ...Where) model.CompanyStaff {
	var user model.CompanyStaff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).First(&user, Id)
	return user
}

func (r *CompanyStaffRepo) WithCompany() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *CompanyStaffRepo) GetByUsername(username string, withs ...Where) model.CompanyStaff {
	var user model.CompanyStaff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).Where("user_name = ?", username).Debug().First(&user)
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
