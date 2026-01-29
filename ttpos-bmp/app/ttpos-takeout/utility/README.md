# Takeout Converter 转换工具

> **包路径**: `ttpos-bmp/app/ttpos-takeout/utility`

## 📦 模块说明

本包提供外卖订单格式的**双向转换方法**,用于在不同平台的订单格式和统一数据模型之间进行转换。

**数据模型定义**: `ttpos-api/ttpos-takeout/message`

---

## 🔧 可用方法

### Grab 平台转换

#### FromGrabSDK

从 Grab SDK 订单请求转换为统一模型

```go
func FromGrabSDK(req *grabsdk.SubmitOrderRequest) (*message.TakeoutOrder, error)
```

**参数**:
- `req`: Grab SDK 订单请求 (`*grabsdk.SubmitOrderRequest`)

**返回**:
- `*message.TakeoutOrder`: 统一订单模型
- `error`: 错误信息

**转换内容**:
- ✅ 价格单位转换 (分 → 元, `/100`)
- ✅ 时间格式转换 (`time.Time` → RFC3339 字符串)
- ✅ 字段名映射 (`Phones` → `Phone`)
- ✅ 原始数据序列化到 `RawData` 字段
- ✅ 设置 `ProviderName` 为 `"grab"`

#### ToGrabSDK

转换为 Grab SDK 订单请求

```go
func ToGrabSDK(o *message.TakeoutOrder) *grabsdk.SubmitOrderRequest
```

**参数**:
- `o`: 统一订单模型 (`*message.TakeoutOrder`)

**返回**:
- `*grabsdk.SubmitOrderRequest`: Grab SDK 订单请求

**转换内容**:
- ✅ 价格单位转换 (元 → 分, `*100`)
- ✅ 时间格式转换 (RFC3339 字符串 → `time.Time`)
- ✅ 字段名映射 (`Phone` → `Phones`)

---

### Lineman 平台转换

#### FromLinemanPlaceOrder

从 Lineman PlaceOrder 请求转换为统一模型

```go
func FromLinemanPlaceOrder(req *linemanv1.PlaceOrderReq) (*message.TakeoutOrder, error)
```

**参数**:
- `req`: Lineman PlaceOrder 请求 (`*linemanv1.PlaceOrderReq`)

**返回**:
- `*message.TakeoutOrder`: 统一订单模型
- `error`: 错误信息

**转换内容**:
- ✅ 字段名映射:
  - `orderId` → `OrderID`
  - `orderShortCode` → `ShortOrderNumber`
  - `storeId` → `PartnerMerchantID` (TTPOS 侧店铺 ID)
  - `MerchantID` 设为空字符串
  - `restaurantRevenue` → `Price.Subtotal` + `Price.EaterPayment`
- ✅ Properties → Modifiers 结构转换 (嵌套 → 扁平)
- ✅ **复用 Grab 字段**: `memo` → `Specifications`, `memberId` → `MembershipID`
- ✅ **OrderType 映射**: `customerType` (DELIVERY/PICKUP) → `FeatureFlags.OrderType` (Delivery/Pickup)
- ✅ **促销信息提升**: `items[].promotionId` + `items[].discount` → `order.Promos[]` (商品级别 → 订单级别)
- ✅ 默认 `PaymentType` 为 `"CASH"`

**参考文档**:
- [Lineman API 映射表](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)

#### ToLinemanPlaceOrder

转换为 Lineman PlaceOrder 请求

```go
func ToLinemanPlaceOrder(o *message.TakeoutOrder) *linemanv1.PlaceOrderReq
```

**参数**:
- `o`: 统一订单模型 (`*message.TakeoutOrder`)

**返回**:
- `*linemanv1.PlaceOrderReq`: Lineman PlaceOrder 请求

**转换内容**:
- ✅ 字段名映射 (`OrderID` → `orderId`, `Price.Subtotal` → `restaurantRevenue`)
- ✅ Modifiers → Properties 结构转换 (扁平 → 嵌套)
- ✅ **字段反向映射**: `Specifications` → `memo`, `MembershipID` → `memberId`
- ✅ **OrderType 反向映射**: `FeatureFlags.OrderType` (Delivery/Pickup) → `customerType` (DELIVERY/PICKUP)
- ✅ **促销信息降级**: `order.Promos[]` → `items[0].promotionId` + `items[0].discount` (仅第一个商品)
- ✅ 默认 `customerType` 为 `"DELIVERY"` (如果未设置)

