> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Legacy Order Adapter for Core 需求文档

> 本文档定义 Legacy Order Adapter 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/2025-12-05-implement-legacy-order-service-on-core.md](../../../../team/proposals/2025-12/2025-12-05-implement-legacy-order-service-on-core.md) |
| **创建日期**      | 2025-12-05                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint Next                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | 待定             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

本功能旨在通过 **Strangler Fig (绞杀榕)** 模式，将现有的 TTPOS 订单服务（`SaleOrder` / `SaleBill`）逻辑适配到新的 `order_core` 模块上。通过构建 `LegacyOrderAdapter` 和利用事件驱动架构，实现业务逻辑的解耦和平滑迁移，确保在不破坏现有 POS 业务的前提下，让核心订单流程运行在新的架构上。

## 🎯 产品对齐

该功能支持"重构核心订单模块，提升系统稳定性与可维护性"的技术战略目标。它允许新旧业务共存，降低了一次性重构的风险，为未来的业务扩展（如小程序点餐）打下坚实基础。

## 📝 用户故事

**作为** 后端开发者
**我想** 有一个适配层将旧的订单服务调用转发给新的核心订单模块
**以便于** 在不修改大量旧业务代码的情况下，逐步迁移到新的架构，并利用新架构的状态机和事件机制。

---

## 功能需求

### Requirement 1: Legacy Order Adapter (适配器)

**用户故事**: 作为后端开发者，我想通过 `LegacyOrderAdapter` 处理旧的订单请求，以便于底层逻辑使用的是新的 `OrderCoreService`。

#### 验收标准

1.  **WHEN** 调用 `LegacyOrderAdapter.CreateOrder` **THEN** 应调用 `OrderCoreService.CreateOrder` 并返回兼容旧接口的数据结构。
2.  **IF** `OrderCoreService` 返回错误 **THEN** Adapter 应将其转换为旧业务层能理解的错误格式。

#### 具体要求

- [ ] 1.1 创建 `modules/order_core/adapter/legacy_order_service.go`。
- [ ] 1.2 实现原有 Service 层的关键接口（Create, Pay, Cancel 等）。
- [ ] 1.3 确保输入参数（DTO）能正确映射到 `CoreSaleOrder`。
- [ ] 1.4 确保输出结果能正确映射回旧的 `SaleOrder` 结构（如有必要）。

---

### Requirement 2: Data Mapping & Sync (数据映射与同步)

**用户故事**: 作为数据库管理员，我想确保新旧模型的数据一致性，以便于报表和旧查询接口能正常工作。

#### 验收标准

1.  **WHEN** 通过 Adapter 创建订单 **THEN** 数据库中应同时存在 `sale_order` (旧表) 和 `core_sale_order` (新表，如果分离) 的数据，或者两者共用同一张表但字段兼容。
2.  **IF** 采用共用表方案 **THEN** 必须确保新模块不破坏旧模块依赖的字段（如 `sign` 等）。

#### 具体要求

- [ ] 2.1 确定数据存储策略（共用表 vs 双写 vs 视图）。建议优先考虑共用表结构，通过 Model 层的字段映射来解决。
- [ ] 2.2 实现数据转换工具函数，用于在 `SaleOrder` 和 `CoreSaleOrder` 之间进行转换。

---

### Requirement 3: Event Migration (事件驱动迁移)

**用户故事**: 作为业务开发者，我想通过监听领域事件来触发积分、打印等逻辑，以便于将被动的业务逻辑从主流程中剥离。

#### 验收标准

1.  **WHEN** `OrderCore` 发布 `OrderPaid` 事件 **THEN** 相关的 Listeners (积分、库存、打印) 应被触发执行。
2.  **IF** 某个 Listener 执行失败 **THEN** 不应影响主订单的状态流转（最终一致性），但应有重试或报警机制。

#### 具体要求

- [ ] 3.1 迁移库存预占逻辑到 `OrderCreated` 事件监听器。
- [ ] 3.2 迁移会员积分增加、实扣库存逻辑到 `OrderPaid` 事件监听器。
- [ ] 3.3 迁移小票打印触发逻辑到 `OrderPaid` 事件监听器。
- [ ] 3.4 迁移 KDS 推送逻辑到 `OrderPaid` 事件监听器。
- [ ] 3.5 迁移经营报表统计逻辑到 `OrderFinished` 事件监听器。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: Adapter 层位于 `modules/order_core/adapter`，依赖 `modules/order_core/service`。
- **依赖管理**: 严禁 Adapter 反向依赖旧的上帝 Service，应仅依赖接口。
- **遵循规范**: `.cursor/rules/go-main.mdc`

### 性能要求

- [ ] 引入适配层和事件总线后的额外延迟不超过 50ms。
- [ ] 事件处理应异步化（对于非关键路径），避免阻塞主线程。

### 测试要求

- [ ] **Adapter 层测试**: 覆盖率 100%，确保与旧 Service 行为一致。
- [ ] **集成测试**: 模拟完整下单支付流程，验证事件是否正确触发。

### 可靠性要求

- [ ] 事件处理需具备幂等性，防止重复消费。
- [ ] 关键业务（如支付后扣库存）需保证事务一致性或最终一致性。

---

## 验收标准

### 功能验收

1.  **POS 兼容性**: 使用 POS 端进行下单、支付、退款操作，流程顺畅无报错。
2.  **事件触发**: 确认下单后库存预占成功，支付后积分增加、小票打印指令发出。
3.  **数据完整**: 数据库中订单状态、金额、时间等关键字段准确无误。

### 测试验收

1.  **单元测试**: Adapter 和 Listeners 的单元测试通过。  
2.  **集成测试**: 通过自动化集成测试脚本验证标准流程。

---

## 风险和缓解

### 风险 1: 数据字段不兼容

**影响**: 高
**概率**: 中
**缓解措施**:
- 在开发初期详细比对 `SaleOrder` 和 `CoreSaleOrder` 的字段定义。
- 编写数据校验脚本，在 CI/CD 中运行。

### 风险 2: 事件丢失或延迟

**影响**: 中
**概率**: 低
**缓解措施**:
- 使用可靠的消息队列（如 RabbitMQ/RocketMQ）作为事件总线底层（如果需要跨进程）。
- 本地事件使用内存总线，但在应用关闭时需优雅处理。

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc`

### 架构文档

- `docs/shared/specs/active/story-main-decoupled-order-module/requirements.md`

---

**版本**: v1.0.0
**创建日期**: 2025-12-05
**作者**: xiezhihuan
**审核者**: 待定

