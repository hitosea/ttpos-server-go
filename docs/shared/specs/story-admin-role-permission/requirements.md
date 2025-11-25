# 新管理端增加角色权限功能 需求文档

> 本文档定义新管理端增加角色权限功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11-24-admin-role-permission.md](../../../team/proposals/2025-11-24-admin-role-permission.md) |
| **创建日期**      | 2025-11-24                                                                                                 |
| **负责人**        | 曾振华                                                                                                       |
| **目标 Sprint**   | Sprint v2.10.0                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |
| **开发范围**      | 仅 Go Main 模块，不涉及 PHP Admin 和 Vue 前端模块                                                              |

---

## 📋 概述

在新管理端增加角色权限管理功能，支持创建角色、配置功能权限、关联员工等操作。通过角色-权限-员工的关联模型，实现灵活的权限管理，提升系统安全性和管理效率。

## 🎯 产品对齐

提升系统安全性，通过权限控制降低误操作风险；支持精细化权限管理，满足不同岗位的权限需求；提高管理效率，便于商户进行员工权限配置；符合企业级应用的权限管理标准。

## 📝 用户故事

**作为** 商户管理员  
**我想** 创建角色并配置功能权限  
**以便于** 为不同岗位的员工分配合适的权限，提升系统安全性和管理效率

---

## 功能需求

### Requirement 0: ShopBase 接口返回权限

**用户故事**: 作为新管理端员工，我想在获取基础信息时获取我的权限信息，以便于前端根据权限显示菜单和功能

#### 验收标准

1. **WHEN** 员工调用 `/api/v1/shop/base` 接口 **THEN** 系统 **SHALL** 在响应中返回权限信息
2. **IF** 员工无任何权限 **THEN** 系统 **SHALL** 返回空权限数组
3. **IF** 员工是超级管理员 **THEN** 系统 **SHALL** 返回所有权限

#### 具体要求

- [x] 0.1 在 `ShopBase` 方法中调用权限服务获取员工权限
- [x] 0.2 `ShopBase` 响应结构已包含 `permissions` 字段（无需修改）
- [x] 0.3 权限数据格式为权限树结构（与 `GetPermission` 返回格式一致）
- [x] 0.4 使用 `constant.ShopAppRouteName`（"管理APP"）作为路由名称获取新管理端权限
- [x] 0.5 权限数据经过筛选（根据商户类型、ERP对接状态、渠道营收统计配置）

---

### Requirement 0.5: 获取权限树接口和角色权限接口

**用户故事**: 作为商户管理员，我想在新增或编辑角色权限前获取店铺的所有权限树和角色的权限列表，以便于配置角色的功能权限

#### 验收标准

1. **WHEN** 管理员调用权限树接口 **THEN** 系统 **SHALL** 返回店铺的权限树（管理APP、收银机、点餐助手三个权限组，不包含管理后台）
2. **WHEN** 管理员调用角色权限接口 **THEN** 系统 **SHALL** 返回该角色的权限UUID列表
3. **权限数据经过筛选**: 根据商户类型、ERP对接状态、渠道营收统计配置等动态筛选
4. **不依赖员工**: 权限树接口不依赖员工权限，返回店铺的所有可用权限
5. **过滤管理后台**: 权限树不返回"管理后台"分组，只返回"管理APP"、"收银机"、"点餐助手"三个分组

#### 具体要求

- [x] 0.5.1 创建获取权限树 API 接口 `/api/v1/shop/permission_tree`（在 shop_staff.go 中）
- [x] 0.5.1.1 创建获取权限组 API 接口 `/api/v1/shop/permission_group`（在 shop_role.go 中，用于角色权限配置）
- [x] 0.5.2 调用 `roleAccessSrv.GetCompanyPermissionGroup(companyUuid, includeRouteNames)` 获取店铺权限树（不依赖员工），传入 `includeRouteNames` 参数指定需要返回的权限组
- [x] 0.5.3 返回格式为 `PermissionGroup`，包含所有权限组的树形结构
- [x] 0.5.4 权限数据经过筛选（根据商户类型、ERP对接状态、渠道营收统计配置）
- [x] 0.5.5 接口需要身份验证（JWT Token）
- [x] 0.5.6 创建获取角色权限 API 接口 `/api/v1/shop/role/permissions?role_uuid=xxx`
- [x] 0.5.7 调用 `roleAccessSrv.GetRolePermissions(roleUuid, companyUuid)` 获取角色权限列表

**技术说明**:
- Service 层新增 `GetCompanyPermissionGroup` 方法（不依赖员工，只依赖店铺设置），接受 `includeRouteNames` 参数用于指定需要返回的权限组
- Service 层新增 `GetRolePermissions` 方法（获取角色的权限UUID列表）
- `/api/v1/shop/permission_tree` 接口在 `shop_staff.go` 中实现，返回不包含"管理后台"的权限树
- `/api/v1/shop/permission_group` 接口在 `shop_role.go` 中实现，用于角色权限配置，通过 `includeRouteNames` 参数指定返回"管理APP"、"收银机"、"点餐助手"三个权限组

