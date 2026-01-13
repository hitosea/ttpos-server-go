# 代码审查：统计时区修复

## 审查范围

本次审查针对以下修复的方法：
1. CountSaleDays
2. CountPaymentDays
3. CountAreaDays
4. CountMemberNumDays
5. CountMemberDays
6. CountMemberPaymentDays
7. CountRefundSummary (orderCountQuery 和 takeoutOrderCountQuery)

## 审查项目

- [x] 1. SQL错误检查
- [x] 2. 是否影响其他接口
- [x] 3. 是否除0
- [x] 4. 统计逻辑是否与原来不一致

---

## 1. SQL错误检查

### CountSaleDays
**SQL查询位置**: `main/app/repository/statistics.go:236-590`

**检查项**:
- ✅ SQL语法正确：使用 `GROUP BY sb.uuid, sb.finish_time`，返回原始数据
- ✅ 字段映射正确：所有字段都有对应的 gorm tag
- ✅ JOIN 正确：LEFT JOIN 连接相关表

### CountPaymentDays
**SQL查询位置**: `main/app/repository/statistics.go:650-780`

**检查项**:
- ✅ SQL语法正确：使用 `LEFT JOIN` 连接支付方式表
- ✅ 字段映射正确：所有字段都有对应的 gorm tag
- ✅ 无 GROUP BY：返回原始数据，在应用层分组

### CountAreaDays
**SQL查询位置**: `main/app/repository/statistics.go:1008-1160`

**检查项**:
- ✅ SQL语法正确：使用 `LEFT JOIN` 连接桌台和区域表
- ✅ 字段映射正确：所有字段都有对应的 gorm tag
- ✅ 无 GROUP BY：返回原始数据，在应用层分组

### CountMemberNumDays
**SQL查询位置**: `main/app/repository/statistics.go:1358-1412`

**检查项**:
- ✅ SQL语法正确：简单的 SELECT 查询
- ✅ 字段映射正确：`create_time` 字段映射正确

### CountMemberDays
**SQL查询位置**: `main/app/repository/statistics.go:1547-1656`

**检查项**:
- ✅ SQL语法正确：简单的 SELECT 查询
- ✅ 字段映射正确：所有字段都有对应的 gorm tag

### CountMemberPaymentDays
**SQL查询位置**: `main/app/repository/statistics.go:1709-1820`

**检查项**:
- ✅ SQL语法正确：使用 `LEFT JOIN` 连接支付方式表
- ✅ 字段映射正确：所有字段都有对应的 gorm tag
- ✅ 无 GROUP BY：返回原始数据，在应用层分组

### CountRefundSummary
**SQL查询位置**: `main/app/repository/statistics.go:3317-3389`

**检查项**:
- ✅ orderCountQuery: SQL语法正确，移除了 GROUP BY FROM_UNIXTIME，返回原始数据
- ✅ takeoutOrderCountQuery: SQL语法正确，移除了 GROUP BY FROM_UNIXTIME，返回原始数据
- ✅ 字段映射正确：所有字段都有对应的 gorm tag

**结论**: ✅ 所有SQL查询语法正确，无错误

---

## 2. 是否影响其他接口

### 方法签名变更检查

#### CountSaleDays
- **原签名**: `CountSaleDays(opts ...DBOption)`
- **新签名**: `CountSaleDays(timezone string, opts ...DBOption)`
- **调用位置检查**:
  - `main/app/service/statistics.go:230` ✅ 已更新，传递 `timezone` 参数
- **影响**: ✅ 无影响，所有调用已更新

#### CountPaymentDays
- **原签名**: `CountPaymentDays(opts ...DBOption)`
- **新签名**: `CountPaymentDays(timezone string, opts ...DBOption)`
- **调用位置检查**:
  - `main/app/service/statistics.go:495` ✅ 已更新，传递 `timezone` 参数
- **影响**: ✅ 无影响，所有调用已更新

