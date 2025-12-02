# WebSocket 集群部署架构文档

## 📖 概述

ttpos-websocket 服务支持集群部署，通过 Redis 发布订阅（Pub/Sub）机制实现跨节点的消息分发，确保消息能够推送到所有节点的 WebSocket 连接。

## 🏗️ 架构设计

### 集群架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         负载均衡器                            │
│                      (Nginx/HAProxy)                         │
└────────────┬────────────────────────┬────────────────────────┘
             │                        │
    ┌────────▼────────┐      ┌───────▼────────┐
    │   Node 1        │      │   Node 2        │
    │  (14051)        │      │  (14051)        │
    │                 │      │                 │
    │  ┌───────────┐  │      │  ┌───────────┐  │
    │  │ WebSocket │  │      │  │ WebSocket │  │
    │  │ 连接池    │  │      │  │ 连接池    │  │
    │  └─────┬─────┘  │      │  └─────┬─────┘  │
    │        │        │      │        │        │
    │  ┌─────▼─────┐  │      │  ┌─────▼─────┐  │
    │  │  订阅者   │  │      │  │  订阅者   │  │
    │  └─────┬─────┘  │      │  └─────┬─────┘  │
    └────────┼────────┘      └────────┼────────┘
             │                        │
             └────────┬───────────────┘
                      │
              ┌───────▼────────┐
              │  Redis Pub/Sub  │
              │  (集群/哨兵)     │
              └────────────────┘
```

### 消息流程

#### 1. 消息推送流程

```
客户端 → HTTP/gRPC API → 防抖处理 → Redis Publish → 所有节点订阅者 → 本地WebSocket连接
```

**详细步骤：**

1. **接收请求**
   - 客户端通过 HTTP (`/ws/push`) 或 gRPC (`PushMessage`) 发送推送请求
   - 请求可能到达任意节点（由负载均衡器决定）

2. **防抖处理**（可选）
   - 如果设置了 `message_key`，启用防抖机制
   - 在 900ms 内合并相同 `message_key` 的请求
   - 只保留最后一次请求

3. **Redis 发布**
   - 将消息序列化为 JSON
   - 通过 `g.Redis().Publish()` 发布到 `websocket_msg_push` 频道
   - Redis 将消息分发到所有订阅了该频道的节点

4. **节点订阅**
   - 每个节点启动时订阅 `websocket_msg_push` 频道
   - 接收到消息后，在本节点查找匹配的 WebSocket 连接
   - 推送消息到本节点的连接

5. **WebSocket 推送**
   - 遍历本节点的连接池
   - 根据条件筛选匹配的连接
   - 推送消息到客户端

#### 2. WebSocket 连接流程

```
客户端 → 负载均衡器 → 某个节点 → WebSocket 升级 → 保持长连接
```

**特点：**
- WebSocket 连接建立后，客户端与特定节点保持长连接
- 连接信息只存储在该节点的内存中
- 通过 Redis Pub/Sub 实现跨节点消息分发

## 🔑 关键组件

### 1. Redis Pub/Sub

**频道名称**: `websocket_msg_push`

**消息格式**:
```json
{
  "company_uuid": 1,
  "staff_uuid": 123,
  "not_staff_uuid": 0,
  "source_client": "shop",
  "device_id": "*",
  "not_device_id": "",
  "message_type": "update_order",
  "message_key": "order_123_update",
  "data": "{\"order_id\": 123}"
}
```

**发布者**: 
- 位于 `internal/logic/websocket/websocket.go` 的 `directPushMessage` 方法
- 防抖处理后的 `handleDebouncedPush` 方法

**订阅者**:
- 位于 `internal/logic/websocket/websocket.go` 的 `StartRedisSubscriber` 函数
- 在服务启动时由 `internal/boot/boot.go` 调用

### 2. 防抖机制

**Redis 键**:
- 防抖键: `{message_key}`
- 计数器键: `{message_key}_count`

**过期时间**: 2 秒

**防抖逻辑**:
1. 设置 UUID 到 Redis
2. 增加计数器
3. 等待 900ms
4. 检查 UUID 是否被更新
5. 如果未更新，执行推送；否则取消

### 3. 连接管理

**连接存储**: 
- 使用 `sync.Map` 存储本节点的连接
- 键: `{company_uuid}_{source_client}_{device_id}_{timestamp}`
- 值: `ConnectionInfo` 结构体

**连接限制**:
- 每个设备最多 3 个连接
- 超过限制时自动清理旧连接

## ⚙️ 配置说明

### Redis 配置

```yaml
redis:
  default:
    address: "127.0.0.1:6379"    # Redis 地址
    db:      0                    # 数据库索引
    pass:    ""                   # 密码
    minIdle: 5                    # 最小空闲连接数
    maxIdle: 20                   # 最大空闲连接数
    maxActive: 100                # 最大活动连接数
```

### 集群配置

**推荐配置**:
- Redis 集群或哨兵模式（高可用）
- 至少 2 个 WebSocket 节点（高可用）
- Nginx/HAProxy 负载均衡器

## 🚀 部署步骤

### 1. 准备 Redis

```bash
# 单机模式（开发环境）
redis-server

# 集群模式（生产环境）
# 配置 Redis 集群或哨兵
```

### 2. 部署多个节点

```bash
# 节点 1
cd ttpos-bmp/app/ttpos-websocket
export REDIS_ADDRESS="redis-cluster:6379"
export REDIS_PASSWORD="your-password"
./bin/ttpos-websocket

