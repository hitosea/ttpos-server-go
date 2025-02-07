package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type IStaffRoleRepo interface {
	GetRoleIds(staffId uint, appId uint) ([]int, error)
}

func NewStaffRoleRepo(dbm *database.DBManager) IStaffRoleRepo {
	return NewStaffRoleRepoImpl(dbm)
}

type StaffRoleRepo struct {
	dbm *database.DBManager
}

func NewStaffRoleRepoImpl(dbm *database.DBManager) *StaffRoleRepo {
	return &StaffRoleRepo{dbm: dbm}
}

func (r *StaffRoleRepo) GetRoleIds(staffId uint, appId uint) ([]int, error) {
	var roleIds []int
	err := r.dbm.GetDB(appId).Model(&model.StaffRole{}).Where("staff_id = ?", staffId).Debug().Pluck("role_id", &roleIds).Error
	return roleIds, err
}
