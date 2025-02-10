package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IBindRecordRepo interface {
	Unbind(source string, deviceId string, staffUuid uint64) error
	GetBindCount(source string) uint
	GetRecordBySourceAndDeviceId(source string, deviceId string) model.BindRecord
	Update(uuid uint64, vars map[string]interface{}) error
	Create(bindRecord model.BindRecord) error
	GetRemark(source string, deviceId string) string
	GetBindRecordUuid(source string, deviceId string) uint64
}

type BindRecordRepo struct {
	db *gorm.DB
}

func NewBindRecordRepo(db *gorm.DB) IBindRecordRepo {
	return NewBindRecordRepoImpl(db)
}

func NewBindRecordRepoImpl(db *gorm.DB) *BindRecordRepo {
	return &BindRecordRepo{db: db}
}

func (r *BindRecordRepo) Unbind(source string, deviceId string, staffUuid uint64) error {
	return r.db.Model(&model.BindRecord{}).Select("finally_login_uuid").
		Where("`source` = ? AND `device_id` = ? AND `finally_login_uuid` = ?", source, deviceId, staffUuid).Debug().
		Updates(map[string]interface{}{
			"finally_login_uuid": 0,
		}).Error
}

func (r *BindRecordRepo) GetBindCount(source string) uint {
	var c int64
	r.db.Model(&model.BindRecord{}).Where("source = ? AND finally_login_id > 0", source).Count(&c)
	return uint(c)
}

func (r *BindRecordRepo) GetRecordBySourceAndDeviceId(source string, deviceId string) model.BindRecord {
	var bindRecord model.BindRecord
	r.db.Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).First(&bindRecord)
	return bindRecord
}

func (r *BindRecordRepo) Update(uuid uint64, vars map[string]interface{}) error {
	return r.db.Model(&model.BindRecord{}).Where("uuid = ?", uuid).Updates(vars).Error
}

func (r *BindRecordRepo) Create(bindRecord model.BindRecord) error {
	return r.db.Create(&bindRecord).Error
}

func (r *BindRecordRepo) GetRemark(source string, deviceId string) string {
	var remark string
	r.db.Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).Select("remark").Scan(&remark)
	return remark
}

func (r *BindRecordRepo) GetBindRecordUuid(source string, deviceId string) uint64 {
	var uuid uint64
	r.db.Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).Select("uuid").Scan(&uuid)
	return uuid
}
