# Payment Service 支付服务说明文档

## 📋 概述

支付服务是 TTPOS 系统的核心功能之一，负责处理订单支付、支付回调、退款等支付相关功能。该服务支持连连支付（微信、支付宝、PromptPay）、余额支付、现金支付等多种支付方式，并提供完整的支付订单管理和退款功能。

**主要服务文件**:
- `/main/app/service/payment.go` - 支付核心服务（815行）
- `/main/app/service/payment_method.go` - 支付方式服务（133行）
- `/main/app/repository/payment_order.go` - 支付订单仓库（239行）
- `/main/app/repository/ll_payment_order.go` - 连连支付订单仓库（119行）

**接口定义**: 
- `PaymentRepo` - 支付仓库（连连支付）
- `IPaymentMethodSrv` - 支付方式服务接口
- `IPaymentOrderRepo` - 支付订单仓库接口
- `ILlPaymentOrderRepo` - 连连支付订单仓库接口

---

## 🏗️ 架构设计

### 支付仓库 (PaymentRepo)

```go
type PaymentRepo struct {
    dbm               *database.DBManager  // 数据库管理器
    ctx               contexts.Context      // 上下文
    payServiceUrl     string                // 支付服务URL
    payCallbackUrl    string                // 支付回调地址
    refundCallbackUrl string                // 退款回调地址
    orderCurrency     string                // 订单币种（THB/USD/JPY/CNY）
}
```

### 支付方式服务 (IPaymentMethodSrv)

```go
type IPaymentMethodSrv interface {
    IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool // 支付方式是否已启用
    GetList(ctx context.Context, typ string) resp.PaymentMethodList                                             // 获取支付方式列表
}
```

### 支付订单仓库 (IPaymentOrderRepo)

```go
type IPaymentOrderRepo interface {
    // 查询接口
    GetPaymentOrder(opts ...DBOption) (model.PaymentOrder, error)
    GetPaymentOrderList(opts ...DBOption) ([]model.PaymentOrder, error)
    GetPaymentOrderListBySaleOrderUuid(saleOrderUuid uint64) ([]*model.PaymentOrder, error)
    
    // 更新接口
    Create(order model.PaymentOrder) (model.PaymentOrder, error)
    UpdateOrCreatePaymentOrderRecord(obj model.PaymentOrder) error
    Update(uuid uint64, vars map[string]any) error
    DeletePaymentOrderRecord(uuid uint64) error
}
```

---

## 🎯 核心功能

### 1. 支付方式类型

系统支持多种支付方式：

| 支付方式代码 | 说明 | 类型 |
|------------|------|-----|
| `PaymentMethodCodeFreePay = -1` | 免单 | 系统默认 |
| `PaymentMethodCodeBalance = 10` | 余额支付 | 系统默认 |
| `PaymentMethodCodeCash = 40` | 现金支付 | 系统默认 |
| `PaymentMethodCodeLianLianWechatPay = 90111` | 连连微信支付 | 连连支付 |
| `PaymentMethodCodeLianLianAliPay = 90222` | 连连支付宝 | 连连支付 |
| `PaymentMethodCodeLianLianQRPromptPay = 90333` | 连连PromptPay | 连连支付 |
| `PaymentMethodCodeWechat = 20` | 微信支付 | 自有支付 |
| `PaymentMethodCodeAliPay = 30` | 支付宝支付 | 自有支付 |
| `PaymentMethodCodeQRPromptPay = 80` | QR PromptPay | 自有支付 |

### 2. 支付方式来源

| 来源常量 | 说明 |
|---------|------|
| `PaymentMethodSourceSystem = 0` | 系统默认 |
| `PaymentMethodSourceDefault = 1` | 自行添加 |
| `PaymentMethodSourceLianLianPay = 2` | LianLianPay |

### 3. 支付订单状态

