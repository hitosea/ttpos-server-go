package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventReturnOrder 用餐订单退款事件名称
const EventReturnOrder EventName = "Event_Return_Order"

// ReturnOrderPayload 用餐订单退款事件数据结构
type ReturnOrderPayload struct {
	BasePayload
	PayTypes   []PayType `json:"pay_type"`
	ReturnType int       `json:"return_type"` // 退款方式：1-整单退款；2-部分退款
}

func (payload *ReturnOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ReturnOrderHandler 用餐订单退款事件处理器
type ReturnOrderHandler func(msg ReturnOrderPayload)

// PublishReturnOrderEvent 发布用餐订单退款事件
func (system *SystemEventBus) PublishReturnOrderEvent(msg ReturnOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventReturnOrder), Payload: msg})
}

// SubscribeReturnOrderEvent 订阅用餐订单退款事件
func (system *SystemEventBus) SubscribeReturnOrderEvent(handler ReturnOrderHandler) {
	system.bus.Subscribe(string(EventReturnOrder), func(event eventbus.Event) {
		msg := event.Payload.(ReturnOrderPayload)
		handler(msg)
	})
}
