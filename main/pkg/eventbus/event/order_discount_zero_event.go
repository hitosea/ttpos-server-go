package event

import (
	"ttpos-server-go/pkg/eventbus"
)

// DiscountZeroSaleOrderHandler 每个事件的处理器
type DiscountZeroSaleOrderHandler func(msg DiscountSaleOrderPayload)

// PublishDiscountZeroSaleOrderEvent 发布抹零事件
func (system *SystemEventBus) PublishDiscountZeroSaleOrderEvent(msg DiscountSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventDiscountZeroSaleOrder), Payload: msg})
}

// SubscribeDiscountZeroSaleOrderEvent 订阅改价事件
func (system *SystemEventBus) SubscribeDiscountZeroSaleOrderEvent(handler DiscountZeroSaleOrderHandler) {
	system.bus.Subscribe(string(EventDiscountZeroSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(DiscountSaleOrderPayload)
		handler(msg)
	})
}
