package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

type IDeviceRepo interface {
	GetDeviceBySn(ctx context.Context, deviceSn string) (*model.Device, error)
	GetDeviceBrand(opts ...DBOption) string

	WhereDeviceId(deviceId string) DBOption
}

func NewDeviceRepo(db *gorm.DB) IDeviceRepo {
	return NewDeviceRepoImpl(db)
}

type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepoImpl(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}
func (r *DeviceRepo) GetDeviceBySn(ctx context.Context, deviceSn string) (*model.Device, error) {
	var device model.Device
	err := r.db.Model(&model.Device{}).Where("device_id = ?", deviceSn).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepo) GetDeviceBrand(opts ...DBOption) string {
	var brand string
	db := r.db.Model(&model.Device{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Select("brand").Scan(&brand)
	return brand
}

func (r *DeviceRepo) WhereDeviceId(deviceId string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_id = ?", deviceId)
	}
}
