package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
)

type IAccessRepo interface {
	WhereUuids(accessUuids []uint64) DBOption
	WhereIsSupplier() DBOption

	GetPermissions(opts ...DBOption) ([]model.Access, error)
	GetAccessUuids(roleUuids []uint64) ([]uint64, error)
	CreateAccess(access *model.Access) error
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

func (r *AccessRepo) GetPermissions(opts ...DBOption) ([]model.Access, error) {
	var access []model.Access
	db := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc, create_time asc").Debug().Find(&access).Error
	return access, err
}

func (r *AccessRepo) GetAccessUuids(roleIds []uint64) ([]uint64, error) {
	var accessUuids []uint64
	err := r.db.Model(&model.RoleAccess{}).Where("role_uuid in (?)", roleIds).Debug().Pluck("access_uuid", &accessUuids).Error
	return accessUuids, err
}

func (r *AccessRepo) WhereUuids(accessIds []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid in (?)", accessIds)
	}
}

func (r *AccessRepo) WhereIsSupplier() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_supplier", 1)
	}
}
