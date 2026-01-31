# Takeout Order 统一数据模型

> **包路径**: `ttpos-api/ttpos-takeout/message`

## 📦 模块说明

本包提供**统一的外卖订单数据模型** (`TakeoutOrder`),用于抽象不同外卖平台(Grab, Lineman 等)的订单格式。

### 设计原则

- ✅ **平台无关**: 以 Grab 订单格式为基准,扩展支持 Lineman 等其他平台的特有字段
- ✅ **公共复用**: 定义在 `ttpos-api` 中,可被多个模块引用
- ✅ **完整性**: 保留所有平台的字段信息,不丢失数据
- ✅ **可扩展**: 易于添加新平台支持

---

## 🏗️ 核心结构

### TakeoutOrder (统一订单模型)

```go
type TakeoutOrder struct {
    // ========== 基础字段 (Grab 和 Lineman 通用) ==========
    OrderID           string            // 订单 ID
    ShortOrderNumber  string            // 短订单号
    MerchantID        string            // 商户 ID (平台侧)
    PartnerMerchantID string            // 合作商户 ID (TTPOS 侧)
    PaymentType       string            // 支付方式
    OrderTime         string            // 下单时间 (RFC3339)
    Items             []TakeoutOrderItem // 订单商品列表
    Price             TakeoutOrderPrice  // 价格信息
    
    // ========== Grab 特有字段 ==========
    Cutlery              *bool                        // 是否需要餐具
    Currency             *TakeoutCurrency             // 货币信息
    FeatureFlags         *TakeoutFeatureFlags         // 功能标识
    DineIn               *TakeoutDineIn               // 堂食信息
    Receiver             *TakeoutReceiver             // 收货人信息
    OrderReadyEstimation *TakeoutOrderReadyEstimation // 准备时间估算
    MembershipID         *string                      // 会员 ID
    
    // ========== Lineman 特有字段 ==========
    CustomerType    *string                   // 订单类型 (DELIVERY/PICKUP)
    AdditionalItems []TakeoutAdditionalItem  // 订单附加项
    MemberID        *string                   // 会员 ID
    
    // ========== 内部字段 ==========
    ProviderName string          // 渠道名称 (grab/lineman)
    RawData      json.RawMessage // 原始数据 (完整保存)
}
```

---

## 🔧 转换工具

**转换方法位于**: `ttpos-bmp/app/ttpos-takeout/utility/takeout_converter.go`

### 可用方法

| 方法 | 说明 | 包路径 |
|------|------|--------|
| `utility.FromGrabSDK()` | Grab SDK → 统一模型 | `ttpos-bmp/app/ttpos-takeout/utility` |
| `utility.ToGrabSDK()` | 统一模型 → Grab SDK | `ttpos-bmp/app/ttpos-takeout/utility` |
| `utility.FromLinemanPlaceOrder()` | Lineman API → 统一模型 | `ttpos-bmp/app/ttpos-takeout/utility` |
| `utility.ToLinemanPlaceOrder()` | 统一模型 → Lineman API | `ttpos-bmp/app/ttpos-takeout/utility` |

---

## 📖 使用示例

### 引入包

```go
import (
    "ttpos-api/ttpos-takeout/message"
    "ttpos-bmp/app/ttpos-takeout/utility"
)
```

### Grab 订单处理

```go
import (
    grabsdk "github.com/grab/grabfood-api-sdk-go"
    "ttpos-api/ttpos-takeout/message"
    "ttpos-bmp/app/ttpos-takeout/utility"
)

func handleGrabOrder(req *grabsdk.SubmitOrderRequest) error {
    // 转换为统一模型
    order, err := utility.FromGrabSDK(req)
    if err != nil {
        return err
    }
    
    // 使用统一字段处理
    fmt.Printf("订单ID: %s\n", order.OrderID)
    fmt.Printf("商户ID: %s\n", order.MerchantID)
    fmt.Printf("订单总额: %.2f\n", order.Price.Subtotal)
    fmt.Printf("渠道: %s\n", order.ProviderName) // "grab"
    
    // 访问 Grab 特有字段
    if order.Cutlery != nil && *order.Cutlery {
        fmt.Println("需要准备餐具")
    }
    
    // 保存到数据库或其他处理
    return saveOrder(order)
}
```

