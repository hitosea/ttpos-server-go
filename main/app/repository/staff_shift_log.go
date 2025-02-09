package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IShiftLogRepo interface {
	GetPreviousShiftCash(companyId uint) (float64, error)
	Create(companyId uint, shiftLog model.StaffShiftLog) (model.StaffShiftLog, error)
}

func NewShiftLogRepo(db *gorm.DB) IShiftLogRepo {
	return NewShiftLogRepoImpl(db)
}

type ShiftLogRepo struct {
	db *gorm.DB
}

func NewShiftLogRepoImpl(db *gorm.DB) *ShiftLogRepo {
	return &ShiftLogRepo{db: db}
}

func (r *ShiftLogRepo) GetPreviousShiftCash(companyId uint) (float64, error) {
	var previewShiftCash float64
	err := r.db.Model(&model.StaffShiftLog{}).Where("status = 1").Order("id desc").Select("previous_shift_cash").Scan(&previewShiftCash).Error
	return previewShiftCash, err
}

func (r *ShiftLogRepo) Create(companyId uint, shiftLog model.StaffShiftLog) (model.StaffShiftLog, error) {
	err := r.db.Model(&model.StaffShiftLog{}).Create(&shiftLog).Error
	return shiftLog, err
}
