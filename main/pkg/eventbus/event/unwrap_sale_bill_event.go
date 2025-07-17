package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventUnwrapSaleBill 取消整单打包事件
const EventUnwrapSaleBill EventName = "Event_Unwrap_Sale_Bill"

// UnwrapSaleBillPayload 每个事件有一个数据结构
type UnwrapSaleBillPayload struct {
	BasePayload
}

func (payload *UnwrapSaleBillPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// UnwrapSaleBillHandler 每个事件的处理器
type UnwrapSaleBillHandler func(msg UnwrapSaleBillPayload)

// PublishUnwrapSaleBillEvent 发布取消整单打包事件
func (system *SystemEventBus) PublishUnwrapSaleBillEvent(msg UnwrapSaleBillPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventUnwrapSaleBill), Payload: msg})
}

// SubscribeUnwrapSaleBillEvent 订阅取消整单打包事件
func (system *SystemEventBus) SubscribeUnwrapSaleBillEvent(handler UnwrapSaleBillHandler) {
	system.bus.Subscribe(string(EventUnwrapSaleBill), func(event eventbus.Event) {
		msg := event.Payload.(UnwrapSaleBillPayload)
		handler(msg)
	})
}
