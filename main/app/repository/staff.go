package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRepo interface {
	GetByIdAndCompanyId(Id, companyId uint, withs ...With) model.Staff
	GetById(companyId uint, id uint, withs ...With) model.Staff
	WithCompanySetting() With
	WithCompany() With
	OfflineGetByUsername(username string, withs ...With) model.Staff
	//CreateStaff(staff model.Staff) error
	GetCurrentCashier(companyId uint, bindKey string) model.Staff
	Update(companyId uint, id uint, vars map[string]any) error
}

func NewStaffRepo(db *gorm.DB) IStaffRepo {
	return NewStaffRepoImpl(db)
}

type StaffRepo struct {
	db *gorm.DB
}

func NewStaffRepoImpl(db *gorm.DB) *StaffRepo {
	return &StaffRepo{db: db}
}

func (r *StaffRepo) GetByIdAndCompanyId(Id, companyId uint, withs ...With) model.Staff {
	var user model.Staff
	r.handleWiths(r.db, withs).First(&user, Id)
	return user
}

func (r *StaffRepo) GetById(companyId uint, id uint, withs ...With) model.Staff {
	var user model.Staff
	r.handleWiths(r.db, withs).Debug().First(&user, id)
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
	r.handleWiths(r.db, withs).Where("username = ?", username).Debug().First(&user)
	return user
}

func (r *StaffRepo) GetCurrentCashier(companyId uint, bindKey string) model.Staff {
	var user model.Staff
	r.db.Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&user)
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
	return r.db.Model(&model.Staff{}).Where("id = ?", id).Updates(vars).Error
}
