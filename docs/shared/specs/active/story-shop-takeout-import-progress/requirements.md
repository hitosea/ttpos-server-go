# 外卖菜单导入进度条显示 需求文档

> 本文档定义 外卖菜单导入进度条显示 功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.11.0-takeout-menu-import-progress.md](../../../../team/proposals/2025-12/v2.11.0-takeout-menu-import-progress.md) |
| **创建日期**      | 2025-12-16                                                                                                 |
| **负责人**        | weifashi                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | weifashi             |
| **审核日期** | 2025-12-16             |
| **审核意见** | 需求清晰,验收标准明确,技术方案可行,批准进入设计阶段         |

---

## 📋 概述

当前外卖菜单导入功能(`ImportMenu`)在执行时,前端无法获知导入进度,用户体验不佳。用户只能看到静态的提示信息,无法了解具体进度,不知道导入是否正在进行中,可能重复点击导入按钮,导致并发问题。

本需求旨在实现外卖菜单导入过程的实时进度反馈机制,通过数据库状态记录、日志表记录历史、后端实时更新进度、前端轮询展示等措施,让用户能够实时看到导入进度,避免重复操作,提升用户体验,并便于问题追溯。

**核心价值**：
- 提升用户体验：用户可以实时看到导入进度(如"正在同步分类 3/10","正在导入商品 50/200"),减少等待焦虑
- 避免重复操作：通过状态标识防止用户在导入进行中重复触发
- 问题定位：导入失败时可以明确告知用户在哪个环节失败,便于排查问题
- 历史记录：提供完整的导入历史记录,支持追溯和统计分析

## 🎯 产品对齐

该功能支持 TTPOS 的外卖平台集成战略,通过提升导入过程的透明度和可追溯性,增强商户对系统的信任度,降低客服成本,为后续扩展到更多外卖平台的导入场景奠定基础。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在导入外卖菜单时看到实时进度并查看历史导入记录  
**以便于** 了解导入状态,避免重复操作,并在出错时快速定位问题

---

## 功能需求

### Requirement 1: 导入状态管理

**用户故事**: 作为商户管理员，我想在点击导入菜单时系统能够检查是否有正在进行的导入任务，以便于避免重复操作导致数据混乱。

#### 验收标准

1. **WHEN** 用户点击"导入菜单"按钮 **THEN** 系统 **SHALL** 检查 `ttpos_takeout` 表中的 `import_status` 字段
2. **IF** `import_status` 为 1(导入中) **THEN** 系统 **SHALL** 提示"导入正在进行中,请稍后再试"并拒绝新的导入请求
3. **IF** `import_status` 为 0、2、3(未导入、成功、失败) **THEN** 系统 **SHALL** 允许开始新的导入任务
4. **WHEN** 导入任务开始 **THEN** 系统 **SHALL** 将 `import_status` 更新为 1(导入中)
5. **WHEN** 导入任务完成 **THEN** 系统 **SHALL** 将 `import_status` 更新为 2(导入成功)或 3(导入失败)
6. **IF** 导入任务异常中断(如服务重启) **THEN** 系统 **SHALL** 在重启后将超时的导入任务状态重置为 3(导入失败)

#### 具体要求

- [x] 1.1 在 `ttpos_takeout` 表中新增 `import_status`、`import_progress`、`import_start_time`、`import_end_time`、`import_error` 字段
- [ ] 1.2 实现导入状态检查逻辑(`checkImportStatus` 方法)
- [ ] 1.3 实现导入状态更新逻辑(`updateImportStatus` 方法)
- [ ] 1.4 在 `ImportMenu` 方法开始前调用状态检查
- [ ] 1.5 在导入完成/失败后更新状态
- [ ] 1.6 实现超时导入任务检测和重置机制(通过定时任务或启动时检查)

---

### Requirement 2: 导入进度实时更新

**用户故事**: 作为商户管理员，我想在导入过程中看到实时的进度信息(当前步骤、完成百分比、已处理数量等)，以便于了解导入是否正常进行。

#### 验收标准

