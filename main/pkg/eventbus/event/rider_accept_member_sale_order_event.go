package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventRiderAcceptMemberSaleOrder “骑手接单”事件
const EventRiderAcceptMemberSaleOrder EventName = "Event_Rider_Accept_Member_Sale_Order"

// RiderAcceptMemberSaleOrderPayload 接单骑手端销售订单事件的数据结构
type RiderAcceptMemberSaleOrderPayload struct {
	BasePayload
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
}

func (payload *RiderAcceptMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// RiderAcceptMemberSaleOrderHandler “骑手接单”事件的处理器
type RiderAcceptMemberSaleOrderHandler func(msg RiderAcceptMemberSaleOrderPayload)

// PublishRiderAcceptMemberSaleOrderEvent 发布“骑手接单”事件
func (system *SystemEventBus) PublishRiderAcceptMemberSaleOrderEvent(msg RiderAcceptMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventRiderAcceptMemberSaleOrder), Payload: msg})
}

// SubscribeRiderAcceptMemberSaleOrderEvent 订阅“骑手接单”事件
func (system *SystemEventBus) SubscribeRiderAcceptMemberSaleOrderEvent(handler RiderAcceptMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventRiderAcceptMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(RiderAcceptMemberSaleOrderPayload)
		handler(msg)
	})
}
