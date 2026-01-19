# 全平台统一账号 任务分解

> 本文档定义全平台统一账号功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 56  
**已完成**: 27  
**进行中**: -  
**完成率**: 48.2%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建统一账号表迁移文件（saas 数据库）

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_staff_table_in_saas.php`
  - Purpose: 在 saas 数据库中创建 ttpos_staff 表，作为统一账号表
  - Requirements: 1.1, 1.2
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 在 saas 数据库中创建 ttpos_staff 表的迁移文件，遵循 design.md 中的数据库设计 | Context: 必须包含 id, uuid, email(唯一), phone(允许空字符串，不建唯一索引), real_name(varchar(255), 姓名), password, password_change_count, password_change_time, is_disable, last_company_uuid(上次登录新管理端的商家UUID), create_time, update_time, delete_time 字段，时间字段使用 int 类型 | Restrictions: 遵循 .cursor/rules/database.mdc，邮箱建立唯一索引，手机号不建立唯一索引（因为允许空字符串），手机号和 last_company_uuid 建立普通索引用于查询优化 | Success: 迁移文件创建成功，字段定义正确，索引正确

- [x] 1.2 确认账号-门店关联表（使用现有 ttpos_company_staff）

  - File: -
  - Purpose: 确认使用现有的 ttpos_company_staff 表作为账号-门店关联表
  - Requirements: 1.1, 2.2
  - Leverage: 现有表: `ttpos_company_staff`（saas 数据库）
  - Prompt: Role: Database Engineer | Task: 确认 ttpos_company_staff 表结构符合需求 | Context: 表已存在，uuid 字段关联到 ttpos_staff.uuid，company_uuid 字段关联到门店 | Restrictions: 表结构暂不需要调整 | Success: 确认表结构符合需求

- [ ] 1.3 执行数据库迁移

  - File: -
  - Purpose: 在 saas 数据库中创建 ttpos_staff 表
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，表已创建

- [x] 1.4 创建 SaasStaff Go Model

  - File: `main/app/model/saas_staff.go`
  - Purpose: 定义 saas 数据库统一账号数据模型，与数据库表对应
  - Requirements: 1.1
  - Leverage: 现有 Model: `main/app/model/staff.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 创建 SaasStaff 结构体，映射到 saas 数据库的 ttpos_staff 表 | Context: 使用 gorm 标签，包含所有字段（包括 real_name, last_company_uuid），实现 TableName() 方法，邮箱字段需要 uniqueIndex 标签，手机号和 last_company_uuid 字段需要 index 标签（不设置 uniqueIndex） | Restrictions: 遵循 .cursor/rules/go-main.mdc，注意此表在 saas 数据库中 | Success: Model 创建成功，字段映射正确，索引标签正确

- [x] 1.5 确认 CompanyStaff Go Model（使用现有模型）

  - File: `main/app/model/company.go`
  - Purpose: 确认现有的 CompanyStaff 模型符合需求
  - Requirements: 1.2
  - Leverage: 现有 Model: `main/app/model/company.go` 中的 CompanyStaff
  - Prompt: Role: Go Developer | Task: 确认 CompanyStaff 模型符合需求，添加与 SaasStaff 的关联 | Context: CompanyStaff.Uuid 关联到 SaasStaff.Uuid，CompanyStaff.CompanyUuid 关联到 Company.Uuid | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 确认完成，关联关系正确

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [x] 2.1 创建 SaasStaff Repository 接口

  - File: `main/app/repository/saas_staff.go`
  - Purpose: 定义 saas 数据库统一账号数据访问接口和实现
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Repository: `main/app/repository/staff.go`（接口和实现在同一文件）
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 创建 ISaasStaffRepo 接口，定义 CRUD 方法和唯一性检查方法 | Context: 使用选项模式(DBOption)，包含 Create, Update, GetByUuid, GetByEmail, GetByPhone, GetList, Delete, CheckEmailExists, CheckPhoneExists 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager，注意需要使用 saas 数据库连接 | Success: 接口定义完整，方法签名正确

