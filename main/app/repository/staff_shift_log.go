package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IShiftLogRepo interface {
	GetPreviousShiftCash() (float64, error)
	Create(shiftLog model.StaffShiftLog) (model.StaffShiftLog, error)
	GetShiftLog(opts ...DBOption) (model.StaffShiftLog, error)
	Update(shiftLog model.StaffShiftLog, updates map[string]interface{}) (model.StaffShiftLog, error)
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

func (r *ShiftLogRepo) GetPreviousShiftCash() (float64, error) {
	var previewShiftCash float64
	err := r.db.Model(&model.StaffShiftLog{}).Where("status = 1").Order("id desc").Select("previous_shift_cash").Scan(&previewShiftCash).Error
	return previewShiftCash, errors.WithMessage(err)
}

func (r *ShiftLogRepo) Create(shiftLog model.StaffShiftLog) (model.StaffShiftLog, error) {
	err := r.db.Model(&model.StaffShiftLog{}).Create(&shiftLog).Error
	return shiftLog, errors.WithMessage(err)
}

// GetShiftLog 获取当班记录
func (r *ShiftLogRepo) GetShiftLog(opts ...DBOption) (model.StaffShiftLog, error) {
	var (
		log model.StaffShiftLog
		db  *gorm.DB = r.db
	)

	for _, opt := range opts {
		db = opt(db)
	}

	err := r.db.First(&log).Error
	return log, errors.WithMessage(err)
}

// Update 更新当班记录
func (r *ShiftLogRepo) Update(shiftLog model.StaffShiftLog, updates map[string]interface{}) (model.StaffShiftLog, error) {
	err := r.db.Model(&model.StaffShiftLog{}).Where("id = ?", shiftLog.ID).Updates(updates).Error
	return shiftLog, errors.WithMessage(err)
}
