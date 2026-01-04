# 统计报表按商户时区查询 任务分解

> 本文档定义统计报表按商户时区查询的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📌 第一阶段实施指南

**优先实施收银机端接口**，详细任务清单请参考：
- `phase1-cashier-tasks.md` - 收银机端接口实施任务清单（15 个任务）

**实施顺序**：
1. Phase 1: 时区工具类扩展（3 个任务）
2. Phase 2: DTO 扩展（4 个任务）
3. Phase 3: Repository 层查询优化（2 个任务）
4. Phase 4: Service 层时区转换（6 个任务）
5. Phase 5: API 层参数处理（6 个任务）

完成第一阶段后，再继续实施新管理后台和旧管理后台接口。

## 📊 进度总览

**总任务数**: 45  
**已完成**: 15+  
**进行中**: -  
**完成率**: 约 35%（核心功能已完成）

## 🎯 分步实施计划

### 第一阶段：收银机端接口（优先实施）

**目标**: 完成收银机端 5 个统计接口的时区转换改造

**涉及接口**:
- `/cashier/statistics/printer` - 打印统计数据
- `/cashier/statistics/business` - 统计营业数据
- `/cashier/statistics/payment_method` - 统计支付方式
- `/cashier/statistics/product_category` - 统计商品分类
- `/cashier/statistics/product` - 统计商品

**预计工作量**: 2-3 天

**任务清单**: 见下方 "第一阶段：收银机端实施" 章节

### 第二阶段：新管理后台接口（后续实施）

**目标**: 完成新管理后台 27 个统计接口的时区转换改造

**预计工作量**: 4-5 天

### 第三阶段：旧管理后台接口（后续实施）

**目标**: 完成旧管理后台 12 个统计接口的时区转换改造

**预计工作量**: 2-3 天

---

## Phase 1: 时区工具类扩展

- [x] 1.1 添加 FormatDateTimeToUnix 方法

  - File: `main/pkg/utils/time.go`
  - Purpose: 支持解析 "YYYY-MM-DD HH:mm:ss" 格式的日期时间字符串，转换为时间戳（使用商户时区）
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有方法 `FormatTimeToUnix` (只支持 YYYY-MM-DD)，参考 `FormatTimeToTime` 方法
  - Prompt: Role: Go Developer specializing in time utilities | Task: 在 Timezone 类型上添加 FormatDateTimeToUnix 方法，支持解析 "YYYY-MM-DD HH:mm:ss" 和 "YYYY-MM-DD" 两种格式 | Context: 使用 time.ParseInLocation 解析，使用商户时区，返回 Unix 时间戳 | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理使用 errors.WithMessage | Success: 方法创建成功，支持两种格式，错误处理正确

- [x] 1.2 添加 FormatDateTimeToTime 方法

  - File: `main/pkg/utils/time.go`
  - Purpose: 支持解析日期时间字符串，转换为 time.Time 对象（使用商户时区）
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有方法 `FormatTimeToTime`，Task 1.1 的实现
  - Prompt: Role: Go Developer specializing in time utilities | Task: 在 Timezone 类型上添加 FormatDateTimeToTime 方法，支持解析日期时间字符串 | Context: 使用 time.ParseInLocation 解析，使用商户时区 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法创建成功，支持两种格式

- [ ] 1.3 添加日期分组工具方法（可选）

  - File: `main/pkg/utils/time.go`
  - Purpose: 提供日期分组相关的工具方法，便于 Service 层使用
  - Requirements: Requirement 4
  - Leverage: 现有 FormatUnixTime 方法
  - Prompt: Role: Go Developer specializing in timezone conversion | Task: 添加日期分组相关的工具方法，如 GroupByDate 等 | Context: 将 UTC 时间戳列表按商户时区的日期分组 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 工具方法创建成功

