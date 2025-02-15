package repository

import (
	"websocket/model"
	"websocket/pkg/database"
)

type DeviceRepository struct {
	dbm *database.DBManager
}

func NewDeviceRepository(dbm *database.DBManager) *DeviceRepository {
	return &DeviceRepository{dbm: dbm}
}

func (r *DeviceRepository) Unbind(companyId uint, source string, key string, staffId uint) error {
	return r.dbm.GetDB(companyId).Model(&model.Device{}).Select("finally_login_id").
		Where("`source` = ? AND `key` = ? AND `finally_login_id` = ?", source, key, staffId).Debug().
		Updates(map[string]interface{}{
			"finally_login_id": 0,
		}).Error
}

func (r *DeviceRepository) GetBindCount(companyId uint, source string) uint {
	var c int64
	r.dbm.GetDB(companyId).Model(&model.Device{}).Where("source = ? AND finally_login_id > 0", source).Count(&c)
	return uint(c)
}

func (r *DeviceRepository) GetRecordBySourceAndDeviceId(companyId uint, source string, deviceId string) model.Device {
	var Device model.Device
	r.dbm.GetDB(companyId).Model(&model.Device{}).Where("source = ? AND device_id = ?", source, deviceId).First(&Device)
	return Device
}

func (r *DeviceRepository) Update(companyId uint, id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyId).Model(&model.Device{}).Where("id = ?", id).Updates(vars).Error
}

func (r *DeviceRepository) Create(companyId uint, Device model.Device) error {
	return r.dbm.GetDB(companyId).Create(&Device).Error
}
