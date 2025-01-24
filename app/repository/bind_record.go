package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type BindRecordRepository struct {
	dbm *database.DBManager
}

func NewBindRecordRepository(dbm *database.DBManager) *BindRecordRepository {
	return &BindRecordRepository{dbm: dbm}
}

func (r *BindRecordRepository) Unbind(appId uint, source string, key string, shopUserId uint) error {
	return r.dbm.GetDB(appId).Model(&model.BindRecord{}).Select("finally_login_id").
		Where("`source` = ? AND `key` = ? AND `finally_login_id` = ?", source, key, shopUserId).Debug().
		Updates(map[string]interface{}{
			"finally_login_id": 0,
		}).Error
}

func (r *BindRecordRepository) GetBindCount(companyId uint, source string) uint {
	var c int64
	r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("source = ? AND finally_login_id > 0", source).Count(&c)
	return uint(c)
}

func (r *BindRecordRepository) GetRecordBySourceAndKey(source string, key string) model.BindRecord {
	var bindRecord model.BindRecord
	r.dbm.GetDB(constant.DefaultDB).Model(&model.BindRecord{}).Where("source = ? AND key = ?", source, key).First(&bindRecord)
	return bindRecord
}

func (r *BindRecordRepository) Update(companyId uint, id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyId).Model(&model.BindRecord{}).Where("id = ?", id).Updates(vars).Error
}
