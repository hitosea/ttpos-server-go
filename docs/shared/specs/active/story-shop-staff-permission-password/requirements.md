# 员工账号增加权限密码 需求文档

> 本文档定义员工账号增加权限密码的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                        |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/staff-permission-password.md](../../../../team/proposals/2025-11/staff-permission-password.md) |
| **创建日期**      | 2025-11-19                                                                                                                  |
| **负责人**        | 曾振华、何翔                                                                                                                |
| **目标 Sprint**   | Sprint {N}                                                                                                                  |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                                  |

---

## 📋 概述

在员工账号管理中新增"权限密码"字段，支持为每个员工设置权限密码。权限密码默认值为 666888，密码必须为 4-8 位数字，系统在设置密码时进行校验。**本功能需要在两个终端实现：1) 新管理端（Go 项目的 shop 目录：`main/app/api/v1/shop/`）；2) (旧)商家后台（PHP 项目的 admin/shop 目录：`admin/app/shop/`）。** 本功能仅包含员工账号的权限密码设置功能，不包含密码验证逻辑的实现。

## 🎯 产品对齐

该功能支持敏感操作权限验证需求：

- **支持权限密码验证**：为敏感操作验证功能提供基础数据支持
- **增强权限控制**：通过权限密码区分普通员工和管理人员
- **降低财务风险**：敏感操作需要授权员工密码验证
- **符合餐饮行业管理规范**：重要操作需要上级授权

## 📝 用户故事

**作为** 门店经理/店长  
**我想** 为员工账号设置权限密码  
**以便于** 在敏感操作（如折扣、退款）时进行权限验证，控制普通员工的操作权限

---

## 功能需求

### Requirement 1: 权限密码字段（新管理端 Go）

**用户故事**: 作为门店经理/店长，我想在新管理端为员工账号设置权限密码，以便于管理员工权限

#### 验收标准

1. **WHEN** 在新管理端添加员工账号 **THEN** 系统 **SHALL** 显示权限密码输入框，权限密码为必填项
2. **WHEN** 在新管理端编辑员工账号 **THEN** 系统 **SHALL** 显示权限密码输入框，但不回显原权限密码（显示为空或占位符），权限密码为非必填项
3. **WHEN** 设置权限密码 **THEN** 系统 **SHALL** 验证密码格式（4-8 位数字）
4. **IF** 密码不符合格式（不是 4-8 位数字） **THEN** 系统 **SHALL** 提示"密码必须为 4 - 8 位数字"
5. **IF** 新建员工时权限密码为空 **THEN** 系统 **SHALL** 提示"权限密码不能为空"
6. **WHEN** 保存新建员工信息 **THEN** 系统 **SHALL** 将权限密码加密后存储到数据库（权限密码必填）
7. **WHEN** 保存编辑员工信息 **THEN** 系统 **SHALL** 仅在设置了权限密码时更新，如未设置则不修改原权限密码

#### 具体要求

- [x] 1.1 在 `ttpos_staff` 表中新增 `permission_password` 字段（varchar(255) NOT NULL DEFAULT ''，加密存储）
- [x] 1.2 在 Go Model（`main/app/model/staff.go`）中添加权限密码字段
- [x] 1.3 在 DTO（`main/app/dto/req/staff.go`）中添加权限密码字段和验证规则（AddStaffReq 中必填，UpdateStaffReq 中非必填）
- [x] 1.4 在 Service（`main/app/service/staff.go`）中添加权限密码处理逻辑（加密存储）
- [x] 1.5 在 API（`main/app/api/v1/shop/shop_staff.go`）中支持权限密码字段的保存和查询（通过 DTO 自动支持）

---

### Requirement 2: 权限密码字段（商家后台 PHP）

**用户故事**: 作为门店经理/店长，我想在商家后台为员工账号设置权限密码，以便于管理员工权限

#### 验收标准

1. **WHEN** 在商家后台添加员工账号 **THEN** 系统 **SHALL** 显示权限密码输入框，权限密码为必填项
2. **WHEN** 在商家后台编辑员工账号 **THEN** 系统 **SHALL** 显示权限密码输入框，但不回显原权限密码（显示为空或占位符），权限密码为非必填项
3. **WHEN** 设置权限密码 **THEN** 系统 **SHALL** 验证密码格式（4-8 位数字）
4. **IF** 密码不符合格式（不是 4-8 位数字） **THEN** 系统 **SHALL** 提示"密码必须为 4 - 8 位数字"
5. **IF** 新建员工时权限密码为空 **THEN** 系统 **SHALL** 提示"权限密码不能为空"
6. **WHEN** 保存新建员工信息 **THEN** 系统 **SHALL** 将权限密码加密后存储到数据库（权限密码必填）
7. **WHEN** 保存编辑员工信息 **THEN** 系统 **SHALL** 仅在设置了权限密码时更新，如未设置则不修改原权限密码

