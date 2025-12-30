# 外卖订单班次自动关联功能

> 记录时间：2025-12-30  
> 功能：外卖订单在接单/呼叫骑手时自动绑定班次

---

## 功能说明

外卖订单根据接单方式不同，在不同时机绑定班次：

### 业务场景

#### 1. 手动接单订单
- **触发时机**：点击"接单"按钮时
- **执行操作**：
  - 绑定当前操作员工的班次 UUID
  - 设置 `accepted_by` 为当前员工
  - 生成 ERP 发票号（后续开发）

#### 2. 自动接单订单
- **触发时机**：点击"呼叫骑手"按钮时
- **执行操作**：
  - 绑定当前操作员工的班次 UUID
  - 设置 `accepted_by` 为当前员工
  - 生成 ERP 发票号（后续开发）

### 设计原则

1. **分场景处理**：手动接单和自动接单在不同时机绑定班次
2. **容错设计**：班次绑定失败不影响主流程，只记录错误日志
3. **领域层实现**：业务逻辑封装在领域层，保持代码整洁
4. **幂等性**：只对 `staff_shift_log_uuid = 0` 的订单设置班次

---

## 代码实现

### 1. 手动接单 - AcceptOrder

**文件**：`main/app/modules/takeout/domain/service/takeout_order_service.go`

**位置**：第 276-340 行

```go
// AcceptOrder 接单
func (s *takeoutOrderSrv) AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error {
    db := ctx.GetDB()
    currentTime := time.Now().Unix()
    userUuid := ctx.GetStaffUuid()

    // 查询订单
    orderRepo := persistence.NewTakeoutOrderRepo(db)
    order, err := orderRepo.GetByUuid(req.Uuid)
    if err != nil {
        logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", req.Uuid))
        return errors.WithMessage(errors.New("查询订单失败"), err.Error())
    }
    
    // 检查订单状态
    if err := order.IsPendingOrder(); err != nil {
        return err
    }

    // 如果不是自动接单，则通知平台接受订单
    if !order.IsAutoAcceptOrder() {
        // 调用 BMP RPC 通知平台接受订单
        // ...
    }

    // ✅ 设置员工班次信息（手动接单在此处绑定班次）
    order.AcceptedBy = userUuid
    if err := orderRepo.SetStaffShiftLogUuid(order, userUuid); err != nil {
        logger.Logger.Error("设置员工班次日志UUID失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
    }

    updateData := map[string]interface{}{
        "order_state":          valueobject.TakeoutOrderStateAccepted,
        "accepted_time":        currentTime,
        "staff_shift_log_uuid": order.StaffShiftLogUuid,
        "accepted_by":          order.AcceptedBy,
        "update_time":          currentTime,
    }
    
    // 更新订单
    if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
        logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
        return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
    }

    // 发布订单接受事件
    event.GetDispatcher().Publish(event.NewOrderAcceptedEvent(
        order.Uuid,
        order.Platform,
        order.PlatformOrderId,
        order.ShortOrderNumber,
        order.TakeoutOrderUuid,
        userUuid,
        valueobject.TakeoutOrderAcceptedTypeManual,
        ctx.GetCompanyUuid(),
    ))

    return nil
}
```

### 2. 自动接单 - CallRider（呼叫骑手）

**文件**：`main/app/modules/takeout/domain/service/takeout_order_service.go`

**位置**：第 411-488 行

