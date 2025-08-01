package event

import (
	"ttpos-server-go/pkg/eventbus"
)

// ShowSaleBillPayload “取单”事件数据结构
type ShowSaleBillPayload struct {
	BasePayload
}

// ShowSaleBillHandler 取单事件处理器
type ShowSaleBillHandler func(msg ShowSaleBillPayload)

// PublishShowSaleBillEvent 发布取单事件
func (system *SystemEventBus) PublishShowSaleBillEvent(msg ShowSaleBillPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventShowSaleBill), Payload: msg})
}

// SubscribeShowSaleBillEvent 订阅取单事件
func (system *SystemEventBus) SubscribeShowSaleBillEvent(handler ShowSaleBillHandler) {
	system.bus.Subscribe(string(EventShowSaleBill), func(event eventbus.Event) {
		msg := event.Payload.(ShowSaleBillPayload)
		handler(msg)
	})
}
