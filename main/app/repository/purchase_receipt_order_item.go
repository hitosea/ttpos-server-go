package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseReceiptOrderItemRepo 采购收货明细Repository接口
type IPurchaseReceiptOrderItemRepo interface {
	// 基础操作
	Create(item *model.PurchaseReceiptOrderItem) error
	CreateBatch(items []model.PurchaseReceiptOrderItem) error
	Update(item *model.PurchaseReceiptOrderItem) error
	Delete(uuid uint64) error
	DeleteByReceiptOrderUuid(receiptOrderUuid uint64) error
	GetByUuid(uuid uint64) (*model.PurchaseReceiptOrderItem, error)

	// 查询操作
	GetByPurchaseReceiptUuid(purchaseReceiptUuid uint64) ([]model.PurchaseReceiptOrderItem, error)
	GetList(opts ...DBOption) ([]model.PurchaseReceiptOrderItem, error)
	Count(opts ...DBOption) (int64, error)

	// 条件查询选项
	WhereUuid(uuid uint64) DBOption
	WherePurchaseReceiptUuid(purchaseReceiptUuid uint64) DBOption
	WherePurchaseOrderItemUuid(purchaseOrderItemUuid uint64) DBOption
	WhereProductUuid(productUuid uint64) DBOption
	WhereQualityStatus(qualityStatus int) DBOption
}

// PurchaseReceiptOrderItemRepoImpl 采购收货明细Repository实现
type PurchaseReceiptOrderItemRepoImpl struct {
	db *gorm.DB
}

// NewPurchaseReceiptItemRepo 创建采购收货明细Repository
func NewPurchaseReceiptOrderItemRepo(db *gorm.DB) IPurchaseReceiptOrderItemRepo {
	return &PurchaseReceiptOrderItemRepoImpl{db: db}
}

// Create 创建收货明细
func (r *PurchaseReceiptOrderItemRepoImpl) Create(item *model.PurchaseReceiptOrderItem) error {
	return r.db.Create(item).Error
}

// CreateBatch 批量创建收货明细
func (r *PurchaseReceiptOrderItemRepoImpl) CreateBatch(items []model.PurchaseReceiptOrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.CreateInBatches(items, 100).Error
}

// Update 更新收货明细
func (r *PurchaseReceiptOrderItemRepoImpl) Update(item *model.PurchaseReceiptOrderItem) error {
	return r.db.Where("uuid = ?", item.Uuid).Updates(item).Error
}

// Delete 删除收货明细
func (r *PurchaseReceiptOrderItemRepoImpl) Delete(uuid uint64) error {
	return r.db.Where("uuid = ?", uuid).Delete(&model.PurchaseReceiptOrderItem{}).Error
}

// DeleteByReceiptOrderUuid 根据收货单UUID删除明细
func (r *PurchaseReceiptOrderItemRepoImpl) DeleteByReceiptOrderUuid(receiptOrderUuid uint64) error {
	return r.db.Where("receipt_order_uuid = ?", receiptOrderUuid).Delete(&model.PurchaseReceiptOrderItem{}).Error
}

// GetByUuid 根据UUID获取收货明细
func (r *PurchaseReceiptOrderItemRepoImpl) GetByUuid(uuid uint64) (*model.PurchaseReceiptOrderItem, error) {
	var item model.PurchaseReceiptOrderItem
	err := r.db.Where("uuid = ?", uuid).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetByPurchaseReceiptUuid 根据收货记录UUID获取明细列表
func (r *PurchaseReceiptOrderItemRepoImpl) GetByPurchaseReceiptUuid(purchaseReceiptUuid uint64) ([]model.PurchaseReceiptOrderItem, error) {
	var items []model.PurchaseReceiptOrderItem
	err := r.db.Where("purchase_receipt_uuid = ?", purchaseReceiptUuid).Find(&items).Error
	return items, err
}

// GetList 获取收货明细列表
func (r *PurchaseReceiptOrderItemRepoImpl) GetList(opts ...DBOption) ([]model.PurchaseReceiptOrderItem, error) {
	var items []model.PurchaseReceiptOrderItem
	db := r.applyOptions(r.db, opts...)
	err := db.Find(&items).Error
	return items, err
}

// Count 统计收货明细数量
func (r *PurchaseReceiptOrderItemRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseReceiptOrderItem{}).Count(&count).Error
	return count, err
}

// 条件查询选项实现

// WhereUuid UUID条件
func (r *PurchaseReceiptOrderItemRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WherePurchaseReceiptUuid 收货记录UUID条件
func (r *PurchaseReceiptOrderItemRepoImpl) WherePurchaseReceiptUuid(purchaseReceiptUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("purchase_receipt_uuid = ?", purchaseReceiptUuid)
	}
}

// WherePurchaseOrderItemUuid 采购订单明细UUID条件
func (r *PurchaseReceiptOrderItemRepoImpl) WherePurchaseOrderItemUuid(purchaseOrderItemUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("purchase_order_item_uuid = ?", purchaseOrderItemUuid)
	}
}

// WhereProductUuid 商品UUID条件
func (r *PurchaseReceiptOrderItemRepoImpl) WhereProductUuid(productUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_uuid = ?", productUuid)
	}
}

// WhereQualityStatus 质检状态条件
func (r *PurchaseReceiptOrderItemRepoImpl) WhereQualityStatus(qualityStatus int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("quality_status = ?", qualityStatus)
	}
}

// applyOptions 应用查询选项
func (r *PurchaseReceiptOrderItemRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}
