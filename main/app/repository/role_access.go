package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IRoleAccessRepo 角色权限仓库接口
type IRoleAccessRepo interface {
	GetRoleAccessList() ([]model.RoleAccess, error)
	CreateRoleAccess(roleAccess model.RoleAccess) (uint64, error)
}

func NewRoleAccessRepo(db *gorm.DB) IRoleAccessRepo {
	return NewRoleAccessRepoImpl(db)
}

func NewRoleAccessRepoImpl(db *gorm.DB) IRoleAccessRepo {
	return &roleAccessRepo{db: db}
}

type roleAccessRepo struct {
	db *gorm.DB
}

func (r *roleAccessRepo) GetRoleAccessList() ([]model.RoleAccess, error) {
	var roleAccesses []model.RoleAccess
	err := r.db.Model(&model.RoleAccess{}).Find(&roleAccesses).Error
	return roleAccesses, errors.WithMessage(err)
}

func (r *roleAccessRepo) CreateRoleAccess(roleAccess model.RoleAccess) (uint64, error) {
	return roleAccess.RoleUuid, r.db.Create(&roleAccess).Error
}
