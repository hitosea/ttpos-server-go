# 品牌采购 Drop Ship 混合模型 需求提案

## 📋 提案信息

| 项目          | 内容                             |
| ------------- | -------------------------------- |
| **提案人**    | BenDayeCoder                     |
| **日期**      | 2026-01-23                       |
| **目标版本**  | 待定                             |
| **状态**      | 待评审                           |
| **关联 Spec** | -                                |

---

## 🎯 背景和动机

### 问题描述

当前品牌采购（内部采购）流程中，子门店向总部采购物资时，存在以下限制：

1. **单一发货路径**：所有商品默认从总部仓库发货，不支持外部供应商直发
2. **无 DN 自动创建**：收货直接从 PO 创建 PR，跳过了 DN（发货单）环节
3. **Drop Ship 标记未生效**：BMP 已有 Drop Ship 商品标记逻辑（`processDripShopItems()`），但无后续处理流程

**现有流程的局限性**：

```
当前流程：所有商品统一处理，无法区分供应类型
─────────────────────────────────────────────────
门店 MR → 门店 PO → 总部 SO → (无 DN) → 门店 PR
                                ↑
                         所有商品走同一路径
                         Drop Ship 标记被忽略
```

### 业务价值

- **供应链灵活性提升**：支持总部代采外部供应商商品，外部供应商可直发门店
- **仓库运营效率提升**：按仓库拆分 DN，各仓库可并行作业
- **财务核算清晰**：内部供应和代采分开处理，账务流转更清晰
- **库存管理优化**：Drop Ship 商品不占用总部仓库库存

### 目标用户

- [x] 仓库管理员（主要）- 总部各仓库的库管/操作人员
- [x] 采购人员（主要）- 处理 Drop Ship 商品的采购对接
- [x] 运营主管（次要）- 监控整体采购履行情况的管理层
- [x] 门店收货人员 - 区分不同来源的收货
- [ ] 收银员
- [ ] 店长
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

利用 ERPNext 原生的 **Drop Ship（供应商直发）** 功能，通过 Item 的 `delivered_by_supplier` 字段标识，实现品牌采购的混合供应模型：

- **内部供应商品**（`delivered_by_supplier = 0`）：从总部仓库发货，按仓库拆分创建 DN
- **Drop Ship 商品**（`delivered_by_supplier = 1`）：总部代采，外部供应商直发门店

**目标流程**：

```
目标流程：按商品类型分流处理
─────────────────────────────────────────────────────────────────
门店 MR → 门店 PO → 总部 SO
                      ↓
         ┌───────────┴───────────┐
         ↓                       ↓
   内部供应 Item            Drop Ship Item
   (delivered_by_supplier=0) (delivered_by_supplier=1)
         ↓                       ↓
   N × DN (按仓库拆分)      N × 外部 PO (按供应商拆分)
         ↓                       ↓
   总部仓库发货              外部供应商直发门店
         ↓                       ↓
   门店 PR (从 DN)          门店 PR (从外部 PO)
```

### 核心功能点

1. **【新增】DN 按仓库拆分创建**
   - SO 创建后，自动为内部供应 Item 创建 DN
   - 按 Item 的 `warehouse` 字段分组，每个仓库一个 DN
   - 过滤掉 `delivered_by_supplier = 1` 的 Item

2. **【新增】外部 PO 按供应商拆分创建**
   - SO 创建后，自动为 Drop Ship Item 创建外部 PO
   - 按 Item 的 `supplier` 字段分组，每个供应商一个 PO
   - 供应商取自 Item 的 `supplier_items[0].supplier`

3. **【修改】总部库存扣减逻辑**
   - 仅对内部供应 Item 扣减库存
   - 跳过 Drop Ship Item（不占用总部库存）

4. **【新增】多来源收货流程**
   - 内部供应：从 DN 创建 PR
   - Drop Ship：从外部 PO 创建 PR，同时更新 SO 履约状态

### 影响范围

