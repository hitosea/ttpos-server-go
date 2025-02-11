package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISettingRepo interface {
	GetAll() ([]model.Setting, error)
	Updates(key string, values string) error
	GetByKey(key string) model.Setting
	Create(setting model.Setting) (model.Setting, error)
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

func (r *SettingRepo) GetAll() ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Find(&settings).Error
	return settings, err
}

func (r *SettingRepo) Updates(key string, values string) error {
	return r.db.Model(&model.Setting{}).Where("`key` = ?", key).Updates(map[string]any{"values": values}).Error
}

func (r *SettingRepo) GetByKey(key string) model.Setting {
	var setting model.Setting
	r.db.Model(&model.Setting{}).Where("`key` = ?", key).First(&setting)
	return setting
}

func (r *SettingRepo) Create(setting model.Setting) (model.Setting, error) {
	err := r.db.Create(&setting).Error
	return setting, err
}
