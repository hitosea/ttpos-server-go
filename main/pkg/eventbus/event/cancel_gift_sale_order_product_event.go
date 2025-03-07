package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCancelGiftSaleOrderProduct 事件名称，每个事件都有一个全局唯一的名称
const EventCancelGiftSaleOrderProduct EventName = "Event_Cancel_Gift_Sale_Order_Product"

// CancelGiftSaleOrderProductPayload 每个事件有一个数据结构
type CancelGiftSaleOrderProductPayload struct {
	BasePayload
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	ProductPrice   float64            `json:"product_price"`    // 商品价格
	TotalNum       uint               `json:"total_num"`        // 总数量
	TotalPrice     float64            `json:"total_price"`      // 总价格
	ParentId       uint64             `json:"parent_id"`        // 父订单ID
	OrderName      string             `json:"order_name"`       // 订单名称
}

func (payload *CancelGiftSaleOrderProductPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CancelGiftSaleOrderProductHandler 每个事件的处理器
type CancelGiftSaleOrderProductHandler func(msg CancelGiftSaleOrderProductPayload)

// PublishCancelGiftSaleOrderProductEvent 发布取消赠菜事件
func (system *SystemEventBus) PublishCancelGiftSaleOrderProductEvent(msg CancelGiftSaleOrderProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelGiftSaleOrderProduct), Payload: msg})
}

// SubscribeCancelGiftSaleOrderProductEvent 订阅取消赠菜事件
func (system *SystemEventBus) SubscribeCancelGiftSaleOrderProductEvent(handler CancelGiftSaleOrderProductHandler) {
	system.bus.Subscribe(string(EventCancelGiftSaleOrderProduct), func(event eventbus.Event) {
		msg := event.Payload.(CancelGiftSaleOrderProductPayload)
		handler(msg)
	})
}
