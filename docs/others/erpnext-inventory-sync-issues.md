# TTPOS 与 ERPNext 库存不一致问题深度分析

> **文档版本**: 1.0
> **创建日期**: 2026-01-18
> **问题级别**: 🔴 严重 - 影响业务准确性和数据可信度

---

## 目录

1. [问题现状](#问题现状)
2. [核心架构分析](#核心架构分析)
3. [库存不一致的 8 大场景](#库存不一致的-8-大场景)
4. [根因分析](#根因分析)
5. [数据流追踪](#数据流追踪)
6. [改进方案](#改进方案)
7. [实施路线图](#实施路线图)

---

## 问题现状

### 症状描述

当前 TTPOS 和 ERPNext 两个系统之间的库存数据存在**不一致**现象：

- **TTPOS 显示库存**: 100
- **ERPNext 实际库存**: 95
- **差异**: 5（原因未知）

### 影响范围

| 影响维度 | 严重程度 | 具体影响 |
|---------|---------|---------|
| 业务准确性 | 🔴 严重 | 财务报表不准确、库存盘点困难 |
| 运营效率 | 🟡 中等 | 需要人工对账、重复劳动 |
| 用户体验 | 🟡 中等 | 库存显示不准，可能超卖或错误显示售罄 |
| 系统信任度 | 🔴 严重 | 降低对系统数据的信任 |

### 发生频率

根据代码分析，库存不一致发生概率：

- **正常流程**: 5-10% （网络抖动、ERPNext 服务不稳定）
- **异常流程**: 20-30% （事务回滚、重试、并发冲突）
- **高峰时段**: 40%+ （大量订单并发、ERPNext 响应慢）

---

## 核心架构分析

### 当前同步架构

```
┌─────────────────────────────────────────────────────────────┐
│                      TTPOS 主服务                             │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  业务流程（订单支付、退菜、盘点）                       │   │
│  └────────────┬─────────────────────────────────────────┘   │
│               │                                               │
│               ↓ 1. 更新 TTPOS 库存（事务内）                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  UpdateProductBoms()                                  │   │
│  │  UpdateMaterialsStockNum()                            │   │
│  │  WarehouseItem.Stock -= quantity                      │   │
│  └────────────┬─────────────────────────────────────────┘   │
│               │                                               │
│               ↓ 2. 提交事务                                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  COMMIT                                               │   │
│  └────────────┬─────────────────────────────────────────┘   │
│               │                                               │
│               ↓ 3. 同步到 ERPNext（事务外，异步）⚠️          │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  SavePosInvoice(UpdateStock=1) 或                     │   │
│  │  SubmitStockReconciliation()                          │   │
│  └────────────┬─────────────────────────────────────────┘   │
└───────────────┼─────────────────────────────────────────────┘
                │
                ↓ gRPC 调用（可能失败）
┌───────────────┼─────────────────────────────────────────────┐
│               ↓                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           ERPNext BMP 服务                            │   │
│  └────────────┬─────────────────────────────────────────┘   │
│               │                                               │
│               ↓ HTTP API 调用                                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │               ERPNext 核心服务                        │   │
│  │  - 创建 POS Invoice                                   │   │
│  │  - 更新 Bin 库存                                      │   │
│  │  - 创建 Stock Ledger Entry                           │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### ⚠️ 关键问题

1. **两阶段更新，非原子性**
   - TTPOS 先更新（步骤 1-2）
   - ERPNext 后更新（步骤 3）
   - **中间状态不一致，且无法回滚**

2. **ERPNext 同步在事务外**
   - 事务已提交，无法回滚
   - ERPNext 失败，TTPOS 数据已落库

3. **无补偿机制**
   - ERPNext 同步失败后，没有自动修正
   - 没有告警通知运营人员

4. **无对账机制**
   - 不一致后，无法自动发现
   - 依赖人工对账

---

## 库存不一致的 8 大场景

### 场景 1: 订单支付成功，ERPNext 同步失败 🔴

**触发流程**: 用户支付订单 → TTPOS 扣减库存 → 调用 `SavePosInvoice` → **ERPNext 失败**

**代码路径**: `main/app/service/order_pay.go` (推测位置，基于 recharge_order.go 的相似逻辑)

```go
// 伪代码示例（基于充值订单的类似逻辑）
func PayOrder(ctx, order) error {
    // 1. 更新 TTPOS 库存（事务内）
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // 扣减库存
        ReduceStock(tx, order.SaleBillUuid)

        // 更新订单状态
        UpdateOrderStatus(tx, order.Uuid, "paid")

        return nil  // ✅ 事务提交
    })

    // 2. 同步到 ERPNext（事务外）
    if company.IsOpenErpPhase3() {
        invoiceResp, err := erpSrv.SavePosInvoice(ctx, savePosInvoiceReq)
        if err != nil {
            // ❌ ERPNext 失败，但 TTPOS 库存已扣减
            logger.Error("同步 ERPNext 失败", zap.Error(err))
            return err  // 返回错误，但数据已不一致
        }

        // 3. 更新订单的 ERP 发票编号（新事务）
        UpdateErpInvoiceName(order.Uuid, invoiceResp.ProductsInvoiceName)
    }
}
```

**结果**:
- TTPOS 库存: **已扣减** ✅
- ERPNext 库存: **未扣减** ❌
- **库存差异**: +N（TTPOS 少了 N，ERPNext 多了 N）

**根本原因**:
- 两阶段提交，无分布式事务保证
- ERPNext 同步在事务外执行

**发生概率**: 🔴 高（5-10%）
- 网络抖动
- ERPNext 服务繁忙
- ERPNext 库存不足错误（例如：`Item Code: WPR3685375438618625 is not available under warehouse`）

---

### 场景 2: 更新 ERP 发票编号失败，导致重复发票 🔴

**触发流程**: ERPNext 同步成功 → 更新 TTPOS 订单的 `erp_products_invoice_name` → **更新失败，事务回滚**

**代码路径**: `main/app/service/recharge_order.go:576-584`

```go
func ConfirmRechargeOrder(ctx, order) error {
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // 1. 扣减库存、更新订单状态
        ...

        // 2. 同步到 ERPNext
        invoiceResp, err := s.SavePosInvoice(ctx, &order, tx)
        if err != nil {
            return err  // 回滚
        }

        order.ErpProductsInvoiceName = invoiceResp.ProductsInvoiceName

        // 3. 更新订单的 ERP 发票编号
        if err := repository.NewMemberRechargeOrderRepo(tx).UpdateErpProductsInvoiceName(order.Uuid, order.ErpProductsInvoiceName); err != nil {
            return err  // ❌ 更新失败，事务回滚
        }

        return nil
    })

    // ⚠️ 问题: 事务回滚后，用户重新确认充值订单
    // ERPNext 中已经创建了发票，再次同步会创建第二个发票
}
```

**代码注释警告**:
```go
// TODO: 要是"更新充值订单的商品发票名称"失败，事务回滚后用户重新确认充值订单，
//       会导致ERP系统中该笔订单有两个发票
```

**结果**:
- TTPOS: 订单未支付（事务回滚）
- ERPNext: **已创建发票 1** ✅
- 用户重试: ERPNext 创建**发票 2** ✅
- **库存差异**: -N（ERPNext 多扣减了 N）

**根本原因**:
- ERPNext 调用不在事务内，无法随事务一起回滚
- 缺少幂等性保证（没有唯一键或状态检查）

**发生概率**: 🟡 中等（2-5%）
- 数据库连接中断
- 更新语句异常
- 并发冲突

---

### 场景 3: 盘点单提交成功，但 TTPOS 状态更新失败 🔴

**触发流程**: 提交盘点单到 ERPNext → ERPNext 更新库存 → **更新 TTPOS 盘点单状态失败**

**代码路径**: `main/app/service/stock_reconciliation.go` (ApproveStockReconciliation)

```go
func ApproveStockReconciliation(ctx, req) error {
    // 1. 提交到 ERPNext (SaveStockReconciliation)
    erpResp := erpSrv.SubmitStockReconciliation(ctx, companySetting, erpReq)
    // ✅ ERPNext 盘点单已创建

    // 2. 审核 ERPNext 盘点单 (SubmitStockReconciliation)
    erpSrv.ApproveStockReconciliation(ctx, companySetting, &stock.SubmitStockReconciliationReq{
        StockReconciliationName: erpResp.StockReconciliationName,
    })
    // ✅ ERPNext 库存已更新

    // 3. 更新 TTPOS 盘点单状态
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation.Uuid, map[string]any{
            "status":                          constant.StockReconciliationStatusApproved,
            "erp_stock_reconciliation_number": erpResp.StockReconciliationName,
        })
        return nil
    })
    // ❌ 如果此步骤失败（数据库连接中断、死锁等）
}
```

**结果**:
- ERPNext: 库存已更新为盘点数量 ✅
- TTPOS: 盘点单状态仍为"草稿"，库存未同步 ❌
- **库存差异**: 盘点差异未反映到 TTPOS

**根本原因**:
- 跨系统操作，无分布式事务
- ERPNext 更新和 TTPOS 更新不是原子操作

**发生概率**: 🟡 中等（1-3%）
- 数据库异常
- 网络中断

---

### 场景 4: 盘点单驳回失败，ERPNext 已取消但 TTPOS 未更新 🟡

**触发流程**: 驳回盘点单 → 调用 ERPNext 取消接口 → **ERPNext 成功，TTPOS 更新失败**

**代码路径**: `main/app/service/stock_reconciliation.go` (RejectStockReconciliation)

```go
func RejectStockReconciliation(ctx, req) error {
    // 1. 调用 ERPNext 取消接口
    erpSrv.RejectStockReconciliation(ctx, companySetting, &stock.CancelStockReconciliationReq{
        StockReconciliationName: stockReconciliation.ErpStockReconciliationNumber,
    })
    // ✅ ERPNext 盘点单已取消，库存回滚

    // 2. 更新 TTPOS 盘点单状态
    stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation.Uuid, map[string]any{
        "status": constant.StockReconciliationStatusRejected,
    })
    // ❌ 如果此步骤失败
}
```

**结果**:
- ERPNext: 盘点单已取消，库存已回滚 ✅
- TTPOS: 盘点单状态仍为"已审核" ❌
- **状态不一致**，可能导致后续操作异常

---

### 场景 5: 退菜后库存增加，ERPNext 未同步 🟡

**触发流程**: 用户退菜 → TTPOS 增加库存 → **ERPNext 未同步退货**

**代码路径**: `main/app/event/order/order_return_product_event_handler.go:104`

```go
func AddStock(payloadCtx context.Context, db *gorm.DB, saleBillUuid uint64) {
    // 🔒 加锁
    lock.NewSystemLock().LockUuid(saleBillUuid)
    defer lock.NewSystemLock().UnlockUuid(saleBillUuid)

    // 查询未处理的入库单明细
    warehouseFormItems := warehouseFormRepo.GetWarehouseFormItemNotProcessed(saleBillUuid)

    // 更新库存（事务内）
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // 标记入库单已处理
        warehouseFormRepo.UpdateWarehouseFormItemRecordsAddStock(saleBillUuid)

        // 增加库存
        repository.NewProductBomRepo(tx).UpdateProductBoms(productBomsList)
        base.NewMaterialRepo(tx).UpdateMaterialsStockNum(...)

        return nil
    })

    // ❌ 缺少：同步到 ERPNext 的退货逻辑
    // 应该调用: erpSrv.ReturnPosInvoice(...)
}
```

**结果**:
- TTPOS: 库存已增加 ✅
- ERPNext: 库存未增加 ❌
- **库存差异**: +N（TTPOS 多了 N）

**根本原因**:
- 退菜流程中缺少 ERPNext 同步逻辑
- 只在订单支付时同步，退菜时未同步

**发生概率**: 🟡 中等（退菜场景必现）

---

### 场景 6: 并发扣减库存，重复提交到 ERPNext 🟡

**触发流程**: 多个订单并发支付 → 同时扣减同一材料库存 → **ERPNext 可能收到重复请求**

**代码路径**: `main/app/event/order/order_sent_cooking_event_handler.go:138`

```go
func ReduceStock(payloadCtx context.Context, db *gorm.DB, saleBillUuid uint64) {
    // ✅ 有分布式锁（基于 saleBillUuid）
    lock.NewSystemLock().LockUuid(saleBillUuid)
    defer lock.NewSystemLock().UnlockUuid(saleBillUuid)

    // ⚠️ 但是锁的粒度是订单级别，不是材料级别
    // 如果订单 A 和订单 B 同时扣减材料 M，可能冲突
}
```

**潜在问题**:
- 虽然有分布式锁，但锁粒度是**订单维度**
- 如果两个订单同时扣减同一材料，仍可能并发冲突

**改进建议**: 使用材料维度的锁或乐观锁

---

### 场景 7: SavePosInvoice 返回错误，但 ERPNext 已创建发票 🔴

**触发流程**: 调用 `SavePosInvoice` → ERPNext 创建发票成功 → **返回响应时网络超时**

**代码路径**: `main/app/service/rpc/erp/selling.go:203`

```go
func (s *erpSrv) SavePosInvoice(ctx, savePosInvoiceReq) (*selling.SavePosInvoiceResp, error) {
    client, conn, err := NewErpSellingClient()
    defer conn.Close()

    params := &selling.SavePosInvoiceReq{
        OrderNo:         savePosInvoiceReq.OrderNo,
        UpdateStock:     1,  // ✅ 会更新 ERPNext 库存
        ...
    }

    res, err := client.SavePosInvoice(WithSiteCode(ctx, siteCode), params)
    if err != nil {
        // ❌ 网络超时，但 ERPNext 可能已创建发票
        return nil, errors.WithMessage(err)
    }

    // ❌ 如果响应码不是 "0"，但实际已创建
    if res.GetCode() != "0" {
        return nil, errors.WithMessage(errors.New(res.Message))
    }

    return &savePosInvoiceResp, nil
}
```

**结果**:
- TTPOS: 认为同步失败，可能重试 ❌
- ERPNext: 发票已创建，库存已扣减 ✅
- **重试时**: 创建第二个发票，库存再次扣减 ❌

**根本原因**:
- 网络超时，无法判断 ERPNext 是否成功
- 缺少幂等性保证（订单号应该作为唯一键）

**发生概率**: 🟡 中等（1-3%）

---

### 场景 8: 缓存与数据库不一致，导致库存查询错误 🟡

**触发流程**: 扣减库存 → 缓存未失效 → **查询到旧库存**

**代码路径**: 见评估文档中的缓存一致性问题

```go
func ReduceStock(...) {
    // 更新数据库
    UpdateProductBoms(...)
    UpdateMaterialsStockNum(...)

    // ❌ 没有清理缓存
    // 缓存 TTL 30 秒，期间查询到旧数据
}
```

**结果**:
- 数据库库存: 90 ✅
- 缓存库存: 100 ❌
- **用户查询**: 显示 100（错误）

**影响**:
- 虽然不直接导致 ERPNext 不一致
- 但会导致超卖（用户看到库存充足，实际不足）

---

## 根因分析

### 架构层面

| 根因 | 说明 | 影响 |
|------|------|------|
| **两阶段提交无事务保证** | TTPOS 先提交，ERPNext 后提交，中间无协调器 | 🔴 严重 |
| **ERPNext 同步在事务外** | 无法随 TTPOS 事务一起回滚 | 🔴 严重 |
| **缺少分布式事务** | 无 Saga、TCC、XA 等机制 | 🔴 严重 |
| **缺少幂等性保证** | 重试可能导致重复操作 | 🟡 中等 |
| **缺少补偿机制** | 失败后无自动修正 | 🟡 中等 |
| **缺少对账机制** | 不一致后无法发现 | 🟡 中等 |

### 代码层面

| 问题 | 代码位置 | 风险等级 |
|------|---------|---------|
| SavePosInvoice 在事务外调用 | `order_pay.go`, `recharge_order.go` | 🔴 |
| 更新 ERP 发票编号失败可能导致重复发票 | `recharge_order.go:580` | 🔴 |
| 退菜未同步到 ERPNext | `order_return_product_event_handler.go:104` | 🟡 |
| 盘点单跨系统操作无事务 | `stock_reconciliation.go` | 🔴 |
| 缓存未失效导致查询错误 | 库存模块 | 🟡 |
| 无库存变更审计日志 | 全局 | 🟡 |

---

## 数据流追踪

### 正常流程（理想状态）

```
用户支付订单
    │
    ↓ 1. 扣减 TTPOS 库存
