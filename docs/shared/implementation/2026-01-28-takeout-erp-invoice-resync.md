# 外卖订单变更时 ERP 发票自动同步实现

> **版本**: v2.15.0  
> **日期**: 2026-01-28  
> **分支**: feature/takeout-lineman-02  
> **需求**: LINEMAN-ERP 的发票信息跟随店内订单变更自动更新

---

## 📋 任务目标

当外卖订单（特别是 LINEMAN）发生变更时（商品增加、减少、修改），自动重新同步 ERP 发票信息，确保 ERP 中的发票数据与最新订单状态保持一致。

---

## 🎯 实现方案

### 核心思路

1. **监听订单更新事件**：在 `OrderUpdatedEvent` 处理器中，当检测到菜品变动时触发
2. **检查同步条件**：验证订单是否已同步到 ERP、是否有班次信息等
3. **重新同步发票**：采用"先取消，再创建"策略：
   - 调用 `SyncOrderCancelledToERP` 取消旧发票
   - 清空订单的 `erp_pos_invoice_resp` 字段
   - 调用 `SyncOrderToERP` 创建新发票
4. **异步执行**：避免阻塞订单更新的主流程

---

## 📁 修改的文件

### 1. ERP 同步领域服务

**文件**: `main/app/modules/takeout/domain/service/takeout_erp_sync_service.go`

**修改内容**:
- 在 `ITakeoutErpSyncService` 接口中添加 `ResyncOrderToERP` 方法
- 实现 `ResyncOrderToERP` 方法，采用"先取消，再创建"策略：
  - 检查 ERP 集成是否启用
  - 查询订单并检查是否已同步到 ERP
  - 记录旧发票信息（用于日志）
  - **Step 1**: 调用 `SyncOrderCancelledToERP` 取消旧发票
  - **Step 2**: 清空订单的 `erp_pos_invoice_resp` 字段
  - **Step 3**: 调用 `SyncOrderToERP` 创建新发票
  - 记录新发票信息到日志

**关键代码**:
```go
// ResyncOrderToERP 重新同步外卖订单到 ERP（订单变更后使用）
// 实现策略：
//  1. 调用 SyncOrderCancelledToERP 取消旧发票
//  2. 清空订单的 erp_pos_invoice_resp 字段
//  3. 调用 SyncOrderToERP 创建新发票
func (s *takeoutErpSyncService) ResyncOrderToERP(ctx appContext.Context, orderUuid uint64) error {
    // 1. 检查 ERP 集成是否启用
    // 2. 查询订单并检查是否已同步到 ERP
    // 3. 记录旧发票信息
    
    // Step 1: 取消旧发票
    if err := s.SyncOrderCancelledToERP(ctx, orderUuid); err != nil {
        return errors.WithMessage(err, "取消旧发票失败")
    }
    
    // Step 2: 清空订单的 erp_pos_invoice_resp 字段
    if err := takeoutOrderRepo.UpdateByMap(orderUuid, map[string]interface{}{
        "erp_pos_invoice_resp": "",
    }); err != nil {
        return errors.WithMessage(err, "清空订单 ERP 响应字段失败")
    }
    
    // Step 3: 创建新发票
    if err := s.SyncOrderToERP(ctx, orderUuid); err != nil {
        return errors.WithMessage(err, "创建新发票失败")
    }
    
    return nil
}
```

**设计优势**:
- ✅ **代码复用**：直接使用现有的取消和创建方法，避免重复代码
- ✅ **逻辑清晰**：分步执行，易于理解和维护
- ✅ **错误处理**：每一步都有独立的错误处理和日志记录
- ✅ **测试友好**：可以单独测试每个步骤

---

### 2. 订单更新事件处理器

**文件**: `main/app/event/takeout/takeout_order_updated_event_handler.go`

**修改内容**:
- 在 `Handle` 方法的 Step 4 后添加 Step 5：重新同步 ERP 发票
- 实现 `resyncErpInvoice` 方法，异步调用 ERP 同步服务

**关键代码**:
```go
// Step 5: 重新同步 ERP 发票（如果订单已同步到 ERP）
s.resyncErpInvoice(ctx, takeoutSrv, orderUpdatedEvent)

// resyncErpInvoice 重新同步 ERP 发票
func (s *takeoutOrderUpdatedEventSubscriber) resyncErpInvoice(
    ctx appContext.Context,
    takeoutSrv takeoutService.ITakeoutSrv,
    orderUpdatedEvent event.OrderUpdatedEvent,
) {
    // 异步执行，避免阻塞主流程
    utils.Go(func() {
        if err := takeoutSrv.ResyncOrderErpInvoice(ctx, orderUpdatedEvent.OrderUuid); err != nil {
            logger.Logger.Error("重新同步 ERP 发票失败", ...)
        }
    })
}
```

---

### 3. 外卖订单服务

**文件**: `main/app/service/takeout/takeout_order.go`

