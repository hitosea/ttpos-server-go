package repository

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseOrderItemRepo 采购订单明细Repository接口
type IPurchaseOrderItemRepo interface {
	// 基础操作
	Create(item *model.PurchaseOrderItem) error
	CreateBatch(items []model.PurchaseOrderItem) error
	Update(item *model.PurchaseOrderItem) error
	Delete(uuid uint64) error
	DeleteByPurchaseOrderUuid(purchaseOrderUuid uint64) error
	GetNotReceivedQuantityByMaterialUuid(materialUuid uint64) (int64, error)
	DeleteByPurchaseOrderUuidAndNumIsZero(purchaseOrderUuid uint64) error
	DeleteByPurchaseOrderUuidAndMaterialUuids(purchaseOrderUuid uint64, materialUuids []uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.PurchaseOrderItem, error)
	GetByPurchaseOrderUuidAndMaterialCode(subUuid uint64, materialCode string) (*model.PurchaseOrderItem, error)

	// 查询操作
	GetByPurchaseOrderUuid(purchaseOrderUuid uint64, opts ...DBOption) ([]model.PurchaseOrderItem, error)
	GetList(opts ...DBOption) ([]model.PurchaseOrderItem, error)
	Count(opts ...DBOption) (int64, error)

	// 条件查询选项
	WhereUuid(uuid uint64) DBOption
	WherePurchaseOrderUuid(purchaseOrderUuid uint64) DBOption
	WhereProductUuid(productUuid uint64) DBOption
	WithPreloadUnits() DBOption
	OrderBySort() DBOption
	OrderByCreateTime(desc bool) DBOption

	// 统计查询
	GetTotalQuantityByPurchaseOrder(purchaseOrderUuid uint64) (float64, error)
	GetTotalAmountByPurchaseOrder(purchaseOrderUuid uint64) (float64, error)
	GetReceivedQuantityByPurchaseOrder(purchaseOrderUuid uint64) (float64, error)

	// 品牌采购限购相关统计
	// SumBrandPurchaseByTimeRange 统计指定时间范围内的品牌采购物品数量（按物品+单位+门店）
	SumBrandPurchaseByTimeRange(
		materialCode string,
		unitCode string,
		startTime, endTime int64,
		excludeOrderUuid uint64,
	) (float64, error)
	// SumBrandPurchaseByMonth 统计指定月份的品牌采购物品数量（按物品+单位+门店）
	SumBrandPurchaseByMonth(
		materialCode string,
		unitCode string,
		month string,
		excludeOrderUuid uint64,
	) (float64, error)
}

// PurchaseOrderItemRepoImpl 采购订单明细Repository实现
type PurchaseOrderItemRepoImpl struct {
	db *gorm.DB
}

// NewPurchaseOrderItemRepo 创建采购订单明细Repository
func NewPurchaseOrderItemRepo(db *gorm.DB) IPurchaseOrderItemRepo {
	return &PurchaseOrderItemRepoImpl{db: db}
}

// Create 创建采购订单明细
func (r *PurchaseOrderItemRepoImpl) Create(item *model.PurchaseOrderItem) error {
	return r.db.Create(item).Error
}

// CreateBatch 批量创建采购订单明细
func (r *PurchaseOrderItemRepoImpl) CreateBatch(items []model.PurchaseOrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.CreateInBatches(items, 100).Error
}

// Update 更新采购订单明细
func (r *PurchaseOrderItemRepoImpl) Update(item *model.PurchaseOrderItem) error {
	return r.db.Where("uuid = ?", item.Uuid).Updates(item).Error
}

// Delete 删除采购订单明细
func (r *PurchaseOrderItemRepoImpl) Delete(uuid uint64) error {
	return r.db.Where("uuid = ?", uuid).Delete(&model.PurchaseOrderItem{}).Error
}

// DeleteByPurchaseOrderUuid 根据采购订单UUID删除明细
func (r *PurchaseOrderItemRepoImpl) DeleteByPurchaseOrderUuid(purchaseOrderUuid uint64) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("purchase_order_uuid = ?", purchaseOrderUuid).Delete(&model.PurchaseOrderItem{}).Error
		if err != nil {
			return err
		}
		return tx.Where("purchase_order_uuid = ?", purchaseOrderUuid).Delete(&model.PurchaseOrderItemUnit{}).Error
	})
	return err
}

