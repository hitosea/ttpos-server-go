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

var once_return_order_event_handler sync.Once

// returnOrderEventHandler 订单"退款"事件处理器
func ReturnOrderEventHandler() {
	once_return_order_event_handler.Do(func() {
		event.NewSystemBus().SubscribeReturnOrderEvent(func(payload event.ReturnOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderRefund,
				Remark:        "退款",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			payload.IsSplitOrder = payload.SaleBill.IsSplit()
			for i, saleOrder := range payload.SaleBill.SaleOrders {
				if saleOrder.Uuid == payload.SaleOrderUuid {
					payload.Index = i + 1
				}
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeReturnOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:退款 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