**限制说明**: 促销信息反向转换会丢失精确对应关系，仅将第一个 Promo 设置到第一个商品上

---

## 📖 使用示例

### Grab 订单转换

```go
package main

import (
    "fmt"
    grabsdk "github.com/grab/grabfood-api-sdk-go"
    "ttpos-api/ttpos-takeout/message"
    "ttpos-bmp/app/ttpos-takeout/utility"
)

func processGrabOrder(req *grabsdk.SubmitOrderRequest) error {
    // 1. 转换为统一模型
    order, err := utility.FromGrabSDK(req)
    if err != nil {
        return fmt.Errorf("转换失败: %v", err)
    }
    
    // 2. 使用统一模型处理订单
    fmt.Printf("订单ID: %s\n", order.OrderID)
    fmt.Printf("商户ID: %s\n", order.MerchantID)
    fmt.Printf("订单总额: %.2f THB\n", order.Price.Subtotal)
    fmt.Printf("渠道: %s\n", order.ProviderName) // "grab"
    
    // 3. 访问平台特有字段
    if order.Cutlery != nil && *order.Cutlery {
        fmt.Println("顾客需要餐具")
    }
    
    if order.Currency != nil {
        fmt.Printf("货币: %s\n", order.Currency.Code)
    }
    
    // 4. 保存到数据库
    return saveToDatabase(order)
}

// 如需回传给 Grab
func convertBackToGrab(order *message.TakeoutOrder) {
    grabReq := utility.ToGrabSDK(order)
    // 使用 grabReq 进行后续操作
}
```

### Lineman 订单转换

```go
package main

import (
    "fmt"
    linemanv1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
    "ttpos-api/ttpos-takeout/message"
    "ttpos-bmp/app/ttpos-takeout/utility"
)

func processLinemanOrder(req *linemanv1.PlaceOrderReq) error {
    // 1. 转换为统一模型
    order, err := utility.FromLinemanPlaceOrder(req)
    if err != nil {
        return fmt.Errorf("转换失败: %v", err)
    }
    
    // 2. 使用统一模型处理订单
    fmt.Printf("订单ID: %s\n", order.OrderID)
    fmt.Printf("商户ID: %s\n", order.MerchantID)
    fmt.Printf("订单总额: %.2f THB\n", order.Price.Subtotal)
    fmt.Printf("渠道: %s\n", order.ProviderName) // "lineman"
    
    // 3. 访问平台特有字段
    if order.CustomerType != nil {
        orderType := "外送"
        if *order.CustomerType == "PICKUP" {
            orderType = "自取"
        }
        fmt.Printf("订单类型: %s\n", orderType)
    }
    
    if order.MemberID != nil {
        fmt.Printf("会员ID: %s\n", *order.MemberID)
    }
    
    // 4. 访问附加项
    for _, item := range order.AdditionalItems {
        fmt.Printf("附加信息: %s\n", item.Name)
    }
    
    // 5. 保存到数据库
    return saveToDatabase(order)
}

// 如需回传给 Lineman
func convertBackToLineman(order *message.TakeoutOrder) {
    linemanReq := utility.ToLinemanPlaceOrder(order)
    // 使用 linemanReq 进行后续操作
}
```

### 平台无关的订单处理