**涉及终端**：
- [x] Shop 商家管理端（总部审批、收货确认）
- [ ] POS 收银端
- [ ] KDS 厨显端
- [ ] QDS 排号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [ ] Kiosk 自助点餐机

**涉及模块**：
- [x] API 接口（gRPC 参数结构调整）
- [x] 数据模型（Item 级别新增 Drop Ship 标识和仓库信息）
- [x] 业务逻辑（Main 层库存扣减、BMP 层 DN/PO 创建）
- [ ] UI 组件（可选：收货界面区分来源）
- [x] 其他：ERPNext 单据关联关系

**涉及系统**：
- TTPOS Main（总部审批逻辑、库存扣减、收货流程）
- TTPOS BMP（DN 创建、外部 PO 创建、PR 创建）
- ERPNext（单据存储、Drop Ship 原生支持）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [x] **高**：涉及架构调整、第三方集成、复杂算法

**复杂度说明**：
- 需要改造 Main 层审批流程，新增 Item 分类和库存扣减过滤
- BMP 层需实现 DN 按仓库创建、外部 PO 按供应商创建
- 收货流程需支持多来源（DN / 外部 PO）
- 涉及跨系统数据一致性（TTPOS ↔ ERPNext）

### 现有代码基础

**已实现（可复用）**：

| 功能               | 文件位置                                              | 方法                       |
| ------------------ | ----------------------------------------------------- | -------------------------- |
| Drop Ship 商品判断 | `ttpos-bmp/.../logic/buying/buying.go`                | `isDripShopItem()` :561    |
| 获取第一供应商     | `ttpos-bmp/.../logic/buying/buying.go`                | `selectFirstSupplier()` :577 |
| Drop Ship 标记     | `ttpos-bmp/.../logic/buying/buying.go`                | `processDripShopItems()` :517 |
| Item 字段定义      | `ttpos-bmp/.../model/dto/erp/item.go`                 | `DeliveredBySupplier` :53  |

**待实现**：

| 功能                 | 涉及模块 | 说明                           |
| -------------------- | -------- | ------------------------------ |
| DN 按仓库创建        | BMP      | 调用 `make_delivery_note` API  |
| 外部 PO 按供应商创建 | BMP      | 调用 `make_purchase_order_for_default_supplier` API |
| 库存扣减过滤         | Main     | 跳过 Drop Ship Item            |
| 从 DN 创建 PR        | BMP      | 调用 `make_purchase_receipt` API（来源 DN） |
| Drop Ship 收货流程   | Main/BMP | 新增收货入口和 gRPC 接口       |

### 工作量预估

- **预估 SP**: 23-26

### 拆分预估

**是否需要拆分**：
- [ ] **否**：单终端，SP ≤ 5，可直接创建 1 个 Spec
- [x] **是**：需要拆分为多个 Spec

**拆分维度**：
- [x] 按功能模块拆分：BMP 层 + Main 层
- [x] 按 Phase 拆分：基础能力 → 收货流程 → UI 优化

**预估 Spec 数量**：5-6 个

**预估 Spec 列表**：

```
Phase 1: 基础能力
─────────────────
┌─────────────────────────┐
│ Spec 1                  │
│ BMP: DN 按仓库拆分创建   │
│ SP: 5                   │
└───────────┬─────────────┘
            │ 依赖
            ↓
┌─────────────────────────┐     ┌─────────────────────────┐
│ Spec 2                  │     │ Spec 3                  │
│ BMP: 外部 PO 按供应商    │     │ Main: 库存扣减逻辑改造  │
│ 拆分创建                │     │ SP: 3                   │
│ SP: 5                   │     └─────────────────────────┘
└───────────┬─────────────┘

Phase 2: 收货流程
─────────────────
┌─────────────────────────┐     ┌─────────────────────────┐
│ Spec 4                  │     │ Spec 5                  │
│ Main/BMP: 内部供应收货   │     │ Main/BMP: Drop Ship     │
│ (从 DN 创建 PR)         │     │ 收货流程                │
│ SP: 5                   │     │ SP: 5                   │
└─────────────────────────┘     └─────────────────────────┘

Phase 3: 可选优化
─────────────────
┌─────────────────────────┐
│ Spec 6 (可选)           │
│ UI: 收货界面区分来源    │
│ SP: 3                   │
└─────────────────────────┘
```

