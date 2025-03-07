package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventChangePriceSaleOrder 事件名称，每个事件都有一个全局唯一的名称
const EventChangePriceSaleOrder EventName = "Event_Change_Price_Sale_Order"

// ChangePriceSaleOrderPayload 每个事件有一个数据结构
type ChangePriceSaleOrderPayload struct {
	BasePayload
	OldPrice        float64 `json:"old_price"`        // 原价。订单改价前的价格
	NewPrice        float64 `json:"new_price"`        // 新价格。订单改价后的价格
	DiscountType    int     `json:"discount_type"`    // 折扣类型。1: 订单改价
	SpecialDiscount float64 `json:"special_discount"` // 优惠金额。订单改价后的优惠金额
}

func (payload *ChangePriceSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ChangePriceSaleOrderHandler 每个事件的处理器
type ChangePriceSaleOrderHandler func(msg ChangePriceSaleOrderPayload)

// PublishChangePriceSaleOrderProductEvent 发布改价事件
func (system *SystemEventBus) PublishChangePriceSaleOrderEvent(msg ChangePriceSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangePriceSaleOrder), Payload: msg})
}

// SubscribeChangePriceSaleOrderEvent 订阅改价事件
func (system *SystemEventBus) SubscribeChangePriceSaleOrderEvent(handler ChangePriceSaleOrderHandler) {
	system.bus.Subscribe(string(EventChangePriceSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(ChangePriceSaleOrderPayload)
		handler(msg)
	})
}
