# 新管理端-报表-门店汇总统计 任务分解

> 本文档定义 新管理端-报表-门店汇总统计 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 21  
**已完成**: 19  
**进行中**: -  
**完成率**: 90.5%

---

## Phase 1: DTO 和 API 定义

- [x] 1.1 创建 Request DTO

  - File: `main/app/dto/req/shop_summary_statistics_req.go`
  - Purpose: 定义门店汇总统计请求参数
  - Requirements: Requirement 1, Requirement 2, Requirement 3, Requirement 4
  - Leverage: 现有 DTO: `main/app/dto/req/statistics_summary_req.go`
  - Prompt: Role: Go Developer | Task: 创建 ShopSummaryStatisticsReq 结构体，包含指标类型、门店UUID列表、开始日期(QueryStartDate)、结束日期(QueryEndDate)、支付方式UUID列表（可选） | Context: 使用 binding 标签验证参数，指标类型使用 oneof 验证，日期格式为 YYYY-MM-DD | Restrictions: 遵循 .cursor/rules/go-main.mdc，注意：ExcludeDataManage 不在请求参数中，由接口自行判断 | Success: DTO 创建成功，validation 正确

- [x] 1.2 创建 Response DTO - 营业数据汇总

  - File: `main/app/dto/resp/shop_summary_statistics_resp.go`
  - Purpose: 定义营业数据汇总响应数据结构
  - Requirements: Requirement 2
  - Leverage: 现有 DTO: `main/app/dto/resp/business_data_resp/base.go` 中的 `StatisticsSummaryItem`
  - Prompt: Role: Go Developer | Task: 创建 BusinessSummaryResp、BusinessSummaryDetailItem 结构体，包含14个字段（营业日、门店名称、订单金额、实付金额等） | Context: 金额字段使用 float64，保留2位小数 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，字段定义正确

- [x] 1.3 创建 Response DTO - 支付方式汇总

  - File: `main/app/dto/resp/shop_summary_statistics_resp.go`
  - Purpose: 定义支付方式汇总响应数据结构
  - Requirements: Requirement 3
  - Leverage: 现有 DTO: `main/app/dto/resp/business_data_resp/base.go` 中的 `StatisticsPaymentMethodItem`
  - Prompt: Role: Go Developer | Task: 创建 PaymentMethodSummaryResp、PaymentMethodSummaryDetailItem 结构体，包含6个字段（营业日、门店名称、支付方式、支付金额、支付笔数、支付占比） | Context: 支付占比使用 float64，保留2位小数 | Restrictions: data 必须是对象 | Success: DTO 创建成功，字段定义正确

- [x] 1.4 创建 Response DTO - 退款金额汇总

  - File: `main/app/dto/resp/shop_summary_statistics_resp.go`
  - Purpose: 定义退款金额汇总响应数据结构
  - Requirements: Requirement 4
  - Leverage: 现有退款相关 DTO
  - Prompt: Role: Go Developer | Task: 创建 RefundSummaryResp、RefundSummaryDetailItem 结构体，包含9个字段（营业日、门店名称、退款金额、退款笔数、退款率、部分退款、整单退款等） | Context: 退款率使用 float64，保留2位小数 | Restrictions: data 必须是对象 | Success: DTO 创建成功，字段定义正确

