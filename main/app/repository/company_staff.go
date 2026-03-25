package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICompanyStaffRepo interface {
	WhereUsername(username string) DBOption
	WhereUsernameOrPhone(username, phone string) DBOption
	WhereNotUuid(uuid uint64) DBOption
	WhereUuid(uuid uint64) DBOption
	WhereStaffUuid(staffUuid uint64) DBOption
	WhereCompanyUuid(companyUuid uint64) DBOption
	WithCompany() DBOption // 关联查询商家信息
	GetCompanyStaff(opts ...DBOption) model.CompanyStaff
	GetByUuid(uuid uint64, options ...DBOption) (*model.CompanyStaff, error)
	GetByStaffUuid(staffUuid uint64, options ...DBOption) ([]*model.CompanyStaff, error)
	GetByCompanyUuid(companyUuid uint64, options ...DBOption) ([]*model.CompanyStaff, error)
	GetByCompanyUuids(companyUuids []uint64, options ...DBOption) ([]*model.CompanyStaff, error)
	GetByStaffAndCompany(staffUuid, companyUuid uint64, options ...DBOption) (*model.CompanyStaff, error)
	FindByStaffAndCompany(staffUuid, companyUuid uint64) (*model.CompanyStaff, error) // 含已删除记录
	Create(companyStaff *model.CompanyStaff) error
	CreateCompanyStaff(companyStaff *model.CompanyStaff) *model.CompanyStaff
	UpdateCompanyStaff(uuid uint64, vars map[string]any) error
	UpdateByStaffAndCompany(staffUuid, companyUuid uint64, vars map[string]any) error
	Delete(uuid uint64) error
	DeleteByStaffAndCompany(staffUuid, companyUuid uint64) error
}

func NewCompanyStaffRepo(db *gorm.DB) ICompanyStaffRepo {
	return NewCompanyStaffRepoImpl(db)
}

type companyStaffRepo struct {
	db *gorm.DB
}

func NewCompanyStaffRepoImpl(db *gorm.DB) ICompanyStaffRepo {
	return &companyStaffRepo{db: db}
}

func (r *companyStaffRepo) WhereUsername(username string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("BINARY username = ? OR phone = ?", username, username)
	}
}

func (r *companyStaffRepo) GetCompanyStaff(opts ...DBOption) model.CompanyStaff {
	var companyStaff model.CompanyStaff
	db := r.db.Model(&model.CompanyStaff{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&companyStaff)
	return companyStaff
}

// Create 创建账号-门店关联记录
func (r *companyStaffRepo) Create(companyStaff *model.CompanyStaff) error {
	return r.db.Model(&model.CompanyStaff{}).Create(companyStaff).Error
}

func (r *companyStaffRepo) CreateCompanyStaff(companyStaff *model.CompanyStaff) *model.CompanyStaff {
	r.db.Model(&model.CompanyStaff{}).Create(companyStaff)
	return companyStaff
}

func (r *companyStaffRepo) UpdateCompanyStaff(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.CompanyStaff{}).Where("uuid = ?", uuid).Updates(vars).Error
}

func (r *companyStaffRepo) WhereNotUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid <> ?", uuid)
	}
}

func (r *companyStaffRepo) WhereUsernameOrPhone(username, phone string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("BINARY username = ? OR phone = ?", username, phone)
	}
}

// WhereUuid UUID条件
func (r *companyStaffRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereStaffUuid 员工UUID条件
func (r *companyStaffRepo) WhereStaffUuid(staffUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", staffUuid)
	}
}

// WhereCompanyUuid 门店UUID条件
func (r *companyStaffRepo) WhereCompanyUuid(companyUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("company_uuid = ?", companyUuid)
	}
}

// WithCompany 关联查询商家信息
func (r *companyStaffRepo) WithCompany() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Company")
	}
}

