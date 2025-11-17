# Go Main 模块开发指南

> 👤 **受众**: 人类开发者  
> 📖 **用途**: Go Main 模块的详细开发规范、代码风格和最佳实践

---

## 📋 目录

1. [命名规范](#命名规范)
2. [代码风格](#代码风格)
3. [错误处理](#错误处理)
4. [API 响应规范](#api-响应规范)
5. [数据库操作](#数据库操作)
6. [性能优化](#性能优化)
7. [多语言国际化](#多语言国际化)
8. [安全规范](#安全规范)
9. [测试规范](#测试规范)
10. [代码提交](#代码提交)

---

## 命名规范

### 1. 结构体命名

**规则**: 使用大驼峰命名（PascalCase），ID 字段特殊处理

✅ **正确示例**:
```go
type Staff struct {
    StaffId   uint64  // ID字段：Id 大写
    StaffName string  // 其他字段：驼峰命名
    CompanyId uint64
    ShopId    uint64
}

type Product struct {
    ProductId   uint64
    ProductName string
    CategoryId  uint64
}
```

❌ **错误示例**:
```go
type staff struct {        // ❌ 小写
    staff_id   uint64     // ❌ 下划线命名
    staff_name string
}

type STAFF struct {        // ❌ 全大写
    STAFF_ID uint64
}
```

**原因**: 
- Go 惯例使用驼峰命名
- ID 大写符合 Go lint 规范
- 提高代码一致性和可读性

---

### 2. URL 命名规范

**规则**: 使用 snake_case（蛇形命名）

✅ **正确示例**:
```go
"/api/v1/passport/server_public_key"
"/api/v1/order/cart_info"
"/api/v1/cashier/order/list"
"/api/v1/shop/product/category"
```

❌ **错误示例**:
```go
"/api/v1/passport/server-public-key"  // ❌ kebab-case
"/api/v1/order/cartInfo"              // ❌ camelCase
"/api/v1/Cashier/Order/List"          // ❌ PascalCase
```

**原因**:
- 与后端其他模块保持一致
- URL 大小写不敏感，使用下划线更清晰
- 避免混淆（- 在某些环境中有特殊含义）

---

### 3. 包名和文件名

**规则**: 使用 snake_case

✅ **正确示例**:
```go
package member_service
// 文件名：member_service.go

package cashier_api
// 文件名：cashier_api.go
```

❌ **错误示例**:
```go
package memberService      // ❌ 驼峰
// 文件名：memberService.go

package MemberService      // ❌ 大驼峰
```

---

### 4. 接口和实现命名

**规则**: 接口以 `I` 开头，实现以 `Impl` 结尾

✅ **正确示例**:
```go
// 接口定义
type IProductCategoryRepo interface {
    CreateProductCategory(category model.ProductCategory) (uint64, error)
    GetProductCategory(opts ...DBOption) (model.ProductCategory, error)
}

// 实现
type ProductCategoryRepoImpl struct {
    db *gorm.DB
}

// Repository 简写成 Repo
type IOrderRepo interface {}
type OrderRepoImpl struct {}

// Service 简写成 Srv
type ICashierSrv interface {}
type CashierSrvImpl struct {}
```

**为什么要加 I 前缀?**

这是项目约定，有以下优点：
1. **明确标识**: 一眼看出是接口还是实现
2. **避免冲突**: 接口和结构体可以有相似名称
3. **IDE 支持**: 许多 IDE 对 I 开头的接口有特殊提示

**注意**: 虽然 Go 官方不推荐 I 前缀，但本项目为了团队一致性采用此规范。

---

## 代码风格

### 1. 包导入顺序

**规则**: 标准库 → 第三方包 → 项目包，用空行分隔

✅ **正确示例**:
```go
import (
    // 标准库
    "context"
    "fmt"
    "time"
    
    // 第三方包
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    // 项目包
    "ttpos-server-go/app/model"
    "ttpos-server-go/app/service"
    "ttpos-server-go/pkg/context"
)
```

❌ **错误示例**:
```go
import (
    "ttpos-server-go/app/model"  // ❌ 顺序错误
    "github.com/gin-gonic/gin"
    "context"
    "fmt"
)
```

**工具**: 使用 `goimports` 自动格式化

---

### 2. 注释规范

**规则**: 所有注释使用中文

✅ **正确示例**:
```go
// CreateOrder 创建订单
// 参数：ctx 上下文，req 创建订单请求
// 返回：订单响应，错误信息
func (s *OrderService) CreateOrder(ctx context.Context, req req.CreateOrderReq) (*resp.OrderResp, error) {
    // 验证请求参数
    if err := req.Validate(); err != nil {
        return nil, errors.WithMessage(err, "参数验证失败")
    }
    
    // 创建订单逻辑
    order := model.Order{
        ProductId: req.ProductId, // 商品ID
        Quantity:  req.Quantity,  // 数量
        Price:     req.Price,     // 价格
    }
    
    return &resp.OrderResp{}, nil
}
```

**注释类型**:

1. **包注释** (package comment):
```go
// Package cashier 提供收银相关的业务逻辑
package cashier
```

2. **函数注释** (function comment):
```go
// NewOrderService 创建订单服务实例
// 参数 dbm 数据库管理器
// 返回订单服务接口
func NewOrderService(dbm *database.DBManager) IOrderSrv {
    // ...
}
```

3. **结构体注释** (struct comment):
```go
// OrderService 订单服务
// 负责处理订单相关的业务逻辑
type OrderService struct {
    dbm        *database.DBManager
    orderRepo  repository.IOrderRepo
}
```

4. **字段注释** (field comment):
```go
type Order struct {
    Id          uint64  `json:"id"`           // 主键ID
    Uuid        uint64  `json:"uuid"`         // 唯一标识
    ProductName string  `json:"product_name"` // 商品名称
    Quantity    int     `json:"quantity"`     // 数量
    TotalAmount float64 `json:"total_amount"` // 总金额
}
```

---

### 3. 代码格式

**缩进**: 使用 Tab（不是空格）

**行宽**: 建议不超过 120 个字符

**函数长度**: 建议不超过 80 行

**使用工具**:
```bash
# 格式化代码
go fmt ./...

# 或使用 gofmt
gofmt -w .

# 更推荐使用 goimports（自动管理导入）
goimports -w .
```

---

## 错误处理

### 1. 不使用 panic

**规则**: 返回 error，不要使用 panic

✅ **正确示例**:
```go
func GetUser(id uint64) (*User, error) {
    if id == 0 {
        return nil, errors.New("用户ID不能为空")
    }
    
    user, err := userRepo.GetUser(id)
    if err != nil {
        return nil, errors.WithMessage(err, "查询用户失败")
    }
    
    return user, nil
}
```

❌ **错误示例**:
```go
func GetUser(id uint64) *User {
    if id == 0 {
        panic("用户ID不能为空")  // ❌ 不要使用 panic
    }
    
    user, err := userRepo.GetUser(id)
    if err != nil {
        panic(err)  // ❌ 不要 panic
    }
    
    return user
}
```

**何时可以 panic?**
- 程序初始化失败（配置加载、数据库连接失败）
- 不可恢复的致命错误

```go
func init() {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("数据库连接失败: " + err.Error())  // ✅ 初始化可以 panic
    }
}
```

---

### 2. 错误包装

**规则**: 使用 `errors.WithMessage` 包装错误，添加上下文信息

✅ **正确示例**:
```go
import "ttpos-server-go/app/errors"

func GetUser(id uint64) (*User, error) {
    user, err := repo.FindUser(id)
    if err != nil {
        // 包装错误，添加上下文
        return nil, errors.WithMessage(err, "查询用户失败")
    }
    return user, nil
}

func CreateOrder(req req.CreateOrderReq) error {
    // 验证商品
    product, err := productRepo.GetProduct(req.ProductId)
    if err != nil {
        return errors.WithMessage(err, fmt.Sprintf("查询商品失败，商品ID: %d", req.ProductId))
    }
    
    // 创建订单
    err = orderRepo.CreateOrder(order)
    if err != nil {
        return errors.WithMessage(err, "创建订单失败")
    }
    
    return nil
}
```

**为什么要包装错误?**
- 提供更多上下文信息
- 便于调试和日志记录
- 保留原始错误堆栈

---

### 3. 错误检查

**规则**: 每个可能返回错误的函数调用都要检查

✅ **正确示例**:
```go
result, err := someFunction()
if err != nil {
    return nil, errors.WithMessage(err, "操作失败")
}
// 使用 result
```

❌ **错误示例**:
```go
result, _ := someFunction()  // ❌ 忽略错误
// 直接使用 result，可能导致 panic
```

---

### 4. 验证错误信息国际化

**规则**: 定义错误消息映射，支持多语言

✅ **正确示例**:
```go
var AddMemberReqMessage = map[string]string{
    "level_uuid.required": "会员等级不存在",
    "phone.max":           "手机号不能超过20个字符",
    "phone.mobile":        "手机号格式不正确",
    "nickname.max":        "昵称不能超过50个字符",
    "email.email":         "邮箱格式不正确",
}

type AddMemberReq struct {
    Nickname  string `json:"nickname" binding:"omitempty,max=50"`
    Phone     string `json:"phone" binding:"required,max=20,mobile"`
    LevelUuid uint64 `json:"level_uuid" binding:"required"`
    Email     string `json:"email" binding:"omitempty,email"`
}

func (req *AddMemberReq) GetMessages() map[string]string {
    return AddMemberReqMessage
}
```

---

## API 响应规范

### 1. 统一响应格式

**规则**: 所有 API 返回统一格式 `{code, message, data}`

✅ **正确的响应格式**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [{"foo": "bar"}],
    "options": {
      "list": [{"key": "k1", "value": "v1"}]
    },
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

**data 字段规则**:
- ✅ **必须是对象** (object)
- ❌ **不能是 null**
- ❌ **不能是数组**

❌ **错误的响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": null           // ❌ 不能为 null
}

{
  "code": 1,
  "message": "success",
  "data": []             // ❌ 不能是数组
}
```

**原因**: 
- 前端可以统一处理 `data` 对象
- 避免 null 导致的前端报错
- 保持响应格式一致性

---

### 2. 使用 helper 返回响应

**规则**: 在 Handler 中使用 helper 统一返回

✅ **正确示例**:
```go
import (
    "ttpos-server-go/app/api/helper"
    "ttpos-server-go/app/constant"
    "github.com/gin-gonic/gin"
)

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // ...业务逻辑
    
    // 成功响应
    helper.Success(c, resp)
    
    // 或者空数据
    helper.Success(c, gin.H{})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
    // ...业务逻辑
    
    // 失败响应
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, err)
        return
    }
    
    helper.Success(c, orderResp)
}
```

**Helper 函数说明**:
```go
// helper.Success(c, data)
// 成功响应，code=1

// helper.ErrorWithDetail(c, code, err)
// 错误响应，自动处理错误信息

// helper.Error(c, code, message)
// 自定义错误消息
```

---

### 3. 分页信息规范

**规则**: 分页信息统一放在 `meta` 中

✅ **正确示例**:
```go
type OrderListPaginationResp struct {
    List []OrderItem    `json:"list"`  // 列表数据
    Meta OrderListMeta  `json:"meta"`  // 分页和统计信息
}

type OrderListMeta struct {
    dto.PageResponse                     // 内嵌分页结构
    TotalNum    int64 `json:"total_num"`     // 总数量
    UnpaidNum   int64 `json:"unpaid_num"`    // 待付款数量
    CompleteNum int64 `json:"complete_num"`  // 已完成数量
    CancelNum   int64 `json:"cancel_num"`    // 已取消数量
}

type PageResponse struct {
    PageNo   int   `json:"page_no"`    // 当前页码
    PageSize int   `json:"page_size"`  // 每页大小
    Total    int64 `json:"total"`      // 总记录数
}
```

**JSON 响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {"id": 1, "name": "订单1"},
      {"id": 2, "name": "订单2"}
    ],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100,
      "total_num": 100,
      "unpaid_num": 10,
      "complete_num": 85,
      "cancel_num": 5
    }
  }
}
```

---

## 数据库操作

### 1. 使用 DBManager

**规则**: Service 层使用 DBManager 获取数据库连接

✅ **正确示例**:
```go
type orderSrv struct {
    dbm *database.DBManager  // 持有 DBManager
}

func (s *orderSrv) CreateOrder(ctx context.Context) error {
    // 根据 Context 获取对应商户的数据库
    db := s.dbm.GetDB(ctx.GetDbId())
    
    // 使用 db 进行操作
    err := db.Create(&order).Error
    return errors.WithMessage(err, "创建订单失败")
}
```

❌ **错误示例**:
```go
// ❌ Repository 不能持有 DBManager
type OrderRepoImpl struct {
    dbm *database.DBManager  // ❌ 错误
}

// ✅ Repository 只能持有 db 实例
type OrderRepoImpl struct {
    db *gorm.DB  // ✅ 正确
}
```

---

### 2. 预加载避免 N+1 查询

**规则**: 使用 Preload 预加载关联数据

✅ **正确示例**:
```go
func (r *orderRepo) GetOrderWithProducts(orderUuid uint64) (*model.Order, error) {
    var order model.Order
    err := r.db.Preload("Products").                // 预加载商品
        Preload("Products.Product").                 // 预加载商品详情
        Preload("Member").                           // 预加载会员
        Where("uuid = ?", orderUuid).
        First(&order).Error
    
    return &order, errors.WithMessage(err, "查询订单失败")
}
```

❌ **错误示例**:
```go
// ❌ N+1 查询问题
func (r *orderRepo) GetOrderWithProducts(orderUuid uint64) (*model.Order, error) {
    var order model.Order
    r.db.Where("uuid = ?", orderUuid).First(&order)
    
    // ❌ 每次都查询数据库，如果有 10 个订单项就查 10 次
    for i := range order.Products {
        r.db.Where("id = ?", order.Products[i].ProductId).First(&order.Products[i].Product)
    }
    
    return &order, nil
}
```

---

### 3. 使用索引和分页

✅ **正确示例**:
```go
func (r *orderRepo) GetOrderList(req req.OrderListReq) ([]model.Order, int64, error) {
    var orders []model.Order
    var total int64
    
    // 构建查询
    db := r.db.Model(&model.Order{}).
        Where("company_uuid = ?", req.CompanyUuid).    // 使用索引字段
        Where("create_time >= ?", req.StartTime).
        Where("create_time <= ?", req.EndTime)
    
    // 先计算总数
    db.Count(&total)
    
    // 分页查询
    err := db.Offset((req.PageNo - 1) * req.PageSize).
        Limit(req.PageSize).
        Order("create_time DESC").                     // 索引排序
        Find(&orders).Error
    
    return orders, total, errors.WithMessage(err, "查询订单列表失败")
}
```

---

## 性能优化

### 1. API 响应时间要求

**规则**: 本地响应时间应在 200ms 以内

✅ **监控示例**:
```go
func (c *Controller) GetOrderList(ctx *gin.Context) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 200*time.Millisecond {
            logger.Logger.Warn("API响应时间过长",
                zap.String("api", ctx.Request.URL.Path),
                zap.Duration("duration", duration),
            )
        }
    }()
    
    // 业务逻辑
    // ...
}
```

**优化建议**:
- 使用缓存减少数据库查询
- 异步处理非关键业务
- 使用索引优化查询
- 避免 N+1 查询
- 使用连接池

---

### 2. 数据库查询优化

**技巧**:

1. **只查询需要的字段**:
```go
// ✅ 只查询需要的字段
db.Select("id, name, price").Find(&products)

// ❌ 查询所有字段（浪费）
db.Find(&products)
```

2. **使用批量操作**:
```go
// ✅ 批量插入
db.CreateInBatches(products, 100)

// ❌ 循环插入（慢）
for _, p := range products {
    db.Create(&p)
}
```

3. **使用原生 SQL（复杂查询）**:
```go
db.Raw("SELECT ... FROM ... WHERE ... ").Scan(&results)
```

---

## 多语言国际化

### 1. 支持的语言列表

系统支持以下 10 种语言：

| 代码 | 语言 | 代码 | 语言 |
|------|------|------|------|
| `zh` | 简体中文 | `en` | 英语 |
| `zhtw` | 繁体中文 | `th` | 泰语 |
| `ja` | 日语 | `ko` | 韩语 |
| `my` | 缅甸语 | `tr` | 土耳其语 |
| `de` | 德语 | `sv` | 瑞典语 |

---

### 2. 获取当前语言

✅ **从 Context 获取**:
```go
func (s *orderSrv) CreateOrder(ctx context.Context) error {
    // 获取当前请求的语言
    language := ctx.GetLanguage()
    
    // 使用语言进行翻译
    message := i18n.Translate(language, "订单创建成功")
    
    return nil
}
```

✅ **从 Gin Context 获取**:
```go
import "ttpos-server-go/i18n"

func (c *Controller) GetOrder(ginCtx *gin.Context) {
    // 从请求头获取语言
    language := i18n.GetAcceptLanguage(ginCtx)
    
    // 使用语言
    message := i18n.Translate(language, "查询成功")
}
```

---

### 3. 使用 i18n 翻译

✅ **基础翻译**:
```go
import "ttpos-server-go/i18n"

// 简单翻译
message := i18n.Translate(ctx.GetLanguage(), "操作成功")

// 带参数的翻译
message := i18n.Translate(
    ctx.GetLanguage(),
    "物品 %s 库存不足",
    materialName,
)

// 多个参数
message := i18n.Translate(
    ctx.GetLanguage(),
    "%s管理员添加会员发卡赠送操作 [%s]",
    adminName,
    operationType,
)
```

---

### 4. 多语言数据结构

✅ **使用 LocaleResponse**:
```go
import "ttpos-server-go/app/dto"

type ProductResp struct {
    ProductId uint64              `json:"product_id"`
    Name      dto.LocaleResponse  `json:"name"`  // 多语言名称
}

// 构建多语言响应
product := ProductResp{
    ProductId: 1001,
    Name: dto.LocaleResponse{
        ZH:   "商品名称",
        EN:   "Product Name",
        TH:   "ชื่อสินค้า",
        JA:   "商品名",
        KO:   "상품명",
        MY:   "ကုန်ပစ္စည်းအမည်",
        TR:   "Ürün Adı",
        DE:   "Produktname",
        SV:   "Produktnamn",
        ZHTW: "商品名稱",
    },
}
```

**常用方法**:
```go
// 获取指定语言的值
name := localeResp.GetLocale("zh")

// 设置指定语言的值
localeResp.SetLocale("en", "Product Name")

// 转换为 JSON 字符串
jsonStr := localeResp.ToJson()

// 检查是否为空
if localeResp.IsNull() {
    // 所有语言都为空
}

// 检查是否包含必需的语言
requiredLocales := []string{"zh", "en", "th"}
if !localeResp.CheckRequiredLocale(requiredLocales) {
    return errors.New("缺少必需的语言翻译")
}
```

---

### 5. 错误消息国际化

✅ **返回国际化错误消息**:
```go
func (s *orderSrv) CancelOrder(ctx context.Context, req req.OrderCancelReq) error {
    // 验证订单状态
    if order.Status != constant.OrderStatusPending {
        return errors.New(
            i18n.Translate(ctx.GetLanguage(), "订单状态不允许取消"),
        )
    }
    
    // 带参数的错误消息
    if stock < required {
        return errors.New(
            i18n.Translate(
                ctx.GetLanguage(),
                "物品 %s 库存不足，当前库存：%d，需要：%d",
                materialName,
                stock,
                required,
            ),
        )
    }
    
    return nil
}
```

---

### 6. 最佳实践

✅ **多语言使用建议**:

1. **所有面向用户的文本都要支持多语言**
   - 错误消息
   - 提示信息
   - 状态文本

2. **使用翻译键而不是硬编码文本**
   ```go
   // ✅ 正确
   message := i18n.Translate(ctx.GetLanguage(), "操作成功")
   
   // ❌ 错误
   message := "操作成功"
   ```

3. **新增翻译键时同步更新所有语言包**
   - 在 `i18n/languages/` 下的所有 JSON 文件中添加

4. **参数化的翻译使用占位符**
   ```go
   // ✅ 正确 - 使用 %s 占位符
   i18n.Translate(ctx.GetLanguage(), "物品 %s 库存不足", name)
   
   // ❌ 错误 - 字符串拼接
   i18n.Translate(ctx.GetLanguage(), "物品") + name + i18n.Translate(ctx.GetLanguage(), "库存不足")
   ```

---

## 安全规范

### 1. 操作权限验证

✅ **验证操作权限**:
```go
func (s *orderSrv) CancelOrder(ctx context.Context, req req.OrderCancelReq) error {
    // 1. 验证订单是否可以取消
    saleBill, err := s.IsCellCancelOrder(ctx, req.SaleBillUuid)
    if err != nil {
        return errors.WithMessage(err, "订单不可取消")
    }
    
    // 2. 验证用户权限
    if !ctx.HasPermission("order.cancel") {
        return errors.New("无权限取消订单")
    }
    
    // 3. 业务逻辑
    // ...
    
    return nil
}
```

---

### 2. 参数验证

✅ **严格的参数验证**:
```go
type OrderCancelReq struct {
    SaleBillUuid uint64 `json:"sale_bill_uuid" binding:"required,gt=0"`
    CancelReason string `json:"cancel_reason" binding:"required,max=500"`
}

func (req *OrderCancelReq) Validate() error {
    if req.SaleBillUuid == 0 {
        return errors.New("销售账单UUID不能为空")
    }
    
    if req.CancelReason == "" {
        return errors.New("取消原因不能为空")
    }
    
    if len(req.CancelReason) > 500 {
        return errors.New("取消原因不能超过500字符")
    }
    
    return nil
}
```

---

## 测试规范

### 1. 单元测试

✅ **测试文件命名**: `*_test.go`

✅ **测试示例**:
```go
// order_service_test.go
package cashier

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestOrderService_CreateOrder(t *testing.T) {
    // Arrange - 准备测试数据
    orderSrv := NewOrderSrv(mockDBM, mockOrderRepo, mockProductRepo)
    req := req.CreateOrderReq{
        ProductUuid: 123,
        Quantity:    2,
    }
    
    // Act - 执行测试
    resp, err := orderSrv.CreateOrder(context.Background(), req)
    
    // Assert - 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Greater(t, resp.OrderUuid, uint64(0))
}
```

---

### 2. 表驱动测试

✅ **表驱动测试示例**:
```go
func TestValidatePhone(t *testing.T) {
    tests := []struct {
        name    string
        phone   string
        wantErr bool
    }{
        {"有效手机号", "13800138000", false},
        {"无效手机号-太短", "138001", true},
        {"无效手机号-格式错误", "12345678901", true},
        {"空手机号", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePhone(tt.phone)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

## 代码提交

### 1. 提交前检查

```bash
# 1. 整理依赖
go mod tidy

# 2. 格式化代码
go fmt ./...
# 或
goimports -w .

# 3. 检查代码问题
go vet ./...

# 4. 运行测试
go test ./...

# 5. 运行 linter (如果有)
golangci-lint run
```

---

### 2. 忽略文件

**禁止提交以下文件和目录**:

- `.idea/`, `.vscode/` - IDE 配置文件
- `tmp/`, `bin/`, `__debug_*` - 构建产物和临时文件
- `*.log` - 日志文件
- 包含敏感信息的配置文件

**.gitignore 示例**:
```gitignore
# IDE
.idea/
.vscode/

# 构建产物
tmp/
bin/
__debug_*
main

# 日志
*.log
log/

# 配置（包含敏感信息）
config/config.yaml
!config/config.example.yaml
```

---

## 相关文档

- [Go Main 架构设计](../architecture/go-main-architecture.md) - 深入理解架构设计
- [API 设计指南](./api-design-guide.md) - RESTful API 设计详解
- [数据库开发指南](./database-guide.md) - 数据库设计和操作
- [安全开发指南](./security-guide.md) - 安全开发详细规范
- [Go 测试标准](../testing/standards/go-testing.md) - Go 测试规范

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

