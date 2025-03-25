package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IOrderAbnormalRecordRepo 订单异常记录
type IOrderAbnormalRecordRepo interface {
	GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleOrderAbnormalRecord, int64, error)
	GetRecordLists(saleBillUuid uint64) ([]model.SaleOrderAbnormalRecord, error)
	GetRecordInfo(saleBillUuid uint64) (model.SaleOrderAbnormalRecord, error)
	UpdateRecord(saleBillUuid uint64, record model.SaleOrderAbnormalRecord) error
	CreateSaleOrderAbnormalLog(Source, dutyNo string, obj model.SaleOrderOperationRecord) (uint64, error)
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

func (r *OrderAbnormalRecordRepoImpl) CreateSaleOrderAbnormalLog(Source, dutyNo string, obj model.SaleOrderOperationRecord) (uint64, error) {

	var record model.SaleOrderAbnormalRecord
	record.SetNil()
	record.Source = Source
	record.SaleBillUuid = obj.SaleBillUuid
	record.SaleOrderUuid = obj.SaleOrderUuid
	record.CashierUuid = obj.OperatorUuid
	record.DutyNo = dutyNo
	record.Action = obj.Action
	// record.SubAction = obj.SubAction
	record.Remark = obj.Remark

	// constant.OrderProductFree
	fmt.Println("CreateSaleOrderAbnormalLog", utils.ToJsonString(obj))

	if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&obj).Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	// 返回
	return obj.Uuid, nil
}

// GetRecordList 获取订单异常记录列表
func (r *OrderAbnormalRecordRepoImpl) GetRecordList(pageNo, pageSize int, opts ...DBOption) ([]model.SaleOrderAbnormalRecord, int64, error) {
	var orderAbnormalRecords []model.SaleOrderAbnormalRecord
	var total int64

	query := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("delete_time = ?", 0)

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
func (r *OrderAbnormalRecordRepoImpl) GetRecordLists(saleBillUuid uint64) ([]model.SaleOrderAbnormalRecord, error) {
	var orderAbnormalRecords []model.SaleOrderAbnormalRecord
	err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Preload("Operator").
		Where("delete_time = ?", 0).Where("sale_bill_uuid = ?", saleBillUuid).Find(&orderAbnormalRecords).Error
	return orderAbnormalRecords, errors.WithMessage(err)
}

// GetRecordInfo 获取订单异常记录信息
func (r *OrderAbnormalRecordRepoImpl) GetRecordInfo(saleBillUuid uint64) (model.SaleOrderAbnormalRecord, error) {
	var orderAbnormalRecord model.SaleOrderAbnormalRecord
	if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("uuid = ?", saleBillUuid).First(&orderAbnormalRecord).Error; err != nil {
		return model.SaleOrderAbnormalRecord{}, errors.WithMessage(err)
	}
	return orderAbnormalRecord, nil
}

// UpdateRecord 更新订单异常记录
func (r *OrderAbnormalRecordRepoImpl) UpdateRecord(saleOrderUuid uint64, record model.SaleOrderAbnormalRecord) error {
	if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("uuid = ?", saleOrderUuid).Updates(record).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// DeleteRecord 软删除订单异常记录
func (r *OrderAbnormalRecordRepoImpl) DeleteRecord(saleOrderUuid uint64) error {
	return r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("uuid = ?", saleOrderUuid).Update("delete_time", uint(time.Now().Unix())).Error
}
