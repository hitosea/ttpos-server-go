package repository

import (
	"encoding/json"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderProductCancelReasonRepo 定义销售订单商品取消原因仓库接口
type IOrderProductCancelReasonRepo interface {
	CreateBatch(boms []*model.SaleOrderProductCancelReason) error // 批量创建
}

// orderProductCancelReasonRepo 销售订单商品取消原因仓库
type orderProductCancelReasonRepo struct {
	db *gorm.DB
}

// NewOrderProductCancelReasonRepo 实例化销售订单商品取消原因仓库
func NewOrderProductCancelReasonRepo(db *gorm.DB) IOrderProductCancelReasonRepo {
	return NewOrderProductCancelReasonRepoImpl(db)
}

// NewOrderProductCancelReasonRepoImpl 实例化销售订单商品取消原因仓库实现
func NewOrderProductCancelReasonRepoImpl(db *gorm.DB) IOrderProductCancelReasonRepo {
	return &orderProductCancelReasonRepo{db: db}
}

// CreateBatch 批量创建
func (o *orderProductCancelReasonRepo) CreateBatch(boms []*model.SaleOrderProductCancelReason) error {
	return o.db.Create(&boms).Error
}

// CreateRecord 创建订单操作记录
func (r *orderProductCancelReasonRepo) CreateRecord(saleBillUuid uint64, Action string, record model.SaleBillOperationRecord, data interface{}) (uint64, error) {
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
