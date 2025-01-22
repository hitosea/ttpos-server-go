package repository

import (
	"jjjshop-server-go/app/constant"
	"jjjshop-server-go/app/model"
	"jjjshop-server-go/pkg/database"

	"gorm.io/gorm"
)

type StaffRepository struct {
	dbm *database.DBManager
}

func NewStaffRepository(dbm *database.DBManager) *StaffRepository {
	return &StaffRepository{dbm: dbm}
}
func (r *StaffRepository) GetByIdAndCompanyId(Id, companyId uint, withs ...Where) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(companyId)
	r.handleWiths(db, withs).First(&user, Id)
	return user
}

func (r *StaffRepository) GetById(Id, companyId uint, withs ...Where) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(companyId)
	r.handleWiths(db, withs).First(&user, Id)
	return user
}

func (r *StaffRepository) WithCompanySetting() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("CompanySetting")
	}
}

func (r *StaffRepository) WithCompany() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *StaffRepository) GetByUsername(username string, withs ...Where) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).Where("user_name = ?", username).Debug().First(&user)
	return user
}

func (r *StaffRepository) GetCurrentCashier(bindKey string) model.Staff {
	var user model.Staff
	r.dbm.GetDB(constant.DefaultDB).Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&user)
	return user
}

func (r *StaffRepository) handleWiths(db *gorm.DB, withs []Where) *gorm.DB {
	if len(withs) == 0 {
		return db
	}
	for _, with := range withs {
		db = with(db)
	}
	return db
}
