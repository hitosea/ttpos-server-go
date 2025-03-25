package event

import (
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
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
			orderRecordRepo.CreateRecord(msg.SaleBillUuid, constant.OrderChangePrice, model.SaleOrderOperationRecord{
				Source:        msg.Source,
				Remark:        "改价",
				SaleBillUuid:  msg.SaleBillUuid,
				SaleOrderUuid: msg.SaleOrderUuid,
				OperatorUuid:  msg.StaffUuid,
			}, map[string]interface{}{
				"order_product_id": msg.OrderProductUuid,
				"product_id":       msg.OrderProductUuid,
				"product_name":     msg.OrderProductName,
				"total_num":        msg.OrderProductNum,
				"price":            msg.Price,
				"product_attr":     msg.AttributeNames,
			})
		})
	})
}
