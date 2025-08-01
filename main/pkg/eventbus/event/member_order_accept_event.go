package event

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// AcceptMemberSaleOrderPayload 外送订单“商家接单”事件的数据结构
type AcceptMemberSaleOrderPayload struct {
	BasePayload
	MemberSaleOrder *model.MemberSaleOrder
}

func (payload *AcceptMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// AcceptMemberSaleOrderHandler 外送订单“商家接单”事件的处理器
type AcceptMemberSaleOrderHandler func(msg AcceptMemberSaleOrderPayload)

// PublishAcceptMemberSaleOrderEvent 发布外送订单“商家接单”事件
func (system *SystemEventBus) PublishAcceptMemberSaleOrderEvent(msg AcceptMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventAcceptMemberSaleOrder), Payload: msg})
}

// SubscribeAcceptMemberSaleOrderEvent 订阅外送订单“商家接单”事件
func (system *SystemEventBus) SubscribeAcceptMemberSaleOrderEvent(handler AcceptMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventAcceptMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(AcceptMemberSaleOrderPayload)
		handler(msg)
	})
}
