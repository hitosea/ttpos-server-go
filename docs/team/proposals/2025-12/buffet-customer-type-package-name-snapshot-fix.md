# 自助餐顾客类型套餐名称快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan |
| **日期**   | 2025-12-09 |
| **目标版本** | v2.11.0 |
| **状态**   | 已批准 → Spec 已创建   |
| **关联任务** | -      |
| **关联 Spec** | [story-main-buffet-customer-type-package-name-snapshot-fix](../../shared/specs/active/story-main-buffet-customer-type-package-name-snapshot-fix/) |
| **父提案** | `order-attribute-snapshot-fix.md` |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，`ttpos_sale_order_buffet_customer_type` 表中的自助餐套餐名称信息会随后台数据变更而改变，导致订单历史信息不准确。这是订单商品信息快照修复需求（`order-attribute-snapshot-fix.md`）的子任务。

**具体场景**：

1. **自助餐套餐被删除**：
   - 订单中 `SaleOrderBuffetCustomerType` 记录关联了 `buffet_package_uuid = 123`（"豪华自助餐"）
   - 后台删除了 UUID 为 123 的自助餐套餐配置
   - 查询订单时，虽然 `SaleBill` 有快照字段，但如果该套餐不在 `BuffetPackage1Uuid` 或 `BuffetPackage2Uuid` 中，或者套餐被删除后关联查询失败，自助餐名称信息可能丢失或显示错误

2. **自助餐套餐被改名**：
   - 订单中 `SaleOrderBuffetCustomerType` 记录关联了 `buffet_package_uuid = 123`（"豪华自助餐"）
   - 后台将 UUID 为 123 的自助餐套餐改名为"超值自助餐"
   - 查询订单时显示："超值自助餐"（显示的是新名称，而非下单时的名称）

3. **数据一致性问题**：
   - `SaleOrderBuffetCustomerType` 表中有 `buffet_package_uuid` 字段，但没有对应的名称快照字段
   - 目前查询时依赖 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 方法，该方法需要 `SaleBill` 已加载且套餐 UUID 匹配 `BuffetPackage1Uuid` 或 `BuffetPackage2Uuid`
   - 如果 `SaleBill` 未加载或套餐 UUID 不匹配，则降级使用关联表数据，可能获取到错误或已删除的数据

**问题影响**：

- ❌ 订单历史信息不准确，无法还原下单时的真实状态
- ❌ 影响对账、统计报表等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 自助餐套餐被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能
- ❌ `SaleOrderBuffetCustomerType` 记录本身缺少快照字段，依赖外部快照可能导致数据不一致

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保 `SaleOrderBuffetCustomerType` 记录的自助餐名称准确反映下单时的状态
- ✅ **数据独立性**：`SaleOrderBuffetCustomerType` 记录不依赖 `SaleBill` 的快照字段，可以独立获取自助餐名称
- ✅ **合规性**：满足财务、税务对订单历史记录的要求
- ✅ **可追溯性**：支持订单历史查询和问题追溯
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

**核心思路**：在 `ttpos_sale_order_buffet_customer_type` 表中添加 `buffet_package_name` 字段（TEXT 类型），保存自助餐套餐名称的多语言 JSON 快照。

**现状分析**：

1. **数据库设计缺失快照字段**：
   - `ttpos_sale_order_buffet_customer_type` 表有 `buffet_package_uuid` 字段，但没有对应的名称快照字段
   - 查询时依赖 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 方法，该方法需要 `SaleBill` 已加载且套餐 UUID 匹配

2. **代码实现依赖外部快照**：
   - 目前查询时通过 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 获取自助餐名称
   - 如果 `SaleBill` 未加载或套餐 UUID 不匹配，降级使用关联表数据（`BuffetPackage.MultiLanguageName`）
   - 导致相关数据被删除或改名时，订单显示信息变化

**解决方案**：

1. **数据库结构变更**：
   - 在 `ttpos_sale_order_buffet_customer_type` 表添加 `buffet_package_name` 字段（TEXT 类型），用于保存自助餐套餐名称的多语言 JSON 快照