```go
// ProcessOrder 统一处理订单(无需关心来源平台)
func ProcessOrder(order *message.TakeoutOrder) error {
    // 1. 验证订单
    if order.OrderID == "" {
        return fmt.Errorf("订单ID不能为空")
    }
    
    if len(order.Items) == 0 {
        return fmt.Errorf("订单商品列表不能为空")
    }
    
    // 2. 计算订单总额
    var totalAmount float64
    for _, item := range order.Items {
        totalAmount += item.Price * float64(item.Quantity)
    }
    
    fmt.Printf("订单总额: %.2f\n", totalAmount)
    
    // 3. 根据渠道执行特定逻辑
    switch order.ProviderName {
    case "grab":
        // Grab 特有逻辑
        if order.Cutlery != nil && *order.Cutlery {
            fmt.Println("需要准备餐具")
        }
        
    case "lineman":
        // Lineman 特有逻辑
        if order.CustomerType != nil && *order.CustomerType == "PICKUP" {
            fmt.Println("自取订单,无需配送")
        }
    }
    
    // 4. 保存到数据库
    return db.SaveOrder(order)
}
```

---

## 🔄 转换规则

### Properties ↔ Modifiers 转换

**Lineman Properties (嵌套结构)**:
```json
{
  "properties": [
    {
      "id": "PROP-FLAVOR",
      "values": [
        {"id": "MILD", "price": 0},
        {"id": "SPICY", "price": 5}
      ]
    }
  ]
}
```

**转换为 Modifiers (扁平结构)**:
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
    }
  ]
}
```

**转换逻辑**:
1. 每个 `property` 映射为一个 `modifier`
2. `property.id` → `modifier.id`
3. `property.values` 保留在 `modifier.values` 中
4. `modifier.price` 为所有 `values` 价格之和
5. `modifier.quantity` 默认为 `1`

---

## 🧪 测试

### 运行测试

```bash
cd /home/coder/workspaces/ttpos-server-go/ttpos-bmp/app/ttpos-takeout
go test -v ./utility/takeout_converter_test.go ./utility/takeout_converter.go
```

### 测试覆盖

- ✅ **TestFromGrabSDK**: Grab SDK → 统一模型转换
- ✅ **TestFromLinemanPlaceOrder**: Lineman API → 统一模型转换
- ✅ **TestRoundTripGrabSDK**: Grab 双向转换 (往返测试)
- ✅ **TestRoundTripLineman**: Lineman 双向转换 (往返测试)
- ✅ **TestJSONSerialization**: JSON 序列化/反序列化

### 测试结果

```
=== RUN   TestFromGrabSDK
--- PASS: TestFromGrabSDK (0.00s)
=== RUN   TestFromLinemanPlaceOrder
--- PASS: TestFromLinemanPlaceOrder (0.00s)
=== RUN   TestRoundTripGrabSDK
--- PASS: TestRoundTripGrabSDK (0.00s)
=== RUN   TestRoundTripLineman
--- PASS: TestRoundTripLineman (0.00s)
=== RUN   TestJSONSerialization
--- PASS: TestJSONSerialization (0.00s)
PASS
```

---

## ⚠️ 注意事项

### 1. 价格单位转换

**Grab**:
- 输入: `33389` (分, int64)
- 转换: `/ 100`
- 输出: `333.89` (元, float64)

**Lineman**:
- 输入: `333.89` (元, float64)
- 无需转换
- 输出: `333.89` (元, float64)

### 2. 时间格式处理

**Grab**:
- `OrderTime`: string (RFC3339) → 直接映射
- `SubmitTime`, `CompleteTime`: `time.Time` → 格式化为 RFC3339
- `OrderReadyEstimation`: `time.Time` → 格式化为 RFC3339

**Lineman**:
- `orderAcceptedTime`: string (ISO 8601) → 直接映射

### 3. 字段名差异与映射

| 概念 | Grab SDK | Lineman API | 统一模型 | 映射说明 |
|------|----------|-------------|---------|---------|
| 电话 | `Phones` | - | `Phone` | 字段名简化 |
| 订单ID | `OrderID` | `orderId` | `OrderID` | 统一驼峰命名 |
| 短单号 | `ShortOrderNumber` | `orderShortCode` | `ShortOrderNumber` | 统一命名 |
| 用户实付 | `EaterPayment` | `restaurantRevenue` | `Subtotal` & `EaterPayment` | 双字段兼容 |
| 商品备注 | `Specifications` | `memo` | `Specifications` | **复用 Grab 字段** |
| 会员ID | `membershipID` | `memberId` | `MembershipID` | **复用 Grab 字段** |
| 订单类型 | `featureFlags.orderType` | `customerType` | `FeatureFlags.OrderType` | **映射转换** |
| 附加属性 | - | `additionalItems` | `AdditionalProperties` | **重命名** |

**关键设计决策**:
- ✅ **最小化冗余**: 删除 Lineman 特有的 `MemberID`、`CustomerType` 字段
- ✅ **复用 Grab 字段**: 使用 `MembershipID`、`Specifications`、`FeatureFlags.OrderType`
- ✅ **统一命名**: `AdditionalItems` → `AdditionalProperties` (对齐 Grab SDK 概念)

### 4. 默认值

- **Lineman**: `PaymentType` 默认为 `"CASH"`
- **Lineman**: `FeatureFlags.OrderType` 默认为 `"Delivery"` (回传时转换为 `customerType: "DELIVERY"`)
- **Lineman**: `FeatureFlags.OrderAcceptedType` 默认为 `"AUTO"` (自动接单)
- **Grab**: `Modifiers.Quantity` 默认为 `1` (如果未提供)

### 5. 原始数据保留

- 所有转换方法都会将原始请求序列化到 `RawData` 字段
- `RawData` 不参与 JSON 序列化 (`json:"-"`)
- 用于问题排查和数据追溯

---

## 📁 文件结构

```
ttpos-bmp/app/ttpos-takeout/utility/
├── takeout_converter.go       # 转换方法实现
├── takeout_converter_test.go  # 单元测试
└── README.md                   # 本文档
```

---

## 🚀 扩展新平台

如需支持新平台 (如 FoodPanda):

### 1. 在 `ttpos-api/ttpos-takeout/message` 中添加平台特有字段

参见: `ttpos-api/ttpos-takeout/message/README.md`

### 2. 在本包中实现转换方法

```go
// FromFoodPanda 从 FoodPanda 订单请求转换为统一模型
func FromFoodPanda(req *foodpanda.OrderRequest) (*message.TakeoutOrder, error) {
    if req == nil {
        return nil, nil
    }
    
    // 序列化原始数据
    rawData, _ := json.Marshal(req)
    
    order := &message.TakeoutOrder{
        OrderID:      req.ID,
        MerchantID:   req.VendorID,
        PaymentType:  req.PaymentMethod,
        ProviderName: "foodpanda",
        RawData:      rawData,
    }
    
    // ... 实现字段映射 ...
    
    return order, nil
}

