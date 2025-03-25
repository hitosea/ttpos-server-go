package repository

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IOrderAbnormalRecordRepo 订单异常记录
type IOrderAbnormalRecordRepo interface {
	GetRecordInfo(saleBillUuid uint64) (model.SaleOrderAbnormalRecord, error)
	CreateSaleOrderAbnormalLog(Source, dutyNo string, obj model.SaleOrderOperationRecord) error
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

func (r *OrderAbnormalRecordRepoImpl) CreateSaleOrderAbnormalLog(Source, dutyNo string, obj model.SaleOrderOperationRecord) error {
	record := model.SaleOrderAbnormalRecord{
		Source:        Source,
		SaleBillUuid:  obj.SaleBillUuid,
		SaleOrderUuid: obj.SaleOrderUuid,
		CashierUuid:   obj.OperatorUuid,
		DutyNo:        dutyNo,
		Action:        obj.Action,
		SubAction:     "",
		Remark:        obj.Remark,
	}

	var data map[string]interface{}
	if obj.Data != "" {
		if err := json.Unmarshal([]byte(obj.Data), &data); err != nil {
			fmt.Println("CreateSaleOrderAbnormalLog", err)
			logger.Logger.Info("CreateSaleOrderAbnormalLog", zap.String("error", err.Error()))
			return errors.WithMessage(err)
		}
	}

	//
	switch obj.Action {
	case constant.OrderRefundProduct:
		// 退菜 对一个商品反复操作，记录为1次
		record.Sign = data["sign"].(string)
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_order_uuid = ? and action = ? and sign = ?", obj.SaleOrderUuid, obj.Action, record.Sign).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderRefundProduct", err)
				return errors.WithMessage(err)
			}
		}
	case constant.OrderCancelRefundProduct:
		// 取消退菜 重置所有所选赠菜操作
		record.Sign = data["sign"].(string)
		if err := r.db.Where("sale_order_uuid = ? and action = ? and sign = ?", obj.SaleOrderUuid, constant.OrderRefundProduct, record.Sign).Delete(&model.SaleOrderAbnormalRecord{}).Error; err != nil {
			fmt.Println("CreateSaleOrderAbnormalLog-OrderCancelRefundProduct", err)
			return errors.WithMessage(err)
		}
	case constant.OrderProductFree:
		// 赠菜 对一个商品反复操作，记录为1次
		record.Sign = fmt.Sprintf("%.0f", data["order_product_id"])
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_order_uuid = ? and action = ? and sign = ?", obj.SaleOrderUuid, obj.Action, record.Sign).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderProductFree", err)
				return errors.WithMessage(err)
			}
		}
	case constant.OrderCancelProductFree:
		// 取消赠菜 重置所有所选赠菜操作
		record.Sign = fmt.Sprintf("%.0f", data["order_product_id"])
		if err := r.db.Where("sale_order_uuid = ? and action = ? and sign = ?", obj.SaleOrderUuid, constant.OrderProductFree, record.Sign).Delete(&model.SaleOrderAbnormalRecord{}).Error; err != nil {
			fmt.Println("CreateSaleOrderAbnormalLog-OrderCancelProductFree", err)
			return errors.WithMessage(err)
		}
	case constant.OrderChangePrice:
		// 单品改价 对一个商品反复操作，记录为1次
		record.Sign = fmt.Sprintf("%.0f", data["order_product_id"])
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_order_uuid = ? and action = ? and sign = ?", obj.SaleOrderUuid, obj.Action, record.Sign).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderChangePrice", err)
				return errors.WithMessage(err)
			}
		}
	case constant.OrderDiscount:
		// 优惠折扣 拆单优惠折扣在主单中重复只算一次，记录分开记，查询时再去重
		if data["discount_type"] != nil {
			record.SubAction = fmt.Sprintf("%.0f", data["discount_type"])
		}
		// 查询该订单是否已经有优惠折扣
		if record.SubAction == "1" || record.SubAction == "2" {
			if err := r.db.Where("sale_order_uuid = ? and action = ? and sub_action in (?)", obj.SaleOrderUuid, obj.Action, []int{1, 2}).Delete(&model.SaleOrderAbnormalRecord{}).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderDiscount", err)
				return errors.WithMessage(err)
			}
		} else if record.SubAction == "3" {
			if err := r.db.Where("sale_order_uuid = ? and action = ? and sub_action in (?)", obj.SaleOrderUuid, obj.Action, []int{1, 3}).Delete(&model.SaleOrderAbnormalRecord{}).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderDiscount", err)
				return errors.WithMessage(err)
			}
		}
		record.Sign = fmt.Sprintf("%d-%s-%s", obj.SaleOrderUuid, obj.Action, record.SubAction)
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_order_uuid = ? and action = ? and sign = ?", obj.SaleOrderUuid, obj.Action, record.Sign).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderDiscount", err)
				return errors.WithMessage(err)
			}
		}
	case constant.OrderCancelDiscount:
		// 撤销优惠折扣 重置所有优惠折扣操作
		if err := r.db.Where("sale_order_uuid = ? and action = ?", obj.SaleOrderUuid, constant.OrderDiscount).Delete(&model.SaleOrderAbnormalRecord{}).Error; err != nil {
			fmt.Println("CreateSaleOrderAbnormalLog-OrderCancelDiscount", err)
			return errors.WithMessage(err)
		}
	case constant.OrderCheckoutDiscount:
		// 结账手动抹零 重置所有优惠折扣操作
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_order_uuid = ? and action = ?", obj.SaleOrderUuid, record.Action).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderDiscount", err)
				return errors.WithMessage(err)
			}
		}
	case constant.OrderFreeSale:
		// 优惠折扣 拆单优惠折扣在主单中重复只算一次，记录分开记，查询时再去重
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_order_uuid = ? and action = ?", obj.SaleOrderUuid, obj.Action).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderFreeSale", err)
				return errors.WithMessage(err)
			}
		}
	case constant.OrderReverseSettle:
		// 反结账 重置该订单免单操作
		if err := r.db.Where("sale_order_uuid = ? and action = ?", obj.SaleOrderUuid, constant.OrderFreeSale).Delete(&model.SaleOrderAbnormalRecord{}).Error; err != nil {
			fmt.Println("CreateSaleOrderAbnormalLog-OrderReverseSettle", err)
			return errors.WithMessage(err)
		}
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_order_uuid = ? and action = ?", obj.SaleOrderUuid, constant.OrderReverseSettle).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderReverseSettle", err)
				return errors.WithMessage(err)
			}
		}
	}

	// 返回
	return nil
}

// GetRecordInfo 获取订单异常记录信息
func (r *OrderAbnormalRecordRepoImpl) GetRecordInfo(saleBillUuid uint64) (model.SaleOrderAbnormalRecord, error) {
	var orderAbnormalRecord model.SaleOrderAbnormalRecord
	if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("uuid = ?", saleBillUuid).First(&orderAbnormalRecord).Error; err != nil {
		return model.SaleOrderAbnormalRecord{}, errors.WithMessage(err)
	}
	return orderAbnormalRecord, nil
}