```go
// CallRider 呼叫骑手（标记订单准备完成）
func (s *takeoutOrderSrv) CallRider(ctx context.Context, req *request.TakeoutOrderCallRiderReq) error {
    db := ctx.GetDB()
    currentTime := time.Now().Unix()

    // 查询订单
    orderRepo := persistence.NewTakeoutOrderRepo(db)
    order, err := orderRepo.GetByUuid(req.Uuid)
    if err != nil {
        logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", req.Uuid))
        return errors.WithMessage(errors.New("查询订单失败"), err.Error())
    }
    if order == nil {
        return errors.New("订单不存在")
    }

    // 检查订单状态 - 只有已接单配餐中的订单才能呼叫骑手
    if order.OrderState != valueobject.TakeoutOrderStateAccepted {
        return errors.New("订单状态不正确，只有已接单配餐中的订单才能呼叫骑手")
    }

    // ✅ 自动接单的订单在呼叫骑手时设置班次
    userUuid := ctx.GetStaffUuid()
    if order.IsAutoAcceptOrder() && order.StaffShiftLogUuid == 0 {
        if err := orderRepo.SetStaffShiftLogUuid(order, userUuid); err != nil {
            logger.Logger.Error("自动接单订单呼叫骑手时设置班次失败",
                zap.Error(err),
                zap.Uint64("order_uuid", order.Uuid))
            // 不中断流程，只记录日志
        }
    }

    // 调用 BMP RPC 标记订单准备完成
    rpcClient, err := rpc.NewBMPTakeoutClient()
    if err != nil {
        logger.Logger.Error("创建 BMP RPC 客户端失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
        return errors.WithMessage(errors.New("创建 BMP RPC 客户端失败"), err.Error())
    }
    defer rpcClient.Close()

    // 调用 MarkOrderReady 接口
    if err := rpcClient.MarkOrderReady(ctx.GetContext(), order.TakeoutOrderUuid); err != nil {
        logger.Logger.Error("调用 BMP MarkOrderReady 接口失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
        return errors.WithMessage(errors.New("呼叫骑手失败"), err.Error())
    }

    // 更新订单状态为待骑手接单
    updateData := map[string]interface{}{
        "order_state": func() int {
            if order.Platform == valueobject.TakeoutPlatformGrab {
                if order.IsDineInOrder() {
                    return valueobject.TakeoutOrderStateCompleted
                }
                return grab.ConvertPlatformStateToOrderState(order.PlatformOrderState, valueobject.TakeoutOrderStateRiderPending)
            }
            return valueobject.TakeoutOrderStateRiderProcessing
        }(),
        "update_time": currentTime,
    }

    // ✅ 如果是自动接单的订单，且已设置班次，则同时更新班次信息
    if order.IsAutoAcceptOrder() && order.StaffShiftLogUuid > 0 {
        updateData["staff_shift_log_uuid"] = order.StaffShiftLogUuid
        updateData["accepted_by"] = order.AcceptedBy
        logger.Logger.Info("自动接单订单呼叫骑手时设置班次成功",
            zap.Uint64("order_uuid", order.Uuid),
            zap.Uint64("shift_log_uuid", order.StaffShiftLogUuid),
            zap.Uint64("accepted_by", order.AcceptedBy))
    }

    if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
        logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
        return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
    }

    // 发布订单呼叫骑手事件
    event.GetDispatcher().Publish(event.NewOrderReadyEvent(
        order.Uuid,
        order.Platform,
        order.PlatformOrderId,
        order.ShortOrderNumber,
        order.TakeoutOrderUuid,
        ctx.GetCompanyUuid(),
    ))

    return nil
}
```

### 3. 仓储层方法（已有）

**文件**：`main/app/modules/takeout/infrastructure/persistence/takeout_order_repo.go`

```go
// SetStaffShiftLogUuid 设置员工班次日志UUID
func (r *TakeoutOrderRepoImpl) SetStaffShiftLogUuid(order *model.TakeoutOrder, staffUuid uint64) error {
    if order.StaffShiftLogUuid != 0 {
        return nil
    }
    // 是否自动接单
    if order.IsAutoAcceptOrder() {
        return r.SetStaffShiftLogUuidByLatestShiftLog(order)
    }
    if staffUuid == 0 {
        return nil
    }
    // 查询员工当前班次
    var shiftLog struct {
        Uuid uint64 `gorm:"column:uuid"`
    }
    err := r.db.Table("ttpos_staff_shift_log").
        Select("uuid").
        Where("staff_uuid = ? AND status = ?", staffUuid, constant.StaffNotHandedOver).
        First(&shiftLog).Error
    if err != nil {
        // 如果没有找到班次记录，不报错，返回 nil
        if err == gorm.ErrRecordNotFound {
            return nil
        }
        return errors.WithMessage(err)
    }
    // 设置订单的班次UUID
    order.StaffShiftLogUuid = shiftLog.Uuid
    return nil
}

// SetStaffShiftLogUuidByLatestShiftLog 根据最新班次记录设置订单的班次UUID
func (r *TakeoutOrderRepoImpl) SetStaffShiftLogUuidByLatestShiftLog(order *model.TakeoutOrder) error {
    if order.StaffShiftLogUuid != 0 {
        return nil
    }
    // 查询最新班次记录
    var shiftLog struct {
        Uuid      uint64 `gorm:"column:uuid"`
        StaffUuid uint64 `gorm:"column:staff_uuid"`
    }
    err := r.db.Table("ttpos_staff_shift_log").
        Select("uuid, staff_uuid").
        Where("status = ?", constant.StaffNotHandedOver).
        Order("id ASC").
        First(&shiftLog).Error

    if err != nil {
        // 如果没有找到班次记录，不报错，返回 nil
        if err == gorm.ErrRecordNotFound {
            return nil
        }
        return errors.WithMessage(err)
    }
    order.AcceptedBy = shiftLog.StaffUuid
    order.StaffShiftLogUuid = shiftLog.Uuid
    return nil
}
```