**修改内容**:
- 在 `ITakeoutOrderSrv` 接口中添加 `ResyncOrderErpInvoice` 方法
- 实现 `ResyncOrderErpInvoice` 方法，调用领域服务

**关键代码**:
```go
// ResyncOrderErpInvoice 重新同步外卖订单到 ERP（订单变更后使用）
func (s *takeoutSrv) ResyncOrderErpInvoice(ctx context.Context, orderUuid uint64) error {
    erpSyncService := domainService.NewTakeoutErpSyncService()
    return erpSyncService.ResyncOrderToERP(ctx, orderUuid)
}
```

---

## 🔄 执行流程

```
订单更新 Webhook
    ↓
应用层处理订单数据
    ↓
触发 OrderUpdatedEvent
    ↓
订单更新事件处理器 (Handle)
    ├─ Step 1: 打印退菜单（如果有退菜）
    ├─ Step 2: 打印送厨单（如果有新菜品）
    ├─ Step 3: 更新生产单
    ├─ Step 4: 重建订单库存和销量
    └─ Step 5: 重新同步 ERP 发票 ✨ NEW
         ↓
    resyncErpInvoice (异步)
         ↓
    ResyncOrderErpInvoice
         ↓
    ResyncOrderToERP (领域服务)
         ↓
    [检查条件]
         ├─ ERP 集成是否启用？
         └─ 订单是否已同步到 ERP？
         ↓
    [Step 1: 取消旧发票]
         ↓
    SyncOrderCancelledToERP
         ├─ 检查订单是否已同步
         ├─ 检查班次是否已交班
         ├─ 调用 ERP CancelPosInvoice
         └─ 取消商品发票和原料发票
         ↓
    [Step 2: 清空 ERP 响应字段]
         ↓
    UpdateByMap(erp_pos_invoice_resp: "")
         ↓
    [Step 3: 创建新发票]
         ↓
    SyncOrderToERP
         ├─ 查询订单完整信息
         ├─ 检查订单是否已同步（跳过）
         ├─ 检查班次信息
         ├─ 构建 POS Invoice 请求
         ├─ 调用 ERP SavePosInvoice
         └─ 更新订单 ERP 响应字段
         ↓
    [记录日志]
         └─ 记录新旧发票名称
```

---

## ✅ 关键检查点

### 同步条件检查

1. **ERP 集成启用**：`companySetting.ErpnextSiteCode != ""`
2. **订单已同步**：`takeoutOrder.IsErpInvoiceSynced()` 返回 `true`
3. **订单有班次**：由 `SyncOrderCancelledToERP` 和 `SyncOrderToERP` 各自检查
4. **班次未交班**：由 `SyncOrderCancelledToERP` 检查，确保可以取消发票

### 三步同步策略

#### Step 1: 取消旧发票
调用 `SyncOrderCancelledToERP` 方法，该方法会：
- 检查订单是否已同步到 ERP
- 检查订单是否有班次信息
- 检查班次是否已交班（已交班无法取消发票）
- 调用 ERP `CancelPosInvoice` 接口取消旧发票

#### Step 2: 清空 ERP 响应字段
```go
takeoutOrderRepo.UpdateByMap(orderUuid, map[string]interface{}{
    "erp_pos_invoice_resp": "",
})
```
这一步很关键，因为 `SyncOrderToERP` 方法会检查订单是否已同步（通过 `IsErpInvoiceSynced()`），如果已同步会跳过。清空这个字段后，`SyncOrderToERP` 就会重新创建发票。

#### Step 3: 创建新发票
调用 `SyncOrderToERP` 方法，该方法会：
- 查询订单完整信息（商品、修饰符、原材料）
- 检查订单是否已同步（此时已清空，会继续执行）
- 检查订单是否有班次信息
- 构建新的 POS Invoice 请求
- 调用 ERP `SavePosInvoice` 接口创建新发票
- 更新订单的 `erp_pos_invoice_resp` 字段

---

## 🧪 测试建议

### 1. 正常流程测试

**场景**：LINEMAN 订单变更后 ERP 发票自动更新

**步骤**：
1. 创建 LINEMAN 外卖订单
2. 接单（触发首次 ERP 同步）
3. 修改订单菜品（通过 Webhook）：
   - 增加菜品数量
   - 减少菜品数量
   - 修改菜品属性（规格、加料）
   - 删除菜品
   - 新增菜品
4. 验证 ERP 发票是否自动更新

**预期结果**：
- 订单更新事件成功触发
- ERP 发票重新同步成功
- 日志中记录新旧发票名称
- 订单的 `erp_pos_invoice_resp` 字段更新

---

### 2. 边界条件测试

#### 测试用例 2.1：订单未同步到 ERP

**场景**：订单变更时，订单尚未同步到 ERP

**预期结果**：
- 检测到 `erp_pos_invoice_resp` 为空
- 跳过重新同步
- 日志记录："外卖订单未同步到 ERP，跳过重新同步"

---

#### 测试用例 2.2：订单没有班次信息

