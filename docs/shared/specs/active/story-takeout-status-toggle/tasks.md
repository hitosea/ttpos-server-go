# 外卖状态切换 任务分解

> 本文档定义 外卖状态切换 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12
**已完成**: 0
**进行中**: -
**完成率**: 0%

---

## Phase 1: 数据库和模型

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [ ] 1.1 创建 ttpos_takeout 表迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_takeout_table.php`
  - Purpose: 定义外卖平台管理表的数据库结构
  - Requirements: 数据库设计要求
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_takeout 表的迁移文件，包含 id, uuid, platform, enabled, menu, create_time, update_time, delete_time 字段 | Context: platform 字段支持多个外卖平台过滤，menu 字段存储 JSON 格式菜单数据 | Restrictions: 遵循 .cursor/rules/database.mdc，包含复合唯一索引 uk_platform | Success: 迁移文件创建成功，字段定义正确，索引完整

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建外卖平台管理表
  - Requirements: 数据库设计要求
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，ttpos_takeout 表已创建

- [ ] 1.3 创建 Takeout Go Model

  - File: `main/app/modules/takeout/domain/model/takeout.go`
  - Purpose: 定义外卖平台管理数据模型
  - Requirements: 数据模型要求
  - Leverage: 现有 Model: `main/app/modules/takeout/domain/model/`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 创建 Takeout 结构体，对应 ttpos_takeout 表 | Context: 包含所有字段，menu 字段使用 interface{} 类型处理 JSON，实现 TableName() 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用正确的 gorm 标签 | Success: Model 创建成功，字段映射正确，JSON 处理正确

---

## Phase 2: 核心实现（Go Main）

### Takeout Domain Service

- [ ] 2.1 创建 Takeout Domain Service 接口

  - File: `main/app/modules/takeout/domain/service/i_takeout_domain_service.go`
  - Purpose: 定义外卖领域业务逻辑接口
  - Requirements: 功能需求 1.1
  - Leverage: 现有 Domain Service 接口: `main/app/modules/takeout/domain/service/`
  - Prompt: Role: Go Developer specializing in Domain Service | Task: 创建 ITakeoutDomainService 接口，定义外卖状态管理的领域方法 | Context: 包含 GetByPlatform, UpdateStatus, UpdateMenu 等方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 2.2 实现 Takeout Domain Service

  - File: `main/app/modules/takeout/domain/service/takeout_domain_service.go`
  - Purpose: 实现外卖领域业务逻辑
  - Requirements: 功能需求 1.1
  - Leverage: Task 2.1 的接口，Takeout Repository
  - Prompt: Role: Go Developer with domain logic expertise | Task: 实现 takeoutDomainService，包含状态管理和菜单数据处理 | Context: 使用 Repository 进行数据访问，处理业务规则验证 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不直接依赖 Repository | Success: 领域逻辑实现正确，业务规则完整

### Takeout Repository 层

- [ ] 2.3 创建 Takeout Repository 接口

  - File: `main/app/modules/takeout/domain/repository/i_takeout_repository.go`
  - Purpose: 定义外卖数据访问接口
  - Requirements: 数据访问要求
  - Leverage: 现有 Repository 接口: `main/app/modules/takeout/domain/repository/`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 ITakeoutRepository 接口，定义 CRUD 方法和平台查询方法 | Context: 使用选项模式(DBOption)，包含 GetByPlatform, UpdateByPlatform 等方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 接口定义完整，方法签名正确

- [ ] 2.4 实现 Takeout Repository

  - File: `main/app/modules/takeout/domain/repository/takeout_repository.go`
  - Purpose: 实现外卖数据访问逻辑
  - Requirements: 数据访问要求
  - Leverage: 现有 Repository 实现: `main/app/modules/takeout/domain/repository/`，使用选项模式
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 TakeoutRepositoryImpl，使用选项模式实现灵活查询 | Context: 只持有 db \*gorm.DB，实现所有接口方法和选项方法 | Restrictions: 不能持有 DBManager，使用 GORM，软删除(delete_time=0) | Success: Repository 实现完整，选项模式正确，软删除正确

- [ ] 2.5 编写 Takeout Repository 单元测试

  - File: `main/app/modules/takeout/domain/repository/takeout_repository_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 TakeoutRepository 编写单元测试，覆盖率 ≥ 80% | Context: 测试 CRUD 方法，测试平台查询，测试软删除 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

