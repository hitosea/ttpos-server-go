# ttpos-bmp 内部采购不自动创建发货单 需求文档

> 本文档定义移除 ERPNext 内部销售订单自动创建发货单逻辑的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                            |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/erp-brand-purchase-no-auto-shipment.md](../../../../team/proposals/2025-11/erp-brand-purchase-no-auto-shipment.md) |
| **创建日期**      | 2025-11-19                                                                                                                                      |
| **负责人**        | rikugun                                                                                                                                         |
| **目标 Sprint**   | Sprint 待定                                                                                                                                     |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                                      |
| **任务编号**      | 36978                                                                                                                                           |

---

## 📋 概述

在当前的 ttpos-bmp 内部采购流程中，ERPNext 系统会在内部销售订单（Inter Company Sales Order）提交后自动创建发货单（Delivery Note）。这种自动创建行为缺乏灵活性，导致无法根据实际发货计划控制发货流程，并可能产生不必要的数据冗余。

本需求旨在通过修改 ttpos-bmp 模块的配置或代码逻辑，禁用 ERPNext 端的自动创建行为，同时保留手动创建发货单的接口，以支持按需发货的业务场景。

**重要约束（遵循 erpnext.mdc）**：

- ❌ 不修改 ERPNext 源代码
- ❌ 不使用 ERPNext Server Scripts 功能
- ✅ 通过修改 ttpos-bmp 模块（ttpos-erp）的代码实现

## 🎯 产品对齐

本功能支持以下产品目标：

1. **提升业务灵活性**：允许根据实际库存和发货计划决定何时创建发货单
2. **减少数据冗余**：避免自动创建不必要的发货单，降低维护成本
3. **精细化流程管控**：支持更复杂的发货场景（分批发货、延期发货等）
4. **提高库存准确性**：避免自动发货导致的库存提前出库问题

## 📝 用户故事

**作为** 品牌方 ERP 系统管理员  
**我想** 在内部销售订单（Inter Company Sales Order）提交后，系统不自动创建发货单（Delivery Note）  
**以便于** 根据实际库存和发货计划灵活控制发货流程，避免不必要的发货单产生

---

## 功能需求

### Requirement 1: 禁用内部销售订单自动创建发货单

**用户故事**: 作为 ERP 系统管理员，我想在提交内部销售订单时系统不自动创建发货单，以便于根据实际情况决定发货时机

#### 验收标准

1. **WHEN** 调用 `CreateInnerSaleOrderFromPurchaseOrder` 接口创建并提交内部销售订单 **THEN** ERPNext 系统 **SHALL NOT** 自动创建发货单（Delivery Note）
2. **WHEN** 内部销售订单创建成功 **THEN** 响应 **SHALL** 只包含销售订单信息，不包含自动创建的发货单信息
3. **WHEN** 在 ERPNext 系统中查看已提交的内部销售订单 **THEN** 订单状态 **SHALL** 为 "Submitted"，**SHALL NOT** 自动关联任何发货单
4. **WHEN** 查询内部销售订单详情 **THEN** 系统 **SHALL** 正常返回订单信息，不受配置变更影响
5. **IF** ERPNext 配置了自动创建发货单的工作流或触发器 **THEN** ttpos-bmp **SHALL** 通过配置或代码逻辑禁用该行为

#### 具体要求

- [ ] 1.1 调研 ERPNext Inter Company Transaction 中自动创建 Delivery Note 的触发机制
- [ ] 1.2 识别 ttpos-bmp 中与自动创建 Delivery Note 相关的配置项或代码逻辑
- [ ] 1.3 修改 ttpos-erp 模块的配置或代码，禁用自动创建行为
- [ ] 1.4 确保禁用后不影响其他内部采购流程（Material Request → Purchase Order → Sales Order）
- [ ] 1.5 验证修改后，现有数据和已创建的发货单不受影响

---

### Requirement 2: 保留手动创建发货单接口

**用户故事**: 作为 仓库管理员，我想在需要发货时手动创建发货单，以便于按实际发货计划执行发货

#### 验收标准

1. **WHEN** 需要为内部销售订单创建发货单 **THEN** 用户 **SHALL** 可以调用 `CreateDeliveryNoteFromInnerSaleOrder` 接口手动创建
2. **WHEN** 手动创建发货单 **THEN** 系统 **SHALL** 正常调用 ERPNext API `make_delivery_note`，创建并关联到对应的内部销售订单
3. **WHEN** 手动创建的发货单成功 **THEN** 响应 **SHALL** 包含发货单完整信息（Name, Items, Status 等）
4. **WHEN** 查询已手动创建的发货单 **THEN** 系统 **SHALL** 正常返回发货单详情
5. **IF** 内部销售订单已手动创建了发货单 **THEN** 系统 **SHALL** 允许再次调用接口创建新的发货单（支持分批发货）

