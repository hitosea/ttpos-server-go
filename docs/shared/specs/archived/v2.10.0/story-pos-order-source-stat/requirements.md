> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# story-pos-order-source-stat 需求文档

> 本文档定义 POS 端“按订单来源拆分点餐方式统计”的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/order-source-split-statistics.md](../../../../team/proposals/2025-11/order-source-split-statistics.md) |
| **创建日期**      | 2025-11-25                                                                                                                          |
| **负责人**        | 待定                                                                                                                                |
| **目标 Sprint**   | Sprint 25                                                                                                                           |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                          |

---

## 📋 概述

旧商家后台统计报表、收银交班详情以及收银机营业数据目前仅按“点餐方式”统计（桌台/即时/外卖），无法区分订单来源（店内、扫码、第三方外卖等）。本项目将 `order_source_uuid` 贯穿至统计链路，在统计数据中新增来源维度，使即时订单能够拆分为店内与外卖来源，满足门店交班、渠道分析与稽核对账需要。

## 🎯 产品对齐

该功能支撑“全渠道经营洞察”，帮助门店管理者识别不同渠道（店内、扫码、自营、第三方外卖）的即时订单贡献度，并在交班与报表中快速对账，降低财务风险。

## 📝 用户故事

**作为** 门店店长/收银主管  
**我想** 在统计报表、交班详情与收银机营业数据中直接看到按订单来源拆分的点餐方式指标  
**以便于** 及时发现渠道异常、优化投放及完成财务核对

---

## 功能需求

### Requirement 1: 统计数据落库需保存订单来源

**用户故事**: 作为 数据分析系统，我想在统计表中保留订单来源 UUID，以便后续按来源聚合。

#### 验收标准

1. **WHEN** `statisticsSrv.SaveSale` 写入 `ttpos_statistics_sale` **THEN** 同步写入 `order_source_uuid`，并与 `ttpos_sale_bill.order_source_uuid` 保持一致。
2. **IF** `order_source_uuid` 为空或不存在 **THEN** 采用默认值 0（表示店内），保证字段非空。
3. **WHEN** 历史统计数据缺少该字段 **THEN** 迁移脚本需将新增列默认设置为 0，后续补偿脚本另行规划。

#### 具体要求

- [x] 1.1 `ttpos_statistics_sale` 新增字段 `order_source_uuid bigint unsigned NOT NULL DEFAULT 0 COMMENT '订单来源UUID（0=店内，>0=外卖/渠道）'`。
  - 迁移文件：`admin/database/migrations/20251126010101_add_order_source_uuid_to_statistics_sale_table.php`
  - 已添加索引：`idx_order_source_uuid`
- [x] 1.2 `main/app/model/statistics.go` 等模型结构新增字段映射。
  - 已添加 `StatisticsSale.OrderSourceUuid` 字段
  - 已更新 `StatisticsSaleData` 结构体，新增外卖来源相关统计字段
- [x] 1.3 `statisticsSrv.SaveSale` 必须从 `saleBill.OrderSourceUuid` 复制数据，未设置时置 0。
  - 已在 `SaveSale` 方法中从 `saleBill.OrderSourceUuid` 赋值

---

### Requirement 2: 即时订单统计按来源拆分

**用户故事**: 作为 门店管理员，我想在报表中区分即时订单（counter order）的店内 vs 外卖来源，以便洞察渠道表现。

#### 验收标准

1. **WHEN** 调用 `CountSale` API **THEN** 现有 `total_instant_order_amount`、`avg_instant_order_amount` 需仅统计 `order_source_uuid = 0`（店内）并保持字段名称不变；同时新增：
   - `total_instant_order_takeaway_num`（order_source_uuid > 0 的即时订单数量）
   - `total_instant_order_takeaway_amount`（order_source_uuid > 0 的即时订单金额）
   - `min_instant_order_takeaway_amount`（最小即时订单金额-外卖）
   - `max_instant_order_takeaway_amount`（最大即时订单金额-外卖）
   - `avg_instant_order_takeaway_amount`（平均即时订单金额-外卖）
2. **WHEN** `order_source_uuid` > 0 且 `desk_uuid = 0` **THEN** 被计入“即时订单-外卖来源”。
3. **WHEN** `order_source_uuid = 0` 且 `desk_uuid = 0` **THEN** 被计入“即时订单-店内”。
4. **WHEN** `CountSaleDays` 被调用 **THEN** 每天的响应也需携带上述四个字段。

#### 具体要求

- [x] 2.1 `CountSaleResp`、`CountSaleDaysResp` DTO 新增字段并补充 JSON 标识。
  - 已添加：`TotalInstantOrderTakeawayNum`、`TotalInstantOrderTakeawayAmount`、`MinInstantOrderTakeawayAmount`、`MaxInstantOrderTakeawayAmount`、`AvgInstantOrderTakeawayAmount`
- [x] 2.2 `repository.statistics.go` 的 `countSaleSubQuerySelect` / `countSaleSelect` 及关联 SQL 需根据 `order_source_uuid` 做条件聚合。
  - 已在子查询中添加 `order_source_uuid` 字段
  - 已使用 `COUNT(CASE WHEN ...)` 和 `SUM(IF(...))` 实现条件聚合
  - 已更新最小/最大/平均金额的计算逻辑，区分店内和外卖来源
