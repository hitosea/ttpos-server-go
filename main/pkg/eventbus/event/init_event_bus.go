package event

import (
	"sync"
	"ttpos-server-go/pkg/eventbus"
)

// EventName 定义了事件名称的类型，用于标识不同的事件。
type EventName string

// systemEventBus 是 SystemEventBus 的单例实例，用于全局访问。
// systemBusOnce 确保 systemEventBus 只被初始化一次。
var systemEventBus *SystemEventBus
var systemBusOnce sync.Once

// SystemEventBus 是一个包含事件总线的结构体。
// bus 是一个嵌入的 EventBus 实例，用于处理事件的发布和订阅。
type SystemEventBus struct {
	bus *eventbus.EventBus
}

// NewSystemBus 创建并返回 SystemEventBus 的单例实例。
// 使用 sync.Once 确保 SystemEventBus 只被初始化一次，避免多次创建实例。
// 初始化时，会创建一个新的 EventBus 实例并赋值给 SystemEventBus 的 bus 字段。
func NewSystemBus() *SystemEventBus {
	systemBusOnce.Do(func() {
		systemEventBus = &SystemEventBus{
			bus: eventbus.NewEventBus(),
		}
	})
	return systemEventBus
}
