# 解耦独立订单模块 - 需求规格

| 属性 | 内容 |
| :--- | :--- |
| **Spec 名称** | story-main-decoupled-order-module |
| **模块** | Main (Go) |
| **状态** | 已通过 |
| **创建日期** | 2025-12-05 |
| **负责人** | xiezhihuan |
| **来源 Proposal** | [解耦独立订单模块](../../../team/proposals/2025-12/2025-12-05-decoupled-order-module.md) |

---

## 1. 概述

### 1.1 背景

当前 `SaleBill` (销售账单) 和 `SaleOrder` (销售订单) 模型承载了过多的业务逻辑，包括会员积分、ERP库存、打印等，形成“上帝对象”，导致维护困难、复用性差且存在测试和性能隐患。

### 1.2 目标

采用 **绞杀榕模式 (Strangler Fig Pattern)**，在不改动现有 `main/app/model` 业务实现的前提下，新建一个独立的 **Order Core Module**。

### 1.3 范围

*   **包含**:
    *   新建 `modules/order_core/` 模块。
    *   定义新的轻量级 Structs (`CoreSaleOrder`, `CoreSaleBill`)，映射现有数据库表。
    *   实现纯净的数据 CRUD 和基础状态机 (Pending -> Paid -> Finished/Canceled)。
    *   引入 Event Bus 发布领域事件 (`OrderCreated`, `OrderPaid`)。
*   **不包含**:
    *   修改现有的 `SaleOrder`/`SaleBill` 代码和逻辑。
    *   迁移现有的 POS 收银端业务逻辑（初期）。

---

## 2. 用户故事 (User Stories)

### 2.1 核心模块构建

*   **As a** 后端开发者
*   **I want** 拥有一个独立的 `order_core` 模块，仅包含订单数据的映射和基础操作
*   **So that** 我可以在新业务中快速构建订单功能，而不受旧业务逻辑的干扰。

### 2.2 状态机管理

*   **As a** 后端开发者
*   **I want** 订单状态流转受到严格的状态机控制
*   **So that** 避免出现“已取消但又支付成功”等非法状态。

### 2.3 事件驱动

*   **As a** 后端开发者
*   **I want** 在订单状态变更时收到标准化的领域事件
*   **So that** 我可以解耦地实现积分赠送、库存扣减等业务逻辑。

### 2.4 业务共存

*   **As a** 运维/测试人员
*   **I want** 新模块的操作不影响旧 POS 业务的正常运行
*   **So that** 系统可以平滑过渡，无风险上线。

---

## 3. 验收标准 (Acceptance Criteria)

### AC1: 模块独立性

*   **Given** 新建的 `order_core` 模块
*   **When** 检查其依赖关系
*   **Then** 它不应依赖 `main/app/model` 中的旧 `SaleOrder` 或 `SaleBill`。
*   **And** 它不应包含具体的业务计算逻辑（如积分规则）。

### AC2: 数据一致性

*   **Given** 一个由新模块创建的订单
*   **When** 使用旧模块查询该订单
*   **Then** 应能正确读取基础字段（如金额、状态）。

### AC3: 状态机约束

*   **Given** 一个状态为 `Paid` 的订单
*   **When** 尝试将其状态变更为 `Pending`
*   **Then** 操作应失败并返回错误。

### AC4: 事件发布

*   **Given** 订单状态从 `Pending` 变为 `Paid`
*   **When** 事务提交成功
*   **Then** 系统应发布 `OrderPaid` 事件，且订阅者能收到消息。

---

## 4. 影响分析

| 模块 | 影响描述 | 风险等级 |
| :--- | :--- | :--- |
| **Main (Go)** | 新增 `modules/order_core`，对旧代码无侵入 | 低 |
| **数据库** | 共用现有表结构，无 Schema 变更 | 中 (并发写风险) |
| **API** | 新业务接口将使用新模块 | 低 |

---

## 5. 备注

*   **技术栈**: Go (Main Module)
*   **风险缓解**: 引入乐观锁机制处理并发写；明确新旧模块的业务边界。

