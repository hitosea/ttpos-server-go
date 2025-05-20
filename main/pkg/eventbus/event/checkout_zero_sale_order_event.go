package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCheckoutZeroSaleOrder 结账手动抹零事件
const EventCheckoutZeroSaleOrder EventName = "Event_Checkout_Zero_Sale_Order"

// CheckoutZeroSaleOrderPayload 每个事件有一个数据结构
type CheckoutZeroSaleOrderPayload struct {
	BasePayload
	Operation       string  `json:"operation"`        // 操作类型。add: 设置结账抹零，cancel: 撤销结账抹零
	RoundingType    int     `json:"rounding_type"`    // 抹零规则。0-实款实收 1-抹分 2-抹角 5-抹元
	SpecialDiscount float64 `json:"special_discount"` // 优惠金额
	Reason          string  `json:"reason"`           // 原因(撤销时使用)
	IsAuto          bool    `json:"is_auto"`          // 是否自动抹零
}

func (payload *CheckoutZeroSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CheckoutZeroSaleOrderHandler 每个事件的处理器
type CheckoutZeroSaleOrderHandler func(msg CheckoutZeroSaleOrderPayload)

// PublishCheckoutZeroSaleOrderEvent 发布结账手动抹零事件
func (system *SystemEventBus) PublishCheckoutZeroSaleOrderEvent(msg CheckoutZeroSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCheckoutZeroSaleOrder), Payload: msg})
}

// SubscribeCheckoutZeroSaleOrderEvent 订阅结账手动抹零事件
func (system *SystemEventBus) SubscribeCheckoutZeroSaleOrderEvent(handler CheckoutZeroSaleOrderHandler) {
	system.bus.Subscribe(string(EventCheckoutZeroSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(CheckoutZeroSaleOrderPayload)
		handler(msg)
	})
}
