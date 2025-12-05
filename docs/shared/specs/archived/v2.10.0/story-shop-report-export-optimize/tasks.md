# 优化新管理端导出报表名称 任务分解

> 本文档定义优化新管理端导出报表名称功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 19  
**已完成**: 16  
**进行中**: -  
**完成率**: 84%

---

## Phase 1: Repository 层扩展

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 确认数据库连接已包含商户隔离

  - File: `main/app/service/business.go`
  - Purpose: 确认数据库连接已包含商户隔离，无需额外的 company_uuid 过滤
  - Requirements: 2.1, 2.5
  - Leverage: 现有代码: `main/app/service/business.go`，Context 的数据库连接
  - Success: 确认数据库连接已包含商户隔离，无需 company_uuid

- [x] 1.2 新增 GetByDateAndType 方法到 IExportRecordRepo 接口

  - File: `main/app/repository/export_record.go`
  - Purpose: 定义查询同一天导出记录的接口方法
  - Requirements: 2.1, 2.5
  - Leverage: 现有 Repository 接口: `main/app/repository/export_record.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 在 IExportRecordRepo 接口中新增 GetByDateAndType 方法 | Context: 方法签名: GetByDateAndType(exportType uint8, startTime, endTime int64) ([]model.ExportRecord, error)，无需 company_uuid 参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口方法定义完整
  - Success: 接口方法定义完成 ✅

- [x] 1.3 实现 GetByDateAndType 方法

  - File: `main/app/repository/export_record.go`
  - Purpose: 实现查询同一天导出记录的逻辑
  - Requirements: 2.1, 2.3, 2.5
  - Leverage: 现有 Repository 实现: `main/app/repository/export_record.go`，参考 GetUnfinishedExportRecord 方法
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 GetByDateAndType 方法，查询指定日期范围内的导出记录 | Context: 查询条件: export_type, create_time 范围, status=1(成功), delete_time=0，无需 company_uuid（数据库连接已包含商户隔离） | Restrictions: 使用 GORM，参数化查询，遵循 .cursor/rules/go-main.mdc | Success: 方法实现完整，查询逻辑正确
  - Success: 方法实现完成，查询逻辑正确 ✅

- [x] 1.4 添加数据库索引优化查询性能

  - File: `admin/database/migrations/20251201164509_add_index_to_export_record.php`
  - Purpose: 优化查询同一天导出记录的性能
  - Requirements: 性能要求
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，为 ttpos_export_record 表添加索引 | Context: 索引字段: export_type, status, create_time（无需 company_uuid） | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查索引是否存在 | Success: 索引创建成功，查询性能提升
  - Command: `cd admin && php think migrate:create AddIndexToExportRecord`
  - Success: 索引创建成功 ✅（索引名称：idx_export_type_status_date，包含 status 字段）

- [ ] 1.5 编写 Repository 单元测试

  - File: `main/app/repository/export_record_test.go`
  - Purpose: 确保 GetByDateAndType 方法正确性
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/repository/export_record_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetByDateAndType 方法编写单元测试，覆盖率 ≥ 80% | Context: 测试查询逻辑，测试日期范围，测试过滤条件 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: 测试覆盖率达标，所有测试通过

---

## Phase 2: Service 层文件名生成

- [x] 2.1 创建 generateExportFileName 工具方法

  - File: `main/app/service/business.go`
  - Purpose: 统一文件名生成逻辑，支持日期格式和自动编号
  - Requirements: 1.1-1.5, 2.1-2.5
  - Leverage: 现有 Service: `main/app/service/business.go`，时区工具: `main/app/utils`，Repository: Task 1.3
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 businessSrv 中创建 generateExportFileName 方法，生成格式为"报表名YYYY-MM-DD.xlsx"的文件名 | Context: 使用商户时区，查询同一天导出记录（无需 company_uuid），计算序号 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法创建成功，文件名格式正确
  - Success: 方法创建成功，逻辑正确 ✅（已修复文件名后缀问题）

- [x] 2.2 修改 ExportBusinessTimePeriod 文件名生成

  - File: `main/app/service/business.go` (约第 2263 行)
  - Purpose: 时段营业统计导出使用新的文件名格式
  - Requirements: 1.1, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Prompt: Role: Go Developer | Task: 修改 ExportBusinessTimePeriod 方法，使用 generateExportFileName 生成文件名 | Context: 替换原有的时间戳格式文件名生成逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 文件名格式改为日期格式
  - Success: 文件名格式修改完成 ✅

- [x] 2.3 修改 ExportBusinessSummary 文件名生成

  - File: `main/app/service/business.go` (约第 2467 行)
  - Purpose: 综合运营统计导出使用新的文件名格式
  - Requirements: 1.2, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Success: 文件名格式修改完成 ✅

- [x] 2.4 修改 ExportBusinessPaymentMethod 文件名生成

  - File: `main/app/service/business.go` (约第 2676 行)
  - Purpose: 营业收款统计导出使用新的文件名格式
  - Requirements: 1.3, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Success: 文件名格式修改完成 ✅

- [x] 2.5 修改 ExportChannelSales 文件名生成

  - File: `main/app/service/business.go` (约第 2953 行)
  - Purpose: 渠道营业统计导出使用新的文件名格式
  - Requirements: 1.4, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Success: 文件名格式修改完成 ✅

- [x] 2.6 修改 ExportProductSales 文件名生成

  - File: `main/app/service/business.go` (约第 789 行)
  - Purpose: 商品销售统计导出使用新的文件名格式
  - Requirements: 1.5, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Success: 文件名格式修改完成 ✅

- [x] 2.7 修改 ExportUserAnalysis 文件名生成

  - File: `main/app/service/business.go` (约第 3554 行)
  - Purpose: 用户分析导出使用新的文件名格式（注意：不修改子表名称）
  - Requirements: 1.6, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Success: 文件名格式修改完成 ✅

- [x] 2.8 修改 ExportKitchenProductionDetail 文件名生成

  - File: `main/app/service/business.go` (约第 1652 行)
  - Purpose: 后厨菜品出品明细导出使用新的文件名格式
  - Requirements: 1.7, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Success: 文件名格式修改完成 ✅

- [x] 2.9 修改 ExportKitchenEfficiencyAnalysis 文件名生成

  - File: `main/app/service/business.go` (约第 1522 行)
  - Purpose: 后厨效率分析导出使用新的文件名格式
  - Requirements: 1.8, 2.1
  - Leverage: Task 2.1 的 generateExportFileName 方法
  - Success: 文件名格式修改完成 ✅

- [ ] 2.10 实现并发导出控制（事务处理）

  - File: `main/app/service/business.go`
  - Purpose: 确保并发导出时文件名唯一性，避免冲突
  - Requirements: 2.4
  - Leverage: 现有事务处理: `main/app/service/business.go`，参考其他导出方法的事务使用
  - Prompt: Role: Go Developer with transaction expertise | Task: 在 generateExportFileName 或导出方法中使用数据库事务，确保文件名唯一性 | Context: 查询和创建导出记录在同一事务中完成 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 并发导出测试通过，文件名不冲突
  - Success: 并发控制实现完成

- [ ] 2.11 编写 Service 单元测试（文件名生成）

  - File: `main/app/service/business_test.go`
  - Purpose: 确保文件名生成逻辑正确性
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/business_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 generateExportFileName 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试日期格式，测试序号计算，测试时区处理，测试并发场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过
  - Success: 测试覆盖率达标，所有测试通过

---

## Phase 3: Excel 子表名称修改

- [x] 3.1 修改 ExportProductSalesTask 子表名称

  - File: `main/app/service/business.go` (约第 908-920 行)
  - Purpose: 商品销售统计 Excel 子表名称改为 Sheet1
  - Requirements: 3.1, 3.3
  - Leverage: 现有代码: `main/app/service/business.go`
  - Prompt: Role: Go Developer | Task: 修改 ExportProductSalesTask 方法，将子表名称从多语言"报表"改为"Sheet1" | Context: 删除 sheetNameMul 多语言定义，直接使用 "Sheet1" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 子表名称修改完成
  - Success: 子表名称修改完成 ✅

- [x] 3.2 修改 ExportBusinessTimePeriodTask 子表名称

  - File: `main/app/service/business.go` (约第 2379 行)
  - Purpose: 时段营业统计 Excel 子表名称改为 Sheet1
  - Requirements: 3.1, 3.3
  - Leverage: Task 3.1 的修改方式
  - Success: 子表名称修改完成 ✅

- [x] 3.3 修改 ExportBusinessSummaryTask 子表名称

  - File: `main/app/service/business.go` (约第 2586 行)
  - Purpose: 综合运营统计 Excel 子表名称改为 Sheet1
  - Requirements: 3.1, 3.3
  - Leverage: Task 3.1 的修改方式
  - Success: 子表名称修改完成 ✅

- [x] 3.4 修改 ExportBusinessPaymentMethodTask 子表名称

  - File: `main/app/service/business.go` (约第 2799 行)
  - Purpose: 营业收款统计 Excel 子表名称改为 Sheet1
  - Requirements: 3.1, 3.3
  - Leverage: Task 3.1 的修改方式
  - Success: 子表名称修改完成 ✅

- [x] 3.5 修改 ExportChannelSalesTask 子表名称

  - File: `main/app/service/business.go` (约第 3120 行)
  - Purpose: 渠道营业统计 Excel 子表名称改为 Sheet1
  - Requirements: 3.1, 3.3
  - Leverage: Task 3.1 的修改方式
  - Success: 子表名称修改完成 ✅

- [x] 3.6 修改 ExportKitchenProductionDetailTask 子表名称

  - File: `main/app/service/business.go` (约第 2004 行)
  - Purpose: 后厨菜品出品明细 Excel 子表名称改为 Sheet1
  - Requirements: 3.1, 3.3
  - Leverage: Task 3.1 的修改方式
  - Success: 子表名称修改完成 ✅

- [x] 3.7 修改 ExportKitchenEfficiencyAnalysisTask 子表名称

  - File: `main/app/service/business.go` (约第 2137 行)
  - Purpose: 后厨效率分析 Excel 子表名称改为 Sheet1
  - Requirements: 3.1, 3.3
  - Leverage: Task 3.1 的修改方式
  - Success: 子表名称修改完成 ✅

- [x] 3.8 验证 ExportUserAnalysisTask 子表名称保持不变

  - File: `main/app/service/business.go` (约第 3647 行)
  - Purpose: 确认用户分析报表的子表名称逻辑保持不变
  - Requirements: 3.2
  - Leverage: 现有代码: `main/app/service/business.go`
  - Prompt: Role: QA Engineer | Task: 检查 ExportUserAnalysisTask 方法，确认子表名称逻辑未修改 | Context: 用户分析报表可能有多个子表，需要保持原有逻辑 | Restrictions: 不修改代码，仅验证 | Success: 确认子表名称逻辑未修改
  - Success: 验证通过，子表名称逻辑保持不变 ✅

---

## Phase 4: 测试和优化

- [ ] 4.1 集成测试（端到端导出流程）

  - File: `test/integration/export_test.go`
  - Purpose: 测试完整的导出流程，验证文件名格式和子表名称
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，测试所有报表类型的导出功能 | Context: 测试文件名格式，测试序号计算，测试子表名称，测试时区处理 | Restrictions: 测试真实用户场景 | Success: 集成测试通过
  - Success: 集成测试通过

- [ ] 4.2 并发测试（文件名冲突处理）

  - File: `test/integration/export_concurrent_test.go`
  - Purpose: 测试并发导出场景，确保文件名不冲突
  - Requirements: 2.4
  - Leverage: Go 并发测试工具
  - Prompt: Role: QA Engineer specializing in concurrency testing | Task: 编写并发测试，模拟同一天多次导出同名报表 | Context: 使用 goroutine 并发执行导出，验证文件名序号递增 | Restrictions: 测试并发安全性 | Success: 并发测试通过，文件名不冲突
  - Success: 并发测试通过

- [ ] 4.3 时区处理测试

  - File: `test/integration/export_timezone_test.go`
  - Purpose: 测试不同时区商户的文件名日期正确性
  - Requirements: 1.1
  - Leverage: 时区工具类测试
  - Prompt: Role: QA Engineer | Task: 编写时区测试，验证不同时区商户的文件名日期正确 | Context: 测试多个时区（UTC+8, UTC+0, UTC-5等），验证日期格式 | Restrictions: 测试时区边界情况 | Success: 时区测试通过
  - Success: 时区测试通过

- [ ] 4.4 性能测试（查询优化）

  - File: -
  - Purpose: 确保查询同一天导出记录的性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Prompt: Role: Performance Engineer | Task: 测试 GetByDateAndType 方法的查询性能 | Context: 使用索引后，查询时间应 < 50ms | Restrictions: 测试大数据量场景 | Success: 查询时间 < 50ms
  - Success: 性能测试通过

- [ ] 4.5 回归测试（现有功能不受影响）

  - File: `test/integration/export_regression_test.go`
  - Purpose: 确保现有导出功能不受影响
  - Requirements: 可靠性要求
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行回归测试，确保现有导出功能正常工作 | Context: 测试所有报表类型的导出功能，验证数据正确性 | Restrictions: 全面覆盖 | Success: 回归测试通过
  - Success: 回归测试通过

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
- [ ] 8 个报表的文件名格式已优化
- [ ] 7 个报表的子表名称已修改（用户分析除外）

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] 技术文档已更新（如有需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-report-export-optimize/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-report-export-optimize/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-report-export-optimize/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-report-export-optimize/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-report-export-optimize/tasks.md)" | bc
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

## 附录：标准 Prompt 模板

### Go 后端开发

```
Role: Go Developer specializing in {具体领域}

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc, .cursor/rules/database.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Service) 或 ≥ 80% (Repository)
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-01  
**维护者**: 后端开发组