- [x] 2.2 实现 SaasStaff Repository（选项模式）

  - File: `main/app/repository/saas_staff.go`
  - Purpose: 实现 saas 数据库统一账号数据访问逻辑
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Repository 实现: `main/app/repository/staff.go`、`main/app/repository/company_staff.go`（参考代码风格）
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 saasStaffRepo，使用选项模式实现灵活查询 | Context: 
    - 代码风格：接口 `ISaasStaffRepo`，实现 `saasStaffRepo`（小写开头），构造函数 `NewSaasStaffRepo` 和 `NewSaasStaffRepoImpl`
    - 只持有 db *gorm.DB（saas 数据库连接），实现所有接口方法
    - 特别注意 CheckPhoneExists 方法需要排除空字符串（只有非空手机号才验证唯一性）
    - 参考现有 Repository 的代码风格和选项模式实现
  - Restrictions: 不能持有 DBManager，使用 GORM，软删除(delete_time=0)，使用 saas 数据库连接，遵循现有代码风格 | Success: Repository 实现完整，选项模式正确，唯一性检查正确（手机号排除空字符串），代码风格符合现有规范

- [x] 2.3 创建 CompanyStaff Repository 接口

  - File: `main/app/repository/company_staff.go`
  - Purpose: 定义账号-门店关联数据访问接口（saas 数据库）
  - Requirements: 2.2
  - Leverage: 现有 Repository: `main/app/repository/company_staff.go`（接口和实现在同一文件）
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 扩展 ICompanyStaffRepo 接口，添加账号-门店关联的 CRUD 方法 | Context: 添加 GetByUuid, GetByStaffUuid, GetByCompanyUuid, GetByStaffAndCompany, Delete 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc，注意需要使用 saas 数据库连接 | Success: 接口定义完整

- [x] 2.4 扩展 CompanyStaff Repository（如需要）

  - File: `main/app/repository/company_staff.go`
  - Purpose: 扩展现有 CompanyStaff Repository，添加缺失的方法
  - Requirements: 2.2
  - Leverage: 现有 Repository: `main/app/repository/company_staff.go`
  - Prompt: Role: Go Developer with GORM expertise | Task: 扩展 CompanyStaffRepo，添加 GetByStaffUuid, GetByCompanyUuid, GetByStaffAndCompany, Delete 等方法 | Context: 参考现有代码风格，确保有 GetByStaffAndCompany 等方法（用于门店切换时验证权限） | Restrictions: 不能持有 DBManager，使用 GORM，软删除，使用 saas 数据库连接，遵循现有代码风格 | Success: Repository 方法完整，代码风格符合现有规范

