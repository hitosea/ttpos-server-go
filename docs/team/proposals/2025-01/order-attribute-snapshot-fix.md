# 订单商品信息快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 开发团队 |
| **日期**   | 2025-01-27 |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | -      |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，商品相关信息（商品名称、规格名称、属性名称、小料名称等）会随后台数据变更而改变，导致订单历史信息不准确。

**具体场景**：

1. **商品属性被删除**：
   - 订单商品："珍珠奶茶,热饮"（下单时选择了"热饮"属性）
   - 后台删除了"热饮"属性
   - 查询订单时显示："珍珠奶茶"（属性信息丢失）

2. **商品属性被改名**：
   - 订单商品："珍珠奶茶,热饮"（下单时选择了"热饮"属性）
   - 后台将"热饮"改名为"热饮30度"
   - 查询订单时显示："珍珠奶茶,热饮30度"（显示的是新名称，而非下单时的名称）

3. **商品规格被删除或改名**：
   - 订单商品："珍珠奶茶（大杯）"（下单时选择了"大杯"规格）
   - 后台删除了"大杯"规格或改名为"超大杯"
   - 查询订单时显示错误信息或新名称

4. **小料被删除或改名**：
   - 订单商品："珍珠奶茶（加珍珠）"（下单时选择了"珍珠"小料）
   - 后台删除了"珍珠"小料或改名为"黑珍珠"
   - 查询订单时显示错误信息或新名称

5. **商品名称被修改**：
   - 订单商品："珍珠奶茶"
   - 后台将商品改名为"经典珍珠奶茶"
   - 查询订单时显示："经典珍珠奶茶"（显示的是新名称，而非下单时的名称）

6. **免单原因被删除或改名**：
   - 订单免单原因："员工福利"
   - 后台删除了"员工福利"原因或改名为"内部测试"
   - 查询订单时显示错误信息或新名称

7. **退菜原因被删除或改名**：
   - 订单退菜原因："菜品质量问题"
   - 后台删除了该原因或改名为"质量问题"
   - 查询订单时显示错误信息或新名称

8. **自助餐名称被修改**：
   - 订单自助餐："豪华自助餐"
   - 后台将自助餐改名为"超值自助餐"
   - 查询订单时显示："超值自助餐"（显示的是新名称，而非下单时的名称）

9. **外卖来源被删除或改名**：
   - 订单来源："美团外卖"
   - 后台删除了"美团外卖"或改名为"美团"
   - 查询订单时显示错误信息或新名称

10. **国籍信息被删除或改名**：
    - 订单国籍："中国"
    - 后台删除了"中国"或改名为"中华人民共和国"
    - 查询订单时显示错误信息或新名称

**问题影响**：

- ❌ 订单历史信息不准确，无法还原下单时的真实状态
- ❌ 影响对账、退款等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 可能导致法律风险（如发票、账单与实际订单不符）
- ❌ 商品、规格、属性、小料、原因、来源等信息被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保订单信息准确反映下单时的状态
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

**核心思路**：订单商品相关信息应该使用快照数据，而不是从关联表实时获取。

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
   - **免单原因**：从 `FreeReason.MultiLanguageName` 获取，`SaleOrderProductReason` 没有快照字段
   - **退菜原因**：从 `ReturnFoodReason.MultiLanguageName` 获取，`SaleOrderProductReason` 没有快照字段
   - **自助餐名称**：从 `BuffetPackage.MultiLanguageName` 获取，`SaleBill` 没有快照字段
   - **外卖来源**：从 `OrderSource.MultiLanguageName` 获取，`SaleBill` 没有快照字段
   - **国籍**：从 `Nationality.MultiLanguageName` 获取，`SaleBill` 没有快照字段
   - 导致相关数据被删除或改名时，订单显示信息变化

**解决方案**：

