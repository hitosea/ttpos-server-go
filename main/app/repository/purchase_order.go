package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseOrderRepo 采购订单Repository接口
type IPurchaseOrderRepo interface {
	// 基础操作
	Create(purchaseOrder *model.PurchaseOrder) error
	Update(purchaseOrder *model.PurchaseOrder) error
	Delete(uuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.PurchaseOrder, error)
	GetBySubUuid(subUuid uint64, opts ...DBOption) (*model.PurchaseOrder, error)
	GetByOrderNo(orderNo string, opts ...DBOption) (*model.PurchaseOrder, error)

	// 查询操作
	GetList(opts ...DBOption) ([]model.PurchaseOrder, error)
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.PurchaseOrder, int64, error)
	Count(opts ...DBOption) (int64, error)

	// 条件查询选项
	WhereUuid(uuid uint64) DBOption
	WhereOrderNo(orderNo string) DBOption
	WhereTitle(title string) DBOption
	WhereStatus(status int) DBOption
	WhereStatusIn(statusIn []int) DBOption
	WherePriority(priority int) DBOption
	WhereSupplierUuid(supplierUuid uint64) DBOption
	WhereSupplierName(supplierName string) DBOption
	WhereApplicantUuid(applicantUuid uint64) DBOption
	WhereApproverUuid(approverUuid uint64) DBOption
	WhereCreateTimeRange(start, end int) DBOption
	WhereDeliveryTimeRange(start, end int) DBOption
	WhereOrderTimeRange(start, end int) DBOption
	WhereExpectArrivalTimeRange(start, end int) DBOption
	WhereReceiveTimeRange(start, end int) DBOption
	WhereUuidIn(uuidIn []uint64) DBOption
	WherePurchaseType(purchaseType int) DBOption
	WhereWarehouseErpCode(warehouseErpCode string) DBOption
	WhereCompanyUuid(companyUuid uint64) DBOption
	WithItems() DBOption
	WithWarehouse() DBOption
	WithLogs() DBOption
	WithReceipts() DBOption
	WithSupplier() DBOption
	OrderByCreateTime(desc bool) DBOption
	OrderByOrderTime(desc bool) DBOption

	OrderByStatus() DBOption
	OrderByPriority() DBOption

	// 统计查询
	GetStatusStats(opts ...DBOption) (map[int]int64, error)
	GetPriorityStats(opts ...DBOption) (map[int]int64, error)
	GetSupplierStats(limit int, opts ...DBOption) ([]map[string]interface{}, error)
	GetAmountStats(opts ...DBOption) (float64, error)
	IsOrderNoExists(orderNo string) (bool, error)

	// 获取今天最新的采购申请
	GetLatestOrderToday() (*model.PurchaseOrder, error)

	// 判断供应商是否存在
	IsSupplierExists(supplierErpCode string) (bool, error)
}

// PurchaseOrderRepoImpl 采购订单Repository实现
type PurchaseOrderRepoImpl struct {
	db *gorm.DB
}

// NewPurchaseOrderRepo 创建采购订单Repository
func NewPurchaseOrderRepo(db *gorm.DB) IPurchaseOrderRepo {
	return &PurchaseOrderRepoImpl{db: db}
}

// Create 创建采购订单
func (r *PurchaseOrderRepoImpl) Create(purchaseOrder *model.PurchaseOrder) error {
	purchaseOrder.SetNil()
	return r.db.Create(purchaseOrder).Error
}

// Update 更新采购订单
func (r *PurchaseOrderRepoImpl) Update(purchaseOrder *model.PurchaseOrder) error {
	return r.db.Model(&model.PurchaseOrder{}).Where("uuid = ?", purchaseOrder.Uuid).Updates(purchaseOrder).Error
}

// Delete 删除采购订单
func (r *PurchaseOrderRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.PurchaseOrder{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

// GetByUuid 根据UUID获取采购订单
func (r *PurchaseOrderRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.PurchaseOrder, error) {
	var purchaseOrder model.PurchaseOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).Where("uuid = ?", uuid).Where("delete_time = ?", constant.NotDeleted).First(&purchaseOrder).Error
	if err != nil {
		return nil, err
	}
	return &purchaseOrder, nil
}

// GetBySubUuid 根据子订单UUID获取采购订单
func (r *PurchaseOrderRepoImpl) GetBySubUuid(subUuid uint64, opts ...DBOption) (*model.PurchaseOrder, error) {
	var purchaseOrder model.PurchaseOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).Where("sub_uuid = ?", subUuid).Where("delete_time = ?", constant.NotDeleted).First(&purchaseOrder).Error
	if err != nil {
		return nil, err
	}
	return &purchaseOrder, nil
}

// GetByOrderNo 根据订单编号获取采购订单
func (r *PurchaseOrderRepoImpl) GetByOrderNo(orderNo string, opts ...DBOption) (*model.PurchaseOrder, error) {
	var purchaseOrder model.PurchaseOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).Where("order_no = ?", orderNo).Where("delete_time = ?", constant.NotDeleted).First(&purchaseOrder).Error
	if err != nil {
		return nil, err
	}
	return &purchaseOrder, nil
}

