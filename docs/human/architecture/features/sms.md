# 短信服务 (SMS Service)

## 概述

`sms.go` 实现了短信发送管理服务，负责处理餐饮系统中各类短信通知的发送。该服务支持会员消费、充值、退款、验证码等多种场景的短信发送，提供手机号格式化（中国/泰国）、短信额度管理、并发控制、多语言支持等功能，是会员服务和通知系统的核心模块。

**文件路径**: `ttpos-server-go/main/app/service/sms.go`

## 核心功能

### 1. 会员业务短信
- 会员消费通知
- 会员充值通知
- 会员充值退款通知
- 会员订单退款通知

### 2. 会员验证码短信
- 通用验证码
- 注册验证码
- 认证验证码

### 3. 营销短信
- 会员积分变动通知
- 会员优惠券发放通知

### 4. 外送短信
- 外送订单取消通知

### 5. 短信额度管理
- 额度检查
- 额度扣减（双库同步）
- 额度不足拦截

### 6. 多地区支持
- 中国手机号（11位，前缀+86）
- 泰国手机号（9-10位，前缀+66）

## 接口定义

### ISmsSrv 接口

```go
type ISmsSrv interface {
    SendMemberConsumptionSMS(ctx context.Context, phone string, params *sms.MemberConsumptionRequest) error
    SendMemberRechargeSMS(ctx context.Context, phone string, params *sms.MemberRechargeRequest) error
    SendMemberRechargeRefundSMS(ctx context.Context, phone string, params *sms.MemberRechargeRefundRequest) error
    SendMemberOrderRefundSMS(ctx context.Context, phone string, params *sms.MemberOrderRefundRequest) error
    SendMemberCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error
    SendMemberRegisterCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error
    SendMemberAuthOrderCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error
    SendMemberPointsSMS(ctx context.Context, phone string, params *sms.MemberPointsRequest) error
    SendMemberCouponSMS(ctx context.Context, phone string, params *sms.MemberCouponRequest) error
    SendDeliveryOrderCancelSMS(ctx context.Context, phone string, params *sms.DeliveryOrderCancel) error
}
```

### smsSrv 结构体

```go
type smsSrv struct {
    bus    *event.SystemEventBus // 事件总线
    dbm    *database.DBManager   // 数据库管理器
    client sms.SMSClient          // SMS客户端
}
```

## 依赖项

### 内部依赖
- **sms.SMSClient**: 短信客户端，实现具体的短信发送逻辑
- **repository.CompanySettingRepo**: 商家配置仓库，管理短信额度
- **srvSetting**: 设置服务，获取门店配置

### 外部依赖
- **database.DBManager**: 数据库管理器
- **lock.Lock**: 并发锁
- **event.SystemEventBus**: 事件总线

## 短信类型

### 1. 业务通知类（需扣费）

| 短信类型 | 方法 | 扣费 | 业务场景 |
|---------|------|------|---------|
| 会员消费 | SendMemberConsumptionSMS | ✅ | 会员结账后通知 |
| 会员充值 | SendMemberRechargeSMS | ✅ | 会员充值成功通知 |
| 充值退款 | SendMemberRechargeRefundSMS | ✅ | 充值订单退款通知 |
| 订单退款 | SendMemberOrderRefundSMS | ✅ | 用餐订单退款通知 |
| 积分变动 | SendMemberPointsSMS | ✅ | 积分增减通知 |
| 优惠券 | SendMemberCouponSMS | ✅ | 优惠券发放通知 |
| 订单取消 | SendDeliveryOrderCancelSMS | ✅ | 外送订单取消通知 |

### 2. 验证码类（不扣费）

| 短信类型 | 方法 | 扣费 | 业务场景 |
|---------|------|------|---------|
| 通用验证码 | SendMemberCodeSMS | ❌ | 登录、验证等 |
| 注册验证码 | SendMemberRegisterCodeSMS | ❌ | 会员注册 |
| 认证验证码 | SendMemberAuthOrderCodeSMS | ❌ | 订单认证 |

**注意**: 验证码短信不检查和扣减额度，由短信服务商计费。

## 核心方法详解

### 1. SendMemberConsumptionSMS - 发送会员消费短信

**方法签名**:
```go
func (s *smsSrv) SendMemberConsumptionSMS(ctx context.Context, phone string, params *sms.MemberConsumptionRequest) error
```

**功能**: 发送会员消费通知短信，告知会员消费金额、积分变动等信息。

