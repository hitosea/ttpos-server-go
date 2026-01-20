# 外卖自动接单后 POS 发票生成失败处理方案

## 问题描述

外卖订单自动接单后，在结账时生成 POS 发票失败，导致整个结账流程失败。

**重要更新**：ERP 端发票生成是异步的，TTPOS 调用后默认都成功，但实际上可能失败。

## 问题分析

### 流程说明

1. **外卖订单支付完成** → 触发 `PayFinishMemberSaleOrderEvent`
2. **自动接单** → 调用 `AcceptMemberSaleOrder`（如果满足自动接单条件）
3. **结账时生成发票** → 调用 `SavePosInvoice` 生成 ERP 发票
4. **问题场景**：
   - **场景 1（同步模式）**：如果发票生成失败，会导致结账事务回滚，结账失败
   - **场景 2（异步模式）**：ERP 返回 `AsyncRecordId`，TTPOS 默认成功，但实际可能失败，TTPOS 无法感知
   - **场景 3（调用失败）**：TTPOS 调用 ERP 接口失败（网络超时、连接失败等），ERP 端没有收到请求，没有创建 Draft 记录，订单已结账但发票未生成 ⚠️

### 代码位置

**结账时发票生成**：
```go
// main/app/service/order_pay.go:863-876
if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
    res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
    if err != nil {
        return errors.WithMessage(err) // 这里会导致整个结账事务回滚
    }
    // ...
}
```

**发票生成失败原因**：
- ERP 系统连接失败
- ERP 系统数据校验失败（如库存不足、商品不存在等）
- 网络超时
- ERP 系统临时故障

### 场景 3：TTPOS 调用失败，ERP 端无数据

**问题描述**：
- TTPOS 调用 ERP 接口时失败（网络超时、连接失败、gRPC 错误等）
- ERP 端没有收到请求，所以没有创建 `ReceivePosInvoice` 记录（没有 Draft）
- TTPOS 端返回错误，但订单可能已结账（如果使用了降级处理）
- 订单没有 `AsyncRecordId`，ERP 端也没有对应记录

**典型错误**：
- `连接ERP系统失败`
- `网络超时`
- `gRPC连接失败`
- `context deadline exceeded`

**影响**：
- 订单已结账，但发票未生成
- 无法通过 `AsyncRecordId` 查询状态
- ERP 端没有记录，无法重试

## 解决方案

### 方案 1：使用异步模式 + 状态查询机制（推荐）

系统已支持异步发票生成模式，但需要完善状态查询和失败处理机制。

#### 当前问题

1. **ERP 异步返回**：ERP 端返回 `AsyncRecordId`，TTPOS 默认成功
2. **缺少状态跟踪**：TTPOS 没有保存 `AsyncRecordId`，无法查询发票生成状态
3. **无法感知失败**：ERP 异步处理失败时，TTPOS 不知道

#### 实现原理

1. **异步模式**：发票生成请求先保存到 `ReceivePosInvoice` 表，状态为 `Draft`
2. **消息队列**：通过消息队列异步处理发票生成
3. **状态更新**：处理完成后更新 `ReceivePosInvoice` 表的 `docstatus` 字段
   - `0` (Draft)：待处理
   - `1` (Submitted)：已成功
   - `2` (Cancelled)：已取消

#### 需要实现的改进

##### 1. 保存 AsyncRecordId 到订单表

**数据库迁移**：
```sql
ALTER TABLE `ttpos_sale_order` 
ADD COLUMN `erp_async_record_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERP异步记录ID' AFTER `erp_material_invoice_name`;
```

**代码修改**：
```go
// main/app/model/sale_order.go
type SaleOrder struct {
    // ... 其他字段
    ErpProductsInvoiceName string `gorm:"column:erp_products_invoice_name;type:varchar(255);comment:商品发票名称;NOT NULL" json:"erp_products_invoice_name"`
    ErpMaterialInvoiceName string `gorm:"column:erp_material_invoice_name;type:varchar(255);comment:原材料发票名称;NOT NULL" json:"erp_material_invoice_name"`
    ErpAsyncRecordId      string `gorm:"column:erp_async_record_id;type:varchar(255);comment:ERP异步记录ID;NOT NULL" json:"erp_async_record_id"` // 新增字段
}

// main/app/service/order_pay.go
if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
    res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
    if err != nil {
        return errors.WithMessage(err)
    }
    
    // 保存异步记录ID（如果返回了）
    if res.AsyncRecordId != "" {
        saleOrder.ErpAsyncRecordId = res.AsyncRecordId
    }
    
    // 如果发票已生成成功，保存发票名称
    if res.ProductsInvoiceName != "" {
        saleOrder.ErpProductsInvoiceName = res.ProductsInvoiceName
    }
    if res.MaterialInvoiceName != "" {
        saleOrder.ErpMaterialInvoiceName = res.MaterialInvoiceName
    }
    
    // 更新订单表
    updateFields := map[string]interface{}{
        "erp_async_record_id": saleOrder.ErpAsyncRecordId,
    }
    if saleOrder.ErpProductsInvoiceName != "" {
        updateFields["erp_products_invoice_name"] = saleOrder.ErpProductsInvoiceName
    }
    if saleOrder.ErpMaterialInvoiceName != "" {
        updateFields["erp_material_invoice_name"] = saleOrder.ErpMaterialInvoiceName
    }
    
    if err := repository.NewSaleOrderRepo(db).Update(saleOrder.Uuid, updateFields); err != nil {
        return errors.WithMessage(err)
    }
}
```

##### 2. 提供查询发票状态的接口

**在 ttpos-bmp 服务中添加查询接口**：
```go
// ttpos-bmp/app/ttpos-erp/internal/logic/selling/async_selling.go
func (s *sAsyncSelling) GetInvoiceStatus(ctx context.Context, asyncRecordId string) (*selling.InvoiceStatusResp, error) {
    recordId, err := gconv.Int64(asyncRecordId)
    if err != nil {
        return nil, gerror.Wrapf(err, "异步记录ID格式错误: %s", asyncRecordId)
    }
    
    record := &entity.ReceivePosInvoice{}
    err = dao.ReceivePosInvoice.Ctx(ctx).WherePri(recordId).Scan(record)
    if err != nil {
        return nil, gerror.Wrapf(err, "查询发票记录失败")
    }
    
    return &selling.InvoiceStatusResp{
        AsyncRecordId:        asyncRecordId,
        Docstatus:           record.Docstatus,
        ProductsInvoiceName: record.ProductsInvoiceName,
        MaterialInvoiceName: record.MaterialInvoiceName,
        RespBody:            record.RespBody,
        CreatedAt:           record.CreatedAt,
    }, nil
}
```

**在 main 服务中添加查询方法**：
```go
// main/app/service/rpc/erp/selling.go
func (s *erpSrv) GetInvoiceStatus(ctx pkgCtx.Context, asyncRecordId string) (*selling.InvoiceStatusResp, error) {
    client, conn, err := NewErpSellingClient()
    if err != nil {
        return nil, errors.WithMessage(err)
    }
    defer conn.Close()
    
    req := &selling.GetInvoiceStatusReq{
        AsyncRecordId: asyncRecordId,
    }
    res, err := client.GetInvoiceStatus(ctx.GetContext(), req)
    if err != nil {
        return nil, errors.WithMessage(err)
    }
    
    if res.GetCode() != "0" {
        return nil, errors.WithMessage(errors.New(res.Message))
    }
    
    if res.Data != nil {
        var statusResp selling.InvoiceStatusResp
        if err := res.Data.UnmarshalTo(&statusResp); err != nil {
            return nil, errors.WithMessage(err)
        }
        return &statusResp, nil
    }
    
    return nil, errors.WithMessage(errors.New("查询发票状态异常, data为空"))
}
```

##### 3. 定时任务查询未完成的发票

**创建定时任务**：
```go
// main/app/job/erp_invoice_status_check.go
package job

