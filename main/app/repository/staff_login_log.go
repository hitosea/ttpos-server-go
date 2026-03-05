package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IStaffLoginLogRepo interface {
	Create(log *model.StaffLoginLog) error
}

type staffLoginLogRepo struct {
	db *gorm.DB
}

func NewStaffLoginLogRepo(db *gorm.DB) IStaffLoginLogRepo {
	return &staffLoginLogRepo{db: db}
}

func (r *staffLoginLogRepo) Create(log *model.StaffLoginLog) error {
	return r.db.Create(log).Error
}
