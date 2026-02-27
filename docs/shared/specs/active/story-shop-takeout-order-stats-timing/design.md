# 调整外卖订单统计时间逻辑 设计方案

## 📋 基本信息

| 项目            | 内容                              |
| --------------- | --------------------------------- |
| **Spec ID**     | story-shop-takeout-order-stats-timing |
| **设计版本**    | v1.0.0                            |
| **创建日期**    | 2026-02-26                        |
| **最后更新**    | 2026-02-26                        |

---

## 1. 概述

### 1.1 设计目标

修改外卖订单统计的**查询逻辑**，实现「实时监控 + 最终准确」的双重目标：

- **未完成订单**：使用 `accepted_time`（接单时间）匹配时段，保证实时监控能力
- **已完成订单**：使用 `completed_time`（完成时间）匹配时段，保证统计准确性

### 1.2 核心变更

当前逻辑：所有订单统一使用 `accepted_time` 进行时间范围筛选

目标逻辑：
```
IF order_state = 40 (已完成) THEN
    使用 completed_time 匹配时段
ELSE
    使用 accepted_time 匹配时段
```

### 1.3 影响范围

| 层级 | 文件 | 变更类型 |
|------|------|----------|
| Repository | `main/app/repository/statistics_takeout.go` | 修改查询条件 |

**不涉及**：
- 数据模型变更
- API 接口变更
- Service 层变更
- 前端变更

---

## 2. 详细设计

### 2.1 查询条件构建策略

#### 2.1.1 时间条件改造

**原时间条件**：
```sql
WHERE accepted_time >= ? AND accepted_time <= ?
```

**新时间条件**（使用 OR 逻辑）：
```sql
WHERE (
    -- 已完成订单：使用完成时间
    (order_state = 40 AND completed_time >= ? AND completed_time <= ?)
    OR
    -- 未完成订单：使用接单时间
    (order_state != 40 AND accepted_time >= ? AND accepted_time <= ?)
)
```

#### 2.1.2 需要改造的方法

以下 `StatisticsTakeoutRepo` 方法需要修改时间筛选逻辑：

| 方法 | 当前逻辑 | 改造点 |
|------|----------|--------|
| `CountTakeoutSale` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutPayment` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutReceivedAmount` | `accepted_time` 范围筛选 | 添加状态判断 |
| `RankTakeoutProduct` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutBusinessTimePeriod` | `accepted_time` 范围筛选 + 时段分组 | 改造复杂 |
| `CountTakeoutBusinessSummary` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutChannelSale` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutChannelSaleByPlatform` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutPaymentMethodRawData` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutCategory` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutProduct` | `accepted_time` 范围筛选 | 添加状态判断 |
| `CountTakeoutRefundAmount` | `accepted_time` 范围筛选 | 添加状态判断 |

### 2.2 辅助函数设计

为避免重复代码，抽取公共的时间条件构建函数：

```go
// buildDynamicTimeCondition 构建动态时间条件
// 已完成订单(order_state=40)使用 completed_time
// 其他状态订单使用 accepted_time
func buildDynamicTimeCondition(tableAlias string, timeStart, timeEnd int64) string {
    if tableAlias != "" {
        tableAlias = tableAlias + "."
    }
    return fmt.Sprintf(`(
        (%sorder_state = %d AND %scompleted_time >= %d AND %scompleted_time <= %d)
        OR
        (%sorder_state != %d AND %saccepted_time >= %d AND %saccepted_time <= %d)
    )`,
        tableAlias, valueobject.TakeoutOrderStateCompleted,
        tableAlias, timeStart, tableAlias, timeEnd,
        tableAlias, valueobject.TakeoutOrderStateCompleted,
        tableAlias, timeStart, tableAlias, timeEnd,
    )
}
```

### 2.3 特殊处理：时段分组方法

`CountTakeoutBusinessTimePeriod` 方法需要特殊处理，因为它使用时间字段进行时段分组：

**原 SQL**：
```sql
SELECT
    FLOOR(accepted_time / period_seconds) * period_seconds AS period_start_time,
    ...
FROM ttpos_takeout_order
WHERE accepted_time >= ? AND accepted_time <= ?
GROUP BY period_start_time
```

**改造后 SQL**：
```sql
SELECT
    FLOOR(
        CASE WHEN order_state = 40 THEN completed_time ELSE accepted_time END
        / period_seconds
    ) * period_seconds AS period_start_time,
    ...
FROM ttpos_takeout_order
WHERE (
    (order_state = 40 AND completed_time >= ? AND completed_time <= ?)
    OR
    (order_state != 40 AND accepted_time >= ? AND accepted_time <= ?)
)
GROUP BY period_start_time
```

### 2.4 原始数据结构改造

部分返回原始数据的结构体需要调整字段名以反映实际语义：

| 结构体 | 当前字段 | 改造 |
|--------|----------|------|
| `takeoutBusinessSummaryRawData` | `AcceptedTime` | 改为 `StatTime`（统计时间） |
| `TakeoutChannelSaleRawData` | `AcceptedTime` | 改为 `StatTime`（统计时间） |
| `TakeoutPaymentMethodRawData` | `AcceptedTime` | 改为 `StatTime`（统计时间） |

**改造后查询**：
```sql
SELECT
    CASE WHEN order_state = 40 THEN completed_time ELSE accepted_time END AS stat_time,
    ...
```

---

## 3. 数据库设计

### 3.1 索引要求

确保 `ttpos_takeout_order` 表存在以下索引：

| 索引 | 字段 | 用途 |
|------|------|------|
| `idx_accepted_time` | `accepted_time` | 未完成订单时间筛选 |
| `idx_completed_time` | `completed_time` | 已完成订单时间筛选 |
| `idx_order_state` | `order_state` | 状态判断 |

**复合索引建议**（可选优化）：
```sql
CREATE INDEX idx_order_state_completed_time ON ttpos_takeout_order(order_state, completed_time);
CREATE INDEX idx_order_state_accepted_time ON ttpos_takeout_order(order_state, accepted_time);
```

### 3.2 无数据迁移

本次改动仅涉及查询逻辑，不需要数据迁移。

---

## 4. 测试设计

### 4.1 单元测试用例

| 用例 | 输入 | 预期输出 |
|------|------|----------|
| 已完成订单按完成时间匹配 | 订单 1:30 接单、2:15 完成，查询 2-3 点 | 包含该订单 |
| 已完成订单不按接单时间匹配 | 订单 1:30 接单、2:15 完成，查询 1-2 点 | 不包含该订单 |
| 未完成订单按接单时间匹配 | 订单 1:30 接单、未完成，查询 1-2 点 | 包含该订单 |
| 取消订单按接单时间匹配 | 订单 1:30 接单、1:45 取消，查询 1-2 点 | 包含该订单 |

### 4.2 集成测试

- 验证各统计接口返回数据的正确性
- 验证时段统计的分组正确性

---

## 5. 风险评估

### 5.1 历史数据口径变化

**风险**：上线后统计数据会与历史数据不一致

**缓解**：
- 这是预期行为，新逻辑对所有订单生效
- 上线前与业务确认数据口径变化

### 5.2 查询性能

**风险**：OR 条件可能影响查询性能

**缓解**：
- 确保索引覆盖
- 测试环境验证查询性能
- 考虑使用 UNION 替代 OR（如性能有问题）

---

## 6. 实现步骤

1. 添加 `buildDynamicTimeCondition` 辅助函数
2. 修改各统计方法的时间条件
3. 修改原始数据结构体字段名
4. 编写单元测试
5. 验证查询性能

---

**版本**: v1.0.0
**创建日期**: 2026-02-26