**实现细节**:
- API 层调用时传入三个权限组常量：`constant.ShopAppRouteName`（"管理APP"）、`constant.CashierRouteName`（"收银机"）、`constant.AssistantRouteName`（"点餐助手"）
- Service 层在构建权限树后，遍历根节点，只保留名称匹配 `includeRouteNames` 的权限组
- 代码位置：`main/app/service/role_access.go:312-315`，通过 `slices.Contains` 判断权限组名称是否在允许列表中

---

### Requirement 1: 添加角色功能

**用户故事**: 作为商户管理员，我想添加角色并配置功能权限，以便于创建不同权限级别的角色

#### 验收标准

1. **WHEN** 管理员点击"添加角色" **THEN** 系统 **SHALL** 显示角色创建表单，包含角色名称和功能权限配置
2. **IF** 角色名称为空或超过50字符 **THEN** 系统 **SHALL** 提示错误信息
3. **IF** 选择管理APP权限 **THEN** 系统 **SHALL** 默认勾选所有权限
4. **IF** 总部商户 **THEN** 系统 **SHALL** 隐藏品采收货权限选项
5. **IF** 未对接ERP **THEN** 系统 **SHALL** 隐藏进销存权限选项
6. **IF** 关闭渠道营收统计 **THEN** 系统 **SHALL** 隐藏首页渠道营收统计的"更多"选项

#### 具体要求

- [ ] 1.1 在角色管理页面添加"添加角色"按钮
- [ ] 1.2 创建角色添加表单，包含角色名称输入框（必填，1-50字符）
- [ ] 1.3 添加功能权限配置区域，支持管理APP、收银机、点餐助手三个模块
- [ ] 1.4 管理APP权限默认勾选所有权限
- [ ] 1.5 根据商户类型和配置动态显示/隐藏权限选项：
  - 总部商户隐藏品采收货权限
  - 未对接ERP隐藏进销存权限
  - 关闭渠道营收统计时隐藏首页渠道营收统计的"更多"
- [ ] 1.6 添加角色时无关联员工功能（仅在编辑时可用）

---

### Requirement 2: 编辑角色功能

**用户故事**: 作为商户管理员，我想编辑角色信息并关联员工，以便于修改角色权限和分配员工

#### 验收标准

1. **WHEN** 管理员点击"编辑角色" **THEN** 系统 **SHALL** 显示角色编辑表单，包含角色名称、功能权限和关联员工
2. **WHEN** 编辑角色时 **THEN** 系统 **SHALL** 显示关联员工功能
3. **IF** 修改角色名称或权限 **THEN** 系统 **SHALL** 验证并保存更改
4. **WHEN** 更新角色权限成功 **THEN** 系统 **SHALL** 推送 WebSocket 通知到前端，通知权限已更新

#### 具体要求

- [ ] 2.1 在角色列表中添加"编辑"操作按钮
- [ ] 2.2 创建角色编辑表单，支持修改角色名称和功能权限
- [ ] 2.3 添加关联员工功能（仅在编辑时显示）
- [ ] 2.4 支持选择多个员工关联到角色
- [ ] 2.5 显示已关联的员工列表
- [x] 2.6 在获取角色详情时，返回已选择叶子节点权限数量（`selected_leaf_count`）
- [x] 2.7 在获取角色详情时，返回公司管理APP、收银机、点餐助手三个权限组的叶子节点数量之和（`total_leaf_count`）
- [x] 2.8 在 `UpdateRole` 方法中，更新角色权限成功后推送 WebSocket 通知
  - 使用 `websocket.PushClient()` 推送
  - 消息类型：`websocket.UPDATE_PERMISSION`
  - 推送范围：`SourceAll`（所有客户端）
  - 数据格式：`{"update_time": timestamp, "role_uuid": roleUuid}`
  - ⚠️ **注意**：仅在更新权限时推送，创建和删除角色不推送
  - ✅ **已完成**：在 `main/app/service/role.go` 的 `UpdateRole` 方法中实现了 WebSocket 推送逻辑

---

### Requirement 3: 删除角色功能

**用户故事**: 作为商户管理员，我想删除不需要的角色，以便于清理无效角色

#### 验收标准

1. **IF** 角色已关联员工 **THEN** 系统 **SHALL** 置灰删除按钮
2. **IF** 角色未关联员工且点击删除 **THEN** 系统 **SHALL** 显示确认提示："确定删除角色吗？"
3. **WHEN** 确认删除 **THEN** 系统 **SHALL** 删除角色及其权限关联

#### 具体要求

