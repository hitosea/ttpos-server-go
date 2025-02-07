package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type IBindRecordRepo interface {
	Unbind(companyId uint, source string, key string, staffId uint) error
	GetBindCount(companyId uint, source string) uint
	GetRecordBySourceAndDeviceId(companyId uint, source string, deviceId string) model.BindRecord
	Update(companyId uint, id uint, vars map[string]interface{}) error
	Create(companyId uint, bindRecord model.BindRecord) error
}

type BindRecordRepo struct {
	dbm *database.DBManager
}

func NewBindRecordRepo(dbm *database.DBManager) IBindRecordRepo {
	return NewBindRecordRepoImpl(dbm)
}

func NewBindRecordRepoImpl(dbm *database.DBManager) *BindRecordRepo {
	return &BindRecordRepo{dbm: dbm}
}

func (r *BindRecordRepo) Unbind(companyId uint, source string, key string, staffId uint) error {
	return r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Select("finally_login_id").
		Where("`source` = ? AND `key` = ? AND `finally_login_id` = ?", source, key, staffId).Debug().
		Updates(map[string]interface{}{
			"finally_login_id": 0,
		}).Error
}

func (r *BindRecordRepo) GetBindCount(companyId uint, source string) uint {
	var c int64
	r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("source = ? AND finally_login_id > 0", source).Count(&c)
	return uint(c)
}

func (r *BindRecordRepo) GetRecordBySourceAndDeviceId(companyId uint, source string, deviceId string) model.BindRecord {
	var bindRecord model.BindRecord
	r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).First(&bindRecord)
	return bindRecord
}

func (r *BindRecordRepo) Update(companyId uint, id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("id = ?", id).Updates(vars).Error
}

func (r *BindRecordRepo) Create(companyId uint, bindRecord model.BindRecord) error {
	return r.dbm.GetDB(companyId).Create(&bindRecord).Error
}