| 状态常量 | 说明 |
|---------|------|
| `PaymentOrderStatusUnPay = 0` | 未支付 |
| `PaymentOrderStatusPaid = 1` | 已支付 |
| `PaymentOrderStatusRefund = 2` | 已退款 |
| `PaymentOrderStatusFailed = 3` | 支付失败 |

### 4. 连连支付订单状态

| 状态值 | 说明 |
|-------|------|
| `PI` | 初始化（未访问支付页操作） |
| `WP` | 等待支付 |
| `PS` | 支付成功 |
| `PF` | 支付失败 |
| `PE` | 支付已过期 |

### 5. 创建支付 (CreatePayment)

**功能描述**: 创建连连支付订单，生成支付二维码或支付链接。

**处理流程**:
```
1. 验证支付配置
   - 检查支付应用配置
   - 检查支付服务URL
   - 检查回调地址
   ↓
2. 生成商户订单号
   - 格式: PS + 时间戳 + 随机数
   ↓
3. 检查是否存在有效支付订单
   - 相同订单、相同金额、未过期
   - 如果存在，直接返回
   ↓
4. 获取支付配置
   - 根据支付方式代码获取支付接口URL
   - 确定订单类型
   ↓
5. 组装请求数据
   - 商家ID、订单号、金额、币种
   - 回调地址、支付方式、跳转地址
   ↓
6. 计算签名并发送请求
   - SHA256签名（sign_salt + JSON数据）
   - 发送HTTP POST请求
   ↓
7. 处理支付响应
   - 解析响应数据
   - 生成二维码（如果需要）
   ↓
8. 创建支付订单记录
   - 保存连连支付订单
   - 设置过期时间
   ↓
9. 返回支付订单
```

**支付方式映射**:
- `PaymentMethodCodeLianLianWechatPay` → `/api/receipts/lianlianWechatPay` → `LIANLIAN_WECHAT`
- `PaymentMethodCodeLianLianAliPay` → `/api/receipts/lianlianAliOfflinePay` → `LIANLIAN_ALI_OFFLINE_PAY`
- `PaymentMethodCodeLianLianQRPromptPay` → `/api/receipts/lianlianQrPromptPay` → `LIANLIAN_QR_PROMPT_PAY`

**二维码生成规则**:
- **微信支付**:
  - H5支付: 直接使用LinkUrl
  - 二维码支付: 将LinkUrl转换为二维码图片（Base64）
- **支付宝支付**:
  - H5支付: `alipays://platformapi/startapp?saId=10000007&qrcode={LinkUrl}`
  - 二维码支付: 将LinkUrl转换为二维码图片（Base64）
- **PromptPay**: 直接使用返回的QR code（Base64）

**二维码有效期**:
- 微信支付: 60分钟（H5支付5分钟）
- 支付宝支付: 15分钟
- PromptPay: 8分钟

### 6. 支付回调处理 (HandleCallback)

**功能描述**: 处理连连支付的支付成功回调。

**处理流程**:
```
1. 验证签名
   - 使用sign_salt计算签名
   - 与回调签名对比
   ↓
2. 验证支付状态
   - PayStatus必须为1（支付成功）
   ↓
3. 查询支付订单
   - 根据商户订单号、金额、用户ID查询
   - 订单必须未支付（pay_time = 0）
   ↓
4. 解析支付时间
   - 将PayAt转换为时间戳
   ↓
5. 计算手续费率
   - 手续费率 = 手续费 / 订单金额
   ↓
6. 更新支付订单（事务）
   - 更新连连支付订单状态为PS
   - 更新支付时间
   - 创建或更新支付订单记录
   ↓
7. 处理会员端订单
   - 如果是会员端订单，更新提交支付时间
   - 发布支付完成事件
   ↓
8. 返回成功
```