- [x] 1.5 创建 API Handler

  - File: `main/app/api/v1/shop/shop_summary_statistics.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: Requirement 1, Requirement 2, Requirement 3, Requirement 4, Requirement 5
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_statistics.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 shopSummaryStatisticsHandler，实现 GetShopSummaryCompanyList、GetShopSummaryStatistics 和 ExportShopSummaryStatistics 方法 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，注意：ExcludeDataManage 不在请求参数中，由 Service 层自行判断 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 1.6 创建 Response DTO - 门店列表

  - File: `main/app/dto/resp/shop_summary_statistics_resp.go`
  - Purpose: 定义门店列表响应数据结构
  - Requirements: Requirement 1
  - Leverage: 现有 DTO: `main/app/dto/resp/saas_staff.go` 中的 `CompanyStaffResp`
  - Prompt: Role: Go Developer | Task: 创建 ShopSummaryCompanyListResp 结构体，包含门店列表 | Context: 复用 CompanyStaffResp 结构 | Restrictions: data 必须是对象 | Success: DTO 创建成功，字段定义正确

- [x] 1.7 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 API 路由
  - Requirements: Requirement 5
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功，路径为 `/api/v1/shop/shop_summary_statistics` 和 `/api/v1/shop/shop_summary_statistics/company_list`

- [x] 1.8 创建获取门店支付方式接口

  - File: `main/app/api/v1/shop/shop_statistics.go`, `main/app/service/business.go`
  - Purpose: 新增获取门店支付方式列表接口
  - Requirements: Requirement 3（支付方式汇总）
  - Leverage: 现有 Service: `main/app/service/business.go` 中的 `GetCompanyList`
  - Prompt: Role: Go Developer | Task: 创建 GetCompanyPaymentMethods 接口，获取有权限的所有门店的支付方式，汇总去重后返回 | Context: 使用 goroutine 并发查询多个门店，按 Sort 升序、CreateTime 倒序、ID 倒序排序 | Restrictions: 遵循 .cursor/rules/go-main.mdc 和 .cursor/rules/api.mdc | Success: 接口创建成功，并发查询实现正确，排序逻辑正确

- [x] 1.9 创建获取门店支付方式响应 DTO

  - File: `main/app/dto/resp/statistics_summary_resp.go`
  - Purpose: 定义门店支付方式列表响应数据结构
  - Requirements: Requirement 3（支付方式汇总）
  - Leverage: 现有 DTO: `main/app/dto/resp/statistics_summary_resp.go`
  - Prompt: Role: Go Developer | Task: 创建 CompanyPaymentMethodListResp 和 CompanyPaymentMethodItem 结构体 | Context: 包含支付方式名称字段 | Restrictions: data 必须是对象 | Success: DTO 创建成功，字段定义正确

---

## Phase 2: Service 层实现 - 核心业务逻辑

- [x] 2.1 扩展 StatisticsService 接口

  - File: `main/app/service/i_statistics_service.go`
  - Purpose: 添加门店汇总统计方法到接口
  - Requirements: Requirement 1, Requirement 2, Requirement 3, Requirement 4
  - Leverage: 现有 Service 接口: `main/app/service/i_statistics_service.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 IStatisticsSrv 接口中添加 GetShopSummaryCompanyList、GetShopSummaryStatistics 和 ExportShopSummaryStatistics 方法 | Context: 方法签名与 API Handler 调用一致 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [x] 2.1.1 在 AuthService 中添加 GetCompanyList 公共方法

  - File: `main/app/service/auth.go`
  - Purpose: 将 getCompanyList 私有方法改为公共方法，供其他 Service 调用
  - Requirements: Requirement 1
  - Leverage: 现有方法: `main/app/service/auth.go` 中的 `getCompanyList`（第1741行）
  - Prompt: Role: Go Developer | Task: 将 getCompanyList 方法改为 GetCompanyList 公共方法，添加到 IAuthSrv 接口 | Context: 该方法用于获取员工可用的商家列表（过滤已过期、异常的商家），子店用户使用 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法改为公共方法，接口定义正确

- [x] 2.2 实现获取门店列表方法

  - File: `main/app/service/statistics.go`
  - Purpose: 实现 GetShopSummaryCompanyList 方法，根据用户角色返回不同的门店列表
  - Requirements: Requirement 1
  - Leverage: 现有 Service: `main/app/service/company.go` 中的 `GetVisibleCompanyList`，`main/app/service/auth.go` 中的 `GetCompanyList`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetShopSummaryCompanyList 方法，总店使用 GetVisibleCompanyList，子店使用 GetCompanyList | Context: 判断 companySetting.IsHeadquarter()，总店返回本店及下级所有子店，子店返回本店及已授权的其他门店 | Restrictions: 遵循 .cursor/rules/go-main.mdc，需要注入 IAuthSrv | Success: 门店列表获取正确，根据角色返回不同列表

- [x] 2.3 实现门店权限验证

  - File: `main/app/service/statistics.go`
  - Purpose: 验证用户选择的门店是否在可见列表中
  - Requirements: Requirement 1
  - Leverage: 现有 Service: `main/app/service/company.go` 中的 `GetVisibleCompanyList`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 validateCompanyUuids 方法，验证用户选择的门店UUID列表是否在可见门店列表中 | Context: 使用 CompanyService 获取可见门店列表，过滤无效门店 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 权限验证正确，返回有效门店列表

- [x] 2.4 实现多门店并发查询框架

  - File: `main/app/service/statistics.go`
  - Purpose: 实现多门店并发查询的基础框架
  - Requirements: Requirement 2, Requirement 3, Requirement 4
  - Leverage: 现有并发查询模式
  - Prompt: Role: Go Developer with concurrency expertise | Task: 实现多门店并发查询框架，使用 goroutine 和 channel，设置超时机制 | Context: 为每个门店创建独立的 context，使用 context.WithTimeout 设置超时，使用 channel 收集查询结果 | Restrictions: 遵循 .cursor/rules/go-main.mdc，正确处理错误和超时 | Success: 并发查询框架实现正确，错误处理完善

