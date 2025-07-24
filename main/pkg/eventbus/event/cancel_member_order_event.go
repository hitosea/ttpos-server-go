package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCancelMemberOrder 整单取消事件名称
const EventCancelMemberOrder EventName = "Event_Cancel_Member_Order"

type CancelMemberOrderPayloadDataRefund struct {
	Name              string  `json:"name"`
	Code              int     `json:"code"`
	Amount            float64 `json:"amount"`
	RefundStatus      int     `json:"refund_status"`
	ReturnAmountUuid  uint64  `json:"return_amount_uuid"`
	ReturnOrderUuid   uint64  `json:"return_order_uuid"`
	PaymentOrderUuid  uint64  `json:"payment_order_uuid"`
	PaymentMethodUuid uint64  `json:"payment_method_uuid"`
}

type CancelMemberOrderPayloadData struct {
	Type    string                               `json:"type"`
	Refunds []CancelMemberOrderPayloadDataRefund `json:"refunds"`
}

// CancelMemberOrderPayload 整单取消事件数据结构
type CancelMemberOrderPayload struct {
	BasePayload
	Data CancelMemberOrderPayloadData `json:"data"`
}

func (payload *CancelMemberOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CancelOrderHandler 开台事件处理器
type CancelMemberOrderHandler func(msg CancelMemberOrderPayload)

// PublishCancelOrderEvent 发布整单取消事件
func (system *SystemEventBus) PublishCancelMemberOrderEvent(msg CancelMemberOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelMemberOrder), Payload: msg})
}

// SubscribeCancelOrderEvent 订阅整单取消事件
func (system *SystemEventBus) SubscribeCancelMemberOrderEvent(handler CancelMemberOrderHandler) {
	system.bus.Subscribe(string(EventCancelMemberOrder), func(event eventbus.Event) {
		msg := event.Payload.(CancelMemberOrderPayload)
		handler(msg)
	})
}
