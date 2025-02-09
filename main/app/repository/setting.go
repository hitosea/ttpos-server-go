package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISettingRepo interface {
	GetAll(companyId uint) ([]model.Setting, error)
	Updates(companyId uint, key string, values string) error
}

func NewSettingRepo(db *gorm.DB) ISettingRepo {
	return NewSettingRepoImpl(db)
}

type SettingRepo struct {
	db *gorm.DB
}

func NewSettingRepoImpl(db *gorm.DB) *SettingRepo {
	return &SettingRepo{db: db}
}

func (r *SettingRepo) GetAll(companyId uint) ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Find(&settings).Error
	return settings, err
}

func (r *SettingRepo) Updates(companyId uint, key string, values string) error {
	return r.db.Model(&model.Setting{}).Where("`key` = ?", key).Updates(map[string]any{"values": values}).Error
}
