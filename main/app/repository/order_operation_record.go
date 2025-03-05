package repository

import (
	"encoding/json"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderOperationRecordRepo 订单操作记录
type IOrderOperationRecordRepo interface {
	GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleBillOperationRecord, int64, error)
	GetRecordLists(saleBillUuid uint64) ([]model.SaleBillOperationRecord, error)
	GetRecordInfo(saleBillUuid uint64) (model.SaleBillOperationRecord, error)
	UpdateRecord(saleBillUuid uint64, record model.SaleBillOperationRecord) error
	CreateRecord(saleBillUuid uint64, Action string, record model.SaleBillOperationRecord, data interface{}) (uint64, error)
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

// GetRecordList 获取订单操作记录列表
func (r *OrderOperationRecordRepoImpl) GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleBillOperationRecord, int64, error) {
	var orderOperationRecords []model.SaleBillOperationRecord
	var total int64

	query := r.db.Model(&model.SaleBillOperationRecord{}).Where("delete_time = ?", 0)

	for _, opt := range opts {
		query = opt(query)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orderOperationRecords).Error
	return orderOperationRecords, total, err
}

// GetRecordLists 获取订单操作记录列表
func (r *OrderOperationRecordRepoImpl) GetRecordLists(saleBillUuid uint64) ([]model.SaleBillOperationRecord, error) {
	var orderOperationRecords []model.SaleBillOperationRecord
	err := r.db.Model(&model.SaleBillOperationRecord{}).Where("delete_time = ?", 0).Where("sale_bill_uuid = ?", saleBillUuid).Find(&orderOperationRecords).Error
	return orderOperationRecords, err
}

// GetRecordInfo 获取订单操作记录信息
func (r *OrderOperationRecordRepoImpl) GetRecordInfo(saleBillUuid uint64) (model.SaleBillOperationRecord, error) {
	var orderOperationRecord model.SaleBillOperationRecord
	if err := r.db.Model(&model.SaleBillOperationRecord{}).Where("uuid = ?", saleBillUuid).First(&orderOperationRecord).Error; err != nil {
		return model.SaleBillOperationRecord{}, err
	}
	return orderOperationRecord, nil
}

// UpdateRecord 更新订单操作记录
func (r *OrderOperationRecordRepoImpl) UpdateRecord(saleBillUuid uint64, record model.SaleBillOperationRecord) error {
	if err := r.db.Model(&model.SaleBillOperationRecord{}).Where("uuid = ?", saleBillUuid).Updates(record).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateRecord 创建订单操作记录
func (r *OrderOperationRecordRepoImpl) CreateRecord(saleBillUuid uint64, Action string, record model.SaleBillOperationRecord, data interface{}) (uint64, error) {
	record.Action = Action
	record.SaleBillUuid = saleBillUuid

	// 将 data 转换为 JSON 字符串
	if data != nil {
		dataJson, err := json.Marshal(data)
		if err != nil {
			return 0, err
		}
		record.Data = string(dataJson)
	}

	if err := r.db.Create(&record).Error; err != nil {
		return 0, err
	}

	// // 添加异常日志
	// orderRecordRepo.CreateRecord(req.SaleBillUuid, constant.OrderChangePrice, model.SaleBillOperationRecord{
	// 	Source:        source,
	// 	Remark:        "改价",
	// 	SaleOrderUuid: req.SaleOrderUuid,
	// 	OperatorUuid:  staffUuid,
	// }, map[string]interface{}{
	// 	"remark": "",
	// })

	return record.Uuid, nil
}

// DeleteRecord 软删除订单操作记录
func (r *OrderOperationRecordRepoImpl) DeleteRecord(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleBillOperationRecord{}).Where("uuid = ?", saleBillUuid).Update("delete_time", uint(time.Now().Unix())).Error
}
