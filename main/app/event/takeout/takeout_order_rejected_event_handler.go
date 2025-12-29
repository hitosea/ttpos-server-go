package event

import (
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/pkg/utils"
)

var once_takeout_order_rejected_event_handler sync.Once

// TakeoutOrderRejectedEventHandler 外卖订单拒单事件处理器
func TakeoutOrderRejectedEventHandler() {
	once_takeout_order_rejected_event_handler.Do(func() {
		// 订阅外卖订单拒单事件
		event.GetDispatcher().Register(&takeoutOrderRejectedEventSubscriber{})
	})
}

// takeoutOrderRejectedEventSubscriber 外卖订单拒单事件订阅者
type takeoutOrderRejectedEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderRejectedEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.rejected"}
}

// Handle 处理事件
func (s *takeoutOrderRejectedEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderRejectedEvent, ok := domainEvent.(event.OrderRejectedEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderRejectedEvent.CompanyUuid,
			orderRejectedEvent.TakeoutOrderUuid,
			orderRejectedEvent.Platform,
			orderRejectedEvent.ShortOrderNumber,
			"rejected",
			map[string]any{},
		)
		// 发送 CUSTOMER_CALL 通知（触发前端未处理提醒）
		sendCustomerCallWebSocketNotification(orderRejectedEvent.CompanyUuid)
	})

	return nil
}
