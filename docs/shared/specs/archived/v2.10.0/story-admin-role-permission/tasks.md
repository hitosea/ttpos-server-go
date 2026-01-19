# 新管理端增加角色权限功能 任务分解

> 本文档定义「新管理端增加角色权限功能」的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求（如 R1.1, R2.3）

## 📊 进度总览

**总任务数**: 26（仅 Go 代码，新增国际化任务和测试）  
**已完成**: 19（Phase 0 + Phase 0.5 + Phase 1 + Phase 2 + Phase 3 + WebSocket推送 + 国际化）  
**进行中**: -  
**完成率**: 73%

> **注意**: 本 Spec 仅涉及 Go Main 模块开发，不涉及 PHP Admin 和 Vue 前端模块。

---

## Phase 0: ShopBase 接口返回权限（Requirement 0）

- [x] 0.1 ShopBase 响应结构确认

  - **Module**: Main - DTO 层
  - **File**: `main/app/dto/resp/base.go`
  - **Purpose**: 确认 `ShopBase` 结构已包含 `permissions` 字段
  - **Requirements**: R0.2
  - **Leverage**: 现有响应结构: `main/app/dto/resp/base.go`
  - **Status**: ✅ 已完成，`ShopBase` 结构已包含 `Permissions []*Permission` 字段，无需修改

- [x] 0.2 在 ShopBase 方法中获取权限

  - **Module**: Main - Service 层
  - **File**: `main/app/service/auth.go`
  - **Purpose**: 在 `ShopBase` 方法中调用权限服务获取员工权限
  - **Requirements**: R0.1, R0.4
  - **Leverage**: 现有权限服务: `main/app/service/role_access.go`，参考收银端权限获取逻辑
  - **Status**: ✅ 已完成，在 ShopBase 方法中调用 roleAccessSrv.GetPermission(constant.ShopAppRouteName, ...) 获取权限

- [x] 0.3 ShopBase API 返回权限

  - **Module**: Main - API 层
  - **File**: `main/app/api/v1/shop/shop_base.go`
  - **Purpose**: ShopBase 接口自动返回权限数据（通过 Service 层）
  - **Requirements**: R0.1-R0.5
  - **Leverage**: Task 0.2 的实现
  - **Status**: ✅ 已完成，ShopBase 接口通过 Service 层自动返回权限数据

---

## Phase 0.5: 获取权限树接口（Requirement 0.5）

- [x] 0.5.1 创建获取权限树 API 接口

  - **Module**: Main - API 层
  - **File**: `main/app/api/v1/shop/shop_staff.go`
  - **Purpose**: 创建 `/api/v1/shop/permission_tree` 接口，返回所有权限组的树形结构
  - **Requirements**: R0.5.1, R0.5.5
  - **Leverage**: 现有 Service 方法: `main/app/service/role_access.go` 的 `GetPermissionGroup` 方法
  - **Status**: ✅ 已完成，创建了 `GetPermissionTree` Handler 方法，注册路由 `/shop/permission_tree`

- [x] 0.5.2 调用权限服务获取权限组

  - **Module**: Main - Service 层 + API 层
  - **File**: `main/app/service/role_access.go`, `main/app/api/v1/shop/shop_staff.go`
  - **Purpose**: 在 Service 层新增 `GetCompanyPermissionTree` 方法（不依赖员工），在 Handler 中调用该方法获取店铺权限树
  - **Requirements**: R0.5.2
  - **Leverage**: Task 0.5.1 的实现
  - **Status**: ✅ 已完成，新增 `GetCompanyPermissionTree` 方法，在 `GetPermissionTree` Handler 中调用

- [x] 0.5.3 返回 PermissionGroup 格式

  - **Module**: Main - API 层
  - **File**: `main/app/api/v1/shop/shop_staff.go`
  - **Purpose**: 返回 `PermissionGroup` 结构，包含所有权限组的树形结构
  - **Requirements**: R0.5.3
  - **Leverage**: 现有响应结构: `main/app/dto/resp/base.go` 的 `PermissionGroup` 类型
  - **Status**: ✅ 已完成，返回 `resp.PermissionGroup` 格式

