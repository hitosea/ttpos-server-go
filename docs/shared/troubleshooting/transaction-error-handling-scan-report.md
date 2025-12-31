# 事务错误处理扫描报告

**扫描时间**: 2025-12-19  
**最后重复检查时间**: 2025-12-19  
**扫描范围**: `main/app/service` 目录  
**扫描目标**: 事务中未处理错误的情况

## 统计

- 扫描文件数: 40 个文件（所有包含事务的文件）
- 发现事务块数: 200+ 个事务块
- 发现问题数: 16 个（已修复 16 个）
- **重复检查结果**: ✅ 未发现新问题

## 问题分类

### 严重程度说明

- **严重**: 事务中使用错误的数据库对象，可能导致数据不一致
- **高**: 事务中未处理错误，可能导致事务失败但无错误提示
- **中**: 调用服务方法前未设置事务上下文，可能导致事务失效
- **低**: 标准库函数未处理错误（在事务中）

## 问题列表

### main/app/service/order_member.go

#### [严重] 事务中使用外部 db 而非 tx - 行号 100-110 ✅ 已修复

**问题描述**:  
事务回调函数的参数是 `tx`，但代码中使用的是外部 `db`，导致操作不在事务中执行。

**修复状态**: ✅ 已修复 - 第102行和第106行已改为使用 `tx`

**代码片段**:
```go
if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // 更新销售订单
    if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
        return errors.WithMessage(err)
    }
    // 设置sort排序
    if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderSort(memberSaleOrderUuid, constant.MemberSaleOrderSortDefault); err != nil {
        return errors.WithMessage(err)
    }
    return nil
}); err != nil {
    return errors.WithMessage(err)
}
```

**问题分析**:  
- 第 100 行：事务回调参数是 `tx`
- 第 102 行和第 106 行：使用外部 `db` 而非 `tx`

**建议修复方案**:  
将第 102 行和第 106 行的 `db` 改为 `tx`：
```go
if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderRecord(*saleOrder); err != nil {
    return errors.WithMessage(err)
}
if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderSort(memberSaleOrderUuid, constant.MemberSaleOrderSortDefault); err != nil {
    return errors.WithMessage(err)
}
```

---

#### [高] 事务调用未检查返回错误 - 行号 1762-1765 ✅ 已修复

**问题描述**:  
事务调用没有检查返回的错误，如果事务失败，错误会被忽略。

**修复状态**: ✅ 已修复 - 已添加错误检查

**代码片段**:
```go
// 更新订单状态
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderReject(*memberSaleOrder); err != nil {
        return errors.WithMessage(err)
    }
    return nil
})
```

**问题分析**:  
- 第 1764 行：事务调用没有赋值给变量，未检查返回错误

**建议修复方案**:  
添加错误检查：
```go
if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderReject(*memberSaleOrder); err != nil {
        return errors.WithMessage(err)
    }
    return nil
}); err != nil {
    return errors.WithMessage(err)
}
```

---

#### [中] returnInventory 调用前未设置事务上下文 - 行号 984-988 ✅ 已修复

**问题描述**:  
在事务中调用 `returnInventory` 方法，但未设置事务上下文，可能导致方法内部使用错误的数据库连接。

**修复状态**: ✅ 已修复 - 已添加 `ctxCopy.SetDB(tx)` 设置事务上下文

**代码片段**:
```go
err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // 退回商品库存
    if err := s.returnInventory(ctx.Copy(), billInfo); err != nil {
        return errors.WithMessage(err)
    }
    // ...
})
```

**问题分析**:  
- 第 984 行：事务回调参数是 `tx`
- 第 986 行：调用 `returnInventory` 前未设置事务上下文

**建议修复方案**:  
在调用 `returnInventory` 前设置事务上下文：
```go
err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    ctxCopy := ctx.Copy()
    ctxCopy.SetDB(tx) // 确保 returnInventory 使用事务
    if err := s.returnInventory(ctxCopy, billInfo); err != nil {
        return errors.WithMessage(err)
    }
    // ...
})
```

---

### main/app/service/order_pay.go

#### [严重] 事务中使用外部 db 而非 tx - 行号 107-126 ✅ 已修复

**问题描述**:  
事务回调函数的参数是 `tx`，但代码中使用的是外部 `db`，导致操作不在事务中执行。同时事务调用未检查返回错误。