**请求参数**:
```go
type MemberConsumptionRequest struct {
    Company        string  // 商家名称
    MemberPay      float64 // 会员支付金额
    IncreasePoints int     // 增加的积分
    TotalPoints    int     // 总积分
}
```

**实现流程**:

```153:195:ttpos-server-go/main/app/service/sms.go
func (s *smsSrv) SendMemberConsumptionSMS(ctx context.Context, phone string, params *sms.MemberConsumptionRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	// 检查短信额度
	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 如果增加的积分和会员支付的金额都为0，则不发送短信
	if params.IncreasePoints == 0 && params.MemberPay == 0 {
		return errors.WithMessage(errors.New("增加的积分和会员支付的金额都为0，不发送短信"))
	}

	// 发送短信
	resp, err := s.client.SendMemberConsumptionSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}
```

**处理步骤**:
1. **并发锁**: 使用公司UUID+后缀作为锁键，防止并发发送
2. **额度检查**: 检查短信是否开启和额度是否充足
3. **手机号格式化**: 根据国家添加正确的国际区号
4. **业务验证**: 积分和金额都为0时不发送
5. **发送短信**: 调用SMS客户端发送
6. **扣减额度**: 发送成功后扣减额度

**短信内容示例**:
```
【XX餐厅】您好！您本次消费50.00元，获得积分10分，当前总积分110分。感谢您的光临！
```

---

### 2. SendMemberRechargeSMS - 发送会员充值短信

**方法签名**:
```go
func (s *smsSrv) SendMemberRechargeSMS(ctx context.Context, phone string, params *sms.MemberRechargeRequest) error
```

**功能**: 发送会员充值成功通知短信。

**请求参数**:
```go
type MemberRechargeRequest struct {
    Company       string  // 商家名称
    RechargeValue float64 // 充值金额
    GiftValue     float64 // 赠送金额
    Balance       float64 // 余额
}
```

**短信内容示例**:
```
【XX餐厅】您好！您已成功充值100.00元，赠送10.00元，当前余额110.00元。
```

---

### 3. SendMemberCodeSMS - 发送会员验证码短信

**方法签名**:
```go
func (s *smsSrv) SendMemberCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error
```

**功能**: 发送通用验证码短信（不扣费）。

**请求参数**:
```go
type MemberSendCodeRequest struct {
    Company string // 商家名称
    Code    string // 验证码
}
```

**实现流程**:

```311:343:ttpos-server-go/main/app/service/sms.go
func (s *smsSrv) SendMemberCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	formattedPhone, language, companyName, err := s.checkFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberCodeSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送失败，返回错误
	if resp.Code != sms.ResponseCodeSuccess {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}
```

**关键差异**:
- 使用 `checkFormatPhone`（不检查额度）
- 发送成功后**不扣减额度**
- 验证码由短信服务商单独计费

**短信内容示例**:
```
【XX餐厅】您的验证码是：123456，5分钟内有效，请勿泄露。
```

---

### 4. formatPhone - 格式化手机号（私有方法）

**方法签名**:
```go
func (s *smsSrv) formatPhone(phone string) (string, error)
```

**功能**: 根据手机号长度判断国家并添加国际区号。

**实现流程**:

```68:91:ttpos-server-go/main/app/service/sms.go
func (s *smsSrv) formatPhone(phone string) (string, error) {
	// 如果手机号已经有前缀，则去掉+66或+86
	if strings.HasPrefix(phone, constant.ThailandPrefix) {
		phone = strings.TrimPrefix(phone, constant.ThailandPrefix)
	}
	if strings.HasPrefix(phone, constant.ChinaPrefix) {
		phone = strings.TrimPrefix(phone, constant.ChinaPrefix)
	}

	if len(phone) == 10 || len(phone) == 9 {
		// 如果手机号以0开头，则去掉0
		if phone[0] == '0' {
			phone = phone[1:]
		}
		return constant.ThailandPrefix + phone, nil
	}
	if len(phone) == 11 {
		return constant.ChinaPrefix + phone, nil
	}
	return "", fmt.Errorf("invalid phone number")
}
```

**格式化规则**:

| 原始手机号 | 处理逻辑 | 格式化结果 | 国家 |
|-----------|---------|-----------|------|
| 13812345678 | 11位 | +8613812345678 | 中国 |
| 0812345678 | 10位，去前缀0 | +66812345678 | 泰国 |
| 812345678 | 9位 | +66812345678 | 泰国 |
| +8613812345678 | 去除前缀后重新格式化 | +8613812345678 | 中国 |
| +66812345678 | 去除前缀后重新格式化 | +66812345678 | 泰国 |