- [x] 2.5 实现单门店营业数据查询

  - File: `main/app/service/statistics.go`
  - Purpose: 查询单个门店的营业数据汇总
  - Requirements: Requirement 2
  - Leverage: 现有 Service: `main/app/service/statistics.go` 中的 `CountBusinessSummary`，现有 Repository: `main/app/repository/statistics.go` 中的 `CountBusinessSummary`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 queryBusinessSummaryByCompany 方法，查询单个门店在指定日期范围内的营业数据汇总 | Context: 获取门店数据库连接，调用 StatisticsRepository.CountBusinessSummary，转换为响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc，正确处理数据库连接失败 | Success: 单门店查询实现正确，数据转换准确

- [x] 2.6 实现营业数据汇总统计主方法

  - File: `main/app/service/statistics.go`
  - Purpose: 实现营业数据汇总统计的主方法
  - Requirements: Requirement 2
  - Leverage: Task 2.3, Task 2.4
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 getBusinessSummaryStatistics 方法，接收开始日期和结束日期字符串，解析为时间戳，调用多门店并发查询框架，收集所有门店数据，计算汇总行 | Context: 使用 goroutine 并发查询多个门店，将日期字符串解析为时间戳后传递给单门店查询方法，收集结果后计算汇总行 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 多门店查询实现正确，日期范围解析正确，汇总计算准确

- [x] 2.7 实现营业数据汇总行计算

  - File: `main/app/service/statistics.go`
  - Purpose: 计算营业数据汇总行的数据
  - Requirements: Requirement 2
  - Leverage: 汇总规则：金额类求和、数量类求和、比率类重新计算
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 calculateBusinessSummaryRow 方法，计算汇总行的所有字段 | Context: 金额类字段求和，数量类字段求和，人均/单均类字段重新计算（汇总金额/汇总人数或订单数） | Restrictions: 遵循汇总规则，避免简单平均导致的偏差 | Success: 汇总计算正确，比率类字段重新计算

- [x] 2.8 实现单门店支付方式查询

  - File: `main/app/service/statistics.go`
  - Purpose: 查询单个门店的支付方式统计
  - Requirements: Requirement 3
  - Leverage: 现有 Repository: `main/app/repository/statistics.go` 中的支付方式统计方法（需要确认是否存在）
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 queryPaymentMethodByCompany 方法，查询单个门店在指定日期范围(startTime, endTime)内的支付方式统计 | Context: 如果现有 Repository 方法不存在，需要先实现 Repository 方法，按日期范围查询返回每日数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 单门店支付方式查询实现正确，按日期范围查询

- [x] 2.9 实现支付方式汇总统计主方法

  - File: `main/app/service/statistics.go`
  - Purpose: 实现支付方式汇总统计的主方法
  - Requirements: Requirement 3
  - Leverage: Task 2.3, Task 2.7
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 getPaymentMethodSummaryStatistics 方法，接收开始日期和结束日期字符串，解析为时间戳，调用多门店并发查询框架，按支付方式分组汇总 | Context: 将日期字符串解析为时间戳后传递给单门店查询方法，收集所有门店的支付方式数据，按支付方式分组，计算每个支付方式的汇总数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 支付方式汇总实现正确，日期范围解析正确，分组计算准确

- [x] 2.10 实现支付方式汇总占比计算

  - File: `main/app/service/statistics.go`
  - Purpose: 计算支付方式汇总占比
  - Requirements: Requirement 3
  - Leverage: 汇总规则：支付占比重新计算（汇总支付金额/汇总总实付金额）
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现支付方式汇总占比计算，确保占比重新计算而不是简单平均 | Context: 先计算所有支付方式的汇总金额和总实付金额，然后计算占比 | Restrictions: 遵循汇总规则 | Success: 占比计算正确，重新计算而非简单平均

- [x] 2.11 实现退款统计 Repository 方法（如不存在）

  - File: `main/app/repository/statistics.go`
  - Purpose: 实现退款统计的 Repository 方法
  - Requirements: Requirement 4
  - Leverage: 现有退款相关 Repository 方法
  - Prompt: Role: Go Developer with GORM expertise | Task: 如果退款统计 Repository 方法不存在，实现 CountRefundSummary 方法，查询退款金额、退款笔数、部分退款、整单退款等 | Context: 查询 ttpos_return_order 表，区分部分退款和整单退款 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用选项模式 | Success: Repository 方法实现正确，查询逻辑准确

- [x] 2.12 实现单门店退款查询

  - File: `main/app/service/statistics.go`
  - Purpose: 查询单个门店的退款统计
  - Requirements: Requirement 4
  - Leverage: Task 2.10
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 queryRefundByCompany 方法，查询单个门店在指定日期范围(startTime, endTime)内的退款统计 | Context: 调用退款统计 Repository 方法，按日期范围查询返回每日数据，计算退款率 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 单门店退款查询实现正确，按日期范围查询

