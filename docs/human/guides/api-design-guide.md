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

| 方法 | 用途 | 示例 | 说明 |
|------|------|------|------|
| GET | 获取资源 | `GET /api/v1/cashier/order/list` | 查询、列表、详情 |
| POST | 创建或执行操作 | `POST /api/v1/cashier/order/cancel` | 创建、更新、执行操作 |
| DELETE | 删除数据 | `DELETE /api/v1/cashier/order/delete` | **所有删除操作必须使用 DELETE** |
| PUT | **禁止使用** | - | **项目禁止使用 PUT 方法** |

### DELETE 方法 - 删除接口规范

**核心规则：所有删除操作必须使用 DELETE HTTP 方法**

#### 路由注册示例

```go
// ✅ 正确：删除角色使用 DELETE 方法
privateApi.DELETE("/shop_role/delete", handler.Delete)

// ✅ 正确：删除其他资源也使用 DELETE 方法
privateApi.DELETE("/order/delete", handler.Delete)
privateApi.DELETE("/product/delete", handler.Delete)

// ❌ 错误：删除操作不能使用 POST
privateApi.POST("/shop_role/delete", handler.Delete)  // ❌ 禁止
```

#### DELETE 请求处理示例

```go
// ✅ 正确：DELETE 请求用于删除数据
func (h *ShopRoleHandler) Delete(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    // DELETE 请求参数可以通过 Query 或 Body 传递
    // 方式1：通过 Query 传递（推荐用于简单删除）
    uuidStr := c.Query("uuid")
    uuid, err := strconv.ParseUint(uuidStr, 10, 64)
    if err != nil || uuid == 0 {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.New("uuid 参数错误"))
        return
    }
    
    // 方式2：通过 Body 传递（适用于批量删除等复杂场景）
    // var deleteReq req.ShopRoleDeleteReq
    // if err := c.ShouldBindJSON(&deleteReq); err != nil {
    //     helper.HandleValidationError(c, err, deleteReq, nil)
    //     return
    // }
    
    if err := h.shopRoleSrv.Delete(ctx, uuid); err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, nil, "删除成功")
}
```

#### 批量删除示例

```go
// ✅ 正确：批量删除也使用 DELETE 方法
func (h *ShopRoleHandler) BatchDelete(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    var deleteReq req.ShopRoleBatchDeleteReq
    if err := c.ShouldBindJSON(&deleteReq); err != nil {
        helper.HandleValidationError(c, err, deleteReq, nil)
        return
    }
    
    if err := h.shopRoleSrv.BatchDelete(ctx, deleteReq.Uuids); err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, nil, "批量删除成功")
}

// 路由注册
privateApi.DELETE("/shop_role/batch_delete", handler.BatchDelete)
```

#### 规则说明

- ✅ **删除接口必须使用 DELETE 方法**（如删除角色、删除订单、删除商品等）
- ❌ **禁止使用 POST 方法进行删除操作**
- DELETE 请求参数可以通过 Query 或 Body 传递（根据实际情况选择）
  - 简单删除（单个 ID）：推荐使用 Query 参数
  - 复杂删除（批量、条件删除）：使用 Body 传递 JSON

### 幂等性

- GET、DELETE 应该是幂等的
- POST 不是幂等的
- 多次执行相同的 DELETE，结果应该相同（删除不存在的资源也应返回成功）

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

### 切片初始化规范

**响应体中的切片必须使用 `make` 初始化，避免返回 null**

#### 问题说明

在 Go 中，未初始化的切片（nil slice）在 JSON 序列化时会变成 `null`，而不是空数组 `[]`。这会导致前端无法统一处理，可能引发错误。

#### 正确示例

```go
// ✅ 正确：使用 make 初始化切片
type ExampleResp struct {
    List     []Item `json:"list"`      // 切片字段
    Tags     []string `json:"tags"`     // 字符串切片
    Ids      []uint64 `json:"ids"`     // 数字切片
}

func GetExample() *ExampleResp {
    resp := &ExampleResp{
        List: make([]Item, 0),      // ✅ 初始化为空切片，JSON 序列化为 []
        Tags: make([]string, 0),    // ✅ 初始化为空切片，JSON 序列化为 []
        Ids:  make([]uint64, 0),    // ✅ 初始化为空切片，JSON 序列化为 []
    }
    return resp
}

// JSON 输出：
// {
//   "list": [],
//   "tags": [],
//   "ids": []
// }
```