### Takeout App Service 层

- [ ] 2.6 创建 Takeout Status App Service 接口

  - File: `main/app/modules/takeout/application/i_takeout_status_app_service.go`
  - Purpose: 定义外卖状态应用服务接口
  - Requirements: 功能需求 1.1
  - Leverage: 现有 App Service 接口: `main/app/modules/takeout/application/`
  - Prompt: Role: Go Developer specializing in Application Service | Task: 创建 ITakeoutStatusAppService 接口，定义外卖状态管理的方法 | Context: 包含 GetTakeoutStatus, GetAllTakeoutStatus, ToggleTakeoutStatus, UpdateTakeoutMenu 等方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 2.7 实现 Takeout Status App Service

  - File: `main/app/modules/takeout/application/takeout_status_app_service.go`
  - Purpose: 实现外卖状态应用服务逻辑
  - Requirements: 功能需求 1.1
  - Leverage: Task 2.6 的接口，Task 2.2 的 Domain Service
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 takeoutStatusAppService，协调领域服务和外部接口 | Context: 持有 Domain Service 实例，处理缓存，发布事件 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只能依赖其他 Service 接口 | Success: 应用服务实现正确，缓存和事件处理完整

- [ ] 2.8 编写 Takeout Status App Service 单元测试

  - File: `main/app/modules/takeout/application/takeout_status_app_service_test.go`
  - Purpose: 确保 App Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 TakeoutStatusAppService 编写单元测试，覆盖率 ≥ 70% | Context: 测试业务逻辑，测试缓存，测试事件发布 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### Takeout API 层

- [ ] 2.9 创建 Takeout Status API 接口

  - File: `main/app/api/v1/shop/takeout_status_api.go`
  - Purpose: 实现外卖状态相关的 HTTP API 接口
  - Requirements: API 设计要求
  - Leverage: 现有 API: `main/app/api/v1/shop/`，Task 2.7 的 App Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 TakeoutStatusAPI，实现 RESTful 接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 2.10 注册 Takeout Status API 路由

  - File: `main/router/router.go`
  - Purpose: 注册外卖状态相关的 API 路由
  - Requirements: API 设计要求
  - Leverage: 现有路由配置: `main/router/router.go`
  - Success: 路由注册成功

- [ ] 2.11 编写 Takeout Status API 集成测试

  - File: `main/app/api/v1/shop/takeout_status_api_test.go`
  - Purpose: 测试外卖状态 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/shop/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 TakeoutStatusAPI 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 2.2 实现外卖状态获取逻辑

  - File: `main/app/modules/setting/application/setting_app_service.go`
  - Purpose: 实现获取外卖状态的业务逻辑
  - Requirements: 功能需求 1.1
  - Leverage: 现有 Setting App Service 实现，Task 2.1 的接口
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetTakeoutStatus 方法，从数据库获取外卖状态设置 | Context: 使用缓存优先策略，支持默认值处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic | Success: 方法实现完成，缓存和数据库逻辑正确

- [ ] 2.3 实现外卖状态切换逻辑

  - File: `main/app/modules/setting/application/setting_app_service.go`
  - Purpose: 实现切换外卖状态的业务逻辑
  - Requirements: 功能需求 1.2
  - Leverage: Task 2.2 的实现，现有 Setting 模块的更新逻辑
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 ToggleTakeoutStatus 方法，支持开启/关闭外卖功能 | Context: 包含状态验证、缓存更新、事件发布 | Restrictions: 使用事务管理，保证数据一致性 | Success: 状态切换逻辑正确，包含完整的业务校验

