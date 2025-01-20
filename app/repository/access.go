package repository

import (
	"gorm.io/gorm"

	"jjjshop-server-go/app/model"
	"jjjshop-server-go/pkg/database"
)

type AccessRepository struct {
	dbm *database.DBManager
}

func NewAccessRepository(dbm *database.DBManager) *AccessRepository {
	return &AccessRepository{dbm: dbm}
}

func (r *AccessRepository) GetPermissions(appId uint, where ...Where) ([]model.ShopAccess, error) {
	var access []model.ShopAccess
	db := r.dbm.GetDB(appId)
	for _, w := range where {
		db = w(db)
	}
	err := db.Order("sort asc, create_time asc").Debug().Find(&access).Error
	return access, err
}

func (r *AccessRepository) GetAccessIds(roleIds []int, appId uint) ([]int, error) {
	var accessIds []int
	err := r.dbm.GetDB(appId).Model(&model.ShopRoleAccess{}).Where("role_id in (?)", roleIds).Debug().Pluck("access_id", &accessIds).Error
	return accessIds, err
}

func (r *AccessRepository) WhereIds(accessIds []int) Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("access_id in (?)", accessIds)
	}
}

func (r *AccessRepository) WherePath(excludePath []string) Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path no in (?)", excludePath)
	}
}

func (r *AccessRepository) WhereIsSupplier() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_supplier", 1)
	}
}
