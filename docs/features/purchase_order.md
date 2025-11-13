# Purchase Order Service 采购单服务说明文档

## 📋 概述

`service/purchase_order/purchase_order.go` 是 TTPOS 系统的采购单管理服务，负责处理采购申请、审核流程、收货管理等核心功能。该服务支持外部采购和内部采购两种类型，支持总部审核流程，并与 ERP 系统集成。

**文件路径**: `/main/app/service/purchase_order/purchase_order.go`  
**代码行数**: 1370 行  
**接口定义**: `IPurchaseOrderSrv`  
**实现结构**: `purchaseOrderSrv`

---

## 🏗️ 架构设计

### 接口定义 (IPurchaseOrderSrv)

```go
type IPurchaseOrderSrv interface {
    // 采购申请管理
    GetPurchaseOrderList(ctx context.Context, req req.PurchaseOrderListReq) (resp.PurchaseOrderListResp, error)
    GetPurchaseOrderDetail(ctx context.Context, req req.PurchaseOrderDetailReq) (resp.PurchaseOrderDetailResp, error)
    CreatePurchaseOrder(ctx context.Context, req req.PurchaseOrderCreateReq) (resp.PurchaseOrderCreateResp, error)
    UpdatePurchaseOrder(ctx context.Context, req req.PurchaseOrderUpdateReq) error
    DeletePurchaseOrder(ctx context.Context, req req.PurchaseOrderDeleteReq) error
    SubmitPurchaseOrder(ctx context.Context, req req.PurchaseOrderSubmitReq) error
    ApprovePurchaseOrder(ctx context.Context, req req.PurchaseOrderApproveReq) error

    // 收货管理
    CreatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptCreateReq) (resp.PurchaseReceiptOrderCreateResp, error)
    GetPurchaseReceiptOrderList(ctx context.Context, req req.PurchaseReceiptOrderListReq) (resp.PurchaseReceiptOrderListResp, error)
    GetPurchaseReceiptOrderDetail(ctx context.Context, req req.PurchaseReceiptOrderDetailReq) (resp.PurchaseReceiptOrderDetailResp, error)
    UpdatePurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderUpdateReq) error
    CancelPurchaseReceiptOrder(ctx context.Context, req req.PurchaseReceiptOrderCancelReq) error
}
```

### 依赖服务

```go
type purchaseOrderSrv struct {
    dbm        *database.DBManager      // 数据库管理器
    validator  *purchaseOrderValidator  // 验证器
    helper     *purchaseOrderHelper     // 辅助方法
    receiptSrv *purchaseReceiptOrderSrv // 收货单服务
    lock       lock.Lock                 // 并发锁
}
```

---

## 🎯 核心功能

### 1. 采购类型

系统支持两种采购类型：

| 类型常量 | 说明 | 特点 |
|---------|------|-----|
| `PurchaseType = 1` | 外部采购 | 向外部供应商采购，需要选择供应商和仓库 |
| `PurchaseType = 2` | 内部采购 | 向总部采购，需要选择仓库，需要总部审核 |

### 2. 采购单状态流转

```
待提交 (Draft)
    ↓ [提交]
待审核 (Pending)
    ↓ [审核通过] / [驳回]
已通过 (Approved) / 已驳回 (Rejected)
    ↓ [创建收货单]
部分收货 (Partial Receipt)
    ↓ [继续收货]
全部收货 (Completed)
```

**状态常量**:
- `PurchaseOrderStatusDraft = 0` - 待提交
- `PurchaseOrderStatusPending = 1` - 待审核
- `PurchaseOrderStatusApproved = 2` - 已通过（待收货）
- `PurchaseOrderStatusRejected = 3` - 已驳回
- `PurchaseOrderStatusCompleted = 4` - 全部收货（已完成）
- `PurchaseOrderStatusHeadquarterPending = 5` - 待总部审核（内部采购）

