package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type ISettingRepo interface {
	GetAll(companyId uint) ([]model.Setting, error)
	Updates(companyId uint, key string, values string) error
}

func NewSettingRepo(dbm *database.DBManager) ISettingRepo {
	return NewSettingRepoImpl(dbm)
}

type SettingRepo struct {
	dbm *database.DBManager
}

func NewSettingRepoImpl(dbm *database.DBManager) *SettingRepo {
	return &SettingRepo{dbm: dbm}
}

func (r *SettingRepo) GetAll(companyId uint) ([]model.Setting, error) {
	var settings []model.Setting
	err := r.dbm.GetDB(companyId).Find(&settings).Error
	return settings, err
}

func (r *SettingRepo) Updates(companyId uint, key string, values string) error {
	return r.dbm.GetDB(companyId).Model(&model.Setting{}).Where("`key` = ?", key).Updates(map[string]any{"values": values}).Error
}
