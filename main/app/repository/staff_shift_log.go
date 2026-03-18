package repository

import (
	"ttpos-server-go/app/constant"
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
	GetShiftLogByStaffUuid(staffUuid uint64, opts ...DBOption) (model.StaffShiftLog, error)
	GetShiftLogList(opts ...DBOption) ([]model.StaffShiftLog, error)
	Update(shiftLog model.StaffShiftLog, updates map[string]interface{}) (model.StaffShiftLog, error)
	GetSnapshot(opts ...DBOption) (model.StaffShiftSnapshot, error)
	GetAvailableShiftInfo() (shiftNo string, staffUuid uint64, staffName string, err error)
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

// GetShiftLogByStaffUuid 根据员工UUID获取班次记录
func (r *ShiftLogRepo) GetShiftLogByStaffUuid(staffUuid uint64, opts ...DBOption) (model.StaffShiftLog, error) {
	var (
		shiftLog model.StaffShiftLog
		db       *gorm.DB = r.db
	)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("staff_uuid = ?", staffUuid).Where("status = ?", constant.StaffNotHandedOver).First(&shiftLog).Error
	return shiftLog, errors.WithMessage(err)
}

// GetAvailableShiftInfo 获取可用的班次信息（班次编号、员工UUID、员工姓名）
// 优先级规则（只管班次，不管登录状态）：
// 1. 主收银机的班次（优先级最高）
// 2. 非主收银机中，最先登录的班次（按 finally_login_time ASC 排序）
// 3. 若都没有未交班的班次，返回空值（等待下次登录时处理）
func (r *ShiftLogRepo) GetAvailableShiftInfo() (shiftNo string, staffUuid uint64, staffName string, err error) {
	// 定义查询结果结构体
	type shiftStaffInfo struct {
		ShiftNo          string `gorm:"column:shift_no"`
		StaffUuid        uint64 `gorm:"column:staff_uuid"`
		StaffName        string `gorm:"column:staff_name"`
		IsMain           uint8  `gorm:"column:is_main"`
		FinallyLoginTime int64  `gorm:"column:finally_login_time"`
	}

	// 一次性查询所有未交班的班次（关联设备信息）
	var results []shiftStaffInfo
	err = r.db.Model(&model.StaffShiftLog{}).
		Select("ttpos_staff_shift_log.shift_no, ttpos_staff_shift_log.staff_uuid, "+
			"COALESCE(NULLIF(ttpos_staff.real_name, ''), ttpos_staff.username) AS staff_name, "+
			"ttpos_device.is_main, ttpos_device.finally_login_time").
		Joins("INNER JOIN ttpos_staff ON ttpos_staff_shift_log.staff_uuid = ttpos_staff.uuid").
		Joins("INNER JOIN ttpos_device ON ttpos_staff.bind_key = ttpos_device.device_id").
		Where("ttpos_staff_shift_log.status = ?", 0).
		Where("ttpos_device.source = ?", "cashier").
		Where("ttpos_device.delete_time = ?", 0).
		Where("ttpos_staff.delete_time = ?", 0).
		Order("ttpos_device.is_main DESC").           // 主收银机优先（1 在前，0 在后）
		Order("ttpos_device.finally_login_time ASC"). // 同为非主收银机时，按登录时间升序
		Order("ttpos_staff_shift_log.uuid DESC").     // 同一设备多个班次时，取最新的
		Find(&results).Error

	if err != nil {
		return "", 0, "", errors.WithMessage(err)
	}
	if len(results) == 0 {
		return "", 0, "", nil
	}

	// 返回优先级最高的班次（已按规则排序，直接取第一个）
	first := results[0]
	return first.ShiftNo, first.StaffUuid, first.StaffName, nil
}
