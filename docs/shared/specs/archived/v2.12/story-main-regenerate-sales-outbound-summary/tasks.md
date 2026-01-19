# 重新生成每日销售出库汇总记录 任务分解

> 本文档定义重新生成每日销售出库汇总记录功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 11  
**进行中**: -  
**完成率**: 61.1%

---

## Phase 1: 核心 Service 实现

### DTO 层

- [x] 1.1 创建 Request DTO

  - File: `main/app/dto/req/sales_outbound_summary_req.go`
  - Purpose: 定义重新生成请求参数
  - Requirements: 5.2
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 RegenerateSalesOutboundSummaryReq 结构体，包含 company_uuid 和 date 字段 | Context: 使用 binding 标签验证参数，date 格式为 YYYY-MM-DD | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [x] 1.2 创建 Response DTO

  - File: `main/app/dto/resp/sales_outbound_summary_resp.go`
  - Purpose: 定义重新生成响应数据
  - Requirements: 5.3
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 RegenerateSalesOutboundSummaryResp 结构体，包含 deleted_count、generated_count、duration_ms 字段 | Context: data 必须是对象，不能是 null 或数组 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: DTO 创建成功，响应格式正确

### Service 层

- [x] 1.3 创建 Service 接口

  - File: `main/app/service/i_sales_outbound_summary_service.go`
  - Purpose: 定义业务逻辑接口
  - Requirements: 2.1, 2.2
  - Leverage: 现有 Service 接口: `main/app/service/i_*_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 ISalesOutboundSummarySrv 接口，定义 RegenerateSalesOutboundSummary 方法 | Context: 方法接收 companyUuid 和 date 参数，返回响应和 error | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [x] 1.4 提取公共方法（从 DailySalesOutboundSummaryTask）

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 从定时任务中提取可复用的统计逻辑
  - Requirements: 2.1, 2.2
  - Leverage: `main/app/tasks/daily_sales_outbound_summary.go` 中的 `getDailySalesOutboundRecords`、`saveOutboundSummaryRecords`、`getOpeningHours`、`generateOrderNo` 方法
  - Prompt: Role: Go Developer with refactoring expertise | Task: 从 DailySalesOutboundSummaryTask 中提取公共方法到 Service，包括获取销售出库记录、保存汇总记录、获取营业时段、生成出库单号等方法 | Context: 保持原有逻辑不变，仅提取为可复用方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只能依赖其他 Service 接口 | Success: 公共方法提取成功，逻辑保持一致

- [x] 1.5 实现删除旧记录方法

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 实现删除指定日期旧记录的逻辑
  - Requirements: 1.1-1.5
  - Leverage: `main/app/repository/warehouse_in_out_log.go`，参考 `main/app/service/cost_card_correction_service.go` 中的删除逻辑
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 deleteOldRecords 方法，根据 opening_hours 匹配指定日期的所有营业时段记录，软删除 log_type=1 且 scene=1 的记录 | Context: 使用事务保证原子性，返回删除的记录数 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用软删除 | Success: 删除逻辑正确，事务管理正确

- [x] 1.6 实现重新生成记录方法

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 实现重新生成汇总记录的逻辑
  - Requirements: 2.1-2.9
  - Leverage: Task 1.4 提取的公共方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 generateNewRecords 方法，调用提取的公共方法获取销售出库数据并保存汇总记录 | Context: 根据门店营业时段配置计算时间范围，生成 opening_hours 字符串，更新 is_summarized 状态 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用事务保证一致性 | Success: 生成逻辑正确，事务管理正确

- [x] 1.7 实现主 Service 方法（整合删除和生成）

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 实现 RegenerateSalesOutboundSummary 主方法，整合删除和生成逻辑
  - Requirements: 1, 2, 6
  - Leverage: Task 1.5, 1.6，`pkg/lock` 分布式锁
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 RegenerateSalesOutboundSummary 主方法，包括分布式锁控制、参数验证、调用删除和生成方法、返回结果 | Context: 使用分布式锁防止并发操作，锁的 key 为 regenerate_sales_outbound_summary:{company_uuid}:{date}，超时时间 5 分钟 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 主方法实现完整，分布式锁正确，错误处理完善

- [ ] 1.8 编写 Service 单元测试

  - File: `main/app/service/sales_outbound_summary_service_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SalesOutboundSummarySrv 编写单元测试，覆盖率 ≥ 70% | Context: 测试删除旧记录、重新生成记录、分布式锁、错误处理等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 2: API 和命令行工具

### API 层

- [x] 2.1 创建 API Controller 方法

  - File: `main/app/api/v1/shop/shop_warehouse.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: 5.1-5.5
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_warehouse.go`，Task 1.3-1.7 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 WarehouseHandler 中新增 RegenerateSalesOutboundSummary 方法，实现 RESTful 接口 | Context: URL 为 /api/shop/inventory/regenerate-sales-outbound-summary，使用 helper.Success() 返回响应，需要权限校验（仅管理员） | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，权限校验正确