#### 错误示例

```go
// ❌ 错误：未初始化的切片会序列化为 null
type ExampleResp struct {
    List []Item `json:"list"`
}

func GetExample() *ExampleResp {
    resp := &ExampleResp{
        List: nil,  // ❌ nil 切片，JSON 序列化为 null
    }
    return resp
}

// JSON 输出：
// {
//   "list": null  // ❌ 错误：应该是 []
// }
```

#### 常见场景

**1. 列表查询接口**

```go
// ✅ 正确
func (h *OrderHandler) GetList(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    result, err := h.orderSrv.GetList(ctx, &req.OrderListReq{})
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, err)
        return
    }
    
    // 确保 List 字段已初始化
    if result.List == nil {
        result.List = make([]OrderItem, 0)
    }
    
    helper.Success(c, result, "获取成功")
}
```

**2. 响应结构体定义**

```go
// ✅ 正确：在结构体定义时就初始化
type OrderListResp struct {
    List []OrderItem `json:"list"`
    Meta OrderMeta   `json:"meta"`
}

func NewOrderListResp() *OrderListResp {
    return &OrderListResp{
        List: make([]OrderItem, 0),  // ✅ 初始化
        Meta: OrderMeta{},
    }
}

// ❌ 错误：未初始化
func NewOrderListResp() *OrderListResp {
    return &OrderListResp{
        List: nil,  // ❌ 错误
    }
}
```

**3. 嵌套结构体中的切片**

```go
// ✅ 正确
type ProductDetailResp struct {
    Uuid       uint64              `json:"uuid"`
    Name       string              `json:"name"`
    Attributes []ProductAttribute  `json:"attributes"`  // 嵌套切片
    Images     []string            `json:"images"`     // 字符串切片
}

func GetProductDetail() *ProductDetailResp {
    return &ProductDetailResp{
        Uuid:       123,
        Name:       "商品名称",
        Attributes: make([]ProductAttribute, 0),  // ✅ 初始化
        Images:     make([]string, 0),            // ✅ 初始化
    }
}
```

**4. 条件赋值**

```go
// ✅ 正确：即使有条件判断，也要初始化
func GetOrderItems(orderId uint64) []OrderItem {
    items, err := orderRepo.GetItems(orderId)
    if err != nil {
        return make([]OrderItem, 0)  // ✅ 错误时返回空切片，不是 nil
    }
    
    if len(items) == 0 {
        return make([]OrderItem, 0)  // ✅ 空结果返回空切片
    }
    
    return items
}
```

#### 规则总结

- ✅ **必须使用 `make([]Type, 0)` 初始化所有响应体中的切片字段**
- ✅ **即使切片为空，也要初始化为空切片，不要使用 nil**
- ✅ **在结构体初始化时就进行切片初始化**
- ✅ **函数返回切片时，确保返回的是空切片而不是 nil**
- ❌ **禁止在响应体中使用 nil 切片**
- ❌ **禁止依赖 Go 的零值初始化（nil slice）**

### 业务数据返回规范

#### 销售订单购物车信息

**涉及销售订单（`ttpos_sale_order`）和销售账单（`ttpos_sale_bill`）的接口，必须在响应中包含购物车信息。**

**规则：**

- ✅ **必须返回购物车信息**：确保前端能够实时获取最新的购物车状态（如商品列表、金额信息、折扣信息等）。
- ✅ **使用 `resp.ShopCart` 结构**：标准化的购物车响应结构。
- ✅ **场景覆盖**：创建订单、更新订单、查询订单详情、结账检查等接口。

**代码参考：**

- **响应结构**：`main/app/dto/resp/shop_cart.go` 中的 `ShopCart` 结构体。
- **实现示例**：`main/app/api/v1/cashier/cashier_instant.go` 中的 `OrderCartInfo` 方法。

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

## 代码示例

### 路由注册

**每个 API 文件必须包含注册函数**

