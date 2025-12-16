# Legacy Order Adapter 任务分解

> 本文档定义 Legacy Order Adapter 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8
**已完成**: 0
**完成率**: 0%

---

## Phase 1: Adapter 基础框架

- [ ] 1.1 定义 Adapter 接口与 DTO 转换
  - File: `main/app/modules/order_core/adapter/i_legacy_order_service.go`, `main/app/modules/order_core/adapter/converter.go`
  - Purpose: 定义适配器接口标准，并实现新旧 DTO 的互转
  - Requirements: 1.1, 1.3, 1.4
  - Prompt: Role: Go Developer | Task: 定义 ILegacyOrderAdapter 接口和 converter.go | Context: 接口包含 CreateOrder, PayOrder, CancelOrder；Converter 实现 old dto -> core dto 和 core model -> old dto 的转换 | Restrictions: 遵循 design.md

- [ ] 1.2 实现 LegacyOrderAdapter (CreateOrder)
  - File: `main/app/modules/order_core/adapter/legacy_order_service.go`
  - Purpose: 实现创建订单的适配逻辑
  - Requirements: 1.1, 1.2
  - Leverage: `main/app/modules/order_core/service/core_order_service.go`
  - Prompt: Role: Go Developer | Task: 实现 LegacyOrderAdapter.CreateOrder | Context: 调用 converter 转换参数，调用 CoreOrderService.CreateOrder，处理错误映射 | Restrictions: 错误码需兼容旧系统

- [ ] 1.3 实现 LegacyOrderAdapter (PayOrder & CancelOrder)
  - File: `main/app/modules/order_core/adapter/legacy_order_service.go`
  - Purpose: 实现支付和取消订单的适配逻辑
  - Requirements: 1.1, 1.2
  - Prompt: Role: Go Developer | Task: 实现 LegacyOrderAdapter 的 PayOrder 和 CancelOrder | Context: 类似 CreateOrder，调用 Core 对应方法

- [ ] 1.4 编写 Adapter 单元测试
  - File: `main/app/modules/order_core/adapter/legacy_order_service_test.go`
  - Purpose: 确保适配逻辑正确
  - Requirements: 测试验收 1
  - Prompt: Role: QA Engineer | Task: 测试 Adapter 的参数转换和 Mock 调用 | Context: Mock CoreOrderService

---

## Phase 2: 事件监听器迁移

- [ ] 2.1 实现 InventoryListener (库存)
  - File: `main/app/modules/order_core/listener/inventory_listener.go`
  - Purpose: 监听订单事件处理库存
  - Requirements: 3.1, 3.2
  - Leverage: 旧的库存 Service
  - Prompt: Role: Go Developer | Task: 实现库存监听器 | Context: 监听 OrderCreated (预占) 和 OrderPaid (实扣) | Restrictions: 异步处理，注意幂等

- [ ] 2.2 实现 MemberListener (积分)
  - File: `main/app/modules/order_core/listener/member_listener.go`
  - Purpose: 监听订单支付事件增加积分
  - Requirements: 3.2
  - Leverage: 旧的会员 Service
  - Prompt: Role: Go Developer | Task: 实现会员监听器 | Context: 监听 OrderPaid，调用 MemberService 增加积分

- [ ] 2.3 实现 PrinterListener (打印)
  - File: `main/app/modules/order_core/listener/printer_listener.go`
  - Purpose: 监听订单支付事件触发打印
  - Requirements: 3.3
  - Leverage: 旧的打印 Service
  - Prompt: Role: Go Developer | Task: 实现打印监听器 | Context: 监听 OrderPaid，构造打印任务

---

## Phase 3: 集成验证

- [ ] 3.1 编写集成测试
  - File: `test/integration/legacy_adapter_test.go`
  - Purpose: 验证全流程
  - Requirements: 测试验收 2
  - Prompt: Role: QA Engineer | Task: 编写集成测试 | Context: 模拟下单 -> 支付流程，断言 DB 数据和 Event 触发结果

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`

