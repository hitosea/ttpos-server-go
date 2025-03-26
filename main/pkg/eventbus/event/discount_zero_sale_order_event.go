package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventDiscountZeroSaleOrder "订单抹零"事件
const EventDiscountZeroSaleOrder EventName = "Event_Discount_Zero_Sale_Order"

// DiscountZeroSaleOrderPayload 每个事件有一个数据结构
type DiscountZeroSaleOrderPayload struct {
	BasePayload
	DiscountType    int     `json:"discount_type"`    // 折扣类型。1: 订单改价 2: 折扣 3:抹零
	RoundingType    int     `json:"rounding_type"`    // 抹零规则 1:抹分 2:抹角 3:四舍五入保留一位小数 4:四舍五入到整数
	SpecialDiscount float64 `json:"special_discount"` // 优惠金额。整单打折后的优惠金额=会员折扣后的订单应收金额-订单应收金额
}

func (payload *DiscountZeroSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// DiscountZeroSaleOrderHandler 每个事件的处理器
type DiscountZeroSaleOrderHandler func(msg DiscountZeroSaleOrderPayload)

// PublishDiscountZeroSaleOrderEvent 发布改价事件
func (system *SystemEventBus) PublishDiscountZeroSaleOrderEvent(msg DiscountZeroSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventDiscountZeroSaleOrder), Payload: msg})
}

// SubscribeDiscountZeroSaleOrderEvent 订阅改价事件
func (system *SystemEventBus) SubscribeDiscountZeroSaleOrderEvent(handler DiscountZeroSaleOrderHandler) {
	system.bus.Subscribe(string(EventDiscountZeroSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(DiscountZeroSaleOrderPayload)
		handler(msg)
	})
}
