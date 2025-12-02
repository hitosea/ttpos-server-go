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

var once_activity_sale_order_event_handler sync.Once

// activitySaleOrderEventHandler "满减活动"事件处理器
func activitySaleOrderEventHandler() {
	once_activity_sale_order_event_handler.Do(func() {
		event.NewSystemBus().SubscribeActivitySaleOrderEvent(func(payload event.ActivitySaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)

			action := constant.OrderActivity
			remark := "满减活动"

			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        action,
				Remark:        remark,
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeActivitySaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:满减活动 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