1. **修复查询逻辑**（使用现有快照字段）：
   - **商品名称**：优先使用 `SaleOrderProduct.Name`，如果为空则使用 `MultiLanguageName`（兼容历史数据）
   - **规格名称**：优先使用 `SaleOrderProduct.FlavorName` 或 `SaleOrderProductBom.Name`，如果为空则使用关联表数据
   - **小料名称**：优先使用 `SaleOrderProductBom.Name`，如果为空则使用关联表数据
   - **属性名称**：优先使用 `SaleOrderProductAttribute.Name`，如果为空则使用关联表数据

2. **新增快照字段**（需要数据库结构变更）：
   - **免单原因**：在 `SaleOrderProductReason` 表添加 `Name` 字段（单语言快照）
   - **退菜原因**：在 `SaleOrderProductReason` 表已有 `MultiLanguageNameUuid`，但需要添加 `Name` 字段作为快照
   - **自助餐名称**：在 `SaleBill` 表添加 `BuffetPackage1Name` 和 `BuffetPackage2Name` 字段（单语言快照）
   - **外卖来源名称**：在 `SaleBill` 表添加 `OrderSourceName` 字段（单语言快照）
   - **国籍名称**：在 `SaleBill` 表添加 `NationalityName` 字段（单语言快照）

3. **修复查询逻辑**（新增字段）：
   - **免单原因**：优先使用 `SaleOrderProductReason.Name`，如果为空则使用关联表数据
   - **退菜原因**：优先使用 `SaleOrderProductReason.Name`，如果为空则使用关联表数据
   - **自助餐名称**：优先使用 `SaleBill.BuffetPackage1Name`/`BuffetPackage2Name`，如果为空则使用关联表数据
   - **外卖来源**：优先使用 `SaleBill.OrderSourceName`，如果为空则使用关联表数据
   - **国籍**：优先使用 `SaleBill.NationalityName`，如果为空则使用关联表数据

4. **实施策略**：
   - **新订单**：下单时保存所有快照字段（包括新增字段）
   - **历史订单**：快照字段为空时，降级使用关联表数据（兼容性处理）
   - **数据迁移**：可选，从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）
   - **渐进式实施**：不需要强制迁移所有历史数据，新订单自动使用快照机制

4. **确保快照数据完整性**：
   - **新订单**：下单时正确保存所有快照字段（包括新增字段）
   - **历史订单**：快照字段为空时，降级使用关联表数据（保证兼容性）
   - **数据迁移**：可选，从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）

5. **实施策略说明**：
   - **渐进式实施**：不需要强制迁移所有历史数据
   - **新订单优先**：新订单自动使用快照机制，确保数据完整性
   - **历史订单兼容**：历史订单通过降级逻辑正常显示，不影响现有功能
   - **可选迁移**：可以根据业务需要，选择性迁移历史数据

6. **多语言支持**：
   - **现状**：快照字段只保存单语言（中文），但接口需要返回多语言格式
   - **解决方案**：采用"主语言快照 + 关联表补充"的混合方案
     - 快照字段保存主语言（中文）作为历史快照
     - 查询时构建 `LocaleResponse`：
       - `ZH`（中文）：优先使用快照字段，如果为空则使用关联表
       - 其他语言（TH、EN等）：优先使用关联表数据，如果关联表不存在或已删除，则使用快照的主语言填充
     - 这样既保证了快照完整性（即使数据被删除也能显示），又尽可能提供了多语言支持

### 核心功能点

#### 一、使用现有快照字段（无需数据库变更）

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

#### 二、新增快照字段（需要数据库变更）

6. **数据库结构变更**
   - 在 `ttpos_sale_order_product_reason` 表添加 `name` 字段（VARCHAR(255)），用于快照免单/退菜原因名称
   - 在 `ttpos_sale_bill` 表添加以下字段：
     - `buffet_package1_name` VARCHAR(255) - 自助餐套餐1名称快照
     - `buffet_package2_name` VARCHAR(255) - 自助餐套餐2名称快照
     - `order_source_name` VARCHAR(255) - 外卖来源名称快照
     - `nationality_name` VARCHAR(255) - 国籍名称快照

