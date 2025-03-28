package event

import (
	"sync"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_statistics_sale_event_handler sync.Once

// init 自动注册"统计销售"事件处理器
func init() {
	// 只初始化一次
	statisticsSaleEventHandler()
}

// splitOrderEventHandler "拆单"事件处理器
func statisticsSaleEventHandler() {
	once_statistics_sale_event_handler.Do(func() {
		event.NewSystemBus().SubscribeStatisticsSaleEvent(func(payload event.StatisticsSalePayload) {
			service.NewStatisticsSrv().SaveSale(payload.Ctx, service.SaveSaleReq{
				SaleBill:   payload.SaleBill,
				OnlyDelete: payload.OnlyDelete,
			})
		})
	})
}