// DeleteByPurchaseOrderUuidAndNumIsZero 根据采购订单UUID删除明细
func (r *PurchaseOrderItemRepoImpl) DeleteByPurchaseOrderUuidAndNumIsZero(purchaseOrderUuid uint64) error {
	return r.db.Where("purchase_order_uuid = ?", purchaseOrderUuid).Where("num = 0").Delete(&model.PurchaseOrderItem{}).Error
}

// DeleteByPurchaseOrderUuidAndMaterialUuids 根据采购订单UUID和物品UUID删除明细
func (r *PurchaseOrderItemRepoImpl) DeleteByPurchaseOrderUuidAndMaterialUuids(purchaseOrderUuid uint64, materialUuids []uint64) error {
	if len(materialUuids) == 0 {
		return nil
	}
	return r.db.Where("purchase_order_uuid = ?", purchaseOrderUuid).Where("material_uuid IN ?", materialUuids).Delete(&model.PurchaseOrderItem{}).Error
}

// GetByUuid 根据UUID获取采购订单明细
func (r *PurchaseOrderItemRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.PurchaseOrderItem, error) {
	var item model.PurchaseOrderItem
	db := r.applyOptions(r.db, opts...)
	err := db.Where("uuid = ?", uuid).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetByPurchaseOrderUuidAndMaterialCode 根据采购订单UUID和物料编码获取采购订单明细
func (r *PurchaseOrderItemRepoImpl) GetByPurchaseOrderUuidAndMaterialCode(purchaseOrderUuid uint64, materialCode string) (*model.PurchaseOrderItem, error) {
	var item model.PurchaseOrderItem
	err := r.db.Where("purchase_order_uuid = ?", purchaseOrderUuid).Where("material_code = ?", materialCode).First(&item).Error
	return &item, err
}

// GetByPurchaseOrderUuid 根据采购订单UUID获取明细列表
func (r *PurchaseOrderItemRepoImpl) GetByPurchaseOrderUuid(purchaseOrderUuid uint64, opts ...DBOption) ([]model.PurchaseOrderItem, error) {
	var items []model.PurchaseOrderItem
	db := r.applyOptions(r.db, opts...)
	err := db.Where("purchase_order_uuid = ?", purchaseOrderUuid).Find(&items).Error
	return items, err
}

// GetList 获取采购订单明细列表
func (r *PurchaseOrderItemRepoImpl) GetList(opts ...DBOption) ([]model.PurchaseOrderItem, error) {
	var items []model.PurchaseOrderItem
	db := r.applyOptions(r.db, opts...)
	err := db.Find(&items).Error
	return items, err
}

// Count 统计采购订单明细数量
func (r *PurchaseOrderItemRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrderItem{}).Count(&count).Error
	return count, err
}

// 条件查询选项实现

// WhereUuid UUID条件
func (r *PurchaseOrderItemRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WherePurchaseOrderUuid 采购订单UUID条件
func (r *PurchaseOrderItemRepoImpl) WherePurchaseOrderUuid(purchaseOrderUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("purchase_order_uuid = ?", purchaseOrderUuid)
	}
}

// WhereProductUuid 商品UUID条件
func (r *PurchaseOrderItemRepoImpl) WhereProductUuid(productUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_uuid = ?", productUuid)
	}
}

// OrderBySort 按排序字段排序
func (r *PurchaseOrderItemRepoImpl) OrderBySort() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC, create_time ASC")
	}
}

// OrderByCreateTime 按创建时间排序
func (r *PurchaseOrderItemRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// 统计查询实现

// GetTotalQuantityByPurchaseOrder 获取采购订单总数量
func (r *PurchaseOrderItemRepoImpl) GetTotalQuantityByPurchaseOrder(purchaseOrderUuid uint64) (float64, error) {
	var totalQuantity float64
	err := r.db.Model(&model.PurchaseOrderItem{}).
		Where("purchase_order_uuid = ?", purchaseOrderUuid).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalQuantity).Error
	return totalQuantity, err
}

// GetTotalAmountByPurchaseOrder 获取采购订单总金额
func (r *PurchaseOrderItemRepoImpl) GetTotalAmountByPurchaseOrder(purchaseOrderUuid uint64) (float64, error) {
	var totalAmount float64
	err := r.db.Model(&model.PurchaseOrderItem{}).
		Where("purchase_order_uuid = ?", purchaseOrderUuid).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&totalAmount).Error
	return totalAmount, err
}

