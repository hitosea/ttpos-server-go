package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCancelOrder 整单取消事件名称
const EventCancelOrder EventName = "Event_Cancel_Order"

// CancelOrderPayload 整单取消事件数据结构
type CancelOrderPayload struct {
	BasePayload
	Data CancelMemberOrderPayloadData `json:"data"`
}

func (payload *CancelOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CancelOrderHandler 开台事件处理器
type CancelOrderHandler func(msg CancelOrderPayload)

// PublishCancelOrderEvent 发布整单取消事件
func (system *SystemEventBus) PublishCancelOrderEvent(msg CancelOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelOrder), Payload: msg})
}

// SubscribeCancelOrderEvent 订阅整单取消事件
func (system *SystemEventBus) SubscribeCancelOrderEvent(handler CancelOrderHandler) {
	system.bus.Subscribe(string(EventCancelOrder), func(event eventbus.Event) {
		msg := event.Payload.(CancelOrderPayload)
		handler(msg)
	})
}
