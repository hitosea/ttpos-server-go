package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// ProductSoldOutPayload 商品沽清事件数据结构
type ProductSoldOutPayload struct {
	BasePayload
	ProductBomUuid uint64   `json:"product_bom_uuid"` // 商品规格UUID
	StockNum       *float64 `json:"stock_num"`        // 库存数量
	IsSoldOut      *bool    `json:"is_sold_out"`      // 是否售罄
}

func (payload *ProductSoldOutPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ProductSoldOutHandler 商品沽清事件处理器
type ProductSoldOutHandler func(msg ProductSoldOutPayload)

// PublishProductSoldOutEvent 发布商品沽清事件
func (system *SystemEventBus) PublishProductSoldOutEvent(msg ProductSoldOutPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventProductSoldOut), Payload: msg})
}

// SubscribeProductSoldOutEvent 订阅商品沽清事件
func (system *SystemEventBus) SubscribeProductSoldOutEvent(handler ProductSoldOutHandler) {
	system.bus.Subscribe(string(EventProductSoldOut), func(event eventbus.Event) {
		msg := event.Payload.(ProductSoldOutPayload)
		handler(msg)
	})
}
