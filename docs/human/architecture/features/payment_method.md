# Payment Method Service 支付方式服务说明文档

## 📋 概述

`service/payment_method.go` 是 TTPOS 系统的支付方式管理服务，负责判断支付方式是否可用以及获取支付方式列表。该服务根据系统设置、会员功能状态、客户端类型等条件，动态返回可用的支付方式，支持收银端、助手端等多个终端。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/payment_method.go`  
**文件大小**: 133行  
**接口定义**: `IPaymentMethodSrv`  
**实现结构**: `paymentMethodSrv`

---

## 🏗️ 架构设计

### 接口定义 (IPaymentMethodSrv)

```go
type IPaymentMethodSrv interface {
    // 判断支付方式是否已启用
    IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool
    
    // 获取支付方式列表
    GetList(ctx context.Context, typ string) resp.PaymentMethodList
}
```

### 依赖服务

```go
type paymentMethodSrv struct {
    dbm        *database.DBManager  // 数据库管理器
    settingSrv setting.ISrv         // 设置服务
}
```

### 服务初始化

```go
func NewPaymentMethodSrv(dbm *database.DBManager, settingSrv setting.ISrv) IPaymentMethodSrv {
    return NewPaymentMethodSrvImpl(dbm, settingSrv)
}

func NewPaymentMethodSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IPaymentMethodSrv {
    return &paymentMethodSrv{
        dbm:        dbm,
        settingSrv: settingSrv,
    }
}
```

**初始化参数**:
- `dbm`: 数据库管理器，用于获取不同公司的数据库连接
- `settingSrv`: 设置服务，用于获取支付相关配置

---

## 🎯 核心功能

### 1. 判断支付方式是否启用 (IsEnabled)

**功能描述**: 根据系统支付设置和支付方式状态，判断某个支付方式是否可用。

#### 方法签名

```go
func (s *paymentMethodSrv) IsEnabled(
    ctx context.Context, 
    paymentMethod model.PaymentMethod, 
    companySetting model.CompanySetting
) bool
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `paymentMethod` | model.PaymentMethod | 支付方式对象 |
| `companySetting` | model.CompanySetting | 公司设置对象 |

#### 返回值

| 类型 | 说明 |
|-----|------|
| bool | `true` - 已启用，`false` - 未启用 |

#### 判断逻辑

```
1. 获取支付设置
   ↓
2. 初始化可用支付方式代码列表
   ↓
3. 判断余额支付是否开启
   - IsBalance == "1" → 添加余额支付代码
   ↓
4. 判断现金支付是否开启
   - IsCash == "1" → 添加现金支付代码
   ↓
5. 判断其他支付方式是否开启
   - IsOther == "1" && 支付方式状态启用 → 添加该支付方式代码
   ↓
6. 检查支付方式代码是否在可用列表中
   ↓
7. 返回判断结果
```

#### 代码实现

```go
func (s *paymentMethodSrv) IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool {
    // 获取支付设置
    paymentSetting, err := s.settingSrv.GetPaymentSetting(ctx, companySetting)
    if err != nil {
        ctx.Log().Error("获取支付设置失败", zap.Error(err))
        return false
    }
    
    var availableCodes []int
    
    // 判断余额支付
    if paymentSetting.IsBalance == "1" {
        availableCodes = append(availableCodes, constant.PaymentMethodCodeBalance)
    }
    
    // 判断现金支付
    if paymentSetting.IsCash == "1" {
        availableCodes = append(availableCodes, constant.PaymentMethodCodeCash)
    }
    
    // 判断其他支付方式
    if paymentSetting.IsOther == "1" && paymentMethod.Status == 1 {
        availableCodes = append(availableCodes, paymentMethod.Code)
    }
    
    return slices.Contains(availableCodes, paymentMethod.Code)
}
```

#### 支付方式代码常量