**总部状态常量**:
- `HeadquarterStatusDraft = 0` - 待提交
- `HeadquarterStatusPending = 1` - 待审核
- `HeadquarterStatusApproved = 2` - 已通过
- `HeadquarterStatusRejected = 3` - 已驳回
- `HeadquarterStatusCompleted = 4` - 已完成

### 3. 创建采购单 (CreatePurchaseOrder)

**功能描述**: 创建新的采购申请，支持外部采购和内部采购两种类型。

#### 创建流程

```
1. 验证请求参数
   - 供应商名称验证
   - 物品明细验证
   - 版本检查（V2.6+需要供应商编码和仓库编码）
   ↓
2. 加锁防止并发创建（使用供应商编码+订单时间字符串锁）
   ↓
3. 获取默认仓库
   ↓
4. 生成采购单编号
   - 外部采购：CSSQ + 年月日 + 4位序号
   - 内部采购：TPHY + 年月日 + 4位序号
   ↓
5. 获取仓库名称（如提供仓库编码）
   ↓
6. 设置期望到货时间（默认为2035-12-31）
   ↓
7. 创建采购申请主表（事务）
   ↓
8. 构建并创建采购申请明细
   ↓
9. 记录操作日志
   ↓
10. 返回采购单UUID和编号
```

#### 编号生成规则

**外部采购**:
```go
// 格式：CSSQ + 8位日期(YYYYMMDD) + 4位序号
// 示例：CSSQ202501120001
orderNo := "CSSQ" + datePart + serialNo
```

**内部采购**:
```go
// 格式：TPHY + 8位日期(YYYYMMDD) + 4位序号
// 示例：TPHY202501120001
orderNo := "TPHY" + datePart + serialNo
```

**编号特点**:
- 按日期重置序号（每天从0001开始）
- 通过查询当天最新订单获取下一个序号
- 支持时区设置

#### 响应数据

```go
type PurchaseOrderCreateResp struct {
    Uuid    uint64 // 采购单UUID
    OrderNo string // 采购单编号
}
```

---

### 4. 更新采购单 (UpdatePurchaseOrder)

**功能描述**: 更新采购申请信息，仅待提交状态可更新。

#### 更新限制

- 只有 `PurchaseOrderStatusDraft` 状态的采购单才能更新
- 更新时会重新创建明细（先删除旧明细，再创建新明细）

#### 更新流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 查询采购单并验证状态
   ↓
4. 版本检查（V2.6+需要供应商编码和仓库编码）
   ↓
5. 获取仓库名称（如提供仓库编码）
   ↓
6. 更新主表信息（事务）
   ↓
7. 删除所有现有明细
   ↓
8. 构建并创建新明细
   ↓
9. 记录操作日志
```

---

### 5. 删除采购单 (DeletePurchaseOrder)

**功能描述**: 删除采购申请，仅待提交和已驳回状态可删除。

#### 删除限制

- 只有 `PurchaseOrderStatusDraft` 或 `PurchaseOrderStatusRejected` 状态的采购单才能删除
- 执行软删除（设置 `delete_time`）

---

### 6. 提交采购单 (SubmitPurchaseOrder)

**功能描述**: 提交采购申请进入审核流程，触发供应商和物品状态验证。

#### 提交流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 查询采购单并验证状态
   ↓
4. 验证供应商状态
   ↓
5. 验证物品状态（如未确认）
   - 如果物品已禁用，提示："物品 %s 的状态已关闭。\n\n提交后将移除该物品，是否继续提交？"
   ↓
6. 删除数量为0的物品
   ↓
7. 如果用户确认提交，删除禁用的物品
   ↓
8. 重新查询采购单获取最新物品列表
   ↓
9. 验证物品数量不能为0
   ↓
10. 更新状态为待审核（事务）
    - 更新状态和单据日期
    - 更新申请人信息
    - 更新物品数量
    ↓
11. 记录操作日志
```

#### 供应商验证

提交时会检查供应商状态：
- 验证供应商是否存在
- 验证供应商是否启用

#### 物品状态验证

