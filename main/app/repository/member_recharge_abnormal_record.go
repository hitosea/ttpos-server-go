package repository

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMemberRechargeAbnormalRecordRepo 充值异常记录
type IMemberRechargeAbnormalRecordRepo interface {
	GetRecordInfo(cashierUuid uint64, dutyNo string) (*business_data_resp.AbnormalData, error)
	CreateRechargeAbnormalLog(obj model.MemberRechargeOrderOperationLog) error
}

type MemberRechargeAbnormalRecordRepoImpl struct {
	db *gorm.DB
}

func NewMemberRechargeAbnormalRecordRepo(db *gorm.DB) IMemberRechargeAbnormalRecordRepo {
	return NewMemberRechargeAbnormalRecordRepoImpl(db)
}

// NewMemberRechargeAbnormalRecordRepoImpl 创建新的充值异常记录仓库实现
func NewMemberRechargeAbnormalRecordRepoImpl(db *gorm.DB) *MemberRechargeAbnormalRecordRepoImpl {
	return &MemberRechargeAbnormalRecordRepoImpl{db: db}
}

// CreateRechargeAbnormalLog 创建充值异常记录
func (r *MemberRechargeAbnormalRecordRepoImpl) CreateRechargeAbnormalLog(obj model.MemberRechargeOrderOperationLog) error {
	rechargeOrder, err := NewMemberRechargeOrderRepo(r.db).GetRechargeOrderByUuid(obj.RechargeOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	//
	record := model.MemberRechargeOrderAbnormalRecord{
		RechargeOrderUuid: obj.RechargeOrderUuid,
		CashierUuid:       rechargeOrder.StaffUuid,
		DutyNo:            rechargeOrder.DutyNo,
		Action:            obj.Action,
	}
	//
	switch obj.Action {
	case constant.RechargeOrderActionRefund, constant.RechargeOrderActionReverseSettle:
		// 退款，反结账
		if err := r.db.Create(&record).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	// 返回
	return nil
}

// GetRecordInfo 获取订单异常记录信息
func (r *MemberRechargeAbnormalRecordRepoImpl) GetRecordInfo(cashierUuid uint64, dutyNo string) (*business_data_resp.AbnormalData, error) {
	var orderAbnormalRecord business_data_resp.AbnormalData
	// 退款次数（操作了几次，数量就是几，不是按照订单来）
	refundTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS refund_times`,
		constant.RechargeOrderActionRefund,
	)
	// 反结账次数
	reverseSettleTimes := fmt.Sprintf(
		`COUNT(CASE WHEN action = '%s' THEN 1 END) AS reverse_settle_times`,
		constant.RechargeOrderActionReverseSettle,
	)
	// 查询
	db := r.db.Model(&model.MemberRechargeOrderAbnormalRecord{})
	if cashierUuid != 0 {
		db = db.Where("cashier_uuid = ?", cashierUuid)
	}
	err := db.Select(refundTimes, reverseSettleTimes).
		Where("duty_no = ?", dutyNo).
		Where(ExcludeTestBusinessSQL("")).  // 排除测试营业时段的异常记录
		First(&orderAbnormalRecord).Error
	if err != nil {
		return &business_data_resp.AbnormalData{}, errors.WithMessage(err)
	}
	// 返回
	return &orderAbnormalRecord, nil
}
