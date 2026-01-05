# 外卖菜单导入进度条显示 任务分解

> 本文档定义 外卖菜单导入进度条显示功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 29  
**已完成**: 8 (Phase 1: 数据库设计和迁移, Phase 2: Go 导入进度管理模块)  
**进行中**: Phase 3 (集成到导入流程)  
**完成率**: 27.6%

---

## Phase 1: 数据库设计和迁移 (预计 0.5 天)

- [x] 1.1 设计数据库字段和表结构

  - File: `docs/shared/specs/active/story-shop-takeout-import-progress/design.md`
  - Purpose: 完成数据库设计,定义 ttpos_takeout 扩展字段和 ttpos_takeout_import_log 新表
  - Requirements: Requirement 1 (导入状态管理), Requirement 3 (历史日志记录)
  - Leverage: 现有表结构: `admin/database/migrations/`
  - Success: 设计文档已包含完整的数据库设计

- [x] 1.2 创建扩展 ttpos_takeout 表的迁移文件

  - File: `admin/database/migrations/20251216055139_add_import_status_to_ttpos_takeout.php`
  - Purpose: 为 ttpos_takeout 表添加进度相关字段
  - Requirements: Requirement 1.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件,为 ttpos_takeout 表添加 import_status, import_progress, import_start_time, import_end_time, import_error 字段和 idx_import_status 索引 | Context: 字段定义参考 design.md | Restrictions: 遵循 .cursor/rules/database.mdc,检查字段是否已存在 | Success: 迁移文件创建成功,字段定义正确

- [x] 1.3 创建 ttpos_takeout_import_log 表的迁移文件

  - File: `admin/database/migrations/20251216055140_create_ttpos_takeout_import_log_table.php`
  - Purpose: 创建导入日志表
  - Requirements: Requirement 3.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_takeout_import_log 表的迁移文件 | Context: 表结构参考 design.md,包含所有必需字段和索引 | Restrictions: 遵循 .cursor/rules/database.mdc,检查表是否已存在 | Success: 迁移文件创建成功,表结构正确

- [ ] 1.4 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建/修改表
  - Requirements: Requirement 1.1, 3.1
  - Leverage: Task 1.2, 1.3 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功,表已创建/修改

- [ ] 1.5 更新 Takeout Model

  - File: `main/app/modules/takeout/domain/model/takeout.go`
  - Purpose: 为 Takeout 模型添加进度相关字段
  - Requirements: Requirement 1.1
  - Leverage: 现有 Model: `main/app/modules/takeout/domain/model/takeout.go`
  - Prompt: Role: Go Developer | Task: 为 Takeout 结构体添加 ImportStatus, ImportProgress, ImportStartTime, ImportEndTime, ImportError 字段 | Context: 字段定义参考 design.md | Restrictions: 遵循 .cursor/rules/go-main.mdc,使用 gorm 标签 | Success: Model 更新成功,字段映射正确

- [ ] 1.6 创建 TakeoutImportLog Model

  - File: `main/app/model/takeout_import_log.go`
  - Purpose: 创建导入日志数据模型
  - Requirements: Requirement 3.1
  - Leverage: 现有 Model: `main/app/model/`，迁移文件: Task 1.3
  - Prompt: Role: Go Developer | Task: 创建 TakeoutImportLog 结构体,映射到 ttpos_takeout_import_log 表 | Context: 使用 gorm 标签,包含所有字段,实现 TableName() 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 创建成功,字段映射正确

---

## Phase 2: Repository 层实现 (预计 1 天)

### TakeoutImportLogRepo