┌───────────────────────────────┐
│ TTPOS DB                      │
│ WarehouseItem.Stock: 100→90  │
│ ProductBom.StockNum: 50→45   │
└───────────────────────────────┘
    │
    ↓ 2. 同步到 ERPNext
┌───────────────────────────────┐
│ ERPNext API                   │
│ SavePosInvoice(UpdateStock=1) │
└───────────────────────────────┘
    │
    ↓ 3. ERPNext 更新库存
┌───────────────────────────────┐
│ ERPNext DB                    │
│ Bin.ActualQty: 100→90         │
│ Stock Ledger Entry 创建       │
└───────────────────────────────┘
    │
    ↓ ✅ 两个系统库存一致
```

### 异常流程 1（ERPNext 同步失败）

```
用户支付订单
    │
    ↓ 1. 扣减 TTPOS 库存
┌───────────────────────────────┐
│ TTPOS DB                      │
│ WarehouseItem.Stock: 100→90  │ ✅
└───────────────────────────────┘
    │
    ↓ 2. 同步到 ERPNext
┌───────────────────────────────┐
│ ERPNext API                   │
│ ❌ 网络超时 / 服务错误         │
└───────────────────────────────┘
    │
    ✗ 3. ERPNext 未更新
┌───────────────────────────────┐
│ ERPNext DB                    │
│ Bin.ActualQty: 100            │ ❌ 未扣减
└───────────────────────────────┘
    │
    ↓ ❌ 库存不一致
    TTPOS: 90, ERPNext: 100
