# story-takeout-peak-hour-timing 技术设计

## 📋 概述

| 项目 | 内容 |
|------|------|
| Spec ID | story-takeout-peak-hour-timing |
| 设计人 | 王昱 |
| 设计日期 | 2026-02-26 |
| 总 SP | 2 |

## 🎯 设计目标

将外卖订单高峰期记录的触发时机从「接单/取消时」调整为「订单完成时」，使统计数据更准确反映实际业务交付时段。

---

## 🔄 代码复用分析

### 可复用代码

| 文件 | 说明 | 复用方式 |
|------|------|---------|
| `main/app/service/takeout/takeout_order.go` | RecordTakeoutOrderPeakTime 方法 | 修改内部逻辑 |
| `main/app/repository/sale_order_peak_time.go` | Record 方法 | 直接复用 |
| `main/app/model/sale_order_peak_time.go` | SaleOrderPeakTime 模型 | 直接复用 |

### 需要修改

| 文件 | 说明 |
|------|------|
| `main/app/event/takeout/takeout_order_accept_event_handler.go` | 移除高峰期记录调用 |
| `main/app/event/takeout/takeout_order_cancel_event_handler.go` | 移除高峰期记录调用 |
| `main/app/event/takeout/takeout_order_completed_event_handler.go` | 添加高峰期记录调用 |
| `main/app/service/takeout/takeout_order.go` | 修改 determineRecordType 和 buildSaleBillFromTakeoutOrder |
| `main/app/modules/takeout/domain/service/takeout_order_service.go` | 修改 BatchAssignShiftLogToPendingOrders 的高峰期过滤逻辑 |

---

## 🏗️ 架构设计

### 当前流程（需移除）

```
接单事件 (OrderAccepted)
    ↓
takeout_order_accept_event_handler.go:93-98
    ↓
RecordTakeoutOrderPeakTime() → inc 增加高峰期
```

```
取消事件 (OrderCancelled)
    ↓
takeout_order_cancel_event_handler.go:82-88
    ↓
RecordTakeoutOrderPeakTime() → dec 减少高峰期
```

### 目标流程（需实现）

```
完成事件 (OrderCompleted)
    ↓
takeout_order_complete_event_handler.go
    ↓
RecordTakeoutOrderPeakTime() → inc 增加高峰期
```

### 分层说明

- **Event Handler**: `main/app/event/takeout/` - 事件处理器，触发高峰期记录
- **Service Layer**: `main/app/service/takeout/` - 业务逻辑，判断记录类型和构建数据
- **Repository Layer**: `main/app/repository/` - 数据持久化，写入高峰期统计表

---

## 🧩 组件和接口

### Service: takeoutSrv

**位置**: `main/app/service/takeout/takeout_order.go`

**现有接口** (无需修改签名):
```go
func (s *takeoutSrv) RecordTakeoutOrderPeakTime(ctx context.Context, orderUuid uint64, companyUuid uint64) error
```

### 需修改的内部函数

#### 1. determineRecordType (行 1639-1658)

**当前逻辑**:
```go
func determineRecordType(order *takeoutModel.TakeoutOrder) string {
    if order.AcceptedBy <= 0 || order.AcceptedTime <= 0 {
        return ""
    }
    if order.OrderState == valueObject.TakeoutOrderStateAccepted {
        return "inc"  // 接单时增加
    } else if order.OrderState == valueObject.TakeoutOrderStateCanceled {
        return "dec"  // 取消时减少
    }
    return ""
}
```

**目标逻辑**:
```go
func determineRecordType(order *takeoutModel.TakeoutOrder) string {
    // 仅在订单完成时记录，且必须有完成时间和接单人
    if order.AcceptedBy <= 0 || order.AcceptedTime <= 0 {
        return ""
    }
    if order.OrderState == valueObject.TakeoutOrderStateCompleted && order.CompleteTime > 0 {
        return "inc"
    }
    return ""
}
```

#### 2. buildSaleBillFromTakeoutOrder (行 1660-1703)

**当前逻辑**: 根据 recordType 使用 AcceptedTime 或 RejectedTime

**目标逻辑**: 统一使用 CompleteTime 作为高峰期记录时间

```go
func buildSaleBillFromTakeoutOrder(order *takeoutModel.TakeoutOrder, recordType string) *saleOrderModel.SaleBill {
    return &saleOrderModel.SaleBill{
        BillTime:      order.CompleteTime,    // 使用完成时间
        CashierUuid:   order.CompleteBy,      // 使用完成操作人（如有）
        PaymentAmount: order.PlatformTotal,
        // ... 其他字段
    }
}
```

---

## 📊 数据模型

### 复用现有模型: SaleOrderPeakTime

**位置**: `main/app/model/sale_order_peak_time.go`

```go
type SaleOrderPeakTime struct {
    ID          uint64  `gorm:"primaryKey"`
    Uuid        uint64  `gorm:"uniqueIndex"`
    Date        int     // 日期（当天 00:00:00 时间戳）
    Hour        int     // 小时（0-23）
    Num         int     // 订单数
    Amount      float64 // 订单金额
    CashierUuid uint64  // 操作人 UUID
    CreateTime  int     `gorm:"autoCreateTime"`
    UpdateTime  int     `gorm:"autoUpdateTime"`
    DeleteTime  int     `gorm:"default:0"`
}
```

**无需修改**，复用现有结构。

---

## 🔌 API 设计

**无需 API 变更**

本需求为纯后端逻辑调整，不涉及 API 接口修改。Shop 端报表查询 API 保持不变，自动展示新逻辑下的统计数据。

---

## ⚠️ 风险识别

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 历史数据口径不一致 | 中 | 功能上线后仅对新完成订单生效，历史数据保持不变 |
| 订单完成时间跨天 | 低 | 按完成时间所在日期和小时统计，与业务已确认 |
| 完成事件处理器不存在 | 低 | 检查是否有现有处理器，若无则需新建或使用其他触发点 |

---

## 🧪 测试策略

### 测试场景

1. **正常完成订单**: 订单完成时，高峰期统计增加
2. **取消订单**: 订单取消时，高峰期统计不变（不再减少）
3. **接单订单**: 订单接单时，高峰期统计不变（不再增加）
4. **跨天完成**: 验证按完成时间日期正确归属

### 测试命令

```bash
cd main && go test -v ./app/service/takeout/... -run TestRecordTakeoutOrderPeakTime
cd main && go test -coverprofile=coverage.out ./app/service/takeout/...
```

---

**版本**: v1.0.0
**设计日期**: 2026-02-26
