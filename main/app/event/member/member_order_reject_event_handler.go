package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_reject_member_sale_order_event_handler sync.Once

// rejectMemberSaleOrderEventHandler "外送拒单"事件处理器
func RejectMemberSaleOrderEventHandler() {
	once_reject_member_sale_order_event_handler.Do(func() {
		// 外送拒单后，自动退款
		event.NewSystemBus().SubscribeRejectMemberSaleOrderEvent(func(payload event.RejectMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			// +++++++++++++++++++++++ 设置SaleBill的当班编号
			SetDutyNoForSaleBill(db, payload.BasePayload)
		})

		// 外送拒单后，创建操作记录
		event.NewSystemBus().SubscribeRejectMemberSaleOrderEvent(func(payload event.RejectMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:              payload.Source,
				Action:              constant.OrderCancelMemberSaleOrder,
				Remark:              "订单拒单",
				SaleBillUuid:        payload.SaleBillUuid,
				SaleOrderUuid:       payload.SaleOrderUuid,
				H5OrderUuid:         payload.H5OrderUuid,
				OperatorUuid:        payload.GetOperatorUuid(),
				MemberUuid:          payload.MemberUuid,
				MemberSaleOrderUuid: payload.MemberSaleOrderUuid,
				Data:                utils.ToJson(payload.Data),
			}
			// 创建操作记录
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCancelOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:订单取消 %+v", payload), zap.Uint64("record", uuid))

		})
	})
}