import (
    "time"
    "ttpos-server-go/app/repository"
    "ttpos-server-go/app/service/rpc/erp"
    "ttpos-server-go/pkg/database"
    "ttpos-server-go/pkg/logger"
    "go.uber.org/zap"
)

// CheckErpInvoiceStatus 检查ERP发票状态
func CheckErpInvoiceStatus() {
    dbm := database.GetDBManager(config.DatabaseConf{})
    
    // 查询所有有 AsyncRecordId 但发票名称为空的订单（可能是异步处理中或失败）
    db := dbm.GetDB(0) // 需要遍历所有公司
    orderRepo := repository.NewSaleOrderRepo(db)
    
    // 查询条件：有 AsyncRecordId，但发票名称为空，且订单已结账
    orders, err := orderRepo.GetOrdersWithAsyncRecordIdButNoInvoice()
    if err != nil {
        logger.Logger.Error("查询待检查发票状态的订单失败", zap.Error(err))
        return
    }
    
    erpSrv := erp.NewIErpSrv(dbm)
    for _, order := range orders {
        if order.ErpAsyncRecordId == "" {
            continue
        }
        
        // 查询发票状态
        status, err := erpSrv.GetInvoiceStatus(ctx, order.ErpAsyncRecordId)
        if err != nil {
            logger.Logger.Error("查询发票状态失败", 
                zap.String("order_no", order.OrderNo),
                zap.String("async_record_id", order.ErpAsyncRecordId),
                zap.Error(err))
            continue
        }
        
        // 如果发票已生成成功，更新订单表
        if status.Docstatus == 1 { // Submitted
            updateFields := map[string]interface{}{}
            if status.ProductsInvoiceName != "" {
                updateFields["erp_products_invoice_name"] = status.ProductsInvoiceName
            }
            if status.MaterialInvoiceName != "" {
                updateFields["erp_material_invoice_name"] = status.MaterialInvoiceName
            }
            
            if len(updateFields) > 0 {
                if err := orderRepo.Update(order.Uuid, updateFields); err != nil {
                    logger.Logger.Error("更新订单发票名称失败",
                        zap.String("order_no", order.OrderNo),
                        zap.Error(err))
                }
            }
        } else if status.Docstatus == 0 { // Draft - 仍然处理中
            // 记录日志，但不需要处理
            logger.Logger.Info("发票仍在处理中",
                zap.String("order_no", order.OrderNo),
                zap.String("async_record_id", order.ErpAsyncRecordId))
        } else {
            // 失败或其他状态，记录告警
            logger.Logger.Warn("发票生成失败",
                zap.String("order_no", order.OrderNo),
                zap.String("async_record_id", order.ErpAsyncRecordId),
                zap.Int("docstatus", status.Docstatus),
                zap.String("error", status.RespBody))
        }
    }
}
```

**Repository 方法**：
```go
// main/app/repository/order.go
func (r *orderRepo) GetOrdersWithAsyncRecordIdButNoInvoice() ([]*model.SaleOrder, error) {
    var orders []*model.SaleOrder
    err := r.db.Where("erp_async_record_id != ''").
        Where("(erp_products_invoice_name = '' OR erp_material_invoice_name = '')").
        Where("status = ?", constant.OrderStatusSettled).
        Where("created_at > ?", time.Now().Add(-24*time.Hour).Unix()). // 只查询24小时内的订单
        Find(&orders).Error
    return orders, err
}
```

##### 4. 失败后的处理机制

**方案 A：自动重试**
- 定时任务检测到失败的发票（`docstatus = 0` 且超过一定时间）
- 自动调用 `RedoPosConsumer` 重试

**方案 B：告警通知**
- 检测到失败的发票时，发送告警通知
- 管理员手动处理或触发重试

**方案 C：回退处理**
- 如果发票生成失败且无法重试，记录到失败表
- 提供手动补单功能

### 方案 2：发票生成失败时允许结账继续

修改结账逻辑，发票生成失败时记录错误但不阻止结账。

#### 实现步骤

1. **修改结账逻辑**：发票生成失败时，记录错误日志，但不返回错误
2. **记录失败信息**：将发票生成失败的信息记录到订单表或日志表
3. **后续处理**：通过定时任务或手动重试机制，重新生成发票

#### 代码修改示例

```go
// main/app/service/order_pay.go
if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
    res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
    if err != nil {
        // 记录错误日志，但不阻止结账
        ctx.Log().Error("发票生成失败，但允许结账继续", 
            zap.String("order_no", saleOrder.OrderNo),
            zap.Error(err))
        
        // 可选：记录到订单表，标记发票生成失败
        // repository.NewSaleOrderRepo(db).UpdateSaleOrderErpInvoiceStatus(
        //     saleOrder.Uuid, 
        //     constant.ErpInvoiceStatusFailed,
        //     err.Error(),
        // )
        
        // 不返回错误，允许结账继续
        // return errors.WithMessage(err) // 删除这行
    } else {
        saleOrder.ErpProductsInvoiceName = res.ProductsInvoiceName
        saleOrder.ErpMaterialInvoiceName = res.MaterialInvoiceName
        if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderErpInvoice(
            saleOrder.Uuid, 
            saleOrder.ErpProductsInvoiceName, 
            saleOrder.ErpMaterialInvoiceName,
        ); err != nil {
            return errors.WithMessage(err)
        }
    }
}
```

### 方案 3：添加自动重试机制

对于发票生成失败的情况，添加自动重试机制。

#### 实现步骤

1. **记录失败订单**：发票生成失败时，将订单信息记录到重试表
2. **定时任务**：创建定时任务，定期重试失败的发票生成
3. **重试策略**：设置重试次数限制和重试间隔

#### 重试表结构示例

```sql
CREATE TABLE `erp_invoice_retry` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `sale_order_uuid` bigint unsigned NOT NULL COMMENT '销售订单UUID',
  `order_no` varchar(64) NOT NULL COMMENT '订单号',
  `retry_count` int NOT NULL DEFAULT 0 COMMENT '重试次数',
  `max_retry_count` int NOT NULL DEFAULT 3 COMMENT '最大重试次数',
  `last_error` text COMMENT '最后一次错误信息',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态：0-待重试，1-重试中，2-成功，3-失败',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sale_order_uuid` (`sale_order_uuid`),
  KEY `idx_status_retry_count` (`status`, `retry_count`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ERP发票重试表';
```

## 推荐方案

**优先使用方案 1（异步模式 + 状态查询机制）**，原因：
1. ✅ 系统已支持异步模式，失败不影响结账流程
2. ✅ 需要完善状态跟踪机制（保存 AsyncRecordId）
3. ✅ 需要添加定时任务查询发票状态
4. ✅ 需要处理失败情况（重试或告警）

**实施步骤**：
1. ✅ 第一步：保存 `AsyncRecordId` 到订单表
2. ✅ 第二步：添加查询发票状态的接口
3. ✅ 第三步：创建定时任务定期查询未完成的发票
4. ✅ 第四步：实现失败后的处理机制（重试/告警）

如果无法使用异步模式，则采用**方案 2（允许结账继续）**，并配合**方案 3（自动重试）**。

### 方案 4：处理 TTPOS 调用失败但 ERP 端无数据的情况

当 TTPOS 调用失败且 ERP 端没有记录时，需要重新生成发票。

#### 问题识别

**识别条件**：
1. 订单已结账（`status = 1`）
2. 订单没有 `erp_async_record_id` 或为空
3. 订单没有 `erp_products_invoice_name` 或为空
4. ERP 端 `ReceivePosInvoice` 表中没有对应 `order_no` 的记录

**查询 SQL**：
```sql
-- 查询需要重新生成发票的订单
SELECT 
    so.uuid,
    so.order_no,
    so.erp_async_record_id,
    so.erp_products_invoice_name,
    so.erp_material_invoice_name,
    so.status,
    so.created_at
FROM ttpos_sale_order so
LEFT JOIN receive_pos_invoice rpi ON so.order_no = rpi.order_no
WHERE so.status = 1  -- 已结账
  AND (so.erp_async_record_id = '' OR so.erp_async_record_id IS NULL)
  AND (so.erp_products_invoice_name = '' OR so.erp_products_invoice_name IS NULL)
  AND rpi.id IS NULL  -- ERP 端没有记录
  AND so.created_at > UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 7 DAY));  -- 最近7天的订单
