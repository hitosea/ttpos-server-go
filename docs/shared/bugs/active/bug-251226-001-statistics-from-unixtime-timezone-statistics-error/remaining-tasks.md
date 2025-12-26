# 待修复的统计方法清单

## 已修复的方法 ✅

1. ✅ **CountBusinessSummary** - 统计综合运营
   - 已改为应用层分组
   - 使用 decimal 进行精确计算

2. ✅ **CountBusinessPaymentMethod** - 统计支付方式
   - 已改为应用层分组
   - 使用 decimal 进行精确计算

## 待修复的方法（8个）

### 1. CountSaleDays - 统计销售天数
- **位置**: `main/app/repository/statistics.go:186`
- **问题**: 使用 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 分组
- **数据源**: `statistics_sale` 表
- **调用**: `service.CountSaleDays` → `repo.CountSaleDays`
- **修复方案**: 改为应用层分组

### 2. CountPaymentDays - 统计支付天数
- **位置**: `main/app/repository/statistics.go:250`
- **问题**: 使用 `FROM_UNIXTIME(sp.complete_time, '%Y-%m-%d')` 分组
- **数据源**: `statistics_payment` 表
- **调用**: `service.CountPaymentDays` → `repo.CountPaymentDays`
- **修复方案**: 改为应用层分组

### 3. CountAreaDays - 统计区域天数
- **位置**: `main/app/repository/statistics.go:470`
- **问题**: 使用 `FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d')` 分组
- **数据源**: `statistics_sale` 表
- **调用**: `service.CountAreaDays` → `repo.CountAreaDays`
- **修复方案**: 改为应用层分组

### 4. Count7Days - 统计7天数据
- **位置**: `main/app/repository/statistics.go:564`
- **问题**: 使用 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 分组
- **数据源**: `statistics_sale` 表
- **调用**: `service.Count7Days` → `repo.Count7Days`
- **修复方案**: 改为应用层分组

### 5. CountMemberNumDays - 统计会员数量天数
- **位置**: `main/app/repository/statistics.go:595`
- **问题**: 使用 `FROM_UNIXTIME(create_time, '%Y-%m-%d')` 分组
- **数据源**: `member` 表（非统计表）
- **调用**: `service.CountMemberNumDays` → `repo.CountMemberNumDays`
- **修复方案**: 改为应用层分组

### 6. CountMemberDays - 统计会员天数
- **位置**: `main/app/repository/statistics.go:732`
- **问题**: 使用 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 分组
- **数据源**: `statistics_member` 表
- **调用**: `service.CountSaleDays` → `repo.CountMemberDays`
- **修复方案**: 改为应用层分组

### 7. CountMemberPaymentDays - 统计会员支付天数
- **位置**: `main/app/repository/statistics.go:790`
- **问题**: 使用 `FROM_UNIXTIME(smp.complete_time, '%Y-%m-%d')` 分组
- **数据源**: `statistics_member_payment` 表
- **调用**: `service.CountMemberPaymentDays` → `repo.CountMemberPaymentDays`
- **修复方案**: 改为应用层分组

### 8. CountFreePaymentDays - 统计免单支付天数
- **位置**: `main/app/repository/statistics.go:1040`
- **问题**: 使用 `FROM_UNIXTIME(complete_time, '%Y-%m-%d')` 分组
- **数据源**: `statistics_sale` 表
- **调用**: `service.CountFreePaymentDays` → `repo.CountFreePaymentDays`
- **修复方案**: 改为应用层分组

## 修复方案

### 统一方案：应用层分组

所有方法都采用与 `CountBusinessSummary` 相同的方案：

1. **查询原始数据**（不分组）
   - 查询时间范围内的所有记录
   - 返回完整字段（时间戳、金额、数量等）

2. **应用层分组统计**
   - 使用 `time.LoadLocation` 加载业务时区
   - 使用 `time.Unix().In(loc)` 将时间戳转换为业务时区的日期
   - 使用 `map` 按日期分组统计
   - 使用 `decimal` 进行精确计算

3. **排序和返回**
   - 使用 `slices.SortFunc` 排序
   - 返回分组后的结果

### 时区参数传递

这些方法使用 `opts ...DBOption` 参数，时区信息需要通过以下方式传递：

**方案 A**: 修改方法签名，添加时区参数（推荐）
```go
func (r *StatisticsRepo) CountSaleDays(timezone string, opts ...DBOption) []model.StatisticsSaleDaysData
```

**方案 B**: 通过 DBOption 传递时区
```go
// 在 common.go 中添加
func (r *commonRepo) WithTimezone(timezone string) DBOption {
    // 存储时区信息，供 Repository 方法使用
}
```

**方案 C**: 从 Service 层传递时区
```go
// Service 层调用时传递时区
timezone := ctx.GetCompanySetting().Timezone
// 需要修改 Repository 方法签名
```

## 优先级

### 高优先级（直接影响业务）
1. ✅ CountBusinessSummary（已修复）
2. ✅ CountBusinessPaymentMethod（已修复）

### 中优先级（统计报表）
3. CountSaleDays - 销售天数统计
4. CountPaymentDays - 支付天数统计
5. CountAreaDays - 区域天数统计
6. CountMemberDays - 会员天数统计

### 低优先级（辅助统计）
7. Count7Days - 7天统计
8. CountMemberNumDays - 会员数量天数
9. CountMemberPaymentDays - 会员支付天数
10. CountFreePaymentDays - 免单支付天数

## 注意事项

1. **数据源差异**：
   - `CountMemberNumDays` 查询的是 `member` 表（非统计表），使用 `create_time`
   - 其他方法查询的是统计表（`statistics_*`），使用 `complete_time`

2. **分组维度**：
   - 大部分方法按日期分组
   - `CountAreaDays` 按区域 + 日期分组
   - `CountPaymentDays` 和 `CountMemberPaymentDays` 按支付方式 + 日期分组

3. **性能考虑**：
   - 如果数据量很大，可能需要分批查询
   - 考虑添加缓存机制

4. **测试**：
   - 需要测试不同时区的统计准确性
   - 需要测试跨日期边界的数据统计