#### 具体要求

- [x] 2.1 在 PHP Model（`admin/app/shop/model/auth/User.php`）中添加权限密码字段处理
- [x] 2.2 在验证器（`admin/extend/help/ValidateHelp.php`）中添加权限密码格式验证逻辑
- [x] 2.3 使用 `salt_hash()` 函数加密权限密码（与登录密码加密方式一致）
- [x] 2.4 在员工添加/编辑方法中处理权限密码字段

---

### Requirement 3: 密码格式验证

**用户故事**: 作为系统，我想验证权限密码格式，以便于确保密码符合要求

#### 验收标准

1. **WHEN** 用户输入权限密码 **THEN** 系统 **SHALL** 验证密码是否为 4-8 位数字
2. **IF** 密码不是纯数字 **THEN** 系统 **SHALL** 提示"密码必须为 4 - 8 位数字"
3. **IF** 密码长度小于 4 位或大于 8 位 **THEN** 系统 **SHALL** 提示"密码必须为 4 - 8 位数字"
4. **WHEN** 密码格式正确 **THEN** 系统 **SHALL** 允许保存

#### 具体要求

- [x] 3.1 前端验证：在输入框失去焦点时验证格式（商家后台已实现）
- [x] 3.2 后端验证：在保存前验证格式（Go 和 PHP 都已实现）
- [x] 3.3 验证规则：使用正则表达式 `/^\d{4,8}$/` 验证
- [x] 3.4 错误提示：统一提示"密码必须为 4 - 8 位数字"

---

### Requirement 4: 必填性和编辑处理

**用户故事**: 作为门店经理/店长，我想新建员工时必须设置权限密码，编辑员工时可以选择是否修改权限密码，以便于管理员工账号

#### 验收标准

1. **WHEN** 新建员工账号 **THEN** 系统 **SHALL** 要求权限密码为必填项
2. **IF** 新建员工时权限密码为空 **THEN** 系统 **SHALL** 提示"权限密码不能为空"并阻止保存
3. **WHEN** 编辑员工账号 **THEN** 系统 **SHALL** 不回显原权限密码（显示为空或占位符），权限密码为非必填项
4. **IF** 编辑员工时用户未设置权限密码（输入框为空） **THEN** 系统 **SHALL** 不修改原权限密码（保持原值）
5. **IF** 编辑员工时用户设置了权限密码 **THEN** 系统 **SHALL** 更新为新设置的权限密码

#### 具体要求

- [x] 4.1 前端页面：新建员工时权限密码输入框为必填项，显示必填标识（商家后台已实现）
- [x] 4.2 前端页面：编辑员工时权限密码输入框显示为空或占位符（不回显原密码），非必填项（商家后台已实现）
- [x] 4.3 前端验证：新建时权限密码必填验证，编辑时权限密码非必填（商家后台已实现）
- [x] 4.4 后端处理（新建）：权限密码必填，如果为空则返回错误（Go 和 PHP 都已实现）
- [x] 4.5 后端处理（编辑）：如果前端未传权限密码或为空，不更新权限密码字段（保持原值）（Go 和 PHP 都已实现）
- [x] 4.6 两个终端（新管理端和商家后台）都需要实现此逻辑（后端已实现，前端商家后台已实现）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层（Go）和 Controller → Model 分层（PHP）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Model 应独立且可复用
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名
- [ ] 响应格式：`{code, message, data{}}`
- [ ] data 字段必须是对象
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 在 `ttpos_staff` 表中新增 `permission_password` 字段
- [x] 字段类型：`varchar(255)`，用于存储加密后的密码
- [x] 字段约束：`NOT NULL DEFAULT ''`（不允许为空，默认值为空字符串）
- [x] 迁移文件：`admin/database/migrations/20251121014418_add_permission_password_field_to_staff_table.php`
- [x] Seeds 文件：已更新 `admin/database/seeds/shop_01.sql`
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 员工编辑页面加载时间 < 500ms
- [ ] 员工信息保存响应时间 < 200ms

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Go Service 层测试覆盖率 ≥ 70%
- [ ] PHP Controller 层测试覆盖核心逻辑
- [ ] 前端组件测试覆盖交互逻辑
- [ ] 集成测试覆盖员工信息保存和读取流程

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `admin/views/admin/src/locales/` - 国际化配置