// GetReceivedQuantityByPurchaseOrder 获取采购订单已收货数量
func (r *PurchaseOrderItemRepoImpl) GetReceivedQuantityByPurchaseOrder(purchaseOrderUuid uint64) (float64, error) {
	var receivedQuantity float64
	err := r.db.Model(&model.PurchaseOrderItem{}).
		Where("purchase_order_uuid = ?", purchaseOrderUuid).
		Select("COALESCE(SUM(received_quantity), 0)").
		Scan(&receivedQuantity).Error
	return receivedQuantity, err
}

// SumBrandPurchaseByTimeRange 统计指定时间范围内的品牌采购物品数量
//
// 参数:
//   - companyUuid: 公司UUID
//   - materialCode: 物品编码
//   - unitCode: 单位编码
//   - startTime: 开始时间（Unix 时间戳）
//   - endTime: 结束时间（Unix 时间戳）
//   - excludeOrderUuid: 排除的订单UUID（避免重复统计）
//
// 返回:
//   - float64: 已使用数量
//   - error: 错误信息
func (r *PurchaseOrderItemRepoImpl) SumBrandPurchaseByTimeRange(
	materialCode string,
	unitCode string,
	startTime, endTime int64,
	excludeOrderUuid uint64,
) (float64, error) {
	var usedQty float64
	err := r.db.Table("ttpos_purchase_order po").
		Select("COALESCE(SUM(poi.num), 0) as used_qty").
		Joins("JOIN ttpos_purchase_order_item poi ON po.uuid = poi.purchase_order_uuid").
		Where("po.purchase_type = ?", 2). // 品牌采购 (constant.PurchaseTypeBrand)
		Where("poi.material_code = ?", materialCode).
		Where("poi.erpnext_uom = ?", unitCode).
		Where("po.status != ?", constant.PurchaseOrderStatusDraft). // 排除草稿
		Where("po.order_time > 0").                                 // order_time 必须有值
		Where("po.uuid != ?", excludeOrderUuid).                    // 排除当前单据
		Where("po.order_time >= ? AND po.order_time <= ?", startTime, endTime).
		Where("po.delete_time = 0").
		Scan(&usedQty).Error

	return usedQty, err
}

// SumBrandPurchaseByMonth 统计指定月份的品牌采购物品数量
//
// 参数:
//   - companyUuid: 公司UUID
//   - materialCode: 物品编码
//   - unitCode: 单位编码
//   - month: 月份（格式：2026-01）
//   - excludeOrderUuid: 排除的订单UUID（避免重复统计）
//
// 返回:
//   - float64: 已使用数量
//   - error: 错误信息
func (r *PurchaseOrderItemRepoImpl) SumBrandPurchaseByMonth(
	materialCode string,
	unitCode string,
	month string,
	excludeOrderUuid uint64,
) (float64, error) {
	var usedQty float64
	err := r.db.Table("ttpos_purchase_order po").
		Select("COALESCE(SUM(poi.num), 0) as used_qty").
		Joins("JOIN ttpos_purchase_order_item poi ON po.uuid = poi.purchase_order_uuid").
		Where("po.purchase_type = ?", 2). // 品牌采购 (constant.PurchaseTypeBrand)
		Where("poi.material_code = ?", materialCode).
		Where("poi.erpnext_uom = ?", unitCode).
		Where("po.status != ?", constant.PurchaseOrderStatusDraft). // 排除草稿
		Where("po.order_time > 0").                                 // order_time 必须有值
		Where("po.uuid != ?", excludeOrderUuid).                    // 排除当前单据
		Where("FROM_UNIXTIME(po.order_time, '%Y-%m') = ?", month).
		Where("po.delete_time = 0").
		Scan(&usedQty).Error

	return usedQty, err
}

// applyOptions 应用查询选项
func (r *PurchaseOrderItemRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

// GetNotReceivedQuantityByMaterialUuid 根据物料UUID获取未收货数量
func (r *PurchaseOrderItemRepoImpl) GetNotReceivedQuantityByMaterialUuid(materialUuid uint64) (int64, error) {
	var count int64
	db := r.db.Model(&model.PurchaseOrderItem{}).Session(&gorm.Session{})
	result := db.Where("material_uuid = ?", materialUuid).Where("num > arrival_num").Count(&count)
	err := result.Error
	if err != nil {
		return 0, fmt.Errorf("GetNotReceivedQuantityByMaterialUuid: %v", err)
	}
	return count, nil
}

// WithPreloadUnits 预加载单位
func (r *PurchaseOrderItemRepoImpl) WithPreloadUnits() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Units")
	}
}