- [ ] 2.5 编写 Repository 单元测试

  - File: `main/app/repository/saas_staff_test.go`, `main/app/repository/company_staff_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 1.1, 1.2, 2.2
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SaasStaffRepo 和 CompanyStaffRepo 编写单元测试，覆盖率 ≥ 80% | Context: 测试 CRUD 方法，测试唯一性检查方法（特别注意手机号空字符串的处理），测试软删除 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

### DTO 层

- [x] 2.6 创建 Request DTO

  - File: `main/app/dto/req/saas_staff_req.go`
  - Purpose: 定义 API 请求参数
  - Requirements: 1.1, 1.2, 2.2, 3.1
  - Leverage: 现有 DTO: `main/app/dto/req/auth.go`
  - Prompt: Role: Go Developer | Task: 创建 Request DTO，包含 SaasStaffCreateReq, SaasStaffUpdateReq, SaasStaffListReq, CompanyStaffBindReq, StoreSwitchReq | Context: 使用 binding 标签验证参数，邮箱需要 email 验证，手机号需要 max=20（允许空字符串），real_name 需要 max=255（可选） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [x] 2.7 创建 Response DTO

  - File: `main/app/dto/resp/saas_staff_resp.go`
  - Purpose: 定义 API 响应数据
  - Requirements: 1.1, 1.2, 2.2, 3.1
  - Leverage: 现有 DTO: `main/app/dto/resp/auth.go`
  - Prompt: Role: Go Developer | Task: 创建 Response DTO，包含 SaasStaffResp, SaasStaffListResp, CompanyStaffResp, StoreSwitchResp | Context: 包含 Meta 分页信息，注意字段映射（email、phone、real_name等） | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

### Service 层

- [x] 2.8 调整现有 Staff Service（AddStaff、UpdateStaff、PaginateGetStaffs 方法）

  - File: `main/app/service/staff.go`
  - Purpose: **调整现有方法**，使其支持统一账号体系和上级门店管理下级门店员工
  - Requirements: 1.1, 1.2, 1.3, 2.2
  - Leverage: 现有 Service: `main/app/service/staff.go` 中的 `AddStaff`、`UpdateStaff`、`PaginateGetStaffs` 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: **调整现有 Staff Service 方法**，使其支持统一账号体系和上级门店管理下级门店员工 | Context: 
    - **PaginateGetStaffs 方法调整**（员工列表查询）：
      - **上级门店**：需要查询本店及所有下级门店的员工
        - 通过 `companyRepo.GetAllSubShopsAndHeadquarterListByCompanyUuid(ctx.GetCompanyUuid())` 获取本店及下级门店UUID列表
        - 查询 `saas.ttpos_company_staff` 表（使用 saas 数据库连接），获取这些门店关联的所有员工UUID
        - 在门店数据库中查询这些员工的信息（需要跨数据库查询，可能需要遍历多个门店数据库）
        - 返回员工列表，包含员工所在门店信息
      - **子店**：只查询本店的员工（现有逻辑保持不变）
        - 查询当前门店数据库的 `ttpos_staff` 表
    - **AddStaff 方法调整**：
      - 验证邮箱和手机号唯一性时，需要查询 `saas.ttpos_staff` 表（使用 SaasStaffRepo.CheckEmailExists 和 CheckPhoneExists）
      - 创建员工时，需要同时创建 `saas.ttpos_staff` 记录（使用 SaasStaffRepo.Create）和 `saas.ttpos_company_staff` 记录
      - 设置 `last_company_uuid` 为当前门店UUID（ctx.GetCompanyUuid()）
      - 保持现有逻辑：在门店数据库中创建 Staff 记录和角色关联
    - **UpdateStaff 方法调整**：
      - 验证邮箱和手机号唯一性时，需要查询 `saas.ttpos_staff` 表（排除当前员工UUID）
      - 更新员工时，需要同时更新 `saas.ttpos_staff` 表（使用 SaasStaffRepo.Update）和 `saas.ttpos_company_staff` 表
      - 如果修改了邮箱或手机号，需要同步更新 `saas.ttpos_staff` 表
      - **上级门店**：可为员工配置本店及下级门店的角色
        - 需要扩展 `UpdateStaffReq`，添加 `company_role_list` 字段（数组，每个元素包含 `company_uuid` 和 `role_uuids`）
        - 更新 `saas.ttpos_company_staff` 表（添加或更新门店关联）
        - 更新各门店数据库的 `ttpos_staff_role` 表（更新角色关联）
      - **子店**：只能配置本店的角色（现有逻辑保持不变）
      - 保持现有逻辑：在门店数据库中更新 Staff 记录和角色关联
    - 参考现有代码风格：接口 `IStaffSrv`，实现 `staffSrv`，构造函数 `NewStaffSrv` 和 `NewStaffSrvImpl`
    - 权限判断：通过 `ctx.GetCompanySetting().IsHeadquarter()` 或 `ctx.GetCompanySetting().HeadquarterUuid` 判断是否为上级门店
  - Restrictions: 遵循 .cursor/rules/go-main.mdc，**不新增方法，只调整现有方法逻辑**，保持方法签名不变（UpdateStaffReq 需要扩展），特别注意手机号唯一性验证（排除空字符串），上级门店查询需要跨数据库查询 | Success: 现有方法调整完成，支持统一账号体系，上级门店可查看和管理下级门店员工，唯一性验证正确，事务管理正确

- [x] 2.9 创建 SaasStaff Service（仅用于门店切换等功能）

  - File: `main/app/service/saas_staff_srv.go`
  - Purpose: 创建新 Service，仅用于门店切换、获取默认门店等新管理端特有功能
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有 Service 实现: `main/app/service/staff.go`（参考代码风格）
  - Prompt: Role: Go Developer with business logic expertise | Task: 创建 ISaasStaffSrv 接口和 saasStaffSrv 实现，仅包含门店切换相关方法 | Context: 
    - 接口方法：UpdateLastCompany（更新上次登录商家UUID）、GetDefaultCompanyUuid（获取默认门店UUID，优先使用 last_company_uuid）
    - 代码风格：接口 `ISaasStaffSrv`，实现 `saasStaffSrv`（小写开头），构造函数 `NewSaasStaffSrv` 和 `NewSaasStaffSrvImpl`
    - 持有 `dbm *database.DBManager` 和 `cache cache.Cache`
    - 使用 saas 数据库连接（s.dbm.GetDB(constant.DefaultDB)）
  - Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 创建成功，代码风格正确，门店切换功能正确

- [x] 2.9.1 新增 SaasUpdateStaff 方法（统一账号修改员工）

  - File: `main/app/service/staff.go`
  - Purpose: 新增统一账号修改员工方法，支持多门店角色配置、IsDisable 更新和 RemoveCompanyList 处理
  - Requirements: 1.1, 1.2, 1.3, 2.2
  - Leverage: 现有方法: `main/app/service/staff.go` 中的 `UpdateStaff`、`SaasAddStaff` 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 IStaffSrv 接口中新增 SaasUpdateStaff 方法，实现统一账号修改员工功能 | Context: 
    - 参数和 UpdateStaff 一致（使用 `req.UpdateStaffReq`）
    - 查询 `saas.ttpos_staff` 是否存在该员工，不存在则报错
    - 修改 `saas.ttpos_staff` 的 Email、Phone、RealName（参考 UpdateStaff）
    - 支持密码更新（同步更新 `saas.ttpos_staff` 和门店数据库）
    - **总部/有子级商家**：
      - 验证 `CompanyRoleList` 不为空
      - 获取当前商家可见的所有门店列表
      - 遍历 `CompanyRoleList`，验证门店是否可见、角色是否存在、员工是否存在于门店数据库
      - 更新各门店数据库的 `ttpos_staff` 表和 `ttpos_staff_role` 关联关系
      - 更新 `saas.ttpos_company_staff` 表
      - 如果 `CompanyRoleList` 中存在当前商家uuid，根据 `IsDisable` 参数更新 `is_disable` 字段
    - **子店**：
      - 验证 `Roles` 不为空
      - 验证角色和员工是否存在
      - 更新当前商家数据库的 `ttpos_staff` 表和 `ttpos_staff_role` 关联关系
      - 更新 `saas.ttpos_company_staff` 表
      - 根据 `IsDisable` 参数更新 `is_disable` 字段
    - **IsDisable 字段更新**：
      - 如果是子店，或者 `CompanyRoleList` 中存在当前商家uuid，则根据参数中的 `IsDisable` 更新：
        - `saas.ttpos_company_staff` 的 `is_disable` 字段
        - 对应商家数据库中的 `ttpos_staff` 的 `is_disable` 字段
    - **RemoveCompanyList 处理**：
      - 如果 `RemoveCompanyList` 有值，遍历每个门店UUID：
        - 验证门店是否为当前公司可见
        - 软删除 `saas.ttpos_company_staff` 中的关联关系（设置 `delete_time`）
        - 软删除对应商家数据库中的 `ttpos_staff` 记录（设置 `delete_time`）
    - 删除收银机缓存并推送 WebSocket 配置更新通知
  - Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，支持事务管理 | Success: SaasUpdateStaff 方法实现成功，支持多门店配置、IsDisable 更新和 RemoveCompanyList 处理

- [x] 2.10 扩展 Auth Service（支持门店切换和默认门店选择）

  - File: `main/app/service/auth.go`
  - Purpose: 扩展登录流程，支持门店选择和多门店切换，新管理端登录时默认进入上次选择的门店
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有 Auth Service: `main/app/service/auth.go`，Task 2.9 的实现
  - Prompt: Role: Go Developer with authentication expertise | Task: 扩展 Auth Service，添加门店切换相关方法 | Context: 
    - 修改 Login 方法支持门店选择（新管理端 source=shop）：
      - 验证账号密码（使用邮箱或手机号）
      - 获取员工关联的门店列表
      - 如果只有一个门店，直接进入，返回 token 和门店信息
      - 如果有多个门店，调用 SaasStaffSrv.GetDefaultCompanyUuid 获取默认门店UUID（优先使用 last_company_uuid）
      - 返回门店列表和 default_company_uuid，前端显示门店选择界面
    - 添加 StoreSwitch 方法（切换成功后调用 SaasStaffSrv.UpdateLastCompany 更新 last_company_uuid）
    - 添加 GetCompanyList 方法
  - Restrictions: 遵循 .cursor/rules/go-main.mdc，保持向后兼容 | Success: Service 扩展成功，门店切换功能正确，登录时默认门店选择逻辑正确，last_company_uuid 更新逻辑正确

- [ ] 2.11 编写 Service 单元测试

  - File: `main/app/service/saas_staff_srv_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 1.1, 1.2, 2.2
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SaasStaffSrv 编写单元测试，覆盖率 ≥ 70% | Context: 测试业务逻辑，测试错误处理，测试事务管理，测试唯一性验证（特别注意手机号空字符串的处理） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### API 层

