> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 订单直接添加商品领域服务 需求文档

> 本文档定义订单直接添加商品领域服务的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [订单直接添加商品领域服务需求提案](../../../../team/proposals/2025-12/order-direct-add-products-domain-service.md) |
| **创建日期**      | 2025-12-19                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核（⚠️ 注意：设计文档已创建，等待需求审核通过） |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

设计一个专门的**订单直接添加商品领域服务（OrderDirectAddProductsDomainService）**，专注于**直接向订单添加指定的商品/实体到数据库**。

该服务将提供一个统一的接口，支持直接向订单添加普通商品、套餐、自助餐顾客、自助餐加钟等多种类型的商品/实体，并确保数据写入的事务性和一致性。

**核心职责**：
- ✅ **数据写入**：直接将商品/实体数据写入数据库，不进行业务规则验证
- ✅ **多类型支持**：支持添加普通商品、套餐、自助餐顾客、自助餐加钟等多种类型
- ✅ **批量添加**：支持一次调用添加多个不同类型的商品/实体
- ❌ **不处理业务规则验证**：库存验证、限购检查、超时检查等由调用方处理
- ❌ **不处理价格计算**：价格计算由订单实体或应用服务处理

## 🎯 产品对齐

通过设计专门的订单直接添加商品领域服务，可以带来以下价值：

- **简化调用**：提供一个统一的接口，直接向订单添加商品/实体，无需关心底层实现
- **提高开发效率**：开发人员只需调用一个方法，即可添加多种类型的商品/实体
- **增强可测试性**：领域服务可以独立测试数据写入逻辑，mock 仓储层即可
- **降低维护成本**：数据写入逻辑集中管理，修改影响范围可控
- **保证数据一致性**：统一的事务管理，确保所有相关表的数据写入原子性
- **支持 DDD 演进**：符合项目 DDD 架构演进方向，为订单模块重构奠定基础

## 📝 用户故事

**作为** 开发人员  
**我想** 通过一个简单的方法直接向订单添加指定的商品/实体（普通商品、套餐、自助餐顾客、自助餐加钟等）  
**以便于** 简化代码调用，提高开发效率和可测试性

**作为** 测试人员  
**我想** 能够直接向订单添加测试数据  
**以便于** 快速准备测试场景，提高测试效率

**作为** 运维人员  
**我想** 能够直接向订单添加数据  
**以便于** 进行数据修复和迁移操作

---

## 功能需求

### Requirement 1: 直接添加商品/实体核心功能

**用户故事**: 作为开发人员，我想通过一个统一的方法直接向订单添加商品/实体，以便于简化代码调用和提高开发效率

#### 验收标准

1. **WHEN** 调用领域服务添加普通商品 **THEN** 系统 **SHALL** 将商品数据写入 `sale_order_product` 及相关表
2. **WHEN** 调用领域服务添加套餐商品 **THEN** 系统 **SHALL** 将套餐数据写入 `sale_order_product` 及相关表
3. **WHEN** 调用领域服务添加自助餐顾客 **THEN** 系统 **SHALL** 将顾客数据写入 `sale_order_buffet_customer_type` 表
4. **WHEN** 调用领域服务添加自助餐加钟 **THEN** 系统 **SHALL** 将加钟数据写入 `sale_order_buffet_delay_product` 表
5. **WHEN** 调用领域服务混合添加多种类型（如：普通商品 + 自助餐顾客 + 自助餐加钟 + 套餐） **THEN** 系统 **SHALL** 将所有类型的数据写入对应的表，并在同一事务中完成
6. **IF** 数据写入失败 **THEN** 系统 **SHALL** 回滚所有相关表的数据，保证数据一致性
7. **IF** products 参数为空 **THEN** 系统 **SHALL** 返回错误，提示至少需要提供一个商品/实体

#### 具体要求

- [ ] 1.1 实现 `AddProductsToOrder` 方法，支持直接向订单添加商品/实体
- [ ] 1.2 支持多种类型的商品/实体添加：
  - ✅ **普通商品**：`SaleOrderProduct`（包含 BOM、属性、备注原因等）
  - ✅ **套餐商品**：`SaleOrderProduct`（套餐类型，可能包含子商品）
  - ✅ **自助餐顾客**：`SaleOrderBuffetCustomerType`（自助餐场景）
  - ✅ **自助餐加钟**：`SaleOrderBuffetDelayProduct`（自助餐场景）
