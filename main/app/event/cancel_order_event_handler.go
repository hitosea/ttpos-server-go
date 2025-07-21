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

var once_cancel_order_event_handler sync.Once

// init 自动注册"添加销售账单记录"事件处理器
func init() {
	// 只初始化一次
	cancelOrderEventHandler()
}

// cancelOrderEventHandler "整单取消"事件处理器
func cancelOrderEventHandler() {
	once_cancel_order_event_handler.Do(func() {

		// 整单取消 - 创建操作记录
		event.NewSystemBus().SubscribeCancelOrderEvent(func(payload event.CancelOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderOrderCancel,
				Remark:        "整单取消",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				H5OrderUuid:   payload.H5OrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = ""
			if payload.Ctx.GetStaff() != (model.Staff{}) {
				record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			}
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCancelOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:整单取消 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
