package persistence

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ITakeoutOrderItemRepo 外卖订单商品仓储接口
type ITakeoutOrderItemRepo interface {
	Create(item *model.TakeoutOrderItem) error
	BatchCreate(items []*model.TakeoutOrderItem) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutOrderItem, error)
	GetByOrderUuid(orderUuid uint64, options ...DBOption) ([]*model.TakeoutOrderItem, error)
	Delete(uuid uint64) error
	UpdateQuantity(uuid uint64, quantity int) error
	UpdateByMap(uuid uint64, fields map[string]interface{}) error
	// 选项方法
	WithModifiers() DBOption
}

// NewTakeoutOrderItemRepo 创建外卖订单商品仓储
func NewTakeoutOrderItemRepo(db *gorm.DB) ITakeoutOrderItemRepo {
	return &TakeoutOrderItemRepoImpl{db: db}
}

// TakeoutOrderItemRepoImpl 外卖订单商品仓储实现
type TakeoutOrderItemRepoImpl struct {
	db *gorm.DB
}

// Create 创建外卖订单商品
func (r *TakeoutOrderItemRepoImpl) Create(item *model.TakeoutOrderItem) error {
	return errors.WithMessage(r.db.Model(&model.TakeoutOrderItem{}).Omit(clause.Associations).Create(&item).Error)
}

// BatchCreate 批量创建外卖订单商品
func (r *TakeoutOrderItemRepoImpl) BatchCreate(items []*model.TakeoutOrderItem) error {
	return errors.WithMessage(r.db.Model(&model.TakeoutOrderItem{}).Create(&items).Error)
}

// GetByUuid 根据UUID获取外卖订单商品
func (r *TakeoutOrderItemRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutOrderItem, error) {
	var item model.TakeoutOrderItem
	db := r.db.Model(&model.TakeoutOrderItem{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("uuid = ?", uuid).First(&item).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}

	return &item, nil
}

// GetByOrderUuid 根据订单UUID获取商品列表
func (r *TakeoutOrderItemRepoImpl) GetByOrderUuid(orderUuid uint64, options ...DBOption) ([]*model.TakeoutOrderItem, error) {
	var items []*model.TakeoutOrderItem
	db := r.db.Model(&model.TakeoutOrderItem{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("takeout_order_uuid = ?", orderUuid).Order("id ASC").Find(&items).Error

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return items, nil
}

// Delete 软删除外卖订单商品
func (r *TakeoutOrderItemRepoImpl) Delete(uuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutOrderItem{}).
			Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

// UpdateQuantity 更新商品数量
func (r *TakeoutOrderItemRepoImpl) UpdateQuantity(uuid uint64, quantity int) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutOrderItem{}).
			Where("uuid = ?", uuid).
			Update("quantity", quantity).Error,
	)
}

// UpdateByMap 根据 Map 更新商品字段
func (r *TakeoutOrderItemRepoImpl) UpdateByMap(uuid uint64, fields map[string]interface{}) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutOrderItem{}).
			Where("uuid = ?", uuid).
			Updates(fields).Error,
	)
}

// WithModifiers 预加载修饰符
func (r *TakeoutOrderItemRepoImpl) WithModifiers() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderItemModifiers", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0)
		})
	}
}