7. **修复免单原因获取逻辑**
   - 修改 `SaleOrder.GetFreeReason()` 方法
   - 优先使用 `SaleOrderProductReason.Name` 字段
   - 如果 `Name` 为空，降级使用 `FreeReason.MultiLanguageName`

8. **修复退菜原因获取逻辑**
   - 修改 `SaleOrderProduct.GetCancelReason()` 方法
   - 优先使用 `SaleOrderProductReason.Name` 字段
   - 如果 `Name` 为空，降级使用 `ReturnFoodReason.MultiLanguageName`

9. **修复自助餐名称获取逻辑**
   - 修改 `SaleBill.GetBuffetNames()` 方法
   - 修改 `SaleOrder.GetBuffetNames()` 方法
   - 优先使用 `SaleBill.BuffetPackage1Name`/`BuffetPackage2Name` 字段
   - 如果快照字段为空，降级使用 `BuffetPackage.MultiLanguageName`

10. **修复外卖来源获取逻辑**
    - 修改订单详情查询中的 `OrderSourceName` 获取逻辑
    - 优先使用 `SaleBill.OrderSourceName` 字段
    - 如果快照字段为空，降级使用 `OrderSource.MultiLanguageName`

11. **修复国籍获取逻辑**
    - 修改订单详情查询中的 `NationalityName` 获取逻辑
    - 优先使用 `SaleBill.NationalityName` 字段
    - 如果快照字段为空，降级使用 `Nationality.MultiLanguageName`

12. **修复自助餐顾客类型套餐名称获取逻辑**
    - 在 `ttpos_sale_order_buffet_customer_type` 表添加 `buffet_package_name` 字段（TEXT 类型，多语言 JSON 快照）
    - 修改 `SaleOrderBuffetCustomerType` 模型，添加 `BuffetPackageName` 字段和 `GetLocaleBuffetPackageName()` 方法
    - 修改所有使用 `SaleOrderBuffetCustomerType` 的地方，优先使用快照字段
    - 修改下单逻辑，创建 `SaleOrderBuffetCustomerType` 时保存自助餐套餐名称快照

#### 三、数据迁移和兼容性

12. **数据完整性检查**
    - 检查历史订单的所有快照字段填充情况
    - 如有缺失，提供数据迁移脚本

13. **兼容性处理**
    - 当快照字段为空时，降级使用关联表数据
    - 确保历史订单正常显示

14. **下单时保存快照**
    - 确保创建订单时正确保存所有快照字段
    - 包括新增的字段（免单原因、退菜原因、自助餐名称、外卖来源、国籍）

#### 四、测试验证

15. **测试验证**
    - 验证商品删除/改名后订单显示
    - 验证规格删除/改名后订单显示
    - 验证小料删除/改名后订单显示
    - 验证属性删除/改名后订单显示
    - 验证免单原因删除/改名后订单显示
    - 验证退菜原因删除/改名后订单显示
    - 验证自助餐名称修改后订单显示
    - 验证外卖来源删除/改名后订单显示
    - 验证国籍删除/改名后订单显示
    - 验证历史订单兼容性

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
- [ ] Member 会员端（历史订单）

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [x] 数据模型（`SaleOrderProductAttribute`、`SaleOrderProductReason`、`SaleBill`）
- [x] 业务逻辑（名称获取方法）
- [ ] 第三方集成
- [x] 数据库迁移（新增快照字段）
- [x] 下单逻辑（保存快照数据）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [ ] **中**：需要修改业务逻辑，涉及数据一致性
- [x] **高**：涉及数据库结构变更、业务逻辑修改、数据迁移

