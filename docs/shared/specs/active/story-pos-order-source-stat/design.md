# story-pos-order-source-stat 设计文档

> 本文档定义“按订单来源拆分点餐方式统计”功能的技术设计与实现方案。

## 📋 概述

通过为 `ttpos_statistics_sale` 增加 `order_source_uuid` 字段，并在统计服务中贯穿该字段，最终在 `CountSale` / `CountSaleDays` 接口中输出即时订单的来源拆分结果，支撑旧商家后台与 POS 端营业数据的渠道分析。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service (`statisticsSrv`) 仅依赖 Repository 接口；`SaveSale` 仍作为唯一入口，禁止跨层 SQL。
- 所有新逻辑返回 `error`，使用 `errors.WithMessage` 包装；禁止 panic。
- DTO/响应字段使用 snake_case，`data` 内返回对象结构。

### API 设计规范 (api.mdc)

- `CountSale`、`CountSaleDays` 新字段沿用 JSON snake_case：`total_instant_order_num_store` 等。
- 响应默认值为 0，不允许返回 `null`。

### 数据库规范 (database.mdc)

- 新增列命名为 `order_source_uuid`，类型 `bigint unsigned NOT NULL DEFAULT 0`，并在模型中保持一致。
- 迁移需包含 `update_time`/`delete_time` 默认值校验；列追加时需要 `AFTER is_takeout` 保持语义。

---

## 🔄 代码复用分析

| 组件 | 路径 | 复用方式 |
| ---- | ---- | ---- |
| 统计保存服务 | `main/app/service/statistics.go::SaveSale` | 在构造 `model.StatisticsSale` 数据时填充 `OrderSourceUuid` |
| 统计仓储聚合 | `main/app/repository/statistics.go` | 扩展 `countSaleSubQuerySelect`、`countSaleSelect`，引入 cond sum |
| DTO 定义 | `main/app/service/statistics.go` (`CountSaleResp`, `CountSaleDaysResp`) | 在原结构上新增字段，复用序列化逻辑 |

---

## 🏗️ 架构设计

### 数据流

```text
ttpos_sale_bill (order_source_uuid)
        ↓ SaveSale
ttpos_statistics_sale.order_source_uuid
        ↓ repository.CountSale(SubQuery)
Aggregate by is_takeout/desk_uuid/order_source_uuid
        ↓ service.CountSale / CountSaleDays
API 响应 (新增拆分字段)
```

### 关键改动

1. **数据入库**：`SaveSale` 从 `saleBill.OrderSourceUuid` 读取值写入 `model.StatisticsSale.OrderSourceUuid`。
2. **统计 SQL**：子查询 `countSaleSubQuerySelect` 新增 `order_source_uuid` 维度；聚合时按 `order_source_uuid = 0` 与 `>0` 条件拆分即时订单。
3. **DTO 扩展**：`CountSaleResp`、`CountSaleDaysResp` 增加 4 个字段用于表达即时订单店内/外卖来源的数量与金额。

---

## 🗄️ 数据库设计

### 表：ttpos_statistics_sale

**迁移文件**: `admin/database/migrations/20251126010101_add_order_source_uuid_to_statistics_sale_table.php`

```sql
ALTER TABLE `ttpos_statistics_sale`
  ADD COLUMN `order_source_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '订单来源UUID（0=店内，>0=外卖/渠道)' AFTER `is_takeout`,
  ADD KEY `idx_order_source_uuid` (`order_source_uuid`);
