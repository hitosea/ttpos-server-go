> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 云平台-日志管理(外卖相关) 需求文档

> 本文档定义云平台外卖日志管理功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                   |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-cloud-platform-takeout-log-management.md](../../../../team/proposals/2025-12/v2.12.0-cloud-platform-takeout-log-management.md) |
| **创建日期**      | 2025-12-17                                                                                                             |
| **负责人**        | 待分配                                                                                                                 |
| **目标 Sprint**   | Sprint 待定                                                                                                            |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                             |

## 📋 审核状态

| 项目         | 内容     |
| ------------ | -------- |
| **审核状态** | 待审核   |
| **审核人**   | 待分配   |
| **审核日期** | -        |
| **审核意见** | -        |

---

## 📋 概述

为云平台(Admin)添加外卖相关的日志管理功能,让平台管理员和商户能够查看、筛选和追溯外卖订单同步日志。当前外卖订单同步(Grab/LINE MAN)已经记录日志到数据库,但缺少统一的管理界面。本功能复用现有的日志基础设施,只需在 Admin 端添加 API 接口和前端页面,即可快速实现日志管理功能。

**用户价值**:
- 快速定位外卖同步失败原因,减少 80% 的排查时间
- 完整追溯外卖订单同步历史,支持业务审计和数据核对
- 无需登录服务器即可查看日志,降低技术支持成本
- 实时监控外卖渠道同步状态,及时发现和处理异常情况

## 🎯 产品对齐

**产品愿景**: 为商户提供稳定可靠的外卖订单同步服务,确保订单数据的准确性和实时性。

**业务目标**:
1. 提升问题排查效率,减少客服工单处理时间
2. 增强数据可追溯性,支持业务审计
3. 降低运维成本,减少服务器登录次数
4. 提高用户满意度,快速解决外卖同步问题

---

## 📝 用户故事

**作为** 平台管理员  
**我想** 查看所有商户的外卖订单同步日志,并按门店、平台、类型、状态进行筛选  
**以便于** 监控外卖渠道的整体运行状况,快速定位和解决问题

**作为** 商户管理员  
**我想** 查看我店铺的外卖订单同步日志,了解同步状态和失败原因  
**以便于** 及时发现和解决外卖订单同步问题,避免影响营业

**作为** 技术支持人员  
**我想** 查看外卖同步失败的详细错误信息  
**以便于** 快速排查问题并给出解决方案,减少客服响应时间

---

## 功能需求

### Requirement 1: 日志列表查询

**用户故事**: 作为平台/商户管理员,我想查看外卖同步日志列表,以便于了解同步历史和状态

#### 验收标准

1. **WHEN** 用户访问日志管理页面 **THEN** 系统 **SHALL** 展示外卖同步日志列表,按时间倒序排列(最新的在前)
2. **IF** 用户是平台管理员 **THEN** 系统 **SHALL** 展示所有商户的日志
3. **IF** 用户是商户管理员 **THEN** 系统 **SHALL** 只展示该商户的日志
4. **WHEN** 日志列表加载失败 **THEN** 系统 **SHALL** 展示错误提示,并提供重试按钮

#### 具体要求

- [x] 1.1 日志列表包含以下字段:
  - 类型(import_direction): TTPOS推送/Grab推送/TTPOS获取
  - 门店(company_uuid): 显示门店名称
  - 平台(platform): Grab/LINE MAN
  - 状态(status): 进行中/成功/失败
  - 统计(success_count/failure_count/total_count): 成功数/失败数/总数
  - 时间(createtime): 创建时间
  - 耗时(duration): 同步耗时
  - 原因(error_message): 失败原因
- [x] 1.2 日志列表默认按创建时间倒序排列
- [x] 1.3 进行中的日志显示进度条(progress 字段)
- [x] 1.4 成功的日志显示绿色标签
- [x] 1.5 失败的日志显示红色标签,并展示失败原因
- [x] 1.6 支持分页查询,每页默认 20 条,可选 20/50/100
- [x] 1.7 显示总记录数和当前页码

---

### Requirement 2: 日志筛选功能

**用户故事**: 作为用户,我想按门店、平台、类型、状态筛选日志,以便于快速找到目标日志

#### 验收标准

1. **WHEN** 用户选择筛选条件并点击"查询" **THEN** 系统 **SHALL** 只展示符合条件的日志
2. **WHEN** 用户点击"重置" **THEN** 系统 **SHALL** 清空筛选条件,恢复默认状态
3. **WHEN** 筛选条件无匹配结果 **THEN** 系统 **SHALL** 展示"暂无数据"提示
4. **WHEN** 用户同时选择多个筛选条件 **THEN** 系统 **SHALL** 按 AND 逻辑进行筛选