**说明**：
- 需要数据库结构变更（新增快照字段）
- 需要修改多个模型方法和查询逻辑
- 需要处理兼容性和数据迁移
- 需要修改下单逻辑，确保保存快照数据
- 需要充分测试确保不影响现有功能

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 5-8 天
- **预估 SP**: 8-13 SP（待技术评审确认）

**任务分解**：
1. **数据库结构变更**（1-2 天）
   - 设计数据库迁移脚本
   - 添加快照字段到相关表
   - 执行迁移并验证

2. **代码修改 - 使用现有快照字段**（2-3 天）
   - 修改商品名称获取方法
   - 修改规格名称获取方法
   - 修改小料名称获取方法
   - 修改属性名称获取方法
   - 修改组合方法（如 `GetProductNameAttributes`）
   - 添加兼容性处理

3. **代码修改 - 新增快照字段**（1-2 天）
   - 修改免单原因获取方法
   - 修改退菜原因获取方法
   - 修改自助餐名称获取方法
   - 修改外卖来源获取方法
   - 修改国籍获取方法
   - 修改下单逻辑，确保保存快照数据

4. **数据检查与迁移**（0.5-1 天）
   - 检查历史数据完整性
   - 编写数据迁移脚本（补充历史数据的快照字段）
   - 执行数据迁移

5. **测试验证**（0.5-1 天）
   - 单元测试（覆盖所有修改的方法）
   - 集成测试（验证订单查询、下单保存快照）
   - 回归测试（确保不影响现有功能）

### 风险识别

**潜在风险**：

1. **数据库结构变更风险**
   - 风险：需要新增快照字段，涉及数据库迁移
   - 影响：需要设计迁移脚本，确保不影响现有数据
   - 缓解：使用 `ALTER TABLE ADD COLUMN` 添加可空字段，不影响现有数据

2. **历史数据不完整**
   - 风险：部分历史订单的快照字段可能为空
   - 影响：需要降级处理或数据迁移
   - 缓解：实现降级逻辑，历史数据可以逐步迁移

3. **多语言支持问题** ⚠️ **关键问题**
   - **问题**：快照字段（如 `Name`、`FlavorName`）只保存单语言（中文），但接口返回需要 `dto.LocaleResponse` 格式（多语言）
   - **影响**：需要设计合理的多语言快照方案
   - **解决方案**：采用"主语言快照 + 关联表补充"的混合方案
     - 快照字段保存主语言（中文）
     - 查询时，优先使用快照字段填充主语言（ZH）
     - 如果关联表数据存在且未删除，使用关联表数据填充其他语言（TH、EN等）
     - 如果关联表数据不存在（已删除），所有语言都用快照的主语言填充
     - 这样既保证了快照完整性（数据被删除也能显示），又尽可能提供了多语言支持

4. **下单逻辑修改风险**
   - 风险：需要修改下单逻辑，确保保存所有快照字段
   - 影响：可能遗漏某些下单场景，导致快照数据不完整
   - 缓解：全面梳理所有下单入口，确保都保存快照数据

5. **性能影响**
   - 风险：如果快照字段为空，降级查询可能增加数据库查询
   - 影响：需要优化查询逻辑，优先使用快照数据

6. **回归风险**
   - 风险：修改核心方法可能影响其他功能（订单查询、打印、导出等）
   - 影响：需要充分测试，特别是订单相关的所有功能

7. **数据一致性**
   - 风险：多个地方需要修改，可能遗漏某些场景
   - 影响：需要全面梳理所有使用这些方法的地方

**缓解措施**：

1. **数据检查**：
   - 先检查历史数据的所有快照字段填充情况
   - 根据检查结果决定是否需要数据迁移

2. **兼容性处理**：
   - 实现降级逻辑，确保历史订单正常显示
   - 逐步迁移，不强制要求所有数据立即完整