---

## 执行流程

### 手动接单流程

```
用户点击"接单"
    ↓
查询订单
    ↓
检查订单状态（必须是待接单）
    ↓
通知外卖平台接单
    ↓
设置班次 UUID 和接单人
    ↓
更新订单状态为"已接单"
    ↓
发布订单接受事件
```

### 自动接单流程

```
平台自动接单
    ↓
订单状态变为"已接单"（staff_shift_log_uuid = 0）
    ↓
用户点击"呼叫骑手"
    ↓
查询订单
    ↓
检查订单状态（必须是已接单）
    ↓
✅ 设置班次 UUID 和接单人（首次设置）
    ↓
通知外卖平台标记订单准备完成
    ↓
更新订单状态为"待骑手接单"或"骑手配送中"
    ↓
同时更新班次信息到数据库
    ↓
发布订单呼叫骑手事件
```

---

## 数据库操作

### 手动接单更新语句

```sql
UPDATE ttpos_takeout_order
SET order_state = 2,           -- 已接单
    accepted_time = ?,
    staff_shift_log_uuid = ?,  -- 班次UUID
    accepted_by = ?,            -- 接单人
    update_time = ?
WHERE uuid = ?;
```

### 自动接单呼叫骑手更新语句

```sql
UPDATE ttpos_takeout_order
SET order_state = ?,           -- 待骑手接单/骑手配送中
    staff_shift_log_uuid = ?,  -- 班次UUID（首次设置）
    accepted_by = ?,            -- 接单人（首次设置）
    update_time = ?
WHERE uuid = ?;
```

---

## 关键判断逻辑

### 如何判断是否自动接单？

```go
func (o *TakeoutOrder) IsAutoAcceptOrder() bool {
    return o.OrderAcceptedType == valueobject.TakeoutOrderAcceptedTypeAuto
}
```

### 如何判断是否已设置班次？

```go
if order.StaffShiftLogUuid == 0 {
    // 未设置班次，需要设置
}
```

---

## 测试要点

### 手动接单场景

1. **正常接单**：验证班次UUID和接单人是否正确设置
2. **无当前班次**：验证是否正常处理（不报错，不设置班次）
3. **已有班次**：验证不会重复设置

### 自动接单场景

1. **呼叫骑手时首次设置**：验证班次UUID和接单人是否正确设置
2. **无当前班次**：验证是否正常处理（不报错，不设置班次）
3. **已有班次**：验证不会重复设置
4. **自动接单后手动点接单**：不应该触发呼叫骑手的班次设置逻辑

### 异常场景

1. **设置班次失败**：验证不影响主流程（接单/呼叫骑手成功）
2. **数据库异常**：验证错误日志记录

---

## 注意事项

1. **容错设计**：班次设置失败不影响接单或呼叫骑手的主流程
2. **幂等性**：只对 `staff_shift_log_uuid = 0` 的订单设置班次
3. **自动接单订单**：在呼叫骑手时才设置班次，而不是接单时
4. **手动接单订单**：在接单时立即设置班次
5. **日志记录**：成功和失败都有详细的日志记录

---

## 相关文档

- [外卖模块数据库架构](../../../main/app/modules/takeout/docs/database-architecture.md)
- [班次管理功能](../../human/guides/shift-management.md)

---

## 修改记录

| 日期 | 修改人 | 说明 |
|------|--------|------|
| 2025-12-30 | AI Assistant | 初始版本 |
| 2025-12-30 | AI Assistant | 修正流程：自动接单在呼叫骑手时绑定班次 |
