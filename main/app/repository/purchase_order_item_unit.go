package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseOrderItemUnitRepo 采购订单明细单位Repository接口
type IPurchaseOrderItemUnitRepo interface {
	CreateBatch(items []model.PurchaseOrderItemUnit) error
	Update(item model.PurchaseOrderItemUnit) error
	DeleteByItemUuids(itemUuids []uint64) error
	GetByUuid(uuid uint64) (*model.PurchaseOrderItemUnit, error)
	GetList(opts ...DBOption) ([]model.PurchaseOrderItemUnit, error)
	WhereUuid(uuid uint64) DBOption
}

// PurchaseOrderItemUnitRepoImpl 采购订单明细单位Repository实现
type PurchaseOrderItemUnitRepoImpl struct {
	db *gorm.DB
}

// NewPurchaseOrderItemUnitRepo 创建采购订单明细单位Repository
func NewPurchaseOrderItemUnitRepo(db *gorm.DB) IPurchaseOrderItemUnitRepo {
	return &PurchaseOrderItemUnitRepoImpl{db: db}
}

// CreateBatch 批量创建采购订单明细单位
func (r *PurchaseOrderItemUnitRepoImpl) CreateBatch(items []model.PurchaseOrderItemUnit) error {
	return r.db.CreateInBatches(items, 100).Error
}

// Update 更新采购订单明细单位
func (r *PurchaseOrderItemUnitRepoImpl) Update(item model.PurchaseOrderItemUnit) error {
	return r.db.Where("uuid = ?", item.Uuid).Updates(item).Error
}

// DeleteByItemUuids 根据明细UUID批量删除单位
func (r *PurchaseOrderItemUnitRepoImpl) DeleteByItemUuids(itemUuids []uint64) error {
	if len(itemUuids) == 0 {
		return nil
	}
	return r.db.Where("item_uuid IN ?", itemUuids).Delete(&model.PurchaseOrderItemUnit{}).Error
}

// GetByUuid 根据UUID获取采购订单明细单位
func (r *PurchaseOrderItemUnitRepoImpl) GetByUuid(uuid uint64) (*model.PurchaseOrderItemUnit, error) {
	var item model.PurchaseOrderItemUnit
	err := r.db.Where("uuid = ?", uuid).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetList 获取采购订单明细列表
func (r *PurchaseOrderItemUnitRepoImpl) GetList(opts ...DBOption) ([]model.PurchaseOrderItemUnit, error) {
	var items []model.PurchaseOrderItemUnit
	db := r.applyOptions(r.db, opts...)
	err := db.Find(&items).Error
	return items, err
}

// 条件查询选项实现

// WhereUuid UUID条件
func (r *PurchaseOrderItemUnitRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// applyOptions 应用查询选项
func (r *PurchaseOrderItemUnitRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}