3. **多语言处理**：
   - 采用"主语言快照 + 关联表补充"方案
   - 快照字段保存主语言（中文）
   - 查询时优先使用快照字段填充主语言
   - 关联表存在时使用关联表数据填充其他语言
   - 关联表不存在时所有语言都用快照的主语言填充

4. **性能优化**：
   - 优先使用快照数据，减少关联查询
   - 如果必须降级查询，使用索引优化

5. **充分测试**：
   - 编写单元测试覆盖所有修改的方法和场景
   - 测试多语言快照逻辑（关联表存在/不存在的情况）
   - 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
   - 在测试环境充分验证后再上线

6. **全面梳理**：
   - 梳理所有使用商品名称、规格、小料、属性的地方
   - 确保所有相关方法都使用快照数据

---

## 🔗 相关资源

### 参考需求

- 类似功能: 订单商品快照机制（商品名称、价格等已有快照）
- 竞品分析: 主流餐饮系统都采用订单快照机制

### 相关文档

- 订单信息获取逻辑分析: `docs/shared/api/cashier-order-info-analysis.md`
- 数据模型定义: `main/app/model/order.go`
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
- `main/app/model/order.go:347-362` - `SaleOrderProductReason` 模型定义（需要新增 `Name` 字段）
- `main/app/model/sale_bill.go:15-112` - `SaleBill` 模型定义（需要新增 `BuffetPackage1Name`、`BuffetPackage2Name`、`OrderSourceName`、`NationalityName` 字段）
- `main/app/model/sale_order.go:20-105` - `SaleOrder` 模型定义（`FreeReason` 字段）
- `main/app/model/reason.go` - `FreeReason`、`ReturnFoodReason` 模型定义
- `main/app/model/order_source.go` - `OrderSource` 模型定义
- `main/app/model/nationality.go` - `Nationality` 模型定义

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

- [ ] 创建 Spec：`story-main-order-attribute-snapshot-fix`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实属性信息  
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
8. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 正确保存所有快照字段（商品名称、规格名称、小料名称、属性名称、免单原因、退菜原因、自助餐名称、外卖来源、国籍）
9. **WHEN** 查询订单商品信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
10. **IF** 关联表数据存在 **THEN** 系统 **SHALL** 使用关联表数据填充其他语言（TH、EN等）
11. **IF** 关联表数据不存在（已删除） **THEN** 系统 **SHALL** 使用快照的主语言填充所有语言字段
12. **WHEN** 查询包含免单原因的订单 **THEN** 系统 **SHALL** 显示下单时保存的免单原因快照
13. **WHEN** 查询包含退菜原因的订单 **THEN** 系统 **SHALL** 显示下单时保存的退菜原因快照
14. **WHEN** 查询包含自助餐的订单 **THEN** 系统 **SHALL** 显示下单时保存的自助餐名称快照
15. **WHEN** 查询包含外卖来源的订单 **THEN** 系统 **SHALL** 显示下单时保存的外卖来源名称快照
16. **WHEN** 查询包含国籍的订单 **THEN** 系统 **SHALL** 显示下单时保存的国籍名称快照
17. **IF** 后台删除了免单/退菜原因、自助餐、外卖来源、国籍 **THEN** 历史订单 **SHALL** 仍然显示该信息的原始名称
18. **IF** 后台修改了免单/退菜原因、自助餐、外卖来源、国籍的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称

### 技术方案要点（初稿）

#### 多语言快照策略

**核心思路**：主语言快照 + 关联表补充

- **快照字段**：保存主语言（中文）作为历史快照
- **查询逻辑**：
  - 主语言（ZH）：优先使用快照字段
  - 其他语言：优先使用关联表，如果关联表不存在或已删除，则使用快照的主语言填充

#### 具体实现方案