```

#### 解决方案

##### 1. 检测并重新生成发票

**定时任务检测**：
```go
// main/app/job/erp_invoice_missing_detect.go
package job

import (
    "context"
    "ttpos-server-go/app/repository"
    "ttpos-server-go/app/service"
    "ttpos-server-go/pkg/database"
    "ttpos-server-go/pkg/logger"
    "go.uber.org/zap"
)

// DetectAndRegenerateMissingInvoices 检测并重新生成缺失的发票
func DetectAndRegenerateMissingInvoices() {
    dbm := database.GetDBManager(config.DatabaseConf{})
    
    // 查询所有公司
    companies, err := repository.NewCompanyRepo(dbm.GetDB(0)).GetAllCompanies()
    if err != nil {
        logger.Logger.Error("获取公司列表失败", zap.Error(err))
        return
    }
    
    orderSrv := service.NewOrderSrv(dbm, ...)
    
    for _, company := range companies {
        db := dbm.GetDB(company.Uuid)
        orderRepo := repository.NewSaleOrderRepo(db)
        
        // 查询需要重新生成发票的订单
        orders, err := orderRepo.GetOrdersWithoutInvoiceAndAsyncRecordId()
        if err != nil {
            logger.Logger.Error("查询订单失败",
                zap.String("company_uuid", company.Uuid),
                zap.Error(err))
            continue
        }
        
        logger.Logger.Info("发现需要重新生成发票的订单",
            zap.String("company_uuid", company.Uuid),
            zap.Int("count", len(orders)))
        
        for _, order := range orders {
            ctx := context.Background()
            ctx.SetDB(db)
            ctx.SetCompanyUuid(company.Uuid)
            
            // 重新生成发票
            err := orderSrv.RegenerateInvoice(ctx, order.Uuid)
            if err != nil {
                logger.Logger.Error("重新生成发票失败",
                    zap.String("order_no", order.OrderNo),
                    zap.Error(err))
                continue
            }
            
            logger.Logger.Info("重新生成发票成功",
                zap.String("order_no", order.OrderNo))
        }
    }
}

// Repository 方法
// main/app/repository/order.go
func (r *orderRepo) GetOrdersWithoutInvoiceAndAsyncRecordId() ([]*model.SaleOrder, error) {
    var orders []*model.SaleOrder
    
    // 查询已结账但没有发票和异步记录ID的订单
    err := r.db.Where("status = ?", constant.OrderStatusSettled).
        Where("(erp_async_record_id = '' OR erp_async_record_id IS NULL)").
        Where("(erp_products_invoice_name = '' OR erp_products_invoice_name IS NULL)").
        Where("created_at > ?", time.Now().Add(-7*24*time.Hour).Unix()).
        Find(&orders).Error
    
    return orders, err
}
```

##### 2. 重新生成发票服务方法

```go
// main/app/service/erp_invoice_regenerate.go
package service

// RegenerateInvoice 重新生成发票
func (s *orderSrv) RegenerateInvoice(ctx context.Context, orderUuid uint64) error {
    db := ctx.GetDB()
    
    // 获取订单信息
    order, err := repository.NewSaleOrderRepo(db).GetSaleOrderByUuid(orderUuid)
    if err != nil {
        return errors.WithMessage(err, "获取订单信息失败")
    }
    
    // 检查订单是否已结账
    if order.Status != constant.OrderStatusSettled {
        return errors.New("订单未结账，无法生成发票")
    }
    
    // 检查是否已有发票（防止重复生成）
    if order.ErpProductsInvoiceName != "" && order.ErpMaterialInvoiceName != "" {
        ctx.Log().Info("订单已有发票，无需重新生成",
            zap.String("order_no", order.OrderNo))
        return nil
    }
    
    // 检查 ERP 端是否已有记录
    erpSrv := erp.NewIErpSrv(s.dbm)
    hasRecord, err := erpSrv.CheckInvoiceExists(ctx, order.OrderNo)
    if err != nil {
        ctx.Log().Warn("检查ERP端发票记录失败",
            zap.String("order_no", order.OrderNo),
            zap.Error(err))
        // 继续执行，尝试重新生成
    } else if hasRecord {
        ctx.Log().Info("ERP端已有发票记录，无需重新生成",
            zap.String("order_no", order.OrderNo))
        return nil
    }
    
    // 获取账单信息
    saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(order.SaleBillUuid)
    if err != nil {
        return errors.WithMessage(err, "获取账单信息失败")
    }
    
    // 重新生成发票
    ctx.Log().Info("开始重新生成发票",
        zap.String("order_no", order.OrderNo))
    
    res, err := s.SavePosInvoice(ctx, order, saleBill, db)
    if err != nil {
        // 记录失败信息，但不阻止后续处理
        ctx.Log().Error("重新生成发票失败",
            zap.String("order_no", order.OrderNo),
            zap.Error(err))
        
        // 记录到重试表
        retryRecord := &model.ErpInvoiceRetry{
            SaleOrderUuid: order.Uuid,
            OrderNo:       order.OrderNo,
            RetryCount:    0,
            LastError:     err.Error(),
            Status:        constant.ErpInvoiceRetryStatusFailed,
            RetryType:     constant.ErpInvoiceRetryTypeRegenerate, // 标记为重新生成
        }
        repository.NewErpInvoiceRetryRepo(db).Create(retryRecord)
        
        return errors.WithMessage(err, "重新生成发票失败")
    }
    
    // 更新订单表
    updateFields := map[string]interface{}{}
    if res.AsyncRecordId != "" {
        updateFields["erp_async_record_id"] = res.AsyncRecordId
    }
    if res.ProductsInvoiceName != "" {
        updateFields["erp_products_invoice_name"] = res.ProductsInvoiceName
    }
    if res.MaterialInvoiceName != "" {
        updateFields["erp_material_invoice_name"] = res.MaterialInvoiceName
    }
    
    if len(updateFields) > 0 {
        if err := repository.NewSaleOrderRepo(db).Update(order.Uuid, updateFields); err != nil {
            return errors.WithMessage(err, "更新订单发票信息失败")
        }
    }
    
    ctx.Log().Info("重新生成发票成功",
        zap.String("order_no", order.OrderNo),
        zap.String("async_record_id", res.AsyncRecordId))
    
    return nil
}
```

##### 3. 检查 ERP 端是否已有记录

```go
// main/app/service/rpc/erp/selling.go
// CheckInvoiceExists 检查订单是否已有发票记录
func (s *erpSrv) CheckInvoiceExists(ctx pkgCtx.Context, orderNo string) (bool, error) {
    client, conn, err := NewErpSellingClient()
    if err != nil {
        return false, errors.WithMessage(err)
    }
    defer conn.Close()
    
    req := &selling.CheckInvoiceExistsReq{
        OrderNo: orderNo,
    }
    
    res, err := client.CheckInvoiceExists(WithSiteCode(ctx.GetContext(), ctx.GetCompanySetting().ErpnextSiteCode), req)
    if err != nil {
        return false, errors.WithMessage(err)
    }
    
    if res.GetCode() != "0" {
        return false, errors.WithMessage(errors.New(res.Message))
    }
    
    if res.Data != nil {
        var checkResp selling.CheckInvoiceExistsResp
        if err := res.Data.UnmarshalTo(&checkResp); err != nil {
            return false, errors.WithMessage(err)
        }
        return checkResp.Exists, nil
    }
    
    return false, errors.WithMessage(errors.New("检查发票记录异常, data为空"))
}
```

##### 4. 在 ERP 端实现检查接口

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/selling/async_selling.go
// CheckInvoiceExists 检查订单是否已有发票记录
func (s *sAsyncSelling) CheckInvoiceExists(ctx context.Context, req *selling.CheckInvoiceExistsReq) (*selling.CheckInvoiceExistsResp, error) {
    // 查询 ReceivePosInvoice 表
    count, err := dao.ReceivePosInvoice.Ctx(ctx).Count(&do.ReceivePosInvoice{
        OrderNo: req.OrderNo,
    })
    if err != nil {
        return nil, gerror.Wrapf(err, "查询发票记录失败")
    }
    
    return &selling.CheckInvoiceExistsResp{
        Exists: count > 0,
    }, nil
}
```

