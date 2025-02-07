package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type IAccessRepo interface {
	GetPermissions(companyId uint, where ...Where) ([]model.Access, error)
	GetAccessIds(roleIds []int, appId uint) ([]int, error)
	WhereIds(accessIds []int) Where
	WherePath(excludePath []string) Where
	WhereIsSupplier() Where
}

func NewAccessRepo(dbm *database.DBManager) IAccessRepo {
	return NewAccessRepoImpl(dbm)
}

type AccessRepo struct {
	dbm *database.DBManager
}

func NewAccessRepoImpl(dbm *database.DBManager) *AccessRepo {
	return &AccessRepo{dbm: dbm}
}

func (r *AccessRepo) GetPermissions(companyId uint, where ...Where) ([]model.Access, error) {
	var access []model.Access
	db := r.dbm.GetDB(companyId)
	for _, w := range where {
		db = w(db)
	}
	err := db.Order("sort asc, create_time asc").Debug().Find(&access).Error
	return access, err
}

func (r *AccessRepo) GetAccessIds(roleIds []int, appId uint) ([]int, error) {
	var accessIds []int
	err := r.dbm.GetDB(appId).Model(&model.RoleAccess{}).Where("role_id in (?)", roleIds).Debug().Pluck("access_id", &accessIds).Error
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
