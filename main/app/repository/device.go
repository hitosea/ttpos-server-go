package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type IDeviceRepo interface {
	WhereSn(deviceId string) DBOption   // sn/device_id 条件
	WhereSource(source string) DBOption // 来源（cashier、tablet、kitchen、assistant）条件

	GetDevice(opts ...DBOption) (model.Device, error)            // 获取设备
	GetDeviceBrand(opts ...DBOption) string                      // 获取设备品牌
	GetBindCountBySource(source string) uint                     // 根据来源获取设备绑定数量
	CreateDevice(device model.Device) (model.Device, error)      // 创建设备
	UpdateDevice(uuid uint64, vars map[string]interface{}) error // 更新设备
}

func NewDeviceRepo(db *gorm.DB) IDeviceRepo {
	return NewDeviceRepoImpl(db)
}

type deviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepoImpl(db *gorm.DB) IDeviceRepo {
	return &deviceRepo{db: db}
}

func (r *deviceRepo) GetDevice(opts ...DBOption) (model.Device, error) {
	var device model.Device
	db := r.db.Model(&model.Device{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&device).Error
	return device, errors.WithMessage(err, "db.First failed")
}
func (r *deviceRepo) GetDeviceBrand(opts ...DBOption) string {
	var brand string
	db := r.db.Model(&model.Device{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Select("brand").Scan(&brand)
	return brand
}

func (r *deviceRepo) WhereSn(deviceId string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_id = ?", deviceId)
	}
}

func (r *deviceRepo) WhereSource(source string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source = ?", source)
	}
}

func (r *deviceRepo) GetBindCountBySource(source string) uint {
	var count int64
	r.db.Model(&model.Device{}).Where("source = ? AND finally_login_uuid > 0", source).Count(&count)
	return uint(count)
}

func (r *deviceRepo) UpdateDevice(uuid uint64, vars map[string]interface{}) error {
	return r.db.Model(&model.Device{}).Where("uuid = ?", uuid).Updates(vars).Error
}

func (r *deviceRepo) CreateDevice(device model.Device) (model.Device, error) {
	err := r.db.Model(&model.Device{}).Create(&device).Error
	return device, errors.WithMessage(err)
}
