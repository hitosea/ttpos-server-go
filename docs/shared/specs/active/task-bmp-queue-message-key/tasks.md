# task-bmp-queue-message-key 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 2 |
| 总任务数 | 7 |
| 已完成 | 6 |
| 完成率 | 86% |

---

## Phase 1: 队列基础设施

### 1.1 扩展 MqProducer 接口

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/internal/pkg/queue/queue.go` |
| Purpose | 在 MqProducer 接口中新增 SendMsgWithKey 方法定义 |
| Requirements | Requirement 1 |
| Leverage | 现有接口定义 |

**变更内容**:
```go
// MqProducer 接口新增方法
SendMsgWithKey(ctx context.Context, topic, key, body string) (mqMsg MqMsg, err error)
```

- [x] 完成

### 1.2 实现 RocketMq.SendMsgWithKey

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/internal/pkg/queue/rocketmq.go` |
| Purpose | 实现带 key 的消息发送，使用 primitive.Message.WithKeys |
| Requirements | Requirement 1 |
| Leverage | 现有 SendByteMsg 方法 |

**变更内容**:
- 新增 `SendMsgWithKey` 方法
- 使用 `msg.WithKeys([]string{key})` 设置 key
- 复用现有的参数验证、主题创建、性能监控逻辑

- [x] 完成

### 1.3 实现 RedisMq.SendMsgWithKey（兼容）

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/internal/pkg/queue/redismq.go` |
| Purpose | Redis 驱动兼容实现，忽略 key 参数 |
| Requirements | 兼容性要求 |
| Leverage | 现有 SendMsg 方法 |

**变更内容**:
```go
// Redis 不支持 key，直接调用 SendMsg
func (r *RedisMq) SendMsgWithKey(ctx context.Context, topic, key, body string) (mqMsg MqMsg, err error) {
    return r.SendMsg(ctx, topic, body)
}
```

- [x] 完成

### 1.4 新增 PushWithKey 和日志方法

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/internal/pkg/queue/producer.go`, `logger.go` |
| Purpose | 新增高层 API 和日志方法 |
| Requirements | Requirement 1 |
| Leverage | 现有 PushWithContext 方法 |

**变更内容**:
- `producer.go`: 新增 `PushWithKey` 方法
- `logger.go`: 新增 `ProducerLogWithKey` 方法

- [x] 完成

---

## Phase 2: Grab/Lineman 集成

### 2.1 Grab 订单使用 Key

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go` |
| Purpose | HandleSubmitOrder 和 HandlePushOrderState 使用订单号作为 key |
| Requirements | Requirement 2 |
| Leverage | 现有代码，只需替换方法调用 |

**变更内容**:
- `HandleSubmitOrder`: `queue.PushWithContext` → `queue.PushWithKey` + `req.GetOrderID()`
- `HandlePushOrderState`: `queue.PushWithContext` → `queue.PushWithKey` + `req.GetOrderID()`

- [x] 完成

### 2.2 Lineman 订单使用 Key

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` |
| Purpose | HandlePlaceOrder、HandleOrderUpdate、HandleOrderStatusUpdate 使用订单号作为 key |
| Requirements | Requirement 3 |
| Leverage | 现有代码，只需替换方法调用 |

**变更内容**:
- `HandlePlaceOrder`: `queue.PushWithContext` → `queue.PushWithKey` + `req.OrderId`
- `HandleOrderUpdate`: `queue.PushWithContext` → `queue.PushWithKey` + `req.OrderId`
- `HandleOrderStatusUpdate`: `queue.PushWithContext` → `queue.PushWithKey` + `req.OrderId`

- [x] 完成

---

## Phase 3: 测试与文档

### 3.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/internal/pkg/queue/producer_test.go` |
| Purpose | 测试 PushWithKey 方法 |
| Requirements | 测试要求 |

**测试用例**:
- [ ] `TestPushWithKey_Success`: 正常发送
- [ ] `TestPushWithKey_EmptyKey`: key 为空时正常发送
- [ ] `TestPushWithKey_EmptyTopic`: topic 为空返回错误
- [ ] `TestPushWithKey_NilData`: data 为 nil 返回错误

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [ ] 测试通过: `go test ./internal/pkg/queue/...`

### 功能完整性
- [x] Requirement 1: PushWithKey 方法可用
- [x] Requirement 2: Grab 订单消息包含 key
- [x] Requirement 3: Lineman 订单消息包含 key
- [x] 日志中包含 message key 信息

### 兼容性
- [x] 现有 PushWithContext 调用不受影响
- [x] Redis 驱动兼容处理

---

## 文件变更汇总

| 文件 | 变更类型 | 任务 | 状态 |
|------|----------|------|------|
| `ttpos-bmp/internal/pkg/queue/queue.go` | 修改 | 1.1 | ✅ |
| `ttpos-bmp/internal/pkg/queue/rocketmq.go` | 修改 | 1.2 | ✅ |
| `ttpos-bmp/internal/pkg/queue/redismq.go` | 修改 | 1.3 | ✅ |
| `ttpos-bmp/internal/pkg/queue/producer.go` | 修改 | 1.4 | ✅ |
| `ttpos-bmp/internal/pkg/queue/logger.go` | 修改 | 1.4 | ✅ |
| `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go` | 修改 | 2.1 | ✅ |
| `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` | 修改 | 2.2 | ✅ |
| `ttpos-bmp/internal/pkg/queue/producer_test.go` | 新建 | 3.1 | ⏳ |

---

**版本**: v1.0.0
**创建日期**: 2026-01-30
**最后更新**: 2026-01-30
