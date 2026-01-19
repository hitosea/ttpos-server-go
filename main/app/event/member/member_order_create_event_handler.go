package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_create_member_sale_order_event_handler sync.Once
// createMemberSaleOrderEventHandler "创建外送订单"事件处理器
func CreateMemberSaleOrderEventHandler() {
	once_create_member_sale_order_event_handler.Do(func() {
		var orderSrv service.IOrderSrv
		_ = orderSrv

		event.NewSystemBus().SubscribeCreateMemberSaleOrderEvent(func(payload event.CreateMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:              payload.Source,
				Action:              constant.OrderCreateMemberSaleOrder,
				Remark:              "创建外送订单",
				SaleBillUuid:        payload.SaleBillUuid,
				SaleOrderUuid:       payload.SaleOrderUuid,
				MemberSaleOrderUuid: payload.MemberSaleOrderUuid,
				OperatorUuid:        payload.GetOperatorUuid(),
				MemberUuid:          payload.MemberUuid,
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCreateMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:创建外送订单 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