#### CountAreaDays
- **原签名**: `CountAreaDays(opts ...DBOption)`
- **新签名**: `CountAreaDays(timezone string, opts ...DBOption)`
- **调用位置检查**:
  - `main/app/service/statistics.go:1278` ✅ 已更新，传递 `timezone` 参数
- **影响**: ✅ 无影响，所有调用已更新

#### CountMemberNumDays
- **原签名**: `CountMemberNumDays(opts ...DBOption)`
- **新签名**: `CountMemberNumDays(timezone string, opts ...DBOption)`
- **调用位置检查**:
  - `main/app/service/statistics.go:1790` ✅ 已更新，传递 `timezone` 参数
- **影响**: ✅ 无影响，所有调用已更新

#### CountMemberDays
- **原签名**: `CountMemberDays(opts ...DBOption)`
- **新签名**: `CountMemberDays(timezone string, opts ...DBOption)`
- **调用位置检查**:
  - `main/app/service/statistics.go:231` ✅ 已更新，传递 `timezone` 参数（在 CountSaleDays 中调用）
- **影响**: ✅ 无影响，所有调用已更新

#### CountMemberPaymentDays
- **原签名**: `CountMemberPaymentDays(opts ...DBOption)`
- **新签名**: `CountMemberPaymentDays(timezone string, opts ...DBOption)`
- **调用位置检查**:
  - `main/app/service/statistics.go:667` ✅ 已更新，传递 `timezone` 参数
- **影响**: ✅ 无影响，所有调用已更新

#### CountRefundSummary
- **签名**: 未变更（使用 `req.Timezone` 字段）
- **调用位置检查**:
  - Service层调用未变更，使用 `CountRefundSummaryReq` 结构体，其中已包含 `Timezone` 字段
- **影响**: ✅ 无影响

**结论**: ✅ 所有接口调用已更新，无遗漏，不影响其他接口

---

## 3. 是否除0

### CountSaleDays
**检查位置**: `main/app/repository/statistics.go:509-523`

**除0检查**:
```go
if group.OrderNumForAvg > 0 {
    group.AvgOrderAmount = group.AvgOrderAmountSum.Div(decimal.NewFromInt(group.OrderNumForAvg)).Round(2)
}
if group.DeskOrderNumForAvg > 0 {
    group.AvgDeskOrderAmount = group.AvgDeskOrderAmountSum.Div(decimal.NewFromInt(group.DeskOrderNumForAvg)).Round(2)
}
// ... 其他平均值计算
```
- ✅ 所有除法操作都有 `> 0` 检查，无除0风险

### CountPaymentDays
**检查位置**: `main/app/repository/statistics.go:726-730`

**除0检查**:
- ✅ 无除法操作，只有加法和减法

### CountAreaDays
**检查位置**: `main/app/repository/statistics.go:1073-1120`

**除0检查**:
- ✅ 无除法操作，只有加法和减法

### CountMemberNumDays
**检查位置**: `main/app/repository/statistics.go:1379-1390`

**除0检查**:
- ✅ 无除法操作，只有计数操作

### CountMemberDays
**检查位置**: `main/app/repository/statistics.go:1583-1625`

**除0检查**:
- ✅ 无除法操作，只有加法和减法

### CountMemberPaymentDays
**检查位置**: `main/app/repository/statistics.go:1780-1784`

**除0检查**:
- ✅ 无除法操作，只有加法和减法

### CountRefundSummary
**检查位置**: `main/app/repository/statistics.go:3398-3437`

**除0检查**:
- ✅ 无除法操作，只有计数和累加操作
- ✅ 退款率计算在其他地方（Service层），不在 Repository 层

**结论**: ✅ 所有修复的方法都无除0风险

---

## 4. 统计逻辑是否与原来不一致

