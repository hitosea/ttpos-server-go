# 在 Order Core 上实现原有 TTPOS 订单服务 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-05   |
| **目标版本** | v2.11.0 |
| **状态**   | 已通过   |
| **关联任务** | - |
| **关联 Spec** | [story-main-legacy-order-adapter](../../../shared/specs/active/story-main-legacy-order-adapter/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

我们已经确立了 [story-main-decoupled-order-module](../../shared/specs/active/story-main-decoupled-order-module/requirements.md) (Order Core) 作为未来订单处理的核心模块。然而，现有的 TTPOS 业务（包括 POS 收银、会员积分、库存扣减等）仍然依赖于庞大且耦合严重的 `SaleOrder` 和 `SaleBill` 旧模型。

目前面临的问题是：如何将现有的复杂业务逻辑平滑迁移到新的 `order_core` 架构上，既能利用新架构的纯净性和稳定性，又不破坏现有的业务流程。直接重写风险过大，需要一种渐进式的适配方案。

### 业务价值

1.  **降低维护成本**：将核心订单流转与周边业务（积分、打印、库存）解耦。
2.  **提高稳定性**：核心订单状态由严格的状态机控制，减少因业务逻辑复杂导致的脏数据。
3.  **平滑过渡**：通过适配层支持现有业务，无需一次性重构所有终端代码。
4.  **复用性**：新业务（如小程序点餐）可以直接使用 Core，旧业务（POS）可以通过适配层使用 Core。

### 目标用户

- [ ] 收银员 (间接受益于系统稳定性)
- [x] 后端开发者 (直接受益于架构清晰)
- [ ] 商户管理员
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

采用 **适配器模式 (Adapter Pattern)** 和 **事件驱动架构 (EDA)** 相结合的方式。

1.  保留原有的 Service 接口层，但在底层将其改造为调用 `order_core` 模块。
2.  将原 `SaleOrder` 模型中的"上帝逻辑"（如积分计算、库存扣减、打印触发）剥离为 **领域事件监听器 (Domain Event Listeners)**。
3.  构建 `LegacyOrderAdapter`，作为旧业务逻辑与新 `order_core` 之间的桥梁。

### 核心功能点

1.  **Legacy Adapter**: 创建 `modules/order_core/adapter/legacy_order_service.go`，实现原有业务对订单的操作接口，内部调用 `OrderCoreService`。
2.  **Data Sync / Mapping**: 确保旧 `SaleOrder` 模型的数据结构能正确读写底层表（新旧模块共用 DB 表，或通过 DTO 转换）。
3.  **Event Migration**:
    *   `OrderCreated` 事件 -> 触发：库存预占、生成取餐号。
    *   `OrderPaid` 事件 -> 触发：会员积分增加、实扣库存、打印小票、推送 KDS。
    *   `OrderFinished` 事件 -> 触发：经营报表统计。

### 影响范围

**涉及终端**：
- [x] POS 收银端 (核心影响)
- [x] Mobile 扫码端
- [x] KDS 厨显端 (通过事件触发)

**涉及模块**：
- [ ] UI 组件
- [x] API 接口 (需要适配)
- [x] 数据模型 (共用表结构)
- [x] 业务逻辑 (核心重构)
- [ ] 第三方集成

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [x] **高**：涉及架构调整、第三方集成、复杂算法 (涉及核心订单流转的迁移，风险较高)

### 工作量预估

- **预计天数**: 10 天
- **预估 SP**: 21（待技术评审确认）

### 风险识别

**潜在风险**：
1.  **数据不一致**：新旧模型共用表可能导致字段读写冲突。
2.  **漏发事件**：某些隐蔽的业务逻辑（如特定促销规则）可能在迁移中被遗漏。
3.  **性能回退**：引入 Event Bus 和适配层可能增加请求延迟。

**缓解措施**：
1.  **双写/灰度验证**：在开发环境中并行运行，比对数据。
2.  **单元测试覆盖**：为 `LegacyOrderAdapter` 编写高覆盖率的测试，确保行为与旧 Service 一致。
3.  **异步优化**：非核心业务（打印、积分）严格走异步消息队列。

---

## 🔗 相关资源

### 参考需求

- 基础 Spec: [story-main-decoupled-order-module](../../shared/specs/active/story-main-decoupled-order-module/requirements.md)

### 相关文档

- 架构设计: Strangler Fig Pattern

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**下一步行动**：

- [x] 创建 Spec：`story-main-legacy-order-adapter`
- [ ] 分配负责人：xiezhihuan
- [ ] 目标 Sprint：Sprint Next

---