- [ ] 1.3 支持批量添加：一次调用可以添加多个不同类型的商品/实体
- [ ] 1.4 所有表的数据写入在同一事务中完成，使用 `repository.CommonRepo.Transaction` 确保原子性
- [ ] 1.5 写入失败时自动回滚所有相关表的数据
- [ ] 1.6 支持写入以下核心表：
  - `ttpos_sale_order_product` - 订单商品（普通商品/套餐时写入）
  - `ttpos_sale_order_product_bom` - 商品BOM（规格、加料，如商品包含 BOM 则写入）
  - `ttpos_sale_order_product_attribute` - 商品属性（如商品包含属性则写入）
  - `ttpos_sale_order_product_reason` - 商品备注原因（如商品包含备注原因则写入）
  - `ttpos_sale_order_buffet_customer_type` - 自助餐顾客类型（添加自助餐顾客时写入）
  - `ttpos_sale_order_buffet_delay_product` - 自助餐加钟商品（添加自助餐加钟时写入）
  - `ttpos_sale_order_operation_record` - 操作记录（记录添加操作，所有场景必写）

---

### Requirement 2: 多类型商品/实体支持

**用户故事**: 作为开发人员，我想通过领域服务添加不同类型的商品/实体（普通商品、套餐、自助餐顾客、自助餐加钟等），以便于统一处理各种添加场景

#### 验收标准

1. **WHEN** 调用领域服务添加普通商品 **THEN** 系统 **SHALL** 识别商品类型并写入对应的表
2. **WHEN** 调用领域服务添加套餐商品 **THEN** 系统 **SHALL** 识别套餐类型并写入对应的表
3. **WHEN** 调用领域服务添加自助餐顾客 **THEN** 系统 **SHALL** 识别自助餐顾客类型并写入对应的表
4. **WHEN** 调用领域服务添加自助餐加钟 **THEN** 系统 **SHALL** 识别自助餐加钟类型并写入对应的表
5. **WHEN** 调用领域服务混合添加多种类型 **THEN** 系统 **SHALL** 正确识别每种类型并写入对应的表

#### 具体要求

- [ ] 2.1 `AddToOrderProduct` 结构体支持以下字段：
  - `Type ProductType` - 商品类型（Normal, Package, BuffetCustomer, BuffetDelay）
  - `Product *SaleOrderProduct` - 普通商品/套餐（Type 为 Normal 或 Package 时使用）
  - `BuffetCustomer *SaleOrderBuffetCustomerType` - 自助餐顾客（Type 为 BuffetCustomer 时使用）
  - `BuffetDelay *SaleOrderBuffetDelayProduct` - 自助餐加钟（Type 为 BuffetDelay 时使用）
- [ ] 2.2 领域服务能够根据 `Type` 字段识别数据类型，并写入对应的表：
  - `ProductTypeNormal` / `ProductTypePackage` → `sale_order_product`、`sale_order_product_bom`、`sale_order_product_attribute`、`sale_order_product_reason`
  - `ProductTypeBuffetCustomer` → `sale_order_buffet_customer_type`
  - `ProductTypeBuffetDelay` → `sale_order_buffet_delay_product`
- [ ] 2.3 验证 `Type` 字段与对应数据字段的一致性，不一致时返回错误
- [ ] 2.4 支持批量添加：一次调用可以添加多个不同类型的商品/实体

---

### Requirement 3: 场景适配功能

**用户故事**: 作为开发人员，我想根据不同场景（POS、H5、自助餐等）设置不同的字段，以便于支持多样化的业务场景

#### 验收标准

1. **WHEN** POS 端添加商品 **THEN** 系统 **SHALL** 设置 `is_accept_order` 为已接单
2. **WHEN** H5 端添加商品 **THEN** 系统 **SHALL** 设置 `is_accept_order` 为未接单
3. **WHEN** 自助餐场景添加商品 **THEN** 系统 **SHALL** 写入自助餐相关表（`sale_order_buffet_customer_type`、`sale_order_buffet_delay_product` 等）
4. **IF** 商品包含 BOM（规格、加料） **THEN** 系统 **SHALL** 写入 `sale_order_product_bom` 表
5. **IF** 商品包含属性 **THEN** 系统 **SHALL** 写入 `sale_order_product_attribute` 表
6. **IF** 商品包含备注原因 **THEN** 系统 **SHALL** 写入 `sale_order_product_reason` 表

