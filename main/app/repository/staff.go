package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRepo interface {
	WithCompany() With
	WithCompanySetting() With

	GetByUuid(uuid uint64, withs ...With) model.Staff
	GetByUsername(username string, withs ...With) model.Staff
	GetByDeviceId(bindKey string) model.Staff
	GetByUuidAndDeviceId(uuid uint64, bindKey string, withs ...With) model.Staff
	CreateStaff(staff model.Staff) error
	Update(uuid uint64, vars map[string]any) error
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

func (r *StaffRepo) CreateStaff(staff model.Staff) error {
	return r.db.Create(&staff).Error
}

func (r *StaffRepo) GetByUuid(uuid uint64, withs ...With) model.Staff {
	var staff model.Staff
	r.handleWiths(r.db, withs).Where("uuid = ?", uuid).Debug().First(&staff)
	return staff
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

func (r *StaffRepo) GetByUsername(username string, withs ...With) model.Staff {
	var staff model.Staff
	r.handleWiths(r.db, withs).Where("BINARY username = ? OR phone = ?", username, username).Debug().First(&staff)
	return staff
}

func (r *StaffRepo) GetByDeviceId(bindKey string) model.Staff {
	var staff model.Staff
	r.db.Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&staff)
	return staff
}

func (r *StaffRepo) GetByUuidAndDeviceId(uuid uint64, bindKey string, withs ...With) model.Staff {
	var staff model.Staff
	r.handleWiths(r.db, withs).Where("uuid = ? AND bind_key = ?", uuid, bindKey).Debug().First(&staff)
	return staff
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

func (r *StaffRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.Staff{}).Where("uuid = ?", uuid).Updates(vars).Error
}
