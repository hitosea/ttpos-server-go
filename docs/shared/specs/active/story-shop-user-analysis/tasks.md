# 用户分析统计 任务分解

> 本文档定义用户分析统计功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_nationality_uuid_to_statistics_sale.php`
  - Purpose: 为 `ttpos_statistics_sale` 表新增 `nationality_uuid` 字段
  - Requirements: R1.4, R2.1, R3.1, R4.1, R5.1
  - Leverage: 现有迁移文件: `admin/database/migrations/20251120083419_alter_ttpos_order_add_source_nationality_fields.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，为 ttpos_statistics_sale 表新增 nationality_uuid 字段 | Context: 字段类型 bigint(20) unsigned，默认值 0，添加索引 idx_nationality_uuid | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中新增字段
  - Requirements: R1.4, R2.1, R3.1, R4.1, R5.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [ ] 1.3 更新 Go Model

  - File: `main/app/model/statistics.go`
  - Purpose: 在 `StatisticsSale` 结构体中新增 `NationalityUuid` 字段
  - Requirements: R1.4, R2.1, R3.1, R4.1, R5.1
  - Leverage: 现有 Model: `main/app/model/statistics.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 StatisticsSale 结构体中新增 NationalityUuid 字段 | Context: 使用 gorm 标签，类型 uint64，默认值 0，添加 json 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [ ] 2.1 实现 Repository `CountUserAnalysis` 方法

  - File: `main/app/repository/statistics.go`
  - Purpose: 实现用户分析统计查询逻辑（四个维度）
  - Requirements: R2.*, R3.*, R4.*, R5.*, R6.3
  - Leverage: 现有统计方法: `main/app/repository/statistics.go`，`WhereNotInDataManageSubQuery`: `main/app/repository/common.go`
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 CountUserAnalysis 方法，按四个维度统计订单数和占比 | Context: 1) 国籍统计：从 statistics_sale 表查询，先检查是否存在 nationality_uuid > 0 的订单，如果所有订单的 nationality_uuid = 0 则返回空数组，否则仅统计 nationality_uuid > 0 的订单并关联 nationality 表获取名称；2) 点餐方式来源：仅点餐订单，order_source_uuid=0为"店内"，>0时关联 order_source 表获取来源名称（通过 multi_language_name_uuid 关联多语言名称表）；3) 桌台方式来源：仅桌台订单，从 sale_bill.source 获取；4) 用餐方式：从 sale_bill.dining_method 获取（0=店内用餐，1=打包），桌台订单统一为店内用餐 | Restrictions: 排除数据管理订单，按订单数升序排序，占比计算使用 decimal 类型（decimal.NewFromInt(orderCount).Div(decimal.NewFromInt(totalCount)).Mul(decimal.NewFromInt(100)).Round(2)），最后转换为 float64 返回 | Success: Repository 实现完整，四个维度统计正确，占比计算精确，国籍统计前置检查正确

- [ ] 2.2 编写 Repository 单元测试

  - File: `main/app/repository/statistics_user_analysis_test.go`
  - Purpose: 确保 Repository 统计逻辑正确
  - Requirements: R2.*, R3.*, R4.*, R5.*
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CountUserAnalysis 编写单元测试，覆盖率 ≥ 80% | Context: 测试四个维度统计，测试数据管理过滤，测试排序和占比计算，测试国籍统计的前置检查（所有订单 nationality_uuid = 0 时返回空数组） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

### DTO 层

- [ ] 2.3 创建 Response DTO

  - File: `main/app/dto/resp/statistics_user_analysis_resp.go`
  - Purpose: 定义 API 响应数据结构
  - Requirements: R6.4
  - Leverage: 现有 DTO: `main/app/dto/resp/statistics_channel_resp.go`
  - Prompt: Role: Go Developer | Task: 创建 UserAnalysisResp 和 UserAnalysisItem 结构体 | Context: 包含四个维度数组，每个维度包含名称、订单数、占比 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

### Service 层

- [ ] 2.4 更新 `SaveSale` 方法保存 `nationality_uuid`

  - File: `main/app/service/statistics.go`
  - Purpose: 在保存统计时保存国籍信息
  - Requirements: R1.4, R2.1
  - Leverage: 现有 SaveSale 方法: `main/app/service/statistics.go`，SaleBill 模型: `main/app/model/sale_bill.go`
  - Prompt: Role: Go Developer | Task: 在 SaveSale 方法中保存 NationalityUuid 字段 | Context: 从 saleBill.NationalityUuid 获取值，赋值给 StatisticsSale.NationalityUuid | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: SaveSale 方法更新成功，nationality_uuid 正确保存

- [ ] 2.5 扩展 `IBusinessSrv` 接口

  - File: `main/app/service/i_business_srv.go`
  - Purpose: 定义用户分析统计接口方法
  - Requirements: R6.*, R7.*
  - Leverage: 现有接口: `main/app/service/i_business_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 IBusinessSrv 接口中新增 CountUserAnalysis 和 ExportUserAnalysis 方法 | Context: CountUserAnalysis 返回 UserAnalysisResp，ExportUserAnalysis 返回 error | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 2.6 实现 Service `CountUserAnalysis` 方法

  - File: `main/app/service/business.go`
  - Purpose: 实现用户分析统计业务逻辑
  - Requirements: R6.*
  - Leverage: 现有 Service 实现: `main/app/service/business.go`，Task 2.1 的 Repository，时间工具: `pkg/timeutil/company_time.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 CountUserAnalysis 方法，调用 Repository 并转换响应格式 | Context: 获取今天时间范围（使用门店时区），调用 CountUserAnalysis，Repository 返回的 Percentage 为 decimal.Decimal 类型，转换为 float64（使用 .InexactFloat64() 或 .Float64()），转换响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，占比计算使用 decimal 保证精度 | Success: Service 实现完整，业务逻辑正确，占比计算精确

- [ ] 2.7 实现 Service `ExportUserAnalysis` 方法

  - File: `main/app/service/business.go`
  - Purpose: 实现用户分析统计导出功能
  - Requirements: R7.*
  - Leverage: 现有导出实现: `main/app/service/business.go` 中的 `ExportChannelSales`，导出工具: `pkg/excel/`
  - Prompt: Role: Go Developer with export expertise | Task: 实现 ExportUserAnalysis 方法，创建导出任务并异步处理 | Context: 检查是否有正在导出的任务，获取统计数据，创建 ExportRecord，异步生成 Excel 文件 | Restrictions: 遵循 .cursor/rules/go-main.mdc，支持多语言文件名 | Success: Service 实现完整，导出功能正确

- [ ] 2.8 编写 Service 单元测试

  - File: `main/app/service/business_user_analysis_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: R6.*, R7.*
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CountUserAnalysis 和 ExportUserAnalysis 编写单元测试，覆盖率 ≥ 70% | Context: 测试统计查询，测试占比计算，测试导出任务创建 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### API 层