#### 具体要求

- [x] 2.1 支持按门店筛选(平台管理员可见,商户管理员不可见)
  - 门店下拉框列出所有商户
  - 支持清空选择(查询所有门店)
- [x] 2.2 支持按平台筛选
  - 选项: 全部/Grab/LINE MAN
  - 默认: 全部
- [x] 2.3 支持按类型筛选
  - 选项: 全部/TTPOS推送/平台推送/TTPOS获取
  - 默认: 全部
- [x] 2.4 支持按状态筛选
  - 选项: 全部/进行中/成功/失败
  - 默认: 全部
- [x] 2.5 筛选条件变更后,页码重置为第 1 页
- [x] 2.6 筛选条件保存到 URL 参数,支持刷新页面后保持筛选状态

---

### Requirement 3: 日志详情查看

**用户故事**: 作为用户,我想查看日志的完整错误信息,以便于深入分析问题

#### 验收标准

1. **WHEN** 用户点击失败日志的"查看详情"按钮 **THEN** 系统 **SHALL** 弹出对话框,展示完整的错误信息
2. **WHEN** 错误信息过长 **THEN** 系统 **SHALL** 使用滚动条展示,而不是截断
3. **WHEN** 用户关闭详情对话框 **THEN** 系统 **SHALL** 返回日志列表页面

#### 具体要求

- [x] 3.1 失败日志在"原因"列显示简短的错误信息(截取前 50 字符)
- [x] 3.2 失败日志提供"查看详情"链接
- [x] 3.3 详情对话框展示完整的 error_message 字段
- [x] 3.4 错误信息使用等宽字体展示,保持格式(pre 标签)
- [x] 3.5 详情对话框支持复制错误信息到剪贴板
- [x] 3.6 详情对话框宽度为 60%,高度自适应,最大高度 400px

---

### Requirement 4: 权限控制

**用户故事**: 作为系统,我要确保用户只能查看有权限的日志,以便于保护数据安全

#### 验收标准

1. **IF** 用户未登录 **THEN** 系统 **SHALL** 返回 401 错误,跳转到登录页面
2. **IF** 用户无权限访问 Admin 端 **THEN** 系统 **SHALL** 返回 403 错误,提示无权限
3. **IF** 商户管理员尝试查询其他商户的日志 **THEN** 系统 **SHALL** 只返回该商户的日志,忽略非法参数
4. **WHEN** 平台管理员查询日志 **AND** 未指定门店 **THEN** 系统 **SHALL** 返回所有商户的日志

#### 具体要求

