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

var once_cancel_member_order_event_handler sync.Once

// CancelMemberOrderEventHandler "订单取消"事件处理器
func CancelMemberOrderEventHandler() {
	once_cancel_member_order_event_handler.Do(func() {

		// 订单取消 - 创建操作记录
		event.NewSystemBus().SubscribeCancelMemberOrderEvent(func(payload event.CancelMemberOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:              payload.Source,
				Action:              constant.OrderCancelMemberSaleOrder,
				Remark:              "订单取消",
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
