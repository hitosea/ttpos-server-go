package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

type IDeviceRepo interface {
	GetDeviceBySn(ctx context.Context, deviceSn string) (*model.Device, error)
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