提交时会检查物品状态：
- 如果物品已禁用，提示："物品 %s 的状态已关闭。\n\n提交后将移除该物品，是否继续提交？"
- 用户确认后，会自动移除这些物品

---

### 7. 审核采购单 (ApprovePurchaseOrder)

**功能描述**: 审核采购申请，支持通过和驳回两种操作，内部采购需要总部审核。

#### 审核流程

```
1. 验证请求参数
   ↓
2. 加锁（UUID锁）
   ↓
3. 查询采购单并验证状态
   ↓
4. 验证审核权限
   ↓
5. 审核通过时：
   - 验证供应商状态
   - 验证物品状态
   ↓
6. 更新采购单状态（事务）
   - 通过：状态改为已通过，记录通过时间
   - 驳回：状态改为已驳回，记录驳回时间
   ↓
7. 内部采购特殊处理：
   - 子店审核通过：状态改为待总部审核，复制到总部
   - 总部审核通过：同步状态到子店
   - 总部驳回：同步驳回状态到子店
   ↓
8. 调用ERP接口（如开启ERP且审核通过）
   ↓
9. 记录操作日志
```

#### 内部采购审核流程

**子店审核通过**:
```
1. 子店审核通过
   ↓
2. 状态改为待总部审核
   ↓
3. 复制采购单到总部数据库
   - 复制主表信息
   - 复制明细信息
   - 验证总部是否存在对应物料
   ↓
4. 总部状态为待审核
```

**总部审核通过**:
```
1. 总部审核通过
   ↓
2. 调用ERP接口（内部采购）
   - 减总部库存
   - 创建物料请求单
   ↓
3. 同步状态到子店
   - 更新子店采购单状态为已通过
   - 更新ERP订单号
```

**总部驳回**:
```
1. 总部驳回
   ↓
2. 同步驳回状态到子店
   - 更新子店采购单状态为已驳回
   - 更新驳回时间
```

#### 外部采购审核流程

**审核通过**:
```
1. 审核通过
   ↓
2. 调用ERP接口（外部采购）
   - 创建采购订单
   - 添加到在途仓库
   ↓
3. 更新ERP订单号
```

---

### 8. 创建收货单 (CreatePurchaseReceiptOrder)

**功能描述**: 基于已通过的采购单创建收货单，支持外部收货和内部收货。

#### 收货类型

| 类型常量 | 说明 | 特点 |
|---------|------|-----|
| `ReceiptType = 1` | 外部收货 | 从外部供应商收货 |
| `ReceiptType = 2` | 内部收货 | 从总部收货 |

#### 创建流程

```
1. 验证请求参数
   ↓
2. 加锁（采购单UUID锁）
   ↓
3. 查询采购单并验证状态（必须是已通过状态）
   ↓
4. 验证收货明细
   - 验证收货数量不能超过采购数量
   - 验证物品状态
   ↓
5. 生成收货单编号（SHRK + 年月日 + 4位序号）
   ↓
6. 创建收货单主表和明细（事务）
   ↓
7. 更新采购单到货数量
   ↓
8. 更新采购单状态（部分收货/全部收货）
   ↓
9. 调用ERP接口（如开启ERP）
   ↓
10. 记录操作日志
```

#### 收货单编号生成

```go
// 格式：SHRK + 8位日期(YYYYMMDD) + 4位序号
// 示例：SHRK202501120001
receiptNo := "SHRK" + datePart + serialNo
```

---

### 9. 更新收货单 (UpdatePurchaseReceiptOrder)

**功能描述**: 更新收货单信息，仅待收货状态可更新。

#### 更新限制

- 只有 `ReceiptOrderStatusPending` 状态的收货单才能更新
- 更新时会重新创建明细

---

### 10. 取消收货单 (CancelPurchaseReceiptOrder)

**功能描述**: 取消收货单，仅待收货状态可取消。

#### 取消限制

- 只有 `ReceiptOrderStatusPending` 状态的收货单才能取消
- 取消后会回退采购单的到货数量

---

