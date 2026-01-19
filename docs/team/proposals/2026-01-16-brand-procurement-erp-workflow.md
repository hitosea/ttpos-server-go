# 品牌采购 ERP 业务流程完整提案

> 品牌采购在 ERPNext 中的完整业务流程梳理，包括集采和直采的详细操作流程、拆单规则、TTPOS 侧收货操作及对应的 ERP 功能。

---

## 📋 目录

1. [业务概述](#1-业务概述)
2. [集采完整流程](#2-集采完整流程)
3. [直采完整流程](#3-直采完整流程)
4. [集采拆单规则](#4-集采拆单规则)
5. [TTPOS 侧收货操作](#5-ttpos-侧收货操作)
6. [在途仓管理](#6-在途仓管理)
7. [ERP 功能映射](#7-erp-功能映射)
8. [实施建议](#8-实施建议)

---

## 1. 业务概述

### 1.1 品牌采购分类

品牌采购分为两个部分：

| 类型 | 业务模式 | 触发条件 | ERPNext 单据类型 |
|------|---------|---------|-----------------|
| **集采** | 总部集中采购后配送给门店 | MR 审批后，物品**未勾选默认供应商**（或默认供应商是总部） | Sales Order → Delivery Note |
| **直采** | 外部供应商直接配送给门店 | MR 审批后，物品**勾选了默认供应商**（外部供应商） | Purchase Order → Purchase Receipt |

### 1.2 核心判断逻辑

```
MR 审批通过
   ↓
遍历 MR 中的每个物品
   ↓
判断物品是否勾选了默认供应商？
   ├─ 是（外部供应商）→ 走直采流程，创建 Purchase Order
   └─ 否（或默认供应商是总部）→ 走集采流程，创建 Sales Order（BOI）
```

**关键字段**：
- Material Request Item 中的 `supplier` 字段（默认供应商）
- 如果 `supplier` 有值且不是总部 → 直采
- 如果 `supplier` 为空或是总部 → 集采

---

## 2. 集采完整流程

### 2.1 集采流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                        集采完整流程                               │
└─────────────────────────────────────────────────────────────────┘

阶段1：申请和审批
├─ 1. 门店创建 MR 申请
│    └─ ERPNext: Material Request (Draft)
│
├─ 2. MR 审批通过
│    └─ ERPNext: Material Request (Submitted)
│    └─ 系统检测：物品未勾选默认供应商 → 自动走集采流程
│
└─ 3. 自动创建 BOI（集采订单）
     └─ ERPNext: Sales Order (Inter Company Sales Order)
     └─ 客户：门店
     └─ 公司：总部（作为销售方）
     └─ 状态：Draft

阶段2：审批和发货准备
├─ 4. 审批 BOI
│    └─ ERPNext: Sales Order (Submitted)
│
└─ 5. 总部仓库人员按仓库拆分创建 Delivery Note
     ├─ 系统自动按仓库分组物品
     ├─ 仓库 A 的物品 → Delivery Note A
     ├─ 仓库 B 的物品 → Delivery Note B
     └─ 每个 Delivery Note 关联 Sales Order
     └─ 状态：Draft

阶段3：发货和收货
├─ 6. 提交 Delivery Note
│    ├─ ERPNext: Delivery Note (Submitted)
│    ├─ 库存从对应的总部仓库扣减
│    └─ 自动在 TTPOS 创建采购收货单（PR）
│
├─ 7. 司机配送货物到门店
│    └─ 不同仓库的物品可能由不同的司机或批次配送
│
└─ 8. 门店在 TTPOS 确认收货
     ├─ 选择对应的采购收货单（关联 Delivery Note）
     ├─ 扫码收货或手动收货
     ├─ 确认收货数量
     └─ 更新库存（门店仓库增加）

阶段4：财务结算
└─ 9. 财务处理
     ├─ 总部创建 Sales Invoice（销售发票）
     └─ 门店创建 Purchase Invoice（采购发票）
```

### 2.2 详细步骤说明

#### 步骤 1-3：申请和创建订单

**操作者**：门店人员 → 采购部门

1. **门店创建 MR 申请**
   - TTPOS 侧：创建采购申请单
   - ERPNext：创建 Material Request
   - 物品信息：可以不设置默认供应商（或设置为总部）
   - 状态：Draft

2. **MR 审批**
   - 操作者：采购部门
   - ERPNext：Material Request 状态 → Submitted
   - 系统自动判断：物品未勾选默认供应商 → 走集采流程

3. **自动创建 BOI（集采订单）**
   - ERPNext：创建 Inter Company Sales Order
   - 客户：门店（作为客户）
   - 公司：总部（作为销售方）
   - 关联：Material Request
   - 状态：Draft

#### 步骤 4-5：审批和发货准备

**操作者**：采购部门 → 总部仓库人员

4. **审批 BOI**
   - 操作者：采购部门
   - ERPNext：Sales Order 状态 → Submitted
   - 审批通过后可以发货

5. **按仓库拆分创建 Delivery Note**
   - 操作者：总部仓库人员
   - 系统自动按仓库分组物品
   - 为每个仓库创建独立的 Delivery Note
   - 每个 Delivery Note 只包含来自同一仓库的物品
   - 详细拆单规则见 [4. 集采拆单规则](#4-集采拆单规则)

#### 步骤 6-8：发货和收货

**操作者**：总部仓库人员 → 司机 → 门店收货员

6. **提交 Delivery Note**
   - 操作者：总部仓库人员
   - ERPNext：Delivery Note 状态 → Submitted
   - 库存从对应的总部仓库扣减
   - **自动触发**：系统在 TTPOS 中创建对应的采购收货单（PR）
   - 采购收货单关联 Delivery Note（通过 `ErpOrderNo` 字段）
   - **在途仓操作**：**物品自动添加到门店在途仓库**
     - 时机：Delivery Note 提交后
     - 操作：系统自动将 Delivery Note 中的物品添加到门店在途仓库
     - 数量：使用 Delivery Note 中的发货数量
     - 目的：表示货物已从总部仓库发出，正在运输途中
     - 详细说明见 [在途仓管理](#在途仓管理)

7. **司机配送**
   - 不同仓库的物品可能由不同的司机或批次配送
   - 每个 Delivery Note 对应一个仓库的发货

8. **门店收货**

   **操作者**：门店收货员  
   **操作位置**：TTPOS 系统（门店在 TTPOS 完成收货操作）

   **TTPOS 侧操作**：
   - 在品采收货中看到发货单对应的采购收货单
   - 选择采购收货单 → 扫码收货或手动收货 → 确认收货数量
   - 点击"确认收货"按钮

   **对应的 ERPNext 操作**（自动同步）：

   | TTPOS 操作 | ERPNext 自动执行的操作 | 同步时机 | 同步方式 |
   |-----------|---------------------|---------|---------|
   | 门店在 TTPOS 查看发货单详情 | 查询 Delivery Note 信息 | 实时查询 | 通过 `ErpOrderNo` 字段关联查询 |
   | 门店扫码收货或手动输入收货数量 | - | - | TTPOS 本地操作，暂不同步 |
   | 门店确认收货，提交收货单 | **自动创建 Inter Company Purchase Receipt**（可选） | **提交收货单时** | **调用 ERPNext API** |
   | TTPOS 更新收货单状态为"已收货" | **Purchase Receipt 状态自动变为 Submitted**（如果创建） | **创建 Purchase Receipt 后** | **自动提交** |
   | TTPOS 更新本地库存 | **Delivery Note 的 `delivered_qty` 自动更新** | **Purchase Receipt 提交后** | **ERPNext 自动执行** |
   | TTPOS 更新采购单进度 | **Sales Order 的 `per_delivered` 自动更新** | **Delivery Note 更新后** | **ERPNext 自动执行** |

   **同步流程**：
   ```
   门店在 TTPOS 点击"确认收货"
      ↓
   1. TTPOS 创建/更新采购收货单（本地操作）
      - 状态：待收货 → 已收货
      - 关联 Delivery Note（通过 ErpOrderNo）
      ↓
   2. 【可选】TTPOS 自动调用 ERPNext API（同步操作）
      ├─ API: POST /api/resource/Purchase Receipt
      ├─ 方法: make_inter_company_purchase_receipt
      ├─ 关联: against_delivery_note = {Delivery Note No}
      ├─ 供应商: 总部供应商（自动填充）
      ├─ items: 收货物品明细（从 TTPOS 收货单获取）
      └─ 创建: Inter Company Purchase Receipt（状态: Draft）
      ↓
   3. 自动提交 Purchase Receipt（如果创建）
      ├─ API: POST /api/resource/Purchase Receipt/{name}
      ├─ action: "submit"
      └─ 状态: Draft → Submitted
      ↓
   4. ERPNext 自动执行后续操作
      ├─ 更新 Delivery Note 的 delivered_qty（自动）
      ├─ 更新 Delivery Note 的 per_delivered（自动）
      ├─ 更新 Sales Order 的 per_delivered（自动）
      └─ 库存已从总部仓库扣减（Delivery Note 提交时已扣减）
      ↓
   5. TTPOS 更新本地数据
      ├─ 更新收货单的 ErpOrderNo（关联 Purchase Receipt，如果创建）
      ├─ 更新本地库存（从在途仓库转入目标仓库）
      └─ 更新采购单进度
      ↓
   6. 同步完成
      ✅ TTPOS 收货单状态：已收货
      ✅ ERPNext Purchase Receipt 状态：Submitted（如果创建）
      ✅ Delivery Note 的 delivered_qty 已更新（如果创建了 Purchase Receipt）
      ✅ 库存已更新（TTPOS 和 ERPNext）
   ```

   **同步性说明**：
   - ✅ **实时查询**：查看 Delivery Note 信息时实时查询，不持久化
   - ✅ **可选同步**：创建 Inter Company Purchase Receipt 是可选的，根据业务需求配置
   - ✅ **自动执行**：ERPNext 侧的操作由系统自动完成，无需门店人员手动操作
   - ✅ **原子性保证**：如果 ERPNext API 调用失败，TTPOS 收货操作会回滚或标记为待同步
   - ✅ **一致性保证**：TTPOS 和 ERPNext 的数据保持一致，通过 `ErpOrderNo` 字段关联

   **详细操作见 [5. TTPOS 侧收货操作](#5-ttpos-侧收货操作) 和 [5.1.4 TTPOS 操作与 ERP 操作对应关系（集采收货）](#514-ttpos-操作与-erp-操作对应关系集采收货)**

#### 步骤 9：财务结算

**操作者**：财务部门

9. **财务处理**
   - 总部创建 Sales Invoice（销售发票）
   - 门店创建 Purchase Invoice（采购发票）
   - 记录应收账款和应付账款

---

## 3. 直采完整流程

### 3.1 直采流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                        直采完整流程                               │
└─────────────────────────────────────────────────────────────────┘

阶段1：申请和审批
├─ 1. 门店创建 MR 申请
│    └─ ERPNext: Material Request (Draft)
│    └─ 物品信息：勾选默认供应商（外部供应商）
│
├─ 2. MR 审批通过
│    └─ ERPNext: Material Request (Submitted)
│    └─ 系统检测：物品勾选了默认供应商 → 自动走直采流程
│
└─ 3. 采购部门创建直采 PO
     └─ ERPNext: Purchase Order
     └─ 供应商：外部供应商（从物品的默认供应商获取）
     └─ 状态：Draft
     └─ 自动在 TTPOS 生成采购单

阶段2：审批和下单
├─ 4. 审批 PO
│    └─ ERPNext: Purchase Order (Submitted)
│    └─ 根据金额自动进入审批流程
│
├─ 5. 打印采购 PDF
│    └─ 调用 ERPNext API 获取 PDF
│
└─ 6. 提交外部供应商
     └─ 外部供应商接收 PO → 准备货物 → 配送

阶段3：收货
├─ 7. 外部供应商配送货物到门店
│
├─ 8. 门店在 TTPOS 创建收货单
│    ├─ 选择直采订单（PO）
│    ├─ 系统自动按供应商拆分收货单
│    └─ 选择收货物品
│
└─ 9. 门店确认收货
     ├─ 扫码收货或手动收货
     ├─ 确认收货数量
     ├─ TTPOS 自动为每个供应商创建 Purchase Receipt
     └─ 更新库存（门店仓库增加）

阶段4：财务结算
└─ 10. 财务处理
     ├─ 门店创建 Purchase Invoice（采购发票）
     └─ 门店创建付款单
```

### 3.2 详细步骤说明

#### 步骤 1-3：申请和创建订单

**操作者**：门店人员 → 采购部门

1. **门店创建 MR 申请**
   - TTPOS 侧：创建采购申请单
   - ERPNext：创建 Material Request
   - 物品信息：**勾选默认供应商**（外部供应商）
   - 状态：Draft

2. **MR 审批**
   - 操作者：采购部门
   - ERPNext：Material Request 状态 → Submitted
   - 系统自动判断：物品勾选了默认供应商 → 走直采流程

3. **创建直采 PO**
   - 操作者：采购部门
   - ERPNext：创建 Purchase Order
   - 供应商：外部供应商（从物品的默认供应商获取）
   - 关联：Material Request
   - 状态：Draft
   - **自动触发**：系统在 TTPOS 的品采收货中生成该采购单
   - **在途仓操作**：**物品自动添加到门店在途仓库**
     - 时机：Purchase Order 创建时
     - 操作：系统自动将 Purchase Order 中的物品添加到门店在途仓库
     - 数量：使用 Purchase Order 中的采购数量
     - 目的：表示已向外部供应商下单，货物预期在途
     - 详细说明见 [在途仓管理](#在途仓管理)

#### 步骤 4-6：审批和下单

**操作者**：采购部门 → 审批人员

4. **审批 PO**
   - 操作者：采购部门/审批人员
   - ERPNext：Purchase Order 状态 → Submitted
   - 审批流程：根据金额自动进入对应审批状态
     - 金额 < 100,000：Pending PMA（采购经理审批）
     - 金额 ≥ 100,000：Pending VP（VP 审批）
   - 审批通过后状态：Approved

5. **打印采购 PDF**
   - 调用 ERPNext API 获取 PDF
   - 打印后提交给供应商

6. **提交外部供应商**
   - 外部供应商接收 PO
   - 准备货物
   - 安排配送

#### 步骤 7-9：收货

**操作者**：外部供应商 → 门店收货员

7. **外部供应商配送**
   - 外部供应商直接配送货物到门店
   - 可能包含多个供应商的物品

8. **门店创建收货单**
   - 操作者：门店收货员
   - TTPOS 侧：在品采收货中选择直采订单（PO）
   - 系统自动按供应商拆分收货单
     - 供应商 A 的物品 → 收货单 A
     - 供应商 B 的物品 → 收货单 B
   - 选择收货物品和数量

9. **门店确认收货**

   **操作者**：门店收货员  
   **操作位置**：TTPOS 系统（门店在 TTPOS 完成收货操作）

   **TTPOS 侧操作**：
   - 扫码收货或手动输入收货数量
   - 确认收货数量
   - 点击"确认收货"按钮

   **对应的 ERPNext 操作**（自动同步）：

   | TTPOS 操作 | ERPNext 自动执行的操作 | 同步时机 | 同步方式 |
   |-----------|---------------------|---------|---------|
   | 门店在 TTPOS 确认收货，提交收货单 | **自动为每个供应商创建 Purchase Receipt** | **提交收货单时** | **调用 ERPNext API** |
   | TTPOS 更新收货单状态为"已收货" | **Purchase Receipt 状态自动变为 Submitted** | **创建 Purchase Receipt 后** | **自动提交** |
   | TTPOS 更新本地库存 | **ERPNext 库存自动更新（门店仓库增加）** | **Purchase Receipt 提交后** | **ERPNext 自动执行** |
   | TTPOS 更新采购单进度 | **Purchase Order 的 `received_qty` 自动更新** | **Purchase Receipt 提交后** | **ERPNext 自动执行** |

   **同步流程**：
   ```
   门店在 TTPOS 点击"确认收货"
      ↓
   1. TTPOS 创建/更新收货单（本地操作）
      - 状态：待收货 → 已收货
      - 按供应商拆分收货单
      ↓
   2. TTPOS 自动调用 ERPNext API（同步操作）
      ├─ 为供应商 A 创建 Purchase Receipt
      │   ├─ API: POST /api/resource/Purchase Receipt
      │   ├─ 方法: make_purchase_receipt
      │   ├─ purchase_order: {Purchase Order No}
      │   ├─ supplier: Supplier A
      │   └─ items: 供应商 A 的物品明细
      │
      ├─ 自动提交 Purchase Receipt A
      │   ├─ API: POST /api/resource/Purchase Receipt/{name}
      │   ├─ action: "submit"
      │   └─ 状态: Draft → Submitted
      │
      └─ 为供应商 B 创建 Purchase Receipt（如果存在）
          ├─ 创建 Purchase Receipt B
          └─ 提交 Purchase Receipt B
      ↓
   3. ERPNext 自动执行后续操作
      ├─ 更新 Purchase Order 的 received_qty（汇总所有 Purchase Receipt）
      ├─ 更新 Purchase Order 的 per_received（自动计算）
      ├─ 更新 Purchase Order 状态（部分收货/全部收货）
      └─ 更新库存（门店仓库增加）
      ↓
   4. TTPOS 更新本地数据
      ├─ 更新收货单的 ErpOrderNo（关联 Purchase Receipt）
      ├─ 更新本地库存（门店仓库增加）
      └─ 更新采购单进度
      ↓
   5. 同步完成
      ✅ TTPOS 收货单状态：已收货
      ✅ ERPNext Purchase Receipt 状态：Submitted
      ✅ 库存已更新（TTPOS 和 ERPNext）
      ✅ Purchase Order 收货进度已更新
   ```

   **同步性说明**：
   - ✅ **实时同步**：门店在 TTPOS 确认收货时，立即同步到 ERPNext
   - ✅ **自动执行**：ERPNext 侧的操作由系统自动完成，无需门店人员手动操作
   - ✅ **原子性保证**：如果 ERPNext API 调用失败，TTPOS 收货操作会回滚或标记为待同步
   - ✅ **一致性保证**：TTPOS 和 ERPNext 的数据保持一致，通过 `ErpOrderNo` 字段关联

   **详细操作见 [5. TTPOS 侧收货操作](#5-ttpos-侧收货操作) 和 [5.2.5 TTPOS 操作与 ERP 操作对应关系（直采收货）](#525-ttpos-操作与-erp-操作对应关系直采收货)**

**ERPNext 界面按钮操作（直采收货）**：

**说明**：以下是在 ERPNext 中手动操作的步骤（通常由 TTPOS 自动完成，门店人员无需手动操作）：

如果门店人员在 ERPNext 中手动创建 Purchase Receipt（通常由 TTPOS 自动完成），操作步骤如下：

1. **进入 Purchase Order 页面**
   - 路径：`Buying > Purchase Order > {Purchase Order No}`
   - 例如：`Buying > Purchase Order > PO-00001`

2. **点击 "Create" 按钮**
   - 位置：页面右上角的 "Create" 下拉菜单

3. **选择 "Purchase Receipt"**
   - 在 "Create" 下拉菜单中，选择 "Purchase Receipt"
   - 系统会自动创建 Purchase Receipt（采购收货单）

4. **查看创建的 Purchase Receipt**
   - 系统会打开新创建的 Purchase Receipt 页面
   - 自动填充了 Purchase Order 中的物品信息
   - 供应商自动从 Purchase Order 继承

5. **按供应商拆分收货单**（如果包含多个供应商）
   - 如果 Purchase Order 包含多个供应商的物品，需要为每个供应商创建独立的 Purchase Receipt
   - 只选择同一供应商的物品创建 Purchase Receipt
   - 重复步骤 2-4，为其他供应商创建 Purchase Receipt

6. **填写收货信息**
   - 核对收货数量（可以与 Purchase Order 数量不同）
   - 检查物品编码和规格
   - 输入实际收货数量

7. **提交 Purchase Receipt**
   - 点击 "Submit" 按钮提交
   - 状态变为 "Submitted"
   - 库存自动更新（门店仓库增加）
   - Purchase Order 的 `received_qty` 自动更新

**关键按钮**：
- **"Create"** → **"Purchase Receipt"**：从 Purchase Order 创建采购收货单
- **"Submit"**：提交 Purchase Receipt，更新库存和采购单进度

#### 步骤 10：财务结算

**操作者**：财务部门

10. **财务处理**
    - 门店创建 Purchase Invoice（采购发票）
    - 门店创建付款单
    - 记录应付账款

---

## 4. 集采拆单规则

### 4.1 拆单目的

集采部分发货时，需要**按仓库拆分 Delivery Note**，原因：

1. **清楚区分货物来源**：门店人员可以清楚地知道哪些物品来自哪个总部仓库
2. **便于收货管理**：不同仓库的物品可能由不同的司机或批次配送，按仓库拆单便于收货
3. **支持分批收货**：门店可以按 Delivery Note 分别确认收货
4. **便于对账和追溯**：每个 Delivery Note 对应一个仓库的发货，便于对账和问题追溯

### 4.2 拆单规则

#### 规则 1：按总部仓库拆分

一个 Sales Order（集采订单）可能包含来自不同总部仓库的物品，系统需要**按仓库自动拆分**：

```
Sales Order (SO-00001)
├─ 物品 A (仓库: 总部仓库A, 数量: 50)
├─ 物品 B (仓库: 总部仓库A, 数量: 30)
├─ 物品 C (仓库: 总部仓库B, 数量: 100)
└─ 物品 D (仓库: 总部仓库B, 数量: 20)
    ↓
按仓库分组：
├─ 总部仓库A:
│   ├─ 物品 A (数量: 50)
│   └─ 物品 B (数量: 30)
│   → Delivery Note A (MAT-DN-2026-00001)
│
└─ 总部仓库B:
    ├─ 物品 C (数量: 100)
    └─ 物品 D (数量: 20)
    → Delivery Note B (MAT-DN-2026-00002)
```

#### 规则 2：拆单时机

- **触发时机**：总部仓库人员创建 Delivery Note 时
- **拆分方式**：系统自动按仓库分组，为每个仓库创建独立的 Delivery Note
- **操作方式**：
  - 方案 A：系统自动拆分（推荐）
  - 方案 B：仓库人员手动选择仓库分组

#### 规则 3：Delivery Note 关联

每个 Delivery Note 需要关联：
- Sales Order（集采订单）
- 仓库信息（源仓库）
- 门店信息（目标仓库）

### 4.3 拆单实现逻辑

#### 伪代码

```python
def create_delivery_notes_by_warehouse(sales_order_name):
    """
    从 Sales Order 按仓库拆分创建多个 Delivery Note
    """
    # 1. 获取 Sales Order 信息
    sales_order = get_sales_order(sales_order_name)
    
    # 2. 按仓库分组物品
    warehouse_groups = {}
    for item in sales_order.items:
        source_warehouse = item.warehouse  # 源仓库（总部仓库）
        if source_warehouse not in warehouse_groups:
            warehouse_groups[source_warehouse] = []
        warehouse_groups[source_warehouse].append(item)
    
    # 3. 为每个仓库创建独立的 Delivery Note
    delivery_notes = []
    for warehouse, items in warehouse_groups.items():
        delivery_note = create_delivery_note({
            "customer": sales_order.customer,  # 门店
            "company": sales_order.company,     # 总部
            "against_sales_order": sales_order_name,
            "set_warehouse": warehouse,  # 源仓库
            "set_target_warehouse": sales_order.set_target_warehouse,  # 目标仓库（门店）
            "items": items  # 只包含该仓库的物品
        })
        delivery_notes.append(delivery_note)
    
    return delivery_notes
```

### 4.4 ERPNext API 调用示例

```python
# 场景：一个 Sales Order 包含来自不同总部仓库的物品
# 需要为每个仓库创建独立的 Delivery Note

# 步骤1：获取 Sales Order 信息，按仓库分组物品
sales_order = get_sales_order("SO-00001")

# 按仓库分组物品
warehouse_items = {}
for item in sales_order.items:
    warehouse = item.warehouse  # 源仓库（总部仓库）
    if warehouse not in warehouse_items:
        warehouse_items[warehouse] = []
    warehouse_items[warehouse].append(item)

# 步骤2：为每个仓库创建独立的 Delivery Note

# 仓库 A 的 Delivery Note
POST /api/resource/Delivery Note
{
    "customer": "Store Branch - Company",  // 门店作为客户
    "company": "Headquarters - Company",   // 总部作为销售方
    "delivery_date": "2025-01-20",
    "against_sales_order": "SO-00001",
    "set_warehouse": "Headquarters Warehouse A - Company",  // 源仓库（总部仓库A）
    "set_target_warehouse": "Store Warehouse - Company",   // 目标仓库（门店）
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 50,
            "uom": "Nos",
            "warehouse": "Headquarters Warehouse A - Company",
            "target_warehouse": "Store Warehouse - Company",
            "sales_order": "SO-00001",
            "sales_order_item": "SO-ITEM-001"
        },
        {
            "item_code": "ITEM-002",
            "qty": 30,
            "uom": "Nos",
            "warehouse": "Headquarters Warehouse A - Company",
            "target_warehouse": "Store Warehouse - Company",
            "sales_order": "SO-00001",
            "sales_order_item": "SO-ITEM-002"
        }
    ]
}

# 仓库 B 的 Delivery Note
POST /api/resource/Delivery Note
{
    "customer": "Store Branch - Company",
    "company": "Headquarters - Company",
    "delivery_date": "2025-01-20",
    "against_sales_order": "SO-00001",
    "set_warehouse": "Headquarters Warehouse B - Company",  // 源仓库（总部仓库B）
    "set_target_warehouse": "Store Warehouse - Company",
    "items": [
        {
            "item_code": "ITEM-003",
            "qty": 100,
            "uom": "Nos",
            "warehouse": "Headquarters Warehouse B - Company",
            "target_warehouse": "Store Warehouse - Company",
            "sales_order": "SO-00001",
            "sales_order_item": "SO-ITEM-003"
        }
    ]
}

# 步骤3：提交每个 Delivery Note
POST /api/resource/Delivery Note/{name}
{
    "action": "submit"
}
# 提交后，库存从对应的总部仓库扣减
```

---

## 5. TTPOS 侧收货操作

### 5.1 集采收货操作

#### 5.1.1 收货前置条件

- Delivery Note 已创建并提交（ERPNext 状态：Submitted）
- 系统已在 TTPOS 中创建对应的采购收货单（PR）
- 采购收货单关联 Delivery Note（通过 `ErpOrderNo` 字段）

#### 5.1.2 收货操作流程

```
门店收货员操作流程：

1. 登录 TTPOS，进入"品采收货"页面
   ↓
2. 查看待收货列表
   ├─ 显示所有待收货的采购收货单
   ├─ 显示发货单号（Delivery Note 编号）
   ├─ 显示发货仓库信息
   └─ 区分集采和直采
   ↓
3. 选择集采收货单
   ├─ 点击查看详情
   ├─ 显示关联的 Delivery Note 信息
   ├─ 显示发货仓库
   └─ 显示物品明细
   ↓
4. 开始收货
   ├─ 方式一：扫码收货
   │   ├─ 扫描物品条形码
   │   ├─ 系统自动匹配物品
   │   ├─ 自动填充收货数量
   │   └─ 可以调整数量
   │
   └─ 方式二：手动收货
       ├─ 选择物品
       ├─ 输入收货数量
       └─ 确认收货
   ↓
5. 确认收货
   ├─ 核对收货数量与 Delivery Note 数量
   ├─ 检查物品编码是否正确
   ├─ 提交收货单
   └─ 更新库存（从在途仓库转入目标仓库）
   ↓
6. 收货完成
   ├─ 采购收货单状态：已收货
   ├─ 库存更新完成
   └─ 可以查看收货记录
```

#### 5.1.3 使用的 ERP 功能

- **查看 Delivery Note 信息**：通过 `ErpOrderNo` 字段关联，显示发货单详情
- **更新 Delivery Note 收货进度**：收货后自动更新 Delivery Note 的 `delivered_qty` 和 `per_delivered`
- **创建 Inter Company Purchase Receipt**（可选）：如果需要双向记录，可以从 Delivery Note 创建

#### 5.1.4 TTPOS 操作与 ERP 操作对应关系（集采收货）

**核心说明**：门店人员在 TTPOS 完成收货操作，系统自动同步到 ERPNext，实现操作的同步性。

**操作对应关系表**：

| TTPOS 操作步骤 | TTPOS 操作内容 | 对应的 ERPNext 操作 | 同步时机 | 同步方式 | 同步结果 |
|--------------|--------------|-------------------|---------|---------|---------|
| 1. 查看待收货列表 | 门店在 TTPOS 查看采购收货单列表 | 查询关联的 Delivery Note 信息 | 实时查询 | 通过 `ErpOrderNo` 字段关联查询 | 显示发货单号和发货仓库 |
| 2. 选择收货单 | 门店选择集采收货单 | 查询 Delivery Note 详情（GET /api/resource/Delivery Note/{name}） | 点击查看详情时 | 按需查询，不持久化 | 显示物品明细和发货信息 |
| 3. 扫码/手动收货 | 门店扫码或手动输入收货数量 | - | - | 本地操作，暂不同步 | TTPOS 本地数据暂存 |
| 4. 确认收货 | 门店点击"确认收货"按钮 | **自动创建 Inter Company Purchase Receipt**（可选） | **提交收货单时** | **调用 ERPNext API** | **ERPNext 创建 Purchase Receipt** |
| 5. 提交收货单 | TTPOS 更新收货单状态为"已收货" | **Purchase Receipt 状态变为 Submitted**（如果创建） | **同步创建 Purchase Receipt 后** | **自动提交** | **ERPNext 库存更新** |
| 6. 更新库存 | TTPOS 更新本地库存 | **Delivery Note 的 `delivered_qty` 自动更新** | **Purchase Receipt 提交后** | **ERPNext 自动更新** | **Delivery Note 收货进度更新** |
| 7. 更新进度 | TTPOS 更新采购单进度 | **Sales Order 的 `per_delivered` 自动更新** | **Delivery Note 更新后** | **ERPNext 自动更新** | **Sales Order 交付进度更新** |

**详细同步流程**：

```
┌─────────────────────────────────────────────────────────────────┐
│                    TTPOS 收货操作同步流程（集采）                  │
└─────────────────────────────────────────────────────────────────┘

阶段1：TTPOS 本地操作（不同步）
├─ 1. 门店在 TTPOS 查看待收货列表
│    └─ 查询关联的 Delivery Note（按需查询，不持久化）
│
├─ 2. 门店选择集采收货单
│    └─ 查询 Delivery Note 详情（GET /api/resource/Delivery Note/{name}）
│
└─ 3. 门店扫码或手动输入收货数量
     └─ TTPOS 本地操作，暂不同步

阶段2：提交收货时同步（关键同步点）
├─ 4. 门店点击"确认收货"按钮
│    └─ TTPOS: 更新收货单状态为"待收货" → "已收货"
│
├─ 5. TTPOS 自动调用 ERPNext API
│    ├─ API: POST /api/resource/Purchase Receipt
│    ├─ 方法: make_inter_company_purchase_receipt
│    ├─ 参数:
│    │   - against_delivery_note: {Delivery Note No}
│    │   - supplier: 总部供应商（自动填充）
│    │   - items: 收货物品明细（从 TTPOS 收货单获取）
│    └─ 创建: Inter Company Purchase Receipt（状态: Draft）
│
├─ 6. 自动提交 Purchase Receipt
│    ├─ API: POST /api/resource/Purchase Receipt/{name}
│    ├─ action: "submit"
│    └─ 状态: Draft → Submitted
│
└─ 7. ERPNext 自动执行操作
     ├─ 更新 Delivery Note 的 delivered_qty（自动）
     ├─ 更新 Delivery Note 的 per_delivered（自动）
     ├─ 更新 Sales Order 的 per_delivered（自动）
     └─ 更新库存（从总部仓库扣减，门店仓库增加）

阶段3：同步完成
└─ 8. 同步结果反馈
     ├─ TTPOS: 更新收货单的 ErpOrderNo（关联 Purchase Receipt）
     ├─ TTPOS: 更新本地库存（从在途仓库转入目标仓库）
     └─ ERPNext: Delivery Note 和 Sales Order 收货进度已更新
```

**同步机制说明**：

1. **查询操作（不同步）**：
   - 查看待收货列表：实时查询，不持久化
   - 查看发货单详情：按需查询，不持久化
   - **特点**：只读操作，不影响 ERPNext 数据

2. **提交操作（同步）**：
   - 确认收货：自动创建 ERPNext Purchase Receipt
   - **同步时机**：门店点击"确认收货"按钮时
   - **同步方式**：调用 ERPNext API 创建 Purchase Receipt
   - **同步结果**：ERPNext 创建 Purchase Receipt 并自动提交

3. **自动更新（ERPNext 侧）**：
   - Delivery Note 的 `delivered_qty`：Purchase Receipt 提交后自动更新
   - Delivery Note 的 `per_delivered`：Purchase Receipt 提交后自动更新
   - Sales Order 的 `per_delivered`：Delivery Note 更新后自动更新
   - **特点**：ERPNext 内部自动更新，无需额外 API 调用

**同步一致性保证**：

- ✅ **原子性**：TTPOS 收货操作和 ERPNext API 调用在同一事务中完成
- ✅ **一致性**：如果 ERPNext API 调用失败，TTPOS 收货操作会回滚或标记为待同步
- ✅ **可追溯性**：通过 `ErpOrderNo` 字段关联 TTPOS 收货单和 ERPNext Delivery Note/Purchase Receipt
- ✅ **幂等性**：相同收货单不会重复创建 Purchase Receipt（通过检查 `ErpOrderNo` 判断）

**错误处理**：

- ❌ **ERPNext API 调用失败**：
  - TTPOS 收货操作回滚（状态恢复为"待收货"）
  - 或标记为待同步状态（`sync_status = pending`）
  - 提示门店操作失败，稍后重试
  
- ❌ **网络超时**：
  - 标记为待同步状态
  - 后台定时任务重试同步
  
- ✅ **同步成功**：
  - 更新 `ErpOrderNo` 字段（关联 Purchase Receipt）
  - 更新同步状态为已同步（`sync_status = synced`）

#### 5.1.4 TTPOS 操作与 ERP 操作对应关系（集采收货）

**说明**：门店人员在 TTPOS 完成收货操作，系统自动同步到 ERPNext，实现操作的同步性。

**操作对应关系**：

| TTPOS 操作 | 对应的 ERPNext 操作 | 同步时机 | 同步方式 |
|-----------|-------------------|---------|---------|
| 门店在 TTPOS 选择采购收货单 | 查询 Delivery Note 信息 | 实时 | 通过 `ErpOrderNo` 字段关联查询 |
| 门店扫码收货或手动输入收货数量 | - | - | TTPOS 本地操作，暂不同步 |
| 门店确认收货，提交收货单 | **自动创建/更新 Purchase Receipt** | **提交收货单时** | **自动调用 ERPNext API** |
| TTPOS 更新库存 | Delivery Note 的 `delivered_qty` 自动更新 | 同步创建 Purchase Receipt 后 | ERPNext 自动更新 |
| TTPOS 更新采购单进度 | Purchase Order 的 `arrival_num` 更新 | 确认收货后 | TTPOS 侧更新 |

**同步流程详解**：

```
门店在 TTPOS 确认收货
   ↓
1. TTPOS 创建/更新采购收货单（PurchaseReceiptOrder）
   - 状态：已收货（ReceiptOrderStatusReceived）
   - 关联 Delivery Note（通过 ErpOrderNo）
   ↓
2. 【可选】TTPOS 自动调用 ERPNext API 创建 Inter Company Purchase Receipt
   - API: POST /api/resource/Purchase Receipt
   - 方法: make_inter_company_purchase_receipt
   - 关联: against_delivery_note = {Delivery Note No}
   - 供应商: 总部供应商（自动填充）
   - 状态: Draft → Submitted
   ↓
3. ERPNext 自动更新 Delivery Note 的 delivered_qty
   - Delivery Note 的 delivered_qty 字段自动更新
   - per_delivered 百分比自动计算
   - Sales Order 的 per_delivered 自动更新
   ↓
4. TTPOS 更新库存（本地库存）
   - 从在途仓库转入目标仓库
   - 库存数量增加
   ↓
5. 同步完成
   - TTPOS 收货单状态：已收货
   - ERPNext Purchase Receipt 状态：Submitted（如果创建）
   - Delivery Note delivered_qty 已更新
   - 库存已更新
```

**同步时机**：
- ✅ **实时同步**：查询 Delivery Note 信息（按需查询，不持久化）
- ✅ **提交时同步**：确认收货后，自动创建/更新 ERPNext Purchase Receipt（可选）
- ✅ **自动更新**：ERPNext 侧 Delivery Note 的 `delivered_qty` 自动更新（如果创建了 Purchase Receipt）

**同步一致性保证**：
- ✅ **原子性**：TTPOS 收货操作和 ERPNext API 调用在同一事务中完成
- ✅ **一致性**：如果 ERPNext API 调用失败，TTPOS 收货操作会回滚或标记为待同步
- ✅ **可追溯性**：通过 `ErpOrderNo` 字段关联 TTPOS 收货单和 ERPNext Delivery Note/Purchase Receipt

#### 5.1.5 ERPNext 界面按钮操作（集采收货）

**说明**：集采收货主要在 TTPOS 侧操作，ERPNext 侧的操作由系统自动完成。如果需要在 ERPNext 中手动操作，可参考以下步骤：

**在 ERPNext 中创建 Inter Company Purchase Receipt（可选）**：

1. **进入 Delivery Note 页面**
   - 路径：`Stock > Delivery Note > {Delivery Note No}`
   - 例如：`Stock > Delivery Note > MAT-DN-2026-00001`

2. **点击 "Create" 按钮**
   - 位置：页面右上角的 "Create" 下拉菜单

3. **选择 "Inter Company Purchase Receipt"**
   - 在 "Create" 下拉菜单中，选择 "Inter Company Purchase Receipt"
   - 系统会自动创建 Inter Company Purchase Receipt（跨公司采购收货单）

4. **查看创建的 Purchase Receipt**
   - 系统会打开新创建的 Purchase Receipt 页面
   - 自动填充了 Delivery Note 中的物品信息
   - 供应商自动设置为总部供应商

5. **提交 Purchase Receipt**
   - 核对收货数量和物品信息
   - 点击 "Submit" 按钮提交
   - 状态变为 "Submitted"，库存自动更新

**关键按钮**：
- **"Create"** → **"Inter Company Purchase Receipt"**：从 Delivery Note 创建跨公司采购收货单
- **"Submit"**：提交 Purchase Receipt，更新库存

### 5.2 直采收货操作

#### 5.2.1 收货前置条件

- Purchase Order 已创建并审批通过（ERPNext 状态：Submitted）
- 系统已在 TTPOS 的品采收货中生成该采购单
- 外部供应商已配送货物到门店

#### 5.2.2 收货操作流程

```
门店收货员操作流程：

1. 登录 TTPOS，进入"品采收货"页面
   ↓
2. 查看待收货列表
   ├─ 显示所有待收货的采购单（包含集采和直采）
   ├─ 显示供应商信息
   └─ 区分集采和直采
   ↓
3. 选择直采采购单
   ├─ 点击查看详情
   ├─ 显示关联的 Purchase Order 信息
   ├─ 显示供应商信息
   └─ 显示物品明细
   ↓
4. 创建收货单
   ├─ 系统自动按供应商拆分
   │   ├─ 供应商 A 的物品 → 收货单 A
   │   └─ 供应商 B 的物品 → 收货单 B
   ├─ 选择收货物品
   └─ 输入收货数量（可以调整）
   ↓
5. 开始收货
   ├─ 方式一：扫码收货
   │   ├─ 扫描物品条形码
   │   ├─ 系统自动匹配物品
   │   ├─ 自动填充收货数量
   │   └─ 可以调整数量
   │
   └─ 方式二：手动收货
       ├─ 选择物品
       ├─ 输入收货数量
       └─ 确认收货
   ↓
6. 确认收货
   ├─ 核对收货数量与 Purchase Order 数量
   ├─ 检查物品编码是否正确
   ├─ 提交收货单
   ├─ TTPOS 自动为每个供应商创建 Purchase Receipt
   └─ 更新库存（门店仓库增加）
   ↓
7. 收货完成
   ├─ Purchase Receipt 状态：Submitted
   ├─ 采购单到货数量更新
   ├─ 库存更新完成
   └─ 可以查看收货记录
```

#### 5.2.3 按供应商拆单规则

直采收货时，一个 Purchase Order 可能包含多个供应商的物品，系统需要**按供应商拆分**：

```
Purchase Order (PO-00001)
├─ 物品 A (供应商: Supplier A, 数量: 50)
├─ 物品 B (供应商: Supplier A, 数量: 30)
├─ 物品 C (供应商: Supplier B, 数量: 100)
└─ 物品 D (供应商: Supplier B, 数量: 20)
    ↓
按供应商分组：
├─ Supplier A:
│   ├─ 物品 A (数量: 50)
│   └─ 物品 B (数量: 30)
│   → Purchase Receipt A (MAT-PRE-2026-00001)
│
└─ Supplier B:
    ├─ 物品 C (数量: 100)
    └─ 物品 D (数量: 20)
    → Purchase Receipt B (MAT-PRE-2026-00002)
```

#### 5.2.4 使用的 ERP 功能

- **查看 Purchase Order 信息**：显示采购单详情和物品明细
- **创建 Purchase Receipt**：为每个供应商创建独立的采购收货单
- **更新 Purchase Order 收货进度**：收货后自动更新采购单的 `received_qty` 和 `per_received`

#### 5.2.5 TTPOS 操作与 ERP 操作对应关系（直采收货）

**说明**：门店人员在 TTPOS 完成收货操作，系统自动同步到 ERPNext，实现操作的同步性。

**操作对应关系**：

| TTPOS 操作 | 对应的 ERPNext 操作 | 同步时机 | 同步方式 |
|-----------|-------------------|---------|---------|
| 门店在 TTPOS 选择采购单 | 查询 Purchase Order 信息 | 实时 | 通过 `ErpOrderNo` 字段关联查询 |
| 门店创建收货单，选择收货物品 | - | - | TTPOS 本地操作，暂不同步 |
| 门店扫码收货或手动输入收货数量 | - | - | TTPOS 本地操作，暂不同步 |
| 门店确认收货，提交收货单 | **自动创建 Purchase Receipt** | **提交收货单时** | **自动调用 ERPNext API** |
| TTPOS 更新库存 | Purchase Receipt 提交后库存自动更新 | 同步创建 Purchase Receipt 后 | ERPNext 自动更新 |
| TTPOS 更新采购单进度 | Purchase Order 的 `received_qty` 自动更新 | Purchase Receipt 提交后 | ERPNext 自动更新 |

**同步流程详解**：

```
门店在 TTPOS 确认收货
   ↓
1. TTPOS 创建收货单（PurchaseReceiptOrder）
   - 状态：待收货（ReceiptOrderStatusPending）
   - 关联 Purchase Order（通过 PurchaseOrderUuid）
   - 系统自动按供应商拆分收货单
   ↓
2. 门店填写收货信息
   - 选择收货物品
   - 输入收货数量（可以调整）
   - 扫码收货或手动输入
   ↓
3. 门店确认收货，提交收货单
   - 状态：已收货（ReceiptOrderStatusReceived）
   ↓
4. TTPOS 自动为每个供应商创建 ERPNext Purchase Receipt
   - API: POST /api/resource/Purchase Receipt
   - 方法: make_purchase_receipt
   - 关联: purchase_order = {Purchase Order No}
   - 供应商: 从 Purchase Order 继承（自动填充）
   - 状态: Draft → Submitted（自动提交）
   - 库存: 自动更新（门店仓库增加）
   ↓
5. ERPNext 自动更新 Purchase Order 的 received_qty
   - Purchase Order 的 received_qty 字段自动更新
   - per_received 百分比自动计算
   - Purchase Order 状态可能更新（部分收货/全部收货）
   ↓
6. TTPOS 更新库存（本地库存）
   - 门店仓库库存增加
   - 在途仓库库存减少（如果在途仓）
   ↓
7. 同步完成
   - TTPOS 收货单状态：已收货
   - ERPNext Purchase Receipt 状态：Submitted
   - Purchase Order received_qty 已更新
   - 库存已更新（TTPOS 和 ERPNext）
```

**按供应商拆分同步逻辑**：

```
一个 Purchase Order 包含多个供应商的物品
   ↓
TTPOS 自动按供应商拆分收货单
   ├─ 供应商 A 的物品 → 收货单 A
   └─ 供应商 B 的物品 → 收货单 B
   ↓
为每个供应商创建独立的 ERPNext Purchase Receipt
   ├─ Purchase Receipt A (供应商 A)
   │   └─ 状态: Submitted
   │   └─ 库存: 门店仓库增加（供应商 A 的物品）
   └─ Purchase Receipt B (供应商 B)
       └─ 状态: Submitted
       └─ 库存: 门店仓库增加（供应商 B 的物品）
   ↓
Purchase Order 的 received_qty 汇总更新
   └─ received_qty = 所有 Purchase Receipt 的收货数量总和
```

**同步时机**：
- ✅ **实时同步**：查询 Purchase Order 信息（按需查询，不持久化）
- ✅ **提交时同步**：确认收货后，自动为每个供应商创建 ERPNext Purchase Receipt
- ✅ **自动更新**：ERPNext 侧 Purchase Order 的 `received_qty` 自动更新（Purchase Receipt 提交后）

**同步一致性保证**：
- ✅ **原子性**：TTPOS 收货操作和 ERPNext API 调用在同一事务中完成
- ✅ **一致性**：如果 ERPNext API 调用失败，TTPOS 收货操作会回滚或标记为待同步
- ✅ **可追溯性**：通过 `ErpOrderNo` 字段关联 TTPOS 收货单和 ERPNext Purchase Receipt
- ✅ **按供应商拆分**：每个供应商创建独立的 Purchase Receipt，确保供应商级别的库存和进度更新

**代码实现位置**：

```go
// 直采收货同步实现（main/app/service/purchase_order/receipt_order.go）
func (s *purchaseReceiptOrderSrv) CreatePurchaseReceiptOrder(...) {
    // 1. TTPOS 创建收货单（本地操作）
    receiptOrder := &model.PurchaseReceiptOrder{...}
    err = receiptOrderRepo.Create(receiptOrder)
    
    // 2. 如果确认收货，同步到 ERPNext
    if req.IsConfirm && ctx.GetCompany().IsOpenErp() {
        // 按供应商分组物品
        supplierGroups := groupItemsBySupplier(receiptOrder.Items)
        
        // 为每个供应商创建 Purchase Receipt
        for supplier, items := range supplierGroups {
            erpReq := buying.SavePurchaseReceiptReq{
                PurchaseOrderName: receiptOrder.PurchaseOrder.ErpOrderNo,
                Items:             convertToErpItems(items),
            }
            
            // 调用 ERPNext API 创建 Purchase Receipt
            resp, err := erp.NewIErpSrv(s.dbm).SavePurchaseReceipt(ctx, &erpReq)
            if err != nil {
                // 错误处理：回滚或标记为待同步
                return err
            }
            
            // 更新收货单的 ERP 订单号
            receiptOrder.ErpOrderNo = resp.PurchaseReceipt.PurchaseReceiptName
        }
        
        // 更新收货单的 ERP 订单号
        receiptOrderRepo.Update(receiptOrder)
    }
}
```

#### 5.2.6 ERPNext 界面按钮操作（直采收货）

**说明**：直采收货主要在 TTPOS 侧操作，ERPNext 侧的操作由系统自动完成。如果需要在 ERPNext 中手动操作，可参考以下步骤：

**在 ERPNext 中创建 Purchase Receipt**：

1. **进入 Purchase Order 页面**
   - 路径：`Buying > Purchase Order > {Purchase Order No}`
   - 例如：`Buying > Purchase Order > PO-00001`

2. **点击 "Create" 按钮**
   - 位置：页面右上角的 "Create" 下拉菜单

3. **选择 "Purchase Receipt"**
   - 在 "Create" 下拉菜单中，选择 "Purchase Receipt"
   - 系统会自动创建 Purchase Receipt（采购收货单）

4. **查看创建的 Purchase Receipt**
   - 系统会打开新创建的 Purchase Receipt 页面
   - 自动填充了 Purchase Order 中的物品信息
   - 供应商自动从 Purchase Order 继承

5. **按供应商拆分收货单**（如果包含多个供应商）
   - 如果 Purchase Order 包含多个供应商的物品，需要为每个供应商创建独立的 Purchase Receipt
   - 只选择同一供应商的物品创建 Purchase Receipt
   - 重复步骤 2-4，为其他供应商创建 Purchase Receipt

6. **填写收货信息**
   - 核对收货数量（可以与 Purchase Order 数量不同）
   - 检查物品编码和规格
   - 输入实际收货数量

7. **提交 Purchase Receipt**
   - 点击 "Submit" 按钮提交
   - 状态变为 "Submitted"
   - 库存自动更新（门店仓库增加）
   - Purchase Order 的 `received_qty` 自动更新

**关键按钮**：
- **"Create"** → **"Purchase Receipt"**：从 Purchase Order 创建采购收货单
- **"Submit"**：提交 Purchase Receipt，更新库存和采购单进度

**Purchase Receipt 页面中的其他相关按钮**：
- **"Create"** → **"Purchase Return"**：创建采购退货单
- **"Create"** → **"Purchase Invoice"**：创建采购发票
- **"Create"** → **"Make Stock Entry"**：创建库存调整单
- **"Create"** → **"Retention Stock Entry"**：创建保留库存调整单（质量检验）
- **"Create"** → **"Landed Cost Voucher"**：创建到岸成本凭证

---

## 5.3 ERPNext 界面操作总结

### 5.3.1 集采收货操作总结

**操作场景**：门店在 ERPNext 中手动操作集采收货（通常由 TTPOS 自动完成）

**操作步骤**：
1. 进入 Delivery Note 页面：`Stock > Delivery Note > {Delivery Note No}`
2. 点击 **"Create"** 按钮 → 选择 **"Inter Company Purchase Receipt"**
3. 系统自动创建 Purchase Receipt，自动填充 Delivery Note 中的物品信息
4. 核对收货数量，点击 **"Submit"** 提交
5. 库存自动更新，Delivery Note 的 `delivered_qty` 自动更新

**关键按钮路径**：
```
Delivery Note 页面
  └─ Create 按钮
      └─ Inter Company Purchase Receipt
          └─ 自动打开 Purchase Receipt 页面
              └─ Submit 按钮（提交收货单）
```

### 5.3.2 直采收货操作总结

**操作场景**：门店在 ERPNext 中手动操作直采收货（通常由 TTPOS 自动完成）

**操作步骤**：
1. 进入 Purchase Order 页面：`Buying > Purchase Order > {Purchase Order No}`
2. 点击 **"Create"** 按钮 → 选择 **"Purchase Receipt"**
3. 系统自动创建 Purchase Receipt，自动填充 Purchase Order 中的物品信息
4. **按供应商拆分**（如果包含多个供应商）：为每个供应商创建独立的 Purchase Receipt
5. 核对收货数量，点击 **"Submit"** 提交
6. 库存自动更新，Purchase Order 的 `received_qty` 自动更新

**关键按钮路径**：
```
Purchase Order 页面
  └─ Create 按钮
      └─ Purchase Receipt
          └─ 自动打开 Purchase Receipt 页面
              └─ Submit 按钮（提交收货单）
              
Purchase Receipt 页面（可选操作）
  └─ Create 按钮
      ├─ Purchase Return（采购退货）
      ├─ Purchase Invoice（采购发票）
      ├─ Make Stock Entry（库存调整）
      ├─ Retention Stock Entry（保留库存）
      └─ Landed Cost Voucher（到岸成本）
```

### 5.3.3 操作对比

| 项目 | 集采收货 | 直采收货 |
|------|---------|---------|
| **入口单据** | Delivery Note | Purchase Order |
| **创建按钮** | Create → Inter Company Purchase Receipt | Create → Purchase Receipt |
| **目标单据** | Purchase Receipt（跨公司采购收货单） | Purchase Receipt（采购收货单） |
| **拆分规则** | 按仓库拆分（在创建 Delivery Note 时） | 按供应商拆分（在创建 Purchase Receipt 时） |
| **提交按钮** | Submit（Purchase Receipt 页面） | Submit（Purchase Receipt 页面） |
| **库存更新** | 门店仓库增加 | 门店仓库增加 |
| **进度更新** | Delivery Note 的 `delivered_qty` 更新 | Purchase Order 的 `received_qty` 更新 |

---

## 6. 在途仓管理

### 6.1 在途仓概述

**在途仓（Transit Warehouse）**：用于记录已下单但尚未到达门店的货物，表示货物正在运输途中。

**作用**：
- 记录预期在途的货物数量
- 便于库存管理和对账
- 支持收货时数量差异处理

### 6.2 集采部分：在途仓操作

#### 6.2.1 进入在途仓的时机

**时机**：**Delivery Note 提交后**

**操作流程**：
```
1. 总部仓库人员创建 Delivery Note
   - 状态：Draft
   ↓
2. 提交 Delivery Note
   - 状态：Draft → Submitted
   - 库存从总部仓库扣减
   ↓
3. 【自动触发】物品添加到门店在途仓库
   - 时机：Delivery Note 提交后
   - 操作：系统自动将 Delivery Note 中的物品添加到门店在途仓库
   - 数量：使用 Delivery Note 中的发货数量
   - 仓库：门店在途仓库（Transit Warehouse）
   ↓
4. 司机配送货物到门店
   - 货物在运输途中
   - 在途仓库数量保持不变
   ↓
5. 门店确认收货
   - 从在途仓库扣减数量
   - 转入目标仓库（门店正常仓库）
   - 使用实际收货数量（可能与 Delivery Note 数量不同）
```

#### 6.2.2 在途仓操作详情

**添加时机**：
- ✅ **Delivery Note 提交后**：表示货物已从总部仓库发出，开始运输
- ❌ **不在 Delivery Note 创建时**：此时货物还未发出

**添加数量**：
- 使用 Delivery Note 中的发货数量
- 每个物品的数量 = Delivery Note Item 的 `qty` 字段

**添加操作**：
```go
// 在 Delivery Note 提交后，自动添加到在途仓库
func HandleDeliveryNoteSubmitted(ctx context.Context, deliveryNote *DeliveryNote) error {
    // 1. 获取门店在途仓库
    transitWarehouse, err := getTransitWarehouse(ctx, storeCompany.Uuid)
    
    // 2. 为每个物品添加到在途仓库
    for _, item := range deliveryNote.Items {
        err := addToTransitWarehouse(ctx, transitWarehouse, deliveryNote, item)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

**收货时处理**：
- 门店确认收货时，从在途仓库扣减数量
- 转入目标仓库（门店正常仓库）
- 如果实际收货数量与 Delivery Note 数量不同，使用实际收货数量

### 6.3 直采部分：在途仓操作

#### 6.3.1 进入在途仓的时机

**时机**：**Purchase Order 创建时**

**操作流程**：
```
1. 采购部门创建 Purchase Order
   - 状态：Draft
   ↓
2. 【自动触发】物品添加到门店在途仓库
   - 时机：Purchase Order 创建时
   - 操作：系统自动将 Purchase Order 中的物品添加到门店在途仓库
   - 数量：使用 Purchase Order 中的采购数量
   - 仓库：门店在途仓库（Transit Warehouse）
   ↓
3. 审批 Purchase Order
   - 状态：Draft → Submitted
   - 在途仓库数量保持不变
   ↓
4. 提交外部供应商
   - 外部供应商准备货物
   - 在途仓库数量保持不变
   ↓
5. 外部供应商配送货物到门店
   - 货物在运输途中
   - 在途仓库数量保持不变
   ↓
6. 门店确认收货
   - 从在途仓库扣减数量
   - 转入目标仓库（门店正常仓库）
   - 使用实际收货数量（可能与 Purchase Order 数量不同）
```

#### 6.3.2 在途仓操作详情

**添加时机**：
- ✅ **Purchase Order 创建时**：表示已向外部供应商下单，货物预期在途
- **原因**：外部供应商发货是线下操作，无法实时跟踪，所以在创建 Purchase Order 时就添加到在途仓库

**添加数量**：
- 使用 Purchase Order 中的采购数量
- 每个物品的数量 = Purchase Order Item 的 `qty` 字段

**添加操作**：
```go
// 在 Purchase Order 创建时，自动添加到在途仓库
func (s *purchaseOrderSrv) handleExternalPurchaseErp(ctx context.Context, tx *gorm.DB, purchaseOrder *model.PurchaseOrder) error {
    // 获取在途仓库
    transitWarehouse, _ := repository.NewWarehouseRepo(tx).GetTransitWarehouse()
    
    // 为每个物品添加到在途仓库
    for _, item := range purchaseOrder.Items {
        actualNum := item.GetConversionRateNum()
        if transitWarehouse != nil {
            err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

**收货时处理**：
- 门店确认收货时，从在途仓库扣减数量
- 转入目标仓库（门店正常仓库）
- 如果实际收货数量与 Purchase Order 数量不同，使用实际收货数量
- 自动处理数量差异（多收/少收）

### 6.4 在途仓数量差异处理

#### 6.4.1 数量差异场景

**场景 1：多收（实际收货数量 > 在途仓数量）**
```
在途仓数量：100
实际收货数量：105
差异：+5（多收 5 个）

处理方式：
1. 从在途仓扣减：100
2. 转入目标仓库：105
3. 差异处理：自动补充在途仓库存 5 个（或直接增加目标仓库 5 个）
```

**场景 2：少收（实际收货数量 < 在途仓数量）**
```
在途仓数量：100
实际收货数量：95
差异：-5（少收 5 个）

处理方式：
1. 从在途仓扣减：95
2. 转入目标仓库：95
3. 差异处理：在途仓剩余 5 个，记录为待收货或退货
```

**场景 3：正常收货（实际收货数量 = 在途仓数量）**
```
在途仓数量：100
实际收货数量：100
差异：0

处理方式：
1. 从在途仓扣减：100
2. 转入目标仓库：100
3. 差异处理：无需处理
```

#### 6.4.2 差异处理逻辑

**处理原则**：
- ✅ 使用实际收货数量更新库存
- ✅ 记录数量差异日志，便于后续对账
- ✅ 支持多收和少收的场景处理

**代码实现**：
```go
// 收货时处理数量差异
func (s *purchaseReceiptOrderSrv) handleQuantityDifference(
    ctx context.Context,
    tx *gorm.DB,
    receiptOrder *model.PurchaseReceiptOrder,
    transitWarehouse *model.Warehouse,
) error {
    for _, item := range receiptOrder.Items {
        transitQty := getTransitWarehouseQty(item.MaterialUuid)  // 在途仓数量
        receivedQty := item.GetUnitsTotalConversionRateNum()     // 实际收货数量
        
        // 从在途仓扣减（使用实际收货数量）
        err := reduceFromTransitWarehouse(ctx, transitWarehouse, item, receivedQty)
        if err != nil {
            return err
        }
        
        // 转入目标仓库
        err = addToTargetWarehouse(ctx, receiptOrder.TargetWarehouse, item, receivedQty)
        if err != nil {
            return err
        }
        
        // 处理数量差异
        if receivedQty > transitQty {
            // 多收：记录差异日志
            logQuantityDifference(item, transitQty, receivedQty, "多收")
        } else if receivedQty < transitQty {
            // 少收：记录差异日志，在途仓剩余数量待处理
            remainingQty := transitQty - receivedQty
            logQuantityDifference(item, transitQty, receivedQty, "少收")
            // 可选：标记在途仓剩余数量待处理
        }
    }
    
    return nil
}
```

### 6.5 在途仓操作总结

| 项目 | 集采部分 | 直采部分 |
|------|---------|---------|
| **进入在途仓时机** | Delivery Note 提交后 | Purchase Order 创建时 |
| **添加数量** | Delivery Note 中的发货数量 | Purchase Order 中的采购数量 |
| **添加操作** | 系统自动添加 | 系统自动添加 |
| **收货时处理** | 从在途仓扣减，转入目标仓库 | 从在途仓扣减，转入目标仓库 |
| **数量差异处理** | 使用实际收货数量，记录差异日志 | 使用实际收货数量，记录差异日志 |
| **代码位置** | `main/app/service/purchase_order/helper.go::AddToTransitWarehouse()` | `main/app/service/purchase_order/purchase_order.go::handleExternalPurchaseErp()` |

**关键点**：
- ✅ **集采**：在 Delivery Note 提交后添加到在途仓（货物已发出）
- ✅ **直采**：在 Purchase Order 创建时添加到在途仓（预期在途）
- ✅ **收货时**：从在途仓扣减，转入目标仓库，使用实际收货数量
- ✅ **差异处理**：自动处理数量差异，记录差异日志

---

## 7. ERP 功能映射

### 6.1 集采功能映射

| TTPOS 操作 | ERPNext 单据/功能 | ERPNext 按钮操作 | 说明 |
|-----------|------------------|-----------------|------|
| 查看待收货列表 | 查询关联 Delivery Note 的采购收货单 | - | 通过 `ErpOrderNo` 字段关联 |
| 查看发货单详情 | `GET /api/resource/Delivery Note/{name}` | 在 Delivery Note 页面查看 | 获取 Delivery Note 详细信息 |
| 扫码收货 | 通过条形码匹配 Delivery Note 物品 | - | 匹配 `barcode_value` 字段 |
| 确认收货 | `CreatePurchaseReceiptOrder` | - | TTPOS 侧创建或更新采购收货单 |
| 更新库存 | 自动更新（不在 ERP 中操作） | - | TTPOS 侧库存更新 |
| 查看收货记录 | 查询采购收货单历史 | - | 通过关联的 Delivery Note 查询 |
| 创建 Inter Company Purchase Receipt（可选） | `make_inter_company_purchase_receipt` | **Delivery Note 页面：Create → Inter Company Purchase Receipt** | 从 Delivery Note 创建跨公司采购收货单（可选） |
| 提交 Purchase Receipt | `POST /api/resource/Purchase Receipt/{name}` (action: submit) | **Purchase Receipt 页面：Submit** | 提交收货单，更新库存 |

### 6.2 直采功能映射

| TTPOS 操作 | ERPNext 单据/功能 | ERPNext 按钮操作 | 说明 |
|-----------|------------------|-----------------|------|
| 查看待收货列表 | 查询 Purchase Order | - | 显示所有待收货的采购单 |
| 查看采购单详情 | `GET /api/resource/Purchase Order/{name}` | 在 Purchase Order 页面查看 | 获取 Purchase Order 详细信息 |
| 扫码收货 | 通过条形码匹配 Purchase Order 物品 | - | 匹配 `barcode_value` 字段 |
| 创建收货单 | `CreatePurchaseReceiptOrder` | - | TTPOS 侧创建收货单（按供应商拆分） |
| 确认收货 | `POST /api/resource/Purchase Receipt` | - | TTPOS 自动为每个供应商创建 Purchase Receipt |
| 提交收货单 | `POST /api/resource/Purchase Receipt/{name}` (action: submit) | **Purchase Receipt 页面：Submit** | 提交 Purchase Receipt，更新库存 |
| 更新库存 | 自动更新（Purchase Receipt 提交后） | 自动执行 | ERPNext 自动更新库存 |
| 更新采购单进度 | 自动更新 `received_qty` | 自动执行 | 收货后自动更新采购单进度 |
| 查看收货记录 | 查询 Purchase Receipt 历史 | - | 通过关联的 Purchase Order 查询 |
| 创建 Purchase Receipt（手动） | `POST /api/resource/Purchase Receipt` | **Purchase Order 页面：Create → Purchase Receipt** | 从 Purchase Order 创建采购收货单（手动操作） |
| 创建采购退货单 | `POST /api/resource/Purchase Invoice` (is_return: 1) | **Purchase Receipt 页面：Create → Purchase Return** | 创建采购退货单 |
| 创建采购发票 | `POST /api/resource/Purchase Invoice` | **Purchase Receipt 页面：Create → Purchase Invoice** | 创建采购发票 |

### 6.3 关键 ERP API 列表

#### 6.3.1 Delivery Note 相关

```python
# 获取 Delivery Note 详情
GET /api/resource/Delivery Note/{name}

# 从 Delivery Note 创建 Inter Company Purchase Receipt（可选）
POST /api/resource/Purchase Receipt
{
    "purchase_receipt_type": "Inter Company Purchase Receipt",
    "against_delivery_note": "MAT-DN-2026-00001",
    ...
}

# 从 Sales Order 创建 Delivery Note
POST /api/resource/Delivery Note
{
    "against_sales_order": "SO-00001",
    ...
}
```

#### 6.3.2 Purchase Receipt 相关

```python
# 从 Purchase Order 创建 Purchase Receipt
POST /api/resource/Purchase Receipt
{
    "purchase_order": "PO-00001",
    "supplier": "Supplier A - Company",
    ...
}

# 提交 Purchase Receipt
POST /api/resource/Purchase Receipt/{name}
{
    "action": "submit"
}
```

#### 6.3.3 库存相关

```python
# 查询库存
GET /api/resource/Stock Ledger Entry
{
    "filters": [["warehouse", "=", "Store Warehouse - Company"]]
}

# 查询物品条形码
GET /api/resource/Item/{item_code}
# 返回中的 barcodes 字段包含条形码信息
```

---

## 8. 实施建议

### 7.1 开发优先级

#### 阶段一：核心流程（P0）

1. **集采流程**
   - ✅ Delivery Note 创建后自动在 TTPOS 创建采购收货单
   - ✅ 门店收货操作界面
   - ✅ 收货后更新库存

2. **直采流程**
   - ✅ Purchase Order 创建后在 TTPOS 显示
   - ✅ 门店收货操作界面
   - ✅ 收货后创建 Purchase Receipt
   - ✅ 收货后更新库存

#### 阶段二：拆单功能（P1）

3. **集采拆单**
   - ✅ 按仓库拆分 Delivery Note
   - ✅ 显示发货仓库信息

4. **直采拆单**
   - ✅ 按供应商拆分 Purchase Receipt
   - ✅ 显示供应商信息

#### 阶段三：优化功能（P2）

5. **扫码收货**
   - ✅ 支持条形码扫码收货
   - ✅ 自动匹配物品和数量

6. **数量差异处理**
   - ✅ 支持多收和少收场景
   - ✅ 自动处理数量差异

### 7.2 技术实现建议

#### 7.2.1 集采收货实现

**方案 A：Webhook 触发（推荐）**

```go
// 处理 Delivery Note 创建的 Webhook
func HandleDeliveryNoteCreated(ctx context.Context, req *DeliveryNoteWebhookReq) error {
    // 1. 获取 Delivery Note 信息
    deliveryNote, err := erpService.GetDeliveryNote(ctx, req.DocumentName)
    
    // 2. 检查是否已创建对应的采购收货单
    existingPR, err := purchaseReceiptOrderRepo.GetByErpOrderNo(deliveryNote.Name)
    if existingPR != nil {
        return nil  // 已存在，跳过
    }
    
    // 3. 获取关联的 Sales Order
    salesOrder, err := erpService.GetSalesOrder(ctx, deliveryNote.AgainstSalesOrder)
    
    // 4. 获取门店信息
    storeCompany, err := getStoreCompanyFromCustomer(ctx, salesOrder.Customer)
    
    // 5. 查找或创建对应的采购单（用于关联）
    purchaseOrder, err := findOrCreatePurchaseOrderFromDeliveryNote(ctx, deliveryNote, storeCompany)
    
    // 6. 创建采购收货单（PR）
    receiptOrder := &model.PurchaseReceiptOrder{
        OrderNo:         generateReceiptOrderNo("SHRK"),
        ErpOrderNo:     deliveryNote.Name,  // 关联 Delivery Note
        PurchaseOrderUuid: purchaseOrder.Uuid,
        ReceiptType:     constant.ReceiptTypeInternal,  // 内部收货（集采）
        Status:          constant.ReceiptOrderStatusPending,
        ...
    }
    
    // 7. 保存采购收货单
    return purchaseReceiptOrderRepo.Create(receiptOrder)
}
```

**方案 B：定时任务同步**

```go
// 定时任务：同步 Delivery Note 并创建采购收货单
func SyncDeliveryNoteAndCreatePurchaseReceiptOrder(ctx context.Context) error {
    // 1. 查询 ERPNext 中已创建但未同步的 Delivery Note
    deliveryNotes, err := erpService.GetCreatedDeliveryNotes(ctx, time.Now().Add(-24*time.Hour))
    
    // 2. 为每个 Delivery Note 创建对应的采购收货单
    for _, dn := range deliveryNotes {
        // 检查是否已创建
        existingPR, _ := purchaseReceiptOrderRepo.GetByErpOrderNo(dn.Name)
        if existingPR != nil {
            continue
        }
        
        // 创建采购收货单
        err := createPurchaseReceiptOrderFromDeliveryNote(ctx, dn)
    }
    
    return nil
}
```

#### 7.2.2 直采收货实现

```go
// 创建直采收货单
func CreatePurchaseReceiptOrderForDirectPurchase(ctx context.Context, req *PurchaseReceiptCreateReq) error {
    // 1. 查询采购单
    purchaseOrder, err := purchaseOrderRepo.GetByUuid(req.PurchaseOrderUuid)
    
    // 2. 按供应商分组物品
    supplierGroups := groupItemsBySupplier(req.Items)
    
    // 3. 为每个供应商创建独立的 Purchase Receipt
    for supplier, items := range supplierGroups {
        // 创建 Purchase Receipt
        purchaseReceipt, err := erpService.CreatePurchaseReceipt(ctx, &CreatePurchaseReceiptReq{
            PurchaseOrderName: purchaseOrder.ErpOrderNo,
            Supplier:          supplier,
            Items:             items,
        })
        
        // 提交 Purchase Receipt
        err = erpService.SubmitPurchaseReceipt(ctx, purchaseReceipt.Name)
        
        // 更新采购单收货进度
        err = updatePurchaseOrderReceivedQty(ctx, purchaseOrder, items)
    }
    
    return nil
}
```

### 7.3 测试建议

#### 7.3.1 功能测试

1. **集采收货测试**
   - ✅ 测试 Delivery Note 创建后自动创建采购收货单
   - ✅ 测试门店收货操作
   - ✅ 测试库存更新

2. **直采收货测试**
   - ✅ 测试 Purchase Order 创建后在 TTPOS 显示
   - ✅ 测试按供应商拆分收货单
   - ✅ 测试收货后创建 Purchase Receipt
   - ✅ 测试库存更新

3. **拆单功能测试**
   - ✅ 测试按仓库拆分 Delivery Note
   - ✅ 测试按供应商拆分 Purchase Receipt

#### 7.3.2 集成测试

1. **完整流程测试**
   - ✅ 测试从 MR 到收货的完整流程
   - ✅ 测试集采和直采并行流程

2. **异常场景测试**
   - ✅ 测试数量差异处理
   - ✅ 测试部分收货场景
   - ✅ 测试退货场景

### 7.4 部署建议

1. **ERPNext 配置**
   - 配置 Delivery Note 创建时的 Webhook（如果使用 Webhook 方案）
   - 配置用户权限（门店收货人员权限）

2. **TTPOS 部署**
   - 部署后端代码
   - 更新前端界面
   - 配置收货流程

3. **数据迁移**
   - 检查历史数据
   - 如需要，创建历史数据的采购收货单

---

## 9. 总结

### 8.1 核心要点

1. **集采流程**：
   - 走销售订单业务线（Sales Order → Delivery Note）
   - 按仓库拆分 Delivery Note
   - Delivery Note 创建后自动在 TTPOS 创建采购收货单

2. **直采流程**：
   - 走采购订单业务线（Purchase Order → Purchase Receipt）
   - 按供应商拆分 Purchase Receipt
   - Purchase Order 创建后在 TTPOS 显示

3. **收货操作**：
   - 统一在 TTPOS 的品采收货页面操作
   - 支持扫码收货和手动收货
   - 自动更新 ERP 单据和库存

### 8.2 关键优势

1. **流程清晰**：集采和直采流程明确，操作人员易于理解
2. **拆单合理**：按仓库和供应商拆单，便于收货管理
3. **自动化高**：自动创建单据、自动更新库存
4. **追溯方便**：每个收货单都有关联的 ERP 单据，便于对账和追溯

---

**文档版本**：v1.0  
**创建日期**：2026-01-16  
**最后更新**：2026-01-16  
**维护者**：TTPOS Team

