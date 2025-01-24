package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type ShiftLogRepository struct {
	dbm *database.DBManager
}

func NewShiftLogRepository(dbm *database.DBManager) *ShiftLogRepository {
	return &ShiftLogRepository{dbm: dbm}
}

func (r *ShiftLogRepository) GetPreviousShiftCash(companyId uint) (float64, error) {
	var previewShiftCash float64
	err := r.dbm.GetDB(companyId).Model(&model.StaffShiftLog{}).Where("status = 1").Order("id desc").Select("previous_shift_cash").Scan(&previewShiftCash).Error
	return previewShiftCash, err
}

func (r *ShiftLogRepository) Create(companyId uint, shiftLog model.StaffShiftLog) (model.StaffShiftLog, error) {
	err := r.dbm.GetDB(companyId).Model(&model.StaffShiftLog{}).Create(&shiftLog).Error
	return shiftLog, err
}
