package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// PayFinishMemberDineInOrderPayload 会员端堂食订单支付完成事件的数据结构
type PayFinishMemberDineInOrderPayload struct {
	BasePayload
	PaymentOrderUuid uint64 `json:"payment_order_uuid"` // 支付单UUID
}

func (payload *PayFinishMemberDineInOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// PayFinishMemberDineInOrderHandler 会员端堂食订单支付完成事件的处理器
type PayFinishMemberDineInOrderHandler func(msg PayFinishMemberDineInOrderPayload)

// PublishPayFinishMemberDineInOrderEvent 发布会员端堂食订单支付完成事件
func (system *SystemEventBus) PublishPayFinishMemberDineInOrderEvent(msg PayFinishMemberDineInOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventPayFinishMemberDineInOrder), Payload: msg})
}

// SubscribePayFinishMemberDineInOrderEvent 订阅会员端堂食订单支付完成事件
func (system *SystemEventBus) SubscribePayFinishMemberDineInOrderEvent(handler PayFinishMemberDineInOrderHandler) {
	system.bus.Subscribe(string(EventPayFinishMemberDineInOrder), func(event eventbus.Event) {
		msg := event.Payload.(PayFinishMemberDineInOrderPayload)
		handler(msg)
	})
}
