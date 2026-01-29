# 云平台-商家管理-LINE MAN外卖控制 需求文档

> 本文档定义云平台商家管理中 LINE MAN 外卖控制功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/cloud-platform-merchant-lineman-delivery-toggle.md](../../../../team/proposals/2026-01/cloud-platform-merchant-lineman-delivery-toggle.md) |
| **创建日期**      | 2026-01-12                                                                                                 |
| **负责人**        | 曾振华                                                                                                       |
| **目标 Sprint**   | Sprint 2026-01                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | 曾振华             |
| **审核日期** | 2026-01-12             |
| **审核意见** | 需求已通过审核，可进入技术设计阶段         |

---

## 📋 概述

在云平台的商家管理模块中，新增/编辑商家时增加 LINE MAN 外卖功能的开启/关闭开关配置项。商家可以通过该配置控制是否启用 LINE MAN 外卖服务，配置后系统会根据该状态控制 LINE MAN 外卖相关的业务逻辑。

**核心功能**：
1. 在商家新增/编辑页面新增 LINE MAN 外卖开启/关闭开关（默认关闭）
2. 关闭后新管理端不可见 LINE MAN 外卖相关功能
3. 关闭后新管理端不可见对应外卖接单功能
4. 与 Grab 外卖开关相互独立，互不影响

**实现范围**：参考 `enable_grab_delivery` 的实现方式，在 `company_setting` 表中添加字段，并更新相关 API 和前端页面。

## 🎯 产品对齐

- 为商家提供 LINE MAN 外卖服务的灵活控制能力
- 支持商家根据业务需求动态开启/关闭 LINE MAN 外卖
- 提升云平台商家管理的完整性和便捷性
- 增强系统的第三方外卖平台集成管理能力
- 支持多外卖平台并存，商家可独立控制各平台

## 📝 用户故事

**作为** 商户管理员  
**我想** 在云平台新增/编辑商家时配置 LINE MAN 外卖的开启/关闭状态  
**以便于** 根据业务需求灵活控制是否启用 LINE MAN 外卖服务

**作为** 云平台运营人员  
**我想** 在商家管理中统一控制 LINE MAN 外卖功能的启用状态  
**以便于** 根据商家实际情况灵活管理 LINE MAN 外卖服务

---

## 功能需求

### Requirement 1: 商家管理 - LINE MAN外卖开关

**用户故事**: 作为商户管理员，我想在新建/编辑商家时配置是否启用 LINE MAN 外卖功能，以便于根据门店实际情况灵活控制 LINE MAN 外卖服务的使用

#### 验收标准

1. **WHEN** 商户管理员在新建商家页面 **THEN** 系统 **SHALL** 提供【LINE MAN外卖】开关参数（默认关闭）
2. **WHEN** 商户管理员在编辑商家页面 **THEN** 系统 **SHALL** 显示【LINE MAN外卖】开关，并显示当前配置状态
3. **WHEN** 商户管理员修改【LINE MAN外卖】开关状态并保存 **THEN** 系统 **SHALL** 成功保存配置到数据库
4. **IF** 未传递【LINE MAN外卖】开关参数 **THEN** 系统 **SHALL** 使用默认值（0-关闭）

#### 具体要求

- [x] 1.1 在 `company_setting` 表中添加 `enable_lineman_delivery` 字段（INT(3)，默认值 0，注释：是否启用LINE MAN外卖：0-否；1-是）
- [x] 1.2 在商家新建接口 (`/api/admin/shop/add`) 中添加 `enable_lineman_delivery` 参数（可选，默认 0）
- [x] 1.3 在商家编辑接口 (`/api/admin/shop/edit`) 中添加 `enable_lineman_delivery` 参数（可选，默认 0）
- [x] 1.4 在商家列表查询接口中返回 `enable_lineman_delivery` 字段
- [x] 1.5 在 `AppValidate` 验证器中添加 `enable_lineman_delivery` 验证规则（in:0,1）
- [x] 1.6 在 `App` Model 中添加 `enable_lineman_delivery` 字段定义
- [x] 1.7 创建数据库迁移脚本，添加 `enable_lineman_delivery` 字段

**参考实现**：`enable_grab_delivery` 字段的实现方式（`story-cloud-platform-merchant-grab-delivery-toggle`）

---

### Requirement 2: Go Main 模块 - LINE MAN外卖状态支持

**用户故事**: 作为系统开发者，我想在 Go Main 模块中支持 LINE MAN 外卖状态，以便于前端和业务逻辑能够获取和使用该配置

#### 验收标准

1. **WHEN** 前端调用 `/shop/base` 接口 **THEN** 响应 **SHALL** 包含 `is_open_lineman_delivery` 字段
2. **WHEN** POS 端调用基础信息接口 **THEN** 响应 **SHALL** 包含 LINE MAN 外卖状态
3. **IF** LINE MAN 外卖已关闭 **THEN** 商品外卖类型过滤 **SHALL** 不包含 LINE MAN 类型
4. **IF** LINE MAN 外卖已关闭 **THEN** 新管理端权限列表 **SHALL** 不包含 LINE MAN 外卖相关权限

