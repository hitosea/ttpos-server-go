# {功能名称} 任务分解

> 本文档定义 {功能} 的详细执行任务清单。

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

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [ ] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_{table_name}_table.php`
  - Purpose: 定义数据库表结构
  - Requirements: {需求编号}
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos\_{table_name} 表的迁移文件，遵循 requirements.md 中的数据库设计 | Context: 必须包含 id, uuid, create_time, update_time, delete_time 字段，时间字段使用 int 类型，金额字段使用 decimal(20,8) | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查表是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建表
  - Requirements: {需求编号}
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，表已创建

- [ ] 1.3 创建 Go Model

  - File: `main/app/model/{name}.go`
  - Purpose: 定义 Go 数据模型，与数据库表对应
  - Requirements: {需求编号}
  - Leverage: 现有 Model: `main/app/model/`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 创建 {ModelName} 结构体，映射到 ttpos\_{table_name} 表 | Context: 使用 gorm 标签，包含所有字段，实现 TableName() 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 创建成功，字段映射正确

- [ ] 1.4 更新 Seeds 文件（可选）

  - File: `admin/database/seeds/shop_01.sql`
  - Purpose: 添加测试数据
  - Requirements: {需求编号}
  - Leverage: 现有 Seeds: `admin/database/seeds/`
  - Success: Seeds 文件更新成功

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [ ] 2.1 创建 Repository 接口

  - File: `main/app/repository/i_{name}_repo.go`
  - Purpose: 定义数据访问接口
  - Requirements: {需求编号}
  - Leverage: 现有 Repository 接口: `main/app/repository/i_*_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 I{Name}Repo 接口，定义 CRUD 方法和选项方法 | Context: 使用选项模式(DBOption)，包含 Create, Update, GetByUuid, GetList, Delete 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 接口定义完整，方法签名正确

- [ ] 2.2 实现 Repository（选项模式）

  - File: `main/app/repository/{name}_repo.go`
  - Purpose: 实现数据访问逻辑
  - Requirements: {需求编号}
  - Leverage: 现有 Repository 实现: `main/app/repository/*_repo.go`，使用选项模式
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 {Name}RepoImpl，使用选项模式实现灵活查询 | Context: 只持有 db \*gorm.DB，实现所有接口方法和选项方法 | Restrictions: 不能持有 DBManager，使用 GORM，软删除(delete_time=0) | Success: Repository 实现完整，选项模式正确，软删除正确

- [ ] 2.3 编写 Repository 单元测试

  - File: `main/app/repository/{name}_repo_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: {需求编号}
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 {Name}Repo 编写单元测试，覆盖率 ≥ 80% | Context: 测试 CRUD 方法，测试选项方法，测试软删除 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

### DTO 层

- [ ] 2.4 创建 Request DTO

  - File: `main/app/dto/req/{name}_req.go`
  - Purpose: 定义 API 请求参数
  - Requirements: {需求编号}
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 Request DTO，包含 Create, Update, Get, List 请求结构体 | Context: 使用 binding 标签验证参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [ ] 2.5 创建 Response DTO

  - File: `main/app/dto/resp/{name}_resp.go`
  - Purpose: 定义 API 响应数据
  - Requirements: {需求编号}
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 Response DTO，包含单条和列表响应结构体 | Context: 包含 Meta 分页信息 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

### Service 层

- [ ] 2.6 创建 Service 接口

  - File: `main/app/service/i_{name}_srv.go`
  - Purpose: 定义业务逻辑接口
  - Requirements: {需求编号}
  - Leverage: 现有 Service 接口: `main/app/service/i_*_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 I{Name}Srv 接口，定义业务方法 | Context: 包含 Create, Update, GetByUuid, GetList, Delete 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 2.7 实现 Service 业务逻辑

  - File: `main/app/service/{name}_srv.go`
  - Purpose: 实现核心业务逻辑
  - Requirements: {需求编号}
  - Leverage: 现有 Service 实现: `main/app/service/*_srv.go`，Task 2.1-2.5 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 {name}Srv，包含完整业务逻辑 | Context: 持有 DBManager，依赖其他 Service 接口（不依赖 Repository），使用事务管理，发布事件（如需要） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确，事务管理正确

- [ ] 2.8 编写 Service 单元测试

  - File: `main/app/service/{name}_srv_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: {需求编号}
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 {Name}Srv 编写单元测试，覆盖率 ≥ 70% | Context: 测试业务逻辑，测试错误处理，测试事务管理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%（Payment/Order 相关 100%），所有测试通过

### API 层

- [ ] 2.9 创建 API Controller

  - File: `main/app/api/{name}_api.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: {需求编号}
  - Leverage: 现有 API: `main/app/api/*_api.go`，Task 2.6-2.7 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 {Name}API，实现 RESTful 接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 2.10 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 API 路由
  - Requirements: {需求编号}
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功

- [ ] 2.11 编写 API 集成测试

  - File: `main/app/api/{name}_api_test.go`
  - Purpose: 测试 API 接口
  - Requirements: {需求编号}
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 {Name}API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: PHP Admin 模块（如适用）

- [ ] 3.1 创建 PHP Controller

  - File: `admin/app/{admin|shop}/controller/{name}Controller.php`
  - Purpose: 实现后台管理接口
  - Requirements: {需求编号}
  - Leverage: 现有 Controller: `admin/app/{admin|shop}/controller/`
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 创建 {Name}Controller，实现后台管理功能 | Context: 遵循 MVC 分层，Controller 不写业务逻辑 | Restrictions: 遵循 .cursor/rules/php.mdc，使用验证器 | Success: Controller 创建成功，接口正确

