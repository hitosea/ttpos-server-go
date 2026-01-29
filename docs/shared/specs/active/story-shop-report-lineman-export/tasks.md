# story-shop-report-lineman-export 任务清单

> 本文档定义统计报表导出增加 LINEMAN 数据的详细执行任务清单。

## 📊 进度总览

| 项目     | 数值 |
| -------- | ---- |
| 总 SP    | 2    |
| 总任务数 | 5    |
| 已完成   | 4    |
| 完成率   | 80%  |

---

## Phase 1: 核心实现

### 1.1 扩展 CountSaleResp 结构体

| 项目        | 内容                                                                                      |
| ----------- | ----------------------------------------------------------------------------------------- |
| File        | `main/app/service/statistics.go`                                                          |
| Purpose     | 在 CountSaleResp 结构体中添加 LINEMAN 统计字段                                            |
| Requirements | Req 1.1, 1.2                                                                              |
| Leverage    | `main/app/service/statistics.go` L122-126 (Grab 字段定义)                                 |

**实现说明**:
在 Grab 字段后添加 LINEMAN 字段（约 L126 后）:
```go
// LINEMAN 平台统计指标
LinemanOrderNum       int64   `json:"lineman_order_num"`        // LINEMAN 订单数
LinemanMinOrderAmount float64 `json:"lineman_min_order_amount"` // LINEMAN 最小订单金额
LinemanMaxOrderAmount float64 `json:"lineman_max_order_amount"` // LINEMAN 最大订单金额
LinemanAvgOrderAmount float64 `json:"lineman_avg_order_amount"` // LINEMAN 平均订单金额
```

- [x] 完成

---

### 1.2 扩展 CountSaleDays 方法 - 添加 LINEMAN 统计变量

| 项目        | 内容                                                                                      |
| ----------- | ----------------------------------------------------------------------------------------- |
| File        | `main/app/service/statistics.go`                                                          |
| Purpose     | 在 CountSaleDays 方法中添加 LINEMAN 统计变量和查询逻辑                                    |
| Requirements | Req 1.1, 1.2                                                                              |
| Leverage    | `main/app/service/statistics.go` L245 (Grab 开关检查), L292-326 (Grab 统计逻辑)           |

**实现说明**:

1. **添加 LINEMAN 开关检查**（约 L245 后）:
```go
enableLineManDelivery := shopSetting.IsOpenLineManDelivery()
```

2. **添加 LINEMAN 统计变量**（约 L296 后）:
```go
// LINEMAN 统计指标
linemanOrderNum       int64
linemanMinOrderAmount decimal.Decimal
linemanMaxOrderAmount decimal.Decimal
linemanAvgOrderAmount decimal.Decimal
```

3. **添加 LINEMAN 数据查询**（约 L326 后，参考 Grab 逻辑）:
```go
var linemanSaleData model.StatisticsTakeoutSaleData
if dayStartTime > 0 && dayEndTime > 0 {
    if enableLineManDelivery {
        linemanSaleData = takeoutRepo.CountTakeoutSale(repository.CountTakeoutReq{
            TimeStart: dayStartTime,
            TimeEnd:   dayEndTime,
            Platform:  "lineman",
        })
        // 提取 LINEMAN 统计指标
        linemanOrderNum = linemanSaleData.TotalOrderNum
        if linemanOrderNum > 0 {
            linemanMinOrderAmount = decimal.NewFromFloat(linemanSaleData.MinOrderAmount)
        }
        if linemanSaleData.MaxOrderAmount > 0 {
            linemanMaxOrderAmount = decimal.NewFromFloat(linemanSaleData.MaxOrderAmount)
        }
        if linemanOrderNum > 0 && linemanSaleData.TotalOrderAmount > 0 {
            linemanAvgOrderAmount = decimal.NewFromFloat(linemanSaleData.TotalOrderAmount).Div(decimal.NewFromInt(linemanOrderNum)).Round(2)
        }
    }
}
```

- [x] 完成

---

### 1.3 扩展 CountSaleDays 方法 - 添加 LINEMAN 数据累加

| 项目        | 内容                                                                                      |
| ----------- | ----------------------------------------------------------------------------------------- |
| File        | `main/app/service/statistics.go`                                                          |
| Purpose     | 在 CountSaleDays 方法中添加 LINEMAN 数据累加到总统计                                      |
| Requirements | Req 3.1, 3.2                                                                              |
| Leverage    | `main/app/service/statistics.go` L379-430 (Grab 累加逻辑)                                 |

**实现说明**:

