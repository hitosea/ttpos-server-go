package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IBindRecordRepo interface {
	Unbind(source string, deviceId string, staffUuid uint64) error          // 解绑
	GetBindCountBySource(source string) uint                                // 根据来源获取绑定数量
	GetBySourceAndDeviceId(source string, deviceId string) model.BindRecord // 根据来源和设备ID获取绑定记录
	Update(uuid uint64, vars map[string]interface{}) error                  // 更新绑定记录
	Create(bindRecord model.BindRecord) error                               // 创建绑定记录
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

func (r *BindRecordRepo) GetBindCountBySource(source string) uint {
	var count int64
	r.db.Model(&model.BindRecord{}).Where("source = ? AND finally_login_uuid > 0", source).Count(&count)
	return uint(count)
}

func (r *BindRecordRepo) GetBySourceAndDeviceId(source string, deviceId string) model.BindRecord {
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