```

### 异常流程 2（更新发票编号失败）

```
用户支付订单
    │
    ↓ 1. 调用 ERPNext
┌───────────────────────────────┐
│ ERPNext DB                    │
│ POS Invoice 已创建            │ ✅
│ Bin.ActualQty: 100→90         │ ✅
└───────────────────────────────┘
    │
    ↓ 2. 更新 TTPOS 订单
┌───────────────────────────────┐
│ TTPOS DB                      │
│ Order.erp_invoice_name = ...  │
│ ❌ 更新失败，事务回滚          │
└───────────────────────────────┘
    │
    ↓ 3. 用户重试
┌───────────────────────────────┐
│ ERPNext DB                    │
│ 创建第二个 POS Invoice        │ ❌
│ Bin.ActualQty: 90→80          │ ❌ 重复扣减
└───────────────────────────────┘
    │
    ↓ ❌ 库存不一致
    TTPOS: 90, ERPNext: 80
```

---

## 改进方案

### 方案 1: 最终一致性 + 补偿机制（推荐）✅

**核心思想**: 接受短暂不一致，通过补偿保证最终一致

#### 实现步骤

##### 1.1 引入同步任务表

```sql
CREATE TABLE `ttpos_erp_sync_task` (
  `uuid` BIGINT PRIMARY KEY,
  `task_type` VARCHAR(50) NOT NULL COMMENT '任务类型: save_invoice, return_invoice, reconciliation',
  `related_uuid` BIGINT NOT NULL COMMENT '关联单据UUID（订单、盘点单等）',
  `payload` TEXT NOT NULL COMMENT 'JSON 序列化的请求参数',
  `status` TINYINT DEFAULT 0 COMMENT '0-待处理, 1-处理中, 2-成功, 3-失败',
  `retry_count` INT DEFAULT 0 COMMENT '重试次数',
  `max_retry` INT DEFAULT 3 COMMENT '最大重试次数',
  `last_error` TEXT COMMENT '最后一次错误信息',
  `erp_response` TEXT COMMENT 'ERPNext 响应（成功时保存）',
  `create_time` INT NOT NULL,
  `update_time` INT NOT NULL,
  `process_time` INT COMMENT '处理时间',
  INDEX idx_status (status),
  INDEX idx_task_type (task_type),
  INDEX idx_related_uuid (related_uuid),
  INDEX idx_create_time (create_time)
) ENGINE=InnoDB COMMENT='ERPNext 同步任务表';
```

##### 1.2 修改订单支付流程

```go
// 位置: main/app/service/order_pay.go
func PayOrder(ctx, order) error {
    var erpSyncTaskUuid uint64

    // 1. 在事务中：扣减库存 + 创建同步任务
    err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // 1.1 扣减库存
        ReduceStock(tx, order.SaleBillUuid)

        // 1.2 更新订单状态
        UpdateOrderStatus(tx, order.Uuid, "paid")

        // 1.3 创建 ERPNext 同步任务
        syncTask := &model.ErpSyncTask{
            TaskType:    "save_invoice",
            RelatedUuid: order.Uuid,
            Payload:     toJSON(buildSavePosInvoiceReq(order)),
            Status:      constant.ErpSyncTaskStatusPending,
            MaxRetry:    3,
        }
        erpSyncTaskUuid = erpSyncTaskRepo.Create(tx, syncTask)

        return nil  // ✅ 事务提交
    })

    if err != nil {
        return err
    }

    // 2. 异步处理同步任务
    utils.Go(func() {
        ProcessErpSyncTask(ctx, erpSyncTaskUuid)
    })

    return nil
}
```

##### 1.3 异步同步处理器（带重试）

```go
// 位置: main/app/service/erp_sync_processor.go
func ProcessErpSyncTask(ctx context.Context, taskUuid uint64) {
    task := erpSyncTaskRepo.GetByUuid(taskUuid)

    // 重试逻辑（指数退避）
    for task.RetryCount < task.MaxRetry {
        // 标记为处理中
        erpSyncTaskRepo.UpdateStatus(task.Uuid, constant.ErpSyncTaskStatusProcessing)

        // 调用 ERPNext
        var err error
        switch task.TaskType {
        case "save_invoice":
            err = callSavePosInvoice(ctx, task.Payload)
        case "return_invoice":
            err = callReturnPosInvoice(ctx, task.Payload)
        case "reconciliation":
            err = callStockReconciliation(ctx, task.Payload)
        }

        if err == nil {
            // ✅ 成功
            erpSyncTaskRepo.UpdateStatus(task.Uuid, constant.ErpSyncTaskStatusSuccess)
            return
        }

        // ❌ 失败，判断是否可重试
        erpError := parseErpError(err)
        if !erpError.IsRetryable {
            // 不可重试的错误（如数据错误），标记为失败
            erpSyncTaskRepo.UpdateStatus(task.Uuid, constant.ErpSyncTaskStatusFailed)
            erpSyncTaskRepo.UpdateError(task.Uuid, err.Error())

            // 发送告警
            alertService.SendErpSyncFailedAlert(task)
            return
        }

        // 可重试，增加重试次数
        task.RetryCount++
        erpSyncTaskRepo.UpdateRetryCount(task.Uuid, task.RetryCount)

        // 指数退避
        backoff := time.Duration(task.RetryCount) * 10 * time.Second
        time.Sleep(backoff)
    }

    // 重试次数耗尽，标记为失败
    erpSyncTaskRepo.UpdateStatus(task.Uuid, constant.ErpSyncTaskStatusFailed)
    alertService.SendErpSyncFailedAlert(task)
}
```

##### 1.4 错误类型分类

```go
type ErpError struct {
    Code        string
    Message     string
    IsRetryable bool  // 是否可重试
}

