# Go Main 模块架构设计

> 👤 **受众**: 人类开发者  
> 📖 **用途**: 深入理解 Go Main 模块的架构设计和实现原理

---

## 📋 目录

1. [架构概览](#架构概览)
2. [项目结构](#项目结构)
3. [分层架构设计](#分层架构设计)
4. [服务层设计](#服务层设计)
5. [Repository层设计](#repository层设计)
6. [事件总线](#事件总线)
7. [并发控制](#并发控制)
8. [数据库管理](#数据库管理)
9. [缓存管理](#缓存管理)
10. [设计原则](#设计原则)

---

## 架构概览

### 整体架构

Go Main 模块采用**模块化单体架构**（Modular Monolith），使用 **Gin** 框架构建，提供核心业务 API。

```
┌─────────────────────────────────────────┐
│         HTTP API 层 (Gin)              │
│     /api/v1/cashier/*                  │
│     /api/v1/shop/*                     │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          中间件层                        │
│   认证、授权、日志、限流、跨域等         │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          API 控制器层                    │
│   参数验证、请求处理、响应封装          │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          服务层 (Service)               │
│   业务逻辑编排、事务管理、事件发布      │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        数据访问层 (Repository)          │
│   数据库CRUD、选项模式查询              │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          数据层 (GORM)                  │
│   多租户数据库、连接池管理              │
└─────────────────────────────────────────┘
```

### 核心特点

1. **依赖注入**: 通过构造函数注入依赖，便于测试和解耦
2. **接口驱动**: Service 和 Repository 都定义接口，提高灵活性
3. **事件驱动**: 使用事件总线实现模块间解耦
4. **多租户**: 每个商户独立数据库，数据完全隔离
5. **并发安全**: UUID锁机制保证资源访问安全

---

## 项目结构

```
main/
├── app/
│   ├── api/v1/              # API 控制器层
│   │   ├── cashier/         # 收银端 API
│   │   ├── shop/            # 店铺端 API
│   │   └── common/          # 共享 API
│   ├── service/             # 业务逻辑层
│   │   ├── cashier/         # 收银服务
│   │   ├── member/          # 会员服务
│   │   ├── order/           # 订单服务
│   │   └── product/         # 商品服务
│   ├── repository/          # 数据访问层
│   │   ├── order_repo.go
│   │   ├── product_repo.go
│   │   └── user_repo.go
│   ├── model/               # 数据模型
│   │   ├── order.go
│   │   ├── product.go
│   │   └── user.go
│   ├── dto/                 # 数据传输对象
│   │   ├── req/             # 请求参数
│   │   └── resp/            # 响应数据
│   ├── constant/            # 常量定义
│   ├── errors/              # 错误定义
│   └── event/               # 事件处理器
├── pkg/                     # 基础设施包
│   ├── database/            # 数据库管理器
│   ├── cache/               # 缓存管理
│   ├── eventbus/            # 事件总线
│   ├── lock/                # 并发锁
│   ├── logger/              # 日志
│   └── utils/               # 工具类
├── middleware/              # 中间件
├── router/                  # 路由配置
├── config/                  # 配置管理
└── main.go                  # 入口文件
```

### 目录职责

| 目录 | 职责 | 可修改 |
|------|------|--------|
| `app/api/` | HTTP 请求处理、参数验证、响应封装 | ✅ |
| `app/service/` | 业务逻辑编排、事务控制、跨 Repository 调用 | ✅ |
| `app/repository/` | 数据库 CRUD、选项模式查询 | ✅ |
| `app/model/` | 数据库表结构映射 | ✅ |
| `app/dto/` | 请求响应结构定义 | ✅ |
| `pkg/` | 基础设施和工具类 | ⚠️ 谨慎修改 |

---

## 分层架构设计

### 1. API 控制器层 (Handler)

**职责**:
- 接收 HTTP 请求
- 参数验证和绑定
- 调用 Service 层
- 封装统一响应格式

**示例**:
```go
// app/api/v1/cashier/order_handler.go
package cashier

import (
    "github.com/gin-gonic/gin"
    "ttpos-server-go/app/api/helper"
    "ttpos-server-go/app/dto/req"
    "ttpos-server-go/app/service/cashier"
)

type OrderHandler struct {
    orderSrv cashier.IOrderSrv
}

func NewOrderHandler(orderSrv cashier.IOrderSrv) *OrderHandler {
    return &OrderHandler{
        orderSrv: orderSrv,
    }
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // 1. 参数绑定和验证
    var req req.CreateOrderReq
    if err := c.ShouldBindJSON(&req); err != nil {
        helper.ErrorWithDetail(c, constant.CodeInvalidParam, err)
        return
    }
    
    // 2. 调用服务层
    ctx := context.GetContext(c)
    resp, err := h.orderSrv.CreateOrder(ctx, req)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, err)
        return
    }
    
    // 3. 返回响应
    helper.Success(c, resp)
}
```

**核心原则**:
- ❌ **不要在 Handler 中编写业务逻辑**
- ✅ **只做参数处理和服务调用**
- ✅ **使用 helper 统一返回格式**

---

### 2. 服务层 (Service)

**职责**:
- 业务逻辑编排
- 事务管理
- 调用多个 Repository
- 发布领域事件

**接口定义**:
```go
// app/service/cashier/order_srv.go
package cashier

import (
    "context"
    "ttpos-server-go/app/dto/req"
    "ttpos-server-go/app/dto/resp"
)

// IOrderSrv 订单服务接口
type IOrderSrv interface {
    CreateOrder(ctx context.Context, req req.CreateOrderReq) (*resp.CreateOrderResp, error)
    GetOrderList(ctx context.Context, req req.OrderListReq) (*resp.OrderListResp, error)
    CancelOrder(ctx context.Context, req req.CancelOrderReq) error
}
```

**实现**:
```go
type orderSrv struct {
    dbm            *database.DBManager
    orderRepo      repository.IOrderRepo
    productRepo    repository.IProductRepo
    memberSrv      IMemberSrv
    paymentSrv     IPaymentSrv
    systemLock     lock.Lock
}

// 构造函数 - 依赖注入
func NewOrderSrv(
    dbm *database.DBManager,
    orderRepo repository.IOrderRepo,
    productRepo repository.IProductRepo,
    memberSrv IMemberSrv,
    paymentSrv IPaymentSrv,
) IOrderSrv {
    return &orderSrv{
        dbm:         dbm,
        orderRepo:   orderRepo,
        productRepo: productRepo,
        memberSrv:   memberSrv,
        paymentSrv:  paymentSrv,
        systemLock:  lock.NewSystemLock(),
    }
}

func (s *orderSrv) CreateOrder(ctx context.Context, req req.CreateOrderReq) (*resp.CreateOrderResp, error) {
    // 1. 获取数据库连接
    db := s.dbm.GetDB(ctx.GetDbId())
    
    // 2. 开启事务
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // 3. 业务逻辑
    // 验证商品库存
    product, err := s.productRepo.GetProduct(
        s.productRepo.WithDB(tx),
        s.productRepo.WhereUuid(req.ProductUuid),
    )
    if err != nil {
        tx.Rollback()
        return nil, errors.WithMessage(err, "查询商品失败")
    }
    
    // 创建订单
    order := model.Order{
        Uuid:        genUuid(),
        CompanyUuid: ctx.GetCompanyUuid(),
        ProductUuid: req.ProductUuid,
        Quantity:    req.Quantity,
        TotalAmount: product.Price * float64(req.Quantity),
    }
    
    if err := s.orderRepo.CreateOrder(order); err != nil {
        tx.Rollback()
        return nil, errors.WithMessage(err, "创建订单失败")
    }
    
    // 4. 提交事务
    if err := tx.Commit().Error; err != nil {
        return nil, errors.WithMessage(err, "提交事务失败")
    }
    
    // 5. 发布事件
    go func() {
        event.NewSystemBus().PublishOrderCreatedEvent(event.OrderCreatedPayload{
            BasePayload: event.BasePayload{
                Ctx:         ctx,
                CompanyUuid: ctx.GetCompanyUuid(),
            },
            OrderUuid:   order.Uuid,
            OrderAmount: order.TotalAmount,
        })
    }()
    
    return &resp.CreateOrderResp{
        OrderUuid: order.Uuid,
    }, nil
}
```

**设计原则**:
- ✅ **通过接口依赖其他服务**
- ✅ **Service 只依赖 Service，不直接依赖 Repository**
- ✅ **事务在 Service 层管理**
- ✅ **发布事件使用 goroutine 异步**
- ❌ **不要在 Service 中直接操作数据库**

---

### 3. Repository 层

**职责**:
- 封装数据库操作
- 提供选项模式查询
- 只操作单表或简单关联

**接口定义**:
```go
// app/repository/order_repo.go
package repository

import (
    "ttpos-server-go/app/model"
    "gorm.io/gorm"
)

// DBOption 数据库选项函数
type DBOption func(*gorm.DB) *gorm.DB

// IOrderRepo 订单仓储接口
type IOrderRepo interface {
    // 查询选项
    WhereUuid(uuid uint64) DBOption
    WhereStatus(status uint8) DBOption
    WhereCompanyUuid(companyUuid uint64) DBOption
    WithDB(db *gorm.DB) DBOption
    
    // CRUD 操作
    CreateOrder(order model.Order) error
    GetOrder(opts ...DBOption) (*model.Order, error)
    GetOrderList(opts ...DBOption) ([]model.Order, int64, error)
    UpdateOrder(order model.Order, opts ...DBOption) error
    DeleteOrder(opts ...DBOption) error
}
```

**实现**:
```go
type OrderRepoImpl struct {
    db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) IOrderRepo {
    return &OrderRepoImpl{db: db}
}

// 选项模式实现
func (r *OrderRepoImpl) WhereUuid(uuid uint64) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("uuid = ?", uuid)
    }
}

func (r *OrderRepoImpl) WhereStatus(status uint8) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", status)
    }
}

func (r *OrderRepoImpl) WithDB(tx *gorm.DB) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return tx
    }
}

// 查询操作
func (r *OrderRepoImpl) GetOrder(opts ...DBOption) (*model.Order, error) {
    db := r.db
    for _, opt := range opts {
        db = opt(db)
    }
    
    var order model.Order
    err := db.First(&order).Error
    return &order, err
}

// 使用示例
// order, err := orderRepo.GetOrder(
//     orderRepo.WhereUuid(orderUuid),
//     orderRepo.WhereStatus(constant.OrderStatusActive),
// )
```

**设计原则**:
- ✅ **使用选项模式提高灵活性**
- ✅ **Repository 只能传入 db 实例，不能传入 dbm**
- ✅ **Repository 不包含业务逻辑**
- ❌ **不要在 Repository 中调用其他 Repository**

---

## 服务层设计

### 依赖注入模式

**原则**: Service 依赖通过构造函数注入

```go
// ✅ 正确：通过构造函数注入
type orderSrv struct {
    dbm            *database.DBManager
    orderRepo      repository.IOrderRepo
    productRepo    repository.IProductRepo
    memberSrv      IMemberSrv      // 依赖其他服务
    paymentSrv     IPaymentSrv
}

func NewOrderSrv(
    dbm *database.DBManager,
    orderRepo repository.IOrderRepo,
    productRepo repository.IProductRepo,
    memberSrv IMemberSrv,
    paymentSrv IPaymentSrv,
) IOrderSrv {
    return &orderSrv{
        dbm:         dbm,
        orderRepo:   orderRepo,
        productRepo: productRepo,
        memberSrv:   memberSrv,
        paymentSrv:  paymentSrv,
    }
}
```

```go
// ❌ 错误：Service 不能直接依赖其他 Service 的 Repository
type orderSrv struct {
    memberRepo   repository.IMemberRepo   // ❌ 错误
}

// ✅ 正确：应该依赖 Service 接口
type orderSrv struct {
    memberSrv    IMemberSrv               // ✅ 正确
}
```

---

## Repository层设计

### 选项模式 (Option Pattern)

选项模式使查询更加灵活和可组合。

**优点**:
- 查询条件可自由组合
- 避免参数爆炸
- 便于扩展新条件
- 代码清晰易读

**实现**:
```go
type DBOption func(*gorm.DB) *gorm.DB

// 定义各种查询条件
func (r *OrderRepoImpl) WhereUuid(uuid uint64) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("uuid = ?", uuid)
    }
}

func (r *OrderRepoImpl) WhereCompanyUuid(companyUuid uint64) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("company_uuid = ?", companyUuid)
    }
}

func (r *OrderRepoImpl) WhereStatus(status uint8) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", status)
    }
}

func (r *OrderRepoImpl) WhereCreateTimeRange(startTime, endTime int) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
    }
}

// 应用选项
func (r *OrderRepoImpl) GetOrderList(opts ...DBOption) ([]model.Order, int64, error) {
    db := r.db
    for _, opt := range opts {
        db = opt(db)
    }
    
    var orders []model.Order
    var total int64
    
    db.Count(&total)
    err := db.Find(&orders).Error
    
    return orders, total, err
}
```

**使用示例**:
```go
// 灵活组合查询条件
orders, total, err := orderRepo.GetOrderList(
    orderRepo.WhereCompanyUuid(companyUuid),
    orderRepo.WhereStatus(constant.OrderStatusActive),
    orderRepo.WhereCreateTimeRange(startTime, endTime),
)
```

---

## 事件总线

### 架构设计

事件总线用于模块间解耦，采用**发布-订阅模式**。

**组件**:
```
┌─────────────┐       ┌──────────────┐       ┌─────────────┐
│  Publisher  │──────▶│  Event Bus   │──────▶│ Subscriber  │
│   (发布者)  │       │  (事件总线)  │       │  (订阅者)   │
└─────────────┘       └──────────────┘       └─────────────┘
      │                      │                       │
      │                      │                       │
   Service              In-Memory                  Event
     层                    Queue                  Handler
```

### 事件定义

**位置**: `pkg/eventbus/event/`

```go
// pkg/eventbus/event/order_created_event.go

const EventOrderCreated EventName = "Event_Order_Created"

// 事件载荷
type OrderCreatedPayload struct {
    BasePayload
    OrderUuid   uint64  `json:"order_uuid"`
    OrderAmount float64 `json:"order_amount"`
}

// 事件处理器类型
type OrderCreatedHandler func(msg OrderCreatedPayload)

// 发布事件
func (system *SystemEventBus) PublishOrderCreatedEvent(msg OrderCreatedPayload) {
    system.bus.Publish(eventbus.Event{
        Name:    string(EventOrderCreated),
        Payload: msg,
    })
}

// 订阅事件
func (system *SystemEventBus) SubscribeOrderCreatedEvent(handler OrderCreatedHandler) {
    system.bus.Subscribe(string(EventOrderCreated), func(event eventbus.Event) {
        msg := event.Payload.(OrderCreatedPayload)
        handler(msg)
    })
}
```

### 事件发布

**在 Service 中发布**:
```go
func (s *orderSrv) CreateOrder(ctx context.Context, req req.CreateOrderReq) error {
    // 业务逻辑
    // ...
    
    // 异步发布事件
    go func() {
        event.NewSystemBus().PublishOrderCreatedEvent(event.OrderCreatedPayload{
            BasePayload: event.BasePayload{
                Ctx:          ctx,
                CompanyUuid:  ctx.GetCompanyUuid(),
            },
            OrderUuid:   orderUuid,
            OrderAmount: totalAmount,
        })
    }()
    
    return nil
}
```

### 事件订阅

**在 app/event/ 目录下创建处理器**:
```go
// app/event/order_created_event_handler.go

func InitOrderCreatedEventHandler() {
    event.NewSystemBus().SubscribeOrderCreatedEvent(func(msg event.OrderCreatedPayload) {
        // 处理订单创建后的逻辑
        // 例如：更新库存、发送通知、积分奖励等
        handleOrderCreated(msg)
    })
}

func handleOrderCreated(msg event.OrderCreatedPayload) {
    // 具体处理逻辑
    log.Printf("订单创建事件: OrderUuid=%d, Amount=%.2f", msg.OrderUuid, msg.OrderAmount)
}
```

**启动时注册**:
```go
// main.go
func main() {
    // 初始化事件处理器
    event_handler.InitOrderCreatedEventHandler()
    event_handler.InitPaymentSuccessEventHandler()
    
    // 启动服务
    r := gin.Default()
    // ...
}
```

**使用原则**:
- ✅ **事件发布使用 goroutine 异步**
- ✅ **事件处理器应该快速返回**
- ✅ **事件处理失败不应影响主流程**
- ❌ **不要在事件处理器中进行长时间操作**

---

## 并发控制

### UUID 锁机制

用于保护共享资源，避免并发冲突。

**使用场景**:
- 桌台操作（开台、关台、转台）
- 订单操作（下单、支付、退款）
- 库存操作（入库、出库、调拨）

**实现**:
```go
type Service struct {
    systemLock lock.Lock
}

func NewService() *Service {
    return &Service{
        systemLock: lock.NewSystemLock(),
    }
}

func (s *Service) OpenDesk(deskUuid uint64) error {
    // 获取锁
    s.systemLock.LockUuid(deskUuid)
    defer s.systemLock.UnlockUuid(deskUuid)
    
    // 业务逻辑（保证串行执行）
    // ...
    
    return nil
}

// 资源销毁时清理锁
func (s *Service) DeleteDesk(deskUuid uint64) error {
    // 业务逻辑
    // ...
    
    // 清理UUID锁资源
    s.systemLock.ClearUuidLock(deskUuid)
    return nil
}
```

### Context 副本

**问题**: Context 在 goroutine 中传递可能导致数据竞争

**解决方案**:
```go
func (s *Service) ProcessAsync(ctx context.Context) {
    go func() {
        // 使用 Context 副本
        ctxCopy := ctx.Copy()
        // ⚠️ 注意：不能对副本进行写操作
        s.processInBackground(ctxCopy)
    }()
}
```

---

## 数据库管理

### DBManager 多租户设计

**每个商户一个独立数据库**:
```go
// pkg/database/db_manager.go

type DBManager struct {
    dbPool map[string]*gorm.DB  // dbId -> DB连接
}

func (m *DBManager) GetDB(dbId string) *gorm.DB {
    if db, ok := m.dbPool[dbId]; ok {
        return db
    }
    
    // 动态创建连接
    db := m.createConnection(dbId)
    m.dbPool[dbId] = db
    return db
}
```

**使用**:
```go
func (s *orderSrv) CreateOrder(ctx context.Context) error {
    // 根据 Context 中的 dbId 获取对应数据库连接
    db := s.dbm.GetDB(ctx.GetDbId())
    
    // 使用 db 进行操作
    // ...
}
```

**规范**:
- ✅ **Service 层持有 DBManager**
- ✅ **Repository 层只持有 db 实例**
- ❌ **Repository 不能持有 DBManager**

---

## 缓存管理

### Redis 缓存 Key 管理

**核心原则**: 所有缓存 key 必须统一使用全局常量管理，禁止硬编码。

**实现方式**:

在 `app/constant/` 目录下创建缓存 key 常量文件：

```go
// app/constant/cache_key.go
package constant

const (
    // 订单相关缓存
    CacheKeyOrderInfo    = "order:info:%d"        // 订单详情，参数：orderUuid
    CacheKeyOrderList    = "order:list:%d:%d"     // 订单列表，参数：companyUuid, page
    CacheKeyOrderStatus  = "order:status:%d"       // 订单状态，参数：orderUuid
    
    // 商品相关缓存
    CacheKeyProductInfo  = "product:info:%d"       // 商品详情，参数：productUuid
    CacheKeyProductList  = "product:list:%d"      // 商品列表，参数：companyUuid
    
    // 会员相关缓存
    CacheKeyMemberInfo   = "member:info:%d"        // 会员信息，参数：memberUuid
    CacheKeyMemberPoints = "member:points:%d"      // 会员积分，参数：memberUuid
    
    // 桌台相关缓存
    CacheKeyDeskInfo     = "desk:info:%d"         // 桌台信息，参数：deskUuid
    CacheKeyDeskStatus   = "desk:status:%d"       // 桌台状态，参数：deskUuid
)
```

**使用示例**:

```go
// ✅ 正确：使用常量
import "ttpos-server-go/app/constant"

func (s *orderSrv) GetOrderInfo(ctx context.Context, orderUuid uint64) (*resp.OrderInfoResp, error) {
    // 构建缓存 key
    cacheKey := fmt.Sprintf(constant.CacheKeyOrderInfo, orderUuid)
    
    // 从缓存获取
    var orderInfo resp.OrderInfoResp
    if err := s.cache.Get(ctx, cacheKey, &orderInfo); err == nil {
        return &orderInfo, nil
    }
    
    // 缓存未命中，从数据库查询
    // ...
    
    // 写入缓存
    s.cache.Set(ctx, cacheKey, orderInfo, time.Hour)
    
    return &orderInfo, nil
}
```

```go
// ❌ 错误：硬编码缓存 key
func (s *orderSrv) GetOrderInfo(ctx context.Context, orderUuid uint64) (*resp.OrderInfoResp, error) {
    // ❌ 禁止硬编码
    cacheKey := fmt.Sprintf("order:info:%d", orderUuid)
    
    // ...
}
```

**命名规范**:

1. **前缀分类**: 使用模块前缀，如 `order:`, `product:`, `member:`
2. **功能标识**: 使用功能标识，如 `:info`, `:list`, `:status`
3. **参数占位**: 使用 `%d` 或 `%s` 作为参数占位符
4. **常量命名**: 使用 `CacheKey` 前缀，大驼峰命名

**规范**:
- ✅ **所有缓存 key 必须在 `app/constant/cache_key.go` 中定义**
- ✅ **使用 `fmt.Sprintf` 构建带参数的 key**
- ✅ **缓存 key 命名清晰，包含模块和功能信息**
- ❌ **禁止在代码中硬编码缓存 key 字符串**
- ❌ **禁止使用字符串拼接构建缓存 key**

**优势**:
- 统一管理，便于维护和修改
- 避免 key 冲突和拼写错误
- 提高代码可读性和可维护性
- 便于全局搜索和替换

---

## 设计原则

### 1. 单一职责原则 (SRP)

每个层、每个类都应该只有一个职责。

- **Handler**: 只负责 HTTP 请求处理
- **Service**: 只负责业务逻辑
- **Repository**: 只负责数据访问

### 2. 依赖倒置原则 (DIP)

高层模块不应该依赖低层模块，都应该依赖抽象。

```go
// ✅ 依赖接口
type orderSrv struct {
    orderRepo    repository.IOrderRepo    // 接口
    memberSrv    IMemberSrv               // 接口
}

// ❌ 依赖具体实现
type orderSrv struct {
    orderRepo    *repository.OrderRepoImpl  // 具体类
}
```

### 3. 接口隔离原则 (ISP)

接口应该小而专注，不要设计臃肿的接口。

```go
// ✅ 小接口
type IOrderReader interface {
    GetOrder(opts ...DBOption) (*model.Order, error)
    GetOrderList(opts ...DBOption) ([]model.Order, int64, error)
}

type IOrderWriter interface {
    CreateOrder(order model.Order) error
    UpdateOrder(order model.Order, opts ...DBOption) error
}

// 组合接口
type IOrderRepo interface {
    IOrderReader
    IOrderWriter
}
```

### 4. 开闭原则 (OCP)

对扩展开放，对修改关闭。

**示例**: 选项模式允许添加新的查询条件而不修改现有代码

```go
// 扩展新的查询条件
func (r *OrderRepoImpl) WherePaymentMethod(method string) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("payment_method = ?", method)
    }
}

// 无需修改现有代码，直接使用
orders, _ := orderRepo.GetOrderList(
    orderRepo.WhereStatus(1),
    orderRepo.WherePaymentMethod("wechat"),  // 新条件
)
```

---

## 最佳实践总结

### ✅ 应该做的

1. **分层清晰**: Handler → Service → Repository → Model
2. **依赖注入**: 通过构造函数注入所有依赖
3. **接口驱动**: Service 和 Repository 都定义接口
4. **事务管理**: 在 Service 层统一管理事务
5. **错误处理**: 使用 `errors.WithMessage` 包装错误
6. **异步事件**: 使用 goroutine 发布事件
7. **并发安全**: 使用 UUID 锁保护共享资源

### ❌ 不应该做的

1. **跨层调用**: Handler 不能直接调用 Repository
2. **循环依赖**: Service 之间避免循环依赖
3. **业务逻辑**: 不要在 Handler 或 Repository 中写业务逻辑
4. **使用 panic**: 使用 error 返回错误，不要 panic
5. **硬编码**: 不要硬编码配置，使用配置文件
6. **同步事件**: 不要同步处理耗时的事件

---

## 相关文档

- [Go Main 开发指南](../guides/go-main-development.md) - 详细的开发规范和代码示例
- [数据库设计](./database-design.md) - 数据库架构和多租户设计
- [API 设计指南](../guides/api-design-guide.md) - RESTful API 设计规范
- [测试标准](../testing/standards/go-testing.md) - Go 代码测试规范

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

