> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 商品属性信息快照修复 需求文档

> 本文档定义商品属性信息快照修复功能的详细需求和验收标准。  
> 本功能聚焦于商品名称、规格名称、小料名称、属性名称快照修复（使用现有快照字段，无需数据库变更）。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/product-attribute-snapshot-fix.md](../../../../team/proposals/2025-12/product-attribute-snapshot-fix.md) |
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

当前订单查询时，商品名称、规格名称、小料名称、属性名称信息会随后台数据变更而改变，导致订单历史信息不准确。虽然数据库表已有快照字段（`SaleOrderProduct.Name`、`SaleOrderProduct.FlavorName`、`SaleOrderProductBom.Name`、`SaleOrderProductAttribute.Name`），但代码实现未使用这些字段，而是从关联表实时获取。

本功能将修复查询逻辑，优先使用现有快照字段，确保订单历史信息准确反映下单时的真实状态，满足财务、税务对订单历史记录的合规性要求。

**核心价值**：
- 确保订单商品信息准确反映下单时的状态
- 满足财务、税务对订单历史记录的合规性要求
- 支持订单历史查询和问题追溯
- 避免因数据变更导致的业务逻辑错误
- **无需数据库变更**：使用现有快照字段，降低实施风险

## 🎯 产品对齐

本功能是"订单商品信息快照修复"（`order-attribute-snapshot-fix.md`）系列需求的一部分，通过修复查询逻辑使用现有快照字段，确保：
- **数据一致性**：订单信息作为历史快照，不随后台配置变更而改变
- **业务可靠性**：支持订单对账、报表、审计等关键业务场景
- **合规性**：满足餐饮行业对历史订单数据的监管要求
- **实施效率**：无需数据库变更，降低实施风险和复杂度

## 📝 用户故事

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实商品信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的商品属性  
**以便于** 准确处理退款和客户咨询

---

## 功能需求

### Requirement 1: 修复商品名称获取逻辑

**用户故事**: 作为开发者，我想修复商品名称获取逻辑，优先使用快照字段，以便于确保订单历史信息准确。

#### 验收标准

1. **WHEN** 查询订单商品名称 **THEN** 系统 **SHALL** 优先使用 `SaleOrderProduct.Name` 字段
2. **IF** `SaleOrderProduct.Name` 为空（历史数据） **THEN** 系统 **SHALL** 降级使用 `MultiLanguageName` 关联对象
3. **WHEN** 返回商品名称 **THEN** 系统 **SHALL** 返回多语言格式（`dto.LocaleResponse`）
4. **IF** 快照字段有值 **THEN** 系统 **SHALL** 使用快照字段填充主语言（ZH），其他语言从关联表补充
5. **IF** 关联表数据不存在（已删除） **THEN** 系统 **SHALL** 使用快照的主语言填充所有语言字段

#### 具体要求

- [ ] 1.1 修改 `SaleOrderProduct.MultiLanguageName.GetNames()` 相关方法，优先使用 `SaleOrderProduct.Name` 字段
- [ ] 1.2 实现多语言响应构建逻辑：主语言使用快照，其他语言使用关联表（如果存在）
- [ ] 1.3 添加兼容性处理：快照为空时降级使用关联表数据
- [ ] 1.4 确保所有使用商品名称的地方都使用修复后的逻辑

---

### Requirement 2: 修复规格名称获取逻辑

**用户故事**: 作为开发者，我想修复规格名称获取逻辑，优先使用快照字段，以便于确保订单历史信息准确。

#### 验收标准

1. **WHEN** 查询订单规格名称 **THEN** 系统 **SHALL** 优先使用 `SaleOrderProduct.FlavorName` 或 `SaleOrderProductBom.Name` 字段
2. **IF** 快照字段为空（历史数据） **THEN** 系统 **SHALL** 降级使用 `ProductBom.ProductFlavor.MultiLanguageName`
3. **WHEN** 返回规格名称 **THEN** 系统 **SHALL** 返回多语言格式（`dto.LocaleResponse`）
4. **IF** 快照字段有值 **THEN** 系统 **SHALL** 使用快照字段填充主语言（ZH），其他语言从关联表补充

#### 具体要求

- [ ] 2.1 修改 `SaleOrderProduct.GetFlavorName()` 方法，优先使用 `SaleOrderProduct.FlavorName` 或 `SaleOrderProductBom.Name` 字段
- [ ] 2.2 修改 `SaleOrderProduct.GetFlavorSaleOrderProductBom()` 方法，确保返回的快照数据优先
- [ ] 2.3 实现多语言响应构建逻辑：主语言使用快照，其他语言使用关联表（如果存在）
- [ ] 2.4 添加兼容性处理：快照为空时降级使用关联表数据

