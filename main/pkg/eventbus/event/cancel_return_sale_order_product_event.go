package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCancelReturnSaleOrderProduct 事件名称，每个事件都有一个全局唯一的名称
const EventCancelReturnSaleOrderProduct EventName = "Event_Cancel_Return_Sale_Order_Product"

// CancelReturnSaleOrderProductPayload 每个事件有一个数据结构
type CancelReturnSaleOrderProductPayload struct {
	BasePayload
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	Num            uint               `json:"num"`              // 退菜数量
	ParentId       uint64             `json:"parent_id"`        // 父订单ID
	OrderName      uint64             `json:"order_name"`       // 订单名称
	Sign           string             `json:"sign"`             // 签名
}

func (payload *CancelReturnSaleOrderProductPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CancelReturnSaleOrderProductHandler 每个事件的处理器
type CancelReturnSaleOrderProductHandler func(msg CancelReturnSaleOrderProductPayload)

// PublishCancelReturnSaleOrderProductEvent 发布取消退菜事件
func (system *SystemEventBus) PublishCancelReturnSaleOrderProductEvent(msg CancelReturnSaleOrderProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelReturnSaleOrderProduct), Payload: msg})
}

// SubscribeCancelReturnSaleOrderProductEvent 订阅取消退菜事件
func (system *SystemEventBus) SubscribeCancelReturnSaleOrderProductEvent(handler CancelReturnSaleOrderProductHandler) {
	system.bus.Subscribe(string(EventCancelReturnSaleOrderProduct), func(event eventbus.Event) {
		msg := event.Payload.(CancelReturnSaleOrderProductPayload)
		handler(msg)
	})
}
