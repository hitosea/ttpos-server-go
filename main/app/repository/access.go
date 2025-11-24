package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type IAccessRepo interface {
	WhereUuids(accessUuids []uint64) DBOption
	WhereIsSupplier() DBOption
	WhereApiPath(apiPath string) DBOption

	GetAccess(opts ...DBOption) model.Access
	GetPermissions(opts ...DBOption) ([]model.Access, error)
	GetAccessUuids(roleUuids []uint64) ([]uint64, error)
	CreateAccess(access *model.Access) error
}

func NewAccessRepo(db *gorm.DB) IAccessRepo {
	return NewAccessRepoImpl(db)
}

type accessRepo struct {
	db *gorm.DB
}

func NewAccessRepoImpl(db *gorm.DB) IAccessRepo {
	return &accessRepo{db: db}
}

func (r *accessRepo) CreateAccess(access *model.Access) error {
	return r.db.Model(&model.Access{}).Create(access).Error
}

func (r *accessRepo) GetPermissions(opts ...DBOption) ([]model.Access, error) {
	var access []model.Access
	db := r.db.Model(&model.Access{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc, create_time asc, id asc").Find(&access).Error
	return access, errors.WithMessage(err)
}

func (r *accessRepo) GetAccess(opts ...DBOption) model.Access {
	var access model.Access
	db := r.db.Model(&model.Access{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&access)
	return access
}

func (r *accessRepo) GetAccessUuids(roleIds []uint64) ([]uint64, error) {
	var accessUuids []uint64
	err := r.db.Model(&model.RoleAccess{}).Where("role_uuid in (?)", roleIds).Pluck("access_uuid", &accessUuids).Error
	return accessUuids, errors.WithMessage(err)
}

func (r *accessRepo) WhereUuids(accessIds []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid in (?)", accessIds)
	}
}

func (r *accessRepo) WhereIsSupplier() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_supplier", 1)
	}
}

func (r *accessRepo) WhereApiPath(apiPath string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("api_path = ?", apiPath)
	}
}