**特殊处理**:
- 泰国手机号以0开头会被去除（本地格式转国际格式）
- 已有国际区号的手机号会重新格式化确保正确

---

### 5. checkQuotaAndFormatPhone - 检查额度并格式化（私有方法）

**方法签名**:
```go
func (s *smsSrv) checkQuotaAndFormatPhone(ctx context.Context, phone string) (string, string, string, error)
```

**功能**: 检查短信额度、格式化手机号、获取语言和商家名称。

**实现流程**:

```101:114:ttpos-server-go/main/app/service/sms.go
func (s *smsSrv) checkQuotaAndFormatPhone(ctx context.Context, phone string) (string, string, string, error) {
	// 检查短信额度
	setting := repository.NewCompanySettingRepo(ctx.GetDB()).Get()
	if !setting.SmsEnabled() {
		err := fmt.Errorf("SMS service is not enabled, EnableSms: %d, SmsQuota: %d", setting.EnableSms, setting.SmsQuota)
		return "", "", "", errors.WithMessage(err, "没有开启短信或没有额度")
	}
	// 检查手机号格式
	formattedPhone, language, companyName, err := s.checkFormatPhone(ctx, phone)
	if err != nil {
		return "", "", "", err
	}
	return formattedPhone, language, companyName, nil
}
```

**返回值**:
- `formattedPhone`: 格式化后的手机号
- `language`: 语言（zh/en）
- `companyName`: 商家名称
- `error`: 错误信息

**额度检查**:
```go
func (s *CompanySetting) SmsEnabled() bool {
    return s.EnableSms == 1 && s.SmsQuota > 0
}
```

---

### 6. deductQuota - 扣减短信额度（私有方法）

**方法签名**:
```go
func (s *smsSrv) deductQuota(ctx context.Context, companyUuid uint64) error
```

**功能**: 扣减短信额度，同时更新商家库和SaaS库。

**实现流程**:

```141:151:ttpos-server-go/main/app/service/sms.go
func (s *smsSrv) deductQuota(ctx context.Context, companyUuid uint64) error {
	if err := repository.NewCompanySettingRepo(s.dbm.GetDB(ctx.GetCompanyUuid())).UpdateSmsQuota(companyUuid, 1); err != nil {
		err := fmt.Errorf("failed to update SMS quota: %v", err)
		return errors.WithMessage(err, "扣减短信额度失败")
	}
	if err := repository.NewCompanySettingRepo(s.dbm.GetDB(constant.DefaultDB)).UpdateSmsQuota(companyUuid, 1); err != nil {
		err := fmt.Errorf("failed to update saas SMS quota: %v", err)
		return errors.WithMessage(err, "saas扣减短信额度失败")
	}
	return nil
}
```

**双库同步**:
1. **商家库**: 商家自己的数据库，扣减 1 条额度
2. **SaaS库**: 全局统一数据库（DefaultDB），扣减 1 条额度

**目的**:
- 商家库：快速查询，业务使用
- SaaS库：统计汇总，计费依据

**扣减逻辑**:
```sql
UPDATE company_setting 
SET sms_quota = sms_quota - 1 
WHERE company_uuid = ? AND sms_quota > 0
```

---

### 7. selectLanguage - 选择语言（私有方法）

**方法签名**:
```go
func (s *smsSrv) selectLanguage(defaultLanguage string) string
```

**功能**: 根据门店默认语言选择短信语言。

**实现流程**:

```94:99:ttpos-server-go/main/app/service/sms.go
func (s *smsSrv) selectLanguage(defaultLanguage string) string {
	if defaultLanguage == "zh" {
		return "zh"
	}
	return "en"
}
```

**语言选择规则**:
- 门店语言为中文 → 发送中文短信
- 其他语言 → 发送英文短信

**支持的语言**:
- `zh`: 中文（简体）
- `en`: 英文

---

## 业务规则

### 1. 并发控制

**锁键生成**:
```go
const ISmsSrvLockSuffix = 1000

lockKey := company.Uuid + ISmsSrvLockSuffix
lock.NewSystemLock().LockUuid(lockKey)
defer lock.NewSystemLock().UnlockUuid(lockKey)
```

**目的**:
- 防止同一商家并发发送短信
- 避免额度重复扣减
- 保证短信发送的原子性

### 2. 短信额度管理

**检查条件**:
```go
EnableSms == 1 && SmsQuota > 0
```

**扣减时机**:
- 短信发送成功后扣减
- 失败不扣减

