# 商品属性信息快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。  
> 本提案是「订单商品信息快照修复」的子任务，聚焦于商品名称、规格名称、小料名称、属性名称快照修复（使用现有快照字段）。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan |
| **日期**   | 2025-12-09 |
| **目标版本** | v2.11.0 |
| **状态**   | 已批准 → Spec 已创建   |
| **关联任务** | -      |
| **关联 Spec** | [story-main-product-attribute-snapshot-fix](../../shared/specs/archived/v2.12/story-main-product-attribute-snapshot-fix/) |
| **父提案** | [订单商品信息快照修复](./2025-01/order-attribute-snapshot-fix.md) |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，商品名称、规格名称、小料名称、属性名称信息会随后台数据变更而改变，导致订单历史信息不准确。虽然数据库表已有快照字段，但代码实现未使用这些字段。

**具体场景**：

1. **商品名称被修改**：
   - 订单商品："珍珠奶茶"
   - 后台将商品改名为"经典珍珠奶茶"
   - 查询订单时显示："经典珍珠奶茶"（显示的是新名称，而非下单时的名称）

2. **规格名称被删除或改名**：
   - 订单商品："珍珠奶茶（大杯）"（下单时选择了"大杯"规格）
   - 后台删除了"大杯"规格或改名为"超大杯"
   - 查询订单时显示错误信息或新名称

3. **小料被删除或改名**：
   - 订单商品："珍珠奶茶（加珍珠）"（下单时选择了"珍珠"小料）
   - 后台删除了"珍珠"小料或改名为"黑珍珠"
   - 查询订单时显示错误信息或新名称

4. **属性名称被删除或改名**：
   - 订单商品："珍珠奶茶,热饮"（下单时选择了"热饮"属性）
   - 后台删除了"热饮"属性或改名为"热饮30度"
   - 查询订单时显示错误信息或新名称

**问题影响**：

- ❌ 订单历史信息不准确，无法还原下单时的真实状态
- ❌ 影响对账、退款等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 可能导致法律风险（如发票、账单与实际订单不符）
- ❌ 商品、规格、小料、属性等信息被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保订单商品信息准确反映下单时的状态
- ✅ **合规性**：满足财务、税务对订单历史记录的要求
- ✅ **可追溯性**：支持订单历史查询和问题追溯
- ✅ **用户体验**：用户查看历史订单时看到的是下单时的真实信息
- ✅ **业务可靠性**：避免因数据变更导致的业务逻辑错误

### 目标用户

- [x] 收银员（查看历史订单）
- [x] 商户管理员（对账、报表）
- [ ] 厨房人员
- [ ] 顾客
- [x] 财务人员（对账、审计）

---

## 💡 解决方案概述

### 方案描述

**核心思路**：订单商品相关信息应该使用快照数据，而不是从关联表实时获取。数据库表已有快照字段，只需修复查询逻辑。

**现状分析**：

1. **数据库设计已支持快照**：
   - `ttpos_sale_order_product` 表已有 `name`（商品名称）、`flavor_name`（规格名称）字段
   - `ttpos_sale_order_product_bom` 表已有 `name` 字段（规格或小料名称），注释："规格或小料规格名称,不随后台更新"
   - `ttpos_sale_order_product_attribute` 表已有 `name` 字段，注释："商品属性名称,不随后台更新"
   - 说明设计时已考虑快照需求

2. **代码实现未使用快照**：
   - **商品名称**：从 `MultiLanguageName` 关联对象获取（通过 UUID 关联，可能被修改）
   - **规格名称**：从 `ProductBom.ProductFlavor.MultiLanguageName` 获取，未使用 `SaleOrderProductBom.Name` 或 `SaleOrderProduct.FlavorName`
   - **小料名称**：从 `ProductBom.ProductSauce.MultiLanguageName` 获取，未使用 `SaleOrderProductBom.Name`
   - **属性名称**：从 `ProductAttribute.MultiLanguageName` 获取，未使用 `SaleOrderProductAttribute.Name`
   - 导致相关数据被删除或改名时，订单显示信息变化