### 风险识别

**潜在风险**：

| 风险                 | 影响 | 概率 | 缓解措施                                                   |
| -------------------- | ---- | ---- | ---------------------------------------------------------- |
| ERPNext API 兼容性   | 高   | 中   | 先在测试环境验证 `make_delivery_note` 和 `make_purchase_order_for_default_supplier` API |
| 并发创建失败         | 中   | 中   | 实现统一回滚机制，失败时取消已创建的单据                   |
| 数据不一致           | 高   | 低   | 依赖 ERPNext 的事务机制，BMP 侧做好错误处理                |
| 收货流程复杂化       | 中   | 高   | 门店端 UI 需要清晰区分两种收货类型                         |
| Item 无 Drop Ship 配置 | 低   | 中   | 提供默认行为：未配置时按内部供应处理                       |

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名 | 签名/日期 |
| ---------- | ---- | --------- |
| 产品经理   |      |           |
| 技术负责人 |      |           |
| 开发代表   |      |           |
| 测试代表   |      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待评审]
```

**下一步行动**：

- [ ] 在测试环境验证 ERPNext Drop Ship API
- [ ] 创建 Spec：`story-erp-brand-purchase-dn-split`
- [ ] 创建 Spec：`story-erp-brand-purchase-dropship-po`
- [ ] 分配负责人：待定
- [ ] 目标 Sprint：待定

---

## 📝 附录

### A. 新旧流程对比

#### A.1 功能对比总表

| 功能点               | 当前实现 | 目标实现 | 变更类型 |
| -------------------- | -------- | -------- | -------- |
| MR 创建              | ✅ 有     | ✅ 保持   | 无变更   |
| PO 创建（门店→总部） | ✅ 有     | ✅ 保持   | 无变更   |
| SO 创建              | ✅ 有     | ✅ 保持   | 无变更   |
| Drop Ship Item 标记  | ✅ 有     | ✅ 保持   | 无变更   |
| 获取第一供应商       | ✅ 有     | ✅ 保持   | 无变更   |
| DN 创建              | ❌ 无     | ✅ 新增   | **新增** |
| DN 按仓库拆分        | ❌ 无     | ✅ 新增   | **新增** |
| DN 过滤 Drop Ship    | ❌ 无     | ✅ 新增   | **新增** |
| 外部 PO 创建         | ❌ 无     | ✅ 新增   | **新增** |
| 外部 PO 按供应商拆分 | ❌ 无     | ✅ 新增   | **新增** |
| PR 从 DN 创建        | ⚠️ 部分  | ✅ 完善   | **修改** |
| PR 从外部 PO 创建    | ❌ 无     | ✅ 新增   | **新增** |
| 总部库存扣减过滤     | ❌ 无     | ✅ 新增   | **修改** |

#### A.2 审批流程对比

```
【当前】审批后处理
────────────────────────────────────────────────────────
handleInternalPurchaseErp()
    │
    ├── reduceHeadquarterStockAndLog()  ← 所有 Item 都扣库存
    │       └── 从指定仓库扣减
    │
    └── SaveMaterialRequest() [gRPC]
            └── CreateMaterialRequest()
            └── CreatePurchaseFromMq()
            └── CreateInnerSaleOrderFromPurchaseOrder()
                    └── processDripShopItems()  ← 仅标记，无后续