**扣减规则**:
- 每条短信扣减 1 个额度
- 验证码短信不扣减
- 双库同步扣减

### 3. 发送前验证

#### 业务通知类短信
1. 检查短信开关（`EnableSms`）
2. 检查短信额度（`SmsQuota > 0`）
3. 格式化手机号
4. 获取语言和商家名称
5. 业务参数验证

#### 验证码类短信
1. 格式化手机号
2. 获取语言和商家名称
3. **不检查额度**

### 4. 消费短信特殊规则

```go
if params.IncreasePoints == 0 && params.MemberPay == 0 {
    return errors.New("增加的积分和会员支付的金额都为0，不发送短信")
}
```

**不发送情况**:
- 会员支付金额为 0
- 增加积分为 0

**目的**: 避免发送无意义的短信，节省额度。

---

## 手机号格式化

### 支持的国家

| 国家 | 国际区号 | 号码长度 | 示例 |
|------|---------|---------|------|
| 中国 | +86 | 11位 | +8613812345678 |
| 泰国 | +66 | 9-10位 | +66812345678 |

### 格式化流程

```
输入手机号
    ↓
去除已有前缀（+86/+66）
    ↓
判断长度
    ↓
9-10位 → 泰国 → 去除前导0 → 添加+66
11位   → 中国 → 添加+86
其他   → 返回错误
```

### 泰国手机号特殊处理

**本地格式**:
```
0812345678  (10位，带0前缀)
```

**国际格式**:
```
+66812345678  (去除0前缀)
```

**转换逻辑**:
```go
if len(phone) == 10 || len(phone) == 9 {
    if phone[0] == '0' {
        phone = phone[1:]  // 去除前导0
    }
    return constant.ThailandPrefix + phone
}
```

---

## 使用场景

### 场景1: 会员消费后发送短信

```go
// 会员结账完成
order := completedOrder
member := order.Member

// 构建短信参数
params := &sms.MemberConsumptionRequest{
    Company:        "XX餐厅",
    MemberPay:      order.MemberPayAmount,  // 50.00
    IncreasePoints: order.EarnedPoints,     // 10
    TotalPoints:    member.Points,          // 110
}

// 发送短信
err := smsSrv.SendMemberConsumptionSMS(ctx, member.Phone, params)
if err != nil {
    // 短信发送失败不影响主流程
    logger.Warn("发送消费短信失败", zap.Error(err))
}
```

### 场景2: 会员充值成功发送短信

```go
// 充值订单支付成功
rechargeOrder := paidOrder

params := &sms.MemberRechargeRequest{
    Company:       "XX餐厅",
    RechargeValue: rechargeOrder.PayAmount,   // 100.00
    GiftValue:     rechargeOrder.GiftAmount,  // 10.00
    Balance:       member.Balance,            // 110.00
}

err := smsSrv.SendMemberRechargeSMS(ctx, member.Phone, params)
```

### 场景3: 会员注册发送验证码

```go
// 会员注册流程
phone := "13812345678"
code := generateVerifyCode()  // "123456"

params := &sms.MemberSendCodeRequest{
    Company: "XX餐厅",
    Code:    code,
}

// 发送验证码（不扣费）
err := smsSrv.SendMemberRegisterCodeSMS(ctx, phone, params)
if err != nil {
    return errors.New("验证码发送失败")
}

// 缓存验证码
cache.Set(fmt.Sprintf("verify_code:%s", phone), code, 5*time.Minute)
```

### 场景4: 订单退款发送短信

```go
// 订单退款完成
refundOrder := completedRefund

params := &sms.MemberOrderRefundRequest{
    Company:     "XX餐厅",
    RefundValue: refundOrder.RefundAmount,  // 50.00
    Balance:     member.Balance,            // 160.00
}

err := smsSrv.SendMemberOrderRefundSMS(ctx, member.Phone, params)
```

### 场景5: 积分变动通知

```go
// 手动调整会员积分
adjustment := pointsAdjustment

params := &sms.MemberPointsRequest{
    Company:     "XX餐厅",
    Points:      adjustment.Points,      // 增加或减少的积分
    TotalPoints: member.Points,          // 总积分
    Reason:      adjustment.Reason,      // 变动原因
}

err := smsSrv.SendMemberPointsSMS(ctx, member.Phone, params)
```

### 场景6: 优惠券发放通知

```go
// 发放优惠券给会员
coupon := distributedCoupon

params := &sms.MemberCouponRequest{
    Company:    "XX餐厅",
    CouponName: coupon.Name,        // "满100减20优惠券"
    ExpireDate: coupon.ExpireDate,  // "2024-12-31"
}

err := smsSrv.SendMemberCouponSMS(ctx, member.Phone, params)
```

