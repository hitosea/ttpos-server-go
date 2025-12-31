# 订单直接添加商品领域服务 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-19   |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [story-order-direct-add-products-domain-service](../../../shared/specs/active/story-order-direct-add-products-domain-service/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前向订单添加商品/实体的逻辑分散且复杂，缺乏统一的、直接的数据写入接口：

1. **缺乏统一接口**：添加普通商品、自助餐顾客、自助餐加钟、套餐等不同类型的实体，需要调用不同的方法，代码分散
2. **数据写入复杂**：每种类型的实体需要写入多个相关表，逻辑分散在不同方法中，容易出现遗漏
3. **难以直接操作**：无法通过一个简单的方法直接向订单添加指定的商品/实体，必须经过复杂的业务逻辑验证流程
4. **场景适配困难**：不同场景（POS、H5、自助餐等）需要设置不同的字段，当前逻辑难以统一管理

**示例问题场景**：
> 开发人员需要给某个订单直接添加一个普通商品、一个自助餐顾客、一个自助餐加钟、一个套餐时，当前需要：
> - 调用 `newSaleOrderProduct` 方法添加普通商品和套餐
> - 调用不同的方法添加自助餐顾客
> - 调用不同的方法添加自助餐加钟
> - 每个方法都需要进行业务规则验证（库存、限购等）
> - 数据写入逻辑分散在多个地方，难以保证一致性
> 
> 这导致：
> - 无法直接向订单添加商品，必须经过复杂的业务逻辑
> - 新增加购场景时需要复制大量代码
> - 难以进行单元测试和集成测试
> - 数据写入逻辑变更时需要修改多处

### 业务价值

通过设计专门的订单直接添加商品领域服务，可以带来以下价值：

- **简化调用**：提供一个统一的接口，直接向订单添加商品/实体，无需关心底层实现
- **提高开发效率**：开发人员只需调用一个方法，即可添加多种类型的商品/实体
- **增强可测试性**：领域服务可以独立测试数据写入逻辑，mock 仓储层即可
- **降低维护成本**：数据写入逻辑集中管理，修改影响范围可控
- **保证数据一致性**：统一的事务管理，确保所有相关表的数据写入原子性
- **支持 DDD 演进**：符合项目 DDD 架构演进方向，为订单模块重构奠定基础

### 目标用户

- [x] 开发人员（代码维护和扩展）
- [x] 测试人员（测试数据准备）
- [x] 运维人员（数据修复和迁移）
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

设计一个专门的**订单直接添加商品领域服务（OrderDirectAddProductsDomainService）**，专注于**直接向订单添加指定的商品/实体到数据库**。

**核心职责**：
- ✅ **数据写入**：直接将商品/实体数据写入数据库，不进行业务规则验证
- ✅ **多类型支持**：支持添加普通商品、套餐、自助餐顾客、自助餐加钟等多种类型
- ✅ **批量添加**：支持一次调用添加多个不同类型的商品/实体
- ❌ **不处理业务规则验证**：库存验证、限购检查、超时检查等由调用方处理
- ❌ **不处理价格计算**：价格计算由订单实体或应用服务处理

该服务将：
1. **提供统一接口**：`AddProductsToOrder(ctx, orderUuid, products)` - 直接向订单添加商品/实体
2. **支持多类型**：支持添加普通商品、套餐、自助餐顾客、自助餐加钟等
3. **批量操作**：支持一次调用添加多个不同类型的商品/实体
4. **事务保证**：所有数据写入在同一事务中完成，确保原子性
5. **符合 DDD 规范**：遵循项目现有的 DDD 模块结构，放在 `main/app/modules/order/domain/service/` 目录下

**参考现有实现**：
- 当前加购数据写入：`main/app/service/order.go:newSaleOrderProduct` 和 `order_base.go` 中的事务写入逻辑
- DDD 模块示例：`main/app/modules/inventory/domain/service/warehouse_domain_service.go`
- 订单领域服务：`main/app/modules/order/domain/service/order_domain_service.go`

### 职责边界（重要）

