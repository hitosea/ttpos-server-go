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

var once_sent_cooking_event_handler sync.Once

// init 自动注册"添加销售账单记录"事件处理器
func init() {
	// 只初始化一次
	sentCookingEventHandler()
}

// sentCookingEventHandler "送厨"事件处理器
func sentCookingEventHandler() {
	once_sent_cooking_event_handler.Do(func() {
		event.NewSystemBus().SubscribeSentCookingEvent(func(payload event.SentCookingPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleBillOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderSendKitchen,
				Remark:        "送厨",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeSentCookingEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:送厨 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