2. **Go Model 修改**：
   - 在 `SaleOrderBuffetCustomerType` 结构体中添加 `BuffetPackageName` 字段
   - 实现 `GetLocaleBuffetPackageName()` 方法：优先使用快照字段，降级使用关联表数据
   - 实现 `SetBuffetPackageNameSnapshot()` 方法：从 `BuffetPackage.MultiLanguageName` 序列化为 JSON 保存

3. **查询逻辑修改**：
   - 修改所有使用 `SaleOrderBuffetCustomerType` 的地方，优先使用 `GetLocaleBuffetPackageName()` 方法
   - 不再依赖 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 方法

4. **下单逻辑修改**：
   - 确保创建 `SaleOrderBuffetCustomerType` 时保存自助餐套餐名称快照
   - 修改 `NewSaleOrderBuffetCustomerType()` 和相关创建方法

5. **多语言支持**：
   - 快照字段保存多语言 JSON（与自助餐名称快照修复方案一致）
   - 查询时优先使用快照字段，如果快照字段为空或无效，降级使用关联表数据

### 核心功能点

1. **数据库结构变更**
   - 在 `ttpos_sale_order_buffet_customer_type` 表添加 `buffet_package_name` 字段（TEXT 类型），用于保存自助餐套餐名称的多语言 JSON 快照

2. **Go Model 修改**
   - 在 `SaleOrderBuffetCustomerType` 结构体中添加 `BuffetPackageName` 字段
   - 实现 `GetLocaleBuffetPackageName()` 方法：优先使用快照字段，降级使用关联表数据
   - 实现 `SetBuffetPackageNameSnapshot()` 方法：从 `BuffetPackage.MultiLanguageName` 序列化为 JSON 保存

3. **查询逻辑修改**
   - 修改所有使用 `SaleOrderBuffetCustomerType` 的地方，使用 `GetLocaleBuffetPackageName()` 方法获取自助餐名称
   - 替换现有的 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 调用

4. **下单逻辑修改**
   - 修改 `NewSaleOrderBuffetCustomerType()` 方法，保存自助餐套餐名称快照
   - 修改 `SaleOrder.GetSaleOrderBuffetCustomerTypes()` 方法，创建 `SaleOrderBuffetCustomerType` 时保存快照
   - 修改 `CreateDeskOrder` 和 `OrderChangeBuffet` 等下单/修改逻辑

5. **数据迁移和兼容性**
   - 历史订单的快照字段为空时，降级使用关联表数据（兼容性处理）
   - 可选：提供数据迁移脚本，从关联表补充历史订单的快照字段

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
- [x] 数据模型（`SaleOrderBuffetCustomerType`）
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

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**任务分解**：
1. **数据库结构变更**（0.5 天）
   - 设计数据库迁移脚本
   - 添加快照字段到 `ttpos_sale_order_buffet_customer_type` 表
   - 执行迁移并验证

2. **代码修改 - Model 层**（0.5 天）
   - 添加 `BuffetPackageName` 字段到 `SaleOrderBuffetCustomerType` 结构体
   - 实现 `GetLocaleBuffetPackageName()` 方法
   - 实现 `SetBuffetPackageNameSnapshot()` 方法

3. **代码修改 - 查询逻辑**（0.5 天）
   - 修改所有使用 `SaleOrderBuffetCustomerType` 的地方，使用 `GetLocaleBuffetPackageName()` 方法
   - 替换现有的 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 调用

4. **代码修改 - 下单逻辑**（0.5 天）
   - 修改 `NewSaleOrderBuffetCustomerType()` 方法
   - 修改 `SaleOrder.GetSaleOrderBuffetCustomerTypes()` 方法
   - 修改 `CreateDeskOrder` 和 `OrderChangeBuffet` 等下单/修改逻辑

5. **测试验证**（0.5 天）
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

3. **多语言支持问题**
   - **问题**：快照字段需要保存多语言 JSON，与自助餐名称快照修复方案保持一致
   - **影响**：需要设计合理的多语言快照方案
   - **解决方案**：采用 JSON 格式保存多语言数据（`dto.LocaleResponse`），与自助餐名称快照修复方案一致

4. **下单逻辑修改风险**
   - 风险：需要修改下单逻辑，确保保存快照字段
   - 影响：可能遗漏某些下单场景，导致快照数据不完整
   - 缓解：全面梳理所有下单入口，确保都保存快照数据