// ToFoodPanda 转换为 FoodPanda 订单请求
func ToFoodPanda(o *message.TakeoutOrder) *foodpanda.OrderRequest {
    if o == nil {
        return nil
    }
    
    req := &foodpanda.OrderRequest{
        ID:            o.OrderID,
        VendorID:      o.MerchantID,
        PaymentMethod: o.PaymentType,
    }
    
    // ... 实现字段映射 ...
    
    return req
}
```

### 3. 添加单元测试

```go
func TestFromFoodPanda(t *testing.T) {
    // 构造测试数据
    req := &foodpanda.OrderRequest{
        ID:       "fp-001",
        VendorID: "vendor-123",
        // ...
    }
    
    // 转换
    order, err := FromFoodPanda(req)
    if err != nil {
        t.Fatalf("FromFoodPanda failed: %v", err)
    }
    
    // 验证
    if order.OrderID != "fp-001" {
        t.Errorf("Expected OrderID 'fp-001', got '%s'", order.OrderID)
    }
    
    if order.ProviderName != "foodpanda" {
        t.Errorf("Expected ProviderName 'foodpanda', got '%s'", order.ProviderName)
    }
}
```

---

## 📚 相关文档

- **数据模型**: [ttpos-api/ttpos-takeout/message/README.md](../../../../ttpos-api/ttpos-takeout/message/README.md)
- **Lineman 映射表**: [Google Sheets](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)
- **Grab SDK**: [GitHub](https://github.com/grab/grabfood-api-sdk-go)
- **GoFrame 规范**: [ttpos-bmp/.cursor/rules/go-rules.mdc](../../../.cursor/rules/go-rules.mdc)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**维护者**: TTPOS Team
