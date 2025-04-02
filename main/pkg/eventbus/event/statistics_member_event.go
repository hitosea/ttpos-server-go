package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventStatisticsMember 统计会员事件
const EventStatisticsMember EventName = "Event_Statistics_Member"

// StatisticsMemberPayload 每个事件有一个数据结构
type StatisticsMemberPayload struct {
	BasePayload
	MemberRechargeOrderUuid uint64
	OnlyDelete              bool
}

func (payload *StatisticsMemberPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// StatisticsMemberHandler 每个事件的处理器
type StatisticsMemberHandler func(msg StatisticsMemberPayload)

// PublishStatisticsMemberEvent 发布统计会员事件
func (system *SystemEventBus) PublishStatisticsMemberEvent(msg StatisticsMemberPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventStatisticsMember), Payload: msg})
}

// SubscribeStatisticsMemberEvent 订阅统计会员事件
func (system *SystemEventBus) SubscribeStatisticsMemberEvent(handler StatisticsMemberHandler) {
	system.bus.Subscribe(string(EventStatisticsMember), func(event eventbus.Event) {
		msg := event.Payload.(StatisticsMemberPayload)
		handler(msg)
	})
}
