# LINE MAN Webhook API 集成文档

> TTPOS 系统与 LINE MAN 外卖平台的 Webhook API 集成指南

---

## 📋 概述

本文档描述 TTPOS 系统接收 LINE MAN 平台 Webhook 回调的 API 接口规范。LINE MAN 通过 Webhook 将订单事件（创建、更新、状态变更）推送到 TTPOS 系统。

### 功能特性

- ✅ **订单创建**：接收 LINE MAN 新订单推送
- ✅ **订单内容更新**：接收订单商品/金额变更通知
- ✅ **订单状态更新**：接收订单完成/取消通知
- ✅ **状态映射**：自动映射 LINE MAN 状态为 TTPOS 内部状态
- ✅ **幂等性保证**：重复请求不重复处理
- ✅ **消息队列集成**：通过 RocketMQ 通知 Main 模块

---

## 🔐 认证机制

### Bearer Token 认证

所有 Webhook 请求需要包含有效的 Bearer Token：

```http
Authorization: Bearer {access_token}
```

**Token 获取方式**：
- LINE MAN OAuth 2.0 认证
- Token 有效期：根据 LINE MAN 平台配置
- Token 刷新：自动刷新机制

**参考文档**：
- Spec: `tech-takeout-lineman-oauth-token`
- 实现：`ttpos-bmp/app/ttpos-takeout/internal/client/lineman/token_client.go`

---

## 📡 Webhook 接口

### 1. 订单状态更新 Webhook

接收 LINE MAN 订单完成或取消通知，更新订单状态到 TTPOS 数据库。

#### 基本信息

| 项目 | 值 |
|------|-----|
| **端点** | `POST /v1/partners/{partnerId}/stores/{storeId}/order/status` |
| **方法** | POST |
| **认证** | Bearer Token |
| **Content-Type** | application/json |

#### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `partnerId` | string | ✅ | 合作伙伴唯一 ID |
| `storeId` | string | ✅ | 门店唯一 ID（LINE MAN 门店 ID） |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `orderId` | string | ✅ | 订单唯一 ID（格式：LMF-yyMMdd-{number}） | `"LMF-260113-338798091"` |
| `orderStatus` | string | ✅ | 订单状态（`FINISH` / `CANCELED`） | `"FINISH"` |

**orderStatus 可选值**：

| LINE MAN 状态 | 说明 | TTPOS 映射状态 |
|--------------|------|----------------|
| `FINISH` | 订单已完成（骑手完成配送） | `COMPLETED` |
| `CANCELED` | 订单已取消（顾客或平台取消） | `CANCELLED` |

#### 请求示例

```http
POST /v1/partners/partner-123/stores/store-456/order/status HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "orderId": "LMF-260113-338798091",
  "orderStatus": "FINISH"
}
```

#### 响应格式

##### 成功响应（200 OK）

```json
{
  "status": "ok",
  "code": "200",
  "message": "Order status updated successfully"
}
```

##### 失败响应（200 OK，业务失败）

订单不存在：

```json
{
  "status": "fail",
  "code": "500",
  "message": "订单不存在"
}
```

数据库更新失败：

```json
{
  "status": "fail",
  "code": "500",
  "message": "更新订单状态失败: {详细错误信息}"
}
```

##### HTTP 错误响应

| HTTP 状态码 | 说明 |
|------------|------|
| 400 | 请求参数错误（缺少必填参数或格式错误） |
| 401 | 未授权（Token 无效或过期） |
| 404 | 路径不存在（partnerId 或 storeId 无效） |

#### 业务逻辑

1. **状态映射**
   - 将 LINE MAN 状态映射为 TTPOS 内部状态
   - `FINISH` → `COMPLETED`
   - `CANCELED` → `CANCELLED`（注意拼写差异）

2. **订单查询**
   - 根据 `orderId` 和 `provider_name = "lineman"` 查询订单
   - 订单不存在时返回错误

