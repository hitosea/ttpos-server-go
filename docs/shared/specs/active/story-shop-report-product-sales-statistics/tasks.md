# 新管理端(商家端)-报表中心-商品销售统计 任务分解

> 本文档定义商品销售统计功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 时间工具方法扩展

- [ ] 1.1 新增近7天时间计算方法

  - File: `main/pkg/utils/time.go`
  - Purpose: 新增 `Last7DaysStartEndUnix()` 方法，计算近7天的开始和结束时间戳
  - Requirements: 1.5
  - Leverage: 现有时间工具方法 `TodayStartEndUnix()`, `YesterdayStartEndUnix()`
  - Prompt: Role: Go Developer | Task: 在 Timezone 类型中新增 Last7DaysStartEndUnix 方法，返回近7天的开始和结束时间戳 | Context: 参考 TodayStartEndUnix 实现，计算从今天往前推7天的时间范围 | Restrictions: 遵循 .cursor/rules/go-main.mdc，考虑时区设置 | Success: 方法实现成功，返回正确的时间戳

- [ ] 1.2 新增上月时间计算方法

  - File: `main/pkg/utils/time.go`
  - Purpose: 新增 `LastMonthStartEndUnix()` 方法，计算上月的开始和结束时间戳
  - Requirements: 1.6
  - Leverage: 现有时间工具方法 `MonthStartEndUnix()`
  - Prompt: Role: Go Developer | Task: 在 Timezone 类型中新增 LastMonthStartEndUnix 方法，返回上月的开始和结束时间戳 | Context: 参考 MonthStartEndUnix 实现，计算上个月的第一天和最后一天 | Restrictions: 遵循 .cursor/rules/go-main.mdc，考虑时区设置 | Success: 方法实现成功，返回正确的时间戳

- [ ] 1.3 新增今年时间计算方法

  - File: `main/pkg/utils/time.go`
  - Purpose: 新增 `YearStartEndUnix()` 方法，计算今年的开始和结束时间戳
  - Requirements: 1.7
  - Leverage: 现有时间工具方法
  - Prompt: Role: Go Developer | Task: 在 Timezone 类型中新增 YearStartEndUnix 方法，返回今年的开始和结束时间戳 | Context: 计算今年1月1日00:00:00到12月31日23:59:59的时间戳 | Restrictions: 遵循 .cursor/rules/go-main.mdc，考虑时区设置 | Success: 方法实现成功，返回正确的时间戳

---

## Phase 2: 请求参数扩展

- [ ] 2.1 扩展 BusinessDataCountProductSalesReq

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 在请求参数中新增时间类型、订单类型、订单来源字段，修改分类字段支持多选
  - Requirements: 1.1, 2.1, 3.1, 4.1
  - Leverage: 现有请求参数结构，design.md 中的参数定义
  - Prompt: Role: Go Developer | Task: 扩展 BusinessDataCountProductSalesReq 结构体，新增 TimeType、OrderType、OrderSource 字段，将 CategoryUuid 改为 CategoryUuids 支持多选 | Context: 参考 design.md 中的参数定义，使用 form 和 json 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 参数结构扩展成功，标签正确

- [ ] 2.2 扩展 CountReq 结构

  - File: `main/app/service/statistics.go`
  - Purpose: 在统计请求参数中新增相应字段
  - Requirements: 1.1, 2.1, 3.1, 4.1
  - Leverage: 现有 CountReq 结构，design.md 中的参数定义
  - Prompt: Role: Go Developer | Task: 扩展 CountReq 结构体，新增 TimeType、OrderTypes、OrderSource 字段，将 CategoryUuid 改为 CategoryUuids | Context: 参考 design.md 中的参数定义 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 参数结构扩展成功

---

## Phase 3: 时间类型处理逻辑

- [ ] 3.1 扩展 buildCountOpts 方法

  - File: `main/app/service/statistics.go`
  - Purpose: 在 buildCountOpts 方法中处理时间类型，自动计算时间范围
  - Requirements: 1.1-1.7
  - Leverage: 现有 buildCountOpts 方法，Phase 1 的时间工具方法
  - Prompt: Role: Go Developer | Task: 在 buildCountOpts 方法中新增时间类型处理逻辑，支持1-7的时间类型 | Context: 参考 design.md 中的时间类型处理逻辑，使用 Phase 1 新增的时间工具方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，优先使用自定义时间范围 | Success: 时间类型处理逻辑实现成功，时间计算准确

---

## Phase 4: 订单类型筛选逻辑

