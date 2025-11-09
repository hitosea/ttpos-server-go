package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITransferOrderItemRepo 调拨单明细Repository接口
type ITransferOrderItemRepo interface {
	// 基础操作
	Create(item *model.TransferOrderItem) error
	CreateBatch(items []*model.TransferOrderItem) error
	Update(item *model.TransferOrderItem) error
	Delete(uuid uint64) error
	DeletePhysical(uuid uint64) error
	DeleteByTransferOrderUuid(transferOrderUuid uint64) error
	DeleteByTransferOrderUuidPhysical(transferOrderUuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.TransferOrderItem, error)

	// 查询操作
	GetList(opts ...DBOption) ([]model.TransferOrderItem, error)
	GetListByTransferOrderUuid(transferOrderUuid uint64, opts ...DBOption) ([]model.TransferOrderItem, error)
	Count(opts ...DBOption) (int64, error)

	// 条件查询选项
	WhereUuid(uuid uint64) DBOption
	WhereTransferOrderUuid(transferOrderUuid uint64) DBOption
	WhereMaterialUuid(materialUuid uint64) DBOption
	WhereCompanyUuid(companyUuid uint64) DBOption
	WhereHeadquarterUuid(headquarterUuid uint64) DBOption

	// 预加载
	WithUnits() DBOption

	// 统计
	CountByTransferOrderUuid(transferOrderUuid uint64) (int64, error)
}

// TransferOrderItemRepoImpl 调拨单明细Repository实现
type TransferOrderItemRepoImpl struct {
	db *gorm.DB
}

// NewTransferOrderItemRepo 创建调拨单明细Repository
func NewTransferOrderItemRepo(db *gorm.DB) ITransferOrderItemRepo {
	return &TransferOrderItemRepoImpl{db: db}
}

// Create 创建调拨单明细
func (r *TransferOrderItemRepoImpl) Create(item *model.TransferOrderItem) error {
	item.SetNil()
	return r.db.Create(item).Error
}

// CreateBatch 批量创建调拨单明细
func (r *TransferOrderItemRepoImpl) CreateBatch(items []*model.TransferOrderItem) error {
	if len(items) == 0 {
		return nil
	}
	for _, item := range items {
		item.SetNil()
	}
	return r.db.Create(items).Error
}

// Update 更新调拨单明细
func (r *TransferOrderItemRepoImpl) Update(item *model.TransferOrderItem) error {
	return r.db.Model(&model.TransferOrderItem{}).Where("uuid = ?", item.Uuid).Updates(item).Error
}

// Delete 删除调拨单明细（软删除）
func (r *TransferOrderItemRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.TransferOrderItem{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

// DeletePhysical 物理删除调拨单明细
func (r *TransferOrderItemRepoImpl) DeletePhysical(uuid uint64) error {
	return r.db.Model(&model.TransferOrderItem{}).Where("uuid = ?", uuid).Delete(&model.TransferOrderItem{}).Error
}

// DeleteByTransferOrderUuid 根据调拨单UUID删除所有明细（软删除）
func (r *TransferOrderItemRepoImpl) DeleteByTransferOrderUuid(transferOrderUuid uint64) error {
	return r.db.Model(&model.TransferOrderItem{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Update("delete_time", time.Now().Unix()).Error
}

// DeleteByTransferOrderUuidPhysical 根据调拨单UUID删除所有明细（物理删除）
func (r *TransferOrderItemRepoImpl) DeleteByTransferOrderUuidPhysical(transferOrderUuid uint64) error {
	return r.db.Model(&model.TransferOrderItem{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Delete(&model.TransferOrderItem{}).Error
}

// GetByUuid 根据UUID获取调拨单明细
func (r *TransferOrderItemRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.TransferOrderItem, error) {
	var item model.TransferOrderItem
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrderItem{}).
		Where("uuid = ?", uuid).
		Where("delete_time = ?", constant.NotDeleted).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetList 获取调拨单明细列表
func (r *TransferOrderItemRepoImpl) GetList(opts ...DBOption) ([]model.TransferOrderItem, error) {
	var items []model.TransferOrderItem
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrderItem{}).
		Where("delete_time = ?", constant.NotDeleted).
		Find(&items).Error
	return items, err
}

// GetListByTransferOrderUuid 根据调拨单UUID获取明细列表
func (r *TransferOrderItemRepoImpl) GetListByTransferOrderUuid(transferOrderUuid uint64, opts ...DBOption) ([]model.TransferOrderItem, error) {
	var items []model.TransferOrderItem
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrderItem{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Where("delete_time = ?", constant.NotDeleted).
		Find(&items).Error
	return items, err
}

// Count 统计调拨单明细数量
func (r *TransferOrderItemRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrderItem{}).
		Where("delete_time = ?", constant.NotDeleted).
		Count(&count).Error
	return count, err
}

// CountByTransferOrderUuid 根据调拨单UUID统计明细数量
func (r *TransferOrderItemRepoImpl) CountByTransferOrderUuid(transferOrderUuid uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.TransferOrderItem{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Where("delete_time = ?", constant.NotDeleted).
		Count(&count).Error
	return count, err
}

// 条件查询选项实现

// WhereUuid UUID条件
func (r *TransferOrderItemRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereTransferOrderUuid 调拨单UUID条件
func (r *TransferOrderItemRepoImpl) WhereTransferOrderUuid(transferOrderUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if transferOrderUuid > 0 {
			return db.Where("transfer_order_uuid = ?", transferOrderUuid)
		}
		return db
	}
}

// WhereMaterialUuid 物品UUID条件
func (r *TransferOrderItemRepoImpl) WhereMaterialUuid(materialUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if materialUuid > 0 {
			return db.Where("material_uuid = ?", materialUuid)
		}
		return db
	}
}

// WhereCompanyUuid 公司UUID条件
func (r *TransferOrderItemRepoImpl) WhereCompanyUuid(companyUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if companyUuid > 0 {
			return db.Where("company_uuid = ?", companyUuid)
		}
		return db
	}
}

// WhereHeadquarterUuid 总部UUID条件
func (r *TransferOrderItemRepoImpl) WhereHeadquarterUuid(headquarterUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if headquarterUuid > 0 {
			return db.Where("headquarter_uuid = ?", headquarterUuid)
		}
		return db
	}
}

// WithUnits 预加载单位信息
func (r *TransferOrderItemRepoImpl) WithUnits() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Units", "delete_time = ?", constant.NotDeleted)
	}
}

// applyOptions 应用查询选项
func (r *TransferOrderItemRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}