- [ ] 3.1 在角色列表中添加"删除"操作按钮
- [ ] 3.2 检查角色是否关联员工，如果已关联则置灰删除按钮
- [ ] 3.3 未关联员工时，点击删除显示确认提示："确定删除角色吗？"
- [ ] 3.4 确认删除后，删除角色记录和角色权限关联记录
- [x] 3.5 删除角色接口使用 DELETE 方法，参数通过请求体传递（`/api/v1/shop/role/delete`，Body: `{"uuid": xxx}`）

---

### Requirement 4: 权限规则处理

**用户故事**: 作为系统，我想根据商户配置动态显示权限选项，以便于提供准确的权限配置

#### 验收标准

1. **IF** 总部商户 **THEN** 系统 **SHALL** 隐藏品采收货权限选项
2. **IF** 未对接ERP **THEN** 系统 **SHALL** 隐藏进销存权限选项
3. **IF** 关闭渠道营收统计 **THEN** 系统 **SHALL** 隐藏首页渠道营收统计的"更多"选项

#### 具体要求

- [ ] 4.1 在权限配置界面，根据商户类型判断是否显示品采收货权限
- [ ] 4.2 在权限配置界面，根据ERP对接状态判断是否显示进销存权限
- [ ] 4.3 在权限配置界面，根据渠道营收统计配置判断是否显示相关权限
- [ ] 4.4 权限筛选逻辑与现有权限筛选逻辑保持一致（参考 `main/app/service/role_access.go`）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/role_info`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### WebSocket 推送要求

- [ ] **更新角色权限时推送通知**
  - 仅在更新角色权限时推送（`UpdateRole` 方法中）
  - 创建角色（`CreateRole`）不推送
  - 删除角色（`DeleteRole`）不推送
  - 使用 `websocket.PushClient()` 异步推送
  - 消息类型：`websocket.UPDATE_PERMISSION`
  - 推送范围：`SourceAll`（所有客户端）
  - 数据格式：`{"update_time": timestamp, "role_uuid": roleUuid}`
  - 参考实现：`main/app/service/staff.go:172-176`

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验）
- [ ] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **ShopBase 返回权限**: 新管理端员工调用 `/api/v1/shop/base` 接口时，响应中包含权限信息
2. **添加角色**: 管理员可以创建角色，配置功能权限，管理APP默认勾选所有权限
3. **编辑角色**: 管理员可以修改角色名称和权限，可以关联员工
4. **删除角色**: 已关联员工的角色不能删除，未关联的角色可以删除（需确认）
5. **权限规则**: 根据商户类型和配置动态显示/隐藏权限选项
6. **WebSocket 推送**: 更新角色权限成功后，推送 WebSocket 通知到前端（创建和删除不推送）

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%）
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 角色名称必填，1-50字符
- 管理APP权限默认勾选所有权限
- 添加角色时无关联员工功能
- 已关联员工的角色不能删除

### 资源约束

- 开发时间: 6-7 天（包含登录接口返回权限）
- Story Point: 5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos_role` 表 - 角色表（已存在）
- `ttpos_role_access` 表 - 角色权限关联表（已存在）
- `ttpos_access` 表 - 权限表（已存在）
- `ttpos_staff_role` 表 - 员工角色关联表（已存在）
- 权限筛选逻辑 - `main/app/service/role_access.go`

### 服务依赖

- **Admin → Main**: HTTP API 调用（角色管理、权限配置接口）

### 业务依赖

- 管理APP权限数据已初始化（`admin/database/migrations/20251124014502_init_management_app_access.php`）
- 权限筛选逻辑正常工作

---

## 风险和缓解

### 风险 1: 权限规则复杂，需要处理多种业务场景

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 详细梳理权限规则，建立权限配置矩阵
- 复用现有的权限筛选逻辑（`main/app/service/role_access.go`）
- 充分测试各种权限组合场景

### 风险 2: 权限变更可能影响现有用户的使用体验

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 提供权限预览功能，确保配置正确
- 权限变更时不影响已有角色权限
- 充分测试权限变更场景

### 风险 3: 需要确保权限验证逻辑的准确性

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 复用现有的权限验证逻辑
- 充分测试权限验证场景
- 代码审查确保逻辑正确

---

## 时间表

- **Phase 0 - 登录接口返回权限**: 0.5 天
- **Phase 1 - 数据库和模型**: 0.5 天（确认表结构）
- **Phase 2 - 核心实现（Go Main）**: 3-4 天
- **Phase 3 - 权限规则处理**: 1 天
- **Phase 4 - 测试和优化**: 1 天
- **总计**: 6-7 天（SP = 5）

> **注意**: 不包含 PHP Admin 和 Vue 前端模块开发时间。

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/features/role_access.md` - 角色权限服务架构
- `main/app/service/role_access.go` - 权限筛选逻辑实现
- `main/app/model/rbac.go` - 角色权限数据模型

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- DooTask #36929 - 产品需求文档
- `admin/database/migrations/20251124014502_init_management_app_access.php` - 管理APP权限初始化

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**作者**: 曾振华  
**审核者**: {审核者}

