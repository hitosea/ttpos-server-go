# Takeout Order 字段映射对照表

> **参考文档**: [Lineman API定义及TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)

本文档详细说明 Grab、Lineman 和 TakeoutOrder 统一模型之间的字段映射关系。

---

## 📊 订单主表字段映射

| Lineman API 字段 | 类型 | 统一模型字段 | 对应 Grab SDK 字段 | 转换说明 |
|------------------|------|-------------|-------------------|---------|
| `orderId` | String(20) | `OrderID` | `orderID` | 直接映射 |
| `orderShortCode` | String(4) | `ShortOrderNumber` | `shortOrderNumber` | 直接映射 |
| `storeId` | String | `PartnerMerchantID` | - | **映射到 PartnerMerchantID** (TTPOS 侧店铺 ID) |
| `partnerId` | String | - | - | **不使用** |
| - | - | `MerchantID` | `merchantID` | **Lineman 时为空** |
| `restaurantRevenue` | double | `Price.Subtotal` | `price.subtotal` | 商户收入总额 |
| `restaurantRevenue` | double | `Price.EaterPayment` | `price.eaterPayment` | 用户实付金额（同值） |
| `orderAcceptedTime` | String | `OrderTime` | `orderTime` | ISO 8601 时间字符串 |
| `memberId` | String(255) | `MembershipID` | `membershipID` | **复用 Grab 字段** |
| `customerType` | String(32) | `FeatureFlags.OrderType` | `featureFlags.orderType` | **转换映射** (见下表) |
| `additionalItems[]` | Array | `AdditionalProperties[]` | - | **重命名** (扩展字段) |

### customerType 转换映射

| Lineman `customerType` | 统一模型 `FeatureFlags.OrderType` | 业务含义 |
|------------------------|----------------------------------|---------|
| `DELIVERY` | `Delivery` | 外送配送 |
| `PICKUP` | `Pickup` | 自取 |

**反向转换**:

| 统一模型 `FeatureFlags.OrderType` | Lineman `customerType` | 说明 |
|----------------------------------|------------------------|------|
| `Delivery` | `DELIVERY` | 外送配送 |
| `Pickup` | `PICKUP` | 自取 |
| `DineIn` | `PICKUP` | 堂食视为自取 |
| (其他) | `DELIVERY` | 默认外送 |

---

## 📦 订单商品字段映射

| Lineman API 字段 | 类型 | 统一模型字段 | 对应 Grab SDK 字段 | 转换说明 |
|------------------|------|-------------|-------------------|---------|
| `items[].id` | String(255) | `Items[].ID` | `items[].id` | 商品 ID |
| `items[].quantity` | int | `Items[].Quantity` | `items[].quantity` | 商品数量 |
| `items[].unitPrice` | double | `Items[].Price` | `items[].price` | 单价 (THB, 已含选项费用和折扣) |
| `items[].memo` | String | `Items[].Specifications` | `items[].specifications` | **字段复用** (顾客备注) |
| `items[].promotionId` | String(255) | `Promos[].Code` | `promos[].code` | **提升到订单级别** |
| `items[].discount` | double | `Promos[].MexFundedAmount` | `promos[].mexFundedAmount` | **提升到订单级别** (商户补贴) |
| `items[].properties[]` | Array | `Items[].Modifiers[]` | `items[].modifiers[]` | **结构转换** (见下节) |

### 促销信息转换（商品级别 → 订单级别）

#### 关键设计决策

Lineman 的促销信息是**商品级别**的（`items[].promotionId` 和 `items[].discount`），而 Grab 的促销信息是**订单级别**的（`order.promos[]`）。

为了保持与 Grab 的结构一致，在转换时将商品级别的促销信息**提升到订单级别**：

#### Lineman → 统一模型（提升转换）

```go
// Lineman 请求
{
  "items": [
    {
      "id": "item-001",
      "promotionId": "PROMO-A",
      "discount": 10.0
    },
    {
      "id": "item-002",
      "promotionId": "PROMO-A",
      "discount": 5.0
    }
  ]
}

// 转换为统一模型（聚合到订单级别）
{
  "items": [
    {"id": "item-001"},
    {"id": "item-002"}
  ],
  "promos": [
    {
      "code": "PROMO-A",
      "mexFundedAmount": 15.0  // 10.0 + 5.0（相同促销累加）
    }
  ]
}
```

**转换规则**：
1. 遍历所有商品的 `promotionId` 和 `discount`
2. 按 `promotionId` 分组，累加相同促销的 `discount`
3. 转换为订单级别的 `Promos[]` 数组

#### 统一模型 → Lineman（降级转换）

