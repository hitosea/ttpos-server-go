package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderSourceRepo 外卖来源仓库接口
//
// 任务: story-order-source-nationality Phase 2.2
// 需求: R1.1-R1.6
//
// @version v2.10.0
type IOrderSourceRepo interface {
	IOrderSourceQueryRepo
	Create(orderSource model.OrderSource) (uint64, error)
	Update(orderSource model.OrderSource) error
	SoftDelete(uuid uint64) error
}

// IOrderSourceQueryRepo 外卖来源查询接口
type IOrderSourceQueryRepo interface {
	FindList() ([]model.OrderSource, error)
	FindByUuid(uuid uint64) (*model.OrderSource, error)
	FindByUuidWithDeleted(uuid uint64) (*model.OrderSource, error)
	FindByUuidsWithDeleted(uuids []uint64) ([]*model.OrderSource, error) // 批量根据UUID查找外卖来源（包含已删除）
	CountOrdersBySourceUuid(uuid uint64) (int64, error)
	Count() (int64, error) // 获取外卖来源数量
}

// NewOrderSourceRepo 创建外卖来源仓库
func NewOrderSourceRepo(db *gorm.DB) IOrderSourceRepo {
	return NewOrderSourceRepoImpl(db)
}

// NewOrderSourceRepoImpl 创建外卖来源仓库实现
func NewOrderSourceRepoImpl(db *gorm.DB) *OrderSourceRepoImpl {
	return &OrderSourceRepoImpl{db: db}
}

// OrderSourceRepoImpl 外卖来源仓库实现
type OrderSourceRepoImpl struct {
	db *gorm.DB
}

// FindList 获取外卖来源列表（JOIN多语言表）
func (r *OrderSourceRepoImpl) FindList() ([]model.OrderSource, error) {
	var orderSources []model.OrderSource
	err := r.db.Model(&model.OrderSource{}).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Where("delete_time = ?", 0).
		Order("create_time DESC").
		Find(&orderSources).Error

	return orderSources, errors.WithMessage(err)
}

// FindByUuid 根据UUID查找外卖来源
func (r *OrderSourceRepoImpl) FindByUuid(uuid uint64) (*model.OrderSource, error) {
	var orderSource model.OrderSource
	err := r.db.Model(&model.OrderSource{}).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		First(&orderSource).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录返回 nil，不报错
		}
		return nil, errors.WithMessage(err)
	}

	return &orderSource, nil
}

// Create 创建外卖来源
func (r *OrderSourceRepoImpl) Create(orderSource model.OrderSource) (uint64, error) {
	if err := r.db.Model(&model.OrderSource{}).Create(&orderSource).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return orderSource.Uuid, nil
}

// Update 更新外卖来源
func (r *OrderSourceRepoImpl) Update(orderSource model.OrderSource) error {
	err := r.db.Model(&model.OrderSource{}).
		Where("uuid = ? AND delete_time = ?", orderSource.Uuid, 0).
		Updates(map[string]interface{}{
			"multi_language_name_uuid": orderSource.MultiLanguageNameUuid,
			"sort":                     orderSource.Sort,
			"status":                   orderSource.Status,
			"update_time":              orderSource.UpdateTime,
		}).Error

	return errors.WithMessage(err)
}

// SoftDelete 软删除外卖来源
func (r *OrderSourceRepoImpl) SoftDelete(uuid uint64) error {
	err := r.db.Model(&model.OrderSource{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error

	return errors.WithMessage(err)
}

// FindByUuidWithDeleted 根据UUID查找外卖来源（包含已删除）
// 用于订单详情查询，保证历史订单仍可显示已删除的配置名称
func (r *OrderSourceRepoImpl) FindByUuidWithDeleted(uuid uint64) (*model.OrderSource, error) {
	var orderSource model.OrderSource
	err := r.db.Model(&model.OrderSource{}).
		Preload("MultiLanguageName").
		Where("uuid = ?", uuid).
		First(&orderSource).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录返回 nil，不报错
		}
		return nil, errors.WithMessage(err)
	}

	return &orderSource, nil
}

// FindByUuidsWithDeleted 批量根据UUID查找外卖来源（包含已删除）
// 用于订单详情查询，保证历史订单仍可显示已删除的配置名称
func (r *OrderSourceRepoImpl) FindByUuidsWithDeleted(uuids []uint64) ([]*model.OrderSource, error) {
	if len(uuids) == 0 {
		return []*model.OrderSource{}, nil
	}

	var orderSources []*model.OrderSource
	err := r.db.Model(&model.OrderSource{}).
		Preload("MultiLanguageName").
		Where("uuid IN (?)", uuids).
		Find(&orderSources).Error

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return orderSources, nil
}

// Count 获取外卖来源数量
func (r *OrderSourceRepoImpl) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.OrderSource{}).Where("delete_time = 0").Count(&count).Error
	return count, errors.WithMessage(err)
}

// CountOrdersBySourceUuid 统计使用该外卖来源的订单数量
func (r *OrderSourceRepoImpl) CountOrdersBySourceUuid(uuid uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.SaleBill{}).
		Where("order_source_uuid = ? AND delete_time = ?", uuid, 0).
		Count(&count).Error

	return count, errors.WithMessage(err)
}
