# 外卖订单信息查询服务 技术设计

> **状态**: 🚧 设计中
> **版本**: v1.0.0
> **创建日期**: 2025-12-22
> **更新日期**: 2025-12-22

---

## 📋 设计概览

| 项目       | 内容                              |
| ---------- | --------------------------------- |
| **Spec 名称** | story-takeout-order-query-service |
| **模块**   | ttpos-takeout (BMP)             |
| **技术栈** | Go + GoFrame + gRPC + MySQL     |
| **复杂度** | 低                               |
| **预估工作量** | 0.5 天 (SP: 1)                  |

**来源需求**: [requirements.md](./requirements.md)

---

## 🏗️ 架构设计

### 系统架构图

```
┌─────────────────┐    gRPC     ┌──────────────────────┐
│   TTPOS Main    │────────────▶│  ttpos-takeout BMP   │
│   (调用方)      │             │  (订单查询服务)     │
└─────────────────┘             └──────────────────────┘
                                       │
                                       ▼
                               ┌──────────────────────┐
                               │   MySQL Database     │
                               │   ttpos_takeout.order│
                               └──────────────────────┘
```

### 技术选型

- **框架**: GoFrame v2.x (微服务框架)
- **通信协议**: gRPC + Protocol Buffers
- **数据库**: MySQL 8.0+
- **ORM**: GoFrame ORM
- **配置管理**: GoFrame Config
- **日志**: GoFrame Log (支持 requestId 追踪)

---

## 📊 数据模型设计

### 数据库表结构

**表名**: `ttpos_takeout.order`

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 说明 |
|--------|------|------|----------|--------|------|
| id | bigint | 20 | NO | AUTO_INCREMENT | 主键ID |
| shop_uuid | varchar | 36 | NO | - | TTPOS店铺UUID |
| order_uuid | varchar | 36 | NO | - | TTPOS订单UUID |
| order_uuid | varchar | 36 | NO | - | TTPOS订单UUID |
| order_status | varchar | 32 | NO | - | 订单状态 |
| order_type | varchar | 32 | NO | - | 订单类型 |
| raw_data | json | - | YES | NULL | 原始平台数据 |
| created_at | datetime | - | NO | CURRENT_TIMESTAMP | 创建时间 |
| updated_at | datetime | - | NO | CURRENT_TIMESTAMP ON UPDATE | 更新时间 |

### 索引设计

```sql
-- 主键索引
PRIMARY KEY (`id`)

-- 业务查询索引（组合索引）
UNIQUE KEY `uk_shop_uuid_order_uuid` (`shop_uuid`, `order_uuid`)

-- 性能优化索引
KEY `idx_status` (`order_status`)
KEY `idx_created_at` (`created_at`)
```

---

## 🔌 接口设计

### Protobuf 定义

**文件位置**: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`

```protobuf
syntax = "proto3";

package ttpos.takeout.order.v1;

option go_package = "github.com/ttpos/ttpos-bmp/app/ttpos-takeout/api/ttpos/takeout/order/v1;v1";

// 获取订单信息请求
message GetOrderInfoReq {
  string shop_uuid = 1;    // TTPOS店铺UUID
  string order_uuid = 2;   // TTPOS订单UUID
  string request_id = 3;   // 请求追踪ID (可选)
}

// 获取订单信息响应
message GetOrderInfoResp {
  string shop_uuid = 1;     // TTPOS店铺UUID
  string order_status = 2;  // 订单状态
  string order_type = 3;    // 订单类型
  string raw_data = 4;      // 原始JSON数据
  string provider_name = 5; // 渠道名称: grab, foodpanda
}

// 订单服务
service OrderService {
  // 获取订单信息
  rpc GetOrderInfo(GetOrderInfoReq) returns (takeout.ApiResponse);
}
```

### gRPC 错误码设计

```go
// 错误码定义
const (
    OrderNotFound    = 10001 // 订单不存在
    InvalidParameter = 10002 // 参数无效
    DatabaseError    = 10003 // 数据库错误
    InternalError    = 10004 // 内部错误
)
```

---

## 🏛️ 代码结构设计

### 目录结构

```
ttpos-bmp/app/ttpos-takeout/
├── api/
│   └── ttpos/takeout/order/v1/          # gRPC API定义
│       ├── order.pb.go                  # 自动生成
│       └── order_grpc.pb.go            # 自动生成
├── internal/
│   ├── controller/                      # gRPC控制器
│   │   └── order_v1_controller.go      # 订单控制器
│   ├── logic/                          # 业务逻辑层
│   │   └── order/                      # 订单业务逻辑
│   │       └── get_order_info_logic.go # 获取订单信息逻辑
│   ├── model/                          # 数据模型
│   │   └── entity/                     # 实体定义
│   │       └── order.go                # 订单实体
│   └── dao/                            # 数据访问层
│       └── order.go                    # 订单DAO
├── manifest/
│   └── protobuf/
│       └── order/
│           └── order.proto             # Protobuf定义
└── router/                             # 路由配置
    └── order.go                        # 订单路由
