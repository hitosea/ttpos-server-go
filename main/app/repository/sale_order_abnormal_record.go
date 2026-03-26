package repository

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IOrderAbnormalRecordRepo 订单异常记录
type IOrderAbnormalRecordRepo interface {
	GetRecordInfo(cashierUuid uint64, dutyNo string) (*business_data_resp.AbnormalData, error)
	CreateSaleOrderAbnormalLog(obj model.SaleOrderOperationRecord) error
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

// CreateSaleOrderAbnormalLog 创建订单异常记录
func (r *OrderAbnormalRecordRepoImpl) CreateSaleOrderAbnormalLog(obj model.SaleOrderOperationRecord) error {
	//
	bill, err := NewSaleBillRepo(r.db).GetSaleBillByUuid(obj.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	//
	record := model.SaleOrderAbnormalRecord{
		SaleBillUuid:  obj.SaleBillUuid,
		SaleOrderUuid: obj.SaleOrderUuid,
		CashierUuid:   obj.OperatorUuid,
		DutyNo:        utils.IfString(obj.GetDutyNo() != "", obj.GetDutyNo(), bill.DutyNo),
		Action:        obj.Action,
		SubAction:     "",
		Remark:        obj.Remark,
	}
	//
	var data map[string]interface{}
	if obj.Data != "" {
		if err := json.Unmarshal([]byte(obj.Data), &data); err != nil {
			fmt.Println("CreateSaleOrderAbnormalLog", err)
			logger.Logger.Error("CreateSaleOrderAbnormalLog", zap.String("error", err.Error()))
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
		if info := r.db.Model(&model.SaleOrderAbnormalRecord{}).Where("sale_bill_uuid = ? and action = ?", obj.SaleBillUuid, constant.OrderReverseSettle).First(&record); info.Error != nil {
			if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
				fmt.Println("CreateSaleOrderAbnormalLog-OrderReverseSettle", err)
				return errors.WithMessage(err)
			}
		}
	case constant.OrderRefund, constant.OrderProductMove:
		// 退款，转菜
		if err := r.db.Model(&model.SaleOrderAbnormalRecord{}).Create(&record).Error; err != nil {
			fmt.Println("CreateSaleOrderAbnormalLog-OrderReverseSettle", err)
			return errors.WithMessage(err)
		}
	}
	// 返回
	return nil
}

// GetRecordInfo 获取订单异常记录信息
func (r *OrderAbnormalRecordRepoImpl) GetRecordInfo(cashierUuid uint64, dutyNo string) (*business_data_resp.AbnormalData, error) {
	var orderAbnormalRecord business_data_resp.AbnormalData
	// 退菜次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次，如果有取消退菜，则要减去取消退菜次数）
	refundProductTimes := fmt.Sprintf(
		`COUNT( CASE WHEN action = '%s' THEN 1 END) AS refund_product_times`,
		constant.OrderRefundProduct,
	)
	// 赠菜次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次）
	productFreeTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS product_free_times`,
		constant.OrderProductFree,
	)
	// 退款次数（操作了几次，数量就是几，不是按照订单来）
	refundTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS refund_times`,
		constant.OrderRefund,
	)
	// 转菜次数
	productMoveTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS product_move_times`,
		constant.OrderProductMove,
	)
	// 单品改价次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次）
	changePriceTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS change_price_times`,
		constant.OrderChangePrice,
	)
	// 整单改价次数
	changeOrderPriceTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' and sub_action = '1' THEN 1 END) AS change_order_price_times`,
		constant.OrderDiscount,
	)
	// 整单折扣次数
	discountOrderTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' and sub_action = '2' THEN 1 END) AS discount_order_times`,
		constant.OrderDiscount,
	)
	// 整单抹零次数 + 整单抹零和结账抹零
	roundOrderTimes := fmt.Sprintf(
		`COUNT(CASE WHEN (action = '%s' and sub_action = '3') or (action = '%s') THEN 1 END) AS round_order_times`,
		constant.OrderDiscount,
		constant.OrderCheckoutDiscount,
	)
	// 免单次数（操作了几次，数量就是几，按照订单来）
	freeOrderTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS free_order_times`,
		constant.OrderFreeSale,
	)
	// 反结账次数（操作了几次，数量就是几，不是按照订单来）
	reverseSettleTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS reverse_settle_times`,
		constant.OrderReverseSettle,
	)
	// 查询
	db := r.db.Model(&model.SaleOrderAbnormalRecord{})
	if cashierUuid != 0 {
		db = db.Where("cashier_uuid = ?", cashierUuid)
	}
	err := db.Select(
		refundProductTimes,
		productFreeTimes,
		refundTimes,
		productMoveTimes,
		changePriceTimes,
		changeOrderPriceTimes,
		discountOrderTimes,
		roundOrderTimes,
		freeOrderTimes,
		reverseSettleTimes,
	).
		Where("duty_no = ?", dutyNo).
		Where(ExcludeTestBusinessSQL("")). // 排除测试营业时段的异常记录
		First(&orderAbnormalRecord).Error
	if err != nil {
		return &business_data_resp.AbnormalData{}, errors.WithMessage(err)
	}

	// 获取会员充值异常记录
	abnormalData, err := NewMemberRechargeAbnormalRecordRepo(r.db).GetRecordInfo(cashierUuid, dutyNo)
	if err != nil {
		return &business_data_resp.AbnormalData{}, errors.WithMessage(err)
	}

	// 加上会员充值异常记录
	orderAbnormalRecord.RefundTimes = orderAbnormalRecord.RefundTimes + abnormalData.RefundTimes
	orderAbnormalRecord.ReverseSettleTimes = orderAbnormalRecord.ReverseSettleTimes + abnormalData.ReverseSettleTimes

	// 返回
	return &orderAbnormalRecord, nil
}
