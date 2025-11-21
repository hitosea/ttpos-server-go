# 收银端优惠券使用流程

> 📖 **文档类型**: 架构设计文档  
> 🎯 **受众**: 开发人员、产品经理  
> 📅 **最后更新**: 2025-11-21

---

## 📋 目录

- [概述](#概述)
- [优惠券类型](#优惠券类型)
- [核心流程](#核心流程)
- [数据模型](#数据模型)
- [API 接口](#api-接口)
- [代码结构](#代码结构)
- [关键业务逻辑](#关键业务逻辑)
- [注意事项](#注意事项)

---

## 概述

收银端优惠券功能允许收银员在结账时为订单选择或取消优惠券。系统支持两种类型的优惠券：

1. **通用优惠券** (`none`): 所有人可用，一个销售账单（SaleBill）只能使用一张
2. **会员优惠券** (`marketing`): 仅会员可用，每个订单可以使用一张

**核心约束：**
- 一个销售订单（SaleOrder）只能使用一个优惠券
- 选择优惠券后，积分自动抵扣失效，改为手动抵扣
- 如果订单已创建支付单，则不能修改优惠券
- 优惠券在订单结账时进行核销

---

## 优惠券类型

### 常量定义

```go
// app/constant/coupon.go
const (
    CouponRequirementNone   = "none"      // 通用优惠券（所有人可用）
    CouponRequirementMember = "marketing" // 会员优惠券
)
```

### 类型说明

| 类型 | Requirement | 说明 | 使用限制 |
|------|------------|------|---------|
| 通用优惠券 | `none` | 所有人可用，由营销活动创建 | 一个 SaleBill 只能使用一张 |
| 会员优惠券 | `marketing` | 会员通过营销活动获得 | 每个 SaleOrder 可以使用一张 |

---

## 核心流程

### 流程图

```
┌─────────────────┐
│  进入结账页面   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│ 获取优惠券列表          │
│ GET /cashier/desk/order/ │
│ payment/info            │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ 选择/取消优惠券         │
│ POST /cashier/desk/order/│
│ payment/coupon          │
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
│ 核销优惠券              │
│ VerifyCoupon()           │
└─────────────────────────┘
```

### 详细步骤

#### 1. 获取优惠券列表

**触发时机**: 进入结账页面时

**接口**: `GET /cashier/desk/order/payment/info`

**逻辑**:
- 检查是否开启优惠券功能
- 查询通用优惠券列表（`requirement = none`）
- 如果订单有会员，查询会员优惠券列表
- 判断优惠券是否在使用时段内
- 判断优惠券是否已被选中
- 排序优惠券列表（可用在前，通用在前，先到期在前）

**代码位置**: `main/app/service/order.go:GetValidMemberCouponList()`

#### 2. 选择/取消优惠券

**触发时机**: 用户点击选择或取消优惠券

**接口**: `POST /cashier/desk/order/payment/coupon`

**请求参数**:
```go
type InstantOrderPaymentCouponReq struct {
    SaleBillUuid      uint64 `json:"sale_bill_uuid"`     // 销售账单UUID
    SaleOrderUuid     uint64 `json:"sale_order_uuid"`    // 销售订单UUID
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

#### 3. 核销优惠券

**触发时机**: 订单结账完成时（`InstantOrderPaymentFinish`）

**代码位置**: `main/app/service/order.go:VerifyCoupon()`

**业务逻辑**:

1. **会员优惠券核销** (`requirement = marketing`):
   - 验证订单是否有会员
   - 验证优惠券是否可用（未过期、未使用、在使用时段内）
   - 更新会员优惠券状态为已使用
   - 创建会员优惠券使用记录
   - 创建营销优惠券使用记录
   - 更新营销优惠券剩余数量

2. **通用优惠券核销** (`requirement = none`):
   - 验证优惠券是否可用（未过期、数量充足）
   - 减少通用优惠券数量（`Count - 1`）
   - 创建通用优惠券使用记录

**关键验证**:
- 优惠券有效期检查
- 使用时段检查（`DayStartTime` ~ `DayEndTime`）
- 优惠券状态检查（未使用、未禁用）
- 会员归属检查（会员优惠券）

---

## 数据模型

### 核心表结构

#### 1. `ttpos_sale_order_coupon` - 销售订单优惠券关联表

```go
type SaleOrderCoupon struct {
    BaseModel
    CouponAmount        float64 // 优惠券抵扣金额（实际抵扣金额）
    CouponOriginAmount  float64 // 优惠券原始金额（面值）
    CouponRequirement   string  // 优惠券类型: "none" 或 "marketing"
    MemberCouponUuid    uint64  // 会员优惠券UUID（marketing时有值）
    MarketingCouponUuid uint64  // 营销优惠券UUID（none时有值）
    SaleOrderUuid       uint64  // 销售订单UUID
}
```

**说明**:
- 一个订单只能有一条有效记录（`delete_time = 0`）
- `CouponAmount` 是实际抵扣金额，可能小于 `CouponOriginAmount`（受订单金额限制）

#### 2. `ttpos_member_coupon` - 会员优惠券表

```go
type MemberCoupon struct {
    BaseModel
    MemberUuid     uint64  // 会员UUID
    CouponUuid     uint64  // 营销优惠券UUID（关联 ttpos_marketing_coupon）
    Name           string  // 优惠券名称
    Amount         float64 // 优惠券面值
    Status         int     // 状态: 0-未使用, 1-已使用
    StartTime      int64   // 开始时间
    EndTime        int64   // 结束时间
    DayStartTime   string  // 每日适用时段开始时间 (HH:mm)
    DayEndTime     string  // 每日适用时段结束时间 (HH:mm)
    UseTime        int64   // 使用时间
}
```

#### 3. `ttpos_marketing_coupon` - 营销优惠券表

```go
type MarketingCoupon struct {
    BaseModel
    Name           string  // 优惠券名称
    Requirement    string  // 类型: "none" 或 "marketing"
    Amount         float64 // 优惠券面值
    Count          int     // 剩余数量（通用优惠券）
    DayStartTime   string  // 每日适用时段开始时间
    DayEndTime     string  // 每日适用时段结束时间
    ValidStartTime int     // 有效开始时间
    ValidEndTime   int     // 有效结束时间
    Status         int     // 状态: 0-禁用, 1-开启
}
```

#### 4. `ttpos_member_coupon_use_record` - 会员优惠券使用记录

```go
type MemberCouponUseRecord struct {
    BaseModel
    MemberUuid     uint64  // 会员UUID
    CouponUuid     uint64  // 优惠券UUID
    UseOrderUuid   uint64  // 使用订单UUID
    UseOrderAmount float64 // 使用订单金额
}
```

---

## API 接口

### 1. 获取结账页面信息（包含优惠券列表）

**接口**: `GET /cashier/desk/order/payment/info`

**Handler**: `DeskHandler.InstantOrderPaymentInfo()`

**Service**: `OrderSrv.InstantOrderPaymentInfo()`

**响应结构**:
```go
type InstantOrderPaymentInfoResp struct {
    CouponList     CouponList     // 优惠券列表
    PaymentOrders  PaymentInfoList // 支付单列表
    PaymentMethods PaymentMethodList // 支付方式列表
    Amounts        PaymentMethodAmountList // 金额信息
    // ...
}
```

### 2. 选择/取消优惠券

**接口**: `POST /cashier/desk/order/payment/coupon`

**Handler**: `DeskHandler.OrderPaymentCoupon()`

**Service**: `OrderSrv.OrderPaymentCoupon()`

**请求参数**:
```json
{
    "sale_bill_uuid": 1234567890,
    "sale_order_uuid": 9876543210,
    "coupon_uuid": 1111111111,
    "coupon_requirement": "marketing"
}
```

**响应**: 返回更新后的结账页面信息

### 3. 完成结账（核销优惠券）

**接口**: `POST /cashier/desk/order/payment/finish`

**Handler**: `DeskHandler.OrderPaymentFinish()`

**Service**: `OrderSrv.InstantOrderPaymentFinish()`

**说明**: 在事务中调用 `VerifyCoupon()` 核销优惠券

---

## 代码结构

### 目录结构

```
main/
├── app/
│   ├── api/v1/cashier/
│   │   ├── cashier_desk.go          # 桌台收银 API Handler
│   │   └── cashier_instant.go       # 点餐收银 API Handler
│   ├── service/
│   │   ├── order.go                  # 订单服务（核销优惠券）
│   │   └── order_pay.go             # 支付服务（选择优惠券）
│   ├── repository/
│   │   ├── member_coupon.go          # 会员优惠券 Repository
│   │   ├── marketing_coupon.go      # 营销优惠券 Repository
│   │   └── sale_order_coupon.go     # 销售订单优惠券 Repository
│   ├── model/
│   │   ├── member_coupon.go          # 会员优惠券模型
│   │   ├── marketing_coupon.go       # 营销优惠券模型
│   │   └── sale_order_coupon.go      # 销售订单优惠券模型
│   ├── constant/
│   │   └── coupon.go                 # 优惠券常量
│   └── dto/
│       ├── req/
│       │   └── instant.go            # 请求结构
│       └── resp/
│           └── instant.go            # 响应结构
```

### 关键方法

#### Service 层

| 方法 | 文件 | 说明 |
|------|------|------|
| `OrderPaymentCoupon()` | `order_pay.go:30` | 选择/取消优惠券 |
| `GetValidMemberCouponList()` | `order.go:3286` | 获取有效优惠券列表 |
| `VerifyCoupon()` | `order.go:3677` | 核销优惠券 |
| `SortCouponList()` | `order.go:3478` | 排序优惠券列表 |

#### Repository 层

| 方法 | 文件 | 说明 |
|------|------|------|
| `GetValidMemberCouponList()` | `member_coupon.go:103` | 获取会员有效优惠券列表 |
| `GetMemberCouponByUuid()` | `member_coupon.go:79` | 根据UUID获取会员优惠券 |
| `VerifyMemberCoupon()` | `member_coupon.go:151` | 核销会员优惠券 |
| `CreateMemberCouponRecord()` | `member_coupon.go:136` | 创建会员优惠券使用记录 |
| `GetValidCouponList()` | `marketing_coupon.go` | 获取有效通用优惠券列表 |
| `GetCouponByUuid()` | `marketing_coupon.go` | 根据UUID获取营销优惠券 |
| `CreateSaleOrderCoupon()` | `sale_order_coupon.go` | 创建销售订单优惠券记录 |
| `UpdateSaleOrderCoupon()` | `sale_order_coupon.go` | 更新销售订单优惠券记录 |

#### Model 层

| 方法 | 文件 | 说明 |
|------|------|------|
| `IsAvailable()` | `member_coupon.go:43` | 判断优惠券是否可用 |
| `IsExpire()` | `member_coupon.go:72` | 判断优惠券是否过期 |
| `IsUsed()` | `member_coupon.go:81` | 判断优惠券是否已使用 |
| `ReplaceCoupon()` | `sale_order_coupon.go:57` | 更换优惠券 |

---

## 关键业务逻辑

### 1. 优惠券可用性判断

**判断条件**:
1. ✅ 优惠券未过期（`StartTime <= now <= EndTime`）
2. ✅ 优惠券未使用（`Status = 0` 且 `UseTime = 0`）
3. ✅ 优惠券未禁用（`MarketingCoupon.Status = 1`）
4. ✅ 优惠券属于该会员（会员优惠券）
5. ✅ 优惠券在使用时段内（`DayStartTime <= nowTime <= DayEndTime`）
6. ✅ 订单未创建支付单（`hasPay = false`）
7. ✅ 订单金额大于0（积分抵扣后）

**代码位置**: `main/app/model/member_coupon.go:IsAvailable()`

### 2. 优惠券选择逻辑

**场景1: 订单已有优惠券，选择相同优惠券**
- 操作: 取消选择（软删除 `sale_order_coupon` 记录）

**场景2: 订单已有优惠券，选择不同优惠券**
- 操作: 替换优惠券（更新 `sale_order_coupon` 记录）

**场景3: 订单没有优惠券，选择优惠券**
- 操作: 新增优惠券使用记录

**代码位置**: `main/app/service/order_pay.go:OrderPaymentCoupon()`

### 3. 优惠券排序规则

**排序优先级**:
1. 按是否可用分组（可用在前）
2. 按优惠券类型分组（通用在前）
3. 按有效期排序（先到期在前）
4. 按 `sort` 字段排序
5. 按 `CouponUuid` 排序（创建时间）

**代码位置**: `main/app/service/order.go:SortCouponList()`

### 4. 优惠券核销逻辑

**会员优惠券核销**:
1. 验证优惠券可用性
2. 更新会员优惠券状态: `Status = 1`, `UseTime = now`
3. 创建会员优惠券使用记录
4. 创建营销优惠券使用记录
5. 更新营销优惠券剩余数量

**通用优惠券核销**:
1. 验证优惠券可用性
2. 减少优惠券数量: `Count = Count - 1`
3. 创建通用优惠券使用记录

**代码位置**: `main/app/service/order.go:VerifyCoupon()`

### 5. 优惠券金额计算

**实际抵扣金额**:
- 优惠券面值（`CouponOriginAmount`）可能大于订单金额
- 实际抵扣金额（`CouponAmount`） = `min(订单金额, 优惠券面值)`

**计算位置**: `main/app/model/sale_order.go:CalcCouponExchangeAmount()`

---

## 注意事项

### 1. 并发控制

- ✅ 选择优惠券时对 `SaleBillUuid` 加锁
- ✅ 核销优惠券时对活动消费加锁（`LockNameActivityConsumption`）
- ✅ 使用数据库事务保证数据一致性

### 2. 错误处理

**常见错误**:
- "优惠券不属于该会员" - 会员优惠券验证失败
- "优惠券已过期" - 优惠券不在有效期内
- "优惠券不在使用时间区间内" - 不在 `DayStartTime ~ DayEndTime` 范围内
- "请刷新优惠券列表" - 优惠券核销失败，需要刷新列表

**错误处理位置**: `main/app/api/v1/cashier/cashier_desk.go:1414`

### 3. 数据一致性

- ✅ 选择优惠券后，自动将积分抵扣改为手动抵扣
- ✅ 订单已创建支付单后，不能修改优惠券
- ✅ 核销优惠券在事务中进行，失败会回滚
- ✅ 核销失败时，会取消优惠券使用记录

### 4. 业务限制

- ❌ 一个订单只能使用一个优惠券
- ❌ 一个 SaleBill 只能使用一张通用优惠券
- ❌ 订单已创建支付单后，不能修改优惠券
- ❌ 积分抵扣后订单金额为0时，不能使用优惠券

### 5. 性能优化

- ✅ 优惠券列表查询使用预加载（Preload）
- ✅ 会员优惠券按 `CouponUuid` 分组聚合
- ✅ 使用缓存减少数据库查询

---

## 相关文档

- [订单支付流程](./ordering_checkout_flow.md)
- [会员系统架构](./member.md)
- [营销活动系统](./marketing.md)

---

**最后更新**: 2025-11-21  
**维护者**: TTPOS Team

