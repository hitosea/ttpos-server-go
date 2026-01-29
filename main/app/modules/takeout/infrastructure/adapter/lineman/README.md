# LINE MAN 订单转换器实现文档

## 概述

本模块实现了 LINE MAN Place Order API 的订单数据转换，将 LINE MAN 的订单格式转换为 TTPOS 系统的通用订单格式。

## 文件结构

```
main/app/modules/takeout/infrastructure/adapter/lineman/
├── lineman_order_converter.go       # LINE MAN 订单转换器实现
└── lineman_order_converter_test.go  # 单元测试
```

## LINE MAN Place Order API 字段映射

### 基础字段映射

| LINE MAN 字段 | TTPOS 字段 | 说明 |
|--------------|-----------|------|
| `orderId` | `PlatformOrderId` | 订单唯一ID，格式：LMF-yyMMdd-{number} |
| `orderShortCode` | `ShortOrderNumber` | 订单短码（orderId 的最后 4 位数字） |
| `restaurantRevenue` | `Subtotal` | 餐厅收入（已扣除平台补贴） |
| `orderAcceptedTime` | `OrderTime` | 订单接受时间（ISO 8601 格式） |
| `customerType` | `OrderType` | 订单类型：DELIVERY → DELIVERY, PICKUP → TAKEAWAY |
| `memberId` | `MembershipId` | LINE MAN 会员 ID |

### 商品字段映射

| LINE MAN 字段 | TTPOS 字段 | 说明 |
|--------------|-----------|------|
| `items[].id` | `PlatformItemId` | 菜品 ID |
| `items[].quantity` | `Quantity` | 数量 |
| `items[].unitPrice` | `Price` | 单价（含选项加价和折扣） |
| `items[].memo` | `Specifications` | 备注 |
| `items[].promotionId` | `PromoCode` | 促销活动 ID |
| `items[].discount` | `PromoAmount` | 折扣金额 |

### 商品选项映射

LINE MAN 的 `properties` 数组被转换为 TTPOS 的修饰符（Modifiers）：

```
items[].properties[] → TakeoutOrderItemModifiers[]
  - properties[].id + values[].id → PlatformModifierId
  - values[].price → Price
```

### 货币信息

LINE MAN 使用泰铢（THB）作为货币单位：

- `CurrencyCode`: "THB"
- `CurrencySymbol`: "฿"
- `CurrencyExponent`: 0（无小数位）

## 核心功能

### 1. ParseOrderWebhook

解析 LINE MAN Webhook 推送的订单数据：

```go
func (c *LineManConverter) ParseOrderWebhook(rawData []byte) (interface{}, error)
```

**输入**: LINE MAN Place Order API 的 JSON 数据
**输出**: `*LineManPlaceOrderRequest` 结构体
**验证**: 检查必填字段（orderId, orderShortCode, customerType）

### 2. ConvertOrderToTakeoutOrder

将 LINE MAN 订单转换为 TTPOS 通用订单格式：

```go
func (c *LineManConverter) ConvertOrderToTakeoutOrder(
    orderUuid uint64,
    platform string,
    platformOrderId string,
    platformOrder interface{},
    rawDataJSON []byte,
    currentTime int64,
) (*takeoutModel.TakeoutOrder, error)
```

**功能**:
- 转换基础订单信息
- 转换商品数据（包括选项/修饰符）
- 转换促销信息
- 设置订单状态为"待接单"（PENDING）

## 订单状态映射

| LINE MAN 状态 | TTPOS 状态 | 说明 |
|--------------|-----------|------|
| NEW/PENDING/RECEIVED | TakeoutOrderStatePending (0) | 待接单 |
| ACCEPTED/PREPARING/READY | TakeoutOrderStateAccepted (1) | 已接单配餐中 |
| COLLECTED/PICKED_UP | TakeoutOrderStateRiderProcessing (3) | 骑手配送中 |
| DELIVERED/COMPLETED | TakeoutOrderStateCompleted (4) | 已完成 |
| REJECTED | TakeoutOrderStateRejected (5) | 已拒单 |
| CANCELLED/CANCELED | TakeoutOrderStateCanceled (6) | 已取消 |