【目标】审批后处理
────────────────────────────────────────────────────────
handleInternalPurchaseErp()
    │
    ├── 【修改】分类 Item
    │       ├── 内部供应 Item (delivered_by_supplier = 0)
    │       └── Drop Ship Item (delivered_by_supplier = 1)
    │
    ├── 【修改】reduceHeadquarterStockAndLog()
    │       └── 仅扣减内部供应 Item 的库存
    │       └── 【新增】按仓库分组库存扣减
    │
    └── SaveMaterialRequest() [gRPC]
            └── CreateMaterialRequest()
            └── CreatePurchaseFromMq()
            └── CreateInnerSaleOrderFromPurchaseOrder()
                    └── processDripShopItems()
                    └── 【新增】createDeliveryNotes()
                    │       └── 过滤 Drop Ship Item
                    │       └── 按仓库分组
                    │       └── 循环创建 DN
                    └── 【新增】createDropShipPurchaseOrders()
                            └── 过滤内部供应 Item
                            └── 按供应商分组
                            └── 循环创建外部 PO
```

#### A.3 收货流程对比

```
【当前】收货流程
────────────────────────────────────────────────────────
CreatePurchaseReceiptOrder()
    │
    └── SavePurchaseReceipt() [gRPC]
            └── CreatePurchaseReceiptFromOrder()
                    └── make_purchase_receipt (从 PO)
                    └── Document.Create(PR)


【目标】收货流程
────────────────────────────────────────────────────────
【路径A】内部供应收货
CreatePurchaseReceiptOrder()
    │
    ├── 【修改】判断收货来源类型
    │       └── 内部供应 → 从 DN 创建 PR
    │
    └── SavePurchaseReceipt() [gRPC]
            └── 【修改】CreatePurchaseReceiptFromDeliveryNote()
                    └── make_purchase_receipt (从 DN)
                    └── Document.Create(PR)


【路径B】Drop Ship 收货
【新增】CreateDropShipReceiptOrder()
    │
    ├── 判断收货来源类型
    │       └── Drop Ship → 从外部 PO 创建 PR
    │
    └── 【新增】SaveDropShipReceipt() [gRPC]
            └── 【新增】CreatePurchaseReceiptFromDropShipPO()
                    └── make_purchase_receipt (从外部 PO)
                    └── Document.Create(PR)
                    └── 【新增】更新 SO Item 状态为 Delivered
```

### B. 目标架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    品牌采购 Drop Ship 混合模型（目标架构）                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  门店 (TTPOS Main)                       总部/BMP                           │
│  ────────────────                        ────────                           │
│                                                                             │
│  1. 创建采购申请                                                            │
│     PurchaseType = 2                                                        │
│     Items: [A, B, C, D]                                                     │
│          ↓                                                                  │
│  2. 提交 → 门店审批                                                         │
│          ↓                                                                  │
│  3. 总部审批 ──────────────────────────► 4. handleInternalPurchaseErp()    │
│                                              ├── reduceHeadquarterStockAndLog()
│                                              │   (仅内部供应 Item)           │
│                                              └── SaveMaterialRequest() [gRPC]
│                                                        ↓                    │
│                                          5. BMP 处理                        │
│                                              ├── CreateMaterialRequest()    │
│                                              ├── CreatePurchaseFromMq()     │
│                                              └── CreateInnerSaleOrderFromPurchaseOrder()
│                                                  └── processDripShopItems() │
│                                                        ↓                    │
│                                          ┌─────────────┴─────────────┐      │
│                                          ↓                           ↓      │
│                                    内部供应 Item              Drop Ship Item │
│                                    (A, B)                     (C, D)        │
│                                          ↓                           ↓      │
│                               【新增】按仓库创建 DN      【新增】按供应商创建外部 PO
│                                    ├── DN-仓库M              ├── PO-供应商X │
│                                    └── DN-仓库N              └── PO-供应商Y │
│                                          ↓                           ↓      │
│                                    总部仓库发货            外部供应商直发门店 │
│                                          ↓                           ↓      │
│  门店收货                                │                           │      │
│  ────────                                │                           │      │
│  6a. 内部供应收货 ←──────────────────────┘                           │      │
│      CreatePurchaseReceiptOrder()                                    │      │
│      来源: DN                                                        │      │
│          ↓                                                           │      │
│  6b. Drop Ship 收货 ←────────────────────────────────────────────────┘      │
│      【新增】CreateDropShipReceiptOrder()                                   │
│      来源: 外部 PO                                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### C. ERPNext 单据关系

```
门店公司                                    总部公司
────────                                    ────────
Material Request (MR)
      ↓
