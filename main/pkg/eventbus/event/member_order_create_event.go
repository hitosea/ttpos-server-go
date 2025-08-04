package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// CreateMemberSaleOrderPayload "创建外送订单"事件数据结构
type CreateMemberSaleOrderPayload struct {
	BasePayload
}

func (payload *CreateMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CreateMemberSaleOrderHandler “创建外送订单”事件处理器
type CreateMemberSaleOrderHandler func(msg CreateMemberSaleOrderPayload)

// PublishCreateMemberSaleOrderEvent 发布“创建外送订单”事件
func (system *SystemEventBus) PublishCreateMemberSaleOrderEvent(msg CreateMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCreateMemberSaleOrder), Payload: msg})
}

// SubscribeCreateMemberSaleOrderEvent 订阅“创建外送订单”事件
func (system *SystemEventBus) SubscribeCreateMemberSaleOrderEvent(handler CreateMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventCreateMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(CreateMemberSaleOrderPayload)
		handler(msg)
	})
}