## 使用示例

### 接收 LINE MAN Webhook

```go
// 从 BMP 服务接收 Webhook
rawData := map[string]interface{}{
    "orderId": "LMF-221031-33879881",
    "orderShortCode": "9881",
    "restaurantRevenue": 250.50,
    "orderAcceptedTime": "2022-11-01T13:08:06+07:00",
    "items": [...],
    "customerType": "DELIVERY",
}

// 调用订单同步服务
err := orderService.SyncNewOrder(ctx, "lineman", takeoutOrderUuid, rawData)
```

### 内部处理流程

1. **平台识别**: 根据 `platform` 参数（"lineman"）选择 LINE MAN 转换器
2. **数据解析**: 调用 `ParseOrderWebhook` 解析 JSON 数据
3. **订单转换**: 调用 `ConvertOrderToTakeoutOrder` 转换为 TTPOS 格式
4. **数据存储**: 保存到 `ttpos_takeout_order` 表及关联表

## 测试覆盖

单元测试文件 `lineman_order_converter_test.go` 包含以下测试用例：

1. ✅ **TestParseOrderWebhook**: 测试 Webhook 数据解析
2. ✅ **TestConvertOrderToTakeoutOrder**: 测试订单完整转换
3. ✅ **TestParseISO8601Time**: 测试 ISO 8601 时间解析
4. ✅ **TestConvertCustomerTypeToOrderType**: 测试客户类型转换

运行测试：

```bash
cd main
go test -v ./app/modules/takeout/infrastructure/adapter/lineman/
```

## 注意事项

### LINE MAN 特有限制

1. **无收货人信息**: LINE MAN 通过平台配送，不提供详细收货人信息
2. **无税费信息**: LINE MAN 只提供餐厅净收入，不单独提供税费明细
3. **无 merchantId**: 需要从其他途径获取商户 ID
4. **货币单位**: 固定使用泰铢（THB），无需货币转换

### 与 Grab 的区别

| 特性 | Grab | LINE MAN |
|-----|------|----------|
| 订单 ID 格式 | 自定义 | LMF-yyMMdd-{number} |
| 货币转换 | 需要（分 → 元） | 不需要（泰铢） |
| 收货人信息 | 提供 | 不提供 |
| 税费信息 | 提供 | 不提供 |
| 商品选项格式 | Modifiers | Properties + Values |

## 集成点

### 服务层注册

在 `takeout_order_service.go` 中注册转换器：

```go
converters := make(map[string]service.IOrderConverter)

// 注册 Grab 转换器
grabConverter := grab.NewGrabConverter(dbm, nil)
converters[value_object.TakeoutPlatformGrab] = grabConverter

// 注册 LINE MAN 转换器
linemanConverter := lineman.NewLineManConverter()
converters[value_object.TakeoutPlatformLineman] = linemanConverter
```

### Webhook 入口

LINE MAN 的 Webhook 通过以下路径接收：

1. **BMP 服务** (`ttpos-bmp`) 接收 LINE MAN Webhook
2. **RPC 调用** BMP 将订单数据推送到 Main 服务
3. **Main 服务** 使用 LINE MAN 转换器处理订单

## 未来扩展

- [ ] 支持 LINE MAN 订单状态更新 Webhook
- [ ] 支持 LINE MAN 订单取消 Webhook
- [ ] 实现 LINE MAN 菜单转换器（如果菜单格式与 Grab 不同）
- [ ] 添加更多促销类型的处理

## 相关文档

- [LINE MAN Place Order API 规范](../../../../../../docs/shared/integrations/lineman/place-order-api.md)
- [外卖模块开发指南](../../README.md)
- [订单转换器接口](../../../domain/service/order_converter.go)

---

**最后更新**: 2026-01-12  
**维护者**: TTPOS Development Team