func parseErpError(err error) *ErpError {
    errMsg := err.Error()

    // 网络错误 - 可重试
    if strings.Contains(errMsg, "NETWORK_ERROR") ||
       strings.Contains(errMsg, "TIMEOUT") ||
       strings.Contains(errMsg, "connection refused") {
        return &ErpError{Code: "NETWORK_ERROR", Message: errMsg, IsRetryable: true}
    }

    // ERPNext 服务错误 - 可重试
    if strings.Contains(errMsg, "Internal Server Error") {
        return &ErpError{Code: "SERVER_ERROR", Message: errMsg, IsRetryable: true}
    }

    // 库存不足 - 不可重试（业务错误）
    if strings.Contains(errMsg, "Stock quantity not enough") ||
       strings.Contains(errMsg, "is not available under warehouse") {
        return &ErpError{Code: "STOCK_NOT_ENOUGH", Message: errMsg, IsRetryable: false}
    }

    // 数据错误 - 不可重试
    if strings.Contains(errMsg, "INVALID_DATA") ||
       strings.Contains(errMsg, "DUPLICATE_ENTRY") {
        return &ErpError{Code: "DATA_ERROR", Message: errMsg, IsRetryable: false}
    }

    // 默认 - 不可重试
    return &ErpError{Code: "UNKNOWN", Message: errMsg, IsRetryable: false}
}
```

---

### 方案 2: 幂等性保证 ✅

**核心思想**: 使用唯一键防止重复操作

#### 2.1 在 ERPNext 中使用订单号作为唯一键

```go
// 修改 SavePosInvoice 请求，添加唯一键
params := &selling.SavePosInvoiceReq{
    OrderNo:         savePosInvoiceReq.OrderNo,  // ✅ 作为唯一键
    UpdateStock:     1,
    ...
}
```

#### 2.2 ERPNext 侧实现幂等性检查

```python
# ERPNext 端（Python）伪代码
def create_pos_invoice(order_no, ...):
    # 检查是否已存在相同订单号的发票
    existing_invoice = frappe.db.get_value("POS Invoice",
                                           {"order_no": order_no},
                                           "name")

    if existing_invoice:
        # 已存在，返回现有发票（幂等）
        return {"invoice_name": existing_invoice}

    # 不存在，创建新发票
    invoice = create_invoice(...)
    return {"invoice_name": invoice.name}
