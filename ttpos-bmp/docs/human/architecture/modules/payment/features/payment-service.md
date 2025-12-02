# 支付服务 (Payment Service)

## 📋 服务概述

支付服务提供统一的支付接口，支持多种支付渠道（微信支付、支付宝、PromptPay 等），并提供订单查询功能。

## 🔧 核心功能

### 1. 微信支付 (WeChatPay)

**接口**: `WeChatPay(ctx, platform, req)`

**功能**: 创建微信扫码支付订单

**支付流程**:
```
1. 验证商户平台配置
2. 生成微信支付订单
3. 调用微信 API 创建二维码
4. 保存支付记录到数据库
5. 返回支付二维码和链接
```

**请求参数**:
```go
type WeChatPayReq struct {
    ShopSupplierID  string  // 商户门店ID（必填）
    MerchantOrderNo string  // 商户订单号（必填，唯一）
    OrderAmount     float64 // 订单金额（必填，THB）
    OrderCurrency   string  // 货币类型（必填：THB/USD/CNY）
    OrderDesc       string  // 订单描述（必填）
    FullName        string  // 订单人名称（必填）
    MerchantUserID  string  // 商户用户ID（必填）
    CallbackURL     string  // 回调地址（必填）
    PaymentMethod   string  // 支付方式（可选）
    RedirectURL     string  // 跳转地址（可选）
}
```

**响应数据**:
```go
type PaymentResp struct {
    PaymentID       string // 支付服务唯一ID
    MerchantOrderNo string // 商户订单号
    OrderID         string // 支付平台订单ID
    OrderStatus     string // 订单状态
    LinkURL         string // 支付链接
    QRCode          string // 二维码 base64
    QRCodeExpireSec int    // 二维码过期时间（秒）
}
```

**使用示例**:
```go
// 创建支付请求
req := &dto.WeChatPayReq{
    ShopSupplierID:  "SHOP_001",
    MerchantOrderNo: "ORDER_20251127_001",
    OrderAmount:     100.00,
    OrderCurrency:   "THB",
    OrderDesc:       "测试订单",
    FullName:        "测试用户",
    MerchantUserID:  "USER_001",
    CallbackURL:     "https://merchant.com/callback",
    PaymentMethod:   "WECHAT_PAY",
    RedirectURL:     "https://merchant.com/success",
}

// 获取平台配置
platform, err := service.Platform().GetPlatformByShopSupplierID(ctx, req.ShopSupplierID)

// 调用支付
resp, err := service.Payment().WeChatPay(ctx, platform, req)
```

### 2. 支付宝离线支付 (AliOfflinePay)

**接口**: `AliOfflinePay(ctx, platform, req)`

**功能**: 创建支付宝离线支付订单（扫码/条码）

**支付流程**:
```
1. 验证商户平台配置
2. 生成支付宝订单
3. 调用支付宝 API
4. 保存支付记录
5. 返回支付二维码
```

**请求参数**:
```go
type AliOfflinePayReq struct {
    ShopSupplierID  string  // 商户门店ID
    MerchantOrderNo string  // 商户订单号
    OrderAmount     float64 // 订单金额
    OrderCurrency   string  // 货币类型
    OrderDesc       string  // 订单描述
    FullName        string  // 订单人名称
    MerchantUserID  string  // 商户用户ID
    CallbackURL     string  // 回调地址
    RedirectURL     string  // 跳转地址
}
```

**特点**:
- 支持扫码支付
- 支持条码支付（用户出示付款码）
- 离线场景优化
- 支付结果实时返回

### 3. QR PromptPay 支付

**接口**: `QrPromptPay(ctx, platform, req)`

**功能**: 创建 PromptPay 二维码支付订单（泰国本地支付）

**支付流程**:
```
1. 验证商户平台配置
2. 生成 PromptPay 订单
3. 创建泰国标准二维码
4. 保存支付记录
5. 返回二维码（Thai QR Payment 格式）
```

