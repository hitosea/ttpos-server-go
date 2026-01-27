package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ICostCardCorrectionLogRepo 成本卡修正日志Repository接口
type ICostCardCorrectionLogRepo interface {
	// 基础操作
	Create(log *model.CostCardCorrectionLog) error
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.CostCardCorrectionLog, int64, error)

	// 条件查询选项
	WhereCorrectionUuid(correctionUuid uint64) DBOption
	WhereOrderUuid(orderUuid uint64) DBOption
	WhereOperationType(operationType string) DBOption
	WhereStatus(status string) DBOption
}

// CostCardCorrectionLogRepoImpl 成本卡修正日志Repository实现
type CostCardCorrectionLogRepoImpl struct {
	db *gorm.DB
}

// NewCostCardCorrectionLogRepo 创建成本卡修正日志Repository
func NewCostCardCorrectionLogRepo(db *gorm.DB) ICostCardCorrectionLogRepo {
	return &CostCardCorrectionLogRepoImpl{db: db}
}

// Create 创建日志记录
func (r *CostCardCorrectionLogRepoImpl) Create(log *model.CostCardCorrectionLog) error {
	return r.db.Create(log).Error
}

// GetListWithPagination 分页获取日志列表
func (r *CostCardCorrectionLogRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.CostCardCorrectionLog, int64, error) {
	var logs []model.CostCardCorrectionLog
	var total int64

	query := r.db.Model(&model.CostCardCorrectionLog{}).Scopes(NotDeleted)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	if err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// WhereCorrectionUuid 按修正UUID查询
func (r *CostCardCorrectionLogRepoImpl) WhereCorrectionUuid(correctionUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("correction_uuid = ?", correctionUuid)
	}
}

// WhereOrderUuid 按订单UUID查询
func (r *CostCardCorrectionLogRepoImpl) WhereOrderUuid(orderUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("order_uuid = ?", orderUuid)
	}
}

// WhereOperationType 按操作类型查询
func (r *CostCardCorrectionLogRepoImpl) WhereOperationType(operationType string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("operation_type = ?", operationType)
	}
}

// WhereStatus 按状态查询
func (r *CostCardCorrectionLogRepoImpl) WhereStatus(status string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}
