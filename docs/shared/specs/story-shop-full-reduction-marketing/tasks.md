# 新管理端-满减营销功能 任务分解

> 本文档定义 新管理端-满减营销功能 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 0  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建活动表迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_full_reduction_activity_table.php`
  - Purpose: 定义活动表结构
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_full_reduction_activity 表的迁移文件，遵循 requirements.md 和 design.md 中的数据库设计 | Context: 必须包含 id, uuid, name, multi_language_name_uuid, start_date, end_date, start_time, end_time, is_all_day, reduction_type, is_disabled, create_time, update_time, delete_time 字段，时间字段使用 int 类型，金额字段使用 decimal(22,4)，注意：不需要 company_uuid 字段（每个商户独立数据库实例） | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查表是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 创建规则表迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_full_reduction_activity_rule_table.php`
  - Purpose: 定义规则表结构
  - Requirements: 2.5, 2.6
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_full_reduction_activity_rule 表的迁移文件 | Context: 必须包含 id, uuid, full_reduction_activity_uuid, threshold, reduction_amount, create_time, update_time, delete_time 字段，threshold 和 reduction_amount 使用 decimal(22,4) | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功

- [ ] 1.3 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建表
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.1, 1.2 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，表已创建

- [ ] 1.4 创建 Go Model - FullReductionActivity

  - File: `main/app/model/full_reduction_activity.go`
  - Purpose: 定义活动数据模型，与数据库表对应
  - Requirements: 1.1
  - Leverage: 现有 Model: `main/app/model/base.go`, `main/app/model/multi_language_name.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 创建 FullReductionActivity 结构体，映射到 ttpos_full_reduction_activity 表 | Context: 使用 gorm 标签，包含所有字段，实现 TableName() 方法，添加 GetStatus() 方法用于判断活动状态 | Restrictions: 遵循 .cursor/rules/go-main.mdc，继承 BaseModel | Success: Model 创建成功，字段映射正确，GetStatus() 方法正确

- [ ] 1.5 创建 Go Model - FullReductionActivityRule

  - File: `main/app/model/full_reduction_activity_rule.go`
  - Purpose: 定义规则数据模型
  - Requirements: 1.2
  - Leverage: 现有 Model: `main/app/model/base.go`，迁移文件: Task 1.2
  - Prompt: Role: Go Developer | Task: 创建 FullReductionActivityRule 结构体，映射到 ttpos_full_reduction_activity_rule 表 | Context: 使用 gorm 标签，包含所有字段，实现 TableName() 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 创建成功

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [ ] 2.1 创建 FullReductionActivity Repository 接口

  - File: `main/app/repository/i_full_reduction_activity_repo.go`
  - Purpose: 定义活动数据访问接口
  - Requirements: 1.1, 1.3, 1.4, 1.5
  - Leverage: 现有 Repository 接口: `main/app/repository/i_*_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 IFullReductionActivityRepo 接口，定义 CRUD 方法和选项方法 | Context: 使用选项模式(DBOption)，包含 Create, Update, GetByUuid, GetList, Delete 方法，包含 WhereUuid, WhereStatus 选项方法，注意：不需要 WhereCompanyUuid（每个商户独立数据库实例） | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 接口定义完整，方法签名正确

- [ ] 2.2 实现 FullReductionActivity Repository（选项模式）

  - File: `main/app/repository/full_reduction_activity_repo.go`
  - Purpose: 实现活动数据访问逻辑
  - Requirements: 2.1
  - Leverage: 现有 Repository 实现: `main/app/repository/*_repo.go`，使用选项模式
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 FullReductionActivityRepoImpl，使用选项模式实现灵活查询 | Context: 只持有 db *gorm.DB，实现所有接口方法和选项方法，GetByUuid 需要 Preload Rules 和 MultiLanguageName | Restrictions: 不能持有 DBManager，使用 GORM，软删除(delete_time=0) | Success: Repository 实现完整，选项模式正确，软删除正确

