# ttpos-bmp 内部采购不再自动创建发货单 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目          | 内容                                                                                                                         |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **提案人**    | rikugun                                                                                                                      |
| **日期**      | 2025-11-19                                                                                                                   |
| **目标版本**  | v2.1.0                                                                                                                       |
| **状态**      | ✅ 已批准 → 已创建 Spec                                                                                                      |
| **关联 Spec** | [story-bmp-inter-company-no-auto-delivery-note](../../../shared/specs/active/story-bmp-inter-company-no-auto-delivery-note/) |
| **任务编号**  | 36978                                                                                                                        |

---

## 🎯 背景和动机

### 问题描述

当前 ttpos-bmp 的内部采购流程中，ERPNext 系统在内部销售订单（Inter Company Sales Order）提交后会自动创建发货单（Delivery Note）。

**当前内部采购完整流程**：

```
1. Material Request（物料请求单）
   ↓ 自动
2. Purchase Order（采购订单）
   ↓ 调用 CreateInnerSaleOrderFromPurchaseOrder
3. Inter Company Sales Order（内部销售订单）
   ↓ ERPNext 自动创建 ✗ 需要移除
4. Delivery Note（发货单）
```

这种自动创建发货单的行为在某些业务场景下不够灵活，导致：

1. **业务灵活性不足**：无法在创建内部销售订单后，根据实际库存和发货计划决定是否立即创建发货单
2. **流程控制缺失**：销售订单提交后立即创建发货单，缺少审核和调整的空间
3. **数据冗余**：自动创建的发货单可能因业务变更（如订单取消、延期发货）而需要删除，增加维护成本
4. **库存管理问题**：自动发货可能导致库存提前出库，影响库存准确性

### 业务价值

移除自动创建发货单的逻辑能带来以下价值：

- 提升业务流程灵活性，支持按实际发货计划创建发货单
- 减少无效发货单的产生，降低数据维护成本
- 支持更精细化的内部销售订单和发货流程管控
- 满足不同业务场景的需求（如分批发货、延期发货等）
- 提高库存管理的准确性

### 目标用户

谁会使用这个功能？

- [ ] 收银员
- [x] 商户管理员（总部）
- [x] 品牌方运营人员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: ttpos-bmp ERP 系统管理员、仓库管理员

---

## 💡 解决方案概述

### 方案描述

修改 ERPNext 系统配置，禁用内部销售订单自动创建发货单的逻辑：

**当前流程**：

```
Inter Company Sales Order 提交(Submit) → ERPNext 自动创建 Delivery Note
```

**目标流程**：

```
Inter Company Sales Order 提交(Submit) → 完成（不创建发货单）
需要发货时 → 手动调用 CreateDeliveryNoteFromInnerSaleOrder 接口
```

**实现方式**：

1. **修改 ERPNext 配置**：

   - 禁用 Inter Company Transaction 的自动创建 Delivery Note 功能
   - 或修改 Sales Order 的工作流，移除自动创建 Delivery Note 的步骤

2. **保留手动创建接口**：

   - `CreateDeliveryNoteFromInnerSaleOrder` 接口保持不变
   - 需要发货时，可以手动调用该接口创建发货单

3. **文档更新**：
   - 更新内部采购流程文档
   - 更新操作手册，说明如何手动创建发货单

### 核心功能点

1. 禁用 ERPNext 端 Inter Company Sales Order 自动创建 Delivery Note 的配置
2. 保留 `CreateDeliveryNoteFromInnerSaleOrder` 接口，支持手动创建发货单
3. 更新相关文档和操作流程
4. 测试修改后对现有业务的影响
5. 确保已创建的数据不受影响

### 影响范围

**涉及模块**：

- [ ] POS 收银端
- [x] Shop 商家管理端（间接影响：总部采购和发货管理）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [x] ttpos-bmp 模块（主要影响）
  - `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - `CreateDeliveryNoteFromInnerSaleOrder` 接口（保留，用于手动创建）
- [x] ERPNext 系统配置（主要修改点）
  - Inter Company Transaction 设置
  - Sales Order 工作流配置

**涉及功能**：

- [x] 内部采购流程
- [x] 内部销售订单管理
- [x] 发货单管理
- [x] ERPNext 工作流配置

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要修改 ERPNext 配置，涉及第三方系统
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**复杂度分析**：

- 需要了解 ERPNext 的 Inter Company Transaction 机制
- 需要修改 ERPNext 端的配置或自定义脚本
- 需要测试修改后对现有内部采购流程的影响
- 涉及 ttpos-bmp 和 ERPNext 两个系统的协调

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 2（待技术评审确认）

**工作量分解**：

1. 调研 ERPNext Inter Company Transaction 自动创建 Delivery Note 的机制（0.5 天）
2. 修改 ERPNext 配置，禁用自动创建（0.5 天）
3. 测试验证和文档更新（0.5 天）

### 风险识别

**潜在风险**：

1. **ERPNext 配置风险**：ERPNext 的自动创建逻辑可能深度集成在系统中，难以完全禁用
2. **业务依赖风险**：现有流程可能依赖自动创建的发货单，需要评估影响
3. **操作习惯变更**：用户需要适应新的手动创建发货单流程
4. **遗漏发货风险**：取消自动创建后，可能出现忘记创建发货单导致未发货的情况

**缓解措施**：

1. 深入研究 ERPNext 文档和源码，找到正确的配置方式
2. 与业务方确认当前流程中是否真的需要自动创建发货单
3. 在测试环境充分验证，确保不影响现有数据和流程
4. 保留 `CreateDeliveryNoteFromInnerSaleOrder` 接口，支持手动创建
5. 制定操作手册，培训相关人员
6. 建立发货提醒机制，避免遗漏

---

## 🔗 相关资源

### 参考文档

- ERPNext Inter Company Transaction: https://docs.erpnext.com/docs/user/manual/en/accounts/inter-company-invoices
- ERPNext Sales Order 文档: https://docs.erpnext.com/docs/user/manual/en/selling/sales-order
- ERPNext Delivery Note 文档: https://docs.erpnext.com/docs/user/manual/en/stock/delivery-note

### 相关代码模块

- 创建内部销售订单: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` (lines 86-151)
  - `CreateInnerSaleOrderFromPurchaseOrder` 方法
