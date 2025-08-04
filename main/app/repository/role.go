package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IRoleRepo 角色
type IRoleRepo interface {
	// 根据ID列表查询角色
	WhereUuids(uuids []uint64) DBOption
	// 获取角色列表，排除逻辑删除的角色
	GetRoleList(opts ...DBOption) ([]model.Role, error)
	UpdateRole(uuid uint, role model.Role) error
	CreateRole(role model.Role) (uint64, error)
	DeleteRole(uuid uint) error
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