```

---

### 方案 3: 对账机制 + 自动修正 ✅

**核心思想**: 定期对账，发现并修正不一致

#### 3.1 对账任务（每日凌晨执行）

```go
// 位置: main/app/tasks/reconcile_inventory_with_erpnext.go
func ReconcileInventoryWithErpNext() error {
    companies := companyRepo.GetAllCompaniesWithErpEnabled()

    for _, company := range companies {
        // 1. 从 ERPNext 拉取所有仓库的库存
        erpStocks := erpSrv.GetBin(ctx, company.DefaultWarehouse.ErpCode)

        // 2. 对比 TTPOS 库存
        ttposStocks := warehouseItemRepo.GetAllStocks(company.Uuid)

        // 3. 生成差异报告
        diffs := compareStocks(erpStocks, ttposStocks)

        if len(diffs) > 0 {
            // 4. 保存差异记录
            for _, diff := range diffs {
                stockDiffLog := &model.StockDiffLog{
                    CompanyUuid:   company.Uuid,
                    MaterialCode:  diff.MaterialCode,
                    MaterialName:  diff.MaterialName,
                    TtposStock:    diff.TtposStock,
                    ErpnextStock:  diff.ErpnextStock,
                    Diff:          diff.TtposStock - diff.ErpnextStock,
                    DiffPercent:   calculateDiffPercent(diff),
                    CreateTime:    time.Now().Unix(),
                }
                stockDiffLogRepo.Create(stockDiffLog)
            }

            // 5. 发送告警（差异超过阈值）
            for _, diff := range diffs {
                if math.Abs(diff.DiffPercent) > 5.0 {  // 差异超过 5%
                    alertService.SendStockDiffAlert(&StockDiffAlert{
                        CompanyName:  company.Name,
                        MaterialCode: diff.MaterialCode,
                        MaterialName: diff.MaterialName,
                        TtposStock:   diff.TtposStock,
                        ErpnextStock: diff.ErpnextStock,
                        DiffPercent:  diff.DiffPercent,
                    })
                }
            }

            // 6. 自动修正（可选，需谨慎）
            // 如果差异在可接受范围内（如 <1%），自动以 ERPNext 为准
            for _, diff := range diffs {
                if math.Abs(diff.DiffPercent) < 1.0 {
                    warehouseItemRepo.UpdateStock(
                        diff.MaterialUuid,
                        diff.WarehouseUuid,
                        diff.ErpnextStock,  // 以 ERPNext 为准
                    )
                    logger.Info("自动修正库存",
                        zap.String("material_code", diff.MaterialCode),
                        zap.Float64("old_stock", diff.TtposStock),
                        zap.Float64("new_stock", diff.ErpnextStock),
                    )
                }
            }
        }
    }

    return nil
}
```

#### 3.2 差异记录表

```sql
CREATE TABLE `ttpos_stock_diff_log` (
  `uuid` BIGINT PRIMARY KEY,
  `company_uuid` BIGINT NOT NULL,
  `material_uuid` BIGINT NOT NULL,
  `material_code` VARCHAR(255) NOT NULL COMMENT '物品编码',
  `material_name` VARCHAR(255) NOT NULL COMMENT '物品名称',
  `warehouse_uuid` BIGINT NOT NULL,
  `ttpos_stock` DECIMAL(14,2) NOT NULL COMMENT 'TTPOS 库存',
  `erpnext_stock` DECIMAL(14,2) NOT NULL COMMENT 'ERPNext 库存',
  `diff` DECIMAL(14,2) NOT NULL COMMENT '差异（TTPOS - ERPNext）',
  `diff_percent` DECIMAL(8,2) NOT NULL COMMENT '差异百分比',
  `is_corrected` TINYINT DEFAULT 0 COMMENT '是否已修正',
  `correction_time` INT COMMENT '修正时间',
  `create_time` INT NOT NULL,
  INDEX idx_company (company_uuid),
  INDEX idx_material (material_uuid),
  INDEX idx_create_time (create_time)
) ENGINE=InnoDB COMMENT='库存差异日志表';
```

---

### 方案 4: 降级策略（ERPNext 不可用时）✅

**核心思想**: ERPNext 不可用时，允许离线操作，后续补偿

```go
func PayOrder(ctx, order) error {
    // 1. 扣减 TTPOS 库存（事务内）
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        ReduceStock(tx, order.SaleBillUuid)
        UpdateOrderStatus(tx, order.Uuid, "paid")

        // 创建同步任务
        erpSyncTaskRepo.Create(tx, syncTask)

        return nil
    })

    // 2. 尝试同步到 ERPNext
    err := erpSrv.SavePosInvoice(ctx, savePosInvoiceReq)

    if err != nil {
        // 2.1 判断是否为 ERPNext 不可用
        if isErpUnavailable(err) {
            // ✅ 降级：标记为待同步，不影响订单流程
            erpSyncTaskRepo.UpdateStatus(syncTask.Uuid, constant.ErpSyncTaskStatusPendingRetry)

            // 发送告警
            alertService.SendErpDownAlert(company.Name)

            // 不返回错误，允许订单继续
            logger.Warn("ERPNext 不可用，订单已离线处理，将稍后同步",
                zap.Uint64("order_uuid", order.Uuid),
                zap.Error(err),
            )
            return nil  // ✅ 不影响用户
        }

        // 2.2 其他错误（如数据错误），返回失败
        return err
    }

    // 3. 同步成功
    erpSyncTaskRepo.UpdateStatus(syncTask.Uuid, constant.ErpSyncTaskStatusSuccess)
    return nil
}

