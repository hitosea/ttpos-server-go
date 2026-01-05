# 旧/新管理端-已选择订单数据隐藏 任务分解

> 本文档定义旧/新管理端已选择订单数据隐藏功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 64  
**已完成**: 23  
**进行中**: -  
**完成率**: 35.9%

**最新更新**: 2025-12-22 - 创建任务分解

---

## Phase 0: DTO 层修改

- [x] 0.1 修改 `BusinessTimePeriodReq` - 添加 `ExcludeDataManage` 字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 2.1, 2.2
  - Leverage: 现有 `BusinessDataCountReq` 中的 `ExcludeDataManage` 字段实现
  - Prompt: Role: Go Developer | Task: 在 `BusinessTimePeriodReq` 结构体中添加 `ExcludeDataManage bool` 字段，使用 `json:"exclude_data_manage"` 标签 | Context: 参考 `BusinessDataCountReq` 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，类型和标签正确

- [x] 0.2 修改 `StatisticsSummaryReq` - 添加 `ExcludeDataManage` 字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 3.1, 3.2
  - Leverage: 现有 `BusinessDataCountReq` 中的 `ExcludeDataManage` 字段实现
  - Prompt: Role: Go Developer | Task: 在 `StatisticsSummaryReq` 结构体中添加 `ExcludeDataManage bool` 字段 | Context: 参考 `BusinessDataCountReq` 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功

- [x] 0.3 修改 `StatisticsPaymentMethodReq` - 添加 `ExcludeDataManage` 字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 4.1, 4.2
  - Leverage: 现有 `BusinessDataCountReq` 中的 `ExcludeDataManage` 字段实现
  - Prompt: Role: Go Developer | Task: 在 `StatisticsPaymentMethodReq` 结构体中添加 `ExcludeDataManage bool` 字段 | Context: 参考 `BusinessDataCountReq` 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功

- [x] 0.4 修改 `ChannelSalesReq` - 添加 `ExcludeDataManage` 字段

  - File: `main/app/dto/req/statistics.go` 或相关文件
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 5.1, 5.2
  - Leverage: 现有 `BusinessDataCountReq` 中的 `ExcludeDataManage` 字段实现
  - Prompt: Role: Go Developer | Task: 在 `ChannelSalesReq` 结构体中添加 `ExcludeDataManage bool` 字段 | Context: 参考 `BusinessDataCountReq` 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功

- [x] 0.5 修改 `BusinessDataCountProductSalesReq` - 添加 `ExcludeDataManage` 字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 6.1, 6.2
  - Leverage: 现有 `BusinessDataCountReq` 中的 `ExcludeDataManage` 字段实现
  - Prompt: Role: Go Developer | Task: 在 `BusinessDataCountProductSalesReq` 结构体中添加 `ExcludeDataManage bool` 字段 | Context: 参考 `BusinessDataCountReq` 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功

- [x] 0.6 修改 `UserAnalysisReq` - 添加 `ExcludeDataManage` 字段

  - File: `main/app/dto/req/statistics.go` 或相关文件
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 7.1, 7.2
  - Leverage: 现有 `BusinessDataCountReq` 中的 `ExcludeDataManage` 字段实现
  - Prompt: Role: Go Developer | Task: 在 `UserAnalysisReq` 结构体中添加 `ExcludeDataManage bool` 字段 | Context: 参考 `BusinessDataCountReq` 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功

---

## Phase 1: API Handler 层修改

- [x] 1.1 修改 `CountHome` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 1.1
  - Leverage: 收银机端实现 `main/app/api/v1/cashier/cashier_statistics.go` 中的 `CountBusiness` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 `CountHome` Handler 方法中判断数据管理功能是否开启，设置 `countReq.ExcludeDataManage` | Context: 判断 `companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage`，参考收银机端的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功，参数设置正确