**解决方案**：

1. **修复查询逻辑**（使用现有快照字段，无需数据库变更）：
   - **商品名称**：优先使用 `SaleOrderProduct.Name`，如果为空则使用 `MultiLanguageName`（兼容历史数据）
   - **规格名称**：优先使用 `SaleOrderProduct.FlavorName` 或 `SaleOrderProductBom.Name`，如果为空则使用关联表数据
   - **小料名称**：优先使用 `SaleOrderProductBom.Name`，如果为空则使用关联表数据
   - **属性名称**：优先使用 `SaleOrderProductAttribute.Name`，如果为空则使用关联表数据

2. **多语言支持**：
   - 采用"主语言快照 + 关联表补充"的混合方案
   - 快照字段保存主语言（中文）作为历史快照
   - 查询时构建 `LocaleResponse`：
     - `ZH`（中文）：优先使用快照字段，如果为空则使用关联表
     - 其他语言（TH、EN等）：优先使用关联表数据，如果关联表不存在或已删除，则使用快照的主语言填充
     - 这样既保证了快照完整性（即使数据被删除也能显示），又尽可能提供了多语言支持

3. **实施策略**：
   - **新订单**：下单时已保存快照字段（无需修改下单逻辑）
   - **历史订单**：快照字段为空时，降级使用关联表数据（兼容性处理）
   - **渐进式实施**：不需要强制迁移所有历史数据，新订单自动使用快照机制

### 核心功能点

1. **修复商品名称获取逻辑**
   - 修改 `SaleOrderProduct.MultiLanguageName.GetNames()` 相关方法
   - 优先使用 `SaleOrderProduct.Name` 字段（单语言）
   - 如果 `Name` 为空，降级使用 `MultiLanguageName` 关联对象

2. **修复规格名称获取逻辑**
   - 修改 `SaleOrderProduct.GetFlavorName()` 方法
   - 修改 `SaleOrderProduct.GetFlavorSaleOrderProductBom()` 方法
   - 优先使用 `SaleOrderProduct.FlavorName` 或 `SaleOrderProductBom.Name` 字段
   - 如果快照字段为空，降级使用 `ProductBom.ProductFlavor.MultiLanguageName`

3. **修复小料名称获取逻辑**
   - 修改 `SaleOrderProduct.GetSauceNamesList()` 方法
   - 修改 `SaleOrderProduct.GetSauceSaleOrderProductBom()` 方法
   - 优先使用 `SaleOrderProductBom.Name` 字段
   - 如果 `Name` 为空，降级使用 `ProductBom.ProductSauce.MultiLanguageName`

4. **修复属性名称获取逻辑**
   - 修改 `SaleOrderProduct.GetAttributeName()` 方法
   - 修改 `SaleOrderProduct.GetAttributeNameList()` 方法
   - 修改 `SaleOrderProduct.GetPureAttributeNameList()` 方法
   - 修改 `SaleOrderProduct.GetAttributeNamesByLang()` 方法
   - 优先使用 `SaleOrderProductAttribute.Name` 字段

5. **修复商品名称属性组合方法**
   - 修改 `SaleOrderProduct.GetProductNameAttributes()` 方法
   - 修改 `SaleOrderProduct.GetNameAndFlavorName()` 方法
   - 确保使用快照数据

### 影响范围

**涉及终端**：
- [x] POS 收银端（订单查询）
- [x] Shop 商家管理端（订单管理、对账）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [x] Member 会员端（历史订单）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [x] 数据模型（`SaleOrderProduct`、`SaleOrderProductBom`、`SaleOrderProductAttribute`）
- [x] 业务逻辑（名称获取方法）
- [ ] 第三方集成
- [ ] 数据库迁移（无需新增字段，使用现有字段）
- [ ] 下单逻辑（无需修改，已保存快照）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要修改业务逻辑，涉及数据一致性
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 无需数据库结构变更（使用现有快照字段）
- 需要修改多个模型方法和查询逻辑
- 需要处理兼容性和多语言支持
- 需要充分测试确保不影响现有功能

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**任务分解**：
1. **代码修改 - 商品名称**（0.5 天）
   - 修改商品名称获取方法
   - 添加兼容性处理

