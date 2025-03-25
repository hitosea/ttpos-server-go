package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderOperationRecordRepo 订单操作记录
type IOrderOperationRecordRepo interface {
	GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleOrderOperationRecord, int64, error)
	GetRecordLists(saleBillUuid uint64) ([]model.SaleOrderOperationRecord, error)
	GetRecordInfo(saleBillUuid uint64) (model.SaleOrderOperationRecord, error)
	UpdateRecord(saleBillUuid uint64, record model.SaleOrderOperationRecord) error
	CreateSaleOrderOperationRecord(model model.SaleOrderOperationRecord) (uint64, error)
	DeleteRecord(saleBillUuid uint64) error
}

type OrderOperationRecordRepoImpl struct {
	db *gorm.DB
}

func NewOrderOperationRecordRepo(db *gorm.DB) IOrderOperationRecordRepo {
	return NewOrderOperationRecordRepoImpl(db)
}

// NewOrderOperationRecordRepoImpl 创建新的订单操作记录仓库实现
func NewOrderOperationRecordRepoImpl(db *gorm.DB) *OrderOperationRecordRepoImpl {
	return &OrderOperationRecordRepoImpl{db: db}
}

func (r *OrderOperationRecordRepoImpl) CreateSaleOrderOperationRecord(obj model.SaleOrderOperationRecord) (uint64, error) {
	obj.SetNil()
	if err := r.db.Model(&model.SaleOrderOperationRecord{}).Create(&obj).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	// 添加异常日志
	go NewOrderAbnormalRecordRepo(r.db).CreateSaleOrderAbnormalLog(obj)
	// 返回
	return obj.Uuid, nil
}

// GetRecordList 获取订单操作记录列表
func (r *OrderOperationRecordRepoImpl) GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleOrderOperationRecord, int64, error) {
	var orderOperationRecords []model.SaleOrderOperationRecord
	var total int64

	query := r.db.Model(&model.SaleOrderOperationRecord{}).Where("delete_time = ?", 0)

	for _, opt := range opts {
		query = opt(query)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orderOperationRecords).Error
	return orderOperationRecords, total, errors.WithMessage(err)
}

// GetRecordLists 获取订单操作记录列表
func (r *OrderOperationRecordRepoImpl) GetRecordLists(saleBillUuid uint64) ([]model.SaleOrderOperationRecord, error) {
	var orderOperationRecords []model.SaleOrderOperationRecord
	err := r.db.Model(&model.SaleOrderOperationRecord{}).Preload("Operator").
		Where("delete_time = ?", 0).Where("sale_bill_uuid = ?", saleBillUuid).Order("create_time desc, id desc").Find(&orderOperationRecords).Error
	return orderOperationRecords, errors.WithMessage(err)
}

// GetRecordInfo 获取订单操作记录信息
func (r *OrderOperationRecordRepoImpl) GetRecordInfo(saleBillUuid uint64) (model.SaleOrderOperationRecord, error) {
	var orderOperationRecord model.SaleOrderOperationRecord
	if err := r.db.Model(&model.SaleOrderOperationRecord{}).Where("uuid = ?", saleBillUuid).First(&orderOperationRecord).Error; err != nil {
		return model.SaleOrderOperationRecord{}, errors.WithMessage(err)
	}
	return orderOperationRecord, nil
}

// UpdateRecord 更新订单操作记录
func (r *OrderOperationRecordRepoImpl) UpdateRecord(saleBillUuid uint64, record model.SaleOrderOperationRecord) error {
	if err := r.db.Model(&model.SaleOrderOperationRecord{}).Where("uuid = ?", saleBillUuid).Updates(record).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// DeleteRecord 软删除订单操作记录
func (r *OrderOperationRecordRepoImpl) DeleteRecord(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleOrderOperationRecord{}).Where("uuid = ?", saleBillUuid).Update("delete_time", uint(time.Now().Unix())).Error
}
