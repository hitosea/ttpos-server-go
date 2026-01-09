# 新管理端-业务设置-调拨规则 任务分解

> 本文档定义调拨规则配置功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 21  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_transfer_rule_table.php`
  - Purpose: 定义 ttpos_transfer_rule 表结构
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考: `admin/database/migrations/*_create_ttpos_transfer_order_table.php`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_transfer_rule 表的迁移文件，遵循 requirements.md 中的数据库设计 | Context: 必须包含 id, uuid, shop_id, merchant_id, allow_transfer_in, allow_transfer_out, create_time, update_time, delete_time 字段，唯一索引 uk_shop(shop_id, delete_time) | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查表是否存在，检查索引是否存在 | Success: 迁移文件创建成功，字段定义正确，索引定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建 ttpos_transfer_rule 表
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，表已创建，索引已创建

- [ ] 1.3 创建 Go Model

  - File: `main/app/model/transfer_rule.go`
  - Purpose: 定义 Go 数据模型，与数据库表对应
  - Requirements: Requirement 1
  - Leverage: 现有 Model: `main/app/model/transfer_order.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 创建 TransferRule 结构体，映射到 ttpos_transfer_rule 表 | Context: 使用 gorm 标签，包含所有字段（Id, Uuid, ShopId, MerchantId, AllowTransferIn, AllowTransferOut, CreateTime, UpdateTime, DeleteTime），实现 TableName() 方法返回 "ttpos_transfer_rule" | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段使用 uint64 和 int8 类型 | Success: Model 创建成功，字段映射正确，TableName() 方法正确

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [ ] 2.1 创建 Repository 接口

  - File: `main/app/repository/i_transfer_rule_repo.go`
  - Purpose: 定义数据访问接口
  - Requirements: Requirement 1, Requirement 3
  - Leverage: 现有 Repository 接口: `main/app/repository/i_transfer_order_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 ITransferRuleRepo 接口，定义 CRUD 方法和选项方法 | Context: 使用选项模式(DBOption)，包含 Create, Update, GetByShopId, GetList 方法，选项方法包含 WhereShopId, WhereMerchantId, Offset, Limit | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 接口定义完整，方法签名正确

- [ ] 2.2 实现 Repository（选项模式）

  - File: `main/app/repository/transfer_rule.go`
  - Purpose: 实现数据访问逻辑
  - Requirements: Requirement 1, Requirement 3
  - Leverage: 现有 Repository 实现: `main/app/repository/transfer_order.go`，使用选项模式
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 TransferRuleRepoImpl，使用选项模式实现灵活查询 | Context: 只持有 db \*gorm.DB，实现所有接口方法和选项方法，GetByShopId 查询单条记录，GetList 查询列表并返回 total | Restrictions: 不能持有 DBManager，使用 GORM，软删除(delete_time=0)，使用选项模式 | Success: Repository 实现完整，选项模式正确，软删除正确

- [ ] 2.3 编写 Repository 单元测试

  - File: `main/app/repository/transfer_rule_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: Requirement 1
  - Leverage: 现有测试: `main/app/repository/transfer_order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 TransferRuleRepo 编写单元测试，覆盖率 ≥ 80% | Context: 测试 Create, Update, GetByShopId, GetList 方法，测试选项方法，测试软删除，测试唯一索引约束 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用测试数据库 | Success: 测试覆盖率 ≥ 80%，所有测试通过

### DTO 层

- [ ] 2.4 创建 Request DTO

  - File: `main/app/dto/req/transfer_rule_req.go`
  - Purpose: 定义 API 请求参数
  - Requirements: Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有 DTO: `main/app/dto/req/transfer_order.go`
  - Prompt: Role: Go Developer | Task: 创建 Request DTO，包含 TransferRuleSaveReq, TransferRuleGetReq, TransferRuleListReq 结构体 | Context: TransferRuleSaveReq 包含 ShopId, AllowTransferIn, AllowTransferOut，使用 binding 标签验证参数（required, oneof=0 1） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [ ] 2.5 创建 Response DTO

  - File: `main/app/dto/resp/transfer_rule_resp.go`
  - Purpose: 定义 API 响应数据
  - Requirements: Requirement 1, Requirement 3
  - Leverage: 现有 DTO: `main/app/dto/resp/transfer_order.go`
  - Prompt: Role: Go Developer | Task: 创建 Response DTO，包含 TransferRuleResp, TransferRuleListResp, TransferRuleGetResp 结构体 | Context: TransferRuleResp 包含 ShopId, ShopName, AllowTransferIn, AllowTransferOut, UpdateTime，TransferRuleListResp 包含 List 和 Meta 分页信息 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

### Service 层

- [ ] 2.6 创建 Service 接口

  - File: `main/app/service/transfer_rule/i_transfer_rule_srv.go`
  - Purpose: 定义业务逻辑接口
  - Requirements: Requirement 1, Requirement 2, Requirement 3, Requirement 4
  - Leverage: 现有 Service 接口: `main/app/service/transfer_order/i_transfer_order_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 ITransferRuleSrv 接口，定义业务方法 | Context: 包含 Save, GetByShopId, GetList 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 2.7 实现 Service 业务逻辑

  - File: `main/app/service/transfer_rule/transfer_rule.go`
  - Purpose: 实现核心业务逻辑
  - Requirements: Requirement 1, Requirement 2, Requirement 3, Requirement 4
  - Leverage: 现有 Service 实现: `main/app/service/transfer_order/transfer_order.go`，Task 2.1-2.5 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 transferRuleSrv，包含完整业务逻辑 | Context: 持有 DBManager 和 ShopSrv，Save 方法验证至少保留一个类型，GetByShopId 方法未配置时返回默认值(1,1)，GetList 方法查询所有门店并匹配规则 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，使用 errors.WithMessage 包装错误 | Success: Service 实现完整，业务逻辑正确（规则验证、默认值处理）

- [ ] 2.8 编写 Service 单元测试

  - File: `main/app/service/transfer_rule/transfer_rule_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有测试: `main/app/service/transfer_order/transfer_order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 TransferRuleSrv 编写单元测试，覆盖率 ≥ 70% | Context: 测试 Save 方法（正常保存、更新、参数验证），测试 GetByShopId 方法（有配置、无配置），测试 GetList 方法（分页、数据组装） | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 Mock | Success: 测试覆盖率 ≥ 70%，所有测试通过

### API 层

- [ ] 2.9 创建 API Controller

  - File: `main/app/api/v1/admin/admin_transfer_rule.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: Requirement 1, Requirement 2, Requirement 3, Requirement 4
  - Leverage: 现有 API: `main/app/api/v1/admin/admin_transfer_order.go`，Task 2.6-2.7 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 TransferRuleAPI，实现 RESTful 接口（Save, Get, List） | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象，Get 方法如果没有传 shop_id 则使用当前登录门店 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON()，使用 ShouldBindJSON 和 ShouldBindQuery | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 2.10 注册 API 路由

  - File: `main/router/admin_router.go`
  - Purpose: 注册 API 路由
  - Requirements: Requirement 1
  - Leverage: 现有路由: `main/router/admin_router.go`
  - Command: 在 admin_router.go 中添加路由组 `/transfer_rule`，注册 Save (POST), Get (GET), List (GET) 路由
  - Success: 路由注册成功，可以访问 `/api/v1/transfer_rule/save`, `/api/v1/transfer_rule/get`, `/api/v1/transfer_rule/list`

- [ ] 2.11 编写 API 集成测试

  - File: `main/app/api/v1/admin/admin_transfer_rule_test.go`
  - Purpose: 测试 API 接口
  - Requirements: Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有测试: `main/app/api/v1/admin/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 TransferRuleAPI 编写集成测试 | Context: 测试 Save 接口（正常保存、参数验证），测试 Get 接口（有配置、无配置），测试 List 接口（分页、数据格式） | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试响应格式 | Success: 所有 API 测试通过

---

## Phase 3: Vue 前端模块

- [ ] 3.1 创建 API 封装

  - File: `admin/views/admin/api/transfer-rule.ts`
  - Purpose: 封装后端 API 调用
  - Requirements: Requirement 1
  - Leverage: 现有 API: `admin/views/admin/api/transfer-order.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装 TransferRule API 调用 | Context: 使用 axios，定义 TransferRuleSaveReq, TransferRuleGetReq, TransferRuleListReq, TransferRuleResp, TransferRuleListResp 类型，导出 getTransferRuleList, saveTransferRule, getTransferRule 方法 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

- [ ] 3.2 创建调拨规则配置页面

  - File: `admin/views/admin/pages/business-setting/transfer-rule/index.vue`
  - Purpose: 实现调拨规则配置 UI
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有页面: `admin/views/admin/pages/`，参考类似的列表页面
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建调拨规则配置页面，实现门店列表和规则勾选功能 | Context: 使用 Element Plus 的 Table + Checkbox，展示所有门店，每个门店有两个 Checkbox（允许调入、允许调出），实现前端规则验证（最后一个选项置灰），使用 Composition API，使用 el-message 提示保存结果 | Restrictions: 遵循 .cursor/rules/vue.mdc，禁用最后一个 Checkbox 时显示提示信息 | Success: 页面创建成功，功能完整，规则验证正确

- [ ] 3.3 注册路由和菜单

  - File: `admin/views/admin/router/index.ts`, 菜单配置文件
  - Purpose: 注册路由和菜单项
  - Requirements: Requirement 1
  - Leverage: 现有路由和菜单配置
  - Command: 在路由配置中添加 `/business-setting/transfer-rule` 路由，在菜单配置中添加"调拨规则"菜单项
  - Success: 路由注册成功，可以访问页面，菜单显示正确

- [ ] 3.4 调拨单发起页面集成规则查询

  - File: 调拨单发起页面相关文件（待确定具体路径）
  - Purpose: 根据规则过滤调拨类型选项
  - Requirements: Requirement 3
  - Leverage: Task 3.1 的 API 封装
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在调拨单发起页面中集成规则查询 | Context: 页面加载时调用 getTransferRule API，根据返回的 allow_transfer_in 和 allow_transfer_out 过滤调拨类型选项（隐藏不允许的类型） | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 门店端根据规则显示可选的调拨类型

---

## Phase 4: 测试和优化

- [ ] 4.1 集成测试

  - File: -
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整流程：配置规则 → 保存 → 查询 → 门店端生效，测试边界场景：未配置规则、最后一个类型置灰 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Command: 使用 wrk 测试 API 性能
  - Success: 本地响应时间 < 100ms

- [ ] 4.3 数据库查询优化

  - File: `main/app/repository/transfer_rule.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Command: 使用 EXPLAIN 分析 GetList 查询，确保使用索引
  - Success: 查询时间 < 30ms，使用索引

- [ ] 4.4 文档更新

  - File: `docs/shared/api/admin_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: 更新 API 文档（新增 3 个接口），更新 CHANGELOG.md（v2.13.0 新增调拨规则功能） | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - ✅ 总店可配置门店调拨规则
  - ✅ 至少保留一个调拨类型（前端 + 后端验证）
  - ✅ 门店端根据规则过滤调拨类型
  - ✅ 未配置规则时显示默认值（所有类型）

### 文档同步

- [ ] API 文档已更新（3 个新接口）
- [ ] 数据库文档已更新（ttpos_transfer_rule 表）
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
grep -c "^- \[" docs/shared/specs/active/story-admin-transfer-rule/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-transfer-rule/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-transfer-rule/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-transfer-rule/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-transfer-rule/tasks.md)" | bc
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

### Vue 前端开发

```
Role: Frontend Developer with Vue 3 + TypeScript expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/vue.mdc

Restrictions:
- 使用 Vue 3 Composition API
- 使用 TypeScript
- 使用 Element Plus 组件库
- 遵循命名规范

Success Criteria:
- {成功标准1}
- 代码通过 ESLint 检查
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository)

Test Cases Required:
- 正常场景测试
- 异常场景测试（参数验证、数据库错误）
- 边界条件测试（未配置规则、两个都为 false）

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2026-01/2026-01-06.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2026-01-06  
**维护者**: weifashi