- [x] 0.5.4 权限数据经过筛选

  - **Module**: Main - Service 层
  - **File**: `main/app/service/role_access.go`
  - **Purpose**: `GetPermissionGroup` 方法已实现权限筛选逻辑（根据商户类型、ERP对接状态、渠道营收统计配置）
  - **Requirements**: R0.5.4
  - **Leverage**: 现有筛选逻辑: `main/app/service/role_access.go` 的 `filterPermission` 方法
  - **Status**: ✅ 已完成，Service 层已实现权限筛选逻辑

- [x] 0.5.5 接口需要身份验证

  - **Module**: Main - API 层
  - **File**: `main/app/api/v1/shop/shop_staff.go`
  - **Purpose**: 接口注册在 `middleware.Auth` 保护的路由组中，需要 JWT Token 验证
  - **Requirements**: R0.5.5
  - **Leverage**: 现有认证中间件: `middleware.Auth`
  - **Status**: ✅ 已完成，接口注册在需要认证的路由组中

- [x] 0.5.6 创建获取角色权限 API 接口

  - **Module**: Main - API 层
  - **File**: `main/app/api/v1/shop/shop_staff.go`
  - **Purpose**: 创建 `/api/v1/shop/role/permissions?role_uuid=xxx` 接口，返回角色的权限UUID列表
  - **Requirements**: R0.5.6
  - **Leverage**: 现有 Service 方法: `main/app/service/role_access.go` 的 `GetRolePermissions` 方法
  - **Status**: ✅ 已完成，创建了 `GetRolePermissions` Handler 方法，注册路由 `/shop/role/permissions`

- [x] 0.5.7 调用权限服务获取角色权限

  - **Module**: Main - Service 层 + API 层
  - **File**: `main/app/service/role_access.go`, `main/app/api/v1/shop/shop_staff.go`
  - **Purpose**: 在 Service 层新增 `GetRolePermissions` 方法，在 Handler 中调用该方法获取角色权限列表
  - **Requirements**: R0.5.7
  - **Leverage**: 现有 Repository: `main/app/repository/access.go` 的 `GetAccessUuids` 方法
  - **Status**: ✅ 已完成，新增 `GetRolePermissions` 方法，在 `GetRolePermissions` Handler 中调用

- [x] 0.5.8 权限名称国际化支持

  - **Module**: Main - Service 层
  - **File**: `main/app/service/role_access.go`
  - **Purpose**: 在获取权限树时，根据请求头的 Accept-Language 返回对应语言的权限名称
  - **Requirements**: R0.5.8
  - **Leverage**: 现有国际化配置: `main/i18n/`，`i18n.Translate` 方法
  - **Status**: ✅ 已完成，在 `GetCompanyPermissionGroup` 方法中遍历权限组，调用 `translatePermission(root, ctx.GetLanguage())` 方法进行国际化翻译。`translatePermission` 递归翻译所有子权限，使用 `i18n.Translate(language, permission.Name)` 实现翻译逻辑

---

## Phase 1: 数据库设计和迁移（如需要）

> 根据 design.md，所有表已存在，无需创建迁移文件。但需要确认表结构是否满足需求。

- [x] 1.1 确认数据库表结构

  - **Module**: Main - 数据库检查
  - **File**: `main/app/model/rbac.go`, `main/app/model/staff.go`
  - **Purpose**: 确认 `ttpos_role`、`ttpos_role_access`、`ttpos_staff_role` 表结构满足需求
  - **Requirements**: R1.1-R3.4
  - **Leverage**: 现有表结构，design.md 中的数据库设计
  - **Status**: ✅ 已完成，表结构已存在且符合需求，包含所有必需字段（id, uuid, create_time, update_time, delete_time）

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [x] 2.1 扩展 Role Repository（如需要）

  - **Module**: Main - Repository 层
  - **File**: `main/app/repository/role.go`
  - **Purpose**: 确认或扩展 Role Repository，支持角色 CRUD 和权限关联
  - **Requirements**: R1.1-R3.4
  - **Leverage**: 现有 Repository: `main/app/repository/role.go`，已存在 `UpdateRoleAccess` 方法
  - **Status**: ✅ 已完成，Role Repository 已包含所有必需方法（CreateRole, UpdateRole, DeleteRole, UpdateRoleAccess等），无需扩展