### 核心变更
所有方法的共同变更：
- **原逻辑**: 使用 MySQL `FROM_UNIXTIME` 在数据库层进行日期分组（使用数据库时区）
- **新逻辑**: 返回原始数据（时间戳），在应用层使用 `utils.SetTimezone(timezone).FormatUnixTime()` 进行日期分组（使用业务时区）

### 详细对比

#### CountSaleDays
**原SQL**:
```sql
GROUP BY FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d')
```

**新逻辑**:
- SQL: `GROUP BY sb.uuid, sb.finish_time`（按订单分组，不按日期分组）
- 应用层: 使用 `timeUtil.FormatUnixTime(item.CompleteTime, "2006-01-02")` 进行日期分组

**逻辑一致性**:
- ✅ 统计字段计算逻辑完全一致
- ✅ 日期分组逻辑一致（只是时区从数据库时区改为业务时区）
- ✅ 排序逻辑一致（按日期排序）

#### CountPaymentDays
**原SQL**:
```sql
GROUP BY sp.payment_method_uuid, FROM_UNIXTIME(sp.complete_time, '%Y-%m-%d')
```

**新逻辑**:
- SQL: 无 GROUP BY，返回原始数据
- 应用层: 使用 `timeUtil.FormatUnixTime(item.CompleteTime, "2006-01-02")` 进行日期分组
- 应用层: 使用 `map[groupKey]*groupData` 按支付方式和日期分组

**逻辑一致性**:
- ✅ 统计字段计算逻辑完全一致：`TotalOrderNum++`, `TotalPaymentAmount += payment_amount - refund_amount`, `TotalRefundAmount += refund_amount`
- ✅ 分组逻辑一致（支付方式 + 日期）
- ✅ 排序逻辑一致（按日期、支付方式排序）

#### CountAreaDays
**原SQL**:
```sql
GROUP BY ss.complete_time, dr.id, FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d')
```

**新逻辑**:
- SQL: 无 GROUP BY，返回原始数据
- 应用层: 使用 `timeUtil.FormatUnixTime(item.CompleteTime, "2006-01-02")` 进行日期分组
- 应用层: 使用 `map[groupKey]*groupData` 按区域和日期分组

**逻辑一致性**:
- ✅ 统计字段计算逻辑完全一致：`AreaSaleAmount`, `AreaBusinessAmount`, `AreaProductNum` 的计算公式一致
- ✅ 分组逻辑一致（区域 + 日期）
- ✅ 排序逻辑一致（按日期、区域排序）

#### CountMemberNumDays
**原SQL**:
```sql
GROUP BY FROM_UNIXTIME(create_time, '%Y-%m-%d')
```

**新逻辑**:
- SQL: 无 GROUP BY，返回原始数据
- 应用层: 使用 `timeUtil.FormatUnixTime(item.CreateTime, "2006-01-02")` 进行日期分组
- 应用层: 使用 `map[string]*groupData` 按日期分组，计数会员数量

**逻辑一致性**:
- ✅ 统计逻辑一致：每个日期统计会员数量（计数）
- ✅ 分组逻辑一致（按日期）
- ✅ 排序逻辑一致（按日期排序）

#### CountMemberDays
**原SQL**:
```sql
GROUP BY FROM_UNIXTIME(complete_time, '%Y-%m-%d')
```

**新逻辑**:
- SQL: 无 GROUP BY，返回原始数据
- 应用层: 使用 `timeUtil.FormatUnixTime(item.CompleteTime, "2006-01-02")` 进行日期分组
- 应用层: 使用 `map[string]*groupData` 按日期分组

**逻辑一致性**:
- ✅ 统计字段计算逻辑完全一致：所有字段的累加逻辑一致
- ✅ 特殊逻辑一致：`TotalRechargeAmount` 的条件判断逻辑一致（`IF(payment_amount - refund_amount = 0, 0, recharge_amount - refund_amount)`）
- ✅ 分组逻辑一致（按日期）
- ✅ 排序逻辑一致（按日期排序）

