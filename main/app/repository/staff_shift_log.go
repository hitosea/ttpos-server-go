package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type IShiftLogRepo interface {
	GetPreviousShiftCash(companyId uint) (float64, error)
	Create(companyId uint, shiftLog model.StaffShiftLog) (model.StaffShiftLog, error)
}

func NewShiftLogRepo(dbm *database.DBManager) IShiftLogRepo {
	return NewShiftLogRepoImpl(dbm)
}

type ShiftLogRepo struct {
	dbm *database.DBManager
}

func NewShiftLogRepoImpl(dbm *database.DBManager) *ShiftLogRepo {
	return &ShiftLogRepo{dbm: dbm}
}

func (r *ShiftLogRepo) GetPreviousShiftCash(companyId uint) (float64, error) {
	var previewShiftCash float64
	err := r.dbm.GetDB(companyId).Model(&model.StaffShiftLog{}).Where("status = 1").Order("id desc").Select("previous_shift_cash").Scan(&previewShiftCash).Error
	return previewShiftCash, err
}

func (r *ShiftLogRepo) Create(companyId uint, shiftLog model.StaffShiftLog) (model.StaffShiftLog, error) {
	err := r.dbm.GetDB(companyId).Model(&model.StaffShiftLog{}).Create(&shiftLog).Error
	return shiftLog, err
}