#### 具体要求

- [ ] 3.1 通过选项模式支持不同场景的特殊处理：
  - `WithH5Product()` - H5 商品标记（设置 `is_accept_order` 为未接单）
  - `WithMemberAdd()` - 会员加购标记
  - `WithTableAdd()` - 桌台加购标记
  - `WithBuffetContext()` - 自助餐场景（需要写入自助餐相关表）
  - `WithBatchCooking()` - 分批制作场景
- [ ] 3.2 支持写入场景相关表（按场景写入）：
  - `ttpos_sale_order_coupon` - 优惠券（如添加时使用了优惠券）
  - `ttpos_sale_order_discount_strategy` - 折扣策略（如添加时应用了折扣策略）
  - `ttpos_sale_order_invoice_info` - 发票信息（如添加时需要记录发票信息）
  - `ttpos_sale_order_material` - 订单材料（如添加时需要记录材料消耗）
  - `ttpos_sale_order_peak_time` - 高峰时段（如添加时需要记录高峰时段）
  - `ttpos_sale_order_abnormal_record` - 异常记录（如添加操作产生异常时记录）

---

### Requirement 4: 数据一致性保证

**用户故事**: 作为开发人员，我想确保添加商品/实体数据写入的一致性，以便于避免数据不一致的问题

#### 验收标准

1. **WHEN** 写入商品/实体数据 **THEN** 系统 **SHALL** 确保所有相关表的数据写入在同一事务中完成
2. **IF** 任何一张表写入失败 **THEN** 系统 **SHALL** 回滚所有已写入的数据
3. **WHEN** 写入数据 **THEN** 系统 **SHALL** 确保写入的数据符合业务约束（如外键关联、快照数据等）

#### 具体要求

- [ ] 4.1 使用事务管理确保所有表的数据写入原子性
- [ ] 4.2 确保外键关联正确（如 `sale_order_product_bom.sale_order_product_uuid` 关联到 `sale_order_product.uuid`）
- [ ] 4.3 确保快照数据正确（如商品名称、价格快照等）
- [ ] 4.4 处理数据写入的异常情况，返回明确的错误信息

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 DDD 分层架构（Domain Service → Repository）
- **单一职责原则**: 订单直接添加商品领域服务只负责数据写入，不负责业务规则验证
- **模块化设计**: 领域服务应独立且可复用
- **依赖管理**: 领域服务依赖 Repository 接口，不直接依赖数据库
- **遵循规范**:
  - `.cursor/rules/go-modules.mdc` - Go Modules 模块开发规范
  - `.cursor/rules/go-main.mdc` - Go Main 核心约束
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 领域服务接口使用 `IOrderDirectAddProductsDomainService` 命名（接口以 `I` 开头）
- [ ] 实现类使用 `orderDirectAddProductsDomainService` 命名（实现以小写开头）
- [ ] 方法参数使用 `context.Context`（pkg/context）
- [ ] 参考: `.cursor/rules/go-modules.mdc` - DDD 模块开发规范

### 数据库设计要求

- [ ] 所有数据写入在同一事务中完成
- [ ] 使用 `repository.CommonRepo.Transaction` 确保原子性
- [ ] 确保外键关联正确
- [ ] 确保快照数据正确（商品名称、价格等）
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 数据写入操作应在合理时间内完成（< 500ms）
- [ ] 事务提交时间应尽可能短
- [ ] 避免不必要的数据库查询

### 测试要求

- [ ] 领域服务层测试覆盖率 ≥ 80%
- [ ] 单元测试：测试领域服务的数据写入逻辑，mock 仓储层
- [ ] 集成测试：测试与数据库的交互，验证数据写入正确性
- [ ] 场景测试：测试不同场景（POS、H5、自助餐等）的数据写入差异
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 错误处理

- [ ] 使用项目统一的错误处理机制
- [ ] 返回明确的错误码和错误信息
- [ ] 支持国际化错误信息
- [ ] 记录错误日志

