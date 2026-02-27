# story-takeout-peak-hour-timing 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 2 |
| 总任务数 | 6 |
| 已完成 | 6 |
| 完成率 | 100% |

---

## Phase 1: 修改事件处理器

### 1.1 移除接单事件中的高峰期记录

| 项目 | 内容 |
|------|------|
| File | `main/app/event/takeout/takeout_order_accept_event_handler.go` |
| Line | 93-98 |
| Purpose | 移除接单时触发高峰期记录的代码 |
| Requirements | AC3: 接单时不再触发高峰期统计更新 |
| Action | 删除或注释 `RecordTakeoutOrderPeakTime` 调用 |

**变更说明**:
```go
// 删除以下代码块（行 93-98）
// 记录高峰期
if err := takeoutSrv.RecordTakeoutOrderPeakTime(ctx, orderAcceptedEvent.OrderUuid, orderAcceptedEvent.CompanyUuid); err != nil {
    logger.Logger.Error("记录外卖订单高峰期失败", ...)
}
```

- [x] 完成

### 1.2 移除取消事件中的高峰期记录

| 项目 | 内容 |
|------|------|
| File | `main/app/event/takeout/takeout_order_cancel_event_handler.go` |
| Line | 82-88 |
| Purpose | 移除取消时触发高峰期记录的代码 |
| Requirements | AC2: 订单取消时不计入高峰期统计 |
| Action | 删除或注释 `RecordTakeoutOrderPeakTime` 调用 |

**变更说明**:
```go
// 删除以下代码块（行 82-88）
// 记录高峰期（自动判断是增加还是减少）
if err := takeoutSrv.RecordTakeoutOrderPeakTime(ctx, orderCancelEvent.OrderUuid, orderCancelEvent.CompanyUuid); err != nil {
    logger.Logger.Error("记录外卖订单高峰期失败", ...)
}
```

- [x] 完成

### 1.3 在完成事件中添加高峰期记录

| 项目 | 内容 |
|------|------|
| File | `main/app/event/takeout/takeout_order_complete_event_handler.go`（或对应文件） |
| Purpose | 在订单完成事件处理中添加高峰期记录调用 |
| Requirements | AC1: 订单完成时更新高峰期统计数据 |
| Leverage | 参考 accept_event_handler 中的调用方式 |

**新增代码**:
```go
// 记录高峰期
if err := takeoutSrv.RecordTakeoutOrderPeakTime(ctx, orderCompletedEvent.OrderUuid, orderCompletedEvent.CompanyUuid); err != nil {
    logger.Logger.Error("记录外卖订单高峰期失败",
        zap.Uint64("order_uuid", orderCompletedEvent.OrderUuid),
        zap.Error(err))
}
```

**注意**: 需先确认完成事件处理器是否存在，若不存在需要查找订单完成的触发点。

- [x] 完成

---

## Phase 2: 修改 Service 层逻辑

### 2.1 修改 determineRecordType 函数

| 项目 | 内容 |
|------|------|
| File | `main/app/service/takeout/takeout_order.go` |
| Line | 1639-1658 |
| Purpose | 调整记录类型判断逻辑，仅在订单完成时返回 "inc" |
| Requirements | AC1, AC2, AC3 |

**当前代码**:
```go
func determineRecordType(order *takeoutModel.TakeoutOrder) string {
    if order.AcceptedBy <= 0 || order.AcceptedTime <= 0 {
        return ""
    }
    if order.OrderState == valueObject.TakeoutOrderStateAccepted {
        return "inc"
    } else if order.OrderState == valueObject.TakeoutOrderStateCanceled {
        return "dec"
    }
    return ""
}
```

**目标代码**:
```go
func determineRecordType(order *takeoutModel.TakeoutOrder) string {
    // 仅在订单完成时记录高峰期，且必须有接单人和接单时间
    if order.AcceptedBy <= 0 || order.AcceptedTime <= 0 {
        return ""
    }
    if order.OrderState == valueObject.TakeoutOrderStateCompleted && order.CompleteTime > 0 {
        return "inc"
    }
    return ""
}
```

- [x] 完成

### 2.2 修改 buildSaleBillFromTakeoutOrder 函数

| 项目 | 内容 |
|------|------|
| File | `main/app/service/takeout/takeout_order.go` |
| Line | 1660-1703 |
| Purpose | 调整时间字段取值，使用完成时间而非接单/取消时间 |
| Requirements | AC4: 按完成时间所在日期统计高峰期 |

**关键变更**:
- `FinishTime`: 使用 `order.CompletedTime`（完成时间）
- `CashierUuid`: 使用 `order.AcceptedBy`（接单人）
- 移除 `recordType` 参数（简化函数）

- [x] 完成

### 2.3 修改 BatchAssignShiftLogToPendingOrders 批量高峰期逻辑

| 项目 | 内容 |
|------|------|
| File | `main/app/modules/takeout/domain/service/takeout_order_service.go` |
| Line | 2127-2146 |
| Purpose | 仅对已完成订单发布高峰期记录事件 |
| Requirements | 与 determineRecordType 逻辑保持一致 |

**变更说明**: 在发布事件前过滤订单，仅保留 `OrderState == Completed && CompletedTime > 0 && AcceptedBy > 0` 的订单

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [x] `go build ./...` 通过
- [x] `go test ./app/service/takeout/...` 通过（8 个测试用例）

### 功能完整性
- [ ] AC1: 订单完成时更新高峰期统计
- [ ] AC2: 订单取消时不计入高峰期
- [ ] AC3: 订单接单时不触发高峰期更新
- [ ] AC4: 跨天订单按完成时间日期统计

### 文档更新
- [ ] 无需迁移文件
- [ ] 无需更新 shop_01.sql

---

## 相关文件索引

| 文件 | 类型 | 操作 |
|------|------|------|
| `main/app/event/takeout/takeout_order_accept_event_handler.go` | 事件处理器 | 修改（移除调用） |
| `main/app/event/takeout/takeout_order_cancel_event_handler.go` | 事件处理器 | 修改（移除调用） |
| `main/app/event/takeout/takeout_order_completed_event_handler.go` | 事件处理器 | 修改（添加调用） |
| `main/app/service/takeout/takeout_order.go` | Service | 修改（调整逻辑） |
| `main/app/modules/takeout/domain/service/takeout_order_service.go` | Domain Service | 修改（过滤已完成订单） |
| `main/app/repository/sale_order_peak_time.go` | Repository | 不变 |
| `main/app/model/sale_order_peak_time.go` | Model | 不变 |
| `main/app/service/takeout/takeout_order_peak_time_test.go` | 单元测试 | 新增 |

---

**版本**: v1.0.0
**创建日期**: 2026-02-26
