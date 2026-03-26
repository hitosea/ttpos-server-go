# 先下单后付 (Order-First-Pay-Later) 前端对接文档

> 版本: v1.3 | 更新日期: 2026-03-23

## 业务流程概览

```
┌─ 会员端 ──────────────────────────────────────────────────┐
│                                                            │
│  GET /base ──→ 获取 is_order_first_pay_later 配置          │
│                                                            │
│  创建订单 ──→ 加购(可选) ──→ GET form_info ──→ 判断模式    │
│  (create)     (create)      (form_info)                    │
│                                                            │
│  ┌──────────────────────────────────────────────────┐      │
│  │ is_order_first_pay_later = true → 调 submit      │      │
│  │ is_order_first_pay_later = false → 调 pay        │      │
│  └──────────────────────────────────────────────────┘      │
│                                                            │
│  submit 后 → pending(待接单)                               │
│    ├── 会员取消 → cancelled(已取消)                         │
│    ├── 收银拒单 → rejected(已拒单)                          │
│    └── 收银接单 → preparing(备餐中)                         │
│                     └── 收银结账 → completed(已完成)         │
└────────────────────────────────────────────────────────────┘

┌─ 收银端 ──────────────────────────────────────────────────┐
│                                                            │
│  H5接单列表 ──→ 接单 ──→ 挂单列表 ──→ 取单 ──→ 即时点餐   │
│  (h5_order/list) (accept) (order/list)  (show)  (正常流程)  │
│                                                            │
│  取单后与即时点餐完全一致: 加购、送厨、结账、反结账         │
│  禁止拆单 (is_split_disabled=true)                         │
└────────────────────────────────────────────────────────────┘
```

## 通用说明

### 基础 URL
```
http://{host}/api/v1
```

### 统一响应格式
```json
{
  "code": 0,
  "message": "Request successful",
  "data": { ... }
}
```
- `code: 0` 成功, 非0为错误
- `data` 始终为对象, 不会是 `null` 或数组

### 认证方式
```
Authorization: Bearer {token}
X-TTPOS-Company-Id: {company_uuid}
Device-Id: {device_id}
TZ: {timezone}
Content-Type: application/json
```

---

## 一、Shop端 — 配置管理

### 1.1 获取门店点餐配置

```
GET /shop/setting/store_scan_order
```

