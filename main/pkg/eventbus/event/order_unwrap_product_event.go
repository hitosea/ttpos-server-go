package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// UnwrapSaleOrderProductPayload 每个事件有一个数据结构
type UnwrapSaleOrderProductPayload struct {
	BasePayload
	SaleOrderProductUuid uint64             `json:"sale_order_product_uuid"` // 订单商品ID
	ProductPackageUuid   uint64             `json:"product_package_uuid"`    // 商品包ID
	ProductName          dto.LocaleResponse `json:"product_name"`            // 商品名称
	ProductAttr          dto.LocaleResponse `json:"product_attr"`            // 商品属性
	ProductPrice         float64            `json:"product_price"`           // 商品价格
	Num                  float64            `json:"num"`                     // 商品数量
}

func (payload *UnwrapSaleOrderProductPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// UnwrapSaleOrderProductHandler 每个事件的处理器
type UnwrapSaleOrderProductHandler func(msg UnwrapSaleOrderProductPayload)

// PublishUnwrapSaleOrderProductEvent 发布取消打包事件
func (system *SystemEventBus) PublishUnwrapSaleOrderProductEvent(msg UnwrapSaleOrderProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventUnwrapSaleOrderProduct), Payload: msg})
}

// SubscribeUnwrapSaleOrderProductEvent 订阅取消打包事件
func (system *SystemEventBus) SubscribeUnwrapSaleOrderProductEvent(handler UnwrapSaleOrderProductHandler) {
	system.bus.Subscribe(string(EventUnwrapSaleOrderProduct), func(event eventbus.Event) {
		msg := event.Payload.(UnwrapSaleOrderProductPayload)
		handler(msg)
	})
}
