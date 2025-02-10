package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IBindRecordRepo interface {
	Unbind(companyId uint64, source string, key string, staffId uint64) error
	GetBindCount(companyId uint64, source string) uint
	GetRecordBySourceAndDeviceId(companyId uint64, source string, deviceId string) model.BindRecord
	Update(companyId uint64, id uint, vars map[string]interface{}) error
	Create(companyId uint64, bindRecord model.BindRecord) error
	GetRemark(companyId uint64, source string, deviceId string) string
	GetBindRecordId(source string, deviceId string) uint
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

func (r *BindRecordRepo) Unbind(companyId uint64, source string, key string, staffId uint64) error {
	return r.db.Model(&model.BindRecord{}).Select("finally_login_id").
		Where("`source` = ? AND `key` = ? AND `finally_login_id` = ?", source, key, staffId).Debug().
		Updates(map[string]interface{}{
			"finally_login_id": 0,
		}).Error
}

func (r *BindRecordRepo) GetBindCount(companyId uint64, source string) uint {
	var c int64
	r.db.Model(&model.BindRecord{}).Where("source = ? AND finally_login_id > 0", source).Count(&c)
	return uint(c)
}

func (r *BindRecordRepo) GetRecordBySourceAndDeviceId(companyId uint64, source string, deviceId string) model.BindRecord {
	var bindRecord model.BindRecord
	r.db.Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).First(&bindRecord)
	return bindRecord
}

func (r *BindRecordRepo) Update(companyId uint64, id uint, vars map[string]interface{}) error {
	return r.db.Model(&model.BindRecord{}).Where("id = ?", id).Updates(vars).Error
}

func (r *BindRecordRepo) Create(companyId uint64, bindRecord model.BindRecord) error {
	return r.db.Create(&bindRecord).Error
}

func (r *BindRecordRepo) GetRemark(companyId uint64, source string, deviceId string) string {
	var remark string
	r.db.Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).Select("remark").Scan(&remark)
	return remark
}

func (r *BindRecordRepo) GetBindRecordId(source string, deviceId string) uint {
	var id uint
	r.db.Model(&model.BindRecord{}).Where("source = ? AND device_id = ?", source, deviceId).Select("id").Scan(&id)
	return id
}