- [x] 1.4 编写时区工具类单元测试

  - File: `main/pkg/utils/time_test.go`
  - Purpose: 确保时区工具类方法正确
  - Requirements: Requirement 1, Requirement 2, Requirement 4
  - Leverage: 现有测试: `main/pkg/utils/time_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为新增的时区工具方法编写单元测试，覆盖率 100% | Context: 测试 FormatDateTimeToUnix、FormatDateTimeToTime、TimezoneToMySQLOffset，覆盖多种时区（UTC+7, UTC+8, UTC+9），测试错误情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

---

## Phase 2: DTO 扩展

- [x] 2.1 扩展 BusinessDataCountReq 添加日期时间字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加 query_start_date 和 query_end_date 字段，支持接收日期时间格式 "YYYY-MM-DD HH:mm:ss"
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有字段 `QueryStartTime`、`QueryEndTime`，参考 `full_reduction_activity_req.go` 中的日期字符串处理
  - Prompt: Role: Go Developer | Task: 在 BusinessDataCountReq 结构体中添加 QueryStartDate 和 QueryEndDate 字段（string 类型） | Context: 字段标签使用 form 和 json，字段注释说明格式要求 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，标签和注释正确

- [x] 2.2 扩展 BusinessDataCountReq.GetParam 方法

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 扩展 GetParam 方法，优先处理 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有 GetParam 方法，Task 1.1 的 FormatDateTimeToUnix 方法
  - Prompt: Role: Go Developer | Task: 扩展 BusinessDataCountReq.GetParam 方法，添加日期时间字符串参数处理逻辑 | Context: 参数优先级：1) query_start_date + query_end_date, 2) query_start_time + query_end_time, 3) time_type | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理使用 errors.WithMessage | Success: 方法扩展成功，参数优先级正确，错误处理完善

- [x] 2.3 扩展 BusinessDataPrinterReq 添加日期时间字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加 query_start_date 和 query_end_date 字段
  - Requirements: Requirement 2
  - Leverage: Task 2.1 的实现
  - Success: 字段添加成功

- [x] 2.4 扩展 BusinessDataPrinterReq.GetParam 方法

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 扩展 GetParam 方法支持日期时间字符串
  - Requirements: Requirement 2
  - Leverage: Task 2.2 的实现
  - Success: 方法扩展成功

- [x] 2.5 扩展 StatisticsSummaryReq 添加日期时间字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加 query_start_date 和 query_end_date 字段
  - Requirements: Requirement 1
  - Leverage: Task 2.1 的实现
  - Success: 字段添加成功

- [x] 2.6 扩展 StatisticsPaymentMethodReq 添加日期时间字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加 query_start_date 和 query_end_date 字段
  - Requirements: Requirement 1
  - Leverage: Task 2.1 的实现
  - Success: 字段添加成功

- [x] 2.7 扩展 BusinessTimePeriodReq 添加日期时间字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加 query_start_date 和 query_end_date 字段
  - Requirements: Requirement 1
  - Leverage: Task 2.1 的实现
  - Success: 字段添加成功

- [x] 2.8 扩展其他统计请求 DTO（如 KitchenEfficiencyAnalysisReq 等）

  - File: `main/app/dto/req/statistics.go`, `main/app/dto/req/statistics_channel_req.go`, `main/app/dto/req/statistics_user_analysis_req.go`, `main/app/dto/req/member_order.go`
  - Purpose: 为所有统计请求 DTO 添加日期时间字段
  - Requirements: Requirement 1
  - Leverage: Task 2.1 的实现
  - Success: 所有 DTO 扩展完成

---

## Phase 3: Repository 层查询优化

- [ ] 3.1 修改 IStatisticsRepo 接口移除时区参数

  - File: `main/app/repository/i_statistics_repo.go`
  - Purpose: Repository 层只接收 UTC 时间戳，不处理时区
  - Requirements: Requirement 4
  - Leverage: 现有接口定义
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 修改 IStatisticsRepo 接口，移除 timezone 参数，Repository 层只处理 UTC 时间戳 | Context: 方法包括 CountBusinessSummary、CountBusinessPaymentMethod 等，只接收时间戳范围 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口修改成功，所有方法移除时区参数

- [ ] 3.2 修改 CountBusinessSummary 实现使用 UTC 时间戳查询

  - File: `main/app/repository/statistics_repo.go`
  - Purpose: 修改 SQL 查询使用 UTC 时间戳，移除 FROM_UNIXTIME 和 CONVERT_TZ
  - Requirements: Requirement 4
  - Leverage: 现有查询实现
  - Prompt: Role: Go Developer with SQL expertise | Task: 修改 CountBusinessSummary 方法的 SQL 查询，移除 FROM_UNIXTIME 和 CONVERT_TZ，直接使用 UTC 时间戳查询 | Context: 查询返回原始数据（包含 create_time 等 UTC 时间戳字段），时区转换在 Service 层完成 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用参数化查询防止 SQL 注入 | Success: SQL 查询修改成功，返回 UTC 时间戳数据

- [ ] 3.3 修改 CountBusinessPaymentMethod 实现使用 UTC 时间戳查询

  - File: `main/app/repository/statistics_repo.go`
  - Purpose: 修改 SQL 查询使用 UTC 时间戳
  - Requirements: Requirement 4
  - Leverage: Task 3.2 的实现
  - Success: SQL 查询修改成功

- [ ] 3.4 修复所有使用 FROM_UNIXTIME 的统计查询（移除时区转换）

  - File: `main/app/repository/statistics_repo.go` 及相关文件
  - Purpose: 统一移除所有使用 FROM_UNIXTIME 的查询中的时区转换
  - Requirements: Requirement 4
  - Leverage: Task 3.2 的实现，使用 grep 搜索 FROM_UNIXTIME
  - Command: `grep -r "FROM_UNIXTIME" main/app/repository/`
  - Success: 所有查询修复完成，返回 UTC 时间戳数据

- [ ] 3.5 编写 Repository 单元测试

  - File: `main/app/repository/statistics_repo_test.go`
  - Purpose: 确保 Repository 查询返回 UTC 时间戳数据正确
  - Requirements: Requirement 4
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为统计 Repository 编写单元测试，覆盖率 ≥ 80% | Context: 测试 UTC 时间戳查询正确性，测试时间范围查询 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

---

## Phase 4: Service 层时区转换和日期分组

- [ ] 4.1 修改 CountBusiness 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换和日期分组
  - Requirements: Requirement 1
  - Leverage: 现有方法实现，`ctx.GetCompanySetting().Timezone`，Task 1.1-1.2 的工具方法
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 CountBusiness 方法，在应用层完成时区转换和日期分组 | Context: 1) 使用 ctx.GetCompanySetting().Timezone 获取时区，2) 调用 req.GetParam(timezone, openingHours) 转换时间范围为 UTC 时间戳，3) 调用 Repository 查询数据，4) 将查询结果中的 UTC 时间戳转换为商户时区，进行日期分组 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改成功，时区转换和日期分组正确

- [x] 4.2 修改 CountPaymentMethod 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换
  - Requirements: Requirement 1, Requirement 2（收银机端使用）
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功

- [x] 4.3 修改 CountProductCategory 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换
  - Requirements: Requirement 1, Requirement 2（收银机端使用）
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功

- [x] 4.4 修改 CountProduct 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换
  - Requirements: Requirement 1, Requirement 2（收银机端使用）
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功

- [x] 4.4.1 修改 Printer 方法实现应用层时区转换（收银机端）

  - File: `main/app/service/business_service.go`
  - Purpose: 打印功能在应用层完成时区转换
  - Requirements: Requirement 2（收银机端）
  - Leverage: Task 4.1 的实现，Task 2.3-2.4 的 GetParam 方法
  - Success: 方法修改成功，时区转换正确

- [x] 4.5 修改 CountBusinessSummary 方法实现应用层时区转换和日期分组

  - File: `main/app/service/statistics.go`
  - Purpose: 在应用层完成时区转换和日期分组（重要：需要按日期分组统计）
  - Requirements: Requirement 1
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功，日期分组正确

- [x] 4.6 修改 CountBusinessPaymentMethod 方法实现应用层时区转换和日期分组

  - File: `main/app/service/statistics.go`
  - Purpose: 在应用层完成时区转换和日期分组
  - Requirements: Requirement 1
  - Leverage: Task 4.5 的实现
  - Success: 方法修改成功

- [x] 4.7 修改 CountBusinessTimePeriod 方法实现应用层时区转换

  - File: `main/app/service/statistics.go`
  - Purpose: 在应用层完成时区转换
  - Requirements: Requirement 1
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功

- [x] 4.8 修改所有其他统计 Service 方法实现应用层时区转换

  - File: `main/app/service/business.go`, `main/app/service/statistics.go`, `main/app/service/recharge_order.go`, `main/app/modules/printer/service/printer_log.go`
  - Purpose: 统一所有统计方法在应用层完成时区转换
  - Requirements: Requirement 1, Requirement 2
  - Leverage: Task 4.1 的实现
  - Success: 所有方法修改完成

- [ ] 4.9 编写 Service 单元测试

  - File: `main/app/service/business_service_test.go`
  - Purpose: 确保 Service 时区转换和日期分组正确
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有测试: `main/app/service/*_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为统计 Service 编写单元测试，覆盖率 ≥ 70% | Context: 测试时区转换正确性，测试日期分组正确性，测试多种时区场景（UTC+7, UTC+8, UTC+9） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 5: API 层参数处理

### 第一阶段：收银机端实施（优先）

- [x] 5.1.1 修改收银机 Printer API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机打印接口支持 query_start_date 和 query_end_date 参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: 现有 API 实现，Task 2.3-2.4 的 GetParam 方法
  - Success: API 修改成功，参数接收正确，时区转换正确

- [x] 5.1.2 修改收银机 CountBusiness API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计营业数据接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1.1 的实现
  - Success: API 修改成功

- [x] 5.1.3 修改收银机 CountPaymentMethod API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计支付方式接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1.1 的实现
  - Success: API 修改成功

- [x] 5.1.4 修改收银机 CountProductCategory API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计商品分类接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1.1 的实现
  - Success: API 修改成功

- [x] 5.1.5 修改收银机 CountProduct API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计商品接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1.1 的实现
  - Success: API 修改成功

- [x] 5.1.6 修改收银机 Order List API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_order.go`
  - Purpose: 收银机订单列表接口支持 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 2
  - Success: API 修改成功（Repository 层处理）

- [x] 5.1.7 修改收银机 RechargeOrder List API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_recharge_order.go`
  - Purpose: 收银机充值订单列表接口支持 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 2
  - Success: API 修改成功（Service 层处理）

- [x] 5.1.8 修改收银机 Printer List API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_printer.go`
  - Purpose: 收银机打印列表接口支持 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 2
  - Success: API 修改成功（Service 层处理）

- [ ] 5.1.6 编写收银机 API 集成测试

  - File: `main/app/api/v1/cashier/cashier_statistics_test.go`
  - Purpose: 测试收银机统计 API 接口参数处理和时区转换
  - Requirements: Requirement 2
  - Leverage: 现有测试: `main/app/api/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为收银机统计 API 编写集成测试 | Context: 测试 query_start_date 和 query_end_date 参数，测试时区转换，测试多种时区场景（UTC+7, UTC+8, UTC+9） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

### 第二阶段：新管理后台实施（后续）

- [x] 5.2.1 修改新管理后台 CountBusiness API 支持新参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: API 层接收 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 1
  - Leverage: Task 5.1.1 的实现（收银机端已完成）
  - Success: API 修改成功，参数接收正确（Service 层通过 GetParam 处理）

- [x] 5.2.2 修改新管理后台 CountPaymentMethod API 支持新参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 支持新参数
  - Requirements: Requirement 1
  - Leverage: Task 5.2.1 的实现
  - Success: API 修改成功（Service 层通过 GetParam 处理）

- [x] 5.2.3 修改所有新管理后台统计 API 支持新参数（27 个接口）

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 统一所有统计接口支持新参数
  - Requirements: Requirement 1
  - Leverage: Task 5.2.1 的实现
  - Success: 所有 API 修改完成（Service 层统一处理）

- [x] 5.2.4 修改新管理后台 MemberOrder List API 支持新参数

  - File: `main/app/api/v1/shop/shop_member_order.go`
  - Purpose: 会员订单列表接口支持 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 1
  - Success: API 修改成功（DTO 层 GetTimeFilterParams 方法处理）

- [x] 5.2.5 修改新管理后台 RechargeOrder List API 支持新参数

  - File: `main/app/api/v1/shop/shop_recharge_order.go`
  - Purpose: 充值订单列表接口支持 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 1
  - Success: API 修改成功（Service 层已处理）

- [x] 5.2.6 修改新管理后台 Order List API 支持新参数

  - File: `main/app/api/v1/shop/shop_order.go`
  - Purpose: 订单列表接口支持 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 1
  - Success: API 修改成功（Repository 层已处理）

- [ ] 5.5 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_statistics_test.go`
  - Purpose: 测试 API 接口参数处理
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有测试: `main/app/api/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为统计 API 编写集成测试 | Context: 测试 query_start_date 和 query_end_date 参数，测试参数优先级，测试时区转换 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 6: PHP Admin 模块时区处理

- [x] 6.1 创建 PHP 时区转换工具类

  - File: `admin/app/common/util/TimezoneUtil.php`
  - Purpose: 统一处理 PHP 时区转换逻辑
  - Requirements: Requirement 3
  - Status: 无需调整（已确认）

- [x] 6.2 修改 Sales Controller 使用商户时区

  - File: `admin/app/shop/controller/statistics/Sales.php`
  - Purpose: 修复 getDays 方法使用商户时区
  - Requirements: Requirement 3
  - Status: 无需调整（已确认）

- [x] 6.3 修改 Access Controller 使用商户时区

  - File: `admin/app/shop/controller/statistics/Access.php`
  - Purpose: 修复 getDays 方法使用商户时区
  - Requirements: Requirement 3
  - Status: 无需调整（已确认）

- [x] 6.4 修改 User Controller 使用商户时区

  - File: `admin/app/shop/controller/statistics/User.php`
  - Purpose: 修复 getDays 方法使用商户时区
  - Requirements: Requirement 3
  - Status: 无需调整（已确认）

- [x] 6.5 修改 Supplier Controller 使用商户时区

  - File: `admin/app/shop/controller/statistics/Supplier.php`
  - Purpose: 修复 getDays 方法使用商户时区
  - Requirements: Requirement 3
  - Status: 无需调整（已确认）

- [x] 6.6 修改 Order Controller 使用商户时区

  - File: `admin/app/shop/controller/statistics/Order.php`
  - Purpose: 修复 getDays 方法使用商户时区
  - Requirements: Requirement 3
  - Status: 无需调整（已确认）

- [x] 6.7 修改 Product Controller 使用商户时区

  - File: `admin/app/shop/controller/statistics/Product.php`
  - Purpose: 修复 getDays 方法使用商户时区
  - Requirements: Requirement 3
  - Status: 无需调整（已确认）

---

## Phase 7: 前端时区显示优化

- [ ] 7.1 前端显示商户时区信息

  - File: `admin/views/{admin|shop}/pages/statistics/*.vue`
  - Purpose: 在统计页面显示商户时区信息
  - Requirements: Requirement 6
  - Leverage: 现有页面组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在统计页面显示商户时区信息（如"商户时区：Asia/Bangkok (UTC+7)"） | Context: 从 API 响应获取时区信息，使用 Element Plus 组件显示 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 时区信息显示正确

- [ ] 7.2 时间选择器默认使用商户时区

  - File: `admin/views/{admin|shop}/pages/statistics/*.vue`
  - Purpose: 时间选择器使用商户时区
  - Requirements: Requirement 6
  - Leverage: Element Plus DatePicker 组件
  - Success: 时间选择器配置正确

- [ ] 7.3 统计图表时间轴使用商户时区

  - File: `admin/views/{admin|shop}/pages/statistics/*.vue`
  - Purpose: 图表时间轴显示商户时区时间
  - Requirements: Requirement 6
  - Leverage: 图表库（如 ECharts）
  - Success: 图表时间轴正确

---

## Phase 8: 测试和优化

- [ ] 8.1 集成测试（多时区场景）

  - File: `test/integration/statistics_timezone_test.go`
  - Purpose: 测试端到端时区转换流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现多时区场景的集成测试 | Context: 测试 UTC+7, UTC+8, UTC+9 时区，测试跨时区查询 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 8.2 性能测试

  - File: -
  - Purpose: 确保时区转换不影响性能
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 本地响应时间 < 200ms

- [ ] 8.3 数据库查询优化

  - File: `main/app/repository/statistics_repo.go`
  - Purpose: 优化 SQL 查询性能
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 50ms

- [ ] 8.4 文档更新

  - File: `docs/shared/api/statistics_api.md`, `CHANGELOG.md`
  - Purpose: 更新 API 文档
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-statistics-merchant-timezone-query/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-statistics-merchant-timezone-query/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-statistics-merchant-timezone-query/tasks.md
```

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-30  
**维护者**: 后端开发组