反向转换时，由于信息丢失（无法知道哪个商品对应哪个促销），采用**折衷方案**：

```go
// 统一模型
{
  "items": [
    {"id": "item-001"},
    {"id": "item-002"}
  ],
  "promos": [
    {
      "code": "PROMO-A",
      "mexFundedAmount": 15.0
    }
  ]
}

// 转换为 Lineman（折衷方案：仅设置到第一个商品）
{
  "items": [
    {
      "id": "item-001",
      "promotionId": "PROMO-A",
      "discount": 15.0  // 将订单级别的促销信息设置到第一个商品
    },
    {
      "id": "item-002"
      // 其他商品不包含促销信息
    }
  ]
}
```

**限制说明**：
- ⚠️ 反向转换会丢失促销与商品的精确对应关系
- ⚠️ 仅用于测试或特殊场景，实际业务中应避免反向转换
- ✅ 正向转换（Lineman → 统一模型 → Grab）无信息丢失

---

### items[].properties ↔ items[].Modifiers 结构转换

#### Lineman Properties (嵌套结构)

```json
{
  "properties": [
    {
      "id": "PROP-FLAVOR",
      "values": [
        {"id": "MILD", "price": 0},
        {"id": "SPICY", "price": 5}
      ]
    },
    {
      "id": "PROP-SIZE",
      "values": [
        {"id": "LARGE", "price": 20}
      ]
    }
  ]
}
```

#### 统一模型 Modifiers (扁平结构 + Values 保留)

```json
{
  "modifiers": [
    {
      "id": "PROP-FLAVOR",
      "quantity": 1,
      "price": 5.0,
      "values": [
        {"id": "MILD", "price": 0},
        {"id": "SPICY", "price": 5}
      ]
    },
    {
      "id": "PROP-SIZE",
      "quantity": 1,
      "price": 20.0,
      "values": [
        {"id": "LARGE", "price": 20}
      ]
    }
  ]
}
```

#### Grab SDK Modifiers (纯扁平结构)

```json
{
  "modifiers": [
    {"id": "MOD-001", "quantity": 1, "price": 1000}
  ]
}
```

**转换规则**:

1. **Lineman → 统一模型**:
   - 每个 `property` 映射为一个 `modifier`
   - `property.id` → `modifier.id`
   - `property.values` 保留在 `modifier.values` 中
   - `modifier.price` = sum of all `values[].price`
   - `modifier.quantity` 默认为 `1`

2. **Grab → 统一模型**:
   - 直接映射，`modifier.values` 为空
   - 价格单位转换 (分 → 元)

3. **统一模型 → Lineman**:
   - 如果 `modifier.values` 存在，转换为 `property` 嵌套结构
   - 如果 `modifier.values` 为空，创建单个 value

4. **统一模型 → Grab**:
   - 转换为纯扁平结构
   - 价格单位转换 (元 → 分)

---

## 💰 价格单位转换

### Grab SDK (最小单位)

- **单位**: 分 (satang/cent)
- **类型**: `int64`
- **示例**: `33389` = 333.89 THB

**转换公式**:
```go
// Grab → 统一模型
priceInBaht := float64(priceInSatang) / 100

// 统一模型 → Grab
priceInSatang := int64(priceInBaht * 100)
```

### Lineman API (标准单位)

- **单位**: 泰铢 (THB)
- **类型**: `double` (float64)
- **示例**: `333.89` = 333.89 THB

**转换**:
```go
// Lineman → 统一模型（无需转换）
order.Price.Subtotal = req.RestaurantRevenue

// 统一模型 → Lineman（无需转换）
req.RestaurantRevenue = order.Price.Subtotal
```

### 统一模型 TakeoutOrder

- **单位**: 泰铢 (THB)
- **类型**: `float64`
- **设计原因**: 避免精度损失，与 Lineman 对齐

---

## 🔄 关键字段映射详解

### 1. memo ↔ specifications (商品备注)

**Lineman**: `items[].memo` - 顾客对商品的备注 (如 "不要辣")

**Grab**: `items[].specifications` - 商品的规格说明/备注

**映射**: 直接复用 Grab 的 `Specifications` 字段

```go
// Lineman → 统一模型
if item.Memo != "" {
    takeoutItem.Specifications = &item.Memo
}

// 统一模型 → Lineman
if item.Specifications != nil {
    linemanItem.Memo = *item.Specifications
}
```

### 2. memberId ↔ membershipID (会员 ID)

**Lineman**: `memberId` - LINE MAN 绑定的会员 ID

**Grab**: `membershipID` - Grab 会员 ID

