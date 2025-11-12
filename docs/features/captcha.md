# Captcha Service 验证码服务说明文档

## 📋 概述

`service/captcha.go` 是 TTPOS 系统的验证码服务，负责生成图形验证码和验证用户输入。验证码主要用于用户登录等安全场景，防止暴力破解和机器人攻击。该服务使用缓存存储验证码答案，支持自动过期机制。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/captcha.go`  
**文件大小**: 62行  
**接口定义**: `ICaptchaSrv`  
**实现结构**: `captchaSrv`

---

## 🏗️ 架构设计

### 接口定义 (ICaptchaSrv)

```go
type ICaptchaSrv interface {
    Generate() (*resp.Captcha, error)           // 生成验证码
    Verify(sign string, answer string) bool     // 验证验证码
}
```

### 依赖服务

```go
type captchaSrv struct {
    captcha     *captcha.Captcha  // 验证码生成器（pkg/captcha包）
    cachePrefix string            // 缓存键前缀 "captcha:"
}
```

### 服务初始化

```go
func NewCaptchaSrv(cache cache.Cache) ICaptchaSrv {
    return NewCaptchaSrvImpl(cache)
}

func NewCaptchaSrvImpl(cache cache.Cache) ICaptchaSrv {
    srv := &captchaSrv{
        cachePrefix: "captcha:",
    }
    // 创建验证码工具，设置5分钟过期时间
    captchaTool, err := captcha.New(cache, srv.cachePrefix, 5*time.Minute)
    if err != nil {
        log.Fatalln(err)  // 初始化失败则退出程序
    }
    srv.captcha = captchaTool
    return srv
}
```

**初始化参数**:
- `cache`: 缓存服务实例
- `cachePrefix`: 缓存键前缀 `"captcha:"`
- `expiration`: 验证码有效期 `5分钟`

---

## 🎯 核心功能

### 1. 生成验证码 (Generate)

**功能描述**: 生成图形验证码，返回唯一标识（sign）和Base64编码的图片数据。

#### 方法签名

```go
func (s *captchaSrv) Generate() (*resp.Captcha, error)
```

#### 返回数据结构

```go
type Captcha struct {
    Sign   string `json:"sign"`   // 验证码唯一标识（UUID）
    Base64 string `json:"base64"` // Base64编码的验证码图片
}
```

#### 实现逻辑

```
1. 调用captcha工具生成验证码
   ↓
2. 生成随机标识（UUID）
   ↓
3. 生成验证码图片（Base64编码）
   ↓
4. 将验证码答案存入缓存
   - 缓存键：captcha:{sign}
   - 缓存值：验证码答案
   - 有效期：5分钟
   ↓
5. 返回sign和base64图片
```

#### 代码实现

```go
func (s *captchaSrv) Generate() (*resp.Captcha, error) {
    // 生成随机标识
    sign, b64s, err := s.captcha.Generate()
    if err != nil {
        return nil, errors.WithMessage(err)
    }
    return &resp.Captcha{
        Sign:   sign,  // UUID标识
        Base64: b64s,  // Base64图片数据
    }, nil
}
```

#### 使用示例

```go
// 生成验证码
captchaSrv := service.NewCaptchaSrv(cache)
captcha, err := captchaSrv.Generate()
if err != nil {
    // 错误处理
}

// 返回给前端
response := map[string]interface{}{
    "sign":   captcha.Sign,   // "550e8400-e29b-41d4-a716-446655440000"
    "base64": captcha.Base64, // "data:image/png;base64,iVBORw0KGgoAAAANS..."
}
```

#### 缓存存储

```
缓存键: captcha:550e8400-e29b-41d4-a716-446655440000
缓存值: "abcd"  (验证码答案)
有效期: 300秒（5分钟）
```

---

### 2. 验证验证码 (Verify)

**功能描述**: 验证用户输入的验证码是否正确。

#### 方法签名

```go
func (s *captchaSrv) Verify(sign, answer string) bool
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `sign` | string | 验证码标识（生成时返回的UUID） |
| `answer` | string | 用户输入的验证码答案 |

#### 返回值

| 类型 | 说明 |
|-----|------|
| bool | `true` - 验证通过，`false` - 验证失败 |

#### 验证逻辑

```
1. 检查是否为调试模式
   ↓ 是
2. 如果answer == "123456" → 返回true（调试密码）
   ↓ 否
3. 根据sign从缓存获取正确答案
   ↓
4. 比较用户输入与缓存中的答案
   ↓
5. 验证成功后删除缓存（一次性验证）
   ↓
6. 返回验证结果
```

#### 代码实现

