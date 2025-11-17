# API 设计详细指南

> 👤 **受众**: 人类开发者  
> 📖 **用途**: RESTful API 设计的详细指南和最佳实践

---

## RESTful API 设计

### 资源命名原则

**使用名词单数 + 操作后缀**

```
✅ 正确：
GET    /api/v1/cashier/order/list          # 获取订单列表
GET    /api/v1/cashier/order/info          # 获取订单详情
POST   /api/v1/cashier/order/cancel        # 取消订单
POST   /api/v1/cashier/order/return        # 退款
DELETE /api/v1/cashier/order/delete        # 删除订单

❌ 错误（RESTful 风格）：
GET    /api/v1/cashier/orders               # 不使用复数
GET    /api/v1/cashier/orders/123           # 不使用ID路径
POST   /api/v1/cashier/orders               # 不使用HTTP方法表示动作
```

**为什么不用标准 RESTful?**
- 前端习惯使用路径表示操作
- 更直观易懂
- 避免 HTTP 方法的语义混淆

---

## URL 命名规范

### 使用 snake_case

```
✅ 正确：
/api/v1/user_profiles
/api/v1/order_items
/api/v1/payment_methods

❌ 错误：
/api/v1/userProfiles      # camelCase
/api/v1/user-profiles     # kebab-case
```

### 版本控制

```
/api/v1/users             # 版本1
/api/v2/users             # 版本2
```

### 查询参数

```
GET /api/v1/cashier/order/list?page=1&page_size=20&sort_by=create_time&order=desc
```

---

## HTTP 方法

### 标准方法映射

| 方法 | 用途 | 示例 |
|------|------|------|
| GET | 获取资源 | `GET /api/v1/cashier/order/list` |
| POST | 创建或执行操作 | `POST /api/v1/cashier/order/cancel` |
| PUT | 完整更新 | `PUT /api/v1/shop/product/edit` |
| PATCH | 部分更新 | `PATCH /api/v1/shop/product/edit` |
| DELETE | 删除 | `DELETE /api/v1/cashier/order/delete` |

### 幂等性

- GET、PUT、DELETE 应该是幂等的
- POST 不是幂等的
- 多次执行相同的 PUT/DELETE，结果应该相同

---

## 请求格式

### Content-Type

```http
Content-Type: application/json
```

### 请求体示例

```json
{
  "username": "test_user",
  "email": "test@example.com",
  "password": "123456",
  "phone": "13800138000"
}
```

### 必需的请求头

```http
Content-Type: application/json
Authorization: Bearer <token>
Accept-Language: zh-CN
```

---

## 响应格式

### 统一响应结构

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {"id": 1, "name": "商品1"},
      {"id": 2, "name": "商品2"}
    ],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

**核心规则**:
- `code`: 1 表示成功，0 表示失败
- `message`: 提示信息
- `data`: **必须是对象，不能是 null 或数组**

### data 字段规则

```json
// ✅ 正确
{"code": 1, "message": "success", "data": {}}
{"code": 1, "message": "success", "data": {"list": []}}

// ❌ 错误
{"code": 1, "message": "success", "data": null}
{"code": 1, "message": "success", "data": []}
```

---

## 分页规范

### 分页信息放在 meta 中

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [...],
    "meta": {
      "page_no": 1,       // 当前页码
      "page_size": 20,    // 每页大小
      "total": 100        // 总记录数
    }
  }
}
```

### Go 响应结构

```go
type OrderListPaginationResp struct {
    List []OrderItem    `json:"list"`
    Meta OrderListMeta  `json:"meta"`
}

type OrderListMeta struct {
    dto.PageResponse
    TotalNum    int64 `json:"total_num"`
    UnpaidNum   int64 `json:"unpaid_num"`
}