**重要说明**：
- **不需要新增 API 文件**：`main/app/api/v1/admin/` 目录下不需要新增文件。
- **不需要管理接口**：新管理端的员工管理功能（添加、编辑）使用现有的 `main/app/api/v1/shop/shop_staff.go` 中的 API 接口。
- **API 层无需修改**：现有的 `StaffHandler.AddStaff` 和 `StaffHandler.UpdateStaff` 方法会调用调整后的 Service 方法，自动支持统一账号体系。

- [x] 2.12 扩展 Auth API（支持新管理端登录和门店切换）

  - File: `main/app/api/v1/auth/auth_api.go`
  - Purpose: 扩展 Auth API，支持新管理端登录（返回门店列表和默认门店）、门店切换、获取账号关联的门店列表
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有 Auth API: `main/app/api/v1/auth/auth_api.go`，Task 2.10 的实现
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 扩展 Auth API，添加新管理端登录逻辑（支持返回门店列表和默认门店）、门店切换接口（/api/v1/auth/store_switch）和获取门店列表接口（/api/v1/auth/company_list） | Context: 新管理端登录需要根据 last_company_uuid 返回默认门店，门店切换需要更新 last_company_uuid 字段 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 扩展成功，登录逻辑正确，门店切换功能正常

- [x] 2.13 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 Auth API 的新路由（门店切换、获取门店列表）
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有路由: `main/router/router.go`
  - Prompt: Role: Go Developer | Task: 注册 Auth API 的新路由（门店切换、获取门店列表），**不需要注册员工管理相关的路由（使用现有的 shop API 路由）** | Restrictions: 遵循现有路由注册方式 | Success: 路由注册成功