**映射**: 直接复用 Grab 的 `MembershipID` 字段

```go
// Lineman → 统一模型
if req.MemberId != "" {
    order.MembershipID = &req.MemberId
}

// 统一模型 → Lineman
if o.MembershipID != nil {
    req.MemberId = *o.MembershipID
}
```

### 3. customerType ↔ featureFlags.orderType (订单类型)

**Lineman**: `customerType` - 订单配送方式
- `DELIVERY`: 外送配送
- `PICKUP`: 自取

**Grab**: `featureFlags.orderType` - 订单类型
- `Delivery`: 配送
- `Pickup`: 自取
- `DineIn`: 堂食

**映射**: 转换为 `FeatureFlags.OrderType`

```go
// Lineman → 统一模型
orderTypeMapping := map[string]string{
    "DELIVERY": "Delivery",
    "PICKUP":   "Pickup",
}
if mappedType, ok := orderTypeMapping[req.CustomerType]; ok {
    order.FeatureFlags = &message.TakeoutFeatureFlags{
        OrderAcceptedType: "AUTO",
        OrderType:         mappedType,
    }
}

// 统一模型 → Lineman
if o.FeatureFlags != nil {
    switch o.FeatureFlags.OrderType {
    case "Delivery":
        req.CustomerType = "DELIVERY"
    case "Pickup":
        req.CustomerType = "PICKUP"
    case "DineIn":
        req.CustomerType = "PICKUP" // 堂食视为自取
    default:
        req.CustomerType = "DELIVERY"
    }
}
```

### 4. additionalItems ↔ additionalProperties (订单附加项)

**Lineman**: `additionalItems[]` - 订单附加信息列表
```json
{
  "additionalItems": [
    {"name": "ไม่รับช้อนส้อมพลาสติก"}
  ]
}
```

**Grab**: 无直接对应字段 (但 SDK 有 `AdditionalProperties map[string]interface{}`)

**映射**: 重命名为 `AdditionalProperties` (对齐 Grab SDK 概念)

```go
// Lineman → 统一模型
for _, addItem := range req.AdditionalItems {
    order.AdditionalProperties = append(order.AdditionalProperties, message.TakeoutAdditionalProperty{
        Name: addItem.Name,
    })
}

// 统一模型 → Lineman
for _, addProp := range o.AdditionalProperties {
    req.AdditionalItems = append(req.AdditionalItems, linemanv1.OrderAdditionalItem{
        Name: addProp.Name,
    })
}
```

---

## 📋 完整映射表

### 订单级别字段

| Lineman 字段 | 统一模型字段 | Grab 字段 | 映射类型 | 说明 |
|-------------|-------------|----------|---------|------|
| `orderId` | `OrderID` | `orderID` | 直接 | 订单 ID |
| `orderShortCode` | `ShortOrderNumber` | `shortOrderNumber` | 直接 | 短单号 |
| `storeId` | `MerchantID` | `merchantID` | 直接 | 商户 ID |
| `partnerId` | `PartnerMerchantID` | `partnerMerchantID` | 直接 | 合作商户 ID |
| `restaurantRevenue` | `Price.Subtotal` | `price.subtotal` | 直接 | 商品小计 |
| `restaurantRevenue` | `Price.EaterPayment` | `price.eaterPayment` | 直接 | 用户实付（同值） |
| `orderAcceptedTime` | `OrderTime` | `orderTime` | 直接 | 下单时间 |
| `memberId` | `MembershipID` | `membershipID` | **复用** | 会员 ID |
| `customerType` | `FeatureFlags.OrderType` | `featureFlags.orderType` | **转换** | 订单类型 |
| `additionalItems[]` | `AdditionalProperties[]` | - | **重命名** | 附加属性 |
| - | `PaymentType` | `paymentType` | 默认 | Lineman 默认 "CASH" |

### 商品级别字段

| Lineman 字段 | 统一模型字段 | Grab 字段 | 映射类型 | 说明 |
|-------------|-------------|----------|---------|------|
| `items[].id` | `Items[].ID` | `items[].id` | 直接 | 商品 ID |
| - | `Items[].GrabItemID` | `items[].grabItemID` | - | Grab 特有 |
| `items[].quantity` | `Items[].Quantity` | `items[].quantity` | 直接 | 数量 |
| `items[].unitPrice` | `Items[].Price` | `items[].price` | 直接/转换 | 价格 (Grab 需 `/100`) |
| `items[].memo` | `Items[].Specifications` | `items[].specifications` | **复用** | 商品备注 |
| `items[].promotionId` | `Promos[].Code` | `promos[].code` | **提升到订单级别** | 促销代码 |
| `items[].discount` | `Promos[].MexFundedAmount` | `promos[].mexFundedAmount` | **提升到订单级别** | 商户补贴金额 |
| `items[].properties[]` | `Items[].Modifiers[]` | `items[].modifiers[]` | **结构转换** | 属性/修改项 |
| - | `Items[].Tax` | `items[].tax` | - | Grab 特有 |
| - | `Items[].OutOfStockInstruction` | `items[].outOfStockInstruction` | - | Grab 特有 |

