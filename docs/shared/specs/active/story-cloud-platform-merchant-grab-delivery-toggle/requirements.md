# 云平台-商家管理-Grab外卖控制 需求文档

> 本文档定义云平台商家管理中 Grab 外卖控制功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/cloud-platform-merchant-grab-delivery-toggle.md](../../../../team/proposals/2025-12/cloud-platform-merchant-grab-delivery-toggle.md) |
| **创建日期**      | 2025-12-08                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | 2025-12-08             |
| **审核意见** | 需求已通过审核，可进入技术设计阶段         |

---

## 📋 概述

在云平台的商家管理模块中，新增/编辑商家时增加 Grab 外卖功能的开启/关闭开关配置项。商家可以通过该配置控制是否启用 Grab 外卖服务，配置后系统会根据该状态控制 Grab 外卖相关的业务逻辑。

**核心功能**：
1. 在商家新增/编辑页面新增 Grab 外卖开启/关闭开关（默认关闭）
2. 关闭后新管理端不可见 Grab 外卖相关功能
3. 关闭后新管理端不可见对应外卖接单功能

**实现范围**：参考 `enable_kiosk` 的实现方式，在 `company_setting` 表中添加字段，并更新相关 API 和前端页面。

## 🎯 产品对齐

- 为商家提供 Grab 外卖服务的灵活控制能力
- 支持商家根据业务需求动态开启/关闭 Grab 外卖
- 提升云平台商家管理的完整性和便捷性
- 增强系统的第三方外卖平台集成管理能力

## 📝 用户故事

**作为** 商户管理员  
**我想** 在云平台新增/编辑商家时配置 Grab 外卖的开启/关闭状态  
**以便于** 根据业务需求灵活控制是否启用 Grab 外卖服务

**作为** 云平台运营人员  
**我想** 在商家管理中统一控制 Grab 外卖功能的启用状态  
**以便于** 根据商家实际情况灵活管理 Grab 外卖服务

---

## 功能需求

### Requirement 1: 商家管理 - Grab外卖开关

**用户故事**: 作为商户管理员，我想在新建/编辑商家时配置是否启用 Grab 外卖功能，以便于根据门店实际情况灵活控制 Grab 外卖服务的使用

#### 验收标准

1. **WHEN** 商户管理员在新建商家页面 **THEN** 系统 **SHALL** 提供【Grab外卖】开关参数（默认关闭）
2. **WHEN** 商户管理员在编辑商家页面 **THEN** 系统 **SHALL** 显示【Grab外卖】开关，并显示当前配置状态
3. **WHEN** 商户管理员修改【Grab外卖】开关状态并保存 **THEN** 系统 **SHALL** 成功保存配置到数据库
4. **IF** 未传递【Grab外卖】开关参数 **THEN** 系统 **SHALL** 使用默认值（0-关闭）

#### 具体要求

- [ ] 1.1 在 `company_setting` 表中添加 `enable_grab_delivery` 字段（INT(3)，默认值 0，注释：是否启用Grab外卖：0-否；1-是）
- [ ] 1.2 在商家新建接口 (`/api/admin/shop/add`) 中添加 `enable_grab_delivery` 参数（可选，默认 0）
- [ ] 1.3 在商家编辑接口 (`/api/admin/shop/edit`) 中添加 `enable_grab_delivery` 参数（可选，默认 0）
- [ ] 1.4 在商家列表查询接口中返回 `enable_grab_delivery` 字段
- [ ] 1.5 在 `AppValidate` 验证器中添加 `enable_grab_delivery` 验证规则（in:0,1）
- [ ] 1.6 在 `App` Model 中添加 `enable_grab_delivery` 字段定义
- [ ] 1.7 创建数据库迁移脚本，添加 `enable_grab_delivery` 字段

**参考实现**：`enable_kiosk` 字段的实现方式（`story-shop-kiosk-management`）

---

### Requirement 2: 新管理端 - Grab外卖功能可见性控制（暂不实现）

> **说明**: 根据项目安排，前端可见性控制功能暂不实现，由前端团队后续根据业务需要自行实现。后端已提供 `enable_grab_delivery` 字段，前端可根据该字段进行控制。

**用户故事**: 作为商户管理员，我想在 Grab 外卖关闭时，新管理端不显示 Grab 外卖相关功能，以便于避免误操作和界面混乱

#### 验收标准（待前端实现）

