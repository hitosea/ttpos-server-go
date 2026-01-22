# 品牌采购（内部采购）及收货流程

> 子门店向总部采购物资的完整技术实现，包括 TTPOS Main、BMP、ERPNext 之间的交互细节。

**相关文档：**
- [采购与调拨完整合集](./purchase-transfer-flow.md)
- [外部采购流程](./external-purchase-flow.md)
- [门店调拨流程](./store-transfer-flow.md)

---

## 目录

1. [系统架构总览](#系统架构总览)
2. [BMP 与 ERPNext 交互机制](#bmp-与-erpnext-交互机制)
3. [业务流程图](#业务流程图)
4. [采购审批调用链路](#采购审批调用链路)
5. [收货调用链路](#收货调用链路)
6. [关键文件索引](#关键文件索引)

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

## 业务流程图

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
│                                     │       ↓          │  (ERPNext)        │
│                                     │ Sales Order      │                    │
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

### 采购类型判断

外部采购与内部采购通过 `PurchaseType` 字段区分：
- `PurchaseType=1`：外部采购，触发 `handleExternalPurchaseErp()` 流程
- `PurchaseType=2`：内部采购（品牌采购），触发 `handleInternalPurchaseErp()` 流程

---

## 采购审批调用链路

### TTPOS Main 处理（子门店）

| 步骤 | 方法                                      | 说明                             | 同步/异步   |
| ---- | ----------------------------------------- | -------------------------------- | ----------- |
| 1    | `purchaseOrderSrv.CreatePurchaseOrder()`  | 创建采购申请单（PurchaseType=2） | 同步-本地DB |
| 2    | `purchaseOrderSrv.SubmitPurchaseOrder()`  | 提交采购申请                     | 同步-本地DB |
| 3    | `purchaseOrderSrv.ApprovePurchaseOrder()` | 子门店审批通过                   | 同步        |
| 3.1  | `└─ 状态变更`                             | Pending → HeadquarterPending     | 同步-本地DB |

### TTPOS Main 处理（总部）

| 步骤  | 方法                                           | 说明                     | 同步/异步     |
| ----- | ---------------------------------------------- | ------------------------ | ------------- |
| 4     | `purchaseOrderSrv.ApprovePurchaseOrder()`      | 总部审批通过             | 同步          |
| 4.1   | `└─ handleInternalPurchaseErp()`               | 处理内部采购 ERP 逻辑    | 同步          |
| 4.1.1 | `    └─ helper.reduceHeadquarterStockAndLog()` | 减总部库存，记录出库日志 | 同步-本地DB   |
| 4.1.2 | `    └─ erp.SaveMaterialRequest()`             | **gRPC 调用 BMP**        | **同步-gRPC** |

### BMP 处理 (gRPC: SaveMaterialRequest)

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

### ERPNext 单据生成

| 步骤 | DocType            | 公司视角 | 状态      | 说明                                   |
| ---- | ------------------ | -------- | --------- | -------------------------------------- |
| 1    | `Material Request` | 子门店   | Submitted | 物料申请单                             |
| 2    | `Purchase Order`   | 子门店   | Submitted | 内部采购订单（关联 MR）                |
| 3    | `Sales Order`      | 总部     | Submitted | 内部销售订单（Inter-Company，关联 PO） |

---

## 收货调用链路

收货流程与外部采购基本一致，差异点：

| 差异项         | 外部采购 | 品牌采购                 |
| -------------- | -------- | ------------------------ |
| 收货单编号前缀 | `PRC`    | `TPHY`                   |
| 总部同步       | 无       | 同步更新总部采购申请明细 |

### TTPOS Main 处理

| 步骤  | 方法                                            | 说明                 | 同步/异步     |
| ----- | ----------------------------------------------- | -------------------- | ------------- |
| 1     | `purchaseOrderSrv.CreatePurchaseReceiptOrder()` | 创建收货单           | 同步          |
| 1.1   | `└─ receiptSrv.CreatePurchaseReceiptOrder()`    | 实际创建逻辑         | 同步          |
| 1.1.1 | `    └─ 创建收货单记录`                         | 保存到本地数据库     | 同步-本地DB   |
| 1.1.2 | `    └─ 更新采购申请明细到货数量`               | 更新 ArrivalNum      | 同步-本地DB   |
| 1.1.3 | `    └─ erp.SavePurchaseReceipt()`              | **gRPC 调用 BMP**    | **同步-gRPC** |
| 1.2   | `└─ helper.checkAndUpdatePurchaseOrderStatus()` | 检查并更新采购单状态 | 同步-本地DB   |
| 1.3   | `└─ syncToHeadquarter()`                        | 同步更新总部申请明细 | 同步-本地DB   |

### BMP 处理 (gRPC: SavePurchaseReceipt)

| 步骤 | 方法                                      | ERPNext API                                               | 说明               | 同步/异步     |
| ---- | ----------------------------------------- | --------------------------------------------------------- | ------------------ | ------------- |
| 1    | `Buying.CreatePurchaseReceiptFromOrder()` | -                                                         | 入口方法           | 同步          |
| 2    | `└─ Rpc.Execute()`                        | `POST /api/v2/method/frappe.model.mapper.make_mapped_doc` | 从 PO 生成 PR 模板 | **同步-HTTP** |
| 3    | `└─ 设置 PR 字段`                         | -                                                         | 设置仓库、数量等   | 内存操作      |
| 4    | `└─ Document.Create()`                    | `POST /api/v2/document/Purchase Receipt`                  | 创建收货单         | **同步-HTTP** |
| 5    | `└─ Document.ChangeDocStatus()`           | `PUT /api/v2/document/Purchase Receipt/{name}`            | 提交收货单         | **同步-HTTP** |

### ERPNext 单据生成

| 步骤 | DocType            | 状态      | 说明                 |
| ---- | ------------------ | --------- | -------------------- |
| 1    | `Purchase Receipt` | Draft     | 从 PO 创建收货单草稿 |
| 2    | `Purchase Receipt` | Submitted | 提交收货单，更新库存 |

---

## 关键文件索引

### TTPOS Main

| 模块     | 文件路径                                            | 说明               |
| -------- | --------------------------------------------------- | ------------------ |
| 采购服务 | `main/app/service/purchase_order/purchase_order.go` | 采购申请主逻辑     |
| 收货服务 | `main/app/service/purchase_order/receipt_order.go`  | 收货单逻辑         |
| 采购辅助 | `main/app/service/purchase_order/helper.go`         | 辅助方法           |
| ERP-库存 | `main/app/service/rpc/erp/stock.go`                 | gRPC 客户端-库存   |

### BMP

| 模块        | 文件路径                                                  | 说明           |
| ----------- | --------------------------------------------------------- | -------------- |
| gRPC-库存   | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/`  | 库存 gRPC 服务 |
| gRPC-采购   | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/buying/` | 采购 gRPC 服务 |
| 库存逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/`           | 库存业务逻辑   |
| 采购逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/`          | 采购业务逻辑   |
| ERPNext通信 | `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/`         | HTTP 客户端    |

### ERPNext DocType

| DocType            | 中文名     | 用途                        |
| ------------------ | ---------- | --------------------------- |
| `Material Request` | 物料申请   | 品牌采购起点                |
| `Purchase Order`   | 采购订单   | 内部采购订单（关联 MR）     |
| `Purchase Receipt` | 采购收货单 | 确认收货、更新库存          |
| `Sales Order`      | 销售订单   | 内部销售（Inter-Company）   |
| `Supplier`         | 供应商     | 内部供应商（总部）          |

---

## 变更记录

| 日期       | 版本 | 变更说明       | 作者     |
| ---------- | ---- | -------------- | -------- |
| 2025-12-25 | 1.0  | 从合集文档拆分 | AI Agent |
