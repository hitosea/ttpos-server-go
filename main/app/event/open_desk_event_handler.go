package event

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
)

var once_open_desk_event_handler sync.Once

// init 自动注册"添加销售账单记录"事件处理器
func init() {
	// 只初始化一次
	openDeskEventHandler()
}

// openDeskEventHandler "开台"事件处理器
func openDeskEventHandler() {
	once_open_desk_event_handler.Do(func() {
		event.NewSystemBus().SubscribeOpenDeskEvent(func(payload event.OpenDeskPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleBillOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderOpenTable,
				Remark:        "开台",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeOpenDeskEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("%+v", payload), zap.Uint64("record", uuid))
		})
	})
}