3. **幂等性检查**
   - 检查当前状态是否与目标状态相同
   - 如果状态未变化，跳过更新（返回成功）

4. **订单状态更新**
   - 更新 `order_status` 字段为映射后的状态
   - 更新 `updated_at` 字段为当前时间戳

5. **RocketMQ 事件发送**
   - 构造 `OrderEvent` 消息
   - 发送到 Topic: `takeout_grab_order`
   - 失败时只记录日志，不影响主流程

#### 幂等性保证

- **场景**：LINE MAN 可能重复发送相同的状态通知
- **处理**：检查订单当前状态，如果已是目标状态则跳过更新
- **返回**：返回成功响应（`{"status": "ok"}`）

**示例**：
1. 第一次收到 `{"orderId": "LMF-xxx", "orderStatus": "FINISH"}`
   - 订单状态从 `ACCEPTED` 更新为 `COMPLETED`
   - 返回 `{"status": "ok"}`
2. 第二次收到相同请求（LINE MAN 重试）
   - 检测到订单状态已是 `COMPLETED`
   - 跳过更新，直接返回 `{"status": "ok"}`

#### 错误处理

| 错误场景 | HTTP 码 | status | code | message |
|---------|---------|--------|------|---------|
| 订单不存在 | 200 | fail | 500 | 订单不存在 |
| 数据库更新失败 | 200 | fail | 500 | 更新订单状态失败: {error} |
| 参数格式错误 | 400 | - | - | Bad Request |
| Token 无效 | 401 | - | - | Unauthorized |

---

### 2. 订单创建 Webhook

接收 LINE MAN 新订单推送。

#### 基本信息

| 项目 | 值 |
|------|-----|
| **端点** | `POST /v1/partners/{partnerId}/stores/{storeId}/orders` |
| **方法** | POST |
| **认证** | Bearer Token |
| **Content-Type** | application/json |

#### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `partnerId` | string | ✅ | 合作伙伴唯一 ID |
| `storeId` | string | ✅ | 门店唯一 ID（LINE MAN 门店 ID） |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `orderId` | string | ✅ | 订单唯一 ID（格式：LMF-yyMMdd-{number}，1-20字符） | `"LMF-260113-338798091"` |
| `orderShortCode` | string | ✅ | 短订单 ID（orderId 后4位） | `"8091"` |
| `restaurantRevenue` | number | ✅ | 商户收入总额（THB，已扣除合作伙伴补贴） | `350.00` |
| `orderAcceptedTime` | string | ✅ | 订单接受时间（ISO 8601 格式） | `"2022-11-01T13:08:06+07:00"` |
| `customerType` | string | ✅ | 订单类型（`DELIVERY` / `PICKUP`） | `"DELIVERY"` |
| `memberId` | string | ❌ | 绑定 LINE MAN 账号的会员 ID（最大255字符） | `"member-123"` |
| `items` | array | ✅ | 订单商品列表（至少1项） | 见下方结构 |
| `additionalItems` | array | ✅ | 订单附加项列表 | 见下方结构 |

#### items 商品项结构

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `id` | string | ✅ | 菜单项 ID（最大255字符） | `"menu-item-001"` |
| `quantity` | int | ✅ | 商品数量（≥1） | `2` |
| `unitPrice` | number | ✅ | 商品单价（THB，含选项费用，已折扣） | `89.00` |
| `memo` | string | ❌ | 顾客备注 | `"不要辣"` |
| `promotionId` | string | ❌ | 促销活动 ID（最大255字符） | `"promo-001"` |
| `discount` | number | ❌ | 促销折扣金额（商户补贴） | `10.00` |
| `properties` | array | ❌ | 商品选项列表 | 见下方结构 |

#### items.properties 商品选项结构

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `id` | string | ✅ | 选项 ID（最大255字符） | `"size"` |
| `values` | array | ✅ | 已选择的选项值列表（至少1项） | 见下方结构 |