- [x] 1.2 修改 `CountBusiness` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 1.2
  - Leverage: 收银机端实现 `main/app/api/v1/cashier/cashier_statistics.go` 中的 `CountBusiness` 方法
  - Status: ⚠️ 可能已实现，需要检查
  - Prompt: Role: Go Developer | Task: 检查并修改 `CountBusiness` Handler，确保判断数据管理功能并设置参数 | Context: 参考收银机端的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.3 修改 `CountArea` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 1.3
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `CountArea` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.4 修改 `ChannelSales` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 1.4, 5.1
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `ChannelSales` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.5 修改 `CountPaymentMethod` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 1.5
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `CountPaymentMethod` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.6 修改 `CountProductRank` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 1.6
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `CountProductRank` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.7 修改 `CountBusinessTimePeriod` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 2.1
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `CountBusinessTimePeriod` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.8 修改 `CountBusinessSummary` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 3.1
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `CountBusinessSummary` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.9 修改 `CountBusinessPaymentMethod` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 4.1
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `CountBusinessPaymentMethod` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.10 修改 `CountProductSales` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 6.1
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `CountProductSales` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.11 修改 `UserAnalysis` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 7.1
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `UserAnalysis` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.12 修改 `ExportBusinessTimePeriod` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 2.2
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `ExportBusinessTimePeriod` Handler 方法中判断数据管理功能并设置参数，传递给 Service 层 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功，参数正确传递

- [x] 1.13 修改 `ExportBusinessSummary` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 3.2
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `ExportBusinessSummary` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.14 修改 `ExportBusinessPaymentMethod` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 4.2
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `ExportBusinessPaymentMethod` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.15 修改 `ExportChannelSales` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 5.2
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `ExportChannelSales` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.16 修改 `ExportProductSales` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 6.2
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `ExportProductSales` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

- [x] 1.17 修改 `ExportUserAnalysis` Handler - 判断数据管理功能并设置参数

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 7.2
  - Leverage: 收银机端实现
  - Prompt: Role: Go Developer | Task: 在 `ExportUserAnalysis` Handler 方法中判断数据管理功能并设置参数 | Context: 参考 Task 1.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Handler 修改成功

---

## Phase 2: Service 层修改

- [ ] 2.1 修改 `CountHome` Service - 传递 `ExcludeDataManage` 参数

  - File: `main/app/service/business.go`
  - Purpose: 在 `CountHome` 方法中传递 `ExcludeDataManage` 参数给统计方法
  - Requirements: 1.1
  - Leverage: 现有 `CountBusiness` 方法的实现
  - Prompt: Role: Go Developer with Service Layer expertise | Task: 修改 `CountHome` 方法，在调用 `CountSale` 时传递 `req.ExcludeDataManage` 参数 | Context: 参考 `CountBusiness` 方法的实现，确保所有统计方法调用都传递参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功，参数正确传递

- [ ] 2.2 修改 `CountBusinessTimePeriod` Service - 添加过滤逻辑

  - File: `main/app/service/statistics.go`
  - Purpose: 在营业时段统计中添加已选择订单过滤逻辑
  - Requirements: 2.1
  - Leverage: 现有 `CountSale` 方法的过滤逻辑实现
  - Prompt: Role: Go Developer | Task: 修改 `CountBusinessTimePeriod` 方法，根据 `req.ExcludeDataManage` 参数添加过滤条件 | Context: 使用 `WhereNotInDataManageSubQuery` 方法，传入独立的 `ctx.GetDB()` 参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc，参考 `CountSale` 的实现 | Success: Service 修改成功，过滤逻辑正确

- [ ] 2.3 修改 `CountBusinessSummary` Service - 添加过滤逻辑

  - File: `main/app/service/statistics.go`
  - Purpose: 在综合运营统计中添加已选择订单过滤逻辑
  - Requirements: 3.1
  - Leverage: 现有 `CountSale` 方法的过滤逻辑实现
  - Prompt: Role: Go Developer | Task: 修改 `CountBusinessSummary` 方法，根据 `req.ExcludeDataManage` 参数添加过滤条件 | Context: 使用 `WhereNotInDataManageSubQuery` 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.4 修改 `CountBusinessPaymentMethod` Service - 添加过滤逻辑

  - File: `main/app/service/statistics.go`
  - Purpose: 在营业收款统计中添加已选择订单过滤逻辑
  - Requirements: 4.1
  - Leverage: 现有 `CountPayment` 方法的过滤逻辑实现
  - Prompt: Role: Go Developer | Task: 修改 `CountBusinessPaymentMethod` 方法，根据 `req.ExcludeDataManage` 参数添加过滤条件 | Context: 使用 `WhereNotInDataManageSubQuery` 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.5 修改 `CountChannelSales` Service - 传递过滤参数给 Repository

  - File: `main/app/service/business.go`
  - Purpose: 在渠道营业统计中传递过滤参数给 Repository 层
  - Requirements: 5.1
  - Leverage: 现有 `CountChannelSale` Repository 方法已支持 `opts` 参数
  - Prompt: Role: Go Developer | Task: 修改 `CountChannelSales` 方法，根据 `req.ExcludeDataManage` 参数构建过滤选项，传递给 `CountChannelSale` Repository 方法 | Context: 使用 `WhereNotInDataManageSubQuery` 构建 `opts`，调用 `statisticsRepo.CountChannelSale(startTime, endTime, opts...)` | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功，过滤参数正确传递

