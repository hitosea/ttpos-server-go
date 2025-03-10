package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type IStaffOperationLogRepo interface {
	SaveStaffOperationLog(operationLog model.StaffOperationLog) error
}

func NewStaffOperationLogRepo(db *gorm.DB) IStaffOperationLogRepo {
	return NewOptLogRepoImpl(db)
}

type optLogRepo struct {
	db *gorm.DB
}

func NewOptLogRepoImpl(db *gorm.DB) IStaffOperationLogRepo {
	return &optLogRepo{db: db}
}

func (r *optLogRepo) SaveStaffOperationLog(operationLog model.StaffOperationLog) error {
	return r.db.Model(&model.StaffOperationLog{}).Create(&operationLog).Error
}