### 可靠性要求

- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 数据写入失败时优雅降级（回滚事务）

---

## 验收标准

### 功能验收

1. **数据写入功能**: 订单直接添加商品领域服务能够将商品/实体数据正确写入所有相关表（13 张表）
2. **多类型支持**: 支持添加普通商品、套餐、自助餐顾客、自助餐加钟等多种类型
3. **批量添加**: 支持一次调用添加多个不同类型的商品/实体
4. **场景适配功能**: 不同场景（POS、H5、自助餐等）能够正确写入对应的数据
5. **事务一致性**: 所有表的数据写入在同一事务中完成，失败时能够正确回滚
6. **操作记录**: 添加操作能够正确记录到 `sale_order_operation_record` 表

### 测试验收

1. **单元测试**: 覆盖率达标（≥ 80%），mock 仓储层测试数据写入逻辑
2. **集成测试**: 端到端流程测试通过，验证数据写入正确性
3. **场景测试**: 不同场景（POS、H5、自助餐等）的数据写入差异测试通过
4. **事务测试**: 事务回滚测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **代码注释**: 领域服务接口和实现有清晰的注释说明
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 DDD 架构，放在 `main/app/modules/order/domain/service/` 目录下
- 接口以 `I` 开头，实现以小写开头
- 所有方法第一个参数必须是 `context.Context`（pkg/context）
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-modules.mdc` - Go Modules 模块开发规范

### 业务约束

- **职责边界**: 订单直接添加商品领域服务只负责数据写入，不负责业务规则验证（库存、限购、超时等）
- **调用流程**: 调用方可以选择性地进行业务规则验证，再调用领域服务写入数据
- **向后兼容**: 重构需要保证向后兼容，避免影响现有功能

### 资源约束

- 开发时间: 5-7 天
- Story Point: 8 SP（待技术评审确认）

---

## 依赖关系

### 技术依赖

- `main/app/modules/order/domain/repository` - 订单仓储接口（需要扩展支持批量写入）
- `main/app/repository` - 现有仓储实现（用于事务管理）

### 服务依赖

- **无外部服务依赖**

### 业务依赖

- 订单实体（`SaleOrder`、`SaleBill`）已存在
- 订单商品模型（`SaleOrderProduct`）已存在
- 自助餐顾客模型（`SaleOrderBuffetCustomerType`）已存在
- 自助餐加钟模型（`SaleOrderBuffetDelayProduct`）已存在
- 相关表结构已存在（13 张表）

---

## 风险和缓解

### 风险 1: 向后兼容风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 渐进式重构：先实现领域服务，再逐步迁移调用方，保留旧方法作为 fallback
- 充分测试：编写完整的单元测试和集成测试，覆盖所有添加场景和表写入逻辑

### 风险 2: 职责边界风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 明确文档说明领域服务的职责边界，只负责数据写入，不负责业务规则验证
- 代码审查：邀请团队进行代码审查，确保设计合理和职责边界清晰

### 风险 3: 性能风险

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 性能监控：添加性能监控，确保重构后性能不下降
- 数据写入逻辑本身不变，只是抽象到领域服务中

---

## 时间表

- **Phase 1 - 领域服务设计和实现**: 2-3 天
- **Phase 2 - 现有代码重构和适配**: 2-3 天
- **Phase 3 - 单元测试和集成测试**: 1 天
- **Phase 4 - 代码审查和优化**: 1 天
- **总计**: 5-7 天（SP = 8，待技术评审确认）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-modules.mdc` - Go Modules 模块开发规范
- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 架构文档

- `main/app/modules/order/README.md` - 订单模块文档
- `main/app/modules/inventory/domain/service/warehouse_domain_service.go` - 仓库领域服务示例
- `main/app/modules/order/domain/service/order_domain_service.go` - 订单领域服务示例

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/agent/workflows/development/service-layer-examples.md` - 服务层示例

### 相关代码

- `main/app/service/order.go:newSaleOrderProduct` - 当前加购数据写入逻辑
- `main/app/service/order_base.go` - 事务写入逻辑
- `main/app/service/order_action.go:actionAdd` - 加购业务逻辑

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-19  
**作者**: xiezhihuan  
**审核者**: {审核者}