**修复状态**: ✅ 已修复 - 已改为使用 `tx`，并添加错误检查

**代码片段**:
```go
db.Transaction(func(tx *gorm.DB) error {
    // 选择优惠券后，将积分自动抵扣失效改为手动抵扣
    saleOrder.AutoPointsExchange = 0

    if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
        return errors.WithMessage(err)
    }
    // ...
    return nil
})
```

**问题分析**:  
- 第 107 行：事务调用没有赋值给变量，未检查返回错误
- 第 111 行：使用外部 `db` 而非事务 `tx`
- `CalcAndSaveSaleBill` 方法需要传入数据库连接，应该使用 `tx` 而非 `db`

**建议修复方案**:  
```go
if err := db.Transaction(func(tx *gorm.DB) error {
    saleOrder.AutoPointsExchange = 0

    if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
        return errors.WithMessage(err)
    }
    // ...
    return nil
}); err != nil {
    return nil, errors.WithMessage(err)
}
```

---

#### [高] 调用方法未检查返回错误 - 行号 828-829 ✅ 已修复

**问题描述**:  
调用 `IncConsumptionAmount` 和 `IncConsumptionCount` 方法但未检查返回错误。

**修复状态**: ✅ 已修复 - 已添加错误检查

**代码片段**:
```go
// 更新会员消费金额和消费次数
consumptionAmount := decimal.NewFromFloat(saleOrder.GetAmountValue()).Sub(decimal.NewFromFloat(saleOrder.ZeroCheckoutFee)).Truncate(2).InexactFloat64()
repository.NewMemberRepo(db).IncConsumptionAmount(saleOrder.ConsumerUuid, consumptionAmount)
repository.NewMemberRepo(db).IncConsumptionCount(saleOrder.ConsumerUuid)
```

**问题分析**:  
- 第 828 行：`IncConsumptionAmount` 返回 error 但未检查
- 第 829 行：`IncConsumptionCount` 返回 error 但未检查
- 这些操作在事务外执行，如果失败会导致数据不一致

**建议修复方案**:  
添加错误检查：
```go
if err := repository.NewMemberRepo(db).IncConsumptionAmount(saleOrder.ConsumerUuid, consumptionAmount); err != nil {
    return nil, errors.WithMessage(err)
}
if err := repository.NewMemberRepo(db).IncConsumptionCount(saleOrder.ConsumerUuid); err != nil {
    return nil, errors.WithMessage(err)
}
```

---

### main/app/service/order_product.go

#### [严重] 事务中使用外部 db 而非 tx - 行号 1097-1098 ✅ 已修复

**问题描述**:  
事务回调函数的参数是 `tx`，但代码中使用的是外部 `db`，导致操作不在事务中执行。

**修复状态**: ✅ 已修复 - 已改为使用 `tx`

**代码片段**:
```go
errUpdate := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductReasons(
        saleOrderProduct.SaleOrderUuid,
        saleOrderProduct.Uuid,
        constant.ProductReasonTypeReturnFood,
    ); err != nil {
        return errors.WithMessage(err)
    }
    // ...
})
```

**问题分析**:  
- 第 1097 行：事务回调参数是 `tx`
- 第 1098 行：使用外部 `db` 而非 `tx`

**建议修复方案**:  
将第 1098 行的 `db` 改为 `tx`：
```go
errUpdate := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    if err := repository.NewSaleOrderProductRepo(tx).DeleteSaleOrderProductReasons(
        saleOrderProduct.SaleOrderUuid,
        saleOrderProduct.Uuid,
        constant.ProductReasonTypeReturnFood,
    ); err != nil {
        return errors.WithMessage(err)
    }
    // ...
})
```

---

### main/app/service/order_base.go

#### [严重] 事务中使用外部 db 而非 tx - 行号 215 ✅ 已修复

**问题描述**:  
事务回调函数的参数是 `tx`，但代码中使用的是外部 `db`，导致操作不在事务中执行。

**修复状态**: ✅ 已修复 - 已改为使用 `tx`

