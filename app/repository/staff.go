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

func (r *StaffRepository) GetById(companyId uint, id uint, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(companyId)
	r.handleWiths(db, withs).Debug().First(&user, id)
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

func (r *StaffRepository) OfflineGetByUsername(username string, withs ...With) model.Staff {
	var user model.Staff
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).Where("username = ?", username).Debug().First(&user)
	return user
}

func (r *StaffRepository) GetCurrentCashier(companyId uint, bindKey string) model.Staff {
	var user model.Staff
	r.dbm.GetDB(companyId).Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&user)
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

func (r *StaffRepository) Update(companyId uint, id uint, vars map[string]any) error {
	return r.dbm.GetDB(companyId).Model(&model.Staff{}).Where("id = ?", id).Updates(vars).Error
}