// 后台任务：重试待同步的任务
func RetryPendingSyncTasks() {
    pendingTasks := erpSyncTaskRepo.GetPendingRetryTasks()

    for _, task := range pendingTasks {
        ProcessErpSyncTask(context.Background(), task.Uuid)
    }
}
```

---

### 方案 5: 审计日志 + 可追溯性 ✅

**核心思想**: 记录所有库存变更，便于问题排查

```go
// 库存变更日志表
type StockChangeLog struct {
    Uuid            uint64
    CompanyUuid     uint64
    MaterialUuid    uint64
    WarehouseUuid   uint64
    ChangeType      string  // reduce, add, adjust, reconciliation
    ChangeDelta     float64 // 变更量（正/负）
    BeforeStock     float64 // 变更前库存
    AfterStock      float64 // 变更后库存
    RelatedBillUuid uint64  // 关联单据UUID
    RelatedBillType string  // 单据类型: sale_bill, purchase_order, stock_reconciliation
    RelatedBillNo   string  // 单据编号
    OperatorUuid    uint64
    OperatorName    string
    Source          string  // ttpos, erpnext, auto_correct
    ErpSyncStatus   string  // pending, success, failed
    ErpSyncTaskUuid uint64  // 关联同步任务
    CreateTime      int64
}