**✅ 订单直接添加商品领域服务负责**：
- 直接将商品/实体数据写入数据库（13 张表）
- 根据数据类型写入对应的表
- 保证数据写入的事务性和一致性
- 处理数据快照（如商品名称、价格快照等）

**❌ 订单直接添加商品领域服务不负责**：
- 库存验证（由调用方在调用前验证）
- 限购检查（由调用方在调用前检查）
- 超时检查（由调用方在调用前检查）
- 订单状态校验（由调用方在调用前校验）
- 价格计算（由订单实体或应用服务处理）
- 业务规则验证（由调用方处理）

**调用流程**：
```
调用方（应用服务）
  ↓
1. 业务规则验证（库存、限购、超时等）- 可选
  ↓
2. 构建商品/实体数据（SaleOrderProduct / SaleOrderBuffetCustomerType / SaleOrderBuffetDelayProduct）
  ↓
3. 调用领域服务（AddProductsToOrder）
  ↓
4. 领域服务直接写入数据库（13 张表）
  ↓
5. 返回结果
```

### 核心功能点

1. **直接添加方法**：`AddProductsToOrder(ctx, orderUuid, products)` - 直接向订单添加商品/实体

   **支持多种类型的商品/实体添加**：
   - ✅ **普通商品**：`SaleOrderProduct`（包含 BOM、属性、备注原因等）
   - ✅ **套餐商品**：`SaleOrderProduct`（套餐类型，可能包含子商品）
   - ✅ **自助餐顾客**：`SaleOrderBuffetCustomerType`（自助餐场景）
   - ✅ **自助餐加钟**：`SaleOrderBuffetDelayProduct`（自助餐场景）
   - ✅ **批量添加**：支持一次调用添加多个不同类型的商品/实体

   **使用示例**：
   ```go
   // 示例1：添加一个普通商品
   products := []AddToOrderProduct{
       {Type: ProductTypeNormal, Product: product1},
   }
   err := domainService.AddProductsToOrder(ctx, orderUuid, products)
   
   // 示例2：添加一个自助餐顾客
   products := []AddToOrderProduct{
       {Type: ProductTypeBuffetCustomer, BuffetCustomer: customer1},
   }
   err := domainService.AddProductsToOrder(ctx, orderUuid, products)
   
   // 示例3：添加一个自助餐加钟
   products := []AddToOrderProduct{
       {Type: ProductTypeBuffetDelay, BuffetDelay: delay1},
   }
   err := domainService.AddProductsToOrder(ctx, orderUuid, products)
   
   // 示例4：混合添加（普通商品 + 自助餐顾客 + 自助餐加钟 + 套餐）
   products := []AddToOrderProduct{
       {Type: ProductTypeNormal, Product: product1},
       {Type: ProductTypePackage, Product: package1},
       {Type: ProductTypeBuffetCustomer, BuffetCustomer: customer1},
       {Type: ProductTypeBuffetDelay, BuffetDelay: delay1},
   }
   err := domainService.AddProductsToOrder(ctx, orderUuid, products)
   ```

2. **多表数据写入**：根据数据类型将数据写入以下表（共 13 张表）：

   **核心表（按类型写入）**：
   - `ttpos_sale_order_product` - 订单商品（普通商品/套餐时写入）
   - `ttpos_sale_order_product_bom` - 商品BOM（规格、加料，如商品包含 BOM 则写入）
   - `ttpos_sale_order_product_attribute` - 商品属性（如商品包含属性则写入）
   - `ttpos_sale_order_product_reason` - 商品备注原因（如商品包含备注原因则写入）
   - `ttpos_sale_order_buffet_customer_type` - 自助餐顾客类型（添加自助餐顾客时写入）
   - `ttpos_sale_order_buffet_delay_product` - 自助餐加钟商品（添加自助餐加钟时写入）
   - `ttpos_sale_order_operation_record` - 操作记录（记录添加操作，所有场景必写）

   **场景相关表（按场景写入）**：
   - `ttpos_sale_order_coupon` - 优惠券（如添加时使用了优惠券）
   - `ttpos_sale_order_discount_strategy` - 折扣策略（如添加时应用了折扣策略）
   - `ttpos_sale_order_invoice_info` - 发票信息（如添加时需要记录发票信息）
   - `ttpos_sale_order_material` - 订单材料（如添加时需要记录材料消耗）
   - `ttpos_sale_order_peak_time` - 高峰时段（如添加时需要记录高峰时段）
   - `ttpos_sale_order_abnormal_record` - 异常记录（如添加操作产生异常时记录）

