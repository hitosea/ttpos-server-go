# ERPNext 对接 - 物品管理增加默认销售单位 需求文档

> 本文档定义 ERPNext 对接 - 物品管理增加默认销售单位的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/shop-erpnext-default-sales-unit.md](../../../../team/proposals/2026-01/shop-erpnext-default-sales-unit.md) |
| **创建日期**      | 2026-01-19                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

本功能旨在完善 ERPNext 与 TTPOS 之间的物品数据同步，增加默认销售单位字段的同步、显示和编辑功能。通过该功能，可以确保 ERPNext 中设置的默认销售单位能够正确同步到 TTPOS，并在物品管理界面中展示和编辑，提升数据一致性和用户体验。

## 🎯 产品对齐

该功能支持以下产品目标：
- **数据一致性**：确保 ERPNext 与 TTPOS 之间的物品数据完全同步
- **用户体验**：提供清晰的默认销售单位信息展示和便捷的编辑功能
- **业务灵活性**：支持总部统一管理，同时允许子店根据实际情况灵活调整

## 📝 用户故事

**作为** 店长/商户管理员  
**我想** 在物品管理中查看和设置默认销售单位（ERPNext），并在创建物品时提交默认销售单位  
**以便于** 确保 ERPNext 与 TTPOS 数据一致，提升物品管理效率，满足总部统一管理和子店灵活调整的业务需求

---

## 功能需求

### Requirement 1: ERPNext 同步默认销售单位字段

**用户故事**: 作为 系统管理员，我想 ERPNext 同步物品数据时同步默认销售单位字段，以便于 确保数据一致性

#### 验收标准

1. **WHEN** ERPNext 同步物品数据 **THEN** 系统 **SHALL** 同步 Item 的 `Default Sales Unit of Measure` 字段到 TTPOS
2. **IF** ERPNext 中 Item 已设置 `Default Sales Unit of Measure` **THEN** 系统 **SHALL** 将对应的单位名称同步到 TTPOS 物品表
3. **IF** ERPNext 中 Item 未设置 `Default Sales Unit of Measure` **THEN** 系统 **SHALL** 将默认销售单位字段设置为空或默认值

#### 具体要求

- [ ] 1.1 在物品同步接口中增加 `default_sales_unit` 字段的同步逻辑
- [ ] 1.2 同步时验证单位是否存在，如不存在则记录日志或使用默认值
- [ ] 1.3 支持增量同步和全量同步两种模式

---

### Requirement 2: 物品详情页显示默认销售单位

**用户故事**: 作为 店长/商户管理员，我想 在物品详情页查看默认销售单位信息，以便于 了解物品的标准销售单位

#### 验收标准

1. **WHEN** 查看物品详情 **THEN** 系统 **SHALL** 在基本信息区域显示"默认销售单位（ERPNext）"字段
2. **IF** ERPNext 中 Item 已设置 `Default Sales Unit of Measure` **THEN** 系统 **SHALL** 显示对应的单位名称（如：箱、件、个等）
3. **IF** ERPNext 中 Item 未设置 `Default Sales Unit of Measure` **THEN** 系统 **SHALL** 显示"无"

#### 具体要求

- [ ] 2.1 在物品详情页基本信息区域添加"默认销售单位（ERPNext）"字段
- [ ] 2.2 字段显示格式：有值时显示单位名称，无值时显示"无"
- [ ] 2.3 字段标签清晰标识来源为 ERPNext

---

### Requirement 3: 权限控制 - 总部来源只读

**用户故事**: 作为 系统管理员，我想 总部来源的物品默认销售单位为只读，以便于 确保总部统一管理

#### 验收标准

1. **IF** 物品来源于总部 **THEN** 系统 **SHALL** 将默认销售单位字段设置为只读状态，子店无法修改
2. **IF** 物品不是总部来源 **THEN** 系统 **SHALL** 允许子店编辑默认销售单位

#### 具体要求

- [ ] 3.1 根据物品的 `source` 或 `is_headquarters` 字段判断是否为总部来源
- [ ] 3.2 前端表单根据权限控制字段的可编辑性
- [ ] 3.3 后端 API 验证权限，拒绝未授权修改

---

### Requirement 4: 创建/编辑物品时设置默认销售单位