1. **WHEN** 导入任务开始 **THEN** 系统 **SHALL** 初始化 `import_progress` 字段,包含开始时间和初始进度
2. **WHEN** 系统在 `syncCategories` 阶段处理分类时 **THEN** 系统 **SHALL** 实时更新进度为"正在同步分类 X/Y"
3. **WHEN** 系统在 `syncProducts` 阶段处理商品时 **THEN** 系统 **SHALL** 实时更新进度为"正在导入商品 X/Y"
4. **WHEN** 完成一个步骤 **THEN** 系统 **SHALL** 更新进度百分比(基于总步骤数计算)
5. **WHEN** 前端轮询查询进度 **THEN** 系统 **SHALL** 返回当前的进度信息(步骤、百分比、预估剩余时间)
6. **IF** 导入过程中发生错误 **THEN** 系统 **SHALL** 在 `import_progress` 中记录错误信息

#### 具体要求

- [ ] 2.1 定义 `ImportProgress` 结构体,包含 `current_step`、`step_name`、`current_count`、`total_count`、`percentage`、`start_time`、`estimated_time` 字段
- [ ] 2.2 实现进度更新逻辑(`updateImportProgress` 方法)
- [ ] 2.3 在 `syncCategories` 方法中每处理 N 个分类(如每 5 个)更新一次进度
- [ ] 2.4 在 `syncProducts` 方法中每处理 N 个商品(如每 10 个)更新一次进度
- [ ] 2.5 实现进度百分比计算逻辑(根据当前步骤和总步骤数)
- [ ] 2.6 实现预估剩余时间计算逻辑(基于已用时间和完成百分比)
- [ ] 2.7 新增 API 接口 `GET /api/takeout/import/progress` 返回当前导入进度

---

### Requirement 3: 导入历史日志记录

**用户故事**: 作为商户管理员，我想查看所有的导入历史记录(成功、失败、进行中)，以便于追溯问题和统计分析。

#### 验收标准

1. **WHEN** 导入任务开始 **THEN** 系统 **SHALL** 在 `ttpos_takeout_import_log` 表中创建新的日志记录
2. **WHEN** 导入过程中 **THEN** 系统 **SHALL** 实时更新日志记录的 `progress` 和 `status` 字段
3. **WHEN** 导入完成 **THEN** 系统 **SHALL** 更新日志记录的 `end_time`、`duration`、`success_count`、`failure_count` 字段
4. **WHEN** 导入失败 **THEN** 系统 **SHALL** 记录 `error_message` 字段
5. **WHEN** 用户请求历史日志 **THEN** 系统 **SHALL** 返回按时间倒序排列的日志列表
6. **WHEN** 用户查看日志列表 **THEN** 系统 **SHALL** 展示导入方向、状态、进度、时间、成功/失败数量等信息

#### 具体要求

- [x] 3.1 创建 `ttpos_takeout_import_log` 表,包含 `id`、`uuid`、`platform`、`import_type`、`import_direction`、`status`、`progress`、`success_count`、`failure_count`、`total_count`、`error_message`、`start_time`、`end_time`、`duration`、`create_time`、`update_time`、`delete_time` 字段
- [ ] 3.2 实现日志记录创建逻辑(`createImportLog` 方法)
- [ ] 3.3 实现日志记录更新逻辑(`updateImportLog` 方法)
- [ ] 3.4 实现日志记录完成逻辑(`completeImportLog` 方法)
- [ ] 3.5 在 `ImportMenu` 方法开始时创建日志记录
- [ ] 3.6 在导入过程中同时更新 `ttpos_takeout` 和 `ttpos_takeout_import_log` 的进度信息
- [ ] 3.7 新增 API 接口 `GET /api/takeout/import/logs` 返回历史导入日志列表(支持分页、按平台筛选)

---

### Requirement 4: 前端进度展示

**用户故事**: 作为商户管理员，我想在前端界面看到美观的进度条和历史日志列表，以便于直观了解导入状态。

#### 验收标准

1. **WHEN** 用户点击"导入菜单"按钮后 **THEN** 系统 **SHALL** 显示进度对话框,包含进度条、当前步骤描述、完成百分比、预估剩余时间
2. **WHEN** 导入进行中 **THEN** 前端 **SHALL** 每 2-3 秒轮询一次进度接口,更新进度条
3. **WHEN** 导入完成 **THEN** 系统 **SHALL** 显示成功提示,包含成功数量和失败数量
4. **WHEN** 导入失败 **THEN** 系统 **SHALL** 显示失败提示和错误信息
5. **WHEN** 用户进入日志页面 **THEN** 系统 **SHALL** 展示历史导入记录列表,包含导入方向、状态、进度、时间等信息
6. **WHEN** 用户查看进行中的导入 **THEN** 日志列表中 **SHALL** 显示实时的进度百分比

