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

var once_add_sale_bill_record_event_handler sync.Once

// init 自动注册"添加销售账单记录"事件处理器
func init() {
	// 只初始化一次
	addSaleBillRecordEventHandler()
}

// addSaleBillRecordEventHandler "添加销售账单记录"事件处理器
func addSaleBillRecordEventHandler() {
	once_add_sale_bill_record_event_handler.Do(func() {
		event.NewSystemBus().SubscribeAddSaleBillRecordEvent(func(msg event.AddSaleBillRecordPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(msg.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			// 添加操作日志
			record := model.SaleOrderOperationRecord{
				Source:        msg.Source,
				Action:        constant.OrderChangePrice,
				Remark:        "改价",
				SaleBillUuid:  msg.SaleBillUuid,
				SaleOrderUuid: msg.SaleOrderUuid,
				OperatorUuid:  msg.StaffUuid,
				Data: utils.ToJson(map[string]interface{}{
					"order_product_id": msg.OrderProductUuid,
					"product_id":       msg.OrderProductUuid,
					"product_name":     msg.OrderProductName,
					"total_num":        msg.OrderProductNum,
					"price":            msg.Price,
					"product_attr":     msg.AttributeNames,
				}),
			}
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeAddSaleBillRecordEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:改价 %+v", msg), zap.Uint64("record", uuid))
		})
	})
}
