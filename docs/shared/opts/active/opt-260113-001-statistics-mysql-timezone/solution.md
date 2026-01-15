# Opt-260113-001 优化方案

## 需求概述

统计功能中使用了 MySQL 的 `FROM_UNIXTIME` 函数进行日期格式化，但该函数使用的是 MySQL 服务器的时区，而不是业务时区（商户设置的时区），导致统计结果错误。

**影响范围**：
- 综合运营统计 (`CountBusinessSummary`) - 已修复
- 支付方式统计 (`CountBusinessPaymentMethod`) - 需修复
- 销售天数统计 (`CountSaleDays`) - 需修复
- 支付天数统计 (`CountPaymentDays`) - 需修复
- 区域统计 (`CountAreaDays`) - 需修复
- 会员数量天数统计 (`CountMemberNumDays`) - 需修复
- 会员支付天数统计 (`CountMemberPaymentDays`) - 需修复
- 退款金额汇总统计 (`CountRefundSummary`) - 部分需修复
- 其他约 14 处使用 `FROM_UNIXTIME` 的统计查询

## 问题分析

### 性能瓶颈分析

**核心问题**：
- MySQL 的 `FROM_UNIXTIME` 函数使用服务器时区进行时间转换
- 当服务器时区与业务时区不一致时，日期分组错误
- 跨时区商户（如日本、泰国、土耳其等）的统计数据不准确

**数据支撑**：
- 代码中搜索到约 105 处 `FROM_UNIXTIME` 的使用
- 其中 14 处关键统计查询受影响
- 已有部分方法（如 `CountBusinessSummary`、`CountRefundSummary`）已采用应用层时区转换方案

**影响因素**：
- MySQL 服务器时区设置（通常为 UTC）
- 商户设置的业务时区（如 `Asia/Shanghai`、`Asia/Tokyo` 等）
- 跨日期边界时间（如 23:00-01:00）的订单统计

## 优化方案

### 方案对比

**方案 1: 使用 MySQL `CONVERT_TZ` 函数**
- 优点: 
  - 在数据库层完成时区转换，性能较好
  - 代码改动相对较小
  - 利用 MySQL 内置时区转换能力
- 缺点: 
  - 需要将时区名称（如 `Asia/Shanghai`）转换为偏移量（如 `+08:00`）
  - 需要处理夏令时等复杂情况
  - 依赖 MySQL 时区数据完整性
- 实施成本: 中等（需要修改 SQL 查询，添加时区转换工具函数）
- 预期收益: 数据准确性 100%，性能良好
- 风险: 
  - 时区名称到偏移量的转换可能不准确（夏令时问题）
  - 需要确保 MySQL 时区数据完整

**方案 2: 在应用层转换时区（✅ 最终选择）**
- 优点: 
  - 完全控制时区转换逻辑，不依赖 MySQL 时区数据
  - 可以处理复杂的时区规则（包括夏令时）
  - 与项目已有实现保持一致（`CountBusinessSummary`、`CountRefundSummary` 已采用此方案）
  - 代码更清晰，时区转换逻辑集中
- 缺点: 
  - 需要重构 SQL 查询，返回原始数据
  - 在应用层计算日期分组，可能增加内存使用
  - 代码改动较大
- 实施成本: 较高（需要重构 SQL 查询，修改数据处理逻辑）
- 预期收益: 数据准确性 100%，完全控制时区转换
- 风险: 
  - 需要重构大量 SQL 查询
  - 可能引入新的 Bug（需要充分测试）

**方案 3: 设置 MySQL Session Timezone + `CONVERT_TZ`**
- 优点: 
  - 结合两种方案的优点
  - 可以在连接级别设置时区
- 缺点: 
  - 需要修改数据库连接配置
  - 可能影响其他查询
- 实施成本: 高（需要修改数据库连接配置）
- 预期收益: 中等
- 风险: 
  - 需要确保每个连接都正确设置时区
  - 连接池复用可能导致时区混乱

**✅ 最终选择: 方案 2（在应用层转换时区）**

理由: 
1. **一致性**：项目已有部分代码采用此方案（`CountBusinessSummary`、`CountRefundSummary`），保持一致更易维护
2. **准确性**：完全控制时区转换逻辑，不依赖 MySQL 时区数据，可以处理复杂的时区规则
3. **可维护性**：时区转换逻辑集中在应用层，代码更清晰
4. **已有工具**：项目已有 `utils.SetTimezone(timezone).FormatUnixTime()` 工具，可直接复用
5. **规范对齐**：符合数据库规范（database.mdc）中的要求：时区转换在应用层完成