1. **修改商品名称获取方法**：
   ```go
   // 构建多语言响应
   func (model *SaleOrderProduct) GetLocaleName() dto.LocaleResponse {
       // 优先使用快照字段（主语言）
       snapshotName := model.Name
       
       // 如果快照字段为空，降级使用关联表（兼容历史数据）
       if snapshotName == "" && model.MultiLanguageName != nil {
           return model.MultiLanguageName.GetNames()
       }
       
       // 如果快照字段有值，构建多语言响应
       result := dto.LocaleResponse{ZH: snapshotName}
       
       // 如果关联表数据存在且未删除，使用关联表数据填充其他语言
       if model.MultiLanguageName != nil && !model.MultiLanguageName.IsNullName() {
           multiLang := model.MultiLanguageName.GetNames()
           result.TH = multiLang.TH
           result.EN = multiLang.EN
           result.ZHTW = multiLang.ZHTW
           result.JA = multiLang.JA
           result.KO = multiLang.KO
           result.MY = multiLang.MY
           result.TR = multiLang.TR
           result.SV = multiLang.SV
       } else {
           // 如果关联表数据不存在（已删除），所有语言都用快照的主语言填充
           result.TH = snapshotName
           result.EN = snapshotName
           result.ZHTW = snapshotName
           result.JA = snapshotName
           result.KO = snapshotName
           result.MY = snapshotName
           result.TR = snapshotName
           result.SV = snapshotName
       }
       
       return result
   }
   ```

2. **修改规格名称获取方法**：
   ```go
   func (model *SaleOrderProduct) GetLocaleFlavorName() dto.LocaleResponse {
       // 优先使用快照字段
       snapshotName := model.FlavorName
       
       // 如果快照字段为空，尝试从 SaleOrderProductBom 获取
       if snapshotName == "" {
           bom := model.GetFlavorSaleOrderProductBom()
           if bom != nil && bom.Name != "" {
               snapshotName = bom.Name
           }
       }
       
       // 如果快照字段仍为空，降级使用关联表
       if snapshotName == "" {
           bom := model.GetFlavorSaleOrderProductBom()
           if bom != nil && bom.ProductBom.ProductFlavor.MultiLanguageName != nil {
               return bom.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
           }
           return dto.LocaleResponse{}
       }
       
       // 构建多语言响应
       result := dto.LocaleResponse{ZH: snapshotName}
       
       // 如果关联表数据存在，使用关联表数据填充其他语言
       bom := model.GetFlavorSaleOrderProductBom()
       if bom != nil && bom.ProductBom.ProductFlavor.MultiLanguageName != nil {
           multiLang := bom.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
           result.TH = multiLang.TH
           result.EN = multiLang.EN
           // ... 其他语言
       } else {
           // 如果关联表数据不存在，所有语言都用快照的主语言填充
           result.TH = snapshotName
           result.EN = snapshotName
           // ... 其他语言都用 snapshotName
       }
       
       return result
   }
   ```

3. **修改小料名称获取方法**：
   ```go
   func (model *SaleOrderProductBom) GetLocaleSauceName() dto.LocaleResponse {
       // 优先使用快照字段
       snapshotName := model.Name
       
       // 如果快照字段为空，降级使用关联表
       if snapshotName == "" {
           if model.ProductBom.ProductSauce.MultiLanguageName != nil {
               return model.ProductBom.ProductSauce.MultiLanguageName.GetNames()
           }
           return dto.LocaleResponse{}
       }
       
       // 构建多语言响应
       result := dto.LocaleResponse{ZH: snapshotName}
       
       // 如果关联表数据存在，使用关联表数据填充其他语言
       if model.ProductBom.ProductSauce.MultiLanguageName != nil {
           multiLang := model.ProductBom.ProductSauce.MultiLanguageName.GetNames()
           result.TH = multiLang.TH
           result.EN = multiLang.EN
           // ... 其他语言
       } else {
           // 如果关联表数据不存在，所有语言都用快照的主语言填充
           result.TH = snapshotName
           result.EN = snapshotName
           // ... 其他语言都用 snapshotName
       }
       
       return result
   }
   ```

