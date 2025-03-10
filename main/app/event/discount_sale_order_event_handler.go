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

var once_discount_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	discountSaleOrderEventHandler()
}

// discountSaleOrderEventHandler "优惠折扣"事件处理器
func discountSaleOrderEventHandler() {
	once_discount_sale_order_event_handler.Do(func() {
		event.NewSystemBus().SubscribeDiscountSaleOrderEvent(func(payload event.DiscountSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleBillOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderDiscount,
				Remark:        "优惠折扣",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeDiscountSaleOrderEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:优惠折扣-打折 %+v", payload), zap.Uint64("record", uuid))
		})
	})
}
