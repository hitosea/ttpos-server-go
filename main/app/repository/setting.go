package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type SettingRepository struct {
	dbm *database.DBManager
}

func NewSettingRepository(dbm *database.DBManager) *SettingRepository {
	return &SettingRepository{dbm: dbm}
}

func (r *SettingRepository) GetAll(companyId uint) ([]model.Setting, error) {
	var settings []model.Setting
	err := r.dbm.GetDB(companyId).Find(&settings).Error
	return settings, err
}

func (r *SettingRepository) Updates(companyId uint, key string, values string) error {
	return r.dbm.GetDB(companyId).Model(&model.Setting{}).Where("`key` = ?", key).Updates(map[string]any{"values": values}).Error
}
