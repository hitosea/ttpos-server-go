# WebSocket 防抖功能迁移文档

## 📋 概述

本文档记录了从旧的 `websocket/api/api.go` 到新的 `ttpos-websocket` 模块的防抖功能迁移过程。

## 🎯 迁移内容

### 1. 核心功能

从原有的 HTTP API 防抖机制迁移到 gRPC 服务，保留了以下核心特性：

- **防抖机制**：900ms 防抖延迟
- **限流保护**：基于 MessageKey 的细粒度锁
- **计数器**：防止频繁推送（10次阈值）
- **Redis 缓存**：用于跨实例的防抖协调

### 2. 架构变化

#### 旧架构（websocket/api/api.go）
```
HTTP POST /push_client
  ↓
防抖处理（goroutine）
  ↓
Redis Publish
  ↓
WebSocket 推送
```

#### 新架构（ttpos-websocket）
```
gRPC PushMessage
  ↓
防抖处理（goroutine）
  ↓
直接 WebSocket 推送
```

## 📝 修改清单

### 1. Protobuf 定义更新

**文件**: `manifest/protobuf/websocket/websocket.proto`

```protobuf
message PushMessageReq {
  uint64 company_uuid = 1;
  uint64 staff_uuid = 2;
  uint64 not_staff_uuid = 3;
  string source_client = 4;
  string device_id = 5;
  string not_device_id = 6;
  string message_type = 7;
  string message_key = 8;        // 新增：消息键，用于防抖去重
  string data = 9;
}
```

### 2. DTO 更新

**文件**: `internal/model/dto/websocket.go`

```go
type PushMessageInput struct {
	CompanyUuid  uint64 `json:"company_uuid"`
	StaffUuid    uint64 `json:"staff_uuid"`
	NotStaffUuid uint64 `json:"not_staff_uuid"`
	SourceClient string `json:"source_client"`
	DeviceId     string `json:"device_id"`
	NotDeviceId  string `json:"not_device_id"`
	MessageType  string `json:"message_type"`
	MessageKey   string `json:"message_key"`    // 新增
	Data         string `json:"data"`
}
```

### 3. 业务逻辑实现

**文件**: `internal/logic/websocket/websocket.go`

#### 新增函数：

1. **getMessageKeyMutex**: 获取 MessageKey 专用锁
2. **handleDebouncedPush**: 处理防抖推送逻辑
3. **directPushMessage**: 直接推送消息（无防抖）

#### 修改函数：

- **PushMessage**: 增加防抖判断逻辑

### 4. 控制器更新

**文件**: `internal/controller/rpc/websocket/websocket.go`

在构建输入参数时添加 `MessageKey` 字段传递。

## 🔧 技术细节

### 防抖算法

```go
1. 接收推送请求，生成唯一 UUID
2. 将 UUID 存入 Redis（key: MessageKey, value: UUID, expire: 2s）
3. 增加计数器（key: MessageKey_count, expire: 2s）
4. 等待 900ms
5. 检查 Redis 中的 UUID 是否被更新
   - 如果被更新：取消推送（有新请求）
   - 如果未更新：执行推送
6. 清理计数器
```

### 限流机制

- 使用 `sync.Map` 存储每个 MessageKey 的专用锁
- 双重检查锁模式避免并发创建
- 计数器超过 10 次后自动推送，不再检查防抖

### Redis 配置

配置文件已包含 Redis 配置：

```yaml
redis:
  default:
    address: "$REDIS_ADDRESS"
    db:      $REDIS_DB
    pass:    "$REDIS_PASSWORD"
    # ... 其他配置
```

## 📊 性能对比

| 特性 | 旧实现 | 新实现 |
|------|--------|--------|
| 通信协议 | HTTP + Redis Pub/Sub | gRPC 直连 |
| 防抖延迟 | 900ms | 900ms |
| 并发控制 | 信号量（500） | 细粒度锁 |
| 跨实例支持 | ✅ Redis | ✅ Redis |
| 响应速度 | 异步返回 | 异步返回 |

## 🚀 使用示例

### gRPC 调用示例

```go
// 带防抖的推送
req := &v1.PushMessageReq{
    CompanyUuid:  1,
    MessageType:  "update_order",
    MessageKey:   "order_123_update",  // 设置 MessageKey 启用防抖
    Data:         `{"order_id": 123}`,
}

resp, err := client.PushMessage(ctx, req)
```

### 不带防抖的推送

```go
// 不设置 MessageKey，直接推送
req := &v1.PushMessageReq{
    CompanyUuid:  1,
    MessageType:  "customer_call",
    Data:         `{"table": "A01"}`,
    // MessageKey 为空，不启用防抖
}

resp, err := client.PushMessage(ctx, req)
```

## ✅ 测试验证

### 1. 功能测试

- [x] MessageKey 为空时直接推送
- [x] MessageKey 不为空时启用防抖
- [x] 900ms 内多次推送只执行最后一次
- [x] 计数器超过 10 次后自动推送
- [x] Redis 连接失败时的降级处理

### 2. 性能测试

- [x] 并发 100 个相同 MessageKey 的请求
- [x] 验证最终只推送一次
- [x] 检查 Redis 键的正确创建和过期

### 3. 集成测试

- [x] 与现有 WebSocket 连接管理集成
- [x] 消息推送到客户端验证
- [x] 日志输出验证

## 📌 注意事项

1. **Redis 依赖**：防抖功能依赖 Redis，确保 Redis 配置正确
2. **MessageKey 规范**：建议使用 `{业务类型}_{业务ID}_{操作}` 格式
3. **过期时间**：Redis 键过期时间设置为 2 秒，确保不会长期占用内存
4. **并发安全**：使用细粒度锁保证同一 MessageKey 的操作串行化
5. **异步处理**：防抖逻辑在 goroutine 中异步执行，不阻塞主流程

## 🔄 迁移步骤

1. ✅ 更新 Protobuf 定义添加 `message_key` 字段
2. ✅ 重新生成 Protobuf 代码
3. ✅ 更新 DTO 添加 `MessageKey` 字段
4. ✅ 实现防抖逻辑
5. ✅ 更新控制器传递参数
6. ✅ 配置 Redis 连接
7. ✅ 测试验证功能

## 📚 相关文件

- `manifest/protobuf/websocket/websocket.proto` - Protobuf 定义
- `internal/model/dto/websocket.go` - 数据传输对象
- `internal/logic/websocket/websocket.go` - 业务逻辑实现
- `internal/controller/rpc/websocket/websocket.go` - gRPC 控制器
- `manifest/config/config.tpl.yaml` - 配置模板

## 🎉 总结

防抖功能已成功从旧的 HTTP API 迁移到新的 gRPC 服务架构，保留了原有的核心特性，并进行了以下优化：

1. **更好的性能**：gRPC 直连，减少了 Redis Pub/Sub 的中间环节
2. **更清晰的架构**：业务逻辑集中在 logic 层
3. **更好的可维护性**：符合 GoFrame 框架规范
4. **向后兼容**：不影响不使用 MessageKey 的现有调用

## 🔗 参考资料

- [GoFrame Redis 文档](https://goframe.org/docs/components/cache-redis)
- [gRPC Go 文档](https://grpc.io/docs/languages/go/)
- [防抖和节流原理](https://css-tricks.com/debouncing-throttling-explained-examples/)

