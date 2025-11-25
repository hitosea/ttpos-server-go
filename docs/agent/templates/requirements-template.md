# {功能名称} 需求文档

> 本文档定义 {功能} 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/{YYYY-MM}/{feature-name}.md](../../../../team/proposals/{YYYY-MM}/{feature-name}.md) |
| **创建日期**      | {YYYY-MM-DD}                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

---

## 📋 概述

[简要说明该功能的目的、价值以及对用户的意义]

## 🎯 产品对齐

[说明该功能如何支持产品愿景和业务目标]

## 📝 用户故事

**作为** [角色]  
**我想** [功能]  
**以便于** [价值/目的]

---

## 功能需求

### Requirement 1: {需求标题}

**用户故事**: 作为 [角色]，我想 [功能]，以便于 [价值]

#### 验收标准

1. **WHEN** [事件] **THEN** [系统] **SHALL** [响应]
2. **IF** [前置条件] **THEN** [系统] **SHALL** [响应]
3. **WHEN** [事件] **AND** [条件] **THEN** [系统] **SHALL** [响应]

#### 具体要求

- [ ] 1.1 {具体要求描述}
- [ ] 1.2 {具体要求描述}
- [ ] 1.3 {具体要求描述}

---

### Requirement 2: {需求标题}

**用户故事**: 作为 [角色]，我想 [功能]，以便于 [价值]

#### 验收标准

1. **WHEN** [事件] **THEN** [系统] **SHALL** [响应]
2. **IF** [前置条件] **THEN** [系统] **SHALL** [响应]

#### 具体要求

- [ ] 2.1 {具体要求描述}
- [ ] 2.2 {具体要求描述}

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

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
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

1. **{验收项 1}**: {验收标准}
2. **{验收项 2}**: {验收标准}
3. **{验收项 3}**: {验收标准}

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

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

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

- [业务约束条件]

### 资源约束

- 开发时间: {预计天数}
- Story Point: {SP 值} (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `{package_name}: {version}` - {用途}
- `{internal_module}` - {用途}

### 服务依赖

- **Main → BMP**: gRPC 调用（如需要）
- **Admin → Main**: HTTP API 调用
- **Frontend → Admin**: HTTP API 调用

### 业务依赖

- [依赖的其他功能或模块]
- [前置条件]

---

## 风险和缓解

### 风险 1: {风险描述}

**影响**: 高/中/低  
**概率**: 高/中/低  
**缓解措施**:

- {缓解措施 1}
- {缓解措施 2}

### 风险 2: {风险描述}

**影响**: 高/中/低  
**概率**: 高/中/低  
**缓解措施**:

- {缓解措施 1}
- {缓解措施 2}

---

## 时间表

- **Phase 1 - {阶段名}**: {天数}
- **Phase 2 - {阶段名}**: {天数}
- **Phase 3 - {阶段名}**: {天数}
- **总计**: {总天数}（SP = {值}）

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

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- [外部参考链接]

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: {YYYY-MM-DD}  
**作者**: {团队/个人}  
**审核者**: {审核者}