```go
// ✅ 正确：包含注册函数
func RegisterExampleHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
    // 1. 初始化所有依赖服务
    captchaSrv := service.NewCaptchaSrv(cache)
    settingSrv := setting.NewSrv(dbm, cache)
    authSrv := service.NewAuthSrv(...)
    
    // 2. 创建 Handler 实例
    handler := &ExampleHandler{
        exampleSrv: service.NewExampleSrv(dbm),
    }
    
    // 3. 注册路由
    // middleware.Auth(authSrv, dbm) (强制)
    privateApi := router.Group("", middleware.Auth(authSrv, dbm))
    {
        // ✅ GET：获取信息（列表、详情）
        privateApi.GET("/example/list", handler.GetList)
        privateApi.GET("/example/get", handler.GetByUuid)
        
        // ✅ POST：创建和修改数据
        privateApi.POST("/example/create", handler.Create)
        privateApi.POST("/example/update", handler.Update)
        
        // ✅ DELETE：删除数据
        privateApi.DELETE("/example/delete", handler.Delete)
        
        // ❌ 错误：禁止使用 PUT
        // privateApi.PUT("/example/update", handler.Update)
    }
}

// ❌ 错误：没有注册函数
// API 文件必须有 Register*Handlers 函数
```

**注册函数必须在 router.go 中调用**

```go
// router/router.go
shopGroup := apiV1.Group("/shop")
{
    shop.RegisterExampleHandlers(shopGroup, dbm, cache)  // ✅ 必须调用
}
```

### GET 请求处理示例

```go
// ✅ 正确：GET 请求使用 Query 参数
func (h *ExampleHandler) GetByUuid(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    // 从 Query 获取参数
    uuidStr := c.Query("uuid")
    uuid, err := strconv.ParseUint(uuidStr, 10, 64)
    if err != nil || uuid == 0 {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.New("uuid 参数错误"))
        return
    }
    
    result, err := h.exampleSrv.GetByUuid(ctx, uuid)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, result, "获取成功")
}

// ✅ 正确：GET 请求列表使用 Query 参数
func (h *ExampleHandler) GetList(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    // 从 Query 获取参数
    pageNoStr := c.DefaultQuery("page_no", "1")
    pageSizeStr := c.DefaultQuery("page_size", "20")
    status := c.Query("status")
    
    pageNo, _ := strconv.Atoi(pageNoStr)
    pageSize, _ := strconv.Atoi(pageSizeStr)
    
    listReq := &req.ExampleListReq{
        PageNo:   pageNo,
        PageSize: pageSize,
        Status:   status,
    }
    
    result, err := h.exampleSrv.GetList(ctx, listReq)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, result, "获取成功")
}

// ❌ 错误：GET 请求不应该使用 Body
func (h *ExampleHandler) GetByUuid(c *gin.Context) {
    var getReq req.ExampleGetReq
    if err := c.ShouldBindJSON(&getReq); err != nil {  // ❌ GET 不应该用 Body
        helper.HandleValidationError(c, err, getReq, nil)
        return
    }
    // ...
}
```

### POST 请求处理示例

```go
// ✅ 正确：POST 请求用于创建数据
func (h *ExampleHandler) Create(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    var createReq req.ExampleCreateReq
    if err := c.ShouldBindJSON(&createReq); err != nil {
        helper.HandleValidationError(c, err, createReq, nil)
        return
    }
    
    err := h.exampleSrv.Create(ctx, &createReq)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, nil, "创建成功")
}

// ✅ 正确：POST 请求用于更新数据
func (h *ExampleHandler) Update(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    var updateReq req.ExampleUpdateReq
    if err := c.ShouldBindJSON(&updateReq); err != nil {
        helper.HandleValidationError(c, err, updateReq, nil)
        return
    }
    
    if err := h.exampleSrv.Update(ctx, &updateReq); err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, nil, "更新成功")
}
```

### DELETE 请求处理示例

