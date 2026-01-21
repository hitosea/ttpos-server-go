# Grab 外卖统计代码审查报告

**审查时间**: 2026-01-20  
**审查范围**: `main/app/service/statistics.go` 中的 `CountSaleDays` 和 `CountPaymentDays` 方法  
**审查重点**: 报错风险、需求符合度、开关判断、空数据影响

---

## 审查结果总览

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 报错风险 | ✅ 通过 | 代码逻辑安全，无报错风险 |
| Grab外卖统计符合需求 | ✅ 通过 | 统计逻辑正确，累加到正确字段 |
| Grab外卖开关判断 | ✅ 通过 | 开关判断正确，使用 `IsOpenGrabDelivery()` |
| 空数据不影响原有逻辑 | ✅ 通过 | 当没有 Grab 订单时，不影响原有统计 |

---

## 详细审查结果

### 1. 报错风险检查 ✅

#### 1.1 CountSaleDays 方法

**检查点**：
- ✅ 变量初始化：所有 Grab 相关变量都有默认值（零值）
- ✅ 条件判断：`enableGrabDelivery` 判断正确
- ✅ 空数据处理：当 `enableGrabDelivery` 为 false 或没有订单时，使用零值累加，不会报错
- ✅ 类型转换：使用 `decimal` 包进行精确计算，避免精度问题

**代码分析**：
```go
// 第305行：初始化 grabSaleData 为零值
var grabSaleData model.StatisticsTakeoutSaleData

// 第307-326行：只有在 enableGrabDelivery 为 true 时才查询
if enableGrabDelivery {
    grabSaleData = takeoutRepo.CountTakeoutSale(...)
    // 提取统计指标
}

// 第387-397行：只有在 enableGrabDelivery 为 true 时才从 grabSaleData 提取数据
if enableGrabDelivery {
    grabTotalSaleAmount = decimal.NewFromFloat(grabSaleData.TotalSaleAmount)
    // ... 其他字段
}

// 第399-417行：累加操作（无论 enableGrabDelivery 是否为 true）
// 如果 enableGrabDelivery 为 false，所有 grab* 变量都是零值，累加时不影响结果
totalOrderNum = totalOrderNum + grabOrderNum  // grabOrderNum 为 0
totalSaleAmount = totalSaleAmount.Add(grabTotalSaleAmount)  // grabTotalSaleAmount 为 decimal.Zero
```

**结论**：✅ 无报错风险。即使 `enableGrabDelivery` 为 false 或没有 Grab 订单，代码也能正常运行。

#### 1.2 CountPaymentDays 方法

**检查点**：
- ✅ 条件判断：`enableGrabDelivery` 判断正确
- ✅ 空数据处理：当 `enableGrabDelivery` 为 false 时，不查询 Grab 数据，不影响原有逻辑
- ✅ 时间计算：`dayStartTime` 和 `dayEndTime` 计算正确，有边界检查

**代码分析**：
```go
// 第778行：只有在 enableGrabDelivery 为 true 时才查询
if enableGrabDelivery {
    dayStartTime, _ := timezoneUtil.FormatTimeToUnix(day)
    dayEndTime := dayStartTime + 86399
    
    if dayStartTime > 0 && dayEndTime > 0 {
        grabPaymentDataList := takeoutRepo.CountTakeoutPayment(...)
        // 查找并追加 Grab 支付数据
    }
}
```

**结论**：✅ 无报错风险。当 `enableGrabDelivery` 为 false 时，不会执行 Grab 相关逻辑。

---

### 2. Grab外卖统计是否符合需求 ✅

#### 2.1 销售统计（CountSaleDays）

**需求**：
1. ✅ Grab 订单数累加到总订单数（不累加到外卖订单数）
2. ✅ Grab 销售额累加到总销售额（不累加到外卖销售额）
3. ✅ Grab 实付金额累加到总实收金额和总营业收入
4. ✅ Grab 其他指标（税费、退款、折扣、商品数量等）累加到总指标
5. ✅ Grab 最小/最大/平均订单金额更新总的最小/最大/平均订单金额（不更新外卖相关字段）
6. ✅ Grab 统计指标单独返回（`GrabOrderNum`, `GrabMinOrderAmount`, `GrabMaxOrderAmount`, `GrabAvgOrderAmount`）

