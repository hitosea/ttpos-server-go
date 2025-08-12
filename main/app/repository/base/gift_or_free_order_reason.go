package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

// IGiftOrFreeOrderReasonRepo 赠品或免费订单原因仓库接口
type IGiftOrFreeOrderReasonRepo interface {
	GetGiftOrFreeOrderReasonList() ([]model.FreeReason, error)                           // 获取赠品或免费订单原因列表
	UpdateGiftOrFreeOrderReason(id uint64, giftOrFreeOrderReason model.FreeReason) error // 更新赠品或免费订单原因
	CreateGiftOrFreeOrderReason(giftOrFreeOrderReason model.FreeReason) (uint64, error)  // 创建赠品或免费订单原因
	DeleteGiftOrFreeOrderReason(id uint64) error                                         // 删除赠品或免费订单原因
	DeleteGiftOrFreeOrderReasons(uuids []uint64) error                                   // 批量删除赠品或免费订单原因
	ExistsByUuids(uuids []uint64) ([][2]uint64, []uint64, error)                         // 根据uuid数组验证赠品或免费订单原因是否存在，返回[uuid, 多语言名称UUID]数组和不存在的UUID列表
	GetFreeOrderReasons(opts ...repository.DBOption) ([]*model.FreeReason, error)        // 获取赠品或免单原因列表
	GetFreeOrderReasonListByUuids(uuids []uint64) ([]*model.FreeReason, error)
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
	return giftOrFreeOrderReasons, errors.WithMessage(err)
}

// UpdateGiftOrFreeOrderReason 更新赠品或免费订单原因
func (r *GiftOrFreeOrderReasonRepoImpl) UpdateGiftOrFreeOrderReason(id uint64, giftOrFreeOrderReason model.FreeReason) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.FreeReason{}).Where("id = ?", id).Updates(giftOrFreeOrderReason).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&giftOrFreeOrderReason.MultiLanguageName).Where("id = ?", giftOrFreeOrderReason.MultiLanguageNameUuid).Updates(giftOrFreeOrderReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// CreateGiftOrFreeOrderReason 创建赠品或免费订单原因
func (r *GiftOrFreeOrderReasonRepoImpl) CreateGiftOrFreeOrderReason(giftOrFreeOrderReason model.FreeReason) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&giftOrFreeOrderReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	// 创建赠品或免费订单原因
	if err := tx.Create(&giftOrFreeOrderReason).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	return giftOrFreeOrderReason.Uuid, tx.Commit().Error // 提交事务
}

// DeleteGiftOrFreeOrderReason 软删除赠品或免费订单原因
func (r *GiftOrFreeOrderReasonRepoImpl) DeleteGiftOrFreeOrderReason(id uint64) error {
	return r.db.Model(&model.FreeReason{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

// ExistsByUuids 根据uuid数组验证赠品或免费订单原因是否存在，返回[uuid, 多语言名称UUID]数组和不存在的UUID列表
func (r *GiftOrFreeOrderReasonRepoImpl) ExistsByUuids(uuids []uint64) ([][2]uint64, []uint64, error) {
	if len(uuids) == 0 {
		return [][2]uint64{}, []uint64{}, nil
	}

	var reasons []model.FreeReason
	err := r.db.Where("uuid IN ? AND delete_time = 0", uuids).Find(&reasons).Error
	if err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	// 创建存在的UUID集合
	existMap := make(map[uint64]model.FreeReason)
	for _, reason := range reasons {
		existMap[reason.Uuid] = reason
	}

	// 找出不存在的UUID
	notFound := make([]uint64, 0)
	for _, uuid := range uuids {
		if _, ok := existMap[uuid]; !ok {
			notFound = append(notFound, uuid)
		}
	}

	// 创建结果数组，每个元素是[uuid, 多语言名称UUID]
	result := make([][2]uint64, len(reasons))
	for i, reason := range reasons {
		result[i] = [2]uint64{reason.Uuid, reason.MultiLanguageNameUuid}
	}

	return result, notFound, nil
}

func (r *GiftOrFreeOrderReasonRepoImpl) GetFreeOrderReasons(opts ...repository.DBOption) ([]*model.FreeReason, error) {
	freeReasons := make([]*model.FreeReason, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&freeReasons)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return freeReasons, nil
}

func (r *GiftOrFreeOrderReasonRepoImpl) GetFreeOrderReasonListByUuids(uuids []uint64) ([]*model.FreeReason, error) {
	reasons, err := r.GetFreeOrderReasons(
		repository.CommonRepo.WhereInUuids(uuids),
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.SortWithCreateTime("desc"),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return reasons, nil
}

func (r *GiftOrFreeOrderReasonRepoImpl) DeleteGiftOrFreeOrderReasons(uuids []uint64) error {
	return r.db.Model(&model.FreeReason{}).Where("uuid IN (?)", uuids).Update("delete_time", uint(time.Now().Unix())).Error
}
