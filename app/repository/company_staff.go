package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type CompanyStaffRepository struct {
	dbm *database.DBManager
}

func NewCompanyStaffRepository(db *database.DBManager) *CompanyStaffRepository {
	return &CompanyStaffRepository{dbm: db}
}

func (r *CompanyStaffRepository) GetById(Id uint, withs ...Where) model.CompanyStaff {
	var user model.CompanyStaff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).First(&user, Id)
	return user
}

func (r *CompanyStaffRepository) WithCompany() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *CompanyStaffRepository) GetByUsername(username string, withs ...Where) model.CompanyStaff {
	var user model.CompanyStaff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).Where("user_name = ?", username).Debug().First(&user)
	return user
}

func (r *CompanyStaffRepository) handleWiths(db *gorm.DB, withs []Where) *gorm.DB {
	if len(withs) == 0 {
		return db
	}
	for _, with := range withs {
		db = with(db)
	}
	return db
}