- [ ] 2.6 修改 `CountProductSales` Service - 添加过滤逻辑

  - File: `main/app/service/statistics.go` 或 `business.go`
  - Purpose: 在商品销售统计中添加已选择订单过滤逻辑
  - Requirements: 6.1
  - Leverage: 现有统计方法的过滤逻辑实现
  - Prompt: Role: Go Developer | Task: 修改 `CountProductSales` 方法，根据 `req.ExcludeDataManage` 参数添加过滤条件 | Context: 使用 `WhereNotInDataManageSubQuery` 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.7 修改 `CountUserAnalysis` Service - 添加过滤逻辑

  - File: `main/app/service/business.go`
  - Purpose: 在用户分析统计中添加已选择订单过滤逻辑
  - Requirements: 7.1
  - Leverage: 现有统计方法的过滤逻辑实现
  - Prompt: Role: Go Developer | Task: 修改 `CountUserAnalysis` 方法，根据 `req.ExcludeDataManage` 参数添加过滤条件 | Context: 使用 `WhereNotInDataManageSubQuery` 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.8 修改 `ExportBusinessTimePeriod` Service - 传递过滤参数

  - File: `main/app/service/business.go`
  - Purpose: 在导出时段营业统计时传递过滤参数给统计方法
  - Requirements: 2.2
  - Leverage: 现有导出方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `ExportBusinessTimePeriod` 方法，确保调用 `CountBusinessTimePeriod` 时传递 `req.ExcludeDataManage` 参数 | Context: 导出接口内部调用统计接口时，必须传递相同的过滤参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功，导出数据与页面数据一致

- [ ] 2.9 修改 `ExportBusinessSummary` Service - 传递过滤参数

  - File: `main/app/service/business.go`
  - Purpose: 在导出综合运营统计时传递过滤参数给统计方法
  - Requirements: 3.2
  - Leverage: 现有导出方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `ExportBusinessSummary` 方法，确保调用 `CountBusinessSummary` 时传递 `req.ExcludeDataManage` 参数 | Context: 导出接口内部调用统计接口时，必须传递相同的过滤参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.10 修改 `ExportBusinessPaymentMethod` Service - 传递过滤参数

  - File: `main/app/service/business.go`
  - Purpose: 在导出营业收款统计时传递过滤参数给统计方法
  - Requirements: 4.2
  - Leverage: 现有导出方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `ExportBusinessPaymentMethod` 方法，确保调用 `CountBusinessPaymentMethod` 时传递 `req.ExcludeDataManage` 参数 | Context: 导出接口内部调用统计接口时，必须传递相同的过滤参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.11 修改 `ExportChannelSales` Service - 传递过滤参数

  - File: `main/app/service/business.go`
  - Purpose: 在导出渠道营业统计时传递过滤参数给统计方法
  - Requirements: 5.2
  - Leverage: 现有导出方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `ExportChannelSales` 方法，确保调用 `CountChannelSales` 时传递 `req.ExcludeDataManage` 参数 | Context: 导出接口内部调用统计接口时，必须传递相同的过滤参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.12 修改 `ExportProductSales` Service - 传递过滤参数

  - File: `main/app/service/business.go`
  - Purpose: 在导出商品销售统计时传递过滤参数给统计方法
  - Requirements: 6.2
  - Leverage: 现有导出方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `ExportProductSales` 方法，确保调用 `CountProductSales` 时传递 `req.ExcludeDataManage` 参数 | Context: 导出接口内部调用统计接口时，必须传递相同的过滤参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