- [x] 2.2 注册 API 路由

  - File: `main/app/api/v1/shop/shop_warehouse.go`
  - Purpose: 注册 API 路由
  - Requirements: 5.1
  - Leverage: 现有路由: `main/app/api/v1/shop/shop_warehouse.go` 的 RegisterWarehouseHandlers
  - Success: 路由注册成功，路径为 `/api/v1/shop/inventory/regenerate-sales-outbound-summary`

- [ ] 2.3 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_warehouse_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/shop/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 RegenerateSalesOutboundSummary API 编写集成测试 | Context: 测试正常场景、参数验证、权限校验、错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

### Command 层

- [x] 2.4 创建命令行工具

  - File: `main/command/regenerate_sales_outbound.go`
  - Purpose: 实现命令行工具，支持批量操作
  - Requirements: 4.1-4.6
  - Leverage: 现有命令: `main/command/add_item_stock.go`、`main/command/statistics_re.go`
  - Prompt: Role: Go Developer with Cobra expertise | Task: 创建 regenerate-sales-outbound 子命令，支持 --company-uuid、--date、--dry-run 参数 | Context: 使用 cobra.Command，在 PreRun 中初始化配置和数据库，在 Run 中调用 Service 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 命令行工具创建成功，参数解析正确，dry-run 模式正确

- [x] 2.5 注册命令行工具

  - File: `main/command/regenerate_sales_outbound.go`
  - Purpose: 将命令注册到 rootCommand
  - Requirements: 4.1
  - Leverage: Task 2.4
  - Success: 命令注册成功，可通过 `./main regenerate-sales-outbound --company-uuid=xxx --date=YYYY-MM-DD` 调用

---

## Phase 3: Vue 前端模块

- [ ] 3.1 创建 API 封装

  - File: `admin/views/shop/api/inventory.ts`
  - Purpose: 封装后端 API 调用
  - Requirements: 3.1-3.7
  - Leverage: 现有 API: `admin/views/shop/api/`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装 regenerateSalesOutboundSummary API 调用 | Context: 使用 axios，定义 TypeScript 类型，包含 company_uuid 和 date 参数 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

- [ ] 3.2 修改出入库记录列表页面

  - File: `admin/views/shop/pages/inventory/log/index.vue`
  - Purpose: 在列表页面添加"重新生成"按钮
  - Requirements: 3.1-3.7
  - Leverage: 现有页面: `admin/views/shop/pages/inventory/log/index.vue`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在出入库记录列表页面添加"重新生成"按钮，点击后弹出日期选择对话框，确认后调用 API | Context: 使用 Element Plus 组件，按钮仅管理员可见，显示确认对话框和操作结果 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Composition API | Success: 页面修改成功，功能完整，权限控制正确

- [ ] 3.3 添加多语言支持

  - File: `admin/views/shop/pages/inventory/log/index.vue` 及相关语言文件
  - Purpose: 添加前端文案的多语言支持
  - Requirements: 国际化要求
  - Leverage: 现有多语言文件
  - Success: 多语言支持完成，所有文案可国际化

---

## Phase 4: 测试和优化

- [ ] 4.1 集成测试

  - File: `test/integration/regenerate_sales_outbound_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试删除旧记录 + 重新生成的完整流程，测试数据一致性，测试并发控制 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 并发测试

  - File: `test/integration/regenerate_sales_outbound_concurrent_test.go`
  - Purpose: 测试分布式锁有效性
  - Requirements: 6.2-6.5
  - Leverage: 现有测试工具
  - Prompt: Role: QA Engineer specializing in concurrency testing | Task: 测试同一门店同一日期的并发操作，验证分布式锁有效性 | Context: 模拟多个并发请求，验证只有一个请求成功，其他请求返回"操作进行中" | Restrictions: 使用 goroutine 模拟并发 | Success: 并发测试通过，分布式锁有效

- [ ] 4.3 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 单次操作响应时间 < 5 秒（正常数据量）

- [ ] 4.4 数据库查询优化

  - File: `main/app/repository/warehouse_in_out_log.go`
  - Purpose: 优化查询性能
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 1 秒，索引使用正确

- [ ] 4.5 文档更新

  - File: `docs/shared/api/inventory_api.md`（如有），`CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%（复用现有，无需新增）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-regenerate-sales-outbound-summary/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-regenerate-sales-outbound-summary/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-regenerate-sales-outbound-summary/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-regenerate-sales-outbound-summary/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-regenerate-sales-outbound-summary/tasks.md)" | bc
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
**最后更新**: 2025-12-15  
**维护者**: 后端开发组

