package event

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventReturnOrder 用餐订单退款事件名称
const EventReturnOrder EventName = "Event_Return_Order"

type RefundPayType struct {
	Name          string  `json:"name"`           // 退款支付方式名称
	Code          int     `json:"code"`           // 退款支付方式代号
	Amount        float64 `json:"amount"`         // 退款金额
	PaymentStatus int     `json:"payment_status"` // 支付状态
}

// ReturnOrderPayload 用餐订单退款事件数据结构
type ReturnOrderPayload struct {
	BasePayload
	SaleBill     *model.SaleBill `json:"-"`
	Products     Products        `json:"products"`       // 退款商品
	PayTypes     []RefundPayType `json:"pay_type"`       // 支付方式
	RefundType   int             `json:"refund_type"`    // 退款方式：1-整单退款；2-部分退款
	IsSplitOrder bool            `json:"is_split_order"` // 是否拆单
	Index        int             `json:"index"`          // 子单索引
}

func (payload *ReturnOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ReturnOrderHandler 用餐订单退款事件处理器
type ReturnOrderHandler func(msg ReturnOrderPayload)

// PublishReturnOrderEvent 发布用餐订单退款事件
func (system *SystemEventBus) PublishReturnOrderEvent(msg ReturnOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventReturnOrder), Payload: msg})
}

// SubscribeReturnOrderEvent 订阅用餐订单退款事件
func (system *SystemEventBus) SubscribeReturnOrderEvent(handler ReturnOrderHandler) {
	system.bus.Subscribe(string(EventReturnOrder), func(event eventbus.Event) {
		msg := event.Payload.(ReturnOrderPayload)
		handler(msg)
	})
}
