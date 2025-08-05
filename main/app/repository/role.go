package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IRoleRepo 角色
type IRoleRepo interface {
	WithAccesses() DBOption // 关联权限

	// 根据ID列表查询角色
	WhereUuids(uuids []uint64) DBOption
	WhereUuid(uuid uint64) DBOption
	WhereName(name string) DBOption
	GetRoleList(opts ...DBOption) ([]model.Role, error)
	PaginateGetRoleList(pageNo int, pageSize int) ([]model.Role, int64, error)
	GetRole(opts ...DBOption) (model.Role, error)
	UpdateRole(uuid uint, role model.Role) error
	CreateRole(role model.Role) (uint64, error)
	DeleteRole(uuid uint) error
	UpdateRoleAccess(roleUuid uint64, accessUuids []uint64) error
}

func NewRoleRepo(db *gorm.DB) IRoleRepo {
	return NewRoleRepoImpl(db)
}

// NewRoleRepoImpl 创建新的角色仓库实现
func NewRoleRepoImpl(db *gorm.DB) *RoleRepoImpl {
	return &RoleRepoImpl{db: db}
}

type RoleRepoImpl struct {
	db *gorm.DB
}

// GetRoleList 获取角色列表，排除逻辑删除的角色
func (r *RoleRepoImpl) GetRoleList(opts ...DBOption) ([]model.Role, error) {
	db := r.db.Model(&model.Role{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	var roles []model.Role
	err := db.Find(&roles).Error
	return roles, errors.WithMessage(err)
}

// UpdateRole 更新角色
func (r *RoleRepoImpl) UpdateRole(uuid uint, role model.Role) error {
	return r.db.Model(&model.Role{}).Where("uuid = ?", uuid).Updates(role).Error
}

// CreateRole 创建角色
func (r *RoleRepoImpl) CreateRole(role model.Role) (uint64, error) {
	return role.Uuid, r.db.Create(&role).Error
}

// DeleteRole 软删除角色
func (r *RoleRepoImpl) DeleteRole(uuid uint) error {
	return r.db.Model(&model.Role{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// WhereUuids 根据ID列表查询角色
func (r *RoleRepoImpl) WhereUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid IN (?)", uuids)
	}
}

// PaginateGetRoleList 分页获取角色列表
func (r *RoleRepoImpl) PaginateGetRoleList(pageNo int, pageSize int) ([]model.Role, int64, error) {
	db := r.db.Model(&model.Role{}).Scopes(NotDeleted)
	var roles []model.Role
	var total int64 = 0
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	db = db.Offset((pageNo - 1) * pageSize).Limit(pageSize)
	err = db.Order("sort ASC, create_time DESC").Find(&roles).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	return roles, total, nil
}

func (r *RoleRepoImpl) GetRole(opts ...DBOption) (model.Role, error) {
	db := r.db.Model(&model.Role{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	var role model.Role
	err := db.First(&role).Error
	return role, errors.WithMessage(err)
}

func (r *RoleRepoImpl) WhereName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	}
}

func (r *RoleRepoImpl) UpdateRoleAccess(roleUuid uint64, accessUuids []uint64) error {
	// 先删除角色权限
	err := r.db.Model(&model.RoleAccess{}).Where("role_uuid = ?", roleUuid).Delete(&model.RoleAccess{}).Error
	if err != nil {
		return err
	}
	// 再添加角色权限
	for _, accessUuid := range accessUuids {
		err = r.db.Model(&model.RoleAccess{}).Create(&model.RoleAccess{
			RoleUuid:   roleUuid,
			AccessUuid: accessUuid,
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *RoleRepoImpl) WithAccesses() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Accesses")
	}
}
