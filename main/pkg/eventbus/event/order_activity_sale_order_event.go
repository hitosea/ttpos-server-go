package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// ActivitySaleOrderPayload 满减活动事件荷载
type ActivitySaleOrderPayload struct {
	BasePayload
	FullReductionActivityUuid    uint64  `json:"full_reduction_activity_uuid"`    // 满减活动UUID
	FullReductionActivityMessage string  `json:"full_reduction_activity_message"` // 满减规则信息（如"满200减20"）
	ActivityAmount               float64 `json:"activity_amount"`                 // 满减活动抵扣金额
	OldPrice                     float64 `json:"old_price"`                       // 使用活动前的订单金额
	NewPrice                     float64 `json:"new_price"`                       // 使用活动后的订单金额
}

func (payload *ActivitySaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ActivitySaleOrderHandler 满减活动事件处理器
type ActivitySaleOrderHandler func(msg ActivitySaleOrderPayload)

// PublishActivitySaleOrderEvent 发布满减活动事件
func (system *SystemEventBus) PublishActivitySaleOrderEvent(msg ActivitySaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventActivitySaleOrder), Payload: msg})
}

// SubscribeActivitySaleOrderEvent 订阅满减活动事件
func (system *SystemEventBus) SubscribeActivitySaleOrderEvent(handler ActivitySaleOrderHandler) {
	system.bus.Subscribe(string(EventActivitySaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(ActivitySaleOrderPayload)
		handler(msg)
	})
}