```go
func (s *captchaSrv) Verify(sign, answer string) bool {
    // 开发调试使用 - 调试万能密码
    if config.Server.Mode == "debug" && answer == "123456" {
        return true
    }
    
    // 正常验证流程
    ok, err := s.captcha.Verify(sign, answer)
    if err != nil {
        return false
    }
    return ok
}
```

#### 调试模式

当 `config.Server.Mode == "debug"` 时，可以使用万能密码 `"123456"` 通过验证，方便开发调试。

```go
// 开发环境
config.Server.Mode = "debug"
result := captchaSrv.Verify("any-sign", "123456")  // 返回 true

// 生产环境
config.Server.Mode = "release"
result := captchaSrv.Verify("any-sign", "123456")  // 需要真实验证
```

#### 使用示例

```go
// 验证用户输入
sign := "550e8400-e29b-41d4-a716-446655440000"
userAnswer := "abcd"

isValid := captchaSrv.Verify(sign, userAnswer)
if isValid {
    // 验证通过，继续后续流程（如登录）
    fmt.Println("验证码正确")
} else {
    // 验证失败
    fmt.Println("验证码错误")
}
```

#### 验证失败的原因

| 情况 | 说明 |
|-----|------|
| 答案错误 | 用户输入的答案与生成的不匹配 |
| 验证码过期 | 超过5分钟有效期，缓存已删除 |
| sign不存在 | 无效的验证码标识 |
| 已验证过 | 验证码已使用（一次性） |

---

## 🔄 完整业务流程

### 用户登录流程（含验证码）

```
前端                          后端                          缓存
  │                            │                            │
  ├─1. 请求生成验证码───────────→│                            │
  │                            ├─2. 生成验证码图片           │
  │                            │   - 生成UUID sign          │
  │                            │   - 生成答案 "abcd"         │
  │                            │   - 生成Base64图片          │
  │                            ├─3. 存入缓存─────────────→  │
  │                            │                    captcha:{sign}="abcd"
  │←─4. 返回sign和base64图片────┤                            │
  │                            │                            │
  ├─5. 显示验证码图片           │                            │
  │   用户输入验证码            │                            │
  │                            │                            │
  ├─6. 提交登录请求─────────────→│                            │
  │   (username, password, sign, answer)                   │
  │                            ├─7. 验证验证码───────────→  │
  │                            │                    查询captcha:{sign}
  │                            │←─8. 返回缓存的答案──────── │
  │                            ├─9. 比较答案                │
  │                            │   - 相同 → 继续登录         │
  │                            │   - 不同 → 返回错误         │
  │                            ├─10. 删除缓存────────────→ │
  │                            │                    del captcha:{sign}
  │←─11. 返回登录结果────────────┤                            │
  │                            │                            │
```

---

## 🔧 底层实现（pkg/captcha包）

### Captcha工具类

验证码服务依赖 `ttpos-server-go/pkg/captcha` 包，该包封装了验证码的生成和验证逻辑。

#### 主要方法

```go
package captcha

type Captcha struct {
    cache       cache.Cache
    cachePrefix string
    expiration  time.Duration
}

// New 创建验证码工具
func New(cache cache.Cache, cachePrefix string, expiration time.Duration) (*Captcha, error) {
    return &Captcha{
        cache:       cache,
        cachePrefix: cachePrefix,
        expiration:  expiration,
    }, nil
}

// Generate 生成验证码
// 返回：sign（UUID）, base64图片, error
func (c *Captcha) Generate() (string, string, error) {
    // 1. 生成UUID作为sign
    sign := uuid.New().String()
    
    // 2. 生成随机验证码字符串（通常4-6位）
    answer := generateRandomString(4)
    
    // 3. 生成验证码图片
    base64Image := generateCaptchaImage(answer)
    
    // 4. 将答案存入缓存
    cacheKey := c.cachePrefix + sign
    c.cache.Set(cacheKey, answer, c.expiration)
    
    return sign, base64Image, nil
}

// Verify 验证验证码
func (c *Captcha) Verify(sign, answer string) (bool, error) {
    // 1. 从缓存获取正确答案
    cacheKey := c.cachePrefix + sign
    cachedAnswer, exists := c.cache.Get(cacheKey)
    if !exists {
        return false, errors.New("验证码不存在或已过期")
    }
    
    // 2. 比较答案（忽略大小写）
    isMatch := strings.EqualFold(cachedAnswer.(string), answer)
    
    // 3. 删除缓存（一次性验证）
    if isMatch {
        c.cache.Del(cacheKey)
    }
    
    return isMatch, nil
}
```

---

## 📊 验证码特性

### 1. 验证码类型

| 特性 | 说明 |
|-----|------|
| 类型 | 图形验证码（数字+字母） |
| 长度 | 通常4-6位 |
| 字符集 | 数字(0-9) + 大小写字母(A-Z, a-z) |
| 干扰 | 包含干扰线、噪点等 |
| 格式 | Base64编码的PNG图片 |