- [ ] 2.3 创建 FullReductionActivityRule Repository 接口

  - File: `main/app/repository/i_full_reduction_activity_rule_repo.go`
  - Purpose: 定义规则数据访问接口
  - Requirements: 2.5, 2.6
  - Leverage: 现有 Repository 接口: `main/app/repository/i_*_repo.go`
  - Prompt: Role: Go Developer | Task: 创建 IFullReductionActivityRuleRepo 接口 | Context: 包含 Create, Update, GetByFullReductionActivityUuid, DeleteByFullReductionActivityUuid, Delete 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整

- [ ] 2.4 实现 FullReductionActivityRule Repository

  - File: `main/app/repository/full_reduction_activity_rule_repo.go`
  - Purpose: 实现规则数据访问逻辑
  - Requirements: 2.3
  - Leverage: 现有 Repository 实现: `main/app/repository/*_repo.go`
  - Prompt: Role: Go Developer | Task: 实现 FullReductionActivityRuleRepoImpl | Context: 只持有 db *gorm.DB，实现所有接口方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Repository 实现完整

- [ ] 2.5 编写 Repository 单元测试

  - File: `main/app/repository/full_reduction_activity_repo_test.go`, `main/app/repository/full_reduction_activity_rule_repo_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 2.2, 2.4
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 FullReductionActivityRepo 和 FullReductionActivityRuleRepo 编写单元测试，覆盖率 ≥ 80% | Context: 测试 CRUD 方法，测试选项方法，测试软删除 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

### DTO 层

- [ ] 2.6 创建 Request DTO

  - File: `main/app/dto/req/full_reduction_activity_req.go`
  - Purpose: 定义 API 请求参数
  - Requirements: 2.1, 2.2, 2.3, 2.4, 4.1, 4.3
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 Request DTO，包含 Create, Update, Get, List, Delete, Disable 请求结构体 | Context: 使用 binding 标签验证参数，Create 和 Update 需要验证活动名称(1-50字符)、活动日期(不可选择以前日期)、规则(至少一个，满减金额范围0.01-999999.99) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [ ] 2.7 创建 Response DTO

  - File: `main/app/dto/resp/full_reduction_activity_resp.go`
  - Purpose: 定义 API 响应数据
  - Requirements: 1.3, 5.1
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 Response DTO，包含单条和列表响应结构体 | Context: 包含 Meta 分页信息，包含活动状态、规则列表等完整信息 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

### Service 层

- [ ] 2.8 创建 Service 接口

  - File: `main/app/service/i_full_reduction_activity_srv.go`
  - Purpose: 定义业务逻辑接口
  - Requirements: 2.1, 2.2, 2.3, 2.4, 4.1, 4.3
  - Leverage: 现有 Service 接口: `main/app/service/i_*_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 IFullReductionActivitySrv 接口，定义业务方法 | Context: 包含 Create, Update, GetByUuid, GetList, Delete, Disable 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 2.9 实现 Service 业务逻辑 - Create

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 实现创建活动业务逻辑
  - Requirements: 2.1, 2.2, 2.3, 2.7
  - Leverage: 现有 Service 实现: `main/app/service/*_srv.go`，Task 2.1-2.7 的实现，MultiLanguageNameService
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 Create 方法，包含完整业务逻辑 | Context: 创建多语言名称，创建活动，创建规则（阶梯满减需要排序），验证活动日期不可选择以前日期，验证规则数量至少一个，验证满减金额范围 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，使用事务 | Success: Service 实现完整，业务逻辑正确，事务管理正确

- [ ] 2.10 实现 Service 业务逻辑 - Update

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 实现更新活动业务逻辑
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: Task 2.9 的实现
  - Prompt: Role: Go Developer | Task: 实现 Update 方法 | Context: 验证活动状态（已结束和已失效的活动不可编辑），更新多语言名称，更新活动，更新规则（删除旧规则，创建新规则，阶梯满减需要排序） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Update 实现完整，状态验证正确

- [ ] 2.11 实现 Service 业务逻辑 - GetList

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 实现获取活动列表业务逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: Task 2.2 的实现
  - Prompt: Role: Go Developer | Task: 实现 GetList 方法 | Context: 根据 status 参数过滤活动（全部/进行中/未开始/已结束），活动状态根据当前时间和商户时区判断，按添加顺序排序（最新在前） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: GetList 实现完整，状态判断正确