### 11. 获取采购单列表 (GetPurchaseOrderList)

**功能描述**: 分页获取采购单列表，支持多条件筛选。

#### 筛选条件

| 参数 | 说明 | 类型 |
|-----|------|-----|
| `order_no` | 订单编号或ERP订单号搜索 | string |
| `purchase_type` | 采购类型：1-外部采购 2-内部采购 | int |
| `status_in` | 状态筛选 | []int |
| `supplier_name` | 供应商名称 | string |
| `warehouse_erp_code` | 仓库编码 | string |
| `order_time_start` | 订单时间开始 | int |
| `order_time_end` | 订单时间结束 | int |
| `expect_arrival_time_start` | 期望到货时间开始 | int |
| `expect_arrival_time_end` | 期望到货时间结束 | int |

#### 状态筛选

| 状态 | 说明 |
|-----|------|
| 待提交 | 新增了保存的单据 |
| 待审核 | 提交了的单据 |
| 已通过 | 审核通过，待收货 |
| 已驳回 | 被驳回的单据 |
| 全部收货 | 收货完成 |
| 待总部审核 | 内部采购，子店审核通过，待总部审核 |

#### 排序规则

- 默认按订单时间倒序排序（`order_time DESC`）

---

### 12. 获取采购单详情 (GetPurchaseOrderDetail)

**功能描述**: 根据UUID获取采购单详细信息，包括明细、仓库、供应商等。

#### 返回数据

```go
type PurchaseOrderDetailResp struct {
    PurchaseOrderInfo
    Items []PurchaseOrderItemInfo // 采购明细
}
```

#### 特殊处理

- 显示收货进度（百分比）
- 显示物品的可用单位列表
- 显示物品的内部编码和条形码
- 显示每个单位的采购数量和到货数量

---

### 13. 获取收货单列表 (GetPurchaseReceiptOrderList)

**功能描述**: 分页获取收货单列表，支持联合查询采购单和收货单。

#### 查询特点

- V2.7.0+版本支持联合查询采购单和收货单
- 可以同时显示已通过的采购单和已创建的收货单
- 支持按收货时间筛选

---

## 🔄 内部采购流程

### 子店发起内部采购

```
1. 子店创建采购单（采购类型=2）
   ↓
2. 子店提交采购单
   ↓
3. 子店审核通过
   ↓
4. 状态改为待总部审核
   ↓
5. 复制采购单到总部数据库
   ↓
6. 总部审核
   - 通过：调用ERP接口，同步状态到子店
   - 驳回：同步驳回状态到子店
```

### 总部审核内部采购

**审核通过**:
```
1. 总部审核通过
   ↓
2. 减总部库存并记录出入库日志
   ↓
3. 调用ERP接口创建物料请求单
   ↓
4. 同步状态到子店
   - 更新子店采购单状态为已通过
   - 更新ERP订单号
```

**审核驳回**:
```
1. 总部驳回
   ↓
2. 同步驳回状态到子店
   - 更新子店采购单状态为已驳回
   - 更新驳回时间
```

---

## 📊 库存流转机制

### 外部采购库存流转

```
1. 审核通过：
   - 库存进入在途仓库
   ↓
2. 创建收货单：
   - 不扣减库存
   ↓
3. 确认收货：
   - 从在途仓库转入目标仓库
   - 调用ERP接口创建采购收货单
```

### 内部采购库存流转

```
1. 总部审核通过：
   - 从总部仓库扣减库存
   - 记录出入库日志
   ↓
2. 创建收货单：
   - 不扣减库存
   ↓
3. 确认收货：
   - 库存进入子店目标仓库
   - 调用ERP接口创建采购收货单
```

---

## 🔐 权限控制

### 可见性控制

**子店**:
- 可见自己创建的采购单
- 可见总部审核后的内部采购单

**总部**:
- 可见所有子店的内部采购单
- 可以审核内部采购单

### 操作权限

**创建/更新/删除**:
- 只有待提交状态的采购单才能操作
- 只有创建人所在门店可以操作