```go
const (
    PaymentMethodCodeBalance = 1   // 余额支付
    PaymentMethodCodeCash    = 2   // 现金支付
    PaymentMethodCodeFreePay = 3   // 免单
    // 其他支付方式代码（微信、支付宝等）
)
```

#### 使用示例

```go
// 判断微信支付是否可用
wechatPayment := model.PaymentMethod{
    Code:   10,  // 微信支付代码
    Status: 1,   // 启用状态
}

isEnabled := paymentMethodSrv.IsEnabled(ctx, wechatPayment, companySetting)
if isEnabled {
    // 微信支付可用
    fmt.Println("微信支付已启用")
} else {
    // 微信支付不可用
    fmt.Println("微信支付未启用")
}
```

#### 启用条件总结

| 支付方式 | 启用条件 |
|---------|---------|
| 余额支付 | `paymentSetting.IsBalance == "1"` |
| 现金支付 | `paymentSetting.IsCash == "1"` |
| 其他支付方式 | `paymentSetting.IsOther == "1" && paymentMethod.Status == 1` |

---

### 2. 获取支付方式列表 (GetList)

**功能描述**: 根据客户端类型和场景类型，返回可用的支付方式列表。

#### 方法签名

```go
func (s *paymentMethodSrv) GetList(ctx context.Context, typ string) resp.PaymentMethodList
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `typ` | string | 场景类型 |

#### 场景类型 (typ)

| 常量 | 值 | 说明 |
|-----|---|------|
| `PaymentMethodShowAll` | "all" | 显示所有支付方式 |
| `PaymentMethodShowRecharge` | "recharge" | 会员充值场景 |
| `PaymentMethodShowCheckout` | "checkout" | 结账场景 |

#### 返回数据结构

```go
type PaymentMethodList struct {
    List []PaymentMethodItem `json:"list"`
}

type PaymentMethodItem struct {
    SourceText    string  `json:"source_text"`     // 来源文本
    Uuid          uint64  `json:"uuid"`            // UUID
    PaymentName   string  `json:"payment_name"`    // 支付名称
    PaymentMethod string  `json:"payment_method"`  // 支付方式
    FeePercent    float64 `json:"fee_percent"`     // 手续费百分比
    Logo          string  `json:"logo"`            // Logo图片URL
    Qrcode        string  `json:"qrcode"`          // 二维码图片URL
    Code          int     `json:"code"`            // 支付方式代码
    Source        int     `json:"source"`          // 来源
}
```

#### 获取逻辑

```
1. 验证场景类型
   - 非法类型 → 返回空列表
   ↓
2. 获取数据库连接和公司设置
   ↓
3. 构建查询条件
   - 基础条件：状态=启用
   ↓
4. 根据客户端类型添加条件
   ├─ 收银端 (SourceCashier)
   │  ├─ 充值场景 → 查询会员充值支付方式
   │  └─ 结账场景 → 查询收银结账支付方式
   │
   └─ 助手端 (SourceAssistant)
      ├─ 充值场景 → 返回空列表
      └─ 结账场景 → 查询助手结账支付方式
   ↓
5. 预加载Logo和二维码文件
   ↓
6. 执行查询获取支付方式列表
   ↓
7. 遍历处理每个支付方式
   ├─ 过滤免单支付方式
   ├─ 处理余额支付显示条件
   ├─ 处理Logo图片URL
   ├─ 处理二维码图片URL
   └─ 国际化来源文本
   ↓
