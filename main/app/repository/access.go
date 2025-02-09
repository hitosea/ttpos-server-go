package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
)

type IAccessRepo interface {
	GetPermissions(companyId uint, where ...Where) ([]model.Access, error)
	GetAccessIds(roleIds []int, appId uint) ([]int, error)
	CreateAccess(access *model.Access) error
	WhereIds(accessIds []int) Where
	WherePath(excludePath []string) Where
	WhereIsSupplier() Where
}

func NewAccessRepo(db *gorm.DB) IAccessRepo {
	return NewAccessRepoImpl(db)
}

type AccessRepo struct {
	db *gorm.DB
}

func NewAccessRepoImpl(db *gorm.DB) *AccessRepo {
	return &AccessRepo{db: db}
}

func (r *AccessRepo) CreateAccess(access *model.Access) error {
	return r.db.Create(access).Error
}

func (r *AccessRepo) GetPermissions(companyId uint, where ...Where) ([]model.Access, error) {
	var access []model.Access
	for _, w := range where {
		r.db = w(r.db)
	}
	err := r.db.Order("sort asc, create_time asc").Debug().Find(&access).Error
	return access, err
}

func (r *AccessRepo) GetAccessIds(roleIds []int, appId uint) ([]int, error) {
	var accessIds []int
	err := r.db.Model(&model.RoleAccess{}).Where("role_id in (?)", roleIds).Debug().Pluck("access_id", &accessIds).Error
	return accessIds, err
}

func (r *AccessRepo) WhereIds(accessIds []int) Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id in (?)", accessIds)
	}
}

func (r *AccessRepo) WherePath(excludePath []string) Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path no in (?)", excludePath)
	}
}

func (r *AccessRepo) WhereIsSupplier() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_supplier", 1)
	}
}
