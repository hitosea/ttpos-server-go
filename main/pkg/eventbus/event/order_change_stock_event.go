package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventChangeStock 加库存事件名称
const EventChangeStock EventName = "Event_Change_Stock"

// ChangeStockPayload 加库存事件数据结构
type ChangeStockPayload struct {
	BasePayload
}

func (payload *ChangeStockPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ChangeStockHandler 加库存事件处理器
type ChangeStockHandler func(msg ChangeStockPayload)

// PublishChangeStockEvent 发布加库存事件
func (system *SystemEventBus) PublishChangeStockEvent(msg ChangeStockPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangeStock), Payload: msg})
}

// SubscribeChangeStockEvent 订阅加库存事件
func (system *SystemEventBus) SubscribeChangeStockEvent(handler ChangeStockHandler) {
	system.bus.Subscribe(string(EventChangeStock), func(event eventbus.Event) {
		msg := event.Payload.(ChangeStockPayload)
		handler(msg)
	})
}
