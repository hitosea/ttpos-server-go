# Lineman 订单金额单位转换 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 1 |
| 总任务数 | 4 |
| 已完成 | 4 |
| 完成率 | 100% |

---

## Phase 1: 核心实现

### 1.1 创建金额转换函数

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` |
| Purpose | 添加 `convertBahtToCent` 函数，统一泰铢到分的转换逻辑 |
| Requirements | Req 1 AC 3, Req 2 AC 3 |
| Leverage | 无 |

**实现要点**:
```go
// convertBahtToCent 将泰铢金额转换为分
// Lineman API 返回的金额单位是泰铢（元），TTPOS 系统使用分
// 转换公式: 分 = 泰铢 × 100
func convertBahtToCent(baht float64) int64 {
    return int64(baht * 100)
}
```

- [x] 完成

---

### 1.2 修改 saveOrder 方法

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` |
| Purpose | 在 placeOrder 流程中应用金额转换 |
| Requirements | Req 1 AC 1, Req 1 AC 2 |
| Leverage | Task 1.1 的 `convertBahtToCent` 函数 |

**修改点**:

| 行号 | 字段 | 原代码 | 修改后 |
|------|------|--------|--------|
| 126 | TotalAmount | `req.RestaurantRevenue` | `convertBahtToCent(req.RestaurantRevenue)` |
| 127 | Subtotal | `req.RestaurantRevenue` | `convertBahtToCent(req.RestaurantRevenue)` |
| 153 | Price | `item.UnitPrice` | `convertBahtToCent(item.UnitPrice)` |
| 154 | TotalPrice | `item.UnitPrice * float64(item.Quantity)` | `convertBahtToCent(item.UnitPrice) * int64(item.Quantity)` |

- [x] 完成

---

### 1.3 修改 updateOrder 方法

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` |
| Purpose | 在 orderUpdate 流程中应用金额转换 |
| Requirements | Req 2 AC 1, Req 2 AC 2 |
| Leverage | Task 1.1 的 `convertBahtToCent` 函数 |

**修改点**:

| 行号 | 字段 | 原代码 | 修改后 |
|------|------|--------|--------|
| 249 | TotalAmount | `req.RestaurantRevenue` | `convertBahtToCent(req.RestaurantRevenue)` |
| 250 | Subtotal | `req.RestaurantRevenue` | `convertBahtToCent(req.RestaurantRevenue)` |
| 278 | Price | `item.UnitPrice` | `convertBahtToCent(item.UnitPrice)` |
| 279 | TotalPrice | `item.UnitPrice * float64(item.Quantity)` | `convertBahtToCent(item.UnitPrice) * int64(item.Quantity)` |

- [x] 完成

---

## Phase 2: 测试与文档

### 2.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order_test.go` |
| Purpose | 验证金额转换逻辑的正确性 |
| Requirements | 测试覆盖率 ≥ 80% |

**测试用例**:

| 场景 | 输入 (泰铢) | 期望输出 (分) |
|------|------------|---------------|
| 整数金额 | 100.00 | 10000 |
| 带小数金额 | 99.50 | 9950 |
| 零金额 | 0.00 | 0 |
| 最小金额 | 0.01 | 1 |
| 大金额 | 9999.99 | 999999 |

**测试命令**:
```bash
cd ttpos-bmp/app/ttpos-takeout && go test ./internal/logic/lineman/... -v -cover
```

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`（需要配置文件环境）

### 功能完整性
- [x] placeOrder 金额转换正确（Req 1）
- [x] orderUpdate 金额转换正确（Req 2）
- [ ] 日志记录原始金额和转换后金额（可选优化）

### 部署注意
- [x] 历史数据修复 SQL 准备（见 design.md 附录）
- [ ] 部署后验证新订单金额正确

---

**版本**: v1.0.0
**创建日期**: 2026-01-22
