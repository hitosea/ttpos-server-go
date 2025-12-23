package event

import "time"

// DomainEvent 领域事件基础接口
type DomainEvent interface {
	// EventName 事件名称
	EventName() string
	// OccurredAt 事件发生时间
	OccurredAt() time.Time
	// AggregateID 聚合根ID
	AggregateID() uint64
}

// BaseDomainEvent 领域事件基类
type BaseDomainEvent struct {
	occurredAt  time.Time
	aggregateID uint64
}

// OccurredAt 返回事件发生时间
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// AggregateID 返回聚合根ID
func (e BaseDomainEvent) AggregateID() uint64 {
	return e.aggregateID
}

// NewBaseDomainEvent 创建基础领域事件
func NewBaseDomainEvent(aggregateID uint64) BaseDomainEvent {
	return BaseDomainEvent{
		occurredAt:  time.Now(),
		aggregateID: aggregateID,
	}
}

// EventPublisher 事件发布者接口
type EventPublisher interface {
	// Publish 发布事件
	Publish(event DomainEvent)
	// PublishAll 批量发布事件
	PublishAll(events []DomainEvent)
}

// EventSubscriber 事件订阅者接口
type EventSubscriber interface {
	// Handle 处理事件
	Handle(event DomainEvent) error
	// SubscribedEvents 订阅的事件类型
	SubscribedEvents() []string
}

// EventDispatcher 事件分发器
type EventDispatcher struct {
	subscribers map[string][]EventSubscriber
}

// NewEventDispatcher 创建事件分发器
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		subscribers: make(map[string][]EventSubscriber),
	}
}

// Register 注册订阅者
func (d *EventDispatcher) Register(subscriber EventSubscriber) {
	for _, eventName := range subscriber.SubscribedEvents() {
		d.subscribers[eventName] = append(d.subscribers[eventName], subscriber)
	}
}

// Publish 发布事件
func (d *EventDispatcher) Publish(event DomainEvent) {
	if subscribers, ok := d.subscribers[event.EventName()]; ok {
		for _, subscriber := range subscribers {
			// 异步处理可以考虑使用 goroutine
			_ = subscriber.Handle(event)
		}
	}
}

// PublishAll 批量发布事件
func (d *EventDispatcher) PublishAll(events []DomainEvent) {
	for _, event := range events {
		d.Publish(event)
	}
}

// 全局事件分发器
var globalDispatcher = NewEventDispatcher()

// GetDispatcher 获取全局事件分发器
func GetDispatcher() *EventDispatcher {
	return globalDispatcher
}
