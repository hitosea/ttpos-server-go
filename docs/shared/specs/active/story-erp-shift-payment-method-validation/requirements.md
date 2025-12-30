# ERP 班次支付方式锁定与验证 需求文档

> 本文档定义 ERP 班次支付方式锁定与验证功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/erp-shift-payment-method-validation.md](../../../../team/proposals/2025-12/erp-shift-payment-method-validation.md) |
| **创建日期**      | 2025-12-30                                                                                                 |
| **负责人**        | 王昱                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

本功能旨在在班次开账时保存当前可用的支付方式配置，并提供检查函数供外部调用，以支持后续的支付方式验证功能。**本次实现仅包含数据保存和检查函数，不包含完整的验证逻辑实现**。

## 🎯 产品对齐

该功能支持 ERP 系统对班次管理的规范要求，为后续的支付方式验证、外卖接单验证等功能提供基础数据支持。

## 📝 用户故事

**作为** 开发人员  
**我想** 在班次开账时保存支付方式列表，并提供检查函数供外部调用  
**以便于** 后续实现支付方式验证、接单验证等功能时能够使用这些数据

---

## 功能需求

### Requirement 1: 班次开账时保存支付方式列表

**用户故事**: 作为 系统，我想 在班次开账时自动保存当前可用的支付方式列表，以便于 后续验证功能使用

#### 验收标准

1. **WHEN** 班次开账时 **AND** 公司开启了 ERP **THEN** 系统 **SHALL** 自动保存当前已启用的支付方式列表到班次记录中

2. **IF** 班次记录中已保存支付方式列表 **THEN** 系统 **SHALL** 能够通过班次记录查询到开账时的支付方式列表

3. **WHEN** 班次开账时 **AND** 公司未开启 ERP **THEN** 系统 **SHALL** 不保存支付方式列表（兼容现有逻辑）

#### 具体要求

- [ ] 1.1 在 `StaffShiftLog` 模型中新增字段用于存储支付方式列表（JSON 格式）
- [ ] 1.2 在 `CreateWorkingLog` 方法中，当公司开启 ERP 时，获取当前已启用的支付方式列表
- [ ] 1.3 将支付方式列表（包含支付方式 UUID、名称、ERP 支付方式标识等关键信息）序列化为 JSON 并保存到班次记录
- [ ] 1.4 支付方式列表应包含：支付方式 UUID、支付方式名称、ERP PaymentId（如有）、ERP ModeOfPayment（如有）
- [ ] 1.5 对于历史班次数据，如果未保存支付方式列表，字段应为空或 nil

---

### Requirement 2: 提供支付方式检查函数

**用户故事**: 作为 开发人员，我想 通过检查函数验证支付方式是否在开账时保存的列表中，以便于 在结账、充值等场景中使用

#### 验收标准

1. **WHEN** 调用检查函数 **AND** 传入班次记录 UUID 和支付方式 UUID **THEN** 系统 **SHALL** 返回该支付方式是否在开账时保存的列表中

2. **IF** 班次记录中未保存支付方式列表（历史数据） **THEN** 系统 **SHALL** 返回 false（不允许使用，正常验证）

3. **WHEN** 调用检查函数 **AND** 传入无效的班次记录 UUID **THEN** 系统 **SHALL** 返回错误

#### 具体要求

- [ ] 2.1 在 `IStaffShiftSrv` 接口中新增方法：`ValidatePaymentMethod(ctx context.Context, shiftNo string, paymentMethodUuid uint64) (bool, error)`
- [ ] 2.2 方法实现逻辑：
  - 根据 `shiftNo`（交班编号）查询班次记录
  - 如果班次记录不存在，返回错误
  - 如果班次记录中未保存支付方式列表（历史数据），返回 `false`（不允许使用，正常验证）
  - 如果班次记录中保存了支付方式列表，检查传入的 `paymentMethodUuid` 是否在列表中
  - 返回检查结果（true/false）和错误（如有）
- [ ] 2.3 方法应支持通过支付方式 UUID 进行匹配
- [ ] 2.4 方法应提供清晰的错误信息，便于调用方处理

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

**注意**: 本次实现不涉及 API 接口，仅涉及 Service 层方法。

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] 金额字段使用 decimal(20,8)
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 新增字段需创建数据库迁移脚本
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（如需要）
- [ ] 并发处理（使用 UUID 锁）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

**注意**: 本次实现不涉及用户可见的提示信息，暂不需要国际化。

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

1. **支付方式保存**: 班次开账时，系统能够正确保存当前已启用的支付方式列表到班次记录中
2. **检查函数**: 提供的检查函数能够正确验证支付方式是否在开账时保存的列表中
3. **历史数据验证**: 对于历史班次数据（未保存支付方式列表），检查函数返回 false（不允许使用）

### 测试验收

1. **单元测试**: 覆盖率达标
2. **集成测试**: 端到端流程测试通过（班次开账 → 保存支付方式 → 检查函数调用）
3. **手动测试**: 验证历史数据兼容性

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: 检查函数的注释和文档完整
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 仅当公司开启 ERP 时才保存支付方式列表
- 历史班次数据正常验证（未保存支付方式列表时，检查函数返回 false，不允许使用）
- 支付方式列表保存时机：班次开账确认时

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/service/staff_shift.go` - 班次服务（需要修改）
- `main/app/model/staff.go` - 班次模型（需要修改）
- `main/app/repository/shift_log.go` - 班次仓储（可能需要修改）
- `main/app/service/payment_method.go` - 支付方式服务（需要调用）

### 服务依赖

- **Main → BMP**: 无（本次不涉及）
- **Admin → Main**: 无（本次不涉及）
- **Frontend → Admin**: 无（本次不涉及）

### 业务依赖

- 班次开账功能（已存在）
- 支付方式管理功能（已存在）
- ERP 集成功能（已存在）

---

## 风险和缓解

### 风险 1: 数据库迁移影响现有数据

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 新增字段设置为可空或提供默认值
- 迁移脚本需要测试验证
- 对于历史数据，检查函数返回 false（正常验证，不允许使用）

### 风险 2: 支付方式列表格式变更

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 使用 JSON 格式存储，便于扩展
- 定义清晰的数据结构
- 提供版本号或格式标识（如需要）

### 风险 3: 性能影响

**影响**: 低  
**概率**: 低  
**缓解措施**:

- JSON 数据量较小（支付方式列表通常不超过 20 项）
- 检查函数使用索引查询班次记录
- 如需要，可考虑缓存优化

---

## 时间表

- **Phase 1 - 数据库设计**: 0.5 天
  - 设计字段结构
  - 创建迁移脚本
- **Phase 2 - 保存逻辑实现**: 1 天
  - 修改 `CreateWorkingLog` 方法
  - 实现支付方式列表保存逻辑
- **Phase 3 - 检查函数实现**: 0.5 天
  - 实现 `ValidatePaymentMethod` 方法
  - 编写单元测试
- **Phase 4 - 测试和文档**: 0.5 天
  - 集成测试
  - 更新文档
- **总计**: 2.5 天（SP = 3）

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

### 相关代码

- `main/app/service/staff_shift.go` - 班次服务实现
- `main/app/model/staff.go` - 班次模型定义
- `main/app/service/payment_method.go` - 支付方式服务
- `main/app/repository/shift_log.go` - 班次仓储实现

### 外部参考

- [ERPNext POS Entry 文档](https://docs.erpnext.com/docs/user/manual/en/point-of-sales)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-30  
**作者**: 王昱  
**审核者**: {审核者}

