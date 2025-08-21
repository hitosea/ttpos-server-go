package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRepo interface {
	WithCompany() DBOption             // 关联集团
	WithCompanySetting() DBOption      // 关联集团设置
	WithDevice(source string) DBOption // 关联设备
	WithRoles() DBOption               // 关联角色

	WhereUuid(uuid uint64) DBOption         // Uuid 条件
	WhereUsername(username string) DBOption // 用户名条件
	WhereCashierOnline() DBOption           // 收银机在线条件
	WhereDeviceId(bindKey string) DBOption  // 设备ID条件
	WhereRoleUuid(roleUuid uint64) DBOption // 角色ID条件
	WhereIsSuper(isSuper int) DBOption      // 是否超级管理员条件
	WhereRealName(realName string) DBOption // 姓名条件
	WhereNotUuid(uuid uint64) DBOption      // 排除UUID条件

	GetStaff(opts ...DBOption) (model.Staff, error) // 查询员工
	GetStaffs(opts ...DBOption) []model.Staff       // 查询员工

	PaginateGetStaffs(pageNo, pageSize int, opts ...DBOption) ([]model.Staff, int64, error) // 分页查询员工

	CreateStaff(staff model.Staff) error                         // 创建员工
	Update(uuid uint64, vars map[string]any) error               // 更新员工
	UpdateStaffRoles(staffUuid uint64, roleUuids []uint64) error // 更新管理员角色
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

func (r *StaffRepo) GetStaff(opts ...DBOption) (model.Staff, error) {
	var staff model.Staff
	db := r.db.Model(&model.Staff{}).Scopes(NotDeleted).Debug()
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Take(&staff).Error
	return staff, err
}

func (r *StaffRepo) GetStaffs(opts ...DBOption) []model.Staff {
	var staffs []model.Staff
	db := r.db.Model(&model.Staff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Find(&staffs)
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

func (r *StaffRepo) PaginateGetStaffs(pageNo, pageSize int, opts ...DBOption) ([]model.Staff, int64, error) {
	var staffs []model.Staff
	var total int64
	db := r.db.Model(&model.Staff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Count(&total)
	db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("create_time desc").Find(&staffs)
	return staffs, total, nil
}

func (r *StaffRepo) WithRoles() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Roles")
	}
}

func (r *StaffRepo) UpdateStaffRoles(staffUuid uint64, roleUuids []uint64) error {
	// 删除此管理员和角色的关联
	r.db.Model(&model.StaffRole{}).Where("staff_uuid = ?", staffUuid).Delete(&model.StaffRole{})

	// 新增此管理员和角色的关联
	for _, roleUuid := range roleUuids {
		r.db.Model(&model.StaffRole{}).Create(&model.StaffRole{
			StaffUuid: int64(staffUuid),
			RoleUuid:  int64(roleUuid),
		})
	}
	return nil
}

func (r *StaffRepo) WhereRoleUuid(roleUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("role_uuid = ?", roleUuid)
	}
}

func (r *StaffRepo) WhereIsSuper(isSuper int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_super = ?", isSuper)
	}
}

func (r *StaffRepo) WhereRealName(realName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("real_name = ?", realName)
	}
}

func (r *StaffRepo) WhereNotUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid <> ?", uuid)
	}
}