2. **代码修改 - 规格名称**（0.5 天）
   - 修改规格名称获取方法
   - 添加兼容性处理

3. **代码修改 - 小料名称**（0.5 天）
   - 修改小料名称获取方法
   - 添加兼容性处理

4. **代码修改 - 属性名称**（0.5-1 天）
   - 修改属性名称获取方法
   - 修改组合方法
   - 添加兼容性处理

5. **测试验证**（0.5-1 天）
   - 单元测试（覆盖所有修改的方法）
   - 集成测试（验证订单查询）
   - 回归测试（确保不影响现有功能）

### 风险识别

**潜在风险**：

1. **历史数据不完整**
   - 风险：部分历史订单的快照字段可能为空
   - 影响：需要降级处理
   - 缓解：实现降级逻辑，历史数据可以逐步迁移

2. **多语言支持问题**
   - **问题**：快照字段（如 `Name`、`FlavorName`）只保存单语言（中文），但接口返回需要 `dto.LocaleResponse` 格式（多语言）
   - **影响**：需要设计合理的多语言快照方案
   - **解决方案**：采用"主语言快照 + 关联表补充"的混合方案
     - 快照字段保存主语言（中文）
     - 查询时，优先使用快照字段填充主语言（ZH）
     - 如果关联表数据存在且未删除，使用关联表数据填充其他语言（TH、EN等）
     - 如果关联表数据不存在（已删除），所有语言都用快照的主语言填充
     - 这样既保证了快照完整性（数据被删除也能显示），又尽可能提供了多语言支持

3. **回归风险**
   - 风险：修改核心方法可能影响其他功能（订单查询、打印、导出等）
   - 影响：需要充分测试，特别是订单相关的所有功能

4. **数据一致性**
   - 风险：多个地方需要修改，可能遗漏某些场景
   - 影响：需要全面梳理所有使用这些方法的地方

**缓解措施**：

1. **兼容性处理**：
   - 实现降级逻辑，确保历史订单正常显示
   - 逐步迁移，不强制要求所有数据立即完整

2. **多语言处理**：
   - 采用"主语言快照 + 关联表补充"方案
   - 快照字段保存主语言（中文）
   - 查询时优先使用快照字段填充主语言
   - 关联表存在时使用关联表数据填充其他语言
   - 关联表不存在时所有语言都用快照的主语言填充

3. **充分测试**：
   - 编写单元测试覆盖所有修改的方法和场景
   - 测试多语言快照逻辑（关联表存在/不存在的情况）
   - 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
   - 在测试环境充分验证后再上线

4. **全面梳理**：
   - 梳理所有使用商品名称、规格、小料、属性的地方
   - 确保所有相关方法都使用快照数据

---

## 🔗 相关资源

### 参考需求

- 父提案: [订单商品信息快照修复](./2025-01/order-attribute-snapshot-fix.md)
- 类似功能: 订单商品快照机制（商品名称、价格等已有快照）
- 竞品分析: 主流餐饮系统都采用订单快照机制

### 相关文档

- 订单信息获取逻辑分析: `docs/shared/api/cashier-order-info-analysis.md`
- 数据模型定义: `main/app/model/order.go`、`main/app/model/sale_order_product.go`
- 属性获取方法: `main/app/model/sale_order_product.go`

### 代码位置