---

### Requirement 3: 修复小料名称获取逻辑

**用户故事**: 作为开发者，我想修复小料名称获取逻辑，优先使用快照字段，以便于确保订单历史信息准确。

#### 验收标准

1. **WHEN** 查询订单小料名称 **THEN** 系统 **SHALL** 优先使用 `SaleOrderProductBom.Name` 字段
2. **IF** `SaleOrderProductBom.Name` 为空（历史数据） **THEN** 系统 **SHALL** 降级使用 `ProductBom.ProductSauce.MultiLanguageName`
3. **WHEN** 返回小料名称 **THEN** 系统 **SHALL** 返回多语言格式（`dto.LocaleResponse`）
4. **IF** 快照字段有值 **THEN** 系统 **SHALL** 使用快照字段填充主语言（ZH），其他语言从关联表补充

#### 具体要求

- [ ] 3.1 修改 `SaleOrderProduct.GetSauceNamesList()` 方法，优先使用 `SaleOrderProductBom.Name` 字段
- [ ] 3.2 修改 `SaleOrderProduct.GetSauceSaleOrderProductBom()` 方法，确保返回的快照数据优先
- [ ] 3.3 实现多语言响应构建逻辑：主语言使用快照，其他语言使用关联表（如果存在）
- [ ] 3.4 添加兼容性处理：快照为空时降级使用关联表数据

---

### Requirement 4: 修复属性名称获取逻辑

**用户故事**: 作为开发者，我想修复属性名称获取逻辑，优先使用快照字段，以便于确保订单历史信息准确。

#### 验收标准

1. **WHEN** 查询订单属性名称 **THEN** 系统 **SHALL** 优先使用 `SaleOrderProductAttribute.Name` 字段
2. **IF** `SaleOrderProductAttribute.Name` 为空（历史数据） **THEN** 系统 **SHALL** 降级使用 `ProductAttribute.MultiLanguageName`
3. **WHEN** 返回属性名称 **THEN** 系统 **SHALL** 返回多语言格式（`dto.LocaleResponse`）
4. **IF** 快照字段有值 **THEN** 系统 **SHALL** 使用快照字段填充主语言（ZH），其他语言从关联表补充

#### 具体要求

- [ ] 4.1 修改 `SaleOrderProduct.GetAttributeName()` 方法，优先使用 `SaleOrderProductAttribute.Name` 字段
- [ ] 4.2 修改 `SaleOrderProduct.GetAttributeNameList()` 方法，优先使用快照字段
- [ ] 4.3 修改 `SaleOrderProduct.GetPureAttributeNameList()` 方法，优先使用快照字段
- [ ] 4.4 修改 `SaleOrderProduct.GetAttributeNamesByLang()` 方法，优先使用快照字段
- [ ] 4.5 实现多语言响应构建逻辑：主语言使用快照，其他语言使用关联表（如果存在）
- [ ] 4.6 添加兼容性处理：快照为空时降级使用关联表数据

---

### Requirement 5: 修复商品名称属性组合方法

**用户故事**: 作为开发者，我想修复商品名称属性组合方法，确保使用快照数据，以便于确保订单历史信息准确。

#### 验收标准

1. **WHEN** 调用 `GetProductNameAttributes()` 方法 **THEN** 系统 **SHALL** 使用快照数据（商品名称、规格、属性）
2. **WHEN** 调用 `GetNameAndFlavorName()` 方法 **THEN** 系统 **SHALL** 使用快照数据（商品名称、规格）
3. **IF** 快照字段为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据

#### 具体要求

- [ ] 5.1 修改 `SaleOrderProduct.GetProductNameAttributes()` 方法，确保使用快照数据
- [ ] 5.2 修改 `SaleOrderProduct.GetNameAndFlavorName()` 方法，确保使用快照数据
- [ ] 5.3 确保组合方法内部调用已修复的单个方法（商品名称、规格、属性）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Model 层方法修改，不涉及 Service/Repository 层变更
- **单一职责原则**: 每个方法应有单一、明确的目的
- **模块化设计**: 方法应独立且可复用
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 所有返回商品信息的接口应返回多语言格式（`dto.LocaleResponse`）
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] **无需数据库变更**：使用现有快照字段（`SaleOrderProduct.Name`、`SaleOrderProduct.FlavorName`、`SaleOrderProductBom.Name`、`SaleOrderProductAttribute.Name`）
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 优先使用快照字段，减少关联查询
- [ ] 如果必须降级查询，使用索引优化
- [ ] 本地响应时间 < 200ms

### 测试要求

