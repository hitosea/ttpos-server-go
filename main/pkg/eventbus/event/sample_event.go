package event

import "ttpos-server-go/pkg/eventbus"

// EventSample 事件名称，每个事件都有一个全局唯一的名称
const EventSample EventName = "Event_Sample"

// SamplePayload 每个事件有一个数据结构
type SamplePayload struct {
	Region string `json:"region"`
}

// SampleHandler 每个事件都有一个事件处理函数
type SampleHandler func(msg SamplePayload)

func (instance *SystemEventBus) PublishSampleEvent(msg SamplePayload) {
	instance.bus.Publish(eventbus.Event{Name: string(EventSample), Payload: msg})
}

func (instance *SystemEventBus) SubscribeSampleEvent(handler SampleHandler) {
	instance.bus.Subscribe(string(EventSample), func(event eventbus.Event) {
		msg := event.Payload.(SamplePayload)
		handler(msg)
	})
}