- [ ] 4.1 解析订单类型字符串

  - File: `main/app/service/business.go` - `CountProductSales`
  - Purpose: 将订单类型字符串（如"1,2,3"）解析为数组
  - Requirements: 2.1-2.5
  - Leverage: 字符串处理工具
  - Prompt: Role: Go Developer | Task: 在 CountProductSales 方法中解析 OrderType 字符串，转换为订单类型数组 | Context: OrderType 格式为逗号分隔的字符串，如"1,2,3"，需要解析为 []uint | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 订单类型解析成功，数组正确

- [ ] 4.2 扩展 CountProductSale 方法支持订单类型筛选

  - File: `main/app/service/statistics.go`
  - Purpose: 在 CountProductSale 方法中传递订单类型参数
  - Requirements: 2.1-2.5
  - Leverage: 现有 CountProductSale 方法
  - Prompt: Role: Go Developer | Task: 在 CountProductSale 方法中新增 OrderTypes 参数，传递给 Repository | Context: 参考 design.md 中的参数传递逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 参数传递成功

- [ ] 4.3 实现订单类型筛选SQL逻辑

  - File: `main/app/repository/statistics.go` - `CountProductSale`
  - Purpose: 在 Repository 层实现订单类型筛选SQL逻辑
  - Requirements: 2.1-2.5
  - Leverage: 现有 CountProductSale 方法，design.md 中的SQL逻辑
  - Prompt: Role: Go Developer | Task: 在 CountProductSale 方法中实现订单类型筛选，关联 sale_bill 表，使用 bill_type 字段筛选 | Context: 参考 design.md 中的订单类型筛选逻辑，映射订单类型到 SaleBillType | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 LEFT JOIN | Success: 订单类型筛选SQL实现成功，查询结果正确

---

## Phase 5: 订单来源筛选逻辑

- [ ] 5.1 扩展 CountProductSaleRepoReq 支持订单来源

  - File: `main/app/repository/statistics.go`
  - Purpose: 在 Repository 请求参数中新增 OrderSource 字段
  - Requirements: 3.1-3.5
  - Leverage: 现有 CountProductSaleRepoReq 结构
  - Prompt: Role: Go Developer | Task: 在 CountProductSaleRepoReq 中新增 OrderSource 字段 | Context: 参考 design.md 中的参数定义 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 参数扩展成功

- [ ] 5.2 实现订单来源筛选SQL逻辑

  - File: `main/app/repository/statistics.go` - `CountProductSale`
  - Purpose: 在 Repository 层实现订单来源筛选SQL逻辑
  - Requirements: 3.1-3.5
  - Leverage: 现有 CountProductSale 方法，需要确认数据模型
  - Prompt: Role: Go Developer | Task: 在 CountProductSale 方法中实现订单来源筛选，仅在订单类型包含点餐订单时生效 | Context: 需要确认订单来源字段在哪个表，可能需要关联 order 表或 sale_bill 表 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 订单来源筛选SQL实现成功，查询结果正确

---

## Phase 6: 商品分类多选支持

- [ ] 6.1 扩展 CountProductSaleRepoReq 支持分类多选

  - File: `main/app/repository/statistics.go`
  - Purpose: 将 CategoryUuid 改为 CategoryUuids 支持多选
  - Requirements: 4.1-4.5
  - Leverage: 现有 CountProductSaleRepoReq 结构
  - Prompt: Role: Go Developer | Task: 在 CountProductSaleRepoReq 中将 CategoryUuid 改为 CategoryUuids []uint64 | Context: 参考 design.md 中的参数定义 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 参数修改成功

- [ ] 6.2 实现分类多选SQL逻辑

  - File: `main/app/repository/statistics.go` - `CountProductSale`
  - Purpose: 在 Repository 层实现分类多选SQL逻辑，支持查询子分类
  - Requirements: 4.1-4.5
  - Leverage: 现有分类筛选逻辑，design.md 中的SQL逻辑
  - Prompt: Role: Go Developer | Task: 在 CountProductSale 方法中实现分类多选逻辑，支持查询子分类 | Context: 参考 design.md 中的分类多选逻辑，如果分类UUID包含父分类，需要同时查询子分类 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 分类多选SQL实现成功，查询结果正确

- [ ] 6.3 处理向后兼容

  - File: `main/app/service/business.go` - `CountProductSales`
  - Purpose: 如果传入单个 CategoryUuid，自动转换为 CategoryUuids 数组
  - Requirements: 4.5
  - Leverage: 现有 CountProductSales 方法
  - Prompt: Role: Go Developer | Task: 在 CountProductSales 方法中处理向后兼容，如果 CategoryUuid > 0，转换为 CategoryUuids 数组 | Context: 保持API向后兼容，支持旧版本客户端 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 向后兼容处理成功

---

## Phase 7: 导出功能实现