```

### 核心文件设计

#### 1. 实体模型 (internal/model/entity/order.go)

```go
package entity

type Order struct {
    Id         int64     `json:"id" db:"id"`
    ShopUuid   string    `json:"shop_uuid" db:"shop_uuid"`
    OrderUuid  string    `json:"order_uuid" db:"order_uuid"`
    OrderStatus string   `json:"order_status" db:"order_status"`
    OrderType   string   `json:"order_type" db:"order_type"`
    RawData    string    `json:"raw_data" db:"raw_data"` // JSON字符串
    CreatedAt  time.Time `json:"created_at" db:"created_at"`
    UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
```

#### 2. DAO 层

**注意**: 根据 GoFrame 规范，DAO 层文件为框架自动生成，不应添加自定义方法。
数据库查询直接在 Logic 层使用 DAO 链式调用实现。

#### 3. 业务逻辑层 (internal/logic/order/get_order_info_logic.go)

```go
package logic

type GetOrderInfoLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetOrderInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderInfoLogic {
    return &GetOrderInfoLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

func (s *sOrder) GetOrderInfo(ctx context.Context, req *api.GetOrderInfoReq) (res *api.GetOrderInfoResp, err error) {
    // 记录请求ID到日志
    if req.RequestId != "" {
        g.Log().Infof(ctx, "GetOrderInfo start, requestId: %s, shopUuid: %s, orderUuid: %s", 
            req.RequestId, req.ShopUuid, req.OrderUuid)
    }

    // 直接使用 DAO 链式调用查询订单
    var orderEntity *entity.Order
    err = dao.Order.Ctx(ctx).
        Where(dao.Order.Columns().ShopUuid, req.ShopUuid).
        Where(dao.Order.Columns().Uuid, req.OrderUuid).
        Scan(&orderEntity)
    if err != nil {
        return nil, gerror.Wrap(err, "查询订单失败")
    }

    if orderEntity == nil {
        return nil, gerror.New("订单不存在")
    }

    return &api.GetOrderInfoResp{
        ShopUuid:     orderEntity.ShopUuid,
        OrderStatus:  orderEntity.OrderStatus,
        OrderType:    orderEntity.OrderType,
        RawData:      orderEntity.RawData,
        ProviderName: orderEntity.ProviderName,
    }, nil
}
```

#### 4. gRPC 控制器 (internal/controller/order_v1_controller.go)

```go
package order

import (
    "context"
    
    "github.com/gogf/gf/contrib/rpc/grpcx/v2"
    "github.com/gogf/gf/v2/frame/g"
    "google.golang.org/protobuf/types/known/anypb"
    
    api "ttpos-bmp/app/ttpos-takeout/api/order"
    "ttpos-bmp/app/ttpos-takeout/api/takeout"
    "ttpos-bmp/app/ttpos-takeout/internal/consts"
    "ttpos-bmp/app/ttpos-takeout/internal/service"
)

type Controller struct {
    api.UnimplementedOrderServiceServer
}

func Register(s *grpcx.GrpcServer) {
    api.RegisterOrderServiceServer(s.Server, &Controller{})
}

// GetOrderInfo 获取订单信息
func (c *Controller) GetOrderInfo(ctx context.Context, req *api.GetOrderInfoReq) (*takeout.ApiResponse, error) {
    res, err := service.Order().GetOrderInfo(ctx, req)
    if err != nil {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeServiceError),
            Message: err.Error(),
        }, nil
    }

    // 将 res 转换为 anypb.Any
    dataAny, err := anypb.New(res)
    if err != nil {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeSerializeError),
            Message: consts.MsgSerializeFailed,
        }, nil
    }

    g.Log().Debugf(ctx, "GetOrderInfo success: %+v", res)
    return &takeout.ApiResponse{
        Code:    string(consts.CodeSuccess),
        Message: consts.MsgSuccess,
        Data:    dataAny,
    }, nil
}
```

---

## 🔧 实现细节

### 依赖注入

使用 GoFrame 的依赖注入容器管理服务依赖：

```go
// internal/svc/service_context.go
type ServiceContext struct {
    Config    config.Config
    OrderDao  dao.OrderDao
    // 其他依赖...
}

