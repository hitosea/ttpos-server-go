package event

import (
	"ttpos-server-go/pkg/eventbus"
)

// ChangeMemberPointsPayload “会员积分变动”事件数据结构
type ChangeMemberPointsPayload struct {
	BasePayload
}

// ChangeMemberPointsHandler 加积分事件处理器
type ChangeMemberPointsHandler func(msg ChangeMemberPointsPayload)

// PublishChangeMemberPointsEvent 发布“会员积分变动”事件
func (system *SystemEventBus) PublishChangeMemberPointsEvent(msg ChangeMemberPointsPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangeMemberPoints), Payload: msg})
}

// SubscribeChangeMemberPointsEvent 订阅“会员积分变动”事件
func (system *SystemEventBus) SubscribeChangeMemberPointsEvent(handler ChangeMemberPointsHandler) {
	system.bus.Subscribe(string(EventChangeMemberPoints), func(event eventbus.Event) {
		msg := event.Payload.(ChangeMemberPointsPayload)
		handler(msg)
	})
}