**回调数据结构**:
```go
type LianLianCallbackRequest struct {
    CompanyUuid     string // 商户ID
    MerchantOrderNo string // 商户订单号
    MerchantUserId  string // 商户用户ID
    PayTypeDesc     string // 支付类型描述
    PayStatus       int    // 支付状态（1=成功）
    PaymentId       string // 支付ID（交易号）
    OrderAmount     string // 订单金额
    OrderCurrency   string // 订单币种
    PayAt           string // 支付时间（格式：2006-01-02 15:04:05）
}
```

### 7. 退款处理 (Refund)

**功能描述**: 发起连连支付退款请求。

**处理流程**:
```
1. 验证支付配置
   ↓
2. 查询支付订单
   - 验证订单存在
   - 验证订单已支付
   ↓
3. 组装退款请求数据
   - 商户订单号、退款订单号
   - 退款金额、退款币种
   - 退款原因、回调地址
   - 银行信息（如果需要）
   ↓
4. 计算签名并发送请求
   ↓
5. 处理退款响应
   - 解析退款结果
   ↓
6. 返回退款结果
```

**退款请求结构**:
```go
type PaymentServiceRefundReq struct {
    PaymentOrderUuid      uint64  // 支付订单UUID
    RelatedType           int     // 支付订单类型
    MerchantRefundOrderNo string  // 商户退款订单号
    RefundAmount          float64 // 退款金额
    RefundOrderId         string  // 退款ID
    BankCode              string  // 银行代码
    AccountNo             string  // 账号
    AccountName           string  // 账号名称
    RefundRequestIndex    int     // 退款请求次数索引（重试用）
}
```

**退款重试机制**:
- 如果退款失败，自动重试一次
- 使用`RefundRequestIndex`标记重试次数

### 8. 退款回调处理 (HandleRefundCallback)

**功能描述**: 处理连连支付的退款回调。

**处理流程**:
```
1. 验证退款签名
   ↓
2. 查询退货单金额记录
   - 根据商户退款订单号查询
   ↓
3. 验证退款状态
   - 如果已处理（状态1或2），直接返回
   ↓
4. 更新退款状态
   - RefundStatus = "RS" → 退款成功（状态1）
   - 其他 → 退款失败（状态2）
   ↓
5. 推送WebSocket消息
   - 通知收银端退款状态更新
   ↓
6. 返回成功
```

**退款回调结构**:
```go
type LianLianRefundCallbackRequest struct {
    CompanyUuid      string // 商户ID
    RefundStatus     string // 退款状态（RS=成功）
    RefundOrderId    string // 退款订单ID
    PaymentOrderId   string // 支付订单ID
    MerchantRefundId string // 商户退款ID
}
```

### 9. 获取有效支付订单 (GetValidPaymentOrderByUuid)

**功能描述**: 根据关联订单UUID获取有效的待支付订单。

**查询条件**:
- 关联UUID、关联类型、商户用户ID
- 订单类型、订单金额、订单币种
- 未过期（过期时间 > 当前时间+5秒）或已支付

**用途**:
- 避免重复创建支付订单
- 复用未过期的支付二维码

### 10. 支付方式管理

#### 获取支付方式列表 (GetList)

**功能描述**: 根据客户端类型和场景获取可用的支付方式列表。

**参数说明**:
- `typ`: 支付方式显示类型
  - `PaymentMethodShowAll = "all"` - 显示所有
  - `PaymentMethodShowCheckout = "checkout"` - 结账时显示
  - `PaymentMethodShowRecharge = "recharge"` - 充值时显示

**处理流程**:
```
1. 验证显示类型
   ↓
2. 根据客户端类型筛选
   - 收银端：支持结账和充值
   - 助手端：只支持结账
   ↓
3. 查询支付方式列表
   - 状态为启用
   - 根据显示类型筛选
   ↓
4. 过滤支付方式
   - 不显示免单
   - 充值时不显示余额
   ↓
5. 构建响应数据
   - 加载Logo和二维码
   - 多语言名称
   ↓
6. 返回支付方式列表
```

#### 判断支付方式是否启用 (IsEnabled)

**功能描述**: 根据支付设置判断支付方式是否可用。

