# 安全开发详细指南

> 👤 **受众**: 人类开发者  
> 📖 **用途**: 安全开发的详细指南和最佳实践

---

## 认证授权

### JWT 认证

**生成 Token (Go)**:
```go
func GenerateToken(userId uint64) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userId,
        "exp":     time.Now().Add(time.Hour * 2).Unix(),
        "iat":     time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.JWTSecret))
}
```

**验证 Token**:
```go
func VerifyToken(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(config.JWTSecret), nil
    })
}
```

**❌ 不要在 Token 中存储敏感信息**:
```go
// ❌ 错误
claims := jwt.MapClaims{
    "password": password,  // 不要存储密码
    "api_key":  apiKey,    // 不要存储密钥
}
```

---

### 权限验证（RBAC）

```go
func CheckPermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userId := c.GetUint64("user_id")
        
        hasPermission, err := permissionService.CheckUserPermission(userId, permission)
        if err != nil || !hasPermission {
            c.JSON(403, gin.H{
                "code":    403,
                "message": "无权限访问",
                "data":    gin.H{},
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## 数据验证

### 输入验证

**Go - 使用验证标签**:
```go
type CreateUserReq struct {
    Username string `json:"username" binding:"required,min=2,max=20,alphanum"`
    Email    string `json:"email" binding:"required,email"`
    Phone    string `json:"phone" binding:"omitempty,len=11,numeric"`
    Age      int    `json:"age" binding:"required,min=1,max=150"`
}

// 自定义验证
func (req *CreateUserReq) Validate() error {
    if strings.ContainsAny(req.Username, "!@#$%^&*()") {
        return errors.New("用户名不能包含特殊字符")
    }
    
    if req.Phone != "" {
        matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, req.Phone)
        if !matched {
            return errors.New("手机号格式不正确")
        }
    }
    
    return nil
}
```

**PHP - 使用验证器**:
```php
protected $rule = [
    'username' => 'require|length:2,20|alphaNum',
    'email'    => 'require|email',
    'phone'    => 'mobile',
    'age'      => 'require|number|between:1,150',
];
```

### 白名单验证

```go
// ✅ 使用白名单
var allowedSortFields = map[string]bool{
    "id":          true,
    "create_time": true,
    "username":    true,
}

func ValidateSortField(field string) error {
    if !allowedSortFields[field] {
        return errors.New("不支持的排序字段")
    }
    return nil
}
```

---

## SQL 注入防护

### 使用参数化查询

**Go - GORM**:
```go
// ✅ 正确：参数化查询
db.Where("username = ?", username).First(&user)
db.Where("id IN ?", ids).Find(&users)

// ❌ 危险：字符串拼接
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
db.Raw(query).Scan(&users)  // SQL 注入风险！
```

**PHP - ThinkPHP**:
```php
// ✅ 正确
User::where('username', $username)->find();

// ❌ 危险
$sql = "SELECT * FROM users WHERE username = '$username'";
Db::query($sql);
```

---

## XSS 防护

### 输出转义

**Go**:
```go
import "html"

// 转义 HTML
safeContent := html.EscapeString(userInput)
```

**PHP**:
```php
// 转义输出
echo htmlspecialchars($userInput, ENT_QUOTES, 'UTF-8');

// Blade 模板自动转义
{{ $content }}         // 自动转义
{!! $content !!}       // 不转义（谨慎使用）
```

**Vue**:
```vue
<!-- 自动转义 -->
<div>{{ userInput }}</div>

<!-- 不转义（谨慎使用） -->
<div v-html="trustedContent"></div>
```

### Content Security Policy

```go
c.Header("Content-Security-Policy", 
    "default-src 'self'; " +
    "script-src 'self' 'unsafe-inline'; " +
    "style-src 'self' 'unsafe-inline';")
```

---

## CSRF 防护

### CSRF Token

```go
// 生成 CSRF Token
func GenerateCSRFToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

// 验证 CSRF Token
func VerifyCSRFToken(token, sessionToken string) bool {
    return token == sessionToken
}
```

**中间件**:
```go
func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method != "GET" {
            token := c.GetHeader("X-CSRF-Token")
            sessionToken := getSessionToken(c)
            
            if !VerifyCSRFToken(token, sessionToken) {
                c.JSON(403, gin.H{"message": "CSRF token invalid"})
                c.Abort()
                return
            }
        }
        c.Next()
    }
}
```

---

## 密码安全

### 密码加密

**Go**:
```go
import "golang.org/x/crypto/bcrypt"

// 加密密码
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}

// 验证密码
func VerifyPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
```

**PHP**:
```php
// 加密密码
$hashedPassword = password_hash($password, PASSWORD_DEFAULT);

// 验证密码
$isValid = password_verify($password, $hashedPassword);
```

### 密码策略

```go
func ValidatePassword(password string) error {
    if len(password) < 8 {
        return errors.New("密码至少8个字符")
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    
    if !hasUpper || !hasLower || !hasNumber {
        return errors.New("密码必须包含大小写字母和数字")
    }
    
    return nil
}
```

---

## 敏感数据处理

### 数据脱敏

```go
// 手机号脱敏
func MaskPhone(phone string) string {
    if len(phone) != 11 {
        return phone
    }
    return phone[:3] + "****" + phone[7:]
}

// 邮箱脱敏
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return email
    }
    
    name := parts[0]
    if len(name) > 2 {
        name = name[:1] + "***" + name[len(name)-1:]
    }
    
    return name + "@" + parts[1]
}
```

### 日志脱敏

```go
// ❌ 不要记录敏感信息
logger.Info("User login", "password", password)        // 危险！
logger.Info("Payment", "card_number", cardNumber)      // 危险！

// ✅ 记录脱敏后的信息
logger.Info("User login", "username", username)
logger.Info("Payment", "masked_card", MaskCard(cardNumber))
```

---

## API 安全

### 限流

```go
import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(10, 100)  // 每秒10个请求，桶容量100

func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"message": "Too many requests"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### HTTPS

```go
// 强制 HTTPS
func HTTPSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Header.Get("X-Forwarded-Proto") != "https" {
            c.Redirect(301, "https://"+c.Request.Host+c.Request.RequestURI)
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### CORS

```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

---

## 安全检查清单

### 开发阶段

- [ ] 所有输入都进行了验证
- [ ] 使用参数化查询防止 SQL 注入
- [ ] 输出进行了转义防止 XSS
- [ ] 实现了 CSRF 防护
- [ ] 密码使用 bcrypt 加密
- [ ] 敏感数据不记录到日志
- [ ] API 实现了限流
- [ ] 使用 HTTPS
- [ ] 实现了权限验证

### 部署阶段

- [ ] 关闭调试模式
- [ ] 配置文件不包含敏感信息
- [ ] 数据库连接使用最小权限
- [ ] 定期更新依赖包
- [ ] 配置防火墙规则
- [ ] 设置日志监控告警

---

## 相关文档

- [Go Main 开发指南](./go-main-development.md) - Go 安全实践
- [PHP 开发指南](./php-development.md) - PHP 安全实践
- [API 设计指南](./api-design-guide.md) - API 安全设计

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

