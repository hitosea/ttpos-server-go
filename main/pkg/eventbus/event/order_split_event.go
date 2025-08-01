package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventSplitOrder 拆单事件名称
const EventSplitOrder EventName = "Event_Split_Order"

type Order struct {
	SaleOrderUuid uint64  `json:"sale_order_uuid"` // 销售订单Uuid
	OrderName     string  `json:"order_name"`      // 订单名称，顺序
	Amount        float64 `json:"amount"`          // 订单金额
}

// SplitOrderPayload 拆单事件数据结构
type SplitOrderPayload struct {
	BasePayload
	Orders []Order `json:"orders"`
}

func (payload *SplitOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// SplitOrderHandler 拆单事件处理器
type SplitOrderHandler func(msg SplitOrderPayload)

// PublishSplitOrderEvent 发布拆单事件
func (system *SystemEventBus) PublishSplitOrderEvent(msg SplitOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventSplitOrder), Payload: msg})
}

// SubscribeSplitOrderEvent 订阅拆单事件
func (system *SystemEventBus) SubscribeSplitOrderEvent(handler SplitOrderHandler) {
	system.bus.Subscribe(string(EventSplitOrder), func(event eventbus.Event) {
		msg := event.Payload.(SplitOrderPayload)
		handler(msg)
	})
}
