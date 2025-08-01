package event

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventRejectMemberSaleOrder 拒单会员端销售订单事件
const EventRejectMemberSaleOrder EventName = "Event_Reject_Member_Sale_Order"

// RejectMemberSaleOrderPayload 拒单会员端销售订单事件的数据结构
type RejectMemberSaleOrderPayload struct {
	BasePayload
	Data                CancelMemberOrderPayloadData `json:"data"`
	MemberSaleOrderUuid uint64                       `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	MemberSaleOrder     *model.MemberSaleOrder
}

func (payload *RejectMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// RejectMemberSaleOrderHandler 拒单会员端销售订单事件的处理器
type RejectMemberSaleOrderHandler func(msg RejectMemberSaleOrderPayload)

// PublishRejectMemberSaleOrderEvent 发布拒单会员端销售订单事件
func (system *SystemEventBus) PublishRejectMemberSaleOrderEvent(msg RejectMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventRejectMemberSaleOrder), Payload: msg})
}

// SubscribeRejectMemberSaleOrderEvent 订阅拒单会员端销售订单事件
func (system *SystemEventBus) SubscribeRejectMemberSaleOrderEvent(handler RejectMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventRejectMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(RejectMemberSaleOrderPayload)
		handler(msg)
	})
}
