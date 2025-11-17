# 订单业务流程

> 👤 **受众**: 开发者  
> 📖 **用途**: 理解订单从创建到完成的完整流程

---

## 流程概览

```
开台 → 点餐 → 加菜/退菜 → 结账 → 支付 → 完成订单 → 清台
```

---

## 1. 开台

**操作**: 顾客入座，收银员选择桌台，开始点餐

**系统流程**:
```
1. 收银端 → POST /api/v1/cashier/desk/open
2. 验证桌台状态（必须是空闲状态）
3. 创建桌台记录，设置状态为"使用中"
4. 返回桌台信息
```

**数据库操作**:
```sql
-- 更新桌台状态
UPDATE ttpos_desk SET status = 1, open_time = {timestamp} WHERE uuid = {desk_uuid}
```

---

## 2. 创建订单（点餐）

**操作**: 收银员选择商品，创建订单

**系统流程**:
```
1. 收银端 → POST /api/v1/cashier/order/create
2. 参数验证
   - 验证桌台是否开台
   - 验证商品是否存在
   - 验证商品库存（如果启用）
3. 创建订单
   - 生成订单号
   - 计算订单金额
   - 创建订单记录
   - 创建订单明细
4. 发布"订单创建"事件
5. 返回订单信息
```

**伪代码**:
```go
func CreateOrder(req CreateOrderReq) (*OrderResp, error) {
    // 1. 验证桌台
    desk := validateDesk(req.DeskUuid)
    
    // 2. 验证商品
    products := validateProducts(req.Items)
    
    // 3. 计算金额
    totalAmount := calculateAmount(req.Items, products)
    
    // 4. 创建订单
    tx := db.Begin()
    order := Order{
        Uuid:        genUuid(),
        OrderNo:     genOrderNo(),
        DeskUuid:    req.DeskUuid,
        TotalAmount: totalAmount,
        Status:      OrderStatusPending,
    }
    tx.Create(&order)
    
    // 5. 创建订单明细
    for _, item := range req.Items {
        orderItem := OrderItem{
            OrderUuid:   order.Uuid,
            ProductUuid: item.ProductUuid,
            Quantity:    item.Quantity,
            Price:       item.Price,
            Amount:      item.Price * item.Quantity,
        }
        tx.Create(&orderItem)
    }
    
    tx.Commit()
    
    // 6. 发布事件
    publishOrderCreatedEvent(order)
    
    return &OrderResp{OrderUuid: order.Uuid}, nil
}
```

---

## 3. 加菜

**操作**: 订单创建后，顾客追加商品

**系统流程**:
```
1. 收银端 → POST /api/v1/cashier/order/add_dish
2. 验证订单状态（必须是待支付状态）
3. 添加订单明细
4. 更新订单金额
5. 返回更新后的订单
```

---

## 4. 退菜

**操作**: 取消订单中的某个商品

**系统流程**:
```
1. 收银端 → POST /api/v1/cashier/order/return_dish
2. 验证订单状态
3. 验证商品是否可退（未上菜）
4. 删除/标记订单明细
5. 更新订单金额
6. 返回更新后的订单
```

---

## 5. 结账

**操作**: 顾客用餐完毕，准备支付

**系统流程**:
```
1. 收银端 → POST /api/v1/cashier/order/checkout
2. 计算最终金额
   - 应付金额 = 订单金额
   - 折扣金额（如果有会员折扣）
   - 优惠金额（如果使用优惠券）
   - 抹零金额
   - 实付金额 = 应付金额 - 折扣 - 优惠 - 抹零
3. 更新订单金额
4. 返回结账信息
```

---

## 6. 支付

**操作**: 顾客选择支付方式进行支付

**系统流程**:
```
1. 收银端 → POST /api/v1/cashier/order/pay
2. 验证订单状态
3. 根据支付方式调用支付接口
   - 现金：直接标记已支付
   - 微信/支付宝：调用第三方支付
   - 会员卡：扣减余额
4. 等待支付结果
5. 更新订单状态为"已支付"
6. 发布"支付成功"事件
7. 返回支付结果
```

**支付流程图**:
```
选择支付方式
    │
    ├─ 现金支付
    │   └─ 直接完成
    │
    ├─ 微信/支付宝
    │   ├─ 调用支付API
    │   ├─ 展示二维码
    │   ├─ 等待支付结果
    │   └─ 回调处理
    │
    └─ 会员卡支付
        ├─ 验证余额
        ├─ 扣减余额
        └─ 完成支付
```

---

## 7. 完成订单

**操作**: 支付成功后，订单完成

**系统流程**:
```
1. 支付成功回调 → 更新订单状态
2. 发布"订单完成"事件
3. 事件处理器：
   - 更新库存（如果启用）
   - 更新会员积分
   - 发送通知
   - 打印小票
4. 返回成功
```

---

## 8. 清台

**操作**: 顾客离开，收银员清理桌台

**系统流程**:
```
1. 收银端 → POST /api/v1/cashier/desk/clean
2. 验证桌台状态
3. 验证桌台订单是否全部完成
4. 更新桌台状态为"空闲"
5. 返回成功
```

---

## 异常处理

### 订单取消

```
1. 收银端 → POST /api/v1/cashier/order/cancel
2. 验证订单状态（只能取消待支付订单）
3. 更新订单状态为"已取消"
4. 如果已支付，发起退款
5. 发布"订单取消"事件
```

### 订单退款

```
1. 收银端 → POST /api/v1/cashier/order/refund
2. 验证订单状态（只能退已支付订单）
3. 调用支付退款接口
4. 更新订单状态为"已退款"
5. 发布"订单退款"事件
```

---

## 事件流

```
订单创建事件 (OrderCreated)
  ↓
  ├─ 发送通知给厨房
  ├─ 记录操作日志
  └─ 更新统计数据

支付成功事件 (PaymentSuccess)
  ↓
  ├─ 更新会员积分
  ├─ 发送支付通知
  ├─ 打印小票
  └─ 更新库存

订单完成事件 (OrderCompleted)
  ↓
  ├─ 更新营业报表
  ├─ 发送满意度调查
  └─ 记录完成日志
```

---

## 状态流转

```
┌─────────┐
│ 待支付  │ ◄─── 创建订单
└────┬────┘
     │ 支付
     ▼
┌─────────┐
│ 已支付  │
└────┬────┘
     │ 完成
     ▼
┌─────────┐
│ 已完成  │
└─────────┘

  取消/退款
     │
     ▼
┌─────────┐
│ 已取消  │
└─────────┘
```

---

## 相关文档

- [支付业务流程](./payment-flow.md) - 支付详细流程
- [业务术语表](../glossary.md) - 业务术语说明
- [订单API文档](../../shared/api/rest-api.md) - 订单API详细说明

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