##### 5. 降级处理：允许结账继续

修改结账逻辑，调用失败时允许结账继续，后续通过定时任务重新生成：

```go
// main/app/service/order_pay.go
if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
    res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
    if err != nil {
        // 判断错误类型
        errorMsg := err.Error()
        isNetworkError := strings.Contains(errorMsg, "连接") ||
                         strings.Contains(errorMsg, "超时") ||
                         strings.Contains(errorMsg, "network") ||
                         strings.Contains(errorMsg, "timeout") ||
                         strings.Contains(errorMsg, "deadline exceeded")
        
        if isNetworkError {
            // 网络错误：允许结账继续，记录到重试表
            ctx.Log().Warn("发票生成失败（网络错误），允许结账继续",
                zap.String("order_no", saleOrder.OrderNo),
                zap.Error(err))
            
            // 记录到重试表，标记为需要重新生成
            retryRecord := &model.ErpInvoiceRetry{
                SaleOrderUuid: saleOrder.Uuid,
                OrderNo:       saleOrder.OrderNo,
                RetryCount:    0,
                LastError:     err.Error(),
                Status:        constant.ErpInvoiceRetryStatusPending,
                RetryType:     constant.ErpInvoiceRetryTypeRegenerate,
            }
            repository.NewErpInvoiceRetryRepo(db).Create(retryRecord)
            
            // 不返回错误，允许结账继续
        } else {
            // 其他错误：根据业务需求决定是否允许结账继续
            // 这里可以根据错误类型决定处理方式
            return errors.WithMessage(err)
        }
    } else {
        // 成功：正常处理
        saleOrder.ErpProductsInvoiceName = res.ProductsInvoiceName
        saleOrder.ErpMaterialInvoiceName = res.MaterialInvoiceName
        if res.AsyncRecordId != "" {
            saleOrder.ErpAsyncRecordId = res.AsyncRecordId
        }
        // ... 更新订单表
    }
}
```

#### 处理流程

1. **检测阶段**（定时任务，每 5 分钟执行一次）
   - 查询已结账但没有发票和 AsyncRecordId 的订单
   - 检查 ERP 端是否有对应记录
   - 如果没有，标记为需要重新生成

2. **重新生成阶段**（定时任务，每 10 分钟执行一次）
   - 查询需要重新生成的订单
   - 调用 `RegenerateInvoice` 重新生成发票
   - 记录成功或失败

3. **监控告警**
   - 监控需要重新生成的订单数量
   - 超过阈值时发送告警

#### 幂等性保证

为了防止重复生成发票，需要保证幂等性：

1. **检查订单是否已有发票**
   - 如果订单已有 `erp_products_invoice_name`，跳过

2. **检查 ERP 端是否已有记录**
   - 调用 `CheckInvoiceExists` 检查
   - 如果已有记录，查询并更新订单表

3. **使用订单号作为唯一标识**
   - ERP 端会检查订单号是否已存在
   - 如果已存在，返回错误而不是创建新记录

## 手动处理失败的发票

### 查看失败的发票记录

**在 ERP 端（ttpos-bmp）**：
```sql
-- 查看未处理的发票记录（状态为 Draft）
SELECT 
    id,
    order_no,
    open_pos_entry_name,
    docstatus,
    resp_body,
    created_at
FROM receive_pos_invoice
WHERE docstatus = 0  -- Draft 状态
ORDER BY created_at DESC;

-- 查看失败的发票记录（有错误信息）
SELECT 
    id,
    order_no,
    open_pos_entry_name,
    docstatus,
    resp_body,
    created_at
FROM receive_pos_invoice
WHERE docstatus = 0 
  AND resp_body LIKE '%失败%'
ORDER BY created_at DESC;
```

**在 TTPOS 端（main）**：
```sql
-- 查看有 AsyncRecordId 但发票名称为空的订单
SELECT 
    uuid,
    order_no,
    erp_async_record_id,
    erp_products_invoice_name,
    erp_material_invoice_name,
    status,
    created_at
FROM ttpos_sale_order
WHERE erp_async_record_id != ''
  AND (erp_products_invoice_name = '' OR erp_material_invoice_name = '')
  AND status = 1  -- 已结账
ORDER BY created_at DESC;

-- 查看 TTPOS 调用失败但 ERP 端无数据的订单（需要重新生成）
SELECT 
    so.uuid,
    so.order_no,
    so.erp_async_record_id,
    so.erp_products_invoice_name,
    so.erp_material_invoice_name,
    so.status,
    so.created_at
FROM ttpos_sale_order so
LEFT JOIN receive_pos_invoice rpi ON so.order_no = rpi.order_no
WHERE so.status = 1  -- 已结账
  AND (so.erp_async_record_id = '' OR so.erp_async_record_id IS NULL)
  AND (so.erp_products_invoice_name = '' OR so.erp_products_invoice_name IS NULL)
  AND rpi.id IS NULL  -- ERP 端没有记录
  AND so.created_at > UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 7 DAY))
ORDER BY so.created_at DESC;
```

### 手动查询发票状态

**通过订单号查询**：
```go
// 调用查询接口
status, err := erpSrv.GetInvoiceStatus(ctx, order.ErpAsyncRecordId)
if err != nil {
    // 处理错误
}
// status.Docstatus: 0-Draft, 1-Submitted, 2-Cancelled
// status.ProductsInvoiceName: 商品发票名称
// status.MaterialInvoiceName: 原材料发票名称
// status.RespBody: 错误信息（如果有）
```

### 手动重试失败的发票

**方式 1：通过 RedoPosConsumer 重试**

系统提供了 `RedoPosConsumer`，可以重做未处理的发票：

```json
{
  "msg_type": "save-pos-invoice",
  "pos_open_entry_name": "POS-OPE-2025-00238"
}
```

发送到消息队列的 `TopicRedoPos` 主题即可。

**方式 2：通过订单号重试**

如果需要重试特定订单的发票：
```json
{
  "msg_type": "save-pos-invoice",
  "record_id": 12345  // ReceivePosInvoice 表的 id
}
```

**方式 3：直接更新订单发票名称**

如果发票已生成成功，但订单表未更新：
```sql
-- 1. 查询 ReceivePosInvoice 表获取发票名称
SELECT 
    id,
    order_no,
    products_invoice_name,
    material_invoice_name,
    docstatus
FROM receive_pos_invoice
WHERE order_no = 'ORDER-2025-001';

-- 2. 更新订单表
UPDATE ttpos_sale_order
SET erp_products_invoice_name = 'POS-INV-2025-001',
    erp_material_invoice_name = 'POS-INV-2025-002'
WHERE order_no = 'ORDER-2025-001';
```

## 发票一直失败的处理方案

当发票生成一直失败时，需要采取更积极的处理措施。

### 失败原因分析

首先需要分析失败的根本原因：

```sql
-- 分析失败原因
SELECT 
    resp_body,
    COUNT(*) as count,
    MIN(created_at) as first_fail_time,
    MAX(created_at) as last_fail_time
FROM receive_pos_invoice
WHERE docstatus = 0 
  AND resp_body IS NOT NULL
  AND resp_body != ''
GROUP BY resp_body
ORDER BY count DESC;
```

**常见失败原因**：

1. **ERP 系统故障**
   - 错误信息：`连接ERP系统失败`、`ERP服务不可用`
   - 处理：检查 ERP 系统状态，等待恢复

2. **数据校验失败**
   - 错误信息：`Item Code: XXX is not available`、`库存不足`
   - 处理：修复 ERP 系统中的商品数据或库存

3. **网络问题**
   - 错误信息：`网络超时`、`连接超时`
   - 处理：检查网络连接，增加超时时间