**代码片段**:
```go
if err := db.Transaction(func(tx *gorm.DB) error {
    // ...
    // 创建销售账单设置
    if _, errCreateSaleBillSetting := repository.NewOrderRepo(db).CreateSaleBillSetting(*saleBillSetting); errCreateSaleBillSetting != nil {
        return errCreateSaleBillSetting
    }
    // ...
})
```

**问题分析**:  
- 第 183 行：事务回调参数是 `tx`
- 第 215 行：使用外部 `db` 而非 `tx`

**建议修复方案**:  
将第 215 行的 `db` 改为 `tx`：
```go
if _, errCreateSaleBillSetting := repository.NewOrderRepo(tx).CreateSaleBillSetting(*saleBillSetting); errCreateSaleBillSetting != nil {
    return errCreateSaleBillSetting
}
```

---

#### [严重] 事务中使用外部 db 而非 tx - 行号 954 ✅ 已修复

**问题描述**:  
事务回调函数的参数是 `tx`，但代码中使用的是外部 `db`，导致操作不在事务中执行。

**修复状态**: ✅ 已修复 - 已改为使用 `tx`

**代码片段**:
```go
if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // ...
    // 更新账单
    if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
        return errUpdateSaleBill
    }
    // ...
})
```

**问题分析**:  
- 第 929 行：事务回调参数是 `tx`
- 第 954 行：使用外部 `db` 而非 `tx`

**建议修复方案**:  
将第 954 行的 `db` 改为 `tx`：
```go
if errUpdateSaleBill := repository.NewSaleBillRepo(tx).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
    return errUpdateSaleBill
}
```

---

### main/app/service/order_manage.go

#### [中] goroutine 中调用方法未检查返回错误 - 行号 1222-1224 ✅ 已修复

**问题描述**:  
在 goroutine 中调用 `UpdateReturnOrderAmount` 方法但未检查返回错误。

**修复状态**: ✅ 已修复 - 已检查 `UpdateReturnOrderAmount` 的返回错误

**代码片段**:
```go
utils.Go(func() {
    payment, err := NewPaymentRepo(ctx, s.dbm).Refund(paymentServiceRefundReq)
    if err != nil {
        returnOrderAmount.RefundStatus = 2
        returnOrderAmount.LlReturnOrderid = "0"
    } else {
        returnOrderAmount.LlReturnOrderid = payment.RefundOrderId
    }
    // 更新退款状态
    returnOrderRepo := repository.NewReturnOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
    returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{
        returnOrderRepo.WhereUuid(returnOrderAmount.Uuid),
    }, returnOrderAmount)
    if err != nil { // 这里检查的是外部的 err，而不是 UpdateReturnOrderAmount 的返回值
        fmt.Println("更新退款状态失败", err)
        logger.Logger.Error("更新退款状态失败", zap.Error(err))
    }
})
```

**问题分析**:  
- 第 1222 行：调用 `UpdateReturnOrderAmount` 但未检查返回错误
- 第 1225 行：检查的是外部的 `err`（来自 `Refund`），而不是 `UpdateReturnOrderAmount` 的返回值

**建议修复方案**:  
检查 `UpdateReturnOrderAmount` 的返回错误：
```go
if err := returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{
    returnOrderRepo.WhereUuid(returnOrderAmount.Uuid),
}, returnOrderAmount); err != nil {
    logger.Logger.Error("更新退款状态失败", zap.Error(err))
}
```

---

### 问题统计

- **严重问题**: 7 个（事务中使用外部 db 而非 tx、事务中调用的方法未使用事务上下文）
  - ✅ `order_member.go:100-110` - 已修复
  - ✅ `order_pay.go:107-126` - 已修复
  - ✅ `order_product.go:1097-1098` - 已修复
  - ✅ `order_base.go:215` - 已修复
  - ✅ `order_base.go:954` - 已修复
  - ✅ `order_product.go:1274` - 已修复
- **中优先级问题**: 3 个（调用服务方法前未设置事务上下文、goroutine 中调用方法未检查返回错误、事务中使用外部 db）
  - ✅ `order_member.go:984-988` - 已修复
  - ✅ `order_manage.go:1222-1224` - 已修复
  - ✅ `transfer_order.go:687, 1482` - 已修复
