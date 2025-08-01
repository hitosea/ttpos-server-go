package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// MergeDeskPayload 开台事件数据结构
type MergeDeskPayload struct {
	BasePayload
	DeskNos []string
}

func (payload *MergeDeskPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// MergeDeskHandler 开台事件处理器
type MergeDeskHandler func(msg MergeDeskPayload)

// PublishMergeDeskEvent 发布开台事件
func (system *SystemEventBus) PublishMergeDeskEvent(msg MergeDeskPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventMergeDesk), Payload: msg})
}

// SubscribeMergeDeskEvent 订阅开台事件
func (system *SystemEventBus) SubscribeMergeDeskEvent(handler MergeDeskHandler) {
	system.bus.Subscribe(string(EventMergeDesk), func(event eventbus.Event) {
		msg := event.Payload.(MergeDeskPayload)
		handler(msg)
	})
}