8. 返回处理后的列表
```

#### 代码实现

```go
func (s *paymentMethodSrv) GetList(ctx context.Context, typ string) resp.PaymentMethodList {
    // 1. 验证场景类型
    if !slices.Contains([]string{
        constant.PaymentMethodShowAll, 
        constant.PaymentMethodShowRecharge, 
        constant.PaymentMethodShowCheckout,
    }, typ) {
        return resp.PaymentMethodList{}
    }
    
    // 2. 获取数据库连接和公司设置
    paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
    companySetting := ctx.GetCompanySetting()
    
    // 3. 构建基础查询条件
    opts := []repository.DBOption{
        paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable),
    }
    
    // 4. 根据客户端类型添加条件
    if ctx.GetSource() == constant.SourceCashier {
        if typ != constant.PaymentMethodShowAll {
            switch typ {
            case constant.PaymentMethodShowRecharge:
                opts = append(opts, paymentMethodRepo.WhereCashierMemberRecharge())
            case constant.PaymentMethodShowCheckout:
                opts = append(opts, paymentMethodRepo.WhereCashier())
            }
        }
    } else if ctx.GetSource() == constant.SourceAssistant {
        if typ != constant.PaymentMethodShowAll {
            switch typ {
            case constant.PaymentMethodShowRecharge:
                return resp.PaymentMethodList{}  // 助手端不支持充值
            case constant.PaymentMethodShowCheckout:
                opts = append(opts, paymentMethodRepo.WhereAssistant())
            }
        }
    }
    
    // 5. 预加载Logo和二维码文件
    opts = append(opts, paymentMethodRepo.WithLogoFile(), paymentMethodRepo.WithQrcodeFile())
    
    // 6. 执行查询
    paymentMethods := paymentMethodRepo.GetPaymentMethodList(opts...)
    
    // 7. 遍历处理每个支付方式
    paymentMethodItems := make([]resp.PaymentMethodItem, 0, len(paymentMethods))
    for _, method := range paymentMethods {
        // 不显示免单
        if method.Code == constant.PaymentMethodCodeFreePay {
            continue
        }
        
        // 充值不显示余额
        if method.Code == constant.PaymentMethodCodeBalance &&
            (companySetting.IsOpenMember != 1 || typ == constant.PaymentMethodShowRecharge) {
            continue
        }
        
        // 处理图片URL
        var logo, qrcode string
        baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
        
        if method.LogoFile != nil {
            logo = method.LogoFile.GetUrl(baseUrl)
        }
        if logo == "" && method.DefaultImg != "" {
            logo = strings.TrimRight(baseUrl, "/") + method.DefaultImg
        }
        if method.QrcodeFile != nil {
            qrcode = method.QrcodeFile.GetUrl(baseUrl)
        }
        
        // 添加到结果列表
        paymentMethodItems = append(paymentMethodItems, resp.PaymentMethodItem{
            SourceText:    i18n.Translate(i18n.GetAcceptLanguage(ctx.GetGin()), constant.PaymentMethodSourceTextMap[method.Source]),
            Uuid:          method.Uuid,
            PaymentName:   method.GetPaymentName(),
            PaymentMethod: method.GetName(),
            FeePercent:    method.FeePercent,
            Logo:          logo,
            Qrcode:        qrcode,
            Code:          method.Code,
            Source:        method.Source,
        })
    }
    
    // 8. 返回结果
    return resp.PaymentMethodList{List: paymentMethodItems}
}
```

#### 客户端支持矩阵

| 客户端类型 | 充值场景 | 结账场景 | 查询方法 |
|-----------|---------|---------|---------|
| 收银端 (Cashier) | ✅ 支持 | ✅ 支持 | `WhereCashierMemberRecharge()` / `WhereCashier()` |
| 助手端 (Assistant) | ❌ 不支持 | ✅ 支持 | - / `WhereAssistant()` |

#### 过滤规则

| 规则 | 说明 |
|-----|------|
| 免单支付 | 始终不显示 (`Code == PaymentMethodCodeFreePay`) |
| 余额支付 | 会员功能未开启或充值场景时不显示 |

#### 使用示例

```go
// 收银端获取结账支付方式列表
ctx.SetSource(constant.SourceCashier)
paymentList := paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowCheckout)

// 遍历支付方式
for _, payment := range paymentList.List {
    fmt.Printf("支付方式: %s, 手续费: %.2f%%\n", payment.PaymentName, payment.FeePercent)
}