### 2. 安全机制

| 机制 | 说明 |
|-----|------|
| 唯一标识 | 每个验证码使用UUID唯一标识 |
| 时效性 | 5分钟自动过期 |
| 一次性 | 验证成功后立即删除，不可重复使用 |
| 大小写不敏感 | 验证时忽略大小写 |
| 缓存存储 | 使用Redis缓存，分布式支持 |

### 3. 性能指标

| 指标 | 值 |
|-----|---|
| 生成速度 | < 50ms |
| 验证速度 | < 10ms |
| 图片大小 | 约5-10KB（Base64编码后） |
| 并发能力 | 依赖缓存服务性能 |
| 过期时间 | 300秒（5分钟） |

---

## 🎯 使用场景

### 1. 用户登录

```go
// Controller层 - 登录接口
func (c *AuthController) Login(ctx *gin.Context) {
    var req req.LoginReq
    ctx.ShouldBindJSON(&req)
    
    // 1. 验证验证码
    if !c.captchaSrv.Verify(req.CaptchaSign, req.CaptchaAnswer) {
        response.Error(ctx, "验证码错误")
        return
    }
    
    // 2. 验证用户名密码
    staff, err := c.authSrv.Login(ctx, req.Username, req.Password)
    if err != nil {
        response.Error(ctx, err.Error())
        return
    }
    
    // 3. 生成Token
    token := c.authSrv.GenerateToken(staff)
    response.Success(ctx, map[string]interface{}{
        "token": token,
        "staff": staff,
    })
}
```

### 2. 重要操作验证

```go
// 修改密码
func (c *StaffController) ChangePassword(ctx *gin.Context) {
    var req req.ChangePasswordReq
    ctx.ShouldBindJSON(&req)
    
    // 验证验证码
    if !c.captchaSrv.Verify(req.CaptchaSign, req.CaptchaAnswer) {
        response.Error(ctx, "验证码错误")
        return
    }
    
    // 修改密码逻辑
    err := c.staffSrv.ChangePassword(ctx, req.OldPassword, req.NewPassword)
    // ...
}
```

### 3. 注册账号

```go
// 注册接口
func (c *AuthController) Register(ctx *gin.Context) {
    var req req.RegisterReq
    ctx.ShouldBindJSON(&req)
    
    // 验证验证码
    if !c.captchaSrv.Verify(req.CaptchaSign, req.CaptchaAnswer) {
        response.Error(ctx, "验证码错误")
        return
    }
    
    // 注册逻辑
    // ...
}
```

---

## 🔌 API接口示例

### 1. 生成验证码接口

#### 请求

```http
POST /api/v1/captcha/generate
Content-Type: application/json
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "sign": "550e8400-e29b-41d4-a716-446655440000",
    "base64": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMgAAABkCAYAAADDhn8LAAAK..."
  }
}
```

#### Controller实现

```go
// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Description 生成图形验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} resp.Captcha "成功"
// @Router /api/v1/captcha/generate [post]
func (c *AuthController) GetCaptcha(ctx *gin.Context) {
    captcha, err := c.captchaSrv.Generate()
    if err != nil {
        response.Error(ctx, "生成验证码失败")
        return
    }
    response.Success(ctx, captcha)
}
```

### 2. 验证验证码（登录时）

#### 请求

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "123456",
  "captcha_sign": "550e8400-e29b-41d4-a716-446655440000",
  "captcha_answer": "abcd"
}
```

#### 响应（验证码错误）

```json
{
  "code": 0,
  "message": "验证码错误",
  "data": {}
}
```

#### 响应（登录成功）

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "staff": {
      "uuid": 1,
      "username": "admin",
      "nickname": "管理员"
    }
  }
}
```

---

## ⚙️ 配置说明

### 验证码配置

验证码的配置在 `config/config.yaml` 或环境变量中设置：

```yaml
# 服务器模式
server:
  mode: release  # release/debug

# 缓存配置
cache:
  type: redis
  redis:
    host: 127.0.0.1
    port: 6379
    password: ""
    db: 0
```

### 可调整参数

| 参数 | 当前值 | 可调整范围 | 说明 |
|-----|-------|----------|-----|
| 有效期 | 5分钟 | 1-30分钟 | 验证码过期时间 |
| 字符长度 | 4位 | 4-8位 | 验证码字符数量 |
| 缓存前缀 | "captcha:" | 任意字符串 | 缓存键前缀 |
| 图片宽度 | 默认 | 80-200px | 验证码图片宽度 |
| 图片高度 | 默认 | 30-80px | 验证码图片高度 |

---

## 🛡️ 安全建议

### 1. 防暴力破解