**请求参数**:
```go
type QrPromptPayReq struct {
    ShopSupplierID  string  // 商户门店ID
    MerchantOrderNo string  // 商户订单号
    OrderAmount     float64 // 订单金额（THB）
    OrderCurrency   string  // 货币类型（必须为 THB）
    OrderDesc       string  // 订单描述
    FullName        string  // 订单人名称
    MerchantUserID  string  // 商户用户ID
    CallbackURL     string  // 回调地址
    RedirectURL     string  // 跳转地址
}
```

**特点**:
- 泰国本地支付标准
- 支持所有泰国银行 APP 扫码
- 实时到账
- 无手续费或低手续费

### 4. 支付订单查询 (PaymentQuery)

**接口**: `PaymentQuery(ctx, platform, req)`

**功能**: 查询支付订单状态和详情

**查询方式**:
- 通过商户订单号查询
- 通过支付平台订单ID查询

**请求参数**:
```go
type PaymentQueryReq struct {
    ShopSupplierID  string // 商户门店ID（必填）
    MerchantOrderNo string // 商户订单号（可选）
    OrderID         string // 支付平台订单ID（可选）
}
```

**响应数据**:
```go
type PaymentQueryResp struct {
    MerchantOrderNo string  // 商户订单号
    OrderID         string  // 支付平台订单ID
    OrderStatus     string  // 订单状态
    OrderAmount     float64 // 订单金额
    OrderCurrency   string  // 订单货币
    PayTime         int     // 支付时间（时间戳）
}
```

**订单状态**:
- `PENDING` - 待支付
- `PROCESSING` - 支付处理中
- `SUCCESS` - 支付成功
- `FAILED` - 支付失败
- `CLOSED` - 订单关闭
- `REFUNDING` - 退款中
- `REFUNDED` - 已退款

## 🔄 支付状态机

```
        创建订单
           ↓
       PENDING (待支付)
           ↓
      用户扫码支付
           ↓
    PROCESSING (处理中)
           ↓
      ┌─────┴─────┐
      ↓           ↓
  SUCCESS      FAILED
  (成功)       (失败)
      ↓           
   可申请退款      
      ↓           
  REFUNDING      
  (退款中)       
      ↓           
  REFUNDED       
  (已退款)       
```

## 💾 数据存储

### payment_receipts 表

支付记录表，存储所有支付订单信息：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| uuid | BIGINT | 唯一标识 |
| shop_supplier_id | VARCHAR(255) | 商户门店ID |
| merchant_order_no | VARCHAR(500) | 商户订单号 |
| order_id | VARCHAR(255) | 支付平台订单ID |
| order_amount | DECIMAL(14,2) | 订单金额 |
| order_currency | VARCHAR(10) | 货币类型 |
| order_status | INT | 订单状态 |
| payment_method | VARCHAR(50) | 支付方式 |
| pay_time | INT | 支付时间 |
| callback_url | VARCHAR(500) | 回调地址 |
| create_time | INT | 创建时间 |
| update_time | INT | 更新时间 |

**索引设计**:
- PRIMARY KEY (id)
- UNIQUE INDEX (merchant_order_no)
- INDEX (shop_supplier_id)
- INDEX (order_id)
- INDEX (order_status)

## 🔐 安全机制

### 1. 签名验证

所有支付请求必须携带签名，服务端验证签名合法性：

```go
// 签名生成
sign := SHA256(sorted_params + sign_salt)

// 签名验证
platform, err := utility.VerifyRequestSign(ctx, requestParams, sign)
```

### 2. 幂等性保证

通过 `merchant_order_no` 唯一约束保证支付幂等性：

```go
// 查询是否已存在
existing, err := dao.PaymentReceipts.Ctx(ctx).
    Where("merchant_order_no", req.MerchantOrderNo).
    One()

if existing != nil {
    // 返回已有订单
    return buildResponse(existing), nil
}
```

### 3. 金额校验

支付金额范围限制：

