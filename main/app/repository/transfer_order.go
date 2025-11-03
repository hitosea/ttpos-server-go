package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITransferOrderRepo 调拨单Repository接口
type ITransferOrderRepo interface {
	// 基础操作
	Create(transferOrder *model.TransferOrder) error
	Update(transferOrder *model.TransferOrder) error
	Delete(uuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.TransferOrder, error)
	GetByOrderNo(orderNo string, opts ...DBOption) (*model.TransferOrder, error)

	// 查询操作
	GetList(opts ...DBOption) ([]model.TransferOrder, error)
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.TransferOrder, int64, error)
	Count(opts ...DBOption) (int64, error)

	// 条件查询选项
	WhereUuid(uuid uint64) DBOption
	WhereUuidIn(uuidIn []uint64) DBOption
	WhereOrderNo(orderNo string) DBOption
	WhereTransferType(transferType int) DBOption
	WhereStatus(status int) DBOption
	WhereStatusIn(statusIn []int) DBOption
	WhereSenderCompanyUuid(senderCompanyUuid uint64) DBOption
	WhereReceiverCompanyUuid(receiverCompanyUuid uint64) DBOption
	WhereOutWarehouseErpCode(outWarehouseErpCode string) DBOption
	WhereInWarehouseErpCode(inWarehouseErpCode string) DBOption
	WhereCreateTimeRange(start, end int) DBOption
	WhereOrderTimeRange(start, end int) DBOption
	WhereSubmitTimeRange(start, end int) DBOption
	WhereCompanyUuid(companyUuid uint64) DBOption
	WhereHeadquarterUuid(headquarterUuid uint64) DBOption

	// 预加载
	WithItems() DBOption
	WithApprovals() DBOption
	WithLogs() DBOption

	// 排序
	OrderByCreateTime(desc bool) DBOption
	OrderByOrderTime(desc bool) DBOption
	OrderBySubmitTime(desc bool) DBOption
	OrderByStatus() DBOption

	// 统计查询
	GetStatusStats(opts ...DBOption) (map[int]int64, error)
	GetTransferTypeStats(opts ...DBOption) (map[int]int64, error)
	IsOrderNoExists(orderNo string) (bool, error)

	// 获取今天最新的调拨单
	GetLatestOrderToday() (*model.TransferOrder, error)
}

// TransferOrderRepoImpl 调拨单Repository实现
type TransferOrderRepoImpl struct {
	db *gorm.DB
}

// NewTransferOrderRepo 创建调拨单Repository
func NewTransferOrderRepo(db *gorm.DB) ITransferOrderRepo {
	return &TransferOrderRepoImpl{db: db}
}

// Create 创建调拨单
func (r *TransferOrderRepoImpl) Create(transferOrder *model.TransferOrder) error {
	transferOrder.SetNil()
	return r.db.Create(transferOrder).Error
}

// Update 更新调拨单
func (r *TransferOrderRepoImpl) Update(transferOrder *model.TransferOrder) error {
	return r.db.Model(&model.TransferOrder{}).Where("uuid = ?", transferOrder.Uuid).Updates(transferOrder).Error
}

// Delete 删除调拨单（软删除）
func (r *TransferOrderRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.TransferOrder{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

// GetByUuid 根据UUID获取调拨单
func (r *TransferOrderRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.TransferOrder, error) {
	var transferOrder model.TransferOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrder{}).Where("uuid = ?", uuid).Where("delete_time = ?", constant.NotDeleted).First(&transferOrder).Error
	if err != nil {
		return nil, err
	}
	return &transferOrder, nil
}

// GetByOrderNo 根据订单编号获取调拨单
func (r *TransferOrderRepoImpl) GetByOrderNo(orderNo string, opts ...DBOption) (*model.TransferOrder, error) {
	var transferOrder model.TransferOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrder{}).Where("order_no = ?", orderNo).Where("delete_time = ?", constant.NotDeleted).First(&transferOrder).Error
	if err != nil {
		return nil, err
	}
	return &transferOrder, nil
}

