# 通用 MQ Topic

## 1) `ttpos-ping`

- **用途**：RocketMQ 生产者连通性探测（启动生产者时发送一条同步消息用于验证 NameServer/ACL 配置）。
- **生产者**：`ttpos-bmp/internal/pkg/queue/rocketmq.go` → `RegisterRocketMqProducer()`
  - `SendSync(..., topic="ttpos-ping", body="1")`
- **消费者**：无（仅用于探测）。

> 运维提示：该 topic 可能在首次启动时被自动创建（取决于 `queue.rocketmq.brokerAddr` 是否配置）。
