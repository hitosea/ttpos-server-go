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

func (r *DeviceRepository) Unbind(companyUuid uint64, source string, key string, staffId uint) error {
	return r.dbm.GetDB(companyUuid).Model(&model.Device{}).Select("finally_login_id").
		Where("`source` = ? AND `key` = ? AND `finally_login_id` = ?", source, key, staffId).
		Updates(map[string]interface{}{
			"finally_login_id": 0,
		}).Error
}

func (r *DeviceRepository) GetBindCount(companyUuid uint64, source string) uint {
	var c int64
	r.dbm.GetDB(companyUuid).Model(&model.Device{}).Where("source = ? AND finally_login_id > 0", source).Count(&c)
	return uint(c)
}

func (r *DeviceRepository) GetRecordBySourceAndDeviceId(companyUuid uint64, source string, deviceId string) model.Device {
	var Device model.Device
	r.dbm.GetDB(companyUuid).Model(&model.Device{}).Where("source = ? AND device_id = ?", source, deviceId).First(&Device)
	return Device
}

func (r *DeviceRepository) Update(companyUuid uint64, id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyUuid).Model(&model.Device{}).Where("id = ?", id).Updates(vars).Error
}

func (r *DeviceRepository) Create(companyUuid uint64, Device model.Device) error {
	return r.dbm.GetDB(companyUuid).Create(&Device).Error
}