### 场景7: 外送订单取消通知

```go
// 外送订单被取消
deliveryOrder := canceledOrder

params := &sms.DeliveryOrderCancel{
    Company:   "XX餐厅",
    OrderNo:   deliveryOrder.OrderNo,
    Reason:    deliveryOrder.CancelReason,
}

err := smsSrv.SendDeliveryOrderCancelSMS(ctx, customer.Phone, params)
```

---

## 最佳实践

### 1. 异步发送短信

```go
// 推荐：异步发送，不阻塞主流程
go func() {
    err := smsSrv.SendMemberConsumptionSMS(ctx, phone, params)
    if err != nil {
        logger.Error("发送消费短信失败", 
            zap.Error(err),
            zap.String("phone", phone),
        )
    }
}()
```

### 2. 错误处理

```go
// 短信发送失败不影响业务主流程
err := smsSrv.SendMemberRechargeSMS(ctx, phone, params)
if err != nil {
    // 记录日志，但不返回错误
    logger.Warn("充值短信发送失败",
        zap.Error(err),
        zap.String("phone", phone),
        zap.Float64("amount", params.RechargeValue),
    )
    // 可选：记录到发送失败队列，后续重试
    // retryQueue.Push(sms.Task{...})
}

// 继续执行主流程
return completeRecharge()
```

### 3. 手机号验证

```go
// 在业务层验证手机号
func validatePhone(phone string) error {
    // 移除空格和特殊字符
    phone = strings.ReplaceAll(phone, " ", "")
    phone = strings.ReplaceAll(phone, "-", "")
    
    // 验证格式
    if !regexp.MustCompile(`^[\d+]+$`).MatchString(phone) {
        return errors.New("手机号格式错误")
    }
    
    // 长度验证
    length := len(strings.TrimPrefix(strings.TrimPrefix(phone, "+86"), "+66"))
    if length < 9 || length > 11 {
        return errors.New("手机号长度错误")
    }
    
    return nil
}
```

### 4. 验证码管理

```go
// 验证码生成和校验
type VerifyCodeManager struct {
    cache cache.Cache
}

func (m *VerifyCodeManager) SendCode(phone string) error {
    // 1. 检查发送频率（1分钟内只能发送1次）
    key := fmt.Sprintf("sms:send_time:%s", phone)
    if m.cache.Exists(key) {
        return errors.New("发送过于频繁，请稍后再试")
    }
    
    // 2. 生成6位数字验证码
    code := fmt.Sprintf("%06d", rand.Intn(1000000))
    
    // 3. 发送短信
    err := smsSrv.SendMemberCodeSMS(ctx, phone, &sms.MemberSendCodeRequest{
        Code: code,
    })
    if err != nil {
        return err
    }
    
    // 4. 缓存验证码（5分钟有效）
    codeKey := fmt.Sprintf("sms:code:%s", phone)
    m.cache.Set(codeKey, code, 5*time.Minute)
    
    // 5. 记录发送时间（1分钟内不能再发）
    m.cache.Set(key, time.Now().Unix(), time.Minute)
    
    return nil
}

func (m *VerifyCodeManager) VerifyCode(phone, code string) bool {
    key := fmt.Sprintf("sms:code:%s", phone)
    cachedCode, err := m.cache.Get(key)
    if err != nil {
        return false
    }
    
    if cachedCode.(string) != code {
        return false
    }
    
    // 验证成功后删除
    m.cache.Delete(key)
    return true
}
```

### 5. 短信模板管理

```go
// 短信内容模板化
type SMSTemplate struct {
    templates map[string]map[string]string
}

func (t *SMSTemplate) Get(templateName, language string, params map[string]interface{}) string {
    template := t.templates[templateName][language]
    
    // 替换变量
    for key, value := range params {
        placeholder := fmt.Sprintf("{%s}", key)
        template = strings.ReplaceAll(template, placeholder, fmt.Sprint(value))
    }
    
    return template
}

// 使用示例
templates := &SMSTemplate{
    templates: map[string]map[string]string{
        "consumption": {
            "zh": "【{company}】您好！您本次消费{amount}元，获得积分{points}分，当前总积分{total_points}分。",
            "en": "Dear customer, you spent {amount} at {company}, earned {points} points. Total: {total_points}.",
        },
    },
}
```

---

## 性能优化

### 1. 批量发送