// GetList 获取采购订单列表
func (r *PurchaseOrderRepoImpl) GetList(opts ...DBOption) ([]model.PurchaseOrder, error) {
	var purchaseOrders []model.PurchaseOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).Where("delete_time = ?", constant.NotDeleted).Find(&purchaseOrders).Error
	return purchaseOrders, err
}

// GetListWithPagination 分页获取采购订单列表
func (r *PurchaseOrderRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.PurchaseOrder, int64, error) {
	var purchaseOrders []model.PurchaseOrder
	var total int64
	db := r.applyOptions(r.db, opts...)
	// 分页查询
	offset := (pageNo - 1) * pageSize
	err := db.Model(&model.PurchaseOrder{}).
		Where("delete_time = ?", constant.NotDeleted).
		Count(&total).
		Offset(offset).
		Limit(pageSize).
		Find(&purchaseOrders).Error
	return purchaseOrders, total, err
}

// Count 统计采购订单数量
func (r *PurchaseOrderRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).Where("delete_time = ?", constant.NotDeleted).Count(&count).Error
	return count, err
}

// 条件查询选项实现

// WhereUuid UUID条件
func (r *PurchaseOrderRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereOrderNo 订单编号条件
func (r *PurchaseOrderRepoImpl) WhereOrderNo(orderNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if orderNo != "" {
			return db.Where("(order_no LIKE ? OR erp_order_no LIKE ?)", "%"+orderNo+"%", "%"+orderNo+"%")
		}
		return db
	}
}

// WhereTitle 标题条件
func (r *PurchaseOrderRepoImpl) WhereTitle(title string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if title != "" {
			return db.Where("title LIKE ?", "%"+title+"%")
		}
		return db
	}
}

// WhereStatusIn 状态条件
func (r *PurchaseOrderRepoImpl) WhereStatusIn(statusIn []int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status IN (?)", statusIn)
	}
}

// WhereStatus 状态条件
func (r *PurchaseOrderRepoImpl) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WherePriority 优先级条件
func (r *PurchaseOrderRepoImpl) WherePriority(priority int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("priority = ?", priority)
	}
}

// WhereSupplierUuid 供应商条件
func (r *PurchaseOrderRepoImpl) WhereSupplierUuid(supplierUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("supplier_uuid = ?", supplierUuid)
	}
}

// WhereApplicantUuid 申请人条件
func (r *PurchaseOrderRepoImpl) WhereApplicantUuid(applicantUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("applicant_uuid = ?", applicantUuid)
	}
}

// WhereApproverUuid 审批人条件
func (r *PurchaseOrderRepoImpl) WhereApproverUuid(approverUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("approver_uuid = ?", approverUuid)
	}
}

// WhereSupplierName 供应商名称条件
func (r *PurchaseOrderRepoImpl) WhereSupplierName(supplierName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("supplier_name LIKE ?", "%"+supplierName+"%")
	}
}

// WhereCreateTimeRange 创建时间范围条件
func (r *PurchaseOrderRepoImpl) WhereCreateTimeRange(start, end int) DBOption {
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

// WhereDeliveryTimeRange 交付时间范围条件
func (r *PurchaseOrderRepoImpl) WhereDeliveryTimeRange(start, end int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if start > 0 {
			db = db.Where("expected_delivery_time >= ?", start)
		}
		if end > 0 {
			db = db.Where("expected_delivery_time <= ?", end)
		}
		return db
	}
}

// WhereOrderTimeRange 订单时间范围条件
func (r *PurchaseOrderRepoImpl) WhereOrderTimeRange(start, end int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if start > 0 {
			db = db.Where("order_time >= ?", start)
		}
		if end > 0 {
			db = db.Where("order_time <= ?", end)
		}
		return db
	}
}

// WhereExpectArrivalTimeRange 期望到货时间范围条件
func (r *PurchaseOrderRepoImpl) WhereExpectArrivalTimeRange(start, end int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if start > 0 {
			db = db.Where("expect_arrival_time >= ?", start)
		}
		if end > 0 {
			db = db.Where("expect_arrival_time <= ?", end)
		}
		return db
	}
}

// WhereReceiveTimeRange 收货时间范围条件
func (r *PurchaseOrderRepoImpl) WhereReceiveTimeRange(start, end int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if start > 0 {
			db = db.Where("final_receive_time >= ?", start)
		}
		if end > 0 {
			db = db.Where("final_receive_time <= ?", end)
		}
		return db
	}
}

// WhereUuidIn 采购订单ID条件
func (r *PurchaseOrderRepoImpl) WhereUuidIn(uuidIn []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid IN (?)", uuidIn)
	}
}

// WithItems 预加载明细
func (r *PurchaseOrderRepoImpl) WithItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items.Material", func(db *gorm.DB) *gorm.DB {
			return db.Order("create_time ASC")
		}).Preload("Items.Material")
	}
}