**代码验证**：
```go
// 第399行：累加订单数（只累加到总订单数）
totalOrderNum = totalOrderNum + grabOrderNum

// 第403行：累加销售额（只累加到总销售额）
totalSaleAmount = totalSaleAmount.Add(grabTotalSaleAmount)

// 第406-407行：累加实付金额到总实收金额和总营业收入
totalReceivedAmount = totalReceivedAmount.Add(grabTotalPayAmount)
totalBusinessAmount = totalBusinessAmount.Add(grabTotalBusinessAmount)

// 第414-417行：累加其他指标
totalTax = totalTax.Add(grabTotalTax)
totalRefundAmount = totalRefundAmount.Add(grabTotalRefundAmount)
totalDiscount = totalDiscount.Add(grabTotalDiscount)
totalProductNum = totalProductNum.Add(grabTotalProductNum)

// 第421-440行：更新最小/最大订单金额（只更新总的最小/最大订单金额）
if grabOrderNum > 0 {
    if !ok {
        minOrderAmount = grabMinOrderAmount
    } else if grabMinOrderAmount.LessThanOrEqual(minOrderAmount) {
        minOrderAmount = grabMinOrderAmount
    }
}

// 第518-521行：返回 Grab 统计指标
GrabOrderNum:       grabOrderNum,
GrabMinOrderAmount: grabMinOrderAmount.InexactFloat64(),
GrabMaxOrderAmount: grabMaxOrderAmount.InexactFloat64(),
GrabAvgOrderAmount: grabAvgOrderAmount.InexactFloat64(),
```

**结论**：✅ 完全符合需求。Grab 数据正确累加到总统计字段，不累加到外卖相关字段，且单独返回 Grab 统计指标。

#### 2.2 支付统计（CountPaymentDays）

**需求**：
1. ✅ Grab 支付数据追加到支付列表
2. ✅ Grab 支付数据排在最后（在排序后追加）

**代码验证**：
```go
// 第764-775行：先对原有支付数据进行排序
sort.SliceStable(paymentList, func(i, j int) bool {
    // 排序逻辑
})

// 第777-808行：在排序后追加 Grab 支付数据
if enableGrabDelivery {
    // 查询 Grab 支付数据
    grabPaymentDataList := takeoutRepo.CountTakeoutPayment(...)
    
    // 查找 Grab 支付数据并追加
    for _, grabPayment := range grabPaymentDataList {
        if grabPayment.PaymentName == constant.PaymentMethodNameGrab || 
           grabPayment.PaymentCode == constant.PaymentMethodCodeGrab {
            paymentList = append(paymentList, CountPaymentRespList{...})
            break
        }
    }
}
```

**结论**：✅ 完全符合需求。Grab 支付数据在排序后追加，确保排在最后。

---

### 3. Grab外卖开关是否判断正确 ✅

#### 3.1 开关判断方法

**代码**：
```go
// 第244-245行（CountSaleDays）
shopSetting := ctx.GetCompanySetting()
enableGrabDelivery := shopSetting.IsOpenGrabDelivery()

// 第739-740行（CountPaymentDays）
shopSetting := ctx.GetCompanySetting()
enableGrabDelivery := shopSetting.IsOpenGrabDelivery()
```

**IsOpenGrabDelivery 实现**：
```go
// main/app/model/company.go:232-234
func (model *CompanySetting) IsOpenGrabDelivery() bool {
    return model.EnableGrabDelivery == 1
}
```

**验证**：
- ✅ 使用 `ctx.GetCompanySetting()` 获取店铺设置
- ✅ 调用 `IsOpenGrabDelivery()` 方法判断开关
- ✅ 方法实现正确，检查 `EnableGrabDelivery == 1`

**结论**：✅ 开关判断正确。使用统一的 `IsOpenGrabDelivery()` 方法，逻辑清晰。

#### 3.2 开关使用位置

**CountSaleDays**：
- ✅ 第307行：查询 Grab 销售数据前判断
- ✅ 第387行：从 `grabSaleData` 提取数据前判断