#### 具体要求

- [x] 2.1 在 `CompanySetting` Model 中添加 `EnableLinemanDelivery` 字段
- [x] 2.2 在 `CompanySetting` Model 中添加 `IsOpenLINEMANDelivery()` 方法
- [x] 2.3 在 `BaseInfo` DTO 中添加 `IsOpenLINEMANDelivery` 字段
- [x] 2.4 在 `/shop/base` 接口中返回 `is_open_lineman_delivery` 字段
- [x] 2.5 在商品服务的外卖类型过滤中支持 LINE MAN 外卖状态
- [x] 2.6 在权限过滤服务中支持 LINE MAN 外卖权限控制

**参考实现**：
- `main/app/model/company.go` - Grab 外卖字段定义
- `main/app/service/product.go` - Grab 外卖类型过滤
- `main/app/service/role_access.go` - Grab 外卖权限过滤

---

### Requirement 3: PHP Admin 模块 - LINE MAN外卖状态支持

**用户故事**: 作为系统开发者，我想在 PHP Admin 模块中支持 LINE MAN 外卖状态，以便于商家端能够获取和使用该配置

#### 验收标准

1. **WHEN** 商家端调用 `/shop/base` 接口 **THEN** 响应 **SHALL** 包含 `enable_lineman_delivery` 字段
2. **WHEN** 商家端调用授权信息接口 **THEN** 响应 **SHALL** 包含 LINE MAN 外卖状态
3. **IF** LINE MAN 外卖已关闭 **THEN** 商家端权限过滤 **SHALL** 隐藏 LINE MAN 外卖相关权限

#### 具体要求

- [x] 3.1 在 `Controller.php` 的查询字段中添加 `enable_lineman_delivery`
- [x] 3.2 在 `App.php` Model 的 `getLicense()` 方法中返回 `enable_lineman_delivery` 字段
- [x] 3.3 在 `Access.php` Model 的权限过滤中支持 LINE MAN 外卖权限控制

**参考实现**：
- `admin/app/shop/controller/Controller.php` - Grab 外卖字段查询
- `admin/app/common/model/app/App.php` - Grab 外卖授权信息
- `admin/app/common/model/shop/Access.php` - Grab 外卖权限过滤

---

### Requirement 4: 前端 - LINE MAN外卖开关显示

**用户故事**: 作为商户管理员，我想在商家管理页面看到 LINE MAN 外卖开关，以便于配置该功能

#### 验收标准

1. **WHEN** 打开新建商家页面 **THEN** 页面 **SHALL** 显示【LINE MAN外卖】开关（默认关闭）
2. **WHEN** 打开编辑商家页面 **THEN** 页面 **SHALL** 显示【LINE MAN外卖】开关，并显示当前状态
3. **WHEN** 切换【LINE MAN外卖】开关并保存 **THEN** 系统 **SHALL** 成功保存配置
4. **开关 UI** **SHALL** 位于【Grab外卖】开关下方，保持一致的样式

#### 具体要求

- [x] 4.1 在商家编辑表单中添加 `enable_lineman_delivery` 表单项
- [x] 4.2 在表单验证规则中添加 `enable_lineman_delivery` 验证
- [x] 4.3 在表单默认值中添加 `enable_lineman_delivery: 0`
- [x] 4.4 在表单数据绑定中支持 `enable_lineman_delivery` 字段
- [x] 4.5 在 TypeScript 接口定义中添加 `enable_lineman_delivery` 字段