- [x] 2.2 扩展 StaffRole Repository（如需要）

  - **Module**: Main - Repository 层
  - **File**: `main/app/repository/staff_role.go`
  - **Purpose**: 确认或扩展 StaffRole Repository，支持员工角色关联查询
  - **Requirements**: R2.3, R3.1
  - **Leverage**: 现有 Repository: `main/app/repository/staff_role.go`
  - **Status**: ✅ 已完成，新增 `GetStaffUuidsByRoleUuid`、`DeleteStaffRolesByRoleUuid`、`CreateStaffRoles` 方法

### DTO 层

- [x] 2.3 创建 Request DTO

  - **Module**: Main - DTO 层
  - **File**: `main/app/dto/req/staff.go`
  - **Purpose**: 定义角色管理 API 请求参数
  - **Requirements**: R1.1-R3.4
  - **Leverage**: 现有 DTO: `main/app/dto/req/staff.go`（已存在 AddRoleReq, UpdateRoleReq, DeleteRoleReq, GetRoleReq）
  - **Status**: ✅ 已完成，完善了验证规则和错误消息（角色名称 1-50 字符，添加了 StaffUuids 字段用于编辑时关联员工）

- [x] 2.4 创建 Response DTO

  - **Module**: Main - DTO 层
  - **File**: `main/app/dto/resp/staff.go`
  - **Purpose**: 定义角色管理 API 响应数据
  - **Requirements**: R1.1-R3.4, R2.6-R2.7
  - **Leverage**: 现有 DTO: `main/app/dto/resp/staff.go`（已存在 RoleListResp, RoleDetailResp）
  - **Status**: ✅ 已完成，完善了 RoleDetailResp，添加了 StaffCount、StaffUuids、SelectedLeafCount、TotalLeafCount 字段

### Service 层

- [x] 2.5 创建或扩展 Role Service 接口

  - **Module**: Main - Service 层
  - **File**: `main/app/service/role.go`
  - **Purpose**: 定义角色管理业务逻辑接口
  - **Requirements**: R1.1-R3.4, R2.6-R2.7
  - **Leverage**: 现有 Service 接口模式
  - **Status**: ✅ 已完成，创建了 IRoleSrv 接口，包含 GetRoleList, GetRoleDetail, CreateRole, UpdateRole, DeleteRole, GetCompanyPermissionGroup 方法

- [x] 2.6 实现 Role Service 业务逻辑

  - **Module**: Main - Service 层
  - **File**: `main/app/service/role.go`
  - **Purpose**: 实现角色管理核心业务逻辑
  - **Requirements**: R1.1-R3.4, R2.6-R2.7
  - **Leverage**: 现有 Service: `main/app/service/role_access.go`（权限筛选逻辑），Task 2.1-2.4 的实现
  - **Status**: ✅ 已完成，实现了完整的角色 CRUD 业务逻辑，包括角色名称验证、角色关联员工检查、权限更新、员工关联管理等功能。在 GetRoleDetail 中实现了叶子节点统计功能（已选择叶子节点数量和管理APP、收银机、点餐助手三个权限组的叶子节点数量之和）