- 手动创建发货单: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` (lines 153-186)
  - `CreateDeliveryNoteFromInnerSaleOrder` 方法（保留用于手动创建）
- 发货单管理逻辑: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/delivery_note.go`

### 相关接口

- `CreateInnerSaleOrderFromPurchaseOrder`: 从采购订单创建内部销售订单
- `CreateDeliveryNoteFromInnerSaleOrder`: 从内部销售订单手动创建发货单（保留）
- ERPNext API: `erpnext.selling.doctype.sales_order.sales_order.make_delivery_note`

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名    | 签名/日期 |
| ---------- | ------- | --------- |
| 产品经理   | 待定    |           |
| 技术负责人 | 待定    |           |
| 开发代表   | rikugun |           |
| 测试代表   | 待定    |           |
| 业务代表   | 待定    |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`story-bmp-inter-company-no-auto-delivery-note`（已完成 2025-11-19）
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint 待定
- [ ] 开始开发：按照 `docs/shared/specs/archived/v2.10.0/story-bmp-inter-company-no-auto-delivery-note/tasks.md` 执行

---

## 📝 附录

### User Story（初稿）

**作为** 品牌方 ERP 系统管理员  
**我想** 在内部销售订单（Inter Company Sales Order）提交后，系统不自动创建发货单（Delivery Note）  
**以便于** 根据实际库存和发货计划灵活控制发货流程，避免不必要的发货单产生

### AC 验收标准（初稿）

1. **WHEN** 内部销售订单（Inter Company Sales Order）提交 **THEN** ERPNext 系统 **SHALL NOT** 自动创建发货单（Delivery Note）
2. **WHEN** 需要创建发货单 **THEN** 用户 **SHALL** 可以调用 `CreateDeliveryNoteFromInnerSaleOrder` 接口手动创建
3. **WHEN** 手动创建发货单 **THEN** 系统 **SHALL** 正常创建并关联到对应的内部销售订单
4. **WHEN** 查询内部销售订单 **THEN** 系统 **SHALL** 正常返回订单信息，不受影响
5. **WHEN** 在 ERPNext 系统中查看内部销售订单 **THEN** 订单状态 **SHALL** 为 Submitted，无自动关联的发货单

### 技术实现要点（初稿）

1. **ERPNext 配置调整**：

   - 方案 A：在 ERPNext 系统设置中禁用 Inter Company Transaction 的自动创建 Delivery Note
   - 方案 B：修改 Sales Order 的工作流，移除自动创建 Delivery Note 的步骤
   - 方案 C：通过自定义脚本覆盖默认行为

2. **配置示例**（待确认）：

   ```python
   # ERPNext 自定义脚本示例（Server Script）
   # Doctype: Sales Order
   # Event: on_submit

   def on_submit(doc, method):
       # 如果是内部公司交易，不自动创建发货单
       if doc.is_internal_customer:
           # 跳过自动创建 Delivery Note 的逻辑
           pass
   ```

3. **保留的接口**：
   ```go
   // ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go
   // lines 153-186
   func (*sBuying) CreateDeliveryNoteFromInnerSaleOrder(
       ctx context.Context,
       req *dto.CreateDeliveryNoteFromInnerSaleOrderReq
   ) (res *erp.DeliveryNote, err error) {
       // 调用 ERPNext API 手动创建发货单
       // method: erpnext.selling.doctype.sales_order.sales_order.make_delivery_note
       // 此接口保留，用于手动创建发货单
   }
   ```

### 操作流程变更

**变更前**：

```
1. 创建 Inter Company Sales Order
2. 提交 → 系统自动创建 Delivery Note ✗
3. 确认发货单
```

**变更后**：

```
1. 创建 Inter Company Sales Order
2. 提交 → 完成（不创建发货单）✓
3. 需要发货时：
   - 手动调用 CreateDeliveryNoteFromInnerSaleOrder 接口
   - 或在 ERPNext 界面手动创建 Delivery Note
4. 确认发货单
```

### 线框图/原型（可选）

[附加 UI 线框图或原型链接]

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
