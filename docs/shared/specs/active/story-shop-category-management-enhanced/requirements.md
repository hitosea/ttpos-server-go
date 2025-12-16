# 新管理端-分类管理-增强分类 需求文档

> 本文档定义新管理端分类管理增强功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.11.0-new-admin-category-management-enhanced.md](../../../../team/proposals/2025-12/v2.11.0-new-admin-category-management-enhanced.md) |
| **创建日期**      | 2025-12-09                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | 待确认             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

随着 Grab 外卖业务的接入，需要在分类管理中添加显示渠道控制功能。本功能仅在分类表增加两个字段，用于控制分类在店内和外卖平台的显示。

**核心价值**：
- 支持多渠道运营：可以灵活控制分类在店内和外卖平台的展示

## 🎯 产品对齐

本功能支持 TTPOS 的多渠道经营战略，帮助商户更好地管理不同销售渠道的商品分类，提升外卖业务的运营效率和用户体验。

---

## 📝 用户故事

**作为** 管理员  
**我想** 灵活控制分类在店内和外卖平台的显示  
**以便于** 根据不同渠道的经营策略优化分类展示

---

## 功能需求

### Requirement 1: 分类多渠道显示配置

**用户故事**: 作为管理员，我想独立配置分类在店内和外卖平台的显示状态，以便于根据不同渠道的经营策略灵活展示分类

#### 验收标准

1. **GIVEN** 我是一个管理员，已登录新管理端  
   **WHEN** 我在分类管理页面创建或编辑一个分类  
   **THEN** 系统应该显示"显示在店内"和"显示在 Grab 外卖"两个独立的开关

2. **GIVEN** 我是一个管理员，正在创建分类  
   **WHEN** 我尝试设置"显示在店内"为关闭（0）  
   **THEN** 系统应该拒绝此操作，提示"店内显示不允许取消"

3. **GIVEN** 我是一个管理员，正在编辑分类  
   **WHEN** 我尝试关闭"显示在店内"开关  
   **THEN** 系统应该拒绝此操作，提示"店内显示不允许取消"

4. **GIVEN** 我是一个管理员，正在编辑分类  
   **WHEN** 该分类已被 Grab 商品选中（takeout_product_count > 0），我尝试关闭"显示在 Grab 外卖"开关  
   **THEN** 系统应该拒绝此操作，提示"该分类已被外卖商品使用，不允许取消外卖显示"

5. **GIVEN** 我是一个管理员，正在查看分类详情  
   **WHEN** 我查看分类详情  
   **THEN** 系统应该返回该分类被外卖商品选中的数量，前端根据此数量判断是否可以控制 `is_display_in_takeout` 字段

6. **GIVEN** 商品管理中有外卖商品选择了某个分类  
   **WHEN** 外卖商品选中该分类时  
   **THEN** 系统应该自动将该分类的 `is_display_in_takeout` 设置为 1

#### 具体要求

- [ ] 1.1 数据库分类表增加 `is_display_in_store` 字段（tinyint，默认1）
- [ ] 1.2 数据库分类表增加 `is_display_in_takeout` 字段（tinyint，默认0）
- [ ] 1.3 分类创建接口支持设置这两个字段
- [ ] 1.4 分类编辑接口支持修改这两个字段
- [ ] 1.5 分类查询接口返回这两个字段
- [ ] 1.6 保存分类时，至少需要开启一个显示渠道（验证逻辑）
- [ ] 1.7 分类详情接口返回被外卖商品选中的数量（`takeout_product_count`），供前端判断是否可控制 `is_display_in_takeout`
- [ ] 1.8 店内显示不允许取消：`is_display_in_store` 必须始终为 1，创建和编辑时都不允许设置为 0
- [ ] 1.9 被 Grab 商品勾选的分类不允许取消外卖显示：如果 `takeout_product_count > 0`，则 `is_display_in_takeout` 不能设置为 0
- [ ] 1.10 自动外卖显示：当外卖商品选中某个分类时，系统自动将该分类的 `is_display_in_takeout` 设置为 1（商品创建/编辑时触发）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/shop/product/category/edit`）
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

**新增字段**：
- `is_display_in_store` (tinyint, 默认1) - 是否在店内显示
- `is_display_in_takeout` (tinyint, 默认0) - 是否在外卖平台显示
- `source_type` (tinyint, 默认2) - 来源类型：1-总部创建，2-分店创建
- `is_takeout_enabled` (tinyint, 默认0) - 分店对总部分类的外卖显示控制

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）- 分类列表缓存
- [ ] 并发处理（使用 UUID 锁）- Grab 同步时防止并发冲突

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖核心流程（分类编辑 → Grab 同步）
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
- [ ] 权限验证：分店只能操作自己店铺的分类
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级（Grab API 失败时不影响分类基本功能）
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制（Grab 同步失败重试）
- [ ] 同步状态持久化（防止服务重启丢失同步状态）

---

## 验收标准

### 功能验收

1. **分类多渠道显示配置**: 可以独立控制分类在店内和外卖平台的显示状态
2. **接口支持**: 分类创建、编辑、查询接口都支持新字段

### 测试验收

1. **单元测试**: Service 和 Repository 层测试覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过（分类编辑 → Grab 同步）
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
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

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 分类至少需要在一个渠道显示（店内或外卖）
- **店内显示不允许取消**：`is_display_in_store` 必须始终为 1，创建和编辑时都不允许设置为 0
- **被 Grab 商品勾选的分类不允许取消外卖显示**：如果分类被外卖商品选中（`takeout_product_count > 0`），则 `is_display_in_takeout` 不能设置为 0

### 资源约束

- 开发时间: 0.5-1 天
- Story Point: 1-2 SP (待技术评审确认)

---

## 依赖关系

### 数据依赖

- **数据依赖**: 分类表（ttpos_product_category）需要扩展字段

---

## 时间表

- **Phase 1 - 数据库迁移**: 0.2 天
- **Phase 2 - Model 和 DTO 更新**: 0.2 天
- **Phase 3 - API 接口修改**: 0.3 天
- **Phase 4 - 测试**: 0.3 天
- **总计**: 1 天（SP = 1-2，待技术评审确认）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- Grab API 文档（待补充）
- 前端仓库分支: `shop-grab-category-product-sync`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: weifashi  
**审核者**: 待确认