```go
// 最小金额：0.01
// 最大金额：200,000.00

if req.OrderAmount < 0.01 || req.OrderAmount > 200000.00 {
    return nil, gerror.New("订单金额超出范围")
}
```

## 📊 性能优化

### 1. 平台配置缓存

平台信息通过 Redis 缓存，减少数据库查询：

```go
// 缓存 Key: platform:shop_supplier_id:{shop_supplier_id}
// TTL: 1 小时

platform, err := service.Platform().GetPlatformByShopSupplierID(ctx, shopSupplierID)
```

### 2. 支付 SDK 连接池

HTTP 客户端使用连接池，提升并发性能：

```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 3. 异步日志记录

支付日志异步记录，不阻塞主流程：

```go
go func() {
    _ = dao.PaymentCallbackLogs.Ctx(ctx).Data(logData).Insert()
}()
```

## 🧪 测试用例

### 单元测试

```go
func Test_sPayment_WeChatPay(t *testing.T) {
    gtest.C(t, func(t *gtest.T) {
        ctx := context.TODO()
        platform := &entity.PaymentPlatforms{
            ShopSupplierId: "test_supplier",
            SignSalt:       "test_salt",
            ApiUrl:         "https://api.test.com",
        }
        req := &dto.WeChatPayReq{
            ShopSupplierID:  "test_supplier",
            MerchantOrderNo: "test_order_wechat",
            OrderAmount:     100.00,
            OrderCurrency:   "THB",
            OrderDesc:       "测试订单",
            FullName:        "测试用户",
            MerchantUserID:  "test_user",
            CallbackURL:     "https://callback.test.com",
        }
        
        res, err := Payment.WeChatPay(ctx, platform, req)
        t.AssertNil(err)
        t.AssertNE(res, nil)
        t.Assert(res.MerchantOrderNo, req.MerchantOrderNo)
    })
}
```

### API 测试

```bash
# 微信支付
curl -X POST http://localhost:14061/api/v1/payment/wechat_pay \
  -H "Content-Type: application/json" \
  -H "X-Sign: {signature}" \
  -d '{
    "shop_supplier_id": "SHOP_001",
    "merchant_order_no": "ORDER_001",
    "order_amount": 100.00,
    "order_currency": "THB",
    "order_desc": "测试订单",
    "full_name": "测试用户",
    "merchant_user_id": "USER_001",
    "callback_url": "https://merchant.com/callback"
  }'
```

## 🚀 使用建议

### 1. 选择支付方式

| 场景 | 推荐方式 | 原因 |
|------|----------|------|
| 中国用户 | 微信支付 | 用户习惯，普及率高 |
| 泰国用户 | PromptPay | 本地支付，无手续费 |
| 国际用户 | 支付宝 | 支持多币种 |

### 2. 异常处理

```go
resp, err := service.Payment().WeChatPay(ctx, platform, req)
if err != nil {
    // 记录错误日志
    g.Log().Error(ctx, "支付失败", err)
    
    // 返回友好错误
    return nil, gerror.Wrap(err, "支付失败，请稍后重试")
}
```

### 3. 订单号生成

建议使用带时间戳和随机数的订单号：

```go
orderNo := fmt.Sprintf("ORDER_%s_%d", 
    time.Now().Format("20060102150405"),
    rand.Intn(1000000),
)
```

## 📝 注意事项

1. ⚠️ **订单号唯一性**：确保 merchant_order_no 全局唯一
2. ⚠️ **金额精度**：使用 DECIMAL(14,2) 避免精度丢失
3. ⚠️ **超时设置**：支付 API 调用设置合理超时（建议 30秒）
4. ⚠️ **错误重试**：网络错误时可重试，业务错误不应重试
5. ⚠️ **日志记录**：完整记录支付请求和响应，便于排查问题

---

**相关文档**:
- [回调服务](./callback-service.md)
- [退款服务](./refund-service.md)
- [通知服务](./notify-service.md)

