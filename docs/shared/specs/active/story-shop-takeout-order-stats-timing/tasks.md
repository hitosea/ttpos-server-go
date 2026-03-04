# story-shop-takeout-order-stats-timing 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 3 |
| 总任务数 | 5 |
| 已完成 | 5 |
| 完成率 | 100% |

---

## Phase 1: 添加辅助函数

### 1.1 添加 buildDynamicTimeCondition 函数

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 构建动态时间条件，根据订单状态选择使用 accepted_time 或 completed_time |
| Requirements | AC1, AC2 |

**新增代码**:
```go
// buildDynamicTimeCondition 构建动态时间条件
// 已完成订单(order_state=40)使用 completed_time
// 其他状态订单使用 accepted_time
func buildDynamicTimeCondition(tableAlias string, timeStart, timeEnd int64) string {
    if tableAlias != "" {
        tableAlias = tableAlias + "."
    }
    completedState := valueobject.TakeoutOrderStateCompleted
    return fmt.Sprintf(`(
        (%sorder_state = %d AND %scompleted_time >= %d AND %scompleted_time <= %d)
        OR
        (%sorder_state != %d AND %saccepted_time >= %d AND %saccepted_time <= %d)
    )`,
        tableAlias, completedState,
        tableAlias, timeStart, tableAlias, timeEnd,
        tableAlias, completedState,
        tableAlias, timeStart, tableAlias, timeEnd,
    )
}
```

- [x] 完成

---

## Phase 2: 修改基础统计方法

### 2.1 修改 CountTakeoutSale 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改时间条件，使用动态时间判断 |
| Requirements | AC1, AC2 |

- [x] 完成

### 2.2 修改 CountTakeoutPayment 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改时间条件，使用动态时间判断 |

- [x] 完成

### 2.3 修改 CountTakeoutReceivedAmount 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改时间条件，使用动态时间判断 |

- [x] 完成

### 2.4 修改 RankTakeoutProduct 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改时间条件，使用动态时间判断 |

- [x] 完成

### 2.5 修改 CountTakeoutPaymentMethodRawData 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改时间条件，使用动态时间判断；添加 completed_time 到子查询；使用 CASE WHEN 计算 stat_time |

- [x] 完成

### 2.6 修改 CountTakeoutCategory 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改时间条件，使用动态时间判断 |

- [x] 完成

### 2.7 修改 CountTakeoutProduct 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改时间条件，使用动态时间判断 |

- [x] 完成

### 2.8 修改 CountTakeoutRefundAmount 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | **不需要修改** - 该方法只统计取消状态（60）的订单，不是已完成状态（40），继续使用 accepted_time |

- [x] 完成（无需改动）

---

## Phase 3: 修改原始 SQL 方法

### 3.1 修改 CountTakeoutBusinessTimePeriod 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改 Raw SQL，使用动态时间进行时段分组 |
| Requirements | AC1, AC2, AC3, AC4 |
| Complexity | 高（需改造时段分组逻辑）|

**实现要点**:
- 时段分组使用 `CASE WHEN order_state = 40 THEN completed_time ELSE accepted_time END`
- WHERE 条件使用 OR 逻辑合并两种状态的时间筛选
- 参数传递 4 个时间值（完成时间开始/结束 + 接单时间开始/结束）

- [x] 完成

### 3.2 修改 CountTakeoutBusinessSummary 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改 Raw SQL，使用动态时间 |

**变更说明**:
- 修改 WHERE 条件使用动态时间
- 修改 SELECT 中 `accepted_time` 为 `CASE WHEN order_state = 40 THEN completed_time ELSE accepted_time END AS stat_time`

- [x] 完成

### 3.3 修改 CountTakeoutChannelSale 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改 Raw SQL，使用动态时间 |

- [x] 完成

### 3.4 修改 CountTakeoutChannelSaleByPlatform 方法

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 修改 Raw SQL，使用动态时间 |

- [x] 完成

---

## Phase 4: 修改数据结构

### 4.1 修改原始数据结构体字段名

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout.go` |
| Purpose | 将 AcceptedTime 字段改为 StatTime，反映实际语义 |

**已修改的结构体**:
- `takeoutBusinessSummaryRawData`: `AcceptedTime` → `StatTime`
- `TakeoutChannelSaleRawData`: `AcceptedTime` → `StatTime`
- `TakeoutPaymentMethodRawData`: `AcceptedTime` → `StatTime`

**同步修改的调用方** (`main/app/repository/statistics.go`):
- `CountBusinessSummary`: `takeoutItem.AcceptedTime` → `takeoutItem.StatTime`
- `CountBusinessPaymentMethod`: `takeoutItem.AcceptedTime` → `takeoutItem.StatTime`

- [x] 完成

---

## Phase 5: 测试验证

### 5.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/statistics_takeout_test.go`（新增） |
| Purpose | 测试动态时间条件构建函数 |

**测试用例**:
1. `TestBuildDynamicTimeCondition_WithTableAlias` - 测试带表别名的条件生成
2. `TestBuildDynamicTimeCondition_WithDifferentAlias` - 测试不同表别名
3. `TestBuildDynamicTimeCondition_EmptyAlias` - 测试空表别名
4. `TestBuildDynamicTimeCondition_CompletedStateValue` - 测试使用正确的完成状态值（40）

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [x] `go build ./...` 通过
- [x] 单元测试通过（4 个测试用例）

### 功能完整性
- [x] AC1: 未完成订单按接单时间匹配
- [x] AC2: 已完成订单按完成时间匹配
- [x] AC3: 订单 1:30 接单、2:15 完成，查询 1-2 点不包含
- [x] AC4: 订单 1:30 接单、2:15 完成，查询 2-3 点包含
- [x] AC5: 未完成订单 1:30 接单，查询 1-2 点包含

### 文档更新
- [x] 无需数据库迁移
- [x] 无需更新 shop_01.sql

---

## 相关文件索引

| 文件 | 类型 | 操作 |
|------|------|------|
| `main/app/repository/statistics_takeout.go` | Repository | 修改（核心变更） |
| `main/app/repository/statistics.go` | Repository | 修改（调用方字段名更新） |
| `main/app/repository/statistics_takeout_test.go` | 单元测试 | 新增 |

---

**版本**: v1.0.0
**创建日期**: 2026-02-26
**完成日期**: 2026-02-26