- [ ] 2.12 实现 Service 业务逻辑 - GetByUuid, Delete, Disable

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 实现其他业务逻辑
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 5.1
  - Leverage: Task 2.2 的实现
  - Prompt: Role: Go Developer | Task: 实现 GetByUuid, Delete, Disable 方法 | Context: GetByUuid 返回完整活动信息，Delete 验证活动状态（仅允许删除未开始、已结束、已失效的活动），Disable 验证活动状态（仅允许失效进行中或未开始的活动） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有方法实现完整

- [ ] 2.13 实现活动规则计算逻辑

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 实现阶梯满减和循环满减的计算逻辑
  - Requirements: 2.5, 2.6
  - Leverage: Task 2.2, 2.4 的实现
  - Prompt: Role: Go Developer | Task: 实现活动规则计算逻辑 | Context: 阶梯满减：订单金额满足多个档次时，按最高档次满减；循环满减：订单金额每满足一个基数，减去相应的减价金额 | Restrictions: 遵循 .cursor/rules/go-main.mdc，需要处理边界情况 | Success: 计算逻辑正确，边界情况处理正确

- [ ] 2.14 编写 Service 单元测试

  - File: `main/app/service/full_reduction_activity_srv_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 2.9, 2.10, 2.11, 2.12, 2.13
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 FullReductionActivitySrv 编写单元测试，覆盖率 ≥ 70%，活动规则计算逻辑覆盖率 100% | Context: 测试业务逻辑，测试错误处理，测试事务管理，测试活动规则计算（各种场景） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%（规则计算 100%），所有测试通过

### API 层

- [ ] 2.15 创建 API Controller

  - File: `main/app/api/v1/shop/shop_full_reduction_activity.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_setting.go`，Task 2.8-2.13 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 FullReductionActivityHandler，实现 RESTful 接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象，实现 Create, Update, GetByUuid, GetList, Delete, Disable 接口 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 2.16 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 API 路由
  - Requirements: 2.15
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功

- [ ] 2.17 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_full_reduction_activity_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 2.15
  - Leverage: 现有测试: `main/app/api/v1/shop/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 FullReductionActivityHandler 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: Flutter 前端模块（新管理端）

> **注意**：新管理端位于 `ttpos-flutter/apps/shop`，使用 Flutter + Dart 开发，不是 Vue。旧管理后台和店铺后台在 `admin/views/` 中使用 Vue3。

- [ ] 3.1 创建 API 封装

  - File: `ttpos-flutter/packages/api/lib/shop/full_reduction_activity_api.dart`
  - Purpose: 封装后端 API 调用
  - Requirements: 所有功能需求
  - Leverage: 现有 API: `ttpos-flutter/packages/api/lib/shop/`
  - Prompt: Role: Flutter Developer with Dart expertise | Task: 封装 FullReductionActivity API 调用 | Context: 使用 Dio，定义 Dart 类型，遵循项目 API 封装规范 | Restrictions: 遵循 Flutter 开发规范 | Success: API 封装完成，类型定义正确

- [ ] 3.2 创建活动列表页面

  - File: `ttpos-flutter/apps/shop/lib/pages/marketing/full_reduction_activity/list_page.dart`
  - Purpose: 实现活动列表页面 UI
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有页面: `ttpos-flutter/apps/shop/lib/pages/`
  - Prompt: Role: Flutter Developer | Task: 创建活动列表页面，实现列表展示、状态筛选功能 | Context: 使用 GetX 状态管理，按添加顺序排序（最新在前），显示活动名称、状态、活动日期、活动时间 | Restrictions: 遵循 Flutter 开发规范 | Success: 页面创建成功，功能完整