- [ ] Model 层方法测试覆盖率 ≥ 80%
- [ ] 单元测试覆盖所有修改的方法和场景（快照存在/不存在、关联表存在/不存在）
- [ ] 集成测试覆盖订单查询接口（验证快照逻辑）
- [ ] 回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 采用"主语言快照 + 关联表补充"的多语言方案
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] SQL 注入防护（使用参数化查询）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级（快照为空时使用关联表）
- [ ] 错误日志记录（使用 Logger）
- [ ] 兼容性处理（历史数据通过降级逻辑正常显示）

---

## 验收标准

### 功能验收

1. **商品名称快照**: 查询订单时，商品名称优先使用快照字段，如果为空则降级使用关联表
2. **规格名称快照**: 查询订单时，规格名称优先使用快照字段，如果为空则降级使用关联表
3. **小料名称快照**: 查询订单时，小料名称优先使用快照字段，如果为空则降级使用关联表
4. **属性名称快照**: 查询订单时，属性名称优先使用快照字段，如果为空则降级使用关联表
5. **多语言支持**: 所有名称返回多语言格式，主语言使用快照，其他语言从关联表补充（如果存在）
6. **兼容性**: 历史订单（快照字段为空）通过降级逻辑正常显示
7. **数据删除场景**: 后台删除商品/规格/小料/属性后，历史订单仍能显示原始名称

### 测试验收

1. **单元测试**: 覆盖率达标，覆盖所有修改的方法和场景
2. **集成测试**: 订单查询接口测试通过
3. **回归测试**: 订单查询、打印、导出、退款等功能测试通过
4. **手动测试**: 多语言场景测试通过（关联表存在/不存在）

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- Model 层方法修改，不涉及 Service/Repository 层
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc`

### 业务约束

- **无需数据库变更**：使用现有快照字段
- **渐进式实施**：不需要强制迁移所有历史数据
- **兼容性优先**：历史订单通过降级逻辑兼容（快照为空时使用关联表）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (待技术评审确认)

---

## 依赖关系

### 技术依赖

- `gorm.io/gorm` - ORM 框架
- `main/app/dto` - DTO 定义（`LocaleResponse`）
- `main/app/model` - 数据模型（`SaleOrderProduct`, `SaleOrderProductBom`, `SaleOrderProductAttribute`, `MultiLanguageName`）

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖 `MultiLanguageName`、`ProductBom`、`ProductAttribute` 模型
- 依赖订单查询流程
- **无需修改下单逻辑**：下单时已保存快照字段

---

## 风险和缓解

### 风险 1: 历史数据不完整

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 实现降级逻辑，快照为空时使用关联表数据
- 渐进式实施，不强制要求所有历史数据立即完整
- 历史订单通过降级逻辑正常显示

### 风险 2: 多语言支持问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 采用"主语言快照 + 关联表补充"的混合方案
- 快照字段保存主语言（中文），其他语言从关联表补充
- 关联表不存在时，所有语言使用快照的主语言填充
- 充分测试多语言场景（关联表存在/不存在）

### 风险 3: 回归风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 编写单元测试覆盖所有修改的方法和场景
- 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
- 在测试环境充分验证后再上线
- 准备回滚方案

### 风险 4: 数据一致性

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 全面梳理所有使用商品名称、规格、小料、属性的地方
- 确保所有相关方法都使用快照数据
- 代码审查确保修改完整

---

## 时间表

- **Phase 1 - 商品名称修复**: 0.5 天
  - 修改商品名称获取方法
  - 添加兼容性处理和多语言支持
- **Phase 2 - 规格名称修复**: 0.5 天
  - 修改规格名称获取方法
  - 添加兼容性处理和多语言支持
- **Phase 3 - 小料名称修复**: 0.5 天
  - 修改小料名称获取方法
  - 添加兼容性处理和多语言支持
- **Phase 4 - 属性名称修复**: 0.5-1 天
  - 修改属性名称获取方法
  - 修改组合方法
  - 添加兼容性处理和多语言支持
- **Phase 5 - 测试验证**: 0.5-1 天
  - 单元测试（覆盖所有修改的方法）
  - 集成测试（验证订单查询）
  - 回归测试（确保不影响现有功能）
- **总计**: 2.5-3 天（SP = 3-5）

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

### 相关文档

- `docs/team/proposals/2025-01/order-attribute-snapshot-fix.md` - 父提案（订单商品信息快照修复）
- `docs/team/proposals/2025-12/product-attribute-snapshot-fix.md` - 本提案
- `docs/shared/api/cashier-order-info-analysis.md` - 订单信息获取逻辑分析
- `main/app/model/sale_order_product.go` - SaleOrderProduct 模型定义
- `main/app/model/order.go` - SaleOrderProductBom、SaleOrderProductAttribute 模型定义

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2025-12/2025-12-09.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}


