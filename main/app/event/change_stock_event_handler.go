package event

import (
	"sync"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_change_stock_event_handler sync.Once

// init 自动注册"加库存"事件处理器
func init() {
	// 只初始化一次
	changeStockEventHandler()
}

// acceptH5OrderEventHandler "接单"事件处理器
func changeStockEventHandler() {
	once_change_stock_event_handler.Do(func() {
		event.NewSystemBus().SubscribeChangeStockEvent(func(payload event.ChangeStockPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			// 发布"加库存"事件
			go func() {
				AddStock(db, payload.SaleBillUuid)
				ReduceStock(db, payload.SaleBillUuid)
			}()
		})
	})
}
