package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IAutoReceiptLogRepo interface {
	Create(log model.AutoReceiptLog) (model.AutoReceiptLog, error)
	GetByUuid(uuid uint64, headquarterUuid uint64) (model.AutoReceiptLog, error)
	GetList(headquarterUuid uint64, shopCompanyUuid uint64, startTime int64, endTime int64, pageNo int, pageSize int) ([]model.AutoReceiptLog, int64, error)
}

func NewAutoReceiptLogRepo(db *gorm.DB) IAutoReceiptLogRepo {
	return &autoReceiptLogRepo{db: db}
}

type autoReceiptLogRepo struct {
	db *gorm.DB
}

func (r *autoReceiptLogRepo) Create(log model.AutoReceiptLog) (model.AutoReceiptLog, error) {
	err := r.db.Create(&log).Error
	return log, errors.WithMessage(err)
}

func (r *autoReceiptLogRepo) GetByUuid(uuid uint64, headquarterUuid uint64) (model.AutoReceiptLog, error) {
	var log model.AutoReceiptLog
	err := r.db.Where("uuid = ? AND headquarter_company_uuid = ? AND delete_time = 0", uuid, headquarterUuid).
		First(&log).Error
	return log, errors.WithMessage(err)
}

func (r *autoReceiptLogRepo) GetList(headquarterUuid uint64, shopCompanyUuid uint64, startTime int64, endTime int64, pageNo int, pageSize int) ([]model.AutoReceiptLog, int64, error) {
	var logs []model.AutoReceiptLog
	var total int64

	db := r.db.Model(&model.AutoReceiptLog{}).
		Where("headquarter_company_uuid = ? AND delete_time = 0", headquarterUuid)
	if shopCompanyUuid > 0 {
		db = db.Where("shop_company_uuid = ?", shopCompanyUuid)
	}
	if startTime > 0 && endTime > 0 {
		db = db.Where("receipt_time >= ? AND receipt_time <= ?", startTime, endTime)
	}

	err := db.Count(&total).Error
	if err != nil {
		return logs, 0, errors.WithMessage(err)
	}

	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).
		Order("create_time desc").Find(&logs).Error
	return logs, total, errors.WithMessage(err)
}