```go
// 批量发送短信（如营销活动）
func BatchSendSMS(members []Member, params *sms.MemberCouponRequest) error {
    // 使用工作池控制并发
    const workerCount = 10
    jobs := make(chan Member, len(members))
    errors := make(chan error, len(members))
    
    // 启动工作协程
    for i := 0; i < workerCount; i++ {
        go func() {
            for member := range jobs {
                err := smsSrv.SendMemberCouponSMS(ctx, member.Phone, params)
                if err != nil {
                    errors <- err
                }
            }
        }()
    }
    
    // 分发任务
    for _, member := range members {
        jobs <- member
    }
    close(jobs)
    
    // 收集错误
    var failedCount int
    for i := 0; i < len(members); i++ {
        select {
        case <-errors:
            failedCount++
        case <-time.After(time.Second):
            break
        }
    }
    
    logger.Info("批量发送完成",
        zap.Int("total", len(members)),
        zap.Int("failed", failedCount),
    )
    
    return nil
}
```

### 2. 短信队列

```go
// 使用队列异步处理
type SMSQueue struct {
    queue chan SMSTask
}

type SMSTask struct {
    Type   string
    Phone  string
    Params interface{}
}

func (q *SMSQueue) Start() {
    for i := 0; i < 5; i++ {  // 5个消费者
        go q.worker()
    }
}

func (q *SMSQueue) worker() {
    for task := range q.queue {
        err := q.sendSMS(task)
        if err != nil {
            // 失败重试或记录
            logger.Error("短信发送失败", zap.Error(err))
        }
        
        // 限流：每秒最多发送10条
        time.Sleep(100 * time.Millisecond)
    }
}

func (q *SMSQueue) Push(task SMSTask) {
    select {
    case q.queue <- task:
        // 成功加入队列
    default:
        // 队列满，记录日志
        logger.Warn("短信队列已满")
    }
}
```

### 3. 缓存优化

```go
// 缓存商家配置，减少数据库查询
type CachedSMSService struct {
    cache      cache.Cache
    smsSrv     ISmsSrv
    cacheTTL   time.Duration
}

func (s *CachedSMSService) checkQuota(ctx context.Context) error {
    cacheKey := fmt.Sprintf("sms:quota:%d", ctx.GetCompanyUuid())
    
    // 尝试从缓存获取
    if quota, err := s.cache.Get(cacheKey); err == nil {
        if quota.(int) > 0 {
            return nil
        }
        return errors.New("短信额度不足")
    }
    
    // 缓存未命中，查询数据库
    setting := repository.NewCompanySettingRepo(ctx.GetDB()).Get()
    
    // 缓存结果（30秒）
    s.cache.Set(cacheKey, setting.SmsQuota, 30*time.Second)
    
    if !setting.SmsEnabled() {
        return errors.New("短信服务未开启或额度不足")
    }
    
    return nil
}
```

---

## 安全考虑

### 1. 频率限制

```go
// 防止短信轰炸
type RateLimiter struct {
    cache cache.Cache
}

func (r *RateLimiter) CheckLimit(phone string) error {
    // 同一手机号1分钟内最多发送1条
    key := fmt.Sprintf("sms:limit:1m:%s", phone)
    count, _ := r.cache.Incr(key)
    if count == 1 {
        r.cache.Expire(key, time.Minute)
    }
    if count > 1 {
        return errors.New("发送过于频繁，请1分钟后再试")
    }
    
    // 同一手机号1小时内最多发送5条
    hourKey := fmt.Sprintf("sms:limit:1h:%s", phone)
    hourCount, _ := r.cache.Incr(hourKey)
    if hourCount == 1 {
        r.cache.Expire(hourKey, time.Hour)
    }
    if hourCount > 5 {
        return errors.New("今日发送次数已达上限")
    }
    
    // 同一IP 1小时内最多发送10条
    ipKey := fmt.Sprintf("sms:limit:ip:%s", getClientIP())
    ipCount, _ := r.cache.Incr(ipKey)
    if ipCount == 1 {
        r.cache.Expire(ipKey, time.Hour)
    }
    if ipCount > 10 {
        return errors.New("操作过于频繁")
    }
    
    return nil
}
```

### 2. 手机号验证

```go
// 防止恶意手机号
func validatePhoneSecurity(phone string) error {
    // 黑名单检查
    if isInBlacklist(phone) {
        return errors.New("该手机号已被禁止")
    }
    
    // 防止使用虚拟号码
    if isVirtualNumber(phone) {
        return errors.New("不支持虚拟号码")
    }
    
    return nil
}
```

