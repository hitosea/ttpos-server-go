# TTPOS API 模块说明

## 📦 模块结构

ttpos-api 采用模块化设计，按服务划分目录，让代码结构更清晰易懂。

### 目录结构

```
ttpos-api/
├── common/                         # 通用组件（所有模块共享）
│   ├── message/                    # 基础消息定义
│   ├── constant/                   # 通用常量
│   └── util/                       # 工具函数
├── ttpos-message/                  # ttpos-message 服务模块
│   ├── message/                    # 消息结构体
│   └── constant/                   # 常量定义
├── ttpos-websocket/                # ttpos-websocket 服务模块
│   ├── message/                    # 消息结构体
│   └── constant/                   # 常量定义
└── examples/                       # 使用示例
```

## 🎯 模块说明

### 1. common - 通用模块

存放所有服务共享的基础组件。

**包含内容**:
- `message/base.go` - 基础消息结构体 `BaseMessage`
- `message/errors.go` - 错误定义
- `constant/status.go` - 通用状态常量（订单状态、会员状态、桌台状态）
- `constant/topic.go` - Topic 常量汇总
- `util/` - 工具函数（JSON、验证、辅助函数）

**使用场景**:
```go
import "ttpos-api/common/message"
import "ttpos-api/common/constant"
import "ttpos-api/util"
```

### 2. ttpos-message - 消息服务模块

专门用于 ttpos-message 服务的消息队列相关定义。

**包含内容**:
- 3 种消息结构体：
  - `MessageSendMessage` - 消息发送
  - `MessageRetryMessage` - 消息重试
  - `MessageStatusChangeMessage` - 状态变更
- 3 个 Topic 常量
- 消息类型常量（email/sms）
- 消息状态常量

**使用场景**:
```go
import (
    msgMessage "ttpos-api/ttpos-message/message"
    msgConstant "ttpos-api/ttpos-message/constant"
)

// 创建消息
msg := msgMessage.NewMessageSendMessage("uuid", msgConstant.MessageTypeEmail)
```

### 3. ttpos-websocket - WebSocket 服务模块

专门用于 ttpos-websocket 服务的实时通信相关定义。

**包含内容**:
- 8 种消息结构体：
  - `WebSocketMessage` - 基础 WebSocket 消息
  - `OrderUpdateMessage` - 订单更新
  - `DeskStatusMessage` - 桌台状态
  - `PrinterNotifyMessage` - 打印机通知
  - `KitchenOrderMessage` - 厨房订单
  - `CallWaiterMessage` - 呼叫服务员
  - `SystemNotifyMessage` - 系统通知
  - `OnlineStatusMessage` - 在线状态
- 7 个 Topic 常量
- 动作类型常量（subscribe/notify 等）
- 业务常量（呼叫类型、通知类型、厨房类型等）

**使用场景**:
```go
import (
    wsMessage "ttpos-api/ttpos-websocket/message"
    wsConstant "ttpos-api/ttpos-websocket/constant"
)

// 创建 WebSocket 消息
msg := wsMessage.NewOrderUpdateMessage("client-id", "order-uuid", 1)
```

## 📖 使用指南

### 在 ttpos-bmp 中使用 ttpos-message 模块

```go
package queue

import (
    "ttpos-api/ttpos-message/message"
    "ttpos-api/ttpos-message/constant"
)

func PublishMessage(messageUUID, messageType string) error {
    msg := message.NewMessageSendMessage(messageUUID, messageType)
    
    if err := msg.Validate(); err != nil {
        return err
    }
    
    data, _ := msg.ToJSON()
    return queue.Push(constant.TopicMessageSend, data)
}
```

### 在 websocket 服务中使用 ttpos-websocket 模块

```go
package service

import (
    "ttpos-api/ttpos-websocket/message"
    "ttpos-api/ttpos-websocket/constant"
)

func BroadcastOrderUpdate(orderUUID string, status int, amount float64) {
    msg := message.NewOrderUpdateMessage("", orderUUID, status)
    msg.OrderAmount = amount
    msg.RoomID = "cashier-room"
    
    data, _ := msg.ToJSON()
    ws.BroadcastToRoom(msg.RoomID, data)
}
```

### 在 main 服务中使用

```go
package event

import (
    msgMessage "ttpos-api/ttpos-message/message"
    msgConstant "ttpos-api/ttpos-message/constant"
    wsMessage "ttpos-api/ttpos-websocket/message"
    wsConstant "ttpos-api/ttpos-websocket/constant"
)

// 发送邮件消息
func SendEmailNotification(messageUUID string) {
    msg := msgMessage.NewMessageSendMessage(messageUUID, msgConstant.MessageTypeEmail)
    // ...
}

// 推送 WebSocket 通知
func PushSystemNotify(title, content string) {
    msg := wsMessage.NewSystemNotifyMessage("", wsConstant.NotifyTypeInfo, title, content)
    // ...
}
```

## 🎨 设计优势

### 1. 模块清晰
- 每个服务有独立的目录
- 一眼就能看出哪些消息属于哪个服务

### 2. 避免混淆
- `ttpos-message/constant` - 消息服务的常量
- `ttpos-websocket/constant` - WebSocket 服务的常量
- 不会再搞混 Topic 和常量的归属

### 3. 易于扩展
- 新增服务？创建新的模块目录即可
- 例如：`ttpos-payment/`、`ttpos-order/` 等

### 4. 独立维护
- 每个模块可以独立演进
- 修改 ttpos-message 不会影响 ttpos-websocket

## 🔍 常见问题

### Q: 为什么要分模块？
A: 让代码结构更清晰，不同服务的消息定义分开管理，避免混淆。

### Q: common 目录是做什么的？
A: 存放所有模块共享的基础组件，如基础消息结构体、工具函数等。

### Q: 如何选择导入哪个模块？
A: 根据您要使用的服务选择：
- 用 ttpos-message 服务 → 导入 `ttpos-api/ttpos-message/...`
- 用 ttpos-websocket 服务 → 导入 `ttpos-api/ttpos-websocket/...`

### Q: 能否混用不同模块？
A: 可以！例如在一个服务中同时使用消息队列和 WebSocket：
```go
import (
    msgMessage "ttpos-api/ttpos-message/message"
    wsMessage "ttpos-api/ttpos-websocket/message"
)
```

## 📚 相关文档

- [README.md](README.md) - 项目总览
- [USAGE.md](USAGE.md) - 详细使用指南
- [INTEGRATION.md](INTEGRATION.md) - 集成指南
- [WEBSOCKET.md](WEBSOCKET.md) - WebSocket 详细文档

