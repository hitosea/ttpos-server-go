package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetDelayRepo 自助餐加钟价格
type IBuffetDelayRepo interface {
	GetBuffetDelayList() ([]model.BuffetDelay, error)
	UpdateBuffetDelay(uuid uint, buffetDelay model.BuffetDelay) error
	CreateBuffetDelay(buffetDelay model.BuffetDelay) (uint64, error)
	DeleteBuffetDelay(uuid uint) error
}

func NewBuffetDelayRepo(db *gorm.DB) IBuffetDelayRepo {
	return NewBuffetDelayRepoImpl(db)
}

// NewProductFlavorRepoImpl 创建新的商品规格仓库实现
func NewBuffetDelayRepoImpl(db *gorm.DB) *BuffetDelayRepoImpl {
	return &BuffetDelayRepoImpl{db: db}
}

type BuffetDelayRepoImpl struct {
	db *gorm.DB
}

// GetProductFlavorList 获取商品规格列表，排除逻辑删除的规格
func (r *BuffetDelayRepoImpl) GetBuffetDelayList() ([]model.BuffetDelay, error) {
	var buffetDelays []model.BuffetDelay
	err := r.db.Model(&model.BuffetDelay{}).Where("delete_time = ?", 0).Find(&buffetDelays).Error
	return buffetDelays, err
}

// UpdateBuffetDelay 更新自助餐加钟价格
func (r *BuffetDelayRepoImpl) UpdateBuffetDelay(uuid uint, buffetDelay model.BuffetDelay) error {
	if err := r.db.Model(&model.BuffetDelay{}).Where("uuid = ?", uuid).Updates(buffetDelay).Error; err != nil {
		return err
	}
	return nil
}

// CreateBuffetDelay 创建自助餐加钟价格
func (r *BuffetDelayRepoImpl) CreateBuffetDelay(buffetDelay model.BuffetDelay) (uint64, error) {

	// 创建自助餐加钟价格
	if err := r.db.Create(&buffetDelay).Error; err != nil {
		return 0, err
	}
	return buffetDelay.Uuid, nil
}

// DeleteBuffetDelay 软删除自助餐加钟价格
func (r *BuffetDelayRepoImpl) DeleteBuffetDelay(uuid uint) error {
	return r.db.Model(&model.BuffetDelay{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