### 3. 验证码安全

```go
// 验证码尝试次数限制
func checkVerifyAttempts(phone string) error {
    key := fmt.Sprintf("sms:verify:attempts:%s", phone)
    attempts, _ := cache.Incr(key)
    
    if attempts == 1 {
        cache.Expire(key, 10*time.Minute)
    }
    
    // 10分钟内最多尝试5次
    if attempts > 5 {
        return errors.New("验证失败次数过多，请稍后再试")
    }
    
    return nil
}
```

### 4. 敏感信息保护

```go
// 日志脱敏
func maskPhone(phone string) string {
    if len(phone) < 7 {
        return "***"
    }
    return phone[:3] + "****" + phone[len(phone)-4:]
}

// 使用
logger.Info("发送短信",
    zap.String("phone", maskPhone(phone)),  // +861****5678
    zap.String("type", "consumption"),
)
```

---

## 错误处理

### 1. 常见错误

| 错误场景 | 错误消息 | 处理方式 |
|---------|---------|---------|
| 短信未开启 | "没有开启短信或没有额度" | 提示充值或开启服务 |
| 手机号格式错误 | "手机号格式错误" | 提示正确格式 |
| 额度不足 | "短信额度不足" | 提示充值 |
| 发送失败 | "发送短信失败" | 记录日志，可选重试 |
| 扣减失败 | "扣减短信额度失败" | 人工核对 |
| 参数无效 | "增加的积分和会员支付的金额都为0" | 不发送短信 |

### 2. 错误处理示例

```go
err := smsSrv.SendMemberConsumptionSMS(ctx, phone, params)
if err != nil {
    // 分类处理错误
    errMsg := err.Error()
    
    switch {
    case strings.Contains(errMsg, "没有开启短信"):
        // 额度不足，提示充值
        return gin.H{
            "code":    1001,
            "message": "短信服务未开启或额度不足，请充值",
            "action":  "recharge",
        }
        
    case strings.Contains(errMsg, "手机号格式"):
        // 手机号错误
        return gin.H{
            "code":    1002,
            "message": "手机号格式不正确",
        }
        
    case strings.Contains(errMsg, "都为0"):
        // 业务参数问题，不发送
        logger.Info("消费短信跳过", zap.String("reason", "金额和积分都为0"))
        return nil
        
    default:
        // 其他错误，记录日志但不影响主流程
        logger.Error("发送短信失败",
            zap.Error(err),
            zap.String("phone", phone),
            zap.Any("params", params),
        )
        return nil  // 不阻断业务
    }
}
```

---

## 监控和统计

### 1. 发送统计

```go
// 短信发送统计
type SMSStatistics struct {
    Date         string
    TotalSent    int
    SuccessCount int
    FailureCount int
    QuotaUsed    int
    QuotaLeft    int
}

func recordSMSStatistics(companyUuid uint64, success bool) {
    key := fmt.Sprintf("sms:stats:%d:%s", companyUuid, time.Now().Format("20060102"))
    
    if success {
        cache.HIncrBy(key, "success_count", 1)
    } else {
        cache.HIncrBy(key, "failure_count", 1)
    }
    
    cache.HIncrBy(key, "total_sent", 1)
    cache.Expire(key, 30*24*time.Hour)  // 保留30天
}
```

### 2. 告警机制

```go
// 额度告警
func checkQuotaAlert(companyUuid uint64, quota int) {
    // 额度低于100条时告警
    if quota < 100 {
        sendAlertToAdmin(companyUuid, fmt.Sprintf("短信额度不足，剩余%d条", quota))
    }
    
    // 额度为0时紧急告警
    if quota == 0 {
        sendUrgentAlert(companyUuid, "短信额度已用完，请立即充值")
    }
}

// 失败率告警
func checkFailureRateAlert(companyUuid uint64) {
    stats := getSMSStatistics(companyUuid, today)
    
    if stats.TotalSent > 0 {
        failureRate := float64(stats.FailureCount) / float64(stats.TotalSent)
        
        // 失败率超过20%时告警
        if failureRate > 0.2 {
            sendAlertToAdmin(companyUuid, 
                fmt.Sprintf("短信发送失败率异常：%.2f%%", failureRate*100))
        }
    }
}
```

---

## 潜在改进点

### 1. 支持更多短信服务商

