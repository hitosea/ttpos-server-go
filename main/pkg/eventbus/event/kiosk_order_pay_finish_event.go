package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// PayFinishKioskOrderPayload 自助点餐机订单支付完成事件的数据结构
type PayFinishKioskOrderPayload struct {
	BasePayload
	PaymentOrderUuid uint64 `json:"payment_order_uuid"` // 支付单UUID
}

func (payload *PayFinishKioskOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// PayFinishKioskOrderHandler 自助点餐机订单支付完成事件的处理器
type PayFinishKioskOrderHandler func(msg PayFinishKioskOrderPayload)

// PublishPayFinishKioskOrderEvent 发布自助点餐机订单支付完成事件
func (system *SystemEventBus) PublishPayFinishKioskOrderEvent(msg PayFinishKioskOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventPayFinishKioskOrder), Payload: msg})
}

// SubscribePayFinishKioskOrderEvent 订阅自助点餐机订单支付完成事件
func (system *SystemEventBus) SubscribePayFinishKioskOrderEvent(handler PayFinishKioskOrderHandler) {
	system.bus.Subscribe(string(EventPayFinishKioskOrder), func(event eventbus.Event) {
		msg := event.Payload.(PayFinishKioskOrderPayload)
		handler(msg)
	})
}