- **高优先级问题**: 7 个（事务调用未检查返回错误、调用方法未检查返回错误）
  - ✅ `order_member.go:1762-1765` - 已修复
  - ✅ `order_pay.go:828-829` - 已修复
  - ✅ `order_product.go:103` - 已修复
  - ✅ `product.go:4506` - 已修复
  - ✅ `printer.go:835` - 已修复
  - ✅ `order.go:1307` - 已修复
  - ✅ `order_buffet.go:120` - 已修复
- **中优先级问题**: 2 个（调用服务方法前未设置事务上下文、goroutine 中调用方法未检查返回错误）
  - ✅ `order_member.go:984-988` - 已修复
  - ✅ `order_manage.go:1222-1224` - 已修复

### 修复状态

已修复 16 个问题：

**已修复的问题**：
1. ✅ `order_member.go:100-110` - 事务中使用外部 `db` 的问题已修复，现在使用 `tx`
2. ✅ `order_member.go:1762-1765` - 事务调用未检查错误的问题已修复，已添加错误检查
3. ✅ `order_member.go:984-988` - `returnInventory` 调用前未设置事务上下文的问题已修复，已添加 `ctxCopy.SetDB(tx)`
4. ✅ `order_pay.go:107-126` - 事务中使用外部 `db` 的问题已修复，已改为使用 `tx`，并添加错误检查
5. ✅ `order_pay.go:828-829` - 调用方法未检查返回错误的问题已修复，已添加错误检查
6. ✅ `order_product.go:1097-1098` - 事务中使用外部 `db` 的问题已修复，已改为使用 `tx`
7. ✅ `order_base.go:215` 和 `order_base.go:954` - 事务中使用外部 `db` 的问题已修复，已改为使用 `tx`
8. ✅ `order_manage.go:1222-1224` - goroutine 中调用方法未检查返回错误的问题已修复，已检查 `UpdateReturnOrderAmount` 的返回错误
9. ✅ `order_product.go:103` - 事务调用未检查返回错误的问题已修复，已添加错误检查
10. ✅ `product.go:4506` - 事务调用未检查返回错误的问题已修复，已添加错误检查
11. ✅ `printer.go:835` - 事务调用未检查返回错误的问题已修复，已添加错误检查
12. ✅ `order_product.go:1274` - 事务中使用外部 `db` 的问题已修复，已改为使用 `tx`，并设置事务上下文
13. ✅ `order.go:1307` - 调用方法未检查返回错误的问题已修复，已添加 `RejectH5Order` 的错误检查
14. ✅ `transfer_order.go:687, 1482` - 事务中使用外部 `db` 的问题已修复，已改为使用 `tx`
15. ✅ `order_buffet.go:120` - 调用方法未检查返回错误的问题已修复，已添加 `DeleteSaleOrderBuffetCustomerType` 的错误检查

**待修复的问题**：
无（所有问题已修复）

---

### main/app/service/order_product.go

#### [高] 事务调用未检查返回错误 - 行号 103 ✅ 已修复

**问题描述**:  
事务调用没有检查返回的错误，如果事务失败，错误会被忽略。

**修复状态**: ✅ 已修复 - 已添加错误检查

**代码片段**:
```go
// 更新订单商品备注
repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
    // ... 事务操作
    return nil
})
// 没有检查返回错误
```

**问题分析**:  
- 第 103 行：事务调用没有赋值给变量，未检查返回错误
- `OrderProductRemark` 函数返回 `(*resp.ShopCart, error)`，如果事务失败，函数会继续执行并返回成功

**建议修复方案**:  
添加错误检查：
```go
if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
    // ... 事务操作
    return nil
}); err != nil {
    return nil, errors.WithMessage(err)
}
```

---

### main/app/service/product.go

#### [高] 事务调用未检查返回错误 - 行号 4506 ✅ 已修复

**问题描述**:  
事务调用没有检查返回的错误，如果事务失败，错误会被忽略。

**修复状态**: ✅ 已修复 - 已添加错误检查

**代码片段**:
```go
db.Transaction(func(tx *gorm.DB) error {
    // 软删除商品规格
    // ... 事务操作
    return nil
})
// 没有检查返回错误
```

**问题分析**:  
- 第 4506 行：事务调用没有赋值给变量，未检查返回错误
- `DeleteProductFlavor` 函数返回 `error`，如果事务失败，函数会返回 nil 而不是错误

