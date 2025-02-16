package event

import "ttpos-server-go/pkg/eventbus"

// EventCancelDoingProduct 事件名称，每个事件都有一个全局唯一的名称
const EventCancelDoingProduct EventName = "Event_Cancel_Doing_Product"

// CancelDoingProductPayload 每个事件有一个数据结构
type CancelDoingProductPayload struct {
	SaleOrderProductUuids []uint64 `json:"sale_order_product_uuids"`
	companyUuid           uint64
}

// CancelDoingProductHandler 每个事件的处理器
type CancelDoingProductHandler func(msg CancelDoingProductPayload)

// PublishCancelDoingProductEvent 发布取消进行中的产品事件
func (system *SystemEventBus) PublishCancelDoingProductEvent(msg CancelDoingProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelDoingProduct), Payload: msg})
}

// SubscribeCancelDoingProductEvent 订阅取消进行中的产品事件
func (system *SystemEventBus) SubscribeCancelDoingProductEvent(handler CancelDoingProductHandler) {
	system.bus.Subscribe(string(EventCancelDoingProduct), func(event eventbus.Event) {
		msg := event.Payload.(CancelDoingProductPayload)
		handler(msg)
	})
}
