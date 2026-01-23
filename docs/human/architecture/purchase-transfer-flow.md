# 采购与调拨完整调用链路

> 本文档详细描述外部采购、品牌采购（内部采购）、门店调拨三种业务场景的完整技术实现，包括 TTPOS Main、BMP、ERPNext 之间的交互细节。

**独立文档索引（按业务流程拆分）：**
- [外部采购及收货流程](./external-purchase-flow.md) - 向外部供应商采购物资
- [品牌采购（内部采购）流程](./internal-purchase-flow.md) - 子门店向总部采购物资
- [门店调拨流程](./store-transfer-flow.md) - 门店之间的物资转移

---

## 目录

1. [系统架构总览](#系统架构总览)
2. [BMP 与 ERPNext 交互机制](#bmp-与-erpnext-交互机制)
3. [外部采购及收货](#一外部采购及收货)
4. [品牌采购（内部采购）及收货](#二品牌采购内部采购及收货)
5. [门店调拨](#三门店调拨)
6. [三种场景对比总结](#四三种场景对比总结)
7. [关键文件索引](#五关键文件索引)

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

## 一、外部采购及收货

> 向外部供应商采购物资

### 1.1 业务流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              外部采购完整流程                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  创建 ──► 提交 ──► 审批 ──► 收货                                              │
│   │        │        │        │                                              │
│   ▼        ▼        ▼        ▼                                              │
│ Draft → Pending → Approved → PartialReceived/AllReceived                    │
│                      │                │                                     │
│                      ▼                ▼                                     │
│               [调用 BMP]        [调用 BMP]                                   │
│                      │                │                                     │
│                      ▼                ▼                                     │
│               Purchase Order   Purchase Receipt                             │
│               (ERPNext)        (ERPNext)                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 采购审批调用链路

#### TTPOS Main 处理

| 步骤  | 方法                                      | 说明                             | 同步/异步     |
| ----- | ----------------------------------------- | -------------------------------- | ------------- |
| 1     | `purchaseOrderSrv.CreatePurchaseOrder()`  | 创建采购申请单（PurchaseType=1） | 同步-本地DB   |
| 2     | `purchaseOrderSrv.SubmitPurchaseOrder()`  | 提交采购申请，状态 Draft→Pending | 同步-本地DB   |
| 3     | `purchaseOrderSrv.ApprovePurchaseOrder()` | 审批通过                         | 同步          |
| 3.1   | `└─ handleExternalPurchaseErp()`          | 处理外部采购 ERP 逻辑            | 同步          |
| 3.1.1 | `    └─ helper.AddToTransitWarehouse()`   | 添加物品到本店在途仓库           | 同步-本地DB   |
| 3.1.2 | `    └─ erp.CreatePurchaseOrder()`        | **gRPC 调用 BMP**                | **同步-gRPC** |

#### BMP 处理 (gRPC: CreatePurchaseOrder)

| 步骤 | 方法                            | ERPNext API                                  | 说明                            | 同步/异步     |
| ---- | ------------------------------- | -------------------------------------------- | ------------------------------- | ------------- |
| 1    | `Buying.CreatePurchaseOrder()`  | -                                            | 入口方法                        | 同步          |
| 2    | `└─ Document.Create()`          | `POST /api/v2/document/Purchase Order`       | 创建采购订单                    | **同步-HTTP** |
| 3    | `└─ Document.ChangeDocStatus()` | `PUT /api/v2/document/Purchase Order/{name}` | 提交采购订单（Draft→Submitted） | **同步-HTTP** |

#### ERPNext 单据生成

| 步骤 | DocType          | 状态      | 说明               |
| ---- | ---------------- | --------- | ------------------ |
| 1    | `Purchase Order` | Draft     | 创建采购订单草稿   |
| 2    | `Purchase Order` | Submitted | 提交采购订单，生效 |

### 1.3 收货调用链路

#### TTPOS Main 处理

| 步骤  | 方法                                            | 说明                 | 同步/异步     |
| ----- | ----------------------------------------------- | -------------------- | ------------- |
| 1     | `purchaseOrderSrv.CreatePurchaseReceiptOrder()` | 创建收货单           | 同步          |
| 1.1   | `└─ receiptSrv.CreatePurchaseReceiptOrder()`    | 实际创建逻辑         | 同步          |
| 1.1.1 | `    └─ 创建收货单记录`                         | 保存到本地数据库     | 同步-本地DB   |
| 1.1.2 | `    └─ 更新采购申请明细到货数量`               | 更新 ArrivalNum      | 同步-本地DB   |
| 1.1.3 | `    └─ erp.SavePurchaseReceipt()`              | **gRPC 调用 BMP**    | **同步-gRPC** |
| 1.2   | `└─ helper.checkAndUpdatePurchaseOrderStatus()` | 检查并更新采购单状态 | 同步-本地DB   |

#### BMP 处理 (gRPC: SavePurchaseReceipt)

| 步骤 | 方法                                      | ERPNext API                                               | 说明               | 同步/异步     |
| ---- | ----------------------------------------- | --------------------------------------------------------- | ------------------ | ------------- |
| 1    | `Buying.CreatePurchaseReceiptFromOrder()` | -                                                         | 入口方法           | 同步          |
| 2    | `└─ Rpc.Execute()`                        | `POST /api/v2/method/frappe.model.mapper.make_mapped_doc` | 从 PO 生成 PR 模板 | **同步-HTTP** |
| 3    | `└─ 设置 PR 字段`                         | -                                                         | 设置仓库、数量等   | 内存操作      |
| 4    | `└─ Document.Create()`                    | `POST /api/v2/document/Purchase Receipt`                  | 创建收货单         | **同步-HTTP** |
| 5    | `└─ Document.ChangeDocStatus()`           | `PUT /api/v2/document/Purchase Receipt/{name}`            | 提交收货单         | **同步-HTTP** |

#### ERPNext 单据生成

| 步骤 | DocType            | 状态      | 说明                 |
| ---- | ------------------ | --------- | -------------------- |
| 1    | `Purchase Receipt` | Draft     | 从 PO 创建收货单草稿 |
| 2    | `Purchase Receipt` | Submitted | 提交收货单，更新库存 |

---

## 二、品牌采购（内部采购）及收货

> 子门店向总部采购物资

### 2.1 业务流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           品牌采购（内部采购）完整流程                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  子门店                              总部                                    │
│  ┌──────────────────────────┐       ┌──────────────────────────┐           │
│  │ 创建 → 提交 → 门店审批    │──────►│ 总部审批                  │           │
│  │ Draft → Pending → HQPending │      │ HQPending → Approved     │           │
│  └──────────────────────────┘       └──────────────────────────┘           │
│                                              │                              │
│                                              ▼                              │
│                                       [调用 BMP]                            │
│                                              │                              │
│                                              ▼                              │
│                                     ┌────────────────┐                     │
│                                     │ Material Request │                    │
│                                     │       ↓          │                    │
│                                     │ Purchase Order   │                    │
│                                     │       ↓          │                    │
│                                     │ Sales Order      │  (ERPNext)        │
│                                     └────────────────┘                     │
│                                                                             │
│  子门店收货                                                                  │
│  ┌──────────────────────────┐                                              │
│  │ 创建收货单 → 确认收货     │ ──────► [调用 BMP] → Purchase Receipt        │
│  │ Approved → AllReceived   │                                              │
│  └──────────────────────────┘                                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 采购审批调用链路

#### TTPOS Main 处理（子门店）

| 步骤 | 方法                                      | 说明                             | 同步/异步   |
| ---- | ----------------------------------------- | -------------------------------- | ----------- |
| 1    | `purchaseOrderSrv.CreatePurchaseOrder()`  | 创建采购申请单（PurchaseType=2） | 同步-本地DB |
| 2    | `purchaseOrderSrv.SubmitPurchaseOrder()`  | 提交采购申请                     | 同步-本地DB |
| 3    | `purchaseOrderSrv.ApprovePurchaseOrder()` | 子门店审批通过                   | 同步        |
| 3.1  | `└─ 状态变更`                             | Pending → HeadquarterPending     | 同步-本地DB |

#### TTPOS Main 处理（总部）

| 步骤  | 方法                                           | 说明                     | 同步/异步     |
| ----- | ---------------------------------------------- | ------------------------ | ------------- |
| 4     | `purchaseOrderSrv.ApprovePurchaseOrder()`      | 总部审批通过             | 同步          |
| 4.1   | `└─ handleInternalPurchaseErp()`               | 处理内部采购 ERP 逻辑    | 同步          |
| 4.1.1 | `    └─ helper.reduceHeadquarterStockAndLog()` | 减总部库存，记录出库日志 | 同步-本地DB   |
| 4.1.2 | `    └─ erp.SaveMaterialRequest()`             | **gRPC 调用 BMP**        | **同步-gRPC** |

#### BMP 处理 (gRPC: SaveMaterialRequest)

| 步骤 | 方法                                                | ERPNext API                                               | 说明                       | 同步/异步     |
| ---- | --------------------------------------------------- | --------------------------------------------------------- | -------------------------- | ------------- |
| 1    | `StockController.SaveMaterialRequest()`             | -                                                         | gRPC 入口（Controller 层） | 同步          |
| 2    | `└─ Stock.CreateMaterialRequest()`                  | -                                                         | Service 层处理             | 同步          |
| 2.1  | `    └─ Document.Create()`                          | `POST /api/v2/document/Material Request`                  | 创建物料申请               | **同步-HTTP** |
| 2.2  | `    └─ Document.ChangeDocStatus()`                 | `PUT /api/v2/document/Material Request/{name}`            | 提交物料申请               | **同步-HTTP** |
| 3    | `└─ Buying.CreatePurchaseFromMq()`                  | -                                                         | 从 MR 创建 PO              | 同步          |
| 3.1  | `    └─ Rpc.Execute()`                              | `POST make_mapped_doc` + `make_purchase_order`            | 从 MR 生成 PO 模板         | **同步-HTTP** |
| 3.2  | `    └─ Document.Create()`                          | `POST /api/v2/document/Purchase Order`                    | 创建采购订单               | **同步-HTTP** |
| 3.3  | `    └─ Document.ChangeDocStatus()`                 | `PUT /api/v2/document/Purchase Order/{name}`              | 提交采购订单               | **同步-HTTP** |
| 4    | `└─ Buying.CreateInnerSaleOrderFromPurchaseOrder()` | -                                                         | 从 PO 创建内部 SO          | 同步          |
| 4.1  | `    └─ Rpc.Execute()`                              | `POST make_mapped_doc` + `make_inter_company_sales_order` | 从 PO 生成 SO 模板         | **同步-HTTP** |
| 4.2  | `    └─ Document.Create()`                          | `POST /api/v2/document/Sales Order`                       | 创建内部销售订单           | **同步-HTTP** |
| 4.3  | `    └─ Document.ChangeDocStatus()`                 | `PUT /api/v2/document/Sales Order/{name}`                 | 提交销售订单               | **同步-HTTP** |

#### ERPNext 单据生成

| 步骤 | DocType            | 公司视角 | 状态      | 说明                                   |
| ---- | ------------------ | -------- | --------- | -------------------------------------- |
| 1    | `Material Request` | 子门店   | Submitted | 物料申请单                             |
| 2    | `Purchase Order`   | 子门店   | Submitted | 内部采购订单（关联 MR）                |
| 3    | `Sales Order`      | 总部     | Submitted | 内部销售订单（Inter-Company，关联 PO） |

### 2.3 收货调用链路

收货流程与外部采购基本一致，差异点：

| 差异项         | 外部采购 | 品牌采购                 |
| -------------- | -------- | ------------------------ |
| 收货单编号前缀 | `PRC`    | `TPHY`                   |
| 总部同步       | 无       | 同步更新总部采购申请明细 |

---

## 三、门店调拨

> 门店之间的物资转移

### 3.1 调拨场景分类

根据门店组织架构，调拨分为三种场景：

| 场景       | 组织关系           | 单据流转              | ERPNext 单据组数 |
| ---------- | ------------------ | --------------------- | ---------------- |
| **Case 1** | 无父级 或 父级相同 | A → B 直接调拨        | 1 组             |
| **Case 2** | 任一方有父级       | A → 上级 → B          | 2 组             |
| **Case 3** | 父级不同           | A → A上级 → B上级 → B | 3 组             |

### 3.2 业务流程图

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

### 3.3 审批通过调用链路（以 Case 3 为例）

#### TTPOS Main 处理

| 步骤 | 方法                                      | 说明                 | 同步/异步     |
| ---- | ----------------------------------------- | -------------------- | ------------- |
| 1    | `transferOrderSrv.ApproveTransferOrder()` | 最终审批通过         | 同步          |
| 1.1  | `└─ helper.UpdateStockInTransit()`        | 更新各节点在途仓库存 | 同步-多DB事务 |
| 1.2  | `└─ helper.SaveMaterialTransfer()`        | **gRPC 调用 BMP**    | **同步-gRPC** |
| 2    | `└─ 状态变更`                             | Pending → Receiving  | 同步-本地DB   |

#### BMP 处理 (gRPC: MaterialTransfer)

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

**CreateInnerTransferReceipt 详细流程：**

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

#### ERPNext 单据生成（Case 3 示例）

| 节点                      | 调出方单据 | 调入方单据 | 自动收货 |
| ------------------------- | ---------- | ---------- | -------- |
| **Step 1: A → A上级**     | SO₁ + DN₁  | PO₁ + PR₁  | ✅        |
| **Step 2: A上级 → B上级** | SO₂ + DN₂  | PO₂ + PR₂  | ✅        |
| **Step 3: B上级 → B**     | SO₃ + DN₃  | PO₃        | ❌        |

### 3.4 收货调用链路

#### TTPOS Main 处理

| 步骤 | 方法                                      | 说明                          | 同步/异步     |
| ---- | ----------------------------------------- | ----------------------------- | ------------- |
| 1    | `transferOrderSrv.ReceiveTransferOrder()` | 收货门店确认收货              | 同步          |
| 1.1  | `└─ 验证收货权限`                         | 只有收货门店可收货            | 同步          |
| 1.2  | `└─ helper.MoveStockToTargetWarehouse()`  | 更新本地库存（在途仓→目标仓） | 同步-本地DB   |
| 1.3  | `└─ helper.SavePurchaseReceipt()`         | **gRPC 调用 BMP**             | **同步-gRPC** |
| 2    | `└─ 状态变更`                             | Receiving → Completed         | 同步-本地DB   |

#### BMP 处理 (gRPC: SavePurchaseReceipt)

| 步骤 | 方法                            | ERPNext API                                      | 说明                      | 同步/异步     |
| ---- | ------------------------------- | ------------------------------------------------ | ------------------------- | ------------- |
| 1    | `Buying.SavePurchaseReceipt()`  | -                                                | 入口方法                  | 同步          |
| 2    | `└─ Rpc.Execute()`              | `POST make_mapped_doc` + `make_purchase_receipt` | 从 ToReceipt.PoNo 生成 PR | **同步-HTTP** |
| 3    | `└─ Document.Create()`          | `POST /api/v2/document/Purchase Receipt`         | 创建收货单                | **同步-HTTP** |
| 4    | `└─ Document.ChangeDocStatus()` | `PUT /api/v2/document/Purchase Receipt/{name}`   | 提交收货单                | **同步-HTTP** |

### 3.5 Case 3 单据流向示意图

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

## 四、三种场景对比总结

### 4.1 业务维度对比

| 维度         | 外部采购         | 品牌采购（内部采购） | 门店调拨               |
| ------------ | ---------------- | -------------------- | ---------------------- |
| **业务场景** | 向外部供应商采购 | 子门店向总部采购     | 门店之间物资转移       |
| **采购类型** | `PurchaseType=1` | `PurchaseType=2`     | `TransferOrder`        |
| **供应商**   | 外部供应商       | 总部（内部供应商）   | 发送门店（内部供应商） |
| **库存来源** | 外部             | 总部仓库             | 发送门店仓库           |

### 4.2 技术维度对比

| 维度              | 外部采购              | 品牌采购              | 门店调拨               |
| ----------------- | --------------------- | --------------------- | ---------------------- |
| **审批流程**      | 本店审批              | 本店 → 总部           | 多级审批（最多4级）    |
| **BMP 入口**      | `CreatePurchaseOrder` | `SaveMaterialRequest` | `MaterialTransfer`     |
| **ERPNext 起点**  | Purchase Order        | Material Request      | Sales Order            |
| **Inter-Company** | ❌                     | ✅ (PO → Inner SO)     | ✅ (SO → PO)            |
| **收货方式**      | 手动创建收货单        | 手动创建收货单        | 中间节点自动，最终手动 |

### 4.3 ERPNext 单据对比

| 场景                  | 生成的 DocType    | 数量        |
| --------------------- | ----------------- | ----------- |
| **外部采购**          | PO → PR           | 2 种        |
| **品牌采购**          | MR → PO → SO → PR | 4 种        |
| **门店调拨 (Case 1)** | SO + DN → PO + PR | 4 种 × 1 组 |
| **门店调拨 (Case 2)** | SO + DN → PO + PR | 4 种 × 2 组 |
| **门店调拨 (Case 3)** | SO + DN → PO + PR | 4 种 × 3 组 |

---

## 五、关键文件索引

### 5.1 TTPOS Main

| 模块     | 文件路径                                            | 说明               |
| -------- | --------------------------------------------------- | ------------------ |
| 采购服务 | `main/app/service/purchase_order/purchase_order.go` | 采购申请主逻辑     |
| 收货服务 | `main/app/service/purchase_order/receipt_order.go`  | 收货单逻辑         |
| 采购辅助 | `main/app/service/purchase_order/helper.go`         | 辅助方法           |
| 调拨服务 | `main/app/service/transfer_order/transfer_order.go` | 调拨单主逻辑       |
| 调拨辅助 | `main/app/service/transfer_order/helper.go`         | 辅助方法、ERP 调用 |
| ERP-采购 | `main/app/service/rpc/erp/buying.go`                | gRPC 客户端-采购   |
| ERP-库存 | `main/app/service/rpc/erp/stock.go`                 | gRPC 客户端-库存   |
| ERP-调拨 | `main/app/service/rpc/erp/material_transfer.go`     | gRPC 客户端-调拨   |

### 5.2 BMP

| 模块        | 文件路径                                                             | 说明           |
| ----------- | -------------------------------------------------------------------- | -------------- |
| gRPC-采购   | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/buying/`            | 采购 gRPC 服务 |
| gRPC-库存   | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/`             | 库存 gRPC 服务 |
| gRPC-调拨   | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/material_transfer/` | 调拨 gRPC 服务 |
| 采购逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/`                     | 采购业务逻辑   |
| 销售逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/`                    | 销售业务逻辑   |
| 库存逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/`                      | 库存业务逻辑   |
| 调拨逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/material_transfer.go`  | 调拨核心逻辑   |
| ERPNext通信 | `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/`                    | HTTP 客户端    |

### 5.3 ERPNext DocType

| DocType            | 中文名     | 用途                        |
| ------------------ | ---------- | --------------------------- |
| `Material Request` | 物料申请   | 品牌采购起点                |
| `Purchase Order`   | 采购订单   | 采购/调入方采购             |
| `Purchase Receipt` | 采购收货单 | 确认收货、更新库存          |
| `Sales Order`      | 销售订单   | 调出方销售/内部销售         |
| `Delivery Note`    | 送货单     | 调出方发货                  |
| `Customer`         | 客户       | 内部客户（调拨）            |
| `Supplier`         | 供应商     | 内部供应商（调拨/品牌采购） |

---

## 变更记录

| 日期       | 版本 | 变更说明                           | 作者     |
| ---------- | ---- | ---------------------------------- | -------- |
| 2025-12-25 | 1.0  | 初始版本                           | AI Agent |
| 2026-01-21 | 1.1  | 添加独立文档索引链接，拆分三个流程 | AI Agent |