**审核**:
- 子店可以审核自己的采购单
- 总部可以审核内部采购单
- 验证 `status` 是否为待审核状态

**收货**:
- 只有已通过状态的采购单才能创建收货单
- 验证收货数量不能超过采购数量

---

## 🚨 错误处理

### 错误码定义

| 错误码 | 说明 | 使用场景 |
|-------|------|---------|
| `CodeMaterialDisabled` | 物品已禁用 | 提交或审核时物品状态已关闭 |

### 错误处理示例

```go
// 物品禁用错误
return errors.NewWithCodeAndData(
    constant.CodeMaterialDisabled,
    disabledMaterials,
    fmt.Sprintf("物品 %s 的状态已关闭。\n\n提交后将移除该物品，是否继续提交？", names),
)
```

---

## 🔧 配置项

### 版本要求

**V2.6.0+**:
- 必须提供供应商编码（`supplier_erp_code`）
- 内部采购必须提供仓库编码（`warehouse_erp_code`）

**V2.7.0+**:
- 收货单列表支持联合查询采购单和收货单

---

## 📝 数据模型

### 主表 (ttpos_purchase_order)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `order_no` | varchar(255) | 采购单编号 |
| `erp_order_no` | varchar(255) | ERP采购单号 |
| `sub_uuid` | bigint | 子订单UUID（内部采购） |
| `supplier_name` | varchar(100) | 供应商名称 |
| `supplier_erp_code` | varchar(255) | 供应商编码 |
| `purchase_type` | int | 采购类型：1-外部采购 2-内部采购 |
| `warehouse_erp_code` | varchar(255) | 仓库编码 |
| `status` | int | 状态：0-待提交 1-待审核 2-已通过 3-已驳回 4-全部收货 5-待总部审核 |
| `headquarter_status` | int | 总部状态 |
| `order_time` | int | 单据日期（提交时间戳） |
| `expect_arrival_time` | int | 期望到货日期（时间戳） |
| `num` | decimal(14,4) | 物品种类数量 |

### 明细表 (ttpos_purchase_order_item)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `purchase_order_uuid` | bigint | 采购单UUID |
| `material_uuid` | bigint | 物品UUID |
| `material_code` | varchar(255) | 物品编码 |
| `material_name` | text | 物品名称JSON |
| `num` | decimal(14,4) | 申请数量 |
| `arrival_num` | decimal(14,4) | 到货数量 |
| `unit_uuid` | bigint | 单位UUID |
| `unit_conversion_rate` | decimal(12,4) | 单位转换率 |
| `base_unit_uuid` | bigint | 基准单位UUID |

### 明细单位表 (ttpos_purchase_order_item_unit)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `item_uuid` | bigint | 采购单明细UUID |
| `purchase_order_uuid` | bigint | 采购单UUID |
| `unit_uuid` | bigint | 单位UUID |
| `num` | decimal(22,4) | 数量 |
| `arrival_num` | decimal(22,4) | 到货数量 |
| `unit_conversion_rate` | decimal(12,4) | 单位转换率 |

### 收货单表 (ttpos_purchase_receipt_order)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `order_no` | varchar(255) | 收货单编号 |
| `erp_order_no` | varchar(255) | ERP收货单号 |
| `purchase_order_uuid` | bigint | 采购单UUID |
| `purchase_order_no` | varchar(255) | 采购单号 |
| `status` | int | 状态：0-待收货 1-已收货 2-已取消 |
| `receipt_type` | int | 收货类型：1-外部收货 2-内部收货 |
| `receive_time` | int | 收货时间（时间戳） |
| `expect_arrival_time` | int | 期望到货日期（时间戳） |

### 收货单明细表 (ttpos_purchase_receipt_order_item)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `receipt_order_uuid` | bigint | 收货单UUID |
| `purchase_order_item_uuid` | bigint | 采购单明细UUID |
| `material_uuid` | bigint | 物品UUID |
| `num` | decimal(14,4) | 收货数量 |
| `unit_uuid` | bigint | 单位UUID |
| `unit_conversion_rate` | decimal(12,4) | 单位转换率 |

