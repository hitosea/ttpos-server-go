package event

import (
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/pkg/utils"
)

var once_takeout_order_created_event_handler sync.Once

// TakeoutOrderCreatedEventHandler 外卖订单创建事件处理器
func TakeoutOrderCreatedEventHandler() {
	once_takeout_order_created_event_handler.Do(func() {
		// 订阅外卖订单创建事件
		event.GetDispatcher().Register(&takeoutOrderCreatedEventSubscriber{})
	})
}

// takeoutOrderCreatedEventSubscriber 外卖订单创建事件订阅者
type takeoutOrderCreatedEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderCreatedEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.created"}
}

// Handle 处理事件
func (s *takeoutOrderCreatedEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderCreatedEvent, ok := domainEvent.(event.OrderCreatedEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 发送 WebSocket 通知（订单数据更新）
		sendTakeoutOrderWebSocketNotification(
			orderCreatedEvent.CompanyUuid,
			orderCreatedEvent.TakeoutOrderUuid,
			orderCreatedEvent.Platform,
			orderCreatedEvent.ShortOrderNumber,
			"create",
			map[string]any{},
		)

		// 发送 CUSTOMER_CALL 通知（触发前端未处理提醒）
		sendCustomerCallWebSocketNotification(orderCreatedEvent.CompanyUuid)
	})

	return nil
}
