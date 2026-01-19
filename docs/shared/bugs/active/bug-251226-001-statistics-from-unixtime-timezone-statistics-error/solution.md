# Bug-251226-001 修复方案

## 问题概述

统计功能中使用了 MySQL 的 `FROM_UNIXTIME` 函数进行日期格式化，但该函数使用的是 MySQL 服务器的时区，而不是业务时区，导致统计结果错误。

**影响范围**：
- 综合运营统计 (`CountBusinessSummary`)
- 支付方式统计 (`CountBusinessPaymentMethod`)
- 其他 14 处使用 `FROM_UNIXTIME` 的统计查询

## 根本原因

1. **MySQL `FROM_UNIXTIME` 函数的时区问题**
   - `FROM_UNIXTIME(timestamp)` 使用 MySQL 服务器的时区进行转换
   - 如果服务器时区与业务时区不一致，会导致日期分组错误

2. **代码设计缺陷**
   - Service 层已获取业务时区：`timezone := ctx.GetCompanySetting().Timezone`
   - Repository 层的方法未接收时区参数
   - SQL 查询中直接使用 `FROM_UNIXTIME`，未考虑时区转换

3. **时区不一致的场景**
   - 跨时区商户（如日本、泰国、土耳其等）
   - MySQL 服务器时区设置为 UTC
   - 业务时区设置为 `Asia/Shanghai`、`Asia/Tokyo` 等

## 修复方案

### 方案选择

**选项 1: 使用 MySQL `CONVERT_TZ` 函数**
- 优点: 
  - 在数据库层完成时区转换，性能较好
  - 代码改动相对较小
  - 利用 MySQL 内置时区转换能力
- 缺点: 
  - 需要将时区名称（如 `Asia/Shanghai`）转换为偏移量（如 `+08:00`）
  - 需要处理夏令时等复杂情况
- 风险: 
  - 时区名称到偏移量的转换可能不准确（夏令时问题）
  - 需要确保 MySQL 时区数据完整

**选项 2: 在应用层转换时区**
- 优点: 
  - 完全控制时区转换逻辑
  - 不依赖 MySQL 时区数据
  - 可以处理复杂的时区规则
- 缺点: 
  - 需要在应用层计算日期分组，SQL 查询逻辑复杂
  - 性能可能不如数据库层转换
  - 代码改动较大
- 风险: 
  - 需要重构大量 SQL 查询
  - 可能引入新的 Bug

**选项 3: 设置 MySQL Session Timezone + `CONVERT_TZ`**
- 优点: 
  - 结合两种方案的优点
  - 可以在连接级别设置时区
  - 使用 `CONVERT_TZ` 时更简洁
- 缺点: 
  - 需要修改数据库连接配置
  - 可能影响其他查询
- 风险: 
  - 需要确保每个连接都正确设置时区
  - 连接池复用可能导致时区混乱

**✅ 最终选择: 选项 1（使用 MySQL `CONVERT_TZ` 函数）**

理由: 
- 性能最优，在数据库层完成转换
- 代码改动相对较小，只需修改 SQL 查询
- 项目已有 `utils.Timezone` 工具类，可以复用
- 对于统计查询，时区转换的准确性更重要，性能影响可接受

### 实施步骤

1. **创建时区转换工具函数**
   - 在 `main/pkg/utils/time.go` 中添加 `TimezoneToMySQLOffset` 函数
   - 将时区名称（如 `Asia/Shanghai`）转换为 MySQL 偏移量（如 `+08:00`）

2. **修改 Repository 层方法签名**
   - `CountBusinessSummary` 添加 `timezone` 参数
   - `CountBusinessPaymentMethod` 添加 `timezone` 参数
   - 其他使用 `FROM_UNIXTIME` 的方法也需要添加时区参数

3. **修改 SQL 查询**
   - 将 `FROM_UNIXTIME(timestamp, format)` 
   - 替换为 `DATE_FORMAT(CONVERT_TZ(FROM_UNIXTIME(timestamp), '+00:00', offset), format)`
   - 其中 `offset` 为业务时区的 MySQL 偏移量

4. **修改 Service 层调用**
   - 在调用 Repository 方法时传递时区参数
   - 从 `ctx.GetCompanySetting().Timezone` 获取时区

5. **全面检查并修复**
   - 搜索所有 `FROM_UNIXTIME` 的使用
   - 逐一修复，确保都使用正确的时区转换

6. **编写单元测试**
   - 测试时区转换的正确性
   - 测试跨日期边界的统计准确性

### 技术方案

#### 数据结构变更

无需变更数据库结构。

#### 代码修改

**1. 添加时区转换工具函数**

```go
// main/pkg/utils/time.go

// TimezoneToMySQLOffset 将时区名称转换为 MySQL 时区偏移量
// 例如: "Asia/Shanghai" -> "+08:00"
func TimezoneToMySQLOffset(timezone string) string {
    loc, err := time.LoadLocation(timezone)
    if err != nil {
        return "+00:00" // 默认 UTC
    }
    
    // 获取当前时间在该时区的偏移量
    now := time.Now()
    _, offset := now.In(loc).Zone()
    
    // 转换为 MySQL 格式: +HH:MM 或 -HH:MM
    hours := offset / 3600
    minutes := (offset % 3600) / 60
    
    sign := "+"
    if hours < 0 {
        sign = "-"
        hours = -hours
    }
    
    return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}
```

