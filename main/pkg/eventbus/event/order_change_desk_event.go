package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// ChangeDeskPayload “转台”事件数据结构
type ChangeDeskPayload struct {
	BasePayload
	Old Table `json:"old"`
	New Table `json:"new"`
}

type Table struct {
	TableId uint64 `json:"table_id"`
	TableNo string `json:"table_no"`
}

func (payload *ChangeDeskPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// ChangeDeskHandler 转台事件处理器
type ChangeDeskHandler func(msg ChangeDeskPayload)

// PublishChangeDeskEvent 发布转台事件
func (system *SystemEventBus) PublishChangeDeskEvent(msg ChangeDeskPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventChangeDesk), Payload: msg})
}

// SubscribeChangeDeskEvent 订阅转台事件
func (system *SystemEventBus) SubscribeChangeDeskEvent(handler ChangeDeskHandler) {
	system.bus.Subscribe(string(EventChangeDesk), func(event eventbus.Event) {
		msg := event.Payload.(ChangeDeskPayload)
		handler(msg)
	})
}
