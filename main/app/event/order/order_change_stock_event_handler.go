package event

import (
	"sync"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"
)

var once_change_stock_event_handler sync.Once

// changeStockEventHandler "变成库存" 事件处理器
func ChangeStockEventHandler() {
	once_change_stock_event_handler.Do(func() {
		event.NewSystemBus().SubscribeChangeStockEvent(func(payload event.ChangeStockPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			// 发布"加库存"事件
			utils.Go(func() {
				AddStock(db, payload.SaleBillUuid)
				ReduceStock(db, payload.SaleBillUuid)
			})
		})
	})
}
