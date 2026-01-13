# Opt-260113-001 优化任务清单

> **当前状态**: 🟢 规划中
> **开始时间**: 2026-01-13
> **预计完成**: 2026-01-20
> **预期收益**: 统计数据准确性从错误 → 100% 准确，完全支持跨时区商户

---

## 📋 任务列表

### 1. 前期准备

- [ ] **性能基线测试**
  - 需求: 记录优化前的统计查询性能数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 记录主要统计查询的响应时间和数据准确性

- [ ] **环境准备**
  - 需求: 准备测试环境和测试数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 准备跨时区测试数据和测试用例

### 2. 代码优化

#### 2.1 修复 CountSaleDays 方法

- [x] **修改 CountSaleDays 方法签名** `main/app/repository/statistics.go`
  - 需求: 添加 `timezone string` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 修改方法签名，添加时区参数
  - 状态: ✅ 已完成

- [x] **修改 CountSaleDays SQL 查询**
  - 需求: 移除 `FROM_UNIXTIME` 的日期分组，返回原始数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 移除 `GROUP BY FROM_UNIXTIME(complete_time, '%Y-%m-%d')`，返回包含 `complete_time` 的原始数据
  - 状态: ✅ 已完成

- [x] **实现应用层时区转换和分组**
  - 需求: 在应用层使用 `utils.SetTimezone(timezone).FormatUnixTime()` 进行时区转换和日期分组
  - 预计时间: 1.5小时
  - 负责人: 
  - 说明: 参考 `CountBusinessSummary` 的实现方式，创建 `saleDaysRawData` 结构体，在应用层进行日期分组和聚合
  - 状态: ✅ 已完成

- [x] **修改 Service 层调用 CountSaleDays**
  - 需求: 在 Service 层传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 修改 `main/app/service/statistics.go` 中的调用，从 `ctx.GetCompanySetting().Timezone` 获取时区并传递
  - 状态: ✅ 已完成

#### 2.2 修复 CountPaymentDays 方法

- [x] **修改 CountPaymentDays 方法签名** `main/app/repository/statistics.go`
  - 需求: 添加 `timezone string` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 状态: ✅ 已完成

- [x] **修改 CountPaymentDays SQL 查询**
  - 需求: 移除 `FROM_UNIXTIME` 的日期分组，返回原始数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 移除 `GROUP BY FROM_UNIXTIME(sp.complete_time, '%Y-%m-%d')`，返回包含 `complete_time` 的原始数据
  - 状态: ✅ 已完成

- [x] **实现应用层时区转换和分组**
  - 需求: 在应用层进行时区转换和日期分组
  - 预计时间: 1.5小时
  - 负责人: 
  - 说明: 创建 `paymentDaysRawData` 结构体，在应用层按支付方式 + 日期分组，使用 `decimal.Decimal` 进行精确计算
  - 状态: ✅ 已完成

- [x] **修改 Service 层调用 CountPaymentDays**
  - 需求: 在 Service 层传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 修改 `main/app/service/statistics.go` 中的调用，从 `ctx.GetCompanySetting().Timezone` 获取时区并传递
  - 状态: ✅ 已完成 

#### 2.3 修复 CountAreaDays 方法

- [x] **修改 CountAreaDays 方法签名** `main/app/repository/statistics.go`
  - 需求: 添加 `timezone string` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 状态: ✅ 已完成

- [x] **修改 CountAreaDays SQL 查询**
  - 需求: 移除 `FROM_UNIXTIME` 的日期分组，返回原始数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 移除 `GROUP BY FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d')`，返回包含 `complete_time` 的原始数据
  - 状态: ✅ 已完成

- [x] **实现应用层时区转换和分组**
  - 需求: 在应用层进行时区转换和日期分组
  - 预计时间: 2小时
  - 负责人: 
  - 说明: 创建 `areaDaysRawData` 结构体，在应用层按区域 + 日期分组，使用 `decimal.Decimal` 进行精确计算，保留 `dr.id` 用于排序
  - 状态: ✅ 已完成

- [x] **修改 Service 层调用 CountAreaDays**
  - 需求: 在 Service 层传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 修改 `main/app/service/statistics.go` 中的调用，从 `ctx.GetCompanySetting().Timezone` 获取时区并传递
  - 状态: ✅ 已完成 

#### 2.4 修复 CountMemberNumDays 方法

- [x] **修改 CountMemberNumDays 方法签名** `main/app/repository/statistics.go`
  - 需求: 添加 `timezone string` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 状态: ✅ 已完成

