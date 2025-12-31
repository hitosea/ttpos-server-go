package event

import (
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/pkg/utils"
)

var once_takeout_order_rider_processing_event_handler sync.Once

// TakeoutOrderRiderProcessingEventHandler 外卖订单骑手配送中事件处理器
func TakeoutOrderRiderProcessingEventHandler() {
	once_takeout_order_rider_processing_event_handler.Do(func() {
		// 订阅外卖订单骑手配送中事件
		event.GetDispatcher().Register(&takeoutOrderRiderProcessingEventSubscriber{})
	})
}

// takeoutOrderRiderProcessingEventSubscriber 外卖订单骑手配送中事件订阅者
type takeoutOrderRiderProcessingEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderRiderProcessingEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.rider.processing"}
}

// Handle 处理事件
func (s *takeoutOrderRiderProcessingEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderRiderProcessingEvent, ok := domainEvent.(event.OrderRiderProcessingEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderRiderProcessingEvent.CompanyUuid,
			orderRiderProcessingEvent.TakeoutOrderUuid,
			orderRiderProcessingEvent.Platform,
			orderRiderProcessingEvent.ShortOrderNumber,
			"rider.processing",
			map[string]any{},
		)
	})

	return nil
}