4. **权限问题**
   - 错误信息：`权限不足`、`无权限访问`
   - 处理：检查 ERP 系统权限配置

5. **数据格式错误**
   - 错误信息：`数据格式错误`、`必填字段缺失`
   - 处理：检查发送给 ERP 的数据格式

### 重试策略

#### 1. 指数退避重试

实现指数退避重试机制，避免频繁重试导致系统压力：

```go
// main/app/service/erp_invoice_retry.go
package service

import (
    "time"
    "math"
    "ttpos-server-go/app/repository"
    "ttpos-server-go/pkg/logger"
    "go.uber.org/zap"
)

const (
    MaxRetryCount = 5           // 最大重试次数
    BaseRetryDelay = 30        // 基础重试延迟（秒）
    MaxRetryDelay = 3600       // 最大重试延迟（秒）
)

// RetryFailedInvoice 重试失败的发票
func (s *orderSrv) RetryFailedInvoice(ctx context.Context, order *model.SaleOrder) error {
    if order.ErpAsyncRecordId == "" {
        return errors.New("订单没有异步记录ID")
    }
    
    // 查询重试记录
    retryRecord, err := repository.NewErpInvoiceRetryRepo(ctx.GetDB()).GetByOrderUuid(order.Uuid)
    if err != nil {
        return errors.WithMessage(err, "查询重试记录失败")
    }
    
    // 如果重试次数已达上限，不再重试
    if retryRecord != nil && retryRecord.RetryCount >= MaxRetryCount {
        logger.Logger.Warn("发票重试次数已达上限",
            zap.String("order_no", order.OrderNo),
            zap.Int("retry_count", retryRecord.RetryCount))
        return errors.New("发票重试次数已达上限")
    }
    
    // 计算重试延迟（指数退避）
    retryCount := 0
    if retryRecord != nil {
        retryCount = retryRecord.RetryCount
    }
    delay := int64(BaseRetryDelay * math.Pow(2, float64(retryCount)))
    if delay > MaxRetryDelay {
        delay = MaxRetryDelay
    }
    
    // 检查是否到了重试时间
    if retryRecord != nil {
        nextRetryTime := retryRecord.UpdatedAt + delay
        if time.Now().Unix() < nextRetryTime {
            logger.Logger.Info("还未到重试时间",
                zap.String("order_no", order.OrderNo),
                zap.Int64("next_retry_time", nextRetryTime))
            return nil
        }
    }
    
    // 执行重试
    ctx.Log().Info("开始重试发票生成",
        zap.String("order_no", order.OrderNo),
        zap.Int("retry_count", retryCount+1))
    
    // 调用重试接口
    err = s.retryInvoiceGeneration(ctx, order)
    if err != nil {
        // 更新重试记录
        if retryRecord == nil {
            retryRecord = &model.ErpInvoiceRetry{
                SaleOrderUuid: order.Uuid,
                OrderNo:       order.OrderNo,
                RetryCount:    1,
                LastError:     err.Error(),
                Status:        constant.ErpInvoiceRetryStatusFailed,
            }
            repository.NewErpInvoiceRetryRepo(ctx.GetDB()).Create(retryRecord)
        } else {
            retryRecord.RetryCount++
            retryRecord.LastError = err.Error()
            retryRecord.Status = constant.ErpInvoiceRetryStatusFailed
            repository.NewErpInvoiceRetryRepo(ctx.GetDB()).Update(retryRecord)
        }
        return errors.WithMessage(err, "重试发票生成失败")
    }
    
    // 重试成功，更新记录
    if retryRecord != nil {
        retryRecord.Status = constant.ErpInvoiceRetryStatusSuccess
        repository.NewErpInvoiceRetryRepo(ctx.GetDB()).Update(retryRecord)
    }
    
    return nil
}

// retryInvoiceGeneration 重试发票生成
func (s *orderSrv) retryInvoiceGeneration(ctx context.Context, order *model.SaleOrder) error {
    // 通过 RedoPosConsumer 重试
    // 发送消息到消息队列
    msg := &mq.AsyncSellingMsg{
        RecordId: gconv.Int64(order.ErpAsyncRecordId),
        MsgType:  mq.MsgTypeSavePosInvoice,
    }
    
    return queue.Push(string(consts.TopicSavePosInvoice), msg)
}
```

#### 2. 定时任务自动重试

创建定时任务，自动重试失败的发票：

```go
// main/app/job/erp_invoice_auto_retry.go
package job

import (
    "time"
    "ttpos-server-go/app/repository"
    "ttpos-server-go/app/service"
    "ttpos-server-go/pkg/database"
    "ttpos-server-go/pkg/logger"
    "go.uber.org/zap"
)

// AutoRetryFailedInvoices 自动重试失败的发票
func AutoRetryFailedInvoices() {
    dbm := database.GetDBManager(config.DatabaseConf{})
    
    // 查询需要重试的订单
    // 条件：有 AsyncRecordId，发票名称为空，已结账，且重试次数未达上限
    db := dbm.GetDB(0) // 需要遍历所有公司
    retryRepo := repository.NewErpInvoiceRetryRepo(db)
    orderRepo := repository.NewSaleOrderRepo(db)
    
    // 查询需要重试的订单
    retryRecords, err := retryRepo.GetPendingRetries(MaxRetryCount)
    if err != nil {
        logger.Logger.Error("查询待重试订单失败", zap.Error(err))
        return
    }
    
    orderSrv := service.NewOrderSrv(dbm, ...)
    
    for _, retryRecord := range retryRecords {
        // 检查是否到了重试时间
        delay := int64(BaseRetryDelay * math.Pow(2, float64(retryRecord.RetryCount)))
        if delay > MaxRetryDelay {
            delay = MaxRetryDelay
        }
        nextRetryTime := retryRecord.UpdatedAt + delay
        
        if time.Now().Unix() < nextRetryTime {
            continue // 还未到重试时间
        }
        
        // 获取订单信息
        order, err := orderRepo.GetSaleOrderByUuid(retryRecord.SaleOrderUuid)
        if err != nil {
            logger.Logger.Error("获取订单信息失败",
                zap.String("order_no", retryRecord.OrderNo),
                zap.Error(err))
            continue
        }
        
        // 执行重试
        ctx := context.Background()
        ctx.SetDB(db)
        ctx.SetCompanyUuid(order.CompanyUuid)
        
        err = orderSrv.RetryFailedInvoice(ctx, order)
        if err != nil {
            logger.Logger.Error("自动重试发票失败",
                zap.String("order_no", retryRecord.OrderNo),
                zap.Error(err))
        }
    }
}
```

### 降级处理方案

当发票生成一直失败且无法修复时，需要采取降级处理：

#### 1. 允许结账继续（已实现）

修改结账逻辑，发票生成失败时允许结账继续：

```go
// main/app/service/order_pay.go
if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
    res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
    if err != nil {
        // 降级处理：记录错误但允许结账继续
        ctx.Log().Error("发票生成失败，降级处理：允许结账继续", 
            zap.String("order_no", saleOrder.OrderNo),
            zap.Error(err))
        
        // 记录到失败表，后续处理
        retryRecord := &model.ErpInvoiceRetry{
            SaleOrderUuid: saleOrder.Uuid,
            OrderNo:       saleOrder.OrderNo,
            RetryCount:    0,
            LastError:     err.Error(),
            Status:        constant.ErpInvoiceRetryStatusPending,
        }
        repository.NewErpInvoiceRetryRepo(db).Create(retryRecord)
        
        // 不返回错误，允许结账继续
        // return errors.WithMessage(err) // 删除这行
    } else {
        // 正常处理：保存发票信息
        saleOrder.ErpProductsInvoiceName = res.ProductsInvoiceName
        saleOrder.ErpMaterialInvoiceName = res.MaterialInvoiceName
        if res.AsyncRecordId != "" {
            saleOrder.ErpAsyncRecordId = res.AsyncRecordId
        }
        // ... 更新订单表
    }
}
```

