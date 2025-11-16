# 事件总线实现设计文档

### 1. 简介

本文档旨在详细描述 TTPOS-Server-Go 项目中事件总线的实现设计。事件总线作为一种发布/订阅模式的实现，旨在解耦系统中的各个模块，提高系统的可扩展性、可维护性和响应性。通过事件总线，不同的业务模块可以在不直接依赖彼此的情况下进行通信和协作。

### 2. 设计目标

*   **模块解耦**: 降低业务模块之间的直接依赖，使各模块能够独立开发、测试和部署。
*   **异步处理**: 支持将耗时操作或非核心业务逻辑通过事件机制进行异步处理，提高主业务流程的响应速度。
*   **可扩展性**: 方便地添加新的事件类型和事件订阅者，以应对业务需求的变化。
*   **并发安全**: 确保事件的发布和订阅在并发环境下是线程安全的。
*   **易于理解和使用**: 提供清晰的事件定义和简单的 API，方便开发者集成。

### 3. 架构概览

事件总线模块位于 `main/pkg/eventbus` 包中，主要包含以下部分：

*   **核心事件总线 (`eventbus.go`)**: 提供了事件发布、订阅、取消订阅的通用机制，并处理并发安全。
*   **事件定义 (`event/` 目录)**: 集中定义了项目中所有业务事件的名称 (`EventName`) 和对应的负载 (Payload) 结构体，以及方便发布和订阅的辅助方法。
*   **系统事件总线 (`event/init_event_bus.go`)**: 提供事件总线的单例实例，作为业务层与核心事件总线交互的统一入口。

`main/pkg/eventbus` 目录结构：

```
main/pkg/eventbus/
├── event/           # 具体业务事件的定义
│   ├── ... (各种事件文件，如 order_created_event.go)
│   └── init_event_bus.go # 系统事件总线的入口
└── eventbus.go      # 核心事件总线实现
```

### 4. 核心事件总线实现 (`eventbus.go`)

#### 4.1 `Event` 结构体

`Event` 结构体是事件总线中事件的最小单元，包含事件的唯一标识和携带的数据。

```go
// ... existing code ...
type Event struct {
	Name    string
	Payload interface{}
}
// ... existing code ...
```

*   `Name`: 事件的唯一名称，用于区分不同类型的事件。
*   `Payload`: 事件携带的数据，类型为 `interface{}`，允许携带任何类型的数据。在实际使用中，通常会定义具体的结构体作为 Payload。

#### 4.2 `EventHandler` 和 `Subscriber`

*   **`EventHandler`**: 定义了事件处理函数的类型签名，即如何处理一个接收到的 `Event`。
*   **`Subscriber`**: 简单封装了 `EventHandler`，用于存储事件的订阅者。

```go
// ... existing code ...
type EventHandler func(event Event)

type Subscriber struct {
	handler EventHandler
}
// ... existing code ...
```

#### 4.3 `EventBus` 核心结构

`EventBus` 是事件总线的核心实现，负责管理事件的订阅者列表并处理并发访问。

```go
// ... existing code ...
type EventBus struct {
	subscribers map[string][]*Subscriber
	mu sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]*Subscriber),
	}
}
// ... existing code ...
```

*   `subscribers`: 一个 `map`，键是事件名称 (`string`)，值是该事件的所有订阅者 (`[]*Subscriber`) 列表。
*   `mu`: `sync.RWMutex` 读写锁，用于保护 `subscribers` 映射的并发访问。写操作 (订阅、取消订阅) 使用写锁，读操作 (发布) 使用读锁，提高了并发性能。

#### 4.4 核心方法

*   **`Subscribe(eventName string, handler EventHandler)`**:
    *   为指定的 `eventName` 注册一个 `handler`。
    *   使用写锁 `eb.mu.Lock()` 确保订阅操作的线程安全。
    *   如果 `eventName` 尚无订阅者列表，则会创建一个新的空列表。
    *   将 `handler` 封装为 `Subscriber` 并添加到对应事件的订阅者列表中。

