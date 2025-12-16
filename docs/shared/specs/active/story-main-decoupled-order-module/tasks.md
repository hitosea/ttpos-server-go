# 解耦独立订单模块 - 任务分解

> 本文档定义解耦独立订单模块 (Order Core Module) 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时
- **独立性**: 任务之间尽量解耦
- **可测试**: 每个任务完成后应包含测试验证

## 📊 进度总览

**总任务数**: 9
**已完成**: 9
**进行中**: -
**完成率**: 100%

---

## Phase 1: 基础构建

- [x] 1.1 创建模块目录结构
  - File: `main/app/modules/order_core/`
  - Purpose: 建立模块基础目录
  - Requirements: AC1
  - Prompt: Role: System Architect | Task: 创建 main/app/modules/order_core 及其子目录 (api, model, repository, service, event, dto) | Success: 目录结构创建完成

- [x] 1.2 定义 Core Models
  - File: `main/app/modules/order_core/model/*.go`
  - Purpose: 创建 `CoreSaleBill`, `CoreSaleOrder`, `CoreSaleOrderProduct` 结构体，映射现有数据库表
  - Requirements: AC1, AC2
  - Leverage: `main/app/model/sale_order.go`, `main/app/model/sale_bill.go` (仅参考字段定义)
  - Prompt: Role: Go Developer | Task: 定义 Core Models，复用 ttpos_sale_* 表名，仅包含 id, uuid, status, amount, create_time 等核心字段，不包含业务计算方法 | Success: Models 定义完成且能正确映射 DB

- [x] 1.3 定义 Repository 接口
  - File: `main/app/modules/order_core/repository/i_core_order_repo.go`
  - Purpose: 定义数据访问接口
  - Requirements: AC1
  - Prompt: Role: Go Developer | Task: 定义 ICoreOrderRepo 接口，包含 CreateBill, GetBill, UpdateBillStatus 等基础方法 | Success: 接口定义清晰

- [x] 1.4 实现 Repository
  - File: `main/app/modules/order_core/repository/core_order_repo.go`
  - Purpose: 实现 Repository 接口，使用 GORM 操作数据
  - Requirements: AC1, AC2
  - Prompt: Role: Go Developer | Task: 实现 CoreOrderRepo，使用 GORM 操作数据库 | Success: CRUD 功能实现，并通过单元测试

---

## Phase 2: 核心逻辑

- [x] 2.1 定义 DTOs
  - File: `main/app/modules/order_core/dto/*.go`
  - Purpose: 定义 Service 层的输入输出结构
  - Requirements: AC1
  - Prompt: Role: Go Developer | Task: 定义 CreateOrderReq, CreateOrderResp 等 DTO | Success: DTO 定义完成

- [x] 2.2 定义领域事件
  - File: `main/app/modules/order_core/event/events.go`
  - Purpose: 定义 OrderCreated, OrderPaid 事件结构
  - Requirements: AC4
  - Prompt: Role: Go Developer | Task: 定义 Order 相关的 Event 结构体 | Success: 事件结构清晰

- [x] 2.3 实现 Service 接口与骨架
  - File: `main/app/modules/order_core/service/core_order_service.go`
  - Purpose: 实现 ICoreOrderService 接口
  - Requirements: AC1, AC3
  - Prompt: Role: Go Developer | Task: 实现 CoreOrderService，注入 Repository 和 EventBus | Success: Service 结构体创建完成

- [x] 2.4 实现状态机逻辑
  - File: `main/app/modules/order_core/service/core_order_service.go`
  - Purpose: 在 UpdateStatus 方法中实现状态机检查
  - Requirements: AC3
  - Prompt: Role: Go Developer | Task: 实现严格的状态流转检查 (Pending->Paid->Finish)，非法状态返回错误 | Success: 状态机逻辑覆盖所有路径

---

## Phase 3: 测试验证

- [x] 3.1 编写 Repository 测试
  - File: `main/app/modules/order_core/repository/core_order_repo_test.go`
  - Purpose: 验证数据库映射和读写
  - Requirements: AC2
  - Prompt: Role: QA Engineer | Task: 编写 Repository 单元测试，验证读写现有表数据 | Success: 测试通过

- [x] 3.2 编写 Service 状态机测试
  - File: `main/app/modules/order_core/service/core_order_service_test.go`
  - Purpose: 验证状态流转逻辑
  - Requirements: AC3
  - Prompt: Role: QA Engineer | Task: 编写 Service 单元测试，重点测试非法状态变更是否报错 | Success: 测试覆盖所有状态跳转路径

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