#### 2. 后续补单机制

提供手动补单功能，允许管理员手动补开发票：

```go
// main/app/service/erp_invoice_manual.go
package service

// ManualCreateInvoice 手动补开发票
func (s *orderSrv) ManualCreateInvoice(ctx context.Context, orderUuid uint64) error {
    // 获取订单信息
    order, err := repository.NewSaleOrderRepo(ctx.GetDB()).GetSaleOrderByUuid(orderUuid)
    if err != nil {
        return errors.WithMessage(err, "获取订单信息失败")
    }
    
    // 检查订单是否已结账
    if order.Status != constant.OrderStatusSettled {
        return errors.New("订单未结账，无法补开发票")
    }
    
    // 检查是否已有发票
    if order.ErpProductsInvoiceName != "" && order.ErpMaterialInvoiceName != "" {
        return errors.New("订单已有发票，无需补开")
    }
    
    // 重新生成发票
    saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(order.SaleBillUuid)
    if err != nil {
        return errors.WithMessage(err, "获取账单信息失败")
    }
    
    res, err := s.SavePosInvoice(ctx, order, saleBill, ctx.GetDB())
    if err != nil {
        return errors.WithMessage(err, "补开发票失败")
    }
    
    // 更新订单表
    updateFields := map[string]interface{}{
        "erp_products_invoice_name": res.ProductsInvoiceName,
        "erp_material_invoice_name": res.MaterialInvoiceName,
    }
    if res.AsyncRecordId != "" {
        updateFields["erp_async_record_id"] = res.AsyncRecordId
    }
    
    return repository.NewSaleOrderRepo(ctx.GetDB()).Update(order.Uuid, updateFields)
}
```

### 数据修复方案

当发票已生成成功但订单表未更新时，需要数据修复：

#### 1. 批量修复脚本

```go
// main/app/job/erp_invoice_data_fix.go
package job

// FixInvoiceData 修复发票数据
func FixInvoiceData() {
    dbm := database.GetDBManager(config.DatabaseConf{})
    
    // 查询所有有 AsyncRecordId 但发票名称为空的订单
    db := dbm.GetDB(0)
    orderRepo := repository.NewSaleOrderRepo(db)
    erpSrv := erp.NewIErpSrv(dbm)
    
    orders, err := orderRepo.GetOrdersWithAsyncRecordIdButNoInvoice()
    if err != nil {
        logger.Logger.Error("查询待修复订单失败", zap.Error(err))
        return
    }
    
    successCount := 0
    failCount := 0
    
    for _, order := range orders {
        ctx := context.Background()
        ctx.SetDB(db)
        ctx.SetCompanyUuid(order.CompanyUuid)
        
        // 查询发票状态
        status, err := erpSrv.GetInvoiceStatus(ctx, order.ErpAsyncRecordId)
        if err != nil {
            logger.Logger.Error("查询发票状态失败",
                zap.String("order_no", order.OrderNo),
                zap.Error(err))
            failCount++
            continue
        }
        
        // 如果发票已生成成功，更新订单表
        if status.Docstatus == 1 { // Submitted
            updateFields := map[string]interface{}{}
            if status.ProductsInvoiceName != "" {
                updateFields["erp_products_invoice_name"] = status.ProductsInvoiceName
            }
            if status.MaterialInvoiceName != "" {
                updateFields["erp_material_invoice_name"] = status.MaterialInvoiceName
            }
            
            if len(updateFields) > 0 {
                if err := orderRepo.Update(order.Uuid, updateFields); err != nil {
                    logger.Logger.Error("更新订单发票名称失败",
                        zap.String("order_no", order.OrderNo),
                        zap.Error(err))
                    failCount++
                } else {
                    logger.Logger.Info("修复订单发票数据成功",
                        zap.String("order_no", order.OrderNo))
                    successCount++
                }
            }
        } else {
            logger.Logger.Warn("发票状态异常",
                zap.String("order_no", order.OrderNo),
                zap.Int("docstatus", status.Docstatus),
                zap.String("error", status.RespBody))
            failCount++
        }
    }
    
    logger.Logger.Info("发票数据修复完成",
        zap.Int("total", len(orders)),
        zap.Int("success", successCount),
        zap.Int("fail", failCount))
}
```

#### 2. SQL 直接修复

如果确定发票已生成成功，可以直接通过 SQL 修复：

```sql
-- 1. 查询需要修复的订单和对应的发票
SELECT 
    so.uuid as order_uuid,
    so.order_no,
    so.erp_async_record_id,
    rpi.products_invoice_name,
    rpi.material_invoice_name,
    rpi.docstatus
FROM ttpos_sale_order so
LEFT JOIN receive_pos_invoice rpi ON so.erp_async_record_id = CAST(rpi.id AS CHAR)
WHERE so.erp_async_record_id != ''
  AND so.status = 1
  AND (so.erp_products_invoice_name = '' OR so.erp_material_invoice_name = '')
  AND rpi.docstatus = 1;  -- 发票已成功

-- 2. 批量更新订单表
UPDATE ttpos_sale_order so
INNER JOIN receive_pos_invoice rpi ON so.erp_async_record_id = CAST(rpi.id AS CHAR)
SET 
    so.erp_products_invoice_name = rpi.products_invoice_name,
    so.erp_material_invoice_name = rpi.material_invoice_name
WHERE so.erp_async_record_id != ''
  AND so.status = 1
  AND (so.erp_products_invoice_name = '' OR so.erp_material_invoice_name = '')
  AND rpi.docstatus = 1;
```

### 紧急处理流程

当发票生成一直失败时，按以下流程处理：

#### 1. 立即处理（0-30分钟）

1. **检查 ERP 系统状态**
   ```bash
   # 检查 ERP 服务是否正常
   curl -X GET http://erp-service/health
   
   # 检查消息队列是否正常
   # 检查数据库连接是否正常
   ```

2. **查看错误日志**
   ```sql
   SELECT * FROM receive_pos_invoice 
   WHERE docstatus = 0 
   ORDER BY created_at DESC 
   LIMIT 20;
   ```

3. **手动重试关键订单**
   - 优先处理金额较大的订单
   - 优先处理客户投诉的订单

#### 2. 短期处理（30分钟-2小时）

1. **启用降级处理**
   - 修改代码，允许结账继续
   - 记录失败订单，后续处理

2. **批量重试**
   - 使用 `RedoPosConsumer` 批量重试
   - 设置合理的重试间隔

3. **监控告警**
   - 设置告警规则
   - 通知相关人员

#### 3. 长期处理（2小时以上）

1. **修复根本原因**
   - 修复 ERP 系统问题
   - 修复数据问题
   - 修复网络问题

2. **数据修复**
   - 运行数据修复脚本
   - 手动补开发票

3. **优化系统**
   - 优化重试机制
   - 优化错误处理
   - 优化监控告警

### 业务影响处理

发票生成失败可能影响的业务：

1. **财务对账**
   - 影响：ERP 系统无法对账
   - 处理：手动对账或延迟对账

2. **库存管理**
   - 影响：库存可能不准确
   - 处理：手动调整库存或等待修复

3. **报表统计**
   - 影响：销售报表不准确
   - 处理：使用 TTPOS 数据或等待修复

4. **客户服务**
   - 影响：客户可能无法获取发票
   - 处理：手动补开发票或说明情况

## 监控和告警

### 监控指标

建议添加以下监控：

1. **发票生成失败率监控**：监控发票生成失败的比例
   ```sql
   SELECT 
       COUNT(*) as total,
       SUM(CASE WHEN docstatus = 0 THEN 1 ELSE 0 END) as pending,
       SUM(CASE WHEN docstatus = 1 THEN 1 ELSE 0 END) as success,
       SUM(CASE WHEN docstatus = 2 THEN 1 ELSE 0 END) as cancelled
   FROM receive_pos_invoice
   WHERE created_at > UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 24 HOUR));
   ```

2. **待处理发票数量**：监控待处理的发票数量
   ```sql
   SELECT COUNT(*) 
   FROM receive_pos_invoice
   WHERE docstatus = 0 
     AND created_at < UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 5 MINUTE));
   ```

