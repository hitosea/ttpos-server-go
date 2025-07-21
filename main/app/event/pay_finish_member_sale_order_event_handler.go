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

var once_pay_finish_member_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	payFinishMemberSaleOrderEventHandler()
}

// payFinishMemberSaleOrderEventHandler "支付完成会员端销售订单"事件处理器
func payFinishMemberSaleOrderEventHandler() {
	once_pay_finish_member_sale_order_event_handler.Do(func() {
		// 创建订单操作日志
		event.NewSystemBus().SubscribePayFinishMemberSaleOrderEvent(func(payload event.PayFinishMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:              payload.Source,
				Action:              constant.OrderPayFinishMemberSaleOrder,
				Remark:              "订单支付成功",
				SaleBillUuid:        payload.SaleBillUuid,
				SaleOrderUuid:       payload.SaleOrderUuid,
				OperatorUuid:        payload.GetOperatorUuid(),
				MemberUuid:          payload.MemberUuid,
				MemberSaleOrderUuid: payload.MemberSaleOrderUuid,
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribePayFinishMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:订单支付成功 %+v", payload), zap.Uint64("record", uuid))
		})

		// 修改订单状态
		event.NewSystemBus().SubscribePayFinishMemberSaleOrderEvent(func(payload event.PayFinishMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRepo := repository.NewMemberSaleOrderRepo(db)
			// 更新订单状态为“待商家接单”
			if err := orderRepo.UpdateMemberSaleOrderPendingMerchantAccept(payload.MemberSaleOrderUuid); err != nil {
				logger.Logger.Error("SubscribePayFinishMemberSaleOrderEvent process, UpdateMemberSaleOrderPendingMerchantAccept failed", zap.Any("payload", utils.ToJson(payload)), zap.Error(err))
				return
			}
		})
	})
}
