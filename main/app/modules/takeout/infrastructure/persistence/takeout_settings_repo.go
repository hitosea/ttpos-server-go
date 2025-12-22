package persistence

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
)

// ITakeoutSettingsRepo 外卖配置仓储接口
type ITakeoutSettingsRepo interface {
	Create(settings *model.TakeoutSettings) error
	UpdateByMap(uuid uint64, data map[string]interface{}) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutSettings, error)
	GetByPlatform(platform string, options ...DBOption) (*model.TakeoutSettings, error)
	Delete(uuid uint64) error
}

// NewTakeoutSettingsRepo 创建外卖配置仓储
func NewTakeoutSettingsRepo(db *gorm.DB) ITakeoutSettingsRepo {
	return &TakeoutSettingsRepoImpl{db: db}
}

// TakeoutSettingsRepoImpl 外卖配置仓储实现
type TakeoutSettingsRepoImpl struct {
	db *gorm.DB
}

// Create 创建外卖配置
func (r *TakeoutSettingsRepoImpl) Create(settings *model.TakeoutSettings) error {
	return errors.WithMessage(r.db.Create(settings).Error)
}

// UpdateByMap 使用 map 更新外卖配置
func (r *TakeoutSettingsRepoImpl) UpdateByMap(uuid uint64, data map[string]interface{}) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutSettings{}).
			Where("uuid = ?", uuid).
			Updates(data).Error,
	)
}

// GetByUuid 根据UUID获取外卖配置
func (r *TakeoutSettingsRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutSettings, error) {
	var settings model.TakeoutSettings
	db := r.db.Model(&model.TakeoutSettings{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("uuid = ?", uuid).First(&settings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}

	return &settings, nil
}

// GetByPlatform 根据平台获取配置
func (r *TakeoutSettingsRepoImpl) GetByPlatform(platform string, options ...DBOption) (*model.TakeoutSettings, error) {
	var settings model.TakeoutSettings
	db := r.db.Model(&model.TakeoutSettings{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("platform = ?", platform).First(&settings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}

	return &settings, nil
}

// Delete 软删除外卖配置
func (r *TakeoutSettingsRepoImpl) Delete(uuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutSettings{}).
			Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}