#### items.properties.values 选项值结构

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `id` | string | ✅ | 选项值 ID（最大255字符） | `"large"` |
| `price` | number | ✅ | 选项值价格（THB，≥0） | `20.00` |

#### additionalItems 附加项结构

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `name` | string | ✅ | 附加信息（最大1024字符） | `"ไม่รับช้อนส้อมพลาสติก"` |

#### 请求示例

```http
POST /v1/partners/partner-123/stores/store-456/orders HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "orderId": "LMF-260113-338798091",
  "orderShortCode": "8091",
  "restaurantRevenue": 350.00,
  "orderAcceptedTime": "2026-01-13T14:30:00+07:00",
  "customerType": "DELIVERY",
  "memberId": "member-123",
  "items": [
    {
      "id": "menu-item-001",
      "quantity": 2,
      "unitPrice": 109.00,
      "memo": "不要辣",
      "promotionId": "promo-winter",
      "discount": 10.00,
      "properties": [
        {
          "id": "size",
          "values": [
            {
              "id": "large",
              "price": 20.00
            }
          ]
        },
        {
          "id": "topping",
          "values": [
            {
              "id": "cheese",
              "price": 15.00
            }
          ]
        }
      ]
    },
    {
      "id": "menu-item-002",
      "quantity": 1,
      "unitPrice": 132.00,
      "properties": []
    }
  ],
  "additionalItems": [
    {
      "name": "ไม่รับช้อนส้อมพลาสติก"
    }
  ]
}
```

#### 响应格式

##### 成功响应（200 OK）

```json
{
  "status": "ok",
  "code": "200",
  "message": "Order created successfully"
}
```

##### 失败响应

```json
{
  "status": "fail",
  "code": "500",
  "message": "Failed to create order: {error details}"
}
```

---

### 3. 订单内容更新 Webhook

接收 LINE MAN 订单内容变更通知（商品/金额调整）。当顾客修改订单或平台调整订单时，LINE MAN 会推送此通知。

#### 基本信息

| 项目 | 值 |
|------|-----|
| **端点** | `PUT /v1/partners/{partnerId}/stores/{storeId}/orders` |
| **方法** | PUT |
| **认证** | Bearer Token |
| **Content-Type** | application/json |

#### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `partnerId` | string | ✅ | 合作伙伴唯一 ID |
| `storeId` | string | ✅ | 门店唯一 ID（LINE MAN 门店 ID） |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `orderId` | string | ✅ | 订单唯一 ID（格式：LMF-yyMMdd-{number}，1-20字符） | `"LMF-260113-338798091"` |
| `orderShortCode` | string | ✅ | 短订单 ID（orderId 后4位） | `"8091"` |
| `restaurantRevenue` | number | ✅ | 更新后的商户收入总额（THB） | `280.00` |
| `orderAcceptedTime` | string | ✅ | 订单接受时间（ISO 8601 格式） | `"2026-01-13T14:30:00+07:00"` |
| `orderUpdatedTime` | string | ✅ | 订单更新时间（ISO 8601 格式） | `"2026-01-13T14:45:00+07:00"` |
| `customerType` | string | ✅ | 订单类型（`DELIVERY` / `PICKUP`） | `"DELIVERY"` |
| `memberId` | string | ❌ | 绑定 LINE MAN 账号的会员 ID | `"member-123"` |
| `items` | array | ✅ | 更新后的订单商品列表（至少1项） | 结构同订单创建 |
| `additionalItems` | array | ✅ | 订单附加项列表 | 结构同订单创建 |

