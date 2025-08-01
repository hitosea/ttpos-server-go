package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventWrapSaleOrderProduct 打包事件
const EventWrapSaleOrderProduct EventName = "Event_Wrap_Sale_Order_Product"

// WrapSaleOrderProductPayload 每个事件有一个数据结构
type WrapSaleOrderProductPayload struct {
	BasePayload
	SaleOrderProductUuid uint64             `json:"sale_order_product_uuid"` // 订单商品ID
	ProductPackageUuid   uint64             `json:"product_package_uuid"`    // 商品包ID
	ProductName          dto.LocaleResponse `json:"product_name"`            // 商品名称
	ProductAttr          dto.LocaleResponse `json:"product_attr"`            // 商品属性
	ProductPrice         float64            `json:"product_price"`           // 商品价格
	Num                  float64            `json:"num"`                     // 商品数量
}

func (payload *WrapSaleOrderProductPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// WrapSaleOrderProductHandler 每个事件的处理器
type WrapSaleOrderProductHandler func(msg WrapSaleOrderProductPayload)

// PublishWrapSaleOrderProductEvent 发布打包事件
func (system *SystemEventBus) PublishWrapSaleOrderProductEvent(msg WrapSaleOrderProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventWrapSaleOrderProduct), Payload: msg})
}

// SubscribeWrapSaleOrderProductEvent 订阅打包事件
func (system *SystemEventBus) SubscribeWrapSaleOrderProductEvent(handler WrapSaleOrderProductHandler) {
	system.bus.Subscribe(string(EventWrapSaleOrderProduct), func(event eventbus.Event) {
		msg := event.Payload.(WrapSaleOrderProductPayload)
		handler(msg)
	})
}