4. **修改属性名称获取方法**：
   ```go
   func (model *SaleOrderProductAttribute) GetLocaleAttributeName() dto.LocaleResponse {
       // 优先使用快照字段
       snapshotName := model.Name
       
       // 如果快照字段为空，降级使用关联表
       if snapshotName == "" {
           if model.ProductAttribute.MultiLanguageName != nil {
               return model.ProductAttribute.MultiLanguageName.GetNames()
           }
           return dto.LocaleResponse{}
       }
       
       // 构建多语言响应
       result := dto.LocaleResponse{ZH: snapshotName}
       
       // 如果关联表数据存在，使用关联表数据填充其他语言
       if model.ProductAttribute.MultiLanguageName != nil {
           multiLang := model.ProductAttribute.MultiLanguageName.GetNames()
           result.TH = multiLang.TH
           result.EN = multiLang.EN
           // ... 其他语言
       } else {
           // 如果关联表数据不存在，所有语言都用快照的主语言填充
           result.TH = snapshotName
           result.EN = snapshotName
           // ... 其他语言都用 snapshotName
       }
       
       return result
   }
   ```

5. **修改免单原因获取方法**：
   ```go
   func (model *SaleOrder) GetLocaleFreeReason() dto.LocaleResponse {
       // 收集所有免单原因
       var reasons []dto.LocaleResponse
       for _, reason := range model.FreeReasons {
           if !reason.IsFreeReason() || reason.IsDelete() {
               continue
           }
           // 优先使用快照字段
           snapshotName := reason.Name
           if snapshotName == "" && reason.MultiLanguageName != nil {
               // 降级使用关联表
               reasons = append(reasons, reason.MultiLanguageName.GetNames())
           } else {
               // 使用快照构建多语言响应
               result := dto.LocaleResponse{ZH: snapshotName}
               if reason.MultiLanguageName != nil {
                   multiLang := reason.MultiLanguageName.GetNames()
                   result.TH = multiLang.TH
                   result.EN = multiLang.EN
                   // ... 其他语言
               } else {
                   result.TH = snapshotName
                   result.EN = snapshotName
                   // ... 其他语言都用 snapshotName
               }
               reasons = append(reasons, result)
           }
       }
       // 添加自定义免单原因
       if model.FreeReason != "" {
           customReason := dto.LocaleResponse{
               ZH: model.FreeReason,
               TH: model.FreeReason,
               EN: model.FreeReason,
               // ... 所有语言都用自定义原因
           }
           reasons = append(reasons, customReason)
       }
       // 合并所有原因
       return getLocaleResponse(reasons, "、")
   }
   ```

6. **修改退菜原因获取方法**：
   ```go
   func (model *SaleOrderProduct) GetLocaleCancelReason() dto.LocaleResponse {
       // 类似免单原因的处理逻辑
       // 优先使用 SaleOrderProductReason.Name 快照字段
   }
   ```

7. **修改自助餐名称获取方法**：
   ```go
   func (model *SaleBill) GetLocaleBuffetNames() dto.LocaleResponse {
       // 优先使用快照字段
       name1 := model.BuffetPackage1Name
       name2 := model.BuffetPackage2Name
       
       // 如果快照字段为空，降级使用关联表
       if name1 == "" && model.BuffetPackage1 != nil {
           name1 = model.BuffetPackage1.MultiLanguageName.ZhName
       }
       if name2 == "" && model.BuffetPackage2 != nil {
           name2 = model.BuffetPackage2.MultiLanguageName.ZhName
       }
       
       // 构建多语言响应
       // 如果关联表存在，使用关联表数据填充其他语言
       // 如果关联表不存在，所有语言都用快照的主语言填充
   }
   ```

