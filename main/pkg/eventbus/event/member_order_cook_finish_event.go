package event

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// CookFinishMemberSaleOrderPayload 外送订单“厨师完成”事件的数据结构
type CookFinishMemberSaleOrderPayload struct {
	BasePayload
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid" binding:"required"` // 会员端销售订单UUID
	MemberSaleOrder     *model.MemberSaleOrder
}

func (payload *CookFinishMemberSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CookFinishMemberSaleOrderHandler 外送订单“厨师完成”事件的处理器
type CookFinishMemberSaleOrderHandler func(msg CookFinishMemberSaleOrderPayload)

// PublishCookFinishMemberSaleOrderEvent 发布外送订单“厨师完成”事件
func (system *SystemEventBus) PublishCookFinishMemberSaleOrderEvent(msg CookFinishMemberSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCookFinishMemberSaleOrder), Payload: msg})
}

// SubscribeCookFinishMemberSaleOrderEvent 订阅外送订单“厨师完成”事件
func (system *SystemEventBus) SubscribeCookFinishMemberSaleOrderEvent(handler CookFinishMemberSaleOrderHandler) {
	system.bus.Subscribe(string(EventCookFinishMemberSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(CookFinishMemberSaleOrderPayload)
		handler(msg)
	})
}