**判断逻辑**:
```
1. 获取支付设置
   ↓
2. 构建可用支付方式代码列表
   - 余额支付：IsBalance == "1"
   - 现金支付：IsCash == "1"
   - 其他支付：IsOther == "1" && 支付方式状态 == 启用
   ↓
3. 判断支付方式代码是否在列表中
```

### 11. 会员端订单退款 (MemberSaleOrderRefund)

**功能描述**: 会员端销售订单发起退款。

**处理流程**:
```
1. 创建退货单
   - 关联销售订单
   - 设置退款类型为全额退款
   ↓
2. 创建退货单金额记录
   - 遍历订单的支付订单
   - 为每个支付方式创建退款金额记录
   ↓
3. 发起退款
   - 遍历退款金额记录
   - 调用连连支付退款接口
   ↓
4. 更新退款金额记录
   - 保存连连退款订单ID
   ↓
5. 返回退货单
```

---

## 🔄 数据流转

### 支付流程

```
用户选择支付方式
    ↓
创建支付订单
    ↓
调用连连支付接口
    ↓
生成支付二维码/链接
    ↓
用户扫码/点击支付
    ↓
连连支付处理支付
    ↓
支付成功回调
    ↓
验证签名和状态
    ↓
更新支付订单状态
    ↓
创建支付订单记录
    ↓
发布支付完成事件
    ↓
订单状态更新
```

### 退款流程

```
用户申请退款
    ↓
创建退货单
    ↓
创建退款金额记录
    ↓
调用连连支付退款接口
    ↓
连连支付处理退款
    ↓
退款成功回调
    ↓
验证签名和状态
    ↓
更新退款状态
    ↓
推送WebSocket消息
```

### 支付订单状态流转

```
创建支付订单（状态：WP）
    ↓
用户支付
    ↓
支付成功回调
    ↓
更新为已支付（状态：PS）
    ↓
创建支付订单记录（状态：已支付）
    ↓
[可选] 退款
    ↓
更新为已退款（状态：已退款）
```

---

## 🔐 权限控制

### 支付接口

- **创建支付**: 需要认证（收银端/助手端/会员端）
- **支付回调**: 公开接口（连连支付服务器调用）
- **退款回调**: 公开接口（连连支付服务器调用）

### 支付方式查询

- **获取支付方式列表**: 需要认证
- **判断支付方式是否启用**: 内部接口

### 签名验证

- **支付回调**: 必须验证签名
- **退款回调**: 必须验证签名
- **签名算法**: SHA256(sign_salt + JSON数据)，取前32位

---

## ⚠️ 错误处理

### 常见错误

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| "未配置支付信息" | 支付应用配置不存在 | 配置支付应用信息 |
| "未配置PAY_SERVICE_URL" | 支付服务URL未配置 | 配置环境变量 |
| "未配置PAY_SERVICE_LIANLIAN_CALLBACK_URL" | 回调地址未配置 | 配置环境变量 |
| "请求支付失败" | 支付服务请求失败 | 检查网络和支付服务状态 |
| "支付签名验证失败" | 签名不匹配 | 检查sign_salt配置 |
| "支付订单不存在" | 订单UUID错误 | 检查订单UUID |
| "支付订单未支付" | 订单状态错误 | 检查订单状态 |
| "不支持的支付方式" | 支付方式代码错误 | 检查支付方式代码 |
| "活动二维码无效" | 二维码数据损坏 | 重新生成二维码 |

### 错误处理机制

1. **参数验证**: 所有接口都进行严格的参数验证
2. **签名验证**: 回调接口必须验证签名
3. **状态验证**: 验证订单状态、支付状态等
4. **事务处理**: 关键操作使用数据库事务保证一致性
5. **错误包装**: 使用 `errors.WithMessage` 包装错误信息
6. **重试机制**: 退款失败自动重试一次

---

## 📊 数据模型

### PaymentOrder - 支付订单

