package persistence

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
)

// ITakeoutOrderItemModifierRepo 外卖订单商品修饰符仓储接口
type ITakeoutOrderItemModifierRepo interface {
	Create(modifier *model.TakeoutOrderItemModifier) error
	BatchCreate(modifiers []*model.TakeoutOrderItemModifier) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutOrderItemModifier, error)
	GetByOrderItemUuid(orderItemUuid uint64, options ...DBOption) ([]*model.TakeoutOrderItemModifier, error)
	Delete(uuid uint64) error
	DeleteByItemUuid(itemUuid uint64) error
}

// NewTakeoutOrderItemModifierRepo 创建外卖订单商品修饰符仓储
func NewTakeoutOrderItemModifierRepo(db *gorm.DB) ITakeoutOrderItemModifierRepo {
	return &TakeoutOrderItemModifierRepoImpl{db: db}
}

// TakeoutOrderItemModifierRepoImpl 外卖订单商品修饰符仓储实现
type TakeoutOrderItemModifierRepoImpl struct {
	db *gorm.DB
}

// Create 创建外卖订单商品修饰符
func (r *TakeoutOrderItemModifierRepoImpl) Create(modifier *model.TakeoutOrderItemModifier) error {
	return errors.WithMessage(r.db.Create(modifier).Error)
}

// BatchCreate 批量创建外卖订单商品修饰符
func (r *TakeoutOrderItemModifierRepoImpl) BatchCreate(modifiers []*model.TakeoutOrderItemModifier) error {
	if len(modifiers) == 0 {
		return nil
	}
	return errors.WithMessage(r.db.Create(&modifiers).Error)
}

// GetByUuid 根据UUID获取外卖订单商品修饰符
func (r *TakeoutOrderItemModifierRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutOrderItemModifier, error) {
	var modifier model.TakeoutOrderItemModifier
	db := r.db.Model(&model.TakeoutOrderItemModifier{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("uuid = ?", uuid).First(&modifier).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}

	return &modifier, nil
}

// GetByOrderItemUuid 根据订单商品UUID获取修饰符列表
func (r *TakeoutOrderItemModifierRepoImpl) GetByOrderItemUuid(orderItemUuid uint64, options ...DBOption) ([]*model.TakeoutOrderItemModifier, error) {
	var modifiers []*model.TakeoutOrderItemModifier
	db := r.db.Model(&model.TakeoutOrderItemModifier{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("takeout_order_item_uuid = ?", orderItemUuid).Order("id ASC").Find(&modifiers).Error

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return modifiers, nil
}

// Delete 软删除外卖订单商品修饰符
func (r *TakeoutOrderItemModifierRepoImpl) Delete(uuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutOrderItemModifier{}).
			Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

// DeleteByItemUuid 根据商品UUID删除所有修饰符（硬删除）
func (r *TakeoutOrderItemModifierRepoImpl) DeleteByItemUuid(itemUuid uint64) error {
	return errors.WithMessage(
		r.db.Where("takeout_order_item_uuid = ?", itemUuid).
			Delete(&model.TakeoutOrderItemModifier{}).Error,
	)
}