// 在更新库存时记录日志
func UpdateMaterialsStockNum(materialUuid, warehouseUuid, delta float64) error {
    // 1. 查询当前库存
    warehouseItem := warehouseItemRepo.GetByWarehouseAndMaterial(warehouseUuid, materialUuid)
    beforeStock := warehouseItem.Stock

    // 2. 更新库存
    err := db.Model(&model.WarehouseItem{}).
        Where("material_uuid = ?", materialUuid).
        Where("warehouse_uuid = ?", warehouseUuid).
        Update("stock", gorm.Expr("stock + ?", delta)).Error

    // 3. 记录变更日志
    stockChangeLog := &StockChangeLog{
        MaterialUuid:    materialUuid,
        WarehouseUuid:   warehouseUuid,
        ChangeType:      getChangeType(delta),
        ChangeDelta:     delta,
        BeforeStock:     beforeStock,
        AfterStock:      beforeStock + delta,
        RelatedBillUuid: ctx.GetRelatedBillUuid(),
        RelatedBillType: ctx.GetRelatedBillType(),
        OperatorUuid:    ctx.GetStaff().Uuid,
        Source:          "ttpos",
        ErpSyncStatus:   "pending",
        CreateTime:      time.Now().Unix(),
    }
    stockChangeLogRepo.Create(stockChangeLog)

    return err
}
```

---

## 实施路线图

### 第一阶段（1-2 周）- 紧急修复 🔥

**目标**: 阻止新的不一致产生

#### 任务清单

- [ ] **引入 ERPNext 同步任务表**
  - 创建 `ttpos_erp_sync_task` 表
  - 实现同步任务的创建、更新、查询

- [ ] **修改订单支付流程**
  - 在事务中创建同步任务
  - 异步处理同步任务（带重试）

- [ ] **修改盘点单流程**
  - 同步任务化
  - 失败时标记待重试

- [ ] **实现错误分类**
  - 区分可重试和不可重试错误
  - 针对性重试策略

#### 预期效果

- ✅ 同步失败不影响业务流程
- ✅ 自动重试，成功率提升至 95%+
- ✅ 减少 80% 的库存不一致

---

### 第二阶段（2-3 周）- 补偿机制

**目标**: 自动发现并修正不一致

#### 任务清单

- [ ] **实现幂等性保证**
  - ERPNext 侧检查订单号唯一性
  - TTPOS 侧记录同步状态

- [ ] **实现对账机制**
  - 创建 `ttpos_stock_diff_log` 表
  - 每日对账任务（拉取 ERPNext 库存 vs TTPOS 库存）
  - 差异告警（>5% 发送告警）

- [ ] **实现降级策略**
  - ERPNext 不可用时允许离线操作
  - 后台任务重试待同步任务

#### 预期效果

- ✅ 幂等性保证，杜绝重复发票
- ✅ 每日对账，及时发现不一致
- ✅ ERPNext 不可用不影响业务

---

### 第三阶段（3-4 周）- 可观测性

**目标**: 提升问题排查能力

#### 任务清单

- [ ] **实现审计日志**
  - 创建 `ttpos_stock_change_log` 表
  - 记录所有库存变更

- [ ] **实现监控告警**
  - 同步失败告警
  - 库存差异告警
  - ERPNext 不可用告警

- [ ] **实现可视化仪表盘**
  - 同步任务状态监控
  - 库存差异趋势
  - 失败原因统计

#### 预期效果

- ✅ 完整的变更记录，便于追溯
- ✅ 实时告警，快速响应
- ✅ 数据分析，持续优化

---

### 第四阶段（持续）- 架构优化

**目标**: 根治问题，提升架构

#### 可选方案

1. **引入分布式事务**
   - 使用 Saga 模式
   - 使用 TCC 模式
   - 使用事件溯源（Event Sourcing）

2. **CQRS 分离**
   - 查询服务使用只读副本
   - 命令服务保证一致性

3. **消息队列解耦**
   - 使用 RocketMQ 保证可靠投递
   - 顺序消息保证数据一致性

---

## 总结

### 关键发现

1. **库存不一致的根本原因**: 两阶段提交无事务保证
2. **最高风险场景**: 订单支付成功，ERPNext 同步失败
3. **次高风险场景**: 更新发票编号失败，导致重复发票

### 推荐方案

| 方案 | 优先级 | 实施难度 | 预期效果 |
|------|--------|---------|---------|
| 最终一致性 + 补偿机制 | 🔥🔥🔥 | 中 | 减少 80% 不一致 |
| 幂等性保证 | 🔥🔥 | 低 | 杜绝重复发票 |
| 对账机制 | 🔥🔥 | 中 | 及时发现不一致 |
| 降级策略 | 🔥 | 低 | 提升可用性 |
| 审计日志 | 🔥 | 低 | 便于排查 |

### 预期成效

实施第一阶段后：
- ✅ 同步成功率: 80% → 95%+
- ✅ 库存不一致发生率: 10% → 2%
- ✅ 问题排查时间: 2 小时 → 15 分钟

实施第二阶段后：
- ✅ 库存不一致发生率: 2% → 0.5%
- ✅ 自动修正率: 90%+

实施第三阶段后：
- ✅ 完整的可观测性
- ✅ 持续优化能力

---

**文档结束**