### 安全要求

- [ ] 权限密码必须加密存储（使用与登录密码相同的加密方式）
- [ ] 权限密码不在 API 响应中返回
- [ ] 编辑页面不显示密码明文
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 员工信息保存失败时提供错误提示
- [ ] 密码格式验证失败时提供友好提示
- [ ] 错误日志记录

---

## 验收标准

### 功能验收

1. **新管理端权限密码设置**: 可以设置权限密码，格式验证正确，密码加密存储，新建时必填，编辑时非必填
2. **商家后台权限密码设置**: 可以设置权限密码，格式验证正确，密码加密存储，新建时必填，编辑时非必填
3. **密码格式验证**: 4-8 位数字验证正确，错误提示友好
4. **必填性处理**: 新建员工时权限密码必填，编辑时不回显原密码且非必填
5. **编辑逻辑**: 编辑员工时如未设置权限密码，则不修改原权限密码
6. **两个终端一致性**: 新管理端和商家后台的行为一致

### 测试验收

1. **功能测试**: 所有功能需求测试通过
2. **集成测试**: 员工信息保存和读取流程测试通过
3. **浏览器测试**: 主流浏览器兼容性测试通过

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
- Repository 只能持有 db 实例
- 不使用 panic，返回 error
- 密码加密使用 `utils.EncryptPassword()` 函数

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 密码加密使用 `salt_hash()` 函数

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 本功能仅包含员工账号的权限密码设置功能
- 密码验证逻辑将在其他任务中实现
- 权限密码与登录密码使用相同的加密方式
- 需要在两个终端（新管理端 Go 和商家后台 PHP）都实现，确保一致性

### 资源约束

- 开发时间: 1-2 天
- Story Point: 1-2 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- Go: Gin 框架、GORM、`utils.EncryptPassword()` 函数
- PHP: ThinkPHP 6.0、`salt_hash()` 函数
- Vue 3 + Element Plus

### 服务依赖

- **Frontend → Go API**: HTTP API 调用（新管理端）
- **Frontend → PHP API**: HTTP API 调用（商家后台）

### 业务依赖

- 依赖现有的员工管理功能
  - Go: `main/app/api/v1/shop/shop_staff.go`
  - PHP: `admin/app/shop/model/auth/User.php`
- 关联功能: 敏感操作设置（story-shop-sensitive-operation-settings）- 需要权限密码支持

---

## 风险和缓解

### 风险 1: 密码需要加密存储，需要确认加密方式

**影响**: 中  
**概率**: 高  
**缓解措施**:

- Go 项目使用 `utils.EncryptPassword()` 函数（MD5 双重加密）
- PHP 项目使用 `salt_hash()` 函数（与登录密码加密方式一致）
- 确保两个终端使用相同的加密算法

### 风险 2: 需要在两个终端（新管理端 Go 和商家后台 PHP）都实现，需要确保一致性

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 统一两个终端的实现方式，确保数据格式一致
- 统一密码加密方式（虽然函数名不同，但算法一致）
- 统一验证规则和错误提示

### 风险 3: 需要确认员工表结构，可能需要新增字段

**影响**: 低  
**概率**: 高  
**缓解措施**:

- 先查看现有 `ttpos_staff` 表结构，确定字段添加方式
- 参考现有的 `password` 字段实现方式

---

## 时间表

- **Phase 1 - 数据库迁移**: 0.5 天（创建迁移文件，添加 permission_password 字段）
- **Phase 2 - Go 新管理端实现**: 0.5 天（Model、DTO、Service、API）
- **Phase 3 - PHP 商家后台实现**: 0.5 天（Model、Controller）
- **Phase 4 - 前端实现**: 0.5-1 天（新管理端和商家后台的密码输入和校验）
- **Phase 5 - 测试和联调**: 0.5 天
- **总计**: 1-2 天（SP = 1-2）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南

### 类似功能参考

- **Go 新管理端**:
  - `main/app/api/v1/shop/shop_staff.go` - 员工管理 API
  - `main/app/service/staff.go` - 员工 Service（已有登录密码处理）
  - `main/app/model/staff.go` - 员工 Model（已有 password 字段）
- **PHP 商家后台**:
  - `admin/app/shop/model/auth/User.php` - 员工 Model（已有登录密码处理）
- **密码加密**:
  - `main/pkg/utils/encrypt.go` - Go 密码加密函数
  - PHP `salt_hash()` 函数（与登录密码加密方式一致）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**作者**: 产品组 + 开发组  
**审核者**: {审核者}
