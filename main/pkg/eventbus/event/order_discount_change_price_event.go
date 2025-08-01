package event

import (
	"ttpos-server-go/pkg/eventbus"
)

// DiscountChangePriceSaleOrderHandler 每个事件的处理器
type DiscountChangePriceSaleOrderHandler func(msg DiscountSaleOrderPayload)

// PublishDiscountChangePriceSaleOrderEvent 发布改价事件
func (system *SystemEventBus) PublishDiscountChangePriceSaleOrderEvent(msg DiscountSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventDiscountChangePriceSaleOrder), Payload: msg})
}

// SubscribeDiscountChangePriceSaleOrderEvent 订阅改价事件
func (system *SystemEventBus) SubscribeDiscountChangePriceSaleOrderEvent(handler DiscountChangePriceSaleOrderHandler) {
	system.bus.Subscribe(string(EventDiscountChangePriceSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(DiscountSaleOrderPayload)
		handler(msg)
	})
}