- [ ] 2.14 编写 API 集成测试

  - File: `main/app/api/v1/auth/auth_api_test.go`
  - Purpose: 测试 Auth API 接口（新管理端登录、门店切换、获取门店列表）
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有测试: `main/app/api/v1/auth/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为扩展的 Auth API 编写集成测试 | Context: 测试新管理端登录（单门店和多门店场景）、门店切换、获取门店列表等 API 接口，测试参数验证，测试响应格式，测试 last_company_uuid 的更新逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: PHP Admin 模块（云平台端账号管理）

**说明**: 云平台端的账号管理功能（列表、创建、编辑、启用禁用）在 PHP Admin 模块实现，参考 `admin/app/admin/controller/admin/User.php` 的实现方式。

- [x] 3.1 创建 PHP Controller（云平台账号管理）

  - File: `admin/app/admin/controller/admin/Staff.php`
  - Purpose: 实现云平台端统一账号管理接口（列表、创建、编辑、启用禁用）
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 Controller: `admin/app/admin/controller/admin/User.php`（参考实现方式）
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 创建 Staff Controller，实现云平台端账号管理功能（列表、创建、编辑、启用禁用） | Context: 参考 admin/app/admin/controller/admin/User.php 的实现方式，API 路径为 /api/admin/admin.staff/index（列表）、/api/admin/admin.staff/add（创建）、/api/admin/admin.staff/edit（编辑）、/api/admin/admin.staff/updateStatus（启用禁用），Controller 不写业务逻辑，调用 Model 方法 | Restrictions: 遵循 .cursor/rules/php.mdc，使用验证器，遵循 MVC 分层 | Success: Controller 创建成功，接口路径正确，参考 User.php 的实现方式

- [x] 3.2 创建 PHP Model（统一账号）

  - File: `admin/app/admin/model/admin/Staff.php`
  - Purpose: 实现统一账号数据模型，操作 saas.ttpos_staff 表
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 Model: `admin/app/admin/model/admin/User.php`（参考实现方式）
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 创建 Staff Model，实现统一账号的 CRUD 方法 | Context: 参考 admin/app/admin/model/admin/User.php 的实现方式，指定数据库连接为 saas（protected $connection = 'saas'），实现 getList（列表查询，支持 keyword 搜索邮箱、手机号、员工ID，**需要关联查询 saas.ttpos_company_staff 获取门店信息，并查询门店数据库的 ttpos_staff_role 和 ttpos_role 获取角色信息，返回格式包含 company_list 字段，每个门店包含 company_uuid、company_name、roles 数组，返回字段包括 real_name**）、add（创建账号，验证邮箱和手机号唯一性，手机号排除空字符串，创建时绑定门店，支持 real_name 字段）、edit（编辑账号，验证邮箱和手机号唯一性，密码可选，支持 real_name 字段，**支持 company_list 参数设置关联门店和角色，需要更新 saas.ttpos_company_staff 表和门店数据库的 ttpos_staff_role 表**）、updateStatus（切换启用禁用状态）、detail（获取账号详情）方法，需要实现 getCompanyListWithRoles（获取账号关联的门店列表和角色）、getCompanyName（获取门店名称）、getStaffRoles（获取员工在门店下的角色列表）、updateCompanyList（更新账号关联的门店列表和角色设置）、updateStaffRoles（更新员工在门店下的角色设置）等辅助方法 | Restrictions: 遵循 .cursor/rules/php.mdc，使用事务管理，邮箱和手机号唯一性验证（手机号排除空字符串），软删除（delete_time=0），门店和角色信息需要跨数据库查询（saas 数据库和门店数据库） | Success: Model 创建成功，方法实现完整，唯一性验证正确，账号列表返回门店和角色信息（包括 real_name），编辑账号支持设置门店和角色（包括 real_name），参考 User.php 的实现方式

- [x] 3.3 创建验证器（统一账号）

  - File: `admin/app/admin/validate/StaffValidate.php`
  - Purpose: 参数验证（邮箱格式、唯一性，手机号唯一性，门店列表和角色）
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Validate: `admin/app/admin/validate/AdminUserValidate.php`（参考实现方式）
  - Prompt: Role: PHP Developer | Task: 创建 StaffValidate，验证邮箱格式和唯一性，验证手机号唯一性，验证门店列表和角色 | Context: 参考 admin/app/admin/validate/AdminUserValidate.php 的实现方式，邮箱需要 email 验证和全平台唯一性验证（checkEmailExist），手机号需要全平台唯一性验证（checkPhoneExist，排除空字符串），real_name 为可选字段（max:255），编辑场景下密码为可选（sceneEdit 方法移除 password 的 require），**company_list 为可选数组，每个元素包含 company_uuid（必填）和 role_uuids（可选数组）** | Restrictions: 遵循 .cursor/rules/php.mdc，邮箱和手机号唯一性验证（手机号排除空字符串） | Success: 验证器创建成功，唯一性验证正确，支持 company_list 验证，支持 real_name 验证，参考 AdminUserValidate.php 的实现方式

---

## Phase 4: Vue 前端模块

- [x] 4.1 创建 API 封装（云平台账号管理）

  - File: `admin/views/admin/src/api/user/staff.ts`
  - Purpose: 封装云平台端统一账号管理 API 调用（PHP Admin 模块）
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 API: `admin/views/admin/api/`（参考 adminUser.ts 的实现方式）
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装 Staff API 调用 | Context: API 路径为 /api/admin/admin.staff/index（列表）、/api/admin/admin.staff/add（创建）、/api/admin/admin.staff/edit（编辑）、/api/admin/admin.staff/updateStatus（启用禁用），使用 axios，定义 TypeScript 类型 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确，API 路径正确

- [x] 4.2 创建页面组件（云平台账号管理）

  - File: `admin/views/admin/src/pages/user/staff.vue`
  - Purpose: 实现云平台端统一账号管理页面 UI（列表、新增、编辑、启用禁用）
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有页面: `admin/views/admin/pages/admin/user/index.vue`（参考实现方式）
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建 Staff 页面，实现列表、新增、编辑、启用禁用功能 | Context: 参考 admin/views/admin/src/pages/user/admin.vue 的实现方式，使用 Element Plus，使用 Composition API，邮箱和手机号需要验证唯一性（手机号允许空字符串），**列表页面需要显示账号关联的门店和角色信息（company_list 字段），编辑页面需要支持设置关联门店和角色（company_list 参数，包含门店选择和角色多选）**，调用 PHP Admin 模块的 API（/api/admin/admin.staff.*） | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 页面创建成功，功能完整，列表显示门店和角色信息，编辑支持设置门店和角色，参考 User 页面的实现方式

- [x] 4.3 创建门店切换组件（各终端）

  - File: `admin/views/{admin|shop}/components/storeSwitch/index.vue`
  - Purpose: 实现门店切换 UI 组件
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有组件: `admin/views/{admin|shop}/components/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建门店切换组件，支持登录时门店选择和登录后门店切换 | Context: 使用 Element Plus，使用 Composition API | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 组件创建成功，功能完整