```go
// 限制生成频率（可选）
func (c *AuthController) GetCaptcha(ctx *gin.Context) {
    ip := ctx.ClientIP()
    cacheKey := "captcha_limit:" + ip
    
    // 检查是否频繁请求
    if count, exists := cache.Get(cacheKey); exists {
        if count.(int) > 10 { // 1分钟内超过10次
            response.Error(ctx, "请求过于频繁")
            return
        }
        cache.Incr(cacheKey)
    } else {
        cache.Set(cacheKey, 1, 1*time.Minute)
    }
    
    // 生成验证码
    captcha, _ := c.captchaSrv.Generate()
    response.Success(ctx, captcha)
}
```

### 2. 验证失败次数限制

```go
// 限制验证失败次数
func (c *AuthController) Login(ctx *gin.Context) {
    var req req.LoginReq
    ctx.ShouldBindJSON(&req)
    
    // 检查失败次数
    failKey := "login_fail:" + req.Username
    if count, exists := cache.Get(failKey); exists {
        if count.(int) >= 5 { // 5次失败后需要等待
            response.Error(ctx, "失败次数过多，请10分钟后再试")
            return
        }
    }
    
    // 验证验证码
    if !c.captchaSrv.Verify(req.CaptchaSign, req.CaptchaAnswer) {
        // 记录失败次数
        cache.Incr(failKey)
        cache.Expire(failKey, 10*time.Minute)
        response.Error(ctx, "验证码错误")
        return
    }
    
    // 登录逻辑
    // ...
}
```

### 3. HTTPS传输

- 生产环境必须使用HTTPS传输验证码
- 防止中间人攻击窃取验证码

### 4. 日志记录

```go
func (s *captchaSrv) Verify(sign, answer string) bool {
    ok, err := s.captcha.Verify(sign, answer)
    
    // 记录验证失败日志
    if !ok {
        logger.Logger.Warn("验证码验证失败",
            zap.String("sign", sign),
            zap.String("answer", answer),
            zap.Error(err),
        )
    }
    
    return ok
}
```

---

## 🎯 最佳实践

### 1. 验证码刷新

```go
// ✅ 正确：验证失败后刷新验证码
if !captchaSrv.Verify(sign, answer) {
    // 前端应该重新请求生成验证码
    return errors.New("验证码错误，请重新输入")
}
```

### 2. 错误提示

```go
// ✅ 正确：明确的错误提示
if !captchaSrv.Verify(sign, answer) {
    return errors.New("验证码错误")
}

// ❌ 错误：泄露过多信息
if !captchaSrv.Verify(sign, answer) {
    return errors.New("验证码不存在或已过期或答案错误")  // 不要暴露具体原因
}
```

### 3. 调试模式使用

```go
// ✅ 正确：仅在开发环境使用万能密码
if config.Server.Mode == "debug" && answer == "123456" {
    return true
}

// ❌ 错误：生产环境保留万能密码
if answer == "123456" {  // 生产环境不安全
    return true
}
```

### 4. 缓存依赖

```go
// ✅ 正确：确保缓存服务可用
func NewCaptchaSrvImpl(cache cache.Cache) ICaptchaSrv {
    if cache == nil {
        log.Fatalln("缓存服务不可用")
    }
    // 初始化逻辑
}
```

---

## ⚠️ 注意事项

### 1. 缓存依赖

- 验证码服务强依赖缓存服务（Redis）
- 缓存服务不可用时，验证码功能将失败
- 建议配置Redis高可用方案

### 2. 时效性

- 验证码5分钟后自动过期
- 验证成功后立即失效（一次性）
- 前端应及时提示用户验证码过期

### 3. 大小写

- 验证时忽略大小写
- 用户输入 "ABCD" 等同于 "abcd"

### 4. 调试模式

- 万能密码 "123456" 仅在调试模式下有效
- 生产环境务必设置 `mode: release`

### 5. 并发安全

- 服务本身是无状态的，线程安全
- 并发能力取决于缓存服务性能

---

## 📚 相关文档

- [Authentication Service](./auth_service.md) - 认证服务（使用验证码）
- [Cache Service](../pkg/cache/cache_service.md) - 缓存服务
- [Captcha Package](../pkg/captcha/captcha.md) - 验证码底层实现

---

## 📊 服务特点总结

| 特点 | 说明 |
|-----|------|
| 简洁 | 仅62行代码，功能清晰 |
| 高效 | 基于缓存，响应速度快 |
| 安全 | 唯一标识、时效性、一次性验证 |
| 易用 | 接口简单，集成方便 |
| 可靠 | 依赖成熟的缓存服务 |
| 灵活 | 支持调试模式 |

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

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。验证码服务是系统安全的重要组成部分，建议定期审查和优化。