```

> 索引用于后续按来源过滤的查询，可与 `desk_uuid` 组合评估联合索引。

**状态**: ✅ 迁移文件已创建，包含字段定义和索引

**字段说明补充**：

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| order_source_uuid | bigint unsigned | 0=店内，>0=各类外卖或第三方来源 |

### 迁移策略

1. 创建 ThinkPHP migration，在 `up` 中追加列与索引，在 `down` 中删除列和索引。
2. 迁移执行顺序：**先数据库，后部署服务**，避免缺列导致服务 panic。

---

## 📊 数据模型

`main/app/model/statistics.go` 中结构体更新：

**StatisticsSale 结构体**：

```go
OrderSourceUuid uint64 `gorm:"column:order_source_uuid;type:bigint(20) unsigned;default:0;comment:订单来源UUID（0=店内，>0=外卖/渠道）;NOT NULL" json:"order_source_uuid"`
```

**StatisticsSaleData 结构体**（新增外卖来源统计字段）：

```go
TotalInstantOrderTakeawayAmount sql.NullFloat64 `gorm:"column:total_instant_order_takeaway_amount;comment:总即时订单金额-外卖"`
TotalInstantOrderTakeawayNum    sql.NullInt64   `gorm:"column:total_instant_order_takeaway_num;comment:总即时订单数量-外卖"`
MinInstantOrderTakeawayAmount   sql.NullFloat64 `gorm:"column:min_instant_order_takeaway_amount;comment:最小即时订单金额-外卖"`
MaxInstantOrderTakeawayAmount   sql.NullFloat64 `gorm:"column:max_instant_order_takeaway_amount;comment:最大即时订单金额-外卖"`
AvgInstantOrderTakeawayAmount   sql.NullFloat64 `gorm:"column:avg_instant_order_takeaway_amount;comment:平均即时订单金额-外卖"`
```

**状态**: ✅ 模型已更新，包含所有必要字段

---

## 🔌 API / DTO 设计

`CountSaleResp` / `CountSaleDaysResp` 字段调整：

**原有字段（仅统计店内，order_source_uuid = 0）**：

- `TotalInstantOrderAmount`：即时订单金额（店内）
- `TotalInstantOrderNum`：即时订单数量（店内）
- `MinInstantOrderAmount`：最小即时订单金额（店内）
- `MaxInstantOrderAmount`：最大即时订单金额（店内）
- `AvgInstantOrderAmount`：平均即时订单金额（店内）

**新增字段（外卖来源，order_source_uuid > 0）**：

- `TotalInstantOrderTakeawayNum`：即时订单数量（外卖）
- `TotalInstantOrderTakeawayAmount`：即时订单金额（外卖）
- `MinInstantOrderTakeawayAmount`：最小即时订单金额（外卖）
- `MaxInstantOrderTakeawayAmount`：最大即时订单金额（外卖）
- `AvgInstantOrderTakeawayAmount`：平均即时订单金额（外卖）

**计算方式（Repository 层 SQL）**：

```sql
-- 子查询中计算外卖来源金额
SUM(IF(desk_uuid = 0 AND order_source_uuid > 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_instant_order_takeaway_amount

-- 聚合查询中拆分统计
COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_meger = 0 THEN 1 END) AS total_instant_order_num
COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_meger = 0 THEN 1 END) AS total_instant_order_takeaway_num
SUM(t.instant_order_takeaway_amount) AS total_instant_order_takeaway_amount
MIN(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND ... THEN t.instant_order_takeaway_amount ELSE NULL END) AS min_instant_order_takeaway_amount
MAX(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND ... THEN t.instant_order_takeaway_amount ELSE NULL END) AS max_instant_order_takeaway_amount
SUM(t.avg_instant_order_takeaway_amount) / COUNT(CASE WHEN ... THEN 1 END) AS avg_instant_order_takeaway_amount
```

**状态**: ✅ DTO 已更新，包含所有新增字段，并在 `CountSale` 和 `CountSaleDays` 中完成赋值

---

## 🧩 组件与实现

### Service 层 (`statisticsSrv`)

- **SaveSale**：在构造 `model.StatisticsSale` 结构时注入 `OrderSourceUuid: saleBill.OrderSourceUuid`。
  - **状态**: ✅ 已实现，从 `saleBill.OrderSourceUuid` 赋值
- **CountSale / CountSaleDays**：读取 Repository 结果中的店内/外卖即时订单统计，并填充所有新增字段。
  - **状态**: ✅ 已实现，包含数量、金额、最小/最大/平均值的完整赋值逻辑

### Repository 层

- 子查询 `countSaleSubQuerySelect` 在聚合字段中对 `order_source_uuid` 做条件判断，确保原有即时订单字段代表店内，并额外产出外卖来源金额/数量。
  - **状态**: ✅ 已在子查询中添加 `order_source_uuid` 字段
- 顶层查询 `countSaleSelect` 聚合新增字段，输出：
  - `total_instant_order_takeaway_num`：外卖来源即时订单数量
  - `total_instant_order_takeaway_amount`：外卖来源即时订单金额
  - `min_instant_order_takeaway_amount`：最小即时订单金额（外卖）
  - `max_instant_order_takeaway_amount`：最大即时订单金额（外卖）
  - `avg_instant_order_takeaway_amount`：平均即时订单金额（外卖）
  - **状态**: ✅ 已实现所有聚合逻辑，使用 `COUNT(CASE WHEN ...)` 和 `SUM(IF(...))` 实现条件聚合
- `CountSaleDays` 同步更新，支持按天拆分统计。
- 若 `ORDER BY` / `GROUP BY` 依赖 `DAY`，需在新增字段之后保持原排序。

### DTO 层

- 修改 `CountSaleResp` 与 `CountSaleDaysResp` 结构体声明，保持 `json` tag 以及默认值（`float64` / `int64`）。
  - **状态**: ✅ 已添加 5 个新字段：`TotalInstantOrderTakeawayNum`、`TotalInstantOrderTakeawayAmount`、`MinInstantOrderTakeawayAmount`、`MaxInstantOrderTakeawayAmount`、`AvgInstantOrderTakeawayAmount`

---

## ⚙️ 测试策略

1. **单元测试**
   - `statisticsSrv.SaveSale`：构造含 `order_source_uuid` 的 `saleBill`，验证写入 `StatisticsSale`。
   - `repository.CountSale`：基于 sqlite 测试数据验证拆分列计算正确。
2. **集成测试**
   - 构造包含店内即时订单 + 外卖即时订单的数据，调用 `CountSale`，确认 `总即时订单 = 店内 + 外卖来源`。
3. **回归测试**
   - 旧接口调用仍然能获取历史字段。

---

## 🔒 错误与回滚

- 若服务部署早于迁移，`order_source_uuid` 列缺失会触发 SQL 报错；需在发布 checklist 中加入“确认迁移执行完成”的步骤。
- 如拆分统计出现性能问题，可通过配置开关暂时隐藏新增字段（由 API 返回 0），同时分析慢查询。

---

## 📈 性能与监控

- 新增字段仅在 `statistics_sale` 表上做条件 `SUM`，不会增加额外 JOIN。
- 监控：Prometheus 统计 `CountSale` 成功率，若因 SQL 出错导致 `code != 1` 提醒。
- 需要在 DBA 层观察 `ttpos_statistics_sale` 表分析，必要时添加复合索引 `(desk_uuid, order_source_uuid)。

---

## 发布与回滚

1. 执行数据库迁移，确认新增列存在。
2. 发布 Main 服务。
3. 若需回滚：
   - 回滚服务版本。
   - 执行 migration `down` 删除列（如确有需要）。

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/王昱/2025-11/2025-11-25.md`

---

**版本**: v1.0.1  
**创建日期**: 2025-11-25  
**最后更新**: 2025-11-26  
**作者**: TTPOS Backend Team  
**审核者**: 待定