// 助手端获取结账支付方式列表
ctx.SetSource(constant.SourceAssistant)
paymentList = paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowCheckout)

// 助手端获取充值支付方式（返回空列表）
paymentList = paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowRecharge)
// paymentList.List 为空
```

---

## 📊 数据模型

### PaymentMethod 模型

```go
type PaymentMethod struct {
    Uuid          uint64    // UUID
    CompanyUuid   uint64    // 公司UUID
    Name          string    // 支付方式名称
    Code          int       // 支付方式代码
    Source        int       // 来源
    Status        int       // 状态 0-禁用 1-启用
    FeePercent    float64   // 手续费百分比
    DefaultImg    string    // 默认图片路径
    LogoFileId    uint64    // Logo文件ID
    QrcodeFileId  uint64    // 二维码文件ID
    LogoFile      *File     // Logo文件对象
    QrcodeFile    *File     // 二维码文件对象
    CreateTime    int64     // 创建时间
    UpdateTime    int64     // 更新时间
    DeleteTime    int64     // 删除时间
}

// GetPaymentName 获取支付名称
func (p *PaymentMethod) GetPaymentName() string {
    // 返回支付名称逻辑
}

// GetName 获取名称
func (p *PaymentMethod) GetName() string {
    return p.Name
}
```

### 支付方式代码枚举

| 代码 | 名称 | 说明 |
|-----|------|------|
| 1 | 余额支付 | 会员余额扣款 |
| 2 | 现金支付 | 现金结账 |
| 3 | 免单 | 免单支付（不显示） |
| 10 | 微信支付 | 第三方微信 |
| 11 | 支付宝支付 | 第三方支付宝 |
| ... | ... | 其他第三方支付 |

### 支付来源枚举

| 来源 | 说明 |
|-----|------|
| 1 | 系统内置 |
| 2 | 用户自定义 |

---

## 🔄 业务流程

### 1. 结账获取支付方式流程

```
用户结账
  ↓
调用GetList(ctx, "checkout")
  ↓
判断客户端类型
  ├─ 收银端
  │  ├─ 查询收银端支付方式
  │  └─ 包含：现金、余额、第三方支付
  │
  └─ 助手端
     ├─ 查询助手端支付方式
     └─ 包含：现金、第三方支付（无余额）
  ↓
过滤支付方式
  ├─ 排除免单
  ├─ 会员未开启排除余额
  └─ 仅显示启用的支付方式
  ↓
处理图片URL
  ├─ Logo图片
  └─ 二维码图片
  ↓
国际化处理
  ↓
返回支付方式列表
  ↓
前端显示支付选项
```

### 2. 会员充值获取支付方式流程

```
会员充值
  ↓
调用GetList(ctx, "recharge")
  ↓
判断客户端类型
  ├─ 收银端
  │  ├─ 查询充值支付方式
  │  └─ 包含：现金、第三方支付（无余额）
  │
  └─ 助手端
     └─ 返回空列表（不支持）
  ↓
过滤支付方式
  ├─ 排除免单
  ├─ 排除余额（充值场景）
  └─ 仅显示启用的支付方式
  ↓
处理图片URL
  ↓
返回支付方式列表
  ↓
前端显示充值支付选项
```

### 3. 判断支付方式可用流程

```
用户选择支付方式
  ↓
调用IsEnabled(ctx, paymentMethod, companySetting)
  ↓
获取支付设置
  ↓
构建可用支付方式代码列表
  ├─ 余额支付（IsBalance == "1"）
  ├─ 现金支付（IsCash == "1"）
  └─ 其他支付方式（IsOther == "1" && Status == 1）
  ↓
检查支付方式代码是否在可用列表中
  ↓
返回判断结果
  ├─ true → 允许使用该支付方式
  └─ false → 提示支付方式不可用
