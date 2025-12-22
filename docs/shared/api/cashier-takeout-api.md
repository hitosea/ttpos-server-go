# 收银端外卖订单 API 文档

**模块**: 外卖订单管理  
**终端**: POS（收银端）  
**版本**: v1.0.0  
**更新时间**: 2025-12-22

---

## 📋 API 列表

### 1. 获取外卖订单列表

**接口**: `GET /api/v1/cashier/takeout/order/list`

**描述**: 获取外卖订单列表，支持多条件筛选和分页

**请求头**:
```
Authorization: Bearer {token}
```

**查询参数**:
- `page_no`: 页码（必填）
- `page_size`: 每页数量（必填）
- `platform`: 平台筛选（可选）- grab, foodpanda, lineman
- `status`: 状态筛选（可选）- 0=全部, 1=待接单, 2=已接单, 3=制作中, 4=已完成, 5=已拒单
- `start_time`: 开始时间（可选）- Unix时间戳
- `end_time`: 结束时间（可选）- Unix时间戳
- `search`: 搜索关键词（可选）- 订单号、平台订单ID

**请求示例**:
```
GET /api/v1/cashier/takeout/order/list?page_no=1&page_size=20&platform=grab&status=1
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 37019885582417930,
        "platform": "grab",
        "platform_order_id": "123-CYNKLPCVRN5",
        "short_order_number": "GF-123",
        "order_state": 1,
        "is_abnormal": 0,
        "abnormal_detail": "",
        "stock_status": 1,
        "subtotal": 2550,
        "delivery_fee": 400,
        "total_amount": 2075,
        "currency_code": "SGD",
        "currency_symbol": "S$",
        "payment_type": "CASH",
        "order_time": 1703232000,
        "accepted_time": 0,
        "cutlery": 1,
        "order_type": "DeliveredByGrab",
        "items": [
          {
            "uuid": 37019885582417931,
            "platform_item_id": "TTPOS-ITEM-37019885582417930",
            "platform_item_name": "汉堡",
            "quantity": 1,
            "price": 2550,
            "tax": 144,
            "specifications": "less sugar and chili",
            "is_mapped": 1
          }
        ]
      }
    ],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

---

### 2. 获取外卖订单详情

**接口**: `GET /api/v1/cashier/takeout/order/detail`

**描述**: 获取指定外卖订单的详细信息

**请求头**:
```
Authorization: Bearer {token}
```

**查询参数**:
- `order_uuid`: 订单UUID（必填）

**请求示例**:
```
GET /api/v1/cashier/takeout/order/detail?order_uuid=37019885582417930
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "uuid": 37019885582417930,
    "platform": "grab",
    "platform_order_id": "123-CYNKLPCVRN5",
    "short_order_number": "GF-123",
    "order_state": 1,
    "is_abnormal": 0,
    "abnormal_detail": "",
    "stock_status": 1,
    "subtotal": 2550,
    "delivery_fee": 400,
    "total_amount": 2075,
    "currency_code": "SGD",
    "currency_symbol": "S$",
    "payment_type": "CASH",
    "order_time": 1703232000,
    "accepted_time": 0,
    "cutlery": 1,
    "order_type": "DeliveredByGrab",
    "items": [...]
  }
}
```

---

### 3. 接单

**接口**: `POST /api/v1/cashier/takeout/order/accept`

**描述**: 接受外卖订单

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "order_uuid": 37019885582417930
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "接单成功",
  "data": null
}
```

**错误示例**:
```json
{
  "code": 400,
  "message": "订单状态不正确，无法接单",
  "data": null
}
```

---

### 4. 拒单

**接口**: `POST /api/v1/cashier/takeout/order/reject`

