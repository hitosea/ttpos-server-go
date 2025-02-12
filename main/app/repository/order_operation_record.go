package repository

import (
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// IOrderOperationRecordRepo 订单操作记录
type IOrderOperationRecordRepo interface {
	GetRecordList(pageNo, pageSize int) ([]model.SaleBillOperationRecord, int64, error)
	GetRecordInfo(saleBillUuid uint64) (model.SaleBillOperationRecord, error)
	UpdateRecord(saleBillUuid uint64, record model.SaleBillOperationRecord) error
	CreateRecord(record model.SaleBillOperationRecord) (uint64, error)
	DeleteRecord(saleBillUuid uint64) error
}

func NewOrderOperationRecordRepo(db *gorm.DB) IOrderOperationRecordRepo {
	return NewOrderOperationRecordRepoImpl(db)
}

// NewOrderOperationRecordRepoImpl 创建新的订单操作记录仓库实现
func NewOrderOperationRecordRepoImpl(db *gorm.DB) *OrderOperationRecordRepoImpl {
	return &OrderOperationRecordRepoImpl{db: db}
}

type OrderOperationRecordRepoImpl struct {
	db *gorm.DB
}

// GetOrderOperationRecordList 获取订单操作记录列表
func (r *OrderOperationRecordRepoImpl) GetRecordList(pageNo, pageSize int) ([]model.SaleBillOperationRecord, int64, error) {
	var orderOperationRecords []model.SaleBillOperationRecord
	var total int64

	query := r.db.Model(&model.SaleBillOperationRecord{}).Where("delete_time = ?", 0)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orderOperationRecords).Error
	return orderOperationRecords, total, err
}

// GetOrderOperationRecordInfo 获取订单操作记录信息
func (r *OrderOperationRecordRepoImpl) GetRecordInfo(saleBillUuid uint64) (model.SaleBillOperationRecord, error) {
	var orderOperationRecord model.SaleBillOperationRecord
	if err := r.db.Model(&model.SaleBillOperationRecord{}).Where("uuid = ?", saleBillUuid).First(&orderOperationRecord).Error; err != nil {
		return model.SaleBillOperationRecord{}, err
	}
	return orderOperationRecord, nil
}

// UpdateOrderOperationRecord 更新订单操作记录
func (r *OrderOperationRecordRepoImpl) UpdateRecord(saleBillUuid uint64, record model.SaleBillOperationRecord) error {
	if err := r.db.Model(&model.SaleBillOperationRecord{}).Where("uuid = ?", saleBillUuid).Updates(record).Error; err != nil {
		return err
	}
	return nil
}

// CreateOrderOperationRecord 创建订单操作记录
func (r *OrderOperationRecordRepoImpl) CreateRecord(record model.SaleBillOperationRecord) (uint64, error) {
	record.Uuid, _ = database.GetID()
	//
	if err := r.db.Create(&record).Error; err != nil {
		return 0, err
	}
	// todo 异常日志
	// OrderAbnormalLog::createLog(OrderAbnormalLog::SOURCE_ORDER, $orderId, $action, $data, $remark);
	//
	return record.Uuid, nil
}

// DeleteOrderOperationRecord 软删除订单操作记录
func (r *OrderOperationRecordRepoImpl) DeleteRecord(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleBillOperationRecord{}).Where("uuid = ?", saleBillUuid).Update("delete_time", uint(time.Now().Unix())).Error
}
