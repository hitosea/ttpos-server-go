package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type StaffRoleRepository struct {
	dbm *database.DBManager
}

func NewStaffRoleRepository(dbm *database.DBManager) *StaffRoleRepository {
	return &StaffRoleRepository{dbm: dbm}
}

func (r *StaffRoleRepository) GetRoleIds(staffId uint, appId uint) ([]int, error) {
	var roleIds []int
	err := r.dbm.GetDB(appId).Model(&model.StaffRole{}).Where("staff_id = ?", staffId).Debug().Pluck("role_id", &roleIds).Error
	return roleIds, err
}
