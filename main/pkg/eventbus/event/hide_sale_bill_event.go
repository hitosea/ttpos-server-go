package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventHideSaleBill “挂单”事件名称
const EventHideSaleBill EventName = "Event_Hide_Sale_Bill"

// HideSaleBillPayload “挂单”事件数据结构
type HideSaleBillPayload struct {
	BasePayload
}

func (payload *HideSaleBillPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// HideSaleBillHandler 挂单事件处理器
type HideSaleBillHandler func(msg HideSaleBillPayload)

// PublishHideSaleBillEvent 发布挂单事件
func (system *SystemEventBus) PublishHideSaleBillEvent(msg HideSaleBillPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventHideSaleBill), Payload: msg})
}

// SubscribeHideSaleBillEvent 订阅挂单事件
func (system *SystemEventBus) SubscribeHideSaleBillEvent(handler HideSaleBillHandler) {
	system.bus.Subscribe(string(EventHideSaleBill), func(event eventbus.Event) {
		msg := event.Payload.(HideSaleBillPayload)
		handler(msg)
	})
}
