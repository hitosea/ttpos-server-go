> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# ERP 支付方式 PaymentID 字段 需求文档

> 本文档定义 ERP 支付方式 PaymentID 字段 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-erp-mode-of-payments-paymentid.md](../../../../team/proposals/2025-12/v2.12.0-erp-mode-of-payments-paymentid.md) |
| **创建日期**      | 2025-12-22                                                                                                   |
| **负责人**        | rikugun                                                                                                      |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2025-12-23             |
| **审核意见** | 需求明确，技术方案可行，批准进入设计阶段         |

---

## 📋 概述

在 ERP 模块的 ModeOfPayments 表中增加自定义字段 PaymentID，用于唯一标识支付方式并关联 TTPOS 系统。该功能实现 ERP 和 TTPOS 支付数据的无缝对接，确保支付方式在两个系统间的准确同步，为后续的财务对账和数据统计提供基础。

## 🎯 产品对齐

该功能支持产品在 ERP 集成的核心战略，实现餐饮行业 ERP 系统的深度集成，提升商户的运营效率和数据管理能力。

## 📝 用户故事

**作为** ERP 管理员和 TTPOS 商户管理员  
**我想** 在两种系统中准确关联和管理支付方式数据  
**以便于** 确保支付数据的同步和财务对账的准确性

---

## 功能需求

### Requirement 1: PaymentID 字段管理

**用户故事**: 作为 ERP 管理员，我想在支付方式中增加 PaymentID 字段，以便于与 TTPOS 系统准确关联

#### 验收标准

1. **WHEN** 创建新的支付方式 **THEN** 系统 **SHALL** 自动生成唯一的 PaymentID
2. **WHEN** 查询支付方式列表 **THEN** 系统 **SHALL** 显示完整的 PaymentID 信息
3. **IF** PaymentID 已存在 **THEN** 系统 **SHALL** 拒绝重复创建

#### 具体要求

- [ ] 1.1 PaymentID 字段为必填项，不可编辑
- [ ] 1.2 生成规则：PID + 16位唯一数字组合
- [ ] 1.3 支持全局唯一性校验
- [ ] 1.4 字段类型：字符串，最大长度 20 字符

---

### Requirement 2: 双向数据同步

**用户故事**: 作为商户管理员，我想支付方式变更能自动同步到对方系统，以便于保持数据一致性

#### 验收标准

1. **WHEN** 在 ERP 中创建支付方式 **THEN** 系统 **SHALL** 自动同步到 TTPOS
2. **WHEN** 在 TTPOS 中更新支付方式 **THEN** 系统 **SHALL** 同步变更到 ERP
3. **IF** 同步失败 **THEN** 系统 **SHALL** 记录错误日志并支持重试

#### 具体要求

- [ ] 2.1 ERP → TTPOS 单向同步（店铺同步本店 ERP 支付数据）
- [ ] 2.2 TTPOS → ERP 单向同步（支付方式变更时同步）
- [ ] 2.3 支持状态同步（启用/禁用）
- [ ] 2.4 同步失败时的错误处理和重试机制

---

### Requirement 3: 状态和权限管理

**用户故事**: 作为系统管理员，我想控制支付方式的启用状态，以便于灵活管理

#### 验收标准

1. **WHEN** 禁用支付方式 **THEN** 系统 **SHALL** 在两个系统中同时禁用
2. **WHEN** 添加支付方式后 **THEN** 系统 **SHALL** 自动将其加入当前公司账号

#### 具体要求

- [ ] 3.1 支持支付方式的启用/禁用状态管理
- [ ] 3.2 状态变更时自动同步到关联系统
- [ ] 3.3 新增支付方式自动关联到当前商家和分支机构
- [ ] 3.4 权限控制：只有管理员能修改支付方式状态

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/database.mdc` - 数据库开发规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/mode_of_payments`）
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
- [ ] ModeOfPayments 表增加 payment_id 字段：varchar(20) not null unique
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
- [ ] 参考: `.cursor/rules/php.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `admin/i18n/` - 国际化配置

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

1. **PaymentID 字段**: 自动生成唯一标识，支持 ERP-TTPOS 关联
2. **双向同步**: ERP ↔ TTPOS 数据自动同步
3. **状态管理**: 支付方式启用/禁用状态同步

### 测试验收

1. **单元测试**: 覆盖率达标，重点测试 PaymentID 生成逻辑
2. **API 测试**: 所有支付方式相关接口测试通过
3. **集成测试**: ERP-TTPOS 同步流程测试通过
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: 支付方式相关 API 文档完整
3. **数据库文档**: ModeOfPayments 表结构变更文档完整
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

- PaymentID 一旦生成不可修改
- 现有旧数据不做处理
- 发票下单仍使用原有 ID，不改变 ERP 取值逻辑
- 支持 ERP-TTPOS 双向同步，但不同步历史数据
- 本次需求仅涉及 PaymentID 字段新增，不包含支付类型扩展

### 资源约束

- 开发时间: 5-7 天
- Story Point: 13 (待技术评审确认)

---

## 依赖关系

### 技术依赖

- `thinkphp/framework: ^6.0` - ThinkPHP 框架
- `predis/predis: ^2.0` - Redis 客户端
- `guzzlehttp/guzzle: ^7.0` - HTTP 客户端（用于同步）

### ERP 迁移脚本

- **位置**: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.12/01_custom_field/01_custom_payment_id.json`
- **说明**: 在 ERPNext 的 Mode of Payment DocType 中添加 custom_payment_id 自定义字段
- **字段配置**:
  - 字段名: `custom_payment_id`
  - 标签: Payment ID
  - 类型: Data (字符串)
  - 插入位置: mode_of_payment 字段之后

### 服务依赖

- **Admin → Main**: HTTP API 调用（同步到 TTPOS）
- **Admin → BMP**: 如需要 gRPC 调用
- **ERP 外部服务**: ERP 系统 API 接口

### 业务依赖

- ERP 系统集成接口
- TTPOS 支付方式管理功能
- 商家和分支机构管理

---

## 风险和缓解

### 风险 1: ERP 系统集成复杂性

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 充分的集成测试和兼容性验证
- 准备回滚方案
- 分阶段实施，先完成基础同步

### 风险 2: 双向同步数据一致性

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 实现事务机制确保数据一致性
- 添加数据校验和冲突解决逻辑
- 建立监控和告警机制

### 风险 3: 第三方支付渠道兼容性

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 充分调研 LianLianPay API
- 准备多种支付渠道的适配方案
- 建立渠道切换机制

---

## 时间表

- **Phase 1 - 需求分析与设计**: 1 天
- **Phase 2 - PaymentID 字段实现**: 2 天
- **Phase 3 - 双向同步机制**: 3 天
- **Phase 4 - 测试与集成**: 1 天
- **总计**: 7 天（SP = 13）

---

## 参考资料

### 核心规范

- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- DooTask #37499 - ERPNext对接-支付方式需求文档
- ERP 系统 API 文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-22  
**作者**: rikugun  
**审核者**: {审核者}