**场景**：订单变更时，订单没有班次信息

**预期结果**：
- 检测到 `staff_shift_log_uuid == 0`
- 跳过重新同步
- 日志记录："外卖订单没有班次信息，跳过重新同步"

---

#### 测试用例 2.3：ERP 集成未启用

**场景**：公司未启用 ERP 集成

**预期结果**：
- 检测到 `ErpnextSiteCode` 为空
- 跳过重新同步
- 日志记录："公司未启用 ERP 集成，跳过重新同步"

---

### 3. 异常处理测试

#### 测试用例 3.1：ERP 服务不可用

**场景**：ERP 服务宕机或网络不可达

**预期结果**：
- `SavePosInvoice` 调用失败
- 记录错误日志（包含错误信息）
- **不影响订单更新主流程**（异步执行）
- 订单状态正常，库存、销量、生产单都已更新

---

#### 测试用例 3.2：旧发票信息解析失败

**场景**：`erp_pos_invoice_resp` 字段数据损坏

**预期结果**：
- `GetErpPosInvoiceResp()` 返回 `nil`
- 跳过重新同步
- 日志记录："无法解析旧的 ERP 发票响应数据，跳过重新同步"

---

### 4. 性能测试

#### 测试用例 4.1：高并发订单变更

**场景**：短时间内多个订单同时变更

**关注点**：
- 异步执行是否正常工作
- 数据库连接是否稳定
- ERP RPC 调用是否有超时

---

## 📊 日志示例

### 成功同步

```
INFO  成功重新同步外卖订单到 ERP
  orderUuid: 123456789
  platformOrderId: LINEMAN-ABC123
  oldProductsInvoice: SINV-2024-00001
  oldMaterialInvoice: SRET-2024-00001
  newProductsInvoice: SINV-2024-00002
  newMaterialInvoice: SRET-2024-00002
```

### 跳过同步

```
INFO  外卖订单未同步到 ERP，跳过重新同步
  orderUuid: 123456789
  platformOrderId: LINEMAN-ABC123
```

### 同步失败

```
ERROR  重新同步外卖订单到 ERP 失败
  orderUuid: 123456789
  platformOrderId: LINEMAN-ABC123
  oldProductsInvoice: SINV-2024-00001
  oldMaterialInvoice: SRET-2024-00001
  error: "连接超时"
```

---

## 🔍 排查指南

### 问题 1：订单变更后 ERP 发票没有更新

**排查步骤**：
1. 检查日志，确认是否触发了 `OrderUpdatedEvent`
2. 检查订单是否已同步到 ERP（`erp_pos_invoice_resp` 字段）
3. 检查订单是否有班次信息（`staff_shift_log_uuid` 字段）
4. 检查 ERP 集成是否启用（公司设置中的 `ErpnextSiteCode`）
5. 检查 ERP RPC 调用是否成功（查看错误日志）

---

### 问题 2：重新同步报错 "查询 ERP 开账名称失败"

**原因**：订单的班次信息中没有 ERP 开账名称

**解决方案**：
1. 检查班次记录（`ttpos_staff_shift_log` 表）
2. 确认 `erpnext_open_pos_entry_name` 字段有值
3. 如果字段为空，可能是班次开启时 ERP 集成未启用

---

### 问题 3：重新同步报错 "支付方式不存在"

**原因**：系统中未配置 LINEMAN 支付方式

**解决方案**：
1. 检查支付方式配置（`ttpos_payment_method` 表）
2. 确认 `code = 10`（LINEMAN）的支付方式存在
3. 确认支付方式的 `erpnext_payment` 字段有值

---

## 📝 后续优化建议

1. **重试机制**：如果 ERP 同步失败，可以考虑添加重试机制
   - 特别是 Step 1（取消）和 Step 3（创建）之间，如果 Step 2 失败，需要能够回滚
2. **队列处理**：使用消息队列（如 RabbitMQ）来处理 ERP 同步任务
3. **监控告警**：添加 ERP 同步失败的监控和告警
   - 特别关注"取消成功但创建失败"的情况
4. **批量同步**：提供命令行工具，批量重新同步历史订单的 ERP 发票
5. **事务一致性**：考虑将三步操作封装在一个分布式事务中
   - 或者使用补偿事务（Saga 模式）确保最终一致性

---

## 📚 相关文档

- [ERP 集成文档](../integrations/erpnext.md)
- [外卖订单更新流程](../guides/takeout-order-update.md)
- [事件驱动架构](../../human/architecture/event-driven.md)

---

## ✍️ 更新记录

| 日期 | 版本 | 修改内容 | 修改人 |
| --- | --- | --- | --- |
| 2026-01-28 | v1.0 | 初始版本，实现外卖订单变更时 ERP 发票自动同步 | weifashi |
| 2026-01-28 | v1.1 | 重构 ResyncOrderToERP 方法，采用"先取消，再创建"策略，复用现有代码 | weifashi |