3. **订单发票同步状态**：监控订单表与发票表的同步情况
   ```sql
   SELECT COUNT(*) 
   FROM ttpos_sale_order so
   LEFT JOIN receive_pos_invoice rpi ON so.erp_async_record_id = CAST(rpi.id AS CHAR)
   WHERE so.erp_async_record_id != ''
     AND so.status = 1
     AND (so.erp_products_invoice_name = '' OR so.erp_material_invoice_name = '')
     AND rpi.docstatus = 1;  -- 发票已成功但订单未更新
   ```

4. **失败订单告警**：发票生成失败的订单数量超过阈值时告警
5. **ERP 连接状态监控**：监控 ERP 系统的连接状态和响应时间

### 告警规则

建议设置以下告警：

1. **待处理发票超过阈值**：超过 10 个待处理发票且超过 5 分钟
2. **发票生成失败率过高**：失败率超过 5%
3. **订单发票不同步**：超过 20 个订单发票未同步
4. **ERP 服务不可用**：连续 3 次查询失败
5. **重试次数过多**：单个订单重试次数超过 3 次
6. **长时间失败**：发票生成失败超过 1 小时

### 告警提示实现方案

#### 告警位置分析

**建议在 TTPOS 侧做告警提示**，原因：
1. ✅ **订单数据在 TTPOS 侧**：订单信息、客户信息都在 TTPOS 侧，便于生成详细的告警信息
2. ✅ **业务影响在 TTPOS 侧**：发票生成失败直接影响 TTPOS 的业务流程
3. ✅ **用户操作在 TTPOS 侧**：管理员需要在 TTPOS 侧查看和处理问题
4. ✅ **已有告警机制**：TTPOS 侧已有库存预警等告警机制，可以复用

**ERP 侧的作用**：
- 记录详细的错误日志
- 提供错误信息供 TTPOS 侧查询
- 不直接发送告警（避免重复告警）

#### 告警方式

参考现有的库存预警机制，提供以下告警方式：

1. **邮件告警**（推荐）
   - 发送给超级管理员或财务人员
   - 包含订单详情、错误信息、处理建议

2. **系统通知**
   - 在管理后台显示告警通知
   - 实时提醒管理员

3. **日志记录**
   - 记录到告警日志表
   - 便于后续分析和统计

#### 实现方案

##### 1. 创建告警日志表

