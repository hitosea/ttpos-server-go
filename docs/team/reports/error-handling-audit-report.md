# Main 模块错误处理审计报告

> 生成时间: 2026-01-11
> 审计范围: `/home/ben/projects/ttpos-server-go/main/`
> 审计目标: 排查被"吞掉"的错误（error 没有被记录到日志）

---

## 目录

- [问题统计概览](#问题统计概览)
- [高严重级别问题](#高严重级别问题)
- [中等严重级别问题](#中等严重级别问题)
- [低严重级别问题](#低严重级别问题)
- [修复指南](#修复指南)
- [最佳实践规范](#最佳实践规范)

---

## 问题统计概览

| 问题类型 | 高 | 中 | 低 | 总计 |
|---------|---|---|-----|-----|
| Panic 未记录 | 1 | - | - | 1 |
| JSON 错误被丢弃 | 1 | - | 9 | 10 |
| Defer 错误被忽略 | 1 | 2 | 1 | 4 |
| Goroutine 错误处理 | 1 | 1 | - | 2 |
| 类型断言失败 | 1 | 1 | - | 2 |
| 业务逻辑继续执行 | 1 | - | - | 1 |
| fmt.Println 混用 | - | - | 40+ | 40+ |
| **合计** | **5** | **4** | **50+** | **59+** |

---

## 高严重级别问题

### H-001: 业务逻辑错误后继续执行（可能导致 nil 指针）

**文件**: `main/app/service/business.go`
**行号**: 99-116
**风险**: 获取设置失败后继续使用可能为 nil 的对象，导致 panic

**问题代码**:
```go
storeSetting, err := setting.GetStoreSetting(ctx)
if err != nil {
    logger.Logger.Error("获取门店设置失败", zap.Error(err))
    fmt.Println("获取门店设置失败", zap.Error(err))
}
// 没有 return，继续使用 storeSetting，可能为 nil
```

**修复方案**:
```go
storeSetting, err := setting.GetStoreSetting(ctx)
if err != nil {
    logger.Logger.Error("获取门店设置失败", zap.Error(err))
    return nil, errors.WithMessage(err, "获取门店设置失败")
}
```

---

### H-002: Panic 未记录日志

**文件**: `main/app/service/rpc/client.go`
**行号**: 156-164
**风险**: 直接 panic 导致错误信息在日志中丢失，难以排查问题

**问题代码**:
```go
client, conn, err := erp.NewErpSellingClient()
if err != nil {
    panic(err)  // 直接 panic，未记录日志
}
defer conn.Close()
result, err := client.GetPosProfileList(...)
if err != nil {
    panic(err)  // 再次直接 panic，未记录日志
}
```

**修复方案**:
```go
client, conn, err := erp.NewErpSellingClient()
if err != nil {
    logger.Logger.Error("创建ERP客户端失败", zap.Error(err))
    return nil, errors.WithMessage(err, "创建ERP客户端失败")
}
defer func() {
    if err := conn.Close(); err != nil {
        logger.Logger.Warn("关闭gRPC连接失败", zap.Error(err))
    }
}()

result, err := client.GetPosProfileList(...)
if err != nil {
    logger.Logger.Error("获取POS配置列表失败", zap.Error(err))
    return nil, errors.WithMessage(err, "获取POS配置列表失败")
}
```

---

### H-003: JSON Marshal 错误被忽略

**文件**: `main/app/service/rpc/client.go`
**行号**: 171
**风险**: JSON 编码失败时数据丢失，调试困难

**问题代码**:
```go
ccccc, _ := json.Marshal(posProfileListResp.ProfileList)
fmt.Println(string(ccccc))
```

**修复方案**:
```go
data, err := json.Marshal(posProfileListResp.ProfileList)
if err != nil {
    logger.Logger.Error("序列化ProfileList失败", zap.Error(err))
    return nil, errors.WithMessage(err, "序列化ProfileList失败")
}
logger.Logger.Debug("ProfileList数据", zap.String("data", string(data)))
```

---

### H-004: 类型断言失败未完整处理

**文件**: `main/app/service/payment.go`
**行号**: 735-748
**风险**: 类型断言失败时错误信息不准确

**问题代码**:
```go
code, ok := responseMap["code"].(float64)
if !ok || code != 1 {
    msg, _ := responseMap["msg"].(string)  // ok 被丢弃
    if msg == "" {
        msg = "请求失败"
    }
    return nil, errors.New(msg)
}

responseData, ok := responseMap["data"].(map[string]interface{})
if !ok {
    return nil, errors.New("响应数据格式错误")
}
```

**修复方案**:
```go
code, ok := responseMap["code"].(float64)
if !ok {
    logger.Logger.Error("响应code类型断言失败", zap.Any("responseMap", responseMap))
    return nil, errors.New("响应格式错误: code字段类型不正确")
}
if code != 1 {
    msg, ok := responseMap["msg"].(string)
    if !ok || msg == "" {
        msg = "请求失败"
    }
    logger.Logger.Error("请求失败", zap.Float64("code", code), zap.String("msg", msg))
    return nil, errors.New(msg)
}

responseData, ok := responseMap["data"].(map[string]interface{})
if !ok {
    logger.Logger.Error("响应data类型断言失败", zap.Any("responseMap", responseMap))
    return nil, errors.New("响应数据格式错误")
}
```

---

### H-005: Defer 文件关闭错误忽略

**文件**: `main/app/model/printer_data.go`
**行号**: 80
**风险**: 文件关闭失败（磁盘满、权限问题等）时错误信息丢失

**问题代码**:
```go
defer file.Close()
```

**修复方案**:
```go
defer func() {
    if err := file.Close(); err != nil {
        logger.Logger.Error("关闭文件失败",
            zap.String("file", file.Name()),
            zap.Error(err))
    }
}()
```

---

## 中等严重级别问题

### M-001: Recover 后未记录日志

**文件**: `main/app/service/order_import_service.go`
**行号**: 90-93
**风险**: Panic 恢复后只回滚事务，未记录错误日志，调用者不知道发生了什么

**问题代码**:
```go
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        // 未记录日志，未设置错误状态
    }
}()
```

**修复方案**:
```go
defer func() {
    if r := recover(); r != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            logger.Logger.Error("回滚事务失败", zap.Error(rbErr))
        }
        logger.Logger.Error("导入订单发生panic",
            zap.Any("panic", r),
            zap.Stack("stack"))
        // 如果需要返回错误，可以使用命名返回值
    }
}()
```

---

### M-002: Goroutine 中错误未处理

**文件**: `main/app/service/payment.go`
**行号**: 582-587
**风险**: 异步操作的错误无法追踪

**问题代码**:
```go
utils.Go(func() {
    websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.UPDATE_REFUND_STATE, map[string]interface{}{
        "uuid":        orderAmount.Uuid,
        "update_time": orderAmount.BaseModel.UpdateTime,
    })
})
```

**修复方案**:
```go
utils.Go(func() {
    if err := websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.UPDATE_REFUND_STATE, map[string]interface{}{
        "uuid":        orderAmount.Uuid,
        "update_time": orderAmount.BaseModel.UpdateTime,
    }); err != nil {
        logger.Logger.Error("推送退款状态更新失败",
            zap.String("companyUuid", companyUuid),
            zap.String("orderUuid", orderAmount.Uuid),
            zap.Error(err))
    }
})
```

---

### M-003: JSON MarshalIndent 错误丢弃

**文件**: `main/middleware/request_logger.go`
**行号**: 100
**风险**: 日志格式化失败时信息丢失

**问题代码**:
```go
jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
```

**修复方案**:
```go
jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
if err != nil {
    logBuffer.WriteString(fmt.Sprintf("JSON格式化失败: %v\n", err))
    // 回退到原始数据
    jsonBytes = []byte(fmt.Sprintf("%+v", jsonData))
}
```

---

### M-004: 数据库连接关闭错误忽略

**文件**: `main/command/verify_setting_service.go`
**行号**: 1294, 1318
**风险**: 数据库连接关闭失败时资源泄漏

**问题代码**:
```go
db, err := sql.Open("mysql", dsn)
if err != nil {
    return fmt.Errorf("连接数据库失败: %w", err)
}
defer db.Close()  // 错误被忽略
```

**修复方案**:
```go
db, err := sql.Open("mysql", dsn)
if err != nil {
    return fmt.Errorf("连接数据库失败: %w", err)
}
defer func() {
    if err := db.Close(); err != nil {
        logger.Logger.Warn("关闭数据库连接失败", zap.Error(err))
    }
}()
```

---

## 低严重级别问题

### L-001: fmt.Println 与日志混用

**影响文件**:

| 文件 | 行号 |
|------|------|
| `main/app/service/business.go` | 101, 108, 115 |
| `main/app/service/payment.go` | 395, 401, 425, 431, 529, 541, 548, 560, 576, 725 |
| `main/app/model/printer_data.go` | 42, 61, 68, 77, 86, 126, 133 |

**问题代码**:
```go
if err != nil {
    logger.Logger.Error("获取门店设置失败", zap.Error(err))
    fmt.Println("获取门店设置失败", zap.Error(err))  // 重复且不规范
}
```

**修复方案**:
移除所有 `fmt.Println()` 调用，统一使用 `logger.Logger`：
```go
if err != nil {
    logger.Logger.Error("获取门店设置失败", zap.Error(err))
}
```

**批量查找命令**:
```bash
grep -rn "fmt.Println" main/app/ | grep -v "_test.go"
grep -rn "fmt.Printf" main/app/ | grep -v "_test.go"
```

---

### L-002: JSON Marshal 错误丢弃（多处）

**影响文件**:

| 文件 | 行号 |
|------|------|
| `main/middleware/encrypt.go` | 76 |
| `main/middleware/request_logger.go` | 100 |
| `main/app/service/encrypt.go` | 59 |
| `main/app/service/setting/setting.go` | 134, 208 |

**问题代码**:
```go
b, _ := json.Marshal(map[string]string{"encrypted": encryptedResponse})
data, _ := json.Marshal(settings)
```

**修复方案**:
```go
b, err := json.Marshal(map[string]string{"encrypted": encryptedResponse})
if err != nil {
    logger.Logger.Error("JSON编码失败", zap.Error(err))
    return nil, errors.WithMessage(err, "JSON编码失败")
}
```

---

### L-003: os.Remove 错误未处理

**文件**: `main/app/model/printer_data.go`
**行号**: 71

**问题代码**:
```go
defer os.Remove(tmpfile.Name())
```

**修复方案**:
```go
defer func() {
    if err := os.Remove(tmpfile.Name()); err != nil {
        logger.Logger.Warn("删除临时文件失败",
            zap.String("file", tmpfile.Name()),
            zap.Error(err))
    }
}()
```

---

## 修复指南

### 1. Defer 资源关闭标准模板

```go
// 文件关闭
defer func() {
    if err := file.Close(); err != nil {
        logger.Logger.Error("关闭文件失败",
            zap.String("file", file.Name()),
            zap.Error(err))
    }
}()

// 数据库连接关闭
defer func() {
    if err := db.Close(); err != nil {
        logger.Logger.Warn("关闭数据库连接失败", zap.Error(err))
    }
}()

// gRPC 连接关闭
defer func() {
    if err := conn.Close(); err != nil {
        logger.Logger.Warn("关闭gRPC连接失败", zap.Error(err))
    }
}()

// HTTP Response Body 关闭
defer func() {
    if err := resp.Body.Close(); err != nil {
        logger.Logger.Warn("关闭HTTP响应体失败", zap.Error(err))
    }
}()

// 数据库行关闭
defer func() {
    if err := rows.Close(); err != nil {
        logger.Logger.Warn("关闭数据库行失败", zap.Error(err))
    }
}()
```

---

### 2. Recover 标准模板

```go
// 基础模板
defer func() {
    if r := recover(); r != nil {
        logger.Logger.Error("发生panic",
            zap.Any("panic", r),
            zap.Stack("stack"))
    }
}()

// 带事务回滚的模板
defer func() {
    if r := recover(); r != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            logger.Logger.Error("回滚事务失败", zap.Error(rbErr))
        }
        logger.Logger.Error("事务执行发生panic",
            zap.Any("panic", r),
            zap.Stack("stack"))
    }
}()

// 带命名返回值的模板（可以返回错误）
func doSomething() (result Result, err error) {
    defer func() {
        if r := recover(); r != nil {
            logger.Logger.Error("发生panic",
                zap.Any("panic", r),
                zap.Stack("stack"))
            err = fmt.Errorf("panic recovered: %v", r)
        }
    }()
    // 业务逻辑...
    return
}
```

---

### 3. 错误处理标准模板

```go
// 必须中断的错误
result, err := someFunc()
if err != nil {
    logger.Logger.Error("操作失败",
        zap.String("operation", "someFunc"),
        zap.Error(err))
    return nil, errors.WithMessage(err, "操作失败")
}

// 可降级的错误
result, err := someFunc()
if err != nil {
    logger.Logger.Warn("操作失败，使用默认值",
        zap.String("operation", "someFunc"),
        zap.Error(err))
    result = defaultValue
}

// Goroutine 中的错误
utils.Go(func() {
    if err := asyncFunc(); err != nil {
        logger.Logger.Error("异步操作失败",
            zap.String("operation", "asyncFunc"),
            zap.Error(err))
    }
})
```

---

### 4. 类型断言标准模板

```go
// 必须成功的断言
val, ok := x.(SomeType)
if !ok {
    logger.Logger.Error("类型断言失败",
        zap.String("expected", "SomeType"),
        zap.Any("actual", x))
    return nil, errors.New("类型断言失败")
}

// 可选的断言（有默认值）
val, ok := x.(string)
if !ok {
    logger.Logger.Debug("类型断言失败，使用默认值",
        zap.Any("value", x))
    val = "default"
}
```

---

### 5. JSON 操作标准模板

```go
// Marshal
data, err := json.Marshal(obj)
if err != nil {
    logger.Logger.Error("JSON编码失败",
        zap.Any("object", obj),
        zap.Error(err))
    return nil, errors.WithMessage(err, "JSON编码失败")
}

// Unmarshal
var result ResultType
if err := json.Unmarshal(data, &result); err != nil {
    logger.Logger.Error("JSON解码失败",
        zap.String("data", string(data)),
        zap.Error(err))
    return nil, errors.WithMessage(err, "JSON解码失败")
}
```

---

## 最佳实践规范

### 1. 错误处理原则

- **永不忽略错误**: 所有返回 error 的函数调用都必须检查
- **使用 `_` 需谨慎**: 只有在确定错误不影响业务时才能使用
- **错误必须记录**: 即使不返回，也要记录日志
- **错误必须处理**: 要么返回，要么降级，要么恢复

### 2. 日志规范

- **统一使用 logger**: 禁止使用 `fmt.Println`、`fmt.Printf`
- **结构化日志**: 使用 zap 的字段方式，而非字符串拼接
- **适当的日志级别**:
  - `Error`: 需要立即关注的错误
  - `Warn`: 异常但可恢复的情况
  - `Info`: 重要的业务事件
  - `Debug`: 调试信息

### 3. 资源管理

- **defer 必须检查错误**: 所有 `Close()` 操作都要检查返回值
- **使用命名返回值**: 在需要 defer 中修改返回值时使用

### 4. Goroutine 安全

- **使用 utils.Go**: 内置 panic 恢复
- **处理返回错误**: 异步操作的错误也要记录
- **考虑超时**: 长时间运行的 goroutine 需要超时控制

---

## 修复优先级建议

1. **立即修复**: 高严重级别问题 (H-001 ~ H-005)
2. **本周修复**: 中等严重级别问题 (M-001 ~ M-004)
3. **逐步清理**: 低严重级别问题 (L-001 ~ L-003)

---

## 附录：快速查找命令

```bash
# 查找所有被忽略的错误
grep -rn ", _ :=" main/app/ | grep -v "_test.go"
grep -rn ", _ =" main/app/ | grep -v "_test.go"

# 查找所有 fmt.Println
grep -rn "fmt.Println" main/app/ | grep -v "_test.go"

# 查找所有 defer xxx.Close()
grep -rn "defer.*\.Close()" main/app/ | grep -v "_test.go"

# 查找所有 panic
grep -rn "panic(" main/app/ | grep -v "_test.go"

# 查找所有 recover 但没有日志的
grep -A5 "recover()" main/app/ | grep -v "logger\|Logger\|log\."
```

---

## 补充审计：错误判断逻辑与日志脱敏问题

> 补充时间: 2026-01-11
> 审计重点: 错误判断逻辑吞错误、日志敏感信息泄露

---

### 高严重级别问题（补充）

#### H-006: errors.Is 使用错误导致条件永远为 false

**文件**: `main/app/modules/takeout/infrastructure/adapter/rpc/takeout_rpc_service.go`
**行号**: 66-68
**风险**: 条件判断永远不成立，错误被错误地传递到上层

**问题代码**:
```go
if errors.Is(err, errors.New("no rows in result set")) {
    return false, 0, "", "", nil
}
```

**问题分析**:
`errors.Is()` 比较的是错误实例的引用，每次 `errors.New()` 都会创建新的错误实例。这意味着这个条件**永远为 false**，导致：
1. 本应返回 nil 的情况返回了错误
2. 调用方收到意外的错误

**修复方案**:
```go
// 方案1: 使用字符串匹配（不推荐但可行）
if err != nil && strings.Contains(err.Error(), "no rows in result set") {
    return false, 0, "", "", nil
}

// 方案2: 定义全局错误变量（推荐）
var ErrNoRows = errors.New("no rows in result set")
// 在产生错误的地方使用这个变量
// 然后:
if errors.Is(err, ErrNoRows) {
    return false, 0, "", "", nil
}
```

---

#### H-007: 日志打印密码哈希值

**文件**: `main/app/service/member.go`
**行号**: 543-545
**风险**: 密码哈希值泄露到日志，存在安全风险

**问题代码**:
```go
md5Password := cryptor.Md5String(discountReq.Password)
ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
if member.Password != md5Password {
    ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
    return errors.New("密码错误")
}
```

**问题分析**:
1. MD5 哈希后的密码仍是敏感信息
2. 日志中同时包含用户输入密码的哈希和数据库存储的哈希
3. 攻击者可利用这些信息进行彩虹表攻击

**修复方案**:
```go
md5Password := cryptor.Md5String(discountReq.Password)
if member.Password != md5Password {
    ctx.Log().Debug("验证密码失败", zap.Uint64("memberUuid", member.Uuid))
    return errors.New("密码错误")
}
// 成功时无需打印任何密码相关信息
```

---

#### H-008: 日志打印手机号未脱敏

**影响文件及行号**:

| 文件 | 行号 | 场景 |
|------|------|------|
| `app/event/order/order_checkout_event_handler.go` | 755-756 | 发送优惠券短信 |
| `app/event/order/order_checkout_event_handler.go` | 819-820 | 发送积分短信 |
| `app/service/order_manage.go` | 1409, 1411 | 发送退款短信 |
| `app/service/order_manage.go` | 2168, 2170 | 消费反结账退款 |
| `app/service/recharge_order.go` | 632, 634 | 发送充值短信 |
| `app/service/recharge_order.go` | 1600, 1602 | 充值反结账 |
| `app/service/recharge_order.go` | 2009, 2011 | 充值退款 |
| `app/service/order_pay.go` | 1078, 1080 | 发送结账短信 |

**问题代码示例**:
```go
ctx.Log().Info("发送退款短信失败", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
```

**风险**: 用户手机号属于个人隐私信息（PII），直接打印违反数据保护合规要求

**修复方案**:
```go
// 创建脱敏函数
func MaskPhone(phone string) string {
    if len(phone) <= 4 {
        return "****"
    }
    return phone[:3] + "****" + phone[len(phone)-4:]
}

// 使用脱敏后的手机号
ctx.Log().Info("发送退款短信失败",
    zap.String("phone", MaskPhone(member.Phone)),
    zap.Error(err))
```

---

#### H-009: 请求日志中间件打印完整请求体

**文件**: `main/middleware/request_logger.go`
**行号**: 86-130
**风险**: 请求体中可能包含密码、token 等敏感信息被完整打印

**问题代码**:
```go
if c.Request.Body != nil {
    bodyBytes, err := io.ReadAll(c.Request.Body)
    if err == nil && len(bodyBytes) > 0 {
        // 重新设置请求体
        c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

        contentType := c.GetHeader("Content-Type")
        logBuffer.WriteString(fmt.Sprintf("------ 请求体 ------\n"))

        // 直接打印完整请求体，无任何脱敏处理
        var jsonData interface{}
        if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
            jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
            logBuffer.WriteString(string(jsonBytes))
        }
    }
}
```

**风险场景**:
- 登录接口：密码明文被打印
- 支付接口：银行卡信息被打印
- 会员注册：身份证号被打印

**修复方案**:
```go
// 定义敏感字段列表
var sensitiveFields = []string{
    "password", "pwd", "secret", "token",
    "access_token", "refresh_token", "api_key",
    "id_card", "idcard", "card_number", "cvv",
}

// 脱敏函数
func maskSensitiveFields(data map[string]interface{}) map[string]interface{} {
    for key, value := range data {
        lowerKey := strings.ToLower(key)
        for _, sensitive := range sensitiveFields {
            if strings.Contains(lowerKey, sensitive) {
                data[key] = "***MASKED***"
                break
            }
        }
        // 递归处理嵌套对象
        if nested, ok := value.(map[string]interface{}); ok {
            data[key] = maskSensitiveFields(nested)
        }
    }
    return data
}
```

---

### 中等严重级别问题（补充）

#### M-005: isMySQLError 判断后非数据库错误处理不当

**文件**: `main/app/api/helper/helper.go`
**行号**: 98-110, 123-135, 143-155, 164-176

**问题代码**:
```go
func Error(c *gin.Context, err error) {
    code := constant.CodeFail

    if isMySQLError(err) {
        // 数据库错误：隐藏细节
        i18n.SendResponse(c, code, err, "system_error", nil, nil)
        return
    }

    // 非数据库错误处理
    messages := []string{err.Error()}
    var appErr errors.AppError
    if pkgerrors.As(err, &appErr) {
        // 只处理 AppError，其他错误类型被当作普通错误
    }
    // ...
}
```

**风险**:
- 非 MySQL 错误且非 AppError 的错误可能包含内部实现细节
- 错误信息可能直接暴露给前端

**建议**: 在 release 模式下，非 AppError 的错误应返回通用错误信息

---

#### M-006: 循环中错误被 continue 跳过但未记录

**影响位置**:

| 文件 | 行号 | 问题 |
|------|------|------|
| `repository/cache_object_controller.go` | 64 | 查询失败后 continue |
| `repository/cache_object_controller.go` | 268 | 查询失败后 continue |
| `service/stock_reconciliation.go` | 122 | copier.Copy 失败后 continue |
| `service/stock_reconciliation.go` | 271 | copier.Copy 失败后 continue |
| `service/auth.go` | 1822-1823 | 查询失败后 continue |

**问题代码示例** (`cache_object_controller.go:61-64`):
```go
setting, err := orderRepo.GetSaleBillSetting(saleBillUuid)
if err != nil {
    // 单个查询失败不影响其他，继续查询
    continue  // 错误被吞掉，无任何日志
}
```

**修复方案**:
```go
setting, err := orderRepo.GetSaleBillSetting(saleBillUuid)
if err != nil {
    logger.Logger.Warn("查询销售账单设置失败，跳过",
        zap.Uint64("saleBillUuid", saleBillUuid),
        zap.Error(err))
    continue
}
```

---

#### M-007: 字符串匹配判断错误类型（不可靠）

**影响位置**:

| 文件 | 行号 | 问题代码 |
|------|------|---------|
| `service/order.go` | 232 | `strings.Contains(err.Error(), "record not found")` |
| `service/product.go` | 2642 | `strings.Contains(err.Error(), "not found")` |
| `service/supplier.go` | 268 | `strings.Contains(err.Error(), "not found")` |
| `service/warehouse.go` | 414 | `strings.Contains(err.Error(), "not found")` |

**问题代码**:
```go
if err != nil && !strings.Contains(err.Error(), "record not found") {
    return nil, false, errors.WithMessage(err, "获取待支付、未挂单的订单失败")
}
```

**风险**:
- 错误消息格式变化会导致判断失效
- 不同语言环境下错误消息可能不同
- 错误消息可能被包装导致匹配失败

**修复方案**:
```go
// 使用 errors.Is 判断 GORM 错误
if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, false, errors.WithMessage(err, "获取待支付、未挂单的订单失败")
}
```

---

### 低严重级别问题（补充）

#### L-004: fmt.Println 与 logger 同时使用（重复输出）

**影响位置**:

| 文件 | 行号 |
|------|------|
| `app/event/order/order_checkout_event_handler.go` | 755-756 |
| `app/event/order/order_checkout_event_handler.go` | 819-820 |

**问题代码**:
```go
fmt.Println("HandleActivitySendReward process, SendMemberCouponSMS failed", zap.Any("activityUuid", activityUuid), zap.Any("phone", member.Phone), zap.Error(err))
logger.Logger.Info("HandleActivitySendReward process, SendMemberCouponSMS failed", zap.Any("activityUuid", activityUuid), zap.Any("phone", member.Phone), zap.Error(err))
```

**问题**: 同一条日志输出两次，且 fmt.Println 无法正确解析 zap 字段

---

### 敏感信息脱敏规范

#### 1. 需要脱敏的字段

| 类型 | 字段名 | 脱敏规则 |
|------|--------|---------|
| 密码 | password, pwd, secret | 完全隐藏: `***` |
| Token | token, access_token, api_key | 仅显示前4后4: `abcd****efgh` |
| 手机号 | phone, mobile, tel | 中间4位隐藏: `138****1234` |
| 身份证 | id_card, idcard | 仅显示前3后4: `310***********1234` |
| 银行卡 | card_number, bank_card | 仅显示后4位: `************1234` |
| 邮箱 | email | 部分隐藏: `a***b@example.com` |

#### 2. 脱敏工具函数

```go
package utils

import "strings"

// MaskPassword 密码脱敏
func MaskPassword(pwd string) string {
    return "***"
}

// MaskPhone 手机号脱敏
func MaskPhone(phone string) string {
    if len(phone) < 7 {
        return "****"
    }
    return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskIDCard 身份证脱敏
func MaskIDCard(idCard string) string {
    if len(idCard) < 8 {
        return "****"
    }
    return idCard[:3] + strings.Repeat("*", len(idCard)-7) + idCard[len(idCard)-4:]
}

// MaskToken Token脱敏
func MaskToken(token string) string {
    if len(token) <= 8 {
        return "****"
    }
    return token[:4] + "****" + token[len(token)-4:]
}

// MaskBankCard 银行卡脱敏
func MaskBankCard(card string) string {
    if len(card) < 4 {
        return "****"
    }
    return strings.Repeat("*", len(card)-4) + card[len(card)-4:]
}
```

#### 3. 请求日志脱敏中间件

```go
// sensitiveFields 敏感字段列表（小写）
var sensitiveFields = map[string]bool{
    "password":      true,
    "pwd":           true,
    "secret":        true,
    "token":         true,
    "access_token":  true,
    "refresh_token": true,
    "api_key":       true,
    "apikey":        true,
    "id_card":       true,
    "idcard":        true,
    "card_number":   true,
    "bank_card":     true,
    "cvv":           true,
    "pin":           true,
}

// MaskSensitiveData 递归脱敏敏感字段
func MaskSensitiveData(data interface{}) interface{} {
    switch v := data.(type) {
    case map[string]interface{}:
        for key, value := range v {
            if sensitiveFields[strings.ToLower(key)] {
                v[key] = "***MASKED***"
            } else {
                v[key] = MaskSensitiveData(value)
            }
        }
        return v
    case []interface{}:
        for i, item := range v {
            v[i] = MaskSensitiveData(item)
        }
        return v
    default:
        return data
    }
}
```

---

### 补充问题统计

| 问题类型 | 高 | 中 | 低 | 总计 |
|---------|---|---|-----|-----|
| errors.Is 使用错误 | 1 | - | - | 1 |
| 密码日志泄露 | 1 | - | - | 1 |
| 手机号日志未脱敏 | 1 | - | - | 1 |
| 请求体未脱敏 | 1 | - | - | 1 |
| 错误判断逻辑不当 | - | 1 | - | 1 |
| 循环中错误被吞 | - | 1 | - | 1 |
| 字符串匹配判断错误 | - | 1 | - | 1 |
| 重复日志输出 | - | - | 1 | 1 |
| **补充合计** | **4** | **3** | **1** | **8** |
| **原报告合计** | **5** | **4** | **50+** | **59+** |
| **总计** | **9** | **7** | **51+** | **67+** |

---

### 补充查找命令

```bash
# 查找 errors.Is 与 errors.New 混用
grep -rn "errors.Is.*errors.New" main/app/

# 查找日志中的手机号字段
grep -rn 'zap\.\(String\|Any\).*phone' main/app/ | grep -v "_test.go"

# 查找日志中的密码字段
grep -rn 'zap\.\(String\|Any\).*[Pp]assword' main/app/ | grep -v "_test.go"

# 查找字符串匹配判断错误
grep -rn 'strings.Contains.*err.Error()' main/app/

# 查找循环中的 continue（需人工判断）
grep -B3 'continue$' main/app/ | grep -A1 "err"
```
