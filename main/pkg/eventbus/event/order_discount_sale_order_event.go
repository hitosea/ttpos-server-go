package event

import (
	"ttpos-server-go/pkg/eventbus"
)

// EventDiscountSaleOrder 优惠折扣事件
const EventDiscountSaleOrder EventName = "Event_Discount_Sale_Order"

// DiscountSaleOrderHandler 每个事件的处理器
type DiscountSaleOrderHandler func(msg DiscountSaleOrderPayload)

// PublishDiscountSaleOrderEvent 发布改价事件
func (system *SystemEventBus) PublishDiscountSaleOrderEvent(msg DiscountSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventDiscountSaleOrder), Payload: msg})
}

// SubscribeDiscountSaleOrderEvent 订阅改价事件
func (system *SystemEventBus) SubscribeDiscountSaleOrderEvent(handler DiscountSaleOrderHandler) {
	system.bus.Subscribe(string(EventDiscountSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(DiscountSaleOrderPayload)
		handler(msg)
	})
}
