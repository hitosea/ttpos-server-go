package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventStatisticsSale 统计销售事件
const EventStatisticsSale EventName = "Event_Statistics_Sale"

// StatisticsSalePayload 每个事件有一个数据结构
type StatisticsSalePayload struct {
	BasePayload
	SaleBillUuid uint64
	OnlyDelete   bool
}

func (payload *StatisticsSalePayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// StatisticsSaleHandler 每个事件的处理器
type StatisticsSaleHandler func(msg StatisticsSalePayload)

// PublishStatisticsSaleEvent 发布统计销售事件
func (system *SystemEventBus) PublishStatisticsSaleEvent(msg StatisticsSalePayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventStatisticsSale), Payload: msg})
}

// SubscribeStatisticsSaleEvent 订阅统计销售事件
func (system *SystemEventBus) SubscribeStatisticsSaleEvent(handler StatisticsSaleHandler) {
	system.bus.Subscribe(string(EventStatisticsSale), func(event eventbus.Event) {
		msg := event.Payload.(StatisticsSalePayload)
		handler(msg)
	})
}
