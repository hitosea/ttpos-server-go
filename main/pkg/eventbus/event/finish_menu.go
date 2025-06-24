package event

import (
	"ttpos-server-go/pkg/eventbus"
	"ttpos-server-go/pkg/utils"
)

// EventFinishMenu 完成制作事件名称
const EventFinishMenu EventName = "Event_Finish_Menu"

// FinishMenuPayload 完成制作事件数据结构
type FinishMenuPayload struct {
	BasePayload
	Products Products `json:"products"` // 商品列表
}

func (payload *FinishMenuPayload) ToJsonString() string {
	return utils.ToJson(payload)
}

// FinishMenuHandler 完成制作事件处理器
type FinishMenuHandler func(msg FinishMenuPayload)

// PublishFinishMenuEvent 发布完成制作事件
func (system *SystemEventBus) PublishFinishMenuEvent(msg FinishMenuPayload) {
	system.bus.Publish(eventbus.Event{Name: string(EventFinishMenu), Payload: msg})
}

// SubscribeFinishMenuEvent 订阅完成制作事件
func (system *SystemEventBus) SubscribeFinishMenuEvent(handler FinishMenuHandler) {
	system.bus.Subscribe(string(EventFinishMenu), func(event eventbus.Event) {
		msg := event.Payload.(FinishMenuPayload)
		handler(msg)
	})
}
