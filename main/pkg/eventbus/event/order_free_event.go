package event

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// FreeSaleOrderPayload 每个事件有一个数据结构
type FreeSaleOrderPayload struct {
	BasePayload
	SaleBill      *model.SaleBill `json:"-"`
	OrderPrice    float64         `json:"order_price"`    // 订单应付金额
	PayPrice      float64         `json:"pay_price"`      // 最终应付金额。最终应付金额=订单应付金额+手续费（手续费=每笔付款单的手续费之和）
	ActualPrice   float64         `json:"actual_price"`   // 最终实付金额。最终实付金额=最终应付金额+找零金额。如果没有找零，则最终实付金额=最终应付金额。最终实付金额=每笔付款单的付款金额之和（含手续费）
	ChangeDue     float64         `json:"change_due"`     // 找零金额
	IsFree        uint            `json:"is_free"`        // 是否免单
	DiscountMoney float64         `json:"discount_money"` // 免单金额
	// 授权员工信息（可选，仅在使用了授权验证时存在）
	AuthorizedStaff *AuthorizedStaffInfo `json:"authorized_staff,omitempty"`
}

func (payload *FreeSaleOrderPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// FreeSaleOrderHandler 每个事件的处理器
type FreeSaleOrderHandler func(msg FreeSaleOrderPayload)

// PublishFreeSaleOrderEvent 发布免单事件
func (system *SystemEventBus) PublishFreeSaleOrderEvent(msg FreeSaleOrderPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventFreeSaleOrder), Payload: msg})
}

// SubscribeFreeSaleOrderEvent 订阅免单事件
func (system *SystemEventBus) SubscribeFreeSaleOrderEvent(handler FreeSaleOrderHandler) {
	system.bus.Subscribe(string(EventFreeSaleOrder), func(event eventbus.Event) {
		msg := event.Payload.(FreeSaleOrderPayload)
		handler(msg)
	})
}
