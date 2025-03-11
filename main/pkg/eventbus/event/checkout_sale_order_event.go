package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCheckoutSaleOrder 事件名称，每个事件都有一个全局唯一的名称
const EventCheckoutSaleOrder EventName = "Event_Checkout_Sale_Order"

// CheckoutSaleOrderPayload 每个事件有一个数据结构
type CheckoutSaleOrderPayload struct {
	BasePayload
	OrderPrice  float64   `json:"order_price"`  // 订单应付金额
	PayPrice    float64   `json:"pay_price"`    // 最终应付金额。最终应付金额=订单应付金额+手续费（手续费=每笔付款单的手续费之和）
	ActualPrice float64   `json:"actual_price"` // 最终实付金额。最终实付金额=最终应付金额+找零金额。如果没有找零，则最终实付金额=最终应付金额。最终实付金额=每笔付款单的付款金额之和（含手续费）
	ChangeDue   float64   `json:"change_due"`   // 找零金额
	PayType     []PayType `json:"pay_type"`     // 支付方式
}

type PayType struct {
	Name           string  `json:"name"`            // 支付方式名称
	Value          int     `json:"value"`           // 支付方式值
	DisabledCancel uint    `json:"disabled_cancel"` // 是否禁用取消
	Price          float64 `json:"price"`           // 支付金额（含手续费）
	FeeMoney       float64 `json:"fee_money"`       // 手续费
}

func (payload *CheckoutSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CheckoutSaleOrderHandler 每个事件的处理器
type CheckoutSaleOrderHandler func(msg CheckoutSaleOrderPayload)

// PublishCheckoutSaleOrderEvent 发布结账事件
func (system *SystemEventBus) PublishCheckoutSaleOrderEvent(msg CheckoutSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCheckoutSaleOrder), Payload: msg})
}

// SubscribeCheckoutSaleOrderEvent 订阅结账事件
func (system *SystemEventBus) SubscribeCheckoutSaleOrderEvent(handler CheckoutSaleOrderHandler) {
	system.bus.Subscribe(string(EventCheckoutSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(CheckoutSaleOrderPayload)
		handler(msg)
	})
}
