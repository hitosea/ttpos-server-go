package event

import (
	"sync"
	"ttpos-server-go/pkg/eventbus/event"
)

var once_cancel_doing_product_event_handler sync.Once

// init 自动注册"取消进行中的产品"事件处理器
func init() {
	// 只初始化一次
	cancelDoingProductEventHandler()
}

// cancelDoingProductEventHandler "取消进行中的产品"事件处理器
func cancelDoingProductEventHandler() {
	once_cancel_doing_product_event_handler.Do(func() {
		event.NewSystemBus().SubscribeCancelDoingProductEvent(func(msg event.CancelDoingProductPayload) {
			// 通知厨房取消制作
			// 给websocket服务端推送消息，再有websocket服务端通知厨房
		})
	})
}