Purchase Order (PO) ←── Inter-Company ──► Sales Order (SO)
      │                                    ├── 内部供应 Item
      │                                    │       ↓
      │                                    │   N × Delivery Note (DN)  【按仓库拆分】
      │                                    │       ↓
      │                                    └── Drop Ship Item
      │                                            ↓
      │                                    N × Purchase Order (外部PO) 【按供应商拆分】
      │                                            ↓
      │                                    外部供应商直发门店
      ↓                                            ↓
N × Purchase Receipt (PR)              Purchase Receipt (外部PO确认)
├── 从 DN 创建 (内部供应)
└── 从外部 PO 创建 (Drop Ship)  【新增来源】
```

### D. 关键代码位置索引

| 功能             | 文件路径                                                  | 方法/行号                                         |
| ---------------- | --------------------------------------------------------- | ------------------------------------------------- |
| 采购审批入口     | `main/app/service/purchase_order/purchase_order.go`       | `ApprovePurchaseOrder()` :1079                    |
| 内部采购处理     | `main/app/service/purchase_order/purchase_order.go`       | `handleInternalPurchaseErp()` :1387               |
| 总部库存扣减     | `main/app/service/purchase_order/helper.go`               | `reduceHeadquarterStockAndLog()` :362             |
| gRPC 调用 MR     | `main/app/service/rpc/erp/stock.go`                       | `SaveMaterialRequest()` :25                       |
| BMP MR 创建      | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock.go`   | `CreateMaterialRequest()` :271                    |
| BMP PO 创建      | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` | `CreatePurchaseFromMq()` :30                      |
| BMP SO 创建      | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` | `CreateInnerSaleOrderFromPurchaseOrder()` :121    |
| Drop Ship 标记   | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` | `processDripShopItems()` :517                     |
| Drop Ship 判断   | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` | `isDripShopItem()` :561                           |
| 获取第一供应商   | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` | `selectFirstSupplier()` :577                      |
| 门店收货入口     | `main/app/service/purchase_order/receipt_order.go`        | `CreatePurchaseReceiptOrder()` :47                |
| gRPC 调用 PR     | `main/app/service/purchase_order/receipt_order.go`        | `SavePurchaseReceipt()` :928                      |
| BMP PR 创建      | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` | `CreatePurchaseReceiptFromOrder()` :272           |

### E. ERPNext API 参考

| 用途                  | API Method                                                                     |
| --------------------- | ------------------------------------------------------------------------------ |
| 从 MR 创建 PO         | `erpnext.stock.doctype.material_request.material_request.make_purchase_order`  |
| 从 PO 创建 SO         | `erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order` |
| 从 SO 创建 DN         | `erpnext.selling.doctype.sales_order.sales_order.make_delivery_note`           |
| 从 SO 创建 Drop Ship PO | `erpnext.selling.doctype.sales_order.sales_order.make_purchase_order_for_default_supplier` |
| 从 PO 创建 PR         | `erpnext.buying.doctype.purchase_order.purchase_order.make_purchase_receipt`   |
| 从 DN 创建 PR         | `erpnext.stock.doctype.delivery_note.delivery_note.make_purchase_receipt`      |

### F. 关键字段说明

| 字段                  | 所属 DocType        | 说明                               |
| --------------------- | ------------------- | ---------------------------------- |
| `delivered_by_supplier` | Item                | 1=Drop Ship 商品，0=普通商品       |
| `supplier_items`      | Item                | 供应商列表，取第一个作为默认外部供应商 |
| `warehouse`           | SO Item / DN Item   | 来源仓库                           |
| `supplier`            | SO Item (Drop Ship) | 外部供应商                         |

---

**版本**: v1.0.0
