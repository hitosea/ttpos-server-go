package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventOrderReverseSettle 用餐订单反结账事件名称
const EventOrderReverseSettle EventName = "Event_Order_Reverse_Settle"

// OrderReverseSettlePayload 用餐订单反结账事件数据结构
type OrderReverseSettlePayload struct {
	BasePayload
	PayTypes []PayType `json:"pay_type"`
}

func (payload *OrderReverseSettlePayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// OrderReverseSettleHandler 用餐订单反结账事件处理器
type OrderReverseSettleHandler func(msg OrderReverseSettlePayload)

// PublishOrderReverseSettleEvent 发布用餐订单反结账事件
func (system *SystemEventBus) PublishOrderReverseSettleEvent(msg OrderReverseSettlePayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventOrderReverseSettle), Payload: msg})
}

// SubscribeOrderReverseSettleEvent 订阅用餐订单反结账事件
func (system *SystemEventBus) SubscribeOrderReverseSettleEvent(handler OrderReverseSettleHandler) {
	system.bus.Subscribe(string(EventOrderReverseSettle), func(event eventbus.Event) {
		msg := event.Payload.(OrderReverseSettlePayload)
		handler(msg)
	})
}
