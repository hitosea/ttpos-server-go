package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// WrapSaleBillPayload 每个事件有一个数据结构
type WrapSaleBillPayload struct {
	BasePayload
}

func (payload *WrapSaleBillPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// WrapSaleBillHandler 每个事件的处理器
type WrapSaleBillHandler func(msg WrapSaleBillPayload)

// PublishWrapSaleBillEvent 发布打包事件
func (system *SystemEventBus) PublishWrapSaleBillEvent(msg WrapSaleBillPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventWrapSaleBill), Payload: msg})
}

// SubscribeWrapSaleBillEvent 订阅打包事件
func (system *SystemEventBus) SubscribeWrapSaleBillEvent(handler WrapSaleBillHandler) {
	system.bus.Subscribe(string(EventWrapSaleBill), func(event eventbus.Event) {
		msg := event.Payload.(WrapSaleBillPayload)
		handler(msg)
	})
}
