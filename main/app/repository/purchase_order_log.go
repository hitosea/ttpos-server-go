package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseOrderLogRepo 采购订单日志Repository接口
type IPurchaseOrderLogRepo interface {
	// 基础操作
	Create(log *model.PurchaseOrderLog) error
	GetList(opts ...DBOption) ([]model.PurchaseOrderLog, error)
	Count(opts ...DBOption) (int64, error)

	// 条件查询选项
	WherePurchaseOrderUuid(purchaseOrderUuid uint64) DBOption
	WhereOperatorUuid(operatorUuid uint64) DBOption
	WhereAction(action string) DBOption
	OrderByCreateTime(desc bool) DBOption
}

// PurchaseOrderLogRepoImpl 采购订单日志Repository实现
type PurchaseOrderLogRepoImpl struct {
	db *gorm.DB
}

// NewPurchaseOrderLogRepo 创建采购订单日志Repository
func NewPurchaseOrderLogRepo(db *gorm.DB) IPurchaseOrderLogRepo {
	return &PurchaseOrderLogRepoImpl{db: db}
}

// Create 创建操作日志
func (r *PurchaseOrderLogRepoImpl) Create(log *model.PurchaseOrderLog) error {
	return r.db.Create(log).Error
}

// GetList 获取操作日志列表
func (r *PurchaseOrderLogRepoImpl) GetList(opts ...DBOption) ([]model.PurchaseOrderLog, error) {
	var logs []model.PurchaseOrderLog
	db := r.applyOptions(r.db, opts...)
	err := db.Find(&logs).Error
	return logs, err
}

// Count 统计操作日志数量
func (r *PurchaseOrderLogRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	db := r.applyOptions(r.db, opts...)
	err := db.Model(&model.PurchaseOrderLog{}).Count(&count).Error
	return count, err
}

// 条件查询选项实现

// WherePurchaseOrderUuid 采购订单UUID条件
func (r *PurchaseOrderLogRepoImpl) WherePurchaseOrderUuid(purchaseOrderUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("purchase_order_uuid = ?", purchaseOrderUuid)
	}
}

// WhereOperatorUuid 操作人UUID条件
func (r *PurchaseOrderLogRepoImpl) WhereOperatorUuid(operatorUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("operator_uuid = ?", operatorUuid)
	}
}

// WhereAction 操作动作条件
func (r *PurchaseOrderLogRepoImpl) WhereAction(action string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("action = ?", action)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *PurchaseOrderLogRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// applyOptions 应用查询选项
func (r *PurchaseOrderLogRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}