- [ ] 3.3 创建活动添加/编辑页面

  - File: `ttpos-flutter/apps/shop/lib/pages/marketing/full_reduction_activity/form_page.dart`
  - Purpose: 实现活动添加和编辑页面 UI
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有页面: `ttpos-flutter/apps/shop/lib/pages/`
  - Prompt: Role: Flutter Developer | Task: 创建活动添加/编辑页面，实现表单验证、规则管理、未保存退出提示、进行中活动编辑确认 | Context: 使用 GetX，验证活动名称(1-50字符)、活动日期(不可选择以前日期)、规则(至少一个，满减金额范围0.01-999999.99)，阶梯满减自动排序，未保存退出提示，进行中活动编辑确认 | Restrictions: 遵循 Flutter 开发规范 | Success: 页面创建成功，功能完整，验证正确

- [ ] 3.4 创建活动详情页面

  - File: `ttpos-flutter/apps/shop/lib/pages/marketing/full_reduction_activity/detail_page.dart`
  - Purpose: 实现活动详情页面 UI
  - Requirements: 5.1, 5.2, 5.3
  - Leverage: 现有页面: `ttpos-flutter/apps/shop/lib/pages/`
  - Prompt: Role: Flutter Developer | Task: 创建活动详情页面 | Context: 显示所有活动信息，阶梯满减显示所有规则（已排序），循环满减显示基数和减价金额 | Restrictions: 遵循 Flutter 开发规范 | Success: 页面创建成功

- [ ] 3.5 实现活动失效和删除功能

  - File: `ttpos-flutter/apps/shop/lib/pages/marketing/full_reduction_activity/list_page.dart`
  - Purpose: 实现活动失效和删除功能
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5
  - Leverage: Task 3.2
  - Prompt: Role: Flutter Developer | Task: 实现活动失效和删除功能 | Context: 失效功能仅允许失效进行中或未开始的活动，删除功能仅允许删除未开始、已结束、已失效的活动，进行中的活动不可删除，所有操作需要确认提示 | Restrictions: 遵循 Flutter 开发规范 | Success: 功能实现完整，提示正确

- [ ] 3.6 在工作台添加营销活动入口

  - File: `ttpos-flutter/apps/shop/lib/pages/dashboard/dashboard_page.dart` (或相关文件)
  - Purpose: 在工作台添加营销活动入口
  - Requirements: 1.1
  - Leverage: 现有工作台页面
  - Prompt: Role: Flutter Developer | Task: 在工作台添加【营销活动】入口 | Context: 点击后跳转到活动列表页面 | Restrictions: 遵循 Flutter 开发规范 | Success: 入口添加成功

---

## Phase 4: 测试和优化

- [ ] 4.1 集成测试

  - File: `test/integration/full_reduction_activity_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程，测试数据一致性 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 4.3 缓存优化

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 实现 Redis 缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现
  - Success: 缓存实现完成，命中率 > 80%

- [ ] 4.4 数据库查询优化

  - File: `main/app/repository/full_reduction_activity_repo.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 50ms

- [ ] 4.5 并发控制

  - File: `main/app/service/full_reduction_activity_srv.go`
  - Purpose: 添加 UUID 锁
  - Requirements: 可靠性要求
  - Leverage: `pkg/lock/system_lock.go`
  - Success: 并发场景测试通过

- [ ] 4.6 文档更新

  - File: `docs/shared/api/shop_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, 数据库文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

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
  - 活动规则计算: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新表）
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
grep -c "^- \[" docs/shared/specs/story-shop-full-reduction-marketing/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-shop-full-reduction-marketing/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-shop-full-reduction-marketing/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/story-shop-full-reduction-marketing/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/story-shop-full-reduction-marketing/tasks.md)" | bc
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

### Flutter 前端开发（新管理端）

```
Role: Flutter Developer with Dart expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 新管理端位于 ttpos-flutter/apps/shop，使用 Flutter + Dart

Restrictions:
- 使用 GetX 状态管理
- 使用 Dart 语言
- 遵循 Flutter 项目结构和开发规范
- 注意：旧管理后台和店铺后台使用 Vue3（位于 admin/views/），但新管理端使用 Flutter

Success Criteria:
- {成功标准1}
- 代码通过 Flutter 分析工具检查
```

### 测试工程师

```
Role: QA Engineer with {Go/Vue} testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试（如适用）

Restrictions:
- 遵循 .cursor/rules/go-main.mdc 或 .cursor/rules/vue.mdc
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
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-21  
**维护者**: 后端开发组

