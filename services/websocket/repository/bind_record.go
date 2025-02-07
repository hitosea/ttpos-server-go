package repository

import (
	"websocket/model"
	"websocket/pkg/database"
)

type BindRecordRepository struct {
	dbm *database.DBManager
}

func NewBindRecordRepository(dbm *database.DBManager) *BindRecordRepository {
	return &BindRecordRepository{dbm: dbm}
}

func (r *BindRecordRepository) Unbind(companyId uint, source string, key string, staffId uint) error {
	return r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Select("finally_login_id").
		Where("`source` = ? AND `key` = ? AND `finally_login_id` = ?", source, key, staffId).Debug().
		Updates(map[string]interface{}{
			"finally_login_id": 0,
		}).Error
}

func (r *BindRecordRepository) GetBindCount(companyId uint, source string) uint {
	var c int64
	r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("source = ? AND finally_login_id > 0", source).Count(&c)
	return uint(c)
}

func (r *BindRecordRepository) GetRecordBySourceAndDeviceId(companyId uint, source string, deviceId string) model.BindRecord {
	var bindRecord model.BindRecord
	r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).First(&bindRecord)
	return bindRecord
}

func (r *BindRecordRepository) Update(companyId uint, id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("id = ?", id).Updates(vars).Error
}

func (r *BindRecordRepository) Create(companyId uint, bindRecord model.BindRecord) error {
	return r.dbm.GetDB(companyId).Create(&bindRecord).Error
}