```go
type PaymentOrder struct {
    BaseModel
    PaymentMethodName    string  // 支付类型名称
    PaymentMethodUuid    uint64  // 支付类型UUID
    PaymentFeePercent    float64 // 支付手续费百分比
    RelatedType          int     // 关联订单类型（0=销售订单, 1=充值订单）
    RelatedUuid          uint64  // 关联订单UUID
    CurrencyUnit         string  // 货币单位
    PaymentAmount        float64 // 支付金额（扣除手续费后）
    PaymentCommissionFee float64 // 支付手续费
    Amount               float64 // 实收金额（支付金额+手续费）
    TransactionNumber    string  // 交易号
    Status               int     // 支付状态
    StatusReason         string  // 支付状态原因
    
    // 余额支付相关
    BalanceAmount     float64 // 主账户金额
    GiftBalanceAmount float64 // 赠送账户金额
}
```

### LlPaymentOrder - 连连支付订单

```go
type LlPaymentOrder struct {
    BaseModel
    MemberSaleOrderUuid uint64  // 会员销售订单UUID
    RelatedUuid         uint64  // 关联订单UUID
    PaymentOrderUuid    uint64  // 支付订单UUID
    PaymentMethodUuid   uint64  // 支付方式UUID
    RelatedType         int     // 关联订单类型
    MerchantId          string  // 连连商户号
    MerchantOrderId     string  // 商户订单号
    OrderId             string  // 连连订单ID
    OrderType           string  // 订单类型
    OrderStatus         string  // 订单状态（PI/WP/PS/PF/PE）
    OrderAmount         float64 // 订单金额
    OrderCurrency       string  // 订单币种
    CommissionFee       float64 // 支付手续费
    FullName            string  // 订单人名称
    OrderDesc           string  // 订单描述
    LinkUrl             string  // 支付链接/二维码
    MerchantUserId      string  // 商户用户ID
    LlCreateTime        string  // 连连订单创建时间
    PayTime             int64   // 支付时间
    ExpiredTime         int64   // 过期时间
}
```

### PaymentMethod - 支付方式

```go
type PaymentMethod struct {
    BaseModel
    Name                 string  // 支付方式名称
    Code                 int     // 支付方式代号
    PaymentName          string  // 支付名称
    Source               int     // 来源（0=系统, 1=手动, 2=LianLianPay）
    LogoFileUuid         uint64  // Logo图片UUID
    QrcodeFileUuid       uint64  // 二维码图片UUID
    FeePercent           float64 // 手续费百分比（0-1）
    IsShowCashier        int     // 收银机结账显示
    IsShowAssistant      int     // 点餐助手结账显示
    IsShowMemberRecharge int     // 收银机会员充值显示
    Status               int     // 状态（0=禁用, 1=启用）
    Sort                 int     // 排序
    DefaultImg           string  // 默认图片
    ErpnextPayment       string  // ERPNext支付方式
}
```

### PaymentApp - 支付应用配置

```go
type PaymentApp struct {
    BaseModel
    CompanyUuid          uint64 // 集团ID
    LlWhiteIp            string // 白名单IP
    LlMerchantId         string // 商户号
    LlStoreId            string // 站点ID
    LlPublicKey          string // LianLianpay公钥
    LlMerchantPrivateKey string // 商户私钥
    LlToken              string // Token
    LlSignSalt           string // 签名盐
}
```

### RefundOrder - 退款单

```go
type RefundOrder struct {
    BaseModel
    SaleOrderUuid    uint64  // 销售订单UUID
    SaleOrderNo      string  // 销售订单号
    PaymentOrderUuid uint64  // 支付单UUID
    RefundType       uint    // 退款类型（1=反结账）
    Amount           float64 // 退款金额
    Reason           string  // 退款原因
    Status           uint    // 退款状态
    ErpInvoiceName   string  // ERP发票名称
}
```

---

## 🚀 性能优化