- [ ] 2.1 创建 TakeoutImportLogRepo 接口

  - File: `main/app/repository/i_takeout_import_log_repo.go`
  - Purpose: 定义导入日志表的数据访问接口
  - Requirements: Requirement 3.2, 3.3, 3.7
  - Leverage: 现有 Repository 接口: `main/app/repository/i_*_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 ITakeoutImportLogRepo 接口,定义 CRUD 方法和选项方法 | Context: 包含 Create, Update, UpdateProgress, GetByUuid, GetList, Delete 方法,选项方法包含 WhereUuid, WherePlatform, WhereImportType, WhereStatus, OrderByCreateTime, Paginate | Restrictions: 遵循 .cursor/rules/go-main.mdc,使用选项模式(DBOption) | Success: 接口定义完整,方法签名正确

- [ ] 2.2 实现 TakeoutImportLogRepo

  - File: `main/app/repository/takeout_import_log_repo.go`
  - Purpose: 实现导入日志表的数据访问逻辑
  - Requirements: Requirement 3.2, 3.3, 3.7
  - Leverage: 现有 Repository 实现: `main/app/repository/*_repo.go`,Task 2.1 的接口
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 TakeoutImportLogRepoImpl,使用选项模式实现灵活查询 | Context: 只持有 db \*gorm.DB,实现所有接口方法和选项方法 | Restrictions: 不能持有 DBManager,使用 GORM,软删除(delete_time=0) | Success: Repository 实现完整,选项模式正确,软删除正确

- [ ] 2.3 编写 TakeoutImportLogRepo 单元测试

  - File: `main/app/repository/takeout_import_log_repo_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 TakeoutImportLogRepo 编写单元测试,覆盖率 ≥ 80% | Context: 测试 CRUD 方法,测试选项方法,测试软删除,测试分页 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%,所有测试通过

### TakeoutRepo 扩展

- [ ] 2.4 扩展 TakeoutRepo 接口

  - File: `main/app/repository/i_takeout_repo.go`
  - Purpose: 添加进度更新相关方法
  - Requirements: Requirement 1.2, 1.3
  - Leverage: 现有接口: `main/app/repository/i_takeout_repo.go`
  - Prompt: Role: Go Developer | Task: 在 ITakeoutRepo 接口中添加 UpdateProgress, UpdateImportStatus 方法 | Context: UpdateProgress 用于更新 import_progress 字段,UpdateImportStatus 用于更新 import_status 和相关字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口扩展完成,方法签名正确

- [ ] 2.5 实现 TakeoutRepo 扩展方法

  - File: `main/app/repository/takeout_repo.go`
  - Purpose: 实现进度更新相关方法
  - Requirements: Requirement 1.2, 1.3
  - Leverage: 现有实现: `main/app/repository/takeout_repo.go`,Task 2.4 的接口
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 UpdateProgress 和 UpdateImportStatus 方法 | Context: UpdateProgress 更新 import_progress 字段(JSON),UpdateImportStatus 更新 import_status, import_start_time, import_end_time, import_error 字段 | Restrictions: 使用 GORM Updates 方法 | Success: 方法实现正确,更新字段准确

---

## Phase 3: DTO 层实现 (预计 0.5 天)

- [ ] 3.1 创建 Request DTO

  - File: `main/app/dto/req/takeout_req.go` (扩展)
  - Purpose: 定义进度查询和日志查询的请求参数
  - Requirements: Requirement 2.7, 3.7
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 GetImportProgressReq 和 GetImportLogsReq 结构体 | Context: GetImportProgressReq 包含 Platform 字段,GetImportLogsReq 包含 Platform, ImportType, Status, PageNo, PageSize 字段,使用 binding 标签验证参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功,validation 正确

- [ ] 3.2 创建 Response DTO

  - File: `main/app/dto/resp/takeout_resp.go` (扩展)
  - Purpose: 定义进度和日志的响应数据
  - Requirements: Requirement 2.7, 3.7
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 ImportProgressResp, ImportLogResp, ImportLogListResp 结构体 | Context: ImportProgressResp 包含状态、进度、步骤信息,ImportLogResp 包含日志详情,ImportLogListResp 包含列表和 Meta 分页信息 | Restrictions: data 必须是对象,不能是 null 或数组 | Success: DTO 创建成功,响应格式正确

---

## Phase 4: Service 层实现 (预计 1.5 天)

### Service 接口扩展

- [ ] 4.1 扩展 TakeoutService 接口

  - File: `main/app/service/i_takeout_srv.go`
  - Purpose: 添加进度管理和日志查询方法
  - Requirements: Requirement 2, 3
  - Leverage: 现有接口: `main/app/service/i_takeout_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 ITakeoutSrv 接口中添加 GetImportProgress 和 GetImportLogs 方法 | Context: GetImportProgress 返回当前导入进度,GetImportLogs 返回历史日志列表 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口扩展完成,方法签名正确

### 核心方法实现

- [ ] 4.2 实现 checkImportStatus 方法

  - File: `main/app/service/takeout_srv.go`
  - Purpose: 检查当前导入状态,防止并发导入
  - Requirements: Requirement 1.1, 1.2
  - Leverage: 现有 Service: `main/app/service/takeout_srv.go`,Task 2.4-2.5 的 Repository
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 checkImportStatus 方法,检查是否有正在进行的导入任务 | Context: 获取 Takeout 记录,检查 import_status,如果为 1(导入中)且未超时则拒绝,如果超时则重置为失败状态 | Restrictions: 遵循 .cursor/rules/go-main.mdc,不使用 panic,返回 error | Success: 方法实现正确,并发控制有效,超时检测准确

- [ ] 4.3 实现 createImportLog 方法

  - File: `main/app/service/takeout_srv.go`
  - Purpose: 创建导入日志记录
  - Requirements: Requirement 3.2
  - Leverage: Task 2.1-2.2 的 TakeoutImportLogRepo
  - Prompt: Role: Go Developer | Task: 实现 createImportLog 方法,创建新的导入日志记录 | Context: 生成 UUID,设置 platform, import_type, import_direction,初始化 status=0(进行中),progress=0,start_time=当前时间 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确,日志创建成功

- [ ] 4.4 实现 updateImportProgress 方法

  - File: `main/app/service/takeout_srv.go`
  - Purpose: 更新导入进度(同时更新 ttpos_takeout 和 ttpos_takeout_import_log)
  - Requirements: Requirement 2.2, 2.3
  - Leverage: Task 2.4-2.5 的 TakeoutRepo,Task 2.1-2.2 的 TakeoutImportLogRepo
  - Prompt: Role: Go Developer | Task: 实现 updateImportProgress 方法,更新进度信息 | Context: 计算预估剩余时间,构建 progress JSON 数据,同时更新 ttpos_takeout.import_progress 和 ttpos_takeout_import_log.progress | Restrictions: 更新失败不阻塞主流程,记录错误日志 | Success: 方法实现正确,进度更新准确,不阻塞主流程

- [ ] 4.5 实现 completeImportLog 方法

  - File: `main/app/service/takeout_srv.go`
  - Purpose: 完成导入日志(记录最终结果)
  - Requirements: Requirement 3.3, 3.4
  - Leverage: Task 2.1-2.2 的 TakeoutImportLogRepo
  - Prompt: Role: Go Developer | Task: 实现 completeImportLog 方法,更新日志最终结果 | Context: 设置 status, progress=100, success_count, failure_count, total_count, end_time, duration, error_message | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确,日志完成记录准确

- [ ] 4.6 实现 updateImportStatus 方法

  - File: `main/app/service/takeout_srv.go`
  - Purpose: 更新导入状态
  - Requirements: Requirement 1.3, 1.4
  - Leverage: Task 2.4-2.5 的 TakeoutRepo
  - Prompt: Role: Go Developer | Task: 实现 updateImportStatus 方法,更新 ttpos_takeout 的导入状态 | Context: 更新 import_status, import_start_time/import_end_time, import_error | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确,状态更新准确

- [x] 4.7 改造 ImportMenu 方法

  - File: `main/app/service/takeout.go`
  - Purpose: 集成进度更新和日志记录到导入流程
  - Requirements: Requirement 1, 2, 3
  - Leverage: 现有 ImportMenu 方法,Task 4.2-4.6 的方法
  - Implementation: 已完成，实现内容：
    1. 导入前调用 CheckImportStatus 检查并发
    2. 调用 StartImport 创建导入日志
    3. 分阶段更新进度：分类(0-10%) → 规格(10-15%) → 单位(15-20%) → 商品(20-100%)
    4. 在 syncProducts 中按商品数量实时更新进度
    5. 导入完成后调用 CompleteImport 标记成功/失败
    6. 支持部分成功场景（有失败商品时标记为失败但返回成功数量）
  - Success: ✅ 方法改造成功,进度更新准确,日志记录完整,支持并发控制

- [ ] 4.8 实现 GetImportProgress 方法

  - File: `main/app/service/takeout_srv.go`
  - Purpose: 获取当前导入进度
  - Requirements: Requirement 2.7
  - Leverage: Task 2.4-2.5 的 TakeoutRepo,Task 3.2 的 Response DTO
  - Prompt: Role: Go Developer | Task: 实现 GetImportProgress 方法,返回当前导入进度 | Context: 获取 Takeout 记录,解析 import_progress JSON,转换为 ImportProgressResp | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确,进度数据准确

- [ ] 4.9 实现 GetImportLogs 方法

  - File: `main/app/service/takeout_srv.go`
  - Purpose: 获取导入日志列表
  - Requirements: Requirement 3.7
  - Leverage: Task 2.1-2.2 的 TakeoutImportLogRepo,Task 3.2 的 Response DTO
  - Prompt: Role: Go Developer | Task: 实现 GetImportLogs 方法,返回历史日志列表 | Context: 使用 TakeoutImportLogRepo 查询列表,支持平台、类型、状态筛选,支持分页,转换为 ImportLogListResp | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确,列表查询准确,分页正确

- [ ] 4.10 编写 TakeoutService 单元测试

  - File: `main/app/service/takeout_srv_test.go` (扩展)
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/takeout_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为新增和改造的方法编写单元测试,覆盖率 ≥ 70% | Context: 测试 checkImportStatus(正常、导入中、超时),测试 updateImportProgress,测试 GetImportProgress,测试 GetImportLogs(分页、筛选) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%,所有测试通过

---

## Phase 5: API 层实现 (预计 0.5 天)

- [ ] 5.1 实现 GetImportProgress API

  - File: `main/app/api/shop/takeout.go` (扩展)
  - Purpose: 实现获取导入进度接口
  - Requirements: Requirement 2.7
  - Leverage: 现有 API: `main/app/api/shop/takeout.go`,Task 4.1, 4.8 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetImportProgress API 方法 | Context: 绑定 Query 参数,调用 Service.GetImportProgress,使用 helper.Success() 返回响应,data 必须是对象 | Restrictions: 遵循 .cursor/rules/api.mdc,URL 使用 snake_case,不直接使用 c.JSON() | Success: API 实现成功,响应格式正确,参数验证正确

- [ ] 5.2 实现 GetImportLogs API

  - File: `main/app/api/shop/takeout.go` (扩展)
  - Purpose: 实现获取导入日志列表接口
  - Requirements: Requirement 3.7
  - Leverage: 现有 API: `main/app/api/shop/takeout.go`,Task 4.1, 4.9 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 实现 GetImportLogs API 方法 | Context: 绑定 Query 参数,调用 Service.GetImportLogs,使用 helper.Success() 返回响应,data 必须是对象,分页信息在 meta 中 | Restrictions: 遵循 .cursor/rules/api.mdc,URL 使用 snake_case | Success: API 实现成功,响应格式正确,分页正确

- [ ] 5.3 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册进度查询和日志查询路由
  - Requirements: Requirement 2.7, 3.7
  - Leverage: 现有路由: `main/router/router.go`
  - Prompt: Role: Go Developer | Task: 在 takeout 路由组中注册 GetImportProgress 和 GetImportLogs 路由 | Context: GET /api/v1/takeout/import/progress, GET /api/v1/takeout/import/logs | Restrictions: 遵循 .cursor/rules/go-main.mdc,URL 使用 snake_case | Success: 路由注册成功,可以访问

- [ ] 5.4 编写 API 集成测试

  - File: `main/app/api/shop/takeout_test.go` (扩展)
  - Purpose: 测试新增 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/shop/takeout_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 GetImportProgress 和 GetImportLogs API 编写集成测试 | Context: 测试正常场景,测试参数验证,测试响应格式,测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 6: 前端实现 (预计 1 天)

### API 封装和类型定义

- [ ] 6.1 创建 API 封装

  - File: `admin/views/shop/api/takeout.ts` (扩展)
  - Purpose: 封装进度查询和日志查询 API 调用
  - Requirements: Requirement 4
  - Leverage: 现有 API: `admin/views/shop/api/takeout.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装 getImportProgress 和 getImportLogs API 调用 | Context: 使用 axios,定义 TypeScript 请求参数类型 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成,类型定义正确

- [ ] 6.2 创建类型定义

  - File: `admin/views/shop/types/takeout.ts` (扩展)
  - Purpose: 定义进度和日志的 TypeScript 类型
  - Requirements: Requirement 4
  - Leverage: 现有类型: `admin/views/shop/types/takeout.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 定义 ImportProgressResp 和 ImportLogResp 接口 | Context: 根据后端 Response DTO 定义,包含所有字段 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 类型定义完成,与后端一致

### 进度对话框组件

- [ ] 6.3 创建 ImportProgressDialog 组件

  - File: `admin/views/shop/components/takeout/ImportProgressDialog.vue`
  - Purpose: 实现进度对话框组件
  - Requirements: Requirement 4.1, 4.2, 4.3
  - Leverage: Task 6.1-6.2 的 API 和类型,Element Plus Progress 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建 ImportProgressDialog 组件,实现进度展示和轮询逻辑 | Context: 使用 Element Plus Dialog 和 Progress,使用 Composition API,每 2.5 秒轮询一次,导入完成后停止轮询 | Restrictions: 遵循 .cursor/rules/vue.mdc,组件命名使用 PascalCase | Success: 组件创建成功,进度展示准确,轮询逻辑正确

- [ ] 6.4 实现轮询逻辑

  - File: `admin/views/shop/components/takeout/ImportProgressDialog.vue`
  - Purpose: 实现轮询获取进度
  - Requirements: Requirement 4.2
  - Leverage: Task 6.1 的 API
  - Prompt: Role: Frontend Developer | Task: 实现轮询逻辑,每 2.5 秒调用 getImportProgress API | Context: 使用 setInterval,导入完成或失败后停止轮询,组件卸载时清理 timer | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 轮询实现正确,资源清理完善

- [ ] 6.5 实现进度条动画

  - File: `admin/views/shop/components/takeout/ImportProgressDialog.vue`
  - Purpose: 实现进度条平滑过渡动画
  - Requirements: Requirement 4.3
  - Leverage: Element Plus Progress 组件
  - Prompt: Role: Frontend Developer | Task: 实现进度条平滑过渡动画 | Context: 使用 Element Plus Progress 组件,percentage 变化时自动平滑过渡 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 动画效果流畅

### 日志列表组件

- [ ] 6.6 创建 ImportLogList 组件

  - File: `admin/views/shop/components/takeout/ImportLogList.vue`
  - Purpose: 实现历史日志列表组件
  - Requirements: Requirement 4.5, 4.6, 4.7
  - Leverage: Task 6.1-6.2 的 API 和类型,Element Plus Timeline 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建 ImportLogList 组件,实现日志列表展示、筛选和分页 | Context: 使用 Element Plus Timeline, Select, Pagination,使用 Composition API,支持按状态筛选,支持分页 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 组件创建成功,列表展示完整,筛选和分页正常

- [ ] 6.7 实现状态区分展示

  - File: `admin/views/shop/components/takeout/ImportLogList.vue`
  - Purpose: 区分不同状态的日志(进行中/成功/失败)
  - Requirements: Requirement 4.7
  - Leverage: Element Plus Tag, Progress, Alert 组件
  - Prompt: Role: Frontend Developer | Task: 实现状态区分展示逻辑 | Context: 进行中显示进度条,成功显示绿色标签和统计信息,失败显示红色标签和错误信息 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 状态区分清晰,视觉效果好

### 页面集成

- [ ] 6.8 集成进度对话框到导入页面

  - File: `admin/views/shop/pages/takeout/import.vue` (修改)
  - Purpose: 在导入页面中使用进度对话框
  - Requirements: Requirement 4.1
  - Leverage: Task 6.3-6.5 的 ImportProgressDialog 组件
  - Prompt: Role: Frontend Developer | Task: 在导入页面中集成 ImportProgressDialog 组件 | Context: 点击导入按钮后显示进度对话框,导入完成后刷新页面或显示结果 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 集成成功,交互流畅

- [ ] 6.9 创建日志列表页面

  - File: `admin/views/shop/pages/takeout/logs.vue` (新增)
  - Purpose: 创建历史日志列表页面
  - Requirements: Requirement 4.5
  - Leverage: Task 6.6-6.7 的 ImportLogList 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建日志列表页面,使用 ImportLogList 组件 | Context: 页面布局使用 Element Plus Card,页面标题"导入日志" | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 页面创建成功,布局美观

- [ ] 6.10 添加日志页面路由

  - File: `admin/views/shop/router/index.ts`
  - Purpose: 添加日志列表页面路由
  - Requirements: Requirement 4.5
  - Leverage: 现有路由配置
  - Success: 路由添加成功,可以访问日志页面

---

## Phase 7: 测试与优化 (预计 0.5 天)

- [ ] 7.1 集成测试

  - File: `test/integration/takeout_import_progress_test.go` (新增)
  - Purpose: 测试端到端导入进度功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整的导入流程(开始→进行中→完成),测试进度查询,测试日志查询,测试并发控制 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 7.2 性能测试

  - File: -
  - Purpose: 确保进度查询和日志查询性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具(如: wrk, ab)
  - Success: 进度查询接口 < 200ms,日志查询接口 < 200ms

- [ ] 7.3 并发测试

  - File: -
  - Purpose: 测试并发导入控制逻辑
  - Requirements: Requirement 1.1, 1.2
  - Leverage: 并发测试工具
  - Success: 同一商户不能同时进行多个导入任务,并发控制有效

- [ ] 7.4 超时测试

  - File: -
  - Purpose: 测试超时导入任务检测和重置机制
  - Requirements: Requirement 1.6
  - Leverage: 模拟超时场景
  - Success: 超过 30 分钟的导入任务被正确检测和重置

- [ ] 7.5 浏览器兼容性测试

  - File: -
  - Purpose: 确保前端在各浏览器中正常工作
  - Requirements: 浏览器兼容性要求
  - Leverage: Chrome, Safari, Firefox, Edge 浏览器
  - Success: 所有浏览器测试通过

- [ ] 7.6 数据库查询优化

  - File: `main/app/repository/takeout_import_log_repo.go`
  - Purpose: 优化日志查询性能
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析,索引优化
  - Success: 查询时间 < 50ms,索引使用正确

- [ ] 7.7 文档更新

  - File: `CHANGELOG.md`
  - Purpose: 更新变更日志
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG
  - Success: 变更日志已更新,包含新功能描述

---

## 提交清单

完成所有任务后,请检查:

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
  - [ ] 导入前检查状态,导入中拒绝新请求
  - [ ] 导入过程中实时更新进度
  - [ ] 所有导入记录保存到日志表
  - [ ] 进度对话框美观流畅
  - [ ] 日志列表展示完整
  - [ ] 并发控制有效
  - [ ] 超时任务自动重置

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] requirements.md 审核状态为"已通过"
- [ ] design.md 完整准确
- [ ] tasks.md 所有任务已完成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-takeout-import-progress/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-takeout-import-progress/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-takeout-import-progress/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-takeout-import-progress/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-takeout-import-progress/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的详细设计
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板,让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`, `npm run lint`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit(参考 `.cursor/rules/version.mdc`)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板:`docs/agent/templates/graphiti-episode.md`
- 活动日志:`docs/team/activities/weifashi/2025-12/2025-12-16.md`
- 在执行任务过程中若总结出经验或规避策略,请记录 Episode,并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: weifashi  
**最后更新**: 2025-12-16

