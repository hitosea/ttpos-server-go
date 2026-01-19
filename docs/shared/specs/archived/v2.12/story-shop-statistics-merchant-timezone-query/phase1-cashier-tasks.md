# 第一阶段：收银机端接口实施任务清单

> 本文档列出收银机端接口的详细实施任务，优先完成。

## 📋 收银机端接口列表

1. `/cashier/statistics/printer` - 打印统计数据
2. `/cashier/statistics/business` - 统计营业数据
3. `/cashier/statistics/payment_method` - 统计支付方式
4. `/cashier/statistics/product_category` - 统计商品分类
5. `/cashier/statistics/product` - 统计商品

## 📊 进度总览

**总任务数**: 15  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 时区工具类扩展（基础）

- [x] 1.1 添加 FormatDateTimeToUnix 方法

  - File: `main/pkg/utils/time.go`
  - Purpose: 支持解析 "YYYY-MM-DD HH:mm:ss" 格式的日期时间字符串，转换为时间戳（使用商户时区）
  - Requirements: Requirement 2
  - Leverage: 现有方法 `FormatTimeToUnix` (只支持 YYYY-MM-DD)，参考 `FormatTimeToTime` 方法
  - Prompt: Role: Go Developer specializing in time utilities | Task: 在 Timezone 类型上添加 FormatDateTimeToUnix 方法，支持解析 "YYYY-MM-DD HH:mm:ss" 和 "YYYY-MM-DD" 两种格式 | Context: 使用 time.ParseInLocation 解析，使用商户时区，返回 Unix 时间戳 | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理使用 errors.WithMessage | Success: 方法创建成功，支持两种格式，错误处理正确

- [x] 1.2 添加 FormatDateTimeToTime 方法

  - File: `main/pkg/utils/time.go`
  - Purpose: 支持解析日期时间字符串，转换为 time.Time 对象（使用商户时区）
  - Requirements: Requirement 2
  - Leverage: 现有方法 `FormatTimeToTime`，Task 1.1 的实现
  - Prompt: Role: Go Developer specializing in time utilities | Task: 在 Timezone 类型上添加 FormatDateTimeToTime 方法，支持解析日期时间字符串 | Context: 使用 time.ParseInLocation 解析，使用商户时区 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法创建成功，支持两种格式

