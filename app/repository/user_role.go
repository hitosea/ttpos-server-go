package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type UserRoleRepository struct {
	dbm *database.DBManager
}

func NewUserRoleRepository(dbm *database.DBManager) *UserRoleRepository {
	return &UserRoleRepository{dbm: dbm}
}

func (r *UserRoleRepository) GetRoleIds(shoUserId uint, appId uint) ([]int, error) {
	var roleIds []int
	err := r.dbm.GetDB(appId).Model(&model.StaffRole{}).Where("staff_id = ?", shoUserId).Debug().Pluck("role_id", &roleIds).Error
	return roleIds, err
}