1. **添加 LINEMAN 累加变量**（约 L382 后）:
```go
var linemanTotalSaleAmount, linemanTotalPayAmount, linemanTotalBusinessAmount, linemanTotalOrderAmount decimal.Decimal
var linemanTotalProductOriginPrice, linemanTotalTax, linemanTotalRefundAmount, linemanTotalDiscount decimal.Decimal
var linemanTotalProductNum decimal.Decimal
```

2. **添加 LINEMAN 数据累加逻辑**（约 L430 后，参考 Grab 逻辑）:
```go
if enableLineManDelivery {
    linemanTotalSaleAmount = decimal.NewFromFloat(linemanSaleData.TotalSaleAmount)
    linemanTotalPayAmount = decimal.NewFromFloat(linemanSaleData.TotalPayAmount)
    linemanTotalBusinessAmount = decimal.NewFromFloat(linemanSaleData.TotalBusinessAmount)
    linemanTotalOrderAmount = decimal.NewFromFloat(linemanSaleData.TotalOrderAmount)
    linemanTotalProductOriginPrice = decimal.NewFromFloat(linemanSaleData.TotalProductOriginPrice)
    linemanTotalTax = decimal.NewFromFloat(linemanSaleData.TotalTax)
    linemanTotalRefundAmount = decimal.NewFromFloat(linemanSaleData.TotalRefundAmount)
    linemanTotalDiscount = decimal.NewFromFloat(linemanSaleData.TotalDiscount)
    linemanTotalProductNum = decimal.NewFromInt(linemanSaleData.TotalProductNum)

    // 累加到总统计
    totalSaleAmount = totalSaleAmount.Add(linemanTotalSaleAmount)
    // ... 其他字段累加（参考 Grab 累加逻辑）
}
```

3. **填充 LINEMAN 字段到响应**（在 CountSaleDaysResp 构建处）:
```go
LinemanOrderNum:       linemanOrderNum,
LinemanMinOrderAmount: linemanMinOrderAmount.InexactFloat64(),
LinemanMaxOrderAmount: linemanMaxOrderAmount.InexactFloat64(),
LinemanAvgOrderAmount: linemanAvgOrderAmount.InexactFloat64(),
```

- [x] 完成

---

### 1.4 扩展 CountPaymentDays 方法

| 项目        | 内容                                                                                      |
| ----------- | ----------------------------------------------------------------------------------------- |
| File        | `main/app/service/statistics.go`                                                          |
| Purpose     | 在 CountPaymentDays 方法中添加 LINEMAN 支付数据统计                                       |
| Requirements | Req 2.1, 2.2                                                                              |
| Leverage    | `main/app/service/statistics.go` CountPayment 方法中集成外卖支付数据的方式                |

**实现说明**:

参考 `CountPayment` 方法中 Grab 支付数据的处理方式，在 `CountPaymentDays` 方法中为每个日期添加 LINEMAN 支付数据:

1. 调用 `CountTakeoutPayment` 获取 LINEMAN 支付数据
2. 按日期筛选数据
3. 将 LINEMAN 支付数据追加到 `PaymentList` 中
4. 确保 LINEMAN 数据排列在最后（在 Grab 之后）

- [x] 完成

---

## Phase 2: 测试和验证

### 2.1 功能验证

| 项目        | 内容                                                       |
| ----------- | ---------------------------------------------------------- |
| File        | -                                                          |
| Purpose     | 验证导出报表包含 LINEMAN 数据                              |
| Requirements | 所有功能需求                                               |

**验证步骤**:

1. 确保测试环境有 LINEMAN 订单数据
2. 调用导出接口 `GET /shop/statistics/export`
3. 验证响应包含 `lineman_order_num`, `lineman_min_order_amount` 等字段
4. 验证 LINEMAN 数据正确累加到总统计

- [ ] 完成

---

## 提交清单

### 代码质量

- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - [ ] 销售统计（按天）导出包含 LINEMAN 数据
  - [ ] 支付数据（按天）导出包含 LINEMAN 数据
  - [ ] LINEMAN 数据纳入汇总统计

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 执行流程

1. **选择任务**: 按顺序执行 1.1 → 1.2 → 1.3 → 1.4 → 2.1
2. **参考 Grab 实现**: 每个任务都有对应的 Leverage 代码行号
3. **运行检查**: 每个任务完成后执行 `go fmt`, `go vet`
4. **标记完成**: 将 `[ ]` 改为 `[x]`
5. **提交代码**: Git commit

---

**模板版本**: v1.0.0
**创建日期**: 2026-01-26
**维护者**: 后端开发组