type PageResponse struct {
    PageNo   int   `json:"page_no"`
    PageSize int   `json:"page_size"`
    Total    int64 `json:"total"`
}
```

---

## 错误处理

### 错误码定义

```go
const (
    CodeSuccess         = 1   // 成功
    CodeFail            = 0   // 失败
    CodeInvalidParam    = 400 // 参数错误
    CodeUnauthorized    = 401 // 未认证
    CodeForbidden       = 403 // 无权限
    CodeNotFound        = 404 // 资源不存在
    CodeServerError     = 500 // 服务器错误
)
```

### 错误响应示例

```json
{
  "code": 0,
  "message": "用户名已存在",
  "data": {}
}

{
  "code": 400,
  "message": "参数验证失败：邮箱格式不正确",
  "data": {}
}

{
  "code": 401,
  "message": "未登录或登录已过期",
  "data": {}
}
```

---

## 认证授权

### JWT Token

**请求头**:
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Token 结构**:
```json
{
  "user_id": 123,
  "exp": 1699999999,
  "iat": 1699900000
}
```

### 权限验证

```go
// 在中间件中验证权限
func CheckPermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !hasPermission(c, permission) {
            helper.Error(c, 403, "无权限访问")
            c.Abort()
            return
        }
        c.Next()
    }
}

// 使用
router.DELETE("/api/v1/order/delete", CheckPermission("order.delete"), handler)
```

---

## Swagger 文档

### Go (swaggo)

```go
// CreateOrder 创建订单
// @Summary 创建订单
// @Description 创建一个新的订单
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param request body req.CreateOrderReq true "创建订单请求"
// @Success 200 {object} resp.CreateOrderResp "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Security JwtToken
// @Router /api/v1/order/create [post]
func (c *Controller) CreateOrder(ctx *gin.Context) {
    // 实现
}
```

### PHP (ApiDoc)

```php
/**
 * @Apidoc\Title("创建订单")
 * @Apidoc\Desc("创建一个新的订单")
 * @Apidoc\Method("POST")
 * @Apidoc\Param("product_id", type="int", require=true, desc="商品ID")
 * @Apidoc\Param("quantity", type="int", require=true, desc="数量")
 * @Apidoc\Returned("order_id", type="int", desc="订单ID")
 */
public function createOrder(Request $request) {
    // 实现
}
```

---

## 数据验证

### Go 验证

```go
type CreateOrderReq struct {
    ProductId uint64  `json:"product_id" binding:"required,gt=0"`
    Quantity  int     `json:"quantity" binding:"required,min=1,max=999"`
    Price     float64 `json:"price" binding:"required,gt=0"`
    Remark    string  `json:"remark" binding:"omitempty,max=200"`
}
```

### PHP 验证

```php
protected $rule = [
    'product_id' => 'require|number|gt:0',
    'quantity'   => 'require|number|between:1,999',
    'price'      => 'require|float|gt:0',
    'remark'     => 'max:200',
];
```

---

## 最佳实践

### 1. 使用 Helper 统一返回

```go
// ✅ 使用 helper
helper.Success(c, data)
helper.ErrorWithDetail(c, constant.CodeFail, err)

// ❌ 直接返回
c.JSON(200, gin.H{"code": 1, "data": data})
```

### 2. 错误信息国际化

```go
message := i18n.Translate(ctx.GetLanguage(), "订单创建成功")
helper.Success(c, gin.H{"message": message})
```

### 3. 参数验证

```go
// ✅ 使用验证标签
type Req struct {
    Username string `binding:"required,min=2,max=20"`
}

// ✅ 自定义验证
func (req *Req) Validate() error {
    if req.Username == "" {
        return errors.New("用户名不能为空")
    }
    return nil
}
```

### 4. 响应时间监控

```go
start := time.Now()
defer func() {
    duration := time.Since(start)
    if duration > 200*time.Millisecond {
        logger.Warn("API响应时间过长", duration)
    }
}()
```

---

## 相关文档

- [Go Main 开发指南](./go-main-development.md) - Go API 实现
- [PHP 开发指南](./php-development.md) - PHP API 实现
- [安全开发指南](./security-guide.md) - API 安全规范
- [docs/shared/api/conventions.md](../../shared/api/conventions.md) - API 规范速查

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