---

## 🎯 设计决策说明

### 为什么复用 Grab 字段？

1. **MembershipID 而非 MemberID**
   - ✅ Grab SDK 使用 `membershipID`
   - ✅ 语义更清晰 (Membership = 会员身份)
   - ✅ 避免字段冗余

2. **Specifications 而非 Memo**
   - ✅ Grab SDK 使用 `specifications`
   - ✅ 语义更通用 (可以是备注，也可以是规格说明)
   - ✅ 根据 Google Sheets 第 24 行映射关系

3. **FeatureFlags.OrderType 而非 CustomerType**
   - ✅ Grab SDK 在 `featureFlags.orderType` 中定义订单类型
   - ✅ 语义对齐 (都表示订单类型/配送方式)
   - ✅ 避免字段冗余

4. **AdditionalProperties 而非 AdditionalItems**
   - ✅ 对齐 Grab SDK 的 `AdditionalProperties` 概念
   - ✅ 命名更一致 (Properties 表示属性)
   - ✅ Grab SDK 在结构体中有 `AdditionalProperties map[string]interface{}`

### 字段优化前后对比

| 概念 | 优化前 | 优化后 | 优势 |
|------|--------|--------|------|
| 会员 ID | `MemberID` (Lineman) + `MembershipID` (Grab) | `MembershipID` (统一) | 减少冗余，统一命名 |
| 订单类型 | `CustomerType` (Lineman) + `FeatureFlags.OrderType` (Grab) | `FeatureFlags.OrderType` (统一) | 语义对齐，结构一致 |
| 商品备注 | `Memo` (Lineman) + `Specifications` (Grab) | `Specifications` (统一) | 复用 Grab 字段 |
| 附加属性 | `AdditionalItems` | `AdditionalProperties` | 对齐 Grab SDK 概念 |

---

## 🔍 使用场景示例

### 场景 1: 判断订单类型

```go
// 优化前（字段冗余）
if order.CustomerType != nil && *order.CustomerType == "PICKUP" {
    // Lineman 自取订单
} else if order.FeatureFlags != nil && order.FeatureFlags.OrderType == "Pickup" {
    // Grab 自取订单
}

// 优化后（统一字段）
if order.FeatureFlags != nil && order.FeatureFlags.OrderType == "Pickup" {
    // 所有平台的自取订单（统一处理）
}
```

### 场景 2: 获取会员 ID

```go
// 优化前（字段冗余）
var memberID string
if order.MemberID != nil {
    memberID = *order.MemberID // Lineman
} else if order.MembershipID != nil {
    memberID = *order.MembershipID // Grab
}

// 优化后（统一字段）
var memberID string
if order.MembershipID != nil {
    memberID = *order.MembershipID // 所有平台
}
```

### 场景 3: 获取商品备注

```go
// 优化前（字段冗余）
var note string
if item.Memo != nil {
    note = *item.Memo // Lineman
} else if item.Specifications != nil {
    note = *item.Specifications // Grab
}

// 优化后（统一字段）
var note string
if item.Specifications != nil {
    note = *item.Specifications // 所有平台
}
```

---

## ✅ 优化成果

### 减少的冗余字段

1. ❌ 删除 `CustomerType` → ✅ 复用 `FeatureFlags.OrderType`
2. ❌ 删除 `MemberID` → ✅ 复用 `MembershipID`
3. ❌ 删除 `Memo` → ✅ 复用 `Specifications`
4. ✅ 重命名 `AdditionalItems` → `AdditionalProperties` (对齐命名)

### 代码简化效果

- **字段数量**: 减少 3 个冗余字段
- **转换复杂度**: 降低 (无需在多个字段间选择)
- **语义清晰度**: 提升 (统一使用 Grab 术语)
- **可维护性**: 提升 (字段映射关系更明确)

---

## 📚 参考文档

- [Google Sheets - Lineman API定义及TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)
- [Grab SDK 文档](https://github.com/grab/grabfood-api-sdk-go)
- [数据模型文档](./README.md)
- [转换工具文档](../../../ttpos-bmp/app/ttpos-takeout/utility/README.md)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**维护者**: TTPOS Team
