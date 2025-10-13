package event

import (
	"sync"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_sent_cooking_pre_event_handler sync.Once

// sentCookingPreEventHandler "预送厨"事件处理器
func sentCookingPreEventHandler() {
	once_sent_cooking_pre_event_handler.Do(func() {

		event.NewSystemBus().SubscribeSentCookingPreEvent(func(payload event.SentCookingPrePayload) {

		})
	})
}
