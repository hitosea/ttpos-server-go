package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// RiderDeliveryMemberSaleOrderPayload “骑手完成取餐”事件的数据结构
type RiderDeliveryMemberSaleOrderPayload struct {
	BasePayload
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	RiderName           string `json:"rider_name" binding:"required"`             // 骑手姓名
	RiderPhone          string `json:"rider_phone" binding:"required"`            // 骑手手机号
}

func (payload *RiderDeliveryMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// RiderDeliveryMemberSaleOrderHandler “骑手完成取餐”事件的处理器
type RiderDeliveryMemberSaleOrderHandler func(msg RiderDeliveryMemberSaleOrderPayload)

// PublishRiderDeliveryMemberSaleOrderEvent 发布“骑手完成取餐”事件
func (system *SystemEventBus) PublishRiderDeliveryMemberSaleOrderEvent(msg RiderDeliveryMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventRiderDeliveryMemberSaleOrder), Payload: msg})
}

// SubscribeRiderDeliveryMemberSaleOrderEvent 订阅“骑手完成取餐”事件
func (system *SystemEventBus) SubscribeRiderDeliveryMemberSaleOrderEvent(handler RiderDeliveryMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventRiderDeliveryMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(RiderDeliveryMemberSaleOrderPayload)
		handler(msg)
	})
}