```go
// ✅ 正确：DELETE 请求用于删除数据
func (h *ExampleHandler) Delete(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    // DELETE 请求参数可以通过 Query 或 Body 传递
    // 方式1：通过 Query 传递（推荐用于简单删除）
    uuidStr := c.Query("uuid")
    uuid, err := strconv.ParseUint(uuidStr, 10, 64)
    if err != nil || uuid == 0 {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.New("uuid 参数错误"))
        return
    }
    
    // 方式2：通过 Body 传递（适用于批量删除等复杂场景）
    // var deleteReq req.ExampleDeleteReq
    // if err := c.ShouldBindJSON(&deleteReq); err != nil {
    //     helper.HandleValidationError(c, err, deleteReq, nil)
    //     return
    // }
    
    if err := h.exampleSrv.Delete(ctx, uuid); err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, nil, "删除成功")
}

// ❌ 错误：删除不应该使用 POST
func (h *ExampleHandler) Delete(c *gin.Context) {
    // ❌ 错误：删除应该使用 DELETE 方法，不应该用 POST
    // 应该在路由注册时使用 privateApi.DELETE(...)
}
```

### 响应处理示例

#### Go 返回响应

```go
// ✅ 正确
helper.Success(c, gin.H{})
helper.Success(c, gin.H{"list": []})
helper.ErrorWithDetail(c, constant.CodeFail, err)

// ❌ 错误
c.JSON(200, data)  // 不要直接用 c.JSON
```

#### PHP 返回响应

```php
// ✅ 正确
return $this->renderSuccess($data);
return $this->renderSuccess([]);
return $this->renderError('错误信息');
```

#### 分页响应格式

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [...],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

### 多语言字段示例

**核心规则：所有多语言字段必须使用 `dto.LocaleResponse` 类型，字段名必须使用 `LocaleName` 或 `LocaleXXXName` 格式**

#### 请求参数示例

```go
// ✅ 正确 - 请求参数
type ExampleReq struct {
    LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 名称（多语言）
}

// ❌ 错误 - 字段名缺少 Locale 前缀
type ExampleReq struct {
    Name dto.LocaleResponse `json:"name"` // ❌ 错误，应使用 LocaleName
}

// ❌ 错误 - 请求参数使用 map
type ExampleReq struct {
    Name map[string]string `json:"name"` // ❌ 错误，应使用 dto.LocaleResponse
}

// ❌ 错误 - 请求参数使用字符串
type ExampleReq struct {
    Name string `json:"name"` // ❌ 错误，应使用 dto.LocaleResponse
}
```

#### 响应参数示例

```go
// ✅ 正确 - 响应参数（基础名称）
type ExampleResp struct {
    Uuid       uint64             `json:"uuid"`
    LocaleName dto.LocaleResponse `json:"locale_name"` // 名称（多语言）
}

// ✅ 正确 - 响应参数（带描述的名称）
type ProductResp struct {
    Uuid                uint64             `json:"uuid"`
    LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称（多语言）
    LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 属性名称（多语言）
    LocaleUnitName      dto.LocaleResponse `json:"locale_unit_name"`      // 单位名称（多语言）
}

// ❌ 错误 - 字段名缺少 Locale 前缀
type ExampleResp struct {
    AttributeName dto.LocaleResponse `json:"attribute_name"` // ❌ 错误，应使用 LocaleAttributeName
}

// ❌ 错误 - 响应参数使用 map
type ExampleResp struct {
    Name map[string]string `json:"name"` // ❌ 错误，应使用 dto.LocaleResponse
}
```

#### 前端提交的 JSON 格式

```json
{
  "locale_name": {
    "zh": "商品名称",
    "th": "ชื่อสินค้า",
    "en": "Product Name",
    "zhtw": "商品名稱",
    "ja": "商品名",
    "ko": "상품명",
    "my": "ကုန်ပစ္စည်းအမည်",
    "tr": "Ürün Adı",
    "sv": "Produktnamn"
  },
  "price": 100.00
}
```

#### 从数据库获取并返回

```go
// 响应 DTO
type ProductResp struct {
    Uuid       uint64             `json:"uuid"`        // 商品UUID
    LocaleName dto.LocaleResponse `json:"locale_name"` // 商品名称（多语言）
    Price      float64            `json:"price"`       // 价格
}

// 从 Model 转换
resp := ProductResp{
    Uuid:       product.Uuid,
    LocaleName: product.MultiLanguageName.GetLocaleResponse(), // 从数据库获取
    Price:      product.Price,
}
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

