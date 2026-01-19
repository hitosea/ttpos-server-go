package event

import (
	"sync"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_statistics_sale_event_handler sync.Once

// StatisticsSaleEventHandler "统计销售"事件处理器
func StatisticsSaleEventHandler() {
	once_statistics_sale_event_handler.Do(func() {
		event.NewSystemBus().SubscribeStatisticsSaleEvent(func(payload event.StatisticsSalePayload) {
			service.NewStatisticsSrv().SaveSale(payload.Ctx, service.SaveSaleReq{
				SaleBillUuid: payload.SaleBillUuid,
				OnlyDelete:   payload.OnlyDelete,
			})
		})
	})
}
