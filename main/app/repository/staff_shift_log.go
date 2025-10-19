package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IShiftLogRepo interface {
	WithStaff() DBOption

	GetPreviousShiftCash() (float64, error)
	Create(shiftLog model.StaffShiftLog) (model.StaffShiftLog, error)
	CreateSnapshot(shiftLogSnapshot model.StaffShiftSnapshot) (model.StaffShiftSnapshot, error)
	GetShiftLog(opts ...DBOption) (model.StaffShiftLog, error)
	GetShiftLogList(opts ...DBOption) ([]model.StaffShiftLog, error)
	Update(shiftLog model.StaffShiftLog, updates map[string]interface{}) (model.StaffShiftLog, error)
	GetSnapshot(opts ...DBOption) (model.StaffShiftSnapshot, error)
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
	err := r.db.Model(&model.StaffShiftLog{}).Where("status = 1").Order("id desc").Limit(1).Select("current_cash_total").Scan(&previewShiftCash).Error
	return previewShiftCash, errors.WithMessage(err)
}

func (r *ShiftLogRepo) Create(shiftLog model.StaffShiftLog) (model.StaffShiftLog, error) {
	err := r.db.Model(&model.StaffShiftLog{}).Create(&shiftLog).Error
	return shiftLog, errors.WithMessage(err)
}

func (r *ShiftLogRepo) CreateSnapshot(shiftLogSnapshot model.StaffShiftSnapshot) (model.StaffShiftSnapshot, error) {
	err := r.db.Model(&model.StaffShiftSnapshot{}).Create(&shiftLogSnapshot).Error
	return shiftLogSnapshot, errors.WithMessage(err)
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

	err := db.First(&log).Error
	return log, errors.WithMessage(err)
}

// Update 更新当班记录
func (r *ShiftLogRepo) Update(shiftLog model.StaffShiftLog, updates map[string]interface{}) (model.StaffShiftLog, error) {
	err := r.db.Model(&model.StaffShiftLog{}).Where("id = ?", shiftLog.ID).Updates(updates).Error
	return shiftLog, errors.WithMessage(err)
}

// WithStaff 预加载收银员
func (r *ShiftLogRepo) WithStaff() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Staff")
	}
}

// GetSnapshot 获取交班快照
func (r *ShiftLogRepo) GetSnapshot(opts ...DBOption) (model.StaffShiftSnapshot, error) {
	var (
		snapshot model.StaffShiftSnapshot
		db       *gorm.DB = r.db
	)

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&snapshot).Error
	return snapshot, errors.WithMessage(err)
}

// GetShiftLogList 获取班次列表
func (r *ShiftLogRepo) GetShiftLogList(opts ...DBOption) ([]model.StaffShiftLog, error) {
	var logs []model.StaffShiftLog
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&logs).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return logs, nil
}
