# POS 收银机 - 外卖平台订单集成 任务分解（多平台）

> 本文档定义外卖平台订单集成的详细执行任务清单（后端部分）。支持 Grab、Foodpanda、Lineman 等多个外卖平台。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求
- **多平台支持**: 优先实现 Grab，设计支持后续扩展其他平台

## 📊 进度总览

**总任务数**: 49  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

**分阶段统计**:
- Phase 1 (数据库): 7 任务
- Phase 2 (Repository): 9 任务
- Phase 3 (Service): 11 任务
- Phase 4 (API): 9 任务
- Phase 5 (RPC 集成): 6 任务
- Phase 6 (测试优化): 7 任务

---

## Phase 1: 数据库设计和迁移（2 天）

### ⚠️ 前置条件

**必须先与 ttpos-bmp 团队确认各平台 RPC 接口定义**（优先 Grab），再开始数据库设计。

---

- [ ] 1.1 创建 ttpos_takeout_order 表迁移文件

  - File: `admin/database/migrations/20251222100000_create_ttpos_takeout_order_table.php`
  - Purpose: 定义外卖订单主表结构（支持多平台）
  - Requirements: Requirement 1 (订单同步), design.md 数据库设计
  - Leverage: 现有迁移文件: `admin/database/migrations/`, `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建 ttpos_takeout_order 表的迁移文件，包含 platform 字段区分平台，包含订单信息、状态、价格、货币、时间等字段 | Context: 必须包含 id, uuid, takeout_order_uuid, platform (grab/foodpanda/lineman), platform_order_id (唯一), create_time, update_time, delete_time，price 字段使用 bigint（单位：分），时间字段使用 int，移除 shop_uuid 和 company_uuid | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查表是否存在，使用软删除，唯一索引包含 platform | Success: 迁移文件创建成功，字段定义正确，索引设计合理，支持多平台

- [ ] 1.2 创建 ttpos_takeout_order_items 表迁移文件

  - File: `admin/database/migrations/20251222100001_create_ttpos_takeout_order_items_table.php`
  - Purpose: 定义外卖订单商品表结构（支持多平台）
  - Requirements: Requirement 1, Requirement 2 (商品关联检查)
  - Leverage: Task 1.1 的迁移文件
  - Prompt: 添加 platform 字段，platform_item_id 替代 grab_item_id，移除 shop_uuid
  - Success: 迁移文件创建成功，支持多平台

- [ ] 1.3 创建 ttpos_takeout_item_mapping 表迁移文件

  - File: `admin/database/migrations/20251222100002_create_ttpos_takeout_item_mapping_table.php`
  - Purpose: 定义外卖商品关联映射表结构（支持多平台）
  - Requirements: Requirement 2 (商品关联检查)
  - Leverage: Task 1.1 的迁移文件
  - Prompt: 添加 platform 字段，唯一索引 (platform, platform_item_id, delete_time)，移除 shop_uuid
  - Success: 迁移文件创建成功，唯一索引正确

- [ ] 1.4 创建 ttpos_takeout_modifier_mapping 表迁移文件

  - File: `admin/database/migrations/20251222100003_create_ttpos_takeout_modifier_mapping_table.php`
  - Purpose: 定义外卖修饰符关联映射表结构（支持多平台）
  - Requirements: Requirement 2 (商品关联检查)
  - Leverage: Task 1.1 的迁移文件
  - Prompt: 添加 platform 字段，platform_modifier_id 替代 grab_modifier_id，移除 shop_uuid
  - Success: 迁移文件创建成功

- [ ] 1.5 创建 ttpos_takeout_sync_logs 表迁移文件

  - File: `admin/database/migrations/20251222100004_create_ttpos_takeout_sync_logs_table.php`
  - Purpose: 定义外卖订单同步日志表结构（支持多平台）
  - Requirements: Requirement 1 (订单同步)
  - Leverage: Task 1.1 的迁移文件
  - Prompt: 添加 platform 字段，移除 shop_uuid
  - Success: 迁移文件创建成功

- [ ] 1.6 创建 ttpos_takeout_settings 表迁移文件

  - File: `admin/database/migrations/20251222100005_create_ttpos_takeout_settings_table.php`
  - Purpose: 定义外卖平台配置表结构（支持多平台）
  - Requirements: Requirement 6 (自动接单配置)
  - Leverage: Task 1.1 的迁移文件
  - Prompt: 添加 platform 和 is_enabled 字段，唯一索引 (platform, delete_time)，移除 shop_uuid 和 company_uuid，每个平台一条全局配置
  - Success: 迁移文件创建成功，唯一索引正确

- [ ] 1.7 执行数据库迁移并创建 Go Model

  - File: `main/app/model/takeout_*.go` (6 个模型文件)
  - Purpose: 在数据库中创建表，并定义 Go 数据模型
  - Requirements: 所有数据库相关需求
  - Leverage: Task 1.1-1.6 的迁移文件，现有 Model: `main/app/model/`
  - Command: `cd admin && php think migrate:run`
  - Prompt: Role: Go Developer | Task: 创建 6 个 Go Model 结构体，映射到 ttpos_takeout_* 表 | Context: 使用 gorm 标签，包含所有字段（含 platform），实现 TableName() 方法，字段类型与数据库对应 | Restrictions: 遵循 .cursor/rules/go-main.mdc，时间字段使用 int64，金额字段使用 int64（分），软删除字段 delete_time，移除 ShopUuid 和 CompanyUuid | Success: 迁移执行成功，所有表已创建，Go Model 创建成功，字段映射正确，支持多平台

---

## Phase 2: Repository 层实现（1 天）

### 2.1 Takeout Order Repository

- [ ] 2.1.1 创建 Takeout Order Repository 接口和实现

  - File: `main/app/repository/takeout_order_repo.go`
  - Purpose: 定义外卖订单数据访问接口和实现（支持多平台）
  - Requirements: Requirement 4 (订单列表), Requirement 5 (接单/拒单)
  - Leverage: 现有 Repository: `main/app/repository/*_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 ITakeoutOrderRepo 接口和实现，定义 CRUD 方法和选项方法 | Context: 使用选项模式(DBOption)，包含 Create, Update, GetByUuid, GetByPlatformOrderID, GetList, UpdateState, Delete 方法，以及 WherePlatform, WhereOrderState, WhereOrderTime 等选项方法，支持按 platform 筛选 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager，接口和实现在同一文件 | Success: 接口定义完整，实现正确，选项模式支持多平台筛选

- [ ] 2.1.2 编写 Takeout Order Repository 单元测试

  - File: `main/app/repository/takeout_order_repo_test.go`
  - Purpose: 确保 Takeout Order Repository 数据访问正确
  - Requirements: 测试要求（Order 模块 100%）
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 TakeoutOrderRepo 编写单元测试，覆盖率 100% | Context: 测试 CRUD 方法，测试选项方法（含 platform 筛选），测试软删除，测试分页和筛选，测试多平台场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 testify 断言 | Success: 测试覆盖率 100%，所有测试通过

### 2.2 Takeout Item Mapping Repository

- [ ] 2.2.1 创建 Takeout Item Mapping Repository 接口和实现

  - File: `main/app/repository/takeout_item_mapping_repo.go`
  - Purpose: 定义商品关联映射数据访问接口和实现（支持多平台）
  - Requirements: Requirement 2 (商品关联检查)
  - Leverage: Task 2.1.1 的模式
  - Prompt: 支持按 platform 和 platform_item_id 查询，移除 shop_uuid 相关逻辑
  - Success: 接口和实现完整，支持多平台

- [ ] 2.2.2 编写 Takeout Item Mapping Repository 单元测试

  - File: `main/app/repository/takeout_item_mapping_repo_test.go`
  - Purpose: 确保数据访问正确
  - Requirements: 测试要求（≥ 80%）
  - Leverage: Task 2.1.2 的测试模式
  - Success: 测试覆盖率 ≥ 80%

### 2.3 Takeout Sync Log Repository

- [ ] 2.3.1 创建 Takeout Sync Log Repository 接口和实现

  - File: `main/app/repository/takeout_sync_log_repo.go`
  - Purpose: 定义同步日志数据访问接口和实现（支持多平台）
  - Requirements: Requirement 1 (订单同步日志)
  - Leverage: Task 2.1.1 的模式
  - Prompt: 支持按 platform 筛选日志
  - Success: 接口和实现完整

- [ ] 2.3.2 编写 Takeout Sync Log Repository 单元测试

  - File: `main/app/repository/takeout_sync_log_repo_test.go`
  - Purpose: 确保数据访问正确
  - Requirements: 测试要求（≥ 80%）
  - Leverage: Task 2.1.2 的测试模式
  - Success: 测试覆盖率 ≥ 80%

---

## Phase 3: Service 层实现（3 天）

### 3.1 DTO 定义和平台常量

- [ ] 3.1.1 创建平台常量定义

  - File: `main/app/constant/takeout_platform.go`
  - Purpose: 定义外卖平台常量
  - Requirements: 多平台支持
  - Prompt: Role: Go Developer | Task: 创建 takeout_platform.go，定义平台代码常量 | Context: 定义 TakeoutPlatformGrab, TakeoutPlatformFoodpanda, TakeoutPlatformLineman, TakeoutPlatformShopeefood，订单状态常量，库存状态常量 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 常量定义完整，易于扩展

- [ ] 3.1.2 创建 Takeout Order Request DTO

  - File: `main/app/dto/req/takeout_order_req.go`
  - Purpose: 定义外卖订单 API 请求参数（支持多平台）
  - Requirements: Requirement 4, 5, 6
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 Request DTO，包含 TakeoutOrderListReq(含 platform 筛选), TakeoutOrderAcceptReq, TakeoutOrderRejectReq, TakeoutSettingsSaveReq(含 platform) | Context: 使用 binding 标签验证参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确，支持多平台

- [ ] 3.1.3 创建 Takeout Order Response DTO

  - File: `main/app/dto/resp/takeout_order_resp/takeout_order_resp.go`
  - Purpose: 定义外卖订单 API 响应数据（支持多平台）
  - Requirements: Requirement 4, 5, 6
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 Response DTO，包含 TakeoutOrderResp(含 platform 字段), TakeoutOrderListResp, TakeoutSettingsResp, TakeoutRejectReasonResp | Context: 包含 Meta 分页信息，金额字段使用 int64（分） | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

### 3.2 平台 RPC Client (优先 Grab)

- [ ] 3.2.1 创建 Grab RPC 客户端接口和实现

  - File: `main/app/modules/takeout/infrastructure/adapter/grab/grab_rpc_client.go`
  - Purpose: 定义和实现 Grab RPC 调用（接口和实现在同一文件）
  - Requirements: Requirement 1 (RPC 对接)
  - Leverage: 现有 Service 模式，ttpos-bmp 生成的 gRPC 客户端代码
  - Prompt: Role: gRPC Developer | Task: 创建 IGrabRPCClient 接口和实现，调用 ttpos-bmp 的 Grab RPC 服务 | Context: 使用 Nacos 服务发现，设置 10 秒超时，实现重试策略（最多 3 次），方法包含 GetNewOrders, AcceptOrder, RejectOrder, GetRejectReasons | Restrictions: 遵循 .cursor/rules/go-main.mdc，记录详细日志，错误使用 errors.WithMessage 包装 | Success: RPC 客户端完整，超时和重试正确

### 3.3 Takeout Order Service

- [ ] 3.3.1 创建 Takeout Order Service 接口和实现

  - File: `main/app/service/takeout_order_srv.go`
  - Purpose: 定义和实现外卖订单业务逻辑（支持多平台）
  - Requirements: Requirement 1-7
  - Leverage: 现有 Service: `main/app/service/*_srv.go`，Task 2.1-2.3 的 Repository，Task 3.2.1 的 RPC Client
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 ITakeoutOrderSrv 接口和实现，包含业务方法 | Context: 持有 DBManager，依赖 GrabRPCClient, ProductService, KDSService（接口），包含 SyncNewOrders(platform), GetOrderList(platform筛选), GetOrderDetail, AcceptOrder, RejectOrder, CheckItemMapping(platform), CheckStock, AutoAcceptOrder(platform) 等方法，使用事务管理 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只依赖其他 Service 接口，不使用 panic，接口和实现在同一文件 | Success: Service 实现完整，业务逻辑正确，支持多平台

- [ ] 3.3.2 实现商品关联检查逻辑（支持多平台）

  - File: `main/app/service/takeout_order_srv.go` (方法: CheckItemMapping)
  - Purpose: 检查平台商品是否关联到 TTPOS 商品
  - Requirements: Requirement 2 (商品关联检查)
  - Leverage: Task 2.2 的 Mapping Repository，现有 ProductService
  - Prompt: Role: Go Developer | Task: 实现 CheckItemMapping 方法，检查订单中所有商品和修饰符是否已关联 | Context: 根据 platform 和 platform_item_id 查询 takeout_item_mapping 和 takeout_modifier_mapping 表，返回未关联的商品/修饰符列表，标记订单异常 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 检查逻辑正确，支持多平台

- [ ] 3.3.3 实现库存检查逻辑

  - File: `main/app/service/takeout_order_srv.go` (方法: CheckStock)
  - Purpose: 检查订单商品库存是否充足
  - Requirements: Requirement 3 (库存检查)
  - Leverage: 现有 ProductService（库存查询方法）
  - Success: 库存检查逻辑正确

- [ ] 3.3.4 实现自动接单判断逻辑（支持多平台）

  - File: `main/app/service/takeout_order_srv.go` (方法: AutoAcceptOrder)
  - Purpose: 根据配置判断是否自动接单
  - Requirements: Requirement 6 (自动接单配置)
  - Leverage: Task 3.4 的 Settings Service
  - Prompt: 根据 platform 读取配置，检查自动接单开关、金额上限、商品关联、库存
  - Success: 自动接单判断逻辑正确，支持多平台配置

- [ ] 3.3.5 编写 Takeout Order Service 单元测试

  - File: `main/app/service/takeout_order_srv_test.go`
  - Purpose: 确保 Takeout Order Service 业务逻辑正确
  - Requirements: 测试要求（Order 模块 100%）
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Mock GrabRPCClient/ProductService/KDSService，测试多平台场景
  - Success: 测试覆盖率 100%，所有测试通过

### 3.4 Takeout Settings Service

- [ ] 3.4.1 创建 Takeout Settings Service 接口和实现

  - File: `main/app/service/takeout_settings_srv.go`
  - Purpose: 定义和实现外卖平台配置管理业务逻辑（支持多平台）
  - Requirements: Requirement 6 (配置管理)
  - Leverage: Task 3.3.1 的模式
  - Prompt: Role: Go Developer | Task: 实现 TakeoutSettingsSrv，包含 GetSettings(platform), SaveSettings(platform, settings) 方法 | Context: 持有 DBManager，实现配置读取和保存，使用 Redis 缓存（Key: ttpos:takeout:settings:{platform}，TTL: 1 小时），保存时更新缓存 | Restrictions: 遵循 .cursor/rules/go-main.mdc，接口和实现在同一文件 | Success: Service 实现完整，缓存逻辑正确，支持多平台

- [ ] 3.4.2 编写 Takeout Settings Service 单元测试

  - File: `main/app/service/takeout_settings_srv_test.go`
  - Purpose: 确保配置管理逻辑正确
  - Requirements: 测试要求（≥ 70%）
  - Prompt: 测试多平台配置的读取和保存
  - Success: 测试覆盖率 ≥ 70%

---

## Phase 4: API 层实现（2 天）

### 4.1 Takeout Order API

- [ ] 4.1.1 创建 Takeout Order API Controller

  - File: `main/app/api/takeout_order_api.go`
  - Purpose: 实现外卖订单 HTTP API 接口（支持多平台）
  - Requirements: Requirement 4, 5
  - Leverage: 现有 API: `main/app/api/*_api.go`，Task 3.3 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 TakeoutOrderAPI，实现 GetOrderList(支持platform参数), GetOrderDetail, AcceptOrder, RejectOrder, GetRejectReasons(:platform) 方法 | Context: URL 使用 snake_case，路由 /api/v1/takeout/orders，使用 helper.Success() 返回响应，data 必须是对象，参数验证使用 ShouldBindJSON | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，支持多平台

- [ ] 4.1.2 注册 Takeout Order API 路由

  - File: `main/router/router.go`
  - Purpose: 注册外卖订单 API 路由
  - Requirements: Requirement 4, 5
  - Leverage: 现有路由: `main/router/router.go`
  - Prompt: Role: Go Developer | Task: 在 router.go 中注册 Takeout Order API 路由 | Context: 路由组 /api/v1/takeout，包含 GET orders(支持platform查询参数), GET orders/:uuid, POST orders/:uuid/accept, POST orders/:uuid/reject, GET :platform/reject_reasons | Restrictions: 使用 JWTAuth 中间件，使用 Permission 中间件（外卖权限） | Success: 路由注册成功

- [ ] 4.1.3 编写 Takeout Order API 集成测试

  - File: `main/app/api/takeout_order_api_test.go`
  - Purpose: 测试外卖订单 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: 测试多平台场景，测试 platform 参数筛选
  - Success: 所有 API 测试通过

### 4.2 Takeout Settings API

- [ ] 4.2.1 创建 Takeout Settings API Controller

  - File: `main/app/api/takeout_settings_api.go`
  - Purpose: 实现外卖平台配置管理 HTTP API 接口（支持多平台）
  - Requirements: Requirement 6
  - Leverage: Task 4.1.1 的 API 模式，Task 3.4 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 TakeoutSettingsAPI，实现 GetSettings(:platform), SaveSettings(:platform) 方法 | Context: URL 使用 snake_case，路由 /api/v1/shop/takeout/:platform/settings，使用 helper.Success()，参数验证 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 创建成功，支持多平台配置

- [ ] 4.2.2 注册 Takeout Settings API 路由

  - File: `main/router/router.go`
  - Purpose: 注册外卖平台配置 API 路由
  - Requirements: Requirement 6
  - Leverage: Task 4.1.2
  - Prompt: 路由 /api/v1/shop/takeout/:platform/settings
  - Success: 路由注册成功

- [ ] 4.2.3 编写 Takeout Settings API 集成测试

  - File: `main/app/api/takeout_settings_api_test.go`
  - Purpose: 测试外卖平台配置 API 接口
  - Requirements: 测试要求
  - Leverage: Task 4.1.3
  - Prompt: 测试多平台配置管理
  - Success: 所有 API 测试通过

### 4.3 API 文档

- [ ] 4.3.1 创建 Takeout API 文档

  - File: `docs/shared/api/takeout_api.md`
  - Purpose: 记录外卖平台 API 接口文档（支持多平台）
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`, design.md 中的 API 设计
  - Prompt: Role: Technical Writer | Task: 创建 Takeout API 文档，包含所有接口的 URL、Method、Request、Response、Error，说明多平台支持 | Context: 参考 design.md 中的 API 设计，包含 platform 参数说明 | Restrictions: 文档准确完整 | Success: API 文档创建完成

---

## Phase 5: 平台 RPC 集成和数据转换（1 天）

### ⚠️ 前置条件

**必须等待 ttpos-bmp 团队完成各平台 RPC 服务开发和部署**（优先 Grab）。

---

- [ ] 5.1 实现 Grab 数据转换器

  - File: `main/app/modules/takeout/infrastructure/adapter/grab/grab_converter.go`
  - Purpose: 实现 Grab RPC 数据与 TTPOS 数据之间的转换
  - Requirements: Requirement 1 (RPC 对接)
  - Leverage: 现有 `grab_converter.go`（已存在，需扩展），design.md 中的数据模型
  - Prompt: Role: Go Developer | Task: 扩展 grab_converter.go，实现 ConvertGrabOrderToModel, ConvertModelToGrabOrder 方法 | Context: 将 Grab RPC 返回的 GrabOrderData 转换为 model.TakeoutOrder（含 platform 字段），处理货币信息、价格信息（分）、时间戳转换 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 转换逻辑正确，字段映射完整，支持多平台

- [ ] 5.2 实现订单同步 Worker（支持多平台）

  - File: `main/app/modules/takeout/application/worker/grab_sync_worker.go`
  - Purpose: 定时同步 Grab 新订单
  - Requirements: Requirement 1 (订单同步)
  - Leverage: Task 3.2.1 的 RPC Client，Task 5.1 的转换器，Task 3.3.1 的 Service
  - Prompt: Role: Go Developer | Task: 创建 GrabSyncWorker，实现定时同步逻辑 | Context: 使用 cron 或 ticker 定时调用 GrabRPCClient.GetNewOrders()，转换数据（设置 platform="grab"），调用 TakeoutOrderService.SyncNewOrders("grab") 保存，执行商品关联检查和库存检查，判断自动接单 | Restrictions: 遵循 .cursor/rules/go-main.mdc，记录同步日志，错误处理 | Success: Worker 实现完整，同步逻辑正确

- [ ] 5.3 实现失败重试 Worker（支持多平台）

  - File: `main/app/modules/takeout/application/worker/takeout_retry_worker.go`
  - Purpose: 处理平台 RPC 调用失败的重试
  - Requirements: Requirement 1 (失败重试)
  - Leverage: Task 3.2.1 的 RPC Client，Redis 队列
  - Prompt: Role: Go Developer | Task: 创建 TakeoutRetryWorker，消费 Redis 重试队列 | Context: 从 Redis List (ttpos:takeout:retry_queue:{platform}) 中取出失败任务（order_id, action, platform），重新调用 RPC，最多重试 3 次，记录日志 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Worker 实现完整，重试逻辑正确，支持多平台

- [ ] 5.4 集成 KDS 通知

  - File: `main/app/service/takeout_order_srv.go` (方法: notifyKDS)
  - Purpose: 接单成功后通知 KDS 厨显系统
  - Requirements: Requirement 7 (KDS 联动)
  - Leverage: 现有 KDS Service
  - Prompt: Role: Go Developer | Task: 实现 notifyKDS 方法，调用 KDSService.NotifyNewOrder() | Context: 接单成功后异步调用，传递订单信息，标识为外卖订单（包含 platform 信息） | Restrictions: 遵循 .cursor/rules/go-main.mdc，依赖 KDSService 接口 | Success: KDS 通知逻辑正确

- [ ] 5.5 注册 Worker 到启动流程

  - File: `main/cmd/server/main.go`
  - Purpose: 在应用启动时启动 Worker
  - Requirements: Requirement 1 (订单同步)
  - Leverage: Task 5.2, 5.3 的 Worker，现有 Worker 启动模式
  - Success: Worker 启动成功

- [ ] 5.6 RPC 集成测试

  - File: `test/integration/grab_rpc_test.go`
  - Purpose: 测试 Grab RPC 集成
  - Requirements: 测试要求
  - Leverage: Task 3.2.1 的 RPC Client，Task 5.1 的转换器
  - Prompt: Role: QA Engineer | Task: 编写 Grab RPC 集成测试 | Context: Mock Grab RPC 服务响应，测试订单同步流程，测试接单/拒单流程，测试错误处理和重试，验证 platform 字段正确 | Restrictions: 使用 Mock，不依赖真实 Grab RPC 服务 | Success: 集成测试通过

---

## Phase 6: 测试和优化（2 天）

- [ ] 6.1 端到端集成测试（多平台场景）

  - File: `test/integration/takeout_order_e2e_test.go`
  - Purpose: 测试完整的订单处理流程
  - Requirements: 所有功能需求
  - Leverage: 所有已实现的模块
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试订单同步 → 商品关联检查 → 库存检查 → 待接单 → 手动接单 → KDS 通知 → 已接单，测试自动接单流程，测试拒单流程，测试多平台数据隔离 | Restrictions: Mock 平台 RPC，测试真实用户场景 | Success: 端到端测试通过

- [ ] 6.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求（订单列表 < 200ms，接单/拒单 < 500ms）
  - Leverage: 性能测试工具（wrk, ab）
  - Command: `wrk -t4 -c100 -d30s http://localhost:8080/api/v1/takeout/orders?platform=grab`
  - Success: 订单列表接口 < 200ms，接单/拒单接口 < 500ms

- [ ] 6.3 数据库查询优化

  - File: `main/app/repository/takeout_order_repo.go`
  - Purpose: 优化 SQL 查询性能
  - Requirements: 性能要求（查询 < 50ms）
  - Leverage: EXPLAIN 分析，现有索引设计
  - Prompt: 优化多平台查询，确保 platform 索引使用正确
  - Success: 查询时间 < 50ms，索引使用正确

- [ ] 6.4 Redis 缓存实现（多平台）

  - File: `main/app/service/takeout_settings_srv.go`, `main/app/service/takeout_order_srv.go`
  - Purpose: 实现 Redis 缓存策略
  - Requirements: 性能要求（缓存命中率 > 80%）
  - Leverage: `pkg/redis/`
  - Prompt: 缓存键包含 platform，如 ttpos:takeout:settings:{platform}
  - Success: 缓存实现完成，配置缓存 TTL 1 小时，命中率 > 80%

- [ ] 6.5 并发控制

  - File: `main/app/service/takeout_order_srv.go` (方法: AcceptOrder)
  - Purpose: 添加 Redis 分布式锁防止并发接单
  - Requirements: 可靠性要求
  - Leverage: `pkg/lock/system_lock.go`
  - Prompt: Role: Go Developer | Task: 在 AcceptOrder 方法中添加 Redis 分布式锁 | Context: 锁 Key: ttpos:takeout:lock:accept:{order_uuid}，TTL: 10 秒，加锁后再执行接单逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 并发场景测试通过，不会重复接单

- [ ] 6.6 监控和日志

  - File: `main/app/service/takeout_order_srv.go`, `main/app/modules/takeout/application/worker/grab_sync_worker.go`
  - Purpose: 添加监控指标和详细日志
  - Requirements: 可靠性要求
  - Leverage: `logger.Logger`, Prometheus metrics
  - Prompt: 日志包含 platform 字段，便于多平台问题排查
  - Success: 日志记录完整，包含关键操作日志和错误日志

- [ ] 6.7 文档更新

  - File: `docs/shared/api/takeout_api.md`, `CHANGELOG.md`, `README.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: Task 4.3.1 的 API 文档
  - Prompt: 文档说明多平台支持和扩展方式
  - Success: 所有文档已更新，CHANGELOG 已记录变更

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - Order 模块: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] RPC 接口已与 ttpos-bmp 团队联调

### 文档同步

- [ ] API 文档已更新
- [ ] 数据库文档已更新（迁移脚本）
- [ ] CHANGELOG.md 已更新
- [ ] README.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/feature-pos-grab-order-integration/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/feature-pos-grab-order-integration/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/feature-pos-grab-order-integration/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/feature-pos-grab-order-integration/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/feature-pos-grab-order-integration/tasks.md)" | bc
```

---

## 执行流程

1. **选择任务**: 选择下一个未完成任务（按 Phase 顺序执行）
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计细节
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2025-12/2025-12-22.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v2.0.0  
**创建日期**: 2025-12-22  
**更新日期**: 2025-12-22  
**作者**: weifashi  
**预计工期**: 12 天（49 任务，已调整为多平台支持）
**当前实现**: 优先 Grab 平台，设计支持 Foodpanda、Lineman、ShopeeFood 扩展