#### 具体要求

- [ ] 4.1 创建进度对话框组件,包含进度条(0-100%)、步骤描述、百分比文本、预估剩余时间
- [ ] 4.2 实现轮询逻辑(每 2-3 秒调用一次进度接口)
- [ ] 4.3 实现进度条动画效果(平滑过渡)
- [ ] 4.4 实现导入完成/失败的提示对话框
- [ ] 4.5 创建日志列表页面,展示历史导入记录(卡片式布局,类似截图样式)
- [ ] 4.6 实现日志列表的分页和筛选功能
- [ ] 4.7 在日志列表中区分不同状态(进行中-显示进度、成功-显示绿色图标、失败-显示红色图标和错误信息)

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/database.mdc` - 数据库开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/takeout/import/progress`, `/api/v1/takeout/import/logs`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 新增字段：`ttpos_takeout.import_status`、`import_progress`、`import_start_time`、`import_end_time`、`import_error`
- [x] 新表：`ttpos_takeout_import_log`
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms（进度查询接口）
- [ ] 批量导入 100 个商品进度更新不超过 10 次（每 10 个商品更新一次）
- [ ] 进度更新不应显著影响导入性能（更新操作应异步或批量）
- [ ] 日志查询接口支持分页，单页最多返回 50 条记录
- [ ] 使用索引优化日志查询（`idx_platform`、`idx_import_type`、`idx_status`、`idx_create_time`）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 单元测试：进度计算逻辑、状态更新逻辑、日志记录逻辑
- [ ] 集成测试：完整的导入流程(从开始到完成,包含进度更新)
- [ ] API 测试：进度查询接口、日志查询接口的正常场景和异常场景
- [ ] 并发测试：防止重复导入的并发控制逻辑
- [ ] 前端测试：进度条组件、轮询逻辑、日志列表展示

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有前端文案使用多语言实现
- [ ] 进度描述文案支持多语言
- [ ] 错误提示信息支持多语言
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证（JWT Token）
- [ ] 导入操作前验证商户权限
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验和输出转义）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 进度更新失败不应阻塞导入流程
- [ ] 事务管理（保证导入数据和日志记录的一致性）
- [ ] 错误日志记录（使用 Logger，记录导入失败详情）
- [ ] 超时导入任务自动重置(避免永久锁定)
- [ ] 前端轮询失败时优雅降级(停止轮询并提示用户刷新)

---

## 验收标准

### 功能验收

1. **导入状态管理**: 导入前检查状态,导入中拒绝新请求,导入完成更新状态
2. **进度实时更新**: 导入过程中实时更新进度,前端可以查询到当前步骤、百分比、预估时间
3. **历史日志记录**: 所有导入记录保存到日志表,支持查询和展示
4. **前端进度展示**: 进度对话框美观流畅,日志列表展示完整,状态区分清晰
5. **并发控制**: 同一商户不能同时进行多个导入任务
6. **异常处理**: 导入失败时记录错误信息,超时任务自动重置

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%，Repository ≥ 80%）
2. **API 测试**: 进度查询接口和日志查询接口的正常场景和异常场景测试通过
3. **集成测试**: 完整的导入流程测试通过（从开始到完成,包含进度更新和日志记录）
4. **并发测试**: 防止重复导入的并发控制逻辑测试通过
5. **前端测试**: 进度条组件、轮询逻辑、日志列表展示测试通过
6. **手动测试**: 浏览器兼容性测试通过（Chrome、Safari、Firefox、Edge）

### 文档验收