// internal/config/config.go
type Config struct {
    Database struct {
        Host     string
        Port     int
        User     string
        Password string
        Database string
    }
    GRPC struct {
        Host string
        Port int
    }
}
```

### 日志追踪

支持 requestId 透传，实现全链路追踪：

```go
// 日志中间件
func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // 从请求中提取 request_id
    if reqWithId, ok := req.(interface{ GetRequestId() string }); ok {
        requestId := reqWithId.GetRequestId()
        if requestId != "" {
            ctx = context.WithValue(ctx, "request_id", requestId)
        }
    }

    // 调用实际处理逻辑
    return handler(ctx, req)
}
```

### 错误处理

统一的错误处理机制：

```go
// internal/errorx/code.go
const (
    Success       = 0
    OrderNotFound = 10001
    InvalidParam  = 10002
    DatabaseError = 10003
    InternalError = 10004
)

func NewCodeError(code int, message string) error {
    return &CodeError{
        Code:    code,
        Message: message,
    }
}
```

---

## 🧪 测试设计

### 单元测试

1. **DAO 层测试**
   - 正常查询场景
   - 订单不存在场景
   - 数据库连接异常场景

2. **Logic 层测试**
   - 业务逻辑正确性
   - 参数验证
   - 错误处理

3. **Controller 层测试**
   - gRPC 接口调用
   - 请求响应转换

### 集成测试

1. **数据库集成测试**
   - 真实数据库查询
   - 索引性能测试

2. **gRPC 接口测试**
   - 端到端调用测试
   - 并发请求测试

### 测试覆盖率目标

- **单元测试**: > 80%
- **集成测试**: > 90%

---

## 📈 性能优化

### 数据库优化

1. **索引优化**
   - 组合唯一索引: `(shop_uuid, order_uuid)`
   - 查询性能: < 10ms

2. **查询优化**
   - 只查询必需字段
   - 避免全表扫描

### 缓存策略 (预留)

如需要更高性能，可考虑添加 Redis 缓存：

```go
// 缓存 Key 设计
cacheKey := fmt.Sprintf("order:info:%s:%s", shopUuid, shopRefNo)

// 缓存过期时间: 5分钟
ttl := 5 * time.Minute
```

### 监控指标

- **响应时间**: P95 < 100ms
- **成功率**: > 99.9%
- **QPS**: 预计 100+ (根据业务规模)

---

## 🚀 部署考虑

### 配置管理

```yaml
# config.yaml
database:
  host: "localhost"
  port: 3306
  user: "ttpos"
  password: "password"
  database: "ttpos_takeout"

grpc:
  host: "0.0.0.0"
  port: 8080
```

### 健康检查

```go
// 健康检查接口
func (c *OrderController) Health(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
    // 检查数据库连接
    if err := g.DB().Ping(ctx); err != nil {
        return nil, status.Error(codes.Unavailable, "database unhealthy")
    }
    return &emptypb.Empty{}, nil
}
```

---

## 🔒 安全考虑

### 数据安全

- 通过内部 gRPC 调用，无外部暴露
- 请求参数校验
- SQL 注入防护 (使用 ORM)

### 访问控制

- 内部服务间调用
- 无需额外认证授权

---

## 📋 验收标准

### 功能验收

- [ ] gRPC 接口可正常调用
- [ ] 返回字段完整准确
- [ ] 错误处理正确
- [ ] requestId 日志追踪正常

### 性能验收

- [ ] 查询响应时间 < 100ms
- [ ] 支持并发请求
- [ ] 内存使用正常

### 代码质量验收

- [ ] 单元测试覆盖率 > 80%
- [ ] 代码规范检查通过
- [ ] 文档完善

---

**设计完成日期**: 2025-12-22  
**设计师**: 系统自动生成
**评审人**: 待定
**最后更新**: 2025-12-22 (方案调整: shopRefNo → orderUuid)
