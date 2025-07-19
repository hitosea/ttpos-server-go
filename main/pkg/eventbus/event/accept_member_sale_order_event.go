package event

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventAcceptMemberSaleOrder 外送订单“商家接单”事件
const EventAcceptMemberSaleOrder EventName = "Event_Accept_Member_Sale_Order"

// AcceptMemberSaleOrderPayload 外送订单“商家接单”事件的数据结构
type AcceptMemberSaleOrderPayload struct {
	BasePayload
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	MemberSaleOrder     *model.MemberSaleOrder
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