### DTO 层

- [ ] 2.4 创建外卖状态 Request DTO

  - File: `main/app/modules/setting/types/request/takeout_status.go`
  - Purpose: 定义外卖状态相关的请求参数
  - Requirements: API 设计要求
  - Leverage: 现有 DTO: `main/app/modules/setting/types/request/`
  - Prompt: Role: Go Developer | Task: 创建外卖状态切换的请求 DTO | Context: 包含 enabled 字段，使用 binding 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，字段验证正确

- [ ] 2.5 创建外卖状态 Response DTO

  - File: `main/app/modules/setting/types/response/takeout_status.go`
  - Purpose: 定义外卖状态相关的响应数据
  - Requirements: API 设计要求
  - Leverage: 现有 DTO: `main/app/modules/setting/types/response/`
  - Prompt: Role: Go Developer | Task: 创建外卖状态的响应 DTO | Context: 包含 enabled 和 updated_at 字段 | Restrictions: data 必须是对象 | Success: DTO 创建成功，响应格式正确

### API 层

- [ ] 2.6 创建外卖状态 API 接口

  - File: `main/app/api/v1/shop/setting_api.go`
  - Purpose: 实现外卖状态相关的 HTTP API 接口
  - Requirements: API 设计要求
  - Leverage: 现有 Setting API: `main/app/api/v1/shop/setting_api.go`，Task 2.1-2.3 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetTakeoutStatus 和 ToggleTakeoutStatus API 方法 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 2.7 注册外卖状态路由

  - File: `main/router/router.go`
  - Purpose: 注册外卖状态相关的 API 路由
  - Requirements: API 设计要求
  - Leverage: 现有路由配置: `main/router/router.go`
  - Success: 路由注册成功

### 测试

- [ ] 2.8 编写 Setting App Service 单元测试

  - File: `main/app/modules/setting/application/setting_app_service_test.go`
  - Purpose: 确保外卖状态相关业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/modules/setting/application/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为外卖状态相关方法编写单元测试，覆盖率 ≥ 70% | Context: 测试正常场景、异常场景、缓存逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 2.9 编写外卖状态 API 集成测试

  - File: `main/app/api/v1/shop/setting_api_test.go`
  - Purpose: 测试外卖状态 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/shop/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为外卖状态 API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: 测试和优化

### 集成测试

- [ ] 3.1 端到端集成测试

  - File: `test/integration/takeout_status_test.go`
  - Purpose: 测试完整的外卖状态管理流程
  - Requirements: 验收标准
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现外卖状态管理的端到端集成测试 | Context: 测试多平台状态管理，后端 API 完整流程 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

### 性能优化

- [ ] 3.2 缓存性能优化

  - File: `main/app/modules/takeout/application/takeout_status_app_service.go`
  - Purpose: 确保缓存性能达标
  - Requirements: 性能要求
  - Leverage: 现有缓存实现
  - Success: 缓存命中率 > 80%，响应时间 < 200ms

### 文档更新

- [ ] 3.3 更新 API 文档

  - File: `docs/shared/api/takeout_api.md`
  - Purpose: 记录新增的外卖状态 API 接口
  - Requirements: 文档验收
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 创建 Takeout API 文档，记录所有外卖状态管理接口 | Context: 记录请求格式、响应格式、错误码，支持多平台管理 | Restrictions: 文档准确完整 | Success: API 文档已创建并更新

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
grep -c "^- \[" docs/shared/specs/active/story-takeout-status-toggle/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-status-toggle/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-status-toggle/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-status-toggle/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-status-toggle/tasks.md)" | bc
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

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试

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
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0
**最后更新**: 2025-12-13
**维护者**: weifashi
