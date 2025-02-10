package eventbus

import (
	"reflect"
	"sync"
)

// Event 定义事件，包含事件名称和负载。
type Event struct {
	Name    string
	Payload interface{}
}

// EventHandler 是一个函数类型，用于处理事件。
type EventHandler func(event Event)

// Subscriber 包含一个事件处理函数。
type Subscriber struct {
	handler EventHandler
}

// EventBus 是一个事件总线，用于管理事件的订阅者。
// 通过读写锁来保证并发安全。
type EventBus struct {
	// subscribers 是一个映射，键为事件名称，值为订阅者列表。
	subscribers map[string][]*Subscriber
	// mu 是一个读写锁，用于保护 subscribers 的并发访问。
	mu sync.RWMutex
}

// NewEventBus 创建并返回一个新的 EventBus 实例。
// 该实例初始化了一个空的订阅者映射，用于存储事件名称与其订阅者列表的关系。
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]*Subscriber),
	}
}

// Subscribe 为指定的事件名称添加一个新的订阅者。
// 该方法接收事件名称和事件处理函数作为参数。
// 使用写锁保护 subscribers 映射的并发访问，确保线程安全。
// 如果事件名称尚无订阅者，则初始化一个新的订阅者列表。
// 最后，将新的订阅者添加到对应事件名称的订阅者列表中。
func (eb *EventBus) Subscribe(eventName string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if _, ok := eb.subscribers[eventName]; !ok {
		eb.subscribers[eventName] = []*Subscriber{}
	}
	subscriber := &Subscriber{handler: handler}
	eb.subscribers[eventName] = append(eb.subscribers[eventName], subscriber)
}

// Unsubscribe 从指定的事件名称中移除一个订阅者。
// 该方法接收事件名称和事件处理函数作为参数。
// 使用写锁保护 subscribers 映射的并发访问，确保线程安全。
// 遍历订阅者列表，匹配处理函数的指针地址以找到目标订阅者。
// 找到后，从订阅者列表中移除该订阅者。
func (eb *EventBus) Unsubscribe(eventName string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if subscribers, ok := eb.subscribers[eventName]; ok {
		for i, subscriber := range subscribers {
			if reflect.ValueOf(subscriber.handler).Pointer() == reflect.ValueOf(handler).Pointer() {
				// Remove the subscriber from the subscribers list
				eb.subscribers[eventName] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
}

// Publish 发布一个事件到所有订阅者。
// 该方法接收一个事件对象作为参数。
// 使用读锁保护 subscribers 映射的并发访问，确保线程安全。
// 遍历指定事件名称的订阅者列表，调用每个订阅者的处理函数。
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if subscribers, ok := eb.subscribers[event.Name]; ok {
		for _, subscriber := range subscribers {
			subscriber.handler(event)
		}
	}
}