```go
// 多服务商支持
type SMSProvider string

const (
    ProviderAliyun  SMSProvider = "aliyun"   // 阿里云
    ProviderTencent SMSProvider = "tencent"  // 腾讯云
    ProviderTwilio  SMSProvider = "twilio"   // Twilio（国际）
)

type SMSClientFactory struct{}

func (f *SMSClientFactory) CreateClient(provider SMSProvider) sms.SMSClient {
    switch provider {
    case ProviderAliyun:
        return aliyun.NewClient()
    case ProviderTencent:
        return tencent.NewClient()
    case ProviderTwilio:
        return twilio.NewClient()
    default:
        return defaultClient
    }
}
```

### 2. 智能重试机制

```go
// 失败自动重试
type RetryConfig struct {
    MaxRetries int
    RetryDelay time.Duration
}

func SendWithRetry(sendFunc func() error, config RetryConfig) error {
    var lastErr error
    
    for i := 0; i < config.MaxRetries; i++ {
        err := sendFunc()
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // 重试前等待
        if i < config.MaxRetries-1 {
            time.Sleep(config.RetryDelay * time.Duration(i+1))
        }
    }
    
    return fmt.Errorf("重试%d次后仍失败: %v", config.MaxRetries, lastErr)
}
```

### 3. 短信模板管理

```go
// 数据库管理短信模板
type SMSTemplate struct {
    ID       uint64
    Name     string  // 模板名称
    Language string  // 语言
    Content  string  // 模板内容（支持变量）
    Status   int     // 状态
}

func (s *smsSrv) SendByTemplate(phone string, templateName string, vars map[string]interface{}) error {
    // 1. 获取模板
    template := getTemplate(templateName, language)
    
    // 2. 替换变量
    content := replaceVars(template.Content, vars)
    
    // 3. 发送
    return s.sendSMS(phone, content)
}
```

### 4. 国际化扩展

```go
// 支持更多国家
type CountryConfig struct {
    Country     string
    Prefix      string
    NumberLen   []int
    RemoveZero  bool
}

var supportedCountries = []CountryConfig{
    {Country: "CN", Prefix: "+86", NumberLen: []int{11}, RemoveZero: false},
    {Country: "TH", Prefix: "+66", NumberLen: []int{9, 10}, RemoveZero: true},
    {Country: "US", Prefix: "+1", NumberLen: []int{10}, RemoveZero: false},
    {Country: "JP", Prefix: "+81", NumberLen: []int{10}, RemoveZero: true},
}
```

### 5. 短信内容审核

```go
// 敏感词过滤
type ContentFilter struct {
    sensitiveWords []string
}

func (f *ContentFilter) Check(content string) error {
    for _, word := range f.sensitiveWords {
        if strings.Contains(content, word) {
            return fmt.Errorf("内容包含敏感词: %s", word)
        }
    }
    return nil
}
```

### 6. A/B测试支持

```go
// 短信内容A/B测试
type ABTestConfig struct {
    GroupA string  // 版本A内容
    GroupB string  // 版本B内容
    Ratio  float64 // A版本比例
}

func (s *smsSrv) SendWithABTest(phone string, config ABTestConfig) error {
    content := config.GroupA
    if rand.Float64() > config.Ratio {
        content = config.GroupB
    }
    
    // 记录使用版本
    recordABTestUsage(phone, content)
    
    return s.sendSMS(phone, content)
}
```

---

## 相关文件

### SMS客户端
- `ttpos-server-go/pkg/sms/client.go` - SMS客户端接口
- `ttpos-server-go/pkg/sms/request.go` - 请求参数定义
- `ttpos-server-go/pkg/sms/response.go` - 响应数据定义

### 数据仓库
- `ttpos-server-go/app/repository/company_setting.go` - 商家配置仓库

### 常量定义
- `ttpos-server-go/app/constant/country.go` - 国家代码和前缀
- `ttpos-server-go/app/constant/database.go` - 数据库常量

---

## 总结

短信服务是会员服务体系的重要组成部分，具有以下特点：

1. **多场景支持**: 10种短信类型，覆盖业务全流程
2. **多地区支持**: 支持中国和泰国手机号
3. **智能格式化**: 自动识别国家并添加国际区号
4. **额度管理**: 完善的额度检查和扣减机制
5. **并发控制**: 使用分布式锁防止重复发送
6. **多语言支持**: 根据门店配置选择短信语言
7. **双库同步**: 商家库和SaaS库同步扣减
8. **验证码特殊处理**: 验证码短信不扣减额度
9. **业务验证**: 智能判断是否需要发送
10. **完善的错误处理**: 详细的错误信息和日志

该服务为会员营销和通知提供了可靠的短信发送能力，提升了用户体验和商家运营效率。

