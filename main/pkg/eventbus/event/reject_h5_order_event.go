package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventRejectH5Order 拒单事件名称
const EventRejectH5Order EventName = "Event_Reject_H5_Order"

// RejectH5OrderPayload 拒单事件数据结构
type RejectH5OrderPayload struct {
	BasePayload
	H5OrderUuid uint64 `json:"h_5_order_uuid"` // h5订单Uuid
}

func (payload *RejectH5OrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// RejectH5OrderHandler 拒单事件处理器
type RejectH5OrderHandler func(msg RejectH5OrderPayload)

// PublishRejectH5OrderEvent 发布拒单事件
func (system *SystemEventBus) PublishRejectH5OrderEvent(msg RejectH5OrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventRejectH5Order), Payload: msg})
}

// SubscribeRejectH5OrderEvent 订阅拒单事件
func (system *SystemEventBus) SubscribeRejectH5OrderEvent(handler RejectH5OrderHandler) {
	system.bus.Subscribe(string(EventRejectH5Order), func(event eventbus.Event) {
		msg := event.Payload.(RejectH5OrderPayload)
		handler(msg)
	})
}
