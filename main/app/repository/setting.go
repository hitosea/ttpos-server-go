package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISettingRepo interface {
	GetAll() ([]model.Setting, error)
	Updates(key string, values string) error
	GetByKey(key string) model.Setting
	GetByKeyNotDeleted(key string) model.Setting // 带 delete_time = 0 过滤
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
	err := r.db.Model(&model.Setting{}).Find(&settings).Error
	return settings, errors.WithMessage(err)
}

func (r *SettingRepo) Updates(key string, values string) error {
	return r.db.Model(&model.Setting{}).Where("`key` = ?", key).Updates(map[string]any{"values": values}).Error
}

func (r *SettingRepo) GetByKey(key string) model.Setting {
	var setting model.Setting
	r.db.Model(&model.Setting{}).Where("`key` = ?", key).First(&setting)
	return setting
}

func (r *SettingRepo) GetByKeyNotDeleted(key string) model.Setting {
	var setting model.Setting
	r.db.Model(&model.Setting{}).Where("`key` = ? AND delete_time = 0", key).First(&setting)
	return setting
}

func (r *SettingRepo) Create(setting model.Setting) (model.Setting, error) {
	err := r.db.Model(&model.Setting{}).Create(&setting).Error
	return setting, errors.WithMessage(err)
}