```

---

## 🎨 API接口示例

### 1. 获取结账支付方式接口

#### 请求

```http
GET /api/v1/payment_method/list?type=checkout
Authorization: Bearer {token}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "source_text": "系统内置",
        "uuid": 1,
        "payment_name": "现金",
        "payment_method": "现金支付",
        "fee_percent": 0.0,
        "logo": "http://example.com/images/cash.png",
        "qrcode": "",
        "code": 2,
        "source": 1
      },
      {
        "source_text": "系统内置",
        "uuid": 2,
        "payment_name": "余额",
        "payment_method": "余额支付",
        "fee_percent": 0.0,
        "logo": "http://example.com/images/balance.png",
        "qrcode": "",
        "code": 1,
        "source": 1
      },
      {
        "source_text": "用户自定义",
        "uuid": 10,
        "payment_name": "微信支付",
        "payment_method": "微信扫码支付",
        "fee_percent": 0.6,
        "logo": "http://example.com/images/wechat.png",
        "qrcode": "http://example.com/qrcodes/wechat_merchant.png",
        "code": 10,
        "source": 2
      }
    ]
  }
}
```

### 2. 获取充值支付方式接口

#### 请求

```http
GET /api/v1/payment_method/list?type=recharge
Authorization: Bearer {token}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "source_text": "系统内置",
        "uuid": 1,
        "payment_name": "现金",
        "payment_method": "现金支付",
        "fee_percent": 0.0,
        "logo": "http://example.com/images/cash.png",
        "qrcode": "",
        "code": 2,
        "source": 1
      },
      {
        "source_text": "用户自定义",
        "uuid": 10,
        "payment_name": "微信支付",
        "payment_method": "微信扫码支付",
        "fee_percent": 0.6,
        "logo": "http://example.com/images/wechat.png",
        "qrcode": "http://example.com/qrcodes/wechat_merchant.png",
        "code": 10,
        "source": 2
      }
    ]
  }
}
```

**注意**: 充值场景不返回余额支付方式。

### 3. Controller实现示例

```go
// GetPaymentMethodList 获取支付方式列表
// @Summary 获取支付方式列表
// @Description 根据场景类型获取可用的支付方式列表
// @Tags 支付方式
// @Accept json
// @Produce json
// @Param type query string true "场景类型" Enums(all, recharge, checkout)
// @Success 200 {object} resp.PaymentMethodList "成功"
// @Security JwtToken
// @Router /api/v1/payment_method/list [get]
func (c *PaymentMethodController) GetPaymentMethodList(ctx *gin.Context) {
    typ := ctx.Query("type")
    
    paymentList := c.paymentMethodSrv.GetList(ctx, typ)
    
    response.Success(ctx, paymentList)
}
```

---

## 🔧 配置说明

### 支付设置配置

支付方式的可用性由系统设置控制：

```go
type Payment struct {
    IsCash    string  // 现金支付开关 "0"-关闭 "1"-开启
    IsBalance string  // 余额支付开关 "0"-关闭 "1"-开启
    IsOther   string  // 其他方式支付开关 "0"-关闭 "1"-开启
}
```

### 公司设置配置

会员功能状态影响余额支付：

```go
type CompanySetting struct {
    IsOpenMember int  // 会员功能开关 0-关闭 1-开启
    // ... 其他字段
}
```

---

## 🎯 使用场景

### 1. 收银结账

```go
// 获取收银端结账支付方式
func (c *OrderController) GetCheckoutPaymentMethods(ctx *gin.Context) {
    // 设置客户端类型
    ctx.SetSource(constant.SourceCashier)
    
    // 获取支付方式列表
    paymentList := c.paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowCheckout)
    
    // 返回给前端
    response.Success(ctx, paymentList)
}
```

### 2. 会员充值

```go
// 获取会员充值支付方式
func (c *MemberController) GetRechargePaymentMethods(ctx *gin.Context) {
    // 设置客户端类型
    ctx.SetSource(constant.SourceCashier)
    
    // 获取支付方式列表
    paymentList := c.paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowRecharge)
    
    // 返回给前端（不包含余额支付）
    response.Success(ctx, paymentList)
}
```

### 3. 验证支付方式可用性

```go
// 处理支付请求
func (c *PaymentController) ProcessPayment(ctx *gin.Context) {
    var req req.PaymentReq
    ctx.ShouldBindJSON(&req)
    
    // 获取支付方式
    paymentMethodRepo := repository.NewPaymentMethodRepo(db)
    paymentMethod, _ := paymentMethodRepo.GetPaymentMethod(
        paymentMethodRepo.WhereUuid(req.PaymentMethodUuid),
    )
    
    // 获取公司设置
    companySetting := ctx.GetCompanySetting()
    
    // 验证支付方式是否可用
    if !c.paymentMethodSrv.IsEnabled(ctx, paymentMethod, companySetting) {
        response.Error(ctx, "该支付方式不可用")
        return
    }
    
    // 处理支付逻辑
    // ...
}
```

---

## 🛡️ 最佳实践

### 1. 获取支付方式列表

```go
// ✅ 正确：使用服务获取列表
paymentList := paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowCheckout)

