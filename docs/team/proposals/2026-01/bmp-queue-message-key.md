# BMP 队列消息 Key 支持 需求提案

## 📋 提案信息

| 项目          | 内容                |
| ------------- | ------------------- |
| **提案人**    | rikugun             |
| **日期**      | 2026-01-30          |
| **目标版本**  | v2.14.0             |
| **状态**      | 待评审              |
| **关联 Spec** | -                   |

---

## 🎯 背景和动机

### 问题描述

当前 `queue.PushWithContext` 方法发送 RocketMQ 消息时不支持设置 message key，导致以下问题：

1. **消息追踪困难**：无法通过订单号快速查找某个订单的所有相关消息
2. **重复消费难排查**：当消息重复消费时，难以快速定位问题根源
3. **日志关联性差**：生产端和消费端的日志缺乏统一的关联标识

在外卖订单场景中（Grab、Lineman），同一订单可能产生多条消息（创建、状态更新等），没有 key 会增加问题排查的复杂度。

### 业务价值

- **提升运维效率**：通过订单号快速定位相关消息，减少排查时间
- **改善可观测性**：完善消息链路追踪，便于监控和告警
- **降低故障恢复时间**：问题发生时能快速定位根因

### 目标用户

- [x] 后端开发者
- [x] 运维团队
- [x] 技术支持

---

## 💡 解决方案概述

### 方案描述

新增 `queue.PushWithKey(ctx, topic, key, data)` 方法，支持在发送消息时指定 message key。在外卖订单处理场景中，使用平台订单号（如 Grab/Lineman 的 orderId）作为 key，确保同一订单的消息可追踪。

### 核心功能点

1. **新增 `PushWithKey` 方法**：在 `ttpos-bmp/internal/pkg/queue/producer.go` 中新增支持 key 的推送方法
2. **Grab 订单处理支持 key**：`HandleSubmitOrder` 和 `HandlePushOrderState` 使用 `provider_order_id` 作为 key
3. **Lineman 订单处理支持 key**：`HandlePlaceOrder` 和 `HandleOrderUpdate` 复用相同逻辑，使用 `orderId` 作为 key

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [ ] Kiosk 自助点餐机
- [x] 无直接终端影响（后端基础设施改进）

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [x] 其他: 消息队列基础设施

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整,无业务逻辑变更
- [x] **中**：需要前后端联调,基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预估 SP**: 2（待技术评审确认）

### 拆分预估

**是否需要拆分**：
- [x] **否**：单模块，SP ≤ 5，可直接创建 1 个 Spec
- [ ] **是**：需要拆分为多个 Spec

**预估 Spec 数量**：1 个

**预估 Spec 列表**：
1. `tech-bmp-queue-message-key` - 队列消息 Key 支持

### 风险识别

**潜在风险**：
1. 底层 RocketMQ 客户端 `SendMsg` 方法是否支持 key 参数
2. 现有调用方需要逐步迁移到新方法

**缓解措施**：
1. 新增方法不修改原有 API，保持向后兼容
2. 优先在外卖订单场景试点，验证后再推广

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名   | 签名/日期 |
| ---------- | ------ | --------- |
| 产品经理   |        |           |
| 技术负责人 |        |           |
| 开发代表   |        |           |
| 测试代表   |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待评审]
```

**下一步行动**：

- [ ] 创建 Spec：`tech-bmp-queue-message-key`
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### User Story（初稿）

**作为** 后端开发者/运维人员
**我想** 在队列消息中添加 message key（使用平台订单号）
**以便于** 快速追踪和定位订单相关的所有消息

### AC 验收标准（初稿）

1. **WHEN** 调用 `queue.PushWithKey(ctx, topic, key, data)` **THEN** 系统 **SHALL** 发送带有指定 key 的 RocketMQ 消息
2. **WHEN** Grab 订单创建/状态更新时 **THEN** 系统 **SHALL** 使用 `provider_order_id` 作为 message key
3. **WHEN** Lineman 订单创建/更新时 **THEN** 系统 **SHALL** 使用 `orderId` 作为 message key
4. **WHEN** 在 RocketMQ 控制台查询消息时 **THEN** 运维人员 **SHALL** 能通过订单号（key）查找相关消息

### 代码变更预览

```go
// ttpos-bmp/internal/pkg/queue/producer.go

// PushWithKey 使用指定 Context 和 Key 推送队列消息
// key 用于消息追踪和顺序消费保证（相同 key 的消息会路由到同一队列）
func PushWithKey(ctx context.Context, topic, key string, data interface{}) error {
    // TODO: 实现
}
```

```go
// ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go
// HandleSubmitOrder 中的消息发送改为：
if err := queue.PushWithKey(ctx, TopicGrabOrder, req.GetOrderID(), event); err != nil {
    // ...
}
```

---

**版本**: v1.0.0