### 操作日志表 (ttpos_purchase_order_log)

| 字段 | 类型 | 说明 |
|-----|------|-----|
| `uuid` | bigint | 主键UUID |
| `purchase_order_uuid` | bigint | 采购单UUID |
| `operator_uuid` | bigint | 操作人UUID |
| `action` | varchar(50) | 操作动作 |
| `action_desc` | varchar(255) | 操作描述 |
| `old_status` | int | 操作前状态 |
| `new_status` | int | 操作后状态 |

---

## 🔄 依赖关系

```
purchaseOrderSrv
  ├── dbm (database.DBManager) - 数据库管理器
  ├── lock (lock.Lock) - 并发锁
  ├── validator (purchaseOrderValidator) - 验证器
  │   ├── validateMaterialStatus - 验证物品状态
  │   ├── validateSupplierStatus - 验证供应商状态
  │   ├── validateReceiptMaterialStatus - 验证收货单物品状态
  │   └── buildPurchaseOrderItems - 构建采购单明细
  ├── helper (purchaseOrderHelper) - 辅助方法
  │   ├── generateOrderNo - 生成采购单编号
  │   ├── generateReceiptNo - 生成收货单编号
  │   ├── createPurchaseOrderLog - 创建操作日志
  │   ├── reduceHeadquarterStockAndLog - 减总部库存并记录日志
  │   ├── AddToTransitWarehouse - 添加到在途仓库
  │   └── handleErpError - 处理ERP错误
  └── receiptSrv (purchaseReceiptOrderSrv) - 收货单服务
      ├── CreatePurchaseReceiptOrder - 创建收货单
      ├── UpdatePurchaseReceiptOrder - 更新收货单
      ├── CancelPurchaseReceiptOrder - 取消收货单
      └── GetPurchaseReceiptOrderList - 获取收货单列表
```

---

## 🌐 ERP集成

### 外部采购ERP接口

**创建采购订单 (CreatePurchaseOrder)**:
- 调用时机：审核通过时
- 功能：在ERP中创建采购订单
- 库存处理：添加到在途仓库

**创建采购收货单 (SavePurchaseReceipt)**:
- 调用时机：确认收货时
- 功能：在ERP中创建采购收货单
- 库存处理：从在途仓库转入目标仓库

### 内部采购ERP接口

**创建物料请求单 (SaveMaterialRequest)**:
- 调用时机：总部审核通过时
- 功能：在ERP中创建物料请求单
- 库存处理：从总部仓库扣减库存，记录出入库日志

**创建采购收货单 (SavePurchaseReceipt)**:
- 调用时机：确认收货时
- 功能：在ERP中创建采购收货单
- 库存处理：库存进入子店目标仓库

### ERP订单号

- `erp_order_no` - ERP采购单号（采购订单名称）
- 收货单的 `erp_order_no` - ERP收货单号（采购收货单名称）

---

## 🧪 测试建议

### 单元测试覆盖

1. **创建采购单测试**
   - 正常创建流程
   - 参数验证
   - 编号生成唯一性
   - 并发创建测试

2. **提交采购单测试**
   - 供应商验证
   - 物品状态验证
   - 数量为0的物品处理

3. **审核流程测试**
   - 外部采购审核
   - 内部采购审核（子店和总部）
   - 审核权限验证
   - 状态同步验证

4. **收货流程测试**
   - 创建收货单
   - 收货数量验证
   - 库存流转验证
   - ERP接口调用

5. **边界条件测试**
   - 并发操作
   - 物品状态变更处理
   - 供应商状态变更处理
   - ERP接口异常处理

---

## 📚 相关文档

- [物品管理](../material/material_service.md)
- [仓库管理](warehouse.md)
- [供应商管理](supplier.md)
- [ERP集成](../erp/erp_integration.md)
- [调拨单管理](./transfer_order.md)

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-01-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。