**问题代码**：
- `main/app/model/sale_order_product.go:235` - `GetProductNameAttributes()` 方法（商品名称、规格、属性）
- `main/app/model/sale_order_product.go:1299` - `GetNameAndFlavorName()` 方法（商品名称、规格）
- `main/app/model/sale_order_product.go:1354` - `GetAttributeNameList()` 方法（规格、属性、小料）
- `main/app/model/sale_order_product.go:1413` - `GetFlavorName()` 方法（规格名称）
- `main/app/model/sale_order_product.go:1427` - `GetSauceNamesList()` 方法（小料名称）
- `main/app/model/sale_order_product.go:1399` - `GetPureAttributeNameList()` 方法（属性名称）
- `main/app/model/sale_order_product.go:1476` - `GetAttributeNamesByLangs()` 方法（规格、属性、小料）
- `main/app/model/sale_order_product.go:754` - `MultiLanguageName.GetNames()` 使用（商品名称）

**数据模型**：
- `main/app/model/sale_order_product.go:22-133` - `SaleOrderProduct` 模型定义（`Name`、`FlavorName` 字段）
- `main/app/model/order.go:60-70` - `SaleOrderProductAttribute` 模型定义（`Name` 字段）
- `main/app/model/order.go:91-106` - `SaleOrderProductBom` 模型定义（`Name` 字段）

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

- [ ] 创建 Spec：`story-main-product-attribute-snapshot-fix`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实商品信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的商品属性  
**以便于** 准确处理退款和客户咨询

### AC 验收标准（初稿）

1. **WHEN** 查询包含商品名称的订单 **THEN** 系统 **SHALL** 显示下单时保存的商品名称快照
2. **WHEN** 查询包含规格的订单 **THEN** 系统 **SHALL** 显示下单时保存的规格名称快照
3. **WHEN** 查询包含小料的订单 **THEN** 系统 **SHALL** 显示下单时保存的小料名称快照
4. **WHEN** 查询包含属性的订单 **THEN** 系统 **SHALL** 显示下单时保存的属性名称快照
5. **IF** 后台删除了某个商品/规格/小料/属性 **THEN** 历史订单 **SHALL** 仍然显示该信息的原始名称
6. **IF** 后台修改了某个商品/规格/小料/属性的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称
7. **IF** 订单快照数据为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据（兼容性）
8. **WHEN** 查询订单商品信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
9. **IF** 关联表数据存在 **THEN** 系统 **SHALL** 使用关联表数据填充其他语言（TH、EN等）
10. **IF** 关联表数据不存在（已删除） **THEN** 系统 **SHALL** 使用快照的主语言填充所有语言字段

### 技术方案要点（初稿）

#### 多语言快照策略

**核心思路**：主语言快照 + 关联表补充

- **快照字段**：保存主语言（中文）作为历史快照
- **查询逻辑**：
  - 主语言（ZH）：优先使用快照字段
  - 其他语言：优先使用关联表，如果关联表不存在或已删除，则使用快照的主语言填充

#### 具体实现方案

1. **修改商品名称获取方法**：
   - 优先使用 `SaleOrderProduct.Name` 字段
   - 如果为空，降级使用 `MultiLanguageName` 关联对象
   - 构建多语言响应：主语言使用快照，其他语言使用关联表（如果存在）

2. **修改规格名称获取方法**：
   - 优先使用 `SaleOrderProduct.FlavorName` 或 `SaleOrderProductBom.Name` 字段
   - 如果为空，降级使用 `ProductBom.ProductFlavor.MultiLanguageName`
   - 构建多语言响应

3. **修改小料名称获取方法**：
   - 优先使用 `SaleOrderProductBom.Name` 字段
   - 如果为空，降级使用 `ProductBom.ProductSauce.MultiLanguageName`
   - 构建多语言响应

4. **修改属性名称获取方法**：
   - 优先使用 `SaleOrderProductAttribute.Name` 字段
   - 如果为空，降级使用 `ProductAttribute.MultiLanguageName`
   - 构建多语言响应

5. **修改组合方法**：
   - `GetProductNameAttributes()` 方法：确保使用快照数据
   - `GetNameAndFlavorName()` 方法：确保使用快照数据

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
**创建日期**: 2025-12-09  
**维护者**: xiezhihuan  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

