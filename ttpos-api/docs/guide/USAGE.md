# TTPOS API 使用指南

## 📚 目录

- [安装配置](#安装配置)
- [基础使用](#基础使用)
- [消息发送](#消息发送)
- [消息接收](#消息接收)
- [常量使用](#常量使用)
- [工具函数](#工具函数)
- [最佳实践](#最佳实践)

## 安装配置

### 在 main 服务中使用

编辑 `main/go.mod`：

```go
module ttpos-server-go/main

go 1.23

require (
    ttpos-api v0.0.0
    // ... 其他依赖
)

replace ttpos-api => ../ttpos-api
```

### 在 ttpos-bmp 服务中使用

编辑 `ttpos-bmp/go.mod`：

```go
module ttpos-bmp

go 1.23

require (
    ttpos-api v0.0.0
    // ... 其他依赖
)

replace ttpos-api => ../ttpos-api
```

### 在 websocket 服务中使用

编辑 `websocket/go.mod`：

```go
module ttpos-server-go/websocket

go 1.23

require (
    ttpos-api v0.0.0
    // ... 其他依赖
)

replace ttpos-api => ../ttpos-api
```

然后在各服务中执行：

```bash
go mod tidy
```

## 基础使用

### 导入包

```go
import (
    "ttpos-api/message"
    "ttpos-api/constant"
    "ttpos-api/util"
)
```

### 创建消息

```go
// 方式1：使用构造函数
msg := message.NewMessageSendMessage("msg-uuid-123", constant.MessageTypeEmail)

// 方式2：手动创建
msg := &message.MessageSendMessage{
    BaseMessage: message.NewBaseMessage(constant.TopicMessageSend),
    MessageUUID: "msg-uuid-123",
    MessageType: constant.MessageTypeEmail,
}

// 设置可选字段
msg.WithCompanyUUID("company-uuid-456")
msg.WithTraceID("trace-id-789")
```

## 消息发送

### 在 ttpos-bmp 中发送消息

```go
package queue

import (
    "context"
    "ttpos-api/message"
    "ttpos-api/constant"
    "ttpos-bmp/internal/pkg/queue"
)

// 发送消息到队列
func PublishMessage(ctx context.Context, messageUUID, messageType string) error {
    // 创建消息
    msg := message.NewMessageSendMessage(messageUUID, messageType)
    msg.WithCompanyUUID(getCompanyUUID(ctx))
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return err
    }
    
    // 序列化
    data, err := msg.ToJSON()
    if err != nil {
        return err
    }
    
    // 发送到队列
    return queue.Push(constant.TopicMessageSend, msg)
}
```

### 在 main 服务中发送消息

```go
package event

import (
    "context"
    "ttpos-api/message"
    "ttpos-api/constant"
)

// 发布订单创建事件
func PublishOrderCreatedEvent(ctx context.Context, orderUUID string, amount float64) error {
    // 创建消息（假设有 OrderCreatedMessage）
    msg := &message.OrderCreatedMessage{
        BaseMessage: message.NewBaseMessage(constant.TopicOrderCreated),
        OrderUUID:   orderUUID,
        OrderAmount: amount,
    }
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return err
    }
    
    // 发送到事件总线
    eventBus.Publish(msg)
    return nil
}
```

## 消息接收

### 在 ttpos-bmp 中订阅消息

```go
package consumer

import (
    "context"
    "ttpos-api/message"
    "ttpos-api/constant"
    "ttpos-bmp/internal/pkg/queue"
)

// MailgunConsumer 邮件消费者
type MailgunConsumer struct{}

// GetTopic 获取订阅的 Topic
func (c *MailgunConsumer) GetTopic() string {
    return constant.TopicMessageSend
}

// Consume 消费消息
func (c *MailgunConsumer) Consume(ctx context.Context, data []byte) error {
    // 解析消息
    var msg message.MessageSendMessage
    if err := msg.FromJSON(data); err != nil {
        return err
    }
    
    // 验证消息
    if err := msg.Validate(); err != nil {
        return err
    }
    
    // 处理消息
    return c.handleMessage(ctx, &msg)
}

// handleMessage 处理消息
func (c *MailgunConsumer) handleMessage(ctx context.Context, msg *message.MessageSendMessage) error {
    // 从数据库获取消息详情
    record := getMessageFromDB(ctx, msg.MessageUUID)
    
    // 发送邮件
    return sendEmail(ctx, record)
}
```

### 在 main 服务中订阅消息

```go
package event

import (
    "context"
    "ttpos-api/message"
    "ttpos-api/constant"
)

// 初始化事件订阅
func InitEventSubscribers() {
    // 订阅消息状态变更事件
    eventBus.Subscribe(constant.TopicMessageStatusChange, handleMessageStatusChange)
}

// 处理消息状态变更
func handleMessageStatusChange(data []byte) {
    var msg message.MessageStatusChangeMessage
    if err := msg.FromJSON(data); err != nil {
        log.Error("解析消息失败", err)
        return
    }
    
    // 处理状态变更
    log.Info("消息状态变更",
        "uuid", msg.MessageUUID,
        "old_status", msg.OldStatus,
        "new_status", msg.NewStatus,
    )
}
```

## 常量使用

### Topic 常量

```go
import "ttpos-api/constant"

// 使用预定义的 Topic
topic := constant.TopicMessageSend

// 检查 Topic 是否有效
if constant.IsValidTopic(topic) {
    // Topic 有效
}

// 获取所有 Topic
allTopics := constant.GetAllTopics()
```

### 消息类型常量

```go
import "ttpos-api/constant"

// 使用消息类型常量
messageType := constant.MessageTypeEmail

// 检查消息类型是否有效
if constant.IsValidMessageType(messageType) {
    // 消息类型有效
}
```

### 状态常量

```go
import "ttpos-api/constant"

// 使用状态常量
status := constant.MessageStatusSuccess

// 获取状态文本
statusText := constant.GetMessageStatusText(status)
// 输出: "发送成功"
```

## 工具函数

### JSON 工具

```go
import "ttpos-api/util"

// 序列化消息
jsonData, err := util.MarshalMessage(msg)

// 格式化输出（用于调试）
prettyJSON, err := util.MarshalMessagePretty(msg)

// 转换为字符串
jsonStr, err := util.ToJSONString(msg)

// 检查 JSON 是否有效
if util.IsValidJSON(jsonStr) {
    // JSON 有效
}
```

### 验证工具

```go
import "ttpos-api/util"

// 验证邮箱
if util.IsValidEmail("test@example.com") {
    // 邮箱格式正确
}

// 验证手机号
if util.IsValidPhone("13800138000") {
    // 手机号格式正确
}

// 验证 UUID
if util.IsValidUUID("550e8400-e29b-41d4-a716-446655440000") {
    // UUID 格式正确
}

// 验证消息类型
if util.IsValidMessageType("email") {
    // 消息类型有效
}
```

### 辅助工具

```go
import "ttpos-api/util"

// 获取当前时间戳
timestamp := util.GetCurrentTimestamp()

// 格式化时间戳
timeStr := util.FormatTimestamp(timestamp, "2006-01-02 15:04:05")

// 生成消息ID
messageID := util.GenerateMessageID("msg")

// 计算 MD5
hash := util.MD5Hash("hello world")

// 检查字符串是否在切片中
if util.StringInSlice("email", []string{"email", "sms"}) {
    // 字符串在切片中
}
```

## 最佳实践

### 1. 消息创建

```go
// ✅ 推荐：使用构造函数
msg := message.NewMessageSendMessage(messageUUID, messageType)
msg.WithCompanyUUID(companyUUID)

// ❌ 不推荐：手动创建所有字段
msg := &message.MessageSendMessage{
    BaseMessage: message.BaseMessage{
        MessageID: uuid.New().String(),
        Topic:     "message.send",
        Timestamp: time.Now().Unix(),
        // ...
    },
    MessageUUID: messageUUID,
    MessageType: messageType,
}
```

### 2. 消息验证

```go
// ✅ 推荐：发送前验证
msg := message.NewMessageSendMessage(messageUUID, messageType)
if err := msg.Validate(); err != nil {
    return fmt.Errorf("消息验证失败: %w", err)
}

// 发送消息
queue.Push(constant.TopicMessageSend, msg)
```

### 3. 错误处理

```go
// ✅ 推荐：完整的错误处理
msg := message.NewMessageSendMessage(messageUUID, messageType)

// 验证消息
if err := msg.Validate(); err != nil {
    log.Error("消息验证失败", "error", err)
    return err
}

// 序列化消息
data, err := msg.ToJSON()
if err != nil {
    log.Error("消息序列化失败", "error", err)
    return err
}

// 发送消息
if err := queue.Push(constant.TopicMessageSend, data); err != nil {
    log.Error("消息发送失败", "error", err)
    return err
}
```

### 4. 使用常量

```go
// ✅ 推荐：使用常量
topic := constant.TopicMessageSend
messageType := constant.MessageTypeEmail

// ❌ 不推荐：使用字符串字面量
topic := "message.send"
messageType := "email"
```

### 5. 日志记录

```go
// ✅ 推荐：记录关键信息
log.Info("发送消息",
    "message_id", msg.MessageID,
    "message_uuid", msg.MessageUUID,
    "message_type", msg.MessageType,
    "topic", msg.Topic,
)
```

### 6. 链路追踪

```go
// ✅ 推荐：传递链路追踪ID
msg := message.NewMessageSendMessage(messageUUID, messageType)
msg.WithTraceID(getTraceIDFromContext(ctx))
msg.WithCompanyUUID(getCompanyUUID(ctx))
```

## 完整示例

### 发送消息完整流程

```go
package service

import (
    "context"
    "fmt"
    "ttpos-api/message"
    "ttpos-api/constant"
    "ttpos-api/util"
)

// SendEmailMessage 发送邮件消息
func SendEmailMessage(ctx context.Context, messageUUID string) error {
    // 1. 创建消息
    msg := message.NewMessageSendMessage(messageUUID, constant.MessageTypeEmail)
    msg.WithCompanyUUID(getCompanyUUID(ctx))
    msg.WithTraceID(getTraceID(ctx))
    
    // 2. 验证消息
    if err := msg.Validate(); err != nil {
        return fmt.Errorf("消息验证失败: %w", err)
    }
    
    // 3. 序列化消息
    data, err := msg.ToJSON()
    if err != nil {
        return fmt.Errorf("消息序列化失败: %w", err)
    }
    
    // 4. 发送到队列
    if err := queue.Push(constant.TopicMessageSend, data); err != nil {
        return fmt.Errorf("消息发送失败: %w", err)
    }
    
    // 5. 记录日志
    log.Info("消息发送成功",
        "message_id", msg.MessageID,
        "message_uuid", msg.MessageUUID,
        "topic", constant.TopicMessageSend,
    )
    
    return nil
}
```

### 接收消息完整流程

```go
package consumer

import (
    "context"
    "fmt"
    "ttpos-api/message"
    "ttpos-api/constant"
)

// MessageConsumer 消息消费者
type MessageConsumer struct{}

// GetTopic 获取订阅的 Topic
func (c *MessageConsumer) GetTopic() string {
    return constant.TopicMessageSend
}

// Consume 消费消息
func (c *MessageConsumer) Consume(ctx context.Context, data []byte) error {
    // 1. 解析消息
    var msg message.MessageSendMessage
    if err := msg.FromJSON(data); err != nil {
        return fmt.Errorf("消息解析失败: %w", err)
    }
    
    // 2. 验证消息
    if err := msg.Validate(); err != nil {
        return fmt.Errorf("消息验证失败: %w", err)
    }
    
    // 3. 记录日志
    log.Info("收到消息",
        "message_id", msg.MessageID,
        "message_uuid", msg.MessageUUID,
        "message_type", msg.MessageType,
    )
    
    // 4. 处理消息
    if err := c.handleMessage(ctx, &msg); err != nil {
        return fmt.Errorf("消息处理失败: %w", err)
    }
    
    return nil
}

// handleMessage 处理消息
func (c *MessageConsumer) handleMessage(ctx context.Context, msg *message.MessageSendMessage) error {
    // 根据消息类型处理
    switch msg.MessageType {
    case constant.MessageTypeEmail:
        return c.handleEmailMessage(ctx, msg)
    case constant.MessageTypeSMS:
        return c.handleSMSMessage(ctx, msg)
    default:
        return fmt.Errorf("不支持的消息类型: %s", msg.MessageType)
    }
}
```

## 故障排查

### 常见问题

1. **消息验证失败**
   - 检查必填字段是否为空
   - 检查消息类型是否有效
   - 检查 Topic 是否正确

2. **消息序列化失败**
   - 检查消息结构体是否正确
   - 检查是否有循环引用

3. **消息发送失败**
   - 检查队列服务是否启动
   - 检查网络连接是否正常
   - 检查权限配置是否正确

### 调试技巧

```go
// 使用格式化输出查看消息内容
prettyJSON, _ := util.ToJSONStringPretty(msg)
fmt.Println("消息内容:", prettyJSON)

// 检查消息是否有效
if err := msg.Validate(); err != nil {
    fmt.Println("验证错误:", err)
}

// 检查 Topic 是否有效
if !constant.IsValidTopic(msg.Topic) {
    fmt.Println("无效的 Topic:", msg.Topic)
}
```

## 更多资源

- [README.md](README.md) - 项目概述
- [examples/](examples/) - 更多示例代码
- [TTPOS 系统文档](../README.md) - 系统整体文档

