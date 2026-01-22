# 门店调拨流程

> 门店之间物资转移的完整技术实现，包括 TTPOS Main、BMP、ERPNext 之间的交互细节。

**相关文档：**
- [采购与调拨完整合集](./purchase-transfer-flow.md)
- [外部采购流程](./external-purchase-flow.md)
- [品牌采购（内部采购）流程](./internal-purchase-flow.md)

---

## 目录

1. [系统架构总览](#系统架构总览)
2. [BMP 与 ERPNext 交互机制](#bmp-与-erpnext-交互机制)
3. [调拨场景分类](#调拨场景分类)
4. [业务流程图](#业务流程图)
5. [审批通过调用链路](#审批通过调用链路)
6. [收货调用链路](#收货调用链路)
7. [Case 3 单据流向示意图](#case-3-单据流向示意图)
8. [关键文件索引](#关键文件索引)

---

## 系统架构总览

```
┌─────────────────┐       gRPC        ┌─────────────────┐     HTTP API     ┌─────────────────┐
│   TTPOS Main    │ ───────────────► │   TTPOS BMP     │ ───────────────► │    ERPNext      │
│  (门店 POS)     │    同步调用       │  (业务中台)     │    同步调用      │   (ERP 系统)    │
└─────────────────┘                   └─────────────────┘                  └─────────────────┘
      │                                      │                                    │
      │ 本地数据库                            │ 本地数据库                          │ Frappe DB
      │ (MySQL)                              │ (MySQL)                            │ (MariaDB)
      ▼                                      ▼                                    ▼
┌─────────────────┐                   ┌─────────────────┐                  ┌─────────────────┐
│ ttpos_purchase  │                   │ ttpos_shop_*    │                  │ DocType Tables  │
│ ttpos_transfer  │                   │ (收银员等)      │                  │ (SO/PO/DN/PR)   │
└─────────────────┘                   └─────────────────┘                  └─────────────────┘
```

---

## BMP 与 ERPNext 交互机制

### 通信方式

BMP 通过 **HTTP 同步调用** 与 ERPNext 交互，所有调用都是阻塞式的。

| 服务类型     | API 路径                     | HTTP 方法           | 用途                    |
| ------------ | ---------------------------- | ------------------- | ----------------------- |
| **Document** | `/api/v2/document/{DocType}` | GET/POST/PUT/DELETE | 文档 CRUD 操作          |
| **Rpc**      | `/api/v2/method/{method}`    | POST                | 调用 ERPNext 服务端方法 |

### 核心服务实现

```go
// Document 服务 - 文档操作
// 文件: ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go

func (s *sDocument) Create(ctx, docType, data)      // POST /api/v2/document/{DocType}
func (s *sDocument) Update(ctx, req, data)          // PUT /api/v2/document/{DocType}/{name}
func (s *sDocument) ChangeDocStatus(ctx, doctype, name, docstatus)  // 修改文档状态

// Rpc 服务 - 方法调用
// 文件: ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go

func (s *sRpc) Execute(ctx, req, params)            // POST /api/v2/method/{method}
```

### ERPNext 常用 API 方法

| 方法名                                | 用途                 | 调用场景           |
| ------------------------------------- | -------------------- | ------------------ |
| `frappe.model.mapper.make_mapped_doc` | 从源文档创建关联文档 | SO→PO, PO→PR 等    |
| `frappe.client.insert`                | 创建文档             | 创建 SO/PO 等      |
| `frappe.client.submit`                | 提交文档             | 提交草稿状态的文档 |
| `frappe.desk.form.save.cancel`        | 取消已提交文档       | 回滚操作           |

### 文档状态常量

```go
const (
    DocstatusDraft     = "0"  // 草稿
    DocstatusSubmitted = "1"  // 已提交
    DocstatusCancelled = "2"  // 已取消
)
```

---

## 调拨场景分类

根据门店组织架构，调拨分为三种场景：

| 场景       | 组织关系           | 单据流转              | ERPNext 单据组数 |
| ---------- | ------------------ | --------------------- | ---------------- |
| **Case 1** | 无父级 或 父级相同 | A → B 直接调拨        | 1 组             |
| **Case 2** | 任一方有父级       | A → 上级 → B          | 2 组             |
| **Case 3** | 父级不同           | A → A上级 → B上级 → B | 3 组             |

---

## 业务流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              门店调拨完整流程                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────┐     │
│  │                        多级审批流程                                │     │
│  │  发送方门店 → 发送方上级 → 接收方上级 → 接收方门店                   │     │
│  │  (可选审批)    (可选审批)    (可选审批)    (最终审批)                │     │
│  └───────────────────────────────────────────────────────────────────┘     │
│                                     │                                       │
│                                     ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────┐     │
│  │                     最终审批通过时                                  │     │
│  │  1. UpdateStockInTransit() - 更新在途仓库存                        │     │
│  │  2. SaveMaterialTransfer() - 调用 BMP 创建 ERPNext 单据            │     │
│  │  3. 状态变更: Pending → Receiving                                  │     │
│  └───────────────────────────────────────────────────────────────────┘     │
│                                     │                                       │
│                                     ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────┐     │
│  │                        收货门店收货                                │     │
│  │  1. ReceiveTransferOrder() - 确认收货                              │     │
│  │  2. UpdateStockForReceive() - 更新本地库存                         │     │
│  │  3. SavePurchaseReceipt() - 调用 BMP 创建收货单                    │     │
│  │  4. 状态变更: Receiving → Completed                                │     │
│  └───────────────────────────────────────────────────────────────────┘     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 审批通过调用链路（以 Case 3 为例）

### TTPOS Main 处理

| 步骤 | 方法                                      | 说明                 | 同步/异步     |
| ---- | ----------------------------------------- | -------------------- | ------------- |
| 1    | `transferOrderSrv.ApproveTransferOrder()` | 最终审批通过         | 同步          |
| 1.1  | `└─ helper.UpdateStockInTransit()`        | 更新各节点在途仓库存 | 同步-多DB事务 |
| 1.2  | `└─ helper.SaveMaterialTransfer()`        | **gRPC 调用 BMP**    | **同步-gRPC** |
| 2    | `└─ 状态变更`                             | Pending → Receiving  | 同步-本地DB   |

### BMP 处理 (gRPC: MaterialTransfer)

**Controller 层入口：**

| 步骤 | 方法                                               | 说明            | 同步/异步 |
| ---- | -------------------------------------------------- | --------------- | --------- |
| 1    | `MaterialTransferController.MaterialTransfer()`    | gRPC 入口       | 同步      |
| 2    | `└─ service.MaterialTransfer().MaterialTransfer()` | 调用 Service 层 | 同步      |

**Service 层逻辑（Case 3 - 跨组织调拨）：**

| 步骤 | 方法                                               | 说明                               | 同步/异步   |
| ---- | -------------------------------------------------- | ---------------------------------- | ----------- |
| 1    | `sMaterialTransfer.MaterialTransfer()`             | 入口，判断 Case 类型               | 同步        |
| 2    | **Step 1: 门店A → A上级**                          |                                    |             |
| 2.1  | `└─ getTransitWarehouse()`                         | 获取 A上级在途仓                   | 同步-本地DB |
| 2.2  | `└─ CreateInnerTransferReceipt(autoReceipt=true)`  | 创建调拨单据                       | 同步        |
| 3    | **Step 2: A上级 → B上级**                          |                                    |             |
| 3.1  | `└─ getTransitWarehouse()`                         | 获取 B上级在途仓                   | 同步-本地DB |
| 3.2  | `└─ CreateInnerTransferReceipt(autoReceipt=true)`  | 创建调拨单据                       | 同步        |
| 4    | **Step 3: B上级 → 门店B**                          |                                    |             |
| 4.1  | `└─ CreateInnerTransferReceipt(autoReceipt=false)` | 创建调拨单据（最终节点不自动收货） | 同步        |

### CreateInnerTransferReceipt 详细流程

| 步骤  | 方法                                             | ERPNext API                                                  | 说明                   | 同步/异步     |
| ----- | ------------------------------------------------ | ------------------------------------------------------------ | ---------------------- | ------------- |
| 1     | **准备交易对象**                                 |                                                              |                        |               |
| 1.1   | `Selling.ListCustomers()`                        | `GET /api/v2/document/Customer`                              | 查询是否存在内部客户   | **同步-HTTP** |
| 1.2   | `Selling.CreateCustomer()`                       | `POST /api/v2/document/Customer`                             | 如不存在则创建         | **同步-HTTP** |
| 1.3   | `Selling.AddCompanyToCustomer()`                 | `PUT /api/v2/document/Customer/{name}`                       | 添加交易公司           | **同步-HTTP** |
| 1.4   | `Supplier.ListSuppliers()`                       | `GET /api/v2/document/Supplier`                              | 查询是否存在内部供应商 | **同步-HTTP** |
| 1.5   | `Supplier.CreateSupplier()`                      | `POST /api/v2/document/Supplier`                             | 如不存在则创建         | **同步-HTTP** |
| 1.6   | `Supplier.AddSupplerTransactCompany()`           | `PUT /api/v2/document/Supplier/{name}`                       | 添加交易公司           | **同步-HTTP** |
| 2     | **创建销售订单（调出方）**                       |                                                              |                        |               |
| 2.1   | `Selling.CreateSalesOrder()`                     | `POST /api/v2/document/Sales Order`                          | 创建销售订单           | **同步-HTTP** |
| 2.2   | `Selling.SubmitSalesOrder()`                     | `PUT /api/v2/document/Sales Order/{name}`                    | 提交销售订单           | **同步-HTTP** |
| 3     | **创建送货单（调出方）**                         |                                                              |                        |               |
| 3.1   | `DeliveryNote.CreateDeliveryNoteFromSaleOrder()` | -                                                            | 从 SO 创建 DN          | 同步          |
| 3.1.1 | `└─ Rpc.Execute()`                               | `POST make_mapped_doc` + `make_delivery_note`                | 从 SO 生成 DN 模板     | **同步-HTTP** |
| 3.1.2 | `└─ Document.Create()`                           | `POST /api/v2/document/Delivery Note`                        | 创建送货单             | **同步-HTTP** |
| 3.1.3 | `└─ Document.ChangeDocStatus()`                  | `PUT /api/v2/document/Delivery Note/{name}`                  | 提交送货单             | **同步-HTTP** |
| 4     | **创建采购订单（调入方）**                       |                                                              |                        |               |
| 4.1   | `Buying.CreatePurchaseOrderFromSalesOrder()`     | -                                                            | 从 SO 创建 PO          | 同步          |
| 4.1.1 | `└─ Rpc.Execute()`                               | `POST make_mapped_doc` + `make_inter_company_purchase_order` | 从 SO 生成 PO 模板     | **同步-HTTP** |
| 4.1.2 | `└─ Document.Create()`                           | `POST /api/v2/document/Purchase Order`                       | 创建采购订单           | **同步-HTTP** |
| 4.1.3 | `└─ Document.ChangeDocStatus()`                  | `PUT /api/v2/document/Purchase Order/{name}`                 | 提交采购订单           | **同步-HTTP** |
| 5     | **创建收货单（如 autoReceipt=true）**            |                                                              |                        |               |
| 5.1   | `Buying.CreatePurchaseReceiptFromOrder()`        | -                                                            | 从 PO 创建 PR          | 同步          |
| 5.1.1 | `└─ Rpc.Execute()`                               | `POST make_mapped_doc` + `make_purchase_receipt`             | 从 PO 生成 PR 模板     | **同步-HTTP** |
| 5.1.2 | `└─ Document.Create()`                           | `POST /api/v2/document/Purchase Receipt`                     | 创建收货单             | **同步-HTTP** |
| 5.1.3 | `└─ Document.ChangeDocStatus()`                  | `PUT /api/v2/document/Purchase Receipt/{name}`               | 提交收货单             | **同步-HTTP** |

### ERPNext 单据生成（Case 3 示例）

| 节点                      | 调出方单据 | 调入方单据 | 自动收货 |
| ------------------------- | ---------- | ---------- | -------- |
| **Step 1: A → A上级**     | SO₁ + DN₁  | PO₁ + PR₁  | ✅        |
| **Step 2: A上级 → B上级** | SO₂ + DN₂  | PO₂ + PR₂  | ✅        |
| **Step 3: B上级 → B**     | SO₃ + DN₃  | PO₃        | ❌        |

---

## 收货调用链路

### TTPOS Main 处理

| 步骤 | 方法                                      | 说明                          | 同步/异步     |
| ---- | ----------------------------------------- | ----------------------------- | ------------- |
| 1    | `transferOrderSrv.ReceiveTransferOrder()` | 收货门店确认收货              | 同步          |
| 1.1  | `└─ 验证收货权限`                         | 只有收货门店可收货            | 同步          |
| 1.2  | `└─ helper.MoveStockToTargetWarehouse()`  | 更新本地库存（在途仓→目标仓） | 同步-本地DB   |
| 1.3  | `└─ helper.SavePurchaseReceipt()`         | **gRPC 调用 BMP**             | **同步-gRPC** |
| 2    | `└─ 状态变更`                             | Receiving → Completed         | 同步-本地DB   |

### BMP 处理 (gRPC: SavePurchaseReceipt)

| 步骤 | 方法                            | ERPNext API                                      | 说明                      | 同步/异步     |
| ---- | ------------------------------- | ------------------------------------------------ | ------------------------- | ------------- |
| 1    | `Buying.SavePurchaseReceipt()`  | -                                                | 入口方法                  | 同步          |
| 2    | `└─ Rpc.Execute()`              | `POST make_mapped_doc` + `make_purchase_receipt` | 从 ToReceipt.PoNo 生成 PR | **同步-HTTP** |
| 3    | `└─ Document.Create()`          | `POST /api/v2/document/Purchase Receipt`         | 创建收货单                | **同步-HTTP** |
| 4    | `└─ Document.ChangeDocStatus()` | `PUT /api/v2/document/Purchase Receipt/{name}`   | 提交收货单                | **同步-HTTP** |

---

## Case 3 单据流向示意图

```
┌──────────────────────────────────────────────────────────────────────────────┐
│               门店A → 门店B (不同父级) 完整单据流向                            │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  门店A              A上级              B上级              门店B               │
│  ┌─────┐           ┌─────┐           ┌─────┐           ┌─────┐              │
│  │仓库A│──出库──► │在途仓│──出库──► │在途仓│──出库──► │仓库B│              │
│  └─────┘           └─────┘           └─────┘           └─────┘              │
│                                                                              │
│  ERPNext 单据:                                                               │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │                                                                     │     │
│  │  Step 1: 门店A → A上级                                              │     │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐                      │     │
│  │  │   SO₁    │───►│   DN₁    │    │   PO₁    │───► PR₁ (自动收货)   │     │
│  │  │(门店A)   │    │(发货)    │    │(A上级)   │                       │     │
│  │  └──────────┘    └──────────┘    └──────────┘                      │     │
│  │                                                                     │     │
│  │  Step 2: A上级 → B上级                                              │     │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐                      │     │
│  │  │   SO₂    │───►│   DN₂    │    │   PO₂    │───► PR₂ (自动收货)   │     │
│  │  │(A上级)   │    │(发货)    │    │(B上级)   │                       │     │
│  │  └──────────┘    └──────────┘    └──────────┘                      │     │
│  │                                                                     │     │
│  │  Step 3: B上级 → 门店B                                              │     │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐                      │     │
│  │  │   SO₃    │───►│   DN₃    │    │   PO₃    │───► PR₃ (手动收货)   │     │
│  │  │(B上级)   │    │(发货)    │    │(门店B)   │                       │     │
│  │  └──────────┘    └──────────┘    └──────────┘                      │     │
│  │                                                                     │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                                                              │
│  图例:                                                                       │
│  SO = Sales Order (销售订单)      DN = Delivery Note (送货单)                │
│  PO = Purchase Order (采购订单)   PR = Purchase Receipt (收货单)             │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## 关键文件索引

### TTPOS Main

| 模块     | 文件路径                                            | 说明               |
| -------- | --------------------------------------------------- | ------------------ |
| 调拨服务 | `main/app/service/transfer_order/transfer_order.go` | 调拨单主逻辑       |
| 调拨辅助 | `main/app/service/transfer_order/helper.go`         | 辅助方法、ERP 调用 |
| ERP-调拨 | `main/app/service/rpc/erp/material_transfer.go`     | gRPC 客户端-调拨   |

### BMP

| 模块        | 文件路径                                                             | 说明           |
| ----------- | -------------------------------------------------------------------- | -------------- |
| gRPC-调拨   | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/material_transfer/` | 调拨 gRPC 服务 |
| 调拨逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/material_transfer.go`  | 调拨核心逻辑   |
| 销售逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/`                    | 销售业务逻辑   |
| 采购逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/`                     | 采购业务逻辑   |
| ERPNext通信 | `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/`                    | HTTP 客户端    |

### ERPNext DocType

| DocType            | 中文名     | 用途                        |
| ------------------ | ---------- | --------------------------- |
| `Sales Order`      | 销售订单   | 调出方销售                  |
| `Delivery Note`    | 送货单     | 调出方发货                  |
| `Purchase Order`   | 采购订单   | 调入方采购                  |
| `Purchase Receipt` | 采购收货单 | 确认收货、更新库存          |
| `Customer`         | 客户       | 内部客户（调拨）            |
| `Supplier`         | 供应商     | 内部供应商（调拨）          |

---

## 变更记录

| 日期       | 版本 | 变更说明       | 作者     |
| ---------- | ---- | -------------- | -------- |
| 2025-12-25 | 1.0  | 从合集文档拆分 | AI Agent |