5. **性能影响**
   - 风险：如果快照字段为空，降级查询可能增加数据库查询
   - 影响：需要优化查询逻辑，优先使用快照数据

6. **回归风险**
   - 风险：修改核心方法可能影响其他功能（订单查询、打印、导出等）
   - 影响：需要充分测试，特别是订单相关的所有功能

**缓解措施**：

1. **数据检查**：
   - 先检查历史数据的所有快照字段填充情况
   - 根据检查结果决定是否需要数据迁移

2. **兼容性处理**：
   - 实现降级逻辑，确保历史订单正常显示
   - 逐步迁移，不强制要求所有数据立即完整

3. **多语言处理**：
   - 采用 JSON 格式保存多语言数据（`dto.LocaleResponse`）
   - 快照字段保存完整的多语言 JSON
   - 查询时优先使用快照字段，如果快照字段为空或无效，降级使用关联表数据

4. **性能优化**：
   - 优先使用快照数据，减少关联查询
   - 如果必须降级查询，使用索引优化

5. **充分测试**：
   - 编写单元测试覆盖所有修改的方法和场景
   - 测试多语言快照逻辑（关联表存在/不存在的情况）
   - 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
   - 在测试环境充分验证后再上线

6. **全面梳理**：
   - 梳理所有使用 `SaleOrderBuffetCustomerType` 的地方
   - 确保所有相关方法都使用快照数据

---

## 🔗 相关资源

### 参考需求

- 父提案: `order-attribute-snapshot-fix.md`
- 类似功能: `buffet-package-name-snapshot-fix.md`（自助餐名称快照修复）

### 相关文档

- 订单信息获取逻辑分析: `docs/shared/api/cashier-order-info-analysis.md`
- 数据模型定义: `main/app/model/sale_order_buffet_customer_type.go`
- 自助餐名称快照修复 Spec: `docs/shared/specs/active/story-main-buffet-package-name-snapshot-fix/`

### 代码位置

**问题代码**：
- `main/app/service/order_manage.go:459` - `GetOrderInfos()` 方法中使用 `SaleBill.GetLocaleBuffetPackageNameByUuid()`
- `main/app/service/order.go:2777` - `checkBuffetCustomerTypePriceChanged()` 方法中使用 `SaleBill.GetLocaleBuffetPackageNameByUuid()`
- `main/app/model/sale_order.go:670` - `GetCustomerList()` 方法中使用 `SaleBill.GetLocaleBuffetPackageNameByUuid()`

**数据模型**：
- `main/app/model/sale_order_buffet_customer_type.go:12-44` - `SaleOrderBuffetCustomerType` 模型定义（需要新增 `BuffetPackageName` 字段）
- `main/app/model/sale_order.go:1178` - `NewSaleOrderBuffetCustomerType()` 方法（需要保存快照）
- `main/app/model/sale_order.go:1291` - `NewSaleOrderBuffetCustomerType()` 函数（需要保存快照）
- `main/app/model/sale_order_ext_getset.go:628` - `GetSaleOrderBuffetCustomerTypes()` 方法（需要保存快照）

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

- [ ] 创建 Spec：`story-main-buffet-customer-type-package-name-snapshot-fix`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实自助餐套餐名称  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的自助餐套餐名称  
**以便于** 准确处理退款和客户咨询

### AC 验收标准（初稿）

1. **WHEN** 查询包含 `SaleOrderBuffetCustomerType` 的订单 **THEN** 系统 **SHALL** 显示下单时保存的自助餐套餐名称快照
2. **IF** 后台删除了某个自助餐套餐 **THEN** 历史订单 **SHALL** 仍然显示该套餐的原始名称
3. **IF** 后台修改了某个自助餐套餐的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称
4. **IF** 订单快照数据为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据（兼容性）
5. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 正确保存 `SaleOrderBuffetCustomerType` 的自助餐套餐名称快照
6. **WHEN** 查询订单商品信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
7. **IF** 关联表数据存在 **THEN** 系统 **SHALL** 使用关联表数据填充其他语言（TH、EN等）
8. **IF** 关联表数据不存在（已删除） **THEN** 系统 **SHALL** 使用快照的多语言数据填充所有语言字段

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**维护者**: 开发团队  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

