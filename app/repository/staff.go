package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

type StaffRepository struct {
	dbm *database.DBManager
}

func NewStaffRepository(dbm *database.DBManager) *StaffRepository {
	return &StaffRepository{dbm: dbm}
}
func (r *StaffRepository) GetByIdAndCompanyId(Id, companyId uint, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(companyId)
	r.handleWiths(db, withs).First(&user, Id)
	return user
}

func (r *StaffRepository) GetById(Id, companyId uint, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(companyId)
	r.handleWiths(db, withs).Debug().First(&user, Id)
	return user
}

func (r *StaffRepository) WithCompanySetting() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company.CompanySetting")
	}
}

func (r *StaffRepository) WithCompany() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *StaffRepository) GetByUsername(username string, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).Where("username = ?", username).Debug().First(&user)
	return user
}

func (r *StaffRepository) GetCurrentCashier(bindKey string) model.Staff {
	var user model.Staff
	r.dbm.GetDB(constant.DefaultDB).Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&user)
	return user
}

func (r *StaffRepository) handleWiths(db *gorm.DB, withs []With) *gorm.DB {
	if len(withs) == 0 {
		return db
	}
	for _, with := range withs {
		db = with(db)
	}
	return db
}