- [ ] 2.13 修改 `ExportUserAnalysis` Service - 传递过滤参数

  - File: `main/app/service/business.go`
  - Purpose: 在导出用户分析统计时传递过滤参数给统计方法
  - Requirements: 7.2
  - Leverage: 现有导出方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `ExportUserAnalysis` 方法，确保调用 `CountUserAnalysis` 时传递 `req.ExcludeDataManage` 参数 | Context: 导出接口内部调用统计接口时，必须传递相同的过滤参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功

---

## Phase 3: Repository 层修改

- [ ] 3.1 修改 `CountBusinessTimePeriod` Repository - 支持 `opts` 参数

  - File: `main/app/repository/statistics.go`
  - Purpose: Repository 方法支持过滤选项参数
  - Requirements: 2.1
  - Leverage: 现有 `CountChannelSale` Repository 方法的 `opts` 参数实现
  - Prompt: Role: Go Developer with GORM expertise | Task: 修改 `CountBusinessTimePeriod` Repository 方法签名，添加 `opts ...DBOption` 参数，并在查询中应用这些选项 | Context: 参考 `CountChannelSale` 的实现，使用选项模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Repository 修改成功，支持过滤选项

- [ ] 3.2 修改 `CountBusinessSummary` Repository - 支持 `opts` 参数

  - File: `main/app/repository/statistics.go`
  - Purpose: Repository 方法支持过滤选项参数
  - Requirements: 3.1
  - Leverage: 现有 `CountChannelSale` Repository 方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `CountBusinessSummary` Repository 方法签名，添加 `opts ...DBOption` 参数 | Context: 参考 Task 3.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Repository 修改成功

- [ ] 3.3 修改 `CountBusinessPaymentMethod` Repository - 支持 `opts` 参数

  - File: `main/app/repository/statistics.go`
  - Purpose: Repository 方法支持过滤选项参数
  - Requirements: 4.1
  - Leverage: 现有 `CountChannelSale` Repository 方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `CountBusinessPaymentMethod` Repository 方法签名，添加 `opts ...DBOption` 参数 | Context: 参考 Task 3.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Repository 修改成功

- [ ] 3.4 修改 `CountProductSales` Repository - 支持 `opts` 参数

  - File: `main/app/repository/statistics.go`
  - Purpose: Repository 方法支持过滤选项参数
  - Requirements: 6.1
  - Leverage: 现有 `CountChannelSale` Repository 方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `CountProductSales` Repository 方法签名，添加 `opts ...DBOption` 参数 | Context: 参考 Task 3.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Repository 修改成功

- [ ] 3.5 修改 `CountUserAnalysis` Repository - 支持 `opts` 参数

  - File: `main/app/repository/statistics.go` 或相关文件
  - Purpose: Repository 方法支持过滤选项参数
  - Requirements: 7.1
  - Leverage: 现有 `CountChannelSale` Repository 方法的实现
  - Prompt: Role: Go Developer | Task: 修改 `CountUserAnalysis` Repository 方法签名，添加 `opts ...DBOption` 参数 | Context: 参考 Task 3.1 的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Repository 修改成功

- [ ] 3.6 验证 `CountChannelSale` Repository - 确认支持 `opts` 参数

  - File: `main/app/repository/statistics.go`
  - Purpose: 确认 `CountChannelSale` Repository 方法已支持 `opts` 参数
  - Requirements: 5.1
  - Leverage: 现有实现
  - Status: ✅ 已支持，需要验证
  - Prompt: Role: Go Developer | Task: 验证 `CountChannelSale` Repository 方法已支持 `opts ...DBOption` 参数 | Context: 检查方法签名和实现 | Restrictions: 如果已支持，标记为完成；如果未支持，需要修改 | Success: 确认方法已支持 `opts` 参数

---

## Phase 4: 测试

- [ ] 4.1 编写 Service 层单元测试

  - File: `main/app/service/statistics_test.go`, `main/app/service/business_test.go`
  - Purpose: 确保 Service 层过滤逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为修改的 Service 方法编写单元测试，测试过滤逻辑 | Context: 测试数据管理功能开启/关闭时的行为，测试已选择订单被正确排除 | Restrictions: 测试覆盖率 ≥ 70% | Success: 测试覆盖率达标，所有测试通过