*   **`Unsubscribe(eventName string, handler EventHandler)`**:
    *   从指定的 `eventName` 中移除一个 `handler`。
    *   使用写锁 `eb.mu.Lock()`。
    *   通过比较 `reflect.ValueOf(subscriber.handler).Pointer()` (函数指针地址) 来精确匹配并移除目标订阅者。

*   **`Publish(event Event)`**:
    *   发布一个 `event`。
    *   使用读锁 `eb.mu.RLock()` 确保发布操作的线程安全，允许多个发布者同时读取订阅者列表。
    *   遍历 `event.Name` 对应的所有订阅者，并调用每个订阅者的 `handler` 函数。

### 5. 事件定义与系统事件总线 (`event/` 目录和 `init_event_bus.go`)

#### 5.1 `EventName` 定义

在 `init_event_bus.go` 中，定义了项目中所有业务事件的名称，使用 `EventName` 类型常量，并按业务领域进行了清晰的划分。

```go
// ... existing code ...
type EventName string

const (
	// =============================================================================
	// H5订单事件
	// =============================================================================
	EventAcceptH5Order EventName = "Event_Accept_H5_Order" // 接单事件
	EventRejectH5Order EventName = "Event_Reject_H5_Order" // 拒单事件
    // ... 其他事件 ...
)
// ... existing code ...
```

这种分类方式使得事件名称一目了然，方便开发者查找和使用。

#### 5.2 `SystemEventBus` 单例

`SystemEventBus` 是对核心 `EventBus` 的封装，并以单例模式提供给整个应用程序。

```go
// ... existing code ...
var systemEventBus *SystemEventBus
var systemBusOnce sync.Once

type SystemEventBus struct {
	bus *eventbus.EventBus
}

func NewSystemBus() *SystemEventBus {
	systemBusOnce.Do(func() {
		systemEventBus = &SystemEventBus{
			bus: eventbus.NewEventBus(),
		}
	})
	return systemEventBus
}
// ... existing code ...
```

*   `sync.Once`: 确保 `SystemEventBus` 实例只被创建一次，避免资源浪费和潜在的并发问题。
*   `bus *eventbus.EventBus`: 嵌入了核心 `EventBus` 实例。

#### 5.3 具体事件的定义和辅助方法 (`event/` 目录下的其他文件)

在 `event/` 目录下的每个具体事件文件（例如 `order_created_event.go`），通常会包含：

*   **事件 Payload 结构体**: 定义事件携带的详细数据。根据项目规范，这些 Payload 结构体通常会嵌入 `BasePayload`，以携带通用的上下文信息 (如 `Ctx`, `CompanyUuid`, `SaleBillUuid` 等)。

    ```go
    // 示例：order_created_event.go
    type OrderCreatedPayload struct {
        BasePayload // 嵌入 BasePayload
        OrderId     uint64  `json:"order_id"`
        OrderAmount float64 `json:"order_amount"`
    }
    ```

*   **`EventHandler` 类型定义**: 为特定的事件定义一个带有具体 Payload 类型的 `EventHandler`。

    ```go
    // 示例：order_created_event.go
    type OrderCreatedHandler func(msg OrderCreatedPayload)
    ```

*   **便捷的 `Publish` 方法**: 在 `SystemEventBus` 上扩展一个 `Publish` 方法，接受具体的 Payload 类型，方便业务层直接发布事件。

    ```go
    // 示例：order_created_event.go (在 SystemEventBus 上)
    func (system *SystemEventBus) PublishOrderCreatedEvent(msg OrderCreatedPayload) {
        system.bus.Publish(eventbus.Event{Name: string(EventOrderCreated), Payload: msg})
    }
    ```