#### 具体要求

- [ ] 2.1 确认 `CreateDeliveryNoteFromInnerSaleOrder` 接口保持不变
- [ ] 2.2 验证接口调用 ERPNext 的 `erpnext.selling.doctype.sales_order.sales_order.make_delivery_note` 方法正常工作
- [ ] 2.3 测试手动创建发货单的完整流程（创建 → 提交 → 查询）
- [ ] 2.4 验证发货单正确关联到内部销售订单
- [ ] 2.5 支持分批发货场景（同一销售订单可创建多个发货单）

---

### Requirement 3: 文档和操作流程更新

**用户故事**: 作为 新入职的开发人员，我想查阅更新后的文档，以便于了解新的内部采购流程

#### 验收标准

1. **WHEN** 查阅内部采购流程文档 **THEN** 文档 **SHALL** 明确说明销售订单提交后不自动创建发货单
2. **WHEN** 查阅操作手册 **THEN** 手册 **SHALL** 包含如何手动创建发货单的步骤
3. **WHEN** 查阅 API 文档 **THEN** 文档 **SHALL** 说明 `CreateDeliveryNoteFromInnerSaleOrder` 接口的用法和参数
4. **WHEN** 查看 CHANGELOG **THEN** 日志 **SHALL** 记录此次流程变更

#### 具体要求

- [ ] 3.1 更新 `docs/human/architecture/features/purchase_order.md`（如存在），说明流程变更
- [ ] 3.2 更新 ttpos-bmp API 文档，明确手动创建发货单的接口说明
- [ ] 3.3 在 CHANGELOG.md 中记录此次变更
- [ ] 3.4 创建操作指南（如需要），说明手动发货流程

---

## 非功能需求

### 代码架构和模块化

- **遵循 ERPNext 集成规范**：
  - ❌ 不修改 ERPNext 源代码
  - ❌ 不使用 ERPNext Server Scripts
  - ✅ 通过 ttpos-bmp/ttpos-erp 模块实现
  - ✅ 与 ERPNext 交互通过 `ttpos-erp/internal/logic/erpnext` 下的通用服务
  - ✅ JSON 数据结构的 struct 在 `model/dto` 包中
- **遵循 GoFrame 规范**：
  - 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
  - 禁止修改 dao/entity/do/ 目录（自动生成）
  - gRPC 服务必须注册到 Nacos

### API 设计要求

- [ ] 保持现有 API 接口签名不变
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

本需求 **不涉及数据库结构变更**，无需创建或修改表。

### 性能要求

- [ ] 接口响应时间不受影响（与修改前一致）
- [ ] 禁用自动创建不应增加 API 调用延迟

### 测试要求

- [ ] 集成测试覆盖新的流程（提交销售订单 → 确认无发货单 → 手动创建发货单）
- [ ] 回归测试确保其他内部采购流程（Material Request、Purchase Order）不受影响
- [ ] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 可靠性要求

- [ ] 配置或代码变更后，现有数据不受影响
- [ ] 已创建的发货单仍可正常查询和管理
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制（如配置回滚）

---

## 验收标准

### 功能验收

1. **内部销售订单提交不自动创建发货单**: 调用 `CreateInnerSaleOrderFromPurchaseOrder` 后，ERPNext 系统中无自动创建的 Delivery Note
2. **手动创建发货单正常工作**: 调用 `CreateDeliveryNoteFromInnerSaleOrder` 接口可成功创建发货单
3. **流程完整性**: Material Request → Purchase Order → Sales Order 流程不受影响
4. **数据完整性**: 现有数据和已创建的发货单不受影响

### 测试验收

1. **集成测试**: 完整内部采购流程测试通过（包括手动创建发货单）
2. **回归测试**: 其他 ERP 功能（Material Request、Purchase Order、其他 Delivery Note 创建）测试通过
3. **性能测试**: API 响应时间与修改前一致

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: `CreateDeliveryNoteFromInnerSaleOrder` 接口文档完整
3. **CHANGELOG**: 变更记录已添加
4. **操作手册**: 手动创建发货单流程说明（如需要）

---

## 约束条件

### 技术约束

