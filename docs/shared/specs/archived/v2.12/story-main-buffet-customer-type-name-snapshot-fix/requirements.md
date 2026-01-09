> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 自助餐顾客类型名称快照修复 需求文档

> 本文档定义自助餐顾客类型名称快照修复功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/buffet-customer-type-name-snapshot-fix.md](../../../../team/proposals/2025-12/buffet-customer-type-name-snapshot-fix.md) |
| **创建日期**      | 2025-12-09                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

当前订单查询时，`ttpos_sale_order_buffet_customer_type` 表中的 `name` 字段（顾客类型名称）会随后台数据变更而改变，导致订单历史信息不准确。本功能将 `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段从 `VARCHAR(255)` 修改为 `TEXT` 类型，保存顾客类型名称的多语言 JSON 快照，确保订单历史信息准确反映下单时的真实状态，满足财务、税务对订单历史记录的合规性要求。

**核心价值**：
- 确保 `SaleOrderBuffetCustomerType` 记录的顾客类型名称准确反映下单时的状态
- 支持多语言快照，满足国际化需求
- 满足财务、税务对订单历史记录的合规性要求
- 支持订单历史查询和问题追溯
- 避免因数据变更导致的业务逻辑错误

## 🎯 产品对齐

本功能是"订单商品信息快照修复"（`order-attribute-snapshot-fix.md`）系列需求的一部分，通过建立完整的订单快照机制，确保：
- **数据一致性**：订单信息作为历史快照，不随后台配置变更而改变
- **多语言支持**：支持多语言快照，满足国际化需求
- **业务可靠性**：支持订单对账、报表、审计等关键业务场景
- **合规性**：满足餐饮行业对历史订单数据的监管要求

## 📝 用户故事

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实顾客类型名称  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的顾客类型名称  
**以便于** 准确处理退款和客户咨询

---

## 功能需求

### Requirement 1: 数据库结构变更 - 修改 name 字段类型为 TEXT

**用户故事**: 作为开发者，我想将 `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段从 `VARCHAR(255)` 修改为 `TEXT` 类型，以便于保存多语言 JSON 快照。

#### 验收标准