- [ ] 2.9 创建 API Handler

  - File: `main/app/api/v1/shop/shop_statistics.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: R6.1, R7.1
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_statistics.go`，Task 2.5-2.7 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 UserAnalysis 和 ExportUserAnalysis API Handler | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 2.10 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 API 路由
  - Requirements: R6.1, R7.1
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功，权限中间件覆盖

- [ ] 2.11 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_statistics_test.go`
  - Purpose: 测试 API 接口
  - Requirements: R6.*, R7.*
  - Leverage: 现有测试: `main/app/api/v1/shop/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 UserAnalysis 和 ExportUserAnalysis API 编写集成测试 | Context: 测试查询接口返回格式，测试导出接口任务创建，测试权限校验 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: 导出功能

- [ ] 3.1 实现 Excel 导出模板

  - File: `pkg/excel/user_analysis_template.go`（或现有导出工具）
  - Purpose: 生成用户分析统计 Excel 文件
  - Requirements: R7.3, R7.5
  - Leverage: 现有导出工具: `pkg/excel/`，参考 `ExportChannelSales` 的实现
  - Prompt: Role: Go Developer with Excel export expertise | Task: 实现用户分析统计 Excel 导出模板 | Context: 包含四个统计维度的工作表，每个工作表包含名称、订单数、占比列，支持多语言表头 | Restrictions: 遵循现有导出格式规范 | Success: Excel 模板实现完成，格式正确

---

## Phase 4: 测试和优化

- [ ] 4.1 集成测试

  - File: `test/integration/user_analysis_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户分析统计查询，测试导出功能，测试数据管理过滤 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 查询响应时间 < 500ms

- [ ] 4.3 数据库查询优化

  - File: `main/app/repository/statistics.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 500ms

---

## Phase 5: 文档更新

- [ ] 5.1 更新 API 文档

  - File: `docs/shared/api/statistics.md`, `main/docs/swagger.yaml`
  - Purpose: 记录新接口
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档已更新

- [ ] 5.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Success: CHANGELOG 已更新

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
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-user-analysis/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-user-analysis/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-user-analysis/tasks.md
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
**最后更新**: 2025-11-26  
**维护者**: 后端开发组

