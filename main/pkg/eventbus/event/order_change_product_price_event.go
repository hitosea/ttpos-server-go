package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// ChangeSaleOrderProductPricePayload “修改销售订单产品价格”事件数据结构
type ChangeSaleOrderProductPricePayload struct {
	BasePayload
	OrderProductId uint64             `json:"order_product_id"` // 销售订单产品ID
	ProductId      uint64             `json:"product_id"`       // 产品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 产品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 产品属性
	TotalNum       float64            `json:"total_num"`        // 数量
	Price          float64            `json:"price"`            // 价格，单价
}

func (payload *ChangeSaleOrderProductPricePayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ChangeSaleOrderProductPriceHandler 修改销售订单产品价格事件处理器
type ChangeSaleOrderProductPriceHandler func(msg ChangeSaleOrderProductPricePayload)

// PublishChangeSaleOrderProductPriceEvent 发布修改销售订单产品价格事件
func (system *SystemEventBus) PublishChangeSaleOrderProductPriceEvent(msg ChangeSaleOrderProductPricePayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangeSaleOrderProductPrice), Payload: msg})
}

// SubscribeChangeSaleOrderProductPriceEvent 订阅修改销售订单产品价格事件
func (system *SystemEventBus) SubscribeChangeSaleOrderProductPriceEvent(handler ChangeSaleOrderProductPriceHandler) {
	system.bus.Subscribe(string(EventChangeSaleOrderProductPrice), func(event eventbus.Event) {
		msg := event.Payload.(ChangeSaleOrderProductPricePayload)
		handler(msg)
	})
}