1. **WHEN** 执行数据库迁移脚本 **THEN** 系统 **SHALL** 将 `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段类型修改为 `TEXT`
2. **WHEN** 修改字段类型后 **THEN** 系统 **SHALL** 不影响现有数据和业务功能
3. **WHEN** 执行迁移 **THEN** 系统 **SHALL** 支持可重复执行（幂等性）

#### 具体要求

- [ ] 1.1 创建数据库迁移文件，将 `name` 字段类型从 `VARCHAR(255)` 修改为 `TEXT`
- [ ] 1.2 迁移脚本支持可重复执行（幂等性）
- [ ] 1.3 字段注释更新为："顾客类型名称快照（JSON），不随后台更新"

---

### Requirement 2: 数据模型修改 - 修改 Name 字段类型

**用户故事**: 作为开发者，我想修改 `SaleOrderBuffetCustomerType` 模型中的 `Name` 字段，以便于在代码中操作多语言 JSON 快照数据。

#### 验收标准

1. **WHEN** 修改 `SaleOrderBuffetCustomerType` 结构体 **THEN** 系统 **SHALL** 将 `Name` 字段的 GORM 标签修改为 `type:text`
2. **WHEN** 字段定义完成 **THEN** GORM 标签 **SHALL** 正确映射到数据库字段 `name`
3. **WHEN** 编译代码 **THEN** 系统 **SHALL** 无编译错误

#### 具体要求

- [ ] 2.1 修改 `SaleOrderBuffetCustomerType` 结构体中的 `Name` 字段，GORM 标签改为 `type:text`
- [ ] 2.2 保持 JSON 标签为 `json:"name"`
- [ ] 2.3 字段注释更新为："顾客类型名称快照（JSON）"

---

### Requirement 3: 查询逻辑修改 - 实现顾客类型名称获取方法（JSON 方案）

**用户故事**: 作为开发者，我想实现顾客类型名称获取方法，以便于优先使用快照数据，降级使用关联表数据。

#### 验收标准

1. **WHEN** 调用 `GetLocaleName()` 方法 **THEN** 系统 **SHALL** 优先返回快照字段 `Name`（JSON 格式）
2. **IF** 快照字段为空或 JSON 解析失败 **THEN** 系统 **SHALL** 降级使用 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 关联表数据
3. **WHEN** 返回数据 **THEN** 格式 **SHALL** 为 `dto.LocaleResponse`（多语言）
4. **IF** 快照字段有值（JSON）**THEN** 系统 **SHALL** 解析 JSON 并返回完整多语言数据（所有语言）

#### 具体要求

- [ ] 3.1 在 `SaleOrderBuffetCustomerType` 实现 `GetLocaleName()` 方法
- [ ] 3.2 实现快照优先逻辑：优先使用 `Name` 字段（JSON）
- [ ] 3.3 实现 JSON 解析：将快照字段反序列化为 `LocaleResponse`
- [ ] 3.4 实现降级逻辑：快照为空或解析失败时使用关联表数据（`BuffetCustomerTypePrice.BuffetCustomerType.Name`）
- [ ] 3.5 方法返回类型：`dto.LocaleResponse`
- [ ] 3.6 添加方法注释，说明逻辑和用途

---

### Requirement 4: 下单逻辑修改 - 实现快照保存方法（JSON 方案）

**用户故事**: 作为开发者，我想实现快照保存方法，以便于在创建 `SaleOrderBuffetCustomerType` 时保存快照字段。

#### 验收标准

1. **WHEN** 调用 `SetNameSnapshot()` 方法 **THEN** 系统 **SHALL** 正确保存快照字段（`Name`）
2. **WHEN** 保存快照字段 **THEN** 数据格式 **SHALL** 为 JSON 字符串（包含所有语言）
3. **WHEN** 序列化多语言数据 **THEN** 系统 **SHALL** 从 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 获取名称并转换为多语言 JSON
4. **IF** 序列化失败 **THEN** 系统 **SHALL** 记录错误日志但不中断流程

#### 具体要求

- [ ] 4.1 在 `SaleOrderBuffetCustomerType` 实现 `SetNameSnapshot()` 方法
- [ ] 4.2 实现序列化逻辑：从 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 获取名称
- [ ] 4.3 实现 JSON 序列化：将名称转换为多语言 JSON 格式（所有语言使用相同值）
- [ ] 4.4 实现错误处理：序列化失败时记录日志但不中断流程
- [ ] 4.5 方法参数：`customerTypeName string` 类型
- [ ] 4.6 添加方法注释，说明逻辑和用途

---

### Requirement 5: 查询逻辑修改 - 替换现有查询方法

**用户故事**: 作为开发者，我想替换所有使用 `orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 的地方，以便于使用 `SaleOrderBuffetCustomerType` 自己的快照方法。

#### 验收标准

1. **WHEN** 查询订单详情 **THEN** 系统 **SHALL** 使用 `SaleOrderBuffetCustomerType.GetLocaleName()` 方法
2. **WHEN** 替换查询方法 **THEN** 系统 **SHALL** 保持原有功能不变
3. **WHEN** 返回数据 **THEN** 格式 **SHALL** 为 `dto.LocaleResponse`（多语言）

#### 具体要求

- [ ] 5.1 修改 `GetOrderInfos()` 方法，使用 `orderBuffetCustomer.GetLocaleName()`
- [ ] 5.2 修改 `checkBuffetCustomerTypePriceChanged()` 方法，使用 `buffetCustomer.GetLocaleName()`
- [ ] 5.3 修改订单导出相关方法，使用 `orderBuffetCustomer.GetLocaleName()`
- [ ] 5.4 修改其他使用 `SaleOrderBuffetCustomerType` 的地方，统一使用快照方法
- [ ] 5.5 确保所有修改不影响现有功能

---

### Requirement 6: 下单逻辑修改 - 保存顾客类型名称快照（JSON 方案）

**用户故事**: 作为开发者，我想在创建 `SaleOrderBuffetCustomerType` 时保存快照字段，以便于新订单自动使用快照机制。

#### 验收标准

1. **WHEN** 创建新 `SaleOrderBuffetCustomerType` **THEN** 系统 **SHALL** 正确保存快照字段（`Name`）
2. **WHEN** 保存快照字段 **THEN** 数据格式 **SHALL** 为 JSON 字符串（包含所有语言）
3. **WHEN** 序列化多语言数据 **THEN** 系统 **SHALL** 从 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 获取名称并转换为多语言 JSON
4. **IF** 序列化失败 **THEN** 系统 **SHALL** 记录错误日志但不中断下单流程

#### 具体要求

