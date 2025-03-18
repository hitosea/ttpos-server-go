package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_checkout_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	checkoutSaleOrderEventHandler()
}

// checkoutSaleOrderEventHandler "结账"事件处理器
func checkoutSaleOrderEventHandler() {
	once_checkout_sale_order_event_handler.Do(func() {
		// 打印
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			_, err := printer.NewPrinterRepo(payload.Ctx).PrintingStatementOrder(
				constant.PrinterTemplateBilling,
				payload.SaleBill,
				payload.SaleOrderUuid,
				0,
			)
			if err != nil {
				logger.Logger.Error("SubscribeCheckoutZeroSaleOrderEvent process, PrintStatementOrder failed", zap.Any("payload", payload), zap.Error(err))
				return
			}
		})
		// 创建操作记录
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleBillOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderSettle,
				Remark:        "结账",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCheckoutZeroSaleOrderEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:结账 %+v", payload), zap.Uint64("record", uuid))
		})
		// 扣减库存
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			ReduceStock(db, payload.SaleBillUuid)
		})
	})
}