**参考实现**：
- `admin/views/admin/src/pages/merchant/components/dialog-edit.vue` - Grab 外卖开关
- `admin/views/admin/src/api/merchant/index.ts` - Grab 外卖类型定义

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/admin/shop/add`）
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

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用索引）
- [x] 缓存策略（Redis）
- [x] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 前端功能测试覆盖所有页面

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 故障恢复机制

---

## 验收标准

### 功能验收

1. **商家管理 - LINE MAN外卖开关**：
   - [x] 新建商家时可以配置 LINE MAN 外卖开关（默认关闭）
   - [x] 编辑商家时可以修改 LINE MAN 外卖开关状态
   - [x] 配置状态正确保存到数据库

2. **Go Main 模块 - LINE MAN外卖状态支持**：
   - [x] `/shop/base` 接口返回 `is_open_lineman_delivery` 字段
   - [x] 商品外卖类型过滤支持 LINE MAN 外卖状态
   - [x] 权限列表支持 LINE MAN 外卖权限控制

3. **PHP Admin 模块 - LINE MAN外卖状态支持**：
   - [x] 商家端 `/shop/base` 接口返回 `enable_lineman_delivery` 字段
   - [x] 授权信息接口返回 LINE MAN 外卖状态
   - [x] 权限过滤支持 LINE MAN 外卖权限控制

4. **前端 - LINE MAN外卖开关显示**：
   - [x] 商家管理页面显示 LINE MAN 外卖开关
   - [x] 开关状态能够正确保存和显示

### 测试验收

1. **单元测试**: 覆盖率达标
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

- 必须使用 Go 1.23+
- 遵循 Gin 框架规范
- 使用 GORM ORM
- 遵循 `.cursor/rules/go-main.mdc`

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- LINE MAN 外卖开关默认关闭，确保不影响现有商家
- 关闭 LINE MAN 外卖后，不影响已存在的 LINE MAN 订单数据
- LINE MAN 外卖开关与 Grab 外卖开关相互独立
- 前端可见性控制需要实时响应配置变更

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 (已完成)

---

## 依赖关系

### 技术依赖

**PHP Admin 模块**:
- `admin/app/admin/controller/Shop.php` - 商家管理 Controller
- `admin/app/admin/validate/AppValidate.php` - 参数验证器
- `admin/app/admin/model/app/App.php` - 商家数据模型
- `admin/app/shop/controller/Controller.php` - 商家端 Controller
- `admin/app/common/model/app/App.php` - 授权信息 Model
- `admin/app/common/model/shop/Access.php` - 权限过滤 Model

**Go Main 模块**:
- `main/app/model/company.go` - 公司设置模型
- `main/app/dto/resp/base.go` - 基础信息 DTO
- `main/app/service/auth.go` - 认证服务
- `main/app/service/product.go` - 商品服务
- `main/app/service/role_access.go` - 权限服务
- `main/app/api/v1/menu/menu_handler.go` - 菜单处理器
- `main/app/service/h5_service.go` - H5 服务

**Vue 前端模块**:
- `admin/views/admin/src/pages/merchant/components/dialog-edit.vue` - 商家编辑页面
- `admin/views/admin/src/api/merchant/index.ts` - 商家 API 定义

### 服务依赖

- **Admin → Main**: HTTP API 调用（如需要）
- **Frontend → Admin**: HTTP API 调用

### 业务依赖

- 依赖商家管理模块的基础功能
- 依赖新管理端外卖模块的现有实现

---

## 风险和缓解

### 风险 1: 现有 LINE MAN 订单数据受影响

**影响**: 中  
**概率**: 低  
**缓解措施**:
- 关闭 LINE MAN 外卖仅影响新订单的接收，不影响已存在的订单数据
- 在关闭前给出明确提示

### 风险 2: 与 Grab 外卖开关产生冲突

**影响**: 低  
**概率**: 低  
**缓解措施**:
- LINE MAN 外卖开关与 Grab 外卖开关完全独立
- 权限过滤使用 OR 逻辑，任一开关开启即显示外卖接单权限
- 外卖类型过滤分别判断

### 风险 3: 多商家环境下的配置混淆

**影响**: 低  
**概率**: 低  
**缓解措施**:
- 确保配置按商家维度存储和读取
- 前端根据当前登录用户的商家信息判断可见性

---

## 时间表

- **Phase 1 - 数据库和API层**: 1 天（已完成）
  - 数据库迁移脚本 ✅
  - Model 和 DTO 更新 ✅
  - API 接口更新 ✅
- **Phase 2 - Go Main 模块**: 0.5 天（已完成）
  - Model 字段定义 ✅
  - Service 逻辑更新 ✅
  - DTO 响应更新 ✅
- **Phase 3 - PHP Admin 模块**: 0.5 天（已完成）
  - Controller 更新 ✅
  - Model 更新 ✅
  - 权限过滤更新 ✅
- **Phase 4 - 前端实现**: 0.5 天（已完成）
  - 表单添加 ✅
  - 验证规则 ✅
  - API 调用 ✅
- **Phase 5 - 测试和文档**: 0.5 天（待完成）
  - 单元测试
  - 集成测试
  - 文档更新
- **总计**: 3 天（功能已完成，测试待补充）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- `story-cloud-platform-merchant-grab-delivery-toggle` - Grab 外卖开关实现
- `admin/database/migrations/20251208191025_add_enable_grab_delivery_to_company_setting.php` - Grab 外卖迁移脚本
- `admin/database/migrations/20260112103704_add_enable_lineman_delivery_to_company_setting.php` - LINE MAN 外卖迁移脚本
- `admin/app/admin/controller/Shop.php` - 商家管理 Controller
- `main/app/model/company.go` - 公司设置模型

### 架构文档

- `docs/human/architecture/go-architecture.md` - Go 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-development.md` - Go 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/曾振华/2026-01/2026-01-12.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: 曾振华  
**审核者**: 曾振华