**2. 修改 Repository 方法签名**

```go
// main/app/repository/statistics.go

// CountBusinessSummaryReq 添加 Timezone 字段
type CountBusinessSummaryReq struct {
    StartTime                    int64    // 查询开始时间戳
    EndTime                      int64    // 查询结束时间戳
    Cycle                        int      // 周期: 0=按日、1=按月
    PageNo, PageSize             int      // 页码, 每页大小
    Timezone                     string   // 业务时区，如 "Asia/Shanghai"
}

// CountBusinessPaymentMethodReq 添加 Timezone 字段
type CountBusinessPaymentMethodReq struct {
    // ... 其他字段
    Timezone                     string   // 业务时区，如 "Asia/Shanghai"
}
```

**3. 修改 SQL 查询**

```go
// 修改前
FROM_UNIXTIME(sb.finish_time, '%Y-%m-%d')

// 修改后
DATE_FORMAT(CONVERT_TZ(FROM_UNIXTIME(sb.finish_time), '+00:00', ?), '%Y-%m-%d')
```

**4. 修改 Service 层调用**

```go
// main/app/service/statistics.go

func (s *statisticsSrv) CountBusinessSummary(ctx context.Context, req req.StatisticsSummaryReq) StatisticsSummaryResp {
    timezone := ctx.GetCompanySetting().Timezone
    
    total, dataList := statisticsRepo.CountBusinessSummary(repository.CountBusinessSummaryReq{
        StartTime: req.QueryStartTime,
        EndTime:   req.QueryEndTime,
        Cycle:     req.Cycle,
        PageNo:    utils.IfInt(req.PageNo > 0, req.PageNo, 1),
        PageSize:  utils.IfInt(req.PageSize > 0, req.PageSize, 10),
        Timezone:  timezone, // 传递时区参数
    })
    // ...
}
```

#### 配置调整

无需配置调整。

## 影响分析

### 兼容性

- ✅ **向后兼容**: 修改后的代码仍然支持原有的时区设置
- ✅ **数据兼容**: 不涉及数据迁移，只修改查询逻辑
- ⚠️ **API 兼容**: Service 层接口不变，只修改内部实现

### 性能影响

- ✅ **性能优化**: `CONVERT_TZ` 在数据库层执行，性能优于应用层转换
- ⚠️ **查询复杂度**: SQL 查询稍微复杂，但影响可接受
- ✅ **索引使用**: 不影响时间戳字段的索引使用

### 安全风险

- ✅ **无安全风险**: 时区转换是纯计算操作，不涉及用户输入

## 测试计划

### 单元测试

1. **时区转换工具函数测试**
   - 测试常见时区：`Asia/Shanghai`, `Asia/Tokyo`, `Asia/Bangkok`, `Europe/Istanbul`
   - 测试边界情况：UTC, 负偏移时区
   - 测试错误处理：无效时区名称

2. **Repository 层测试**
   - 测试 `CountBusinessSummary` 在不同时区下的日期分组
   - 测试 `CountBusinessPaymentMethod` 在不同时区下的日期分组
   - 测试跨日期边界的统计准确性

### 集成测试

1. **端到端测试**
   - 设置商户时区为 `Asia/Shanghai`
   - 在跨日期边界时间创建订单（如 23:00-01:00）
   - 查询统计报表，验证日期分组正确

2. **多时区测试**
   - 测试不同时区商户的统计准确性
   - 验证时区转换不影响其他功能

### 手动测试

1. **功能测试**
   - 在测试环境创建不同时区的商户
   - 创建测试订单，覆盖跨日期边界场景
   - 验证统计报表的日期分组正确性

2. **回归测试**
   - 验证现有统计功能不受影响
   - 验证其他使用 `FROM_UNIXTIME` 的功能正常

## 上线计划

### 发布时间

- **开发时间**: 预计 1-2 天
- **测试时间**: 预计 1 天
- **发布时间**: 待定

### 回滚方案

1. **代码回滚**: 如果发现问题，立即回滚到上一个版本
2. **数据回滚**: 无需数据回滚，只修改查询逻辑
3. **监控指标**: 监控统计查询的响应时间和错误率

### 监控指标

- 统计查询的响应时间
- 统计查询的错误率
- 时区转换的准确性（通过日志验证）

## 预防措施

1. **代码规范**
   - 在代码规范中明确：**禁止直接使用 `FROM_UNIXTIME`，必须使用时区转换**
   - 添加代码审查检查点

2. **工具函数**
   - 创建统一的时区转换工具函数
   - 封装 `FROM_UNIXTIME` 的时区转换逻辑

3. **单元测试**
   - 为时区相关功能编写单元测试
   - 在 CI/CD 中自动运行测试

4. **文档更新**
   - 更新开发文档，说明时区处理规范
   - 完善故障排查指南

5. **代码审查**
   - 在代码审查中重点关注时区处理
   - 确保所有时间相关查询都使用正确的时区

## 相关链接

- Bug 报告: `bug.md`
- 任务清单: `tasks.md`
- 相关文件: `main/app/repository/statistics.go`
- 相关文件: `main/app/service/statistics.go`
- 排查文档: `docs/shared/troubleshooting/mysql-fromunixtime-timezone.md`

