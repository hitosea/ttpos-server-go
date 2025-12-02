# 收银机结账流程

> 📖 **文档类型**: 架构设计文档  
> 🎯 **受众**: 开发人员、产品经理、测试人员  
> 📅 **最后更新**: 2025-01-27

---

## 📋 目录

- [概述](#概述)
- [核心流程](#核心流程)
- [详细步骤](#详细步骤)
- [数据模型](#数据模型)
- [API 接口](#api-接口)
- [关键业务逻辑](#关键业务逻辑)
- [错误处理](#错误处理)
- [注意事项](#注意事项)

---

## 概述

收银机结账流程是 TTPOS 系统的核心业务流程之一，支持多种支付方式、优惠券、积分抵扣、会员优惠等功能。本文档详细描述了从订单检查到完成结账的完整流程。

### 核心特性

- ✅ 多种支付方式（现金、微信、支付宝、会员余额等）
- ✅ 优惠券支持（通用优惠券、会员优惠券）
- ✅ 积分抵扣
- ✅ 会员优惠（折扣、积分）
- ✅ 结账抹零
- ✅ 找零计算
- ✅ 手续费计算
- ✅ 并发控制

### 业务约束

- 一个订单只能使用一个优惠券
- 一个销售账单（SaleBill）只能使用一张通用优惠券
- 订单已创建支付单后，不能修改优惠券和积分抵扣
- 拆单后只能在收银端操作结账
- 有未送厨商品时不能结账

---

## 核心流程

### 流程图

```
┌─────────────────┐
│  点击结账按钮   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│ 订单检查                │
│ GET /cashier/desk/order/ │
│ check                   │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ 获取结账页面信息        │
│ GET /cashier/desk/order/ │
│ payment/info            │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ 选择优惠/会员/积分      │
│ - 选择优惠券            │
│ - 使用会员优惠          │
│ - 设置积分抵扣          │
│ - 设置抹零规则          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ 创建支付单              │
│ POST /cashier/desk/order/│
│ payment/create          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ 完成结账                │
│ POST /cashier/desk/order/│
│ payment/finish          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ 核销优惠券/扣减积分/    │
│ 扣减会员余额           │
└─────────────────────────┘
```

---

## 详细步骤

### 1. 订单检查（OrderCheck）

**接口**: `GET /cashier/desk/order/check`

**Handler**: `DeskHandler.OrderCheck()`

**Service**: `OrderSrv.OrderCheck()`

**触发时机**: 点击结账按钮时

**检查内容**:

1. **订单状态检查**
   - 销售账单是否已结束
   - 销售订单状态是否可结账
   - 拆单后是否在收银端操作

2. **商品检查**
   - 商品是否删除或下架
   - 商品库存是否充足
   - 商品价格是否变动
   - 商品是否超过限购
   - 必点商品是否已选择

3. **H5 订单检查**（如果存在）
   - H5 订单商品是否已接单
   - H5 订单商品金额是否有变化

**代码位置**: `main/app/service/order_cooking.go:OrderCheck()`

**响应结构**:
```go
type OrderCheckRes struct {
    Code          int    // 检查结果代码
    OrderCheckRes *OrderCheckRes // 检查结果详情
}
```

---

### 2. 获取结账页面信息（OrderPaymentInfo）

**接口**: `GET /cashier/desk/order/payment/info`

**Handler**: `DeskHandler.OrderPaymentInfo()`

**Service**: `OrderSrv.InstantOrderPaymentInfo()`

**触发时机**: 进入结账页面时

**返回信息**:

1. **订单金额信息**
   - 应收金额
   - 已收金额
   - 未收金额
   - 优惠券抵扣金额
   - 积分抵扣金额
   - 会员折扣金额
   - 最终应收金额

2. **优惠券列表**
   - 通用优惠券列表
   - 会员优惠券列表（如果有会员）
   - 已选中的优惠券

3. **支付方式列表**
   - 可用支付方式
   - 支付方式手续费率
   - 支付方式是否可用

4. **支付单列表**
   - 已创建的支付单
   - 支付单状态
   - 支付单金额

5. **会员信息**（如果有会员）
   - 会员余额
   - 会员积分
   - 会员等级
   - 会员折扣信息

**代码位置**: `main/app/service/order_pay.go:InstantOrderPaymentInfo()`

**响应结构**:
```go
type InstantOrderPaymentInfoResp struct {
    CouponList     CouponList              // 优惠券列表
    PaymentOrders  PaymentInfoList          // 支付单列表
    PaymentMethods PaymentMethodList        // 支付方式列表
    Amounts        PaymentMethodAmountList  // 金额信息
    MemberInfo     *MemberInfo              // 会员信息
    PointsExchange PointsExchangeInfo       // 积分抵扣信息
    // ...
}
```

---

### 3. 选择/取消优惠券（OrderPaymentCoupon）

**接口**: `POST /cashier/desk/order/payment/coupon`

**Handler**: `DeskHandler.OrderPaymentCoupon()`

**Service**: `OrderSrv.OrderPaymentCoupon()`

**请求参数**:
```go
type InstantOrderPaymentCouponReq struct {
    SaleBillUuid      uint64 `json:"sale_bill_uuid"`     // 销售账单UUID
    SaleOrderUuid     uint64 `json:"sale_order_uuid"`   // 销售订单UUID
    CouponUuid        uint64 `json:"coupon_uuid"`        // 优惠券UUID
    CouponRequirement string `json:"coupon_requirement"` // 优惠券类型: "none" 或 "marketing"
}
```

**业务逻辑**:

1. **加锁**: 对 `SaleBillUuid` 加锁，防止并发操作

2. **验证优惠券**:
   - 通用优惠券: 查询 `ttpos_marketing_coupon` 表
   - 会员优惠券: 查询 `ttpos_member_coupon` 表，验证是否属于该会员

3. **处理优惠券**:
   - 如果订单已有优惠券:
     - 如果是同一个优惠券: 取消选择（软删除）
     - 如果是不同优惠券: 替换为新优惠券
   - 如果订单没有优惠券: 新增优惠券使用记录

4. **更新订单**:
   - 将积分自动抵扣改为手动抵扣（`AutoPointsExchange = 0`）
   - 重新计算订单金额
   - 保存优惠券使用记录到 `ttpos_sale_order_coupon` 表

**代码位置**: `main/app/service/order_pay.go:OrderPaymentCoupon()`

**注意事项**:
- 选择优惠券后，积分自动抵扣失效
- 订单已创建支付单后，不能修改优惠券

---

### 4. 设置积分抵扣（OrderPaymentPoints）

**接口**: `POST /cashier/desk/order/payment/points`

**Handler**: `DeskHandler.OrderPaymentPoints()`

**Service**: `OrderSrv.OrderPaymentPoints()`

**请求参数**:
```go
type InstantOrderPaymentPointsReq struct {
    SaleBillUuid  uint64  `json:"sale_bill_uuid"`  // 销售账单UUID
    SaleOrderUuid uint64  `json:"sale_order_uuid"` // 销售订单UUID
    Points        float64 `json:"points"`          // 抵扣积分数量
}
```

**业务逻辑**:

1. **加锁**: 对 `SaleBillUuid` 加锁

2. **验证**:
   - 检查是否开启积分抵扣功能
   - 检查订单是否有会员
   - 检查订单是否已创建支付单
   - 检查会员可用积分是否充足
   - 检查积分数量是否超过最大抵扣数

3. **更新订单**:
   - 更新销售订单的抵扣积分和抵扣金额
   - 取消抹零规则
   - 取消所有优惠券

**代码位置**: `main/app/service/order_pay.go:OrderPaymentPoints()`

**注意事项**:
- 设置积分抵扣后，优惠券和抹零规则会被取消
- 订单已创建支付单后，不能修改积分抵扣

---

### 5. 使用会员优惠（OrderUseMember）

**接口**: `POST /cashier/desk/order/member/confirm`

**Handler**: `DeskHandler.OrderUseMember()`

**Service**: `OrderSrv.OrderUseMember()`

**请求参数**:
```go
type CheckMemberPasswordReq struct {
    SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
    SaleOrderUuid uint64 `json:"sale_order_uuid"`  // 销售订单UUID
    MemberUuid    uint64 `json:"member_uuid"`      // 会员UUID
    Password      string `json:"password"`          // 会员密码
}
```

**业务逻辑**:

1. **验证会员密码**

2. **获取会员优惠信息**:
   - 会员折扣
   - 会员积分抵扣比例
   - 会员等级优惠

3. **应用会员优惠**:
   - 计算会员折扣金额
   - 更新订单金额
   - 重新计算应收金额

**代码位置**: `main/app/service/order_cashier_member.go:OrderUseMember()`

**注意事项**:
- 如果改价/抹零已失效，会提示重新进行改价/抹零操作

---

### 6. 设置结账抹零规则（OrderPaymentZeroRule）

**接口**: `POST /cashier/desk/order/payment/zero_rule`

**Handler**: `DeskHandler.OrderPaymentZeroRule()`

**Service**: `OrderSrv.InstantOrderPaymentZeroRule()`

**请求参数**:
```go
type InstantOrderPaymentZeroRuleReq struct {
    SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
    SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
    ZeroRule      int    `json:"zero_rule"`       // 抹零规则: 1-不抹零, 2-抹分, 3-抹角, 4-抹元
}
```

**业务逻辑**:

1. **验证抹零规则**

2. **计算抹零金额**:
   - 不抹零: 0
   - 抹分: 抹掉分位
   - 抹角: 抹掉角位
   - 抹元: 抹掉元位

3. **更新订单**:
   - 更新销售订单的抹零规则和抹零金额
   - 重新计算应收金额

**代码位置**: `main/app/service/order_pay.go:InstantOrderPaymentZeroRule()`

**注意事项**:
- 有手续费时不能抹零
- 订单已创建支付单后，不能修改抹零规则

---

### 7. 创建支付单（OrderPaymentCreate）

**接口**: `POST /cashier/desk/order/payment/create`

**Handler**: `DeskHandler.OrderPaymentCreate()`

**Service**: `OrderSrv.InstantOrderPaymentCreate()`

**请求参数**:
```go
type InstantOrderPaymentCreateReq struct {
    SaleBillUuid      uint64  `json:"sale_bill_uuid"`      // 销售账单UUID
    SaleOrderUuid     uint64  `json:"sale_order_uuid"`     // 销售订单UUID
    PaymentMethodUuid uint64  `json:"payment_method_uuid"` // 支付方式UUID
    PaymentAmount     float64 `json:"payment_amount"`       // 支付金额
    PaymentOrderUuid  uint64  `json:"payment_order_uuid"`  // 支付单UUID（可选）
}
```

**业务逻辑**:

1. **验证订单状态**:
   - 检查订单是否已结束
   - 检查订单是否可操作
   - 检查支付方式是否可用

2. **计算支付金额**:
   - 计算手续费
   - 计算实际支付金额
   - 检查支付金额是否大于未收金额

3. **创建支付单**:
   - 现金支付: 直接创建支付单，状态为已支付
   - 会员余额支付: 检查余额是否充足，创建支付单
   - 第三方支付（微信/支付宝）: 创建支付订单，返回二维码

**代码位置**: `main/app/service/order_pay.go:InstantOrderPaymentCreate()`

**支付方式处理**:

- **现金支付** (`cash`):
  - 直接创建支付单，状态为 `PaymentOrderStatusPaid`
  - 不需要额外处理

- **会员余额支付** (`balance`):
  - 检查会员余额是否充足
  - 创建支付单，状态为 `PaymentOrderStatusPaid`
  - 记录主账户和赠送账户扣款金额

- **第三方支付** (`wechat`/`alipay`):
  - 创建连连支付订单
  - 返回支付二维码
  - 支付单状态为 `PaymentOrderStatusUnpaid`
  - 等待支付回调更新状态

---

### 8. 撤销支付单（OrderPaymentCancel）

**接口**: `POST /cashier/desk/order/payment/cancel`

**Handler**: `DeskHandler.OrderPaymentCancel()`

**Service**: `OrderSrv.InstantOrderPaymentCancel()`

**请求参数**:
```go
type InstantOrderPaymentCancelReq struct {
    SaleBillUuid     uint64 `json:"sale_bill_uuid"`     // 销售账单UUID
    SaleOrderUuid    uint64 `json:"sale_order_uuid"`    // 销售订单UUID
    PaymentOrderUuid uint64 `json:"payment_order_uuid"` // 支付单UUID
}
```

**业务逻辑**:

1. **验证支付单状态**:
   - 检查支付单是否存在
   - 检查支付单是否可撤销

2. **撤销支付单**:
   - 更新支付单状态为已撤销
   - 清空支付单金额
   - 重新计算订单金额

**代码位置**: `main/app/service/order_pay.go:InstantOrderPaymentCancel()`

**注意事项**:
- 已支付的第三方支付单不能撤销
- 撤销支付单后，可以重新选择优惠券和积分抵扣

---

### 9. 完成结账（OrderPaymentFinish）

**接口**: `POST /cashier/desk/order/payment/finish`

**Handler**: `DeskHandler.OrderPaymentFinish()`

**Service**: `OrderSrv.InstantOrderPaymentFinish()`

**请求参数**:
```go
type InstantOrderPaymentFinishReq struct {
    SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
    SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}
```

**业务逻辑**:

1. **加锁**: 对 `SaleBillUuid` 加锁

2. **重新计算销售账单**:
   - 重新计算订单金额
   - 重新计算优惠金额
   - 重新计算积分抵扣金额

3. **验证订单状态**:
   - 检查订单是否已结束
   - 检查是否有未送厨商品
   - 检查拆单状态

4. **验证支付状态**:
   - 检查未付款金额是否为 0
   - 检查会员积分是否充足（如果使用积分抵扣）
   - 检查会员余额是否充足（如果使用余额支付）

5. **计算最终金额**:
   - 最终应收 = 应收金额 + 手续费 - 结账抹零金额
   - 总付款金额 = 各个付款单的实收金额之和
   - 找零金额 = 总付款金额 - 最终应收（如果总付款 > 最终应收）

6. **处理找零**:
   - 如果找零金额 > 0，修改现金付款单的金额
   - 总付款金额 = 总付款金额 - 找零金额

7. **在事务中完成结账**:
   - 更新现金支付单（如果有找零）
   - 更新销售订单状态为已结清
   - 更新销售账单状态（如果可以结束）
   - 扣减会员余额（如果使用余额支付）
   - 扣减会员积分（如果使用积分抵扣）
   - 核销优惠券
   - 更新会员消费金额和消费次数
   - 处理会员升级
   - 更新钱箱（如果使用现金支付）
   - 发送结账完成事件

**代码位置**: `main/app/service/order_pay.go:InstantOrderPaymentFinish()`

**关键计算**:

```go
// 最终应收 = 应收金额 + 手续费 - 结账抹零金额
finalAmount = saleOrder.GetAmountValue() + commissionFee - saleOrder.ZeroCheckoutFee

// 总付款金额 = 各个付款单的实收金额之和
totalPay = sum(paymentOrder.Amount)

// 找零金额 = 总付款金额 - 最终应收（如果总付款 > 最终应收）
if totalPay > finalAmount {
    changeAmount = totalPay - finalAmount
    // 修改现金付款单金额
    cashPaymentOrder.Amount = cashPaymentOrder.Amount - changeAmount
    totalPay = totalPay - changeAmount
}
```

**错误处理**:

- "销售订单未结清": 未付款金额 > 0
- "有未送厨的商品": 存在未送厨的商品
- "当前会员抵扣积分不足": 会员积分 < 订单抵扣积分
- "会员余额不足": 会员余额 < 订单余额支付金额
- "收款金额大于最终应收，请先修改收款金额": 超付金额 > 现金支付金额
- "优惠券信息变化，请重新确认": 优惠券核销失败

---

### 10. 免单（OrderFree）

**接口**: `POST /cashier/desk/order/free`

**Handler**: `DeskHandler.OrderFree()`

**Service**: `OrderSrv.InstantOrderFree()`

**请求参数**:
```go
type InstantOrderFreeReq struct {
    SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
    SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}
```

**业务逻辑**:

1. **验证订单状态**

2. **创建免单支付单**:
   - 创建支付方式为"免单"的支付单
   - 支付金额 = 应收金额
   - 支付单状态为已支付

3. **完成结账**:
   - 调用 `InstantOrderPaymentFinish()` 完成结账

**代码位置**: `main/app/service/order_pay.go:InstantOrderFree()`

**注意事项**:
- 免单需要特殊权限
- 免单后订单状态为已结清

---

## 数据模型

### 核心表结构

#### 1. `ttpos_sale_bill` - 销售账单表

```go
type SaleBill struct {
    BaseModel
    DeskUuid      uint64 // 桌台UUID
    Status        int    // 状态: 0-进行中, 1-已结束
    MealNum       int    // 用餐人数
    // ...
}
```

#### 2. `ttpos_sale_order` - 销售订单表

```go
type SaleOrder struct {
    BaseModel
    SaleBillUuid      uint64  // 销售账单UUID
    ConsumerUuid      uint64  // 会员UUID
    Amount            float64 // 应收金额
    PayPoints         float64 // 抵扣积分数量
    PayPointsAmount   float64 // 积分抵扣金额
    AutoPointsExchange int    // 是否自动积分抵扣: 0-手动, 1-自动
    ZeroCheckoutFee   float64 // 结账抹零金额
    Status            int     // 状态: 0-进行中, 1-已结清
    // ...
}
```

#### 3. `ttpos_payment_order` - 支付单表

```go
type PaymentOrder struct {
    BaseModel
    SaleOrderUuid     uint64  // 销售订单UUID
    PaymentMethodUuid uint64  // 支付方式UUID
    PaymentAmount     float64 // 支付金额
    Amount            float64 // 实收金额
    Status            int     // 状态: 0-未支付, 1-已支付, 2-已撤销
    BalanceAmount     float64 // 主账户扣款金额（余额支付）
    GiftBalanceAmount float64 // 赠送账户扣款金额（余额支付）
    // ...
}
```

#### 4. `ttpos_sale_order_coupon` - 销售订单优惠券关联表

```go
type SaleOrderCoupon struct {
    BaseModel
    SaleOrderUuid      uint64  // 销售订单UUID
    CouponUuid         uint64  // 优惠券UUID
    CouponRequirement  string  // 优惠券类型: "none" 或 "marketing"
    CouponAmount       float64 // 优惠券抵扣金额
    CouponOriginAmount float64 // 优惠券原始金额
    // ...
}
```

---

## API 接口

### 接口列表

| 接口 | 方法 | 说明 |
|------|------|------|
| `/cashier/desk/order/check` | GET | 订单检查 |
| `/cashier/desk/order/payment/info` | GET | 获取结账页面信息 |
| `/cashier/desk/order/payment/coupon` | POST | 选择/取消优惠券 |
| `/cashier/desk/order/payment/points` | POST | 设置积分抵扣 |
| `/cashier/desk/order/member/confirm` | POST | 使用会员优惠 |
| `/cashier/desk/order/payment/zero_rule` | POST | 设置抹零规则 |
| `/cashier/desk/order/payment/create` | POST | 创建支付单 |
| `/cashier/desk/order/payment/cancel` | POST | 撤销支付单 |
| `/cashier/desk/order/payment/finish` | POST | 完成结账 |
| `/cashier/desk/order/free` | POST | 免单 |

### 接口详情

详见代码文件: `main/app/api/v1/cashier/cashier_desk.go`

---

## 关键业务逻辑

### 1. 金额计算顺序

1. **计算订单金额**（商品金额 + 服务费 - 折扣）
2. **应用会员折扣**（如果有会员）
3. **应用优惠券**（如果有优惠券）
4. **应用积分抵扣**（如果使用积分）
5. **计算手续费**（根据支付方式）
6. **应用结账抹零**（如果有手续费则不能抹零）
7. **计算最终应收** = 订单金额 + 手续费 - 抹零金额

### 2. 支付单创建规则

- **现金支付**: 直接创建，状态为已支付
- **会员余额支付**: 检查余额，创建支付单，状态为已支付
- **第三方支付**: 创建支付订单，返回二维码，等待支付回调

### 3. 找零计算

- 找零金额 = 总付款金额 - 最终应收（如果总付款 > 最终应收）
- 找零只能从现金支付中扣除
- 如果超付金额 > 现金支付金额，则拒绝完成订单

### 4. 并发控制

- 所有操作都对 `SaleBillUuid` 加锁
- 会员余额操作对 `MemberUuid` 加锁
- 优惠券核销对 `LockNameActivityConsumption` 加锁
- 使用数据库事务保证数据一致性

### 5. 优惠券、积分、抹零互斥规则

- 选择优惠券后，积分自动抵扣失效
- 设置积分抵扣后，优惠券和抹零规则会被取消
- 有手续费时不能抹零

---

## 错误处理

### 常见错误码

| 错误码 | 说明 | 处理方式 |
|--------|------|---------|
| `CodeOrderCheckSplit` | 订单已拆单 | 提示前往收银机操作 |
| `CodeOrderCheckProductStockZero` | 商品库存不足 | 提示删除商品 |
| `CodeOrderCheckProductPriceChanged` | 商品价格变动 | 提示刷新商品价格 |
| `CodeCouponInvalid` | 优惠券无效 | 提示刷新优惠券列表 |
| `CodeOrderPayError` | 支付失败 | 提示重试 |

### 错误处理位置

- **订单检查错误**: `main/app/api/v1/cashier/cashier_desk.go:1193`
- **优惠券错误**: `main/app/api/v1/cashier/cashier_desk.go:1414`
- **库存不足错误**: `main/app/api/v1/cashier/cashier_desk.go:1425`

---

## 注意事项

### 1. 数据一致性

- ✅ 所有金额计算使用 `decimal` 类型，避免浮点数精度问题
- ✅ 使用数据库事务保证数据一致性
- ✅ 使用分布式锁防止并发操作

### 2. 性能优化

- ✅ 使用预加载（Preload）减少数据库查询
- ✅ 使用缓存减少重复计算
- ✅ 异步处理会员升级等非关键操作

### 3. 业务限制

- ❌ 订单已创建支付单后，不能修改优惠券和积分抵扣
- ❌ 有未送厨商品时不能结账
- ❌ 拆单后只能在收银端操作结账
- ❌ 有手续费时不能抹零

### 4. 安全考虑

- ✅ 所有操作都需要认证
- ✅ 验证订单归属
- ✅ 验证支付金额
- ✅ 验证会员权限

---

## 相关文档

- [收银端优惠券使用流程](./cashier-coupon-flow.md)
- [订单系统架构](./order.md)
- [会员系统架构](./member.md)
- [支付系统架构](./payment.md)

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