**建议修复方案**:  
添加错误检查：
```go
if err := db.Transaction(func(tx *gorm.DB) error {
    // ... 事务操作
    return nil
}); err != nil {
    return errors.WithMessage(err)
}
```

---

### main/app/service/printer.go

#### [高] 事务调用未检查返回错误 - 行号 835 ✅ 已修复

**问题描述**:  
事务调用没有检查返回的错误，如果事务失败，错误会被忽略。

**修复状态**: ✅ 已修复 - 已添加错误检查

**代码片段**:
```go
// 更新打印机定制
db.Transaction(func(tx *gorm.DB) error {
    err = repository.NewPrinterCustomizeRepo(tx).UpdatePrinterCustomize(customizeInfo)
    // ... 事务操作
    return nil
})
// 没有检查返回错误
```

**问题分析**:  
- 第 835 行：事务调用没有赋值给变量，未检查返回错误
- `EditPrinterCustomize` 函数返回 `error`，如果事务失败，函数会返回 nil 而不是错误

**建议修复方案**:  
添加错误检查：
```go
if err := db.Transaction(func(tx *gorm.DB) error {
    // ... 事务操作
    return nil
}); err != nil {
    return errors.WithMessage(err)
}
```

---

### main/app/service/order_product.go

#### [严重] 事务中使用外部 db 而非 tx - 行号 1274 ✅ 已修复

**问题描述**:  
在事务回调函数中调用 `deleteOrRejectH5OrderProduct` 时传入的是外部 `db` 而不是事务 `tx`，导致操作不在事务中执行。

**修复状态**: ✅ 已修复 - 已改为使用 `tx`，并设置事务上下文

**代码片段**:
```go
if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // ...
    if saleOrderProduct.H5OrderUuid != 0 {
        s.deleteOrRejectH5OrderProduct(ctx, db, saleOrderProduct) // 错误：使用 db 而非 tx
    }
    // ...
})
```

**问题分析**:  
- 第 1274 行：在事务回调函数中调用 `deleteOrRejectH5OrderProduct` 时传入的是外部 `db` 而不是事务 `tx`
- 这会导致 `deleteOrRejectH5OrderProduct` 内部的数据库操作不在事务中执行，可能导致数据不一致

**建议修复方案**:  
使用事务 `tx` 并设置事务上下文：
```go
if saleOrderProduct.H5OrderUuid != 0 {
    ctxCopy := ctx.Copy()
    ctxCopy.SetDB(tx)
    if err := s.deleteOrRejectH5OrderProduct(ctxCopy, tx, saleOrderProduct); err != nil {
        return errors.WithMessage(err)
    }
}
```

---

### main/app/service/order.go

#### [高] 调用方法未检查返回错误 - 行号 1307 ✅ 已修复

**问题描述**:  
`deleteOrRejectH5OrderProduct` 方法中调用 `RejectH5Order` 未检查返回错误。

**修复状态**: ✅ 已修复 - 已添加错误检查

**代码片段**:
```go
func (s *orderSrv) deleteOrRejectH5OrderProduct(ctx context.Context, db *gorm.DB, saleOrderProduct *model.SaleOrderProduct) error {
    if saleOrderProduct.H5OrderUuid != 0 {
        h5OrderProductCount, err := repository.NewH5OrderRepo(db).GetH5OrderProductCount(saleOrderProduct.H5OrderUuid)
        if err != nil {
            return errors.WithMessage(err)
        }
        if h5OrderProductCount == 1 {
            s.RejectH5Order(ctx, saleOrderProduct.H5OrderUuid) // 错误：未检查返回错误
        } else {
            // ...
        }
    }
    return nil
}
```

**问题分析**:  
- 第 1307 行：调用 `RejectH5Order` 方法但未检查返回错误
- 如果 `RejectH5Order` 失败，错误会被忽略，函数会返回 `nil` 而不是错误

**建议修复方案**:  
添加错误检查：
```go
if h5OrderProductCount == 1 {
    if err := s.RejectH5Order(ctx, saleOrderProduct.H5OrderUuid); err != nil {
        return errors.WithMessage(err)
    }
}
```

---

### main/app/service/transfer_order/transfer_order.go

#### [中] 事务中使用外部 db 而非 tx - 行号 687, 1482 ✅ 已修复