- [ ] 4.2 编写 Repository 层单元测试

  - File: `main/app/repository/statistics_test.go`
  - Purpose: 确保 Repository 层过滤逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer | Task: 为修改的 Repository 方法编写单元测试 | Context: 测试过滤选项正确应用 | Restrictions: 测试覆盖率 ≥ 80% | Success: 测试覆盖率达标

- [ ] 4.3 API 测试 - 统计接口

  - File: `test/api/` 或相关测试文件
  - Purpose: 测试所有统计接口的过滤功能
  - Requirements: 所有功能需求
  - Leverage: 现有 API 测试
  - Prompt: Role: QA Automation Engineer | Task: 编写 API 测试，验证所有统计接口正确排除已选择订单 | Context: 测试数据管理功能开启/关闭时的接口响应 | Restrictions: 覆盖所有统计接口 | Success: 所有 API 测试通过

- [ ] 4.4 API 测试 - 导出接口

  - File: `test/api/` 或相关测试文件
  - Purpose: 测试所有导出接口的过滤功能和数据一致性
  - Requirements: 所有导出相关需求
  - Leverage: 现有 API 测试
  - Prompt: Role: QA Automation Engineer | Task: 编写 API 测试，验证导出接口正确排除已选择订单，且导出数据与页面数据一致 | Context: 对比导出数据和统计接口返回数据 | Restrictions: 覆盖所有导出接口 | Success: 所有导出接口测试通过，数据一致性验证通过

- [ ] 4.5 集成测试 - 端到端流程

  - File: `test/integration/` 或相关测试文件
  - Purpose: 测试完整的业务流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 编写端到端集成测试 | Context: 1. 选择订单进行数据管理 2. 查看首页统计数据（应排除已选择订单）3. 查看各类报表统计（应排除已选择订单）4. 导出报表（导出数据应与页面数据一致） | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.6 数据一致性测试 - 导出数据与页面数据一致

  - File: `test/integration/` 或相关测试文件
  - Purpose: 验证导出数据与页面展示数据完全一致
  - Requirements: 所有导出相关需求
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 编写数据一致性测试，对比导出 Excel 文件中的数据与统计接口返回的数据 | Context: 确保导出数据 = 页面数据（都排除已选择订单） | Restrictions: 覆盖所有导出接口 | Success: 数据一致性验证通过

---

## Phase 5: 文档和验证

- [ ] 5.1 确认旧后台使用的统计接口列表

  - File: `docs/` 或代码审查
  - Purpose: 确认旧后台是否使用相同的统计接口
  - Requirements: 8.1, 8.2, 8.3
  - Leverage: 代码搜索和文档
  - Prompt: Role: Technical Writer | Task: 梳理旧后台使用的统计接口，确认是否与新管理端使用相同接口 | Context: 搜索 PHP Admin 模块中的接口调用 | Restrictions: 确保接口列表完整 | Success: 接口列表确认完成

- [ ] 5.2 验证旧后台和新管理端的数据一致性

  - File: `test/integration/` 或手动测试
  - Purpose: 确保旧后台和新管理端使用相同接口时数据一致
  - Requirements: 8.3
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 验证旧后台和新管理端的数据一致性 | Context: 如果旧后台使用相同接口，确保也应用相同的过滤逻辑 | Restrictions: 数据必须一致 | Success: 数据一致性验证通过

- [ ] 5.3 更新 API 文档（如有）

  - File: `docs/shared/api/` 或相关文档
  - Purpose: 更新 API 文档，说明过滤功能
  - Requirements: 文档要求
  - Leverage: 现有 API 文档
  - Prompt: Role: Technical Writer | Task: 更新相关 API 文档，说明已选择订单过滤功能 | Context: 说明 `ExcludeDataManage` 参数的作用 | Restrictions: 文档准确完整 | Success: API 文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - 统计相关模块: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] 导出数据与页面数据一致

### 文档同步

- [ ] API 文档已更新（如有新接口或参数）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-management-hidden-selected-orders-data/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-management-hidden-selected-orders-data/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-management-hidden-selected-orders-data/tasks.md
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
**最后更新**: 2025-12-22  
**维护者**: 后端开发组

