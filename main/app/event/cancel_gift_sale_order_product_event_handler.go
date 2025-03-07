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

var once_cancel_gift_sale_order_product_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	cancelGiftSaleOrderProductEventHandler()
}

// cancelGiftSaleOrderProductEventHandler "取消赠菜"事件处理器
func cancelGiftSaleOrderProductEventHandler() {
	once_cancel_gift_sale_order_product_event_handler.Do(func() {
		event.NewSystemBus().SubscribeCancelGiftSaleOrderProductEvent(func(payload event.CancelGiftSaleOrderProductPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleBillOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderCancelProductFree,
				Remark:        "取消赠菜",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCancelGiftSaleOrderProductEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:取消赠菜 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
