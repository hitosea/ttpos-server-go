package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCancelSaleOrderDiscount “取消优惠折扣”事件名称
const EventCancelSaleOrderDiscount EventName = "Event_Cancel_Sale_Order_Discount"

// CancelSaleOrderDiscountPayload 每个事件有一个数据结构
type CancelSaleOrderDiscountPayload struct {
	BasePayload
	ParentId  uint64 `json:"parent_id"`  // 销售账单uuid
	OrderName uint64 `json:"order_name"` // 销售订单uuid
}

func (payload *CancelSaleOrderDiscountPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CancelSaleOrderDiscountHandler 每个事件的处理器
type CancelSaleOrderDiscountHandler func(msg CancelSaleOrderDiscountPayload)

// PublishCancelSaleOrderDiscountEvent 发布取消优惠折扣事件
func (system *SystemEventBus) PublishCancelSaleOrderDiscountEvent(msg CancelSaleOrderDiscountPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelSaleOrderDiscount), Payload: msg})
}

// SubscribeCancelSaleOrderDiscountEvent 订阅取消优惠折扣事件
func (system *SystemEventBus) SubscribeCancelSaleOrderDiscountEvent(handler CancelSaleOrderDiscountHandler) {
	system.bus.Subscribe(string(EventCancelSaleOrderDiscount), func(event eventbus.Event) {
		msg := event.Payload.(CancelSaleOrderDiscountPayload)
		handler(msg)
	})
}
