# Takeout Order 快速参考

> 一页纸搞懂 Grab/Lineman 订单字段映射

---

## 🚀 快速使用

```go
import (
    "ttpos-api/ttpos-takeout/message"          // 数据模型
    "ttpos-bmp/app/ttpos-takeout/utility"      // 转换方法
)

// Grab 订单转换
grabOrder, _ := utility.FromGrabSDK(grabReq)

// Lineman 订单转换
linemanOrder, _ := utility.FromLinemanPlaceOrder(linemanReq)
```

---

## 📋 关键字段映射速查

| 概念 | Lineman | 统一模型 | Grab | 映射说明 |
|------|---------|---------|------|---------|
| **订单 ID** | `orderId` | `OrderID` | `orderID` | 直接映射 |
| **短单号** | `orderShortCode` | `ShortOrderNumber` | `shortOrderNumber` | 直接映射 |
| **商户 ID** | `storeId` | `MerchantID` | `merchantID` | 直接映射 |
| **订单总额** | `restaurantRevenue` | `Price.Subtotal` | `price.eaterPayment` | 直接映射 |
| **会员 ID** | `memberId` | `MembershipID` ✅ | `membershipID` | **复用 Grab** |
| **商品备注** | `items[].memo` | `Items[].Specifications` ✅ | `items[].specifications` | **复用 Grab** |
| **订单类型** | `customerType` | `FeatureFlags.OrderType` ✅ | `featureFlags.orderType` | **转换映射** |
| **附加属性** | `additionalItems[]` | `AdditionalProperties[]` ✅ | - | **重命名** |

---

## 🔄 订单类型映射

### Lineman → 统一模型

```go
Lineman              统一模型
-----------------------------------
DELIVERY      →      FeatureFlags.OrderType = "Delivery"
PICKUP        →      FeatureFlags.OrderType = "Pickup"
```

### 统一模型 → Lineman

```go
统一模型                           Lineman
---------------------------------------------------
FeatureFlags.OrderType = "Delivery"   →   DELIVERY
FeatureFlags.OrderType = "Pickup"     →   PICKUP
FeatureFlags.OrderType = "DineIn"     →   PICKUP (堂食视为自取)
```

---

## 💰 价格单位

| 平台 | 单位 | 类型 | 示例 | 转换 |
|------|------|------|------|------|
| **Grab** | 分 (satang) | `int64` | `33389` = 333.89 THB | `/100` |
| **Lineman** | 泰铢 (THB) | `float64` | `333.89` | 无需转换 |
| **统一模型** | 泰铢 (THB) | `float64` | `333.89` | - |

---

## 🏷️ Properties vs Modifiers

### Lineman (嵌套)

```json
{
  "properties": [
    {
      "id": "FLAVOR",
      "values": [
        {"id": "MILD", "price": 0},
        {"id": "SPICY", "price": 5}
      ]
    }
  ]
}
```

### 统一模型 (扁平 + Values)

```json
{
  "modifiers": [
    {
      "id": "FLAVOR",
      "quantity": 1,
      "price": 5.0,
      "values": [
        {"id": "MILD", "price": 0},
        {"id": "SPICY", "price": 5}
      ]
    }
  ]
}
```

### Grab (纯扁平)

```json
{
  "modifiers": [
    {"id": "MOD-001", "quantity": 1, "price": 1000}
  ]
}
```

---

## ✅ 优化成果

### 删除的冗余字段

| 删除字段 | 替代方案 | 优势 |
|---------|---------|------|
| ❌ `CustomerType` | ✅ `FeatureFlags.OrderType` | 复用 Grab 结构 |
| ❌ `MemberID` | ✅ `MembershipID` | 复用 Grab 字段 |
| ❌ `Items[].Memo` | ✅ `Items[].Specifications` | 复用 Grab 字段 |

### 重命名的字段

| 原名称 | 新名称 | 原因 |
|--------|--------|------|
| `AdditionalItems` | `AdditionalProperties` | 对齐 Grab SDK 命名 |

---

## 📚 相关文档

- **详细映射**: [MAPPING.md](./MAPPING.md)
- **使用指南**: [README.md](./README.md)
- **转换工具**: [utility/README.md](../../../ttpos-bmp/app/ttpos-takeout/utility/README.md)
- **Google Sheets**: [Lineman API定义及TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)

---

**版本**: v1.0.0  
**最后更新**: 2026-01-12  
**维护者**: rikugun
