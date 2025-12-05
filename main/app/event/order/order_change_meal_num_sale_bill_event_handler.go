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

var once_change_meal_num_sale_bill_event_handler sync.Once

// changeMealNumSaleBillEventHandler "修改桌台就餐人数"事件处理器
func ChangeMealNumSaleBillEventHandler() {
	once_change_meal_num_sale_bill_event_handler.Do(func() {
		event.NewSystemBus().SubscribeChangeMealNumSaleBillEvent(func(payload event.ChangeMealNumSaleBillPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:       payload.Source,
				Action:       constant.OrderUpdateMealNum,
				Remark:       "修改桌台就餐人数",
				SaleBillUuid: payload.SaleBillUuid,
				OperatorUuid: payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeChangeMealNumSaleBillEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:修改桌台就餐人数 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
