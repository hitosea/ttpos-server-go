package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// CancelSplitOrderPayload 拆单事件数据结构
type CancelSplitOrderPayload struct {
	BasePayload
}

func (payload *CancelSplitOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CancelSplitOrderHandler 拆单事件处理器
type CancelSplitOrderHandler func(msg CancelSplitOrderPayload)

// PublishCancelSplitOrderEvent 发布拆单事件
func (system *SystemEventBus) PublishCancelSplitOrderEvent(msg CancelSplitOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelSplitOrder), Payload: msg})
}

// SubscribeCancelSplitOrderEvent 订阅拆单事件
func (system *SystemEventBus) SubscribeCancelSplitOrderEvent(handler CancelSplitOrderHandler) {
	system.bus.Subscribe(string(EventCancelSplitOrder), func(event eventbus.Event) {
		msg := event.Payload.(CancelSplitOrderPayload)
		handler(msg)
	})
}
