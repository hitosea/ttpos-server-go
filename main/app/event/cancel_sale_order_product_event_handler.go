package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

var once_cancel_sale_order_product_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	cancelSaleOrderProductEventHandler()
}

// cancelSaleOrderProductEventHandler "退菜"事件处理器
func cancelSaleOrderProductEventHandler() {
	once_cancel_sale_order_product_event_handler.Do(func() {
		event.NewSystemBus().SubscribeCancelSaleOrderProductEvent(func(payload event.CancelSaleOrderProductPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			// 创建退菜操作记录
			go func() {
				orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
				record := model.SaleBillOperationRecord{
					Source:        payload.Source,
					Action:        constant.OrderRefundProduct,
					Remark:        "退菜",
					SaleBillUuid:  payload.SaleBillUuid,
					SaleOrderUuid: payload.SaleOrderUuid,
					OperatorUuid:  payload.GetOperatorUuid(),
				}
				record.Data = payload.ToJsonString()
				uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
				if err != nil {
					logger.Logger.Error("SubscribeCancelSaleOrderProductEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
					return
				}
				logger.Logger.Info(fmt.Sprintf("操作记录:退菜 %+v", payload), zap.Uint64("record", uuid))
			}()

			// 创建退菜打印记录
			go func() {
				products := printer_model.OrderProduct{}
				copier.Copy(&products, payload)
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypeBackFood,
					payload.SaleBillUuid,
					[]printer_model.OrderProduct{products},
				)
			}()
		})
	})
}