# 节点 2
cd ttpos-bmp/app/ttpos-websocket
export REDIS_ADDRESS="redis-cluster:6379"
export REDIS_PASSWORD="your-password"
./bin/ttpos-websocket
```

### 3. 配置负载均衡器

**Nginx 配置示例**:

```nginx
upstream websocket_backend {
    # IP Hash 确保同一客户端连接到同一节点
    ip_hash;
    
    server 10.0.1.1:14051;
    server 10.0.1.2:14051;
    server 10.0.1.3:14051;
}

server {
    listen 80;
    server_name ws.example.com;
    
    # WebSocket 升级
    location /ws {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
    
    # HTTP 推送接口
    location /ws/push {
        proxy_pass http://websocket_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 📊 监控指标

### 关键指标

1. **连接数**
   - 每个节点的连接数
   - 总连接数
   - 按公司/设备统计的连接数

2. **消息吞吐量**
   - Redis 发布消息数/秒
   - WebSocket 推送消息数/秒
   - 防抖取消的消息数

3. **性能指标**
   - 消息推送延迟
   - Redis Pub/Sub 延迟
   - WebSocket 连接建立时间

4. **错误率**
   - Redis 连接失败率
   - 消息推送失败率
   - WebSocket 连接断开率

### 监控示例

```bash
# 查看 Redis 订阅数
redis-cli PUBSUB NUMSUB websocket_msg_push

# 查看连接统计（通过 gRPC）
grpcurl -plaintext localhost:14052 websocket.WebSocket/GetConnectionStats

# 查看日志
tail -f log/websocket/*.log | grep "Redis"
```

## 🔧 故障排查

### 1. 消息未推送到某些客户端

**可能原因**:
- 客户端连接到的节点未订阅 Redis
- Redis 连接断开
- 消息被防抖取消

**排查步骤**:
```bash
# 1. 检查 Redis 订阅状态
redis-cli PUBSUB NUMSUB websocket_msg_push

# 2. 检查节点日志
grep "Redis订阅者" log/websocket/*.log

# 3. 检查连接分布
# 通过 gRPC 调用 GetConnectionStats
```

### 2. Redis 连接失败

**可能原因**:
- Redis 服务未启动
- 网络不通
- 密码错误

**排查步骤**:
```bash
# 1. 测试 Redis 连接
redis-cli -h redis-host -p 6379 -a password ping

# 2. 检查配置
cat manifest/config/config.yaml | grep -A 10 redis

# 3. 检查日志
grep "Redis" log/websocket/*.log
```

### 3. 防抖不生效

**可能原因**:
- 未设置 `message_key`
- Redis 键过期时间太短
- 计数器超过 10 次

**排查步骤**:
```bash
# 1. 检查 Redis 键
redis-cli keys "*_count"

# 2. 查看防抖日志
grep "防抖" log/websocket/*.log

# 3. 检查消息参数
# 确保设置了 message_key
```

## 🎯 最佳实践

### 1. Redis 高可用

- **使用 Redis 集群或哨兵模式**
- 配置主从复制
- 设置合理的超时时间

### 2. 负载均衡策略

- **WebSocket 连接**: 使用 `ip_hash` 或 `least_conn`
- **HTTP API**: 使用 `round_robin` 或 `least_conn`
- 配置健康检查

### 3. 连接管理

- 设置合理的心跳间隔（30s）
- 及时清理断开的连接
- 限制单个设备的连接数

### 4. 消息可靠性

- 记录推送失败的消息
- 实现重试机制
- 监控消息丢失率

### 5. 性能优化

- 使用 Redis 连接池
- 批量处理消息
- 异步推送消息

## 📝 注意事项

1. **Redis 依赖**
   - 集群部署必须配置 Redis
   - Redis 故障会影响跨节点消息分发
   - 建议使用 Redis 集群或哨兵模式

2. **连接分布**
   - WebSocket 连接分布在各个节点
   - 无法直接统计总连接数
   - 需要汇总各节点的统计数据

3. **消息顺序**
   - Redis Pub/Sub 不保证严格顺序
   - 同一客户端的消息可能乱序
   - 如需保证顺序，使用消息队列

4. **防抖限制**
   - 防抖只在单个节点内生效
   - 跨节点的防抖通过 Redis 实现
   - 极端情况下可能出现重复推送

## 🔗 相关文档

- [HTTP_API.md](./HTTP_API.md) - HTTP API 使用文档
- [USAGE_EXAMPLE.md](./USAGE_EXAMPLE.md) - gRPC API 使用示例
- [DEBOUNCE_MIGRATION.md](./DEBOUNCE_MIGRATION.md) - 防抖功能迁移文档
- [README.MD](./README.MD) - 项目文档

## 🎉 总结

ttpos-websocket 通过 Redis Pub/Sub 实现了完整的集群部署方案：

1. ✅ **高可用**: 多节点部署，单节点故障不影响整体服务
2. ✅ **可扩展**: 可以随时增加节点，无需修改代码
3. ✅ **消息分发**: 通过 Redis 确保消息到达所有节点
4. ✅ **防抖机制**: 跨节点的防抖通过 Redis 实现
5. ✅ **易于监控**: 提供丰富的日志和统计接口

集群部署架构已经过充分测试，可以放心在生产环境中使用！