- [x] **修改 CountMemberNumDays SQL 查询**
  - 需求: 移除 `FROM_UNIXTIME` 的日期分组，返回原始数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 移除 `GROUP BY FROM_UNIXTIME(create_time, '%Y-%m-%d')`，返回包含 `create_time` 的原始数据，注意数据源是 `member` 表
  - 状态: ✅ 已完成

- [x] **实现应用层时区转换和分组**
  - 需求: 在应用层进行时区转换和日期分组
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 创建 `memberNumDaysRawData` 结构体，在应用层按日期分组统计会员数量
  - 状态: ✅ 已完成

- [x] **修改 Service 层调用 CountMemberNumDays**
  - 需求: 在 Service 层传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 修改 `main/app/service/statistics.go` 中的调用，从 `ctx.GetCompanySetting().Timezone` 获取时区并传递
  - 状态: ✅ 已完成 

#### 2.5 修复 CountMemberDays 方法

- [x] **修改 CountMemberDays 方法签名** `main/app/repository/statistics.go`
  - 需求: 添加 `timezone string` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 状态: ✅ 已完成

- [x] **修改 CountMemberDays SQL 查询**
  - 需求: 移除 `FROM_UNIXTIME` 的日期分组，返回原始数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 移除 `GROUP BY FROM_UNIXTIME(complete_time, '%Y-%m-%d')`，返回包含 `complete_time` 的原始数据
  - 状态: ✅ 已完成

- [x] **实现应用层时区转换和分组**
  - 需求: 在应用层进行时区转换和日期分组
  - 预计时间: 1.5小时
  - 负责人: 
  - 说明: 创建 `memberDaysRawData` 结构体，在应用层按日期分组，使用 `decimal.Decimal` 进行精确计算，实现所有统计字段的聚合
  - 状态: ✅ 已完成

- [x] **修改 Service 层调用 CountMemberDays**
  - 需求: 在 Service 层传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 修改 `main/app/service/statistics.go` 中的调用（在 CountSaleDays 方法中），从 `ctx.GetCompanySetting().Timezone` 获取时区并传递
  - 状态: ✅ 已完成 

#### 2.6 修复 CountMemberPaymentDays 方法

- [x] **修改 CountMemberPaymentDays 方法签名** `main/app/repository/statistics.go`
  - 需求: 添加 `timezone string` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 状态: ✅ 已完成

- [x] **修改 CountMemberPaymentDays SQL 查询**
  - 需求: 移除 `FROM_UNIXTIME` 的日期分组，返回原始数据
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 移除 `GROUP BY FROM_UNIXTIME(smp.complete_time, '%Y-%m-%d')`，返回包含 `complete_time` 的原始数据
  - 状态: ✅ 已完成

- [x] **实现应用层时区转换和分组**
  - 需求: 在应用层进行时区转换和日期分组
  - 预计时间: 1.5小时
  - 负责人: 
  - 说明: 创建 `memberPaymentDaysRawData` 结构体，在应用层按支付方式 + 日期分组，使用 `decimal.Decimal` 进行精确计算
  - 状态: ✅ 已完成

- [x] **修改 Service 层调用 CountMemberPaymentDays**
  - 需求: 在 Service 层传递时区参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 修改 `main/app/service/statistics.go` 中的调用，从 `ctx.GetCompanySetting().Timezone` 获取时区并传递
  - 状态: ✅ 已完成 

#### 2.7 修复 CountBusinessPaymentMethod 方法（如未完成）

- [x] **检查 CountBusinessPaymentMethod 是否已修复**
  - 需求: 确认是否已采用应用层时区转换方案
  - 预计时间: 0.5小时
  - 负责人: 
  - 状态: ✅ 已检查，**无需修复**
  - 说明: `CountBusinessPaymentMethod` 方法已经使用应用层时区转换方案，SQL 查询返回原始数据，在应用层使用 `utils.SetTimezone(req.Timezone).FormatUnixTime()` 进行时区转换和日期分组。详见 `countbusinesspaymentmethod-check.md`

- [ ] **如未修复，执行修复**
  - 状态: ⏭️ 跳过（已确认无需修复）
  - 需求: 参考其他方法的修复方式
  - 预计时间: 2小时
  - 负责人: 

#### 2.8 修复 CountRefundSummary 方法中的 FROM_UNIXTIME

- [x] **修复 CountRefundSummary 中的 orderCountQuery**
  - 需求: 移除 `GROUP BY FROM_UNIXTIME(sb.finish_time, ...)`，在应用层分组
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 该方法已部分采用应用层分组，但 orderCountQuery 仍使用 FROM_UNIXTIME。修复：返回原始数据（包含 finish_time 和 sale_bill_uuid），在应用层使用 map 去重统计每个日期的订单数量
  - 状态: ✅ 已完成