**CountPaymentDays**：
- ✅ 第778行：查询 Grab 支付数据前判断

**结论**：✅ 开关使用位置正确。所有 Grab 相关操作都在 `enableGrabDelivery` 判断内。

---

### 4. 如果没有Grab外卖订单是否影响原有统计逻辑 ✅

#### 4.1 CountSaleDays 方法

**场景分析**：

**场景1：`enableGrabDelivery` 为 false**
- `grabSaleData` 为零值
- `grabOrderNum` 为 0
- `grabTotalSaleAmount` 等变量为 `decimal.Zero`
- 累加时：`totalOrderNum = totalOrderNum + 0`（不影响）
- 累加时：`totalSaleAmount = totalSaleAmount.Add(decimal.Zero)`（不影响）

**场景2：`enableGrabDelivery` 为 true，但没有 Grab 订单**
- `grabSaleData` 为零值（所有字段为 0）
- `grabOrderNum` 为 0
- `grabTotalSaleAmount` 等变量为 `decimal.Zero`
- 累加时：`totalOrderNum = totalOrderNum + 0`（不影响）
- 累加时：`totalSaleAmount = totalSaleAmount.Add(decimal.Zero)`（不影响）

**场景3：`enableGrabDelivery` 为 true，有 Grab 订单**
- `grabSaleData` 有数据
- `grabOrderNum > 0`
- `grabTotalSaleAmount` 等变量有值
- 累加时：正确累加到总统计字段

**结论**：✅ 不影响原有统计逻辑。无论是否有 Grab 订单，原有统计逻辑都不受影响。

#### 4.2 CountPaymentDays 方法

**场景分析**：

**场景1：`enableGrabDelivery` 为 false**
- 不执行 Grab 相关逻辑
- 原有支付数据正常返回

**场景2：`enableGrabDelivery` 为 true，但没有 Grab 订单**
- 查询 `CountTakeoutPayment` 返回空列表
- `grabPaymentDataList` 为空
- 循环不执行，不追加 Grab 支付数据
- 原有支付数据正常返回

**场景3：`enableGrabDelivery` 为 true，有 Grab 订单**
- 查询 `CountTakeoutPayment` 返回 Grab 支付数据
- 追加到支付列表

**结论**：✅ 不影响原有统计逻辑。无论是否有 Grab 订单，原有支付统计逻辑都不受影响。

---

## 潜在问题和建议

### 1. 代码优化建议（非必需）

#### 1.1 减少重复代码

**当前**：`CountSaleDays` 和 `CountPaymentDays` 都获取 `shopSetting` 和 `enableGrabDelivery`

**建议**：可以考虑在方法开始处统一获取，但当前实现已经足够清晰。

#### 1.2 错误处理

**当前**：`timezoneUtil.FormatTimeToUnix(day)` 的第二个返回值（错误）被忽略

**建议**：虽然当前实现不会报错（因为 `day` 格式固定），但可以考虑添加错误处理：
```go
dayStartTime, err := timezoneUtil.FormatTimeToUnix(day)
if err != nil {
    // 记录日志或使用默认值
    continue
}
```

**优先级**：低（当前实现已经足够安全）

---

## 总结

### ✅ 审查通过

所有检查项均通过：

1. **报错风险**：✅ 无报错风险，代码逻辑安全
2. **Grab外卖统计符合需求**：✅ 完全符合需求，累加逻辑正确
3. **Grab外卖开关判断**：✅ 开关判断正确，使用统一方法
4. **空数据不影响原有逻辑**：✅ 无论是否有 Grab 订单，原有统计逻辑都不受影响

### 代码质量

- ✅ 代码逻辑清晰，易于维护
- ✅ 使用 `decimal` 包进行精确计算
- ✅ 条件判断完整，无遗漏
- ✅ 变量初始化正确，无未初始化使用

### 建议

- ⚠️ 可以考虑添加 `FormatTimeToUnix` 的错误处理（低优先级）
- ✅ 当前代码可以直接使用，无需修改

---

**审查人**: AI Assistant  
**审查日期**: 2026-01-20  
**审查版本**: 当前代码版本
