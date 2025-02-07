package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

type IStaffRepo interface {
	GetByIdAndCompanyId(Id, companyId uint, withs ...With) model.Staff
	GetById(companyId uint, id uint, withs ...With) model.Staff
	WithCompanySetting() With
	WithCompany() With
	OfflineGetByUsername(username string, withs ...With) model.Staff
	GetCurrentCashier(companyId uint, bindKey string) model.Staff
	Update(companyId uint, id uint, vars map[string]any) error
}

func NewStaffRepo(dbm *database.DBManager) IStaffRepo {
	return NewStaffRepoImpl(dbm)
}

type StaffRepo struct {
	dbm *database.DBManager
}

func NewStaffRepoImpl(dbm *database.DBManager) *StaffRepo {
	return &StaffRepo{dbm: dbm}
}

func (r *StaffRepo) GetByIdAndCompanyId(Id, companyId uint, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(companyId)
	r.handleWiths(db, withs).First(&user, Id)
	return user
}

func (r *StaffRepo) GetById(companyId uint, id uint, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(companyId)
	r.handleWiths(db, withs).Debug().First(&user, id)
	return user
}

func (r *StaffRepo) WithCompanySetting() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company.CompanySetting")
	}
}

func (r *StaffRepo) WithCompany() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *StaffRepo) OfflineGetByUsername(username string, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).Where("username = ?", username).Debug().First(&user)
	return user
}

func (r *StaffRepo) GetCurrentCashier(companyId uint, bindKey string) model.Staff {
	var user model.Staff
	r.dbm.GetDB(companyId).Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&user)
	return user
}

func (r *StaffRepo) handleWiths(db *gorm.DB, withs []With) *gorm.DB {
	if len(withs) == 0 {
		return db
	}
	for _, with := range withs {
		db = with(db)
	}
	return db
}

func (r *StaffRepo) Update(companyId uint, id uint, vars map[string]any) error {
	return r.dbm.GetDB(companyId).Model(&model.Staff{}).Where("id = ?", id).Updates(vars).Error
}