---

## Phase 5: 新管理端功能扩展

- [x] 5.1 扩展新管理端门店管理功能

  - File: `main/app/api/v1/shop/shop_setting.go`, `main/app/service/setting/setting.go`, `admin/views/shop/pages/storeManagement/`
  - Purpose: 实现门店管理功能（总部可查看和修改所有门店）
  - Requirements: 2.1
  - Leverage: 现有门店管理: `main/app/api/v1/shop/shop_setting.go`, `main/app/service/setting/setting.go::GetStoreSetting` 和 `EditStoreSetting`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 扩展门店管理功能，添加门店列表、查看门店信息、修改门店信息 API | Context: 
    - **API 1: 获取总部下所有门店列表** (`GET /api/v1/shop/company/list`)
      - 仅总部管理员可访问
      - 返回总部下所有门店列表，包含 `uuid`、`name`、`is_super`（是否有超管）、`is_headquarter`（是否为总部）
      - `is_super` 需要从 `saas.ttpos_company_staff` 表中查询 `is_super=1` 的记录
      - `is_headquarter` 从 `ttpos_company_setting.headquarter_uuid` 判断
    - **API 2: 获取门店信息** (`GET /api/v1/shop/company/info?company_uuid=xxx`)
      - 总部管理员可传入 `company_uuid` 查看任意门店信息
      - 分店管理员只能查看自己门店信息（忽略 `company_uuid` 参数）
      - 返回结构为 `setting.Store` 类型
      - 需要扩展 `settingSrv.GetStoreSetting` 方法支持指定 `company_uuid`
    - **API 3: 修改门店信息** (`POST /api/v1/shop/company/update`)
      - 总部管理员可传入 `company_uuid` 修改任意门店信息
      - 分店管理员只能修改自己门店信息（忽略 `company_uuid` 参数）
      - 可修改字段：商家名称、商家LOGO、地址、经纬度、公司名称、联系电话、税号、语言、时区
      - 参考现有的 `SaveStoreSetting` 接口实现
      - 需要扩展 `settingSrv.EditStoreSetting` 方法支持指定 `company_uuid`
    - 权限验证：需要验证是否为总部管理员（通过 `ctx.GetCompanySetting().IsHeadquarter()` 判断）
  - Restrictions: 遵循 .cursor/rules/go-main.mdc，遵循现有代码风格，参考 `SaveStoreSetting` 的实现方式 | Success: 门店管理功能扩展成功，API 接口正确，权限验证正确，总部可查看和修改所有门店