- [ ] 3.2 创建 PHP Service

  - File: `admin/app/{admin|shop}/service/{Name}Service.php`
  - Purpose: 实现业务逻辑
  - Requirements: {需求编号}
  - Leverage: 现有 Service: `admin/app/{admin|shop}/service/`
  - Success: Service 创建成功

- [ ] 3.3 创建 PHP Model

  - File: `admin/app/{admin|shop}/model/{Name}.php`
  - Purpose: 实现数据模型
  - Requirements: {需求编号}
  - Leverage: 现有 Model: `admin/app/{admin|shop}/model/`
  - Success: Model 创建成功

- [ ] 3.4 创建验证器

  - File: `admin/app/{admin|shop}/validate/{Name}Validate.php`
  - Purpose: 参数验证
  - Requirements: {需求编号}
  - Leverage: 现有 Validate: `admin/app/{admin|shop}/validate/`
  - Success: 验证器创建成功

---

## Phase 4: Vue 前端模块（如适用）

- [ ] 4.1 创建 API 封装

  - File: `admin/views/{admin|shop}/api/{name}.ts`
  - Purpose: 封装后端 API 调用
  - Requirements: {需求编号}
  - Leverage: 现有 API: `admin/views/{admin|shop}/api/`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装 {Name} API 调用 | Context: 使用 axios，定义 TypeScript 类型 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

- [ ] 4.2 创建页面组件

  - File: `admin/views/{admin|shop}/pages/{name}/index.vue`
  - Purpose: 实现页面 UI
  - Requirements: {需求编号}
  - Leverage: 现有页面: `admin/views/{admin|shop}/pages/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建 {Name} 页面，实现列表、新增、编辑、删除功能 | Context: 使用 Element Plus，使用 Composition API | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 页面创建成功，功能完整

- [ ] 4.3 创建业务组件（如需要）

  - File: `admin/views/{admin|shop}/components/{name}/*.vue`
  - Purpose: 实现可复用组件
  - Requirements: {需求编号}
  - Leverage: 现有组件: `admin/views/{admin|shop}/components/`
  - Success: 组件创建成功

---

## Phase 5: 微服务集成（如适用）

- [ ] 5.1 定义 Protobuf

  - File: `ttpos-bmp/app/ttpos-{service}/manifest/protobuf/{name}.proto`
  - Purpose: 定义 gRPC 接口
  - Requirements: {需求编号}
  - Leverage: 现有 Protobuf: `ttpos-bmp/app/ttpos-{service}/manifest/protobuf/`
  - Prompt: Role: gRPC Developer | Task: 定义 {Name} gRPC 服务 | Context: 使用 proto3 语法 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc | Success: Protobuf 定义完成

- [ ] 5.2 生成 gRPC 代码

  - File: -
  - Purpose: 生成 gRPC Go 代码
  - Requirements: {需求编号}
  - Leverage: Task 5.1 的 Protobuf
  - Command: `cd ttpos-bmp/app/ttpos-{service} && make dao`
  - Success: 代码生成成功

- [ ] 5.3 实现 gRPC Controller

  - File: `ttpos-bmp/app/ttpos-{service}/internal/controller/rpc/{name}_controller.go`
  - Purpose: 实现 gRPC 接口
  - Requirements: {需求编号}
  - Leverage: 现有 RPC Controller: `ttpos-bmp/app/ttpos-{service}/internal/controller/rpc/`
  - Success: gRPC Controller 实现完成

- [ ] 5.4 注册服务到 Nacos

  - File: `ttpos-bmp/app/ttpos-{service}/manifest/config/config.yaml`
  - Purpose: 配置服务注册
  - Requirements: {需求编号}
  - Leverage: 现有配置
  - Success: 服务注册成功

---

## Phase 6: 测试和优化

- [ ] 6.1 集成测试

  - File: `test/integration/{name}_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程，测试数据一致性 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 6.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 6.3 缓存优化

  - File: {需要优化的文件}
  - Purpose: 实现 Redis 缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现
  - Success: 缓存实现完成，命中率 > 80%

- [ ] 6.4 数据库查询优化

  - File: `main/app/repository/{name}_repo.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 50ms

- [ ] 6.5 并发控制

  - File: `main/app/service/{name}_srv.go`
  - Purpose: 添加 UUID 锁
  - Requirements: 可靠性要求
  - Leverage: `pkg/lock/system_lock.go`
  - Success: 并发场景测试通过

- [ ] 6.6 文档更新

  - File: `docs/shared/api/{module}_api.md`, `CHANGELOG.md`
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
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - Payment/Order: 100%
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
- [ ] 遵循 `.cursor/rules/go-bmp.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/php.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/vue.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-{module}-{feature}/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-{module}-{feature}/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-{module}-{feature}/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-{module}-{feature}/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-{module}-{feature}/tasks.md)" | bc
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

### PHP 后端开发

```
Role: PHP Developer with ThinkPHP expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/php.mdc

Restrictions:
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除
- 遵循 PSR-2 代码风格

Success Criteria:
- {成功标准1}
- 代码通过 PSR-2 格式检查
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
Role: QA Engineer with {Go/PHP/Vue} testing expertise

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
- 遵循 .cursor/rules/go-main.mdc 或 .cursor/rules/php.mdc
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
**最后更新**: 2025-11-17  
**维护者**: 后端开发组
