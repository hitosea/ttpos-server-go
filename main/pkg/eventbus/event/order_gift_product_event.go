package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// GiftSaleOrderProductPayload 每个事件有一个数据结构
type GiftSaleOrderProductPayload struct {
	BasePayload
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	ProductPrice   float64            `json:"product_price"`    // 商品价格
	TotalNum       float64            `json:"total_num"`        // 总数量
	TotalPrice     float64            `json:"total_price"`      // 总价格
	FreeTagIds     []uint64           `json:"free_tag_ids"`     // 赠菜原因ID
	FreeRemark     string             `json:"free_remark"`      // 赠菜自定义原因
}

func (payload *GiftSaleOrderProductPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// GiftSaleOrderProductHandler 每个事件的处理器
type GiftSaleOrderProductHandler func(msg GiftSaleOrderProductPayload)

// PublishGiftSaleOrderProductEvent 发布赠菜事件
func (system *SystemEventBus) PublishGiftSaleOrderProductEvent(msg GiftSaleOrderProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventGiftSaleOrderProduct), Payload: msg})
}

// SubscribeGiftSaleOrderProductEvent 订阅赠菜事件
func (system *SystemEventBus) SubscribeGiftSaleOrderProductEvent(handler GiftSaleOrderProductHandler) {
	system.bus.Subscribe(string(EventGiftSaleOrderProduct), func(event eventbus.Event) {
		msg := event.Payload.(GiftSaleOrderProductPayload)
		handler(msg)
	})
}
