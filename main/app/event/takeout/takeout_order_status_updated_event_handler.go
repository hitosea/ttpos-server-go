package event

import (
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/pkg/utils"
)

var once_takeout_order_status_updated_event_handler sync.Once

// TakeoutOrderStatusUpdatedEventHandler 外卖订单状态更新事件处理器
func TakeoutOrderStatusUpdatedEventHandler() {
	once_takeout_order_status_updated_event_handler.Do(func() {
		// 订阅外卖订单状态更新事件
		event.GetDispatcher().Register(&takeoutOrderStatusUpdatedEventSubscriber{})
	})
}

// takeoutOrderStatusUpdatedEventSubscriber 外卖订单状态更新事件订阅者
type takeoutOrderStatusUpdatedEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderStatusUpdatedEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.status_updated"}
}

// Handle 处理事件
func (s *takeoutOrderStatusUpdatedEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderStatusUpdatedEvent, ok := domainEvent.(event.OrderStatusUpdatedEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderStatusUpdatedEvent.CompanyUuid,
			orderStatusUpdatedEvent.TakeoutOrderUuid,
			orderStatusUpdatedEvent.Platform,
			orderStatusUpdatedEvent.ShortOrderNumber,
			"status_updated",
			map[string]any{
				"old_order_state": orderStatusUpdatedEvent.OldOrderState,
				"new_order_state": orderStatusUpdatedEvent.NewOrderState,
				"platform_state":  orderStatusUpdatedEvent.PlatformState,
			},
		)
	})

	return nil
}
