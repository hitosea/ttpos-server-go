package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventPayFinishMemberSaleOrder 支付完成会员端销售订单事件
const EventPayFinishMemberSaleOrder EventName = "Event_Pay_Finish_Member_Sale_Order"

// PayFinishMemberSaleOrderPayload 支付完成会员端销售订单事件的数据结构
type PayFinishMemberSaleOrderPayload struct {
	BasePayload
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
}

func (payload *PayFinishMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// PayFinishMemberSaleOrderHandler 支付完成会员端销售订单事件的处理器
type PayFinishMemberSaleOrderHandler func(msg PayFinishMemberSaleOrderPayload)

// PublishPayFinishMemberSaleOrderEvent 发布支付完成会员端销售订单事件
func (system *SystemEventBus) PublishPayFinishMemberSaleOrderEvent(msg PayFinishMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventPayFinishMemberSaleOrder), Payload: msg})
}

// SubscribePayFinishMemberSaleOrderEvent 订阅支付完成会员端销售订单事件
func (system *SystemEventBus) SubscribePayFinishMemberSaleOrderEvent(handler PayFinishMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventPayFinishMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(PayFinishMemberSaleOrderPayload)
		handler(msg)
	})
}
