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

var once_show_sale_bill_event_handler sync.Once

// showSaleBillEventHandler "取单"事件处理器
func ShowSaleBillEventHandler() {
	once_show_sale_bill_event_handler.Do(func() {
		event.NewSystemBus().SubscribeShowSaleBillEvent(func(payload event.ShowSaleBillPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:       payload.Source,
				Action:       constant.OrderPickOrder,
				Remark:       "取单",
				SaleBillUuid: payload.SaleBillUuid,
				OperatorUuid: payload.GetOperatorUuid(),
			}
			record.Data = ""
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeShowSaleBillEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:取单 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
