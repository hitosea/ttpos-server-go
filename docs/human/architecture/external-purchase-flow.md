# 外部采购及收货流程

> 向外部供应商采购物资的完整技术实现，包括 TTPOS Main、BMP、ERPNext 之间的交互细节。

**相关文档：**
- [采购与调拨完整合集](./purchase-transfer-flow.md)
- [品牌采购（内部采购）流程](./internal-purchase-flow.md)
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

### 采购类型判断

外部采购与内部采购通过 `PurchaseType` 字段区分：
- `PurchaseType=1`：外部采购，触发 `handleExternalPurchaseErp()` 流程
- `PurchaseType=2`：内部采购（品牌采购），触发 `handleInternalPurchaseErp()` 流程

---

## 采购审批调用链路

### TTPOS Main 处理

| 步骤  | 方法                                      | 说明                             | 同步/异步     |
| ----- | ----------------------------------------- | -------------------------------- | ------------- |
| 1     | `purchaseOrderSrv.CreatePurchaseOrder()`  | 创建采购申请单（PurchaseType=1） | 同步-本地DB   |
| 2     | `purchaseOrderSrv.SubmitPurchaseOrder()`  | 提交采购申请，状态 Draft→Pending | 同步-本地DB   |
| 3     | `purchaseOrderSrv.ApprovePurchaseOrder()` | 审批通过                         | 同步          |
| 3.1   | `└─ handleExternalPurchaseErp()`          | 处理外部采购 ERP 逻辑            | 同步          |
| 3.1.1 | `    └─ helper.AddToTransitWarehouse()`   | 添加物品到本店在途仓库           | 同步-本地DB   |
| 3.1.2 | `    └─ erp.CreatePurchaseOrder()`        | **gRPC 调用 BMP**                | **同步-gRPC** |

### 子店审批流程（可选）

当子店创建的采购订单需要总部审批时：

| 步骤 | 方法                           | 说明                           | 同步/异步   |
| ---- | ------------------------------ | ------------------------------ | ----------- |
| 3.2  | `└─ handleSubShopApproval()`   | 子店审批后等待总部             | 同步        |
| 3.3  | `└─ syncToSubShop()`           | 总部审批后同步回子店           | 同步        |

### BMP 处理 (gRPC: CreatePurchaseOrder)

| 步骤 | 方法                            | ERPNext API                                  | 说明                            | 同步/异步     |
| ---- | ------------------------------- | -------------------------------------------- | ------------------------------- | ------------- |
| 1    | `Buying.CreatePurchaseOrder()`  | -                                            | 入口方法                        | 同步          |
| 2    | `└─ Document.Create()`          | `POST /api/v2/document/Purchase Order`       | 创建采购订单                    | **同步-HTTP** |
| 3    | `└─ Document.ChangeDocStatus()` | `PUT /api/v2/document/Purchase Order/{name}` | 提交采购订单（Draft→Submitted） | **同步-HTTP** |

### ERPNext 单据生成

| 步骤 | DocType          | 状态      | 说明               |
| ---- | ---------------- | --------- | ------------------ |
| 1    | `Purchase Order` | Draft     | 创建采购订单草稿   |
| 2    | `Purchase Order` | Submitted | 提交采购订单，生效 |

### 错误处理机制

当 ERP 调用失败时，`handleErpError()` 负责处理回滚逻辑：

| 错误类型       | 处理方式                         |
| -------------- | -------------------------------- |
| 供应商被禁用   | 返回错误，提示用户检查供应商状态 |
| 物品被禁用     | 返回错误，提示用户检查物品状态   |
| 网络/超时错误  | 重试或回滚本地状态变更           |
| 其他 ERP 错误  | 记录日志，返回详细错误信息       |

---

## 收货调用链路

### TTPOS Main 处理

| 步骤  | 方法                                            | 说明                 | 同步/异步     |
| ----- | ----------------------------------------------- | -------------------- | ------------- |
| 1     | `purchaseOrderSrv.CreatePurchaseReceiptOrder()` | 创建收货单           | 同步          |
| 1.1   | `└─ receiptSrv.CreatePurchaseReceiptOrder()`    | 实际创建逻辑         | 同步          |
| 1.1.1 | `    └─ 创建收货单记录`                         | 保存到本地数据库     | 同步-本地DB   |
| 1.1.2 | `    └─ 更新采购申请明细到货数量`               | 更新 ArrivalNum      | 同步-本地DB   |
| 1.1.3 | `    └─ erp.SavePurchaseReceipt()`              | **gRPC 调用 BMP**    | **同步-gRPC** |
| 1.2   | `└─ helper.checkAndUpdatePurchaseOrderStatus()` | 检查并更新采购单状态 | 同步-本地DB   |

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
| ERP-采购 | `main/app/service/rpc/erp/buying.go`                | gRPC 客户端-采购   |

### BMP

| 模块        | 文件路径                                                  | 说明           |
| ----------- | --------------------------------------------------------- | -------------- |
| gRPC-采购   | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/buying/` | 采购 gRPC 服务 |
| 采购逻辑    | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/`          | 采购业务逻辑   |
| ERPNext通信 | `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/`         | HTTP 客户端    |

### ERPNext DocType

| DocType            | 中文名     | 用途                   |
| ------------------ | ---------- | ---------------------- |
| `Purchase Order`   | 采购订单   | 外部采购订单           |
| `Purchase Receipt` | 采购收货单 | 确认收货、更新库存     |
| `Supplier`         | 供应商     | 外部供应商             |

---

## 变更记录

| 日期       | 版本 | 变更说明                       | 作者     |
| ---------- | ---- | ------------------------------ | -------- |
| 2025-12-25 | 1.0  | 从合集文档拆分                 | AI Agent |
| 2026-01-21 | 1.1  | 补充子店审批流程和错误处理机制 | AI Agent |
