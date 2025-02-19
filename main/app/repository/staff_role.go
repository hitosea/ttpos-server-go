package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRoleRepo interface {
	CreateStaffRole(staffRole model.StaffRole) error
	GetRoleUuidsByStaffUuid(staffUuid uint64) ([]uint64, error)
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
	return r.db.Model(&model.StaffRole{}).Create(&staffRole).Error
}

func (r *StaffRoleRepo) GetRoleUuidsByStaffUuid(staffUuid uint64) ([]uint64, error) {
	var roleIds []uint64
	err := r.db.Model(&model.StaffRole{}).Where("staff_uuid = ?", staffUuid).Debug().Pluck("role_uuid", &roleIds).Error
	return roleIds, err
}