#### Go BMP 模块（ttpos-bmp）

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- **遵循 `ttpos-bmp/.cursor/rules/erpnext.mdc`**：
  - ❌ 不修改 ERPNext 源代码
  - ❌ 不使用 ERPNext Server Scripts
  - ✅ 通过 ttpos-erp 模块代码实现

#### ERPNext 集成约束

- 不修改 ERPNext 系统源代码
- 不使用 ERPNext 的 Server Scripts 功能
- 所有变更通过 ttpos-bmp/ttpos-erp 模块实现
- 与 ERPNext 交互通过 `ttpos-erp/internal/logic/erpnext` 下的通用服务
- JSON struct 在 `model/dto` 包中

### 业务约束

- 变更后的流程需保持向后兼容（现有数据不受影响）
- 手动创建发货单的操作需易于理解和执行
- 需支持分批发货场景

### 资源约束

- 开发时间: 1-2 天
- Story Point: 2 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` - 内部采购逻辑
  - `CreateInnerSaleOrderFromPurchaseOrder` - 创建内部销售订单
  - `CreateDeliveryNoteFromInnerSaleOrder` - 手动创建发货单
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/delivery_note.go` - 发货单管理逻辑
- ERPNext API:
  - `erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order` - 创建内部销售订单
  - `erpnext.selling.doctype.sales_order.sales_order.make_delivery_note` - 创建发货单

### 服务依赖

- **ttpos-bmp → ERPNext**: HTTP API 调用

### 业务依赖

- 依赖 ERPNext 系统正常运行
- 依赖内部公司（Inter Company）配置正确

---

## 风险和缓解

### 风险 1: ERPNext 自动创建逻辑难以禁用

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 深入研究 ERPNext 文档和源码，找到正确的配置方式
- 如 ERPNext 端无法禁用，考虑在 ttpos-bmp 层面拦截或忽略自动创建的 Delivery Note
- 与 ERPNext 社区或官方咨询最佳实践

### 风险 2: 现有业务流程依赖自动创建的发货单

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 与业务方确认当前是否真的需要自动创建
- 在测试环境充分验证，确保不影响现有数据和流程
- 制定回滚方案，如需要可快速恢复原有行为

### 风险 3: 用户操作习惯变更导致遗漏发货

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 制定操作手册，培训相关人员
- 建立发货提醒机制（如定时检查未发货的销售订单）
- 在 ERP 界面添加手动创建发货单的快捷入口

### 风险 4: 修改配置或代码影响其他 ERP 功能

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 充分的回归测试，确保其他功能不受影响
- 仅修改与内部销售订单相关的逻辑，避免影响外部销售订单
- 保留详细的日志记录，便于问题排查

---

## 时间表

- **Phase 1 - 调研和分析**: 0.5 天
  - 调研 ERPNext Inter Company Transaction 自动创建 Delivery Note 的机制
  - 分析 ttpos-bmp 中相关代码和配置
- **Phase 2 - 实现和测试**: 1 天
  - 修改配置或代码，禁用自动创建
  - 验证手动创建接口正常工作
  - 集成测试和回归测试
- **Phase 3 - 文档和部署**: 0.5 天
  - 更新相关文档
  - 部署到测试环境验证
  - 准备生产环境部署方案
- **总计**: 2 天（SP = 2）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- **`ttpos-bmp/.cursor/rules/erpnext.mdc` - ERPNext 集成规范（核心约束）**
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/documentation.mdc` - 文档创建规范

### 架构文档

- `docs/human/architecture/features/purchase_order.md` - 采购订单架构文档
- ERPNext 官方文档:
  - https://docs.erpnext.com/docs/user/manual/en/accounts/inter-company-invoices
  - https://docs.erpnext.com/docs/user/manual/en/selling/sales-order
  - https://docs.erpnext.com/docs/user/manual/en/stock/delivery-note

### 相关代码模块

- `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - `CreateInnerSaleOrderFromPurchaseOrder` (lines 86-151)
  - `CreateDeliveryNoteFromInnerSaleOrder` (lines 153-186)
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/delivery_note.go`
  - 发货单管理逻辑

### 外部参考

- ERPNext Inter Company Transaction: https://docs.erpnext.com/docs/user/manual/en/accounts/inter-company-invoices
- ERPNext Sales Order: https://docs.erpnext.com/docs/user/manual/en/selling/sales-order
- ERPNext Delivery Note: https://docs.erpnext.com/docs/user/manual/en/stock/delivery-note

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**作者**: rikugun  
**审核者**: 待定