- [x] 1.3 编写时区工具类单元测试

  - File: `main/pkg/utils/time_test.go`
  - Purpose: 确保时区工具类方法正确
  - Requirements: Requirement 2
  - Leverage: 现有测试: `main/pkg/utils/time_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为新增的时区工具方法编写单元测试，覆盖率 100% | Context: 测试 FormatDateTimeToUnix、FormatDateTimeToTime，覆盖多种时区（UTC+7, UTC+8, UTC+9），测试错误情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

---

## Phase 2: DTO 扩展（收银机端相关）

- [ ] 2.1 扩展 BusinessDataCountReq 添加日期时间字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加 query_start_date 和 query_end_date 字段，支持接收日期时间格式 "YYYY-MM-DD HH:mm:ss"
  - Requirements: Requirement 2
  - Leverage: 现有字段 `QueryStartTime`、`QueryEndTime`，参考 `full_reduction_activity_req.go` 中的日期字符串处理
  - Prompt: Role: Go Developer | Task: 在 BusinessDataCountReq 结构体中添加 QueryStartDate 和 QueryEndDate 字段（string 类型） | Context: 字段标签使用 form 和 json，字段注释说明格式要求 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，标签和注释正确

- [ ] 2.2 扩展 BusinessDataCountReq.GetParam 方法

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 扩展 GetParam 方法，优先处理 query_start_date 和 query_end_date 参数
  - Requirements: Requirement 2
  - Leverage: 现有 GetParam 方法，Task 1.1 的 FormatDateTimeToUnix 方法
  - Prompt: Role: Go Developer | Task: 扩展 BusinessDataCountReq.GetParam 方法，添加日期时间字符串参数处理逻辑 | Context: 参数优先级：1) query_start_date + query_end_date, 2) query_start_time + query_end_time, 3) time_type | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理使用 errors.WithMessage | Success: 方法扩展成功，参数优先级正确，错误处理完善

- [ ] 2.3 扩展 BusinessDataPrinterReq 添加日期时间字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加 query_start_date 和 query_end_date 字段
  - Requirements: Requirement 2
  - Leverage: Task 2.1 的实现
  - Success: 字段添加成功

- [ ] 2.4 扩展 BusinessDataPrinterReq.GetParam 方法

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 扩展 GetParam 方法支持日期时间字符串
  - Requirements: Requirement 2
  - Leverage: Task 2.2 的实现
  - Success: 方法扩展成功

---

## Phase 3: Repository 层查询优化（收银机端相关）

- [ ] 3.1 修改 IStatisticsRepo 接口移除时区参数（收银机端相关方法）

  - File: `main/app/repository/i_statistics_repo.go`
  - Purpose: Repository 层只接收 UTC 时间戳，不处理时区（收银机端使用的统计方法）
  - Requirements: Requirement 2
  - Leverage: 现有接口定义
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 修改 IStatisticsRepo 接口中收银机端使用的方法，移除 timezone 参数 | Context: 方法包括 CountBusiness、CountPaymentMethod、CountProductCategory、CountProduct 等，只接收时间戳范围 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口修改成功，收银机端方法移除时区参数

- [ ] 3.2 修改收银机端相关 Repository 实现使用 UTC 时间戳查询

  - File: `main/app/repository/statistics_repo.go` 或相关文件
  - Purpose: 修改 SQL 查询使用 UTC 时间戳，移除 FROM_UNIXTIME 和 CONVERT_TZ
  - Requirements: Requirement 2
  - Leverage: 现有查询实现
  - Prompt: Role: Go Developer with SQL expertise | Task: 修改收银机端使用的统计查询方法，移除 FROM_UNIXTIME 和 CONVERT_TZ，直接使用 UTC 时间戳查询 | Context: 查询返回原始数据（包含 create_time 等 UTC 时间戳字段），时区转换在 Service 层完成 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用参数化查询防止 SQL 注入 | Success: SQL 查询修改成功，返回 UTC 时间戳数据

---

## Phase 4: Service 层时区转换（收银机端相关）

- [ ] 4.1 修改 CountBusiness 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换和日期分组
  - Requirements: Requirement 2
  - Leverage: 现有方法实现，`ctx.GetCompanySetting().Timezone`，Task 1.1-1.2 的工具方法
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 CountBusiness 方法，在应用层完成时区转换 | Context: 1) 使用 ctx.GetCompanySetting().Timezone 获取时区，2) 调用 req.GetParam(timezone, openingHours) 转换时间范围为 UTC 时间戳，3) 调用 Repository 查询数据，4) 将查询结果中的 UTC 时间戳转换为商户时区 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改成功，时区转换正确

- [ ] 4.2 修改 CountPaymentMethod 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换
  - Requirements: Requirement 2
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功

- [ ] 4.3 修改 CountProductCategory 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换
  - Requirements: Requirement 2
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功

- [ ] 4.4 修改 CountProduct 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 在应用层完成时区转换
  - Requirements: Requirement 2
  - Leverage: Task 4.1 的实现
  - Success: 方法修改成功

- [ ] 4.5 修改 Printer 方法实现应用层时区转换

  - File: `main/app/service/business_service.go`
  - Purpose: 打印功能在应用层完成时区转换
  - Requirements: Requirement 2
  - Leverage: Task 4.1 的实现，Task 2.3-2.4 的 GetParam 方法
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 Printer 方法，在应用层完成时区转换 | Context: 1) 获取商户时区，2) 调用 req.GetParam(timezone, openingHours) 转换时间范围为 UTC 时间戳，3) 调用 Repository 查询数据，4) 将查询结果中的 UTC 时间戳转换为商户时区 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改成功，时区转换正确

- [ ] 4.6 编写 Service 单元测试（收银机端相关方法）

  - File: `main/app/service/business_service_test.go`
  - Purpose: 确保 Service 时区转换正确
  - Requirements: Requirement 2
  - Leverage: 现有测试: `main/app/service/*_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为收银机端相关的统计 Service 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试时区转换正确性，测试多种时区场景（UTC+7, UTC+8, UTC+9） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 5: API 层参数处理（收银机端）

- [ ] 5.1 修改收银机 Printer API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机打印接口支持 query_start_date 和 query_end_date 参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: 现有 API 实现，Task 2.3-2.4 的 GetParam 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 修改收银机 Printer API，支持接收 query_start_date 和 query_end_date 参数，并获取商户时区传递给 Service | Context: 1) 使用 ShouldBindJSON 绑定参数，2) 获取商户时区 ctx.GetCompanySetting().Timezone，3) 调用 req.GetParam(timezone, openingHours) 转换时间范围，4) 调用 Service 层方法 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 修改成功，参数接收正确，时区转换正确

- [ ] 5.2 修改收银机 CountBusiness API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计营业数据接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1 的实现
  - Success: API 修改成功

- [ ] 5.3 修改收银机 CountPaymentMethod API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计支付方式接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1 的实现
  - Success: API 修改成功

- [ ] 5.4 修改收银机 CountProductCategory API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计商品分类接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1 的实现
  - Success: API 修改成功

- [ ] 5.5 修改收银机 CountProduct API 支持新参数和时区转换

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 收银机统计商品接口支持新参数，使用商户时区
  - Requirements: Requirement 2
  - Leverage: Task 5.1 的实现
  - Success: API 修改成功

- [ ] 5.6 编写收银机 API 集成测试

  - File: `main/app/api/v1/cashier/cashier_statistics_test.go`
  - Purpose: 测试收银机统计 API 接口参数处理和时区转换
  - Requirements: Requirement 2
  - Leverage: 现有测试: `main/app/api/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为收银机统计 API 编写集成测试 | Context: 测试 query_start_date 和 query_end_date 参数，测试时区转换，测试多种时区场景（UTC+7, UTC+8, UTC+9） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] 收银机端 5 个接口全部支持时区转换
- [ ] 收银机端 5 个接口全部支持 query_start_date 和 query_end_date 参数
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有变更）
- [ ] CHANGELOG.md 已更新

---

**版本**: v1.0.0  
**创建日期**: 2025-12-30  
**维护者**: 后端开发组

