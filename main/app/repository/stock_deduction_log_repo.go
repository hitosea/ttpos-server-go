package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IStockDeductionLogRepo 库存扣减日志仓库接口
type IStockDeductionLogRepo interface {
	// BatchCreate 批量创建扣减日志
	BatchCreate(logs []*model.StockDeductionLog) error
	// GetByOrderUuids 批量查询订单的已扣减记录
	GetByOrderUuids(orderUuids []uint64) ([]*model.StockDeductionLog, error)
	// DeleteByOrderUuid 软删除指定订单的所有扣减记录（反结账用）
	DeleteByOrderUuid(orderUuid uint64) error
}

// NewStockDeductionLogRepo 创建新的库存扣减日志仓库
func NewStockDeductionLogRepo(db *gorm.DB) IStockDeductionLogRepo {
	return &stockDeductionLogRepoImpl{db: db}
}

type stockDeductionLogRepoImpl struct {
	db *gorm.DB
}

func (r *stockDeductionLogRepoImpl) BatchCreate(logs []*model.StockDeductionLog) error {
	if len(logs) == 0 {
		return nil
	}
	if err := r.db.CreateInBatches(logs, 100).Error; err != nil {
		return errors.WithMessage(err, "批量创建库存扣减日志失败")
	}
	return nil
}

func (r *stockDeductionLogRepoImpl) GetByOrderUuids(orderUuids []uint64) ([]*model.StockDeductionLog, error) {
	if len(orderUuids) == 0 {
		return nil, nil
	}
	var logs []*model.StockDeductionLog
	err := r.db.Scopes(NotDeleted).
		Where("sale_order_uuid IN ?", orderUuids).
		Find(&logs).Error
	if err != nil {
		return nil, errors.WithMessage(err, "查询库存扣减日志失败")
	}
	return logs, nil
}

func (r *stockDeductionLogRepoImpl) DeleteByOrderUuid(orderUuid uint64) error {
	return r.db.Model(&model.StockDeductionLog{}).
		Where("sale_order_uuid = ? AND delete_time = 0", orderUuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}