- [x] 2.13 实现退款金额汇总统计主方法

  - File: `main/app/service/statistics.go`
  - Purpose: 实现退款金额汇总统计的主方法
  - Requirements: Requirement 4
  - Leverage: Task 2.3, Task 2.11
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 getRefundSummaryStatistics 方法，接收开始日期和结束日期字符串，解析为时间戳，调用多门店并发查询框架，收集所有门店退款数据，计算汇总行 | Context: 将日期字符串解析为时间戳后传递给单门店查询方法，使用 goroutine 并发查询多个门店，收集结果后计算汇总行 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 退款汇总实现正确，日期范围解析正确，汇总计算准确

- [x] 2.14 实现退款金额汇总行计算

  - File: `main/app/service/statistics.go`
  - Purpose: 计算退款金额汇总行的数据
  - Requirements: Requirement 4
  - Leverage: 汇总规则：退款金额求和、退款笔数求和、退款率重新计算
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 calculateRefundSummaryRow 方法，计算汇总行的所有字段 | Context: 退款金额、退款笔数求和，退款率重新计算（汇总退款订单数量 ÷ 汇总总订单数量）× 100% | Restrictions: 遵循汇总规则 | Success: 汇总计算正确，退款率重新计算

- [x] 2.15 实现数据管理订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在查询时过滤已被数据管理的订单
  - Requirements: Requirement 2, Requirement 3, Requirement 4
  - Leverage: 现有过滤方法: `main/app/repository/common.go` 中的 `WhereNotInDataManageSubQuery`
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 GetShopSummaryStatistics 方法中，根据公司设置和数据管理设置判断是否排除数据管理订单，然后传递给统计查询方法 | Context: 判断 companySetting.IsOpenDataManagement() && dataManageSetting.IsEnableDataManage，将结果传递给所有统计查询方法，使用 WhereNotInDataManageSubQuery 方法传递到 Repository 层 | Restrictions: 遵循现有数据管理过滤模式 | Success: 数据过滤判断正确，已选择订单不参与计算

- [x] 2.16 实现导出功能

  - File: `main/app/service/statistics.go`
  - Purpose: 实现门店汇总统计导出功能
  - Requirements: Requirement 5
  - Leverage: 现有导出功能: `main/app/service/statistics.go` 中的 `CountExport`
  - Prompt: Role: Go Developer with Excel export expertise | Task: 实现 ExportShopSummaryStatistics 方法，导出 Excel 文件 | Context: 调用 GetShopSummaryStatistics 获取数据，使用 Excel 库生成文件 | Restrictions: 遵循现有导出模式 | Success: 导出功能实现正确，Excel 格式正确

---

## Phase 3: 测试和优化

- [ ] 3.1 编写 Service 单元测试

  - File: `main/app/service/statistics_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为门店汇总统计 Service 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试多门店查询、数据聚合、汇总计算、权限验证 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 3.2 编写 Repository 单元测试（如新增方法）

  - File: `main/app/repository/statistics_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: Requirement 4（如新增退款统计方法）
  - Leverage: 现有测试: `main/app/repository/*_test.go`
  - Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 3.3 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_summary_statistics_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为门店汇总统计 API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试权限控制 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 3.4 性能测试和优化

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Prompt: Role: Performance Engineer | Task: 进行性能测试，优化多门店并发查询性能 | Context: 测试5个门店并发查询性能，优化超时设置，优化缓存策略 | Restrictions: 本地响应时间 < 200ms（单门店），多门店查询 < 1s | Success: 性能达标，优化完成

- [x] 3.5 文档更新

  - File: `docs/shared/specs/active/story-admin-report-shop-summary-statistics/design.md`, `docs/shared/specs/active/story-admin-report-shop-summary-statistics/tasks.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, 设计文档, 任务文档 | Restrictions: 文档准确完整 | Success: 所有文档已更新

- [x] 3.6 手动测试获取门店支付方式接口

  - File: -
  - Purpose: 手动测试接口功能
  - Requirements: Requirement 3（支付方式汇总）
  - Success: 接口测试通过，功能正常

- [x] 3.7 生成 API 文档

  - File: -
  - Purpose: 生成 Swagger API 文档
  - Requirements: 文档要求
  - Success: API 文档已生成

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

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-report-shop-summary-statistics/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-report-shop-summary-statistics/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-report-shop-summary-statistics/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-report-shop-summary-statistics/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-report-shop-summary-statistics/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-05  
**维护者**: 后端开发组