- [x] 2.6.1 在 UpdateRole 方法中添加 WebSocket 推送

  - **Module**: Main - Service 层
  - **File**: `main/app/service/role.go`
  - **Purpose**: 在更新角色权限成功后推送 WebSocket 通知到前端
  - **Requirements**: R2.8
  - **Leverage**: 现有 WebSocket 推送: `main/app/service/staff.go:172-176`，`main/pkg/websocket/websocket.go`
  - **Status**: ✅ 已完成，在 UpdateRole 方法中添加了 WebSocket 推送逻辑，使用 `websocket.PushClient` 异步推送 `UPDATE_PERMISSION` 消息，数据包含 `update_time` 和 `role_uuid`。仅在更新权限时推送，创建和删除角色不推送。

- [ ] 2.7 编写 Service 单元测试

  - **Module**: Main - Service 层测试
  - **File**: `main/app/service/role_srv_test.go`
  - **Purpose**: 确保 Service 业务逻辑正确
  - **Requirements**: R1.1-R3.4
  - **Leverage**: 现有测试: `main/app/service/*_srv_test.go`
  - **Prompt**: Role: QA Engineer with Go testing expertise | Task: 为 RoleSrv 编写单元测试，覆盖率 ≥ 70% | Context: 测试业务逻辑，测试错误处理，测试权限筛选，测试角色关联员工检查 | Restrictions: 遵循 `.cursor/rules/go-main.mdc` | Success: 测试覆盖率 ≥ 70%，所有测试通过

### API 层

- [x] 2.8 创建 Role API Controller

  - **Module**: Main - API 层
  - **File**: `main/app/api/v1/shop/shop_role.go`
  - **Purpose**: 实现角色管理 HTTP API 接口
  - **Requirements**: R1.1-R3.4
  - **Leverage**: 现有 API: `main/app/api/v1/shop/*_api.go`，Task 2.5-2.6 的 Service
  - **Status**: ✅ 已完成，创建了 RoleHandler，实现了 GetRoleDetail, CreateRole, UpdateRole, DeleteRole, GetPermissionGroup 接口，URL 使用 snake_case。DeleteRole 接口使用 DELETE 方法，参数通过请求体传递（`/api/v1/shop/role/delete`，Body: `{"uuid": xxx}`）。GetPermissionGroup 接口通过 `includeRouteNames` 参数指定返回"管理APP"、"收银机"、"点餐助手"三个权限组。注意：GetRoleList 已移除（角色列表功能在其他地方实现）

- [x] 2.9 注册 API 路由

  - **Module**: Main - 路由注册
  - **File**: `main/router/router.go`
  - **Purpose**: 注册角色管理 API 路由
  - **Requirements**: R1.1-R3.4
  - **Leverage**: 现有路由: `main/router/router.go`
  - **Status**: ✅ 已完成，在 router.go 中注册了 RegisterRoleHandlers，路由路径正确

- [ ] 2.10 编写 API 集成测试

  - **Module**: Main - API 层测试
  - **File**: `main/app/api/v1/shop/shop_role_test.go`
  - **Purpose**: 测试 API 接口
  - **Requirements**: R1.1-R3.4
  - **Leverage**: 现有测试: `main/app/api/v1/shop/*_api_test.go`
  - **Prompt**: Role: QA Engineer specializing in API testing | Task: 为 RoleAPI 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试权限筛选逻辑 | Restrictions: 遵循 `.cursor/rules/go-main.mdc` | Success: 所有 API 测试通过

---

## Phase 3: 权限规则处理

- [x] 3.1 实现权限动态显示/隐藏逻辑

  - **Module**: Main - Service 层 + Admin - Model 层
  - **File**: `main/app/service/role_access.go`, `admin/app/common/model/shop/Access.php`
  - **Purpose**: 根据商户配置动态筛选权限
  - **Requirements**: R4.1-R4.5
  - **Leverage**: 现有权限筛选逻辑: `main/app/service/role_access.go` 的 `filterPermission` 方法
  - **Status**: ✅ 已完成，`GetCompanyPermissionGroup` 方法中已调用 `filterPermission` 方法（传入 `companySetting` 和 `company` 参数），根据商户类型、ERP对接状态、授权配置、渠道营收统计配置动态筛选权限。筛选规则包括：商家后台隐藏管理APP权限（在 `admin/app/common/model/shop/Access.php:305` 中过滤 uuid=2856266502144000）、总部商户隐藏品采收货权限、已对接ERP隐藏进销存权限、授权配置动态隐藏相关权限等。权限组过滤通过 `includeRouteNames` 参数实现，只返回"管理APP"、"收银机"、"点餐助手"三个权限组（代码位置：`main/app/service/role_access.go:312-315`）