```sql
CREATE TABLE `ttpos_erp_invoice_alert_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `company_uuid` bigint unsigned NOT NULL COMMENT '公司UUID',
  `sale_order_uuid` bigint unsigned NOT NULL COMMENT '销售订单UUID',
  `order_no` varchar(64) NOT NULL COMMENT '订单号',
  `alert_type` tinyint NOT NULL DEFAULT 1 COMMENT '告警类型：1-发票生成失败 2-发票长时间未生成 3-重试次数过多',
  `error_message` text COMMENT '错误信息',
  `erp_async_record_id` varchar(255) DEFAULT '' COMMENT 'ERP异步记录ID',
  `retry_count` int NOT NULL DEFAULT 0 COMMENT '重试次数',
  `last_alert_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '上次告警时间（时间戳）',
  `alert_count` int unsigned NOT NULL DEFAULT 0 COMMENT '告警次数',
  `send_status` tinyint NOT NULL DEFAULT 0 COMMENT '发送状态：0-待发送 1-发送成功 2-发送失败',
  `recipient` varchar(255) DEFAULT '' COMMENT '收件人邮箱',
  `message_uuid` bigint unsigned DEFAULT 0 COMMENT '消息UUID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `updated_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_company_order` (`company_uuid`, `sale_order_uuid`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_alert_type_time` (`alert_type`, `last_alert_time`),
  KEY `idx_send_status` (`send_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ERP发票告警日志表';
```

##### 2. 告警服务实现

```go
// main/app/service/erp_invoice_alert.go
package service

import (
    "context"
    "fmt"
    "strconv"
    "time"
    "ttpos-server-go/app/constant"
    "ttpos-server-go/app/model"
    "ttpos-server-go/app/repository"
    "ttpos-server-go/pkg/logger"
    "ttpos-server-go/pkg/utils"
    v1 "ttpos-message/api/v1"
    "go.uber.org/zap"
)

// SendInvoiceAlert 发送发票生成失败告警
func (s *orderSrv) SendInvoiceAlert(ctx context.Context, order *model.SaleOrder, alertType int, errorMsg string) error {
    db := ctx.GetDB()
    alertLogRepo := repository.NewErpInvoiceAlertLogRepo(db)
    
    // 检查是否需要发送告警（24小时内最多2次）
    shouldSend, existingLog, err := alertLogRepo.ShouldSendAlert(order.CompanyUuid, order.Uuid, alertType)
    if err != nil {
        logger.Logger.Error("检查是否需要发送告警失败",
            zap.String("order_no", order.OrderNo),
            zap.Error(err))
        return err
    }
    
    if !shouldSend {
        logger.Logger.Info("跳过发送告警（24小时内已发送过）",
            zap.String("order_no", order.OrderNo),
            zap.Int("alert_count", existingLog.AlertCount))
        return nil
    }
    
    // 获取收件人邮箱
    recipient := s.getInvoiceAlertRecipient(ctx, order.CompanyUuid)
    if recipient == "" {
        logger.Logger.Warn("未配置发票告警邮箱，跳过发送",
            zap.String("order_no", order.OrderNo))
        return nil
    }
    
    // 准备邮件内容
    subject, messageArgs, templateUuid := s.buildAlertEmailContent(order, alertType, errorMsg)
    
    messageUuid, _ := utils.GetID()
    sendMessageReq := &v1.SendMessageReq{
        MessageUuid:  strconv.FormatUint(messageUuid, 10),
        TemplateUuid: templateUuid,
        MessageArgs:  messageArgs,
        MessageType:  "email",
        Recipient:    recipient,
        Subject:      subject,
        CompanyUuid:  strconv.FormatUint(order.CompanyUuid, 10),
        OperatorUuid: recipient,
    }
    
    // 发送邮件（带重试）
    sendSuccess := false
    var lastErr error
    maxRetries := 2
    
    for i := 0; i <= maxRetries; i++ {
        if i > 0 {
            time.Sleep(time.Second * 2)
        }
        
        _, err := s.messageSrv.SendMessage(ctx.GetContext(), sendMessageReq)
        if err == nil {
            sendSuccess = true
            break
        }
        lastErr = err
    }
    
    // 记录告警日志
    now := uint64(time.Now().Unix())
    if existingLog == nil {
        newLog := &model.ErpInvoiceAlertLog{
            CompanyUuid:      order.CompanyUuid,
            SaleOrderUuid:    order.Uuid,
            OrderNo:          order.OrderNo,
            AlertType:        alertType,
            ErrorMessage:     errorMsg,
            ErpAsyncRecordId: order.ErpAsyncRecordId,
            RetryCount:       0,
            LastAlertTime:    now,
            AlertCount:       1,
            SendStatus:       utils.IfUint8(sendSuccess, model.SendStatusSuccess, model.SendStatusFailed),
            Recipient:        recipient,
            MessageUuid:      messageUuid,
            CreatedAt:        now,
            UpdatedAt:        now,
        }
        if !sendSuccess && lastErr != nil {
            newLog.ErrorMessage = fmt.Sprintf("%s; 发送告警失败: %v", errorMsg, lastErr)
        }
        alertLogRepo.CreateAlertLog(newLog)
    } else {
        existingLog.ErrorMessage = errorMsg
        existingLog.ErpAsyncRecordId = order.ErpAsyncRecordId
        existingLog.Recipient = recipient
        existingLog.MessageUuid = messageUuid
        existingLog.UpdatedAt = now
        
        if sendSuccess {
            existingLog.AlertCount++
            existingLog.LastAlertTime = now
            existingLog.SendStatus = model.SendStatusSuccess
        } else {
            existingLog.SendStatus = model.SendStatusFailed
            if lastErr != nil {
                existingLog.ErrorMessage = fmt.Sprintf("%s; 发送告警失败: %v", errorMsg, lastErr)
            }
        }
        alertLogRepo.UpdateAlertLog(existingLog)
    }
    
    if sendSuccess {
        logger.Logger.Info("发票告警邮件发送成功",
            zap.String("order_no", order.OrderNo),
            zap.String("recipient", recipient))
    } else {
        logger.Logger.Error("发票告警邮件发送失败",
            zap.String("order_no", order.OrderNo),
            zap.Error(lastErr))
    }
    
    return nil
}

// buildAlertEmailContent 构建告警邮件内容
func (s *orderSrv) buildAlertEmailContent(order *model.SaleOrder, alertType int, errorMsg string) (subject, messageArgs, templateUuid string) {
    company := ctx.GetCompany()
    orderTime := time.Unix(order.CreateTime, 0).Format("2006-01-02 15:04:05")
    
    switch alertType {
    case constant.ErpInvoiceAlertTypeGenerationFailed:
        templateUuid = constant.TplErpInvoiceGenerationFailed
        subject = fmt.Sprintf("[告警] %s - 订单 %s 发票生成失败", company.Name, order.OrderNo)
        messageArgs = fmt.Sprintf(`{
            "company": "%s",
            "order_no": "%s",
            "order_time": "%s",
            "order_amount": "%.2f",
            "error_message": "%s",
            "async_record_id": "%s"
        }`, company.Name, order.OrderNo, orderTime, order.Amount, errorMsg, order.ErpAsyncRecordId)
        
    case constant.ErpInvoiceAlertTypeLongTimePending:
        templateUuid = constant.TplErpInvoiceLongTimePending
        subject = fmt.Sprintf("[告警] %s - 订单 %s 发票长时间未生成", company.Name, order.OrderNo)
        messageArgs = fmt.Sprintf(`{
            "company": "%s",
            "order_no": "%s",
            "order_time": "%s",
            "order_amount": "%.2f",
            "pending_hours": "%d"
        }`, company.Name, order.OrderNo, orderTime, order.Amount, 1)
        
    case constant.ErpInvoiceAlertTypeRetryExceeded:
        templateUuid = constant.TplErpInvoiceRetryExceeded
        subject = fmt.Sprintf("[告警] %s - 订单 %s 发票重试次数过多", company.Name, order.OrderNo)
        messageArgs = fmt.Sprintf(`{
            "company": "%s",
            "order_no": "%s",
            "order_time": "%s",
            "order_amount": "%.2f",
            "retry_count": "%d",
            "error_message": "%s"
        }`, company.Name, order.OrderNo, orderTime, order.Amount, 3, errorMsg)
    }
    
    return subject, messageArgs, templateUuid
}

// getInvoiceAlertRecipient 获取发票告警邮件收件人
func (s *orderSrv) getInvoiceAlertRecipient(ctx context.Context, companyUuid uint64) string {
    db := s.dbm.GetDB(companyUuid)
    staffRepo := repository.NewStaffRepo(db)
    
    // 查询超级管理员或财务人员
    staffs := staffRepo.GetStaffs(
        staffRepo.WhereIsSuper(1),
        // 或者查询特定角色的员工
        // staffRepo.WhereRole(constant.RoleFinance),
    )
    
    if len(staffs) == 0 {
        logger.Logger.Warn("未找到告警收件人",
            zap.Uint64("company_uuid", companyUuid))
        return ""
    }
    
    // 返回第一个超级管理员的邮箱
    for _, staff := range staffs {
        if staff.Username != "" {
            return staff.Username
        }
    }
    
    return ""
}
```

##### 3. 在发票生成失败时触发告警

```go
// main/app/service/order_pay.go
if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
    res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
    if err != nil {
        // 判断错误类型
        errorMsg := err.Error()
        isNetworkError := strings.Contains(errorMsg, "连接") ||
                         strings.Contains(errorMsg, "超时") ||
                         strings.Contains(errorMsg, "network") ||
                         strings.Contains(errorMsg, "timeout")
        
        if isNetworkError {
            // 网络错误：允许结账继续，发送告警
            ctx.Log().Warn("发票生成失败（网络错误），允许结账继续",
                zap.String("order_no", saleOrder.OrderNo),
                zap.Error(err))
            
            // 发送告警
            utils.Go(func() {
                s.SendInvoiceAlert(ctx, saleOrder, constant.ErpInvoiceAlertTypeGenerationFailed, errorMsg)
            })
            
            // 记录到重试表
            // ...
        } else {
            // 其他错误：发送告警并决定是否允许结账继续
            utils.Go(func() {
                s.SendInvoiceAlert(ctx, saleOrder, constant.ErpInvoiceAlertTypeGenerationFailed, errorMsg)
            })
            
            // 根据业务需求决定是否返回错误
            return errors.WithMessage(err)
        }
    }
    // ... 正常处理
}
```

##### 4. 定时任务检测并发送告警

```go
// main/app/job/erp_invoice_alert_check.go
package job

// CheckInvoiceAlerts 检查发票告警
func CheckInvoiceAlerts() {
    dbm := database.GetDBManager(config.DatabaseConf{})
    orderSrv := service.NewOrderSrv(dbm, ...)
    
    // 检查长时间未生成的发票
    checkLongTimePendingInvoices(orderSrv, dbm)
    
    // 检查重试次数过多的发票
    checkRetryExceededInvoices(orderSrv, dbm)
}

// checkLongTimePendingInvoices 检查长时间未生成的发票
func checkLongTimePendingInvoices(orderSrv *service.OrderSrv, dbm *database.DBManager) {
    // 查询有 AsyncRecordId 但超过1小时未生成的订单
    // 发送告警
}

// checkRetryExceededInvoices 检查重试次数过多的发票
func checkRetryExceededInvoices(orderSrv *service.OrderSrv, dbm *database.DBManager) {
    // 查询重试次数超过3次的订单
    // 发送告警
}
```

##### 5. 告警常量定义

```go
// main/app/constant/erp_invoice_alert.go
package constant

// ERP发票告警类型
const (
    ErpInvoiceAlertTypeGenerationFailed  = 1 // 发票生成失败
    ErpInvoiceAlertTypeLongTimePending   = 2 // 发票长时间未生成
    ErpInvoiceAlertTypeRetryExceeded     = 3 // 重试次数过多
)

// 邮件模板UUID
const (
    TplErpInvoiceGenerationFailed = "erp-invoice-generation-failed"
    TplErpInvoiceLongTimePending  = "erp-invoice-long-time-pending"
    TplErpInvoiceRetryExceeded    = "erp-invoice-retry-exceeded"
)
```

#### 告警触发条件

1. **发票生成失败时立即告警**
   - 触发时机：`SavePosInvoice` 返回错误时
   - 告警频率：24小时内最多2次

2. **发票长时间未生成告警**
   - 触发条件：有 `AsyncRecordId` 但超过1小时未生成
   - 检查频率：每30分钟检查一次

3. **重试次数过多告警**
   - 触发条件：重试次数超过3次
   - 检查频率：每次重试后检查

#### 告警内容

告警邮件应包含：
- 公司名称
- 订单号
- 订单时间
- 订单金额
- 错误信息
- ERP异步记录ID（如果有）
- 处理建议

#### 告警处理流程

当收到告警时，按以下流程处理：

1. **确认告警**
   - 查看告警详情
   - 确认是否为真实问题

2. **分析原因**
   - 查看错误日志
   - 分析失败原因
   - 检查 ERP 系统状态

3. **采取行动**
   - 根据失败原因采取相应措施
   - 手动重试或修复数据
   - 记录处理过程

4. **验证修复**
   - 验证问题是否已解决
   - 确认告警是否消除
   - 更新告警状态

## 相关文档

- [ERP 集成文档](../integrations/erpnext/)
- [异步处理机制](../../human/architecture/async-processing.md)
- [错误处理规范](../../.cursor/rules/go-main.mdc)