#### CountMemberPaymentDays
**原SQL**:
```sql
GROUP BY smp.payment_method_uuid, FROM_UNIXTIME(smp.complete_time, '%Y-%m-%d')
```

**新逻辑**:
- SQL: 无 GROUP BY，返回原始数据
- 应用层: 使用 `timeUtil.FormatUnixTime(item.CompleteTime, "2006-01-02")` 进行日期分组
- 应用层: 使用 `map[groupKey]*groupData` 按支付方式和日期分组

**逻辑一致性**:
- ✅ 统计字段计算逻辑完全一致：`TotalOrderNum++`, `TotalPaymentAmount += payment_amount - refund_amount`, `TotalRefundAmount += refund_amount`
- ✅ 分组逻辑一致（支付方式 + 日期）
- ✅ 排序逻辑一致（按日期、支付方式排序）

#### CountRefundSummary
**原SQL (orderCountQuery)**:
```sql
GROUP BY FROM_UNIXTIME(sb.finish_time, IF(? = 1, '%Y-%m', '%Y-%m-%d'))
```

**新逻辑**:
- SQL: 无 GROUP BY，返回原始数据（每个订单一行）
- 应用层: 使用 `map[string]map[uint64]bool` 去重统计每个日期的订单数量
- 应用层: 使用 `timeUtil.FormatUnixTime(item.FinishTime, "2006-01-02")` 或 `"2006-01"` 进行日期分组

**逻辑一致性**:
- ✅ 统计逻辑一致：每个日期统计订单数量（使用 DISTINCT uuid 去重）
- ✅ 去重逻辑一致：使用 map 确保每个订单只被统计一次
- ✅ 分组逻辑一致（按日期，支持按日和按月）

**原SQL (takeoutOrderCountQuery)**:
```sql
GROUP BY FROM_UNIXTIME(accepted_time, IF(? = 1, '%%Y-%%m', '%%Y-%%m-%%d'))
```

**新逻辑**:
- SQL: 无 GROUP BY，返回原始数据（每个订单一行）
- 应用层: 使用 `map[string]map[uint64]bool` 去重统计每个日期的订单数量
- 应用层: 使用 `timeUtil.FormatUnixTime(item.AcceptedTime, "2006-01-02")` 或 `"2006-01"` 进行日期分组

**逻辑一致性**:
- ✅ 统计逻辑一致：每个日期统计订单数量（使用 DISTINCT uuid 去重）
- ✅ 去重逻辑一致：使用 map 确保每个订单只被统计一次
- ✅ 分组逻辑一致（按日期，支持按日和按月）

**结论**: ✅ 所有统计逻辑与原来一致，只是时区转换从数据库层移到了应用层

---

## 总结

### ✅ 审查通过

1. **SQL错误检查**: ✅ 通过
   - 所有SQL查询语法正确
   - 字段映射正确
   - 无语法错误

2. **是否影响其他接口**: ✅ 通过
   - 所有方法签名变更的调用位置已更新
   - 无遗漏的调用
   - 不影响其他接口

3. **是否除0**: ✅ 通过
   - 所有除法操作都有检查
   - 无除0风险

4. **统计逻辑是否与原来不一致**: ✅ 通过
   - 所有统计字段计算逻辑一致
   - 分组逻辑一致（只是时区转换从数据库层移到应用层）
   - 排序逻辑一致

### ⚠️ 注意事项

1. **性能考虑**: 
   - 应用层分组可能增加内存使用（需要加载更多数据到内存）
   - 但对于统计场景，数据量通常可控
   - 时区转换的性能影响较小

2. **测试建议**:
   - 建议进行时区转换准确性测试
   - 建议进行跨日期边界测试
   - 建议进行统计准确性对比测试（修复前后对比）

### 📝 建议

1. 所有修改的代码已通过语法检查（linter无错误）
2. 建议在实际环境中进行功能测试
3. 建议对比修复前后的统计数据，确保一致性（除了时区相关的差异）
