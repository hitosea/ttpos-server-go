# 订单领域（Order Domain）

## 概述

订单领域负责管理订单的创建和基本操作，使用**链式调用（Fluent Interface）**模式简化订单创建流程。

---

## 目录结构

```
order/
├── domain/
│   ├── entity/
│   │   ├── order.go           # 订单聚合根（支持链式调用）
│   │   └── order_test.go      # 单元测试
│   ├── valueobject/
│   │   ├── order_status.go    # 订单状态
│   │   ├── order_item.go      # 订单项
│   │   └── discount.go        # 订单优惠
│   ├── repository/
│   │   └── order_repository.go # 仓储接口
│   └── service/
│       └── order_domain_service.go # 领域服务（含 Builder）
│
├── infrastructure/
│   └── persistence/
│       └── order_repository_impl.go # 仓储实现（含 PO 定义）
│
└── application/
    └── order_app_service.go   # 应用服务
```

---

## ⚠️ Context 约束

在 `modules` 目录中，所有方法必须使用 `pkg/context.Context`，禁止使用标准库 `context.Context`。

详细规范参见：`.cursor/rules/go-modules.mdc`

---

## 链式调用使用示例

### 1. 基础创建订单

```go
// 注入应用服务
appService := order.NewOrderAppService(orderRepo, domainService)

// 链式创建订单
order, err := appService.CreateOrder(ctx).
    WithOrderNo("ORD20241204001").
    WithCustomer(memberUuid).
    WithDesk(deskUuid).
    AddItem(productUuid1, "红烧肉", 1, 68.00).
    AddItem(productUuid2, "米饭", 2, 3.00).
    Save()

if err != nil {
    return err
}
fmt.Printf("订单创建成功: %s, 总金额: %.2f\n", order.OrderNo(), order.Total())
```

### 2. 带优惠的订单

```go
order, err := appService.CreateOrder(ctx).
    WithCustomer(memberUuid).
    AddItem(productUuid1, "红烧肉", 1, 68.00).
    AddItem(productUuid2, "可乐", 2, 8.00).
    ApplyPercentDiscount(10, "新客优惠").      // 10% 折扣
    ApplyFixedDiscount(5, "优惠券").           // 再减 5 元
    Save()
```

### 3. 带备注的商品

```go
order, err := appService.CreateOrder(ctx).
    AddItemWithRemark(productUuid, "红烧肉", 1, 68.00, "不要辣").
    AddItem(productUuid2, "米饭", 1, 3.00).
    WithRemark("尽快送达").
    Save()
```

### 4. 创建并立即确认

```go
order, err := appService.CreateOrder(ctx).
    WithDesk(deskUuid).
    AddItem(productUuid, "套餐A", 1, 99.00).
    SaveAndConfirm()  // 保存后自动确认
```

---

## 订单状态流转

```
   ┌─────────┐
   │ Pending │  待处理（初始状态）
   └────┬────┘
        │ Confirm()
        ▼
   ┌──────────┐
   │Confirmed │  已确认
   └────┬─────┘
        │ StartPreparing()
        ▼
   ┌──────────┐
   │Preparing │  制作中
   └────┬─────┘
        │ Complete()
        ▼
   ┌──────────┐
   │Completed │  已完成
   └──────────┘

   任意状态（除已完成/已取消）可调用 Cancel() → Cancelled
```

---

## 业务规则

| 规则 | 说明 |
|------|------|
| 添加商品 | 只能在 `Pending` 或 `Confirmed` 状态下添加 |
| 应用优惠 | 只能在 `Pending` 或 `Confirmed` 状态下应用 |
| 确认订单 | 必须包含至少一个商品 |
| 取消订单 | 已完成或已取消的订单不能取消 |
| 优惠上限 | 优惠金额不能超过订单小计 |

---

## 金额计算

```go
// 商品小计（不含优惠）
SubTotal = Σ(item.Quantity × item.UnitPrice)

// 总优惠金额
TotalDiscount = Σ(discount.Calculate(SubTotal))

// 订单总金额
Total = SubTotal - TotalDiscount
```

---

## 对外方法一览

### OrderAppService

| 方法 | 功能 |
|------|------|
| `CreateOrder(ctx)` | 创建订单（返回 Builder） |
| `GetOrder(ctx, uuid)` | 获取订单详情 |
| `GetOrderByOrderNo(ctx, orderNo)` | 根据订单号获取订单 |
| `GetOrderList(ctx, req)` | 获取订单列表 |
| `ConfirmOrder(ctx, uuid)` | 确认订单 |
| `CancelOrder(ctx, uuid)` | 取消订单 |

### OrderBuilder（链式方法）

| 方法 | 功能 |
|------|------|
| `WithOrderNo(orderNo)` | 设置订单号 |
| `WithCustomer(uuid)` | 设置客户 |
| `WithDesk(uuid)` | 设置桌台 |
| `WithRemark(remark)` | 设置备注 |
| `AddItem(...)` | 添加商品 |
| `AddItemWithRemark(...)` | 添加商品（带备注） |
| `ApplyPercentDiscount(percent, reason)` | 应用百分比折扣 |
| `ApplyFixedDiscount(amount, reason)` | 应用固定金额折扣 |
| `Save()` | 保存订单 |
| `SaveAndConfirm()` | 保存并确认订单 |

---

## 运行测试

```bash
cd main
go test ./app/modules/order/... -v
```

---

**最后更新**: 2025-12-04
**维护者**: TTPOS Team