### 实施步骤

1. **修改 Repository 层方法签名**
   - 为所有受影响的统计方法添加 `timezone` 参数
   - 确保 Request 结构体包含 `Timezone` 字段

2. **修改 SQL 查询**
   - 移除 `FROM_UNIXTIME` 的日期分组（`GROUP BY FROM_UNIXTIME(...)`）
   - 返回原始数据（包含时间戳字段，不进行日期格式化）
   - SQL 查询只负责数据过滤和聚合，不进行日期分组

3. **在应用层进行时区转换和日期分组**
   - 使用 `utils.SetTimezone(timezone).FormatUnixTime()` 进行时区转换
   - 在应用层使用 `map[string]*groupData` 进行日期分组
   - 使用 `decimal.Decimal` 进行精确计算

4. **修改 Service 层调用**
   - 在调用 Repository 方法时传递时区参数
   - 从 `ctx.GetCompanySetting().Timezone` 获取时区

5. **全面检查并修复**
   - 搜索所有 `FROM_UNIXTIME` 的使用
   - 逐一修复，确保都使用应用层时区转换

6. **编写单元测试**
   - 测试时区转换的正确性
   - 测试跨日期边界的统计准确性
   - 测试不同时区的统计准确性

### 技术方案

#### 数据结构变更

无需变更数据库结构。

#### 代码修改

**1. 参考实现：`CountBusinessSummary` 方法**

已有实现展示了应用层时区转换的模式：

```go
// 1. SQL 查询返回原始数据（包含时间戳，不进行日期分组）
rawQuery := `
    SELECT
        sb.finish_time,
        sb.uuid AS sale_bill_uuid,
        ...
    FROM ttpos_sale_bill AS sb
    WHERE ...
    GROUP BY sb.uuid, sb.finish_time  -- 只按订单分组，不按日期分组
`

// 2. 在应用层进行时区转换和日期分组
timeUtil := utils.SetTimezone(req.Timezone)
groupedData := make(map[string]*groupData)
for _, item := range rawData {
    var dateKey string
    if req.Cycle == 1 {
        dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01")  // 按月
    } else {
        dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01-02")  // 按日
    }
    // 初始化分组数据并累加统计
    if groupedData[dateKey] == nil {
        groupedData[dateKey] = &groupData{}
    }
    group := groupedData[dateKey]
    // ... 累加统计数据
}
```

**2. 需要修复的方法**

以下方法需要从 `FROM_UNIXTIME` 改为应用层时区转换：

- `CountSaleDays` - 移除 `GROUP BY FROM_UNIXTIME(complete_time, '%Y-%m-%d')`，在应用层分组
- `CountPaymentDays` - 移除 `GROUP BY FROM_UNIXTIME(sp.complete_time, '%Y-%m-%d')`，在应用层分组
- `CountAreaDays` - 移除 `GROUP BY FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d')`，在应用层分组
- `CountMemberNumDays` - 移除 `GROUP BY FROM_UNIXTIME(create_time, '%Y-%m-%d')`，在应用层分组
- `CountMemberPaymentDays` - 移除 `GROUP BY FROM_UNIXTIME(smp.complete_time, '%Y-%m-%d')`，在应用层分组
- `CountBusinessPaymentMethod` - 移除 `GROUP BY FROM_UNIXTIME(po.create_time, '%Y-%m-%d')`，在应用层分组
- `CountRefundSummary` - 修复 `orderCountQuery` 和 `takeoutOrderCountQuery` 中的 `FROM_UNIXTIME`

**3. 时区工具类**

项目已有 `utils.SetTimezone(timezone)` 工具类，提供以下方法：
- `FormatUnixTime(timestamp int64, layout string) string` - 将时间戳转换为指定格式的时间字符串（使用业务时区）
- 支持常见的时区：`Asia/Shanghai`、`Asia/Tokyo`、`Asia/Bangkok`、`Europe/Istanbul` 等

**4. Service 层修改示例**