**描述**: 拒绝外卖订单

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "order_uuid": 37019885582417930,
  "reject_reason_code": "OUT_OF_STOCK"
}
```

**拒单原因代码**:
- `OUT_OF_STOCK` - 缺货
- `TOO_BUSY` - 太忙
- `RESTAURANT_CLOSED` - 餐厅关闭
- `CANNOT_DELIVER` - 无法配送
- `OTHER` - 其他原因

**响应示例**:
```json
{
  "code": 200,
  "message": "拒单成功",
  "data": null
}
```

---

### 5. 获取外卖配置

**接口**: `GET /api/v1/cashier/takeout/settings`

**描述**: 获取指定平台的外卖配置

**请求头**:
```
Authorization: Bearer {token}
```

**查询参数**:
- `platform`: 平台名称（必填）- grab, foodpanda, lineman

**请求示例**:
```
GET /api/v1/cashier/takeout/settings?platform=grab
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "uuid": 37019885582417930,
    "platform": "grab",
    "is_enabled": true,
    "auto_accept": false,
    "max_amount": 100000
  }
}
```

---

### 6. 保存外卖配置

**接口**: `POST /api/v1/cashier/takeout/settings`

**描述**: 保存外卖配置

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "platform": "grab",
  "is_enabled": true,
  "auto_accept": false,
  "max_amount": 100000
}
```

**字段说明**:
- `platform`: 平台名称（必填）
- `is_enabled`: 是否启用外卖功能
- `auto_accept`: 是否自动接单
- `max_amount`: 自动接单的最大金额（单位：分，0表示不限制）

**响应示例**:
```json
{
  "code": 200,
  "message": "保存成功",
  "data": null
}
```

---

## 📊 数据结构

### 订单状态（order_state）

| 值 | 状态 | 说明 |
|----|------|------|
| 1 | 待接单 | 新订单，等待收银员处理 |
| 2 | 已接单 | 订单已被接受 |
| 3 | 制作中 | 订单正在制作 |
| 4 | 已完成 | 订单已完成 |
| 5 | 已拒单 | 订单被拒绝 |

### 库存状态（stock_status）

| 值 | 状态 | 说明 |
|----|------|------|
| 1 | 充足 | 所有商品库存充足 |
| 2 | 不足 | 部分商品库存不足 |

### 异常标记（is_abnormal）

| 值 | 状态 | 说明 |
|----|------|------|
| 0 | 正常 | 订单正常 |
| 1 | 异常 | 订单存在异常（商品未映射等） |

### 商品映射状态（is_mapped）

| 值 | 状态 | 说明 |
|----|------|------|
| 0 | 未映射 | 商品ID未包含TTPOS前缀，不能接单 |
| 1 | 已映射 | 商品ID已映射到TTPOS商品 |

---

## 🔐 权限要求

所有接口都需要收银端认证（Bearer Token），通过 `/api/v1/cashier/login` 登录获取。

---

## 🎯 业务流程

### 订单接收流程

```
1. 新订单到达（由BMP推送）
   ↓
2. 收银端查询订单列表
   ↓
3. 检查订单状态和商品映射
   ↓
4. 如果正常：接单（AcceptOrder）
   如果异常：拒单（RejectOrder）
   ↓
5. 订单状态更新
   ↓
6. 通知KDS（厨显）出单
```

### 自动接单流程

```
1. 配置自动接单参数
   - is_enabled = true
   - auto_accept = true
   - max_amount = 限额
   ↓
2. 新订单到达
   ↓
3. 系统自动检查
   - 订单金额 ≤ max_amount
   - 商品全部已映射
   - 库存充足
   ↓
4. 条件满足：自动接单
   条件不满足：待人工处理
```

---

## ⚠️ 注意事项

1. **商品ID格式**：所有商品必须包含 `TTPOS-ITEM-` 前缀才能正常接单
2. **时间戳格式**：所有时间字段均为Unix时间戳（秒）
3. **金额单位**：所有金额字段单位为"分"（1元 = 100分）
4. **并发处理**：接单/拒单操作会检查订单状态，避免重复操作
5. **异常订单**：包含未映射商品的订单会被标记为异常，需要人工处理

---

## 📚 相关文档

- 外卖模块设计: `docs/shared/specs/active/feature-pos-grab-order-integration/design.md`
- 字段映射: `docs/shared/specs/active/feature-pos-grab-order-integration/GRAB_FIELD_MAPPING.md`
- ID解析规则: 见 `domain/value_object/item_id_parser.go`

---

**维护者**: TTPOS Team  
**最后更新**: 2025-12-22