- [x] 2.3 所有调用方在短期内无需改动即可读取新增字段，保持向下兼容（默认值 0）。
  - 新增字段使用 `int64` 和 `float64` 类型，默认值为 0

---

### Requirement 3: 数据输出与兼容性

**用户故事**: 作为 收银机与后台报表消费者，我希望新增字段不会破坏现有接口，并能在 UI 中逐步接入。

#### 验收标准

1. **WHEN** 新字段未在前端使用 **THEN** 响应 JSON 默认值为 0，不影响旧版客户端。
2. **WHEN** 现网环境升级 **THEN** 需提供一次性脚本或操作说明，确保统计表新增列与应用启动顺序不会导致查询失败。
3. **WHEN** 单门店关闭外卖渠道 **THEN** 拆分字段仍返回 0，且总即时订单数=店内+外卖来源。

#### 具体要求

- [ ] 3.1 在发布说明中标注“依赖统计库新增列”，要求 DBA 先执行迁移再发布应用。
- [ ] 3.2 补充监控/报警说明：如查询 SQL 因缺列报错需快速回滚。

---

## 非功能需求

- **代码架构**: 严格遵循 Controller → Service → Repository 分层，Service 仅依赖接口，不直接依赖 Repository。
- **API 规范**: URL snake_case，`data` 返回对象，分页在 `meta` 中；参考 `.cursor/rules/api.mdc`。
- **数据库规范**: 表字段遵循 `.cursor/rules/database.mdc`，新增列需建索引（如后续按来源筛选）。
- **性能**: `CountSale` 单次响应需 <200ms，聚合 SQL 需评估是否需要复合索引（`(is_takeout, desk_uuid, order_source_uuid)`）。
- **测试**: 统计模块为高风险模块，Service & Repository 单元测试覆盖率需 ≥90%，并包含 order_source 的边界用例。
- **国际化**: 无新增文案；如后续 UI 显示需引用 i18n 资源。
- **安全**: API 需要授权校验，SQL 使用参数化，禁止字符串拼接。

---

## 验收标准

### 功能验收

1. **即时订单拆分**: `CountSale` 返回的即时订单总量等于拆分后两类之和，且展示正确。
2. **数据一致性**: 新增列写入值与 `ttpos_sale_bill.order_source_uuid` 完全一致（抽样校验）。
3. **兼容稳定**: 旧版客户端/脚本调用 `CountSale` 不报错，新增字段默认为 0。

### 测试验收

1. **单元测试**: `statisticsSrv`、`statisticsRepo` 新增逻辑覆盖率 ≥90%。
2. **API 测试**: `CountSale`、`CountSaleDays` 接口新增字段均在测试用例中覆盖。
3. **集成测试**: 模拟含店内/外卖订单的端到端流程，确认保存与查询一致。
4. **手动测试**: 后台报表、交班导出在无前端改造的情况下仍正常加载。

### 文档验收

1. **design.md** 描述数据库变更与统计逻辑。
2. **tasks.md** 列出并跟踪迁移、代码改造、测试任务。
3. **发布说明** 标注数据库依赖与回滚策略。

---

## 约束条件

### 技术约束

- 必须使用现有 `ttpos_sale_bill.order_source_uuid` 数据源，不新增额外表。
- 禁止在 Service 中直接拼装 SQL，所有聚合在 Repository 完成。
- 不允许使用 panic；所有错误需向上返回 `error`，并记录日志。

### 业务约束

- 拆分范围仅限“即时订单”；桌台与外送订单沿用现有逻辑（后续迭代再扩展）。
- 新增列仅用于统计，不影响收费逻辑。

### 资源约束

- 开发时间：5 天
- Story Point：8（中风险）

---

## 依赖关系

### 技术依赖

- `main/app/service/statistics.go` - 保存统计数据的入口。
- `main/app/repository/statistics.go` - 统计聚合 SQL。
- `main/app/model/statistics_sale.go` - 统计表模型定义。

### 服务依赖

- 本次仅涉及 Main 服务内部调用，不依赖 BMP/PHP。

### 业务依赖

- `ttpos_sale_bill` 已经存在 `order_source_uuid` 字段（由上游订单模块保障正确性）。

---

## 风险和缓解

### 风险 1: 历史数据缺失

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 新增列默认值 0，保证统计查询不报错。
- 另行计划 backfill 任务，或在 UI 中允许“未知来源”提示。

### 风险 2: SQL 性能下降

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在 `ttpos_statistics_sale` 上评估是否需要 `(desk_uuid, order_source_uuid)` 组合索引。
- 发布前在 1w+ 记录数据集上完成压测。

---

## 时间表

- **Phase 1 - 数据库与模型**: 1 日
- **Phase 2 - 服务/仓储改造**: 2 日
- **Phase 3 - 测试与发布准备**: 2 日
- **总计**: 5 日（SP = 8）

---

## 参考资料

- `.cursor/rules/go-main.mdc`
- `.cursor/rules/api.mdc`
- `.cursor/rules/database.mdc`
- `docs/team/proposals/2025-11/order-source-split-statistics.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/王昱/2025-11/2025-11-25.md`

---

**版本**: v1.0.1  
**创建日期**: 2025-11-25  
**最后更新**: 2025-11-26  
**作者**: TTPOS Team (Backend)  
**审核者**: 待定