3. **场景适配策略**：通过选项模式支持不同场景的特殊处理
   - `WithH5Product()` - H5 商品标记（设置 `is_accept_order` 为未接单）
   - `WithMemberAdd()` - 会员加购标记
   - `WithTableAdd()` - 桌台加购标记
   - `WithBuffetContext()` - 自助餐场景（需要写入自助餐相关表）
   - `WithBatchCooking()` - 分批制作场景

4. **事务管理**：确保所有相关表的数据写入在同一事务中完成
5. **数据一致性**：确保写入的数据符合业务约束（如外键关联、快照数据等）

### 影响范围

**涉及终端**：
- [x] POS 收银端（通过应用服务调用）
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [x] Tablet 平板端（通过应用服务调用）
- [x] Mobile 扫码端（通过应用服务调用）
- [x] Menu 电子菜单端（通过应用服务调用）
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（需要适配新的领域服务）
- [ ] 数据模型
- [x] 业务逻辑（核心重构）
- [ ] 第三方集成
- [x] 其他: DDD 模块结构

**涉及文件**：
- `main/app/service/order.go` - 重构 `newSaleOrderProduct` 方法，调用领域服务
- `main/app/service/order_base.go` - 重构事务写入逻辑，调用领域服务
- `main/app/service/order_action.go` - 重构 `actionAdd` 方法，调用领域服务
- `main/app/service/order_product.go` - 重构 `OrderCartProductAdd` 等方法，调用领域服务
- `main/app/modules/order/domain/service/order_direct_add_products_domain_service.go` - **新增**领域服务
- `main/app/modules/order/domain/repository/order_repository.go` - 可能需要扩展仓储接口，支持批量写入

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 需要重构现有加购数据写入逻辑，但业务规则验证逻辑不变（仍由调用方处理）
- 需要适配多个调用方（POS、H5、平板等）的数据写入需求
- 需要保证向后兼容，避免影响现有功能
- 符合项目已有的 DDD 架构模式，有参考实现
- 职责边界清晰：只负责数据写入，不负责业务规则验证

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 5-7 天
- **预估 SP**: 8 SP（待技术评审确认）

**分解**：
- 领域服务设计和实现：2-3 天
- 现有代码重构和适配：2-3 天
- 单元测试和集成测试：1 天
- 代码审查和优化：1 天

### 风险识别

**潜在风险**：
1. **向后兼容风险**：重构可能影响现有功能，需要充分测试
2. **性能风险**：领域服务调用链可能增加性能开销，但数据写入逻辑本身不变
3. **迁移风险**：需要逐步迁移，不能一次性替换所有调用
4. **职责边界风险**：需要明确区分业务规则验证和数据写入的边界，避免职责混乱

**缓解措施**：
1. **充分测试**：编写完整的单元测试和集成测试，覆盖所有添加场景和表写入逻辑
2. **渐进式重构**：先实现领域服务，再逐步迁移调用方，保留旧方法作为 fallback
3. **性能监控**：添加性能监控，确保重构后性能不下降
4. **代码审查**：邀请团队进行代码审查，确保设计合理和职责边界清晰
5. **文档说明**：明确文档说明领域服务的职责边界，只负责数据写入，不负责业务规则验证

---

## 🔗 相关资源

### 参考需求

- 类似功能: `main/app/modules/inventory/domain/service/warehouse_domain_service.go` - 仓库领域服务示例
- 相关提案: [加购领域服务需求提案](./add-to-cart-domain-service.md)
- 相关 Spec: [story-order-add-to-cart-domain-service](../../../shared/specs/active/story-order-add-to-cart-domain-service/requirements.md)
- 竞品分析: 无

### 相关文档

