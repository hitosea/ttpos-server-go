package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITransferOrderLogRepo 调拨单操作日志Repository接口
type ITransferOrderLogRepo interface {
	Create(log *model.TransferOrderLog) error
	GetListByTransferOrderUuid(transferOrderUuid uint64, pageNo, pageSize int) ([]model.TransferOrderLog, int64, error)
}

// TransferOrderLogRepoImpl 调拨单操作日志Repository实现
type TransferOrderLogRepoImpl struct {
	db *gorm.DB
}

// NewTransferOrderLogRepo 创建调拨单操作日志Repository
func NewTransferOrderLogRepo(db *gorm.DB) ITransferOrderLogRepo {
	return &TransferOrderLogRepoImpl{db: db}
}

func (r *TransferOrderLogRepoImpl) Create(log *model.TransferOrderLog) error {
	return r.db.Create(log).Error
}

func (r *TransferOrderLogRepoImpl) GetListByTransferOrderUuid(transferOrderUuid uint64, pageNo, pageSize int) ([]model.TransferOrderLog, int64, error) {
	var logs []model.TransferOrderLog
	var total int64

	offset := (pageNo - 1) * pageSize
	err := r.db.Model(&model.TransferOrderLog{}).
		Where("transfer_order_uuid = ?", transferOrderUuid).
		Where("delete_time = ?", constant.NotDeleted).
		Count(&total).
		Offset(offset).
		Limit(pageSize).
		Order("create_time DESC").
		Find(&logs).Error

	return logs, total, err
}
