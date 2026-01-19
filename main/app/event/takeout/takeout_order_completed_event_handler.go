package event

import (
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/pkg/utils"
)

var once_takeout_order_completed_event_handler sync.Once

// TakeoutOrderCompletedEventHandler 外卖订单完成事件处理器
func TakeoutOrderCompletedEventHandler() {
	once_takeout_order_completed_event_handler.Do(func() {
		// 订阅外卖订单完成事件
		event.GetDispatcher().Register(&takeoutOrderCompletedEventSubscriber{})
	})
}

// takeoutOrderCompletedEventSubscriber 外卖订单完成事件订阅者
type takeoutOrderCompletedEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderCompletedEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.completed"}
}

// Handle 处理事件
func (s *takeoutOrderCompletedEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderCompletedEvent, ok := domainEvent.(event.OrderCompletedEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderCompletedEvent.CompanyUuid,
			orderCompletedEvent.TakeoutOrderUuid,
			orderCompletedEvent.Platform,
			orderCompletedEvent.ShortOrderNumber,
			"completed",
			map[string]any{},
		)
	})

	return nil
}
