# BMP 队列消息 Key 支持 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | task-bmp-queue-message-key                                           |
| **来源 Proposal** | [bmp-queue-message-key](../../../team/proposals/2026-01/bmp-queue-message-key.md) |
| **创建日期**      | 2026-01-30                                                           |
| **负责人**        | rikugun                                                              |
| **目标 Sprint**   | -                                                                    |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发中     |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-30 |

---

## 📝 用户故事

**作为** 后端开发者/运维人员
**我想** 在队列消息中添加 message key（使用平台订单号）
**以便于** 快速追踪和定位订单相关的所有消息，提升问题排查效率

---

## 功能需求

### Requirement 1: 新增 PushWithKey 方法

**用户故事**: 作为后端开发者，我想使用带 key 的消息推送方法，以便于消息可追踪

#### 验收标准

1. **WHEN** 调用 `queue.PushWithKey(ctx, topic, key, data)` **THEN** 系统 **SHALL** 发送带有指定 key 的 RocketMQ 消息
2. **IF** key 参数为空字符串 **THEN** 系统 **SHALL** 正常发送消息（key 可选）
3. **WHEN** 消息发送成功 **THEN** 系统 **SHALL** 在日志中记录 topic、key、msgId

### Requirement 2: Grab 订单消息支持 Key

**用户故事**: 作为运维人员，我想通过 Grab 订单号查找相关消息，以便于快速定位问题

#### 验收标准

1. **WHEN** `HandleSubmitOrder` 发送 MQ 消息 **THEN** 系统 **SHALL** 使用 `provider_order_id` 作为 message key
2. **WHEN** `HandlePushOrderState` 发送 MQ 消息 **THEN** 系统 **SHALL** 使用 `provider_order_id` 作为 message key
3. **WHEN** 在 RocketMQ 控制台查询消息 **THEN** 运维人员 **SHALL** 能通过订单号（key）查找相关消息

### Requirement 3: Lineman 订单消息支持 Key

**用户故事**: 作为运维人员，我想通过 Lineman 订单号查找相关消息，以便于快速定位问题

#### 验收标准

1. **WHEN** `HandlePlaceOrder` 发送 MQ 消息 **THEN** 系统 **SHALL** 使用 `orderId` 作为 message key
2. **WHEN** `HandleOrderUpdate` 发送 MQ 消息 **THEN** 系统 **SHALL** 使用 `orderId` 作为 message key
3. **WHEN** `HandleOrderStatusUpdate` 发送 MQ 消息 **THEN** 系统 **SHALL** 使用 `orderId` 作为 message key

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖 `PushWithKey` 方法
- [ ] 集成测试验证消息 key 正确设置

### 兼容性要求

- [ ] 新方法不影响现有 `PushWithContext` 调用
- [ ] 向后兼容，不修改现有 API 签名

### 可观测性要求

- [ ] 日志中包含 message key 信息
- [ ] 支持在 RocketMQ 控制台通过 key 查询消息

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 分层架构: Controller → Logic → DAO
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/go-rules.mdc 规范
- 使用 `gerror` 处理错误（不用标准库 errors）

### 资源约束

- Story Point: 2

---

## 风险和缓解

### 风险 1: 底层 RocketMQ 客户端是否支持 key

**影响**: 中
**缓解措施**: 查阅 RocketMQ Go 客户端文档，确认 `SendMsg` 方法是否支持 key 参数；如不支持需扩展底层实现

### 风险 2: 现有调用方迁移成本

**影响**: 低
**缓解措施**: 新增方法不修改原有 API，保持向后兼容；优先在外卖订单场景试点

---

## 涉及文件（预估）

| 文件路径 | 变更类型 | 说明 |
|----------|----------|------|
| `ttpos-bmp/internal/pkg/queue/producer.go` | 修改 | 新增 `PushWithKey` 方法 |
| `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go` | 修改 | Grab 订单使用 key |
| `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` | 修改 | Lineman 订单使用 key |

---

**版本**: v1.0.0
**创建日期**: 2026-01-30