**问题描述**:  
在事务回调函数中调用 `CreateLog` 时传入的是外部 `db` 而不是事务 `tx`。

**修复状态**: ✅ 已修复 - 已改为使用 `tx`

**代码片段**:
```go
// transfer_order.go:687
err = db.Transaction(func(tx *gorm.DB) error {
    // ...
    // 记录操作日志
    if err := s.helper.CreateLog(ctx, db, transferOrder.Uuid, constant.TransferActionCreate, "创建调拨单", 0, constant.TransferOrderStatusDraft); err != nil {
        logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
    }
    return nil
})

// transfer_order.go:1482
err = db.Transaction(func(tx *gorm.DB) error {
    // ...
    // 记录操作日志
    if err := s.helper.CreateLog(ctx, db, req.Uuid, constant.TransferActionReceive, "收货完成", transferOrder.Status, constant.TransferOrderStatusCompleted); err != nil {
        logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
    }
    return nil
})
```

**问题分析**:  
- 第 687 行和第 1482 行：在事务回调函数中调用 `CreateLog` 时传入的是外部 `db` 而不是事务 `tx`
- 这会导致日志记录不在事务中执行
- 如果事务回滚，日志可能已经记录，导致数据不一致
- 但日志记录失败不应该影响主业务逻辑，所以错误只记录到日志而不返回

**建议修复方案**:  
如果日志记录需要在事务中执行（以便在事务回滚时也回滚日志），应该使用 `tx`：
```go
if err := s.helper.CreateLog(ctx, tx, transferOrder.Uuid, ...); err != nil {
    logger.Logger.Error("记录调拨单日志失败", zap.Error(err))
    // 注意：日志记录失败不应该导致事务失败
}
```

**修复说明**:  
已修复，将 `db` 改为 `tx`，确保日志记录在事务中执行。如果事务回滚，日志记录也会回滚，保持数据一致性。日志记录失败仍然只记录到日志而不返回错误，不影响主业务逻辑。

---

### main/app/service/order_buffet.go

#### [高] 调用方法未检查返回错误 - 行号 120 ✅ 已修复

**问题描述**:  
在事务中调用 `DeleteSaleOrderBuffetCustomerType` 方法但未检查返回错误。

**修复状态**: ✅ 已修复 - 已添加错误检查

**代码片段**:
```go
if err := db.Transaction(func(tx *gorm.DB) error {
    // 删除原来的 CustomerType
    repository.NewOrderRepo(tx).DeleteSaleOrderBuffetCustomerType(saleOrder.Uuid) // 错误：未检查返回错误
    saleBill.DeleteSaleOrderBuffetCustomerTypeAll(saleOrder.Uuid)
    // ...
})
```

**问题分析**:  
- 第 120 行：调用 `DeleteSaleOrderBuffetCustomerType` 方法但未检查返回错误
- 如果删除失败，错误会被忽略，函数会继续执行，可能导致数据不一致

**建议修复方案**:  
添加错误检查：
```go
if err := repository.NewOrderRepo(tx).DeleteSaleOrderBuffetCustomerType(saleOrder.Uuid); err != nil {
    return errors.WithMessage(err)
}
```

---

## 重复检查记录

**检查时间**: 2025-12-19 10:26  
**检查范围**: 所有包含事务的文件（40 个）  
**检查方法**: 
- 使用 grep 搜索所有事务调用模式（`repository.CommonRepo.Transaction`, `db.Transaction`, `db.Begin()`, `tx.Begin()`）
- 使用语义搜索查找潜在问题（事务中使用外部 db、未检查错误、未设置上下文等）
- 逐文件检查事务块

**检查结果**: ✅ **未发现新问题**

**检查说明**:
- ✅ 所有使用 `db.Begin()` 的地方都正确（GORM v2 的 `Begin()` 不返回错误，这是正常用法）
- ✅ 所有事务回调函数都正确使用了 `tx` 而非外部 `db`
- ✅ 所有事务调用都正确检查了返回错误
- ✅ 所有服务方法调用前都正确设置了事务上下文（`ctx.SetDB(tx)` 或 `ctx.Copy().SetDB(tx)`）
- ✅ 所有 goroutine 中的方法调用都正确检查了返回错误

**结论**: 经过重复检查，确认所有之前发现的问题都已修复，未发现新的问题。

---