- [x] 5.2 扩展新管理端员工管理功能（支持账号导入）

  - File: `main/app/api/v1/shop/shop_staff.go`, `admin/views/shop/pages/staffManagement/`
  - Purpose: 实现员工账号导入功能，支持关联已有账号
  - Requirements: 2.2, 2.3
  - Leverage: 现有员工管理: `main/app/api/v1/shop/shop_staff.go`
  - Prompt: Role: Go Developer | Task: 扩展员工管理功能，添加账号导入逻辑 | Context: 在添加员工时，检查手机号/邮箱是否已存在，如果存在则提示导入关联 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 账号导入功能实现成功

- [x] 5.3 扩展新管理端员工管理功能（支持多门店角色配置和员工详情页）

  - File: `main/app/api/v1/shop/shop_staff.go`, `main/app/service/staff.go`, `admin/views/shop/pages/staffManagement/`
  - Purpose: 实现上级门店可配置员工在多个门店的角色，以及员工详情页
  - Requirements: 2.2
  - Leverage: 现有员工管理: `main/app/api/v1/shop/shop_staff.go`, `main/app/service/staff.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 扩展员工管理功能，添加员工详情页 API 和多门店角色配置 | Context: 
    - **新增 API: 获取员工详情** (`GET /shop/staff/detail?uuid=xxx`)
      - 返回员工基本信息及在所有门店的角色配置
      - **上级门店**：可查看员工在本店及下级门店的角色配置
        - 查询 `saas.ttpos_company_staff` 表，获取员工关联的所有门店
        - 查询各门店数据库的 `ttpos_staff_role` 和 `ttpos_role` 表，获取角色信息
        - 返回格式包含 `company_role_list` 字段，每个元素包含 `company_uuid`、`company_name`、`roles`（角色列表）
      - **子店**：只能查看员工在本店的角色配置
    - **扩展 UpdateStaffReq DTO**：
      - 添加 `company_role_list` 字段（可选数组）
      - 每个元素包含 `company_uuid`（uint64）和 `role_uuids`（[]uint64）
      - 用于上级门店为员工配置多个门店的角色
    - **UpdateStaff 方法已调整**（见任务2.8），这里只需要确保 API 层正确传递参数
  - Restrictions: 遵循 .cursor/rules/go-main.mdc，遵循现有代码风格 | Success: 员工详情页功能实现成功，多门店角色配置功能实现成功，上级门店可查看和配置员工在多个门店的角色

- [x] 5.4 扩展新管理端员工状态管理

  - File: `main/app/api/v1/shop/shop_staff.go`, `admin/views/shop/pages/staffManagement/`
  - Purpose: 实现员工状态管理（启用/禁用，仅对当前门店有效）
  - Requirements: 2.5
  - Leverage: 现有员工管理: `main/app/api/v1/shop/shop_staff.go`
  - Success: 员工状态管理功能实现成功

---

## Phase 6: 数据迁移

- [x] 6.1 创建数据迁移脚本（现有用户数据迁移）

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_migrate_existing_users_to_saas_staff.php`
  - Purpose: 将现有用户数据迁移到 saas 数据库的统一账号表
  - Requirements: 1.4
  - Leverage: 现有用户数据: saas 数据库的 `ttpos_company_staff` 表，门店数据库的 `ttpos_staff` 表
  - Prompt: Role: Database Engineer | Task: 创建数据迁移脚本，将现有用户数据迁移到 saas 数据库的 ttpos_staff 表 | Context: 数据来源包括：1) saas.ttpos_company_staff 表（id, uuid, username→email, phone, create_time, update_time, delete_time），2) 门店数据库的 ttpos_staff 表（real_name, password, password_change_count, password_change_time, is_disable），3) last_company_uuid 初始值来源于 ttpos_company_staff 中该员工关联的第一个门店的 company_uuid
    - 从 saas.ttpos_company_staff 迁移：id、uuid、email（来自username字段）、phone、create_time、update_time、delete_time
    - 从门店数据库的 ttpos_staff 表迁移：real_name、password、password_change_count、password_change_time、is_disable
    - last_company_uuid 初始数据来源于 saas.ttpos_company_staff 中的 company_uuid（取该员工关联的第一个门店，按 create_time 排序）
    - 需要遍历所有门店数据库，通过 uuid 关联获取密码等信息
    - 处理重复数据（同一个 uuid 可能在多个门店有记录，只取一条）
    - 处理邮箱重复的情况（如果 username 重复）
    - 处理手机号空字符串的情况
    - 处理密码为空的情况（设置默认值）
  - Restrictions: 遵循 .cursor/rules/database.mdc，注意数据一致性，注意手机号空字符串的处理，注意 email 来自 username 字段，注意 real_name 从门店数据库的 ttpos_staff 表迁移，注意 last_company_uuid 初始值来源于第一个门店 | Success: 迁移脚本创建成功，数据迁移正确，所有字段映射正确（包括 real_name）