// ❌ 错误：直接查询数据库
paymentMethods := paymentMethodRepo.GetPaymentMethodList(...)
```

### 2. 验证支付方式可用性

```go
// ✅ 正确：使用IsEnabled验证
if !paymentMethodSrv.IsEnabled(ctx, paymentMethod, companySetting) {
    return errors.New("支付方式不可用")
}

// ❌ 错误：仅检查Status字段
if paymentMethod.Status != 1 {
    return errors.New("支付方式不可用")  // 不够全面
}
```

### 3. 场景类型使用

```go
// ✅ 正确：使用常量
paymentList := paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowCheckout)

// ❌ 错误：使用硬编码字符串
paymentList := paymentMethodSrv.GetList(ctx, "checkout")  // 容易拼写错误
```

### 4. 客户端类型设置

```go
// ✅ 正确：在Controller层设置客户端类型
func (c *Controller) GetPaymentList(ctx *gin.Context) {
    ctx.SetSource(constant.SourceCashier)
    paymentList := c.paymentMethodSrv.GetList(ctx, typ)
}

// ✅ 正确：通过中间件自动设置
// middleware会根据token解析客户端类型
```

---

## ⚠️ 注意事项

### 1. 会员功能依赖

- 余额支付依赖会员功能开启
- 会员功能关闭时，余额支付不可用
- 充值场景下，余额支付始终不显示

### 2. 客户端支持差异

- 助手端不支持会员充值场景
- 不同客户端返回的支付方式列表不同
- 需要正确设置客户端类型

### 3. 免单支付

- 免单支付方式不会在列表中显示
- 免单通常在后台操作中使用
- 代码: `PaymentMethodCodeFreePay = 3`

### 4. 图片URL处理

- Logo和二维码需要添加域名前缀
- 优先使用上传的Logo文件
- 未上传时使用默认图片路径

### 5. 状态检查

- 支付方式必须是启用状态（Status = 1）
- 系统设置必须开启对应的支付开关
- 两个条件都满足才可用

---

## 📚 相关文档

- [Setting Service](./setting_service.md) - 设置服务（获取支付配置）
- [Payment Service](./payment_service.md) - 支付服务（处理支付逻辑）
- [Member Service](./member_service.md) - 会员服务（会员充值）

---

## 📊 服务特点总结

| 特点 | 说明 |
|-----|------|
| 简洁 | 仅133行代码，功能清晰 |
| 灵活 | 支持多客户端、多场景 |
| 可控 | 基于设置动态控制可用性 |
| 安全 | 多重验证确保支付方式合法 |
| 易用 | 接口简单，易于集成 |
| 国际化 | 支持多语言显示 |

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-11-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。支付方式服务是结账流程的重要组成部分，修改时需确保不影响现有支付功能。