- [ ] 6.1 修改 `NewSaleOrderBuffetCustomerType()` 方法，保存顾客类型名称快照
- [ ] 6.2 修改 `SaleOrder.GetSaleOrderBuffetCustomerTypes()` 方法，创建 `SaleOrderBuffetCustomerType` 时保存快照
- [ ] 6.3 修改 `CreateDeskOrder` 下单逻辑，确保保存快照
- [ ] 6.4 修改 `OrderChangeBuffet` 修改逻辑，确保保存快照
- [ ] 6.5 确保所有创建 `SaleOrderBuffetCustomerType` 的地方都保存快照

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
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
- [ ] 优先使用快照数据，减少关联查询

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 快照字段保存完整的多语言 JSON（`dto.LocaleResponse`）
- [ ] 查询时返回多语言格式
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
- [ ] 快照字段为空时降级使用关联表数据（兼容性）

---

## 验收标准

### 功能验收

1. **数据库迁移**: 迁移脚本执行成功，字段类型修改正确
2. **快照保存**: 创建 `SaleOrderBuffetCustomerType` 时正确保存快照字段
3. **快照查询**: 查询订单时优先使用快照字段，降级使用关联表数据
4. **多语言支持**: 快照字段保存和返回完整的多语言数据
5. **兼容性**: 历史订单的快照字段为空时，降级使用关联表数据正常显示
6. **数据一致性**: 顾客类型被删除或改名后，历史订单仍显示原始名称

### 测试验收

1. **单元测试**: 覆盖率达标，覆盖所有修改的方法
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过（下单、查询、修改）
4. **回归测试**: 确保不影响现有功能（订单查询、打印、导出、退款等）

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **数据库文档**: 迁移脚本和表结构文档完整
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

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

- 快照字段保存多语言 JSON，与自助餐套餐名称快照修复方案保持一致
- 历史订单的 `name` 字段是单语言文本，需要降级处理
- 新订单必须保存快照字段
- 注意：`BuffetCustomerType.Name` 是单语言字段，需要转换为多语言 JSON 格式（所有语言使用相同值）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `encoding/json` - JSON 序列化/反序列化
- `ttpos-server-go/app/dto` - `LocaleResponse` 类型
- `ttpos-server-go/app/model` - `SaleOrderBuffetCustomerType` 类型

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖订单商品信息快照修复系列需求（`order-attribute-snapshot-fix.md`）
- 依赖自助餐顾客类型套餐名称快照修复方案（`story-main-buffet-customer-type-package-name-snapshot-fix`）

---

## 风险和缓解

### 风险 1: 数据库结构变更风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用 `ALTER TABLE MODIFY COLUMN` 修改字段类型，不影响现有数据（VARCHAR 可以转换为 TEXT）
- 迁移脚本支持可重复执行（幂等性）
- 在测试环境充分验证后再上线

### 风险 2: 历史数据不完整

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 实现降级逻辑，确保历史订单正常显示
- 逐步迁移，不强制要求所有数据立即完整
- 可选：提供数据迁移脚本，从关联表补充历史订单的快照字段

### 风险 3: 多语言支持问题

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 采用 JSON 格式保存多语言数据（`dto.LocaleResponse`）
- 快照字段保存完整的多语言 JSON（所有语言使用相同值）
- 查询时优先使用快照字段，如果快照字段为空或无效，降级使用关联表数据

### 风险 4: 下单逻辑修改风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 全面梳理所有创建 `SaleOrderBuffetCustomerType` 的地方
- 确保所有下单入口都保存快照数据
- 编写单元测试和集成测试覆盖所有场景

### 风险 5: 回归风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 充分测试，特别是订单相关的所有功能
- 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
- 在测试环境充分验证后再上线

---

## 时间表

- **Phase 1 - 数据库结构变更**: 0.5 天
- **Phase 2 - Model 层修改**: 0.5 天
- **Phase 3 - 查询逻辑修改**: 0.5 天
- **Phase 4 - 下单逻辑修改**: 0.5 天
- **Phase 5 - 测试验证**: 0.5 天
- **总计**: 2.5 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- 自助餐顾客类型套餐名称快照修复 Spec: `docs/shared/specs/active/story-main-buffet-customer-type-package-name-snapshot-fix/`
- 商品属性信息快照修复 Spec: `docs/shared/specs/active/story-main-product-attribute-snapshot-fix/`
- 原因信息快照修复 Spec: `docs/shared/specs/active/story-main-reason-snapshot-fix/`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}