- DDD 模块规范: `.cursor/rules/go-modules.mdc`
- 订单模块文档: `main/app/modules/order/README.md`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-order-direct-add-products-domain-service`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 开发人员  
**我想** 通过一个简单的方法直接向订单添加指定的商品/实体（普通商品、套餐、自助餐顾客、自助餐加钟等）  
**以便于** 简化代码调用，提高开发效率和可测试性

**作为** 测试人员  
**我想** 能够直接向订单添加测试数据  
**以便于** 快速准备测试场景，提高测试效率

### AC 验收标准（初稿）

1. **WHEN** 调用领域服务添加普通商品 **THEN** 系统 **SHALL** 将商品数据写入 `sale_order_product` 及相关表
2. **WHEN** 调用领域服务添加套餐商品 **THEN** 系统 **SHALL** 将套餐数据写入 `sale_order_product` 及相关表
3. **WHEN** 调用领域服务添加自助餐顾客 **THEN** 系统 **SHALL** 将顾客数据写入 `sale_order_buffet_customer_type` 表
4. **WHEN** 调用领域服务添加自助餐加钟 **THEN** 系统 **SHALL** 将加钟数据写入 `sale_order_buffet_delay_product` 表
5. **WHEN** 调用领域服务混合添加多种类型（如：普通商品 + 自助餐顾客 + 自助餐加钟 + 套餐） **THEN** 系统 **SHALL** 将所有类型的数据写入对应的表，并在同一事务中完成
6. **IF** 数据写入失败 **THEN** 系统 **SHALL** 回滚所有相关表的数据，保证数据一致性
7. **IF** 业务规则验证失败（库存不足、限购等） **THEN** 系统 **SHALL** 在调用领域服务之前返回错误，不执行数据写入
8. **IF** products 参数为空 **THEN** 系统 **SHALL** 返回错误，提示至少需要提供一个商品/实体

### 技术设计要点（初稿）

1. **领域服务接口设计**：
   ```go
   type IOrderDirectAddProductsDomainService interface {
       // AddProductsToOrder 直接向订单添加商品/实体到数据库
       // 注意：此方法不进行业务规则验证（库存、限购等），只负责数据写入
       AddProductsToOrder(
           ctx context.Context,
           orderUuid uint64,
           products []AddToOrderProduct,
           options ...AddToOrderOption,
       ) error
   }
   
   type AddToOrderProduct struct {
       Type ProductType // 商品类型：Normal, Package, BuffetCustomer, BuffetDelay
       
       // 根据 Type 使用对应的字段
       Product        *SaleOrderProduct              // 普通商品/套餐
       BuffetCustomer *SaleOrderBuffetCustomerType  // 自助餐顾客
       BuffetDelay    *SaleOrderBuffetDelayProduct  // 自助餐加钟
   }
   
   type ProductType int
   const (
       ProductTypeNormal ProductType = iota
       ProductTypePackage
       ProductTypeBuffetCustomer
       ProductTypeBuffetDelay
   )
   ```

2. **选项模式支持**：
   ```go
   type AddToOrderOption struct {
       IsH5Product      bool // H5 商品标记
       IsMemberAdd      bool // 会员加购标记
       IsTableAdd       bool // 桌台加购标记
       IsBuffetContext  bool // 自助餐场景
       BatchCookingMode string // 分批制作模式
   }
   
   func WithH5Product() AddToOrderOption { ... }
   func WithMemberAdd() AddToOrderOption { ... }
   func WithBuffetContext() AddToOrderOption { ... }
   ```

3. **事务管理**：
   - 所有表的数据写入在同一事务中完成
   - 使用 `repository.CommonRepo.Transaction` 确保原子性
   - 写入失败时自动回滚

4. **错误处理**：
   - 使用项目统一的错误处理机制
   - 返回明确的错误码和错误信息
   - 支持国际化错误信息

5. **测试策略**：
   - 单元测试：测试领域服务的数据写入逻辑，mock 仓储层
   - 集成测试：测试与数据库的交互，验证数据写入正确性
   - 场景测试：测试不同场景（POS、H5、自助餐等）的数据写入差异

### 线框图/原型（可选）

[无需 UI 变更，纯后端重构]

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-19  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/go-modules.mdc`, `.cursor/rules/specs.mdc`