*   **便捷的 `Subscribe` 方法**: 在 `SystemEventBus` 上扩展一个 `Subscribe` 方法，接受具体的 `EventHandler` 类型，并处理类型断言。

    ```go
    // 示例：order_created_event.go (在 SystemEventBus 上)
    func (system *SystemEventBus) SubscribeOrderCreatedEvent(handler OrderCreatedHandler) {
        system.bus.Subscribe(string(EventOrderCreated), func(event eventbus.Event) {
            msg := event.Payload.(OrderCreatedPayload) // 类型断言
            handler(msg)
        })
    }
    ```

### 6. 使用规范 (根据项目架构规范)

*   **定义事件**: 在 `pkg/eventbus/event` 目录下定义新的事件，包括 `EventName` 常量、`Payload` 结构体、`EventHandler` 类型以及 `Publish`/`Subscribe` 辅助方法。
*   **发布事件**: 在服务层 (Service) 中，通过 `go func() { ... }()` 协程异步发布事件，确保不阻塞主业务流程。在发布事件时，需要将当前的 `context.Context` 封装到 `BasePayload` 中。

    ```go
    // 示例：在服务中发布事件
    func (s *orderSrv) CreateOrder(ctx context.Context) error {
        // ... 业务逻辑 ...

        go func() {
            event.NewSystemBus().PublishOrderCreatedEvent(event.OrderCreatedPayload{
                BasePayload: event.BasePayload{
                    Ctx:          ctx,
                    CompanyUuid:  ctx.GetCompanyUuid(),
                    SaleBillUuid: saleBillUuid,
                },
                OrderId:     orderId,
                OrderAmount: amount,
            })
        }()

        return nil
    }
    ```

*   **订阅事件**: 在 `app/event` 目录下创建事件处理器 (例如 `order_created_event_handler.go`)，并在应用程序启动时进行订阅。事件处理逻辑应在独立的函数中实现。

    ```go
    // 示例：在app/event/order_created_event_handler.go 中订阅事件
    func InitOrderCreatedEventHandler() {
        event.NewSystemBus().SubscribeOrderCreatedEvent(func(msg event.OrderCreatedPayload) {
            // 处理订单创建事件
            handleOrderCreated(msg)
        })
    }

    func handleOrderCreated(msg event.OrderCreatedPayload) {
        // 事件处理逻辑
    }
    ```

### 7. 关键考量

*   **并发安全**: `EventBus` 内部使用 `sync.RWMutex` 确保发布和订阅操作的线程安全。
*   **异步处理**: 发布事件时推荐使用 `go func()` 协程进行异步处理，避免阻塞主线程。需要注意协程中的 `Context` 传递问题，通常需要 `ctx.Copy()`。
*   **错误处理**: 事件处理函数内部的错误应妥善处理，避免影响其他订阅者或导致整个应用程序崩溃。
*   **事件负载 (`Payload`)**: `Payload` 的设计应包含事件发生时所需的所有必要信息，但也要避免过于庞大。
*   **订阅者顺序**: 事件总线不保证订阅者被调用的顺序。如果存在顺序依赖，则需要重新考虑设计或在处理函数内部进行协调。
*   **内存泄漏**: `Unsubscribe` 功能对于长期运行的应用程序很重要，以防止不再需要的订阅者占用内存。
*   **分布式场景**: 当前的事件总线是基于内存的，仅适用于单个应用程序实例内的解耦。如果需要跨进程或跨服务的事件通信，则需要引入消息队列 (如 RabbitMQ、Kafka) 等分布式事件系统。

### 8. 总结

本项目中的事件总线设计提供了一个高效、可扩展且并发安全的模块解耦机制。通过清晰的事件定义、单例模式的系统事件总线以及规范的发布/订阅流程，极大地提高了代码的可维护性和业务流程的灵活性。在实际应用中，开发者应严格遵循使用规范，并关注异步处理、错误处理和分布式场景下的扩展性考量。
