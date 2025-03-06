package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetDelayRepo 自助餐加钟价格
type IBuffetDelayRepo interface {
	GetBuffetDelayList() ([]model.BuffetDelay, error)
	GetBuffetDelayListByUuids(uuids []uint64) ([]*model.BuffetDelay, error) // 获取自助餐列表通过UUID列表，用于开台写入顾客数据
	UpdateBuffetDelay(uuid uint, buffetDelay model.BuffetDelay) error
	CreateBuffetDelay(buffetDelay model.BuffetDelay) (uint64, error)
	DeleteBuffetDelay(uuid uint) error
}

func NewBuffetDelayRepo(db *gorm.DB) IBuffetDelayRepo {
	return NewBuffetDelayRepoImpl(db)
}

// NewBuffetDelayRepoImpl 创建新的商品规格仓库实现
func NewBuffetDelayRepoImpl(db *gorm.DB) *BuffetDelayRepoImpl {
	return &BuffetDelayRepoImpl{db: db}
}

type BuffetDelayRepoImpl struct {
	db *gorm.DB
}

// GetBuffetDelayList 获取商品规格列表，排除逻辑删除的规格
func (r *BuffetDelayRepoImpl) GetBuffetDelayList() ([]model.BuffetDelay, error) {
	var buffetDelays []model.BuffetDelay
	err := r.db.Model(&model.BuffetDelay{}).Where("delete_time = ?", 0).Find(&buffetDelays).Error
	return buffetDelays, errors.WithMessage(err)
}

// GetBuffetDelayListByUuids 获取商品规格列表，排除逻辑删除的规格
func (r *BuffetDelayRepoImpl) GetBuffetDelayListByUuids(uuids []uint64) ([]*model.BuffetDelay, error) {
	var buffetDelays []*model.BuffetDelay
	err := r.db.Model(&model.BuffetDelay{}).Where("delete_time = ?", 0).Where("uuid in ?", uuids).Find(&buffetDelays).Error
	return buffetDelays, err
}

// UpdateBuffetDelay 更新自助餐加钟价格
func (r *BuffetDelayRepoImpl) UpdateBuffetDelay(uuid uint, buffetDelay model.BuffetDelay) error {
	if err := r.db.Model(&model.BuffetDelay{}).Where("uuid = ?", uuid).Updates(buffetDelay).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateBuffetDelay 创建自助餐加钟价格
func (r *BuffetDelayRepoImpl) CreateBuffetDelay(buffetDelay model.BuffetDelay) (uint64, error) {

	// 创建自助餐加钟价格
	if err := r.db.Create(&buffetDelay).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return buffetDelay.Uuid, nil
}

// DeleteBuffetDelay 软删除自助餐加钟价格
func (r *BuffetDelayRepoImpl) DeleteBuffetDelay(uuid uint) error {
	return r.db.Model(&model.BuffetDelay{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
