package event

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCheckoutZeroCancelSaleOrder 事件名称，每个事件都有一个全局唯一的名称
const EventCheckoutZeroCancelSaleOrder EventName = "Event_Checkout_Zero_Cancel_Sale_Order"

// CheckoutZeroCancelSaleOrderPayload 每个事件有一个数据结构
type CheckoutZeroCancelSaleOrderPayload struct {
	BasePayload
	Operation string `json:"operation"` // 操作类型。add: 设置结账抹零，cancel: 撤销结账抹零
	Remark    string `json:"remark"`    // 备注
}

const (
	CheckoutZeroCancelSaleOrderRemark = "选择含手续费的支付方式"
)

func (payload *CheckoutZeroCancelSaleOrderPayload) ToJsonString() string {
	payload.Operation = constant.OrderCheckoutDiscountCancel
	payload.Remark = CheckoutZeroCancelSaleOrderRemark
	return utils.ToJson(payload)
}

// CheckoutZeroCancelSaleOrderHandler 每个事件的处理器
type CheckoutZeroCancelSaleOrderHandler func(msg CheckoutZeroCancelSaleOrderPayload)

// PublishCheckoutZeroCancelSaleOrderEvent 发布结账撤销抹零事件
func (system *SystemEventBus) PublishCheckoutZeroCancelSaleOrderEvent(msg CheckoutZeroCancelSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCheckoutZeroCancelSaleOrder), Payload: msg})
}

// SubscribeCheckoutZeroCancelSaleOrderEvent 订阅结账撤销抹零事件
func (system *SystemEventBus) SubscribeCheckoutZeroCancelSaleOrderEvent(handler CheckoutZeroCancelSaleOrderHandler) {
	system.bus.Subscribe(string(EventCheckoutZeroCancelSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(CheckoutZeroCancelSaleOrderPayload)
		handler(msg)
	})
}
