package repository

import (
	"encoding/json"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderAbnormalRecordRepo 订单异常记录
type IOrderAbnormalRecordRepo interface {
	GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleBillAbnormalRecord, int64, error)
	GetRecordLists(saleBillUuid uint64) ([]model.SaleBillAbnormalRecord, error)
	GetRecordInfo(saleBillUuid uint64) (model.SaleBillAbnormalRecord, error)
	UpdateRecord(saleBillUuid uint64, record model.SaleBillAbnormalRecord) error
	CreateRecord(saleBillUuid uint64, Action string, record model.SaleBillAbnormalRecord, data interface{}) (uint64, error)
	CreateSaleBillAbnormalRecord(model model.SaleBillAbnormalRecord) (uint64, error)
	DeleteRecord(saleBillUuid uint64) error
}

type OrderAbnormalRecordRepoImpl struct {
	db *gorm.DB
}

func NewOrderAbnormalRecordRepo(db *gorm.DB) IOrderAbnormalRecordRepo {
	return NewOrderAbnormalRecordRepoImpl(db)
}

// NewOrderAbnormalRecordRepoImpl 创建新的订单异常记录仓库实现
func NewOrderAbnormalRecordRepoImpl(db *gorm.DB) *OrderAbnormalRecordRepoImpl {
	return &OrderAbnormalRecordRepoImpl{db: db}
}

func (r *OrderAbnormalRecordRepoImpl) CreateSaleBillAbnormalRecord(obj model.SaleBillAbnormalRecord) (uint64, error) {
	obj.SetNil()
	if err := r.db.Model(&model.SaleBillAbnormalRecord{}).Create(&obj).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return obj.Uuid, nil
}

// GetRecordList 获取订单异常记录列表
func (r *OrderAbnormalRecordRepoImpl) GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleBillAbnormalRecord, int64, error) {
	var orderAbnormalRecords []model.SaleBillAbnormalRecord
	var total int64

	query := r.db.Model(&model.SaleBillAbnormalRecord{}).Where("delete_time = ?", 0)

	for _, opt := range opts {
		query = opt(query)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orderAbnormalRecords).Error
	return orderAbnormalRecords, total, errors.WithMessage(err)
}

// GetRecordLists 获取订单异常记录列表
func (r *OrderAbnormalRecordRepoImpl) GetRecordLists(saleBillUuid uint64) ([]model.SaleBillAbnormalRecord, error) {
	var orderAbnormalRecords []model.SaleBillAbnormalRecord
	err := r.db.Model(&model.SaleBillAbnormalRecord{}).Preload("Operator").
		Where("delete_time = ?", 0).Where("sale_bill_uuid = ?", saleBillUuid).Find(&orderAbnormalRecords).Error
	return orderAbnormalRecords, errors.WithMessage(err)
}

// GetRecordInfo 获取订单异常记录信息
func (r *OrderAbnormalRecordRepoImpl) GetRecordInfo(saleBillUuid uint64) (model.SaleBillAbnormalRecord, error) {
	var orderAbnormalRecord model.SaleBillAbnormalRecord
	if err := r.db.Model(&model.SaleBillAbnormalRecord{}).Where("uuid = ?", saleBillUuid).First(&orderAbnormalRecord).Error; err != nil {
		return model.SaleBillAbnormalRecord{}, errors.WithMessage(err)
	}
	return orderAbnormalRecord, nil
}

// UpdateRecord 更新订单异常记录
func (r *OrderAbnormalRecordRepoImpl) UpdateRecord(saleBillUuid uint64, record model.SaleBillAbnormalRecord) error {
	if err := r.db.Model(&model.SaleBillAbnormalRecord{}).Where("uuid = ?", saleBillUuid).Updates(record).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateRecord 创建订单异常记录
func (r *OrderAbnormalRecordRepoImpl) CreateRecord(saleBillUuid uint64, Action string, record model.SaleBillAbnormalRecord, data interface{}) (uint64, error) {
	record.Action = Action
	record.SaleBillUuid = saleBillUuid

	// 将 data 转换为 JSON 字符串
	if data != nil {
		dataJson, err := json.Marshal(data)
		if err != nil {
			return 0, errors.WithMessage(err)
		}
		record.Data = string(dataJson)
	}

	if err := r.db.Create(&record).Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	return record.Uuid, nil
}

// DeleteRecord 软删除订单异常记录
func (r *OrderAbnormalRecordRepoImpl) DeleteRecord(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleBillAbnormalRecord{}).Where("uuid = ?", saleBillUuid).Update("delete_time", uint(time.Now().Unix())).Error
}
