package event

import (
	"ttpos-server-go/pkg/eventbus"
)

// EventChangeMemberBalance “会员余额变动”事件
const EventChangeMemberBalance EventName = "Event_Change_Member_Balance"

// ChangeMemberBalancePayload “会员余额变动”事件数据结构
type ChangeMemberBalancePayload struct {
	BasePayload
}

// ChangeMemberBalanceHandler 加库存事件处理器
type ChangeMemberBalanceHandler func(msg ChangeMemberBalancePayload)

// PublishChangeMemberBalanceEvent 发布“会员余额变动”事件
func (system *SystemEventBus) PublishChangeMemberBalanceEvent(msg ChangeMemberBalancePayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangeMemberBalance), Payload: msg})
}

// SubscribeChangeMemberBalanceEvent 订阅“会员余额变动”事件
func (system *SystemEventBus) SubscribeChangeMemberBalanceEvent(handler ChangeMemberBalanceHandler) {
	system.bus.Subscribe(string(EventChangeMemberBalance), func(event eventbus.Event) {
		msg := event.Payload.(ChangeMemberBalancePayload)
		handler(msg)
	})
}