- [x] 4.1 使用 `middleware.Internal()` 中间件保护 API
- [x] 4.2 平台管理员可以查询所有商户的日志
- [x] 4.3 平台管理员可以指定 company_uuid 查询指定商户的日志
- [x] 4.4 商户管理员只能查询当前商户的日志(强制设置 company_uuid)
- [x] 4.5 商户管理员看不到"门店"筛选下拉框
- [x] 4.6 其他角色无法访问该功能,返回 403 错误

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 API → Application Service → Domain Service → Repository 四层架构
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Application Service 调用 Domain Service,Domain Service 访问 Repository
- **代码复用**: 复用现有的 Shop 端日志基础设施,无需重复开发
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/admin/takeout/logs`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息使用 `page`, `page_size`, `total` 字段
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 复用现有表: `ttpos_takeout_import_log` (无需创建新表)
- [x] 表结构已包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 索引已建立: platform, import_type, status, create_time
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] API 响应时间 < 500ms (10万条数据)
- [x] 数据库查询优化（使用已有索引）
- [x] 支持并发查询 100 TPS
- [x] 分页查询避免全表扫描
- [x] 当数据量超过 100 万条时,考虑日志归档策略

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [x] API Handler 层测试覆盖率 ≥ 80%
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [x] 权限测试覆盖所有角色
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、泰语等）
- [x] 所有文案使用多语言实现
- [x] 错误信息支持多语言
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证(Internal Token)
- [x] 权限控制(基于角色)
- [x] SQL 注入防护（使用参数化查询/GORM）
- [x] XSS 防护（前端输入校验/Element Plus）
- [x] 敏感数据不返回(如完整的 API Token)
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级(展示错误提示)
- [x] 错误日志记录（使用 Logger）
- [x] 数据一致性保证(软删除)
- [x] 提供重试机制(查询失败时)

---

## 验收标准

### 功能验收

1. **日志列表展示**: 
   - ✅ 展示所有必需字段(类型、门店、平台、状态、统计、时间、耗时、原因)
   - ✅ 按时间倒序排列
   - ✅ 状态可视化(进度条、标签)
   - ✅ 支持分页(20/50/100 条/页)

2. **筛选功能**: 
   - ✅ 支持按门店、平台、类型、状态筛选
   - ✅ 筛选条件组合生效(AND 逻辑)
   - ✅ 无匹配结果时展示"暂无数据"
   - ✅ 重置功能正常

3. **日志详情**: 
   - ✅ 失败日志可查看完整错误信息
   - ✅ 详情对话框展示格式化的错误信息
   - ✅ 支持复制错误信息

4. **权限控制**: 
   - ✅ 平台管理员可查询所有商户日志
   - ✅ 商户管理员只能查询自己的日志
   - ✅ 无权限用户无法访问

### 测试验收

1. **单元测试**: 
   - ✅ API Handler 测试覆盖率 ≥ 80%
   - ✅ 权限测试覆盖所有角色

2. **API 测试**: 
   - ✅ 所有筛选条件测试通过
   - ✅ 分页测试通过
   - ✅ 权限测试通过

3. **集成测试**: 
   - ✅ 端到端流程测试通过
   - ✅ 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整(Swagger 注释)
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以小写字母开头
- Application Service 调用 Domain Service
- Domain Service 访问 Repository
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 使用 Composition API
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 无需创建新表,复用现有 `ttpos_takeout_import_log` 表
- 无需修改表结构
- 复用 Shop 端的 Application Service,无需重复开发

### 资源约束

- 开发时间: 5-8 天
- Story Point: 3-5

---

## 依赖关系

### 技术依赖

**后端**:
- `ttpos-server-go/main` - Go Main 模块
- `github.com/gin-gonic/gin` - Web 框架
- `gorm.io/gorm` - ORM 框架
- `ttpos-server-go/app/modules/takeout` - 外卖模块(已存在)

**前端**:
- `vue@3.x` - 前端框架
- `element-plus` - UI 组件库
- `typescript` - 类型系统
- `axios` - HTTP 客户端

### 服务依赖

- **Admin → Main**: HTTP API 调用 (`GET /api/v1/admin/takeout/logs`)
- **Frontend → Admin**: HTTP API 调用

### 业务依赖

- 依赖现有的外卖同步功能(Shop 端)
- 依赖现有的日志记录功能(ImportProgressService)
- 依赖现有的权限系统(middleware.Internal)

---

## 风险和缓解

### 风险 1: 日志数据量过大影响查询性能

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用已有索引优化查询(platform, status, create_time)
- 限制每页最大数量(100 条)
- 添加查询超时限制(3 秒)
- 未来考虑日志归档策略(如按月分表或归档到 Elasticsearch)

### 风险 2: 权限系统集成复杂度

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在开发前确认现有权限系统的实现方式
- 参考现有 Admin 端 API 的权限控制实现
- 预留 2 天时间用于权限集成和测试
- 如权限系统不完善,先使用简单的 Internal Token 验证

### 风险 3: 多租户数据隔离问题

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 在 API 层严格校验用户权限
- 商户管理员强制设置 company_uuid 为当前商户
- 添加单元测试覆盖数据隔离场景
- Code Review 重点检查权限相关代码

---

## 时间表

- **Phase 1 - 后端开发**: 3-4 天
  - Admin API Handler 开发 (1 天)
  - 权限控制集成 (1 天)
  - 单元测试和集成测试 (1-2 天)
  
- **Phase 2 - 前端开发**: 2-3 天
  - 日志管理页面 UI 实现 (1.5 天)
  - 筛选和分页功能 (0.5 天)
  - 联调和测试 (1 天)
  
- **Phase 3 - 测试和部署**: 1 天
  - 端到端测试 (0.5 天)
  - 浏览器兼容性测试 (0.5 天)
  
- **总计**: 5-8 天（SP = 3-5）

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

### 参考设计文档

- [外卖菜单导入进度设计文档](../story-shop-takeout-import-progress/design.md) - Shop 端日志功能参考

### 外部参考

- [Element Plus 文档](https://element-plus.org/) - UI 组件库
- [Vue 3 文档](https://vuejs.org/) - 前端框架

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-17.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-17  
**作者**: weifashi  
**审核者**: 待分配