1. **技术文档**: design.md 完整且准确，包含架构设计和实现细节
2. **API 文档**: 进度查询接口和日志查询接口文档完整（请求参数、响应格式、错误码）
3. **数据库文档**: 迁移脚本和表结构变更文档完整
4. **测试文档**: tasks.md 中的测试任务完成，测试用例文档完整

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以小写字母开头（如 `ITakeoutSrv` → `takeoutSrv`）
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc`

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 组件命名使用 PascalCase
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 导入操作必须在事务中执行，保证数据一致性
- 进度更新操作不应显著影响导入性能（批量更新或异步更新）
- 同一商户同一时间只能有一个导入任务进行
- 历史日志记录不应被删除（软删除）
- 超时阈值设置为 30 分钟（超过 30 分钟的导入任务视为异常）

### 资源约束

- 开发时间: 3-4 天
- Story Point: 5 (待技术评审确认)

---

## 依赖关系

### 技术依赖

- **Gin 框架**: Web 框架
- **GORM**: ORM 框架
- **Redis**: 分布式锁(可选,用于并发控制)
- **Element Plus**: 前端 UI 组件库
- **Vue 3**: 前端框架

### 服务依赖

- **Takeout Service**: 处理外卖导入相关业务逻辑
- **Product Service**: 处理商品相关业务逻辑

### 业务依赖

- **外卖功能**: 商户必须开启外卖功能才能使用导入功能
- **导入菜单功能**: 依赖现有的 `ImportMenu` 方法

---

## 风险和缓解

### 风险 1: 进度更新频繁影响导入性能

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 批量更新进度（每处理 N 个商品更新一次，而不是每个商品都更新）
- 使用异步方式更新进度（不阻塞主导入流程）
- 对进度更新操作进行性能测试，确保不超过 50ms

### 风险 2: 前端轮询频率过高导致服务器压力

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 设置合理的轮询间隔（2-3 秒）
- 导入完成后立即停止轮询
- 在进度查询接口中增加缓存（1-2 秒缓存）
- 限制单个商户的轮询频率（防止恶意请求）

### 风险 3: 超时导入任务未被重置导致永久锁定

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 实现超时检测机制（服务启动时或定时任务）
- 设置合理的超时阈值（30 分钟）
- 记录超时任务日志，便于问题追溯
- 提供手动重置接口（运维使用）

### 风险 4: 日志表数据量过大影响查询性能

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用合适的索引（`idx_platform`、`idx_create_time`等）
- 实现分页查询，限制单页数量（最多 50 条）
- 定期归档或清理旧日志（如保留最近 6 个月的记录）
- 对日志查询接口进行性能测试

### 风险 5: 并发导入控制失效

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 使用数据库行锁或 Redis 分布式锁
- 在导入开始前检查状态并更新为"导入中"（原子操作）
- 增加幂等性检查，避免重复导入相同数据
- 记录并发冲突日志，便于问题追溯

---

## 时间表

- **Phase 1 - 数据库设计和迁移**: 0.5 天
  - 创建迁移脚本
  - 添加 `ttpos_takeout` 字段
  - 创建 `ttpos_takeout_import_log` 表
  
- **Phase 2 - 后端状态管理和进度更新**: 1.5 天
  - 实现导入状态检查逻辑
  - 实现进度更新逻辑
  - 实现日志记录逻辑
  - 修改 `ImportMenu` 方法集成进度更新
  
- **Phase 3 - 后端 API 接口开发**: 0.5 天
  - 实现进度查询接口
  - 实现日志查询接口
  - API 测试
  
- **Phase 4 - 前端进度对话框**: 0.5 天
  - 创建进度对话框组件
  - 实现轮询逻辑
  - 实现进度条动画
  
- **Phase 5 - 前端日志列表页面**: 0.5 天
  - 创建日志列表页面
  - 实现分页和筛选
  - 状态展示和样式调整
  
- **Phase 6 - 测试与联调**: 0.5 天
  - 单元测试
  - 集成测试
  - 前后端联调
  - 浏览器兼容性测试
  
- **总计**: 4 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 相关代码

- `main/app/service/takeout.go` - 外卖服务实现（`ImportMenu` 方法）
- `main/app/modules/takeout/domain/menu/valueobject/` - 外卖菜单领域对象
- `main/app/model/` - 数据模型定义

### 外部参考

- Element Plus Progress 组件: https://element-plus.org/zh-CN/component/progress.html
- Element Plus Timeline 组件: https://element-plus.org/zh-CN/component/timeline.html

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2025-12/2025-12-16.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: weifashi  
**审核者**: 待审核

