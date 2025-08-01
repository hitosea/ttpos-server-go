package event

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// RiderCompletedMemberSaleOrderPayload “骑手送达”事件的数据结构
type RiderCompletedMemberSaleOrderPayload struct {
	BasePayload
	SaleBill            *model.SaleBill `json:"-"`
	MemberSaleOrderUuid uint64          `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	RiderName           string          `json:"rider_name" binding:"required"`             // 骑手名称
	RiderPhone          string          `json:"rider_phone" binding:"required"`            // 骑手手机号
}

func (payload *RiderCompletedMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// RiderCompletedMemberSaleOrderHandler “骑手送达”事件的处理器
type RiderCompletedMemberSaleOrderHandler func(msg RiderCompletedMemberSaleOrderPayload)

// PublishRiderCompletedMemberSaleOrderEvent 发布“骑手送达”事件
func (system *SystemEventBus) PublishRiderCompletedMemberSaleOrderEvent(msg RiderCompletedMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventRiderCompletedMemberSaleOrder), Payload: msg})
}

// SubscribeRiderCompletedMemberSaleOrderEvent 订阅“骑手送达”事件
func (system *SystemEventBus) SubscribeRiderCompletedMemberSaleOrderEvent(handler RiderCompletedMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventRiderCompletedMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(RiderCompletedMemberSaleOrderPayload)
		handler(msg)
	})
}
