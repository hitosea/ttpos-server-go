package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventAcceptH5Order 接单事件名称
const EventAcceptH5Order EventName = "Event_Accept_H5_Order"

// AcceptH5OrderPayload 接单事件数据结构
type AcceptH5OrderPayload struct {
	BasePayload
	IsAutoOrder bool   `json:"is_auto_order"` // 是否自动接单
	H5OrderUuid uint64 `json:"h5_order_uuid"` // h5订单Uuid
}

func (payload *AcceptH5OrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// AcceptH5OrderHandler 接单事件处理器
type AcceptH5OrderHandler func(msg AcceptH5OrderPayload)

// PublishAcceptH5OrderEvent 发布接单事件
func (system *SystemEventBus) PublishAcceptH5OrderEvent(msg AcceptH5OrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventAcceptH5Order), Payload: msg})
}

// SubscribeAcceptH5OrderEvent 订阅接单事件
func (system *SystemEventBus) SubscribeAcceptH5OrderEvent(handler AcceptH5OrderHandler) {
	system.bus.Subscribe(string(EventAcceptH5Order), func(event eventbus.Event) {
		msg := event.Payload.(AcceptH5OrderPayload)
		handler(msg)
	})
}
