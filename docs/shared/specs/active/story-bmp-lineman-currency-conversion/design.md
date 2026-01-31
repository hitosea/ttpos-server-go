# Lineman 订单金额单位转换 技术设计

## 概述

| 项目 | 内容 |
|------|------|
| Spec ID | story-bmp-lineman-currency-conversion |
| 设计人 | rikugun |
| 设计日期 | 2026-01-22 |
| 总 SP | 1 |

---

## 代码复用分析

### 可复用代码

| 文件 | 说明 | 复用方式 |
|------|------|---------|
| 无 | 这是简单的乘法运算 | 直接实现 |

### 需要新建/修改

| 文件 | 说明 |
|------|------|
| `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` | 添加转换函数，修改 saveOrder 和 updateOrder |

---

## 架构设计

### 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Lineman Webhook                          │
│                 (金额单位: 泰铢/元)                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              lineman_order.go                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  convertBahtToCent(amount float64) int64            │   │
│  │  - 泰铢 × 100 = 分                                   │   │
│  │  - 统一转换入口                                       │   │
│  └─────────────────────────────────────────────────────┘   │
│                              │                              │
│         ┌────────────────────┼────────────────────┐        │
│         ▼                    ▼                    ▼        │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐ │
│  │ saveOrder   │      │ updateOrder │      │ (future)    │ │
│  │ - L126-127  │      │ - L249-250  │      │             │ │
│  │ - L153-154  │      │ - L278-279  │      │             │ │
│  └─────────────┘      └─────────────┘      └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Database                                │
│                 (金额单位: 分)                               │
│  ┌─────────────────────┐  ┌─────────────────────────────┐  │
│  │ ttpos_takeout_order │  │ ttpos_takeout_order_item    │  │
│  │ - total_amount      │  │ - price                     │  │
│  │ - subtotal          │  │ - total_price               │  │
│  └─────────────────────┘  └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 设计决策

**为什么抽取通用函数？**

1. **单一职责**: 转换逻辑集中在一处，便于维护
2. **可测试性**: 可以单独为转换函数编写单元测试
3. **可扩展性**: 未来如果有其他需要转换的场景，可以直接复用
4. **日志记录**: 统一的转换入口方便添加日志记录

---

## 组件和接口

### 新增函数: convertBahtToCent

**位置**: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`

**函数签名**:

```go
// convertBahtToCent 将泰铢金额转换为分
// Lineman API 返回的金额单位是泰铢（元），TTPOS 系统使用分
// 转换公式: 分 = 泰铢 × 100
func convertBahtToCent(baht float64) int64 {
    return int64(baht * 100)
}
```

**设计要点**:

| 要点 | 说明 |
|------|------|
| 输入类型 | `float64` - 与 Lineman API 字段类型一致 |
| 输出类型 | `int64` - 整数分，避免浮点精度问题 |
| 转换公式 | `baht × 100` |
| 精度处理 | 使用 `int64()` 截断小数（泰铢通常为整数或两位小数）|

---

## 修改点详情

### saveOrder 方法修改

**文件**: `lineman_order.go:116-157`

| 行号 | 原代码 | 修改后 |
|------|--------|--------|
| 126 | `TotalAmount: req.RestaurantRevenue` | `TotalAmount: convertBahtToCent(req.RestaurantRevenue)` |
| 127 | `Subtotal: req.RestaurantRevenue` | `Subtotal: convertBahtToCent(req.RestaurantRevenue)` |
| 153 | `Price: item.UnitPrice` | `Price: convertBahtToCent(item.UnitPrice)` |
| 154 | `TotalPrice: item.UnitPrice * float64(item.Quantity)` | `TotalPrice: convertBahtToCent(item.UnitPrice) * int64(item.Quantity)` |

### updateOrder 方法修改

**文件**: `lineman_order.go:248-279`

| 行号 | 原代码 | 修改后 |
|------|--------|--------|
| 249 | `TotalAmount: req.RestaurantRevenue` | `TotalAmount: convertBahtToCent(req.RestaurantRevenue)` |
| 250 | `Subtotal: req.RestaurantRevenue` | `Subtotal: convertBahtToCent(req.RestaurantRevenue)` |
| 278 | `Price: item.UnitPrice` | `Price: convertBahtToCent(item.UnitPrice)` |
| 279 | `TotalPrice: item.UnitPrice * float64(item.Quantity)` | `TotalPrice: convertBahtToCent(item.UnitPrice) * int64(item.Quantity)` |

---

## 数据模型

### 无变更

现有表结构保持不变，仅修改写入数据库的值：

| 表 | 字段 | 类型 | 修改前值 | 修改后值 |
|---|---|---|---|---|
| ttpos_takeout_order | total_amount | DECIMAL | 100 (泰铢) | 10000 (分) |
| ttpos_takeout_order | subtotal | DECIMAL | 100 (泰铢) | 10000 (分) |
| ttpos_takeout_order_item | price | DECIMAL | 50 (泰铢) | 5000 (分) |
| ttpos_takeout_order_item | total_price | DECIMAL | 100 (泰铢) | 10000 (分) |

---

## API 设计

### 无变更

Webhook API 入参和响应保持不变，仅内部处理逻辑变更。

---

## 风险识别

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 历史数据金额错误 | 中 | 提供数据修复 SQL 脚本（见附录） |
| 精度丢失 | 低 | 使用 `int64` 整数运算，避免浮点误差 |
| 类型不匹配 | 低 | 确认 DAO 层字段类型兼容 `int64` |

---

## 测试策略

### 单元测试

**目标覆盖率**: 80%+

**测试用例**:

| 测试场景 | 输入 | 期望输出 |
|---------|------|---------|
| 整数金额 | 100.00 | 10000 |
| 小数金额 | 99.50 | 9950 |
| 零金额 | 0.00 | 0 |
| 小金额 | 0.01 | 1 |

**测试命令**:

```bash
cd ttpos-bmp/app/ttpos-takeout && go test ./internal/logic/lineman/... -v -cover
```

---

## 附录

### 历史数据修复 SQL

```sql
-- 修复历史 Lineman 订单金额（泰铢 → 分）
-- 执行前请备份数据！

-- 1. 修复订单主表
UPDATE ttpos_takeout_order
SET total_amount = total_amount * 100,
    subtotal = subtotal * 100
WHERE provider_name = 'lineman'
  AND total_amount < 100000;  -- 假设原金额小于 1000 泰铢的需要修复

-- 2. 修复订单明细表
UPDATE ttpos_takeout_order_item oi
INNER JOIN ttpos_takeout_order o ON oi.order_uuid = o.uuid
SET oi.price = oi.price * 100,
    oi.total_price = oi.total_price * 100
WHERE o.provider_name = 'lineman'
  AND oi.price < 100000;  -- 假设原金额小于 1000 泰铢的需要修复
```

---

**版本**: v1.0.0
**创建日期**: 2026-01-22
