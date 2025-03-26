package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventDiscountSaleOrder 优惠折扣事件
const EventDiscountSaleOrder EventName = "Event_Discount_Sale_Order"

// DiscountSaleOrderPayload 每个事件有一个数据结构
type DiscountSaleOrderPayload struct {
	BasePayload
	OldPrice        float64 `json:"old_price"`        // 进行整单打折前的总金额
	NewPrice        float64 `json:"new_price"`        // 整单打折后的总金额
	DiscountType    int     `json:"discount_type"`    // 折扣类型。1: 订单改价 2: 折扣 3:抹零
	SpecialDiscount float64 `json:"special_discount"` // 优惠金额。整单打折后的优惠金额=会员折扣后的订单应收金额-订单应收金额
	RoundingRate    float64 `json:"rounding_rate"`    // 打折率。如八折，则打折率是20； 如30%off，则打折率是30。统一展示格式为“优惠折扣：折扣-80%（￥50）”，无论是百分比打折还是百分比减免，都统一展示为百分比减免。
}

func (payload *DiscountSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// DiscountSaleOrderHandler 每个事件的处理器
type DiscountSaleOrderHandler func(msg DiscountSaleOrderPayload)

// PublishDiscountSaleOrderEvent 发布改价事件
func (system *SystemEventBus) PublishDiscountSaleOrderEvent(msg DiscountSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventDiscountSaleOrder), Payload: msg})
}

// SubscribeDiscountSaleOrderEvent 订阅改价事件
func (system *SystemEventBus) SubscribeDiscountSaleOrderEvent(handler DiscountSaleOrderHandler) {
	system.bus.Subscribe(string(EventDiscountSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(DiscountSaleOrderPayload)
		handler(msg)
	})
}