- [x] 3.2 实现管理APP默认勾选所有权限（前端处理）

  - **Module**: Main - Service 层（逻辑说明）
  - **File**: `main/app/service/role_access.go`
  - **Purpose**: 权限树返回时，前端可以根据权限组名称（"管理APP"）判断是否默认勾选所有子权限
  - **Requirements**: R1.4
  - **Leverage**: Task 3.1 的权限树获取方法
  - **Status**: ✅ 已完成，权限树正确返回"管理APP"分组及其所有子权限，前端可以根据分组名称实现默认勾选逻辑

---

## Phase 4: 测试和优化

- [ ] 4.1 集成测试

  - **File**: `test/integration/role_test.go`（如需要）
  - **Purpose**: 测试端到端功能
  - **Requirements**: 所有功能需求
  - **Leverage**: 现有集成测试
  - **Prompt**: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试创建角色 → 配置权限 → 关联员工 → 删除角色完整流程 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - **File**: -
  - **Purpose**: 确保性能达标
  - **Requirements**: 性能要求
  - **Leverage**: 性能测试工具（如：wrk, ab）
  - **Success**: 本地响应时间 < 200ms

- [ ] 4.3 权限筛选逻辑测试

  - **File**: `main/app/service/role_srv_test.go`
  - **Purpose**: 测试权限筛选逻辑在各种场景下的正确性
  - **Requirements**: R4.1-R4.4
  - **Leverage**: Task 2.7 的单元测试，Task 3.1 的权限筛选实现
  - **Prompt**: Role: QA Engineer | Task: 测试权限筛选逻辑 | Context: 测试总部商户隐藏品采收货权限（UUID: 2858548203520000），测试已对接ERP隐藏进销存权限（UUID: 2857919057920000），测试关闭渠道营收统计隐藏相关权限，测试授权配置动态隐藏权限 | Restrictions: 覆盖所有权限筛选场景 | Success: 权限筛选逻辑测试通过

- [ ] 4.3.1 权限名称国际化测试

  - **File**: `main/app/service/role_srv_test.go` 或 `main/app/api/v1/shop/shop_staff_test.go`
  - **Purpose**: 测试权限名称国际化功能是否正常工作
  - **Requirements**: R0.5.8
  - **Leverage**: Task 0.5.8 的国际化实现
  - **Prompt**: Role: QA Engineer | Task: 测试权限名称国际化 | Context: 测试请求头 Accept-Language=zh 返回中文权限名称，测试 Accept-Language=en 返回英文权限名称，测试不存在的语言回退到中文，测试未指定语言默认使用中文 | Restrictions: 覆盖所有国际化场景 | Success: 国际化测试通过

- [ ] 4.4 文档更新

  - **File**: `docs/shared/api/role_api.md`（如需要），`CHANGELOG.md`
  - **Purpose**: 确保文档与代码同步
  - **Requirements**: 文档要求
  - **Leverage**: `docs/agent/templates/api-doc-template.md`
  - **Prompt**: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

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
  - [ ] 添加角色功能正常
  - [ ] 编辑角色功能正常
  - [ ] 删除角色功能正常（关联员工时置灰）
  - [ ] 权限规则处理正确（动态显示/隐藏）
  - [ ] 管理APP默认勾选所有权限

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-role-permission/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-role-permission/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-role-permission/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-role-permission/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-role-permission/tasks.md)" | bc
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
- 在执行任务过程中若总结出角色权限管理、权限筛选逻辑等经验，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-24  
**维护者**: 后端开发组

