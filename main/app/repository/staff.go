package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRepo interface {
	WithCompany() With
	WithCompanySetting() With
	WithDevice(source string) With

	GetByUuid(uuid uint64, withs ...With) model.Staff                            // 根据Uuid查询员工
	GetByUsername(username string, withs ...With) model.Staff                    // 根据用户名查询员工
	GetByDeviceId(bindKey string) model.Staff                                    // 根据员工绑定的设备ID查询员工
	GetByUuidAndDeviceId(uuid uint64, bindKey string, withs ...With) model.Staff // 根据员工Uuid和绑定的设备ID查询员工
	GetOnlineCashiers(withs ...With) []model.Staff                               // 获取在线收银机

	CreateStaff(staff model.Staff) error           // 创建员工
	Update(uuid uint64, vars map[string]any) error // 更新员工
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
	handleWiths(r.db, withs).Where("uuid = ?", uuid).Debug().First(&staff)
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

func (r *StaffRepo) WithDevice(source string) With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Device", "source = ?", source)
	}
}

func (r *StaffRepo) GetByUsername(username string, withs ...With) model.Staff {
	var staff model.Staff
	handleWiths(r.db, withs).Where("BINARY username = ? OR phone = ?", username, username).Debug().First(&staff)
	return staff
}

func (r *StaffRepo) GetByDeviceId(bindKey string) model.Staff {
	var staff model.Staff
	r.db.Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&staff)
	return staff
}

func (r *StaffRepo) GetByUuidAndDeviceId(uuid uint64, bindKey string, withs ...With) model.Staff {
	var staff model.Staff
	handleWiths(r.db, withs).Where("uuid = ? AND bind_key = ?", uuid, bindKey).Debug().First(&staff)
	return staff
}
func (r *StaffRepo) GetOnlineCashiers(withs ...With) []model.Staff {
	var staff []model.Staff
	handleWiths(r.db, withs).Where("cashier_online = 1").Debug().Find(&staff)
	return staff
}

func (r *StaffRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.Staff{}).Where("uuid = ?", uuid).Updates(vars).Error
}