8. **修改外卖来源获取方法**：
   ```go
   func (model *SaleBill) GetLocaleOrderSourceName() dto.LocaleResponse {
       // 优先使用快照字段
       snapshotName := model.OrderSourceName
       
       // 如果快照字段为空，降级使用关联表
       if snapshotName == "" && model.OrderSource != nil {
           return model.OrderSource.MultiLanguageName.GetNames()
       }
       
       // 构建多语言响应
       result := dto.LocaleResponse{ZH: snapshotName}
       if model.OrderSource != nil && model.OrderSource.MultiLanguageName != nil {
           multiLang := model.OrderSource.MultiLanguageName.GetNames()
           result.TH = multiLang.TH
           result.EN = multiLang.EN
           // ... 其他语言
       } else {
           result.TH = snapshotName
           result.EN = snapshotName
           // ... 其他语言都用 snapshotName
       }
       
       return result
   }
   ```

9. **修改国籍获取方法**：
   ```go
   func (model *SaleBill) GetLocaleNationalityName() dto.LocaleResponse {
       // 类似外卖来源的处理逻辑
       // 优先使用 SaleBill.NationalityName 快照字段
   }
   ```

5. **数据库迁移脚本**：
   ```sql
   -- 添加免单/退菜原因快照字段
   ALTER TABLE `ttpos_sale_order_product_reason` 
   ADD COLUMN `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因名称快照（单语言），不随后台更新' AFTER `gift_reason_uuid`;
   
   -- 添加自助餐名称快照字段
   ALTER TABLE `ttpos_sale_bill` 
   ADD COLUMN `buffet_package1_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐套餐1名称快照（单语言），不随后台更新' AFTER `buffet_package2_uuid`,
   ADD COLUMN `buffet_package2_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐套餐2名称快照（单语言），不随后台更新' AFTER `buffet_package1_name`;
   
   -- 添加外卖来源名称快照字段
   ALTER TABLE `ttpos_sale_bill` 
   ADD COLUMN `order_source_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '外卖来源名称快照（单语言），不随后台更新' AFTER `order_source_uuid`;
   
   -- 添加国籍名称快照字段
   ALTER TABLE `ttpos_sale_bill` 
   ADD COLUMN `nationality_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '国籍名称快照（单语言），不随后台更新' AFTER `nationality_uuid`;
   ```

6. **数据检查脚本**：
   - 检查 `ttpos_sale_order_product.name` 字段的填充率
   - 检查 `ttpos_sale_order_product.flavor_name` 字段的填充率
   - 检查 `ttpos_sale_order_product_bom.name` 字段的填充率
   - 检查 `ttpos_sale_order_product_attribute.name` 字段的填充率
   - 检查新增字段的填充率（免单原因、退菜原因、自助餐名称、外卖来源、国籍）
   - 识别需要补充数据的订单

7. **数据迁移脚本**（可选）：
   - **策略**：只对之后的订单做处理，历史订单通过降级逻辑兼容
   - **可选迁移**：从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）
   - **迁移范围**：只迁移关联表数据存在的记录，对于已删除的数据，保持快照字段为空（使用降级逻辑）
   - **迁移时机**：可以在系统空闲时执行，不影响正常业务

8. **多语言处理**：
   - **方案选择**：采用"主语言快照 + 关联表补充"的混合方案
   - **快照字段**：保存主语言（中文）作为历史快照
   - **查询逻辑**：
     - 主语言（ZH）：优先使用快照字段
     - 其他语言：优先使用关联表数据，如果关联表不存在或已删除，则使用快照的主语言填充
   - **优势**：
     - ✅ 保证快照完整性（即使数据被删除也能显示）
     - ✅ 尽可能提供多语言支持（关联表存在时）
     - ✅ 新增字段使用单语言字段（与现有设计一致）
     - ✅ 兼容性好（历史数据可以降级处理）

### 线框图/原型（可选）

无需 UI 变更，主要是后端逻辑修复。

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
**创建日期**: 2025-01-27  
**维护者**: 开发团队  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

