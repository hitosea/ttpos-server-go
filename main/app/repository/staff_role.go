package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffRoleRepo interface {
	CreateStaffRole(staffRole model.StaffRole) error
	GetRoleUuidsByStaffUuid(staffUuid uint64) ([]uint64, error)
	GetStaffUuidsByRoleUuid(roleUuid uint64) ([]uint64, error) // 根据角色UUID查询关联的员工UUID列表
	DeleteStaffRolesByRoleUuid(roleUuid uint64) error          // 删除角色的所有员工关联
	CreateStaffRoles(staffUuid uint64, roleUuids []uint64) error // 批量创建员工角色关联
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
	err := r.db.Model(&model.StaffRole{}).Where("staff_uuid = ?", staffUuid).Pluck("role_uuid", &roleIds).Error
	return roleIds, errors.WithMessage(err)
}

// GetStaffUuidsByRoleUuid 根据角色UUID查询关联的员工UUID列表
func (r *StaffRoleRepo) GetStaffUuidsByRoleUuid(roleUuid uint64) ([]uint64, error) {
	var staffUuids []uint64
	err := r.db.Model(&model.StaffRole{}).Where("role_uuid = ?", roleUuid).Pluck("staff_uuid", &staffUuids).Error
	return staffUuids, errors.WithMessage(err)
}

// DeleteStaffRolesByRoleUuid 删除角色的所有员工关联
func (r *StaffRoleRepo) DeleteStaffRolesByRoleUuid(roleUuid uint64) error {
	return r.db.Model(&model.StaffRole{}).Where("role_uuid = ?", roleUuid).Delete(&model.StaffRole{}).Error
}

// CreateStaffRoles 批量创建员工角色关联
func (r *StaffRoleRepo) CreateStaffRoles(staffUuid uint64, roleUuids []uint64) error {
	for _, roleUuid := range roleUuids {
		err := r.db.Model(&model.StaffRole{}).Create(&model.StaffRole{
			StaffUuid: int64(staffUuid),
			RoleUuid:  int64(roleUuid),
		}).Error
		if err != nil {
			return errors.WithMessage(err, "创建员工角色关联失败")
		}
	}
	return nil
}