### Lineman 订单处理

```go
import (
    linemanv1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
    "ttpos-api/ttpos-takeout/message"
    "ttpos-bmp/app/ttpos-takeout/utility"
)

func handleLinemanOrder(req *linemanv1.PlaceOrderReq) error {
    // 转换为统一模型
    order, err := utility.FromLinemanPlaceOrder(req)
    if err != nil {
        return err
    }
    
    // 使用统一字段处理
    fmt.Printf("订单ID: %s\n", order.OrderID)
    fmt.Printf("商户ID: %s\n", order.MerchantID)
    fmt.Printf("订单总额: %.2f\n", order.Price.Subtotal)
    fmt.Printf("渠道: %s\n", order.ProviderName) // "lineman"
    
    // 访问 Lineman 特有字段
    if order.CustomerType != nil && *order.CustomerType == "PICKUP" {
        fmt.Println("自取订单,无需配送")
    }
    
    // 保存到数据库或其他处理
    return saveOrder(order)
}
```

### 平台无关的订单处理

```go
// saveOrder 统一保存订单(无需关心来源平台)
func saveOrder(order *message.TakeoutOrder) error {
    // 1. 验证订单
    if order.OrderID == "" {
        return fmt.Errorf("订单ID不能为空")
    }
    
    // 2. 计算订单总额
    var totalAmount float64
    for _, item := range order.Items {
        totalAmount += item.Price * float64(item.Quantity)
    }
    
    // 3. 根据渠道执行特定逻辑
    switch order.ProviderName {
    case "grab":
        // Grab 特有逻辑
    case "lineman":
        // Lineman 特有逻辑
    }
    
    // 4. 保存到数据库
    return db.SaveOrder(order)
}
```

---

## 🔄 字段映射

### Grab SDK → TakeoutOrder

| Grab SDK 字段 | TakeoutOrder 字段 | 转换说明 |
|--------------|------------------|---------|
| `OrderID` | `OrderID` | 直接映射 |
| `Price.Subtotal` (int64, 分) | `Price.Subtotal` (float64, 元) | `/100` 转换 |
| `Items[].Price` (int64, 分) | `Items[].Price` (float64, 元) | `/100` 转换 |
| `Items[].Modifiers` | `Items[].Modifiers` | 结构体转换 |
| `Receiver.Phones` | `Receiver.Phone` | 字段名映射 |
| `OrderReadyEstimation.*` (time.Time) | `OrderReadyEstimation.*` (string) | RFC3339 格式化 |

### Lineman API → TakeoutOrder

| Lineman API 字段 | TakeoutOrder 字段 | 对应 Grab 字段 | 转换说明 |
|------------------|------------------|---------------|---------|
| `orderId` | `OrderID` | `orderID` | 直接映射 |
| `orderShortCode` | `ShortOrderNumber` | `shortOrderNumber` | 直接映射 |
| `storeId` | `PartnerMerchantID` | - | **Lineman storeId → PartnerMerchantID** |
| `partnerId` | - | - | **不使用（Lineman partnerId 已废弃）** |
| - | `MerchantID` | `merchantID` | **Lineman 时为空** |
| `restaurantRevenue` | `Price.Subtotal` / `Price.EaterPayment` | `price.eaterPayment` | 直接映射(已是元) |
| `orderAcceptedTime` | `OrderTime` | `orderTime` | 直接映射 |
| `items[].id` | `Items[].ID` | `items[].id` | 直接映射 |
| `items[].quantity` | `Items[].Quantity` | `items[].quantity` | 直接映射 |
| `items[].unitPrice` | `Items[].Price` | `items[].price` | 直接映射 |
| `items[].memo` | `Items[].Specifications` | `items[].specifications` | **字段复用** |
| `items[].properties` | `Items[].Modifiers` | `items[].modifiers` | 属性组展开 |
| `items[].promotionId` + `items[].discount` | `Promos[]` | `promos[]` | **提升到订单级别** (详见 MAPPING.md) |
| `customerType` (DELIVERY/PICKUP) | `FeatureFlags.OrderType` | `featureFlags.orderType` | **转换映射** |
| `memberId` | `MembershipID` | `membershipID` | **字段复用** |
| `additionalItems` | `AdditionalProperties` | - | **重命名** (扩展字段) |

