package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IGiftOrFreeOrderReasonRepo 赠品或免费订单原因仓库接口
type IGiftOrFreeOrderReasonRepo interface {
	GetGiftOrFreeOrderReasonList() ([]model.FreeReason, error)                         // 获取赠品或免费订单原因列表
	UpdateGiftOrFreeOrderReason(id uint, giftOrFreeOrderReason model.FreeReason) error // 更新赠品或免费订单原因
	CreateGiftOrFreeOrderReason(giftOrFreeOrderReason model.FreeReason) (uint, error)  // 创建赠品或免费订单原因
	DeleteGiftOrFreeOrderReason(id uint) error                                         // 删除赠品或免费订单原因
}

// NewGiftOrFreeOrderReasonRepo 创建新的赠品或免费订单原因仓库
func NewGiftOrFreeOrderReasonRepo(db *gorm.DB) IGiftOrFreeOrderReasonRepo {
	return NewGiftOrFreeOrderReasonRepoImpl(db)
}

// NewGiftOrFreeOrderReasonRepoImpl 创建新的赠品或免费订单原因仓库实现
func NewGiftOrFreeOrderReasonRepoImpl(db *gorm.DB) *GiftOrFreeOrderReasonRepoImpl {
	return &GiftOrFreeOrderReasonRepoImpl{db: db}
}

type GiftOrFreeOrderReasonRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetGiftOrFreeOrderReasonList 获取赠品或免费订单原因列表，排除逻辑删除的原因
func (r *GiftOrFreeOrderReasonRepoImpl) GetGiftOrFreeOrderReasonList() ([]model.FreeReason, error) {
	var giftOrFreeOrderReasons []model.FreeReason
	err := r.db.Model(&model.FreeReason{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&giftOrFreeOrderReasons).Error
	return giftOrFreeOrderReasons, err
}

// UpdateGiftOrFreeOrderReason 更新赠品或免费订单原因
func (r *GiftOrFreeOrderReasonRepoImpl) UpdateGiftOrFreeOrderReason(id uint, giftOrFreeOrderReason model.FreeReason) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.FreeReason{}).Where("id = ?", id).Updates(giftOrFreeOrderReason).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return err
	}

	if err := tx.Model(&giftOrFreeOrderReason.MultiLanguageName).Where("id = ?", giftOrFreeOrderReason.MultiLanguageNameUuid).Updates(giftOrFreeOrderReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return err
	}

	return tx.Commit().Error // 提交事务
}

// CreateGiftOrFreeOrderReason 创建赠品或免费订单原因
func (r *GiftOrFreeOrderReasonRepoImpl) CreateGiftOrFreeOrderReason(giftOrFreeOrderReason model.FreeReason) (uint, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&giftOrFreeOrderReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, err
	}

	// 创建赠品或免费订单原因
	if err := tx.Create(&giftOrFreeOrderReason).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, err
	}

	return giftOrFreeOrderReason.Uuid, tx.Commit().Error // 提交事务
}

// DeleteGiftOrFreeOrderReason 软删除赠品或免费订单原因
func (r *GiftOrFreeOrderReasonRepoImpl) DeleteGiftOrFreeOrderReason(id uint) error {
	return r.db.Model(&model.FreeReason{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