- [ ] 6.2 执行数据迁移（需要手动执行：cd admin && php think migrate:run）

  - File: -
  - Purpose: 执行数据迁移脚本
  - Requirements: 1.4
  - Leverage: Task 6.1 的迁移脚本
  - Command: `cd admin && php think migrate:run`
  - Success: 数据迁移执行成功，现有用户数据已迁移

---

## Phase 7: 测试和优化

- [ ] 7.1 集成测试

  - File: `test/integration/saas_staff_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程（创建账号、绑定门店、门店切换），测试数据一致性，测试手机号空字符串的处理 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 7.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 7.3 缓存优化

  - File: `main/app/service/saas_staff_srv.go`
  - Purpose: 实现 Redis 缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现
  - Success: 缓存实现完成，命中率 > 80%

- [ ] 7.4 数据库查询优化

  - File: `main/app/repository/saas_staff.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 50ms

- [ ] 7.5 并发控制

  - File: `main/app/service/saas_staff_srv.go`
  - Purpose: 添加 UUID 锁防止并发创建重复账号
  - Requirements: 可靠性要求
  - Leverage: `pkg/lock/system_lock.go`
  - Success: 并发场景测试通过

- [ ] 7.6 文档更新

  - File: `docs/shared/api/saas_staff_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档（使用 SaasStaff 命名），数据库文档, CHANGELOG | Restrictions: 文档准确完整，统一使用 SaasStaff 命名 | Success: 所有文档已更新

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
  - 账号唯一性验证: 100%
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
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-auth-unified-account/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-auth-unified-account/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-auth-unified-account/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-auth-unified-account/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-auth-unified-account/tasks.md)" | bc
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
**最后更新**: 2025-12-10  
**维护者**: 后端开发组