---

## 📁 文件结构

```
ttpos-api/
└── ttpos-takeout/
    └── message/
        ├── takeout_order.go  # 统一数据模型定义
        └── README.md         # 本文档

ttpos-bmp/app/ttpos-takeout/
└── utility/
    ├── takeout_converter.go      # 转换方法实现
    └── takeout_converter_test.go # 单元测试
```

---

## 🧪 测试

```bash
# 运行转换器测试
cd /home/coder/workspaces/ttpos-server-go/ttpos-bmp/app/ttpos-takeout
go test -v ./utility/takeout_converter_test.go ./utility/takeout_converter.go
```

**测试覆盖**:
- ✅ Grab SDK → TakeoutOrder 转换
- ✅ Lineman API → TakeoutOrder 转换
- ✅ TakeoutOrder → Grab SDK 双向转换
- ✅ TakeoutOrder → Lineman API 双向转换
- ✅ JSON 序列化/反序列化

---

## 🚀 扩展新平台

如需支持新平台(如 FoodPanda, Shopee Food):

### 1. 在 `message/takeout_order.go` 中添加平台特有字段

```go
type TakeoutOrder struct {
    // ... 现有字段 ...
    
    // ========== 新平台特有字段 ==========
    NewPlatformField *string `json:"newPlatformField,omitempty"`
}
```

### 2. 在 `utility/takeout_converter.go` 中实现转换方法

```go
// FromNewPlatform 从新平台订单请求转换为统一模型
func FromNewPlatform(req *NewPlatformOrderRequest) (*message.TakeoutOrder, error) {
    // 实现转换逻辑
}

// ToNewPlatform 转换为新平台订单请求
func ToNewPlatform(o *message.TakeoutOrder) *NewPlatformOrderRequest {
    // 实现转换逻辑
}
```

### 3. 添加单元测试

```go
func TestFromNewPlatform(t *testing.T) {
    // 测试转换逻辑
}
```

---

## ⚠️ 注意事项

### 1. 价格单位

- **Grab**: 使用最小单位(分), `int64` 类型 → 转换为元(`/100`)
- **Lineman**: 使用标准单位(元), `float64` 类型 → 无需转换
- **TakeoutOrder**: 统一使用元, `float64` 类型

### 2. Properties vs Modifiers

- **Lineman**: 属性组 → 属性值列表(嵌套结构)
- **Grab**: 扁平的修改项列表
- **转换**: Lineman properties 展开为扁平结构, `Values` 字段保留原始嵌套

### 3. 时间格式

- **Grab**: 部分字段为 `time.Time` 类型
- **Lineman**: 字符串 (ISO 8601)
- **TakeoutOrder**: 统一使用字符串 (RFC3339)

### 4. 原始数据保留

- `RawData` 字段保存完整的原始请求 JSON
- 用于问题排查和数据追溯
- 不参与 JSON 序列化 (`json:"-"`)

---

## 📚 相关文档

- [Lineman API 映射表](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)
- [Grab SDK 文档](https://github.com/grab/grabfood-api-sdk-go)
- [GoFrame 开发规范](../../../ttpos-bmp/.cursor/rules/go-rules.mdc)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**维护者**: TTPOS Team