- [ ] 7.1 新增导出类型常量

  - File: `main/app/model/export_record.go` 或 `main/app/constant/export_record.go`
  - Purpose: 新增商品销售统计导出类型常量
  - Requirements: 6.8
  - Leverage: 现有导出类型常量
  - Prompt: Role: Go Developer | Task: 新增 ExportTypeProductSales 常量，用于标识商品销售统计导出类型 | Context: 参考现有导出类型常量定义 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 常量定义成功

- [ ] 7.2 创建导出API Handler

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 新增 ExportProductSales API Handler
  - Requirements: 6.1
  - Leverage: 现有导出API Handler，如 ExportKitchenProductionDetail
  - Prompt: Role: Go Developer | Task: 新增 ExportProductSales Handler，调用 BusinessService 的导出方法 | Context: 参考 ExportKitchenProductionDetail 实现，路径为 /shop/statistics/product_sales/export | Restrictions: 遵循 .cursor/rules/go-main.mdc 和 api.mdc | Success: API Handler 创建成功，路由注册正确

- [ ] 7.3 实现导出方法

  - File: `main/app/service/business.go`
  - Purpose: 实现 ExportProductSales 方法，创建导出任务
  - Requirements: 6.1-6.8
  - Leverage: 现有导出方法 ExportKitchenProductionDetail，design.md 中的实现逻辑
  - Prompt: Role: Go Developer | Task: 实现 ExportProductSales 方法，检查数据量，创建导出任务，异步处理 | Context: 参考 design.md 中的导出功能实现，最多导出1000条数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 导出方法实现成功，任务创建正确

- [ ] 7.4 实现导出任务处理

  - File: `main/app/service/business.go`
  - Purpose: 实现 ExportProductSalesTask 方法，生成Excel文件
  - Requirements: 6.1-6.7
  - Leverage: 现有导出任务处理方法，如 ExportKitchenProductionDetailTask
  - Prompt: Role: Go Developer | Task: 实现 ExportProductSalesTask 方法，查询数据，生成Excel文件，上传到存储，更新导出记录状态 | Context: 参考 ExportKitchenProductionDetailTask 实现，导出字段包括：序号、商品名称、商品分类、销售数量、原价销售额、赠菜、实际销售额、营业收入 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 导出任务处理成功，Excel文件生成正确

- [ ] 7.5 注册导出API路由

  - File: `main/app/api/v1/shop/shop_statistics.go` - `RegisterStatisticsHandlers`
  - Purpose: 注册导出API路由
  - Requirements: 6.1
  - Leverage: 现有路由注册逻辑
  - Prompt: Role: Go Developer | Task: 在 RegisterStatisticsHandlers 中注册导出API路由 | Context: 路由路径为 /shop/statistics/product_sales/export，使用 GET 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 路由注册成功

---

## Phase 8: 测试和验证

- [ ] 8.1 单元测试：时间类型计算

  - File: `main/pkg/utils/time_test.go`
  - Purpose: 测试新增的时间工具方法
  - Requirements: 1.1-1.7
  - Leverage: 现有测试文件
  - Prompt: Role: Go Developer | Task: 编写时间工具方法的单元测试，验证时间计算准确性 | Context: 测试近7天、上月、今年的时间计算 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试通过，时间计算准确

- [ ] 8.2 集成测试：筛选功能

  - File: 测试文件或Postman集合
  - Purpose: 测试所有筛选条件的组合
  - Requirements: 1-5
  - Leverage: 现有API测试
  - Prompt: Role: QA Engineer | Task: 编写集成测试，测试时间类型、订单类型、订单来源、商品分类等筛选条件 | Context: 测试各种筛选条件的组合，验证查询结果正确性 | Restrictions: 遵循测试规范 | Success: 所有测试用例通过

- [ ] 8.3 集成测试：导出功能

  - File: 测试文件或Postman集合
  - Purpose: 测试导出功能
  - Requirements: 6
  - Leverage: 现有导出功能测试
  - Prompt: Role: QA Engineer | Task: 编写导出功能测试，验证导出文件格式和内容 | Context: 测试导出文件下载、Excel格式、数据准确性 | Restrictions: 遵循测试规范 | Success: 导出功能测试通过

---

## 📝 开发注意事项

### 代码复用

- 时间工具方法：复用 `utils.SetTimezone` 和相关方法
- 导出功能：参考 `ExportKitchenProductionDetail` 实现
- 筛选逻辑：复用现有 `buildCountOpts` 和 DBOption 模式

### 性能优化

- 时间范围筛选：使用索引字段
- 分类多选：使用 `IN` 查询，避免多次查询
- 订单类型筛选：使用 `IN` 查询，利用索引

### 向后兼容

- 保持现有API参数兼容
- 如果传入旧参数格式，自动转换

### 错误处理

- 参数验证：使用 binding 标签
- 数据量限制：导出最多1000条
- 时间范围验证：不能选择未来时间

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**维护者**: 开发组