**响应 data:**
```json
{
  "is_enabled": 1,
  "enable_delivery": 1,
  "enable_self_pickup": 1,
  "is_order_first_pay_later": 1,
  "delivery_available": 1,
  "self_pickup_available": 1
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `is_enabled` | int | 门店点餐启用状态: 0-关闭, 1-开启 |
| `enable_delivery` | int | 外送服务: 0-关闭, 1-开启 |
| `enable_self_pickup` | int | 到店自取: 0-关闭, 1-开启 |
| **`is_order_first_pay_later`** | int | **先下单后付: 0-先付后下单(默认), 1-先下单后付** |
| `delivery_available` | int | 外送服务是否可用(只读, 云平台控制): 0-不可用, 1-可用 |
| `self_pickup_available` | int | 到店自取是否可用(只读, 云平台控制): 0-不可用, 1-可用 |

---

### 1.2 保存门店点餐配置

```
POST /shop/setting/store_scan_order
```

**请求 body:**
```json
{
  "is_enabled": 1,
  "enable_delivery": 1,
  "enable_self_pickup": 1,
  "is_order_first_pay_later": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `is_enabled` | int | 是 | 启用状态: 0-关闭, 1-开启 |
| `enable_delivery` | int | 是 | 外送服务: 0-关闭, 1-开启 |
| `enable_self_pickup` | int | 是 | 到店自取: 0-关闭, 1-开启 |
| **`is_order_first_pay_later`** | int | 是 | **先下单后付: 0-关闭, 1-开启** |

**响应 data:** `"保存成功"` (字符串)

---

## 二、会员端 — 堂食订单

### 2.0 获取基础信息 (判断模式开关)

```
GET /member/base
```

**响应 data 中 `member` 字段新增:**

```json
{
  "member": {
    "is_order_first_pay_later": true,
    "is_open_store_scan_order": true,
    "is_open_rider": true,
    ...
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| **`is_order_first_pay_later`** | bool | **堂食是否开启先下单后付款** |

> 前端在进入堂食点餐页面前，读取此字段判断模式。

---

### 2.1 创建堂食订单

> 不论哪种模式，create 只创建订单(购物车)。可多次调用加购。

```
POST /member/order/dine_in/create
```

**请求 body:**
```json
{
  "sale_bill_uuid": 0,
  "sale_order_uuid": 0,
  "products": [
    {
      "flavor_uuid": 3717245672884228,
      "num": 1,
      "price": 8,
      "product_type": 0,
      "sauce_uuid": [],
      "attribute_uuid": []
    }
  ]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `sale_bill_uuid` | uint64 | 是 | 销售账单UUID. **0=新建订单, 非0=加购到已有订单** |
| `sale_order_uuid` | uint64 | 是 | 销售订单UUID. 加购时必填(从首次创建的响应中获取) |
| `products` | array | 是 | 商品列表 |

**products 子项 — 普通商品:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `flavor_uuid` | uint64 | 是 | 商品规格UUID(从商品列表接口获取) |
| `num` | number | 是 | 数量 |
| `price` | number | 否 | 单价(会做价格校验) |
| `product_type` | int | 是 | 商品类型: 0-普通商品, 1-套餐 |
| `sauce_uuid` | uint64[] | 否 | 加料UUID列表 |
| `attribute_uuid` | uint64[] | 否 | 属性UUID列表 |
| `remark` | string | 否 | 备注 |
| `remark_uuids` | uint64[] | 否 | 预设备注UUID列表 |

**products 子项 — 套餐商品 (product_type=1):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `flavor_uuid` | uint64 | 是 | 套餐规格UUID |
| `num` | number | 是 | 数量 |
| `price` | number | 否 | 套餐单价 |
| `product_type` | int | 是 | 固定为 1 |
| `products` | array | 是 | 套餐子商品列表 |

**套餐子商品:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `product_package_group_uuid` | uint64 | 是 | 套餐分组UUID |
| `flavor_uuid` | uint64 | 是 | 子商品规格UUID |
| `num` | number | 是 | 数量 |
| `unit_num` | number | 否 | 单位数量 |
| `sauce_uuid` | uint64[] | 否 | 加料UUID列表 |
| `attribute_uuid` | uint64[] | 否 | 属性UUID列表 |

**响应 data:**
```json
{
  "sale_bill_uuid": 3720886418933761,
  "sale_order_uuid": 3720886418933763
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `sale_bill_uuid` | uint64 | 销售账单UUID(后续所有操作使用此值) |
| `sale_order_uuid` | uint64 | 销售订单UUID |

---

### 2.2 获取提交表单信息 (判断 submit/pay)

> **核心**: 前端通过此接口返回的 `is_order_first_pay_later` 字段判断调 submit 还是 pay。

```
GET /member/order/dine_in/form_info?sale_bill_uuid={uuid}&sale_order_uuid={uuid}
```

**请求参数 (Query):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `sale_bill_uuid` | uint64 | 是 | 销售账单UUID |
| `sale_order_uuid` | uint64 | 是 | 销售订单UUID |

**响应 data:**
```json
{
  "sale_bill_uuid": 3720886418933761,
  "sale_order_uuid": 3720886418933763,
  "dining_method": 0,
  "product_list": { "list": [...] },
  "amount_info": {
    "product_amount": 8,
    "tax_amount": 0,
    "service_amount": 0,
    "member_discount": 0,
    "total_amount": 8
  },
  "payment_methods": { "list": [...] },
  "remark": "",
  "is_order_first_pay_later": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `sale_bill_uuid` | uint64 | 销售账单UUID |
| `sale_order_uuid` | uint64 | 销售订单UUID |
| `dining_method` | uint | 用餐方式 0:堂食 1:打包 |
| `product_list` | object | 商品列表 |
| `amount_info` | object | 金额信息 |
| `payment_methods` | object | 支付方式列表 |
| `remark` | string | 订单备注 |
| **`is_order_first_pay_later`** | bool | **是否先下单后付款。`true` → 显示提交按钮调 submit, `false` → 显示支付按钮调 pay** |

**前端判断逻辑:**
```
if (form_info.is_order_first_pay_later) {
  // 显示"提交订单"按钮 → POST /member/order/dine_in/submit
} else {
  // 显示"立即支付"按钮 → POST /member/order/dine_in/pay
}
```

---

### 2.3 提交订单到收银机 (先下单后付专用)

> 调用后生成H5订单, 收银端可见. 同时后端将 `sale_bill.is_order_first_pay_later` 标记为 1。

```
POST /member/order/dine_in/submit
```

**请求 body:**
```json
{
  "sale_bill_uuid": 3720886418933761,
  "sale_order_uuid": 3720886418933763
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `sale_bill_uuid` | uint64 | 是 | 销售账单UUID |
| `sale_order_uuid` | uint64 | 是 | 销售订单UUID |

**响应 data:** 空 (`{}`)

**后端处理:**
1. 验证门店配置 `is_order_first_pay_later=1`
2. 将 `sale_bill.is_order_first_pay_later` 写入 1（数据库标记）
3. 创建 H5 订单, 设置 `submit_pay_time`

**错误码:**
- `"当前模式不支持此操作"` — 门店配置未开启先下单后付
- `"订单状态不允许提交"` — 订单已完成或取消
- `"订单已提交，请勿重复操作"` — 已提交过

---

### 2.4 获取订单详情

```
GET /member/order/dine_in/detail?sale_bill_uuid={sale_bill_uuid}
```

**请求参数 (Query):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `sale_bill_uuid` | uint64 | 是 | 销售账单UUID |

**响应 data:**
```json
{
  "sale_bill_uuid": 3720886418933761,
  "sale_order_uuid": 3720886418933763,
  "company_name": "白日梦想家周边店",
  "serial_no": "8575",
  "order_no": "202603231177399123",
  "status_info": {
    "status": "pending",
    "status_text": "Waiting for order acceptance"
  },
  "dining_method": 0,
  "remark": "",
  "create_time": 1774256907,
  "submit_pay_time": 1774256910,
  "pay_time": 0,
  "cancel_time": 0,
  "remaining_payment_time": 0,
  "refund_amount": 0,
  "is_order_first_pay_later": true,
  "amount_info": {
    "discount_amount": 0,
    "service_fee": 0,
    "tax_fee": 0,
    "amount": 8,
    "payment_method_name": ""
  },
  "product_list": { "list": [...] },
  "payment_methods": {"list": []}
}
```

**新增/关键字段:**

| 字段 | 类型 | 说明 |
|------|------|------|
| **`is_order_first_pay_later`** | bool | **是否先下单后付款的订单。submit 后为 true, 普通支付订单为 false** |
| `status_info.status` | string | 订单状态(见状态枚举) |
| `submit_pay_time` | int64 | 提交时间戳. 0=未提交, >0=已提交 |
| `remaining_payment_time` | int64 | 剩余支付时间(秒). **先下单后付模式始终为0** |
| `payment_methods.list` | array | 支付方式列表. **先下单后付模式为空数组** |
| `amount_info.discount_amount` | float64 | 整单折扣金额(与收银机购物车 discount_amount 一致) |

---

### 2.5 获取订单列表

```
GET /member/order/dine_in/list?status={status}&page_no={page}&page_size={size}
```

**请求参数 (Query):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `status` | string | 否 | 状态过滤: `unpaid`/`inprogress`/`completed`/`cancelled`. 不传=全部 |
| `page_no` | int | 否 | 页码, 默认1 |
| `page_size` | int | 否 | 每页数量, 默认20 |

**状态过滤映射:**

| 前端传值 | 包含的状态 |
|---------|-----------|
| `unpaid` | 待支付(普通模式未付款, **不含先下单后付的待接单**) |
| `inprogress` | 待接单(pending) + 备餐中(preparing) |
| `completed` | 已完成(completed) + 部分退款(partial_refund) + 全部退款(full_refund) |
| `cancelled` | 已取消(cancelled) + 已拒单(rejected) |

---

### 2.6 取消订单

> 先下单后付模式: 仅"待接单"(H5状态=Order)时可取消. 接单后不可取消.

```
POST /member/order/dine_in/cancel
```

**请求 body:**
```json
{
  "sale_bill_uuid": 3720886418933761
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `sale_bill_uuid` | uint64 | 是 | 销售账单UUID |

**响应 data:** 空 (`{}`)

**错误提示:**
- `"订单类型错误"` — 不是会员端堂食订单
- `"无权操作此订单"` — 不是当前会员的订单
- `"订单状态不可取消"` — 订单不是 Pending 状态
- `"订单已支付，不可取消"` — 已有支付记录
- `"订单已被接单，不可取消"` — 收银端已接单

---

## 三、收银端 — 接单管理

### 3.1 H5 接单列表

```
GET /cashier/h5_order/list?status={status}&page_no={page}&page_size={size}
```

**请求参数 (Query):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `status` | int | 是 | **0=待处理, 1=已处理** |
| `page_no` | int | 否 | 页码 |
| `page_size` | int | 否 | 每页数量 |
| `desk_region_uuid` | uint64 | 否 | 桌台区域过滤 |
| `order_type` | int | 否 | 订单类型: -1=全部, 0=桌台扫码, 1=会员端堂食 |

---

### 3.2 接单

> 先下单后付模式: 接单后订单转为即时点餐挂单(不送厨), 等待收银端取单后操作.

```
POST /cashier/h5_order/accept
```

**请求 body:**
```json
{
  "h5_order_uuid": 3720886421030915
}
```

---

### 3.3 拒单

```
POST /cashier/h5_order/reject
```

**请求 body:**
```json
{
  "h5_order_uuid": 3720886421030915
}
```

---

### 3.4 即时点餐挂单列表

```
GET /cashier/instant/order/list?page_no={page}&page_size={size}&keyword={keyword}
```

**请求参数 (Query):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| `page_no` | int | 否 | 页码 |
| `page_size` | int | 否 | 每页数量 |
| **`keyword`** | string | 否 | **搜索关键词(按流水号模糊搜索)** |

---

### 3.5 取单

> 取单后订单进入即时点餐编辑模式, 后续操作与即时点餐完全一致.

```
POST /cashier/instant/order/show
```

**请求 body:**
```json
{
  "sale_bill_uuid": 3720886418933761
}
```

**取单后:**
- 订单从挂单列表消失
- 进入即时点餐编辑界面
- 可执行: 加购、送厨、结账、反结账等所有即时点餐操作
- **禁止拆单** — 购物车 `is_split_disabled=true`
- 收银端商品变更会实时反映到会员端订单详情

---

### 3.6 购物车 `is_split_disabled` 字段

购物车响应中:

```json
{
  "is_split_disabled": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| **`is_split_disabled`** | bool | **是否禁止拆单. 会员端堂食订单为 `true`, 普通即时点餐为 `false`** |

前端在 `is_split_disabled=true` 时应隐藏拆单按钮.

---

## 四、状态枚举速查

### 会员端订单状态 (status_info.status)

| 值 | 含义 | 触发条件 |
|---|------|---------|
| `unpaid` | 待支付 | 普通模式: 创建后未支付 |
| `pending` | 待接单 | 先下单后付: submit后等待收银接单 |
| `preparing` | 备餐中 | 收银端已接单, 生产单未全部完成 |
| `completed` | 已完成 | 订单结账完成 或 生产单全部完成 |
| `partial_refund` | 部分退款 | 有退款但未全额退 |
| `full_refund` | 全部退款 | 全额退款 |
| `cancelled` | 已取消 | 会员主动取消 |
| `rejected` | 已拒单 | 收银端拒单 |

### 会员端状态流转 (先下单后付模式)

```
create → (未提交, 不在列表显示)
  ↓
form_info → is_order_first_pay_later=true → 显示"提交"按钮
  ↓
submit → pending(待接单), sale_bill.is_order_first_pay_later=1
  ├── 会员取消 → cancelled(已取消)
  ├── 收银拒单 → rejected(已拒单)
  └── 收银接单 → preparing(备餐中)
                   └── 收银结账 → completed(已完成)
```

### 先下单后付 vs 普通模式对比

| 行为 | 普通模式(先付后下单) | 先下单后付模式 |
|------|-------------------|--------------|
| 模式判断 | `form_info.is_order_first_pay_later=false` | `form_info.is_order_first_pay_later=true` |
| 创建订单后 | 显示"待支付", 需要支付 | 显示购物车, 可加购 |
| 提交方式 | 调 pay 接口支付 | 调 submit 接口提交 |
| 收银端可见时机 | 支付完成后 | submit 后 |
| 接单后 | 送厨, 订单完成 | **转为挂单, 不送厨** |
| 支付时机 | 下单前 | **收银端取单后结账** |
| 剩余支付时间 | 15分钟倒计时 | **始终为0(不显示)** |
| 支付方式列表 | 返回可用方式 | **返回空数组** |
| 取消限制 | 未支付可取消 | **仅待接单可取消, 接单后不可取消** |
| 拆单 | 支持拆单 | **禁止拆单** (`is_split_disabled=true`) |
| detail 标识 | `is_order_first_pay_later=false` | `is_order_first_pay_later=true` |

---

## 五、数据库变更

### sale_bill 表新增字段

```sql
ALTER TABLE ttpos_sale_bill
ADD COLUMN is_order_first_pay_later tinyint(1) NOT NULL DEFAULT 0
COMMENT '是否先下单后付款, 0-否 1-是'
AFTER is_split_order;
```

> 迁移文件: `admin/database/migrations/20260323120000_add_is_order_first_pay_later_to_sale_bill.php`

---

## 六、变更记录

### v1.3 (2026-03-23)

| 变更 | 说明 | 影响接口 |
|------|------|---------|
| `GET /member/base` 新增 `member.is_order_first_pay_later` | 前端可在进入点餐前获取模式 | `GET /member/base` |
| `GET /member/order/dine_in/form_info` 新增 `is_order_first_pay_later` | **前端用此字段判断调 submit 还是 pay** | `GET /member/order/dine_in/form_info` |
| `GET /member/order/dine_in/detail` 新增 `is_order_first_pay_later` | 从 sale_bill 读取, submit 后为 true | `GET /member/order/dine_in/detail` |
| submit 时写 `sale_bill.is_order_first_pay_later=1` | create 不再写标记, 改为 submit 时才写 | `POST /member/order/dine_in/submit` |
| create req 移除 `is_order_first_pay_later` 参数 | 前端无需在 create 时传此参数 | `POST /member/order/dine_in/create` |
| `sale_bill` 表新增 `is_order_first_pay_later` 列 | 数据库迁移 | - |

### v1.2 (2026-03-23)

| 变更 | 说明 |
|------|------|
| 购物车新增 `is_split_disabled` 字段 | 会员端堂食订单为 true, 前端隐藏拆单按钮 |
| 拆单接口拒绝会员端堂食订单 | 返回错误 "会员端堂食订单不支持拆单" |
| `discount_amount` 改为仅包含整单折扣 | 与收银机购物车 discount_amount 口径一致 |
| 挂单列表新增 `keyword` 搜索参数 | 按流水号模糊搜索 |

### v1.1 (2026-03-23)

基于提交 `5f4f4c35c` 拆分 create/submit 接口。

### v1.0 (2026-03-23)

初始版本, 实现先下单后付模式基础流程。