1. **IF** Grab外卖已关闭（`enable_grab_delivery = 0`） **THEN** 新管理端 **SHALL** 不显示 Grab 外卖相关菜单和页面
2. **IF** Grab外卖已关闭 **THEN** 新管理端外卖订单列表 **SHALL** 不显示 Grab 渠道的订单
3. **IF** Grab外卖已关闭 **THEN** 新管理端外卖接单功能 **SHALL** 不显示 Grab 渠道的接单入口
4. **IF** Grab外卖已开启（`enable_grab_delivery = 1`） **THEN** 新管理端 **SHALL** 正常显示 Grab 外卖相关功能

**涉及页面**：
- `admin/views/admin/src/pages/takeout/order.vue` - 外卖订单列表
- `admin/views/admin/src/pages/takeout/shop.vue` - 外卖商家管理
- `admin/views/admin/src/pages/takeout/setting.vue` - 外卖设置
- `admin/views/admin/src/router/routes.ts` - 路由配置
- `admin/views/admin/src/layouts/components/sidebar.vue` - 侧边栏菜单

**参考实现**：参考其他功能开关的可见性控制实现方式

**后端支持**：
- `/shop/base` 接口已返回 `enable_grab_delivery` 字段（Go Main 和 PHP Admin 均已支持）
- 前端可通过该字段判断是否显示 Grab 外卖相关功能

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/admin/shop/add`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] 金额字段使用 decimal(20,8)
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 前端功能测试覆盖所有页面

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

1. **商家管理 - Grab外卖开关**：
   - 新建商家时可以配置 Grab 外卖开关（默认关闭）
   - 编辑商家时可以修改 Grab 外卖开关状态
   - 配置状态正确保存到数据库

2. **新管理端 - Grab外卖功能可见性**：
   - Grab 外卖关闭时，新管理端不显示 Grab 外卖相关菜单
   - Grab 外卖关闭时，新管理端外卖订单列表不显示 Grab 渠道订单
   - Grab 外卖关闭时，新管理端不显示 Grab 外卖接单功能
   - Grab 外卖开启时，新管理端正常显示所有 Grab 外卖相关功能

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

- Grab 外卖开关默认关闭，确保不影响现有商家
- 关闭 Grab 外卖后，不影响已存在的 Grab 订单数据
- 前端可见性控制需要实时响应配置变更

### 资源约束

- 开发时间: 2-3 天
- Story Point: {SP 值} (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `admin/app/admin/controller/Shop.php` - 商家管理 Controller
- `admin/app/admin/validate/AppValidate.php` - 参数验证器
- `admin/app/admin/model/app/App.php` - 商家数据模型
- `admin/views/admin/src/pages/takeout/` - 新管理端外卖相关页面

### 服务依赖

- **Admin → Main**: HTTP API 调用（如需要）
- **Frontend → Admin**: HTTP API 调用

### 业务依赖

- 依赖商家管理模块的基础功能
- 依赖新管理端外卖模块的现有实现

---

## 风险和缓解

### 风险 1: 现有 Grab 订单数据受影响

**影响**: 中  
**概率**: 低  
**缓解措施**:
- 关闭 Grab 外卖仅影响新订单的接收，不影响已存在的订单数据
- 在关闭前给出明确提示

### 风险 2: 前端可见性控制不完整

**影响**: 中  
**概率**: 中  
**缓解措施**:
- 全面梳理新管理端所有 Grab 外卖相关页面和功能点
- 建立完整的测试用例覆盖所有可见性控制场景
- 代码审查时重点关注路由和菜单的权限控制

### 风险 3: 多商家环境下的配置混淆

**影响**: 低  
**概率**: 低  
**缓解措施**:
- 确保配置按商家维度存储和读取
- 前端根据当前登录用户的商家信息判断可见性

---

## 时间表

- **Phase 1 - 数据库和API层**: 1 天
  - 数据库迁移脚本
  - Model 和 DTO 更新
  - API 接口更新
- **Phase 2 - 前端功能实现**: 1-2 天
  - 商家管理页面开关添加
  - 新管理端可见性控制实现
  - 路由和菜单权限控制
- **Phase 3 - 测试和文档**: 0.5 天
  - 单元测试
  - 集成测试
  - 文档更新
- **总计**: 2.5-3.5 天（SP = {值}）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- `story-shop-kiosk-management` - 自助点餐机开关实现（参考 `enable_kiosk` 字段）
- `admin/database/migrations/20251205185229_add_enable_kiosk_to_company_setting.php` - 数据库迁移脚本参考
- `admin/app/admin/controller/Shop.php` - 商家管理 Controller
- `admin/views/admin/src/pages/takeout/` - 新管理端外卖相关页面

### 架构文档

- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: 产品组 + 开发团队  
**审核者**: {审核者}

