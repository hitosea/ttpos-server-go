> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 自助点餐机管理 需求文档

> 本文档定义自助点餐机管理功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/kiosk-management.md](../../../../team/proposals/2025-12/kiosk-management.md) |
| **创建日期**      | 2025-12-05                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | 2025-12-05                        |
| **审核意见** | 需求已通过审核，可进入技术设计阶段                        |

---

## 📋 概述

在云平台中增加自助点餐机的管理功能，包括：
1. 在商家管理模块中增加【自助点餐机】开关，允许商家在新建或编辑时控制是否启用自助点餐机功能（默认关闭）
2. 在客户端管理模块中增加【自助点餐机】版本管理功能，支持版本发布、更新和回滚

**实现范围**：仅实现后端 API，参考 `enable_data_management` 的实现方式。

## 🎯 产品对齐

- 提升云平台对自助点餐机的统一管理能力
- 为商家提供更灵活的自助点餐机配置选项
- 统一管理自助点餐机客户端版本，提升版本发布效率
- 支持按商家维度控制自助点餐机功能的启用/禁用

## 📝 用户故事

**作为** 商户管理员  
**我想** 在云平台中配置是否启用自助点餐机功能  
**以便于** 根据门店实际情况灵活控制自助点餐机的使用

**作为** 云平台运营人员  
**我想** 在客户端管理中统一管理自助点餐机的版本  
**以便于** 高效地发布、更新和回滚自助点餐机客户端版本

---

## 功能需求

### Requirement 1: 商家管理 - 自助点餐机开关

**用户故事**: 作为商户管理员，我想在新建/编辑商家时配置是否启用自助点餐机功能，以便于根据门店实际情况灵活控制自助点餐机的使用

#### 验收标准

1. **WHEN** 商户管理员在新建商家页面 **THEN** 系统 **SHALL** 提供【自助点餐机】开关参数（默认关闭）
2. **WHEN** 商户管理员在编辑商家页面 **THEN** 系统 **SHALL** 显示【自助点餐机】开关，并显示当前配置状态
3. **WHEN** 商户管理员修改【自助点餐机】开关状态并保存 **THEN** 系统 **SHALL** 成功保存配置到数据库
4. **IF** 未传递【自助点餐机】开关参数 **THEN** 系统 **SHALL** 使用默认值（0-关闭）

#### 具体要求

- [ ] 1.1 在 `company_setting` 表中添加 `enable_kiosk` 字段（INT(3)，默认值 0，注释：是否启用自助点餐机：0-否；1-是）
- [ ] 1.2 在商家新建接口 (`/api/admin/shop/add`) 中添加 `enable_kiosk` 参数（可选，默认 0）
- [ ] 1.3 在商家编辑接口 (`/api/admin/shop/edit`) 中添加 `enable_kiosk` 参数（可选，默认 0）
- [ ] 1.4 在商家列表查询接口中返回 `enable_kiosk` 字段
- [ ] 1.5 在 `AppValidate` 验证器中添加 `enable_kiosk` 验证规则（in:0,1）
- [ ] 1.6 在 `App` Model 中添加 `enable_kiosk` 字段定义
- [ ] 1.7 创建数据库迁移脚本，添加 `enable_kiosk` 字段

**参考实现**：`enable_data_management` 字段的实现方式

---

### Requirement 2: 客户端管理 - 自助点餐机版本管理

**用户故事**: 作为云平台运营人员，我想在客户端管理中统一管理自助点餐机的版本，以便于高效地发布、更新和回滚自助点餐机客户端版本

#### 验收标准

1. **WHEN** 运营人员在客户端管理模块选择自助点餐机类型 **THEN** 系统 **SHALL** 显示自助点餐机版本列表
2. **WHEN** 运营人员上传自助点餐机版本包 **THEN** 系统 **SHALL** 支持版本号解析和版本信息保存
3. **WHEN** 运营人员发布自助点餐机版本 **THEN** 系统 **SHALL** 支持版本发布、更新和回滚操作
4. **WHEN** 自助点餐机客户端查询新版本 **THEN** 系统 **SHALL** 返回最新发布的版本信息

#### 具体要求

- [ ] 2.1 在客户端版本管理接口 (`/api/admin/client.client/index`) 中支持自助点餐机类型（type=6，或根据现有类型编号规则确定）
- [ ] 2.2 在客户端版本添加接口 (`/api/admin/client.client/add`) 中支持自助点餐机版本包上传
- [ ] 2.3 在客户端版本发布接口 (`/api/admin/client.client/publish`) 中支持自助点餐机版本发布
- [ ] 2.4 在客户端版本查询接口 (`/api/admin/client.client/getNewVersion`) 中支持自助点餐机版本查询
- [ ] 2.5 更新客户端类型注释，添加自助点餐机类型说明

**参考实现**：现有客户端版本管理功能（收银端、平板端等）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/php.mdc` - PHP 开发规范
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

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/php.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `admin/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **商家管理 - 自助点餐机开关**: 
   - 新建商家时可设置 `enable_kiosk` 参数（默认 0）
   - 编辑商家时可修改 `enable_kiosk` 参数
   - 商家列表查询返回 `enable_kiosk` 字段
   - 参数验证正确（仅允许 0 或 1）

2. **客户端管理 - 自助点餐机版本管理**:
   - 支持自助点餐机版本列表查询
   - 支持自助点餐机版本包上传和版本信息解析
   - 支持自助点餐机版本发布、更新和回滚
   - 支持自助点餐机客户端查询新版本

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

### 业务约束

- 自助点餐机开关默认关闭（0）
- 客户端版本管理需与现有版本管理机制保持一致
- 历史商家数据默认 `enable_kiosk` 为 0

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5-8 SP (必须 ≤ 5，需拆分)

---

## 依赖关系

### 技术依赖

- `thinkphp/think-orm` - 数据库操作
- `thinkphp/think-validate` - 参数验证

### 服务依赖

- **Admin → Main**: HTTP API 调用（如需要）

### 业务依赖

- 商家管理模块（新建/编辑商家功能）
- 客户端版本管理模块（现有版本管理功能）

---

## 风险和缓解

### 风险 1: 商家配置开关后，需要确保自助点餐机端能正确读取配置

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在技术方案设计阶段明确配置同步机制
- 确保 Main 模块能正确读取 `enable_kiosk` 配置

### 风险 2: 版本管理功能需要与现有的客户端版本管理机制保持一致

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 参考现有客户端版本管理的实现方式，保持一致性
- 复用现有的版本管理逻辑

### 风险 3: 需要考虑历史数据的兼容性（已有商家默认状态处理）

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 对历史商家数据设置默认值（关闭状态）
- 数据库迁移脚本设置默认值

---

## 时间表

- **Phase 1 - 商家管理开关**: 2 天
- **Phase 2 - 客户端版本管理**: 2-3 天
- **总计**: 3-5 天（SP = 5-8，需拆分）

---

## 参考资料

### 核心规范

- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- `enable_data_management` 字段实现：
  - `admin/database/migrations/20251120013811_add_table_map_fields_to_company_setting.php`
  - `admin/app/admin/controller/Shop.php` (add/edit 接口)
  - `admin/app/admin/validate/AppValidate.php`
  - `admin/app/admin/model/app/App.php`

- 客户端版本管理实现：
  - `admin/app/admin/controller/client/Client.php`
  - `admin/app/common/model/client/ClientVersion.php`

### 架构文档

- `docs/human/architecture/php-architecture.md` - PHP 架构

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
**创建日期**: 2025-12-05  
**作者**: 王昱  
**审核者**: {审核者}