**items / additionalItems 结构**：与订单创建接口相同，参见上方 [订单创建 Webhook](#2-订单创建-webhook) 的详细说明。

#### 请求示例

```http
PUT /v1/partners/partner-123/stores/store-456/orders HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "orderId": "LMF-260113-338798091",
  "orderShortCode": "8091",
  "restaurantRevenue": 280.00,
  "orderAcceptedTime": "2026-01-13T14:30:00+07:00",
  "orderUpdatedTime": "2026-01-13T14:45:00+07:00",
  "customerType": "DELIVERY",
  "items": [
    {
      "id": "menu-item-001",
      "quantity": 1,
      "unitPrice": 109.00,
      "properties": []
    },
    {
      "id": "menu-item-002",
      "quantity": 1,
      "unitPrice": 132.00,
      "properties": []
    }
  ],
  "additionalItems": []
}
```

#### 响应格式

##### 成功响应（200 OK）

```json
{
  "status": "ok",
  "code": "200",
  "message": "Order updated successfully"
}
```

##### 失败响应

```json
{
  "status": "fail",
  "code": "500",
  "message": "Order not found"
}
```

#### 业务逻辑

1. **订单查询**：根据 `orderId` 查找已存在的订单
2. **订单验证**：确认订单状态允许更新（未完成/未取消）
3. **数据更新**：
   - 更新商品列表（全量替换）
   - 更新商户收入金额
   - 记录更新时间
4. **事件推送**：发送 RocketMQ 消息通知 Main 模块

---

## 🔄 订单生命周期

```mermaid
sequenceDiagram
    participant LM as LINE MAN Platform
    participant BMP as TTPOS BMP
    participant MQ as RocketMQ
    participant Main as TTPOS Main Module

    Note over LM,Main: 1. 订单创建
    LM->>BMP: POST /orders (PlaceOrder)
    BMP->>BMP: 保存订单到数据库
    BMP->>MQ: 发送 OrderEvent (action: create)
    MQ->>Main: 消费订单创建事件
    BMP-->>LM: {"status": "ok"}

    Note over LM,Main: 2. 订单内容更新（可选）
    LM->>BMP: PUT /orders (OrderUpdate)
    BMP->>BMP: 更新订单商品/金额
    BMP->>MQ: 发送 OrderEvent (action: update)
    MQ->>Main: 消费订单更新事件
    BMP-->>LM: {"status": "ok"}

    Note over LM,Main: 3. 订单状态更新（完成/取消）
    LM->>BMP: POST /order/status (OrderStatusUpdate)
    BMP->>BMP: 更新订单状态 (FINISH → COMPLETED)
    BMP->>MQ: 发送 OrderEvent (action: status_update)
    MQ->>Main: 消费订单状态变更事件
    BMP-->>LM: {"status": "ok"}
```

---

## 📊 状态映射表

### LINE MAN → TTPOS 状态映射

| LINE MAN 状态 | TTPOS 内部状态 | 说明 | 对应 OrderEvent.Status |
|--------------|----------------|------|----------------------|
| `FINISH` | `COMPLETED` | 订单已完成 | `COMPLETED` |
| `CANCELED` | `CANCELLED` | 订单已取消 | `CANCELLED` |

**注意**：
- LINE MAN 使用 `CANCELED`（一个 L）
- TTPOS 内部使用 `CANCELLED`（两个 L）
- 状态映射在 `mapLinemanStatusToTTPOS` 函数中实现

---

## 🚀 RocketMQ 事件

### OrderEvent 结构

订单状态更新后，TTPOS BMP 会发送事件到 RocketMQ：

```json
{
  "action": "status_update",
  "providerName": "lineman",
  "shopUUID": "1234567890",
  "orderUUID": "order-uuid-123",
  "orderID": "LMF-260113-338798091",
  "status": "COMPLETED",
  "timestamp": 1736751234
}
```

### Topic 配置

| 项目 | 值 |
|------|-----|
| **Topic** | `takeout_grab_order` |
| **消息格式** | JSON |
| **消费者** | Main 模块订单状态监听器 |

**注意**：复用 Grab 的 Topic，通过 `providerName` 字段区分平台。

---

## 🧪 测试指南

### Postman 测试

#### 1. 获取 Token

```http
POST /oauth/token HTTP/1.1
Host: lineman-auth.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&client_id={client_id}&client_secret={client_secret}
```

#### 2. 测试订单状态更新

```http
POST /v1/partners/partner-123/stores/store-456/order/status HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "orderId": "LMF-260113-338798091",
  "orderStatus": "FINISH"
}
```

#### 3. 验证数据库

```sql
SELECT uuid, provider_order_id, order_status, updated_at
FROM takeout_order
WHERE provider_name = 'lineman'
  AND provider_order_id = 'LMF-260113-338798091';
```

**预期结果**：`order_status = 'COMPLETED'`

#### 4. 验证 RocketMQ

- 检查 RocketMQ 控制台：Topic `takeout_grab_order` 有新消息
- 检查消息内容：`action = "status_update"`, `status = "COMPLETED"`
- 检查 Main 模块日志：确认接收到消息

---

## 🐛 故障排查

### 问题 1: 订单不存在

**症状**：返回 `{"status": "fail", "message": "订单不存在"}`

**原因**：
- LINE MAN 发送的 `orderId` 在 TTPOS 中不存在
- 订单尚未创建（PlaceOrder Webhook 未触发）

**解决**：
1. 检查订单是否已通过 PlaceOrder 创建
2. 查询数据库：`SELECT * FROM takeout_order WHERE provider_order_id = '{orderId}'`
3. 确认 LINE MAN 门店是否已激活

### 问题 2: 状态未更新

**症状**：数据库订单状态未变化

**原因**：幂等性跳过（状态已是目标状态）

**解决**：
1. 检查订单当前状态：`SELECT order_status FROM takeout_order WHERE ...`
2. 如果已是 `COMPLETED` 或 `CANCELLED`，说明是重复请求
3. 检查日志：应有 "订单状态未变化，跳过" 的 Info 日志

### 问题 3: RocketMQ 消息未送达

**症状**：Main 模块未收到订单状态变更通知

**原因**：RocketMQ 发送失败或连接问题

**解决**：
1. 检查 BMP 日志：是否有 "发送订单状态更新 MQ 事件失败" 的 Warning 日志
2. 检查 RocketMQ 配置和连接状态
3. 检查 Main 模块是否正常监听 Topic `takeout_grab_order`

**注意**：RocketMQ 发送失败不影响订单状态更新（订单状态已成功保存到数据库）

### 问题 4: Token 过期

**症状**：返回 401 Unauthorized

**原因**：Bearer Token 已过期

**解决**：
1. 重新获取 Token（通过 OAuth 2.0）
2. 检查 Token 刷新机制是否正常工作
3. 参考 `tech-takeout-lineman-oauth-token` Spec

---

## 📚 相关文档

### 内部文档

- **Spec 需求文档**: `docs/shared/specs/active/tech-takeout-lineman-order-status-update/requirements.md`
- **Spec 设计文档**: `docs/shared/specs/active/tech-takeout-lineman-order-status-update/design.md`
- **Spec 任务分解**: `docs/shared/specs/active/tech-takeout-lineman-order-status-update/tasks.md`
- **CHANGELOG**: `ttpos-bmp/CHANGELOG.md`

### 代码实现

- **Controller**: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_order_status_update.go`
- **Service 接口**: `ttpos-bmp/app/ttpos-takeout/internal/service/lineman.go`
- **Logic 实现**: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
- **API 定义**: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go`
- **常量定义**: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`

### 外部参考

- **LINE MAN API 规范**: [Google Sheets - Order Status Update Notification API](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=102046225#gid=102046225)
- **LINE MAN 开发者文档**: [LINE MAN Developer Portal](https://developers.lineman.in.th/)

---

## 📝 版本历史

| 版本 | 日期 | 作者 | 变更内容 |
|------|------|------|----------|
| v1.1.0 | 2026-01-14 | Claude | 补充订单创建和订单更新接口的详细参数说明，包括嵌套结构和请求示例 |
| v1.0.0 | 2026-01-13 | rikugun | 初始版本，添加订单状态更新 Webhook 文档 |

---

**维护者**: TTPOS 后端开发组
**最后更新**: 2026-01-14