**用户故事**: 作为 店长/商户管理员，我想 在创建和编辑物品时设置默认销售单位，以便于 提升操作效率

#### 验收标准

1. **WHEN** 创建物品 **THEN** 系统 **SHALL** 允许提交默认销售单位字段
2. **WHEN** 编辑默认销售单位 **THEN** 系统 **SHALL** 在下拉选项中显示该物品的所有单位（基准单位 + 非基准单位）
3. **IF** 物品只有基准单位 **THEN** 系统 **SHALL** 在下拉选项中只显示基准单位
4. **WHEN** 提交默认销售单位 **THEN** 系统 **SHALL** 验证该单位必须是该物品已配置的单位之一

#### 具体要求

- [ ] 4.1 在创建物品表单中添加"默认销售单位"字段
- [ ] 4.2 在编辑物品表单中添加"默认销售单位"字段
- [ ] 4.3 下拉选项动态加载该物品的所有单位（基准单位 + 非基准单位）
- [ ] 4.4 后端验证默认销售单位必须是该物品已配置的单位
- [ ] 4.5 保存时更新物品表的 `default_sales_unit` 字段

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

- [x] URL 使用 snake_case 命名（如：`/api/v1/product_info`）
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

**新增字段要求**：
- 在物品表中增加 `default_sales_unit` 字段（bigint unsigned，关联单位表）
- 字段允许为空（NULL），表示未设置默认销售单位

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

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

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

1. **ERPNext 同步功能**: ERPNext 同步物品数据时，能够正确同步 `Default Sales Unit of Measure` 字段到 TTPOS
2. **物品详情页显示**: 物品详情页能够正确显示默认销售单位信息，有值显示单位名称，无值显示"无"
3. **权限控制**: 总部来源的物品默认销售单位为只读，非总部来源的物品允许编辑
4. **创建/编辑功能**: 创建和编辑物品时能够设置默认销售单位，下拉选项正确显示所有单位
5. **数据验证**: 提交的默认销售单位必须是该物品已配置的单位之一

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%）
2. **API 测试**: 所有接口测试通过（同步接口、创建接口、更新接口、详情接口）
3. **集成测试**: 端到端流程测试通过（ERPNext 同步 → 查看详情 → 编辑 → 保存）
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

- 不涉及 BMP 模块

#### PHP 模块

- 不涉及 PHP 模块

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 默认销售单位必须是该物品已配置的单位之一
- 总部来源的物品默认销售单位不允许子店修改
- ERPNext 同步时，如果单位不存在，需要记录日志或使用默认值

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ERPNext API` - 物品同步接口
- `物品单位管理模块` - 获取物品的所有单位列表

### 服务依赖

- **Main → ERPNext**: HTTP API 调用（同步接口）
- **Frontend → Main**: HTTP API 调用（物品 CRUD 接口）

### 业务依赖

- 物品单位管理功能（已存在）
- ERPNext 同步功能（已存在）
- 物品权限控制功能（已存在）

---

## 风险和缓解

### 风险 1: ERPNext API 可能不支持该字段

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 提前查阅 ERPNext API 文档，确认字段可用性
- 如不可用，需要与 ERPNext 团队沟通，确认是否可以通过其他方式获取
- 如果确实无法获取，考虑使用其他字段或手动配置

### 风险 2: 现有物品数据没有默认销售单位

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 对于历史数据，设置合理的默认值（如使用基准单位）或允许为空
- 提供数据迁移脚本，将基准单位设置为默认销售单位（可选）
- 在 UI 中明确标识未设置的情况

### 风险 3: 权限控制逻辑与现有体系不一致

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 参考现有的物品权限控制逻辑，确保一致性
- 与权限管理模块负责人确认权限判断标准
- 充分测试权限控制的各种场景

---

## 时间表

- **Phase 1 - 数据库设计和迁移**: 0.5 天
- **Phase 2 - 后端 API 开发**: 2 天
- **Phase 3 - 前端 UI 开发**: 1.5 天
- **Phase 4 - 测试和文档**: 1 天
- **总计**: 5 天（SP = 5）

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

- ERPNext API 文档
- ERPNext Item 同步接口文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2026-01/2026-01-19.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-19  
**作者**: xiezhihuan  
**审核者**: {审核者}
