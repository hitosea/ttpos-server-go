package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseReceiptRepo 采购收货记录Repository接口
type IPurchaseReceiptOrderRepo interface {
	// 基础操作
	Create(receipt *model.PurchaseReceiptOrder) error
	Update(receipt *model.PurchaseReceiptOrder) error
	Delete(uuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.PurchaseReceiptOrder, error)
	GetByReceiptNo(receiptNo string, opts ...DBOption) (*model.PurchaseReceiptOrder, error)

	// 查询操作
	GetList(opts ...DBOption) ([]model.PurchaseReceiptOrder, error)
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.PurchaseReceiptOrder, int64, error)
	Count(opts ...DBOption) (int64, error)
	IsOrderNoExists(orderNo string) (bool, error)

	// 获取今天最新的收货单
	GetLatestReceiptToday() (*model.PurchaseReceiptOrder, error)

	// 条件查询选项
	WhereUuid(uuid uint64) DBOption
	WhereReceiptNo(receiptNo string) DBOption
	WherePurchaseOrderUuid(purchaseOrderUuid uint64) DBOption
	WhereReceiverUuid(receiverUuid uint64) DBOption
	WhereQualityStatus(qualityStatus int) DBOption
	WhereReceiptTimeRange(start, end int) DBOption
	WhereCreateTimeRange(start, end int) DBOption
	WhereReceiptType(receiptType int) DBOption
	WhereStatusIn(statusIn []int) DBOption
	WithItems() DBOption
	OrderByReceiptTime(desc bool) DBOption
	OrderByCreateTime(desc bool) DBOption
}

// PurchaseReceiptRepoImpl 采购收货记录Repository实现
type PurchaseReceiptOrderRepoImpl struct {
	db *gorm.DB
}

// NewPurchaseReceiptRepo 创建采购收货记录Repository
func NewPurchaseReceiptOrderRepo(db *gorm.DB) IPurchaseReceiptOrderRepo {
	return &PurchaseReceiptOrderRepoImpl{db: db}
}

// Create 创建收货记录
func (r *PurchaseReceiptOrderRepoImpl) Create(receipt *model.PurchaseReceiptOrder) error {
	return r.db.Create(receipt).Error
}

// Update 更新收货记录
func (r *PurchaseReceiptOrderRepoImpl) Update(receipt *model.PurchaseReceiptOrder) error {
	return r.db.Model(&model.PurchaseReceiptOrder{}).Where("uuid = ?", receipt.Uuid).Updates(receipt).Error
}

// Delete 删除收货记录
func (r *PurchaseReceiptOrderRepoImpl) Delete(uuid uint64) error {
	return r.db.Where("uuid = ?", uuid).Delete(&model.PurchaseReceiptOrder{}).Error
}

// GetByUuid 根据UUID获取收货记录
func (r *PurchaseReceiptOrderRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.PurchaseReceiptOrder, error) {
	var receipt model.PurchaseReceiptOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Where("uuid = ?", uuid).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetByReceiptNo 根据收货单号获取收货记录
func (r *PurchaseReceiptOrderRepoImpl) GetByReceiptNo(receiptNo string, opts ...DBOption) (*model.PurchaseReceiptOrder, error) {
	var receipt model.PurchaseReceiptOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Where("receipt_no = ?", receiptNo).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetList 获取收货记录列表
func (r *PurchaseReceiptOrderRepoImpl) GetList(opts ...DBOption) ([]model.PurchaseReceiptOrder, error) {
	var receipts []model.PurchaseReceiptOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Find(&receipts).Error
	return receipts, err
}

// GetListWithPagination 分页获取收货记录列表
func (r *PurchaseReceiptOrderRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.PurchaseReceiptOrder, int64, error) {
	var receipts []model.PurchaseReceiptOrder
	var total int64

	db := r.applyOptions(r.db, opts...)

	// 计算总数
	err := db.Model(&model.PurchaseReceiptOrder{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Find(&receipts).Error
	return receipts, total, err
}

// Count 统计收货记录数量
func (r *PurchaseReceiptOrderRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseReceiptOrder{}).Count(&count).Error
	return count, err
}

// IsOrderNoExists 检查订单编号是否存在
func (r *PurchaseReceiptOrderRepoImpl) IsOrderNoExists(orderNo string) (bool, error) {
	var count int64
	err := r.db.Model(&model.PurchaseReceiptOrder{}).Where("receipt_no = ?", orderNo).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLatestReceiptToday 获取今天最新的收货单
func (r *PurchaseReceiptOrderRepoImpl) GetLatestReceiptToday() (*model.PurchaseReceiptOrder, error) {
	var receiptOrder model.PurchaseReceiptOrder

	// 获取今天的开始和结束时间戳
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).Unix()

	err := r.db.Where("create_time >= ? AND create_time <= ?", startOfDay, endOfDay).
		Order("create_time DESC").
		First(&receiptOrder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &receiptOrder, nil
}

// 条件查询选项实现

// WhereUuid UUID条件
func (r *PurchaseReceiptOrderRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereReceiptNo 收货单号条件
func (r *PurchaseReceiptOrderRepoImpl) WhereReceiptNo(receiptNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if receiptNo != "" {
			return db.Where("order_no LIKE ?", "%"+receiptNo+"%")
		}
		return db
	}
}

// WherePurchaseOrderUuid 采购订单UUID条件
func (r *PurchaseReceiptOrderRepoImpl) WherePurchaseOrderUuid(purchaseOrderUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("purchase_order_uuid = ?", purchaseOrderUuid)
	}
}

// WhereReceiverUuid 收货人UUID条件
func (r *PurchaseReceiptOrderRepoImpl) WhereReceiverUuid(receiverUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("receiver_uuid = ?", receiverUuid)
	}
}

// WhereQualityStatus 质检状态条件
func (r *PurchaseReceiptOrderRepoImpl) WhereQualityStatus(qualityStatus int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("quality_status = ?", qualityStatus)
	}
}

// WhereReceiptTimeRange 收货时间范围条件
func (r *PurchaseReceiptOrderRepoImpl) WhereReceiptTimeRange(start, end int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if start > 0 {
			db = db.Where("receive_time >= ?", start)
		}
		if end > 0 {
			db = db.Where("receive_time <= ?", end)
		}
		return db
	}
}

// WhereCreateTimeRange 创建时间范围条件
func (r *PurchaseReceiptOrderRepoImpl) WhereCreateTimeRange(start, end int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if start > 0 {
			db = db.Where("create_time >= ?", start)
		}
		if end > 0 {
			db = db.Where("create_time <= ?", end)
		}
		return db
	}
}

// WhereReceiptType 收货类型条件
func (r *PurchaseReceiptOrderRepoImpl) WhereReceiptType(receiptType int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("receipt_type = ?", receiptType)
	}
}

// WhereStatusIn 状态条件
func (r *PurchaseReceiptOrderRepoImpl) WhereStatusIn(statusIn []int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status IN (?)", statusIn)
	}
}

// WithItems 预加载收货明细
func (r *PurchaseReceiptOrderRepoImpl) WithItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items.PurchaseOrderItem")
	}
}

// OrderByReceiptTime 按收货时间排序
func (r *PurchaseReceiptOrderRepoImpl) OrderByReceiptTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("receipt_time DESC")
		}
		return db.Order("receipt_time ASC")
	}
}

// OrderByCreateTime 按创建时间排序
func (r *PurchaseReceiptOrderRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// applyOptions 应用查询选项
func (r *PurchaseReceiptOrderRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}
