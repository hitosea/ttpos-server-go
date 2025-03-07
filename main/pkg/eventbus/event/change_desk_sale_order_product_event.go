package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventChangeDeskSaleOrderProduct 事件名称，每个事件都有一个全局唯一的名称
const EventChangeDeskSaleOrderProduct EventName = "Event_Change_Desk_Sale_Order_Product"

// ChangeDeskSaleOrderProductPayload 每个事件有一个数据结构
type ChangeDeskSaleOrderProductPayload struct {
	BasePayload
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	TotalNum       uint               `json:"total_num"`        // 总数量
	ToOrderId      uint64             `json:"to_order_id"`      // 目标订单ID
	ToTableId      uint64             `json:"to_table_id"`      // 目标桌台ID
	ToTableNo      string             `json:"to_table_no"`      // 目标桌台编号
}

func (payload *ChangeDeskSaleOrderProductPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ChangeDeskSaleOrderProductHandler 每个事件的处理器
type ChangeDeskSaleOrderProductHandler func(msg ChangeDeskSaleOrderProductPayload)

// PublishChangeDeskSaleOrderProductEvent 发布转菜事件
func (system *SystemEventBus) PublishChangeDeskSaleOrderProductEvent(msg ChangeDeskSaleOrderProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangeDeskSaleOrderProduct), Payload: msg})
}

// SubscribeChangeDeskSaleOrderProductEvent 订阅转菜事件
func (system *SystemEventBus) SubscribeChangeDeskSaleOrderProductEvent(handler ChangeDeskSaleOrderProductHandler) {
	system.bus.Subscribe(string(EventChangeDeskSaleOrderProduct), func(event eventbus.Event) {
		msg := event.Payload.(ChangeDeskSaleOrderProductPayload)
		handler(msg)
	})
}