### 支付订单复用

1. **有效订单检查**:
   - 创建支付前检查是否存在有效订单
   - 避免重复创建支付订单
   - 减少支付服务请求

2. **过期时间设置**:
   - 根据支付方式设置不同的过期时间
   - 微信：60分钟
   - 支付宝：15分钟
   - PromptPay：8分钟

### 签名计算优化

1. **签名算法**:
   - 使用SHA256哈希
   - 只取前32位作为签名
   - 提高计算效率

2. **签名缓存**:
   - 相同参数的签名可以缓存
   - 减少重复计算

### 数据库优化

1. **索引优化**:
   - 商户订单号索引
   - 关联订单UUID索引
   - 支付方式UUID索引

2. **查询优化**:
   - 使用预加载减少N+1查询
   - 分页查询优化

---

## 🧪 测试建议

### 单元测试

1. **支付创建测试**:
   - 测试不同支付方式的创建
   - 测试支付订单复用
   - 测试二维码生成

2. **签名验证测试**:
   - 测试签名计算
   - 测试签名验证
   - 测试签名错误处理

3. **回调处理测试**:
   - 测试支付回调处理
   - 测试退款回调处理
   - 测试回调签名验证

### 集成测试

1. **完整支付流程测试**:
   - 测试创建支付-支付-回调完整流程
   - 测试退款-回调完整流程
   - 测试支付订单状态流转

2. **异常场景测试**:
   - 测试支付超时
   - 测试支付失败
   - 测试退款失败
   - 测试重复回调

### 性能测试

1. **并发支付测试**:
   - 测试并发创建支付订单
   - 测试并发支付回调处理

2. **支付服务请求测试**:
   - 测试支付服务响应时间
   - 测试支付服务异常处理

---

## 📝 注意事项

1. **支付配置**:
   - 必须配置支付应用信息（PaymentApp）
   - 必须配置支付服务URL（PAY_SERVICE_URL）
   - 必须配置回调地址（PAY_SERVICE_LIANLIAN_CALLBACK_URL）
   - 必须配置签名盐（LlSignSalt）

2. **签名安全**:
   - 签名盐必须保密
   - 回调接口必须验证签名
   - 签名计算使用SHA256算法

3. **支付订单复用**:
   - 相同订单、相同金额、未过期的支付订单会被复用
   - 避免重复创建支付订单
   - 提高用户体验

4. **二维码有效期**:
   - 不同支付方式的二维码有效期不同
   - 过期后需要重新创建支付订单
   - 过期时间有5秒缓冲

5. **支付回调**:
   - 回调接口必须返回"success"
   - 回调处理必须幂等
   - 已支付的订单不能重复处理

6. **退款处理**:
   - 只能对已支付的订单退款
   - 退款金额不能超过支付金额
   - 退款有自动重试机制

7. **手续费计算**:
   - 手续费率范围：0-1（或1-100，自动转换）
   - 手续费计算保留3位小数，四舍五入到2位
   - 实收金额 = 支付金额 + 手续费

8. **支付方式显示**:
   - 收银端：支持结账和充值
   - 助手端：只支持结账
   - 充值时不显示余额支付
   - 不显示免单支付方式

9. **会员端订单支付**:
   - 会员端订单支付成功会更新提交支付时间
   - 会发布支付完成事件
   - 触发订单状态更新

10. **币种支持**:
    - 默认币种：THB（泰铢）
    - 支持币种：THB、USD、JPY、CNY
    - 币种在支付订单中记录

---

## 🔗 相关文档

- [支付方式服务文档](./payment_method.md) - 支付方式管理相关功能
- [订单服务文档](./order.md) - 订单支付相关功能
- [充值订单文档](./recharge_order.md) - 充值支付相关功能
- [会员服务文档](./member.md) - 会员余额支付相关功能

---

**文档版本**: v1.0  
**最后更新**: 2025-01-27  
**维护者**: TTPOS开发团队

