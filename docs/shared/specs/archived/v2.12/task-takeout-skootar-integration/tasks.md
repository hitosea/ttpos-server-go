# 整合 Skootar 订单逻辑到现有订单模型 任务分解

> 本文档定义 整合 Skootar 订单逻辑到现有订单模型 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 0  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件 (Schema & Data)

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/sql/20251205_migrate_skootar_structure.up.sql`
  - Purpose: 创建 `takeout_order_skootar` 表，并编写 SQL 语句将 `takeout_job` 数据迁移到 `takeout_order` 和 `takeout_order_skootar`。
  - Requirements: 1.1, 1.2
  - Prompt: Role: Database Engineer | Task: Create SQL migration script for Skootar integration. 1. Create table `takeout_order_skootar` (cols: id, uuid, order_uuid, skootar_id, skootar_name, skootar_phone, skootar_rating, skootar_image_url, time cols). 2. Write INSERT SELECT statements to migrate data from `takeout_job` to `takeout_order` (mapping fields) and `takeout_order_skootar`. | Context: `takeout_job` is the old table, `takeout_order` is the new main table.
  - Restrictions: Ensure idempotency if possible.

- [x] 1.2 执行数据库迁移与代码生成

  - File: -
  - Purpose: 应用数据库变更并生成 DAO/Entity。
  - Requirements: 1.1
  - Command: `make db_migrate` (or manual run), then `gf gen dao`
  - Success: `internal/dao/order_skootar.go` and entities created.

---

## Phase 2: 核心业务逻辑适配

- [x] 2.1 修改 Skootar CreateOrder 逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/create_order.go`
  - Purpose: 适配新的双表写入逻辑。
  - Requirements: 2.1, 2.2
  - Leverage: `dao.Order`, `dao.OrderSkootar`
  - Prompt: Role: Go Developer | Task: Refactor CreateOrder to write to `takeout_order` and `takeout_order_skootar` tables within a transaction. | Context: Replace `dao.Job` usage. Map response fields correctly.

- [x] 2.2 修改 Skootar JobStatusChange 逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/job_status.go`
  - Purpose: 更新状态时同步更新主表状态，如有司机信息更新则更新扩展表。
  - Requirements: 2.1
  - Context: Skootar callback may contain driver info updates.

- [x] 2.3 修改查询与 GetDriverInfo 逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/skootar.go` (or related file)
  - Purpose: 从新表结构读取订单和司机信息。
  - Requirements: 3.1
  - Leverage: `dao.Order`, `dao.OrderSkootar`
  - Note: This might involve updating `GetDriverInfo` implementation to query `takeout_order_skootar` by `order_uuid`.

---

## Phase 3: 兼容性与清理

- [x] 3.1 验证 API 兼容性 (Manual/Test)

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/takeout/takeout.go`
  - Purpose: 确保 Controller 层调用新的 Logic 后，返回的 DTO 结构数据填充正确。
  - Requirements: 3.2
  - Action: Review controller code, ensure it maps fields from the new return types correctly.

- [ ] 3.2 清理旧代码 (Optional/Deprecate)

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/entity/job.go`
  - Purpose: 标记旧 `Job` 相关代码为废弃，或在确认无引用后删除（建议先保留几个版本）。
  - Requirements: Non-functional

---

## 提交清单

- [ ] 数据库迁移成功且数据准确
- [ ] Skootar 下单流程在新模型下跑通
- [ ] 司机信息获取正常
- [ ] 回调状态更新正常

