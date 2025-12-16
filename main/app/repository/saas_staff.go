package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ISaasStaffRepo saas数据库统一账号数据访问接口
type ISaasStaffRepo interface {
	// 选项方法
	WhereUuid(uuid uint64) DBOption
	WhereUuids(uuids []uint64) DBOption
	WhereEmail(email string) DBOption
	WherePhone(phone string) DBOption
	WhereEmailOrPhone(email, phone string) DBOption // 姓名、邮箱、手机号条件
	WhereNotUuid(uuid uint64) DBOption

	// CRUD 方法
	Create(staff *model.SaasStaff) error
	Update(uuid uint64, vars map[string]any) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.SaasStaff, error)
	GetByEmail(email string, options ...DBOption) (*model.SaasStaff, error)
	GetByPhone(phone string, options ...DBOption) (*model.SaasStaff, error)
	GetByEmailOrPhone(keyword string, options ...DBOption) (*model.SaasStaff, error)
	PaginateGetStaffs(pageNo, pageSize int, opts ...DBOption) ([]model.SaasStaff, int64, error) // 分页查询统一账号

	GetSaasStaff(opts ...DBOption) model.SaasStaff

	// 唯一性检查方法
	CheckEmailExists(email string, excludeUuid uint64) (bool, error)
	CheckPhoneExists(phone string, excludeUuid uint64) (bool, error)
}

func NewSaasStaffRepo(db *gorm.DB) ISaasStaffRepo {
	return NewSaasStaffRepoImpl(db)
}

type saasStaffRepo struct {
	db *gorm.DB // saas 数据库连接
}

func NewSaasStaffRepoImpl(db *gorm.DB) ISaasStaffRepo {
	return &saasStaffRepo{db: db}
}

// Create 创建统一账号
func (r *saasStaffRepo) Create(staff *model.SaasStaff) error {
	return r.db.Model(&model.SaasStaff{}).Create(staff).Error
}

// Update 更新统一账号
func (r *saasStaffRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.SaasStaff{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// GetByUuid 根据UUID获取统一账号
func (r *saasStaffRepo) GetByUuid(uuid uint64, options ...DBOption) (*model.SaasStaff, error) {
	var staff model.SaasStaff
	db := r.db.Model(&model.SaasStaff{}).Scopes(NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Where("uuid = ?", uuid).First(&staff).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByEmail 根据邮箱获取统一账号
func (r *saasStaffRepo) GetByEmail(email string, options ...DBOption) (*model.SaasStaff, error) {
	var staff model.SaasStaff
	db := r.db.Model(&model.SaasStaff{}).Scopes(NotDeleted).Where("email = ?", email)

	for _, option := range options {
		db = option(db)
	}

	if err := db.First(&staff).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByPhone 根据手机号获取统一账号
func (r *saasStaffRepo) GetByPhone(phone string, options ...DBOption) (*model.SaasStaff, error) {
	var staff model.SaasStaff
	db := r.db.Model(&model.SaasStaff{}).Scopes(NotDeleted).Where("phone = ?", phone)

	for _, option := range options {
		db = option(db)
	}

	if err := db.First(&staff).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByEmailOrPhone 根据邮箱或手机号获取统一账号
func (r *saasStaffRepo) GetByEmailOrPhone(keyword string, options ...DBOption) (*model.SaasStaff, error) {
	var staff model.SaasStaff
	db := r.db.Model(&model.SaasStaff{}).Scopes(NotDeleted).Where("email = ? OR phone = ?", keyword, keyword)
	for _, option := range options {
		db = option(db)
	}
	if err := db.First(&staff).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

// CheckEmailExists 检查邮箱是否存在
func (r *saasStaffRepo) CheckEmailExists(email string, excludeUuid uint64) (bool, error) {
	var count int64
	db := r.db.Model(&model.SaasStaff{}).Scopes(NotDeleted).Where("email = ?", email)
	if excludeUuid > 0 {
		db = db.Where("uuid != ?", excludeUuid)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckPhoneExists 检查手机号是否存在（排除空字符串）
func (r *saasStaffRepo) CheckPhoneExists(phone string, excludeUuid uint64) (bool, error) {
	// 重要：手机号唯一性验证需要排除空字符串
	// 只有非空手机号才需要验证唯一性
	if phone == "" {
		return false, nil // 空字符串不验证唯一性
	}

	var count int64
	db := r.db.Model(&model.SaasStaff{}).
		Scopes(NotDeleted).
		Where("phone = ?", phone).
		Where("phone != ?", "") // 排除空字符串
	if excludeUuid > 0 {
		db = db.Where("uuid != ?", excludeUuid)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// WhereUuid UUID条件
func (r *saasStaffRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereEmail 邮箱条件
func (r *saasStaffRepo) WhereEmail(email string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("email = ?", email)
	}
}

// WherePhone 手机号条件
func (r *saasStaffRepo) WherePhone(phone string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("phone = ?", phone)
	}
}

// WhereUuids UUID列表条件
func (r *saasStaffRepo) WhereUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(uuids) == 0 {
			return db.Where("1 = 0") // 空列表返回空结果
		}
		return db.Where("uuid IN (?)", uuids)
	}
}

// PaginateGetStaffs 分页查询统一账号
func (r *saasStaffRepo) PaginateGetStaffs(pageNo, pageSize int, opts ...DBOption) ([]model.SaasStaff, int64, error) {
	var staffs []model.SaasStaff
	var total int64
	db := r.db.Model(&model.SaasStaff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Count(&total)
	db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("create_time desc").Find(&staffs)
	return staffs, total, nil
}

func (r *saasStaffRepo) WhereEmailOrPhone(email, phone string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("BINARY email = ? OR phone = ?", email, phone)
	}
}

func (r *saasStaffRepo) WhereNotUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid != ?", uuid)
	}
}

func (r *saasStaffRepo) GetSaasStaff(opts ...DBOption) model.SaasStaff {
	var staff model.SaasStaff
	db := r.db.Model(&model.SaasStaff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&staff)
	return staff
}