```go
func (s *statisticsSrv) CountSaleDays(ctx context.Context, req CountReq, days []string) []CountSaleDaysResp {
    repo := repository.NewStatisticsRepo(ctx.GetDB())
    timezone := ctx.GetCompanySetting().Timezone  // 获取业务时区
    
    opts := s.buildCountOpts(ctx, req)
    saleData := repo.CountSaleDays(timezone, opts...)  // 传递时区参数
    memberData := repo.CountMemberDays(timezone, opts...)
    
    // ... 后续处理
}
```

## 收益评估

### 性能提升

- **数据准确性**: 从不准确 → 100% 准确
- **跨时区支持**: 从不支持 → 完全支持
- **时区转换准确性**: 完全控制时区转换逻辑，不依赖 MySQL 时区数据

### 代码质量

- **一致性**: 所有统计查询采用统一的时区转换方案
- **可维护性**: 时区转换逻辑集中在应用层，代码更清晰
- **可测试性**: 应用层逻辑更容易编写单元测试

### 技术债务

- **消除约 14 处关键统计查询的时区问题**
- **统一时区转换方案，便于后续维护**

## 影响分析

### 兼容性

- **数据库兼容性**: 无影响，不依赖数据库时区设置
- **API 兼容性**: 无影响，接口参数和返回格式不变
- **业务兼容性**: 统计数据将更加准确，对业务有正面影响

### 风险评估

- **代码改动风险**: 中等（需要重构多个方法，需要充分测试）
- **性能风险**: 低（应用层分组可能增加少量内存使用，但对统计查询性能影响可接受）
- **数据准确性风险**: 低（已有实现验证了方案的可行性）

### 回滚方案

如果出现问题，可以：
1. 回滚代码到修改前的版本
2. 数据库结构未变更，无需数据迁移
3. 影响范围限于统计查询，不影响核心业务功能

## 测试计划

### 功能测试

- **时区转换测试**
  - 测试不同时区（`Asia/Shanghai`、`Asia/Tokyo`、`Asia/Bangkok` 等）的统计准确性
  - 测试跨日期边界时间（如 23:00-01:00）的订单统计

- **统计准确性测试**
  - 对比修复前后的统计结果
  - 验证日期分组的正确性
  - 验证金额汇总的准确性

- **回归测试**
  - 确保所有统计功能正常工作
  - 确保不影响其他功能

### 性能测试

- **查询性能测试**
  - 对比修复前后的查询性能
  - 验证应用层分组的性能影响

- **内存使用测试**
  - 验证应用层分组的内存使用情况
  - 确保大数据量下的性能可接受

### 灰度发布

- **灰度策略**: 先在小范围商户测试，验证统计准确性后再全量发布
- **监控指标**: 
  - 统计查询的响应时间
  - 统计数据的一致性
  - 错误日志和异常情况

## 上线计划

### 发布时间

- **预计开发时间**: 3-5 天
- **预计测试时间**: 2-3 天
- **预计发布时间**: 开发测试完成后

### 监控指标

- **统计查询响应时间**: 监控查询性能
- **统计数据准确性**: 对比修复前后的数据
- **错误日志**: 监控是否有异常错误

### 应急预案

1. **发现问题立即回滚**
2. **监控统计查询的响应时间和错误率**
3. **准备数据对比脚本，验证统计准确性**

## 经验沉淀

### 实施要点

1. **参考已有实现**：`CountBusinessSummary` 和 `CountRefundSummary` 已经采用应用层时区转换方案，可作为参考
2. **使用工具类**：使用 `utils.SetTimezone(timezone).FormatUnixTime()` 进行时区转换
3. **精确计算**：使用 `decimal.Decimal` 进行金额计算，避免精度问题
4. **充分测试**：重点测试跨日期边界和不同时区的场景

### 关键技术

- **时区转换**：使用 Go 标准库 `time.LoadLocation` 和 `time.Time.In()` 进行时区转换
- **日期分组**：在应用层使用 `map[string]*groupData` 进行日期分组
- **精确计算**：使用 `shopspring/decimal` 库进行精确的金额计算

### 注意事项

1. **时区工具类**：确保使用 `utils.SetTimezone(timezone)` 而不是直接使用 `time.LoadLocation`
2. **日期格式**：按日使用 `"2006-01-02"`，按月使用 `"2006-01"`
3. **时间戳字段**：SQL 查询返回的时间戳字段名称要一致，便于应用层处理
4. **测试覆盖**：重点测试跨日期边界和不同时区的场景