// WithWarehouse 预加载仓库
func (r *PurchaseOrderRepoImpl) WithWarehouse() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Warehouse", func(db *gorm.DB) *gorm.DB {
			return db.Order("create_time ASC")
		})
	}
}

// WithLogs 预加载日志
func (r *PurchaseOrderRepoImpl) WithLogs() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Logs", func(db *gorm.DB) *gorm.DB {
			return db.Order("create_time DESC")
		})
	}
}

// WithReceipts 预加载收货记录
func (r *PurchaseOrderRepoImpl) WithReceipts() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Receipts", func(db *gorm.DB) *gorm.DB {
			return db.Order("receipt_time DESC")
		})
	}
}

// WithSupplier 预加载供应商
func (r *PurchaseOrderRepoImpl) WithSupplier() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Supplier", func(db *gorm.DB) *gorm.DB {
			return db.Order("create_time ASC")
		})
	}
}

// OrderByCreateTime 按创建时间排序
func (r *PurchaseOrderRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// OrderByOrderTime 按订单时间排序
func (r *PurchaseOrderRepoImpl) OrderByOrderTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("order_time DESC")
		}
		return db.Order("order_time ASC")
	}
}

// WherePurchaseType 按采购类型条件
func (r *PurchaseOrderRepoImpl) WherePurchaseType(purchaseType int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("purchase_type = ?", purchaseType)
	}
}

// WhereWarehouseErpCode 仓库编码条件
func (r *PurchaseOrderRepoImpl) WhereWarehouseErpCode(warehouseErpCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("warehouse_erp_code = ?", warehouseErpCode)
	}
}

// WhereCompanyUuid 公司UUID条件
func (r *PurchaseOrderRepoImpl) WhereCompanyUuid(companyUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("company_uuid = ?", companyUuid)
	}
}

// OrderByStatus 按状态排序
func (r *PurchaseOrderRepoImpl) OrderByStatus() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("status ASC")
	}
}

// OrderByPriority 按优先级排序
func (r *PurchaseOrderRepoImpl) OrderByPriority() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("priority DESC")
	}
}

// 统计查询实现

// GetStatusStats 获取状态统计
func (r *PurchaseOrderRepoImpl) GetStatusStats(opts ...DBOption) (map[int]int64, error) {
	var results []struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}

	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[int]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}

// GetPriorityStats 获取优先级统计
func (r *PurchaseOrderRepoImpl) GetPriorityStats(opts ...DBOption) (map[int]int64, error) {
	var results []struct {
		Priority int   `json:"priority"`
		Count    int64 `json:"count"`
	}

	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[int]int64)
	for _, result := range results {
		stats[result.Priority] = result.Count
	}

	return stats, nil
}

// GetSupplierStats 获取供应商统计
func (r *PurchaseOrderRepoImpl) GetSupplierStats(limit int, opts ...DBOption) ([]map[string]interface{}, error) {
	var results []struct {
		SupplierUuid uint64  `json:"supplier_uuid"`
		SupplierName string  `json:"supplier_name"`
		Count        int64   `json:"count"`
		Amount       float64 `json:"amount"`
	}

	db := r.applyOptions(r.db, opts...)
	query := db.Model(&model.PurchaseOrder{}).
		Select("supplier_uuid, supplier_name, COUNT(*) as count, SUM(total_amount) as amount").
		Group("supplier_uuid, supplier_name").
		Order("count DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make([]map[string]interface{}, len(results))
	for i, result := range results {
		stats[i] = map[string]interface{}{
			"supplier_uuid": result.SupplierUuid,
			"supplier_name": result.SupplierName,
			"count":         result.Count,
			"amount":        result.Amount,
		}
	}

	return stats, nil
}

// GetAmountStats 获取金额统计
func (r *PurchaseOrderRepoImpl) GetAmountStats(opts ...DBOption) (float64, error) {
	var totalAmount float64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrder{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalAmount).Error
	return totalAmount, err
}

// IsOrderNoExists 检查订单编号是否存在
func (r *PurchaseOrderRepoImpl) IsOrderNoExists(orderNo string) (bool, error) {
	var count int64
	err := r.db.Model(&model.PurchaseOrder{}).Where("order_no = ?", orderNo).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLatestOrderToday 获取今天最新的采购申请
func (r *PurchaseOrderRepoImpl) GetLatestOrderToday() (*model.PurchaseOrder, error) {
	var purchaseOrder model.PurchaseOrder

	// 获取今天的开始和结束时间戳
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).Unix()

	err := r.db.Where("create_time >= ? AND create_time <= ?", startOfDay, endOfDay).
		Order("create_time DESC").
		First(&purchaseOrder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &purchaseOrder, nil
}

// applyOptions 应用查询选项
func (r *PurchaseOrderRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

func (r *PurchaseOrderRepoImpl) IsSupplierExists(supplierErpCode string) (bool, error) {
	var count int64
	err := r.db.Model(&model.PurchaseOrder{}).Where("supplier_erp_code = ?", supplierErpCode).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
