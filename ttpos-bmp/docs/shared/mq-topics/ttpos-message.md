# ttpos-message MQ Topic

## 总览

- **模块**：`app/ttpos-message`
- **队列驱动**：RocketMQ（见 `app/ttpos-message/manifest/config/config.tpl.yaml`）
- **消费者注册入口**：`app/ttpos-message/internal/logic/queue/rocketmq.go` → `(*sQueue).Init()`
  - 注册：`queue.RegisterConsumer(&consumer.MailgunConsumer{})`
  - 启动：`go queue.StartConsumersListener(ctx)`

## Topic 清单

### 1) `ttpos-message-send`

- **用途**：异步触发消息发送（当前实现以邮件为主；短信预留）。
- **生产者**：`app/ttpos-message/internal/logic/queue/rocketmq.go` → `(*sQueue).PublishMessage()`
  - 发送：`queue.Push("ttpos-message-send", msg)`
- **消费者**：`app/ttpos-message/internal/logic/consumer/mailgun_consumer.go` → `MailgunConsumer`
  - 行为：
    1) 解析消息体为 `dto.RocketMQMessage`
    2) 按 `message_uuid` 从 DB 查询消息详情
    3) 更新状态为 sending
    4) 调用 Mailgun 发送
    5) 更新状态为 success/failed
- **消息体**：`app/ttpos-message/internal/model/dto/message_dto.go` → `dto.RocketMQMessage`

```json
{
  "message_uuid": "...",
  "message_type": "email"
}
```

## 备注

- 目前消费者 `MailgunConsumer` 的 `GetTopic()` 固定返回 `ttpos-message-send`；若未来拆分短信消费者，可新增 topic 或按 tag/字段分流。
