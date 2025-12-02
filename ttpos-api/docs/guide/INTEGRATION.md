# TTPOS API 集成指南

本文档说明如何在各个服务中集成 `ttpos-api` 包。

## 📋 目录

- [在 ttpos-bmp 中集成](#在-ttpos-bmp-中集成)
- [在 main 服务中集成](#在-main-服务中集成)
- [在 websocket 服务中集成](#在-websocket-服务中集成)
- [迁移现有代码](#迁移现有代码)

## 在 ttpos-bmp 中集成

### 1. 修改 go.mod

编辑 `/home/coder/workspaces/ttpos-server-go/ttpos-bmp/go.mod`：

```go
module ttpos-bmp

go 1.23

require (
    ttpos-api v0.0.0
    github.com/gogf/gf/v2 v2.8.4
    // ... 其他依赖
)

replace ttpos-api => ../ttpos-api
```

### 2. 执行依赖更新

```bash
cd /home/coder/workspaces/ttpos-server-go/ttpos-bmp
go mod tidy
```

### 3. 迁移消息结构体

#### 原代码（ttpos-bmp/app/ttpos-message/internal/model/dto/message_dto.go）

```go
// RocketMQMessage RocketMQ 消息体
type RocketMQMessage struct {
	MessageUuid string `json:"message_uuid" description:"消息UUID"`
	MessageType string `json:"message_type" description:"消息类型"`
}
```

#### 新代码（使用 ttpos-api）

```go
import (
    "ttpos-api/message"
    "ttpos-api/constant"
)

// 发送消息
func PublishMessage(ctx context.Context, messageUUID, messageType string) error {
    // 使用 ttpos-api 的消息结构体
    msg := message.NewMessageSendMessage(messageUUID, messageType)
    msg.WithCompanyUUID(getCompanyUUID(ctx))
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return err
    }
    
    // 发送到队列
    return queue.Push(constant.TopicMessageSend, msg)
}
```

### 4. 更新队列服务

#### 修改 internal/logic/queue/rocketmq.go

```go
package queue

import (
    "context"
    "ttpos-api/message"
    "ttpos-api/constant"
    "ttpos-bmp/internal/pkg/queue"
)

// PublishMessage 发布消息到队列
func (s *sQueue) PublishMessage(ctx context.Context, messageUUID, messageType string) error {
    if !s.enabled {
        return gerror.New("队列服务未启用")
    }

    // 使用 ttpos-api 的消息结构体
    msg := message.NewMessageSendMessage(messageUUID, messageType)
    msg.WithCompanyUUID(getCompanyUUID(ctx))
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return gerror.Wrap(err, "消息验证失败")
    }

    // 发送消息
    if err := queue.Push(constant.TopicMessageSend, msg); err != nil {
        g.Log().Errorf(ctx, "发送消息失败: %v", err)
        return gerror.Wrapf(err, "发送消息失败")
    }

    g.Log().Info(ctx, "消息已发送",
        "message_id", msg.MessageID,
        "message_uuid", msg.MessageUUID,
        "type", msg.MessageType,
    )

    return nil
}
```

### 5. 更新消费者

#### 修改 internal/logic/consumer/mailgun_consumer.go

```go
package consumer

import (
    "context"
    "ttpos-api/message"
    "ttpos-api/constant"
)

// MailgunConsumer 邮件消费者
type MailgunConsumer struct{}

// GetTopic 获取订阅的 Topic
func (c *MailgunConsumer) GetTopic() string {
    return constant.TopicMessageSend
}

// Consume 消费消息
func (c *MailgunConsumer) Consume(ctx context.Context, data []byte) error {
    // 使用 ttpos-api 的消息结构体
    var msg message.MessageSendMessage
    if err := msg.FromJSON(data); err != nil {
        return gerror.Wrap(err, "解析消息失败")
    }
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return gerror.Wrap(err, "消息验证失败")
    }
    
    // 处理消息
    return c.handleMessage(ctx, &msg)
}

// handleMessage 处理消息
func (c *MailgunConsumer) handleMessage(ctx context.Context, msg *message.MessageSendMessage) error {
    g.Log().Info(ctx, "开始处理消息",
        "message_id", msg.MessageID,
        "message_uuid", msg.MessageUUID,
        "type", msg.MessageType,
    )
    
    // 从数据库获取消息详情
    record, err := service.Message().GetMessageByUuid(ctx, msg.MessageUUID)
    if err != nil {
        return gerror.Wrap(err, "获取消息详情失败")
    }
    
    // 发送邮件
    return service.Mailgun().SendEmail(ctx, record.Uuid, record.Recipient, record.Subject, record.Content)
}
```

### 6. 更新常量使用

#### 原代码（internal/consts/message_consts.go）

```go
const (
    TopicMessageSend = "message.send"
    MessageTypeEmail = "email"
    MessageTypeSMS   = "sms"
)
```

#### 新代码（使用 ttpos-api）

```go
import "ttpos-api/constant"

// 直接使用 ttpos-api 的常量
// constant.TopicMessageSend
// constant.MessageTypeEmail
// constant.MessageTypeSMS
```

## 在 main 服务中集成

### 1. 修改 go.mod

编辑 `/home/coder/workspaces/ttpos-server-go/main/go.mod`：

```go
module ttpos-server-go/main

go 1.23

require (
    ttpos-api v0.0.0
    github.com/gin-gonic/gin v1.9.1
    // ... 其他依赖
)

replace ttpos-api => ../ttpos-api
```

### 2. 执行依赖更新

```bash
cd /home/coder/workspaces/ttpos-server-go/main
go mod tidy
```

### 3. 在事件总线中使用

#### 创建消息发布器

```go
// pkg/eventbus/event/message_event.go
package event

import (
    "ttpos-api/message"
    "ttpos-api/constant"
    "ttpos-server-go/main/pkg/eventbus"
)

// PublishMessageSendEvent 发布消息发送事件
func (system *SystemEventBus) PublishMessageSendEvent(messageUUID, messageType string) {
    msg := message.NewMessageSendMessage(messageUUID, messageType)
    
    // 序列化消息
    data, err := msg.ToJSON()
    if err != nil {
        log.Error("消息序列化失败", err)
        return
    }
    
    // 发布到事件总线
    system.bus.Publish(eventbus.Event{
        Name:    constant.TopicMessageSend,
        Payload: data,
    })
}
```

#### 创建消息订阅器

```go
// app/event/message_event_handler.go
package event

import (
    "ttpos-api/message"
    "ttpos-api/constant"
)

// InitMessageEventHandler 初始化消息事件处理器
func InitMessageEventHandler() {
    event.NewSystemBus().SubscribeMessageSendEvent(handleMessageSend)
}

// handleMessageSend 处理消息发送事件
func handleMessageSend(data []byte) {
    var msg message.MessageSendMessage
    if err := msg.FromJSON(data); err != nil {
        log.Error("解析消息失败", err)
        return
    }
    
    // 处理消息
    log.Info("收到消息发送事件",
        "message_uuid", msg.MessageUUID,
        "message_type", msg.MessageType,
    )
}
```

## 在 websocket 服务中集成

### 1. 修改 go.mod

编辑 `/home/coder/workspaces/ttpos-server-go/websocket/go.mod`：

```go
module ttpos-server-go/websocket

go 1.23

require (
    ttpos-api v0.0.0
    github.com/gorilla/websocket v1.5.0
    // ... 其他依赖
)

replace ttpos-api => ../ttpos-api
```

### 2. 执行依赖更新

```bash
cd /home/coder/workspaces/ttpos-server-go/websocket
go mod tidy
```

### 3. 在 WebSocket 服务中使用

```go
package service

import (
    "ttpos-api/message"
    "ttpos-api/constant"
)

// BroadcastMessageStatus 广播消息状态变更
func (s *WebSocketService) BroadcastMessageStatus(messageUUID string, oldStatus, newStatus int) {
    // 使用 ttpos-api 的消息结构体
    msg := message.NewMessageStatusChangeMessage(messageUUID, constant.MessageTypeEmail, oldStatus, newStatus)
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        log.Error("消息验证失败", err)
        return
    }
    
    // 序列化消息
    data, err := msg.ToJSON()
    if err != nil {
        log.Error("消息序列化失败", err)
        return
    }
    
    // 广播到所有连接的客户端
    s.broadcast(data)
}
```

## 迁移现有代码

### 迁移检查清单

- [ ] 更新 go.mod 文件
- [ ] 执行 `go mod tidy`
- [ ] 替换自定义消息结构体为 ttpos-api 的结构体
- [ ] 替换常量定义为 ttpos-api 的常量
- [ ] 更新消息发送代码
- [ ] 更新消息接收代码
- [ ] 更新测试代码
- [ ] 运行测试确保功能正常
- [ ] 删除旧的消息结构体定义（可选）

### 迁移步骤

#### 1. 识别需要迁移的代码

查找项目中所有定义消息结构体的地方：

```bash
# 在 ttpos-bmp 中
grep -r "type.*Message struct" ttpos-bmp/app/ttpos-message/

# 在 main 中
grep -r "type.*Message struct" main/app/
```

#### 2. 逐步替换

不要一次性替换所有代码，建议按模块逐步迁移：

1. 先迁移 ttpos-message 服务
2. 再迁移 main 服务
3. 最后迁移 websocket 服务

#### 3. 保持兼容性

在迁移过程中，可以同时保留旧的和新的代码，确保平滑过渡：

```go
// 临时兼容代码
func convertOldToNew(old *OldMessage) *message.MessageSendMessage {
    msg := message.NewMessageSendMessage(old.MessageUUID, old.MessageType)
    msg.WithCompanyUUID(old.CompanyUUID)
    return msg
}
```

#### 4. 测试验证

每迁移一个模块，都要进行充分的测试：

```bash
# 运行单元测试
go test ./...

# 运行集成测试
make test

# 手动测试关键功能
```

## 常见问题

### Q1: 如何处理版本冲突？

A: 使用 `replace` 指令指向本地路径，确保所有服务使用相同版本的 ttpos-api。

### Q2: 如何添加新的消息类型？

A: 在 ttpos-api 中添加新的消息结构体，然后在各服务中更新依赖即可。

### Q3: 如何保证消息格式兼容性？

A: 遵循语义化版本号，不要修改已有字段，只能新增可选字段。

### Q4: 迁移过程中出现编译错误怎么办？

A: 检查导入路径是否正确，确保执行了 `go mod tidy`。

## 技术支持

如有问题，请联系 TTPOS 开发团队。

