# story-pos-order-source-stat 任务分解

> 本文档定义“按订单来源拆分点餐方式统计”功能的执行任务清单。

## 📋 任务分解原则

- 任务粒度 1-4 小时，SP ≤ 1。
- 每个任务需指向 requirements.md 中的需求编号。
- 使用 `- [ ]` / `- [x]` 跟踪进度。

## 📊 进度总览

**总任务数**: 10  
**已完成**: 5  
**进行中**: 0  
**完成率**: 50%

---

## Phase 1: 数据库设计与模型

- [x] 1.1 创建统计表迁移

  - File: `admin/database/migrations/20251126010101_add_order_source_uuid_to_statistics_sale_table.php`
  - Purpose: 在 `ttpos_statistics_sale` 中新增 `order_source_uuid` 列与索引。
  - Requirements: 1.1
  - Leverage: `docs/agent/templates/database-migration-template.md`、近期统计类迁移。
  - Status: ✅ 已创建迁移文件，包含字段定义和索引

- [ ] 1.2 执行迁移并验证

  - File: -
  - Purpose: 保证新列实际存在，避免服务启动失败。
  - Requirements: 1.1
  - Command: `cd admin && php think migrate:run`
  - Success: 列与索引存在，DESC 表结果正确。

- [x] 1.3 更新 Go 模型

  - File: `main/app/model/statistics.go`
  - Purpose: 添加 `OrderSourceUuid` 字段映射，更新 `StatisticsSaleData` 结构体。
  - Requirements: 1.2
  - Leverage: 现有字段定义。
  - Status: ✅ 已添加 `OrderSourceUuid` 字段，新增外卖来源相关统计字段（`TotalInstantOrderTakeawayAmount`、`TotalInstantOrderTakeawayNum`、`AvgInstantOrderTakeawayAmount` 等）

---

## Phase 2: 核心实现（Go Main）

- [x] 2.1 SaveSale 注入 order_source_uuid

  - File: `main/app/service/statistics.go`
  - Purpose: 在 `statisticsSrv.SaveSale` 写入 `OrderSourceUuid`。
  - Requirements: 1.3
  - Leverage: `model.StatisticsSale` 构造代码块。
  - Status: ✅ 已在构造 `StatisticsSale` 时从 `saleBill.OrderSourceUuid` 赋值

- [x] 2.2 Repository 子查询补充字段

  - File: `main/app/repository/statistics.go`
  - Purpose: 在 `countSaleSubQuerySelect`/`countSaleSelect` 中新增 `order_source_uuid` 以及按来源的 `SUM(IF...)`。
  - Requirements: 2.1, 2.2
  - Leverage: 现有 `is_takeout`、`desk_uuid` 条件聚合写法。
  - Status: ✅ 已在子查询中添加 `order_source_uuid`，在聚合查询中使用 `COUNT(CASE WHEN ...)` 和 `SUM(IF(...))` 拆分店内/外卖来源统计

- [x] 2.3 CountSale/CountSaleDays DTO 更新

  - File: `main/app/service/statistics.go`
  - Purpose: 在 `CountSaleResp`、`CountSaleDaysResp` 增加字段，并在组装逻辑中赋值。
  - Requirements: 2.1, 2.2
  - Status: ✅ 已添加 `TotalInstantOrderTakeawayNum`、`TotalInstantOrderTakeawayAmount`、`MinInstantOrderTakeawayAmount`、`MaxInstantOrderTakeawayAmount`、`AvgInstantOrderTakeawayAmount` 等字段，并在 `CountSale` 和 `CountSaleDays` 中完成赋值

- [ ] 2.4 API 输出兼容性确认

  - File: `main/app/api/v1/...`（依赖统计 API）
  - Purpose: 确保新增字段在响应中存在且默认 0，不影响旧调用方。
  - Requirements: 3.1, 3.3
  - Action: 手动/自动测试接口，记录响应示例。

---

## Phase 3: 测试

- [ ] 3.1 SaveSale 单元测试

  - File: `main/app/service/statistics_test.go`
  - Purpose: 构造含不同 `order_source_uuid` 的订单，验证写入正确。
  - Requirements: 1.3

- [ ] 3.2 Repository 聚合测试

  - File: `main/app/repository/statistics_repo_test.go`
  - Purpose: 使用内存数据库模拟数据，验证拆分列数值正确且相加=总数。
  - Requirements: 2.2

- [ ] 3.3 API 集成测试

  - File: `main/app/api/v1/statistics_api_test.go`
  - Purpose: 通过 `CountSale` / `CountSaleDays` 接口验证响应含新字段。
  - Requirements: 2.1, 2.2, 3.1

---

## Phase 4: 发布与文档

- [ ] 4.1 更新发布说明 / Graphiti

  - File: `docs/shared/specs/active/story-pos-order-source-stat/design.md` + Release Note
  - Purpose: 标记“需先跑统计表迁移”，并记录经验。
  - Requirements: 3.1

- [ ] 4.2 活动日志&回滚预案

  - File: `docs/team/activities/王昱/2025-11/2025-11-25.md`
  - Purpose: 记录发布关键节点，附带回滚步骤。
  - Requirements: 3.2

---

## 提交清单

- [ ] 所有任务标记完成，tests 通过。
- [ ] `go fmt && go test ./...`。
- [ ] API 文档 / 发布说明已更新。

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/王昱/2025-11/2025-11-25.md`

---

**模板版本**: v1.0.0（定制）  
**最后更新**: 2025-11-26  
**维护者**: 后端开发组
