package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRoleRepo interface {
	CreateStaffRole(staffRole model.StaffRole) error
	GetRoleIds(staffId uint64) ([]uint64, error)
}

func NewStaffRoleRepo(db *gorm.DB) IStaffRoleRepo {
	return NewStaffRoleRepoImpl(db)
}

type StaffRoleRepo struct {
	db *gorm.DB
}

func NewStaffRoleRepoImpl(db *gorm.DB) *StaffRoleRepo {
	return &StaffRoleRepo{db: db}
}

func (r *StaffRoleRepo) CreateStaffRole(staffRole model.StaffRole) error {
	return r.db.Create(&staffRole).Error
}

func (r *StaffRoleRepo) GetRoleIds(staffId uint64) ([]uint64, error) {
	var roleIds []uint64
	err := r.db.Model(&model.StaffRole{}).Where("staff_id = ?", staffId).Debug().Pluck("role_id", &roleIds).Error
	return roleIds, err
}
