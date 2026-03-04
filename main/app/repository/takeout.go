package repository

import (
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
)

type ITakeoutRepo interface {
	GetEnabledTakeouts() ([]takeoutModel.Takeout, error) // 获取已启用的外卖平台
}

func NewTakeoutRepo(db *gorm.DB) ITakeoutRepo {
	return &takeoutRepo{db: db}
}

type takeoutRepo struct {
	db *gorm.DB
}

// GetEnabledTakeouts 获取已启用的外卖平台
func (r *takeoutRepo) GetEnabledTakeouts() ([]takeoutModel.Takeout, error) {
	var takeouts []takeoutModel.Takeout
	err := r.db.Model(&takeoutModel.Takeout{}).Where("enabled = 1").Find(&takeouts).Error
	return takeouts, err
}
