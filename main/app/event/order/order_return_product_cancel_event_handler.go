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

var once_cancel_return_sale_order_product_event_handler sync.Once

// cancelReturnSaleOrderProductEventHandler "取消退菜"事件处理器
func CancelReturnSaleOrderProductEventHandler() {
	once_cancel_return_sale_order_product_event_handler.Do(func() {
		event.NewSystemBus().SubscribeCancelReturnSaleOrderProductEvent(func(payload event.CancelReturnSaleOrderProductPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderCancelRefundProduct,
				Remark:        "取消退菜",
				SaleBillUuid:  payload.ParentId,
				SaleOrderUuid: payload.OrderName,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCancelSaleOrderProductEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:退菜 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
