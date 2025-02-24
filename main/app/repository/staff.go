package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRepo interface {
	WithCompany() DBOption             // 关联集团
	WithCompanySetting() DBOption      // 关联集团设置
	WithDevice(source string) DBOption // 关联设备

	WhereUuid(uuid uint64) DBOption         // Uuid 条件
	WhereUsername(username string) DBOption // 用户名条件
	WhereCashierOnline() DBOption           // 收银机在线条件
	WhereDeviceId(bindKey string) DBOption  // 设备ID条件

	GetStaff(opts ...DBOption) model.Staff    // 查询员工
	GetStaffs(opts ...DBOption) []model.Staff // 查询员工

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
	return r.db.Model(&model.Staff{}).Create(&staff).Error
}

func (r *StaffRepo) GetStaff(opts ...DBOption) model.Staff {
	var staff model.Staff
	db := r.db.Model(&model.Staff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Debug().First(&staff)
	return staff
}

func (r *StaffRepo) GetStaffs(opts ...DBOption) []model.Staff {
	var staffs []model.Staff
	db := r.db.Model(&model.Staff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Debug().Find(&staffs)
	return staffs
}

func (r *StaffRepo) WithCompanySetting() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company.CompanySetting")
	}
}

func (r *StaffRepo) WithCompany() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

func (r *StaffRepo) WithDevice(source string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Device", "source = ?", source)
	}
}

func (r *StaffRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *StaffRepo) WhereUsername(username string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("BINARY username = ? OR phone = ?", username, username)
	}
}

func (r *StaffRepo) WhereCashierOnlineDeviceId(bindKey string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("bind_key = ? AND cashier_online = 1", bindKey)
	}
}

func (r *StaffRepo) WhereCashierOnline() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("cashier_online = 1")
	}
}

func (r *StaffRepo) WhereDeviceId(bindKey string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("bind_key = ?", bindKey)
	}
}

func (r *StaffRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.Staff{}).Where("uuid = ?", uuid).Updates(vars).Error
}
