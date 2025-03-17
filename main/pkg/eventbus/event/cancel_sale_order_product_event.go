package event

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventCancelSaleOrderProduct “退菜”订单商品事件名称
const EventCancelSaleOrderProduct EventName = "Event_Cancel_Sale_Order_Product"

// CancelSaleOrderProductPayload “退菜”订单商品事件数据结构
type CancelSaleOrderProductPayload struct {
	BasePayload
	OrderProductId  uint64               `json:"order_product_id"` // 订单商品ID
	ProductId       uint64               `json:"product_id"`       // 商品ID
	ProductName     dto.LocaleResponse   `json:"product_name"`     // 商品名称
	ProductAttr     dto.LocaleResponse   `json:"product_attr"`     // 商品属性
	ProductAttrList []dto.LocaleResponse `json:"product_attrs"`    // 商品属性, 包含规格、属性、小料
	TotalNum        uint                 `json:"total_num"`        // 总数量。退菜的数量
	IsBuffet        bool                 `json:"is_buffet"`        // 是否自助餐
	Remark          string               `json:"remark"`           // 备注
	Reason          dto.LocaleResponse   `json:"reason"`           // 退菜原因
	CustomReason    string               `json:"custom_reason"`    // 自定义退菜原因
}

func (payload *CancelSaleOrderProductPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// CancelSaleOrderProductHandler 退菜事件处理器
type CancelSaleOrderProductHandler func(msg CancelSaleOrderProductPayload)

// PublishCancelSaleOrderProductEvent 发布退菜事件
func (system *SystemEventBus) PublishCancelSaleOrderProductEvent(msg CancelSaleOrderProductPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventCancelSaleOrderProduct), Payload: msg})
}

// SubscribeCancelSaleOrderProductEvent 订阅退菜事件
func (system *SystemEventBus) SubscribeCancelSaleOrderProductEvent(handler CancelSaleOrderProductHandler) {
	system.bus.Subscribe(string(EventCancelSaleOrderProduct), func(event eventbus.Event) {
		msg := event.Payload.(CancelSaleOrderProductPayload)
		handler(msg)
	})
}