// GetList 获取调拨单列表
func (r *TransferOrderRepoImpl) GetList(opts ...DBOption) ([]model.TransferOrder, error) {
	var transferOrders []model.TransferOrder
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrder{}).Where("delete_time = ?", constant.NotDeleted).Find(&transferOrders).Error
	return transferOrders, err
}

// GetListWithPagination 分页获取调拨单列表
func (r *TransferOrderRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.TransferOrder, int64, error) {
	var transferOrders []model.TransferOrder
	var total int64
	db := r.applyOptions(r.db, opts...)
	// 分页查询
	offset := (pageNo - 1) * pageSize
	err := db.Model(&model.TransferOrder{}).
		Where("delete_time = ?", constant.NotDeleted).
		Count(&total).
		Offset(offset).
		Limit(pageSize).
		Find(&transferOrders).Error
	return transferOrders, total, err
}

// Count 统计调拨单数量
func (r *TransferOrderRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrder{}).Where("delete_time = ?", constant.NotDeleted).Count(&count).Error
	return count, err
}

// 条件查询选项实现

// WhereUuid UUID条件
func (r *TransferOrderRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereUuidIn UUID列表条件
func (r *TransferOrderRepoImpl) WhereUuidIn(uuidIn []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(uuidIn) > 0 {
			return db.Where("uuid IN (?)", uuidIn)
		}
		return db
	}
}

// WhereOrderNo 订单编号条件
func (r *TransferOrderRepoImpl) WhereOrderNo(orderNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if orderNo != "" {
			return db.Where("(order_no LIKE ? OR erp_order_no LIKE ?)", "%"+orderNo+"%", "%"+orderNo+"%")
		}
		return db
	}
}

// WhereTransferType 调拨类型条件
func (r *TransferOrderRepoImpl) WhereTransferType(transferType int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if transferType > 0 {
			return db.Where("transfer_type = ?", transferType)
		}
		return db
	}
}

// WhereStatus 状态条件
func (r *TransferOrderRepoImpl) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereStatusIn 状态列表条件
func (r *TransferOrderRepoImpl) WhereStatusIn(statusIn []int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(statusIn) > 0 {
			return db.Where("status IN (?)", statusIn)
		}
		return db
	}
}

// WhereSenderCompanyUuid 发货门店条件
func (r *TransferOrderRepoImpl) WhereSenderCompanyUuid(senderCompanyUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if senderCompanyUuid > 0 {
			return db.Where("sender_company_uuid = ?", senderCompanyUuid)
		}
		return db
	}
}

// WhereReceiverCompanyUuid 收货门店条件
func (r *TransferOrderRepoImpl) WhereReceiverCompanyUuid(receiverCompanyUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if receiverCompanyUuid > 0 {
			return db.Where("receiver_company_uuid = ?", receiverCompanyUuid)
		}
		return db
	}
}

// WhereOutWarehouseErpCode 出库仓库ERP编码条件
func (r *TransferOrderRepoImpl) WhereOutWarehouseErpCode(outWarehouseErpCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if outWarehouseErpCode != "" {
			return db.Where("out_warehouse_erp_code = ?", outWarehouseErpCode)
		}
		return db
	}
}

// WhereInWarehouseErpCode 入库仓库ERP编码条件
func (r *TransferOrderRepoImpl) WhereInWarehouseErpCode(inWarehouseErpCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if inWarehouseErpCode != "" {
			return db.Where("in_warehouse_erp_code = ?", inWarehouseErpCode)
		}
		return db
	}
}