// GetByUuid 根据UUID获取账号-门店关联记录
func (r *companyStaffRepo) GetByUuid(uuid uint64, options ...DBOption) (*model.CompanyStaff, error) {
	var companyStaff model.CompanyStaff
	db := r.db.Model(&model.CompanyStaff{}).Scopes(NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Where("uuid = ?", uuid).First(&companyStaff).Error; err != nil {
		return nil, err
	}
	return &companyStaff, nil
}

// GetByStaffUuid 根据员工UUID获取关联的门店列表
func (r *companyStaffRepo) GetByStaffUuid(staffUuid uint64, options ...DBOption) ([]*model.CompanyStaff, error) {
	var list []*model.CompanyStaff
	db := r.db.Model(&model.CompanyStaff{}).Scopes(NotDeleted).Where("uuid = ?", staffUuid)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByCompanyUuid 根据门店UUID获取关联的员工列表
func (r *companyStaffRepo) GetByCompanyUuid(companyUuid uint64, options ...DBOption) ([]*model.CompanyStaff, error) {
	var list []*model.CompanyStaff
	db := r.db.Model(&model.CompanyStaff{}).Scopes(NotDeleted).Where("company_uuid = ?", companyUuid)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByCompanyUuids 根据门店UUID列表获取关联的员工列表
func (r *companyStaffRepo) GetByCompanyUuids(companyUuids []uint64, options ...DBOption) ([]*model.CompanyStaff, error) {
	var list []*model.CompanyStaff
	db := r.db.Model(&model.CompanyStaff{}).Scopes(NotDeleted).Where("company_uuid IN ?", companyUuids)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByStaffAndCompany 根据员工和门店UUID获取关联记录（用于验证权限）
func (r *companyStaffRepo) GetByStaffAndCompany(staffUuid, companyUuid uint64, options ...DBOption) (*model.CompanyStaff, error) {
	var companyStaff model.CompanyStaff
	db := r.db.Model(&model.CompanyStaff{}).
		Scopes(NotDeleted).
		Where("uuid = ?", staffUuid).
		Where("company_uuid = ?", companyUuid)

	for _, option := range options {
		db = option(db)
	}

	if err := db.First(&companyStaff).Error; err != nil {
		return nil, err
	}
	return &companyStaff, nil
}

// FindByStaffAndCompany 根据员工和门店UUID获取关联记录（含已删除记录，用于恢复场景）
func (r *companyStaffRepo) FindByStaffAndCompany(staffUuid, companyUuid uint64) (*model.CompanyStaff, error) {
	var companyStaff model.CompanyStaff
	if err := r.db.Model(&model.CompanyStaff{}).
		Where("uuid = ?", staffUuid).
		Where("company_uuid = ?", companyUuid).
		First(&companyStaff).Error; err != nil {
		return nil, err
	}
	return &companyStaff, nil
}

// UpdateByStaffAndCompany 根据员工和门店UUID更新关联记录
func (r *companyStaffRepo) UpdateByStaffAndCompany(staffUuid, companyUuid uint64, vars map[string]any) error {
	return r.db.Model(&model.CompanyStaff{}).
		Where("uuid = ?", staffUuid).
		Where("company_uuid = ?", companyUuid).
		Updates(vars).Error
}

// Delete 删除账号-门店关联记录（软删除）
func (r *companyStaffRepo) Delete(uuid uint64) error {
	return r.db.Model(&model.CompanyStaff{}).
		Where("uuid = ?", uuid).
		Update("delete_time", time.Now().Unix()).Error
}

// DeleteByStaffAndCompany 根据员工和门店UUID删除关联记录（软删除）
func (r *companyStaffRepo) DeleteByStaffAndCompany(staffUuid, companyUuid uint64) error {
	return r.db.Model(&model.CompanyStaff{}).
		Where("uuid = ?", staffUuid).
		Where("company_uuid = ?", companyUuid).
		Update("delete_time", time.Now().Unix()).Error
}