- [x] **修复 CountRefundSummary 中的 takeoutOrderCountQuery**
  - 需求: 移除 `GROUP BY FROM_UNIXTIME(accepted_time, ...)`，在应用层分组
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 修复：返回原始数据（包含 accepted_time 和 order_uuid），在应用层使用 map 去重统计每个日期的订单数量
  - 状态: ✅ 已完成 

### 3. 测试验证

- [ ] **时区转换单元测试**
  - 需求: 测试不同时区的时区转换准确性
  - 预计时间: 2小时
  - 负责人: 
  - 说明: 测试 `Asia/Shanghai`、`Asia/Tokyo`、`Asia/Bangkok` 等时区

- [x] **跨日期边界测试**
  - 需求: 测试跨日期边界时间（如 23:00-01:00）的订单统计
  - 预计时间: 2小时
  - 负责人: 
  - 说明: 验证订单被正确归类到对应的日期
  - 状态: ✅ 已完成

- [ ] **统计准确性测试**
  - 需求: 对比修复前后的统计结果
  - 预计时间: 3小时
  - 负责人: 
  - 说明: 使用测试数据验证统计结果的准确性

- [ ] **功能回归测试**
  - 需求: 确保所有统计功能正常工作
  - 预计时间: 2小时
  - 负责人: 
  - 说明: 测试所有统计接口，确保无回归

- [ ] **性能测试**
  - 需求: 对比修复前后的查询性能
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 验证应用层分组的性能影响在可接受范围内

### 4. 代码审查

- [x] **代码审查**
  - 需求: 通过 Code Review
  - 预计时间: 2小时
  - 负责人: 
  - 说明: 确保代码质量和一致性
  - 状态: ✅ 已完成
  - 审查文档: `code-review.md`

### 5. 文档更新

- [x] **更新代码注释**
  - 需求: 添加时区转换相关注释
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 说明时区转换的实现方式
  - 状态: ✅ 已完成

- [ ] **更新技术文档**
  - 需求: 记录时区转换的最佳实践
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 更新相关技术文档

### 6. 部署上线

- [ ] **发布到测试环境**
  - 需求: 部署到测试环境并验证
  - 预计时间: 1小时
  - 负责人: 

- [ ] **灰度发布**
  - 需求: 小范围商户灰度发布
  - 预计时间: 2小时
  - 负责人: 
  - 说明: 在小范围商户验证统计准确性

- [ ] **全量发布**
  - 需求: 全量发布到生产环境
  - 预计时间: 1小时
  - 负责人: 
  - 说明: 监控统计查询性能和准确性

---

## 📊 任务统计

- **总任务数**: 38
- **已完成**: 31
- **进行中**: 0
- **未开始**: 7
- **完成率**: 81.6%

**已完成任务清单**:
- ✅ 代码优化（6个方法的修复，共28个任务）
- ✅ CountBusinessPaymentMethod 检查（1个任务）
- ✅ 跨日期边界测试（1个任务）
- ✅ 代码审查（1个任务）
- ✅ 更新代码注释（1个任务）

**待完成任务**:
- ⏳ 时区转换单元测试
- ⏳ 统计准确性测试
- ⏳ 功能回归测试
- ⏳ 性能测试
- ⏳ 更新技术文档
- ⏳ 部署上线相关任务（3个任务）

---

## 📈 性能指标

| 指标       | 优化前 | 目标   | 当前   | 提升   |
| ---------- | ------ | ------ | ------ | ------ |
| 数据准确性 | 不准确 | 100%   | -      | -      |
| 跨时区支持 | 不支持 | 完全支持 | -      | -      |
| 查询响应时间 | -     | 保持或提升 | -      | -      |

---

## 🔗 相关链接

- 优化需求: `optimize.md`
- 优化方案: `solution.md`
- 相关 Bug: `bug-251226-001-statistics-from-unixtime-timezone-statistics-error`
- 相关 Spec: `story-shop-statistics-merchant-timezone-query` (已归档)
- 参考实现: `main/app/repository/statistics.go:CountBusinessSummary`、`CountRefundSummary`

---

## 📝 注意事项

1. **参考已有实现**：`CountBusinessSummary` 和 `CountRefundSummary` 已采用应用层时区转换方案，可作为参考
2. **时区工具类**：使用 `utils.SetTimezone(timezone).FormatUnixTime()` 进行时区转换
3. **精确计算**：使用 `decimal.Decimal` 进行金额计算
4. **测试重点**：重点测试跨日期边界和不同时区的场景
5. **代码一致性**：确保所有方法采用统一的时区转换方案