// WhereCreateTimeRange 创建时间范围条件
func (r *TransferOrderRepoImpl) WhereCreateTimeRange(start, end int) DBOption {
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

// WhereOrderTimeRange 单据时间范围条件
func (r *TransferOrderRepoImpl) WhereOrderTimeRange(start, end int) DBOption {
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

// WhereSubmitTimeRange 提交时间范围条件
func (r *TransferOrderRepoImpl) WhereSubmitTimeRange(start, end int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if start > 0 {
			db = db.Where("submit_time >= ?", start)
		}
		if end > 0 {
			db = db.Where("submit_time <= ?", end)
		}
		return db
	}
}

// WhereCompanyUuid 公司UUID条件
func (r *TransferOrderRepoImpl) WhereCompanyUuid(companyUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if companyUuid > 0 {
			return db.Where("company_uuid = ?", companyUuid)
		}
		return db
	}
}

// WhereHeadquarterUuid 总部UUID条件
func (r *TransferOrderRepoImpl) WhereHeadquarterUuid(headquarterUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if headquarterUuid > 0 {
			return db.Where("headquarter_uuid = ?", headquarterUuid)
		}
		return db
	}
}

// WithItems 预加载明细
func (r *TransferOrderRepoImpl) WithItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items").Preload("Items.Units")
	}
}

// WithApprovals 预加载审批流程
func (r *TransferOrderRepoImpl) WithApprovals() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Approvals", "delete_time = ?", constant.NotDeleted).Order("Approvals.sequence ASC")
	}
}

// WithLogs 预加载操作日志
func (r *TransferOrderRepoImpl) WithLogs() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Logs", "delete_time = ?", constant.NotDeleted).Order("Logs.create_time DESC")
	}
}

// OrderByCreateTime 按创建时间排序
func (r *TransferOrderRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// OrderByOrderTime 按单据时间排序
func (r *TransferOrderRepoImpl) OrderByOrderTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("order_time DESC")
		}
		return db.Order("order_time ASC")
	}
}

// OrderBySubmitTime 按提交时间排序
func (r *TransferOrderRepoImpl) OrderBySubmitTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("submit_time DESC")
		}
		return db.Order("submit_time ASC")
	}
}

// OrderByStatus 按状态排序
func (r *TransferOrderRepoImpl) OrderByStatus() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("status ASC")
	}
}

// GetStatusStats 获取状态统计
func (r *TransferOrderRepoImpl) GetStatusStats(opts ...DBOption) (map[int]int64, error) {
	var results []struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrder{}).
		Select("status, COUNT(*) as count").
		Where("delete_time = ?", constant.NotDeleted).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make(map[int]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}
	return stats, nil
}

// GetTransferTypeStats 获取调拨类型统计
func (r *TransferOrderRepoImpl) GetTransferTypeStats(opts ...DBOption) (map[int]int64, error) {
	var results []struct {
		TransferType int   `json:"transfer_type"`
		Count        int64 `json:"count"`
	}
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.TransferOrder{}).
		Select("transfer_type, COUNT(*) as count").
		Where("delete_time = ?", constant.NotDeleted).
		Group("transfer_type").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make(map[int]int64)
	for _, result := range results {
		stats[result.TransferType] = result.Count
	}
	return stats, nil
}

// IsOrderNoExists 检查订单编号是否存在
func (r *TransferOrderRepoImpl) IsOrderNoExists(orderNo string) (bool, error) {
	var count int64
	err := r.db.Model(&model.TransferOrder{}).Where("order_no = ?", orderNo).Where("delete_time = ?", constant.NotDeleted).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLatestOrderToday 获取今天最新的调拨单
func (r *TransferOrderRepoImpl) GetLatestOrderToday() (*model.TransferOrder, error) {
	var transferOrder model.TransferOrder

	// 获取今天的开始和结束时间戳
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).Unix()

	err := r.db.Where("create_time >= ? AND create_time <= ?", startOfDay, endOfDay).
		Where("delete_time = ?", constant.NotDeleted).
		Order("create_time DESC").
		First(&transferOrder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &transferOrder, nil
}

// applyOptions 应用查询选项
func (r *TransferOrderRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}
