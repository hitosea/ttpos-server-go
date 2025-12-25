package event

import (
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/pkg/utils"
)

var once_takeout_order_ready_event_handler sync.Once

// TakeoutOrderReadyEventHandler 外卖订单准备完成事件处理器（呼叫骑手）
func TakeoutOrderReadyEventHandler() {
	once_takeout_order_ready_event_handler.Do(func() {
		// 订阅外卖订单准备完成事件
		event.GetDispatcher().Register(&takeoutOrderReadyEventSubscriber{})
	})
}

// takeoutOrderReadyEventSubscriber 外卖订单准备完成事件订阅者
type takeoutOrderReadyEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderReadyEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.ready"}
}

// Handle 处理事件
func (s *takeoutOrderReadyEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderReadyEvent, ok := domainEvent.(event.OrderReadyEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderReadyEvent.CompanyUuid,
			orderReadyEvent.TakeoutOrderUuid,
			orderReadyEvent.Platform,
			orderReadyEvent.ShortOrderNumber,
			"ready",
			map[string]any{},
		)
	})

	return nil
}
